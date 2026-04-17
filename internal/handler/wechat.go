package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WechatHandler 处理微信公众号扫码登录流程。
type WechatHandler struct {
	cfg *config.ServerConfig
}

func NewWechatHandler(cfg *config.ServerConfig) *WechatHandler {
	return &WechatHandler{cfg: cfg}
}

const wechatStatePrefix = "wechat:state:"
const wechatStateTTL = 10 * time.Minute

// POST /auth/wechat/init — 生成微信 OAuth 链接（前端将其渲染为二维码）
func (h *WechatHandler) Init(c *gin.Context) {
	appid := getSettingValue("wechat_appid")
	redirectBase := getSettingValue("wechat_redirect_base_url")
	if appid == "" || redirectBase == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "微信登录未配置"})
		return
	}

	state := uuid.New().String()

	// 存储 state -> pending，10 分钟有效
	cache.Client.Set(context.Background(), wechatStatePrefix+state, "pending", wechatStateTTL)

	// 构造微信 OAuth URL（用户在微信内打开时触发授权）
	redirectURI := url.QueryEscape(redirectBase + "/api/auth/wechat/callback")
	qrURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect",
		appid, redirectURI, state,
	)

	c.JSON(http.StatusOK, gin.H{"state": state, "qr_url": qrURL})
}

// GET /auth/wechat/callback — 微信 OAuth 回调（微信扫码后服务器端接收授权码）
func (h *WechatHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	// 验证 state 存在
	val, err := cache.Client.Get(context.Background(), wechatStatePrefix+state).Result()
	if err != nil || val == "" || val == "done" {
		c.String(http.StatusBadRequest, "state invalid or expired")
		return
	}

	appid := getSettingValue("wechat_appid")
	secret := getSettingValue("wechat_secret")

	// 用 code 换取 access_token + openid
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appid, secret, code,
	)
	resp, err := http.Get(tokenURL) //nolint:noctx
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "微信授权失败")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil || tokenResp.OpenID == "" {
		c.String(http.StatusBadGateway, "解析微信响应失败")
		return
	}

	// 获取用户信息（昵称）
	nickname := ""
	infoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		tokenResp.AccessToken, tokenResp.OpenID,
	)
	if infoResp, err := http.Get(infoURL); err == nil { //nolint:noctx
		defer infoResp.Body.Close()
		var info struct {
			Nickname string `json:"nickname"`
		}
		if b, err := io.ReadAll(infoResp.Body); err == nil {
			json.Unmarshal(b, &info) //nolint:errcheck
			nickname = info.Nickname
		}
	}

	// 登录或注册（无邀请码）
	token, _, err := service.LoginOrRegisterWithOpenID(c.Request.Context(), tokenResp.OpenID, nickname, nil, h.cfg)
	if err != nil {
		c.String(http.StatusInternalServerError, "登录失败: "+err.Error())
		return
	}

	// 把 token 写入 Redis，以 state 为 key
	result, _ := json.Marshal(map[string]string{"token": token})
	cache.Client.Set(context.Background(), wechatStatePrefix+state, string(result), 5*time.Minute)

	// 回调前端页面（可配置）
	frontendURL := getSettingValue("wechat_frontend_url")
	if frontendURL == "" {
		frontendURL = "/"
	}
	c.Redirect(http.StatusFound, frontendURL+"?wechat_login=1")
}

// GET /auth/wechat/poll?state=xxx — 前端轮询登录结果
func (h *WechatHandler) Poll(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 state 参数"})
		return
	}
	val, err := cache.Client.Get(context.Background(), wechatStatePrefix+state).Result()
	if err != nil || val == "" {
		c.JSON(http.StatusOK, gin.H{"status": "expired"})
		return
	}
	if val == "pending" {
		c.JSON(http.StatusOK, gin.H{"status": "pending"})
		return
	}
	// val 是 JSON {"token": "..."}
	var data map[string]string
	if err := json.Unmarshal([]byte(val), &data); err != nil || data["token"] == "" {
		c.JSON(http.StatusOK, gin.H{"status": "pending"})
		return
	}
	// 消费掉 state，防止重复使用
	cache.Client.Del(context.Background(), wechatStatePrefix+state)
	c.JSON(http.StatusOK, gin.H{"status": "success", "token": data["token"]})
}
