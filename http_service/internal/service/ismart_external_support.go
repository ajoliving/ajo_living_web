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
	"errors"
	"io"
	"net/http"
	"net/url"
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
	return s.requestJSON(ctx, http.MethodPost, requestURL, nil, payload)
}

// 4. getIntegration sends one GET request to the iSmart integration API.
func (s *IsmartExternalService) getIntegration(ctx context.Context, path string, query url.Values) (*ismartProxyResult, error) {
	return s.requestJSON(ctx, http.MethodGet, s.integrationURL(path), query, nil)
}

// 5. postIntegration sends one POST request to the iSmart integration API.
func (s *IsmartExternalService) postIntegration(ctx context.Context, path string, payload map[string]any) (*ismartProxyResult, error) {
	return s.requestJSON(ctx, http.MethodPost, s.integrationURL(path), nil, payload)
}

// 6. postIntegrationAccepted posts an OwnerReg request where HTTP 2xx alone means acceptance.
func (s *IsmartExternalService) postIntegrationAccepted(ctx context.Context, path string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.integrationURL(path), bytes.NewReader(raw))
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient().Do(request)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to call ismart service")
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to read ismart response")
	}
	_, decodeErr := decodeIsmartProxyResponse(response.StatusCode, responseBody)
	if decodeErr != nil {
		return decodeErr
	}

	return errcode.New(errcode.CodeInternalError, "ismart service request failed")
}

// 7. requestJSON sends one JSON request to iSmart and decodes the response.
func (s *IsmartExternalService) requestJSON(ctx context.Context, method string, requestURL string, query url.Values, payload any) (*ismartProxyResult, error) {
	var requestBody io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
		}
		requestBody = bytes.NewReader(raw)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}
	if len(query) > 0 {
		values := request.URL.Query()
		for key, entries := range query {
			for _, entry := range entries {
				values.Add(key, entry)
			}
		}
		request.URL.RawQuery = values.Encode()
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := s.httpClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call ismart service")
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read ismart response")
	}

	return decodeIsmartProxyResponse(response.StatusCode, responseBody)
}

// 7. externalURL builds an iSmart external app API URL.
func (s *IsmartExternalService) externalURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com/api/v1/external"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 8. integrationURL builds an iSmart integration API URL.
func (s *IsmartExternalService) integrationURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartIntegrationAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com/api/v1/integration"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 9. rootURL builds an iSmart root API URL.
func (s *IsmartExternalService) rootURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 10. httpClient returns the iSmart HTTP client.
func (s *IsmartExternalService) httpClient() *http.Client {
	timeout := s.runtime.Config.IsmartExternalAppTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 11. ismartProxyResult stores a normalized iSmart response.
type ismartProxyResult struct {
	Payload any
	Message string
}

// 12. decodeIsmartProxyResponse normalizes iSmart success and error payloads.
func decodeIsmartProxyResponse(statusCode int, body []byte) (*ismartProxyResult, error) {
	var payload any = map[string]any{}
	if len(strings.TrimSpace(string(body))) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "invalid ismart service response")
		}
	}

	payloadMap := paymentMapValue(payload)
	message := strings.TrimSpace(paymentStringValue(payloadMap["message"]))
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(ismartHTTPErrorCode(statusCode), message)
	}

	if strings.EqualFold(paymentStringValue(payloadMap["status"]), "error") {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}
	if code := strings.TrimSpace(paymentStringValue(payloadMap["code"])); code != "" && code != "200" && !strings.EqualFold(code, errcode.CodeOK) {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}

	if data, ok := payloadMap["data"]; ok {
		if dataMap := paymentMapValue(data); len(dataMap) > 0 {
			return &ismartProxyResult{Payload: dataMap, Message: message}, nil
		}
		return &ismartProxyResult{Payload: data, Message: message}, nil
	}

	return &ismartProxyResult{Payload: payload, Message: message}, nil
}

// 13. ismartHTTPErrorCode maps upstream status into AJO error codes.
func ismartHTTPErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest, http.StatusConflict:
		return errcode.CodeValidationError
	case http.StatusForbidden:
		return errcode.CodeAuthForbidden
	case http.StatusNotFound:
		return errcode.CodeNotFound
	default:
		return errcode.CodeInternalError
	}
}

// 14. ismartFallbackAllowed keeps write fallback conservative.
func ismartFallbackAllowed(err error, isWrite bool) bool {
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) {
		return false
	}
	if appErr.Code == errcode.CodeNotFound {
		return true
	}
	return !isWrite && appErr.Code == errcode.CodeInternalError
}
