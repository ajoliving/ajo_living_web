/*
 * iSmart 註冊聯絡方式檢查服務。
 * 1. 以 iSmart integration API 檢查電郵及手提電話是否可建立帳戶。
 * 2. 副戶授權只查 iSmart：電郵與電話分別走 check-contact；電話先按提交的國際號碼查，再兼容舊格式。
 * 3. 將上游回應收斂為 AJO 穩定的帳戶可用性與副戶解析結果。
 */
package service

import (
	"context"
	"strings"
	"unicode"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. IsmartRegistrationContactParams defines the contact values checked before iSmart registration.
type IsmartRegistrationContactParams struct {
	Email string
	Phone string
}

// 2. IsmartRegistrationContactResult reports iSmart contact availability.
type IsmartRegistrationContactResult struct {
	EmailAvailable bool
	PhoneAvailable bool
}

// 3. IsmartSubaccountContactParams defines contact values used to resolve an iSmart target user.
type IsmartSubaccountContactParams struct {
	Email string
	Phone string
}

// 4. CheckRegistrationContact checks whether iSmart can create an account with the submitted contact values.
func (s *IsmartExternalService) CheckRegistrationContact(ctx context.Context, params IsmartRegistrationContactParams) (*IsmartRegistrationContactResult, error) {
	result, err := s.postIntegration(ctx, "/auth/check-contact/", map[string]any{
		"email": normalizeEmail(params.Email),
		"phone": strings.TrimSpace(params.Phone),
	})
	if err != nil {
		return nil, err
	}

	payload := paymentMapValue(result.Payload)
	emailAvailable, emailOK := payload["email_available"].(bool)
	phoneAvailable, phoneOK := payload["phone_available"].(bool)
	if !emailOK || !phoneOK {
		return nil, errcode.New(errcode.CodeInternalError, "invalid ismart contact availability response")
	}

	return &IsmartRegistrationContactResult{
		EmailAvailable: emailAvailable,
		PhoneAvailable: phoneAvailable,
	}, nil
}

// 5. ResolveSubaccountTargetUser resolves an existing iSmart user by matching phone and email user IDs.
func (s *IsmartExternalService) ResolveSubaccountTargetUser(ctx context.Context, params IsmartSubaccountContactParams) (int64, error) {
	email := normalizeEmail(params.Email)
	phone := strings.TrimSpace(params.Phone)
	if !isValidEmail(email) || phone == "" {
		return 0, errcode.New(errcode.CodeValidationError, "valid target phone and email are required")
	}

	emailUserID, err := s.lookupSubaccountContactUserID(ctx, map[string]any{"email": email}, "email_available", "email_user_id")
	if err != nil {
		return 0, err
	}
	if emailUserID <= 0 {
		return 0, errcode.New(errcode.CodeNotFound, "ismart target user was not found for the supplied email")
	}

	phoneUserID, err := s.lookupSubaccountPhoneUserID(ctx, phone)
	if err != nil {
		return 0, err
	}
	if phoneUserID <= 0 {
		return 0, errcode.New(errcode.CodeNotFound, "ismart target user was not found for the supplied phone")
	}
	if phoneUserID != emailUserID {
		return 0, errcode.New(errcode.CodeValidationError, "target phone and email belong to different ismart users")
	}
	return phoneUserID, nil
}

// 6. lookupSubaccountPhoneUserID finds one iSmart user across stored and legacy phone formats.
func (s *IsmartExternalService) lookupSubaccountPhoneUserID(ctx context.Context, phone string) (int64, error) {
	var matchedID int64
	for _, candidate := range ismartSubaccountPhoneCandidates(phone) {
		userID, err := s.lookupSubaccountContactUserID(ctx, map[string]any{"phone": candidate}, "phone_available", "phone_user_id")
		if err != nil {
			if isIsmartPhoneNormalizationError(err) {
				continue
			}
			return 0, err
		}
		if userID <= 0 {
			continue
		}
		if matchedID > 0 && matchedID != userID {
			return 0, errcode.New(errcode.CodeValidationError, "target phone matches multiple ismart users")
		}
		matchedID = userID
	}
	return matchedID, nil
}

// 7. lookupSubaccountContactUserID reads one check-contact field from the subaccount API.
func (s *IsmartExternalService) lookupSubaccountContactUserID(ctx context.Context, payload map[string]any, availableKey string, userIDKey string) (int64, error) {
	result, err := s.postSubaccountIntegration(ctx, "/auth/check-contact/", payload)
	if err != nil {
		return 0, err
	}
	data := paymentMapValue(result.Payload)
	if available, ok := data[availableKey].(bool); ok && available {
		return 0, nil
	}
	for _, key := range []string{userIDKey, "user_id", "target_user_id"} {
		if id := paymentInt64Value(data[key]); id > 0 {
			return id, nil
		}
	}
	return 0, nil
}

// 8. ismartSubaccountPhoneCandidates builds iSmart phone lookup values for one submitted number.
func ismartSubaccountPhoneCandidates(phone string) []string {
	raw := strings.TrimSpace(phone)
	if raw == "" {
		return nil
	}

	digits := digitsOnly(raw)
	national := ismartSubaccountNationalNumber(digits)
	candidates := make([]string, 0, 12)
	seen := map[string]struct{}{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, exists := seen[value]; exists {
			return
		}
		seen[value] = struct{}{}
		candidates = append(candidates, value)
	}

	add(raw)
	add(digits)
	if digits != "" {
		add("+" + digits)
	}
	add(national)
	if national != "" {
		add("+" + national)
		add("+86" + national)
		add("86" + national)
		add("+852" + national)
		add("852" + national)
		add("+1" + national)
		add("1" + national)
	}
	return candidates
}

// 9. ismartSubaccountNationalNumber strips known country prefixes from one phone value.
func ismartSubaccountNationalNumber(digits string) string {
	switch {
	case strings.HasPrefix(digits, "86") && len(digits) == 13:
		return digits[2:]
	case strings.HasPrefix(digits, "852") && len(digits) == 11:
		return digits[3:]
	case strings.HasPrefix(digits, "1") && len(digits) == 12:
		return digits[1:]
	default:
		return digits
	}
}

// 10. digitsOnly keeps numeric characters from one phone value.
func digitsOnly(value string) string {
	var builder strings.Builder
	for _, item := range value {
		if unicode.IsDigit(item) {
			builder.WriteRune(item)
		}
	}
	return builder.String()
}

// 11. isIsmartPhoneNormalizationError ignores invalid candidate formats from iSmart.
func isIsmartPhoneNormalizationError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	if appErr, ok := err.(*errcode.AppError); ok && appErr != nil {
		message = appErr.Message
	}
	return strings.Contains(message, "電話號碼格式無效") || strings.Contains(message, "未能識別國家碼")
}
