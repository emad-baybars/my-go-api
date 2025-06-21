package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-backend-template/handlers"
	"go-backend-template/middleware"
	"go-backend-template/utils"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	healthHandler *handlers.HealthHandler,
	logger utils.Logger,
) {
	// Add rate limiting and timeout middleware
	router.Use(middleware.RateLimiter())
	router.Use(middleware.Timeout(30 * time.Second))

	// API version 1 group
	v1 := router.Group("/api/v1")

	// Public routes
	{
		// Health check
		v1.GET("/health", healthHandler.HealthCheck)

		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	// Protected routes (require authentication)
	{
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth())

		// User routes
		users := protected.Group("/users")
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)

			// Admin only routes
			adminUsers := users.Group("/")
			adminUsers.Use(middleware.RequireRole("admin", "superadmin"))
			{
				adminUsers.GET("", userHandler.GetUsers)
			}
		}
	}

	// API documentation route (development only)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	logger.Info("Routes configured successfully")
}
