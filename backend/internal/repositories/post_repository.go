package repositories

import (
	"context"
	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var postRepoTracer = otel.Tracer("post_repository")

// PostWithAuthor represents a post with author information
type PostWithAuthor struct {
	db.Post
	Handle      string  `db:"handle"`
	DisplayName *string `db:"display_name"`
}

// PostRepository handles post data access operations
type PostRepository struct {
	db *sqlx.DB
}

// NewPostRepository creates a new PostRepository instance
func NewPostRepository(database *sqlx.DB) *PostRepository {
	return &PostRepository{
		db: database,
	}
}

// GetPostsByDateRange fetches posts within a specified date range for a user
// Returns posts ordered chronologically (published_at ASC)
// Uses RLS to ensure user can only access their own posts
func (r *PostRepository) GetPostsByDateRange(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]PostWithAuthor, error) {
	ctx, span := postRepoTracer.Start(ctx, "GetPostsByDateRange")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("date_from", dateFrom.Format(time.RFC3339)),
		attribute.String("date_to", dateTo.Format(time.RFC3339)),
	)

	query := `
		SELECT 
			p.user_id,
			p.x_post_id,
			p.author_id,
			p.published_at,
			p.url,
			p.text,
			p.conversation_id,
			p.ingested_at,
			p.first_visible_at,
			p.edited_seen,
			a.handle,
			a.display_name
		FROM posts p
		JOIN authors a ON p.author_id = a.x_author_id
		WHERE p.user_id = $1 
		  AND p.published_at >= $2 
		  AND p.published_at <= $3
		ORDER BY p.published_at ASC
		LIMIT 100
	`

	var posts []PostWithAuthor
	err := r.db.SelectContext(ctx, &posts, query, userID, dateFrom, dateTo)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch posts by date range: %w", err)
	}

	span.SetAttributes(attribute.Int("post_count", len(posts)))

	return posts, nil
}
