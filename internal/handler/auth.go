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
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// POST /auth/register — 仅需用户名 + 密码，无需邮箱验证
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username   string `json:"username" binding:"required,min=3,max=32"`
		Password   string `json:"password" binding:"required,min=8"`
		InviteCode string `json:"invite_code"` // 邀请码（可选）
		// 广告追踪参数（可选，用于 OCPC 转化上报）
		PlatformID int64  `json:"platform_id"` // ocpc_platforms.id（落地页 URL 中的 ocpc_id）
		BdVid      string `json:"bd_vid"`
		QhClickID  string `json:"qh_click_id"`
		SourceID   string `json:"source_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 解析邀请人
	var inviterID *int64
	var inviterQR string
	if req.InviteCode != "" {
		inviter := &model.User{}
		if found, _ := db.Engine.Where("invite_code = ?", req.InviteCode).Cols("id", "wechat_qr").Get(inviter); found {
			inviterID = &inviter.ID
			inviterQR = inviter.WechatQR
		}
	}
	user, err := service.Register(c.Request.Context(), req.Username, req.Password, inviterID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// 记录广告追踪参数（用于 OCPC 转化上报）
	ip := clientIP(c)
	ua := c.GetHeader("User-Agent")
	service.CreateOrUpdateOcpcRecord(c.Request.Context(), user.ID, req.PlatformID, req.BdVid, req.QhClickID, req.SourceID, ip, ua)

	// 注册后自动登录
	token, _, tokenErr := service.Login(c.Request.Context(), req.Username, req.Password, h.cfg)
	if tokenErr != nil {
		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username})
		return
	}
	resp := gin.H{"token": token, "user": gin.H{"id": user.ID, "username": user.Username, "role": user.Role}}
	if inviterQR != "" {
		resp["inviter_wechat_qr"] = inviterQR
	}
	c.JSON(http.StatusCreated, resp)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名或邮筱"})
		return
	}
	token, user, err := service.Login(c.Request.Context(), usernameOrEmail, req.Password, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// 如果用户是被客服邀请的，返回该客服的微信二维码
	var inviterQR string
	if user.InviterID != nil {
		inviter := &model.User{}
		if found, _ := db.Engine.ID(*user.InviterID).Cols("wechat_qr").Get(inviter); found {
			inviterQR = inviter.WechatQR
		}
	}
	resp := gin.H{"token": token, "user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email, "role": user.Role}}
	if inviterQR != "" {
		resp["inviter_wechat_qr"] = inviterQR
	}
	c.JSON(http.StatusOK, resp)
}

// GET /user/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	user := &model.User{}
	found, err := db.Engine.ID(userID).Cols("id", "username", "email", "role", "group").Get(user)
	if err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
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
	c.JSON(http.StatusOK, gin.H{"message": "邮箱绑定成功"})
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
	c.JSON(http.StatusOK, gin.H{"message": "如果该邮箱已绑定账号，重置验证码将发送至您的邮筱"})
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
	c.JSON(http.StatusOK, gin.H{"message": "密码已重置，请使用新密码登录"})
}

// POST /user/apikeys  (requires auth)
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		KeyType string `json:"key_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.KeyType != "stable" {
		req.KeyType = "low_price"
	}
	userID := c.MustGet("user_id").(int64)
	rawKey, err := service.GenerateAPIKey(c.Request.Context(), userID, req.Name, req.KeyType, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"key": rawKey, "note": "store this key safely, it will not be shown again"})
}

// GetBalance 查询余额
// @Summary      查询账户余额
// @Description  返回当前 API Key 对应账户的剩余余额，1 CNY = 1,000,000 credits。
// @Tags         用户
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  object{balance_credits=int,balance_cny=number}
// @Router       /user/balance [get]
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
	corrID := c.Query("corr_id")
	taskID := c.Query("task_id")
	txs, err := service.ListTransactions(c.Request.Context(), userID, page, size, corrID, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	total, _ := service.CountTransactions(c.Request.Context(), userID, corrID, taskID)
	c.JSON(http.StatusOK, gin.H{"transactions": txs, "total": total})
}

// ListModels 获取可用渠道列表
// @Summary      获取渠道列表并查询价格
// @Description  登录用户可看到其分组专属价（group_price）；请将 routing_model 填入请求的 model 字段进行加载均衡路由。
// @Tags         用户
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  object{channels=[]object}
// @Router       /user/channels [get]
func (h *AuthHandler) ListModels(c *gin.Context) {
	var channels []model.Channel
	if err := db.Engine.Where("is_active = true").
		Cols("id", "name", "model", "type", "protocol", "billing_type", "billing_config", "icon_url", "description").
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
		Type         string `json:"type"`
		Protocol     string `json:"protocol"`
		BillingType  string `json:"billing_type"`
		PriceDisplay string `json:"price_display"`         // 默认价格
		GroupPrice   string `json:"group_price,omitempty"` // 用户专属价格（与默认不同时才返回）
		IconURL      string `json:"icon_url"`
		Description  string `json:"description"`
	}

	// 按路由键（ch.Model）去重，同一路由键的多个渠道只展示第一个作为代表
	seen := make(map[string]bool)
	result := make([]channelInfo, 0, len(channels))
	for _, ch := range channels {
		if seen[ch.Model] {
			continue
		}
		seen[ch.Model] = true

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
			RoutingModel: ch.Model,
			Type:         ch.Type,
			Protocol:     ch.Protocol,
			BillingType:  ch.BillingType,
			PriceDisplay: defaultPrice,
			GroupPrice:   groupPrice,
			IconURL:      ch.IconURL,
			Description:  ch.Description,
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
			base := fmt.Sprintf("¥%.4f / 1M 输入 + ¥%.4f / 1M 输出", in, out)
			cacheCreate := toF("cache_creation_price_per_1m_tokens") / 1000000
			cacheRead := toF("cache_read_price_per_1m_tokens") / 1000000
			if cacheCreate > 0 || cacheRead > 0 {
				cacheStr := ""
				if cacheCreate > 0 && cacheRead > 0 {
					cacheStr = fmt.Sprintf("缓存写入 ¥%.4f + 缓存读取 ¥%.4f / 1M", cacheCreate, cacheRead)
				} else if cacheCreate > 0 {
					cacheStr = fmt.Sprintf("缓存写入 ¥%.4f / 1M", cacheCreate)
				} else {
					cacheStr = fmt.Sprintf("缓存读取 ¥%.4f / 1M", cacheRead)
				}
				return base + "\n" + cacheStr
			}
			return base
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
		Cols("id", "name", "key_hash", "raw_key_enc", "key_type", "is_active", "last_used_at", "created_at").
		Desc("id").
		Find(&keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type apiKeyItem struct {
		ID         int64       `json:"id"`
		Name       string      `json:"name"`
		KeyType    string      `json:"key_type"`
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
		keyType := k.KeyType
		if keyType == "" {
			keyType = "low_price"
		}
		items = append(items, apiKeyItem{
			ID:         k.ID,
			Name:       k.Name,
			KeyType:    keyType,
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

// PUT /user/password — 当前登录用户修改自己的密码（已登录状态下无需提供旧密码）
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	db.Engine.ID(userID).Cols("password_hash").Update(&model.User{PasswordHash: string(hash)})
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// DELETE /user/apikeys/:id
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	keyID := strings.TrimSpace(c.Param("id"))
	affected, err := db.Engine.Where("id = ? AND user_id = ?", keyID, userID).
		Delete(&model.APIKey{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API Key 不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API Key 已删除"})
}

// GET /user/stats — 用户仪表盘统计（最近7天消耗趋势 + 累计/今日积分）
func (h *AuthHandler) GetUserStats(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	// 累计消耗积分
	var totalConsumed, todayConsumed int64
	if rows, err := db.Engine.QueryString(`SELECT COALESCE(SUM(CASE
		WHEN type IN ('charge','hold','settle') THEN credits
		WHEN type = 'refund' THEN -credits
		ELSE 0 END), 0) AS total
		FROM billing_transactions WHERE user_id = ?`, userID); err == nil && len(rows) > 0 {
		totalConsumed, _ = strconv.ParseInt(rows[0]["total"], 10, 64)
	}

	// 今日消耗
	if rows, err := db.Engine.QueryString(`SELECT COALESCE(SUM(CASE
		WHEN type IN ('charge','hold','settle') THEN credits
		WHEN type = 'refund' THEN -credits
		ELSE 0 END), 0) AS total
		FROM billing_transactions WHERE user_id = ? AND created_at >= CURRENT_DATE`, userID); err == nil && len(rows) > 0 {
		todayConsumed, _ = strconv.ParseInt(rows[0]["total"], 10, 64)
	}

	// 最近7天每日消耗趋势
	dailyCredits := []gin.H{}
	if rows, err := db.Engine.QueryString(`SELECT TO_CHAR(created_at::date, 'MM-DD') AS day,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','hold','settle') THEN credits
			WHEN type = 'refund' THEN -credits
			ELSE 0 END), 0) AS credits
		FROM billing_transactions
		WHERE user_id = ? AND created_at >= CURRENT_DATE - INTERVAL '6 days'
		GROUP BY created_at::date ORDER BY created_at::date`, userID); err == nil {
		for _, row := range rows {
			v, _ := strconv.ParseInt(row["credits"], 10, 64)
			dailyCredits = append(dailyCredits, gin.H{"day": row["day"], "credits": v})
		}
	}

	// 最近7天每日请求次数（成功/失败）
	dailyRequests := []gin.H{}
	if rows, err := db.Engine.QueryString(`SELECT TO_CHAR(created_at::date, 'MM-DD') AS day,
		COUNT(CASE WHEN status = 'ok' THEN 1 END) AS success,
		COUNT(CASE WHEN status = 'error' THEN 1 END) AS failed
		FROM llm_logs
		WHERE user_id = ? AND created_at >= CURRENT_DATE - INTERVAL '6 days'
		GROUP BY created_at::date ORDER BY created_at::date`, userID); err == nil {
		for _, row := range rows {
			s, _ := strconv.ParseInt(row["success"], 10, 64)
			f, _ := strconv.ParseInt(row["failed"], 10, 64)
			dailyRequests = append(dailyRequests, gin.H{"day": row["day"], "success": s, "failed": f})
		}
	}

	c.JSON(200, gin.H{
		"total_consumed": totalConsumed,
		"today_consumed": todayConsumed,
		"daily_credits":  dailyCredits,
		"daily_requests": dailyRequests,
	})
}

// clientIP 从请求头获取真实客户端 IP。
func clientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if idx := strings.Index(ip, ","); idx != -1 {
		ip = ip[:idx]
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}
