package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// ListKeyPools GET /admin/key-pools?channel_id=xxx
func ListKeyPools(c *gin.Context) {
	channelIDStr := c.Query("channel_id")
	if channelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 channel_id"})
		return
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 格式错误"})
		return
	}
	pools, err := service.ListKeyPools(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pools)
}

// CreateKeyPool POST /admin/key-pools
func CreateKeyPool(c *gin.Context) {
	var pool model.KeyPool
	if err := c.ShouldBindJSON(&pool); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if pool.ChannelID == 0 || pool.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 channel_id 和号池名称"})
		return
	}
	if err := service.CreateKeyPool(c.Request.Context(), &pool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pool)
}

// DeleteKeyPool DELETE /admin/key-pools/:id
func DeleteKeyPool(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	if err := service.DeleteKeyPool(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListPoolKeys GET /admin/key-pools/:id/keys
func ListPoolKeys(c *gin.Context) {
	poolID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "号池 ID 格式错误"})
		return
	}
	keys, err := service.ListPoolKeys(c.Request.Context(), poolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// AddPoolKey POST /admin/key-pools/:id/keys
func AddPoolKey(c *gin.Context) {
	poolID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "号池 ID 格式错误"})
		return
	}
	var key model.PoolKey
	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if key.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 Key 值"})
		return
	}
	key.PoolID = poolID
	if err := service.AddPoolKey(c.Request.Context(), &key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, key)
}

// RemovePoolKey DELETE /admin/pool-keys/:id
func RemovePoolKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	if err := service.RemovePoolKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
