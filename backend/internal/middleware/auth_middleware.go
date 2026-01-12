package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
	"github.com/sopeal/AskYourFeed/pkg/logger"
)

// AuthMiddleware creates a middleware that validates session tokens
func AuthMiddleware(authService services.AuthService, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header or cookie
		token := extractToken(c)
		if token == "" {
			logger.Warn("unauthorized access attempt - no token",
				"ip", c.ClientIP(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"user_agent", c.Request.UserAgent())
			c.JSON(http.StatusUnauthorized, dto.ErrorResponseDTO{
				Error: dto.ErrorDetailDTO{
					Code:    "UNAUTHORIZED",
					Message: "Brak tokenu autoryzacji",
				},
			})
			c.Abort()
			return
		}

		// Validate session
		user, err := authService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			logger.Warn("session validation failed",
				"error", err,
				"ip", c.ClientIP(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponseDTO{
				Error: dto.ErrorDetailDTO{
					Code:    "UNAUTHORIZED",
					Message: "Nieprawidłowy lub wygasły token sesji",
				},
			})
			c.Abort()
			return
		}

		// Set user_id in PostgreSQL session for RLS
		// Note: SET command doesn't support parameterized queries, so we use fmt.Sprintf
		// This is safe because user.ID is a UUID from the database
		_, err = db.ExecContext(c.Request.Context(), fmt.Sprintf("SET LOCAL app.user_id = '%s'", user.ID))
		if err != nil {
			logger.Error("failed to set database user context",
				err,
				"user_id", user.ID,
				"path", c.Request.URL.Path)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
				Error: dto.ErrorDetailDTO{
					Code:    "INTERNAL_ERROR",
					Message: "Błąd podczas ustawiania kontekstu użytkownika",
				},
			})
			c.Abort()
			return
		}

		// Log successful authentication at debug level
		logger.Debug("user authenticated",
			"user_id", user.ID,
			"path", c.Request.URL.Path,
			"method", c.Request.Method)

		// Set user_id in Gin context for handlers to use
		c.Set("user_id", user.ID)
		c.Set("user", user)

		c.Next()
	}
}

// extractToken extracts the bearer token from Authorization header or cookie
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try cookie as fallback
	cookie, err := c.Cookie("session_token")
	if err != nil {
		logger.Debug("failed to extract cookie",
			"error", err,
			"cookies", c.Request.Header.Get("Cookie"))
	}
	if err == nil && cookie != "" {
		logger.Debug("found session_token in cookie",
			"token_prefix", cookie[:min(10, len(cookie))]+"...")
		return cookie
	}

	return ""
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
