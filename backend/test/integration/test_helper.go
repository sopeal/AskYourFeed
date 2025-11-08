package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/sopeal/AskYourFeed/internal/handlers"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"github.com/sopeal/AskYourFeed/internal/services"
)

// TestRouter creates a test router with the ingest handler
type TestRouter struct {
	engine *gin.Engine
}

// NewTestRouter creates a new test router
func NewTestRouter(db *sqlx.DB) *TestRouter {
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize dependencies
	ingestRepo := repositories.NewIngestRepository(db)
	ingestService := services.NewIngestService(ingestRepo)
	ingestHandler := handlers.NewIngestHandler(ingestService)

	// Setup routes with auth middleware
	v1 := router.Group("/api/v1")
	ingest := v1.Group("/ingest")
	ingest.Use(testAuthMiddleware())
	{
		ingest.GET("/status", ingestHandler.GetIngestStatus)
	}

	return &TestRouter{engine: router}
}

// GetEngine returns the Gin engine
func (tr *TestRouter) GetEngine() *gin.Engine {
	return tr.engine
}

// testAuthMiddleware is a test version of auth middleware that sets user_id from header
func testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		// Extract user_id from X-Test-User-ID header for testing
		userIDStr := c.GetHeader("X-Test-User-ID")
		if userIDStr == "" {
			// Default test user
			userIDStr = "00000000-0000-0000-0000-000000000001"
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_USER_ID",
					"message": "Invalid user ID format",
				},
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// TestDataHelper provides utilities for inserting test data
type TestDataHelper struct {
	db *sqlx.DB
}

// NewTestDataHelper creates a new test data helper
func NewTestDataHelper(db *sqlx.DB) *TestDataHelper {
	return &TestDataHelper{db: db}
}

// InsertIngestRun inserts a test ingest run into the database
func (tdh *TestDataHelper) InsertIngestRun(
	t *testing.T,
	userID uuid.UUID,
	startedAt time.Time,
	completedAt *time.Time,
	status string,
	fetchedCount, retried, rateLimitHits int,
	errText *string,
) string {
	t.Helper()

	id := ulid.Make().String()
	sinceID := int64(1000000000)

	query := `
		INSERT INTO ingest_runs (id, user_id, started_at, completed_at, status, since_id, fetched_count, retried, rate_limit_hits, err_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tdh.db.Exec(query, id, userID, startedAt, completedAt, status, sinceID, fetchedCount, retried, rateLimitHits, errText)
	if err != nil {
		t.Fatalf("Failed to insert test ingest run: %v", err)
	}

	return id
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}
