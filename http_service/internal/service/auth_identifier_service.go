/*
 * Unified password login business logic.
 * 1. Classify email, phone, and username identifiers.
 * 2. Verify credentials against the iSmart integration login endpoint only.
 * 3. Keep credential failures generic while preserving internal service errors.
 */
package service

import (
	"context"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. IdentifierLoginParams defines unified password login input.
type IdentifierLoginParams struct {
	Identifier string
	Password   string
}

// 2. LoginWithIdentifier classifies one identifier and verifies it via iSmart.
// Local AJO credentials are intentionally not consulted: iSmart is the source
// of truth and a successful login upserts the local member afterwards.
func (s *AuthService) LoginWithIdentifier(ctx context.Context, params IdentifierLoginParams) (*VerifyOTPResult, error) {
	identifier := strings.TrimSpace(params.Identifier)
	password := strings.TrimSpace(params.Password)
	if identifier == "" || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "identifier and password are required")
	}

	ismartParams := IsmartLoginParams{Account: identifier, Password: password}
	if email := normalizeEmail(identifier); isValidEmail(email) {
		ismartParams.Email = email
	}
	if countryCode, phoneNumber, ok := parseLoginPhoneIdentifier(identifier); ok {
		ismartParams.Account = joinPhone(countryCode, phoneNumber)
		ismartParams.Phone = ismartParams.Account
	}

	result, err := s.LoginWithIsmart(ctx, ismartParams)
	if err == nil {
		return result, nil
	}
	if isIsmartCredentialError(err) {
		return nil, errcode.New(errcode.CodeValidationError, "account or password is incorrect")
	}
	return nil, err
}

// 3. parseLoginPhoneIdentifier normalizes Hong Kong and mainland China phone input.
func parseLoginPhoneIdentifier(identifier string) (string, string, bool) {
	var digits strings.Builder
	hasExplicitCountryCode := false
	for _, item := range strings.TrimSpace(identifier) {
		switch {
		case item >= '0' && item <= '9':
			digits.WriteRune(item)
		case item == '+' && digits.Len() == 0 && !hasExplicitCountryCode:
			hasExplicitCountryCode = true
			continue
		case item == ' ' || item == '-' || item == '(' || item == ')':
			continue
		default:
			return "", "", false
		}
	}

	value := digits.String()
	switch {
	case len(value) == 11 && strings.HasPrefix(value, "852"):
		return "+852", strings.TrimPrefix(value, "852"), true
	case len(value) == 13 && strings.HasPrefix(value, "86") && value[2] == '1':
		return "+86", strings.TrimPrefix(value, "86"), true
	case hasExplicitCountryCode:
		return "", "", false
	case len(value) == 8:
		return "+852", value, true
	case len(value) == 11 && strings.HasPrefix(value, "1"):
		return "+86", value, true
	default:
		return "", "", false
	}
}
