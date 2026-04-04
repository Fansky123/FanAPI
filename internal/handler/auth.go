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

// POST /auth/send-code
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

// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.VerifyEmailCode(c.Request.Context(), req.Email, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, user, err := service.Login(c.Request.Context(), req.Email, req.Password, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": user.ID, "email": user.Email, "role": user.Role}})
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

// GET /user/channels — 返回所有启用渠道的公开信息（含简化价格），
// 用户据此选择渠道并决定传哪个 channel_id。
// 同一模型可有多个渠道（如 nano-1001 / nano-1002），价格不同，用户按预算选择。
func (h *AuthHandler) ListModels(c *gin.Context) {
	var channels []model.Channel
	if err := db.Engine.Where("is_active = true").
		Cols("id", "name", "model", "type", "protocol", "billing_type", "billing_config").
		Find(&channels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// channelInfo 是暴露给用户的渠道公开信息，隐藏脚本/密钥/上游地址。
	type channelInfo struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Model        string `json:"model"`
		Type         string `json:"type"`
		Protocol     string `json:"protocol"`
		BillingType  string `json:"billing_type"`
		PriceDisplay string `json:"price_display"`
	}

	result := make([]channelInfo, 0, len(channels))
	for _, ch := range channels {
		result = append(result, channelInfo{
			ID:           ch.ID,
			Name:         ch.Name,
			Model:        ch.Model,
			Type:         ch.Type,
			Protocol:     ch.Protocol,
			BillingType:  ch.BillingType,
			PriceDisplay: buildPriceDisplay(ch.BillingType, ch.BillingConfig),
		})
	}
	c.JSON(http.StatusOK, gin.H{"channels": result})
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
