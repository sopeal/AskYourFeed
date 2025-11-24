package main

import (
	"context"
	"fmt"
	"log"
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
)

func main() {
	// Load configuration from environment variables
	config := loadConfig()

	// Initialize database connection
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
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

	// Initialize services
	llmService := services.NewLLMService()
	qaService := services.NewQAService(db, postRepo, qaRepo, llmService)
	ingestStatusService := services.NewIngestStatusService(ingestRepo)
	followingService := services.NewFollowingService(followingRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, *twitterClient)

	// Initialize ingestion service
	ingestService := services.NewIngestService(
		twitterClient,
		ingestRepo,
		followingRepo,
		postRepo,
		authorRepo,
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
		log.Printf("Starting server on port %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Config holds application configuration
type Config struct {
	Port          string
	DatabaseURL   string
	TwitterAPIKey string
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/askyourfeed?sslmode=disable"),
		TwitterAPIKey: getEnv("TWITTER_API_KEY", ""),
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

	log.Println("Database connection established")
	return db, nil
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
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
