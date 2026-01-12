package repositories

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var authorRepoTracer = otel.Tracer("author_repository")

// AuthorRepository handles author data access operations
type AuthorRepository struct {
	db *sqlx.DB
}

// NewAuthorRepository creates a new AuthorRepository instance
func NewAuthorRepository(database *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		db: database,
	}
}

// GetAuthor retrieves an author by ID
func (r *AuthorRepository) GetAuthor(ctx context.Context, authorID int64) (*db.Author, error) {
	ctx, span := authorRepoTracer.Start(ctx, "GetAuthor")
	defer span.End()

	span.SetAttributes(attribute.Int64("author_id", authorID))

	query := `
		SELECT x_author_id, handle, display_name, last_seen_at
		FROM authors
		WHERE x_author_id = $1
	`

	var author db.Author
	err := r.db.GetContext(ctx, &author, query, authorID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // Author not found
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	return &author, nil
}

// GetAuthorByHandle retrieves an author by handle
func (r *AuthorRepository) GetAuthorByHandle(ctx context.Context, handle string) (*db.Author, error) {
	ctx, span := authorRepoTracer.Start(ctx, "GetAuthorByHandle")
	defer span.End()

	span.SetAttributes(attribute.String("handle", handle))

	query := `
		SELECT x_author_id, handle, display_name, last_seen_at
		FROM authors
		WHERE handle = $1
	`

	var author db.Author
	err := r.db.GetContext(ctx, &author, query, handle)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // Author not found
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get author by handle: %w", err)
	}

	return &author, nil
}

// InsertAuthor inserts a new author
func (r *AuthorRepository) InsertAuthor(ctx context.Context, userDTO *dto.UserDTO) (int64, error) {
	ctx, span := authorRepoTracer.Start(ctx, "InsertAuthor")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("author_id", userDTO.ID),
		attribute.String("handle", userDTO.Handle),
	)

	query := `
		INSERT INTO authors (x_author_id, handle, display_name, last_seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (x_author_id) DO UPDATE SET
			handle = EXCLUDED.handle,
			display_name = EXCLUDED.display_name,
			last_seen_at = EXCLUDED.last_seen_at
		RETURNING x_author_id
	`

	var authorID int64
	err := r.db.GetContext(ctx, &authorID, query,
		userDTO.ID,
		userDTO.Handle,
		userDTO.DisplayName,
		userDTO.LastSeenAt,
	)

	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to insert author: %w", err)
	}

	return authorID, nil
}

// UpdateAuthorLastSeen updates the last seen timestamp for an author
func (r *AuthorRepository) UpdateAuthorLastSeen(ctx context.Context, authorID int64, lastSeenAt interface{}) error {
	ctx, span := authorRepoTracer.Start(ctx, "UpdateAuthorLastSeen")
	defer span.End()

	span.SetAttributes(attribute.Int64("author_id", authorID))

	query := `
		UPDATE authors
		SET last_seen_at = $2
		WHERE x_author_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, authorID, lastSeenAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update author last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("author not found: %d", authorID)
	}

	return nil
}
