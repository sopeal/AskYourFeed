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

var qaRepoTracer = otel.Tracer("qa_repository")

// QARepository handles Q&A data access operations
type QARepository struct {
	db *sqlx.DB
}

// NewQARepository creates a new QARepository instance
func NewQARepository(database *sqlx.DB) *QARepository {
	return &QARepository{
		db: database,
	}
}

// CreateQA inserts a new Q&A message record within a transaction
func (r *QARepository) CreateQA(ctx context.Context, tx *sqlx.Tx, qa db.QAMessage) error {
	ctx, span := qaRepoTracer.Start(ctx, "CreateQA")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("qa_id", qa.ID),
		attribute.String("user_id", qa.UserID.String()),
	)
	
	query := `
		INSERT INTO qa_messages (id, user_id, question, answer, date_from, date_to, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := tx.ExecContext(ctx, query, qa.ID, qa.UserID, qa.Question, qa.Answer, qa.DateFrom, qa.DateTo, qa.CreatedAt)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to insert Q&A message: %w", err)
	}
	
	return nil
}

// CreateQASources batch inserts Q&A source records within a transaction
func (r *QARepository) CreateQASources(ctx context.Context, tx *sqlx.Tx, sources []db.QASource) error {
	if len(sources) == 0 {
		return nil // No sources to insert
	}
	
	ctx, span := qaRepoTracer.Start(ctx, "CreateQASources")
	defer span.End()
	
	span.SetAttributes(attribute.Int("source_count", len(sources)))
	
	query := `
		INSERT INTO qa_sources (qa_id, user_id, x_post_id)
		VALUES (:qa_id, :user_id, :x_post_id)
	`
	
	_, err := tx.NamedExecContext(ctx, query, sources)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to insert Q&A sources: %w", err)
	}
	
	return nil
}

// GetQAByID retrieves a Q&A record with its sources
func (r *QARepository) GetQAByID(ctx context.Context, userID uuid.UUID, qaID string) (*dto.QADetailDTO, error) {
	ctx, span := qaRepoTracer.Start(ctx, "GetQAByID")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("qa_id", qaID),
		attribute.String("user_id", userID.String()),
	)
	
	// First, get the Q&A message
	var qa db.QAMessage
	qaQuery := `
		SELECT id, user_id, question, answer, date_from, date_to, created_at
		FROM qa_messages
		WHERE id = $1 AND user_id = $2
	`
	
	err := r.db.GetContext(ctx, &qa, qaQuery, qaID, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch Q&A message: %w", err)
	}
	
	// Then, get the sources with post and author information
	sourcesQuery := `
		SELECT 
			p.x_post_id,
			a.handle,
			a.display_name,
			p.published_at,
			p.url,
			p.text
		FROM qa_sources qs
		JOIN posts p ON qs.x_post_id = p.x_post_id AND qs.user_id = p.user_id
		JOIN authors a ON p.author_id = a.x_author_id
		WHERE qs.qa_id = $1 AND qs.user_id = $2
		ORDER BY p.published_at ASC
	`
	
	type SourceRow struct {
		XPostID     int64   `db:"x_post_id"`
		Handle      string  `db:"handle"`
		DisplayName *string `db:"display_name"`
		PublishedAt string  `db:"published_at"`
		URL         string  `db:"url"`
		Text        string  `db:"text"`
	}
	
	var sourceRows []SourceRow
	err = r.db.SelectContext(ctx, &sourceRows, sourcesQuery, qaID, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch Q&A sources: %w", err)
	}
	
	// Build response DTO
	sources := make([]dto.QASourceDTO, len(sourceRows))
	for i, row := range sourceRows {
		displayName := ""
		if row.DisplayName != nil {
			displayName = *row.DisplayName
		}
		
		sources[i] = dto.QASourceDTO{
			XPostID:           row.XPostID,
			AuthorHandle:      row.Handle,
			AuthorDisplayName: displayName,
			PublishedAt:       mustParseTime(row.PublishedAt),
			URL:               row.URL,
			Text:              row.Text,
		}
	}
	
	return &dto.QADetailDTO{
		ID:        qa.ID,
		Question:  qa.Question,
		Answer:    qa.Answer,
		DateFrom:  qa.DateFrom,
		DateTo:    qa.DateTo,
		CreatedAt: qa.CreatedAt,
		Sources:   sources,
	}, nil
}

// mustParseTime is a helper to parse time strings (should not fail with valid DB data)
func mustParseTime(s string) time.Time {
	// Parse PostgreSQL timestamp format
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try alternative format if RFC3339 fails
		t, err = time.Parse("2006-01-02 15:04:05.999999-07", s)
		if err != nil {
			// If parsing still fails, panic as this indicates invalid DB data
			panic(fmt.Sprintf("failed to parse time %q: %v", s, err))
		}
	}
	return t
}
