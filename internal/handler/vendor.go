package handler

import (
	"net/http"
	"strconv"

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

	var keys []model.PoolKey
	if err := db.Engine.Where("vendor_id = ?", vendorID).Find(&keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}

	// 对每个 Key 查询累计消耗（按 pool_key_id 汇总 billing_transactions）
	type KeyStat struct {
		model.PoolKey
		TotalCredits int64 `json:"total_credits"`
		TotalCost    int64 `json:"total_cost"`
	}

	result := make([]KeyStat, 0, len(keys))
	for _, k := range keys {
		k.Value = "" // 不暴露 Key 原值
		var totalCredits, totalCost int64
		db.Engine.SQL(
			`SELECT COALESCE(SUM(credits),0), COALESCE(SUM(cost),0) FROM billing_transactions WHERE pool_key_id = ? AND type IN ('settle','charge')`,
			k.ID,
		).Get(&totalCredits, &totalCost) //nolint:errcheck
		result = append(result, KeyStat{PoolKey: k, TotalCredits: totalCredits, TotalCost: totalCost})
	}

	c.JSON(http.StatusOK, gin.H{"keys": result})
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
