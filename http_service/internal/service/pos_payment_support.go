/*
 * POS 物業繳費輔助函式。
 * 1. 建立 POS relay URL 與會員單位顯示值。
 * 2. 正規化 POS 與 H5 支付服務回傳的鬆散資料。
 * 3. 校驗 H5 訂單與目前會員單位的可見性。
 */
package service

import (
	"errors"
	"regexp"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

var posPermissionDigitsPattern = regexp.MustCompile(`\D+`)

// 1. posRelayURL builds a POS relay URL.
func posRelayURL(baseURL string, path string) string {
	value := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if value == "" {
		value = "https://pos.ismart.skylinedances.com/api"
	}

	return value + "/" + strings.TrimLeft(path, "/")
}

// 2. posUnitMatchesProfile checks floor and unit values against POS unit rows.
func posUnitMatchesProfile(item POSUnitSummary, floor string, unitName string) bool {
	itemFloor := posUnitFloor(item)
	itemUnit := posUnitName(item)
	return strings.EqualFold(itemFloor, strings.TrimSpace(floor)) && strings.EqualFold(itemUnit, strings.TrimSpace(unitName))
}

// 3. posUnitID returns a stable POS unit id.
func posUnitID(item POSUnitSummary) string {
	return strings.TrimSpace(paymentFirstNonEmpty(item.UnitID, item.ID))
}

// 4. posCommunityName returns a member-visible building name.
func posCommunityName(community *model.Community) string {
	if community == nil {
		return ""
	}
	return paymentFirstNonEmpty(community.NameZH, community.NameEN, community.AddressText, community.PublicID)
}

// 5. posPaymentBuildingName returns local or POS building id as display name.
func posPaymentBuildingName(community *model.Community, buildingID string) string {
	name := posCommunityName(community)
	if name != "" {
		return name
	}

	return strings.TrimSpace(buildingID)
}

// 6. posPaymentClientType returns the old POS web client type.
func posPaymentClientType(account *model.UserIsmartAccount) string {
	if account != nil && account.IsStaff {
		return "web_staff"
	}
	return "web_client"
}

// 7. posUnitFloor returns a normalized unit floor.
func posUnitFloor(item POSUnitSummary) string {
	return strings.TrimSpace(item.Floor)
}

// 8. posUnitName returns a normalized unit name.
func posUnitName(item POSUnitSummary) string {
	return strings.TrimSpace(paymentFirstNonEmpty(item.Unit, item.UnitName, item.Name))
}

// 8.1 posBuildingID returns a stable POS building id.
func posBuildingID(item POSBuildingSummary) string {
	return strings.TrimSpace(paymentFirstNonEmpty(item.BuildingID, item.ID))
}

// 9. posUnitLabel returns the display label for one unit.
func posUnitLabel(floor string, unit string) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(floor) != "" {
		parts = append(parts, strings.TrimSpace(floor))
	}
	if strings.TrimSpace(unit) != "" {
		parts = append(parts, strings.TrimSpace(unit))
	}
	return strings.Join(parts, " / ")
}

// 10. containsString checks exact membership after trimming.
func containsString(values []string, target string) bool {
	needle := strings.TrimSpace(target)
	if needle == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == needle {
			return true
		}
	}
	return false
}

// 10.1 posDigitsOnly keeps the numeric part of POS permission codes.
func posDigitsOnly(value string) string {
	return posPermissionDigitsPattern.ReplaceAllString(strings.TrimSpace(value), "")
}

// 10.2 posUnitIDVisibleForPermissions checks unit id against POS permission ids.
func posUnitIDVisibleForPermissions(buildingID string, unitID string, permissions []string) bool {
	if len(permissions) == 0 {
		return true
	}

	normalizedBuildingID := posDigitsOnly(buildingID)
	if len(normalizedBuildingID) > 7 {
		normalizedBuildingID = normalizedBuildingID[:7]
	}
	normalizedUnitID := posDigitsOnly(unitID)
	if len(normalizedUnitID) > 11 {
		normalizedUnitID = normalizedUnitID[:11]
	}
	if normalizedUnitID == "" {
		return false
	}
	compactUnitID := strings.ReplaceAll(normalizedUnitID, "0", "")

	for _, rawPermission := range permissions {
		permission := posDigitsOnly(rawPermission)
		if permission == "" {
			continue
		}
		compactPermission := strings.ReplaceAll(permission, "0", "")
		if permission == normalizedBuildingID ||
			normalizedUnitID == permission ||
			compactUnitID == compactPermission ||
			strings.HasPrefix(normalizedUnitID, permission) ||
			strings.HasPrefix(compactUnitID, compactPermission) {
			return true
		}
	}

	return false
}

// 10.3 posUnitVisibleForPermissions checks one POS unit row against permissions.
func posUnitVisibleForPermissions(buildingID string, item POSUnitSummary, permissions []string) bool {
	unitID := posUnitID(item)
	if unitID != "" && posUnitIDVisibleForPermissions(buildingID, unitID, permissions) {
		return true
	}

	normalizedBuildingID := posDigitsOnly(buildingID)
	if len(normalizedBuildingID) > 7 {
		normalizedBuildingID = normalizedBuildingID[:7]
	}
	floor := posDigitsOnly(posUnitFloor(item))
	unit := posDigitsOnly(posUnitName(item))
	if normalizedBuildingID == "" || floor == "" || unit == "" {
		return false
	}
	if len(floor) > 2 {
		floor = floor[len(floor)-2:]
	}
	if len(unit) > 2 {
		unit = unit[len(unit)-2:]
	}
	code := normalizedBuildingID + leftPadPOSCode(floor, 2) + leftPadPOSCode(unit, 2)
	return posUnitIDVisibleForPermissions(buildingID, code, permissions)
}

// 10.4 posUnitsFromFlatUnitPermissions builds unit rows from permission ids.
func posUnitsFromFlatUnitPermissions(buildingID string, permissions []string) []POSUnitSummary {
	normalizedBuildingID := posDigitsOnly(buildingID)
	if len(normalizedBuildingID) > 7 {
		normalizedBuildingID = normalizedBuildingID[:7]
	}
	if normalizedBuildingID == "" || len(permissions) == 0 {
		return []POSUnitSummary{}
	}

	seen := map[string]struct{}{}
	units := make([]POSUnitSummary, 0)
	for _, rawPermission := range permissions {
		permission := posDigitsOnly(rawPermission)
		if permission == "" {
			continue
		}

		unitID := ""
		if strings.HasPrefix(permission, normalizedBuildingID) && len(permission) >= 11 {
			unitID = permission[:11]
		} else if strings.HasPrefix(permission, normalizedBuildingID[:minInt(len(normalizedBuildingID), 6)]) && len(permission) > 6 {
			suffix := permission[6:]
			if len(suffix) > 0 && len(suffix) <= 4 {
				unitID = normalizedBuildingID + leftPadPOSCode(suffix, 4)
			}
		}
		if unitID == "" || !strings.HasPrefix(unitID, normalizedBuildingID) {
			continue
		}
		if _, ok := seen[unitID]; ok {
			continue
		}
		seen[unitID] = struct{}{}

		floor := unitID[7:9]
		unit := strings.TrimLeft(unitID[9:11], "0")
		if unit == "" {
			unit = unitID[9:11]
		}
		if floor == "00" {
			floor = ""
		}
		units = append(units, POSUnitSummary{
			UnitID:   unitID,
			Floor:    floor,
			Unit:     unit,
			UnitName: unit,
		})
	}

	return units
}

// 10.5 posBuildingSummariesFromIDs builds fallback building rows.
func posBuildingSummariesFromIDs(buildingIDs []string) []POSBuildingSummary {
	rows := make([]POSBuildingSummary, 0, len(buildingIDs))
	for _, buildingID := range buildingIDs {
		value := strings.TrimSpace(buildingID)
		if value == "" {
			continue
		}
		rows = append(rows, POSBuildingSummary{
			BuildingID: value,
			Buildname:  value,
			ID:         value,
			Name:       value,
		})
	}
	return rows
}

// 10.6 leftPadPOSCode pads POS numeric permission fragments.
func leftPadPOSCode(value string, size int) string {
	value = posDigitsOnly(value)
	for len(value) < size {
		value = "0" + value
	}
	if len(value) > size {
		return value[len(value)-size:]
	}
	return value
}

// 10.7 minInt returns the smaller integer.
func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

// 11. isPOSAuthRequiredError reports whether a POS relay call needs token refresh.
func isPOSAuthRequiredError(err error) bool {
	var appErr *errcode.AppError
	return errors.As(err, &appErr) && appErr.Code == errcode.CodeAuthRequired
}

// 12. normalizePOSRows extracts a generic row list from a POS payload.
func normalizePOSRows(payload map[string]any, key string) []map[string]any {
	if payload == nil {
		return []map[string]any{}
	}
	if rows, ok := anyToMapRows(payload[key]); ok {
		return rows
	}
	nested := paymentMapValue(payload[key])
	if rows, ok := anyToMapRows(nested[key]); ok {
		return rows
	}
	if rows, ok := anyToMapRows(payload["data"]); ok {
		return rows
	}
	return []map[string]any{}
}

// 13. normalizePOSAccountingRows extracts cash and cheque accounting rows.
func normalizePOSAccountingRows(payload map[string]any) []map[string]any {
	rows := normalizePOSRows(payload, "payment_objs")
	rows = append(rows, normalizePOSRows(payload, "payment_objs_cash")...)
	rows = append(rows, normalizePOSRows(payload, "payment_objs_cheque")...)
	if len(rows) > 0 {
		return rows
	}
	if direct, ok := anyToMapRows(payload); ok {
		return direct
	}
	return []map[string]any{}
}

// 14. normalizePaymentServiceRows extracts order snapshots from H5 payment service.
func normalizePaymentServiceRows(payload map[string]any) []map[string]any {
	if rows, ok := anyToMapRows(payload["data"]); ok {
		return rows
	}
	return normalizePOSRows(payload, "items")
}

// 15. anyToMapRows converts loose JSON list values into map rows.
func anyToMapRows(value any) ([]map[string]any, bool) {
	switch typed := value.(type) {
	case []map[string]any:
		return typed, true
	case []any:
		rows := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if row := paymentMapValue(item); len(row) > 0 {
				rows = append(rows, row)
			}
		}
		return rows, true
	case map[string]any:
		if nested, ok := anyToMapRows(typed["items"]); ok {
			return nested, true
		}
		if nested, ok := anyToMapRows(typed["records"]); ok {
			return nested, true
		}
		if nested, ok := anyToMapRows(typed["payment_objs"]); ok {
			return nested, true
		}
		return []map[string]any{typed}, true
	default:
		return nil, false
	}
}

// 16. h5OrderContextIDs reads POS unit context from one H5 order payload.
func h5OrderContextIDs(payload map[string]any) (string, string) {
	data := paymentMapValue(payload["data"])
	if len(data) == 0 {
		data = payload
	}
	requestPayload := paymentMapValue(data["request_payload"])
	reportPayload := paymentMapValue(requestPayload["report_payment_data"])
	buildingID := paymentFirstNonEmpty(
		paymentStringValue(requestPayload["blg_id"]),
		paymentStringValue(requestPayload["building_id"]),
		paymentStringValue(requestPayload["BLG_ID"]),
		paymentStringValue(reportPayload["BLG_ID"]),
		paymentStringValue(data["blg_id"]),
		paymentStringValue(data["building_id"]),
		paymentStringValue(data["BLG_ID"]),
	)
	unitID := paymentFirstNonEmpty(
		paymentStringValue(requestPayload["unit_id"]),
		paymentStringValue(requestPayload["UNIT_ID"]),
		paymentStringValue(reportPayload["UNIT_ID"]),
		paymentStringValue(data["unit_id"]),
		paymentStringValue(data["UNIT_ID"]),
	)
	return buildingID, unitID
}

// 17. h5OrderHasContext checks whether the order response contains unit context.
func h5OrderHasContext(payload map[string]any) bool {
	buildingID, unitID := h5OrderContextIDs(payload)
	return strings.TrimSpace(buildingID) != "" || strings.TrimSpace(unitID) != ""
}

// 18. h5OrderMatchesContext checks whether one H5 order belongs to the current unit.
func h5OrderMatchesContext(payload map[string]any, contextValue *POSUnitContext) bool {
	if contextValue == nil {
		return false
	}
	buildingID, unitID := h5OrderContextIDs(payload)
	if strings.TrimSpace(unitID) == "" || strings.TrimSpace(unitID) != contextValue.UnitID {
		return false
	}
	if strings.TrimSpace(buildingID) != "" && strings.TrimSpace(buildingID) != contextValue.BuildingID {
		return false
	}
	return true
}
