package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DatabaseHelper manages test database connections and migrations
type DatabaseHelper struct {
	db *sqlx.DB
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(t *testing.T) *DatabaseHelper {
	t.Helper()

	// Use test database URL from environment or default to test container
	defaultURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		testDBUser, testDBPassword, testDBPort, testDBName)
	dbURL := getEnv("TEST_DATABASE_URL", defaultURL)

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	helper := &DatabaseHelper{db: db}
	helper.applyMigrations(t)

	return helper
}

// GetDB returns the database connection
func (dh *DatabaseHelper) GetDB() *sqlx.DB {
	return dh.db
}

// Close closes the database connection
func (dh *DatabaseHelper) Close() {
	if dh.db != nil {
		// Log error but don't fail - this is cleanup code
		// In tests, we can't use t.Logf here since we don't have access to *testing.T
		_ = dh.db.Close()
	}
}

// applyMigrations applies database schema migrations
func (dh *DatabaseHelper) applyMigrations(t *testing.T) {
	t.Helper()

	migrationSQL := `
-- Create global table: authors (no row level security)
CREATE TABLE IF NOT EXISTS authors (
    x_author_id bigint PRIMARY KEY CHECK (x_author_id > 0),
    handle text NOT NULL,
    display_name text,
    last_seen_at timestamptz
);

-- Create user-scoped table: user_following
CREATE TABLE IF NOT EXISTS user_following (
    user_id uuid NOT NULL,
    x_author_id bigint NOT NULL,
    last_checked_at timestamptz,
    PRIMARY KEY (user_id, x_author_id),
    FOREIGN KEY (x_author_id) REFERENCES authors(x_author_id) ON DELETE CASCADE
);

-- Create user-scoped table: ingest_runs
CREATE TABLE IF NOT EXISTS ingest_runs (
    id char(26) PRIMARY KEY,
    user_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    status text NOT NULL CHECK (status IN ('ok','rate_limited','error')),
    since_id bigint NOT NULL CHECK (since_id > 0),
    fetched_count int NOT NULL,
    retried int NOT NULL,
    rate_limit_hits int NOT NULL,
    err_text text
);

-- Create index for ingest_runs on (user_id, started_at desc)
CREATE INDEX IF NOT EXISTS idx_ingest_runs_user_started ON ingest_runs (user_id, started_at DESC);

-- Create user-scoped table: posts
CREATE TABLE IF NOT EXISTS posts (
    user_id uuid NOT NULL,
    x_post_id bigint NOT NULL CHECK (x_post_id > 0),
    author_id bigint NOT NULL,
    published_at timestamptz NOT NULL,
    url text NOT NULL CHECK (url ~ '^https?://(x|twitter)\\.com/.+/status/\\d+'),
    text text NOT NULL,
    conversation_id bigint,
    ingested_at timestamptz NOT NULL,
    first_visible_at timestamptz NOT NULL,
    edited_seen boolean NOT NULL DEFAULT false,
    ts tsvector GENERATED ALWAYS AS (to_tsvector('english', text)) STORED,
    PRIMARY KEY (user_id, x_post_id),
    FOREIGN KEY (author_id) REFERENCES authors(x_author_id) ON DELETE CASCADE
);

-- Create index for posts on (user_id, published_at desc)
CREATE INDEX IF NOT EXISTS idx_posts_user_published ON posts (user_id, published_at DESC);

-- Create user-scoped table: qa_messages
CREATE TABLE IF NOT EXISTS qa_messages (
    id char(26) PRIMARY KEY,
    user_id uuid NOT NULL,
    question text NOT NULL,
    answer text NOT NULL,
    date_from timestamptz NOT NULL,
    date_to timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

-- Create index for qa_messages on (user_id, created_at desc)
CREATE INDEX IF NOT EXISTS idx_qa_messages_user_created ON qa_messages (user_id, created_at DESC);

-- Create user-scoped junction table: qa_sources
CREATE TABLE IF NOT EXISTS qa_sources (
    qa_id char(26) NOT NULL,
    user_id uuid NOT NULL,
    x_post_id bigint NOT NULL,
    PRIMARY KEY (qa_id, user_id, x_post_id),
    FOREIGN KEY (qa_id) REFERENCES qa_messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id, x_post_id) REFERENCES posts(user_id, x_post_id) ON DELETE CASCADE
);

-- Enable row-level security and create policies for user-scoped tables
ALTER TABLE user_following ENABLE ROW LEVEL SECURITY;
ALTER TABLE ingest_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE qa_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE qa_sources ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS user_isolation_user_following ON user_following;
DROP POLICY IF EXISTS user_isolation_ingest_runs ON ingest_runs;
DROP POLICY IF EXISTS user_isolation_posts ON posts;
DROP POLICY IF EXISTS user_isolation_qa_messages ON qa_messages;
DROP POLICY IF EXISTS user_isolation_qa_sources ON qa_sources;

-- Create policies for user-scoped tables
CREATE POLICY user_isolation_user_following ON user_following
    USING (user_id = current_setting('app.user_id', true)::uuid);

CREATE POLICY user_isolation_ingest_runs ON ingest_runs
    USING (user_id = current_setting('app.user_id', true)::uuid);

CREATE POLICY user_isolation_posts ON posts
    USING (user_id = current_setting('app.user_id', true)::uuid);

CREATE POLICY user_isolation_qa_messages ON qa_messages
    USING (user_id = current_setting('app.user_id', true)::uuid);

CREATE POLICY user_isolation_qa_sources ON qa_sources
    USING (user_id = current_setting('app.user_id', true)::uuid);
`

	_, err := dh.db.Exec(migrationSQL)
	if err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}
}

// CleanupTestData removes all test data from the database
func (dh *DatabaseHelper) CleanupTestData(t *testing.T) {
	t.Helper()

	_, err := dh.db.Exec("TRUNCATE TABLE qa_sources, qa_messages, posts, ingest_runs, authors CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup test data: %v", err)
	}
}

// Helper function
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
