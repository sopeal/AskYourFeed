package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/services"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// InitiateLogin handles POST /api/v1/auth/login endpoint
// Initiates OAuth login flow with X (Twitter)
func (h *AuthHandler) InitiateLogin(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Call service to initiate login
	response, err := h.authService.InitiateLogin(ctx)
	if err != nil {
		span.RecordError(err)
		h.respondWithError(c, http.StatusInternalServerError, "OAUTH_INIT_FAILED", "Nie udało się zainicjować logowania OAuth", nil)
		return
	}

	span.SetAttributes(
		attribute.String("auth_url", response.AuthURL[:50]+"..."), // Partial for security
		attribute.String("state", response.State[:8]+"..."),
	)

	// Return 302 redirect to X OAuth authorization URL
	c.Redirect(http.StatusFound, response.AuthURL)
}

// HandleCallback handles GET /api/v1/auth/callback endpoint
// Processes OAuth callback and creates session
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract query parameters
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDescription := c.Query("error_description")

	span.SetAttributes(
		attribute.String("code_present", boolToString(code != "")),
		attribute.String("state", state[:8]+"..."),
	)

	// Check for OAuth errors
	if errorParam != "" {
		span.SetAttributes(
			attribute.String("oauth_error", errorParam),
			attribute.String("error_description", errorDescription),
		)
		h.respondWithError(c, http.StatusBadRequest, "OAUTH_ERROR", "Błąd autoryzacji OAuth", map[string]interface{}{
			"error":             errorParam,
			"error_description": errorDescription,
		})
		return
	}

	// Validate required parameters
	if code == "" {
		h.respondWithError(c, http.StatusBadRequest, "MISSING_CODE", "Brak wymaganego parametru 'code'", nil)
		return
	}

	if state == "" {
		h.respondWithError(c, http.StatusBadRequest, "MISSING_STATE", "Brak wymaganego parametru 'state'", nil)
		return
	}

	// Process OAuth callback
	response, err := h.authService.HandleCallback(ctx, code, state)
	if err != nil {
		span.RecordError(err)
		// Determine error type for appropriate response
		if strings.Contains(err.Error(), "invalid or expired state") {
			h.respondWithError(c, http.StatusBadRequest, "INVALID_STATE", "Nieprawidłowy lub wygasły token stanu", nil)
			return
		}
		if strings.Contains(err.Error(), "failed to exchange code") {
			h.respondWithError(c, http.StatusUnauthorized, "OAUTH_EXCHANGE_FAILED", "Nie udało się wymienić kodu autoryzacyjnego", nil)
			return
		}
		h.respondWithError(c, http.StatusInternalServerError, "OAUTH_CALLBACK_FAILED", "Wystąpił błąd podczas przetwarzania callback OAuth", nil)
		return
	}

	span.SetAttributes(
		attribute.String("user_id", response.UserID.String()),
		attribute.String("x_handle", response.XHandle),
	)

	// Set session token as HTTP-only cookie
	c.SetCookie(
		"session_token",           // name
		response.SessionToken,     // value
		7*24*60*60,               // max age (7 days in seconds)
		"/",                      // path
		"",                       // domain (empty = current domain)
		false,                    // secure (set to true in production with HTTPS)
		true,                     // httpOnly
	)

	// Redirect to frontend dashboard
	c.Redirect(http.StatusFound, response.RedirectURL)
}

// Logout handles POST /api/v1/auth/logout endpoint
// Terminates user session
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract session token from Authorization header or cookie
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		h.respondWithError(c, http.StatusUnauthorized, "MISSING_SESSION_TOKEN", "Brak tokena sesji", nil)
		return
	}

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	// Call service to logout
	if err := h.authService.Logout(ctx, sessionToken); err != nil {
		span.RecordError(err)
		h.respondWithError(c, http.StatusInternalServerError, "LOGOUT_FAILED", "Wystąpił błąd podczas wylogowywania", nil)
		return
	}

	// Clear session cookie
	c.SetCookie(
		"session_token", // name
		"",             // value (empty to clear)
		-1,            // max age (negative to delete)
		"/",           // path
		"",            // domain
		false,         // secure
		true,          // httpOnly
	)

	// Return success response
	c.JSON(http.StatusOK, dto.MessageResponseDTO{
		Message: "Wylogowano pomyślnie",
	})
}

// GetCurrentSession handles GET /api/v1/session/current endpoint
// Returns current authenticated user session information
func (h *AuthHandler) GetCurrentSession(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	// Extract session token from Authorization header or cookie
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		h.respondWithError(c, http.StatusUnauthorized, "MISSING_SESSION_TOKEN", "Brak tokena sesji", nil)
		return
	}

	span.SetAttributes(
		attribute.String("session_token", sessionToken[:8]+"..."),
	)

	// Call service to get session info
	session, err := h.authService.GetCurrentSession(ctx, sessionToken)
	if err != nil {
		span.RecordError(err)
		h.respondWithError(c, http.StatusUnauthorized, "INVALID_SESSION", "Nieprawidłowa lub wygasła sesja", nil)
		return
	}

	span.SetAttributes(
		attribute.String("user_id", session.UserID.String()),
		attribute.String("x_handle", session.XHandle),
	)

	// Return session information
	c.JSON(http.StatusOK, session)
}

// extractSessionToken extracts session token from Authorization header or cookie
func (h *AuthHandler) extractSessionToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try session_token cookie
	if cookie, err := c.Cookie("session_token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// boolToString converts bool to string for tracing
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// respondWithError sends a standardized error response
func (h *AuthHandler) respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := dto.ErrorResponseDTO{
		Error: dto.ErrorDetailDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(statusCode, response)
}
