/*
 * Authentication business logic.
 * 1. Manage mock OTP request and verification flows.
 * 2. Create member accounts and issue JWT tokens.
 * 3. Authenticate bearer tokens for middleware usage.
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const accessTokenTTL = 2 * time.Hour

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
	Email            string
	Password         string
	DisplayName      string
	PhoneCountryCode string
	PhoneNumber      string
}

// 7. PhonePasswordParams defines phone password auth input.
type PhonePasswordParams struct {
	PhoneCountryCode string
	PhoneNumber      string
	Password         string
}

// 8. VerifyOTPResult defines the OTP verify result.
type VerifyOTPResult struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int              `json:"expires_in"`
	User         AuthUserResponse `json:"user"`
}

// 9. AuthUserResponse defines the auth response user payload.
type AuthUserResponse struct {
	PublicID         string   `json:"public_id"`
	MemberStatus     string   `json:"member_status"`
	MemberType       string   `json:"member_type"`
	IsStaff          bool     `json:"is_staff"`
	Role             string   `json:"role"`
	Roles            []string `json:"roles"`
	Permissions      []string `json:"permissions"`
	ProfileCompleted bool     `json:"profile_completed"`
}

// 10. AuthIdentity defines the middleware-facing auth identity.
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
}

// 11. accessTokenClaims stores custom JWT claims.
type accessTokenClaims struct {
	UserID     int64  `json:"user_id"`
	MemberType string `json:"member_type"`
	IsStaff    bool   `json:"is_staff"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// 12. NewAuthService creates an auth service instance.
func NewAuthService(runtime *Runtime) *AuthService {
	return &AuthService{runtime: runtime}
}

// 13. RequestOTP stores a mock OTP code and delegates delivery.
func (s *AuthService) RequestOTP(ctx context.Context, params RequestOTPParams) (*RequestOTPResult, error) {
	if strings.TrimSpace(params.PhoneCountryCode) == "" || strings.TrimSpace(params.PhoneNumber) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "phone number is required")
	}

	scene := strings.TrimSpace(params.Scene)
	if scene == "" {
		scene = "login"
	}

	code := s.runtime.Config.OTPMockCode
	if s.runtime.Config.OTPProvider != "mock" {
		code = utils.NewNumericCode(6)
	}

	key := otpKey(params.PhoneCountryCode, params.PhoneNumber, scene)
	expiresAt := s.runtime.Now().Add(5 * time.Minute)
	s.runtime.OTPStore.Save(key, OTPCode{
		Code:      code,
		ExpiresAt: expiresAt,
	})

	if err := s.runtime.OTPProvider.SendCode(ctx, params.PhoneCountryCode+params.PhoneNumber, scene, code); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to send otp")
	}

	result := &RequestOTPResult{
		ExpiresIn: int(time.Until(expiresAt).Seconds()),
	}

	if s.runtime.Config.OTPProvider == "mock" {
		result.MockCode = code
	}

	return result, nil
}

// 14. VerifyOTP validates the OTP code and issues access tokens.
func (s *AuthService) VerifyOTP(ctx context.Context, params VerifyOTPParams) (*VerifyOTPResult, error) {
	key := otpKey(params.PhoneCountryCode, params.PhoneNumber, params.Scene)
	record, ok := s.runtime.OTPStore.Get(key)
	if !ok || s.runtime.Now().After(record.ExpiresAt) {
		return nil, errcode.New(errcode.CodeValidationError, "otp is invalid or expired")
	}

	if strings.TrimSpace(params.Code) != record.Code {
		return nil, errcode.New(errcode.CodeValidationError, "otp is invalid or expired")
	}

	result, err := s.upsertUserByPhone(ctx, params.PhoneCountryCode, params.PhoneNumber)
	if err != nil {
		return nil, err
	}

	s.runtime.OTPStore.Delete(key)
	return result, nil
}

// 15. RequestEmailOTP stores a code and sends it to the target email.
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
	if !s.runtime.Config.MailEnabled {
		result.MockCode = code
	}

	return result, nil
}

// 16. VerifyEmailOTP validates an email code and signs the member in.
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

// 17. RegisterWithEmail creates an email and phone password account.
func (s *AuthService) RegisterWithEmail(ctx context.Context, params EmailPasswordParams) (*VerifyOTPResult, error) {
	email := normalizeEmail(params.Email)
	password := strings.TrimSpace(params.Password)
	phoneCountryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	if !isValidEmail(email) || !isValidPhone(phoneCountryCode, phoneNumber) || len(password) < 8 {
		return nil, errcode.New(errcode.CodeValidationError, "valid email, phone number, and password are required")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}

	var user model.User
	var profile model.UserProfile
	accessService := NewAccessService(s.runtime)
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

		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: phoneCountryCode,
			PhoneNumber:      phoneNumber,
			MemberStatus:     "active",
			MemberType:       MemberTypeUser,
			IsStaff:          false,
			IsVerifiedPhone:  false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		credential := model.UserCredential{
			UserID:       user.ID,
			Email:        email,
			PasswordHash: passwordHash,
			IsVerified:   true,
		}
		if err := tx.Create(&credential).Error; err != nil {
			return err
		}

		profile = model.UserProfile{
			UserID:      user.ID,
			DisplayName: strings.TrimSpace(params.DisplayName),
		}
		if profile.DisplayName == "" {
			profile.DisplayName = email
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		return accessService.EnsureUserRoles(ctx, tx, &user, nil, "")
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to register email account")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 18. LoginWithEmail validates an email password account and issues tokens.
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

// 19. LoginWithPhone validates a phone password account and issues tokens.
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

// 20. Logout returns a stable success path without token invalidation storage.
func (s *AuthService) Logout(context.Context, int64) error {
	return nil
}

// 21. AuthenticateToken parses and validates a bearer token.
func (s *AuthService) AuthenticateToken(ctx context.Context, tokenString string) (*AuthIdentity, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessTokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.runtime.Config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errcode.New(errcode.CodeAuthRequired, "invalid access token")
	}

	claims, ok := token.Claims.(*accessTokenClaims)
	if !ok {
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

	access, err := NewAccessService(s.runtime).ResolveUserAccess(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &AuthIdentity{
		UserID:             user.ID,
		PublicID:           user.PublicID,
		MemberStatus:       user.MemberStatus,
		MemberType:         normalizeMemberType(user.MemberType),
		IsStaff:            access.IsStaff,
		Role:               resolveUserRole(user.MemberType, access.IsStaff),
		Roles:              access.RoleCodes,
		Permissions:        access.Permissions,
		PrimaryCommunityID: profile.PrimaryCommunityID,
	}, nil
}

// 22. upsertUserByPhone creates or updates a user during OTP verify.
func (s *AuthService) upsertUserByPhone(ctx context.Context, countryCode string, phoneNumber string) (*VerifyOTPResult, error) {
	var user model.User
	var profile model.UserProfile
	accessService := NewAccessService(s.runtime)
	preferredStaffRole := ""
	if s.isBootstrapStaffPhone(countryCode, phoneNumber) {
		preferredStaffRole = model.RoleCodeSuperAdmin
	}

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

			return accessService.EnsureUserRoles(ctx, tx, &user, nil, preferredStaffRole)
		}

		if err := tx.Model(&user).Updates(map[string]any{
			"is_verified_phone": true,
			"member_status":     "active",
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

		return accessService.EnsureUserRoles(ctx, tx, &user, nil, preferredStaffRole)
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to verify otp")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 23. upsertUserByEmail creates or updates a user during email OTP verify.
func (s *AuthService) upsertUserByEmail(ctx context.Context, email string, displayName string) (*VerifyOTPResult, error) {
	var user model.User
	var profile model.UserProfile
	accessService := NewAccessService(s.runtime)

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
				Email:      email,
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

		return accessService.EnsureUserRoles(ctx, tx, &user, nil, "")
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to verify email otp")
	}

	return s.issueAuthResult(ctx, &user, &profile)
}

// 24. issueAuthResult resolves access and returns a token payload.
func (s *AuthService) issueAuthResult(ctx context.Context, user *model.User, profile *model.UserProfile) (*VerifyOTPResult, error) {
	access, err := NewAccessService(s.runtime).ResolveUserAccess(ctx, user)
	if err != nil {
		return nil, err
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
			Permissions:      access.Permissions,
			ProfileCompleted: profile.PrimaryCommunityID != nil,
		},
	}, nil
}

// 25. issueTokens creates access and refresh JWT tokens.
func (s *AuthService) issueTokens(userID int64, memberType string, isStaff bool) (string, string, error) {
	accessToken, err := s.signToken(userID, memberType, isStaff, accessTokenTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.signToken(userID, memberType, isStaff, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// 26. signToken signs a JWT token with the configured secret.
func (s *AuthService) signToken(userID int64, memberType string, isStaff bool, ttl time.Duration) (string, error) {
	now := s.runtime.Now()
	claims := &accessTokenClaims{
		UserID:     userID,
		MemberType: normalizeMemberType(memberType),
		IsStaff:    isStaff,
		Role:       resolveUserRole(memberType, isStaff),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.runtime.Config.JWTSecret))
}

// 27. isBootstrapStaffPhone checks whether the verified phone should become staff.
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

// 28. normalizeEmail prepares email lookup input.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// 29. isValidEmail checks the minimum email shape needed for auth flows.
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) <= 255
}

// 30. normalizePhoneCountryCode prepares a phone country code for lookup.
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

// 31. normalizePhoneNumber prepares a phone number for lookup.
func normalizePhoneNumber(phoneNumber string) string {
	value := strings.TrimSpace(phoneNumber)
	value = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(value)
	return value
}

// 32. isValidPhone checks the minimum phone shape needed for password auth.
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

// 33. normalizeAuthScene returns a stable verification scene.
func normalizeAuthScene(scene string) string {
	value := strings.TrimSpace(scene)
	if value == "" {
		return "login"
	}

	return value
}

// 34. buildEmailOTPContent returns the email verification message.
func buildEmailOTPContent(code string, scene string) (string, string) {
	subject := "AJO Living email verification code"
	body := fmt.Sprintf("Your AJO Living verification code is %s. It expires in 5 minutes.", code)
	if normalizeAuthScene(scene) != "login" {
		body = fmt.Sprintf("Your AJO Living verification code for %s is %s. It expires in 5 minutes.", scene, code)
	}

	return subject, body
}

// 35. otpKey builds the in-memory OTP lookup key.
func otpKey(countryCode string, phoneNumber string, scene string) string {
	return strings.TrimSpace(countryCode) + ":" + strings.TrimSpace(phoneNumber) + ":" + normalizeAuthScene(scene)
}

// 36. emailOTPKey builds the in-memory email OTP lookup key.
func emailOTPKey(email string, scene string) string {
	return "email:" + normalizeEmail(email) + ":" + normalizeAuthScene(scene)
}
