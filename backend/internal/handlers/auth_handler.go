package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var cmd dto.RegisterCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "INVALID_REQUEST",
				Message: "Nieprawidłowe dane wejściowe",
				Details: map[string]interface{}{"error": err.Error()},
			},
		})
		return
	}

	// Register user
	response, err := h.authService.Register(c.Request.Context(), cmd)
	if err != nil {
		// Determine appropriate error code and status
		statusCode := http.StatusInternalServerError
		errorCode := "REGISTRATION_ERROR"
		message := "Błąd podczas rejestracji"

		errMsg := err.Error()
		if strings.Contains(errMsg, "email already registered") {
			statusCode = http.StatusConflict
			errorCode = "EMAIL_ALREADY_EXISTS"
			message = "Email jest już zarejestrowany"
		} else if strings.Contains(errMsg, "X username validation failed") {
			statusCode = http.StatusUnprocessableEntity
			errorCode = "INVALID_X_USERNAME"
			message = "Nazwa użytkownika X nie istnieje"
		} else if strings.Contains(errMsg, "password") {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_PASSWORD"
			message = errMsg
		} else if strings.Contains(errMsg, "temporarily unavailable") {
			statusCode = http.StatusServiceUnavailable
			errorCode = "SERVICE_UNAVAILABLE"
			message = "Usługa twitterapi.io jest tymczasowo niedostępna"
		}

		c.JSON(statusCode, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    errorCode,
				Message: message,
				Details: map[string]interface{}{"error": errMsg},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var cmd dto.LoginCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "INVALID_REQUEST",
				Message: "Nieprawidłowe dane wejściowe",
				Details: map[string]interface{}{"error": err.Error()},
			},
		})
		return
	}

	// Login user
	response, err := h.authService.Login(c.Request.Context(), cmd)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "LOGIN_ERROR"
		message := "Błąd podczas logowania"

		if strings.Contains(err.Error(), "invalid email or password") {
			statusCode = http.StatusUnauthorized
			errorCode = "INVALID_CREDENTIALS"
			message = "Nieprawidłowy email lub hasło"
		}

		c.JSON(statusCode, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    errorCode,
				Message: message,
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from Authorization header
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "UNAUTHORIZED",
				Message: "Brak tokenu autoryzacji",
			},
		})
		return
	}

	// Logout user
	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "LOGOUT_ERROR",
				Message: "Błąd podczas wylogowania",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponseDTO{
		Message: "Wylogowano pomyślnie",
	})
}

// GetCurrentSession returns current session information
// GET /api/v1/session/current
func (h *AuthHandler) GetCurrentSession(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "UNAUTHORIZED",
				Message: "Brak autoryzacji",
			},
		})
		return
	}

	// Get current session
	session, err := h.authService.GetCurrentSession(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "SESSION_ERROR",
				Message: "Błąd podczas pobierania sesji",
			},
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// extractToken extracts the bearer token from the Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
