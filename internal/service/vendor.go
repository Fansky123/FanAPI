package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterVendor 注册新号商账号。
func RegisterVendor(ctx context.Context, username, password string) (*model.Vendor, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	inviteCode := hex.EncodeToString(b)

	vendor := &model.Vendor{
		Username:     username,
		PasswordHash: string(hash),
		IsActive:     true,
		InviteCode:   inviteCode,
	}
	if _, err := db.Engine.Insert(vendor); err != nil {
		log.Printf("[vendor-register] db insert error: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("用户名已被占用")
		}
		return nil, fmt.Errorf("注册失败，请稍后重试")
	}
	return vendor, nil
}

// LoginVendor 验证号商密码，成功返回 JWT。
func LoginVendor(ctx context.Context, username, password string, cfg *config.ServerConfig) (string, *model.Vendor, error) {
	vendor := &model.Vendor{}
	found, err := db.Engine.Where("username = ?", username).Get(vendor)
	if err != nil {
		return "", nil, fmt.Errorf("内部错误，请稍后重试")
	}
	if !found {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}
	if !vendor.IsActive {
		return "", nil, fmt.Errorf("账号已被禁用，请联系管理员")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(vendor.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	exp := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	claims := jwt.MapClaims{
		"sub":  vendor.ID,
		"role": "vendor",
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	return signed, vendor, err
}
