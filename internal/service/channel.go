package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/db"
	"fanapi/internal/model"
)

const channelCacheTTL = 10 * time.Minute

// GetChannel loads a channel by ID, using Redis as cache.
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

// InvalidateChannelCache removes the Redis cache for a channel.
func InvalidateChannelCache(ctx context.Context, channelID int64) {
	cache.Client.Del(ctx, fmt.Sprintf("channel:%d", channelID))
}

// ListChannels returns all channels for the admin.
func ListChannels(ctx context.Context) ([]model.Channel, error) {
	var channels []model.Channel
	err := db.Engine.Find(&channels)
	return channels, err
}

// CreateChannel inserts a new channel.
func CreateChannel(ctx context.Context, ch *model.Channel) error {
	_, err := db.Engine.Insert(ch)
	return err
}

// GetChannelByName loads a channel by its Name field, which serves as the
// model routing key for API calls. Cached under "channel:name:{name}".
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

// UpdateChannel updates an existing channel and invalidates cache.
func UpdateChannel(ctx context.Context, ch *model.Channel) error {
	_, err := db.Engine.Where("id = ?", ch.ID).AllCols().Update(ch)
	if err == nil {
		InvalidateChannelCache(ctx, ch.ID)
		cache.Client.Del(ctx, fmt.Sprintf("channel:name:%s", ch.Name))
	}
	return err
}

// DeleteChannel permanently removes a channel from the database.
func DeleteChannel(ctx context.Context, channelID int64) error {
	// Load name first so we can invalidate the name-based cache entry.
	var ch model.Channel
	_, _ = db.Engine.ID(channelID).Cols("name").Get(&ch)
	_, err := db.Engine.Where("id = ?", channelID).Delete(new(model.Channel))
	if err == nil {
		InvalidateChannelCache(ctx, channelID)
		if ch.Name != "" {
			cache.Client.Del(ctx, fmt.Sprintf("channel:name:%s", ch.Name))
		}
	}
	return err
}
