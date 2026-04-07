package handler

import (
	"net/http"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// publicSettingKeys lists keys that are safe to expose to all visitors.
var publicSettingKeys = map[string]bool{
	"site_name":    true,
	"logo_url":     true,
	"header_html":  true,
	"footer_html":  true,
	"epay_enabled": true,
}

// GetSettings returns all settings (admin only).
func GetSettings(c *gin.Context) {
	var settings []model.SystemSetting
	if err := db.Engine.Find(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	c.JSON(http.StatusOK, gin.H{"settings": result})
}

// UpdateSettings upserts one or more settings (admin only).
func UpdateSettings(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for key, value := range updates {
		existing := &model.SystemSetting{}
		found, err := db.Engine.Where("key = ?", key).Get(existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if found {
			existing.Value = value
			if _, err := db.Engine.Where("key = ?", key).Cols("value").Update(existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if _, err := db.Engine.Insert(&model.SystemSetting{Key: key, Value: value}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "设置已更新"})
}

// GetPublicSettings returns only the public-facing settings (no admin required).
func GetPublicSettings(c *gin.Context) {
	var settings []model.SystemSetting
	if err := db.Engine.Find(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make(map[string]string)
	for _, s := range settings {
		if publicSettingKeys[s.Key] {
			result[s.Key] = s.Value
		}
	}
	c.JSON(http.StatusOK, gin.H{"settings": result})
}
