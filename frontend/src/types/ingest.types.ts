/**
 * Ingest Types
 * Type definitions for feed ingestion and sync status
 */

/**
 * Request to manually trigger feed ingestion
 * POST /api/v1/ingest/trigger
 */
export interface TriggerIngestCommand {
  backfill_hours: number; // Min 1, Max 720 (30 days)
}

/**
 * Response after triggering ingestion
 * Maps to backend TriggerIngestResponseDTO
 */
export interface TriggerIngestResponseDTO {
  ingest_run_id: string; // ULID
  status: string;
  started_at: string; // ISO 8601 format
}

/**
 * Single ingestion run details
 * Maps to backend IngestRunDTO
 */
export interface IngestRunDTO {
  id: string; // ULID
  status: string;
  started_at: string; // ISO 8601 format
  completed_at?: string; // ISO 8601 format, nullable
  fetched_count: number;
  retried?: number;
  rate_limit_hits?: number;
  error?: string;
}

/**
 * Current and recent ingestion status
 * Maps to backend IngestStatusDTO
 * GET /api/v1/ingest/status
 */
export interface IngestStatusDTO {
  last_sync_at?: string; // ISO 8601 format, nullable
  current_run?: IngestRunDTO;
  recent_runs: IngestRunDTO[];
}
