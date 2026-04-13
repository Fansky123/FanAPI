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
		return fmt.Errorf("请等待 1 分钟后再重新获取验证码")
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	codeKey := fmt.Sprintf("email:verify:%s", email)

	if err := cache.Client.Set(ctx, codeKey, code, verifyCodeTTL).Err(); err != nil {
		return err
	}
	if err := cache.Client.Set(ctx, limitKey, "1", sendLimitTTL).Err(); err != nil {
		return err
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
</head>
<body style="margin:0;padding:0;background:#f4f6f9;font-family:'PingFang SC','Helvetica Neue',Arial,sans-serif">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f6f9;padding:40px 0">
  <tr>
    <td align="center">
      <table width="480" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,0.08)">
        <!-- 顶部色条 -->
        <tr>
          <td style="background:linear-gradient(135deg,#4f46e5 0%%,#7c3aed 100%%);height:6px"></td>
        </tr>
        <!-- 内容区 -->
        <tr>
          <td style="padding:40px 48px 32px">
            <!-- 标题 -->
            <h2 style="margin:0 0 8px;font-size:22px;font-weight:700;color:#1a1a2e">邮箱验证码</h2>
            <p style="margin:0 0 28px;font-size:14px;color:#6b7280">您正在进行邮箱验证操作，请使用以下验证码完成验证。</p>
            <!-- 验证码卡片 -->
            <div style="background:#f5f3ff;border:1.5px dashed #7c3aed;border-radius:10px;padding:24px;text-align:center;margin-bottom:28px">
              <div style="font-size:36px;font-weight:800;letter-spacing:10px;color:#4f46e5;font-family:monospace">%s</div>
            </div>
            <!-- 说明 -->
            <p style="margin:0 0 8px;font-size:13px;color:#6b7280;line-height:1.6">
              验证码有效期为 <strong style="color:#374151">5 分钟</strong>，请勿泄露给他人。<br/>
              如果您没有发起此请求，请忽略此邮件。
            </p>
          </td>
        </tr>
        <!-- 底部 -->
        <tr>
          <td style="background:#f9fafb;padding:16px 48px;border-top:1px solid #e5e7eb">
            <p style="margin:0;font-size:12px;color:#9ca3af;text-align:center">此邮件由系统自动发送，请勿直接回复</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`, code)
	return m.Send(email, "邮箱验证码", body)
}

func VerifyEmailCode(ctx context.Context, email, code string) error {
	key := fmt.Sprintf("email:verify:%s", email)
	stored, err := cache.Client.Get(ctx, key).Result()
	if err != nil {
		// redis.Nil 表示 key 不存在（验证码已过期或未发送）
		return fmt.Errorf("验证码已过期或不存在")
	}
	if stored != code {
		return fmt.Errorf("验证码错误")
	}
	cache.Client.Del(ctx, key)
	return nil
}
