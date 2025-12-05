package repositories

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
)

// SessionRepository handles database operations for sessions
type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*db.Session, error)
	GetSessionByToken(ctx context.Context, token string) (*db.Session, error)
	RevokeSession(ctx context.Context, token string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type sessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// hashToken creates a SHA-256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateSession creates a new session in the database
func (r *sessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*db.Session, error) {
	tokenHash := hashToken(token)

	query := `
		INSERT INTO sessions (user_id, token_hash, created_at, expires_at)
		VALUES ($1, $2, NOW(), $3)
		RETURNING id, user_id, token_hash, created_at, expires_at, revoked_at
	`

	var session db.Session
	err := r.db.QueryRowxContext(ctx, query, userID, tokenHash, expiresAt).StructScan(&session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByToken retrieves a session by token
func (r *sessionRepository) GetSessionByToken(ctx context.Context, token string) (*db.Session, error) {
	tokenHash := hashToken(token)

	query := `
		SELECT id, user_id, token_hash, created_at, expires_at, revoked_at
		FROM sessions
		WHERE token_hash = $1
		AND revoked_at IS NULL
		AND expires_at > NOW()
	`

	var session db.Session
	err := r.db.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found or expired
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return &session, nil
}

// RevokeSession revokes a session by token
func (r *sessionRepository) RevokeSession(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1
		AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `
		DELETE FROM sessions
		WHERE expires_at < NOW() - INTERVAL '30 days'
		OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days')
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}
