package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
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
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "Nieprawidłowe dane żądania", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Set default backfill hours if not provided
	if req.BackfillHours == 0 {
		req.BackfillHours = 24 // Default to 24 hours
	}

	span.SetAttributes(attribute.Int("backfill_hours", req.BackfillHours))

	// Trigger ingestion asynchronously (don't wait for completion)
	go func() {
		// Create a new context for the background operation
		backgroundCtx := context.Background()
		err := h.ingestService.IngestUserData(backgroundCtx, userID)
		if err != nil {
			// Log error - in a real implementation, you'd want proper logging here
			// For now, we'll just ignore background errors
			_ = err
		}
	}()

	// Return immediate response indicating ingestion was triggered
	response := dto.TriggerIngestResponseDTO{
		IngestRunID: "background", // Placeholder - in real implementation, return actual run ID
		Status:      "triggered",
		StartedAt:   time.Now(),
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
