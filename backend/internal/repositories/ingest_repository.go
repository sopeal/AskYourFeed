package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ingestRepoTracer = otel.Tracer("ingest_repository")

// IngestRepository handles ingest_runs data access operations
type IngestRepository struct {
	db *sqlx.DB
}

// NewIngestRepository creates a new IngestRepository instance
func NewIngestRepository(database *sqlx.DB) *IngestRepository {
	return &IngestRepository{
		db: database,
	}
}

// GetLastSyncTime retrieves the most recent completed_at timestamp for a user
// Returns nil if no completed runs exist
func (r *IngestRepository) GetLastSyncTime(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	ctx, span := ingestRepoTracer.Start(ctx, "GetLastSyncTime")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	query := `
		SELECT completed_at
		FROM ingest_runs
		WHERE user_id = $1 AND completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var lastSyncAt *time.Time
	err := r.db.GetContext(ctx, &lastSyncAt, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No completed runs found - this is not an error
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch last sync time: %w", err)
	}

	return lastSyncAt, nil
}

// GetCurrentRun retrieves the currently running ingest for a user (if any)
// Returns nil if no run is currently in progress
func (r *IngestRepository) GetCurrentRun(ctx context.Context, userID uuid.UUID) (*db.IngestRun, error) {
	ctx, span := ingestRepoTracer.Start(ctx, "GetCurrentRun")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	query := `
		SELECT id, user_id, started_at, completed_at, status, cursor,
		       last_cursor, fetched_count, retried, rate_limit_hits, err_text
		FROM ingest_runs
		WHERE user_id = $1 AND completed_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`

	var run db.IngestRun
	err := r.db.GetContext(ctx, &run, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No current run - this is not an error
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch current run: %w", err)
	}

	return &run, nil
}

// GetRecentRuns retrieves recent completed ingest runs for a user
// Returns runs ordered by started_at DESC, limited by the limit parameter
func (r *IngestRepository) GetRecentRuns(ctx context.Context, userID uuid.UUID, limit int) ([]db.IngestRun, error) {
	ctx, span := ingestRepoTracer.Start(ctx, "GetRecentRuns")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int("limit", limit),
	)

	query := `
		SELECT id, user_id, started_at, completed_at, status, cursor,
		       last_cursor, fetched_count, retried, rate_limit_hits, err_text
		FROM ingest_runs
		WHERE user_id = $1 AND completed_at IS NOT NULL
		ORDER BY started_at DESC
		LIMIT $2
	`

	var runs []db.IngestRun
	err := r.db.SelectContext(ctx, &runs, query, userID, limit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch recent runs: %w", err)
	}

	// Return empty slice if no runs found (not an error)
	if runs == nil {
		runs = []db.IngestRun{}
	}

	span.SetAttributes(attribute.Int("runs_found", len(runs)))

	return runs, nil
}

// CreateIngestRun creates a new ingest run with cursor-based pagination
func (r *IngestRepository) CreateIngestRun(ctx context.Context, userID uuid.UUID, runID string, lastCursor string) error {
	ctx, span := ingestRepoTracer.Start(ctx, "CreateIngestRun")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("run_id", runID),
		attribute.String("last_cursor", lastCursor),
	)

	query := `
		INSERT INTO ingest_runs (id, user_id, started_at, status, cursor, last_cursor, fetched_count, retried, rate_limit_hits)
		VALUES ($1, $2, NOW(), 'ok', '', $3, 0, 0, 0)
	`

	_, err := r.db.ExecContext(ctx, query, runID, userID, lastCursor)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create ingest run: %w", err)
	}

	return nil
}

// UpdateIngestRunCursor updates the cursor for an ongoing ingest run
func (r *IngestRepository) UpdateIngestRunCursor(ctx context.Context, runID string, cursor string, fetchedCount int) error {
	ctx, span := ingestRepoTracer.Start(ctx, "UpdateIngestRunCursor")
	defer span.End()

	span.SetAttributes(
		attribute.String("run_id", runID),
		attribute.String("cursor", cursor),
		attribute.Int("fetched_count", fetchedCount),
	)

	query := `
		UPDATE ingest_runs
		SET cursor = $2, fetched_count = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, runID, cursor, fetchedCount)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update ingest run cursor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ingest run not found: %s", runID)
	}

	return nil
}

// CompleteIngestRun marks an ingest run as completed
func (r *IngestRepository) CompleteIngestRun(ctx context.Context, runID string, status string, finalFetchedCount int, errText *string) error {
	ctx, span := ingestRepoTracer.Start(ctx, "CompleteIngestRun")
	defer span.End()

	span.SetAttributes(
		attribute.String("run_id", runID),
		attribute.String("status", status),
		attribute.Int("final_fetched_count", finalFetchedCount),
	)

	query := `
		UPDATE ingest_runs
		SET completed_at = NOW(), status = $2, fetched_count = $3, err_text = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, runID, status, finalFetchedCount, errText)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to complete ingest run: %w", err)
	}

	return nil
}
