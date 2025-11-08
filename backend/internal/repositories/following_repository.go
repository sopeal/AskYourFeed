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

// GetFollowing retrieves paginated list of authors the user follows
// Returns items ordered by x_author_id DESC for cursor-based pagination
func (r *FollowingRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursor *int64) ([]FollowingItem, error) {
	ctx, span := followingRepoTracer.Start(ctx, "GetFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int("limit", limit),
	)
	if cursor != nil {
		span.SetAttributes(attribute.Int64("cursor", *cursor))
	}

	var query string
	var args []interface{}

	if cursor == nil {
		// First page - no cursor
		query = `
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
			LIMIT $2
		`
		args = []interface{}{userID, limit}
	} else {
		// Subsequent pages - with cursor
		query = `
			SELECT 
				a.x_author_id,
				a.handle,
				a.display_name,
				a.last_seen_at,
				uf.last_checked_at
			FROM user_following uf
			INNER JOIN authors a ON uf.x_author_id = a.x_author_id
			WHERE uf.user_id = $1 AND a.x_author_id < $2
			ORDER BY a.x_author_id DESC
			LIMIT $3
		`
		args = []interface{}{userID, *cursor, limit}
	}

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
