package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FollowingHandler handles following-related HTTP requests
type FollowingHandler struct {
	followingService *services.FollowingService
}

// NewFollowingHandler creates a new FollowingHandler instance
func NewFollowingHandler(followingService *services.FollowingService) *FollowingHandler {
	return &FollowingHandler{
		followingService: followingService,
	}
}

// GetFollowing handles GET /api/v1/following endpoint
// Returns paginated list of authors the user follows
func (h *FollowingHandler) GetFollowing(c *gin.Context) {
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
	limit := 50 // default value
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

		if parsedLimit > 200 {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_LIMIT", "Parametr 'limit' nie może przekraczać 200", map[string]interface{}{
				"provided_value": parsedLimit,
				"max_value":      200,
			})
			return
		}

		limit = parsedLimit
	}

	span.SetAttributes(attribute.Int("limit", limit))

	// Parse and validate cursor query parameter
	var cursor *int64
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		parsedCursor, err := strconv.ParseInt(cursorStr, 10, 64)
		if err != nil {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_CURSOR", "Parametr 'cursor' musi być liczbą całkowitą", map[string]interface{}{
				"provided_value": cursorStr,
			})
			return
		}

		if parsedCursor <= 0 {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_CURSOR", "Parametr 'cursor' musi być większy niż 0", map[string]interface{}{
				"provided_value": parsedCursor,
			})
			return
		}

		cursor = &parsedCursor
		span.SetAttributes(attribute.Int64("cursor", parsedCursor))
	}

	// Call service layer to get following list
	response, err := h.followingService.GetFollowing(ctx, userID, limit, cursor)
	if err != nil {
		span.RecordError(err)
		h.respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Wystąpił błąd podczas pobierania listy obserwowanych", nil)
		return
	}

	// Return 200 OK with response
	c.JSON(http.StatusOK, response)
}

// respondWithError sends a standardized error response
func (h *FollowingHandler) respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := dto.ErrorResponseDTO{
		Error: dto.ErrorDetailDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(statusCode, response)
}
