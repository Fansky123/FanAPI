package handler

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WechatMPHandler 处理微信公众号扫码登录流程。
// 流程：前端请求二维码 → 用户扫码关注/扫码 → 微信推送事件到 webhook →
// 服务端完成登录并缓存 token → 前端轮询获取 token。
type WechatMPHandler struct {
	cfg *config.ServerConfig
}

func NewWechatMPHandler(cfg *config.ServerConfig) *WechatMPHandler {
	return &WechatMPHandler{cfg: cfg}
}

const (
	mpSessionPrefix     = "wechat_mp:session:"
	mpTrackPrefix       = "wechat_mp:track:"
	mpAccessTokenPrefix = "wechat_mp:access_token:"
	mpSessionTTL        = 10 * time.Minute
	mpSuccessSessionTTL = 5 * time.Minute
)

// GET /auth/wechat-mp/qrcode — 生成公众号场景二维码，供前端展示。
func (h *WechatMPHandler) GetQRCode(c *gin.Context) {
	appid := getSettingValue("wechat_mp_appid")
	secret := getSettingValue("wechat_mp_secret")
	if appid == "" || secret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "微信公众号登录未配置"})
		return
	}

	accessToken, err := getMPAccessToken(appid, secret)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "获取微信 access_token 失败: " + err.Error()})
		return
	}

	// 使用 UUID 作为二维码场景值，唯一标识本次扫码会话
	sceneStr := uuid.New().String()

	ticket, err := createMPQRCode(accessToken, sceneStr, 600)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "生成微信二维码失败: " + err.Error()})
		return
	}

	imgBase64, err := getMPQRCodeImage(ticket)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "获取微信二维码图片失败: " + err.Error()})
		return
	}

	// 将会话状态写入 Redis，等待用户扫码
	ctx := context.Background()
	cache.Client.Set(ctx, mpSessionPrefix+sceneStr, "pending", mpSessionTTL)

	// 保存广告追踪参数（可选），供扫码后上报 OCPC 使用
	bdVid := c.Query("bd_vid")
	qhClickID := c.Query("qh_click_id")
	sourceID := c.Query("source_id")
	ocpcID := c.Query("ocpc_id") // 对应 ocpc_platforms.id
	if bdVid != "" || qhClickID != "" || sourceID != "" || ocpcID != "" {
		trackData, _ := json.Marshal(map[string]string{
			"bd_vid":      bdVid,
			"qh_click_id": qhClickID,
			"source_id":   sourceID,
			"ocpc_id":     ocpcID,
			"ip":          clientIP(c),
			"ua":          c.GetHeader("User-Agent"),
		})
		cache.Client.Set(ctx, mpTrackPrefix+sceneStr, string(trackData), mpSessionTTL)
	}

	c.JSON(http.StatusOK, gin.H{"uuid": sceneStr, "qr_img": imgBase64})
}

// mpWXEvent 表示微信推送的事件消息。
type mpWXEvent struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
	EventKey     string   `xml:"EventKey"`
}

// GET /auth/wechat-mp/event — 微信服务器验证 webhook 有效性。
// POST /auth/wechat-mp/event — 接收微信推送事件（扫码/关注）。
func (h *WechatMPHandler) Event(c *gin.Context) {
	mpToken := getSettingValue("wechat_mp_token")

	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	if !verifyMPSignature(mpToken, timestamp, nonce, signature) {
		c.String(http.StatusForbidden, "invalid signature")
		return
	}

	// GET 请求：返回 echostr 完成 webhook 验证
	if c.Request.Method == http.MethodGet {
		c.String(http.StatusOK, c.Query("echostr"))
		return
	}

	// POST 请求：处理推送事件
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusOK, "")
		return
	}

	var event mpWXEvent
	if err := xml.Unmarshal(body, &event); err != nil || event.MsgType != "event" {
		c.String(http.StatusOK, "")
		return
	}

	// 解析场景值（subscribe 事件 EventKey 有 "qrscene_" 前缀）
	sceneStr := ""
	switch strings.ToUpper(event.Event) {
	case "SUBSCRIBE":
		// 新用户扫码并关注
		if strings.HasPrefix(event.EventKey, "qrscene_") {
			sceneStr = strings.TrimPrefix(event.EventKey, "qrscene_")
		}
	case "SCAN":
		// 已关注用户扫码
		sceneStr = event.EventKey
	}

	if sceneStr == "" {
		c.String(http.StatusOK, "")
		return
	}

	ctx := context.Background()

	// 检查会话是否仍处于 pending 状态
	val, err := cache.Client.Get(ctx, mpSessionPrefix+sceneStr).Result()
	if err != nil || val != "pending" {
		c.String(http.StatusOK, "")
		return
	}

	// 获取用户昵称（丰富注册信息）
	nickname := ""
	appid := getSettingValue("wechat_mp_appid")
	secret := getSettingValue("wechat_mp_secret")
	if at, err := getMPAccessToken(appid, secret); err == nil {
		nickname = getMPUserNickname(at, event.FromUserName)
	}

	// 通过 openid 登录或自动注册用户
	token, user, err := service.LoginOrRegisterWithOpenID(ctx, event.FromUserName, nickname, nil, h.cfg)
	if err != nil {
		c.String(http.StatusOK, "")
		return
	}

	// 上报 OCPC 注册转化（若本次扫码携带了广告追踪参数）
	if trackRaw, trackErr := cache.Client.Get(ctx, mpTrackPrefix+sceneStr).Result(); trackErr == nil && trackRaw != "" {
		var track map[string]string
		if json.Unmarshal([]byte(trackRaw), &track) == nil {
			platformID, _ := strconv.ParseInt(track["ocpc_id"], 10, 64)
			service.CreateOrUpdateOcpcRecord(ctx, user.ID, platformID,
				track["bd_vid"], track["qh_click_id"], track["source_id"],
				track["ip"], track["ua"])
		}
		cache.Client.Del(ctx, mpTrackPrefix+sceneStr)
	}

	// 将 token 写入 Redis，前端轮询时取走
	result, _ := json.Marshal(map[string]string{"token": token})
	cache.Client.Set(ctx, mpSessionPrefix+sceneStr, string(result), mpSuccessSessionTTL)

	// 向用户回复文字消息
	reply := buildMPTextReply(event.ToUserName, event.FromUserName, "登录成功！")
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, reply)
}

// GET /auth/wechat-mp/poll?uuid=xxx — 前端轮询扫码登录结果。
func (h *WechatMPHandler) Poll(c *gin.Context) {
	sceneStr := c.Query("uuid")
	if sceneStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 uuid 参数"})
		return
	}

	val, err := cache.Client.Get(context.Background(), mpSessionPrefix+sceneStr).Result()
	if err != nil || val == "" {
		c.JSON(http.StatusOK, gin.H{"status": "expired"})
		return
	}
	if val == "pending" {
		c.JSON(http.StatusOK, gin.H{"status": "pending"})
		return
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(val), &data); err != nil || data["token"] == "" {
		c.JSON(http.StatusOK, gin.H{"status": "pending"})
		return
	}

	// 消费 session，防止重复使用
	cache.Client.Del(context.Background(), mpSessionPrefix+sceneStr)
	c.JSON(http.StatusOK, gin.H{"status": "success", "token": data["token"]})
}

// getMPAccessToken 从 Redis 缓存中获取，或向微信重新申请公众号 access_token。
func getMPAccessToken(appid, secret string) (string, error) {
	ctx := context.Background()
	cacheKey := mpAccessTokenPrefix + appid

	if cached, err := cache.Client.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		return cached, nil
	}

	resp, err := http.Get(fmt.Sprintf( //nolint:noctx
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(appid), url.QueryEscape(secret),
	))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.AccessToken == "" {
		return "", fmt.Errorf("微信返回异常: %s", string(body))
	}

	// 提前 200 秒过期，避免临界情况下 token 失效
	ttl := time.Duration(result.ExpiresIn-200) * time.Second
	cache.Client.Set(ctx, cacheKey, result.AccessToken, ttl)
	return result.AccessToken, nil
}

// createMPQRCode 调用微信接口创建临时字符串场景二维码，返回 ticket。
func createMPQRCode(accessToken, sceneStr string, expireSeconds int) (string, error) {
	reqBody := fmt.Sprintf(
		`{"expire_seconds":%d,"action_name":"QR_STR_SCENE","action_info":{"scene":{"scene_str":"%s"}}}`,
		expireSeconds, sceneStr,
	)
	apiURL := "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=" + url.QueryEscape(accessToken)
	resp, err := http.Post(apiURL, "application/json", strings.NewReader(reqBody)) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Ticket  string `json:"ticket"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Ticket == "" {
		return "", fmt.Errorf("微信返回异常: %s", string(body))
	}
	return result.Ticket, nil
}

// getMPQRCodeImage 通过 ticket 下载二维码图片，返回 base64 DataURL。
func getMPQRCodeImage(ticket string) (string, error) {
	imgURL := "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=" + url.QueryEscape(ticket)
	resp, err := http.Get(imgURL) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(imgBytes)
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("非图片响应: %s", contentType)
	}

	encoded := base64.StdEncoding.EncodeToString(imgBytes)
	return "data:" + contentType + ";base64," + encoded, nil
}

// getMPUserNickname 查询公众号用户信息，获取昵称。
func getMPUserNickname(accessToken, openid string) string {
	infoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(accessToken), url.QueryEscape(openid),
	)
	resp, err := http.Get(infoURL) //nolint:noctx
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		Nickname string `json:"nickname"`
	}
	json.Unmarshal(body, &info) //nolint:errcheck
	return info.Nickname
}

// verifyMPSignature 验证微信推送消息的签名。
func verifyMPSignature(token, timestamp, nonce, signature string) bool {
	strs := []string{token, timestamp, nonce}
	sort.Strings(strs)
	h := sha1.New()
	h.Write([]byte(strings.Join(strs, "")))
	computed := fmt.Sprintf("%x", h.Sum(nil))
	return computed == signature
}

// buildMPTextReply 构造微信被动回复文本消息 XML。
func buildMPTextReply(toUser, fromUser, content string) string {
	return fmt.Sprintf(`<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%d</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[%s]]></Content>
</xml>`, fromUser, toUser, time.Now().Unix(), content)
}
