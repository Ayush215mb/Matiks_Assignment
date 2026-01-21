package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/handlers"
	"backend/internal/services"
	"backend/pkg/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize in-memory store
	memoryStore := store.NewMemoryStore()
	log.Println("✓ Initialized in-memory store")

	// Initialize services
	leaderboardService := services.NewLeaderboardService(memoryStore)

	// Initialize handlers
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)

	// Set up Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"store":  "in-memory",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Seed data
		api.POST("/seed", leaderboardHandler.SeedData)

		// Leaderboard
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// User operations
		api.GET("/users/:username", leaderboardHandler.GetUserRank)
		api.POST("/users/:username/score", leaderboardHandler.UpdateScore)

		// Search
		api.GET("/search", leaderboardHandler.SearchUser)

		// Stats
		api.GET("/stats", leaderboardHandler.GetStats)
	}

	// Start random score update simulation
	ctx := context.Background()
	go leaderboardService.StartRandomUpdates(ctx)

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
