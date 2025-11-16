package dto

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Auth DTOs
// =============================================================================

// AuthInitiateResponseDTO represents response from login initiation
// Contains OAuth authorization URL and state token
type AuthInitiateResponseDTO struct {
	AuthURL string `json:"auth_url"` // X OAuth authorization URL
	State   string `json:"state"`    // CSRF protection token
}

// AuthCallbackResponseDTO represents response from OAuth callback
// Contains session token and user information
type AuthCallbackResponseDTO struct {
	SessionToken string    `json:"session_token"` // Session token for API authentication
	UserID       uuid.UUID `json:"user_id"`       // Internal user ID
	XHandle      string    `json:"x_handle"`      // X (Twitter) username
	RedirectURL  string    `json:"redirect_url"`  // Frontend redirect URL
}

// =============================================================================
// Session DTOs
// =============================================================================

// SessionDTO represents current user session information
// Maps to: OAuth tokens (external storage) + user metadata
type SessionDTO struct {
	UserID           uuid.UUID `json:"user_id"`
	XHandle          string    `json:"x_handle"`
	XDisplayName     string    `json:"x_display_name"`
	AuthenticatedAt  time.Time `json:"authenticated_at"`
	SessionExpiresAt time.Time `json:"session_expires_at"`
}

// =============================================================================
// Ingest DTOs and Commands
// =============================================================================

// TriggerIngestCommand represents request to manually trigger feed ingestion
// Command model for POST /api/v1/ingest/trigger
type TriggerIngestCommand struct {
	BackfillHours int `json:"backfill_hours" validate:"min=1,max=720"` // Max 30 days
}

// TriggerIngestResponseDTO represents response after triggering ingestion
// Maps to: ingest_runs table
type TriggerIngestResponseDTO struct {
	IngestRunID string    `json:"ingest_run_id"` // ULID from ingest_runs.id
	Status      string    `json:"status"`        // From ingest_runs.status
	StartedAt   time.Time `json:"started_at"`    // From ingest_runs.started_at
}

// IngestRunDTO represents a single ingestion run
// Maps to: ingest_runs table
type IngestRunDTO struct {
	ID            string     `json:"id"`                       // From ingest_runs.id (ULID)
	Status        string     `json:"status"`                   // From ingest_runs.status
	StartedAt     time.Time  `json:"started_at"`               // From ingest_runs.started_at
	CompletedAt   *time.Time `json:"completed_at,omitempty"`   // From ingest_runs.completed_at (nullable)
	FetchedCount  int        `json:"fetched_count"`            // From ingest_runs.fetched_count
	Retried       int        `json:"retried,omitempty"`        // From ingest_runs.retried
	RateLimitHits int        `json:"rate_limit_hits,omitempty"` // From ingest_runs.rate_limit_hits
	Error         string     `json:"error,omitempty"`          // From ingest_runs.err_text (nullable)
}

// IngestStatusDTO represents current and recent ingestion status
// Combines data from multiple ingest_runs table rows
type IngestStatusDTO struct {
	LastSyncAt  *time.Time      `json:"last_sync_at,omitempty"` // Most recent completed_at from ingest_runs
	CurrentRun  *IngestRunDTO   `json:"current_run,omitempty"`  // Currently running ingest (status != 'ok'|'rate_limited'|'error' or completed_at IS NULL)
	RecentRuns  []IngestRunDTO  `json:"recent_runs"`            // Recent completed runs from ingest_runs
}

// =============================================================================
// Q&A DTOs and Commands
// =============================================================================

// CreateQACommand represents request to create a new Q&A interaction
// Command model for POST /api/v1/qa
type CreateQACommand struct {
	Question string     `json:"question" validate:"required,min=1,max=2000"`
	DateFrom *time.Time `json:"date_from,omitempty"` // Optional, defaults to 24 hours ago
	DateTo   *time.Time `json:"date_to,omitempty"`   // Optional, defaults to now
}

// QASourceDTO represents a source post for Q&A answer
// Maps to: posts table joined with authors table via qa_sources junction table
type QASourceDTO struct {
	XPostID           int64     `json:"x_post_id"`            // From posts.x_post_id
	AuthorHandle      string    `json:"author_handle"`        // From authors.handle
	AuthorDisplayName string    `json:"author_display_name"`  // From authors.display_name
	PublishedAt       time.Time `json:"published_at"`         // From posts.published_at
	URL               string    `json:"url"`                  // From posts.url
	TextPreview       string    `json:"text_preview,omitempty"` // Truncated posts.text (for list view)
	Text              string    `json:"text,omitempty"`       // Full posts.text (for detail view)
}

// QADetailDTO represents full Q&A interaction details
// Maps to: qa_messages table with sources from qa_sources -> posts -> authors
type QADetailDTO struct {
	ID        string        `json:"id"`        // From qa_messages.id (ULID)
	Question  string        `json:"question"`  // From qa_messages.question
	Answer    string        `json:"answer"`    // From qa_messages.answer
	DateFrom  time.Time     `json:"date_from"` // From qa_messages.date_from
	DateTo    time.Time     `json:"date_to"`   // From qa_messages.date_to
	CreatedAt time.Time     `json:"created_at"` // From qa_messages.created_at
	Sources   []QASourceDTO `json:"sources"`   // From qa_sources joined with posts and authors
}

// QAListItemDTO represents Q&A item in paginated list
// Maps to: qa_messages table with source count from qa_sources
type QAListItemDTO struct {
	ID            string    `json:"id"`             // From qa_messages.id (ULID)
	Question      string    `json:"question"`       // From qa_messages.question
	AnswerPreview string    `json:"answer_preview"` // Truncated qa_messages.answer
	DateFrom      time.Time `json:"date_from"`      // From qa_messages.date_from
	DateTo        time.Time `json:"date_to"`        // From qa_messages.date_to
	CreatedAt     time.Time `json:"created_at"`     // From qa_messages.created_at
	SourcesCount  int       `json:"sources_count"`  // COUNT from qa_sources
}

// QAListResponseDTO represents paginated Q&A list response
type QAListResponseDTO struct {
	Items      []QAListItemDTO `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"` // ULID of last item for pagination
	HasMore    bool            `json:"has_more"`
}

// =============================================================================
// Following DTOs
// =============================================================================

// FollowingItemDTO represents an author the user follows
// Maps to: user_following table joined with authors table
type FollowingItemDTO struct {
	XAuthorID      int64      `json:"x_author_id"`      // From authors.x_author_id
	Handle         string     `json:"handle"`           // From authors.handle
	DisplayName    string     `json:"display_name"`     // From authors.display_name
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"` // From authors.last_seen_at (nullable)
	LastCheckedAt  *time.Time `json:"last_checked_at,omitempty"` // From user_following.last_checked_at (nullable)
}

// FollowingListResponseDTO represents paginated following list response
type FollowingListResponseDTO struct {
	Items      []FollowingItemDTO `json:"items"`
	NextCursor int64              `json:"next_cursor,omitempty"` // x_author_id of last item for pagination
	HasMore    bool               `json:"has_more"`
	TotalCount int                `json:"total_count"` // Total count of followed authors
}

// =============================================================================
// System Health DTOs
// =============================================================================

// ComponentHealthDTO represents health status of a system component
type ComponentHealthDTO struct {
	Status            string     `json:"status"`                        // "healthy", "unhealthy", "rate_limited"
	ResponseTimeMs    int        `json:"response_time_ms,omitempty"`    // Response time in milliseconds
	Error             string     `json:"error,omitempty"`               // Error message if unhealthy
	RateLimitRemaining int       `json:"rate_limit_remaining,omitempty"` // For X API component
	RateLimitResetAt  *time.Time `json:"rate_limit_reset_at,omitempty"` // For X API component when rate limited
}

// HealthResponseDTO represents system health check response
type HealthResponseDTO struct {
	Status     string                        `json:"status"`    // "healthy", "degraded", "unhealthy"
	Timestamp  time.Time                     `json:"timestamp"` // Current server time
	Version    string                        `json:"version"`   // Application version
	Components map[string]ComponentHealthDTO `json:"components"` // Component-specific health
}

// =============================================================================
// Common Response DTOs
// =============================================================================

// MessageResponseDTO represents simple message responses
// Used for: logout, delete operations
type MessageResponseDTO struct {
	Message string `json:"message"`
}

// DeleteAllQAResponseDTO represents response after deleting all Q&A history
type DeleteAllQAResponseDTO struct {
	Message      string `json:"message"`
	DeletedCount int    `json:"deleted_count"`
}

// ErrorResponseDTO represents standard error response format
type ErrorResponseDTO struct {
	Error ErrorDetailDTO `json:"error"`
}

// ErrorDetailDTO contains error details
type ErrorDetailDTO struct {
	Code    string                 `json:"code"`              // Error code (e.g., "INVALID_DATE_RANGE")
	Message string                 `json:"message"`           // Human-readable error message in Polish
	Details map[string]interface{} `json:"details,omitempty"` // Additional error context
}
