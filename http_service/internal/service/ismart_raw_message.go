/*
 * iSmart 原始回應安全快照。
 * 1. 保留上游未知業務欄位，以供會員本人查閱。
 * 2. 遞迴移除密碼、token、密鑰與授權資料。
 * 3. 合併註冊與 POS 登入回應，避免缺失欄位覆蓋已保存資料。
 */
package service

import (
	"encoding/json"
	"strings"

	"gorm.io/datatypes"

	"ajoliving_web/http_service/internal/model"
)

// 1. ismartRawResponseSection extracts and sanitizes one upstream response section.
func ismartRawResponseSection(body []byte, key string) map[string]any {
	payload := map[string]any{}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	if decoder.Decode(&payload) != nil {
		return nil
	}
	return sanitizeIsmartRawMessage(paymentMapValue(payload[key]))
}

// 2. ismartRawMessage decodes one persisted raw payload and applies current redaction rules.
func ismartRawMessage(raw datatypes.JSON) map[string]any {
	payload := map[string]any{}
	if json.Unmarshal(raw, &payload) != nil {
		return map[string]any{}
	}
	return sanitizeIsmartRawMessage(payload)
}

// 3. ismartProfileSnapshot returns the existing compatibility snapshot or the legacy raw payload.
func ismartProfileSnapshot(account model.UserIsmartAccount) map[string]any {
	profile := ismartRawMessage(account.ProfileSnapshot)
	if len(profile) > 0 {
		return profile
	}
	return ismartRawMessage(account.RawMessage)
}

// 4. sanitizeIsmartRawMessage removes sensitive values while preserving all business fields.
func sanitizeIsmartRawMessage(payload map[string]any) map[string]any {
	result := make(map[string]any, len(payload))
	for key, value := range payload {
		if ismartSensitiveField(key) {
			continue
		}
		result[key] = sanitizeIsmartRawValue(value)
	}
	return result
}

// 5. sanitizeIsmartRawValue recursively redacts nested objects and arrays.
func sanitizeIsmartRawValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeIsmartRawMessage(typed)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, sanitizeIsmartRawValue(item))
		}
		return result
	default:
		return value
	}
}

// 6. ismartSensitiveField identifies credentials that must never leave AJO.
func ismartSensitiveField(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(strings.TrimSpace(key)))
	for _, marker := range []string{"password", "token", "secret", "authorization", "credential", "apikey", "privatekey", "cookie", "session"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

// 7. mergeIsmartRawMessages overlays available incoming business fields onto the saved snapshot.
func mergeIsmartRawMessages(existing map[string]any, incoming map[string]any) map[string]any {
	result := sanitizeIsmartRawMessage(existing)
	for key, value := range sanitizeIsmartRawMessage(incoming) {
		if ismartMissingRawValue(value) {
			continue
		}
		if current, ok := result[key].(map[string]any); ok {
			if next, ok := value.(map[string]any); ok {
				result[key] = mergeIsmartRawMessages(current, next)
				continue
			}
		}
		result[key] = value
	}
	return result
}

// 8. ismartMissingRawValue keeps existing data when an upstream field is absent, null, or blank.
func ismartMissingRawValue(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}
