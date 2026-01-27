package integration_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/pkg/logger"
	"github.com/sopeal/AskYourFeed/internal/testutil"
)

// TestIngestTriggerIntegration contains all integration tests for the ingest trigger endpoint
func TestIngestTriggerIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger for tests
	logger.Init(slog.LevelInfo)

	// Initialize test database
	dbHelper := testutil.NewDatabaseHelper(t)
	defer dbHelper.Close()

	conn := dbHelper.GetDB()
	dataHelper := testutil.NewTestDataHelper(conn)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Run test suite
	//t.Run("HappyPath", func(t *testing.T) {
	//	testTriggerHappyPath(t, dbHelper, dataHelper, userID)
	//})

	t.Run("BackfillHours", func(t *testing.T) {
		testTriggerBackfillHours(t, dbHelper, dataHelper, userID)
	})

	t.Run("NoAuthorization", func(t *testing.T) {
		testTriggerNoAuth(t, dbHelper, userID)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		testTriggerInvalidUserID(t, dbHelper, userID)
	})

	//t.Run("InvalidJSON", func(t *testing.T) {
	//	testTriggerInvalidJSON(t, dbHelper, userID)
	//})

	//t.Run("ConflictRunningIngest", func(t *testing.T) {
	//	testTriggerConflict(t, dbHelper, dataHelper, userID)
	//})
}

// testTriggerHappyPath tests successful trigger - creates ingest run, background fails due to API key but completes 'error'
//func testTriggerHappyPath(t *testing.T, dbHelper *testutil.DatabaseHelper, dataHelper *testutil.TestDataHelper, userID uuid.UUID) {
//	dbHelper.CleanupTestData(t)
//
//	conn := dbHelper.GetDB()
//	router := testutil.NewTestRouter(conn).GetEngine()
//
//	// Count before
//	var countBefore int
//	if err := conn.QueryRow("SELECT COUNT(*) FROM ingest_runs WHERE user_id = $1", userID).Scan(&countBefore); err != nil {
//		t.Fatalf("Failed to count runs before: %v", err)
//	}
//
//	// Trigger
//	w := httptest.NewRecorder()
//	body := `{"backfill_hours": 24}`
//	req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", bytes.NewBufferString(body))
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer test-token")
//	req.Header.Set("X-Test-User-ID", userID.String())
//	router.ServeHTTP(w, req)
//
//	// Assert immediate response
//	if w.Code != http.StatusAccepted {
//		t.Errorf("Expected 202 Accepted, got %d", w.Code)
//	}
//
//	var resp map[string]interface{}
//	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
//		t.Fatalf("Failed to unmarshal response: %v", err)
//	}
//	if resp["status"] != "triggered" {
//		t.Errorf("Expected status 'triggered', got %v", resp["status"])
//	}
//
//	// Wait for background goroutine
//	time.Sleep(3 * time.Second)
//
//	// Check new run created and completed with error (API fail)
//	var countAfter int
//	if err := conn.QueryRow("SELECT COUNT(*) FROM ingest_runs WHERE user_id = $1", userID).Scan(&countAfter); err != nil {
//		t.Fatalf("Failed to count runs after: %v", err)
//	}
//	if countAfter != countBefore+1 {
//		t.Errorf("Expected 1 new run, got %d (before %d)", countAfter-countBefore, countBefore)
//	}
//
//	var run db.IngestRun
//	err := conn.Get(&run, `
//		SELECT * FROM ingest_runs
//		WHERE user_id = $1 ORDER BY started_at DESC LIMIT 1`, userID)
//	if err != nil {
//		t.Fatalf("Failed to get run: %v", err)
//	}
//	if run.Status != "error" {
//		t.Errorf("Expected status 'error' due to API fail, got '%s'", run.Status)
//	}
//	if run.CompletedAt == nil {
//		t.Error("Expected run completed")
//	}
//	if run.FetchedCount != 0 {
//		t.Errorf("Expected 0 fetched, got %d", run.FetchedCount)
//	}
//}

// testTriggerBackfillHours tests parsing backfill_hours (not used but binds)
func testTriggerBackfillHours(t *testing.T, dbHelper *testutil.DatabaseHelper, dataHelper *testutil.TestDataHelper, userID uuid.UUID) {
	dbHelper.CleanupTestData(t)

	conn := dbHelper.GetDB()
	router := testutil.NewTestRouter(conn).GetEngine()

	tests := []struct {
		name string
		body string
	}{
		{"default", `{}`},
		{"zero", `{"backfill_hours": 0}`},
		{"positive", `{"backfill_hours": 48}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")
			req.Header.Set("X-Test-User-ID", userID.String())
			router.ServeHTTP(w, req)

			if w.Code != http.StatusAccepted {
				t.Errorf("%s: Expected 202, got %d", tt.name, w.Code)
			}
		})
	}
}

// testTriggerNoAuth tests missing auth header
func testTriggerNoAuth(t *testing.T, dbHelper *testutil.DatabaseHelper, userID uuid.UUID) {
	dbHelper.CleanupTestData(t)

	conn := dbHelper.GetDB()
	router := testutil.NewTestRouter(conn).GetEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

// testTriggerInvalidUserID tests invalid user ID
func testTriggerInvalidUserID(t *testing.T, dbHelper *testutil.DatabaseHelper, userID uuid.UUID) {
	dbHelper.CleanupTestData(t)

	conn := dbHelper.GetDB()
	router := testutil.NewTestRouter(conn).GetEngine()

	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", "invalid")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

// testTriggerInvalidJSON tests invalid JSON
//func testTriggerInvalidJSON(t *testing.T, dbHelper *testutil.DatabaseHelper, userID uuid.UUID) {
//	dbHelper.CleanupTestData(t)
//
//	conn := dbHelper.GetDB()
//	router := testutil.NewTestRouter(conn).GetEngine()
//
//	w := httptest.NewRecorder()
//	body := `{invalid json}`
//	req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", bytes.NewBufferString(body))
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer test-token")
//	req.Header.Set("X-Test-User-ID", userID.String())
//	router.ServeHTTP(w, req)
//
//	if w.Code != http.StatusBadRequest {
//		t.Errorf("Expected 400, got %d", w.Code)
//	}
//}

// testTriggerConflict tests trigger when running ingest exists (service skips)
//func testTriggerConflict(t *testing.T, dbHelper *testutil.DatabaseHelper, dataHelper *testutil.TestDataHelper, userID uuid.UUID) {
//	dbHelper.CleanupTestData(t)
//
//	conn := dbHelper.GetDB()
//
//	// Insert running ingest
//	now := time.Now().UTC()
//	_ = dataHelper.InsertIngestRun(t, userID, now.Add(-30*time.Second), nil, "ok", 5, 0, 0, nil)
//
//	var countBefore int
//	if err := conn.QueryRow("SELECT COUNT(*) FROM ingest_runs WHERE user_id = $1", userID).Scan(&countBefore); err != nil {
//		t.Fatalf("Failed to count runs before: %v", err)
//	}
//
//	router := testutil.NewTestRouter(conn).GetEngine()
//
//	w := httptest.NewRecorder()
//	body := `{}`
//	req, _ := http.NewRequest("POST", "/api/v1/ingest/trigger", bytes.NewBufferString(body))
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer test-token")
//	req.Header.Set("X-Test-User-ID", userID.String())
//	router.ServeHTTP(w, req)
//
//	if w.Code != http.StatusAccepted {
//		t.Errorf("Expected 202 even with running (no sync check in handler), got %d", w.Code)
//	}
//
//	// Wait
//	time.Sleep(3 * time.Second)
//
//	// Check no new run created (service skipped)
//	var countAfter int
//	if err := conn.QueryRow("SELECT COUNT(*) FROM ingest_runs WHERE user_id = $1", userID).Scan(&countAfter); err != nil {
//		t.Fatalf("Failed to count runs after: %v", err)
//	}
//	if countAfter != countBefore {
//		t.Errorf("Expected no new run (service skipped), got %d (before %d)", countAfter, countBefore)
//	}
//
//	// Running still running or errored, but no new
//}
