/*
 * iCCTV 視像監控代理 URL 處理。
 * 1. 將 Orange Pi 明文 FRP 地址改寫為 AJO 可嵌入的 HTTPS 入口。
 * 2. 限制只處理 iCCTV Orange Pi auth 服務端口。
 * 3. 保留非目標 URL 原樣返回。
 */
package service

import (
	"net/url"
	"strconv"
	"strings"
)

// 1. icctvProxyURL rewrites FRP camera URLs to the public HTTPS proxy.
func (s *SecurityICCTVService) icctvProxyURL(rawURL string) string {
	proxyBaseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.ICCTVStreamProxyBaseURL), "/")
	if proxyBaseURL == "" {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return rawURL
	}
	parsedPort := parsed.Port()
	if parsedPort == "" || !icctvStreamPortAllowed(parsedPort) {
		return rawURL
	}

	proxyBase, err := url.Parse(proxyBaseURL)
	if err != nil || proxyBase.Scheme == "" || proxyBase.Host == "" {
		return rawURL
	}
	proxyPath := strings.TrimRight(proxyBase.EscapedPath(), "/")
	cameraPath := strings.TrimLeft(parsed.EscapedPath(), "/")
	if cameraPath == "" {
		return rawURL
	}

	return proxyBase.Scheme + "://" + proxyBase.Host + proxyPath + "/opi/" + parsedPort + "/" + cameraPath + icctvRawQuery(parsed.RawQuery)
}

// 2. icctvStreamPortAllowed checks the configured FRP stream port range.
func icctvStreamPortAllowed(portValue string) bool {
	port, err := strconv.Atoi(portValue)
	if err != nil {
		return false
	}

	return port >= 29000 && port <= 29999
}

// 3. icctvRawQuery returns one URL query suffix.
func icctvRawQuery(rawQuery string) string {
	value := strings.TrimSpace(rawQuery)
	if value == "" {
		return ""
	}

	return "?" + value
}
