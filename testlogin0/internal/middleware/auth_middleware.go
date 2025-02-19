package middleware

import (
	"net/http"
	"strings"

	"login/internal/config"
	"login/pkg/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := utils.VerifyToken(parts[1], cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.Subject)
		c.Next()
	}
}

func APIKeyMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("API-Key")

		// log.Printf("apiKey = %v, cfg.APIKey = %v", apiKey, cfg.APIKey)

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is missing"})
			return
		}
		if apiKey != cfg.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}
		c.Next()
	}
}
