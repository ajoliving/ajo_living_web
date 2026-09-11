/*
 * 即時聊天廣播適配器。
 * 1. 以 Redis Pub/Sub 在多個 HTTP 實例之間轉發已提交的聊天事件。
 * 2. Redis 只作瞬時廣播，消息真相仍保存在 PostgreSQL。
 * 3. 啟動時驗證 Redis 連線，關閉時釋放客戶端資源。
 */
package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"ajoliving_web/http_service/internal/config"
)

const realtimeBroadcastChannel = "ajo:realtime:chat-events"

// 1. RealtimeBroker defines cross-instance realtime event delivery.
type RealtimeBroker interface {
	Publish(ctx context.Context, payload []byte) error
	Subscribe(ctx context.Context) (<-chan []byte, func(), error)
	Close() error
}

// 2. RedisRealtimeBroker publishes realtime events through Redis Pub/Sub.
type RedisRealtimeBroker struct {
	client *redis.Client
}

// 3. NewRedisRealtimeBroker creates an optional Redis broadcaster that can recover after startup outages.
func NewRedisRealtimeBroker(ctx context.Context, cfg *config.Config) (RealtimeBroker, error) {
	if cfg == nil || !cfg.RedisEnabled {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	broker := &RedisRealtimeBroker{client: client}
	if err := client.Ping(pingCtx).Err(); err != nil {
		// Keep the client alive: go-redis reconnects on subsequent publish and subscribe calls.
		return broker, err
	}
	return broker, nil
}

// 4. Publish broadcasts one already-persisted event.
func (b *RedisRealtimeBroker) Publish(ctx context.Context, payload []byte) error {
	return b.client.Publish(ctx, realtimeBroadcastChannel, payload).Err()
}

// 5. Subscribe returns a channel for cross-instance events and its close function.
func (b *RedisRealtimeBroker) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	pubsub := b.client.Subscribe(ctx, realtimeBroadcastChannel)
	if err := pubsub.Ping(ctx); err != nil {
		_ = pubsub.Close()
		return nil, nil, err
	}
	result := make(chan []byte)
	go func() {
		defer close(result)
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				select {
				case result <- []byte(message.Payload):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return result, func() { _ = pubsub.Close() }, nil
}

// 6. Close releases the Redis client.
func (b *RedisRealtimeBroker) Close() error {
	return b.client.Close()
}
