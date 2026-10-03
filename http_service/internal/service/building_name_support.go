/*
 * 大廈顯示名稱解析。
 * 1. 以共享 POS 大廈目錄為主來源，本地社區資料為備援來源。
 * 2. 提供大廈 ID 至中英文名稱的索引與單筆查詢。
 * 3. 名稱缺失或僅為大廈 ID 佔位值時回傳空字串，由呼叫端決定回退。
 */
package service

import (
	"context"
	"strings"

	"ajoliving_web/http_service/internal/model"
)

// 1. buildingDisplayName holds the resolved Chinese and English names for one building.
type buildingDisplayName struct {
	ZH string
	EN string
}

// 2. buildingIDLookupKeys returns lookup keys covering zero padded ID variants.
func buildingIDLookupKeys(value string) []string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil
	}

	keys := []string{raw}
	if !isNumericBuildingID(raw) {
		return keys
	}

	padded := raw
	if len(raw) < 7 {
		padded = strings.Repeat("0", 7-len(raw)) + raw
	}
	trimmed := strings.TrimLeft(raw, "0")
	if trimmed == "" {
		trimmed = "0"
	}
	for _, candidate := range []string{padded, trimmed} {
		if !containsBuildingKey(keys, candidate) {
			keys = append(keys, candidate)
		}
	}
	return keys
}

// 3. isNumericBuildingID reports whether the value is a pure digit identifier.
func isNumericBuildingID(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

// 4. containsBuildingKey reports whether one lookup key already exists.
func containsBuildingKey(keys []string, value string) bool {
	for _, key := range keys {
		if key == value {
			return true
		}
	}
	return false
}

// 5. usableBuildingName rejects empty values and pure building ID placeholders.
func usableBuildingName(name string, keys []string) string {
	value := strings.TrimSpace(name)
	if value == "" {
		return ""
	}
	if containsBuildingKey(keys, value) {
		return ""
	}
	return value
}

// 6. buildingNameIndex builds the name index from the shared POS directory.
func (s *ChatService) buildingNameIndex(ctx context.Context) map[string]buildingDisplayName {
	result := make(map[string]buildingDisplayName)
	buildings, err := NewPOSBuildingService(s.runtime).ListBuildings(ctx)
	if err != nil {
		return result
	}

	for _, item := range buildings {
		buildingID := strings.TrimSpace(item.BuildingID)
		if buildingID == "" {
			buildingID = strings.TrimSpace(item.ID)
		}
		keys := buildingIDLookupKeys(buildingID)
		if len(keys) == 0 {
			continue
		}
		name := buildingDisplayName{
			ZH: usableBuildingName(item.BuildnameCHI, keys),
			EN: usableBuildingName(item.Buildname, keys),
		}
		if name.ZH == "" && name.EN == "" {
			continue
		}
		for _, key := range keys {
			result[key] = name
		}
	}
	return result
}

// 7. lookupBuildingName returns the indexed name for one building ID.
func lookupBuildingName(names map[string]buildingDisplayName, buildingID string) (buildingDisplayName, bool) {
	for _, key := range buildingIDLookupKeys(buildingID) {
		if name, ok := names[key]; ok {
			return name, true
		}
	}
	return buildingDisplayName{}, false
}

// 8. loadCommunityBuildingName resolves one building name from local communities.
func (s *ChatService) loadCommunityBuildingName(ctx context.Context, buildingID string) buildingDisplayName {
	keys := buildingIDLookupKeys(buildingID)
	if len(keys) == 0 {
		return buildingDisplayName{}
	}

	var community model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("public_id IN ?", keys).First(&community).Error; err != nil {
		return buildingDisplayName{}
	}
	return buildingDisplayName{
		ZH: usableBuildingName(community.NameZH, keys),
		EN: usableBuildingName(community.NameEN, keys),
	}
}

// 9. buildingNameFor resolves one building name from the index with local fallback.
func (s *ChatService) buildingNameFor(ctx context.Context, names map[string]buildingDisplayName, buildingID string) buildingDisplayName {
	if name, ok := lookupBuildingName(names, buildingID); ok {
		return name
	}
	return s.loadCommunityBuildingName(ctx, buildingID)
}
