/*
 * Good-price API client.
 * 1. Fetch public supermarket price data from the deployed good-price service.
 * 2. Apply short in-memory cache for read-heavy public endpoints.
 * 3. Keep external API errors converted into AJO service errors.
 */
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. SupermarketOfferService handles supermarket offer integration.
type SupermarketOfferService struct {
	runtime *Runtime
	cacheMu sync.Mutex
	cache   map[string]supermarketCacheEntry
}

// 2. NewSupermarketOfferService creates a supermarket offer service.
func NewSupermarketOfferService(runtime *Runtime) *SupermarketOfferService {
	return &SupermarketOfferService{
		runtime: runtime,
		cache:   make(map[string]supermarketCacheEntry),
	}
}

// 3. fetchGoodPriceJSON loads one JSON object from the deployed good-price API.
func (s *SupermarketOfferService) fetchGoodPriceJSON(ctx context.Context, path string, query url.Values, useCache bool) (map[string]any, error) {
	requestURL := s.goodPriceURL(path, query)
	cacheTTL := s.goodPriceCacheTTL(path)
	if useCache && cacheTTL > 0 {
		if payload, ok := s.cachedPayload(requestURL); ok {
			return decodeGoodPricePayload(payload)
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare good-price request")
	}

	response, err := s.goodPriceHTTPClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call good-price service")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read good-price response")
	}
	if response.StatusCode == http.StatusNotFound {
		return nil, errcode.New(errcode.CodeNotFound, "product not found")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errcode.New(errcode.CodeInternalError, "good-price service request failed")
	}
	if useCache && cacheTTL > 0 {
		s.saveCachedPayload(requestURL, body, cacheTTL)
	}

	return decodeGoodPricePayload(body)
}

// 4. goodPriceURL builds one deployed good-price API URL.
func (s *SupermarketOfferService) goodPriceURL(path string, query url.Values) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.GoodPriceAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://good.price.skylinedances.com/api"
	}
	requestURL := baseURL + "/" + strings.TrimLeft(path, "/")
	if encoded := query.Encode(); encoded != "" {
		requestURL += "?" + encoded
	}

	return requestURL
}

// 5. goodPriceHTTPClient returns an HTTP client with configured timeout.
func (s *SupermarketOfferService) goodPriceHTTPClient() *http.Client {
	timeout := s.runtime.Config.GoodPriceRequestTimeout
	if timeout <= 0 {
		timeout = s.runtime.Config.POSLoginTimeout
	}
	return &http.Client{Timeout: timeout}
}

// 6. cachedPayload returns a non-expired cached payload.
func (s *SupermarketOfferService) cachedPayload(key string) ([]byte, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	entry, ok := s.cache[key]
	if !ok || s.runtime.Now().After(entry.expiresAt) {
		delete(s.cache, key)
		return nil, false
	}

	return append([]byte(nil), entry.payload...), true
}

// 7. saveCachedPayload stores one raw JSON payload.
func (s *SupermarketOfferService) saveCachedPayload(key string, payload []byte, ttl time.Duration) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache[key] = supermarketCacheEntry{
		expiresAt: s.runtime.Now().Add(ttl),
		payload:   append([]byte(nil), payload...),
	}
}

// 8. goodPriceCacheTTL returns a path-specific cache TTL.
func (s *SupermarketOfferService) goodPriceCacheTTL(path string) time.Duration {
	normalized := "/" + strings.TrimLeft(path, "/")
	switch {
	case strings.HasPrefix(normalized, "/summary"):
		return s.runtime.Config.GoodPriceSummaryCacheTTL
	case strings.HasPrefix(normalized, "/search"):
		return s.runtime.Config.GoodPriceSearchCacheTTL
	case strings.HasPrefix(normalized, "/products/"):
		return s.runtime.Config.GoodPriceDetailCacheTTL
	default:
		return s.runtime.Config.GoodPriceCacheTTL
	}
}

// 9. decodeGoodPricePayload decodes one object response from good-price.
func decodeGoodPricePayload(payload []byte) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, fmt.Sprintf("invalid good-price response: %v", err))
	}

	return data, nil
}
