package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// VendorHandler 号商相关路由处理器。
type VendorHandler struct {
	cfg *config.ServerConfig
}

func NewVendorHandler(cfg *config.ServerConfig) *VendorHandler {
	return &VendorHandler{cfg: cfg}
}

// Register 号商注册。
//
// @Summary      号商注册
// @Tags         号商
// @Param        body  body  object{username=string,password=string}  true  "注册信息"
// @Success      200   {object}  object{id=int,username=string}
// @Router       /vendor/auth/register [post]
func (h *VendorHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vendor, err := service.RegisterVendor(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": vendor.ID, "username": vendor.Username})
}

// Login 号商登录，返回 JWT。
//
// @Summary      号商登录
// @Tags         号商
// @Param        body  body  object{username=string,password=string}  true  "登录凭证"
// @Success      200   {object}  object{token=string}
// @Router       /vendor/auth/login [post]
func (h *VendorHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, vendor, err := service.LoginVendor(c.Request.Context(), req.Username, req.Password, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "username": vendor.Username, "id": vendor.ID})
}

// GetProfile 查看当前号商信息（余额、邀请码等）。
//
// @Summary      号商个人信息
// @Tags         号商
// @Security     BearerAuth
// @Success      200  {object}  model.Vendor
// @Router       /vendor/profile [get]
func (h *VendorHandler) GetProfile(c *gin.Context) {
	vendorID := c.MustGet("vendor_id").(int64)
	var vendor model.Vendor
	if found, err := db.Engine.ID(vendorID).Get(&vendor); err != nil || !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取信息失败"})
		return
	}
	vendor.PasswordHash = "" // 不返回密码哈希
	c.JSON(http.StatusOK, vendor)
}

// GetPoolKeys 查询绑定到当前号商的所有号池 Key 及其使用统计。
//
// @Summary      号商查看自己的 Key 列表
// @Tags         号商
// @Security     BearerAuth
// @Success      200  {object}  object{keys=[]object}
// @Router       /vendor/keys [get]
func (h *VendorHandler) GetPoolKeys(c *gin.Context) {
	vendorID := c.MustGet("vendor_id").(int64)

	// 查询当前号商的 commission_ratio（用于计算净收益）
	var vendor model.Vendor
	if found, _ := db.Engine.ID(vendorID).Get(&vendor); !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}
	commissionRatio := 0.0
	if vendor.CommissionRatio != nil {
		commissionRatio = *vendor.CommissionRatio
	} else {
		// 使用系统全局手续费比例
		var setting model.SystemSetting
		if found, _ := db.Engine.Where("key = ?", "default_vendor_commission").Get(&setting); found && setting.Value != "" {
			fmt.Sscanf(setting.Value, "%f", &commissionRatio)
		}
	}

	var keys []model.PoolKey
	if err := db.Engine.Where("vendor_id = ?", vendorID).Find(&keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}

	type KeyStat struct {
		ID          int64     `json:"id"`
		PoolID      int64     `json:"pool_id"`
		ChannelName string    `json:"channel_name"`
		MaskedValue string    `json:"masked_value"`
		TotalCost   int64     `json:"total_cost"` // 累计平台进价（credits）
		MyEarn      float64   `json:"my_earn"`    // 号商净收益（credits，已扣手续费）
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}

	result := make([]KeyStat, 0, len(keys))
	for _, k := range keys {
		// 查询关联渠道名称
		channelName := ""
		var pool model.KeyPool
		if found, _ := db.Engine.ID(k.PoolID).Get(&pool); found {
			var ch model.Channel
			if found2, _ := db.Engine.ID(pool.ChannelID).Get(&ch); found2 {
				channelName = ch.Name
			}
		}

		// 查询累计进价成本
		var totalCost int64
		db.Engine.SQL(
			`SELECT COALESCE(SUM(cost),0) FROM billing_transactions WHERE pool_key_id = ? AND type IN ('settle','charge')`,
			k.ID,
		).Get(&totalCost) //nolint:errcheck

		myEarn := float64(totalCost) * (1 - commissionRatio)

		result = append(result, KeyStat{
			ID:          k.ID,
			PoolID:      k.PoolID,
			ChannelName: channelName,
			MaskedValue: maskKeyValue(k.Value),
			TotalCost:   totalCost,
			MyEarn:      myEarn,
			IsActive:    k.IsActive,
			CreatedAt:   k.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"keys": result})
}

// maskKeyValue 将 Key 原文打码，保留首 6 位和末 4 位。
func maskKeyValue(v string) string {
	if len(v) <= 10 {
		return "****"
	}
	return v[:6] + "..." + v[len(v)-4:]
}

// GetSubmittablePools 列出允许号商自助上传 Key 的号池。
//
// @Summary      号商获取可提交号池列表
// @Tags         号商
// @Security     BearerAuth
// @Success      200  {object}  object{pools=[]object}
// @Router       /vendor/pools [get]
func (h *VendorHandler) GetSubmittablePools(c *gin.Context) {
	type PoolInfo struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		ChannelName string `json:"channel_name"`
		ChannelType string `json:"channel_type"`
	}

	var pools []model.KeyPool
	if err := db.Engine.Where("vendor_submittable = true AND is_active = true").Find(&pools); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}

	result := make([]PoolInfo, 0, len(pools))
	for _, pool := range pools {
		var ch model.Channel
		if found, _ := db.Engine.ID(pool.ChannelID).Get(&ch); !found {
			continue
		}
		result = append(result, PoolInfo{
			ID:          pool.ID,
			Name:        pool.Name,
			ChannelName: ch.Name,
			ChannelType: ch.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{"pools": result})
}

// SubmitKey 号商自助上传 Key：先测试 Key 有效性，通过后加入号池。
//
// @Summary      号商提交 Key
// @Tags         号商
// @Security     BearerAuth
// @Param        body  body  object{pool_id=int,value=string}  true  "Key 信息"
// @Success      201   {object}  object{message=string}
// @Router       /vendor/keys [post]
func (h *VendorHandler) SubmitKey(c *gin.Context) {
	vendorID := c.MustGet("vendor_id").(int64)

	var req struct {
		PoolID int64  `json:"pool_id" binding:"required"`
		Value  string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 验证号池存在且开放自助上传
	var pool model.KeyPool
	found, err := db.Engine.Where("id = ? AND vendor_submittable = true AND is_active = true", req.PoolID).Get(&pool)
	if err != nil || !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "号池不存在或不支持自助上传"})
		return
	}

	// 2. 获取关联渠道
	var ch model.Channel
	if found2, _ := db.Engine.ID(pool.ChannelID).Get(&ch); !found2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "关联渠道不存在"})
		return
	}

	// 3. 检查是否已存在（防重复）
	exists, _ := db.Engine.Where("pool_id = ? AND value = ?", req.PoolID, req.Value).Exist(&model.PoolKey{})
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "该 Key 已存在于号池中，请勿重复提交"})
		return
	}

	// 4. 验证 Key 有效性
	if err := testKeyAgainstChannel(c.Request.Context(), &ch, req.Value); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// 5. 写入号池
	key := model.PoolKey{
		PoolID:   req.PoolID,
		VendorID: &vendorID,
		Value:    req.Value,
		Priority: 0,
		IsActive: true,
	}
	if err := service.AddPoolKey(c.Request.Context(), &key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Key 验证通过，已成功加入号池"})
}

// testKeyAgainstChannel 向渠道上游发送最小测试请求验证 Key 是否有效。
// 仅当上游明确返回 401 Unauthorized 或 403 Forbidden 时认定 Key 无效；
// 其他状态码（400 参数错误、5xx 服务错误、200 正常）均视为 Key 本身可用。
func testKeyAgainstChannel(ctx context.Context, ch *model.Channel, keyValue string) error {
	var reqBody io.Reader
	method := "POST"

	if ch.Type == "llm" {
		payload := map[string]interface{}{
			"model":      ch.Model,
			"messages":   []map[string]string{{"role": "user", "content": "hi"}},
			"max_tokens": 1,
		}
		b, _ := json.Marshal(payload)
		reqBody = bytes.NewReader(b)
	} else {
		// 非 LLM 渠道：GET 探测连通性
		method = "GET"
		reqBody = http.NoBody
	}

	testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(testCtx, method, ch.BaseURL, reqBody)
	if err != nil {
		// 构建请求失败，忽略测试
		return nil
	}
	if ch.Type == "llm" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Authorization", "Bearer "+keyValue)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		// 网络错误无法判断 Key 有效性，允许通过（可能是短暂网络问题）
		return nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("Key 无效或已过期（上游返回 401 Unauthorized）")
	case http.StatusForbidden:
		return fmt.Errorf("Key 权限不足（上游返回 403 Forbidden）")
	}
	return nil
}

// ---- 管理员接口 ----

// AdminListVendors 列出所有号商（管理员）。
//
// @Summary      管理员列出号商
// @Tags         管理-号商
// @Security     BearerAuth
// @Success      200  {object}  object{vendors=[]model.Vendor}
// @Router       /admin/vendors [get]
func AdminListVendors(c *gin.Context) {
	var vendors []model.Vendor
	if err := db.Engine.OrderBy("id DESC").Find(&vendors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}
	for i := range vendors {
		vendors[i].PasswordHash = ""
	}
	c.JSON(http.StatusOK, gin.H{"vendors": vendors})
}

// AdminUpdateVendor 更新号商信息（is_active、commission_ratio）。
//
// @Summary      管理员更新号商
// @Tags         管理-号商
// @Security     BearerAuth
// @Param        id    path  int  true  "号商 ID"
// @Param        body  body  object  true  "更新字段"
// @Success      200   {object}  object{message=string}
// @Router       /admin/vendors/:id [patch]
func AdminUpdateVendor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var req struct {
		IsActive        *bool    `json:"is_active"`
		CommissionRatio *float64 `json:"commission_ratio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var vendor model.Vendor
	if found, _ := db.Engine.ID(id).Get(&vendor); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "号商不存在"})
		return
	}

	cols := []string{}
	if req.IsActive != nil {
		vendor.IsActive = *req.IsActive
		cols = append(cols, "is_active")
	}
	if req.CommissionRatio != nil {
		vendor.CommissionRatio = req.CommissionRatio
		cols = append(cols, "commission_ratio")
	}
	if len(cols) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
		return
	}

	if _, err := db.Engine.ID(id).Cols(cols...).Update(&vendor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// AdminSetPoolKeyVendor 将号池 Key 与号商关联。
//
// @Summary      管理员绑定号池 Key 到号商
// @Tags         管理-号商
// @Security     BearerAuth
// @Param        id        path  int  true   "号池 Key ID"
// @Param        body      body  object{vendor_id=int}  true  "号商 ID（0 解绑）"
// @Success      200       {object}  object{message=string}
// @Router       /admin/pool-keys/:id/vendor [patch]
func AdminSetPoolKeyVendor(c *gin.Context) {
	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 Key ID"})
		return
	}

	var req struct {
		VendorID *int64 `json:"vendor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果 vendor_id 非零，验证号商存在
	if req.VendorID != nil && *req.VendorID != 0 {
		count, _ := db.Engine.ID(*req.VendorID).Count(&model.Vendor{})
		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "号商不存在"})
			return
		}
	}

	vendorID := req.VendorID
	if vendorID != nil && *vendorID == 0 {
		vendorID = nil
	}

	pk := &model.PoolKey{VendorID: vendorID}
	if _, err := db.Engine.ID(keyID).Cols("vendor_id").Update(pk); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "绑定成功"})
}
