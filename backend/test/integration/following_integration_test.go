package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/dto"
)

// TestFollowingIntegration contains all integration tests for the following endpoint
func TestFollowingIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	dbHelper := NewDatabaseHelper(t)
	defer dbHelper.Close()

	db := dbHelper.GetDB()

	// Run test suite
	t.Run("HappyPath", func(t *testing.T) {
		testFollowingHappyPath(t, dbHelper)
	})

	t.Run("NoData", func(t *testing.T) {
		testFollowingNoData(t, dbHelper)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		testFollowingErrorCases(t, db)
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		testFollowingMultipleUsers(t, dbHelper)
	})
}

// testFollowingHappyPath tests the happy path scenario with following data
func testFollowingHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now().UTC()

	// Insert test authors
	author1ID := int64(123456789)
	author2ID := int64(987654321)
	author3ID := int64(555666777)

	lastSeen1 := now.Add(-1 * time.Hour)
	lastSeen2 := now.Add(-2 * time.Hour)
	lastSeen3 := now.Add(-3 * time.Hour)

	dataHelper.InsertAuthor(t, author1ID, "@author1", StringPtr("Author One"), &lastSeen1)
	dataHelper.InsertAuthor(t, author2ID, "@author2", StringPtr("Author Two"), &lastSeen2)
	dataHelper.InsertAuthor(t, author3ID, "@author3", StringPtr("Author Three"), &lastSeen3)

	// Insert user following relationships
	lastChecked1 := now.Add(-30 * time.Minute)
	lastChecked2 := now.Add(-45 * time.Minute)
	lastChecked3 := now.Add(-60 * time.Minute)

	dataHelper.InsertUserFollowing(t, userID, author1ID, &lastChecked1)
	dataHelper.InsertUserFollowing(t, userID, author2ID, &lastChecked2)
	dataHelper.InsertUserFollowing(t, userID, author3ID, &lastChecked3)

	// Setup router and make request
	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/following", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.FollowingListResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if len(response.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(response.Items))
	}

	// Verify items are ordered by x_author_id DESC (highest first)
	if response.Items[0].XAuthorID != author2ID {
		t.Errorf("Expected first item x_author_id %d, got %d", author2ID, response.Items[0].XAuthorID)
	}
	if response.Items[1].XAuthorID != author3ID {
		t.Errorf("Expected second item x_author_id %d, got %d", author3ID, response.Items[1].XAuthorID)
	}
	if response.Items[2].XAuthorID != author1ID {
		t.Errorf("Expected third item x_author_id %d, got %d", author1ID, response.Items[2].XAuthorID)
	}

	// Verify first item details (should be author2 with highest ID)
	item := response.Items[0]
	if item.Handle != "@author2" {
		t.Errorf("Expected handle '@author2', got '%s'", item.Handle)
	}
	if item.DisplayName != "Author Two" {
		t.Errorf("Expected display_name 'Author Two', got '%s'", item.DisplayName)
	}
	if item.LastSeenAt == nil || !item.LastSeenAt.Truncate(time.Millisecond).Equal(lastSeen2.Truncate(time.Millisecond)) {
		t.Errorf("Expected last_seen_at %v, got %v", lastSeen2, item.LastSeenAt)
	}
	if item.LastCheckedAt == nil || !item.LastCheckedAt.Truncate(time.Millisecond).Equal(lastChecked2.Truncate(time.Millisecond)) {
		t.Errorf("Expected last_checked_at %v, got %v", lastChecked2, item.LastCheckedAt)
	}
}

// testFollowingNoData tests the scenario when user follows no authors
func testFollowingNoData(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/following", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.FollowingListResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify empty response
	if len(response.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(response.Items))
	}
}

// testFollowingErrorCases tests error scenarios
func testFollowingErrorCases(t *testing.T, db *sqlx.DB) {
	router := NewTestRouter(db).GetEngine()

	// Test missing authorization header
	t.Run("MissingAuthHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/following", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

	var response dto.ErrorResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != "UNAUTHORIZED" {
			t.Errorf("Expected error code 'UNAUTHORIZED', got '%s'", response.Error.Code)
		}
	})

	// Test invalid user ID format
	t.Run("InvalidUserIDFormat", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/following", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", "invalid-uuid")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

	var response dto.ErrorResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != "INVALID_USER_ID" {
			t.Errorf("Expected error code 'INVALID_USER_ID', got '%s'", response.Error.Code)
		}
	})
}

// testFollowingMultipleUsers tests data isolation between users
func testFollowingMultipleUsers(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	user1ID := uuid.MustParse("00000000-0000-0000-0000-000000000021")
	user2ID := uuid.MustParse("00000000-0000-0000-0000-000000000022")
	now := time.Now().UTC()

	// Insert authors
	author1ID := int64(111111111)
	author2ID := int64(222222222)
	author3ID := int64(333333333)

	dataHelper.InsertAuthor(t, author1ID, "@shared1", StringPtr("Shared Author 1"), &now)
	dataHelper.InsertAuthor(t, author2ID, "@user1only", StringPtr("User1 Only"), &now)
	dataHelper.InsertAuthor(t, author3ID, "@user2only", StringPtr("User2 Only"), &now)

	// User 1 follows authors 1 and 2
	dataHelper.InsertUserFollowing(t, user1ID, author1ID, &now)
	dataHelper.InsertUserFollowing(t, user1ID, author2ID, &now)

	// User 2 follows authors 1 and 3
	dataHelper.InsertUserFollowing(t, user2ID, author1ID, &now)
	dataHelper.InsertUserFollowing(t, user2ID, author3ID, &now)

	router := NewTestRouter(db).GetEngine()

	// Request for user 1
	t.Run("User1Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/following", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user1ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

	var response dto.FollowingListResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// User 1 should see 2 authors
		if len(response.Items) != 2 {
			t.Errorf("Expected 2 items for user 1, got %d", len(response.Items))
		}
		// Verify authors (ordered by x_author_id DESC)
		expectedHandles := map[string]bool{"@shared1": true, "@user1only": true}
		for _, item := range response.Items {
			if !expectedHandles[item.Handle] {
				t.Errorf("Unexpected handle for user 1: %s", item.Handle)
			}
		}
	})

	// Request for user 2
	t.Run("User2Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/following", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user2ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

	var response dto.FollowingListResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// User 2 should see 2 authors
		if len(response.Items) != 2 {
			t.Errorf("Expected 2 items for user 2, got %d", len(response.Items))
		}

		// Verify authors (ordered by x_author_id DESC)
		expectedHandles := map[string]bool{"@shared1": true, "@user2only": true}
		for _, item := range response.Items {
			if !expectedHandles[item.Handle] {
				t.Errorf("Unexpected handle for user 2: %s", item.Handle)
			}
		}
	})
}
