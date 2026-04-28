/*
 * Sensitive data helpers.
 * 1. Encrypt and decrypt contact values with AES-GCM.
 * 2. Build masked output for protected phone and WhatsApp numbers.
 */
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// 1. EncryptString encrypts plain text with the provided key material.
func EncryptString(secret string, value string) (string, error) {
	block, err := aes.NewCipher(normalizeKey(secret))
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 2. DecryptString decrypts an AES-GCM encrypted value.
func DecryptString(secret string, value string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(normalizeKey(secret))
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	nonce, ciphertext := payload[:nonceSize], payload[nonceSize:]
	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

// 3. MaskPhone keeps the first and last digits visible.
func MaskPhone(value string) string {
	if len(value) <= 4 {
		return value
	}

	if len(value) <= 8 {
		return value[:2] + "****" + value[len(value)-2:]
	}

	return value[:3] + "****" + value[len(value)-3:]
}

// 4. HashPassword hashes a password with PBKDF2-HMAC-SHA256.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	iterations := 120000
	hash := pbkdf2SHA256([]byte(password), salt, iterations, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(hash)), nil
}

// 5. VerifyPassword compares a plain password with a stored hash.
func VerifyPassword(password string, storedHash string) bool {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expectedHash, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actualHash := pbkdf2SHA256([]byte(password), salt, iterations, len(expectedHash))
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

// 6. pbkdf2SHA256 derives a key using PBKDF2-HMAC-SHA256.
func pbkdf2SHA256(password []byte, salt []byte, iterations int, keyLength int) []byte {
	hashLength := sha256.Size
	blocks := (keyLength + hashLength - 1) / hashLength
	derivedKey := make([]byte, 0, blocks*hashLength)

	for block := 1; block <= blocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		sum := mac.Sum(nil)
		blockResult := append([]byte(nil), sum...)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(sum)
			sum = mac.Sum(nil)
			for j := range blockResult {
				blockResult[j] ^= sum[j]
			}
		}

		derivedKey = append(derivedKey, blockResult...)
	}

	return derivedKey[:keyLength]
}

// 7. normalizeKey normalizes arbitrary key material to 32 bytes.
func normalizeKey(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}
