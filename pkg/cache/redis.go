package cache

import (
	"context"
	"fmt"
	"marketplace/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultTTL = 10 * time.Minute // Время жизни кеша по умолчанию
)

type CacheClient struct {
	Client *redis.Client
	ttl    time.Duration
}

// NewRedisClient создает и возвращает нового клиента для Redis.
func NewRedisClient(cfg config.Redis) (*CacheClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	rdb.ConfigSet(ctx, "maxmemory", "100mb")
	rdb.ConfigSet(ctx, "maxmemory-policy", "allkeys-lru")

	// Проверяем соединение
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &CacheClient{
		Client: rdb,
		ttl:    defaultTTL,
	}, nil
}