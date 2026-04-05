package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/pkg/mailer"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user with username + password (no email required).
func Register(ctx context.Context, username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "user",
		IsActive:     true,
	}
	if _, err := db.Engine.Insert(user); err != nil {
		return nil, fmt.Errorf("username already taken")
	}
	return user, nil
}

// Login verifies credentials by username or email, returns a JWT.
func Login(ctx context.Context, usernameOrEmail, password string, cfg *config.ServerConfig) (string, *model.User, error) {
	user := &model.User{}
	// try username first, then email
	found, err := db.Engine.Where("username = ?", usernameOrEmail).Get(user)
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}
	if !found {
		found, err = db.Engine.Where("email = ?", usernameOrEmail).Get(user)
		if err != nil || !found {
			return "", nil, fmt.Errorf("invalid username or password")
		}
	}
	if !user.IsActive {
		return "", nil, fmt.Errorf("account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid username or password")
	}

	exp := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	return signed, user, err
}

// BindEmail binds an email to an existing user after verifying the code.
func BindEmail(ctx context.Context, userID int64, email, code string) error {
	if err := VerifyEmailCode(ctx, email, code); err != nil {
		return err
	}
	// check not already taken
	var count int64
	count, err := db.Engine.Where("email = ? AND id != ?", email, userID).Count(new(model.User))
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("email already bound to another account")
	}
	_, err = db.Engine.Where("id = ?", userID).Cols("email").Update(&model.User{Email: email})
	return err
}

// SendPasswordResetCode sends a reset code to the given email if it's bound to an account.
func SendPasswordResetCode(ctx context.Context, email string, m *mailer.Mailer) error {
	var count int64
	count, err := db.Engine.Where("email = ?", email).Count(new(model.User))
	if err != nil {
		return err
	}
	if count == 0 {
		// Don't reveal whether email exists — just return success silently
		return nil
	}
	return SendVerifyCode(ctx, email, m)
}

// ResetPasswordByEmail resets the password after verifying the email code.
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
		return fmt.Errorf("email not bound to any account")
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
		return "", fmt.Errorf("empty cipher text")
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
		return "", fmt.Errorf("invalid cipher text")
	}
	nonce := raw[:gcm.NonceSize()]
	data := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// GenerateAPIKey creates a new API key and stores an encrypted copy for later viewing.
func GenerateAPIKey(ctx context.Context, userID int64, name, secret string) (string, error) {
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

	apiKey := &model.APIKey{
		UserID:    userID,
		KeyHash:   keyHash,
		RawKeyEnc: rawKeyEnc,
		Name:      name,
		IsActive:  true,
	}
	if _, err := db.Engine.Insert(apiKey); err != nil {
		return "", err
	}
	return rawHex, nil
}

// LookupAPIKey finds an active APIKey by raw key value (via hash). Caches in Redis.
func LookupAPIKey(ctx context.Context, rawKey string) (*model.APIKey, error) {
	h := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(h[:])
	cacheKey := fmt.Sprintf("apikey:%s", keyHash)

	// Try Redis cache first
	userIDStr, err := cache.Client.Get(ctx, cacheKey).Result()
	if err == nil && userIDStr != "" {
		// Parse cached user_id
		var userID int64
		fmt.Sscanf(userIDStr, "%d", &userID)
		now := time.Now()
		db.Engine.Where("key_hash = ?", keyHash).Cols("last_used_at").Update(&model.APIKey{LastUsedAt: &now})
		return &model.APIKey{KeyHash: keyHash, UserID: userID, IsActive: true}, nil
	}

	apiKey := &model.APIKey{}
	found, err := db.Engine.Where("key_hash = ? AND is_active = true", keyHash).Get(apiKey)
	if err != nil || !found {
		return nil, fmt.Errorf("invalid api key")
	}

	// Cache {userID} for 30 minutes
	cache.Client.Set(ctx, cacheKey, fmt.Sprintf("%d", apiKey.UserID), 30*time.Minute)
	now := time.Now()
	apiKey.LastUsedAt = &now
	db.Engine.Where("id = ?", apiKey.ID).Cols("last_used_at").Update(apiKey)
	return apiKey, nil
}
