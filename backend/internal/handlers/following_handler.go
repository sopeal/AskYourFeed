package handlers

import (
	"net/http"

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

	// Call service layer to get following list
	response, err := h.followingService.GetFollowing(ctx, userID)
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
