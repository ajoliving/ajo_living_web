/*
 * 超市商品圖片導入命令測試。
 * 1. 驗證來源圖片副檔名與實際 MIME 類型判斷。
 * 2. 驗證商品代號與固定 OSS object key。
 */
package main

import "testing"

// 1. TestSupportedSupermarketImageFiles validates supported source extensions.
func TestSupportedSupermarketImageFiles(t *testing.T) {
	tests := map[string]bool{
		"P000000001.jpg":  true,
		"P000000001.JPEG": true,
		"P000000001.png":  true,
		"P000000001.webp": true,
		"products.xlsx":   false,
	}
	for name, expected := range tests {
		if actual := isSupportedSupermarketImageFile(name); actual != expected {
			t.Fatalf("isSupportedSupermarketImageFile(%q) = %v, want %v", name, actual, expected)
		}
	}
}

// 2. TestDetectSupermarketImageMimeTypeRecognizesWebP validates disguised WebP files.
func TestDetectSupermarketImageMimeTypeRecognizesWebP(t *testing.T) {
	body := []byte("RIFF\x10\x00\x00\x00WEBPVP8 ")
	if actual := detectSupermarketImageMimeType(body); actual != "image/webp" {
		t.Fatalf("mime type = %q, want image/webp", actual)
	}
}

// 3. TestSupermarketImageProductCodeAcceptsPNG validates code extraction from new PNG files.
func TestSupermarketImageProductCodeAcceptsPNG(t *testing.T) {
	code, err := supermarketImageProductCode("/tmp/P000000121.png")
	if err != nil {
		t.Fatalf("supermarketImageProductCode returned error: %v", err)
	}
	if code != "P000000121" {
		t.Fatalf("code = %q, want P000000121", code)
	}
	if objectKey := supermarketProductImageObjectKey(code); objectKey != "ajo_living/supermarket/products/P000000121.jpg" {
		t.Fatalf("object key = %q", objectKey)
	}
}
