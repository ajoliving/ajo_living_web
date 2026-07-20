/*
 * POS 大廈與單位目錄快取服務。
 * 1. 以固定 key 快取共用大廈目錄及按大廈拆分的單位目錄。
 * 2. 合併同一程序內相同目錄的並行回源請求。
 * 3. Redis 失敗時直接回源 POS，不影響會員權限校驗。
 */
package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

const (
	posBuildingDirectoryCacheKey   = "ajo:pos:buildings:v1"
	posBuildingUnitsCacheKeyPrefix = "ajo:pos:building-units:v1:"
)

// 1. loadPOSBuildings returns the shared building directory from cache or POS.
func (s *POSBuildingService) loadPOSBuildings(ctx context.Context) ([]POSBuildingSummary, error) {
	var cached []POSBuildingSummary
	if s.readPOSDirectoryCache(ctx, posBuildingDirectoryCacheKey, &cached) {
		return cached, nil
	}

	value, err, _ := s.directoryGroup.Do(posBuildingDirectoryCacheKey, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.Background(), s.posDirectoryFetchTimeout())
		defer cancel()
		var secondCached []POSBuildingSummary
		if s.readPOSDirectoryCache(fetchCtx, posBuildingDirectoryCacheKey, &secondCached) {
			return secondCached, nil
		}
		rows, fetchErr := s.fetchPOSBuildings(fetchCtx)
		if fetchErr != nil {
			return nil, fetchErr
		}
		s.writePOSDirectoryCache(fetchCtx, posBuildingDirectoryCacheKey, rows)
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	rows, ok := value.([]POSBuildingSummary)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "invalid pos building cache result")
	}
	return rows, nil
}

// 2. loadPOSUnits returns one building unit directory from cache or POS.
func (s *POSBuildingService) loadPOSUnits(ctx context.Context, buildingID string) ([]POSUnitSummary, error) {
	cacheKey := posBuildingUnitsCacheKeyPrefix + strings.TrimSpace(buildingID)
	var cached []POSUnitSummary
	if s.readPOSDirectoryCache(ctx, cacheKey, &cached) {
		return cached, nil
	}

	value, err, _ := s.directoryGroup.Do(cacheKey, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.Background(), s.posDirectoryFetchTimeout())
		defer cancel()
		var secondCached []POSUnitSummary
		if s.readPOSDirectoryCache(fetchCtx, cacheKey, &secondCached) {
			return secondCached, nil
		}
		rows, fetchErr := s.fetchPOSUnits(fetchCtx, buildingID)
		if fetchErr != nil {
			return nil, fetchErr
		}
		s.writePOSDirectoryCache(fetchCtx, cacheKey, rows)
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	rows, ok := value.([]POSUnitSummary)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "invalid pos unit cache result")
	}
	return rows, nil
}

// 3. fetchPOSBuildings loads the uncached public building directory.
func (s *POSBuildingService) fetchPOSBuildings(ctx context.Context) ([]POSBuildingSummary, error) {
	var rows []POSBuildingSummary
	if err := s.getPOS(ctx, "/building", &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// 4. fetchPOSUnits loads the uncached public unit directory for one building.
func (s *POSBuildingService) fetchPOSUnits(ctx context.Context, buildingID string) ([]POSUnitSummary, error) {
	var rows []POSUnitSummary
	path := "/building/" + url.PathEscape(strings.TrimSpace(buildingID)) + "/units"
	if err := s.getPOS(ctx, path, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// 5. readPOSDirectoryCache decodes one shared directory cache value.
func (s *POSBuildingService) readPOSDirectoryCache(ctx context.Context, cacheKey string, target any) bool {
	if s.runtime.CacheStore == nil {
		return false
	}
	raw, found, err := s.runtime.CacheStore.Get(ctx, cacheKey)
	if err != nil {
		s.logPOSDirectoryCacheError("read", cacheKey, err)
		return false
	}
	if !found {
		return false
	}
	if err := json.Unmarshal(raw, target); err != nil {
		s.logPOSDirectoryCacheError("decode", cacheKey, err)
		return false
	}
	return true
}

// 6. writePOSDirectoryCache stores shared directory metadata with an expiry.
func (s *POSBuildingService) writePOSDirectoryCache(ctx context.Context, cacheKey string, value any) {
	if s.runtime.CacheStore == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		s.logPOSDirectoryCacheError("encode", cacheKey, err)
		return
	}
	if err := s.runtime.CacheStore.Set(ctx, cacheKey, raw, s.posDirectoryCacheTTL()); err != nil {
		s.logPOSDirectoryCacheError("write", cacheKey, err)
	}
}

// 7. posDirectoryCacheTTL returns the configured five-minute default.
func (s *POSBuildingService) posDirectoryCacheTTL() time.Duration {
	if s.runtime.Config != nil && s.runtime.Config.POSDirectoryCacheTTL > 0 {
		return s.runtime.Config.POSDirectoryCacheTTL
	}
	return 5 * time.Minute
}

// 8. posDirectoryFetchTimeout bounds detached cache-fill requests.
func (s *POSBuildingService) posDirectoryFetchTimeout() time.Duration {
	if s.runtime.Config != nil && s.runtime.Config.POSLoginTimeout > 0 {
		return s.runtime.Config.POSLoginTimeout
	}
	return 10 * time.Second
}

// 9. logPOSDirectoryCacheError records cache failures without blocking POS reads.
func (s *POSBuildingService) logPOSDirectoryCacheError(operation string, cacheKey string, err error) {
	if s.runtime.Logger != nil {
		s.runtime.Logger.Warn("pos directory cache unavailable", "operation", operation, "cache_key", cacheKey, "error", err)
	}
}
