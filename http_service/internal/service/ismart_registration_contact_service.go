/*
 * iSmart 註冊聯絡方式檢查服務。
 * 1. 註冊預檢與副戶授權統一復用 CheckSubaccountContacts，走同一條 subaccount integration check-contact 上游與解析路徑。
 * 2. 副戶授權以一次 check-contact 同時檢查電郵及電話，電話先規範化為 E.164。
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

// 4. IsmartSubaccountContactResult reports independent iSmart contact lookups.
type IsmartSubaccountContactResult struct {
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	EmailExists    bool   `json:"email_exists"`
	PhoneExists    bool   `json:"phone_exists"`
	EmailUserID    int64  `json:"email_user_id,omitempty"`
	PhoneUserID    int64  `json:"phone_user_id,omitempty"`
	SameUser       bool   `json:"same_user"`
	CanAuthorize   bool   `json:"can_authorize"`
	ConflictReason string `json:"conflict_reason,omitempty"`
}

// 5. CheckRegistrationContact checks iSmart contact availability via the shared subaccount check-contact path.
func (s *IsmartExternalService) CheckRegistrationContact(ctx context.Context, params IsmartRegistrationContactParams) (*IsmartRegistrationContactResult, error) {
	result, err := s.CheckSubaccountContacts(ctx, IsmartSubaccountContactParams{
		Email: params.Email,
		Phone: params.Phone,
	})
	if err != nil {
		return nil, err
	}

	return &IsmartRegistrationContactResult{
		EmailAvailable: !result.EmailExists,
		PhoneAvailable: !result.PhoneExists,
	}, nil
}

// 6. CheckSubaccountContacts resolves whichever iSmart contacts were supplied.
func (s *IsmartExternalService) CheckSubaccountContacts(ctx context.Context, params IsmartSubaccountContactParams) (*IsmartSubaccountContactResult, error) {
	email := normalizeEmail(params.Email)
	phone := strings.TrimSpace(params.Phone)
	if email == "" && phone == "" {
		return nil, errcode.New(errcode.CodeValidationError, "email or phone is required")
	}

	result := &IsmartSubaccountContactResult{Email: email, Phone: phone}
	if email != "" {
		if !isValidEmail(email) {
			return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
		}
	}
	if phone != "" {
		phone = normalizeSubaccountRegistrationPhone(phone)
		result.Phone = phone
	}

	payload := map[string]any{}
	if email != "" {
		payload["email"] = email
	}
	if phone != "" {
		payload["phone"] = phone
	}
	upstream, err := s.postSubaccountIntegration(ctx, "/auth/check-contact/", payload)
	if err != nil {
		return nil, err
	}
	data := paymentMapValue(upstream.Payload)
	if email != "" {
		available, ok := data["email_available"].(bool)
		if !ok {
			return nil, errcode.New(errcode.CodeInternalError, "invalid ismart email contact response")
		}
		result.EmailExists = !available
		result.EmailUserID = paymentInt64Value(data["email_user_id"])
		if result.EmailUserID <= 0 {
			result.EmailUserID = paymentInt64Value(data["user_id"])
		}
	}
	if phone != "" {
		available, ok := data["phone_available"].(bool)
		if !ok {
			return nil, errcode.New(errcode.CodeInternalError, "invalid ismart phone contact response")
		}
		result.PhoneExists = !available
		result.PhoneUserID = paymentInt64Value(data["phone_user_id"])
		if result.PhoneUserID <= 0 {
			result.PhoneUserID = paymentInt64Value(data["user_id"])
		}
	}
	if email == "" || phone == "" {
		return result, nil
	}
	switch {
	case result.EmailExists && result.PhoneExists && result.EmailUserID == result.PhoneUserID:
		result.SameUser = true
		result.CanAuthorize = true
	case result.EmailExists && result.PhoneExists:
		result.ConflictReason = "target phone and email belong to different ismart users"
	case result.EmailExists || result.PhoneExists:
		result.ConflictReason = "target phone and email do not match the same ismart account"
	}
	return result, nil
}

// 7. ResolveSubaccountTargetUser resolves an existing iSmart user by matching phone and email user IDs.
func (s *IsmartExternalService) ResolveSubaccountTargetUser(ctx context.Context, params IsmartSubaccountContactParams) (int64, error) {
	email := normalizeEmail(params.Email)
	phone := strings.TrimSpace(params.Phone)
	if !isValidEmail(email) || phone == "" {
		return 0, errcode.New(errcode.CodeValidationError, "valid target phone and email are required")
	}

	result, err := s.CheckSubaccountContacts(ctx, IsmartSubaccountContactParams{Phone: phone, Email: email})
	if err != nil {
		return 0, err
	}
	if !result.EmailExists || result.EmailUserID <= 0 {
		return 0, errcode.New(errcode.CodeNotFound, "ismart target user was not found for the supplied email")
	}
	if !result.PhoneExists || result.PhoneUserID <= 0 {
		return 0, errcode.New(errcode.CodeNotFound, "ismart target user was not found for the supplied phone")
	}
	if result.PhoneUserID != result.EmailUserID {
		return 0, errcode.New(errcode.CodeValidationError, "target phone and email belong to different ismart users")
	}
	return result.PhoneUserID, nil
}

// 8. digitsOnly keeps numeric characters from one phone value.
func digitsOnly(value string) string {
	var builder strings.Builder
	for _, item := range value {
		if unicode.IsDigit(item) {
			builder.WriteRune(item)
		}
	}
	return builder.String()
}
