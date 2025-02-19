package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"login/internal/config"
	"login/internal/handler"
	"login/internal/middleware"
	"login/internal/repository"
	"login/internal/service"
	"login/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Periodic database connection check
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := db.Ping(); err != nil {
				log.Printf("Database connection lost: %v", err)
				// Try to reconnect
				if reconnErr := db.Reconnect(); reconnErr != nil {
					log.Printf("Failed to reconnect: %v", reconnErr)
				} else {
					log.Printf("Successfully reconnected to the database")
				}
			}
		}
	}()
	userRepo := repository.NewUserRepository(db.DB)
	// userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS configuration
	configCors := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:4000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(configCors))
	r.Use(TimeoutMiddleware(5 * time.Second))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.GET("/client-id", middleware.APIKeyMiddleware(cfg), authHandler.GetClientID)
			auth.POST("/google/verify", authHandler.VerifyGoogleToken)
			auth.POST("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)
		}
		users := v1.Group("/users")
		{
			users.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetCurrentUser)
		}
	}

	log.Printf("Server running on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
