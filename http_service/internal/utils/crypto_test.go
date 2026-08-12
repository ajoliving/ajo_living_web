/*
 * 敏感資料加解密測試。
 * 1. 驗證有效 AES-GCM 密文可正常還原。
 * 2. 驗證空白或過短密文只返回錯誤，不觸發 panic。
 */
package utils

import "testing"

// 1. TestEncryptDecryptString verifies a normal encryption round trip.
func TestEncryptDecryptString(t *testing.T) {
	encrypted, err := EncryptString("test-secret", "sensitive-value")
	if err != nil {
		t.Fatalf("encrypt string: %v", err)
	}
	plain, err := DecryptString("test-secret", encrypted)
	if err != nil {
		t.Fatalf("decrypt string: %v", err)
	}
	if plain != "sensitive-value" {
		t.Fatalf("unexpected plain value %q", plain)
	}
}

// 2. TestDecryptStringRejectsShortPayload verifies malformed ciphertext is handled safely.
func TestDecryptStringRejectsShortPayload(t *testing.T) {
	for _, value := range []string{"", "YQ=="} {
		if _, err := DecryptString("test-secret", value); err == nil {
			t.Fatalf("expected malformed value %q to fail", value)
		}
	}
}
