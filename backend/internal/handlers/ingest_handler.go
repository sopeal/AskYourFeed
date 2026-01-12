package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
	"github.com/sopeal/AskYourFeed/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IngestHandler handles ingestion-related HTTP requests
type IngestHandler struct {
	ingestStatusService *services.IngestStatusService
	ingestService       *services.IngestService
}

// NewIngestHandler creates a new IngestHandler instance
func NewIngestHandler(ingestStatusService *services.IngestStatusService, ingestService *services.IngestService) *IngestHandler {
	return &IngestHandler{
		ingestStatusService: ingestStatusService,
		ingestService:       ingestService,
	}
}

// GetIngestStatus handles GET /api/v1/ingest/status endpoint
// Returns current and recent ingestion run information
func (h *IngestHandler) GetIngestStatus(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Nieprawidłowy lub wygasły token sesji", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Invalid user ID format", nil)
		return
	}

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// Parse and validate limit query parameter
	limit := 10 // default value
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_LIMIT", "Parametr 'limit' musi być liczbą całkowitą", map[string]interface{}{
				"provided_value": limitStr,
			})
			return
		}

		if parsedLimit < 1 {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_LIMIT", "Parametr 'limit' musi być większy niż 0", map[string]interface{}{
				"provided_value": parsedLimit,
				"min_value":      1,
			})
			return
		}

		if parsedLimit > 50 {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_LIMIT", "Parametr 'limit' nie może przekraczać 50", map[string]interface{}{
				"provided_value": parsedLimit,
				"max_value":      50,
			})
			return
		}

		limit = parsedLimit
	}

	span.SetAttributes(attribute.Int("limit", limit))

	// Call service layer to get ingest status
	status, err := h.ingestStatusService.GetIngestStatus(ctx, userID, limit)
	if err != nil {
		span.RecordError(err)
		logger.Error("failed to get ingest status",
			err,
			"user_id", userID,
			"limit", limit,
			"path", c.Request.URL.Path)
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Wystąpił błąd podczas pobierania statusu ingestion", nil)
		return
	}

	// Return 200 OK with response
	c.JSON(http.StatusOK, status)
}

// TriggerIngest handles POST /api/v1/ingest/trigger endpoint
// Triggers a new ingestion run for the authenticated user
func (h *IngestHandler) TriggerIngest(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Nieprawidłowy lub wygasły token sesji", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Invalid user ID format", nil)
		return
	}

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// Parse request body (optional backfill hours)
	var req dto.TriggerIngestCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		// If body is empty or invalid, use defaults
		req.BackfillHours = 24
	}

	// Set default backfill hours if not provided
	if req.BackfillHours == 0 {
		req.BackfillHours = 24 // Default to 24 hours
	}

	// Validate backfill hours range
	if req.BackfillHours < 0 || req.BackfillHours > 720 {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_BACKFILL_HOURS", "Backfill hours must be between 0 and 720", map[string]interface{}{
			"provided_value": req.BackfillHours,
			"min_value":      0,
			"max_value":      720,
		})
		return
	}

	span.SetAttributes(attribute.Int("backfill_hours", req.BackfillHours))

	// Check if there's already a running ingest (409 Conflict)
	currentRun, err := h.ingestStatusService.GetIngestStatus(ctx, userID, 1)
	if err != nil {
		span.RecordError(err)
		logger.Error("failed to check current ingest status",
			err,
			"user_id", userID)
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Wystąpił błąd podczas sprawdzania statusu ingestion", nil)
		return
	}

	if currentRun.CurrentRun != nil {
		h.respondWithError(c, http.StatusConflict, "INGEST_IN_PROGRESS", "Ingestion już trwa dla tego użytkownika", map[string]interface{}{
			"current_run_id": currentRun.CurrentRun.ID,
			"started_at":     currentRun.CurrentRun.StartedAt,
		})
		return
	}

	// Generate run ID for response
	runID := ulid.Make().String()
	startedAt := time.Now()

	// Trigger ingestion asynchronously (don't wait for completion)
	go func() {
		// Create a new context for the background operation
		backgroundCtx := context.Background()
		err := h.ingestService.IngestUserData(backgroundCtx, userID, req.BackfillHours)
		if err != nil {
			logger.Error("background ingestion failed",
				err,
				"user_id", userID,
				"backfill_hours", req.BackfillHours)
		} else {
			logger.Info("background ingestion completed successfully",
				"user_id", userID,
				"backfill_hours", req.BackfillHours)
		}
	}()

	// Return immediate response indicating ingestion was triggered
	response := dto.TriggerIngestResponseDTO{
		IngestRunID: runID,
		Status:      "triggered",
		StartedAt:   startedAt,
	}

	c.JSON(http.StatusAccepted, response)
}

// respondWithError sends a standardized error response
func (h *IngestHandler) respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := dto.ErrorResponseDTO{
		Error: dto.ErrorDetailDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(statusCode, response)
}
