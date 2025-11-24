package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
)

// AuthMiddleware creates a middleware that validates session tokens
func AuthMiddleware(authService services.AuthService, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header or cookie
		token := extractToken(c)
		if token == "" {
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
		_, err = db.ExecContext(c.Request.Context(), "SET LOCAL app.user_id = $1", user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
				Error: dto.ErrorDetailDTO{
					Code:    "INTERNAL_ERROR",
					Message: "Błąd podczas ustawiania kontekstu użytkownika",
				},
			})
			c.Abort()
			return
		}

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
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}
