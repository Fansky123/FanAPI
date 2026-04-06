package billing

import (
	"context"
	"fmt"

	"fanapi/internal/cache"
	"fanapi/internal/db"

	"github.com/redis/go-redis/v9"
)

const balanceKeyFmt = "user:balance:%d"

// SyncBalanceToRedis 将用户的 DB 余额加载到 Redis（在启动时或缓存错过时调用）。
func SyncBalanceToRedis(ctx context.Context, userID int64) (int64, error) {
	var result struct{ Balance int64 }
	_, err := db.Engine.SQL("SELECT balance FROM users WHERE id = ?", userID).Get(&result)
	if err != nil {
		return 0, err
	}
	key := fmt.Sprintf(balanceKeyFmt, userID)
	cache.Client.Set(ctx, key, result.Balance, 0)
	return result.Balance, nil
}

// GetBalance 返回 Redis 缓存的余额，缓存未命中时自动从 DB 同步。
func GetBalance(ctx context.Context, userID int64) (int64, error) {
	key := fmt.Sprintf(balanceKeyFmt, userID)
	val, err := cache.Client.Get(ctx, key).Int64()
	if err == nil {
		return val, nil
	}
	return SyncBalanceToRedis(ctx, userID)
}

// luaCharge 原子地扣减 credits，余额不足时返回失败。
var luaCharge = redis.NewScript(`
local bal = tonumber(redis.call("GET", KEYS[1]))
if not bal then return -2 end
if bal < tonumber(ARGV[1]) then return -1 end
return redis.call("DECRBY", KEYS[1], ARGV[1])
`)

// Charge 原子扣减 credits。余额不足时返回错误。
func Charge(ctx context.Context, userID, credits int64) error {
	if credits <= 0 {
		return nil
	}
	key := fmt.Sprintf(balanceKeyFmt, userID)
	// 确保 Redis 中存在该键
	if _, err := cache.Client.Get(ctx, key).Int64(); err != nil {
		if _, syncErr := SyncBalanceToRedis(ctx, userID); syncErr != nil {
			return syncErr
		}
	}
	result, err := luaCharge.Run(ctx, cache.Client, []string{key}, credits).Int64()
	if err != nil {
		return err
	}
	if result == -1 {
		return fmt.Errorf("insufficient credits")
	}
	if result == -2 {
		return fmt.Errorf("balance key missing")
	}
	return nil
}

// Refund 退还 credits（用于 LLM 输出实际量少于预扣时的差额退款）。
func Refund(ctx context.Context, userID, credits int64) error {
	if credits <= 0 {
		return nil
	}
	key := fmt.Sprintf(balanceKeyFmt, userID)
	// 确保 Redis 键存在，避免 IncrBy 在键不存在时创建只含退款金额的新键
	// （正确行为应为：实际余额 + 退款金额）。
	if _, err := cache.Client.Get(ctx, key).Int64(); err != nil {
		if _, syncErr := SyncBalanceToRedis(ctx, userID); syncErr != nil {
			return syncErr
		}
	}
	return cache.Client.IncrBy(ctx, key, credits).Err()
}
