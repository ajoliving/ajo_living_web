/*
 * Property content translation service tests.
 * 1. Verify DeepL request authentication, language selection, and field order.
 * 2. Verify translated title and description are returned separately.
 * 3. Verify missing source content is rejected before calling a provider.
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
)

// 1. TestTranslatePropertyContentWithDeepL verifies the normal translation flow.
func TestTranslatePropertyContentWithDeepL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v2/translate" {
			t.Fatalf("unexpected translation request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "DeepL-Auth-Key test-key" {
			t.Fatalf("unexpected authorization header")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("parse translation form: %v", err)
		}
		if request.Form.Get("source_lang") != "ZH" || request.Form.Get("target_lang") != "EN-US" {
			t.Fatalf("unexpected language pair: %v", request.Form)
		}
		texts := request.Form["text"]
		if len(texts) != 2 || texts[0] != "康怡花園高層兩房" || texts[1] != "實用兩房，交通方便。" {
			t.Fatalf("unexpected translation texts: %v", texts)
		}

		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{
			"translations": []map[string]string{
				{"text": "High-floor two-bedroom unit in Kornhill"},
				{"text": "Practical two-bedroom unit with convenient transport."},
			},
		})
	}))
	defer server.Close()

	propertyService := NewPropertyService(&Runtime{Config: &config.Config{
		TranslationProvider:       "deepl",
		DeepLAPIBaseURL:           server.URL,
		DeepLAuthKey:              "test-key",
		TranslationRequestTimeout: time.Second,
	}})
	result, err := propertyService.TranslatePropertyContent(context.Background(), PropertyContentTranslationInput{
		Title:       "康怡花園高層兩房",
		Description: "實用兩房，交通方便。",
	})
	if err != nil {
		t.Fatalf("translate property content: %v", err)
	}
	if result.TitleEn != "High-floor two-bedroom unit in Kornhill" {
		t.Fatalf("unexpected translated title: %q", result.TitleEn)
	}
	if result.DescriptionEn != "Practical two-bedroom unit with convenient transport." {
		t.Fatalf("unexpected translated description: %q", result.DescriptionEn)
	}
}

// 2. TestTranslatePropertyContentRequiresSource verifies empty requests fail early.
func TestTranslatePropertyContentRequiresSource(t *testing.T) {
	propertyService := NewPropertyService(&Runtime{Config: &config.Config{}})
	if _, err := propertyService.TranslatePropertyContent(context.Background(), PropertyContentTranslationInput{}); err == nil {
		t.Fatal("expected empty translation source to fail")
	}
}
