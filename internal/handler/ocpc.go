package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// ── 平台账户 CRUD ──────────────────────────────────────────────────────────────

// ListOcpcPlatforms 列出所有广告平台账户。GET /admin/ocpc/platforms
func ListOcpcPlatforms(c *gin.Context) {
	var list []model.OcpcPlatform
	if err := db.Engine.OrderBy("id ASC").Find(&list); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 脱敏：隐藏 secret 字段（只保留前6位）
	for i := range list {
		if len(list[i].E360Secret) > 6 {
			list[i].E360Secret = list[i].E360Secret[:6] + "******"
		}
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// CreateOcpcPlatform 新增广告平台账户。POST /admin/ocpc/platforms
func CreateOcpcPlatform(c *gin.Context) {
	var p model.OcpcPlatform
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p.Platform != "baidu" && p.Platform != "360" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform 只能是 baidu 或 360"})
		return
	}
	p.ID = 0
	if _, err := db.Engine.Insert(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// UpdateOcpcPlatform 更新广告平台账户。PUT /admin/ocpc/platforms/:id
func UpdateOcpcPlatform(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var p model.OcpcPlatform
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p.ID = id
	// 若 secret 末尾含 *** 则说明前端未修改，从 DB 取原值
	if len(p.E360Secret) > 0 && len(p.E360Secret) < 12 && p.E360Secret[len(p.E360Secret)-6:] == "******" {
		orig := &model.OcpcPlatform{}
		if found, _ := db.Engine.ID(id).Get(orig); found {
			p.E360Secret = orig.E360Secret
		}
	}
	cols := []string{"platform", "name", "enabled", "baidu_token", "baidu_page_url",
		"baidu_reg_type", "baidu_order_type",
		"e360_key", "e360_secret", "e360_jzqs", "e360_so_type",
		"e360_reg_event", "e360_order_event"}
	if _, err := db.Engine.ID(id).Cols(cols...).Update(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

// DeleteOcpcPlatform 删除广告平台账户。DELETE /admin/ocpc/platforms/:id
func DeleteOcpcPlatform(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := db.Engine.ID(id).Delete(&model.OcpcPlatform{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ToggleOcpcPlatform 切换启用/禁用状态。PATCH /admin/ocpc/platforms/:id/toggle
func ToggleOcpcPlatform(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p := &model.OcpcPlatform{}
	if found, _ := db.Engine.ID(id).Get(p); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	p.Enabled = !p.Enabled
	if _, err := db.Engine.ID(id).Cols("enabled").Update(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": p.Enabled})
}

// ── OCPC 上报触发 ──────────────────────────────────────────────────────────────

// TriggerOcpcUpload 触发 OCPC 转化数据上报（管理员接口）。
// POST /admin/ocpc/upload
func TriggerOcpcUpload(c *gin.Context) {
	regOK, regFail, orderOK, orderFail := service.UploadOcpcConversions(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"message":    "上报完成",
		"reg_ok":     regOK,
		"reg_fail":   regFail,
		"order_ok":   orderOK,
		"order_fail": orderFail,
	})
}

// GetOcpcSchedule GET /admin/ocpc/schedule — 读取定时上报配置
func GetOcpcSchedule(c *gin.Context) {
	keys := []string{"ocpc_schedule_enabled", "ocpc_schedule_interval", "ocpc_last_run_at"}
	result := map[string]string{}
	for _, k := range keys {
		s := &model.SystemSetting{}
		if found, _ := db.Engine.Where("key = ?", k).Get(s); found {
			result[k] = s.Value
		}
	}
	c.JSON(http.StatusOK, gin.H{"schedule": result})
}

// UpdateOcpcSchedule PUT /admin/ocpc/schedule — 更新定时上报配置
func UpdateOcpcSchedule(c *gin.Context) {
	var req struct {
		Enabled  *bool `json:"enabled"`
		Interval *int  `json:"interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Interval != nil && (*req.Interval < 5 || *req.Interval > 1440) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval 需在 5~1440 分钟之间"})
		return
	}
	if req.Enabled != nil {
		v := "false"
		if *req.Enabled {
			v = "true"
		}
		upsertScheduleSetting("ocpc_schedule_enabled", v)
	}
	if req.Interval != nil {
		upsertScheduleSetting("ocpc_schedule_interval", strconv.Itoa(*req.Interval))
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

func upsertScheduleSetting(key, value string) {
	s := &model.SystemSetting{}
	found, _ := db.Engine.Where("key = ?", key).Get(s)
	if found {
		db.Engine.Where("key = ?", key).Cols("value").Update(&model.SystemSetting{Value: value}) //nolint:errcheck
	} else {
		db.Engine.Insert(&model.SystemSetting{Key: key, Value: value}) //nolint:errcheck
	}
}
