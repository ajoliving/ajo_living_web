/*
 * 受控資料快照重加密工具測試。
 * 1. 驗證全部應用加密欄位使用新密鑰保存。
 * 2. 驗證空白欄位維持空白且不計入更新數量。
 */
package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/utils"
)

// 1. TestRekeyDatabase verifies atomic re-encryption across managed tables.
func TestRekeyDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:rekey-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	statements := []string{
		`CREATE TABLE listing_contacts (listing_id INTEGER PRIMARY KEY, phone_encrypted TEXT, phone2_encrypted TEXT, whats_app_encrypted TEXT, we_chat_encrypted TEXT, email_encrypted TEXT, updated_at DATETIME)`,
		`CREATE TABLE user_credentials (user_id INTEGER PRIMARY KEY, password_encrypted TEXT, updated_at DATETIME)`,
		`CREATE TABLE user_ismart_accounts (user_id INTEGER PRIMARY KEY, relay_token_encrypted TEXT, password_encrypted TEXT, updated_at DATETIME)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}

	fromKey := "source-key"
	toKey := "testing-key"
	encrypted := func(value string) string {
		result, encryptErr := utils.EncryptString(fromKey, value)
		if encryptErr != nil {
			t.Fatalf("encrypt fixture: %v", encryptErr)
		}
		return result
	}
	encryptedWithKey := func(key string, value string) string {
		result, encryptErr := utils.EncryptString(key, value)
		if encryptErr != nil {
			t.Fatalf("encrypt fixture: %v", encryptErr)
		}
		return result
	}
	if err := db.Exec(`INSERT INTO listing_contacts VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		1,
		encrypted("61234567"),
		"",
		encrypted("61234567"),
		encrypted("listing-wechat"),
		encrypted("listing@example.com"),
	).Error; err != nil {
		t.Fatalf("insert listing contact: %v", err)
	}
	if err := db.Exec(`INSERT INTO listing_contacts VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		2,
		encryptedWithKey(toKey, "68889999"),
		"",
		"",
		"",
		"",
	).Error; err != nil {
		t.Fatalf("insert testing-key listing contact: %v", err)
	}
	if err := db.Exec(`INSERT INTO user_credentials VALUES (?, ?, CURRENT_TIMESTAMP)`, 1, encrypted("credential")).Error; err != nil {
		t.Fatalf("insert user credential: %v", err)
	}
	if err := db.Exec(`INSERT INTO user_ismart_accounts VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, 1, encrypted("relay-token"), encrypted("upstream-password")).Error; err != nil {
		t.Fatalf("insert ismart account: %v", err)
	}

	updated, err := rekeyDatabase(db, fromKey, toKey)
	if err != nil {
		t.Fatalf("rekey database: %v", err)
	}
	if updated != 7 {
		t.Fatalf("expected 7 rekeyed values, got %d", updated)
	}

	var contact encryptedListingContact
	if err := db.Raw(`SELECT listing_id, phone_encrypted, phone2_encrypted, whats_app_encrypted, we_chat_encrypted, email_encrypted FROM listing_contacts WHERE listing_id = 1`).Scan(&contact).Error; err != nil {
		t.Fatalf("load rekeyed contact: %v", err)
	}
	if contact.Phone2Encrypted.String != "" {
		t.Fatalf("expected empty secondary phone, got %q", contact.Phone2Encrypted.String)
	}
	for label, item := range map[string]struct {
		value string
		want  string
	}{
		"phone":    {contact.PhoneEncrypted.String, "61234567"},
		"whatsapp": {contact.WhatsAppEncrypted.String, "61234567"},
		"wechat":   {contact.WeChatEncrypted.String, "listing-wechat"},
		"email":    {contact.EmailEncrypted.String, "listing@example.com"},
	} {
		plain, decryptErr := utils.DecryptString(toKey, item.value)
		if decryptErr != nil || plain != item.want {
			t.Fatalf("unexpected %s after rekey: plain=%q err=%v", label, plain, decryptErr)
		}
	}
}
