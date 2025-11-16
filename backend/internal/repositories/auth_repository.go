package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var authRepoTracer = otel.Tracer("auth_repository")

// AuthRepository handles authentication-related data access operations
type AuthRepository struct {
	db *sqlx.DB
}

// NewAuthRepository creates a new AuthRepository instance
func NewAuthRepository(database *sqlx.DB) *AuthRepository {
	return &AuthRepository{
		db: database,
	}
}

// OAuthState represents OAuth state data for PKCE flow
type OAuthState struct {
	StateToken    string
	UserID        *uuid.UUID
	CodeVerifier  string
	CodeChallenge string
	RedirectURI   string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

// UserSession represents a user session
type UserSession struct {
	SessionToken            string
	UserID                  uuid.UUID
	XUserID                 string
	XHandle                 string
	XDisplayName            *string
	EncryptedAccessToken    string
	EncryptedRefreshToken   *string
	AccessTokenExpiresAt    *time.Time
	RefreshTokenExpiresAt   *time.Time
	AuthenticatedAt         time.Time
	CreatedAt               time.Time
	ExpiresAt               time.Time
}

// UserOAuthTokens represents stored OAuth tokens for a user
type UserOAuthTokens struct {
	UserID                uuid.UUID
	XUserID               string
	XHandle               string
	XDisplayName          *string
	EncryptedAccessToken  string
	EncryptedRefreshToken *string
	AccessTokenExpiresAt  *time.Time
	RefreshTokenExpiresAt *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// StoreOAuthState stores OAuth state for PKCE flow
func (r *AuthRepository) StoreOAuthState(ctx context.Context, state *OAuthState) error {
	ctx, span := authRepoTracer.Start(ctx, "StoreOAuthState")
	defer span.End()

	span.SetAttributes(
		attribute.String("state_token", state.StateToken[:8]+"..."), // Partial for security
	)

	query := `
		INSERT INTO oauth_state (
			state_token, user_id, code_verifier, code_challenge,
			redirect_uri, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		state.StateToken, state.UserID, state.CodeVerifier,
		state.CodeChallenge, state.RedirectURI, state.CreatedAt, state.ExpiresAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to store OAuth state: %w", err)
	}

	return nil
}

// GetOAuthState retrieves OAuth state by token
func (r *AuthRepository) GetOAuthState(ctx context.Context, stateToken string) (*OAuthState, error) {
	ctx, span := authRepoTracer.Start(ctx, "GetOAuthState")
	defer span.End()

	span.SetAttributes(
		attribute.String("state_token", stateToken[:8]+"..."),
	)

	query := `
		SELECT state_token, user_id, code_verifier, code_challenge,
			   redirect_uri, created_at, expires_at
		FROM oauth_state
		WHERE state_token = $1 AND expires_at > NOW()
	`

	var state OAuthState
	err := r.db.GetContext(ctx, &state, query, stateToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OAuth state not found or expired")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get OAuth state: %w", err)
	}

	return &state, nil
}

// DeleteOAuthState removes OAuth state after use
func (r *AuthRepository) DeleteOAuthState(ctx context.Context, stateToken string) error {
	ctx, span := authRepoTracer.Start(ctx, "DeleteOAuthState")
	defer span.End()

	span.SetAttributes(
		attribute.String("state_token", stateToken[:8]+"..."),
	)

	query := `DELETE FROM oauth_state WHERE state_token = $1`

	result, err := r.db.ExecContext(ctx, query, stateToken)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete OAuth state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OAuth state not found")
	}

	return nil
}

// StoreUserSession creates a new user session
func (r *AuthRepository) StoreUserSession(ctx context.Context, session *UserSession) error {
	ctx, span := authRepoTracer.Start(ctx, "StoreUserSession")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", session.UserID.String()),
		attribute.String("x_handle", session.XHandle),
	)

	query := `
		INSERT INTO user_sessions (
			session_token, user_id, x_user_id, x_handle, x_display_name,
			encrypted_access_token, encrypted_refresh_token,
			access_token_expires_at, refresh_token_expires_at,
			authenticated_at, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.SessionToken, session.UserID, session.XUserID, session.XHandle, session.XDisplayName,
		session.EncryptedAccessToken, session.EncryptedRefreshToken,
		session.AccessTokenExpiresAt, session.RefreshTokenExpiresAt,
		session.AuthenticatedAt, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to store user session: %w", err)
	}

	return nil
}

// GetUserSession retrieves session by token
func (r *AuthRepository) GetUserSession(ctx context.Context, sessionToken string) (*UserSession, error) {
	ctx, span := authRepoTracer.Start(ctx, "GetUserSession")
	defer span.End()

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	query := `
		SELECT session_token, user_id, x_user_id, x_handle, x_display_name,
			   encrypted_access_token, encrypted_refresh_token,
			   access_token_expires_at, refresh_token_expires_at,
			   authenticated_at, created_at, expires_at
		FROM user_sessions
		WHERE session_token = $1 AND expires_at > NOW()
	`

	var session UserSession
	err := r.db.GetContext(ctx, &session, query, sessionToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	return &session, nil
}

// DeleteUserSession removes a session (logout)
func (r *AuthRepository) DeleteUserSession(ctx context.Context, sessionToken string) error {
	ctx, span := authRepoTracer.Start(ctx, "DeleteUserSession")
	defer span.End()

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	query := `DELETE FROM user_sessions WHERE session_token = $1`

	result, err := r.db.ExecContext(ctx, query, sessionToken)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete user session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// StoreUserOAuthTokens stores or updates OAuth tokens for a user
func (r *AuthRepository) StoreUserOAuthTokens(ctx context.Context, tokens *UserOAuthTokens) error {
	ctx, span := authRepoTracer.Start(ctx, "StoreUserOAuthTokens")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", tokens.UserID.String()),
		attribute.String("x_handle", tokens.XHandle),
	)

	query := `
		INSERT INTO user_oauth_tokens (
			user_id, x_user_id, x_handle, x_display_name,
			encrypted_access_token, encrypted_refresh_token,
			access_token_expires_at, refresh_token_expires_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id) DO UPDATE SET
			x_user_id = EXCLUDED.x_user_id,
			x_handle = EXCLUDED.x_handle,
			x_display_name = EXCLUDED.x_display_name,
			encrypted_access_token = EXCLUDED.encrypted_access_token,
			encrypted_refresh_token = EXCLUDED.encrypted_refresh_token,
			access_token_expires_at = EXCLUDED.access_token_expires_at,
			refresh_token_expires_at = EXCLUDED.refresh_token_expires_at,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		tokens.UserID, tokens.XUserID, tokens.XHandle, tokens.XDisplayName,
		tokens.EncryptedAccessToken, tokens.EncryptedRefreshToken,
		tokens.AccessTokenExpiresAt, tokens.RefreshTokenExpiresAt,
		tokens.CreatedAt, tokens.UpdatedAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to store user OAuth tokens: %w", err)
	}

	return nil
}

// GetUserOAuthTokens retrieves OAuth tokens for a user
func (r *AuthRepository) GetUserOAuthTokens(ctx context.Context, userID uuid.UUID) (*UserOAuthTokens, error) {
	ctx, span := authRepoTracer.Start(ctx, "GetUserOAuthTokens")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
	)

	query := `
		SELECT user_id, x_user_id, x_handle, x_display_name,
			   encrypted_access_token, encrypted_refresh_token,
			   access_token_expires_at, refresh_token_expires_at,
			   created_at, updated_at
		FROM user_oauth_tokens
		WHERE user_id = $1
	`

	var tokens UserOAuthTokens
	err := r.db.GetContext(ctx, &tokens, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OAuth tokens not found for user")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user OAuth tokens: %w", err)
	}

	return &tokens, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CleanupExpiredStates removes expired OAuth states (should be called periodically)
func (r *AuthRepository) CleanupExpiredStates(ctx context.Context) error {
	ctx, span := authRepoTracer.Start(ctx, "CleanupExpiredStates")
	defer span.End()

	query := `DELETE FROM oauth_state WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to cleanup expired OAuth states: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.SetAttributes(attribute.Int64("states_cleaned", rowsAffected))
	return nil
}

// CleanupExpiredSessions removes expired sessions (should be called periodically)
func (r *AuthRepository) CleanupExpiredSessions(ctx context.Context) error {
	ctx, span := authRepoTracer.Start(ctx, "CleanupExpiredSessions")
	defer span.End()

	query := `DELETE FROM user_sessions WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.SetAttributes(attribute.Int64("sessions_cleaned", rowsAffected))
	return nil
}
