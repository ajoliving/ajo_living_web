/*
 * iSmart 大廈資料快取服務。
 * 1. 以大廈編號快取共享的大廈資料與文件清單。
 * 2. 合併同一程序內同一大廈的並行回源請求。
 * 3. 快取失敗時直接回源，不影響會員功能與權限校驗。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

const ismartBuildingInfoCacheKeyPrefix = "ajo:ismart:building-info:v1:"

// 1. ismartBuildingInfoCacheEntry stores only shared upstream building data.
type ismartBuildingInfoCacheEntry struct {
	Payload json.RawMessage `json:"payload"`
	Message string          `json:"message,omitempty"`
}

// 2. loadBuildingInfo returns cached data or fetches and fills the cache.
func (s *IsmartExternalService) loadBuildingInfo(ctx context.Context, buildingID string) (*ismartProxyResult, error) {
	cacheKey := ismartBuildingInfoCacheKeyPrefix + strings.TrimSpace(buildingID)
	if result, ok := s.readBuildingInfoCache(ctx, cacheKey); ok {
		return result, nil
	}

	value, err, _ := s.buildingInfoGroup.Do(cacheKey, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.Background(), s.ismartBuildingFetchTimeout())
		defer cancel()

		if result, ok := s.readBuildingInfoCache(fetchCtx, cacheKey); ok {
			return result, nil
		}
		result, fetchErr := s.fetchBuildingInfo(fetchCtx, buildingID)
		if fetchErr != nil {
			return nil, fetchErr
		}
		result.Payload = normalizeBuildingInfoPayload(result.Payload)
		s.writeBuildingInfoCache(fetchCtx, cacheKey, result)
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	result, ok := value.(*ismartProxyResult)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "invalid ismart building cache result")
	}
	return result, nil
}

// 3. fetchBuildingInfo loads one building from the current or legacy iSmart API.
func (s *IsmartExternalService) fetchBuildingInfo(ctx context.Context, buildingID string) (*ismartProxyResult, error) {
	query := url.Values{}
	query.Set("building_id", buildingID)
	result, err := s.getIntegration(ctx, "/buildings/info/", query)
	if err != nil && ismartFallbackAllowed(err, false) {
		return s.postExternal(ctx, "/building-info/", map[string]any{
			"building_id": buildingID,
		})
	}

	return result, err
}

// 4. readBuildingInfoCache decodes one shared building cache entry.
func (s *IsmartExternalService) readBuildingInfoCache(ctx context.Context, cacheKey string) (*ismartProxyResult, bool) {
	if s.runtime.CacheStore == nil {
		return nil, false
	}
	raw, found, err := s.runtime.CacheStore.Get(ctx, cacheKey)
	if err != nil {
		s.logBuildingCacheError("read", cacheKey, err)
		return nil, false
	}
	if !found {
		return nil, false
	}

	var entry ismartBuildingInfoCacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		s.logBuildingCacheError("decode", cacheKey, err)
		return nil, false
	}
	var payload any
	decoder := json.NewDecoder(bytes.NewReader(entry.Payload))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		s.logBuildingCacheError("decode payload", cacheKey, err)
		return nil, false
	}

	return &ismartProxyResult{Payload: payload, Message: entry.Message}, true
}

// 5. writeBuildingInfoCache stores normalized shared data without member metadata.
func (s *IsmartExternalService) writeBuildingInfoCache(ctx context.Context, cacheKey string, result *ismartProxyResult) {
	if s.runtime.CacheStore == nil || result == nil {
		return
	}
	payload, err := json.Marshal(result.Payload)
	if err != nil {
		s.logBuildingCacheError("encode payload", cacheKey, err)
		return
	}
	entry, err := json.Marshal(ismartBuildingInfoCacheEntry{Payload: payload, Message: result.Message})
	if err != nil {
		s.logBuildingCacheError("encode", cacheKey, err)
		return
	}
	if err := s.runtime.CacheStore.Set(ctx, cacheKey, entry, s.ismartBuildingCacheTTL()); err != nil {
		s.logBuildingCacheError("write", cacheKey, err)
	}
}

// 6. ismartBuildingCacheTTL returns the configured five-minute default.
func (s *IsmartExternalService) ismartBuildingCacheTTL() time.Duration {
	if s.runtime.Config != nil && s.runtime.Config.IsmartBuildingCacheTTL > 0 {
		return s.runtime.Config.IsmartBuildingCacheTTL
	}

	return 5 * time.Minute
}

// 7. ismartBuildingFetchTimeout bounds detached cache-fill requests.
func (s *IsmartExternalService) ismartBuildingFetchTimeout() time.Duration {
	if s.runtime.Config != nil && s.runtime.Config.IsmartExternalAppTimeout > 0 {
		return s.runtime.Config.IsmartExternalAppTimeout
	}

	return 10 * time.Second
}

// 8. logBuildingCacheError records cache failures without blocking upstream data.
func (s *IsmartExternalService) logBuildingCacheError(operation string, cacheKey string, err error) {
	if s.runtime.Logger != nil {
		s.runtime.Logger.Warn("ismart building cache unavailable", "operation", operation, "cache_key", cacheKey, "error", err)
	}
}
