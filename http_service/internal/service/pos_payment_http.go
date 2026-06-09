/*
 * POS 物業繳費 HTTP 代理輔助。
 * 1. 封裝 POS relay 的授權 JSON 請求。
 * 2. 封裝獨立 H5 支付服務 JSON 請求。
 * 3. 統一第三方服務錯誤轉換。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. posRelayJSON sends a JSON request to POS relay.
func (s *POSPaymentService) posRelayJSON(ctx context.Context, method string, path string, token string, query url.Values, payload any, target any) error {
	requestURL := posRelayURL(s.runtime.Config.POSAPIBaseURL, path)
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to prepare pos request")
		}
		body = bytes.NewReader(raw)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare pos request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return s.decodeJSONResponse(request, target, "pos service request failed")
}

// 2. posPaymentServiceJSON sends a JSON request to the H5 payment service.
func (s *POSPaymentService) posPaymentServiceJSON(ctx context.Context, method string, path string, payload any, target any) error {
	requestURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.POSPaymentServiceURL), "/") + "/" + strings.TrimLeft(path, "/")

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to prepare payment request")
		}
		body = bytes.NewReader(raw)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare payment request")
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return s.decodeJSONResponse(request, target, "payment service request failed")
}

// 3. decodeJSONResponse executes an HTTP request and decodes JSON.
func (s *POSPaymentService) decodeJSONResponse(request *http.Request, target any, failureMessage string) error {
	client := &http.Client{Timeout: s.httpTimeout()}
	response, err := client.Do(request)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, failureMessage)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return errcode.New(errcode.CodeAuthRequired, "ismart login is required")
	}
	if response.StatusCode == http.StatusForbidden {
		return errcode.New(errcode.CodeAuthForbidden, "pos access denied")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return errcode.New(errcode.CodeInternalError, failureMessage)
	}
	if target == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return errcode.New(errcode.CodeInternalError, "invalid service response")
	}

	return nil
}

// 4. httpTimeout returns the configured POS timeout.
func (s *POSPaymentService) httpTimeout() time.Duration {
	if s.runtime.Config.POSLoginTimeout > 0 {
		return s.runtime.Config.POSLoginTimeout
	}

	return 10 * time.Second
}
