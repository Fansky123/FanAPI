package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"fanapi/internal/cache"
	"fanapi/pkg/mailer"
)

const (
	verifyCodeTTL = 5 * time.Minute
	sendLimitTTL  = 1 * time.Minute
)

func SendVerifyCode(ctx context.Context, email string, m *mailer.Mailer) error {
	limitKey := fmt.Sprintf("email:limit:%s", email)
	exists, err := cache.Client.Exists(ctx, limitKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("please wait 1 minute before requesting another code")
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	codeKey := fmt.Sprintf("email:verify:%s", email)

	if err := cache.Client.Set(ctx, codeKey, code, verifyCodeTTL).Err(); err != nil {
		return err
	}
	if err := cache.Client.Set(ctx, limitKey, "1", sendLimitTTL).Err(); err != nil {
		return err
	}

	body := fmt.Sprintf("<p>Your verification code: <strong>%s</strong> (valid for 5 minutes)</p>", code)
	return m.Send(email, "FanAPI Verification Code", body)
}

func VerifyEmailCode(ctx context.Context, email, code string) error {
	key := fmt.Sprintf("email:verify:%s", email)
	stored, err := cache.Client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("code expired or not found")
	}
	if stored != code {
		return fmt.Errorf("invalid verification code")
	}
	cache.Client.Del(ctx, key)
	return nil
}
