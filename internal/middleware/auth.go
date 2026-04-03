package middleware

import (
	"net/http"
	"strings"

	"fanapi/internal/config"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth supports both X-API-Key header and Authorization: Bearer JWT.
func Auth(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API Key first
		rawKey := c.GetHeader("X-API-Key")
		if rawKey != "" {
			apiKey, err := service.LookupAPIKey(c.Request.Context(), rawKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return
			}
			c.Set("user_id", apiKey.UserID)
			c.Set("api_key_id", apiKey.ID)
			c.Set("auth_type", "apikey")
			c.Next()
			return
		}

		// Try JWT Bearer
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				return
			}
			userID := int64(claims["sub"].(float64))
			role, _ := claims["role"].(string)
			c.Set("user_id", userID)
			c.Set("role", role)
			c.Set("auth_type", "jwt")
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
	}
}

// APIKeyOnly rejects requests that are not authenticated via API Key.
func APIKeyOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authType, _ := c.Get("auth_type"); authType != "apikey" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "api key required"})
			return
		}
		c.Next()
	}
}
