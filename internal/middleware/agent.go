package middleware

import (
	"net/http"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// Agent 检查已认证用户是否具有 "agent" 或 "admin" 角色。
// 管理员拥有至少等同客服的所有权限。
func Agent() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, exists := c.Get("role"); exists {
			r, _ := role.(string)
			if r == "agent" || r == "admin" {
				c.Next()
				return
			}
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无访问权限"})
			return
		}

		user := &model.User{}
		found, err := db.Engine.Where("id = ?", userID).Cols("role").Get(user)
		if err != nil || !found || (user.Role != "agent" && user.Role != "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要客服权限"})
			return
		}
		c.Set("role", user.Role)
		c.Next()
	}
}
