/*
 * Unified password login business logic.
 * 1. Classify email, local username, and supported phone identifiers.
 * 2. Fall back to iSmart only after a local credential rejection.
 * 3. Keep credential failures generic while preserving internal service errors.
 */
package service

import (
	"context"
	"errors"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. IdentifierLoginParams defines unified password login input.
type IdentifierLoginParams struct {
	Identifier string
	Password   string
}

// 2. LoginWithIdentifier classifies one identifier and applies controlled iSmart fallback.
func (s *AuthService) LoginWithIdentifier(ctx context.Context, params IdentifierLoginParams) (*VerifyOTPResult, error) {
	identifier := strings.TrimSpace(params.Identifier)
	password := strings.TrimSpace(params.Password)
	if identifier == "" || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "identifier and password are required")
	}

	localResult, localErr, attemptedLocal := s.loginWithLocalIdentifier(ctx, identifier, password)
	if localErr == nil && attemptedLocal {
		return localResult, nil
	}
	if attemptedLocal && !isLocalCredentialRejection(localErr) {
		return nil, localErr
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
		return nil, unifiedLoginCredentialError()
	}
	return nil, err
}

// 3. loginWithLocalIdentifier attempts exactly one classified local login path.
func (s *AuthService) loginWithLocalIdentifier(ctx context.Context, identifier string, password string) (*VerifyOTPResult, error, bool) {
	if strings.Contains(identifier, "@") {
		result, err := s.LoginWithEmail(ctx, EmailPasswordParams{Email: identifier, Password: password})
		return result, err, true
	}
	if countryCode, phoneNumber, ok := parseLoginPhoneIdentifier(identifier); ok {
		result, err := s.LoginWithPhone(ctx, PhonePasswordParams{
			PhoneCountryCode: countryCode,
			PhoneNumber:      phoneNumber,
			Password:         password,
		})
		return result, err, true
	}
	if isValidUsername(normalizeUsername(identifier)) {
		result, err := s.LoginWithUsername(ctx, EmailPasswordParams{Username: identifier, Password: password})
		return result, err, true
	}

	return nil, nil, false
}

// 4. parseLoginPhoneIdentifier normalizes Hong Kong and mainland China phone input.
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

// 5. isLocalCredentialRejection limits fallback to known local credential failures.
func isLocalCredentialRejection(err error) bool {
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeValidationError {
		return false
	}

	switch appErr.Message {
	case "email or password is incorrect", "username or password is incorrect", "phone or password is incorrect":
		return true
	default:
		return false
	}
}

// 6. unifiedLoginCredentialError returns the public credential failure response.
func unifiedLoginCredentialError() error {
	return errcode.New(errcode.CodeValidationError, "account or password is incorrect")
}
