/*
 * 即時事件 outbox worker。
 * 1. 讀取與消息同一交易提交的待發送事件。
 * 2. 發布成功後標記完成，失敗按退避時間重試。
 * 3. Redis 僅作傳輸層，PostgreSQL 保留可恢復的事件記錄。
 */
package service

import (
	"context"
	"math"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/model"
	"gorm.io/gorm"
)

// 1. RealtimeOutboxWorker publishes durable realtime events.
type RealtimeOutboxWorker struct {
	runtime  *Runtime
	broker   RealtimeBroker
	interval time.Duration
}

// 2. MarkRealtimeEventPublished acknowledges an event sent synchronously by the hub.
func (s *ChatService) MarkRealtimeEventPublished(ctx context.Context, aggregateID string) {
	if s == nil || s.runtime == nil || s.runtime.DB == nil || !s.runtime.DB.Migrator().HasTable(&model.RealtimeOutbox{}) {
		return
	}
	_ = s.runtime.DB.WithContext(ctx).Model(&model.RealtimeOutbox{}).Where("aggregate_id = ? AND status <> ?", strings.TrimSpace(aggregateID), "published").Updates(map[string]any{"status": "published", "published_at": s.runtime.Now(), "last_error": ""}).Error
}

// 3. NewRealtimeOutboxWorker creates a worker with a short polling interval.
func NewRealtimeOutboxWorker(runtime *Runtime, broker RealtimeBroker) *RealtimeOutboxWorker {
	return &RealtimeOutboxWorker{runtime: runtime, broker: broker, interval: time.Second}
}

// 4. Run polls until the process context is cancelled.
func (w *RealtimeOutboxWorker) Run(ctx context.Context) {
	if w == nil || w.runtime == nil || w.broker == nil || w.runtime.DB == nil {
		return
	}
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		w.processBatch(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// 5. processBatch claims and publishes a bounded batch of pending events.
func (w *RealtimeOutboxWorker) processBatch(ctx context.Context) {
	if !w.runtime.DB.Migrator().HasTable(&model.RealtimeOutbox{}) {
		return
	}
	now := w.runtime.Now()
	var rows []model.RealtimeOutbox
	if err := w.runtime.DB.WithContext(ctx).Where("((status IN ? AND next_attempt_at <= ?) OR (status = ? AND next_attempt_at <= ?))", []string{"pending", "failed"}, now, "processing", now).Order("id asc").Limit(50).Find(&rows).Error; err != nil {
		w.warn("load realtime outbox", err)
		return
	}
	for _, row := range rows {
		if !w.claim(ctx, row.ID) {
			continue
		}
		if err := w.broker.Publish(ctx, []byte(row.Payload)); err != nil {
			w.fail(ctx, row, err)
			continue
		}
		publishedAt := w.runtime.Now()
		_ = w.runtime.DB.WithContext(ctx).Model(&model.RealtimeOutbox{}).Where("id = ?", row.ID).Updates(map[string]any{"status": "published", "published_at": publishedAt, "last_error": ""}).Error
	}
}

// 6. claim marks one event as processing to avoid duplicate workers.
func (w *RealtimeOutboxWorker) claim(ctx context.Context, id int64) bool {
	now := w.runtime.Now()
	result := w.runtime.DB.WithContext(ctx).Model(&model.RealtimeOutbox{}).Where("id = ? AND status IN ? AND next_attempt_at <= ?", id, []string{"pending", "failed", "processing"}, now).Updates(map[string]any{"status": "processing", "attempts": gorm.Expr("attempts + ?", 1), "next_attempt_at": now.Add(5 * time.Minute)})
	return result.Error == nil && result.RowsAffected == 1
}

// 7. fail schedules an exponential retry with a bounded delay.
func (w *RealtimeOutboxWorker) fail(ctx context.Context, row model.RealtimeOutbox, publishErr error) {
	attempt := row.Attempts + 1
	delay := time.Duration(math.Pow(2, float64(outboxMinInt(attempt, 8)))) * time.Second
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	next := w.runtime.Now().Add(delay)
	message := strings.TrimSpace(publishErr.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	_ = w.runtime.DB.WithContext(ctx).Model(&model.RealtimeOutbox{}).Where("id = ?", row.ID).Updates(map[string]any{"status": "failed", "next_attempt_at": next, "last_error": message}).Error
	w.warn("publish realtime outbox", publishErr)
}

// 8. warn keeps worker failures observable without stopping message writes.
func (w *RealtimeOutboxWorker) warn(action string, err error) {
	if w.runtime.Logger != nil {
		w.runtime.Logger.Warn(action, "error", err)
	}
}

// 9. outboxMinInt bounds retry exponent without another utility package.
func outboxMinInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
