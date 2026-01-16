package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
	"github.com/sopeal/AskYourFeed/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// QAHandler handles Q&A related HTTP requests
type QAHandler struct {
	qaService *services.QAService
	validator *validator.Validate
}

// NewQAHandler creates a new QAHandler instance
func NewQAHandler(qaService *services.QAService) *QAHandler {
	return &QAHandler{
		qaService: qaService,
		validator: validator.New(),
	}
}

// CreateQA handles POST /api/v1/qa endpoint
// Creates a new Q&A interaction by submitting a question to the LLM service
func (h *QAHandler) CreateQA(c *gin.Context) {
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

	// Bind and validate request body
	var cmd dto.CreateQACommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_INPUT", "Nieprawidłowe dane wejściowe", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Validate question field
	if err := h.validator.Struct(cmd); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		details := make(map[string]interface{})

		for _, fieldErr := range validationErrors {
			switch fieldErr.Field() {
			case "Question":
				if fieldErr.Tag() == "required" {
					h.respondWithError(c, http.StatusBadRequest, "QUESTION_REQUIRED", "Pytanie jest wymagane i nie może być puste", map[string]interface{}{
						"field": "question",
					})
					return
				} else if fieldErr.Tag() == "max" {
					h.respondWithError(c, http.StatusBadRequest, "QUESTION_TOO_LONG", "Pytanie przekracza maksymalną długość 2000 znaków", map[string]interface{}{
						"field":      "question",
						"max_length": 2000,
					})
					return
				}
			}
		}

		h.respondWithError(c, http.StatusBadRequest, "INVALID_INPUT", "Walidacja nie powiodła się", details)
		return
	}

	// Apply date defaults
	now := time.Now()
	dateFrom := cmd.DateFrom
	dateTo := cmd.DateTo

	if dateFrom == nil {
		defaultFrom := now.Add(-24 * time.Hour)
		dateFrom = &defaultFrom
	}

	if dateTo == nil {
		dateTo = &now
	}

	// Validate date range
	if dateFrom.After(*dateTo) {
		h.respondWithError(c, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Data początkowa musi być wcześniejsza lub równa dacie końcowej", map[string]interface{}{
			"date_from": dateFrom.Format(time.RFC3339),
			"date_to":   dateTo.Format(time.RFC3339),
		})
		return
	}

	span.SetAttributes(
		attribute.String("date_from", dateFrom.Format(time.RFC3339)),
		attribute.String("date_to", dateTo.Format(time.RFC3339)),
		attribute.Int("question_length", len(cmd.Question)),
	)

	// Call service layer to create Q&A
	qaDetail, err := h.qaService.CreateQA(ctx, userID, cmd.Question, *dateFrom, *dateTo)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Return 201 Created with response
	c.JSON(http.StatusCreated, qaDetail)
}

// respondWithError sends a standardized error response
func (h *QAHandler) respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := dto.ErrorResponseDTO{
		Error: dto.ErrorDetailDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(statusCode, response)
}

// ListQA handles GET /api/v1/qa endpoint
// Returns paginated list of user's Q&A history
func (h *QAHandler) ListQA(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context
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

	// Parse query parameters
	limit := 20 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit < 1 {
				limit = 1
			} else if parsedLimit > 100 {
				limit = 100
			} else {
				limit = parsedLimit
			}
		}
	}

	cursor := c.Query("cursor")

	span.SetAttributes(
		attribute.Int("limit", limit),
		attribute.String("cursor", cursor),
	)

	// Call service layer
	result, err := h.qaService.ListQA(ctx, userID, limit, cursor)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetQAByID handles GET /api/v1/qa/{id} endpoint
// Returns full details of a specific Q&A interaction
func (h *QAHandler) GetQAByID(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context
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

	// Extract Q&A ID from URL parameter
	qaID := c.Param("id")
	if qaID == "" {
		h.respondWithError(c, http.StatusBadRequest, "MISSING_ID", "Brak identyfikatora Q&A", nil)
		return
	}

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("qa_id", qaID),
	)

	// Call service layer
	result, err := h.qaService.GetQAByID(ctx, userID, qaID)
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, services.ErrQANotFound) {
			h.respondWithError(c, http.StatusNotFound, "NOT_FOUND", "Q&A o podanym ID nie zostało znalezione lub nie należy do użytkownika", nil)
			return
		}
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteQA handles DELETE /api/v1/qa/{id} endpoint
// Deletes a specific Q&A interaction
func (h *QAHandler) DeleteQA(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context
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

	// Extract Q&A ID from URL parameter
	qaID := c.Param("id")
	if qaID == "" {
		h.respondWithError(c, http.StatusBadRequest, "MISSING_ID", "Brak identyfikatora Q&A", nil)
		return
	}

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("qa_id", qaID),
	)

	// Call service layer
	err := h.qaService.DeleteQA(ctx, userID, qaID)
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, services.ErrQANotFound) {
			h.respondWithError(c, http.StatusNotFound, "NOT_FOUND", "Q&A o podanym ID nie zostało znalezione lub nie należy do użytkownika", nil)
			return
		}
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponseDTO{
		Message: "Q&A usunięte pomyślnie",
	})
}

// DeleteAllQA handles DELETE /api/v1/qa endpoint
// Deletes all Q&A history for the authenticated user
func (h *QAHandler) DeleteAllQA(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user_id from context
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

	// Call service layer
	deletedCount, err := h.qaService.DeleteAllQA(ctx, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	span.SetAttributes(attribute.Int("deleted_count", deletedCount))

	c.JSON(http.StatusOK, dto.DeleteAllQAResponseDTO{
		Message:      "Cała historia Q&A została usunięta",
		DeletedCount: deletedCount,
	})
}

// handleServiceError maps service errors to appropriate HTTP status codes
func (h *QAHandler) handleServiceError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)

	// Log error with context
	userID, _ := c.Get("user_id")
	logger.Error("service error in QA handler",
		err,
		"user_id", userID,
		"path", c.Request.URL.Path,
		"method", c.Request.Method)

	// Map service errors to HTTP status codes
	switch err {
	case services.ErrLLMUnavailable:
		h.respondWithError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Usługa LLM jest tymczasowo niedostępna", nil)
	case services.ErrRateLimitExceeded:
		h.respondWithError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Przekroczono limit zapytań. Spróbuj ponownie później", nil)
	default:
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Wystąpił błąd serwera. Spróbuj ponownie później", nil)
	}
}
