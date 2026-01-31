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

	t.Run("BackfillHours", func(t *testing.T) {
		testTriggerBackfillHours(t, dbHelper, dataHelper, userID)
	})

	t.Run("NoAuthorization", func(t *testing.T) {
		testTriggerNoAuth(t, dbHelper, userID)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		testTriggerInvalidUserID(t, dbHelper, userID)
	})
}

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
