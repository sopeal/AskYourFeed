package db

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Database Entity Models (for reference and mapping)
// =============================================================================

// Author represents the authors table (global, no RLS)
type Author struct {
	XAuthorID   int64      `db:"x_author_id"`
	Handle      string     `db:"handle"`
	DisplayName *string    `db:"display_name"` // Nullable in DB
	LastSeenAt  *time.Time `db:"last_seen_at"` // Nullable in DB
}

// UserFollowing represents the user_following table (user-scoped, RLS enabled)
type UserFollowing struct {
	UserID        uuid.UUID  `db:"user_id"`
	XAuthorID     int64      `db:"x_author_id"`
	LastCheckedAt *time.Time `db:"last_checked_at"` // Nullable in DB
}

// Post represents the posts table (user-scoped, RLS enabled)
type Post struct {
	UserID         uuid.UUID `db:"user_id"`
	XPostID        int64     `db:"x_post_id"`
	AuthorID       int64     `db:"author_id"`
	PublishedAt    time.Time `db:"published_at"`
	URL            string    `db:"url"`
	Text           string    `db:"text"`
	ConversationID *int64    `db:"conversation_id"` // Nullable in DB
	IngestedAt     time.Time `db:"ingested_at"`
	FirstVisibleAt time.Time `db:"first_visible_at"`
	EditedSeen     bool      `db:"edited_seen"`
	// ts field (tsvector) not included as it's internal to PostgreSQL
}

// QAMessage represents the qa_messages table (user-scoped, RLS enabled)
type QAMessage struct {
	ID        string    `db:"id"` // ULID as string
	UserID    uuid.UUID `db:"user_id"`
	Question  string    `db:"question"`
	Answer    string    `db:"answer"`
	DateFrom  time.Time `db:"date_from"`
	DateTo    time.Time `db:"date_to"`
	CreatedAt time.Time `db:"created_at"`
}

// QASource represents the qa_sources junction table (user-scoped, RLS enabled)
type QASource struct {
	QAID    string    `db:"qa_id"`
	UserID  uuid.UUID `db:"user_id"`
	XPostID int64     `db:"x_post_id"`
}

// IngestRun represents the ingest_runs table (user-scoped, RLS enabled)
type IngestRun struct {
	ID            string     `db:"id"` // ULID as string
	UserID        uuid.UUID  `db:"user_id"`
	StartedAt     time.Time  `db:"started_at"`
	CompletedAt   *time.Time `db:"completed_at"` // Nullable in DB
	Status        string     `db:"status"`       // CHECK: 'ok', 'rate_limited', 'error'
	SinceID       int64      `db:"since_id"`
	FetchedCount  int        `db:"fetched_count"`
	Retried       int        `db:"retried"`
	RateLimitHits int        `db:"rate_limit_hits"`
	ErrText       *string    `db:"err_text"` // Nullable in DB
}
