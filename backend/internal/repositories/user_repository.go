package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
)

// UserRepository handles database operations for users
type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, xUsername, xDisplayName string) (*db.User, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*db.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *userRepository) CreateUser(ctx context.Context, email, passwordHash, xUsername, xDisplayName string) (*db.User, error) {
	query := `
		INSERT INTO users (email, password_hash, x_username, x_display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, email, password_hash, x_username, x_display_name, created_at, updated_at
	`

	var user db.User
	err := r.db.QueryRowxContext(ctx, query, email, passwordHash, xUsername, xDisplayName).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	query := `
		SELECT id, email, password_hash, x_username, x_display_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user db.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *userRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*db.User, error) {
	query := `
		SELECT id, email, password_hash, x_username, x_display_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user db.User
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// EmailExists checks if an email already exists in the database
func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// GetFollowingCount returns the count of authors the user is following
func (r *userRepository) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_following
		WHERE user_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get following count: %w", err)
	}

	return count, nil
}
