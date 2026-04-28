/*
 * Public identifier helpers.
 * 1. Generate sortable public IDs for external resources.
 * 2. Keep ID generation consistent across all modules.
 */
package utils

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropyMu sync.Mutex
)

// 1. NewPublicID generates a ULID string for public APIs.
func NewPublicID() string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// 2. NewNumericCode generates a fixed-length numeric verification code.
func NewNumericCode(length int) string {
	if length <= 0 {
		length = 6
	}

	code := make([]byte, 0, length)
	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			code = append(code, '0')
			continue
		}

		code = append(code, byte('0'+n.Int64()))
	}

	return string(code)
}
