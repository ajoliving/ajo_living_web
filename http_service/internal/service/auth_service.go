/*
 * Authentication business logic.
 * 1. Manage phone and email OTP request and verification flows.
 * 2. Create member accounts and issue JWT tokens.
 * 3. Authenticate bearer tokens for middleware usage.
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	accessTokenTTL          = 2 * time.Hour
	passwordResetEmailScene = "password_reset"
)

// 1. AuthService handles OTP and token operations.
type AuthService struct {
	runtime *Runtime
}

// 2. RequestOTPParams defines the OTP request input.
type RequestOTPParams struct {
	PhoneCountryCode string
	PhoneNumber      string
	Scene            string
}

// 3. RequestOTPResult defines the OTP request result.
type RequestOTPResult struct {
	ExpiresIn int    `json:"expires_in"`
	MockCode  string `json:"mock_code,omitempty"`
}

// 4. VerifyOTPParams defines the OTP verify input.
type VerifyOTPParams struct {
	PhoneCountryCode string
	PhoneNumber      string
	Scene            string
	Code             string
}

// 5. EmailOTPParams defines email verification input.
type EmailOTPParams struct {
	Email       string
	Scene       string
	Code        string
	DisplayName string
}

// 6. EmailPasswordParams defines email password auth input.
type EmailPasswordParams struct {
	Email                 string
	Password              string
	EngName               string
	ChiName               string
	DisplayName           string
	PhoneCountryCode      string
	PhoneNumber           string
	Username              string
	PublisherIdentityType string
	AccountType           string
	IDCard                string
	Remark                string
	Gender                string
	IsReceiveEmail        *bool
	PrimaryCommunityID    string
	PrimaryCommunityName  string
	ResidenceFloor        string
	ResidenceUnit         string
}

// 6.1 PasswordResetParams defines email password reset input.
type PasswordResetParams struct {
	Email    string
	Code     string
	Password string
}

// 6.2 PasswordResetResult defines password reset output.
type PasswordResetResult struct {
	PasswordReset bool `json:"password_reset"`
}

// 7. PhonePasswordParams defines phone password auth input.
type PhonePasswordParams struct {
	PhoneCountryCode string
	PhoneNumber      string
	Password         string
}

// 8. IsmartLoginParams defines POS Web login input.
type IsmartLoginParams struct {
	Account  string
	Password string
	Phone    string
	Email    string
}

// 9. IsmartMessage defines the POS Web account payload.
type IsmartMessage struct {
	UserID                             int64          `json:"user_id"`
	Username                           string         `json:"username"`
	Email                              string         `json:"email,omitempty"`
	Phone                              string         `json:"phone,omitempty"`
	IsStaff                            bool           `json:"is_staff"`
	Building                           []string       `json:"building"`
	StaffBuildingPermissions           []string       `json:"staff_building_permissions"`
	ClientBuildingPermissions          []string       `json:"client_building_permissions"`
	ClientBuildingFlatUnitsPermissions []string       `json:"client_building_flat_units_permissions"`
	RawMessage                         map[string]any `json:"-"`
	ProfileSnapshot                    map[string]any `json:"-"`
}

// 10. IsmartLoginResponse defines the POS Web response shape.
type IsmartLoginResponse struct {
	Code    any            `json:"code"`
	Message string         `json:"message"`
	Msg     *IsmartMessage `json:"msg"`
}

// 11. posRelayLoginResponse defines the authenticated POS relay session response.
type posRelayLoginResponse struct {
	Token string `json:"token"`
}

// 12. VerifyOTPResult defines the OTP verify result.
type VerifyOTPResult struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int              `json:"expires_in"`
	User         AuthUserResponse `json:"user"`
}

// 13. AuthUserResponse defines the auth response user payload.
type AuthUserResponse struct {
	PublicID         string         `json:"public_id"`
	MemberStatus     string         `json:"member_status"`
	MemberType       string         `json:"member_type"`
	IsStaff          bool           `json:"is_staff"`
	Role             string         `json:"role"`
	Roles            []string       `json:"roles"`
	Permissions      []string       `json:"permissions"`
	ProfileCompleted bool           `json:"profile_completed"`
	AccountType      string         `json:"account_type"`
	IsmartMsg        *IsmartMessage `json:"ismart_msg,omitempty"`
	IsmartRaw        map[string]any `json:"ismart_raw,omitempty"`
}

// 14. AuthIdentity defines the middleware-facing auth identity.
type AuthIdentity struct {
	UserID             int64
	PublicID           string
	MemberStatus       string
	MemberType         string
	IsStaff            bool
	Role               string
	Roles              []string
	Permissions        []string
	PrimaryCommunityID *int64
	AccountType        string
}

// 15. accessTokenClaims stores custom JWT claims.
type accessTokenClaims struct {
	UserID     int64  `json:"user_id"`
	MemberType string `json:"member_type"`
	IsStaff    bool   `json:"is_staff"`
	Role       string `json:"role"`
	TokenType  string `json:"token_type"`
	jwt.RegisteredClaims
}

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// 16. NewAuthService creates an auth service instance.
func NewAuthService(runtime *Runtime) *AuthService {
	return &AuthService{runtime: runtime}
}

// 17. RequestOTP delivers one phone OTP after phone-level cooldown validation.
func (s *AuthService) RequestOTP(ctx context.Context, params RequestOTPParams) (*RequestOTPResult, error) {
	countryCode, phoneNumber, err := normalizePhoneOTPInput(params.PhoneCountryCode, params.PhoneNumber)
	if err != nil {
		return nil, errcode.New(errcode.CodeValidationError, "valid phone number is required")
	}

	scene := strings.TrimSpace(params.Scene)
	if scene == "" {
		scene = "login"
	}

	now := s.runtime.Now()
	key := otpKey(countryCode, phoneNumber, scene)
	if !s.runtime.OTPStore.ReservePhoneSend(key, now, s.runtime.Config.OTPResendCooldown) {
		return nil, errcode.New(errcode.CodeRateLimited, "please wait before requesting another verification code")
	}

	code := s.runtime.Config.OTPMockCode
	if !strings.EqualFold(strings.TrimSpace(s.runtime.Config.OTPProvider), "mock") {
		code = utils.NewNumericCode(6)
	}

	if err := s.runtime.OTPProvider.SendCode(ctx, countryCode+phoneNumber, scene, code); err != nil {
		s.runtime.OTPStore.CancelPhoneSend(key)
		return nil, errcode.New(errcode.CodeInternalError, "failed to send otp")
	}

	expiresAt := now.Add(5 * time.Minute)
	s.runtime.OTPStore.CommitPhoneSend(key, OTPCode{Code: code, ExpiresAt: expiresAt}, now)

	result := &RequestOTPResult{
		ExpiresIn: int((5 * time.Minute).Seconds()),
	}

	if strings.EqualFold(strings.TrimSpace(s.runtime.Config.OTPProvider), "mock") && otpMockDisclosureAllowed(s.runtime.Config.AppEnv) {
		result.MockCode = code
	}

	return result, nil
}

// 18. VerifyOTP validates the OTP code and issues access tokens.
func (s *AuthService) VerifyOTP(ctx context.Context, params VerifyOTPParams) (*VerifyOTPResult, error) {
	countryCode, phoneNumber, err := normalizePhoneOTPInput(params.PhoneCountryCode, params.PhoneNumber)
	if err != nil {
		return nil, errcode.New(errcode.CodeValidationError, "otp is invalid or expired")
	}
	key := otpKey(countryCode, phoneNumber, params.Scene)
	if !s.runtime.OTPStore.ConsumePhoneCode(key, strings.TrimSpace(params.Code), s.runtime.Now()) {
		return nil, errcode.New(errcode.CodeValidationError, "otp is invalid or expired")
	}

	result, err := s.upsertUserByPhone(ctx, countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 19. RequestEmailOTP stores a code and sends it to the target email.
func (s *AuthService) RequestEmailOTP(ctx context.Context, params EmailOTPParams) (*RequestOTPResult, error) {
	email := normalizeEmail(params.Email)
	if !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}

	scene := normalizeAuthScene(params.Scene)
	code := s.runtime.Config.OTPMockCode
	if s.runtime.Config.MailEnabled {
		code = utils.NewNumericCode(6)
	}

	expiresAt := s.runtime.Now().Add(5 * time.Minute)
	s.runtime.OTPStore.Save(emailOTPKey(email, scene), OTPCode{
		Code:      code,
		ExpiresAt: expiresAt,
	})

	subject, body := buildEmailOTPContent(code, scene)
	if err := s.runtime.MailSender.Send(ctx, email, subject, body); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to send email otp")
	}

	result := &RequestOTPResult{ExpiresIn: int(time.Until(expiresAt).Seconds())}
	if !s.runtime.Config.MailEnabled && otpMockDisclosureAllowed(s.runtime.Config.AppEnv) {
		result.MockCode = code
	}

	return result, nil
}

// 20. VerifyEmailOTP validates an email code and signs the member in.
func (s *AuthService) VerifyEmailOTP(ctx context.Context, params EmailOTPParams) (*VerifyOTPResult, error) {
	email := normalizeEmail(params.Email)
	if !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}

	scene := normalizeAuthScene(params.Scene)
	key := emailOTPKey(email, scene)
	record, ok := s.runtime.OTPStore.Get(key)
	if !ok || s.runtime.Now().After(record.ExpiresAt) {
		return nil, errcode.New(errcode.CodeValidationError, "email otp is invalid or expired")
	}
	if strings.TrimSpace(params.Code) != record.Code {
		return nil, errcode.New(errcode.CodeValidationError, "email otp is invalid or expired")
	}

	result, err := s.upsertUserByEmail(ctx, email, params.DisplayName)
	if err != nil {
		return nil, err
	}

	s.runtime.OTPStore.Delete(key)
	return result, nil
}

// 21. RegisterWithEmail creates an email and phone password account.
func (s *AuthService) RegisterWithEmail(ctx context.Context, params EmailPasswordParams) (*VerifyOTPResult, error) {
	email := normalizeEmail(params.Email)
	password := strings.TrimSpace(params.Password)
	engName := strings.TrimSpace(params.EngName)
	if engName == "" {
		engName = strings.TrimSpace(params.DisplayName)
	}
	phoneCountryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	username := strings.TrimSpace(params.Username)
	if username == "" {
		username = registrationUsernameCandidate(phoneNumber, email, engName)
	}
	normalizedUsername := normalizeUsername(username)
	accountType := normalizeRegistrationAccountType(params.AccountType)
	if accountType == "" {
		return nil, errcode.New(errcode.CodeValidationError, "account type is invalid")
	}
	memberStatus := "active"
	if accountType == AccountTypeIndividualAgent || accountType == AccountTypeAgencyCompany {
		memberStatus = "pending_profile"
	}
	primaryCommunityPublicID := strings.TrimSpace(params.PrimaryCommunityID)
	primaryCommunityName := strings.TrimSpace(params.PrimaryCommunityName)
	residenceFloor := strings.TrimSpace(params.ResidenceFloor)
	residenceUnit := strings.TrimSpace(params.ResidenceUnit)
	residenceBindingRequested, err := validateRegistrationResidenceBinding(primaryCommunityPublicID, residenceFloor, residenceUnit)
	if err != nil {
		return nil, err
	}
	if !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}
	if !isValidPhone(phoneCountryCode, phoneNumber) || !isValidEnglishName(engName) || !isValidUsername(normalizedUsername) || len(password) < 8 {
		return nil, errcode.New(errcode.CodeValidationError, "valid phone number, English name, username, and password are required")
	}
	gender := strings.ToUpper(strings.TrimSpace(params.Gender))
	if gender != "" && gender != "M" && gender != "F" {
		return nil, errcode.New(errcode.CodeValidationError, "gender must be M or F")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	passwordEncrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	if err := s.ensureEmailRegistrationAvailable(ctx, email, phoneCountryCode, phoneNumber, normalizedUsername); err != nil {
		return nil, err
	}
	if residenceBindingRequested && strings.TrimSpace(s.runtime.Config.IsmartIntegrationAPIBaseURL) == "" {
		return nil, errcode.New(errcode.CodeInternalError, "ismart integration is not configured")
	}

	var ismartMessage *IsmartMessage
	if strings.TrimSpace(s.runtime.Config.IsmartIntegrationAPIBaseURL) != "" {
		isReceiveEmail := true
		if params.IsReceiveEmail != nil {
			isReceiveEmail = *params.IsReceiveEmail
		}
		ismartMessage, err = NewIsmartExternalService(s.runtime).RegisterDirectAccount(ctx, IsmartDirectRegistrationParams{
			Phone:          phoneNumber,
			Email:          email,
			EngName:        engName,
			ChiName:        strings.TrimSpace(params.ChiName),
			LegalEntity:    ismartLegalEntity(accountType),
			IDCard:         strings.TrimSpace(params.IDCard),
			Remark:         strings.TrimSpace(params.Remark),
			Gender:         gender,
			IsReceiveEmail: isReceiveEmail,
			Password:       password,
		})
		if err != nil {
			return nil, err
		}
	}

	var user model.User
	var profile model.UserProfile
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.UserCredential{}).Where("email = ?", email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "email is already registered")
		}
		if err := tx.Model(&model.User{}).Where("phone_country_code = ? AND phone_number = ?", phoneCountryCode, phoneNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "phone number is already registered")
		}
		if err := tx.Model(&model.UserCredential{}).Where("username = ?", normalizedUsername).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "username is already registered")
		}

		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: phoneCountryCode,
			PhoneNumber:      phoneNumber,
			MemberStatus:     memberStatus,
			MemberType:       MemberTypeUser,
			IsStaff:          false,
			IsVerifiedPhone:  false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		credential := model.UserCredential{
			UserID:            user.ID,
			Username:          &normalizedUsername,
			PasswordHash:      passwordHash,
			PasswordEncrypted: passwordEncrypted,
			IsVerified:        true,
		}
		credential.Email = &email
		if err := tx.Select("UserID", "Username", "Email", "PasswordHash", "PasswordEncrypted", "IsVerified").Create(&credential).Error; err != nil {
			return err
		}

		profile = model.UserProfile{
			UserID:                user.ID,
			DisplayName:           engName,
			AccountType:           accountType,
			PublisherIdentityType: derivedPublisherIdentity(accountType),
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		if ismartMessage != nil {
			if err := s.saveIsmartAccount(ctx, tx, user.ID, ismartMessage, "", password); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to register email account")
	}
	if residenceBindingRequested {
		if err := s.submitRegistrationResidenceBinding(ctx, user.ID, primaryCommunityPublicID, primaryCommunityName, residenceFloor, residenceUnit, phoneNumber, email, engName, params); err != nil {
			return nil, err
		}
		if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load property binding request")
		}
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 21.1 ensureEmailRegistrationAvailable checks local identities before creating an iSmart account.
func (s *AuthService) ensureEmailRegistrationAvailable(ctx context.Context, email string, phoneCountryCode string, phoneNumber string, username string) error {
	checks := []struct {
		model   any
		query   string
		args    []any
		message string
	}{
		{&model.UserCredential{}, "email = ?", []any{email}, "email is already registered"},
		{&model.User{}, "phone_country_code = ? AND phone_number = ?", []any{phoneCountryCode, phoneNumber}, "phone number is already registered"},
		{&model.UserCredential{}, "username = ?", []any{username}, "username is already registered"},
	}
	for _, check := range checks {
		var count int64
		if err := s.runtime.DB.WithContext(ctx).Model(check.model).Where(check.query, check.args...).Count(&count).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to validate registration account")
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, check.message)
		}
	}
	return nil
}

// 22. LoginWithEmail validates an email password account and issues tokens.
func (s *AuthService) LoginWithEmail(ctx context.Context, params EmailPasswordParams) (*VerifyOTPResult, error) {
	email := normalizeEmail(params.Email)
	password := strings.TrimSpace(params.Password)
	if !isValidEmail(email) || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "valid email and password are required")
	}

	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("email = ?", email).First(&credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "email or password is incorrect")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load email account")
	}

	if !utils.VerifyPassword(password, credential.PasswordHash) {
		return nil, errcode.New(errcode.CodeValidationError, "email or password is incorrect")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, credential.UserID).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load email account")
	}

	var profile model.UserProfile
	profileErr := s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&profile).Error
	if errors.Is(profileErr, gorm.ErrRecordNotFound) {
		profile = model.UserProfile{UserID: user.ID, DisplayName: email}
		if err := s.runtime.DB.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create email profile")
		}
	} else if profileErr != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load email profile")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 23. LoginWithUsername validates a username password account and issues tokens.
func (s *AuthService) LoginWithUsername(ctx context.Context, params EmailPasswordParams) (*VerifyOTPResult, error) {
	username := normalizeUsername(params.Username)
	password := strings.TrimSpace(params.Password)
	if !isValidUsername(username) || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "valid username and password are required")
	}

	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("username = ?", username).First(&credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "username or password is incorrect")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load username account")
	}

	if !utils.VerifyPassword(password, credential.PasswordHash) {
		return nil, errcode.New(errcode.CodeValidationError, "username or password is incorrect")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, credential.UserID).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load username account")
	}

	var profile model.UserProfile
	profileErr := s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&profile).Error
	if errors.Is(profileErr, gorm.ErrRecordNotFound) {
		profile = model.UserProfile{UserID: user.ID, DisplayName: username}
		if err := s.runtime.DB.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create username profile")
		}
	} else if profileErr != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load username profile")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 24. LoginWithPhone validates a phone password account and issues tokens.
func (s *AuthService) LoginWithPhone(ctx context.Context, params PhonePasswordParams) (*VerifyOTPResult, error) {
	phoneCountryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	password := strings.TrimSpace(params.Password)
	if !isValidPhone(phoneCountryCode, phoneNumber) || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "valid phone number and password are required")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).Where("phone_country_code = ? AND phone_number = ?", phoneCountryCode, phoneNumber).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "phone or password is incorrect")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load phone account")
	}

	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "phone or password is incorrect")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load phone account")
	}
	if !utils.VerifyPassword(password, credential.PasswordHash) {
		return nil, errcode.New(errcode.CodeValidationError, "phone or password is incorrect")
	}

	var profile model.UserProfile
	profileErr := s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&profile).Error
	if errors.Is(profileErr, gorm.ErrRecordNotFound) {
		profile = model.UserProfile{UserID: user.ID, DisplayName: phoneCountryCode + phoneNumber}
		if err := s.runtime.DB.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create phone profile")
		}
	} else if profileErr != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load phone profile")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 25. LoginWithIsmart validates POS Web credentials and signs into the local account.
func (s *AuthService) LoginWithIsmart(ctx context.Context, params IsmartLoginParams) (*VerifyOTPResult, error) {
	account := strings.TrimSpace(params.Account)
	password := strings.TrimSpace(params.Password)
	if account == "" || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "ismart account and password are required")
	}

	phone := strings.TrimSpace(params.Phone)
	ismartMsg, loginAccount, err := s.loginPOSWithAccountFallback(ctx, account, phone, password)
	if err != nil {
		return nil, err
	}
	relayToken, _ := s.callPOSRelayLogin(ctx, loginAccount, password)

	if phone == "" {
		phone = strings.TrimSpace(ismartMsg.Phone)
	}
	if phone == "" && looksLikePhoneAccount(loginAccount) {
		phone = loginAccount
	}
	if phone == "" && looksLikePhoneAccount(account) {
		phone = account
	}
	email := params.Email
	if email == "" && strings.Contains(account, "@") {
		email = account
	}
	return s.upsertUserByIsmart(ctx, ismartMsg, phone, email, relayToken, password)
}

// 25.1 RequestEmailPasswordReset sends a reset code to an existing email credential.
func (s *AuthService) RequestEmailPasswordReset(ctx context.Context, params EmailOTPParams) (*RequestOTPResult, error) {
	email := normalizeEmail(params.Email)
	if !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}

	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("email = ?", email).First(&credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "email account not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load email account")
	}

	return s.RequestEmailOTP(ctx, EmailOTPParams{
		Email: email,
		Scene: passwordResetEmailScene,
	})
}

// 25.2 ResetPasswordWithEmail verifies the reset code and stores a new password.
func (s *AuthService) ResetPasswordWithEmail(ctx context.Context, params PasswordResetParams) (*PasswordResetResult, error) {
	email := normalizeEmail(params.Email)
	code := strings.TrimSpace(params.Code)
	password := strings.TrimSpace(params.Password)
	if !isValidEmail(email) || code == "" || len(password) < 8 {
		return nil, errcode.New(errcode.CodeValidationError, "valid email, code, and password are required")
	}

	key := emailOTPKey(email, passwordResetEmailScene)
	record, ok := s.runtime.OTPStore.Get(key)
	if !ok || s.runtime.Now().After(record.ExpiresAt) || code != record.Code {
		return nil, errcode.New(errcode.CodeValidationError, "email otp is invalid or expired")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	passwordEncrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var credential model.UserCredential
		if err := tx.Where("email = ?", email).First(&credential).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.New(errcode.CodeValidationError, "email account not found")
			}
			return err
		}

		return tx.Model(&credential).Updates(map[string]any{
			"password_hash":      passwordHash,
			"password_encrypted": passwordEncrypted,
			"is_verified":        true,
		}).Error
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to reset password")
	}

	s.runtime.OTPStore.Delete(key)
	return &PasswordResetResult{PasswordReset: true}, nil
}

// 26. BindIsmartAccount links POS Web credentials to the current AJO user.
func (s *AuthService) BindIsmartAccount(ctx context.Context, userID int64, params IsmartLoginParams) (*VerifyOTPResult, error) {
	account := strings.TrimSpace(params.Account)
	password := strings.TrimSpace(params.Password)
	if account == "" || password == "" {
		return nil, errcode.New(errcode.CodeValidationError, "ismart account and password are required")
	}

	phone := strings.TrimSpace(params.Phone)
	ismartMsg, loginAccount, err := s.loginPOSWithAccountFallback(ctx, account, phone, password)
	if err != nil {
		return nil, err
	}
	relayToken, _ := s.callPOSRelayLogin(ctx, loginAccount, password)

	if phone == "" {
		phone = strings.TrimSpace(ismartMsg.Phone)
	}
	if phone == "" && looksLikePhoneAccount(loginAccount) {
		phone = loginAccount
	}
	if phone == "" && looksLikePhoneAccount(account) {
		phone = account
	}
	normalizedEmail := normalizeEmail(params.Email)
	if normalizedEmail == "" {
		normalizedEmail = normalizeEmail(ismartMsg.Email)
	}
	if normalizedEmail == "" && strings.Contains(account, "@") {
		normalizedEmail = normalizeEmail(account)
	}
	if normalizedEmail != "" && !isValidEmail(normalizedEmail) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}

	var user model.User
	var profile model.UserProfile
	phoneCountryCode, phoneNumber := normalizeOptionalIsmartPhone(phone)
	ismartMsg.Phone = joinPhone(phoneCountryCode, phoneNumber)
	ismartMsg.Email = normalizedEmail

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		var existing model.UserIsmartAccount
		findErr := tx.Where("ismart_user_id = ?", ismartMsg.UserID).First(&existing).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		if existing.UserID > 0 && existing.UserID != userID {
			return errcode.New(errcode.CodeValidationError, "ismart account is already linked")
		}

		userUpdates := map[string]any{"member_type": normalizeMemberType(user.MemberType)}
		if isPlaceholderPhone(user.PhoneCountryCode) && phoneCountryCode != "" && phoneNumber != "" {
			if err := s.ensurePhoneAvailableForIsmart(ctx, tx, user.ID, phoneCountryCode, phoneNumber); err != nil {
				return err
			}
			userUpdates["phone_country_code"] = phoneCountryCode
			userUpdates["phone_number"] = phoneNumber
			userUpdates["is_verified_phone"] = true
		}
		if err := tx.Model(&user).Updates(userUpdates).Error; err != nil {
			return err
		}
		user.MemberType = normalizeMemberType(user.MemberType)

		if err := s.saveIsmartAccount(ctx, tx, user.ID, ismartMsg, relayToken, password); err != nil {
			return err
		}
		if err := s.saveIsmartCredential(ctx, tx, user.ID, ismartMsg, password); err != nil {
			return err
		}
		nextProfile, err := s.syncProfileIsmartBindings(ctx, tx, user.ID, ismartMsg)
		if err != nil {
			return err
		}
		profile = *nextProfile
		return nil
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to bind ismart account")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 27. Logout returns a stable success path without token invalidation storage.
func (s *AuthService) Logout(context.Context, int64) error {
	return nil
}

// 28. AuthenticateToken parses and validates a bearer token.
func (s *AuthService) AuthenticateToken(ctx context.Context, tokenString string) (*AuthIdentity, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessTokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.runtime.Config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errcode.New(errcode.CodeAuthRequired, "invalid access token")
	}

	claims, ok := token.Claims.(*accessTokenClaims)
	if !ok || claims.TokenType != tokenTypeAccess {
		return nil, errcode.New(errcode.CodeAuthRequired, "invalid access token")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeAuthRequired, "invalid access token")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load current user")
	}

	var profile model.UserProfile
	err = s.runtime.DB.WithContext(ctx).Where("user_id = ?", user.ID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load current user profile")
	}
	if normalizeAccountType(profile.AccountType) == AccountTypeAgencyCompanySubaccount {
		if err := NewAgencyCompanyService(s.runtime).ValidateSubaccountAccess(ctx, user.ID); err != nil {
			return nil, err
		}
	}

	access := buildAccessSnapshot(user.IsStaff)
	permissions := access.Permissions
	if normalizeAccountType(profile.AccountType) == AccountTypeAgencyCompanySubaccount {
		permissions, err = NewAgencyCompanyService(s.runtime).LoadSubaccountPermissions(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	return &AuthIdentity{
		UserID:             user.ID,
		PublicID:           user.PublicID,
		MemberStatus:       user.MemberStatus,
		MemberType:         normalizeMemberType(user.MemberType),
		IsStaff:            access.IsStaff,
		Role:               resolveUserRole(user.MemberType, access.IsStaff),
		Roles:              access.RoleCodes,
		Permissions:        permissions,
		PrimaryCommunityID: profile.PrimaryCommunityID,
		AccountType:        normalizeAccountType(profile.AccountType),
	}, nil
}

// 28.1 normalizeRegistrationAccountType validates public registration account types.
func normalizeRegistrationAccountType(value string) string {
	switch strings.TrimSpace(value) {
	case "", AccountTypePersonal:
		return AccountTypePersonal
	case AccountTypeIndividualAgent, AccountTypeAgencyCompany:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

// 28.2 derivedPublisherIdentity derives the compatibility field from account type.
func derivedPublisherIdentity(accountType string) string {
	if accountType == AccountTypePersonal {
		return "owner"
	}
	return "agent"
}

// 28.3 ismartLegalEntity derives iSmart legal entity classification from AJO account type.
func ismartLegalEntity(accountType string) string {
	if accountType == AccountTypeIndividualAgent || accountType == AccountTypeAgencyCompany {
		return "LE"
	}
	return "NA"
}

// 29. upsertUserByPhone creates or updates a user during OTP verify.
func (s *AuthService) upsertUserByPhone(ctx context.Context, countryCode string, phoneNumber string) (*VerifyOTPResult, error) {
	var user model.User
	var profile model.UserProfile

	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		findErr := tx.Where("phone_country_code = ? AND phone_number = ?", countryCode, phoneNumber).First(&user).Error
		if findErr != nil {
			if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}

			user = model.User{
				PublicID:         utils.NewPublicID(),
				PhoneCountryCode: countryCode,
				PhoneNumber:      phoneNumber,
				MemberStatus:     "active",
				MemberType:       MemberTypeUser,
				IsStaff:          s.isBootstrapStaffPhone(countryCode, phoneNumber),
				IsVerifiedPhone:  true,
			}

			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			profile = model.UserProfile{
				UserID: user.ID,
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}

			return nil
		}

		if err := tx.Model(&user).Updates(map[string]any{
			"is_verified_phone": true,
			"member_type":       normalizeMemberType(user.MemberType),
			"is_staff":          user.IsStaff || s.isBootstrapStaffPhone(countryCode, phoneNumber),
		}).Error; err != nil {
			return err
		}

		user.MemberType = normalizeMemberType(user.MemberType)
		user.IsStaff = user.IsStaff || s.isBootstrapStaffPhone(countryCode, phoneNumber)

		profileErr := tx.Where("user_id = ?", user.ID).First(&profile).Error
		if errors.Is(profileErr, gorm.ErrRecordNotFound) {
			profile = model.UserProfile{UserID: user.ID}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
		}

		if profileErr != nil {
			return profileErr
		}

		return nil
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to verify otp")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 30. upsertUserByEmail creates or updates a user during email OTP verify.
func (s *AuthService) upsertUserByEmail(ctx context.Context, email string, displayName string) (*VerifyOTPResult, error) {
	var user model.User
	var profile model.UserProfile

	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var credential model.UserCredential
		findErr := tx.Where("email = ?", email).First(&credential).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		if credential.UserID > 0 {
			if err := tx.First(&user, credential.UserID).Error; err != nil {
				return err
			}
			if err := tx.Model(&credential).Update("is_verified", true).Error; err != nil {
				return err
			}
		} else {
			user = model.User{
				PublicID:         utils.NewPublicID(),
				PhoneCountryCode: "email",
				PhoneNumber:      utils.NewPublicID(),
				MemberStatus:     "active",
				MemberType:       MemberTypeUser,
				IsStaff:          false,
				IsVerifiedPhone:  false,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			credential = model.UserCredential{
				UserID:     user.ID,
				Email:      &email,
				IsVerified: true,
			}
			if err := tx.Create(&credential).Error; err != nil {
				return err
			}
		}

		profileErr := tx.Where("user_id = ?", user.ID).First(&profile).Error
		if errors.Is(profileErr, gorm.ErrRecordNotFound) {
			profile = model.UserProfile{
				UserID:      user.ID,
				DisplayName: strings.TrimSpace(displayName),
			}
			if profile.DisplayName == "" {
				profile.DisplayName = email
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
		} else if profileErr != nil {
			return profileErr
		}

		return nil
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to verify email otp")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 31. upsertUserByIsmart creates or updates a user from POS Web data.
func (s *AuthService) upsertUserByIsmart(ctx context.Context, ismartMsg *IsmartMessage, phone string, email string, relayToken string, password string) (*VerifyOTPResult, error) {
	if ismartMsg == nil || ismartMsg.UserID <= 0 || strings.TrimSpace(ismartMsg.Username) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "invalid ismart account response")
	}

	var user model.User
	var profile model.UserProfile
	var account model.UserIsmartAccount
	phoneCountryCode, phoneNumber := normalizeOptionalIsmartPhone(phone)
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" {
		normalizedEmail = normalizeEmail(ismartMsg.Email)
	}
	if normalizedEmail != "" && !isValidEmail(normalizedEmail) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}
	ismartMsg.Phone = joinPhone(phoneCountryCode, phoneNumber)
	ismartMsg.Email = normalizedEmail

	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		findErr := tx.Where("ismart_user_id = ?", ismartMsg.UserID).First(&account).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		if account.UserID > 0 {
			if err := tx.First(&user, account.UserID).Error; err != nil {
				return err
			}
		}

		if user.ID == 0 {
			ajoIsStaff := s.isBootstrapStaffPhone(phoneCountryCode, phoneNumber)
			user = model.User{
				PublicID:         utils.NewPublicID(),
				PhoneCountryCode: phoneCountryCode,
				PhoneNumber:      phoneNumber,
				MemberStatus:     "active",
				MemberType:       MemberTypeUser,
				IsStaff:          ajoIsStaff,
				IsVerifiedPhone:  phoneCountryCode != "" && phoneNumber != "",
			}
			if user.PhoneCountryCode == "" {
				user.PhoneCountryCode = "ismart"
				user.PhoneNumber = fmt.Sprintf("%d", ismartMsg.UserID)
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else {
			userUpdates := map[string]any{"member_type": normalizeMemberType(user.MemberType)}
			if account.UserID > 0 && phoneCountryCode != "" && phoneNumber != "" {
				if err := s.ensurePhoneAvailableForIsmart(ctx, tx, user.ID, phoneCountryCode, phoneNumber); err != nil {
					return err
				}
				userUpdates["phone_country_code"] = phoneCountryCode
				userUpdates["phone_number"] = phoneNumber
				userUpdates["is_verified_phone"] = true
			}
			if err := tx.Model(&user).Updates(userUpdates).Error; err != nil {
				return err
			}
			user.MemberType = normalizeMemberType(user.MemberType)
		}

		profileErr := tx.Where("user_id = ?", user.ID).First(&profile).Error
		if errors.Is(profileErr, gorm.ErrRecordNotFound) {
			profile = model.UserProfile{
				UserID:      user.ID,
				DisplayName: strings.TrimSpace(ismartMsg.Username),
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
		} else if profileErr != nil {
			return profileErr
		} else if strings.TrimSpace(profile.DisplayName) == "" {
			if err := tx.Model(&profile).Update("display_name", strings.TrimSpace(ismartMsg.Username)).Error; err != nil {
				return err
			}
			profile.DisplayName = strings.TrimSpace(ismartMsg.Username)
		}

		if err := s.saveIsmartAccount(ctx, tx, user.ID, ismartMsg, relayToken, password); err != nil {
			return err
		}
		if err := s.saveIsmartCredential(ctx, tx, user.ID, ismartMsg, password); err != nil {
			return err
		}
		nextProfile, err := s.syncProfileIsmartBindings(ctx, tx, user.ID, ismartMsg)
		if err != nil {
			return err
		}
		profile = *nextProfile

		return nil
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to login with ismart account")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 32. issueAuthResult returns a token payload from the is_staff account flag.
func (s *AuthService) issueAuthResult(ctx context.Context, user *model.User, profile *model.UserProfile) (*VerifyOTPResult, error) {
	access := buildAccessSnapshot(user.IsStaff)
	permissions := access.Permissions
	if normalizeAccountType(profile.AccountType) == AccountTypeAgencyCompanySubaccount {
		var err error
		permissions, err = NewAgencyCompanyService(s.runtime).LoadSubaccountPermissions(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	accessToken, refreshToken, err := s.issueTokens(user.ID, user.MemberType, access.IsStaff)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to issue access token")
	}

	return &VerifyOTPResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		User: AuthUserResponse{
			PublicID:         user.PublicID,
			MemberStatus:     user.MemberStatus,
			MemberType:       normalizeMemberType(user.MemberType),
			IsStaff:          access.IsStaff,
			Role:             resolveUserRole(user.MemberType, access.IsStaff),
			Roles:            access.RoleCodes,
			Permissions:      permissions,
			ProfileCompleted: isProfileCompleted(profile),
			AccountType:      normalizeAccountType(profile.AccountType),
			IsmartMsg:        s.loadIsmartMessage(ctx, user.ID),
			IsmartRaw:        s.loadIsmartRaw(ctx, user.ID),
		},
	}, nil
}

// 33. issueTokens creates access and refresh JWT tokens.
func (s *AuthService) issueTokens(userID int64, memberType string, isStaff bool) (string, string, error) {
	accessToken, err := s.signToken(userID, memberType, isStaff, tokenTypeAccess, accessTokenTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.signToken(userID, memberType, isStaff, tokenTypeRefresh, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// 34. signToken signs a JWT token with the configured token type.
func (s *AuthService) signToken(userID int64, memberType string, isStaff bool, tokenType string, ttl time.Duration) (string, error) {
	now := s.runtime.Now()
	claims := &accessTokenClaims{
		UserID:     userID,
		MemberType: normalizeMemberType(memberType),
		IsStaff:    isStaff,
		Role:       resolveUserRole(memberType, isStaff),
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.runtime.Config.JWTSecret))
}

// 35. loginPOSWithAccountFallback tries account and optional Hong Kong mobile candidates once.
func (s *AuthService) loginPOSWithAccountFallback(ctx context.Context, account string, phone string, password string) (*IsmartMessage, string, error) {
	attempts := buildIsmartLoginAttempts(account, phone)
	for _, attempt := range attempts {
		ismartMsg, err := s.callPOSLogin(ctx, attempt, password)
		if err == nil {
			return ismartMsg, attempt, nil
		}
		if !isIsmartCredentialError(err) {
			return nil, "", err
		}
	}

	return nil, "", errcode.New(errcode.CodeValidationError, "ismart account or password is incorrect")
}

// 36. buildIsmartLoginAttempts returns stable POS login account candidates.
func buildIsmartLoginAttempts(values ...string) []string {
	attempts := make([]string, 0, len(values)*2)
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}

		phoneNumber, isPhone := normalizeHongKongMobileAccount(value)
		if isPhone {
			attempts = append(attempts, phoneNumber, value)
			continue
		}

		attempts = append(attempts, value)
	}

	return normalizeStringSlice(attempts)
}

// 37. isIsmartCredentialError reports whether POS rejected only the credentials.
func isIsmartCredentialError(err error) bool {
	var appErr *errcode.AppError
	return errors.As(err, &appErr) &&
		appErr.Code == errcode.CodeValidationError &&
		appErr.Message == "ismart account or password is incorrect"
}

// 38. callPOSLogin sends the credential check to the configured POS Web endpoint.
func (s *AuthService) callPOSLogin(ctx context.Context, account string, password string) (*IsmartMessage, error) {
	loginURL := strings.TrimSpace(s.runtime.Config.POSLoginURL)
	if loginURL == "" {
		baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.POSAPIBaseURL), "/")
		if baseURL == "" {
			baseURL = "https://pos.ismart.skylinedances.com/api"
		}
		loginURL = baseURL + "/poslogin"
	}

	usernameField := strings.TrimSpace(s.runtime.Config.POSLoginUsernameField)
	if usernameField == "" {
		usernameField = "login_name"
	}
	passwordField := strings.TrimSpace(s.runtime.Config.POSLoginPasswordField)
	if passwordField == "" {
		passwordField = "password"
	}

	loginType := "username"
	if looksLikePhoneAccount(account) {
		loginType = "phone"
	}
	payload := map[string]string{
		usernameField: account,
		passwordField: password,
		"login_type":  loginType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare pos login request")
	}

	timeout := s.runtime.Config.POSLoginTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare pos login request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-POS-Client", "pos-web")

	response, err := client.Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call pos login")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return nil, errcode.New(errcode.CodeValidationError, "ismart account or password is incorrect")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call pos login")
	}

	rawResponse, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read pos login response")
	}
	var result IsmartLoginResponse
	if err := json.Unmarshal(rawResponse, &result); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "invalid pos login response")
	}
	code := strings.TrimSpace(paymentStringValue(result.Code))
	message := strings.ToLower(strings.TrimSpace(result.Message))
	success := code == "1" || code == "200" || message == "success" || message == "ok"
	if !success || result.Msg == nil {
		return nil, errcode.New(errcode.CodeValidationError, "ismart account or password is incorrect")
	}
	result.Msg.RawMessage = ismartRawResponseSection(rawResponse, "msg")

	return normalizeIsmartMessage(result.Msg), nil
}

// 39. callPOSRelayLogin obtains the relay token used for member POS data reads.
func (s *AuthService) callPOSRelayLogin(ctx context.Context, account string, password string) (string, error) {
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if account == "" || password == "" {
		return "", errcode.New(errcode.CodeValidationError, "ismart account and password are required")
	}

	loginType := "username"
	if looksLikePhoneAccount(account) {
		loginType = "phone"
	}
	payload := map[string]string{
		"login_name": account,
		"password":   password,
		"login_type": loginType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to prepare pos relay login request")
	}

	timeout := s.runtime.Config.POSLoginTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, posRelayURL(s.runtime.Config.POSAPIBaseURL, "/login"), bytes.NewReader(body))
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to prepare pos relay login request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := (&http.Client{Timeout: timeout}).Do(request)
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to call pos relay login")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", errcode.New(errcode.CodeValidationError, "ismart account or password is incorrect")
	}

	var result posRelayLoginResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", errcode.New(errcode.CodeInternalError, "invalid pos relay login response")
	}
	if strings.TrimSpace(result.Token) == "" {
		return "", errcode.New(errcode.CodeInternalError, "pos relay token is empty")
	}

	return strings.TrimSpace(result.Token), nil
}

// 40. saveIsmartAccount persists the latest POS Web identity snapshot.
func (s *AuthService) saveIsmartAccount(ctx context.Context, tx *gorm.DB, userID int64, ismartMsg *IsmartMessage, relayToken string, password string) error {
	building, err := marshalJSON(ismartMsg.Building)
	if err != nil {
		return err
	}
	staffBuildingPermissions, err := marshalJSON(ismartMsg.StaffBuildingPermissions)
	if err != nil {
		return err
	}
	clientBuildingPermissions, err := marshalJSON(ismartMsg.ClientBuildingPermissions)
	if err != nil {
		return err
	}
	clientBuildingFlatUnitsPermissions, err := marshalJSON(ismartMsg.ClientBuildingFlatUnitsPermissions)
	if err != nil {
		return err
	}
	var existing model.UserIsmartAccount
	if result := tx.WithContext(ctx).Where("user_id = ?", userID).First(&existing); result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	rawMessage, err := marshalJSON(mergeIsmartRawMessages(ismartRawMessage(existing.RawMessage), ismartMsg.RawMessage))
	if err != nil {
		return err
	}
	profileMessage, err := marshalJSON(mergeIsmartRawMessages(ismartProfileSnapshot(existing), ismartMsg.ProfileSnapshot))
	if err != nil {
		return err
	}
	relayTokenEncrypted := ""
	var relayTokenSyncedAt *time.Time
	if strings.TrimSpace(relayToken) != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, strings.TrimSpace(relayToken))
		if err != nil {
			return err
		}
		now := s.runtime.Now()
		relayTokenEncrypted = encrypted
		relayTokenSyncedAt = &now
	}
	passwordEncrypted := ""
	if strings.TrimSpace(password) != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, strings.TrimSpace(password))
		if err != nil {
			return err
		}
		passwordEncrypted = encrypted
	}

	account := model.UserIsmartAccount{
		UserID:                             userID,
		IsmartUserID:                       ismartMsg.UserID,
		Username:                           strings.TrimSpace(ismartMsg.Username),
		Email:                              normalizeEmail(ismartMsg.Email),
		Phone:                              strings.TrimSpace(ismartMsg.Phone),
		IsStaff:                            ismartMsg.IsStaff,
		Building:                           building,
		StaffBuildingPermissions:           staffBuildingPermissions,
		ClientBuildingPermissions:          clientBuildingPermissions,
		ClientBuildingFlatUnitsPermissions: clientBuildingFlatUnitsPermissions,
		RawMessage:                         rawMessage,
		ProfileSnapshot:                    profileMessage,
		RelayTokenEncrypted:                relayTokenEncrypted,
		RelayTokenSyncedAt:                 relayTokenSyncedAt,
		PasswordEncrypted:                  passwordEncrypted,
	}
	return tx.WithContext(ctx).Where("user_id = ?", userID).Assign(account).FirstOrCreate(&account).Error
}

// 41. saveIsmartCredential syncs POS login credentials to local password login.
func (s *AuthService) saveIsmartCredential(ctx context.Context, tx *gorm.DB, userID int64, ismartMsg *IsmartMessage, password string) error {
	username := normalizeUsername(ismartMsg.Username)
	email := normalizeEmail(ismartMsg.Email)
	passwordValue := strings.TrimSpace(password)
	if !isValidUsername(username) || passwordValue == "" {
		return nil
	}

	passwordHash, err := utils.HashPassword(passwordValue)
	if err != nil {
		return err
	}
	passwordEncrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, passwordValue)
	if err != nil {
		return err
	}

	usernamePtr := &username
	var emailPtr *string
	if isValidEmail(email) {
		emailPtr = &email
	}

	var credential model.UserCredential
	findErr := tx.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return findErr
	}
	if credential.UserID == 0 {
		credential = model.UserCredential{UserID: userID}
	}

	if err := s.ensureCredentialOwner(ctx, tx, "username", username, userID); err == nil {
		credential.Username = usernamePtr
	}
	if emailPtr != nil {
		if err := s.ensureCredentialOwner(ctx, tx, "email", email, userID); err == nil {
			credential.Email = emailPtr
		}
	}
	credential.PasswordHash = passwordHash
	credential.PasswordEncrypted = passwordEncrypted
	credential.IsVerified = true

	return tx.WithContext(ctx).Save(&credential).Error
}

// 42. ensureCredentialOwner checks whether a local credential value is free or owned by this user.
func (s *AuthService) ensureCredentialOwner(ctx context.Context, tx *gorm.DB, column string, value string, userID int64) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	var credential model.UserCredential
	query := tx.WithContext(ctx)
	switch column {
	case "username":
		query = query.Where("username = ?", value)
	case "email":
		query = query.Where("email = ?", value)
	default:
		return errcode.New(errcode.CodeValidationError, "invalid credential column")
	}
	err := query.First(&credential).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && credential.UserID == userID) {
		return nil
	}
	if err != nil {
		return err
	}

	return errcode.New(errcode.CodeValidationError, column+" is already registered")
}

// 43. syncProfileIsmartBindings stores POS building and unit bindings on the local profile.
func (s *AuthService) syncProfileIsmartBindings(ctx context.Context, tx *gorm.DB, userID int64, ismartMsg *IsmartMessage) (*model.UserProfile, error) {
	var profile model.UserProfile
	profileErr := tx.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(profileErr, gorm.ErrRecordNotFound) {
		profile = model.UserProfile{
			UserID:      userID,
			DisplayName: strings.TrimSpace(ismartMsg.Username),
		}
		if err := tx.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, err
		}
	} else if profileErr != nil {
		return nil, profileErr
	}

	updates := map[string]any{}
	if strings.TrimSpace(profile.DisplayName) == "" && strings.TrimSpace(ismartMsg.Username) != "" {
		updates["display_name"] = strings.TrimSpace(ismartMsg.Username)
	}
	if hasLocalProfileBinding(&profile) {
		if len(updates) > 0 {
			if err := tx.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return nil, err
		}

		return &profile, nil
	}

	boundBuildingIDs := resolveIsmartBoundBuildings(ismartMsg)
	boundFlatUnitIDs := resolveIsmartBoundUnits(ismartMsg)
	buildingJSON, err := marshalJSON(boundBuildingIDs)
	if err != nil {
		return nil, err
	}
	unitJSON, err := marshalJSON(boundFlatUnitIDs)
	if err != nil {
		return nil, err
	}

	updates["bound_building_ids"] = buildingJSON
	updates["bound_flat_unit_ids"] = unitJSON
	if profile.ResidenceBindingStatus == residenceBindingStatusPending && (len(boundBuildingIDs) > 0 || len(boundFlatUnitIDs) > 0) {
		updates["residence_binding_status"] = "approved"
	}
	if len(boundBuildingIDs) > 0 {
		community, err := s.resolveRegistrationCommunity(ctx, tx, boundBuildingIDs[0], boundBuildingIDs[0])
		if err != nil {
			return nil, err
		}
		updates["primary_community_id"] = community.ID
		updates["district_code"] = community.DistrictCode
	}
	if len(boundFlatUnitIDs) > 0 {
		floor, unit := splitPOSFlatUnitID(boundFlatUnitIDs[0])
		updates["residence_floor"] = floor
		updates["residence_unit"] = unit
	}

	if err := tx.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}

	return &profile, nil
}

// 43.1 hasLocalProfileBinding returns whether the member center already owns residence data.
func hasLocalProfileBinding(profile *model.UserProfile) bool {
	if profile == nil {
		return false
	}
	if profile.ResidenceBindingStatus == residenceBindingStatusPending {
		return false
	}
	if profile.PrimaryCommunityID != nil {
		return true
	}
	if len(normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs))) > 0 {
		return true
	}
	if len(normalizeStringSlice(unmarshalStringSlice(profile.BoundFlatUnitIDs))) > 0 {
		return true
	}

	return strings.TrimSpace(profile.ResidenceFloor) != "" || strings.TrimSpace(profile.ResidenceUnit) != ""
}

// 44. resolveIsmartBoundBuildings returns resident buildings allowed by POS.
func resolveIsmartBoundBuildings(message *IsmartMessage) []string {
	if message == nil {
		return []string{}
	}

	return normalizeStringSlice(message.ClientBuildingPermissions)
}

// 45. resolveIsmartBoundUnits returns client flat unit bindings.
func resolveIsmartBoundUnits(message *IsmartMessage) []string {
	if message == nil {
		return []string{}
	}

	return normalizeStringSlice(message.ClientBuildingFlatUnitsPermissions)
}

// 46. splitPOSFlatUnitID derives floor and unit from a POS flat unit id.
func splitPOSFlatUnitID(unitID string) (string, string) {
	value := strings.TrimSpace(unitID)
	if len(value) <= 7 {
		return "", value
	}
	suffix := strings.TrimLeft(value[7:], "0")
	if suffix == "" {
		return "", ""
	}
	if len(suffix) <= 2 {
		return "", suffix
	}

	return suffix[:len(suffix)-2], suffix[len(suffix)-2:]
}

// 47. isProfileCompleted checks whether account profile has payment-ready bindings.
func isProfileCompleted(profile *model.UserProfile) bool {
	if profile == nil {
		return false
	}
	if len(unmarshalStringSlice(profile.BoundBuildingIDs)) > 0 || len(unmarshalStringSlice(profile.BoundFlatUnitIDs)) > 0 {
		return true
	}

	return profile.PrimaryCommunityID != nil && profile.ResidenceBindingStatus != residenceBindingStatusPending
}

// 48. isPlaceholderPhone checks locally generated non-phone login placeholders.
func isPlaceholderPhone(countryCode string) bool {
	value := strings.ToLower(strings.TrimSpace(countryCode))
	return value == "" || value == "email" || value == "ismart"
}

// 49. ensurePhoneAvailableForIsmart keeps ismart identity authoritative when rebinding a phone.
func (s *AuthService) ensurePhoneAvailableForIsmart(ctx context.Context, tx *gorm.DB, userID int64, countryCode string, phoneNumber string) error {
	var count int64
	if err := tx.WithContext(ctx).Model(&model.User{}).
		Where("id <> ? AND phone_country_code = ? AND phone_number = ?", userID, countryCode, phoneNumber).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errcode.New(errcode.CodeValidationError, "phone number is already registered")
	}

	return nil
}

// 50. resolveRegistrationCommunity loads or mirrors a POS building as local community.
func (s *AuthService) resolveRegistrationCommunity(ctx context.Context, tx *gorm.DB, publicID string, name string) (*model.Community, error) {
	community, err := s.findCommunityByPublicID(ctx, tx, publicID)
	if err == nil {
		return community, nil
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) {
		return nil, err
	}
	if len(publicID) == 0 || len(publicID) > 26 {
		return nil, appErr
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = publicID
	}

	community = &model.Community{
		PublicID:      publicID,
		CommunityType: "building",
		NameZH:        displayName,
		NameEN:        displayName,
		DistrictCode:  "unknown",
		AddressText:   displayName,
	}
	if err := tx.WithContext(ctx).Create(community).Error; err != nil {
		found, findErr := s.findCommunityByPublicID(ctx, tx, publicID)
		if findErr == nil {
			return found, nil
		}
		return nil, err
	}

	return community, nil
}

// 51. findCommunityByPublicID loads a community for registration profile binding.
func (s *AuthService) findCommunityByPublicID(ctx context.Context, tx *gorm.DB, publicID string) (*model.Community, error) {
	var community model.Community
	if err := tx.WithContext(ctx).Where("public_id = ?", publicID).First(&community).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "community not found")
		}
		return nil, err
	}

	return &community, nil
}

// 52. loadIsmartRaw returns the sanitized iSmart upstream business payload.
func (s *AuthService) loadIsmartRaw(ctx context.Context, userID int64) map[string]any {
	var account model.UserIsmartAccount
	result := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).Limit(1).Find(&account)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil
	}

	return ismartRawMessage(account.RawMessage)
}

// 53. loadIsmartMessage returns the linked POS Web payload for auth responses.
func (s *AuthService) loadIsmartMessage(ctx context.Context, userID int64) *IsmartMessage {
	var account model.UserIsmartAccount
	result := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).Limit(1).Find(&account)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil
	}

	return &IsmartMessage{
		UserID:                             account.IsmartUserID,
		Username:                           account.Username,
		Email:                              account.Email,
		Phone:                              account.Phone,
		IsStaff:                            account.IsStaff,
		Building:                           unmarshalStringSlice(account.Building),
		StaffBuildingPermissions:           unmarshalStringSlice(account.StaffBuildingPermissions),
		ClientBuildingPermissions:          unmarshalStringSlice(account.ClientBuildingPermissions),
		ClientBuildingFlatUnitsPermissions: unmarshalStringSlice(account.ClientBuildingFlatUnitsPermissions),
	}
}

// 53. normalizeIsmartMessage keeps POS Web array fields non-null.
func normalizeIsmartMessage(message *IsmartMessage) *IsmartMessage {
	if message == nil {
		return nil
	}

	message.Username = strings.TrimSpace(message.Username)
	message.Email = normalizeEmail(message.Email)
	message.Phone = strings.TrimSpace(message.Phone)
	message.Building = normalizeStringSlice(message.Building)
	message.StaffBuildingPermissions = normalizeStringSlice(message.StaffBuildingPermissions)
	message.ClientBuildingPermissions = normalizeStringSlice(message.ClientBuildingPermissions)
	message.ClientBuildingFlatUnitsPermissions = normalizeStringSlice(message.ClientBuildingFlatUnitsPermissions)
	message.IsStaff = resolveIsmartStaffStatus(message)
	return message
}

// 55. resolveIsmartStaffStatus normalizes the upstream iSmart staff status.
func resolveIsmartStaffStatus(message *IsmartMessage) bool {
	if message == nil {
		return false
	}
	if message.IsStaff {
		return true
	}

	return len(normalizeStringSlice(message.StaffBuildingPermissions)) > 0
}

// 56. normalizeStringSlice splits, trims, deduplicates, and keeps empty arrays stable.
func normalizeStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, item := range values {
		parts := strings.FieldsFunc(item, func(r rune) bool {
			return r == ',' || r == '，' || r == '\n' || r == '\r'
		})
		for _, part := range parts {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

// 57. normalizeOptionalIsmartPhone parses an optional phone string for local account linking.
func normalizeOptionalIsmartPhone(phone string) (string, string) {
	value := strings.TrimSpace(phone)
	if value == "" {
		return "", ""
	}
	if phoneNumber, ok := normalizeHongKongMobileAccount(value); ok {
		return "+852", phoneNumber
	}

	parts := strings.Fields(value)
	if len(parts) >= 2 {
		countryCode := normalizePhoneCountryCode(parts[0])
		phoneNumber := normalizePhoneNumber(strings.Join(parts[1:], ""))
		if countryCode == "+852" {
			if normalizedPhoneNumber, ok := normalizeHongKongMobileAccount(countryCode + phoneNumber); ok {
				return countryCode, normalizedPhoneNumber
			}
			return "", ""
		}
		if isValidPhone(countryCode, phoneNumber) {
			return countryCode, phoneNumber
		}
	}

	return "", ""
}

// 58. normalizeHongKongMobileAccount normalizes likely Hong Kong mobile input to 8 local digits.
func normalizeHongKongMobileAccount(account string) (string, bool) {
	value := strings.TrimSpace(account)
	if value == "" {
		return "", false
	}

	var digits strings.Builder
	for _, item := range value {
		if item >= '0' && item <= '9' {
			digits.WriteRune(item)
			continue
		}
		if item == '+' && digits.Len() == 0 {
			continue
		}
		if item == ' ' || item == '-' || item == '(' || item == ')' {
			continue
		}
		return "", false
	}

	phoneNumber := digits.String()
	if strings.HasPrefix(phoneNumber, "852") && len(phoneNumber) == 11 {
		phoneNumber = strings.TrimPrefix(phoneNumber, "852")
	}
	if len(phoneNumber) != 8 || !strings.ContainsRune("456789", rune(phoneNumber[0])) {
		return "", false
	}

	return phoneNumber, true
}

// 59. looksLikePhoneAccount checks whether an ismart account input is a Hong Kong mobile number.
func looksLikePhoneAccount(account string) bool {
	_, ok := normalizeHongKongMobileAccount(account)
	return ok
}

// 60. joinPhone formats a normalized phone pair for ismart payload storage.
func joinPhone(countryCode string, phoneNumber string) string {
	if countryCode == "" || phoneNumber == "" {
		return ""
	}

	return countryCode + phoneNumber
}

// 61. isBootstrapStaffPhone checks whether the verified phone should become staff.
func (s *AuthService) isBootstrapStaffPhone(countryCode string, phoneNumber string) bool {
	target := strings.TrimSpace(countryCode) + strings.TrimSpace(phoneNumber)
	if target == "" {
		return false
	}

	for _, item := range strings.Split(s.runtime.Config.BootstrapStaffPhones, ",") {
		if strings.TrimSpace(item) == target {
			return true
		}
	}

	return false
}

// 62. normalizeUsername prepares username lookup input.
func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// 63. registrationUsernameCandidate chooses a stable local username fallback.
func registrationUsernameCandidate(phoneNumber string, email string, engName string) string {
	if strings.TrimSpace(phoneNumber) != "" {
		return strings.TrimSpace(phoneNumber)
	}
	if strings.TrimSpace(email) != "" {
		localPart := strings.Split(strings.TrimSpace(email), "@")[0]
		if strings.TrimSpace(localPart) != "" {
			return localPart
		}
	}
	return strings.ReplaceAll(strings.TrimSpace(engName), " ", "")
}

// 64. isValidUsername checks the minimum username shape needed for password auth.
func isValidUsername(username string) bool {
	return len(username) >= 2 && len(username) <= 120 && !strings.ContainsAny(username, " \t\r\n")
}

// 65. isValidEnglishName checks the registration name required by iSmart.
func isValidEnglishName(name string) bool {
	return len(strings.TrimSpace(name)) >= 2 && len(strings.TrimSpace(name)) <= 120
}

// 66. normalizeEmail prepares email lookup input.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// 67. isValidEmail checks the minimum email shape needed for auth flows.
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) <= 255
}

// 68. normalizePhoneCountryCode prepares a phone country code for lookup.
func normalizePhoneCountryCode(countryCode string) string {
	value := strings.TrimSpace(countryCode)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "+") {
		return value
	}
	return "+" + value
}

// 69. normalizePhoneNumber prepares a phone number for lookup.
func normalizePhoneNumber(phoneNumber string) string {
	value := strings.TrimSpace(phoneNumber)
	value = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(value)
	return value
}

// 70. isValidPhone checks the minimum phone shape needed for password auth.
func isValidPhone(countryCode string, phoneNumber string) bool {
	if !strings.HasPrefix(countryCode, "+") || len(countryCode) > 8 || len(phoneNumber) < 4 || len(phoneNumber) > 32 {
		return false
	}
	for _, item := range strings.TrimPrefix(countryCode, "+") + phoneNumber {
		if item < '0' || item > '9' {
			return false
		}
	}
	return true
}

// 71. normalizeAuthScene returns a stable verification scene.
func normalizeAuthScene(scene string) string {
	value := strings.TrimSpace(scene)
	if value == "" {
		return "login"
	}

	return value
}

// 72. buildEmailOTPContent returns the email verification message.
func buildEmailOTPContent(code string, scene string) (string, string) {
	if normalizeAuthScene(scene) == passwordResetEmailScene {
		return "AJO Living password reset code", fmt.Sprintf("Your AJO Living password reset code is %s. It expires in 5 minutes.", code)
	}

	subject := "AJO Living email verification code"
	body := fmt.Sprintf("Your AJO Living verification code is %s. It expires in 5 minutes.", code)
	if normalizeAuthScene(scene) != "login" {
		body = fmt.Sprintf("Your AJO Living verification code for %s is %s. It expires in 5 minutes.", scene, code)
	}

	return subject, body
}

// 73. otpMockDisclosureAllowed allows mock code output only in local test environments.
func otpMockDisclosureAllowed(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "", "development", "dev", "local", "test", "testing":
		return true
	default:
		return false
	}
}

// 74. otpKey builds the in-memory OTP lookup key.
func otpKey(countryCode string, phoneNumber string, scene string) string {
	return strings.TrimSpace(countryCode) + ":" + strings.TrimSpace(phoneNumber) + ":" + normalizeAuthScene(scene)
}

// 75. normalizePhoneOTPInput normalizes form input before provider delivery and lookup.
func normalizePhoneOTPInput(countryCode string, phoneNumber string) (string, string, error) {
	normalizedCountryCode := strings.TrimSpace(countryCode)
	normalizedPhoneNumber := strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(phoneNumber))
	if len(normalizedCountryCode) < 2 || normalizedCountryCode[0] != '+' || len(normalizedCountryCode) > 8 || len(normalizedPhoneNumber) < 4 || len(normalizedPhoneNumber) > 15 {
		return "", "", fmt.Errorf("invalid phone number")
	}
	for _, value := range normalizedCountryCode[1:] + normalizedPhoneNumber {
		if value < '0' || value > '9' {
			return "", "", fmt.Errorf("invalid phone number")
		}
	}
	if normalizedCountryCode == "+86" && (len(normalizedPhoneNumber) != 11 || normalizedPhoneNumber[0] != '1') {
		return "", "", fmt.Errorf("invalid mainland China mobile number")
	}

	return normalizedCountryCode, normalizedPhoneNumber, nil
}

// 76. emailOTPKey builds the in-memory email OTP lookup key.
func emailOTPKey(email string, scene string) string {
	return "email:" + normalizeEmail(email) + ":" + normalizeAuthScene(scene)
}
