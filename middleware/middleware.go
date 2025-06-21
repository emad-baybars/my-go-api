package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-backend-template/config"
	"go-backend-template/jwt"
	"go-backend-template/models"
	"go-backend-template/utils"
)

// Logger middleware for request logging
func Logger(logger utils.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// Recovery middleware for panic recovery
func Recovery(logger utils.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered", "error", recovered)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error:   "Something went wrong",
		})
	})
}

// CORS middleware for cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Localization middleware for language support
func Localization(localizer *utils.Localizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get language from header, query parameter, or use default
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = c.Query("lang")
		}
		if lang == "" {
			lang = localizer.DefaultLanguage
		}

		// Extract the primary language code
		if strings.Contains(lang, ",") {
			lang = strings.Split(lang, ",")[0]
		}
		if strings.Contains(lang, "-") {
			lang = strings.Split(lang, "-")[0]
		}

		c.Set("language", lang)
		c.Set("localizer", localizer)
		c.Next()
	}
}

// RequireRole middleware for role-based authorization
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found",
				Error:   "Authorization failed",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Insufficient permissions",
			Error:   "You don't have permission to access this resource",
		})
		c.Abort()
	}
}

// RateLimiter middleware for rate limiting (simple in-memory implementation)
func RateLimiter() gin.HandlerFunc {
	type client struct {
		requests  int
		resetTime time.Time
	}

	clients := make(map[string]*client)
	limit := 100 // requests per minute
	window := time.Minute

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		if clients[clientIP] == nil {
			clients[clientIP] = &client{
				requests:  1,
				resetTime: now.Add(window),
			}
			c.Next()
			return
		}

		if now.After(clients[clientIP].resetTime) {
			clients[clientIP].requests = 1
			clients[clientIP].resetTime = now.Add(window)
			c.Next()
			return
		}

		if clients[clientIP].requests >= limit {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "Rate limit exceeded",
				Error:   "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		clients[clientIP].requests++
		c.Next()
	}
}

// Timeout middleware adds timeout to requests
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, models.APIResponse{
				Success: false,
				Message: "Request timeout",
				Error:   "Request took too long to process",
			})
			return
		}
	}
}

// JWTAuth middleware for JWT authentication
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header required",
				Error:   "No authorization header provided",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid authorization header format",
				Error:   "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Parse and validate JWT token
		cfg := config.Load()
		claims, err := jwt.ValidateToken(cfg.JWTSecret, tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token",
				Error:   "Authentication failed",
			})
			c.Abort()
			return
		}

		// Extract claims and set in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}
