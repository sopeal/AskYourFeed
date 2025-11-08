package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/db"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ingestServiceTracer = otel.Tracer("ingest_service")

// IngestService handles business logic for ingestion operations
type IngestService struct {
	ingestRepo *repositories.IngestRepository
}

// NewIngestService creates a new IngestService instance
func NewIngestService(ingestRepo *repositories.IngestRepository) *IngestService {
	return &IngestService{
		ingestRepo: ingestRepo,
	}
}

// GetIngestStatus retrieves current and recent ingestion status for a user
func (s *IngestService) GetIngestStatus(ctx context.Context, userID uuid.UUID, limit int) (*dto.IngestStatusDTO, error) {
	ctx, span := ingestServiceTracer.Start(ctx, "GetIngestStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int("limit", limit),
	)

	// Fetch last sync time
	lastSyncAt, err := s.ingestRepo.GetLastSyncTime(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get last sync time: %w", err)
	}

	// Fetch current running ingest (if any)
	currentRunDB, err := s.ingestRepo.GetCurrentRun(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get current run: %w", err)
	}

	// Fetch recent completed runs
	recentRunsDB, err := s.ingestRepo.GetRecentRuns(ctx, userID, limit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get recent runs: %w", err)
	}

	// Build response DTO
	response := &dto.IngestStatusDTO{
		LastSyncAt: lastSyncAt,
		RecentRuns: make([]dto.IngestRunDTO, 0, len(recentRunsDB)),
	}

	// Map current run to DTO if exists
	if currentRunDB != nil {
		response.CurrentRun = mapIngestRunToDTO(currentRunDB)
	}

	// Map recent runs to DTOs
	for _, run := range recentRunsDB {
		response.RecentRuns = append(response.RecentRuns, *mapIngestRunToDTO(&run))
	}

	span.SetAttributes(
		attribute.Bool("has_current_run", currentRunDB != nil),
		attribute.Int("recent_runs_count", len(response.RecentRuns)),
	)

	return response, nil
}

// mapIngestRunToDTO converts a database IngestRun entity to IngestRunDTO
func mapIngestRunToDTO(run *db.IngestRun) *dto.IngestRunDTO {
	runDTO := &dto.IngestRunDTO{
		ID:            run.ID,
		Status:        run.Status,
		StartedAt:     run.StartedAt,
		CompletedAt:   run.CompletedAt,
		FetchedCount:  run.FetchedCount,
		Retried:       run.Retried,
		RateLimitHits: run.RateLimitHits,
	}

	// Include error text if present
	if run.ErrText != nil && *run.ErrText != "" {
		runDTO.Error = *run.ErrText
	}

	return runDTO
}
