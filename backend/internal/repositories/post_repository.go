package repositories

import (
	"context"
	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var postRepoTracer = otel.Tracer("post_repository")

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
func (r *PostRepository) GetPostsByDateRange(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]db.PostWithAuthor, error) {
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

	var posts []db.PostWithAuthor
	err := r.db.SelectContext(ctx, &posts, query, userID, dateFrom, dateTo)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch posts by date range: %w", err)
	}

	span.SetAttributes(attribute.Int("post_count", len(posts)))

	return posts, nil
}

// PostExists checks if a post already exists for a user
func (r *PostRepository) PostExists(ctx context.Context, userID uuid.UUID, postID int64) (bool, error) {
	ctx, span := postRepoTracer.Start(ctx, "PostExists")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int64("post_id", postID),
	)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM posts
			WHERE user_id = $1 AND x_post_id = $2
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, postID)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}

	return exists, nil
}

// InsertPost inserts a new post for a user
func (r *PostRepository) InsertPost(ctx context.Context, userID uuid.UUID, tweetDTO *dto.TweetDTO) error {
	ctx, span := postRepoTracer.Start(ctx, "InsertPost")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int64("post_id", tweetDTO.ID),
		attribute.Int64("author_id", tweetDTO.AuthorID),
	)

	query := `
		INSERT INTO posts (
			user_id, x_post_id, author_id, published_at, url, text,
			conversation_id, ingested_at, first_visible_at, edited_seen
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), false
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		userID,
		tweetDTO.ID,
		tweetDTO.AuthorID,
		tweetDTO.PublishedAt,
		tweetDTO.URL,
		tweetDTO.Text,
		tweetDTO.ConversationID,
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to insert post: %w", err)
	}

	return nil
}
