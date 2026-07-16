/*
 * Redis 快取儲存服務。
 * 1. 建立共用 Redis 連線並驗證可用性。
 * 2. 提供二進位快取讀寫與關閉能力。
 * 3. 將快取未命中與連線錯誤分開處理。
 */
package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"ajoliving_web/http_service/internal/config"
)

// 1. CacheStore defines the shared cache operations used by services.
type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Close() error
}

// 2. RedisCacheStore implements CacheStore with Redis.
type RedisCacheStore struct {
	client *redis.Client
}

// 3. NewRedisCacheStore creates and verifies the configured Redis client.
func NewRedisCacheStore(ctx context.Context, cfg *config.Config) (CacheStore, error) {
	if cfg == nil || !cfg.RedisEnabled {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisCacheStore{client: client}, nil
}

// 4. Get returns one cached value and distinguishes a normal cache miss.
func (s *RedisCacheStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return value, true, nil
}

// 5. Set stores one cached value with an explicit expiry.
func (s *RedisCacheStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// 6. Close releases the Redis client resources.
func (s *RedisCacheStore) Close() error {
	return s.client.Close()
}
