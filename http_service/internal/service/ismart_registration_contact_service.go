/*
 * iSmart 註冊聯絡方式檢查服務。
 * 1. 以 iSmart integration API 檢查電郵及手提電話是否可建立帳戶。
 * 2. 將上游回應收斂為 AJO 註冊可用性判斷所需的穩定結構。
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

// 3. CheckRegistrationContact checks whether iSmart can create an account with the submitted contact values.
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
