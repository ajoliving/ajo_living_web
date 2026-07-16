/*
 * iSmart 原始回應安全快照測試。
 * 1. 驗證遞迴敏感欄位清洗。
 * 2. 驗證 POS 回應補充欄位時保留註冊資料。
 */
package service

import "testing"

// 1. TestIsmartRawMessageSanitizesAndMerges verifies safe raw payload retention.
func TestIsmartRawMessageSanitizesAndMerges(t *testing.T) {
	registered := map[string]any{
		"username":      "200336",
		"owner_name_en": "CHAN TAI MAN",
		"password":      "generated-password",
		"profile": map[string]any{
			"billing_email": "owner@example.com",
			"authorization": "Bearer hidden",
		},
	}
	posLogin := map[string]any{
		"username":      "200336",
		"owner_name_en": "",
		"building":      "BLG-001",
		"access_token":  "hidden",
		"items":         []any{map[string]any{"secret": "hidden", "unit": "01"}},
	}

	result := mergeIsmartRawMessages(registered, posLogin)
	if result["owner_name_en"] != "CHAN TAI MAN" || result["building"] != "BLG-001" {
		t.Fatalf("expected registered and POS business fields, got %#v", result)
	}
	if _, exists := result["password"]; exists {
		t.Fatalf("password must be removed: %#v", result)
	}
	if _, exists := result["access_token"]; exists {
		t.Fatalf("token must be removed: %#v", result)
	}
	profile := result["profile"].(map[string]any)
	if _, exists := profile["authorization"]; exists {
		t.Fatalf("nested authorization must be removed: %#v", profile)
	}
	item := result["items"].([]any)[0].(map[string]any)
	if _, exists := item["secret"]; exists || item["unit"] != "01" {
		t.Fatalf("nested array item must retain only business fields: %#v", item)
	}
}
