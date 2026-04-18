package middleware

import (
	"net/http"
	"strings"

	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// VendorAuth verifies JWT tokens issued to vendor accounts.
func VendorAuth(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "需要登录"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["role"] != "vendor" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}

		sub, _ := claims["sub"].(float64)
		vendorID := int64(sub)

		var vendor model.Vendor
		if found, _ := db.Engine.ID(vendorID).Get(&vendor); !found || !vendor.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "账号不存在或已被禁用"})
			return
		}

		c.Set("vendor_id", vendorID)
		c.Next()
	}
}
