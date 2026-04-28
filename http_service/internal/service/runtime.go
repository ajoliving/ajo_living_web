/*
 * Shared service runtime dependencies.
 * 1. Keep construction dependencies explicit.
 * 2. Reuse providers, config, and storage across services.
 */
package service

import (
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
)

// 1. Runtime groups shared service dependencies.
type Runtime struct {
	Config          *config.Config
	DB              *gorm.DB
	Logger          *slog.Logger
	OTPProvider     OTPProvider
	MailSender      MailSender
	StorageProvider StorageProvider
	OTPStore        *OTPStore
	Now             func() time.Time
}

// 2. OTPStore stores mock OTP verification codes in memory.
type OTPStore struct {
	mu    sync.Mutex
	codes map[string]OTPCode
}

// 3. OTPCode stores a pending OTP code and expiry.
type OTPCode struct {
	Code      string
	ExpiresAt time.Time
}

// 4. NewOTPStore creates the in-memory OTP store.
func NewOTPStore() *OTPStore {
	return &OTPStore{
		codes: make(map[string]OTPCode),
	}
}

// 5. Save stores an OTP code for a key.
func (s *OTPStore) Save(key string, value OTPCode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.codes[key] = value
}

// 6. Get returns an OTP code by key.
func (s *OTPStore) Get(key string) (OTPCode, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.codes[key]
	return value, ok
}

// 7. Delete removes an OTP code by key.
func (s *OTPStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.codes, key)
}
