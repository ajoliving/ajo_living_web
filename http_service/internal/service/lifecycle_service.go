/*
 * Lifecycle and scheduled task business logic.
 * 1. Expire active listings after their deadline.
 * 2. Provide reusable task entrypoints for background workers.
 */
package service

import (
	"context"

	"ajoliving_web/http_service/internal/model"
)

// 1. LifecycleService handles scheduled listing lifecycle updates.
type LifecycleService struct {
	runtime *Runtime
}

// 2. NewLifecycleService creates a lifecycle service instance.
func NewLifecycleService(runtime *Runtime) *LifecycleService {
	return &LifecycleService{runtime: runtime}
}

// 3. ExpireListings marks overdue active listings as expired.
func (s *LifecycleService) ExpireListings(ctx context.Context) (int64, error) {
	result := s.runtime.DB.WithContext(ctx).
		Model(&model.Listing{}).
		Where("publication_status = ? AND expire_at IS NOT NULL AND expire_at < ? AND is_deleted = ?", "active", s.runtime.Now(), false).
		Updates(map[string]any{
			"publication_status": "expired",
			"updated_at":         s.runtime.Now(),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
