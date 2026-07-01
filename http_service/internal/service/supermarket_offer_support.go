/*
 * Supermarket offer support helpers.
 * 1. Normalize public query values.
 * 2. Extract typed values from good-price JSON payloads.
 * 3. Build member-facing fallback and notification values.
 */
package service

import (
	"fmt"
	"strconv"
	"strings"
)

// 1. normalizeSupermarketSort maps empty or unsupported sorts to good-price defaults.
func normalizeSupermarketSort(value string) string {
	switch strings.TrimSpace(value) {
	case "maxPrice", "name", "brand", "diff", "discount", "effective", "minPrice":
		if value == "minPrice" {
			return "effective"
		}
		return strings.TrimSpace(value)
	default:
		return "discount"
	}
}

// 2. boolString returns a lowercase bool string.
func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// 3. intString returns a base-10 integer string.
func intString(value int) string {
	return strconv.Itoa(value)
}

// 4. stringFromMap reads a string field from a JSON object.
func stringFromMap(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

// 5. floatFromMap reads a numeric field from a JSON object.
func floatFromMap(data map[string]any, key string) float64 {
	switch value := data[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case jsonNumber:
		result, _ := strconv.ParseFloat(string(value), 64)
		return result
	default:
		return 0
	}
}

// 6. formatHKPriceForMail formats a price in HKD for plain text email.
func formatHKPriceForMail(value float64) string {
	return fmt.Sprintf("HK$%.2f", value)
}

// 7. supermarketImageObjectKey builds the canonical supermarket image object key.
func supermarketImageObjectKey(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}

	return "ajo_living/supermarket/products/" + strings.ToUpper(code) + ".jpg"
}

type jsonNumber string
