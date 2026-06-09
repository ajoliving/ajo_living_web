/*
 * EasyLink payment helper utilities.
 * 1. Normalize loose gateway payload values.
 * 2. Sign and verify EasyLink MD5 signatures.
 * 3. Keep wallet recharge payment state mapping aligned with POS Web.
 */
package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const (
	PaymentStatePaying   = "PAYING"
	PaymentStateSuccess  = "SUCCESS"
	PaymentStateFailed   = "FAILED"
	PaymentStateClosed   = "CLOSED"
	PaymentStateExpired  = "EXPIRED"
	PaymentStateRevoked  = "REVOKED"
	PaymentStateRefunded = "REFUNDED"
)

// 1. paymentMapValue normalizes object-like values.
func paymentMapValue(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if typed, ok := value.(map[string]any); ok {
		return paymentCloneMap(typed)
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		return map[string]any{}
	}

	return result
}

// 2. paymentCloneMap copies a JSON-compatible map.
func paymentCloneMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return map[string]any{}
	}

	raw, err := json.Marshal(source)
	if err != nil {
		result := make(map[string]any, len(source))
		for key, value := range source {
			result[key] = value
		}
		return result
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		result = make(map[string]any, len(source))
		for key, value := range source {
			result[key] = value
		}
	}

	return result
}

// 3. paymentStringValue converts gateway values to strings for signing.
func paymentStringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		if math.Trunc(typed) == typed {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		if math.Trunc(float64(typed)) == float64(typed) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(typed)
	}
}

// 4. paymentIntValue converts gateway values to ints.
func paymentIntValue(value any) int {
	return int(paymentInt64Value(value))
}

// 5. paymentInt64Value converts gateway values to int64s.
func paymentInt64Value(value any) int64 {
	switch typed := value.(type) {
	case nil:
		return 0
	case int64:
		return typed
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	case float32:
		return int64(typed)
	case json.Number:
		numeric, _ := typed.Int64()
		return numeric
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0
		}
		if numeric, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return numeric
		}
		if numeric, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return int64(numeric)
		}
	}

	return 0
}

// 6. paymentFirstNonNil returns the first non-nil value.
func paymentFirstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}

	return nil
}

// 7. paymentFirstNonEmpty returns the first non-empty string.
func paymentFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// 8. paymentIsEmptyValue returns whether a value is excluded from signing.
func paymentIsEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}

	return false
}

// 9. paymentSignPayload creates an EasyLink MD5 signature.
func paymentSignPayload(payload map[string]any, appSecret string) string {
	signString := paymentToSignString(payload) + "&key=" + appSecret
	hash := md5.Sum([]byte(signString))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// 10. paymentVerifySignature verifies an EasyLink MD5 signature.
func paymentVerifySignature(payload map[string]any, sign string, appSecret string) bool {
	if strings.TrimSpace(sign) == "" {
		return false
	}

	return paymentSignPayload(payload, appSecret) == strings.ToUpper(strings.TrimSpace(sign))
}

// 11. paymentToSignString serializes fields according to EasyLink rules.
func paymentToSignString(payload map[string]any) string {
	keys := make([]string, 0, len(payload))
	for key, value := range payload {
		if key == "sign" || paymentIsEmptyValue(value) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+paymentStringValue(payload[key]))
	}
	return strings.Join(parts, "&")
}

// 12. normalizePaymentGatewayState maps EasyLink numeric state to local state.
func normalizePaymentGatewayState(stateCode int) string {
	switch stateCode {
	case 2:
		return PaymentStateSuccess
	case 3:
		return PaymentStateFailed
	case 4:
		return PaymentStateRevoked
	case 5:
		return PaymentStateRefunded
	case 6:
		return PaymentStateClosed
	default:
		return PaymentStatePaying
	}
}
