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

// TestRouter creates a test router with the ingest and QA handlers
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

	// Initialize QA dependencies
	qaRepo := repositories.NewQARepository(db)
	postRepo := repositories.NewPostRepository(db)
	llmService := services.NewLLMService() // Mock service for testing
	qaService := services.NewQAService(db, postRepo, qaRepo, llmService)
	qaHandler := handlers.NewQAHandler(qaService)

	// Setup routes with auth middleware
	v1 := router.Group("/api/v1")
	ingest := v1.Group("/ingest")
	ingest.Use(testAuthMiddleware())
	{
		ingest.GET("/status", ingestHandler.GetIngestStatus)
	}

	qa := v1.Group("/qa")
	qa.Use(testAuthMiddleware())
	{
		qa.POST("", qaHandler.CreateQA)
		qa.GET("", qaHandler.ListQA)
		qa.GET("/:id", qaHandler.GetQAByID)
		qa.DELETE("/:id", qaHandler.DeleteQA)
		qa.DELETE("", qaHandler.DeleteAllQA)
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

// InsertAuthor inserts a test author into the database
func (tdh *TestDataHelper) InsertAuthor(
	t *testing.T,
	xAuthorID int64,
	handle string,
	displayName *string,
	lastSeenAt *time.Time,
) {
	t.Helper()

	query := `
		INSERT INTO authors (x_author_id, handle, display_name, last_seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (x_author_id) DO NOTHING
	`

	_, err := tdh.db.Exec(query, xAuthorID, handle, displayName, lastSeenAt)
	if err != nil {
		t.Fatalf("Failed to insert test author: %v", err)
	}
}

// InsertPost inserts a test post into the database
func (tdh *TestDataHelper) InsertPost(
	t *testing.T,
	userID uuid.UUID,
	xPostID int64,
	authorID int64,
	publishedAt time.Time,
	url string,
	text string,
	conversationID *int64,
	ingestedAt time.Time,
	firstVisibleAt time.Time,
	editedSeen bool,
) {
	t.Helper()

	query := `
		INSERT INTO posts (user_id, x_post_id, author_id, published_at, url, text, conversation_id, ingested_at, first_visible_at, edited_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tdh.db.Exec(query, userID, xPostID, authorID, publishedAt, url, text, conversationID, ingestedAt, firstVisibleAt, editedSeen)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
}

// InsertQAMessage inserts a test QA message into the database
func (tdh *TestDataHelper) InsertQAMessage(
	t *testing.T,
	userID uuid.UUID,
	question string,
	answer string,
	dateFrom time.Time,
	dateTo time.Time,
	createdAt time.Time,
) string {
	t.Helper()

	id := ulid.Make().String()

	query := `
		INSERT INTO qa_messages (id, user_id, question, answer, date_from, date_to, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tdh.db.Exec(query, id, userID, question, answer, dateFrom, dateTo, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert test QA message: %v", err)
	}

	return id
}

// InsertQASource inserts a test QA source into the database
func (tdh *TestDataHelper) InsertQASource(
	t *testing.T,
	qaID string,
	userID uuid.UUID,
	xPostID int64,
) {
	t.Helper()

	query := `
		INSERT INTO qa_sources (qa_id, user_id, x_post_id)
		VALUES ($1, $2, $3)
	`

	_, err := tdh.db.Exec(query, qaID, userID, xPostID)
	if err != nil {
		t.Fatalf("Failed to insert test QA source: %v", err)
	}
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}
