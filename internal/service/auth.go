package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user after verifying the email code.
func Register(ctx context.Context, email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		IsActive:     true,
	}
	if _, err := db.Engine.Insert(user); err != nil {
		return nil, fmt.Errorf("email already registered")
	}
	return user, nil
}

// Login verifies credentials and returns a JWT.
func Login(ctx context.Context, email, password string, cfg *config.ServerConfig) (string, *model.User, error) {
	user := &model.User{}
	found, err := db.Engine.Where("email = ?", email).Get(user)
	if err != nil || !found {
		return "", nil, fmt.Errorf("invalid email or password")
	}
	if !user.IsActive {
		return "", nil, fmt.Errorf("account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
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

// GenerateAPIKey creates a new API key, returns the raw key (shown once).
func GenerateAPIKey(ctx context.Context, userID int64, name string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawHex := hex.EncodeToString(raw)
	h := sha256.Sum256([]byte(rawHex))
	keyHash := hex.EncodeToString(h[:])

	apiKey := &model.APIKey{
		UserID:   userID,
		KeyHash:  keyHash,
		Name:     name,
		IsActive: true,
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
