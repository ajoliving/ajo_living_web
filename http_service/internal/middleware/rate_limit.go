/*
 * In-memory rate limiting middleware.
 * 1. Limit sensitive endpoints by IP or user key.
 * 2. Keep the implementation simple for local and single-node deployments.
 */
package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. InMemoryLimiter stores request counters by logical key.
type InMemoryLimiter struct {
	mu      sync.Mutex
	windows map[string]rateWindow
}

// 2. rateWindow stores count and expiry for a key.
type rateWindow struct {
	Count   int
	ResetAt time.Time
}

// 3. NewInMemoryLimiter creates a new limiter instance.
func NewInMemoryLimiter() *InMemoryLimiter {
	return &InMemoryLimiter{
		windows: make(map[string]rateWindow),
	}
}

// 4. Limit creates a middleware for a specific route policy.
func (l *InMemoryLimiter) Limit(limit int, window time.Duration, keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}

		if !l.allow(key, limit, window) {
			errcode.WriteError(c, errcode.New(errcode.CodeRateLimited, "rate limit exceeded"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// 5. allow checks and updates the current key window.
func (l *InMemoryLimiter) allow(key string, limit int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	current, exists := l.windows[key]
	if !exists || now.After(current.ResetAt) {
		l.windows[key] = rateWindow{
			Count:   1,
			ResetAt: now.Add(window),
		}
		return true
	}

	if current.Count >= limit {
		return false
	}

	current.Count++
	l.windows[key] = current
	return true
}
