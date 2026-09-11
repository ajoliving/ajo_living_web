/*
 * iSmart 大廈資料讀取服務。
 * 1. 每次回源取得共享的大廈資料與文件清單。
 * 2. 正規化舊新文件分組後再交給會員可見範圍組裝。
 */
package service

import (
	"context"
	"net/url"
)

// 1. loadBuildingInfo fetches one building and normalizes document groups.
func (s *IsmartExternalService) loadBuildingInfo(ctx context.Context, buildingID string) (*ismartProxyResult, error) {
	result, err := s.fetchBuildingInfo(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	result.Payload = normalizeBuildingInfoPayload(result.Payload)
	return result, nil
}

// 2. fetchBuildingInfo loads one building from the current or legacy iSmart API.
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
