/*
 * iSmart 註冊聯絡方式檢查服務。
 * 1. 以 iSmart integration API 檢查電郵及手提電話是否可建立帳戶。
 * 2. 以同一上游聯絡檢查接口解析副戶授權所需的 target_user_id。
 * 3. 將上游回應收斂為 AJO 穩定的帳戶可用性與副戶解析結果。
 */
package service

import (
	"context"
	"strings"

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

// 5. ResolveSubaccountTargetUser resolves an existing iSmart user by phone and email.
func (s *IsmartExternalService) ResolveSubaccountTargetUser(ctx context.Context, params IsmartSubaccountContactParams) (int64, error) {
	email := normalizeEmail(params.Email)
	phone := strings.TrimSpace(params.Phone)
	if !isValidEmail(email) || phone == "" {
		return 0, errcode.New(errcode.CodeValidationError, "valid target phone and email are required")
	}

	result, err := s.postIntegration(ctx, "/auth/check-contact/", map[string]any{
		"email": email,
		"phone": phone,
	})
	if err != nil {
		return 0, err
	}

	payload := paymentMapValue(result.Payload)
	for _, key := range []string{"target_user_id", "user_id"} {
		if targetUserID := paymentInt64Value(payload[key]); targetUserID > 0 {
			return targetUserID, nil
		}
	}
	for _, key := range []string{"target_user", "user"} {
		if targetUser := paymentMapValue(payload[key]); len(targetUser) > 0 {
			for _, idKey := range []string{"target_user_id", "user_id", "id"} {
				if targetUserID := paymentInt64Value(targetUser[idKey]); targetUserID > 0 {
					return targetUserID, nil
				}
			}
		}
	}
	return 0, errcode.New(errcode.CodeNotFound, "ismart target user was not found for the supplied phone and email")
}
