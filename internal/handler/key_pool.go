package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// ListKeyPools GET /admin/key-pools?channel_id=xxx  (channel_id 可选，不传则返回全部号池)
func ListKeyPools(c *gin.Context) {
	var channelID int64
	if s := c.Query("channel_id"); s != "" {
		var err error
		channelID, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 格式错误"})
			return
		}
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
	pool.IsActive = true
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
	key.IsActive = true
	if err := service.AddPoolKey(c.Request.Context(), &key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, key)
}

// ToggleKeyPool PATCH /admin/key-pools/:id/toggle
func ToggleKeyPool(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	if err := service.ToggleKeyPool(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
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

// ToggleVendorSubmittable PATCH /admin/key-pools/:id/vendor-toggle
// 切换号池的"允许号商自助上传 Key"开关。
func ToggleVendorSubmittable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	var pool model.KeyPool
	if found, _ := db.Engine.ID(id).Get(&pool); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "号池不存在"})
		return
	}
	pool.VendorSubmittable = !pool.VendorSubmittable
	if _, err := db.Engine.ID(id).Cols("vendor_submittable").Update(&pool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vendor_submittable": pool.VendorSubmittable})
}
