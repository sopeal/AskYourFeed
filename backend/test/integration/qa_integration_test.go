package integration

import (
	"bytes"
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

// TestQAIntegration contains all integration tests for the QA endpoints
func TestQAIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	dbHelper := NewDatabaseHelper(t)
	defer dbHelper.Close()

	db := dbHelper.GetDB()

	// Run test suite
	//t.Run("CreateQAHappyPath", func(t *testing.T) {
	//	testCreateQAHappyPath(t, dbHelper)
	//})

	t.Run("CreateQAValidationErrors", func(t *testing.T) {
		testCreateQAValidationErrors(t, dbHelper)
	})

	t.Run("CreateQAInvalidDateRange", func(t *testing.T) {
		testCreateQAInvalidDateRange(t, dbHelper)
	})

	t.Run("ListQAEmpty", func(t *testing.T) {
		testListQAEmpty(t, dbHelper)
	})

	t.Run("ListQAPagination", func(t *testing.T) {
		testListQAPagination(t, dbHelper)
	})

	t.Run("ListQAMultipleUsers", func(t *testing.T) {
		testListQAMultipleUsers(t, dbHelper)
	})

	t.Run("GetQAByIDHappyPath", func(t *testing.T) {
		testGetQAByIDHappyPath(t, dbHelper)
	})

	t.Run("GetQAByIDNotFound", func(t *testing.T) {
		testGetQAByIDNotFound(t, dbHelper)
	})

	t.Run("GetQAByIDWrongUser", func(t *testing.T) {
		testGetQAByIDWrongUser(t, dbHelper)
	})

	t.Run("DeleteQAHappyPath", func(t *testing.T) {
		testDeleteQAHappyPath(t, dbHelper)
	})

	t.Run("DeleteQANotFound", func(t *testing.T) {
		testDeleteQANotFound(t, dbHelper)
	})

	t.Run("DeleteAllQAHappyPath", func(t *testing.T) {
		testDeleteAllQAHappyPath(t, dbHelper)
	})

	t.Run("AuthenticationErrors", func(t *testing.T) {
		testAuthenticationErrors(t, db)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		testQAEdgeCases(t, dbHelper)
	})
}

// testCreateQAHappyPath tests successful QA creation
//func testCreateQAHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
//	dbHelper.CleanupTestData(t)
//
//	db := dbHelper.GetDB()
//	dataHelper := NewTestDataHelper(db)
//	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
//	now := time.Now().UTC()
//
//	// Insert test author and posts
//	dataHelper.InsertAuthor(t, 12345, "testuser", StringPtr("Test User"), &now)
//	dataHelper.InsertPost(t, userID, 1001, 12345, now.Add(-1*time.Hour), "https://twitter.com/testuser/status/1001", "This is a test post about AI and machine learning.", nil, now, now, false)
//	dataHelper.InsertPost(t, userID, 1002, 12345, now.Add(-2*time.Hour), "https://twitter.com/testuser/status/1002", "Another post discussing artificial intelligence trends.", nil, now, now, false)
//
//	router := NewTestRouter(db).GetEngine()
//
//	// Create QA request
//	requestBody := dto.CreateQACommand{
//		Question: "What are the latest trends in AI?",
//		DateFrom: &now,
//		DateTo:   &now,
//	}
//
//	jsonBody, _ := json.Marshal(requestBody)
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest("POST", "/api/v1/qa", bytes.NewBuffer(jsonBody))
//	req.Header.Set("Authorization", "Bearer test-token")
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("X-Test-User-ID", userID.String())
//
//	router.ServeHTTP(w, req)
//
//	// Assert response
//	if w.Code != http.StatusCreated {
//		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
//	}
//
//	var response dto.QADetailDTO
//	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
//		t.Fatalf("Failed to unmarshal response: %v", err)
//	}
//
//	// Verify response structure
//	if response.Question != requestBody.Question {
//		t.Errorf("Expected question %s, got %s", requestBody.Question, response.Question)
//	}
//	if response.Answer == "" {
//		t.Error("Expected non-empty answer")
//	}
//	if len(response.Sources) == 0 {
//		t.Error("Expected sources to be populated")
//	}
//}

// testCreateQAValidationErrors tests validation error scenarios
func testCreateQAValidationErrors(t *testing.T, dbHelper *DatabaseHelper) {
	db := dbHelper.GetDB()
	router := NewTestRouter(db).GetEngine()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	t.Run("EmptyQuestion", func(t *testing.T) {
		dbHelper.CleanupTestData(t)

		requestBody := dto.CreateQACommand{
			Question: "",
		}

		jsonBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/qa", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "QUESTION_REQUIRED" {
			t.Errorf("Expected error code 'QUESTION_REQUIRED', got '%s'", response.Error.Code)
		}
	})

	t.Run("QuestionTooLong", func(t *testing.T) {
		dbHelper.CleanupTestData(t)

		longQuestion := string(make([]byte, 2001)) // 2001 characters
		requestBody := dto.CreateQACommand{
			Question: longQuestion,
		}

		jsonBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/qa", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response dto.ErrorResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Error.Code != "QUESTION_TOO_LONG" {
			t.Errorf("Expected error code 'QUESTION_TOO_LONG', got '%s'", response.Error.Code)
		}
	})
}

// testCreateQAInvalidDateRange tests invalid date range scenarios
func testCreateQAInvalidDateRange(t *testing.T, dbHelper *DatabaseHelper) {
	db := dbHelper.GetDB()
	router := NewTestRouter(db).GetEngine()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	dbHelper.CleanupTestData(t)

	now := time.Now().UTC()
	dateFrom := now.Add(1 * time.Hour) // Future date
	dateTo := now                      // Past date

	requestBody := dto.CreateQACommand{
		Question: "Test question",
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/qa", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", w.Code)
	}

	var response dto.ErrorResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "INVALID_DATE_RANGE" {
		t.Errorf("Expected error code 'INVALID_DATE_RANGE', got '%s'", response.Error.Code)
	}
}

// testListQAEmpty tests listing QA when user has no QAs
func testListQAEmpty(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/qa", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.QAListResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(response.Items))
	}
	if response.HasMore {
		t.Error("Expected has_more to be false")
	}
}

// testListQAPagination tests QA listing with pagination
func testListQAPagination(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	now := time.Now().UTC()

	// Insert test data: 5 QA messages
	for i := 0; i < 5; i++ {
		createdAt := now.Add(-time.Duration(i) * time.Hour)
		dataHelper.InsertQAMessage(t, userID, "Question "+string(rune(i+65)), "Answer "+string(rune(i+65)), now.Add(-24*time.Hour), now, createdAt)
	}

	router := NewTestRouter(db).GetEngine()

	// Test default pagination (limit 20)
	t.Run("DefaultPagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.QAListResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Items) != 5 {
			t.Errorf("Expected 5 items, got %d", len(response.Items))
		}
		if response.HasMore {
			t.Error("Expected has_more to be false")
		}
	})

	// Test custom limit
	t.Run("CustomLimit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa?limit=2", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.QAListResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(response.Items))
		}
		if !response.HasMore {
			t.Error("Expected has_more to be true")
		}
		if response.NextCursor == "" {
			t.Error("Expected next_cursor to be set")
		}
	})
}

// testListQAMultipleUsers tests data isolation between users
func testListQAMultipleUsers(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	user1ID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	user2ID := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	now := time.Now().UTC()

	// Insert data for user 1
	dataHelper.InsertQAMessage(t, user1ID, "User 1 Question", "User 1 Answer", now.Add(-24*time.Hour), now, now.Add(-1*time.Hour))
	dataHelper.InsertQAMessage(t, user1ID, "User 1 Question 2", "User 1 Answer 2", now.Add(-24*time.Hour), now, now.Add(-2*time.Hour))

	// Insert data for user 2
	dataHelper.InsertQAMessage(t, user2ID, "User 2 Question", "User 2 Answer", now.Add(-24*time.Hour), now, now.Add(-1*time.Hour))

	router := NewTestRouter(db).GetEngine()

	// Request for user 1
	t.Run("User1Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user1ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.QAListResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Items) != 2 {
			t.Errorf("Expected 2 items for user 1, got %d", len(response.Items))
		}
		if response.Items[0].Question != "User 1 Question" {
			t.Errorf("Expected first item to be 'User 1 Question', got '%s'", response.Items[0].Question)
		}
	})

	// Request for user 2
	t.Run("User2Data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", user2ID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response dto.QAListResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Items) != 1 {
			t.Errorf("Expected 1 item for user 2, got %d", len(response.Items))
		}
		if response.Items[0].Question != "User 2 Question" {
			t.Errorf("Expected item to be 'User 2 Question', got '%s'", response.Items[0].Question)
		}
	})
}

// testGetQAByIDHappyPath tests successful QA retrieval by ID
func testGetQAByIDHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000006")
	now := time.Now().UTC()

	// Insert test data
	dataHelper.InsertAuthor(t, 12346, "testuser2", StringPtr("Test User 2"), &now)
	dataHelper.InsertPost(t, userID, 2001, 12346, now.Add(-1*time.Hour), "https://twitter.com/testuser2/status/2001", "Test post content", nil, now, now, false)

	qaID := dataHelper.InsertQAMessage(t, userID, "Test question", "Test answer", now.Add(-24*time.Hour), now, now)
	dataHelper.InsertQASource(t, qaID, userID, 2001)

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/qa/"+qaID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.QADetailDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != qaID {
		t.Errorf("Expected ID %s, got %s", qaID, response.ID)
	}
	if response.Question != "Test question" {
		t.Errorf("Expected question 'Test question', got '%s'", response.Question)
	}
	if response.Answer != "Test answer" {
		t.Errorf("Expected answer 'Test answer', got '%s'", response.Answer)
	}
	if len(response.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(response.Sources))
	}
}

// testGetQAByIDNotFound tests QA retrieval when ID doesn't exist
func testGetQAByIDNotFound(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000007")

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/qa/01HXXXXXXXXXXXXXXXXXXXXX", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response dto.ErrorResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got '%s'", response.Error.Code)
	}
}

// testGetQAByIDWrongUser tests QA retrieval when QA belongs to different user
func testGetQAByIDWrongUser(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	user1ID := uuid.MustParse("00000000-0000-0000-0000-000000000008")
	user2ID := uuid.MustParse("00000000-0000-0000-0000-000000000009")
	now := time.Now().UTC()

	// Insert QA for user 1
	qaID := dataHelper.InsertQAMessage(t, user1ID, "User 1 question", "User 1 answer", now.Add(-24*time.Hour), now, now)

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/qa/"+qaID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", user2ID.String()) // Try to access with user 2

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response dto.ErrorResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got '%s'", response.Error.Code)
	}
}

// testDeleteQAHappyPath tests successful QA deletion
func testDeleteQAHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	now := time.Now().UTC()

	// Insert QA
	qaID := dataHelper.InsertQAMessage(t, userID, "Test question", "Test answer", now.Add(-24*time.Hour), now, now)

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/qa/"+qaID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.MessageResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Message == "" {
		t.Error("Expected success message")
	}
}

// testDeleteQANotFound tests QA deletion when ID doesn't exist
func testDeleteQANotFound(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000011")

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/qa/01HXXXXXXXXXXXXXXXXXXXXX", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response dto.ErrorResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got '%s'", response.Error.Code)
	}
}

// testDeleteAllQAHappyPath tests successful deletion of all user QAs
func testDeleteAllQAHappyPath(t *testing.T, dbHelper *DatabaseHelper) {
	dbHelper.CleanupTestData(t)

	db := dbHelper.GetDB()
	dataHelper := NewTestDataHelper(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	now := time.Now().UTC()

	// Insert multiple QAs
	dataHelper.InsertQAMessage(t, userID, "Question 1", "Answer 1", now.Add(-24*time.Hour), now, now)
	dataHelper.InsertQAMessage(t, userID, "Question 2", "Answer 2", now.Add(-24*time.Hour), now, now)
	dataHelper.InsertQAMessage(t, userID, "Question 3", "Answer 3", now.Add(-24*time.Hour), now, now)

	router := NewTestRouter(db).GetEngine()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/qa", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-User-ID", userID.String())

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response dto.DeleteAllQAResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.DeletedCount != 3 {
		t.Errorf("Expected deleted_count 3, got %d", response.DeletedCount)
	}
	if response.Message == "" {
		t.Error("Expected success message")
	}
}

// testAuthenticationErrors tests authentication error scenarios
func testAuthenticationErrors(t *testing.T, db *sqlx.DB) {
	router := NewTestRouter(db).GetEngine()

	t.Run("MissingAuthHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa", nil)

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

	t.Run("InvalidUserIDFormat", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa", nil)
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

// testQAEdgeCases tests edge cases and boundary conditions for QA endpoints
func testQAEdgeCases(t *testing.T, dbHelper *DatabaseHelper) {
	db := dbHelper.GetDB()
	router := NewTestRouter(db).GetEngine()

	t.Run("InvalidQAID", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000013")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/qa/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	//t.Run("EmptyQAID", func(t *testing.T) {
	//	dbHelper.CleanupTestData(t)
	//	userID := uuid.MustParse("00000000-0000-0000-0000-000000000014")
	//
	//	w := httptest.NewRecorder()
	//	req, _ := http.NewRequest("GET", "/api/v1/qa/", nil)
	//	req.Header.Set("Authorization", "Bearer test-token")
	//	req.Header.Set("X-Test-User-ID", userID.String())
	//
	//	router.ServeHTTP(w, req)
	//
	//	if w.Code != http.StatusNotFound {
	//		t.Errorf("Expected status 404, got %d", w.Code)
	//	}
	//})

	t.Run("MalformedJSON", func(t *testing.T) {
		dbHelper.CleanupTestData(t)
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000015")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/qa", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", userID.String())

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
