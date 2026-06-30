/*
 * iSmart 對外介面代理支援函式。
 * 1. 建立 iSmart external app 與 root API 請求地址。
 * 2. 發送 JSON 請求並解析舊系統回應。
 * 3. 將舊系統 HTTP 與業務錯誤轉為 AJO 錯誤碼。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. postExternal posts to the iSmart external app API base URL.
func (s *IsmartExternalService) postExternal(ctx context.Context, path string, payload map[string]any) (*ismartProxyResult, error) {
	return s.postJSON(ctx, s.externalURL(path), payload)
}

// 2. postRoot posts to an iSmart root API path outside the external app base path.
func (s *IsmartExternalService) postRoot(ctx context.Context, path string, payload map[string]any) (*ismartProxyResult, error) {
	return s.postJSON(ctx, s.rootURL(path), payload)
}

// 3. postJSON sends one JSON request to iSmart and decodes the response.
func (s *IsmartExternalService) postJSON(ctx context.Context, requestURL string, payload map[string]any) (*ismartProxyResult, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(raw))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call ismart service")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read ismart response")
	}

	return decodeIsmartProxyResponse(response.StatusCode, body)
}

// 4. externalURL builds an iSmart external app API URL.
func (s *IsmartExternalService) externalURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com/api/v1/external"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 5. rootURL builds an iSmart root API URL.
func (s *IsmartExternalService) rootURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 6. httpClient returns the iSmart HTTP client.
func (s *IsmartExternalService) httpClient() *http.Client {
	timeout := s.runtime.Config.IsmartExternalAppTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 7. ismartProxyResult stores a normalized iSmart response.
type ismartProxyResult struct {
	Payload map[string]any
	Message string
}

// 8. decodeIsmartProxyResponse normalizes iSmart success and error payloads.
func decodeIsmartProxyResponse(statusCode int, body []byte) (*ismartProxyResult, error) {
	payload := map[string]any{}
	if len(strings.TrimSpace(string(body))) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "invalid ismart service response")
		}
	}

	message := strings.TrimSpace(paymentStringValue(payload["message"]))
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(ismartHTTPErrorCode(statusCode), message)
	}

	if strings.EqualFold(paymentStringValue(payload["status"]), "error") {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}
	if code := strings.TrimSpace(paymentStringValue(payload["code"])); code != "" && code != "200" && !strings.EqualFold(code, errcode.CodeOK) {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}

	if data, ok := payload["data"]; ok {
		if dataMap := paymentMapValue(data); len(dataMap) > 0 {
			return &ismartProxyResult{Payload: dataMap, Message: message}, nil
		}
		return &ismartProxyResult{Payload: map[string]any{"result": data}, Message: message}, nil
	}

	return &ismartProxyResult{Payload: payload, Message: message}, nil
}

// 9. ismartHTTPErrorCode maps upstream status into AJO error codes.
func ismartHTTPErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return errcode.CodeValidationError
	case http.StatusForbidden:
		return errcode.CodeAuthForbidden
	case http.StatusNotFound:
		return errcode.CodeNotFound
	default:
		return errcode.CodeInternalError
	}
}
