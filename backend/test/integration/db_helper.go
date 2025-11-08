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
		dh.db.Close()
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

-- Enable row-level security and create policies for user-scoped tables
ALTER TABLE ingest_runs ENABLE ROW LEVEL SECURITY;

-- Drop existing policy if it exists
DROP POLICY IF EXISTS user_isolation_ingest_runs ON ingest_runs;

CREATE POLICY user_isolation_ingest_runs ON ingest_runs
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

	_, err := dh.db.Exec("TRUNCATE TABLE ingest_runs CASCADE")
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
