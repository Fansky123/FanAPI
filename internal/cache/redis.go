package cache

import (
	"context"

	"fanapi/internal/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return Client.Ping(context.Background()).Err()
}
