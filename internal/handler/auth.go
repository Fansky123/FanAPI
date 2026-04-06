package handler

import (
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"
	"fanapi/pkg/mailer"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg    *config.ServerConfig
	mailer *mailer.Mailer
}

func NewAuthHandler(cfg *config.ServerConfig, m *mailer.Mailer) *AuthHandler {
	return &AuthHandler{cfg: cfg, mailer: m}
}

// POST /auth/send-code  — 公用：注册绑定邮箱 / 找回密码前发验证码
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.SendVerifyCode(c.Request.Context(), req.Email, h.mailer); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "code sent"})
}

// POST /auth/register — 仅需用户名 + 密码，无需邮箱验证
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	// 注册后自动登录
	token, _, tokenErr := service.Login(c.Request.Context(), req.Username, req.Password, h.cfg)
	if tokenErr != nil {
		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": gin.H{"id": user.ID, "username": user.Username, "role": user.Role}})
}

// POST /auth/login — 用户名或邮箱 + 密码
// 接受 {username, password} 或 {email, password}，兼容两种调用方
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usernameOrEmail := req.Username
	if usernameOrEmail == "" {
		usernameOrEmail = req.Email
	}
	if usernameOrEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email required"})
		return
	}
	token, user, err := service.Login(c.Request.Context(), usernameOrEmail, req.Password, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email, "role": user.Role}})
}

// GET /user/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	user := &model.User{}
	found, err := db.Engine.ID(userID).Get(user)
	if err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"group":    user.Group,
	})
}

// POST /user/bind-email — 登录后绑定邮箱（需先调 /auth/send-code 获取验证码）
func (h *AuthHandler) BindEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(int64)
	if err := service.BindEmail(c.Request.Context(), userID, req.Email, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "email bound"})
}

// POST /auth/forgot-password — 向已绑定邮箱发送重置验证码（邮箱不存在时静默成功，防枚举）
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = service.SendPasswordResetCode(c.Request.Context(), req.Email, h.mailer)
	c.JSON(http.StatusOK, gin.H{"message": "if this email is bound to an account, a reset code has been sent"})
}

// POST /auth/reset-password — 通过邮箱验证码重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ResetPasswordByEmail(c.Request.Context(), req.Email, req.Code, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset"})
}

// POST /user/apikeys  (requires auth)
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(int64)
	rawKey, err := service.GenerateAPIKey(c.Request.Context(), userID, req.Name, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"key": rawKey, "note": "store this key safely, it will not be shown again"})
}

// GET /user/balance
func (h *AuthHandler) GetBalance(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	balance, err := service.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"balance_credits": balance,
		"balance_cny":     float64(balance) / 1_000_000,
	})
}

// GET /user/transactions
func (h *AuthHandler) GetTransactions(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	page := 1
	size := 20
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if s := c.Query("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			size = n
		}
	}
	txs, err := service.ListTransactions(c.Request.Context(), userID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	total, _ := service.CountTransactions(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"transactions": txs, "total": total})
}

// GET /public/channels  — 公开（无需登录），显示默认价格
// GET /user/channels    — 登录用户，按其 group 显示专属价格
func (h *AuthHandler) ListModels(c *gin.Context) {
	var channels []model.Channel
	if err := db.Engine.Where("is_active = true").
		Cols("id", "name", "model", "type", "protocol", "billing_type", "billing_config").
		Find(&channels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 已登录时从 context 取用户分组，用于展示专属价格
	userGroup := ""
	if raw, ok := c.Get("user_group"); ok {
		userGroup, _ = raw.(string)
	}

	type channelInfo struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		RoutingModel string `json:"routing_model"`
		Model        string `json:"model"`
		Type         string `json:"type"`
		Protocol     string `json:"protocol"`
		BillingType  string `json:"billing_type"`
		PriceDisplay string `json:"price_display"`         // 默认价格
		GroupPrice   string `json:"group_price,omitempty"` // 用户专属价格（与默认不同时才返回）
	}

	result := make([]channelInfo, 0, len(channels))
	for _, ch := range channels {
		defaultPrice := buildPriceDisplay(ch.BillingType, ch.BillingConfig)
		groupPrice := ""
		if userGroup != "" {
			groupCfg := applyGroupPricingMap(map[string]interface{}(ch.BillingConfig), userGroup)
			gp := buildPriceDisplay(ch.BillingType, groupCfg)
			if gp != defaultPrice {
				groupPrice = gp
			}
		}
		result = append(result, channelInfo{
			ID:           ch.ID,
			Name:         ch.Name,
			RoutingModel: ch.Name,
			Model:        ch.Model,
			Type:         ch.Type,
			Protocol:     ch.Protocol,
			BillingType:  ch.BillingType,
			PriceDisplay: defaultPrice,
			GroupPrice:   groupPrice,
		})
	}
	c.JSON(http.StatusOK, gin.H{"channels": result})
}

// applyGroupPricingMap 与 billing.applyGroupPricing 逻辑相同，此处避免包循环依赖而内联。
func applyGroupPricingMap(cfg map[string]interface{}, group string) model.JSON {
	if group == "" || cfg == nil {
		return model.JSON(cfg)
	}
	pgs, ok := cfg["pricing_groups"].(map[string]interface{})
	if !ok {
		return model.JSON(cfg)
	}
	overrides, ok := pgs[group].(map[string]interface{})
	if !ok {
		return model.JSON(cfg)
	}
	merged := make(map[string]interface{}, len(cfg))
	for k, v := range cfg {
		merged[k] = v
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return model.JSON(merged)
}

// buildPriceDisplay 根据计费类型和配置生成人类可读的价格描述字符串。
// credits 换算：1 CNY = 1,000,000 credits。
func buildPriceDisplay(billingType string, cfg model.JSON) string {
	if cfg == nil {
		return ""
	}
	toF := func(key string) float64 {
		v, ok := cfg[key]
		if !ok {
			return 0
		}
		switch n := v.(type) {
		case float64:
			return n
		case int64:
			return float64(n)
		}
		return 0
	}
	switch billingType {
	case "token":
		in := toF("input_price_per_1m_tokens") / 1000000 // credits → ¥
		out := toF("output_price_per_1m_tokens") / 1000000
		if in > 0 && out > 0 {
			return fmt.Sprintf("¥%.4f / 1M 输入 + ¥%.4f / 1M 输出", in, out)
		}
	case "image":
		base := toF("base_price") / 1000000
		if base > 0 {
			return fmt.Sprintf("¥%.4f / 张起", base)
		}
	case "video":
		perSec := toF("price_per_second") / 1000000
		if perSec > 0 {
			return fmt.Sprintf("¥%.4f / 秒", perSec)
		}
	case "audio":
		perSec := toF("price_per_second") / 1000000
		if perSec > 0 {
			return fmt.Sprintf("¥%.4f / 秒", perSec)
		}
	case "count":
		p := toF("price_per_call") / 1000000
		if p > 0 {
			return fmt.Sprintf("¥%.4f / 次", p)
		}
	}
	return ""
}

// GET /user/apikeys
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	var keys []model.APIKey
	if err := db.Engine.Where("user_id = ?", userID).
		Cols("id", "name", "key_hash", "raw_key_enc", "is_active", "last_used_at", "created_at").
		Desc("id").
		Find(&keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type apiKeyItem struct {
		ID         int64       `json:"id"`
		Name       string      `json:"name"`
		KeyPrefix  string      `json:"key_prefix"`
		RawKey     string      `json:"raw_key"`
		Viewable   bool        `json:"viewable"`
		IsActive   bool        `json:"is_active"`
		LastUsedAt interface{} `json:"last_used_at"`
		CreatedAt  interface{} `json:"created_at"`
	}

	items := make([]apiKeyItem, 0, len(keys))
	for _, k := range keys {
		rawKey := ""
		viewable := false
		if k.RawKeyEnc != "" {
			if decrypted, err := service.DecryptAPIKey(k.RawKeyEnc, h.cfg.JWTSecret); err == nil {
				rawKey = decrypted
				viewable = true
			}
		}
		prefix := ""
		if len(k.KeyHash) >= 12 {
			prefix = k.KeyHash[:12]
		} else {
			prefix = k.KeyHash
		}
		items = append(items, apiKeyItem{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  prefix,
			RawKey:     rawKey,
			Viewable:   viewable,
			IsActive:   k.IsActive,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": items})
}

// PUT /user/password — 当前登录用户修改自己的密码（需提供旧密码）
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &model.User{}
	found, err := db.Engine.ID(userID).Get(user)
	if err != nil || !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码不正确"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}
	db.Engine.ID(userID).Cols("password_hash").Update(&model.User{PasswordHash: string(hash)})
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

// DELETE /user/apikeys/:id
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	keyID := c.Param("id")
	// Sanitize keyID
	keyID = strings.TrimSpace(keyID)
	affected, err := db.Engine.Where("id = ? AND user_id = ?", keyID, userID).
		Cols("is_active").
		Update(&model.APIKey{IsActive: false})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "api key revoked"})
}
