/*
 * 受控資料快照重加密工具。
 * 1. 使用來源密鑰讀取快照中的加密欄位。
 * 2. 使用測試環境密鑰重新保存，不輸出任何明文或密鑰。
 * 3. 僅供測試資料庫快照導入或修復時執行。
 */
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/utils"
)

type encryptedListingContact struct {
	ListingID         int64
	PhoneEncrypted    sql.NullString
	Phone2Encrypted   sql.NullString
	WhatsAppEncrypted sql.NullString
	WeChatEncrypted   sql.NullString
	EmailEncrypted    sql.NullString
}

type encryptedCredential struct {
	UserID            int64
	PasswordEncrypted sql.NullString
}

type encryptedIsmartAccount struct {
	UserID              int64
	RelayTokenEncrypted sql.NullString
	PasswordEncrypted   sql.NullString
}

// 1. main performs the controlled snapshot re-encryption transaction.
func main() {
	dsn := os.Getenv("REKEY_DSN")
	fromKey := os.Getenv("REKEY_FROM_KEY")
	toKey := os.Getenv("REKEY_TO_KEY")
	if dsn == "" || fromKey == "" || toKey == "" {
		fatal("REKEY_DSN, REKEY_FROM_KEY and REKEY_TO_KEY are required")
	}
	if fromKey == toKey {
		fatal("source and destination encryption keys must differ")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fatal("open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fatal("open database handle: %v", err)
	}
	defer sqlDB.Close()

	updated, err := rekeyDatabase(db, fromKey, toKey)
	if err != nil {
		fatal("rekey encrypted data: %v", err)
	}
	fmt.Printf("rekeyed %d encrypted values\n", updated)
}

// 2. rekeyDatabase updates every application-managed encrypted column atomically.
func rekeyDatabase(db *gorm.DB, fromKey string, toKey string) (int, error) {
	updated := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		listingUpdated, err := rekeyListingContacts(tx, fromKey, toKey)
		if err != nil {
			return err
		}
		credentialUpdated, err := rekeyCredentials(tx, fromKey, toKey)
		if err != nil {
			return err
		}
		ismartUpdated, err := rekeyIsmartAccounts(tx, fromKey, toKey)
		if err != nil {
			return err
		}
		updated = listingUpdated + credentialUpdated + ismartUpdated
		return nil
	})
	return updated, err
}

// 3. rekeyListingContacts re-encrypts listing phone, WhatsApp, Wechat, and email values.
func rekeyListingContacts(tx *gorm.DB, fromKey string, toKey string) (int, error) {
	var rows []encryptedListingContact
	if err := tx.Raw(`
		SELECT listing_id, phone_encrypted, phone2_encrypted, whats_app_encrypted,
		       we_chat_encrypted, email_encrypted
		FROM listing_contacts
	`).Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("load listing contacts: %w", err)
	}
	updated := 0
	for _, row := range rows {
		values, count, err := rekeyValues(fromKey, toKey,
			row.PhoneEncrypted,
			row.Phone2Encrypted,
			row.WhatsAppEncrypted,
			row.WeChatEncrypted,
			row.EmailEncrypted,
		)
		if err != nil {
			return 0, fmt.Errorf("listing contact %d: %w", row.ListingID, err)
		}
		if count == 0 {
			continue
		}
		result := tx.Exec(`
			UPDATE listing_contacts
			SET phone_encrypted = ?, phone2_encrypted = ?, whats_app_encrypted = ?,
			    we_chat_encrypted = ?, email_encrypted = ?, updated_at = CURRENT_TIMESTAMP
			WHERE listing_id = ?
		`, values[0], values[1], values[2], values[3], values[4], row.ListingID)
		if result.Error != nil {
			return 0, fmt.Errorf("save listing contact %d: %w", row.ListingID, result.Error)
		}
		updated += count
	}
	return updated, nil
}

// 4. rekeyCredentials re-encrypts any retained credential ciphertext.
func rekeyCredentials(tx *gorm.DB, fromKey string, toKey string) (int, error) {
	var rows []encryptedCredential
	if err := tx.Raw(`SELECT user_id, password_encrypted FROM user_credentials`).Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("load user credentials: %w", err)
	}
	updated := 0
	for _, row := range rows {
		value, count, err := rekeyValues(fromKey, toKey, row.PasswordEncrypted)
		if err != nil {
			return 0, fmt.Errorf("user credential %d: %w", row.UserID, err)
		}
		if count == 0 {
			continue
		}
		if err := tx.Exec("UPDATE user_credentials SET password_encrypted = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?", value[0], row.UserID).Error; err != nil {
			return 0, fmt.Errorf("save user credential %d: %w", row.UserID, err)
		}
		updated += count
	}
	return updated, nil
}

// 5. rekeyIsmartAccounts re-encrypts any retained upstream credentials.
func rekeyIsmartAccounts(tx *gorm.DB, fromKey string, toKey string) (int, error) {
	var rows []encryptedIsmartAccount
	if err := tx.Raw(`SELECT user_id, relay_token_encrypted, password_encrypted FROM user_ismart_accounts`).Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("load ismart accounts: %w", err)
	}
	updated := 0
	for _, row := range rows {
		values, count, err := rekeyValues(fromKey, toKey, row.RelayTokenEncrypted, row.PasswordEncrypted)
		if err != nil {
			return 0, fmt.Errorf("ismart account %d: %w", row.UserID, err)
		}
		if count == 0 {
			continue
		}
		if err := tx.Exec("UPDATE user_ismart_accounts SET relay_token_encrypted = ?, password_encrypted = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?", values[0], values[1], row.UserID).Error; err != nil {
			return 0, fmt.Errorf("save ismart account %d: %w", row.UserID, err)
		}
		updated += count
	}
	return updated, nil
}

// 6. rekeyValues preserves empty fields and re-encrypts non-empty ciphertext.
func rekeyValues(fromKey string, toKey string, values ...sql.NullString) ([]string, int, error) {
	result := make([]string, len(values))
	updated := 0
	for index, value := range values {
		if !value.Valid || value.String == "" {
			if value.Valid {
				result[index] = value.String
			}
			continue
		}
		plain, err := decryptSafely(fromKey, value.String)
		if err != nil {
			if _, destinationErr := decryptSafely(toKey, value.String); destinationErr == nil {
				result[index] = value.String
				continue
			}
			// 既有測試資料可能使用更早的測試密鑰；保留原值，避免無法識別的密文被覆蓋。
			result[index] = value.String
			continue
		}
		encrypted, err := utils.EncryptString(toKey, plain)
		if err != nil {
			return nil, 0, err
		}
		result[index] = encrypted
		updated++
	}
	return result, updated, nil
}

// 7. decryptSafely converts malformed ciphertext into an ordinary migration error.
func decryptSafely(secret string, value string) (plain string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("invalid encrypted value")
		}
	}()
	plain, err = utils.DecryptString(secret, value)
	if err != nil {
		return "", errors.New("unable to decrypt existing value")
	}
	return plain, nil
}

// 8. fatal terminates without printing sensitive configuration.
func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
