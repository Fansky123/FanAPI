package service

import (
	"context"
	"fmt"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/db"
	"fanapi/internal/model"
)

const (
	// exhaustedTTL 是三方 Key 被标记为耗尽后的自动恢复时间。
	// 超过此时间后该 Key 会重新参与分配轮转。
	exhaustedTTL = time.Hour

	assignKeyFmt    = "pool:assign:%d:%d" // Redis 键格式：pool:assign:{pool_id}:{entity_id}
	exhaustedKeyFmt = "pool:exhausted:%d" // Redis 键格式：pool:exhausted:{pool_key_id}
)

// GetOrAssignPoolKey 返回分配给 entityID 的三方 PoolKey（Sticky Assignment）。
//
// entityID 取值规则：
//   - LLM 场景：优先使用 api_key_id，若无则使用 user_id
//   - 异步任务场景：使用 user_id
//
// 分配策略：
//  1. 若 entityID 已有分配且该 Key 未耗尽 → 直接复用（保证上下文延续 / 缓存命中）
//  2. 若已分配但 Key 已耗尽，或尚未分配 → 按 priority ASC, id ASC 选择第一个可用 Key 并绑定
func GetOrAssignPoolKey(ctx context.Context, poolID, entityID int64) (*model.PoolKey, error) {
	// 检查号池本身是否启用
	pool := &model.KeyPool{}
	found, err := db.Engine.ID(poolID).Get(pool)
	if err != nil {
		return nil, fmt.Errorf("号池 %d: 读取失败: %w", poolID, err)
	}
	if !found {
		return nil, fmt.Errorf("号池 %d 不存在", poolID)
	}
	if !pool.IsActive {
		return nil, fmt.Errorf("号池 %d 已停用", poolID)
	}

	assignKey := fmt.Sprintf(assignKeyFmt, poolID, entityID)

	// 1. 查询当前分配
	assignedIDStr, err := cache.Client.Get(ctx, assignKey).Result()
	if err == nil {
		var assignedID int64
		fmt.Sscanf(assignedIDStr, "%d", &assignedID)

		exhaustedKey := fmt.Sprintf(exhaustedKeyFmt, assignedID)
		exhausted, _ := cache.Client.Exists(ctx, exhaustedKey).Result()
		if exhausted == 0 {
			// 已分配且未耗尽 → 直接复用
			key := &model.PoolKey{}
			found, dbErr := db.Engine.Where("id = ? AND is_active = true", assignedID).Get(key)
			if dbErr == nil && found {
				return key, nil
			}
		}
	}

	// 2. 尚未分配 or 当前分配已耗尽 → 重新轮转
	return rotatePoolKey(ctx, poolID, entityID, assignKey, 0)
}

// MarkExhaustedAndRotate 将 poolKeyID 标记为耗尽（带 TTL），同时为 entityID 轮转到下一可用 Key。
// 用于检测到上游 429 / 配额不足时主动触发轮转。
func MarkExhaustedAndRotate(ctx context.Context, poolID, poolKeyID, entityID int64) (*model.PoolKey, error) {
	exhaustedKey := fmt.Sprintf(exhaustedKeyFmt, poolKeyID)
	cache.Client.Set(ctx, exhaustedKey, 1, exhaustedTTL)

	assignKey := fmt.Sprintf(assignKeyFmt, poolID, entityID)
	return rotatePoolKey(ctx, poolID, entityID, assignKey, poolKeyID)
}

// rotatePoolKey 从池中选择第一个未耗尽的可用 Key（跳过 skipKeyID），并写入分配记录。
func rotatePoolKey(ctx context.Context, poolID, entityID int64, assignKey string, skipKeyID int64) (*model.PoolKey, error) {
	var keys []model.PoolKey
	if err := db.Engine.Where("pool_id = ? AND is_active = true", poolID).
		OrderBy("priority ASC, id ASC").Find(&keys); err != nil {
		return nil, fmt.Errorf("key pool %d: db error: %w", poolID, err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("号池 %d 暂无可用 Key", poolID)
	}

	for i := range keys {
		k := &keys[i]
		if k.ID == skipKeyID {
			continue
		}
		exhaustedKey := fmt.Sprintf(exhaustedKeyFmt, k.ID)
		exists, _ := cache.Client.Exists(ctx, exhaustedKey).Result()
		if exists == 0 {
			// 绑定分配（无过期时间 = Sticky，永久持有直到主动轮转）
			cache.Client.Set(ctx, assignKey, fmt.Sprintf("%d", k.ID), 0)
			return k, nil
		}
	}
	return nil, fmt.Errorf("号池 %d 的所有 Key 已耗尽", poolID)
}

// ---- 管理接口 ----

// ListKeyPools 返回号池列表（管理端，不过滤 is_active，软删除通过前端展示状态区分）。
// channelID > 0 时按渠道过滤，否则返回全部号池。
func ListKeyPools(ctx context.Context, channelID int64) ([]model.KeyPool, error) {
	pools := make([]model.KeyPool, 0)
	var err error
	if channelID > 0 {
		err = db.Engine.Where("channel_id = ?", channelID).OrderBy("id ASC").Find(&pools)
	} else {
		err = db.Engine.OrderBy("id ASC").Find(&pools)
	}
	return pools, err
}

// CreateKeyPool 创建一个新号池。
func CreateKeyPool(ctx context.Context, pool *model.KeyPool) error {
	_, err := db.Engine.Insert(pool)
	return err
}

// ToggleKeyPool 切换号池启用/停用状态。
func ToggleKeyPool(ctx context.Context, poolID int64) error {
	pool := &model.KeyPool{}
	found, err := db.Engine.ID(poolID).Get(pool)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("号池 %d 不存在", poolID)
	}
	_, err = db.Engine.ID(poolID).Cols("is_active").Update(&model.KeyPool{IsActive: !pool.IsActive})
	return err
}

// DeleteKeyPool 删除号池及其所有 Key。
func DeleteKeyPool(ctx context.Context, poolID int64) error {
	if _, err := db.Engine.Where("pool_id = ?", poolID).Delete(&model.PoolKey{}); err != nil {
		return err
	}
	_, err := db.Engine.ID(poolID).Delete(&model.KeyPool{})
	return err
}

// ListPoolKeys 返回号池内所有 Key（含排序，供管理界面展示）。
func ListPoolKeys(ctx context.Context, poolID int64) ([]model.PoolKey, error) {
	keys := make([]model.PoolKey, 0)
	err := db.Engine.Where("pool_id = ?", poolID).OrderBy("priority ASC, id ASC").Find(&keys)
	return keys, err
}

// AddPoolKey 向号池添加一个三方 Key。
func AddPoolKey(ctx context.Context, key *model.PoolKey) error {
	_, err := db.Engine.Insert(key)
	return err
}

// RemovePoolKey 删除号池中的一个 Key。
func RemovePoolKey(ctx context.Context, keyID int64) error {
	_, err := db.Engine.ID(keyID).Delete(&model.PoolKey{})
	return err
}
