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
	MediaProcessor  MediaProcessor
	MediaScanner    MediaScanner
	CacheStore      CacheStore
	OTPStore        *OTPStore
	WalletService   *WalletService
	Now             func() time.Time
}

// 2. OTPStore stores short-lived OTP codes and phone delivery cooldowns in memory.
type OTPStore struct {
	mu         sync.Mutex
	codes      map[string]OTPCode
	sending    map[string]bool
	lastSentAt map[string]time.Time
}

// 3. OTPCode stores a pending OTP code and expiry.
type OTPCode struct {
	Code      string
	ExpiresAt time.Time
}

// 4. NewOTPStore creates the in-memory OTP store.
func NewOTPStore() *OTPStore {
	return &OTPStore{
		codes:      make(map[string]OTPCode),
		sending:    make(map[string]bool),
		lastSentAt: make(map[string]time.Time),
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

// 8. ReservePhoneSend prevents concurrent or frequent delivery to one phone number.
func (s *OTPStore) ReservePhoneSend(key string, now time.Time, cooldown time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sending[key] || (!s.lastSentAt[key].IsZero() && now.Sub(s.lastSentAt[key]) < cooldown) {
		return false
	}

	s.sending[key] = true
	return true
}

// 9. CommitPhoneSend stores a code only after the provider accepts delivery.
func (s *OTPStore) CommitPhoneSend(key string, value OTPCode, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.codes[key] = value
	s.lastSentAt[key] = now
	delete(s.sending, key)
}

// 10. CancelPhoneSend releases a failed delivery reservation.
func (s *OTPStore) CancelPhoneSend(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sending, key)
}

// 11. ConsumePhoneCode validates and removes a phone OTP in one lock scope.
func (s *OTPStore) ConsumePhoneCode(key string, code string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.codes[key]
	if !ok || now.After(record.ExpiresAt) || code != record.Code {
		if ok && now.After(record.ExpiresAt) {
			delete(s.codes, key)
		}
		return false
	}

	delete(s.codes, key)
	return true
}
