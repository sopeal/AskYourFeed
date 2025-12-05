package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/db"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService interface {
	Register(ctx context.Context, cmd dto.RegisterCommand) (*dto.RegisterResponseDTO, error)
	Login(ctx context.Context, cmd dto.LoginCommand) (*dto.LoginResponseDTO, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*db.User, error)
	GetCurrentSession(ctx context.Context, userID uuid.UUID) (*dto.SessionDTO, error)
}

type authService struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	twitterAPI  TwitterClient
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	twitterAPI TwitterClient,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		twitterAPI:  twitterAPI,
	}
}

const (
	sessionDuration = 7 * 24 * time.Hour // 7 days
	bcryptCost      = 12                 // bcrypt cost factor
)

// Register creates a new user account
func (s *authService) Register(ctx context.Context, cmd dto.RegisterCommand) (*dto.RegisterResponseDTO, error) {
	// Validate password confirmation
	if cmd.Password != cmd.PasswordConfirmation {
		return nil, fmt.Errorf("password confirmation does not match")
	}

	// Validate password strength
	if err := validatePasswordStrength(cmd.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate X username via Twitter API
	userInfo, err := s.twitterAPI.GetUserInfo(ctx, cmd.XUsername)
	if err != nil {
		return nil, fmt.Errorf("X username validation failed: %w", err)
	}

	// Hash password
	passwordHash, err := hashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, cmd.Email, passwordHash, userInfo.UserName, userInfo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	expiresAt := time.Now().Add(sessionDuration)
	_, err = s.sessionRepo.CreateSession(ctx, user.ID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &dto.RegisterResponseDTO{
		UserID:       user.ID,
		Email:        user.Email,
		XUsername:    user.XUsername,
		XDisplayName: user.XDisplayName,
		CreatedAt:    user.CreatedAt,
		SessionToken: token,
	}, nil
}

// Login authenticates a user and creates a session
func (s *authService) Login(ctx context.Context, cmd dto.LoginCommand) (*dto.LoginResponseDTO, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := verifyPassword(user.PasswordHash, cmd.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	expiresAt := time.Now().Add(sessionDuration)
	session, err := s.sessionRepo.CreateSession(ctx, user.ID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &dto.LoginResponseDTO{
		UserID:           user.ID,
		Email:            user.Email,
		XUsername:        user.XUsername,
		XDisplayName:     user.XDisplayName,
		SessionToken:     token,
		SessionExpiresAt: session.ExpiresAt,
	}, nil
}

// Logout revokes a session
func (s *authService) Logout(ctx context.Context, token string) error {
	if err := s.sessionRepo.RevokeSession(ctx, token); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// ValidateSession validates a session token and returns the user
func (s *authService) ValidateSession(ctx context.Context, token string) (*db.User, error) {
	// Get session by token
	session, err := s.sessionRepo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Get user by ID
	user, err := s.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetCurrentSession returns current session information
func (s *authService) GetCurrentSession(ctx context.Context, userID uuid.UUID) (*dto.SessionDTO, error) {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get following count
	followingCount, err := s.userRepo.GetFollowingCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following count: %w", err)
	}

	return &dto.SessionDTO{
		UserID:           user.ID,
		Email:            user.Email,
		XUsername:        user.XUsername,
		XDisplayName:     user.XDisplayName,
		AuthenticatedAt:  user.CreatedAt,
		SessionExpiresAt: time.Now().Add(sessionDuration),
		FollowingCount:   followingCount,
		FollowingLimit:   150, // Fixed limit as per requirements
	}, nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// validatePasswordStrength validates password meets requirements
// Requirements: min 8 chars, uppercase, lowercase, number, special char
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	)

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasNumber {
		missing = append(missing, "number")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain: %s", strings.Join(missing, ", "))
	}

	return nil
}
