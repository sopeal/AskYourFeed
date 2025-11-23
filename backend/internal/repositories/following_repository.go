package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var followingRepoTracer = otel.Tracer("following_repository")

// FollowingRepository handles user_following and authors data access operations
type FollowingRepository struct {
	db *sqlx.DB
}

// NewFollowingRepository creates a new FollowingRepository instance
func NewFollowingRepository(database *sqlx.DB) *FollowingRepository {
	return &FollowingRepository{
		db: database,
	}
}

// FollowingItem represents a joined result from user_following and authors tables
type FollowingItem struct {
	XAuthorID     int64      `db:"x_author_id"`
	Handle        string     `db:"handle"`
	DisplayName   *string    `db:"display_name"`
	LastSeenAt    *time.Time `db:"last_seen_at"`
	LastCheckedAt *time.Time `db:"last_checked_at"`
}

// GetFollowing retrieves list of authors the user follows
// Returns items ordered by x_author_id DESC
func (r *FollowingRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]FollowingItem, error) {
	ctx, span := followingRepoTracer.Start(ctx, "GetFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
	)

	query := `
		SELECT
			a.x_author_id,
			a.handle,
			a.display_name,
			a.last_seen_at,
			uf.last_checked_at
		FROM user_following uf
		INNER JOIN authors a ON uf.x_author_id = a.x_author_id
		WHERE uf.user_id = $1
		ORDER BY a.x_author_id DESC
	`
	args := []interface{}{userID}

	var items []FollowingItem
	err := r.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch following list: %w", err)
	}

	// Return empty slice if no items found (not an error)
	if items == nil {
		items = []FollowingItem{}
	}

	span.SetAttributes(attribute.Int("items_found", len(items)))

	return items, nil
}

// UpsertFollowing inserts or updates a following relationship
func (r *FollowingRepository) UpsertFollowing(ctx context.Context, userID uuid.UUID, authorID int64, lastCheckedAt time.Time) error {
	ctx, span := followingRepoTracer.Start(ctx, "UpsertFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int64("author_id", authorID),
	)

	query := `
		INSERT INTO user_following (user_id, x_author_id, last_checked_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, x_author_id) DO UPDATE SET
			last_checked_at = EXCLUDED.last_checked_at
	`

	_, err := r.db.ExecContext(ctx, query, userID, authorID, lastCheckedAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upsert following: %w", err)
	}

	return nil
}

// RemoveFollowing removes a following relationship
func (r *FollowingRepository) RemoveFollowing(ctx context.Context, userID uuid.UUID, authorID int64) error {
	ctx, span := followingRepoTracer.Start(ctx, "RemoveFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int64("author_id", authorID),
	)

	query := `
		DELETE FROM user_following
		WHERE user_id = $1 AND x_author_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, authorID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to remove following: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("following relationship not found")
	}

	return nil
}

// GetTotalFollowingCount retrieves total count of authors the user follows
func (r *FollowingRepository) GetTotalFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, span := followingRepoTracer.Start(ctx, "GetTotalFollowingCount")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	query := `
		SELECT COUNT(*)
		FROM user_following
		WHERE user_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to fetch total following count: %w", err)
	}

	span.SetAttributes(attribute.Int("total_count", count))

	return count, nil
}
