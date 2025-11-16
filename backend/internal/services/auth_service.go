package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/oauth2"
)

var authServiceTracer = otel.Tracer("auth_service")

// AuthService handles authentication business logic
type AuthService struct {
	authRepo     *repositories.AuthRepository
	oauthConfig  *oauth2.Config
	encryptionKey []byte
	baseURL       string
}

// OAuthTokenResponse represents the response from X OAuth token endpoint
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// XUserInfo represents user information from X API
type XUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// NewAuthService creates a new AuthService instance
func NewAuthService(authRepo *repositories.AuthRepository, oauthConfig *oauth2.Config, encryptionKey string, baseURL string) *AuthService {
	// Derive encryption key using PBKDF2
	salt := []byte("askyourfeed-auth-salt") // In production, use a proper salt from config
	key := pbkdf2.Key([]byte(encryptionKey), salt, 100000, 32, sha256.New)

	return &AuthService{
		authRepo:     authRepo,
		oauthConfig:  oauthConfig,
		encryptionKey: key,
		baseURL:       baseURL,
	}
}

// InitiateLogin initiates the OAuth login flow
func (s *AuthService) InitiateLogin(ctx context.Context) (*dto.AuthInitiateResponseDTO, error) {
	ctx, span := authServiceTracer.Start(ctx, "InitiateLogin")
	defer span.End()

	// Generate PKCE code verifier and challenge
	codeVerifier, err := repositories.GenerateSecureToken(32)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Create code challenge (SHA256 hash of verifier, base64url encoded)
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.EncodeToString(hash[:])
	codeChallenge = strings.TrimRight(codeChallenge, "=") // Remove padding

	// Generate state token for CSRF protection
	stateToken, err := repositories.GenerateSecureToken(32)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate state token: %w", err)
	}

	// Store OAuth state
	oauthState := &repositories.OAuthState{
		StateToken:    stateToken,
		UserID:        nil, // Anonymous initiation
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		RedirectURI:   s.baseURL + "/api/v1/auth/callback",
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(10 * time.Minute), // 10 minutes expiry
	}

	if err := s.authRepo.StoreOAuthState(ctx, oauthState); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to store OAuth state: %w", err)
	}

	// Build authorization URL
	authURL := s.oauthConfig.AuthCodeURL(stateToken,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	span.SetAttributes(
		attribute.String("state_token", stateToken[:8]+"..."),
		attribute.String("redirect_uri", oauthState.RedirectURI),
	)

	return &dto.AuthInitiateResponseDTO{
		AuthURL: authURL,
		State:   stateToken,
	}, nil
}

// HandleCallback processes the OAuth callback
func (s *AuthService) HandleCallback(ctx context.Context, code, state string) (*dto.AuthCallbackResponseDTO, error) {
	ctx, span := authServiceTracer.Start(ctx, "HandleCallback")
	defer span.End()

	span.SetAttributes(
		attribute.String("state", state[:8]+"..."),
	)

	// Validate state parameter
	oauthState, err := s.authRepo.GetOAuthState(ctx, state)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid or expired state: %w", err)
	}

	// Exchange authorization code for tokens
	token, err := s.exchangeCodeForToken(ctx, code, oauthState.CodeVerifier, oauthState.RedirectURI)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from X API
	userInfo, err := s.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Generate user ID (in production, this might be based on X user ID or existing user mapping)
	userID := uuid.New()

	// Encrypt tokens
	encryptedAccessToken, err := s.encryptToken(token.AccessToken)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken *string
	if token.RefreshToken != "" {
		encrypted, err := s.encryptToken(token.RefreshToken)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		encryptedRefreshToken = &encrypted
	}

	// Calculate token expiry times
	now := time.Now()
	var accessTokenExpiresAt *time.Time
	if token.Expiry.After(now) {
		accessTokenExpiresAt = &token.Expiry
	}

	var refreshTokenExpiresAt *time.Time
	if token.RefreshToken != "" {
		// Assume refresh tokens last 6 months (X doesn't specify)
		refreshExpiry := now.Add(6 * 30 * 24 * time.Hour)
		refreshTokenExpiresAt = &refreshExpiry
	}

	// Store OAuth tokens
	oauthTokens := &repositories.UserOAuthTokens{
		UserID:                userID,
		XUserID:               userInfo.ID,
		XHandle:               userInfo.Username,
		XDisplayName:          &userInfo.Name,
		EncryptedAccessToken:  encryptedAccessToken,
		EncryptedRefreshToken: encryptedRefreshToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.authRepo.StoreUserOAuthTokens(ctx, oauthTokens); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to store OAuth tokens: %w", err)
	}

	// Generate session token
	sessionToken, err := repositories.GenerateSecureToken(32)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create user session
	session := &repositories.UserSession{
		SessionToken:            sessionToken,
		UserID:                  userID,
		XUserID:                 userInfo.ID,
		XHandle:                 userInfo.Username,
		XDisplayName:            &userInfo.Name,
		EncryptedAccessToken:    encryptedAccessToken,
		EncryptedRefreshToken:   encryptedRefreshToken,
		AccessTokenExpiresAt:    accessTokenExpiresAt,
		RefreshTokenExpiresAt:   refreshTokenExpiresAt,
		AuthenticatedAt:         now,
		CreatedAt:               now,
		ExpiresAt:               now.Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.authRepo.StoreUserSession(ctx, session); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Clean up OAuth state
	if err := s.authRepo.DeleteOAuthState(ctx, state); err != nil {
		// Log but don't fail the request
		span.RecordError(err)
	}

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("x_handle", userInfo.Username),
	)

	return &dto.AuthCallbackResponseDTO{
		SessionToken: sessionToken,
		UserID:       userID,
		XHandle:      userInfo.Username,
		RedirectURL:  "/dashboard", // Frontend redirect URL
	}, nil
}

// Logout terminates the user session
func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	ctx, span := authServiceTracer.Start(ctx, "Logout")
	defer span.End()

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	if err := s.authRepo.DeleteUserSession(ctx, sessionToken); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetCurrentSession retrieves current session information
func (s *AuthService) GetCurrentSession(ctx context.Context, sessionToken string) (*dto.SessionDTO, error) {
	ctx, span := authServiceTracer.Start(ctx, "GetCurrentSession")
	defer span.End()

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	session, err := s.authRepo.GetUserSession(ctx, sessionToken)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	span.SetAttributes(
		attribute.String("user_id", session.UserID.String()),
		attribute.String("x_handle", session.XHandle),
	)

	return &dto.SessionDTO{
		UserID:           session.UserID,
		XHandle:          session.XHandle,
		XDisplayName:     s.stringPtrToString(session.XDisplayName),
		AuthenticatedAt:  session.AuthenticatedAt,
		SessionExpiresAt: session.ExpiresAt,
	}, nil
}

// ValidateSession validates a session token and returns user ID
func (s *AuthService) ValidateSession(ctx context.Context, sessionToken string) (uuid.UUID, error) {
	ctx, span := authServiceTracer.Start(ctx, "ValidateSession")
	defer span.End()

	session, err := s.authRepo.GetUserSession(ctx, sessionToken)
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, fmt.Errorf("invalid session: %w", err)
	}

	span.SetAttributes(
		attribute.String("user_id", session.UserID.String()),
		attribute.String("x_handle", session.XHandle),
	)

	return session.UserID, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (s *AuthService) exchangeCodeForToken(ctx context.Context, code, codeVerifier, redirectURI string) (*oauth2.Token, error) {
	ctx, span := authServiceTracer.Start(ctx, "exchangeCodeForToken")
	defer span.End()

	// Prepare token exchange request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.oauthConfig.ClientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", s.oauthConfig.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add client credentials if needed
	if s.oauthConfig.ClientSecret != "" {
		req.SetBasicAuth(s.oauthConfig.ClientID, s.oauthConfig.ClientSecret)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		span.RecordError(fmt.Errorf("token exchange failed: %s", string(body)))
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Convert to oauth2.Token
	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
	}

	if tokenResp.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return token, nil
}

// getUserInfo retrieves user information from X API
func (s *AuthService) getUserInfo(ctx context.Context, accessToken string) (*XUserInfo, error) {
	ctx, span := authServiceTracer.Start(ctx, "getUserInfo")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitter.com/2/users/me", nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		span.RecordError(fmt.Errorf("user info request failed: %s", string(body)))
		return nil, fmt.Errorf("user info request failed with status %d", resp.StatusCode)
	}

	var xResp struct {
		Data XUserInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&xResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return &xResp.Data, nil
}

// encryptToken encrypts a token using AES-256-GCM
func (s *AuthService) encryptToken(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken decrypts a token using AES-256-GCM
func (s *AuthService) decryptToken(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted token: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return string(plaintext), nil
}

// stringPtrToString converts *string to string, returning empty string if nil
func (s *AuthService) stringPtrToString(sPtr *string) string {
	if sPtr == nil {
		return ""
	}
	return *sPtr
}
