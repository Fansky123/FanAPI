package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/db"
	"fanapi/internal/model"
)

const channelCacheTTL = 10 * time.Minute

// ─────────────────────────────────────────────────────────────────────────────
// 基础 CRUD + 单次渠道查询
// ─────────────────────────────────────────────────────────────────────────────

// GetChannel 通过 ID 加载渠道，使用 Redis 作为缓存层。
func GetChannel(ctx context.Context, channelID int64) (*model.Channel, error) {
	cacheKey := fmt.Sprintf("channel:%d", channelID)

	data, err := cache.Client.Get(ctx, cacheKey).Bytes()
	if err == nil {
		ch := &model.Channel{}
		if jsonErr := json.Unmarshal(data, ch); jsonErr == nil {
			return ch, nil
		}
	}

	ch := &model.Channel{}
	found, err := db.Engine.Where("id = ? AND is_active = true", channelID).Get(ch)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("channel not found")
	}

	if b, jsonErr := json.Marshal(ch); jsonErr == nil {
		cache.Client.Set(ctx, cacheKey, b, channelCacheTTL)
	}
	return ch, nil
}

// InvalidateChannelCache 删除渠道对应的 Redis 缓存。
func InvalidateChannelCache(ctx context.Context, channelID int64) {
	cache.Client.Del(ctx, fmt.Sprintf("channel:%d", channelID))
}

// ListChannels 返回所有渠道（管理员接口）。
func ListChannels(ctx context.Context) ([]model.Channel, error) {
	var channels []model.Channel
	err := db.Engine.Find(&channels)
	return channels, err
}

// CreateChannel 插入一个新渠道。
func CreateChannel(ctx context.Context, ch *model.Channel) error {
	_, err := db.Engine.Insert(ch)
	return err
}

// GetChannelByName 通过 Name 字段加载渠道，Name 即路由模型名。
// 缓存键为 "channel:name:{name}"。
// 保留向后兼容；新路由逻辑请使用 SelectChannel。
func GetChannelByName(ctx context.Context, name string) (*model.Channel, error) {
	cacheKey := fmt.Sprintf("channel:name:%s", name)

	data, err := cache.Client.Get(ctx, cacheKey).Bytes()
	if err == nil {
		ch := &model.Channel{}
		if jsonErr := json.Unmarshal(data, ch); jsonErr == nil {
			return ch, nil
		}
	}

	ch := &model.Channel{}
	found, err := db.Engine.Where("name = ? AND is_active = true", name).Get(ch)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("channel %q not found", name)
	}

	if b, jsonErr := json.Marshal(ch); jsonErr == nil {
		cache.Client.Set(ctx, cacheKey, b, channelCacheTTL)
	}
	return ch, nil
}

// UpdateChannel 更新渠道并删除缓存。
func UpdateChannel(ctx context.Context, ch *model.Channel) error {
	_, err := db.Engine.Where("id = ?", ch.ID).AllCols().Update(ch)
	if err == nil {
		InvalidateChannelCache(ctx, ch.ID)
		cache.Client.Del(ctx, fmt.Sprintf("channel:name:%s", ch.Name))
		// 删除模型路由列表缓存
		cache.Client.Del(ctx, fmt.Sprintf("channel:model:%s", ch.Model))
	}
	return err
}

// DeleteChannel 永久删除数据库中的渠道。
func DeleteChannel(ctx context.Context, channelID int64) error {
	// 先加载渠道的 name/model，以便删除相关缓存条目。
	var ch model.Channel
	_, _ = db.Engine.ID(channelID).Cols("name", "model").Get(&ch)
	_, err := db.Engine.Where("id = ?", channelID).Delete(new(model.Channel))
	if err == nil {
		InvalidateChannelCache(ctx, channelID)
		if ch.Name != "" {
			cache.Client.Del(ctx, fmt.Sprintf("channel:name:%s", ch.Name))
		}
		if ch.Model != "" {
			cache.Client.Del(ctx, fmt.Sprintf("channel:model:%s", ch.Model))
		}
	}
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// 负载均衡渠道选择
// ─────────────────────────────────────────────────────────────────────────────

const (
	channelModelListTTL = 30 * time.Second
	// 错误率窗口：当渠道在 errRateWindow 内的错误率超过 errRateThreshold（需满足最少 errRateMinReqs 次请求）时跳过该渠道。
	errRateWindow    = 5 * time.Minute
	errRateThreshold = 0.5 // 错误率 50%
	errRateMinReqs   = 5   // 触发错误率过滤的最小请求数
)

// SelectChannel 使用以下策略选取最优渠道：
//  1. 按优先级降序排序（选最高优先级组）
//  2. 错误率过滤（跳过近 5 分钟内错误率 >50% 的渠道）
//  3. 最高可用优先级组内按权重随机选取
//
// excludeIDs 允许调用方厒除已失败的渠道（用于重试）。
func SelectChannel(ctx context.Context, modelName string, excludeIDs ...int64) (*model.Channel, error) {
	channels, err := listChannelsByModel(ctx, modelName)
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("no active channels for model %q", modelName)
	}

	excluded := make(map[int64]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excluded[id] = true
	}

	// 删除已排除和不健康的渠道
	var candidates []model.Channel
	for _, ch := range channels {
		if excluded[ch.ID] {
			continue
		}
		if isChannelUnhealthy(ctx, ch.ID) {
			continue
		}
		candidates = append(candidates, ch)
	}

	// 若所有健康渠道均已厒除，则回退至所有未厒除渠道
	if len(candidates) == 0 {
		for _, ch := range channels {
			if !excluded[ch.ID] {
				candidates = append(candidates, ch)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("all channels for model %q are exhausted", modelName)
	}

	// 按优先级降序排序，选取最高优先级组
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Priority > candidates[j].Priority
	})
	topPriority := candidates[0].Priority
	var topTier []model.Channel
	for _, ch := range candidates {
		if ch.Priority == topPriority {
			topTier = append(topTier, ch)
		}
	}

	// 在最高优先级组内按权重随机选取
	selected := weightedRandom(topTier)
	return selected, nil
}

// RecordChannelSuccess 记录一次成功请求用于错误率统计。
func RecordChannelSuccess(ctx context.Context, channelID int64) {
	recordChannelEvent(ctx, channelID, "ok")
}

// RecordChannelError 记录一次失败请求用于错误率统计。
func RecordChannelError(ctx context.Context, channelID int64) {
	recordChannelEvent(ctx, channelID, "err")
}

// ─────────────────────────────────────────────────────────────────────────────
// 内部辅助函数
// ─────────────────────────────────────────────────────────────────────────────

func listChannelsByModel(ctx context.Context, modelName string) ([]model.Channel, error) {
	cacheKey := fmt.Sprintf("channel:model:%s", modelName)
	data, err := cache.Client.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var channels []model.Channel
		if jsonErr := json.Unmarshal(data, &channels); jsonErr == nil {
			return channels, nil
		}
	}

	var channels []model.Channel
	err = db.Engine.Where("model = ? AND is_active = true", modelName).Find(&channels)
	if err != nil {
		return nil, err
	}
	if b, jsonErr := json.Marshal(channels); jsonErr == nil {
		cache.Client.Set(ctx, cacheKey, b, channelModelListTTL)
	}
	return channels, nil
}

func weightedRandom(channels []model.Channel) *model.Channel {
	if len(channels) == 1 {
		return &channels[0]
	}
	total := 0
	for _, ch := range channels {
		w := ch.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	r := rand.Intn(total)
	for i, ch := range channels {
		w := ch.Weight
		if w <= 0 {
			w = 1
		}
		r -= w
		if r < 0 {
			return &channels[i]
		}
	}
	return &channels[0]
}

// isChannelUnhealthy 当渠道近期错误率超过 errRateThreshold 时返回 true。
// 使用每个渠道每个时间窗口对应的两个 Redis 计数器。
func isChannelUnhealthy(ctx context.Context, channelID int64) bool {
	window := time.Now().Truncate(errRateWindow).Unix()
	okKey := fmt.Sprintf("ch_stat:%d:%d:ok", channelID, window)
	errKey := fmt.Sprintf("ch_stat:%d:%d:err", channelID, window)

	okStr, _ := cache.Client.Get(ctx, okKey).Result()
	errStr, _ := cache.Client.Get(ctx, errKey).Result()
	okCount, _ := strconv.ParseInt(okStr, 10, 64)
	errCount, _ := strconv.ParseInt(errStr, 10, 64)
	total := okCount + errCount
	if total < errRateMinReqs {
		return false
	}
	return float64(errCount)/float64(total) > errRateThreshold
}

func recordChannelEvent(ctx context.Context, channelID int64, event string) {
	window := time.Now().Truncate(errRateWindow).Unix()
	key := fmt.Sprintf("ch_stat:%d:%d:%s", channelID, window, event)
	cache.Client.Incr(ctx, key)
	// TTL = 2 个窗口周期，确保旧数据干净过期
	cache.Client.Expire(ctx, key, errRateWindow*2)
}
