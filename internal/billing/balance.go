package billing

import (
	"context"
	"fmt"

	"fanapi/internal/cache"
	"fanapi/internal/db"

	"github.com/redis/go-redis/v9"
)

const balanceKeyFmt = "user:balance:%d"

// SyncBalanceToRedis loads a user's DB balance into Redis (call on startup / cache miss).
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

// GetBalance returns the Redis-cached balance, syncing from DB on miss.
func GetBalance(ctx context.Context, userID int64) (int64, error) {
	key := fmt.Sprintf(balanceKeyFmt, userID)
	val, err := cache.Client.Get(ctx, key).Int64()
	if err == nil {
		return val, nil
	}
	return SyncBalanceToRedis(ctx, userID)
}

// luaCharge atomically deducts credits, failing if insufficient.
var luaCharge = redis.NewScript(`
local bal = tonumber(redis.call("GET", KEYS[1]))
if not bal then return -2 end
if bal < tonumber(ARGV[1]) then return -1 end
return redis.call("DECRBY", KEYS[1], ARGV[1])
`)

// Charge deducts credits atomically. Returns error if balance insufficient.
func Charge(ctx context.Context, userID, credits int64) error {
	if credits <= 0 {
		return nil
	}
	key := fmt.Sprintf(balanceKeyFmt, userID)
	// Ensure key exists in Redis
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

// Refund adds credits back (for LLM output over-hold refund).
func Refund(ctx context.Context, userID, credits int64) error {
	if credits <= 0 {
		return nil
	}
	key := fmt.Sprintf(balanceKeyFmt, userID)
	// Ensure key exists in Redis so IncrBy doesn't create a new key with just
	// the refund amount instead of (actual_balance + refund_amount).
	if _, err := cache.Client.Get(ctx, key).Int64(); err != nil {
		if _, syncErr := SyncBalanceToRedis(ctx, userID); syncErr != nil {
			return syncErr
		}
	}
	return cache.Client.IncrBy(ctx, key, credits).Err()
}
