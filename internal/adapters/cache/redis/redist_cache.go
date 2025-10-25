package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 100,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (r *RedisCache) GetURL(ctx context.Context, shortCode string) (*domain.URL, error) {
	key := r.urlKey(shortCode)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, domain.ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}

	var url domain.URL
	if err := json.Unmarshal([]byte(val), &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *RedisCache) SetURL(ctx context.Context, url *domain.URL, ttl int) error {
	key := r.urlKey(url.ShortCode)

	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	cacheTTL := r.ttl
	if ttl > 0 {
		cacheTTL = time.Duration(ttl) * time.Second
	}

	return r.client.Set(ctx, key, data, cacheTTL).Err()
}

func (r *RedisCache) DeleteURL(ctx context.Context, shortCode string) error {
	key := r.urlKey(shortCode)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) IncrementClickCount(ctx context.Context, shortCode string) error {
	key := r.clickCountKey(shortCode)
	_, err := r.client.Incr(ctx, key).Result()
	return err
}

func (r *RedisCache) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := r.clickCountKey(shortCode)
	return r.client.Get(ctx, key).Int64()
}

func (r *RedisCache) urlKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}

func (r *RedisCache) clickCountKey(shortCode string) string {
	return fmt.Sprintf("clicks:%s", shortCode)
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
