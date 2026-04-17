package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/pkg/mailer"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register 创建新用户（用户名 + 密码，无需邀算验证）。
// inviterID 非 nil 时记录邀请人。
func Register(ctx context.Context, username, password string, inviterID *int64) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "user",
		IsActive:     true,
		InviteCode:   generateInviteCode(),
		InviterID:    inviterID,
	}
	if _, err := db.Engine.Insert(user); err != nil {
		log.Printf("[register] db insert error: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("用户名已被占用")
		}
		return nil, fmt.Errorf("注册失败，请稍后重试")
	}
	return user, nil
}

// generateInviteCode 生成 16 位十六进制邀请码。
func generateInviteCode() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

// Login 验证用户名或邂算密码，验证成功返回 JWT。
func Login(ctx context.Context, usernameOrEmail, password string, cfg *config.ServerConfig) (string, *model.User, error) {
	user := &model.User{}
	// 先尝试用户名登录，失败再尝试邂算
	found, err := db.Engine.Where("username = ?", usernameOrEmail).Get(user)
	if err != nil {
		log.Printf("[login] db error (username lookup): %v", err)
		return "", nil, fmt.Errorf("内部错误，请稍后重试")
	}
	if !found {
		found, err = db.Engine.Where("email = ?", usernameOrEmail).Get(user)
		if err != nil {
			log.Printf("[login] db error (email lookup): %v", err)
			return "", nil, fmt.Errorf("内部错误，请稍后重试")
		}
		if !found {
			return "", nil, fmt.Errorf("用户名或密码错误")
		}
	}
	if !user.IsActive {
		return "", nil, fmt.Errorf("账号已被禁用，请联系管理员")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	exp := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"role":  user.Role,
		"group": user.Group,
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	return signed, user, err
}

// LoginOrRegisterWithOpenID 通过微信 OpenID 获取或创建用户，返回 JWT。
// nickname 为微信昵称（首次注册时用作用户名前缀）。
func LoginOrRegisterWithOpenID(ctx context.Context, openid, nickname string, inviterID *int64, cfg *config.ServerConfig) (string, *model.User, error) {
	user := &model.User{}
	found, err := db.Engine.Where("wechat_openid = ?", openid).Get(user)
	if err != nil {
		return "", nil, fmt.Errorf("内部错误，请稍后重试")
	}
	if !found {
		// 首次微信登录：自动注册
		base := nickname
		if base == "" {
			base = "wx"
		}
		b := make([]byte, 3)
		rand.Read(b) //nolint:errcheck
		username := fmt.Sprintf("%s_%s", base, hex.EncodeToString(b))
		// 生成随机密码（用户无需密码登录，但字段不能为空）
		rawPwd := make([]byte, 16)
		rand.Read(rawPwd)
		hash, _ := bcrypt.GenerateFromPassword(rawPwd, bcrypt.DefaultCost)
		user = &model.User{
			Username:     username,
			PasswordHash: string(hash),
			Role:         "user",
			IsActive:     true,
			InviteCode:   generateInviteCode(),
			InviterID:    inviterID,
			WechatOpenID: openid,
		}
		if _, err := db.Engine.Insert(user); err != nil {
			// 并发重复注册时再次查询
			if found2, _ := db.Engine.Where("wechat_openid = ?", openid).Get(user); !found2 {
				return "", nil, fmt.Errorf("注册失败，请稍后重试")
			}
		}
	}
	if !user.IsActive {
		return "", nil, fmt.Errorf("账号已被禁用，请联系管理员")
	}
	exp := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"role":  user.Role,
		"group": user.Group,
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	return signed, user, err
}

// BindEmail 验证代码后将邂算绑定到已登录用户。
func BindEmail(ctx context.Context, userID int64, email, code string) error {
	if err := VerifyEmailCode(ctx, email, code); err != nil {
		return err
	}
	// 检查邂算是否已被其他账户绑定
	var count int64
	count, err := db.Engine.Where("email = ? AND id != ?", email, userID).Count(new(model.User))
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该邮箱已被其他账号绑定")
	}
	_, err = db.Engine.Where("id = ?", userID).Cols("email").Update(&model.User{Email: &email})
	return err
}

// SendPasswordResetCode 如果邂算已绑定账户，就向该邂算发送重置验证码。
func SendPasswordResetCode(ctx context.Context, email string, m *mailer.Mailer) error {
	var count int64
	count, err := db.Engine.Where("email = ?", email).Count(new(model.User))
	if err != nil {
		return err
	}
	if count == 0 {
		// 不透露邂算是否存在，静默返回成功防止枚举
		return nil
	}
	return SendVerifyCode(ctx, email, m)
}

// ResetPasswordByEmail 通过邂算验证码重置密码。
func ResetPasswordByEmail(ctx context.Context, email, code, newPassword string) error {
	if err := VerifyEmailCode(ctx, email, code); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	affected, err := db.Engine.Where("email = ?", email).Cols("password_hash").
		Update(&model.User{PasswordHash: string(hash)})
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("该邮箱未绑定任何账号")
	}
	return nil
}

func encryptAPIKey(rawKey, secret string) (string, error) {
	key := sha256.Sum256([]byte(secret + ":fanapi:apikey"))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(rawKey), nil)
	buf := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(buf), nil
}

func DecryptAPIKey(cipherText, secret string) (string, error) {
	if cipherText == "" {
		return "", fmt.Errorf("重置链接无效")
	}
	key := sha256.Sum256([]byte(secret + ":fanapi:apikey"))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("重置链接无效或已过期")
	}
	nonce := raw[:gcm.NonceSize()]
	data := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// GenerateAPIKey 创建新 API Key 并将加密副本存入 DB（供用户后续查看）。
func GenerateAPIKey(ctx context.Context, userID int64, name, keyType, secret string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawHex := hex.EncodeToString(raw)
	h := sha256.Sum256([]byte(rawHex))
	keyHash := hex.EncodeToString(h[:])
	rawKeyEnc, err := encryptAPIKey(rawHex, secret)
	if err != nil {
		return "", err
	}

	if keyType == "" {
		keyType = "low_price"
	}
	apiKey := &model.APIKey{
		UserID:    userID,
		KeyHash:   keyHash,
		RawKeyEnc: rawKeyEnc,
		Name:      name,
		KeyType:   keyType,
		IsActive:  true,
	}
	if _, err := db.Engine.Insert(apiKey); err != nil {
		return "", err
	}
	return rawHex, nil
}

// LookupAPIKey 通过哈希查找活跃的 APIKey（Redis 缓存加速）。
func LookupAPIKey(ctx context.Context, rawKey string) (*model.APIKey, error) {
	h := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(h[:])
	cacheKey := fmt.Sprintf("apikey2:%s", keyHash)

	// 先查 Redis 缓存（存储完整 APIKey JSON）
	cached, err := cache.Client.Get(ctx, cacheKey).Bytes()
	if err == nil && len(cached) > 0 {
		var apiKey model.APIKey
		if jsonErr := json.Unmarshal(cached, &apiKey); jsonErr == nil {
			now := time.Now()
			db.Engine.Where("key_hash = ?", keyHash).Cols("last_used_at").Update(&model.APIKey{LastUsedAt: &now})
			return &apiKey, nil
		}
	}

	apiKey := &model.APIKey{}
	found, err := db.Engine.Where("key_hash = ? AND is_active = true", keyHash).Get(apiKey)
	if err != nil {
		log.Printf("[apikey] db error: %v", err)
		return nil, fmt.Errorf("内部错误，请稍后重试")
	}
	if !found {
		return nil, fmt.Errorf("API Key 无效")
	}

	// 缓存完整 APIKey 30 分钟
	if b, jsonErr := json.Marshal(apiKey); jsonErr == nil {
		cache.Client.Set(ctx, cacheKey, b, 30*time.Minute)
	}
	now := time.Now()
	apiKey.LastUsedAt = &now
	db.Engine.Where("id = ?", apiKey.ID).Cols("last_used_at").Update(apiKey)
	return apiKey, nil
}
