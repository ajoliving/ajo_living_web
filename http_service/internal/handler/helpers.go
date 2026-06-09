/*
 * HTTP handler helpers.
 * 1. Normalize pagination and query parsing.
 * 2. Keep current user access consistent across handlers.
 */
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/middleware"
)

// 1. parsePagination reads page and page_size from query parameters.
func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page_size", "20")))
	return page, pageSize
}

// 2. parseBoolQuery reads a bool query parameter with false fallback.
func parseBoolQuery(c *gin.Context, key string) bool {
	value, _ := strconv.ParseBool(strings.TrimSpace(c.DefaultQuery(key, "false")))
	return value
}

// 3. parseOptionalBoolQuery reads a bool query parameter when present.
func parseOptionalBoolQuery(c *gin.Context, key string) *bool {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}

	return &value
}

// 4. parsePositiveIntQuery reads a positive integer query parameter.
func parsePositiveIntQuery(c *gin.Context, key string) int {
	value, _ := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if value < 0 {
		return 0
	}
	return value
}

// 5. currentUser returns the current authenticated user from gin context.
func currentUser(c *gin.Context) *middleware.CurrentUser {
	return middleware.GetCurrentUser(c)
}
