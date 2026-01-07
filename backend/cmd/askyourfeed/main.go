package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sopeal/AskYourFeed/internal/handlers"
	"github.com/sopeal/AskYourFeed/internal/middleware"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"github.com/sopeal/AskYourFeed/internal/services"
	"github.com/sopeal/AskYourFeed/pkg/logger"
)

func main() {
	// Initialize logger
	logLevel := getLogLevel()
	logger.Init(logLevel)

	// Load configuration from environment variables
	config := loadConfig()

	logger.Info("starting AskYourFeed backend",
		"version", "1.0.0",
		"port", config.Port,
		"log_level", logLevel.String(),
		"database_url", maskDatabaseURL(config.DatabaseURL))

	// Initialize database connection
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
		logger.Error("failed to initialize database", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	postRepo := repositories.NewPostRepository(db)
	qaRepo := repositories.NewQARepository(db)
	ingestRepo := repositories.NewIngestRepository(db)
	followingRepo := repositories.NewFollowingRepository(db)
	authorRepo := repositories.NewAuthorRepository(db)
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)

	// Initialize Twitter API client
	twitterClient := services.NewTwitterClient(config.TwitterAPIKey, nil)

	// Initialize OpenRouter client for ingestion (optional - only if API key is provided)
	var openRouterClient *services.OpenRouterClient
	if config.OpenRouterAPIKey != "" {
		openRouterClient = services.NewOpenRouterClient(config.OpenRouterAPIKey, nil)
		logger.Info("OpenRouter client initialized for media processing")
	} else {
		logger.Warn("OpenRouter API key not provided - media processing will be skipped")
	}

	// Initialize OpenRouter client for Q&A (separate API key)
	var openRouterQAClient *services.OpenRouterClient
	if config.OpenRouterQAAPIKey != "" {
		openRouterQAClient = services.NewOpenRouterClient(config.OpenRouterQAAPIKey, nil)
		logger.Info("OpenRouter Q&A client initialized")
	} else {
		logger.Warn("OpenRouter Q&A API key not provided - Q&A functionality will be unavailable")
	}

	// Initialize services
	llmService := services.NewLLMService(openRouterQAClient)
	qaService := services.NewQAService(db, postRepo, qaRepo, llmService)
	ingestStatusService := services.NewIngestStatusService(ingestRepo)
	followingService := services.NewFollowingService(followingRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, *twitterClient)

	// Initialize ingestion service
	ingestService := services.NewIngestService(
		twitterClient,
		openRouterClient,
		ingestRepo,
		followingRepo,
		postRepo,
		authorRepo,
		userRepo,
	)

	// Initialize handlers
	qaHandler := handlers.NewQAHandler(qaService)
	ingestHandler := handlers.NewIngestHandler(ingestStatusService, ingestService)
	followingHandler := handlers.NewFollowingHandler(followingService)
	authHandler := handlers.NewAuthHandler(authService)

	// Set up HTTP router
	router := setupRouter(db, authService, authHandler, qaHandler, ingestHandler, followingHandler)

	// Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("HTTP server listening", "port", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("server exited successfully")
}

// Config holds application configuration
type Config struct {
	Port               string
	DatabaseURL        string
	TwitterAPIKey      string
	OpenRouterAPIKey   string
	OpenRouterQAAPIKey string
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/askyourfeed?sslmode=disable"),
		TwitterAPIKey:      getEnv("TWITTER_API_KEY", ""),
		OpenRouterAPIKey:   getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterQAAPIKey: getEnv("OPENROUTER_QA_API_KEY", ""),
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initDatabase initializes database connection with connection pooling
func initDatabase(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		"max_open_conns", 25,
		"max_idle_conns", 5,
		"conn_max_lifetime", "5m")
	return db, nil
}

// getLogLevel returns the log level from environment variable
func getLogLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// maskDatabaseURL masks sensitive parts of the database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking - in production, use a proper URL parser
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-10:]
	}
	return "***"
}

// setupRouter configures the Gin router with routes and middleware
func setupRouter(
	db *sqlx.DB,
	authService services.AuthService,
	authHandler *handlers.AuthHandler,
	qaHandler *handlers.QAHandler,
	ingestHandler *handlers.IngestHandler,
	followingHandler *handlers.FollowingHandler,
) *gin.Engine {
	// Set Gin to release mode for production (can be overridden with GIN_MODE env var)
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", healthCheckHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication endpoints (public - no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout) // Requires auth but handled in handler
		}

		// Session endpoints (protected by auth middleware)
		session := v1.Group("/session")
		session.Use(middleware.AuthMiddleware(authService, db))
		{
			session.GET("/current", authHandler.GetCurrentSession)
		}

		// Q&A endpoints (protected by auth middleware)
		qa := v1.Group("/qa")
		qa.Use(middleware.AuthMiddleware(authService, db))
		{
			qa.POST("", qaHandler.CreateQA)       // Create new Q&A
			qa.GET("", qaHandler.ListQA)          // List Q&A history
			qa.GET("/:id", qaHandler.GetQAByID)   // Get specific Q&A
			qa.DELETE("/:id", qaHandler.DeleteQA) // Delete specific Q&A
			qa.DELETE("", qaHandler.DeleteAllQA)  // Delete all Q&A
		}

		// Ingest endpoints (protected by auth middleware)
		ingest := v1.Group("/ingest")
		ingest.Use(middleware.AuthMiddleware(authService, db))
		{
			ingest.GET("/status", ingestHandler.GetIngestStatus)
			ingest.POST("/trigger", ingestHandler.TriggerIngest)
		}

		// Following endpoints (protected by auth middleware)
		following := v1.Group("/following")
		following.Use(middleware.AuthMiddleware(authService, db))
		{
			following.GET("", followingHandler.GetFollowing) // Get list of followed authors
		}
	}

	return router
}

// healthCheckHandler returns basic health status
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow specific origin for development (Vite dev server)
		// In production, this should be configurable
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5174")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
