package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/sopeal/AskYourFeed/internal/db"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var qaServiceTracer = otel.Tracer("qa_service")

const noContentMessage = "Brak treści w wybranym zakresie dat. Spróbuj rozszerzyć zakres dat."

// QAService orchestrates Q&A creation workflow
type QAService struct {
	database   *sqlx.DB
	postRepo   *repositories.PostRepository
	qaRepo     *repositories.QARepository
	llmService *LLMService
}

// NewQAService creates a new QAService instance
func NewQAService(
	database *sqlx.DB,
	postRepo *repositories.PostRepository,
	qaRepo *repositories.QARepository,
	llmService *LLMService,
) *QAService {
	return &QAService{
		database:   database,
		postRepo:   postRepo,
		qaRepo:     qaRepo,
		llmService: llmService,
	}
}

// CreateQA creates a new Q&A interaction
// Fetches posts, generates answer via LLM, persists Q&A record, and returns response
func (s *QAService) CreateQA(
	ctx context.Context,
	userID uuid.UUID,
	question string,
	dateFrom time.Time,
	dateTo time.Time,
) (*dto.QADetailDTO, error) {
	ctx, span := qaServiceTracer.Start(ctx, "CreateQA")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("date_from", dateFrom.Format(time.RFC3339)),
		attribute.String("date_to", dateTo.Format(time.RFC3339)),
	)

	// Step 1: Fetch posts from date range
	posts, err := s.postRepo.GetPostsByDateRange(ctx, userID, dateFrom, dateTo)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	span.SetAttributes(attribute.Int("posts_found", len(posts)))

	var answer string
	var sourcePostIDs []int64

	// Step 2: Generate answer or use "no content" message
	if len(posts) == 0 {
		// No posts found - use predefined message
		answer = noContentMessage
		sourcePostIDs = []int64{}
	} else {
		// Posts found - call LLM service
		answer, sourcePostIDs, err = s.llmService.GenerateAnswer(ctx, question, posts)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to generate answer: %w", err)
		}
	}

	// Step 3: Generate ULID for Q&A record
	qaID := ulid.Make().String()
	createdAt := time.Now()

	span.SetAttributes(
		attribute.String("qa_id", qaID),
		attribute.Int("source_count", len(sourcePostIDs)),
	)

	// Step 4: Persist Q&A record in transaction
	tx, err := s.database.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert Q&A message
	qaMessage := db.QAMessage{
		ID:        qaID,
		UserID:    userID,
		Question:  question,
		Answer:    answer,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		CreatedAt: createdAt,
	}

	if err := s.qaRepo.CreateQA(ctx, tx, qaMessage); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create Q&A record: %w", err)
	}

	// Insert Q&A sources (if any)
	if len(sourcePostIDs) > 0 {
		sources := make([]db.QASource, len(sourcePostIDs))
		for i, postID := range sourcePostIDs {
			sources[i] = db.QASource{
				QAID:    qaID,
				UserID:  userID,
				XPostID: postID,
			}
		}

		if err := s.qaRepo.CreateQASources(ctx, tx, sources); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to create Q&A sources: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 5: Build response DTO with source details
	sourceDTOs := s.buildSourceDTOs(posts, sourcePostIDs)

	response := &dto.QADetailDTO{
		ID:        qaID,
		Question:  question,
		Answer:    answer,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		CreatedAt: createdAt,
		Sources:   sourceDTOs,
	}

	return response, nil
}

// buildSourceDTOs creates QASourceDTO objects from posts and selected source IDs
func (s *QAService) buildSourceDTOs(posts []repositories.PostWithAuthor, sourcePostIDs []int64) []dto.QASourceDTO {
	if len(sourcePostIDs) == 0 {
		return []dto.QASourceDTO{}
	}

	// Create a map for quick lookup of source post IDs
	sourceIDMap := make(map[int64]bool)
	for _, id := range sourcePostIDs {
		sourceIDMap[id] = true
	}

	// Build DTOs for selected source posts
	sourceDTOs := make([]dto.QASourceDTO, 0, len(sourcePostIDs))
	for _, post := range posts {
		if sourceIDMap[post.XPostID] {
			displayName := post.Handle
			if post.DisplayName != nil && *post.DisplayName != "" {
				displayName = *post.DisplayName
			}

			// Create text preview (first 200 chars)
			textPreview := post.Text
			if len(textPreview) > 200 {
				textPreview = textPreview[:200] + "..."
			}

			sourceDTO := dto.QASourceDTO{
				XPostID:           post.XPostID,
				AuthorHandle:      post.Handle,
				AuthorDisplayName: displayName,
				PublishedAt:       post.PublishedAt,
				URL:               post.URL,
				TextPreview:       textPreview,
				Text:              post.Text,
			}

			sourceDTOs = append(sourceDTOs, sourceDTO)
		}
	}

	return sourceDTOs
}
