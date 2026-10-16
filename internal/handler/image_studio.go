package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const imageStudioTTL = 60 * time.Second

type imageStudioTicket struct {
	UserID  int64   `json:"user_id"`
	KeyID   int64   `json:"key_id"`
	AllKeys bool    `json:"all_keys,omitempty"`
	KeyIDs  []int64 `json:"key_ids,omitempty"`
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
}

func imageStudioURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid image studio URL")
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")) {
		return nil, errors.New("image studio requires HTTPS")
	}
	return u, nil
}

func imageStudioSecretMatches(actual, expected string) bool {
	if len(expected) < 32 {
		return false
	}
	a, b := sha256.Sum256([]byte(actual)), sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

func imageStudioTicketKey(code string) string {
	sum := sha256.Sum256([]byte(code))
	return "image-studio:handoff:" + hex.EncodeToString(sum[:])
}

// Only a configured, trusted deployment can receive keys. No caller-supplied redirect.
func (h *AuthHandler) CreateImageStudioHandoff(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if c.GetString("auth_type") != "jwt" || c.GetString("role") != "user" {
		c.JSON(http.StatusForbidden, gin.H{"error": "请使用用户账户登录后进入图片工作台"})
		return
	}
	target, err := imageStudioURL(os.Getenv("AIPIC_URL"))
	if err != nil || len(os.Getenv("AIPIC_HANDOFF_SECRET")) < 32 || cache.Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "图片工作台尚未配置，请稍后再试"})
		return
	}
	var ticket imageStudioTicket
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	if c.ShouldBindJSON(&ticket) != nil || (!ticket.AllKeys && (ticket.KeyID <= 0 || strings.TrimSpace(ticket.Model) == "")) || len(ticket.Model) > 200 || len(ticket.Prompt) > 32000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片工作台参数无效"})
		return
	}
	ticket.UserID = c.MustGet("user_id").(int64)
	if ticket.AllKeys {
		profiles, err := h.imageStudioProfiles(ticket.UserID, nil)
		if err != nil || len(profiles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "暂无可用密钥，请先创建 API 密钥"})
			return
		}
		ticket.KeyIDs = nil
		for _, profile := range profiles {
			ticket.KeyIDs = append(ticket.KeyIDs, profile.KeyID)
		}
	} else if _, err := h.imageStudioAPIKey(ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所选密钥不可用，请重试"})
		return
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(bytes)
	data, _ := json.Marshal(ticket)
	if err := cache.Client.Set(c.Request.Context(), imageStudioTicketKey(code), data, imageStudioTTL).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "暂时无法进入图片工作台，请重试"})
		return
	}
	target.Fragment = "midcode_code=" + code
	c.JSON(http.StatusOK, gin.H{"url": target.String()})
}

// Called by the AIPic gateway, not by a cross-origin browser request.
func (h *AuthHandler) RedeemImageStudioHandoff(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !imageStudioSecretMatches(c.GetHeader("X-AIPic-Handoff-Secret"), os.Getenv("AIPIC_HANDOFF_SECRET")) {
		c.Status(http.StatusForbidden)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	if c.ShouldBindJSON(&req) != nil || len(req.Code) != 64 {
		c.Status(http.StatusBadRequest)
		return
	}
	if _, err := hex.DecodeString(req.Code); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if cache.Client == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	// Atomic consumption prevents simultaneous or repeated exchanges, across replicas.
	data, err := cache.Client.GetDel(c.Request.Context(), imageStudioTicketKey(req.Code)).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "交接服务暂时不可用，请稍后从 MidCode 重新进入"})
		return
	}
	var ticket imageStudioTicket
	if err != nil || json.Unmarshal(data, &ticket) != nil {
		c.JSON(http.StatusGone, gin.H{"error": "跳转已过期或已使用，请从 MidCode 重新进入"})
		return
	}
	if ticket.AllKeys {
		if len(ticket.KeyIDs) == 0 {
			c.Status(http.StatusForbidden)
			return
		}
		profiles, err := h.imageStudioProfiles(ticket.UserID, ticket.KeyIDs)
		if err != nil || len(profiles) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "密钥或账户已不可用，请从 MidCode 重新进入"})
			return
		}
		active := profiles[0]
		for _, profile := range profiles {
			if len(profile.Models) > 0 {
				active = profile
				break
			}
		}
		c.JSON(http.StatusOK, gin.H{"profiles": profiles, "activeProfileId": active.ID, "apiKey": active.APIKey, "model": active.Model, "prompt": ticket.Prompt})
		return
	}
	rawKey, err := h.imageStudioAPIKey(ticket)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "密钥或账户已不可用，请从 MidCode 重新进入"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiKey": rawKey, "model": ticket.Model, "prompt": ticket.Prompt})
}

func (h *AuthHandler) imageStudioAPIKey(ticket imageStudioTicket) (string, error) {
	var key model.APIKey
	found, err := db.Engine.Where("id = ? AND user_id = ? AND is_active = true", ticket.KeyID, ticket.UserID).Get(&key)
	if err != nil || !found || key.RawKeyEnc == "" {
		return "", errors.New("unavailable key")
	}
	var user model.User
	found, err = db.Engine.Where("id = ? AND is_active = true", ticket.UserID).Get(&user)
	if err != nil || !found {
		return "", errors.New("unavailable user")
	}
	var binding model.APIKeyModelGroup
	found, err = db.Engine.Where("api_key_id = ?", ticket.KeyID).Get(&binding)
	if err != nil || !found {
		return "", errors.New("unbound key")
	}
	return service.DecryptAPIKey(key.RawKeyEnc, h.cfg.APIKeySecret)
}

// Resolve credentials again on redemption, scoped to the authenticated owner and issued key IDs.
type imageStudioProfile struct {
	ID     string   `json:"id"`
	KeyID  int64    `json:"keyId"`
	Name   string   `json:"name"`
	APIKey string   `json:"apiKey"`
	Model  string   `json:"model"`
	Models []string `json:"models"`
}

func (h *AuthHandler) imageStudioProfiles(userID int64, ids []int64) ([]imageStudioProfile, error) {
	var user model.User
	found, err := db.Engine.Where("id = ? AND is_active = true", userID).Get(&user)
	if err != nil || !found {
		return nil, errors.New("unavailable user")
	}
	var keys []model.APIKey
	query := db.Engine.Where("user_id = ? AND is_active = true", userID).OrderBy("id ASC")
	if ids != nil {
		query.In("id", ids)
	}
	if err := query.Find(&keys); err != nil {
		return nil, err
	}
	grouped, err := listGroupedModelChannels()
	if err != nil {
		return nil, err
	}
	profiles := make([]imageStudioProfile, 0, len(keys))
	for _, key := range keys {
		if key.RawKeyEnc == "" {
			continue
		}
		raw, err := service.DecryptAPIKey(key.RawKeyEnc, h.cfg.APIKeySecret)
		if err != nil || raw == "" {
			continue
		}
		var bindings []model.APIKeyModelGroup
		if err := db.Engine.Where("api_key_id = ?", key.ID).OrderBy("priority ASC, id ASC").Find(&bindings); err != nil {
			return nil, err
		}
		models := []string{}
		seen := map[string]bool{}
		for _, binding := range bindings {
			for _, item := range grouped {
				name := strings.TrimSpace(item.RoutingModel)
				if item.Group.ID == binding.GroupID && (strings.HasPrefix(name, "gpt-image-") || strings.HasPrefix(name, "grok-imagine-image")) && !seen[name] {
					models = append(models, name)
					seen[name] = true
				}
			}
		}
		selected := "gpt-image-2"
		if len(models) > 0 {
			selected = models[0]
		}
		name := key.Name
		if name == "" {
			name = fmt.Sprintf("密钥 #%d", key.ID)
		}
		profiles = append(profiles, imageStudioProfile{ID: fmt.Sprintf("midcode-%d-%d", userID, key.ID), KeyID: key.ID, Name: name, APIKey: raw, Model: selected, Models: models})
	}
	return profiles, nil
}
