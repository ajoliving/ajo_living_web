/*
 * 郵件格式測試。
 * 1. 驗證繁體中文郵件主題使用標準 RFC 2047 編碼。
 * 2. 驗證正文維持 UTF-8 純文字內容。
 */
package service

import (
	"strings"
	"testing"
)

// 1. TestBuildEmailMessageEncodesUnicodeSubject verifies review subjects display correctly.
func TestBuildEmailMessageEncodesUnicodeSubject(t *testing.T) {
	message := string(buildEmailMessage("no-reply@ajo.test", "member@ajo.test", "[AJO Living] 代理資料審核已通過", "審核結果正文"))
	if !strings.Contains(message, "Subject: =?UTF-8?q?") || !strings.Contains(message, "審核結果正文") {
		t.Fatalf("unexpected email message: %q", message)
	}
}
