/*
 * 超市優惠服務測試。
 * 1. 驗證 Good Price 上游不可用時公開摘要維持可渲染空態。
 * 2. 驗證 Good Price 上游不可用時公開搜尋維持可渲染空態。
 */
package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
)

// 1. TestSupermarketPublicEndpointsFallbackWhenGoodPriceUnavailable verifies public fallback payloads.
func TestSupermarketPublicEndpointsFallbackWhenGoodPriceUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	supermarketService := NewSupermarketOfferService(&Runtime{
		Config: &config.Config{
			GoodPriceAPIBaseURL:     server.URL,
			GoodPriceRequestTimeout: time.Second,
		},
		Now: time.Now,
	})

	summary, err := supermarketService.Summary(context.Background())
	if err != nil {
		t.Fatalf("expected summary fallback without error, got %v", err)
	}
	if summary["offers"] == nil || summary["stats"] == nil {
		t.Fatalf("expected renderable summary fallback, got %#v", summary)
	}

	search, err := supermarketService.Search(context.Background(), SupermarketSearchFilters{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("expected search fallback without error, got %v", err)
	}
	if search["total"] != 0 || search["page"] != 2 || search["pageSize"] != 5 {
		t.Fatalf("expected paginated empty search fallback, got %#v", search)
	}
}
