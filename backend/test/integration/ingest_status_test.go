package integration

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/dto"
)

// TestMain handles setup and teardown for all tests
func TestMain(m *testing.M) {
	dockerMgr := NewDockerManager()

	if err := dockerMgr.SetupDatabase(); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	dockerMgr.Cleanup()

	os.Exit(code)
}

// TestIngestStatusIntegration contains all integration tests for the ingest status endpoint
func TestIngestStatusIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	dbHelper := NewDatabaseHelper(t)
	defer dbHelper.Close()

	db := dbHelper.GetDB()

	// Run test suite
	t.Run("HappyPath", func(t *testing.T) {
		testHappyPath(t, dbHelper)
	})

	t.Run("NoData", func(t *testing.T) {
		testNoData(t, dbHelper)
	})

	t.Run("CurrentRunning", func(t *testing.T) {
		testCurrentRunning(t, dbHelper)
	})

	t.Run("LimitParameter", func(t *testing.T) {
		testLimitParameter(t, dbHelper)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		testEdgeCases(t, dbHelper)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		testErrorCases(t, db)
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		testMultipleUsers(t, dbHelper)
	})
}

// testHappyPath tests the happy path scenario with completed runs
func testHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now().UTC()

	// Insert test data: 3 completed runs
	completedAt1 := now.Add(-1 * time.Hour)
	completedAt2 := now.Add(-2 * time.Hour)
	completedAt3 := now.Add(-3 * time.Hour)

	dataHelper.InsertIngestRun(t, userID, now.Add(-1*time.Hour-5*time.Minute), &completedAt1, "ok", 15, 0, 0, nil)
	dataHelper.InsertIngestRun(t, userID, now.Add(-2*time.Hour-5*time.Minute), &completedAt2, "ok", 20, 0, 0, nil)
	dataHelper.InsertIngestRun(t, userID, now.Add(-3*time.Hour-5*time.Minute), &completedAt3, "rate_limited", 8, 2, 3, StringPtr("Przekroczono limit żądań API X"))

	// Setup router and make request
	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.IngestStatusDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify last_sync_at
	if response.LastSyncAt == nil {
		t.Error("Expected last_sync_at to be set")
	} else {
		// Compare with millisecond precision to account for database timestamp precision
		expected := completedAt1.Truncate(time.Millisecond)
		actual := response.LastSyncAt.Truncate(time.Millisecond)
		if expected != actual {
			t.Errorf("Expected last_sync_at to be %v, got %v", expected, actual)
		}
	}

	// Verify no current run
	if response.CurrentRun != nil {
		t.Error("Expected no current run")
	}

	// Verify recent runs
	if len(response.RecentRuns) != 3 {
		t.Errorf("Expected 3 recent runs, got %d", len(response.RecentRuns))
	}

	// Verify first run (most recent)
	if response.RecentRuns[0].Status != "ok" {
		t.Errorf("Expected first run status 'ok', got '%s'", response.RecentRuns[0].Status)
	}
	if response.RecentRuns[0].FetchedCount != 15 {
		t.Errorf("Expected first run fetched_count 15, got %d", response.RecentRuns[0].FetchedCount)
	}

	// Verify third run (rate limited)
	if response.RecentRuns[2].Status != "rate_limited" {
		t.Errorf("Expected third run status 'rate_limited', got '%s'", response.RecentRuns[2].Status)
	}
	if response.RecentRuns[2].Error != "Przekroczono limit żądań API X" {
		t.Errorf("Expected error message, got '%s'", response.RecentRuns[2].Error)
	}
}

// testNoData tests the scenario when user has no ingest runs
func testNoData(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.IngestStatusDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify empty response
	if response.LastSyncAt != nil {
		t.Error("Expected last_sync_at to be nil")
	}
	if response.CurrentRun != nil {
		t.Error("Expected no current run")
	}
	if len(response.RecentRuns) != 0 {
		t.Errorf("Expected 0 recent runs, got %d", len(response.RecentRuns))
	}
}

// testCurrentRunning tests the scenario with a currently running ingest
func testCurrentRunning(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	now := time.Now().UTC()

	// Insert current running ingest (no completed_at)
	currentID := dataHelper.InsertIngestRun(t, userID, now.Add(-5*time.Minute), nil, "ok", 42, 0, 0, nil)

	// Insert completed runs
	completedAt1 := now.Add(-1 * time.Hour)
	dataHelper.InsertIngestRun(t, userID, now.Add(-1*time.Hour-5*time.Minute), &completedAt1, "ok", 15, 0, 0, nil)

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.IngestStatusDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify current run
	if response.CurrentRun == nil {
		t.Fatal("Expected current run to be set")
	}
	if response.CurrentRun.ID != currentID {
		t.Errorf("Expected current run ID %s, got %s", currentID, response.CurrentRun.ID)
	}
	if response.CurrentRun.Status != "ok" {
		t.Errorf("Expected current run status 'ok', got '%s'", response.CurrentRun.Status)
	}
	if response.CurrentRun.FetchedCount != 42 {
		t.Errorf("Expected current run fetched_count 42, got %d", response.CurrentRun.FetchedCount)
	}
	if response.CurrentRun.CompletedAt != nil {
		t.Error("Expected current run completed_at to be nil")
	}

	// Verify recent runs (should not include current run)
	if len(response.RecentRuns) != 1 {
		t.Errorf("Expected 1 recent run, got %d", len(response.RecentRuns))
	}
}

// testLimitParameter tests the limit query parameter
func testLimitParameter(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Now().UTC()

	// Insert 15 completed runs
	for i := 0; i < 15; i++ {
		completedAt := now.Add(-time.Duration(i+1) * time.Hour)
		dataHelper.InsertIngestRun(t, userID, now.Add(-time.Duration(i+1)*time.Hour-5*time.Minute), &completedAt, "ok", 10+i, 0, 0, nil)
	}

	router := NewTestRouter(db).GetEngine()

	// Test default limit (10)
	t.Run("DefaultLimit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.IngestStatusDTO
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response.RecentRuns) != 10 {
			t.Errorf("Expected 10 recent runs (default), got %d", len(response.RecentRuns))
		}
	})

	// Test custom limit (5)
	t.Run("CustomLimit5", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=5", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.IngestStatusDTO
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response.RecentRuns) != 5 {
			t.Errorf("Expected 5 recent runs, got %d", len(response.RecentRuns))
		}
	})

	// Test max limit (50)
	t.Run("MaxLimit50", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=50", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.IngestStatusDTO
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Should return all 15 runs (less than limit)
		if len(response.RecentRuns) != 15 {
			t.Errorf("Expected 15 recent runs, got %d", len(response.RecentRuns))
		}
	})
}

// testEdgeCases tests edge cases and boundary conditions
func testEdgeCases(t *testing.T, dbHelper *DatabaseHelper) {
	db := dbHelper.GetDB()
	router := NewTestRouter(db).GetEngine()

	// Test limit exceeds maximum (51)
	t.Run("LimitExceedsMax", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=51", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "INVALID_LIMIT" {
			t.Errorf("Expected error code 'INVALID_LIMIT', got '%s'", response.Error.Code)
		}
	})

	// Test limit is zero
	t.Run("LimitZero", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000006")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=0", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "INVALID_LIMIT" {
			t.Errorf("Expected error code 'INVALID_LIMIT', got '%s'", response.Error.Code)
		}
	})

	// Test limit is negative
	t.Run("LimitNegative", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000007")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=-5", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "INVALID_LIMIT" {
			t.Errorf("Expected error code 'INVALID_LIMIT', got '%s'", response.Error.Code)
		}
	})

	// Test limit is not a number
	t.Run("LimitNotNumber", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000008")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status?limit=abc", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "INVALID_LIMIT" {
			t.Errorf("Expected error code 'INVALID_LIMIT', got '%s'", response.Error.Code)
		}
	})
}

// testErrorCases tests error scenarios
func testErrorCases(t *testing.T, db *sqlx.DB) {
	router := NewTestRouter(db).GetEngine()

	// Test missing authorization header
	t.Run("MissingAuthHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "UNAUTHORIZED" {
			t.Errorf("Expected error code 'UNAUTHORIZED', got '%s'", response.Error.Code)
		}
	})

	// Test invalid user ID format
	t.Run("InvalidUserIDFormat", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", "invalid-uuid")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "INVALID_USER_ID" {
			t.Errorf("Expected error code 'INVALID_USER_ID', got '%s'", response.Error.Code)
		}
	})
}

// testMultipleUsers tests data isolation between users
func testMultipleUsers(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	user1ID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	user2ID := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	now := time.Now().UTC()

	// Insert data for user 1
	completedAt1 := now.Add(-1 * time.Hour)
	dataHelper.InsertIngestRun(t, user1ID, now.Add(-1*time.Hour-5*time.Minute), &completedAt1, "ok", 10, 0, 0, nil)
	dataHelper.InsertIngestRun(t, user1ID, now.Add(-2*time.Hour-5*time.Minute), &completedAt1, "ok", 20, 0, 0, nil)

	// Insert data for user 2
	completedAt2 := now.Add(-30 * time.Minute)
	dataHelper.InsertIngestRun(t, user2ID, now.Add(-30*time.Minute-5*time.Minute), &completedAt2, "ok", 30, 0, 0, nil)

	router := NewTestRouter(db).GetEngine()

	// Request for user 1
	t.Run("User1Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user1ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.IngestStatusDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		// User 1 should see 2 runs
		if len(response.RecentRuns) != 2 {
			t.Errorf("Expected 2 recent runs for user 1, got %d", len(response.RecentRuns))
		}
	})

	// Request for user 2
	t.Run("User2Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ingest/status", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user2ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.IngestStatusDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		// User 2 should see 1 run
		if len(response.RecentRuns) != 1 {
			t.Errorf("Expected 1 recent run for user 2, got %d", len(response.RecentRuns))
		}

		// Verify it's the correct run
		if response.RecentRuns[0].FetchedCount != 30 {
			t.Errorf("Expected fetched_count 30 for user 2, got %d", response.RecentRuns[0].FetchedCount)
		}
	})
}
