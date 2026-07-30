/*
 * Authentication HTTP handlers.
 * 1. Bind OTP request and verify payloads.
 * 2. Delegate unified and legacy auth flows to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. AuthHandler handles auth endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// 2. otpRequest defines the OTP request payload.
type otpRequest struct {
	PhoneCountryCode string `json:"phone_country_code" binding:"required"`
	PhoneNumber      string `json:"phone_number" binding:"required"`
	Scene            string `json:"scene"`
}

// 3. otpVerifyRequest defines the OTP verify payload.
type otpVerifyRequest struct {
	PhoneCountryCode string `json:"phone_country_code" binding:"required"`
	PhoneNumber      string `json:"phone_number" binding:"required"`
	Scene            string `json:"scene"`
	Code             string `json:"code" binding:"required"`
}

// 4. emailOTPRequest defines email OTP auth payload.
type emailOTPRequest struct {
	Email       string `json:"email" binding:"required"`
	Scene       string `json:"scene"`
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
}

// 5. emailPasswordRequest defines email password auth payload.
type emailPasswordRequest struct {
	Email                 string `json:"email"`
	Password              string `json:"password" binding:"required"`
	EngName               string `json:"eng_name"`
	ChiName               string `json:"chi_name"`
	DisplayName           string `json:"display_name"`
	PhoneCountryCode      string `json:"phone_country_code"`
	PhoneNumber           string `json:"phone_number"`
	Username              string `json:"username"`
	PublisherIdentityType string `json:"publisher_identity_type"`
	AccountType           string `json:"account_type"`
	IDCard                string `json:"id_card"`
	Remark                string `json:"remark"`
	Gender                string `json:"gender"`
	IsReceiveEmail        *bool  `json:"is_receive_email"`
	PrimaryCommunityID    string `json:"primary_community_id"`
	PrimaryCommunityName  string `json:"primary_community_name"`
	ResidenceFloor        string `json:"residence_floor"`
	ResidenceUnit         string `json:"residence_unit"`
}

// 5.1 registrationAvailabilityRequest defines identity fields checked before account creation.
type registrationAvailabilityRequest struct {
	Email            string `json:"email"`
	PhoneCountryCode string `json:"phone_country_code" binding:"required"`
	PhoneNumber      string `json:"phone_number" binding:"required"`
}

// 6. passwordResetRequest defines email password reset payload.
type passwordResetRequest struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 7. phonePasswordRequest defines phone password auth payload.
type phonePasswordRequest struct {
	PhoneCountryCode string `json:"phone_country_code" binding:"required"`
	PhoneNumber      string `json:"phone_number" binding:"required"`
	Password         string `json:"password" binding:"required"`
}

// 8. ismartLoginRequest defines POS Web auth payload.
type ismartLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// 8.1 identifierLoginRequest defines unified password login input.
type identifierLoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// 9. NewAuthHandler creates an auth handler instance.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 10. RequestOTP handles OTP request calls.
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var request otpRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.RequestOTP(c.Request.Context(), service.RequestOTPParams{
		PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:      strings.TrimSpace(request.PhoneNumber),
		Scene:            strings.TrimSpace(request.Scene),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 11. VerifyOTP handles OTP verify calls.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var request otpVerifyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.VerifyOTP(c.Request.Context(), service.VerifyOTPParams{
		PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:      strings.TrimSpace(request.PhoneNumber),
		Scene:            strings.TrimSpace(request.Scene),
		Code:             strings.TrimSpace(request.Code),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 12. RequestEmailOTP handles email OTP request calls.
func (h *AuthHandler) RequestEmailOTP(c *gin.Context) {
	var request emailOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.RequestEmailOTP(c.Request.Context(), service.EmailOTPParams{
		Email: strings.TrimSpace(request.Email),
		Scene: strings.TrimSpace(request.Scene),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 13. VerifyEmailOTP handles email OTP verify calls.
func (h *AuthHandler) VerifyEmailOTP(c *gin.Context) {
	var request emailOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Code) == "" {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.VerifyEmailOTP(c.Request.Context(), service.EmailOTPParams{
		Email:       strings.TrimSpace(request.Email),
		Scene:       strings.TrimSpace(request.Scene),
		Code:        strings.TrimSpace(request.Code),
		DisplayName: strings.TrimSpace(request.DisplayName),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 14. RequestEmailPasswordReset handles email password reset code requests.
func (h *AuthHandler) RequestEmailPasswordReset(c *gin.Context) {
	var request emailOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.RequestEmailPasswordReset(c.Request.Context(), service.EmailOTPParams{
		Email: strings.TrimSpace(request.Email),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 15. ResetPasswordWithEmail handles verified email password reset.
func (h *AuthHandler) ResetPasswordWithEmail(c *gin.Context) {
	var request passwordResetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.ResetPasswordWithEmail(c.Request.Context(), service.PasswordResetParams{
		Email:    strings.TrimSpace(request.Email),
		Code:     strings.TrimSpace(request.Code),
		Password: request.Password,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 16. RegisterEmail handles email and phone password account creation.
func (h *AuthHandler) RegisterEmail(c *gin.Context) {
	var request emailPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.RegisterWithEmail(c.Request.Context(), service.EmailPasswordParams{
		Email:                 strings.TrimSpace(request.Email),
		Password:              request.Password,
		EngName:               strings.TrimSpace(request.EngName),
		ChiName:               strings.TrimSpace(request.ChiName),
		DisplayName:           strings.TrimSpace(request.DisplayName),
		PhoneCountryCode:      strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:           strings.TrimSpace(request.PhoneNumber),
		Username:              strings.TrimSpace(request.Username),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		AccountType:           strings.TrimSpace(request.AccountType),
		IDCard:                strings.TrimSpace(request.IDCard),
		Remark:                strings.TrimSpace(request.Remark),
		Gender:                strings.TrimSpace(request.Gender),
		IsReceiveEmail:        request.IsReceiveEmail,
		PrimaryCommunityID:    strings.TrimSpace(request.PrimaryCommunityID),
		PrimaryCommunityName:  strings.TrimSpace(request.PrimaryCommunityName),
		ResidenceFloor:        strings.TrimSpace(request.ResidenceFloor),
		ResidenceUnit:         strings.TrimSpace(request.ResidenceUnit),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 16.1 CheckRegistrationAvailability checks whether registration email and phone values are already in use.
func (h *AuthHandler) CheckRegistrationAvailability(c *gin.Context) {
	var request registrationAvailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.CheckRegistrationAvailability(c.Request.Context(), service.RegistrationAvailabilityParams{
		Email:            strings.TrimSpace(request.Email),
		PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:      strings.TrimSpace(request.PhoneNumber),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 17. LoginEmail handles email password sign-in.
func (h *AuthHandler) LoginEmail(c *gin.Context) {
	var request emailPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.LoginWithEmail(c.Request.Context(), service.EmailPasswordParams{
		Email:    strings.TrimSpace(request.Email),
		Password: request.Password,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 18. LoginUsername handles username password sign-in.
func (h *AuthHandler) LoginUsername(c *gin.Context) {
	var request emailPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.LoginWithUsername(c.Request.Context(), service.EmailPasswordParams{
		Username: strings.TrimSpace(request.Username),
		Password: request.Password,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 19. LoginPhone handles phone password sign-in.
func (h *AuthHandler) LoginPhone(c *gin.Context) {
	var request phonePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.LoginWithPhone(c.Request.Context(), service.PhonePasswordParams{
		PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:      strings.TrimSpace(request.PhoneNumber),
		Password:         request.Password,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 20. LoginIsmart handles POS Web account sign-in.
func (h *AuthHandler) LoginIsmart(c *gin.Context) {
	var request ismartLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.LoginWithIsmart(c.Request.Context(), service.IsmartLoginParams{
		Account:  strings.TrimSpace(request.Account),
		Password: request.Password,
		Phone:    strings.TrimSpace(request.Phone),
		Email:    strings.TrimSpace(request.Email),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 20.1 LoginIdentifier handles unified password sign-in.
func (h *AuthHandler) LoginIdentifier(c *gin.Context) {
	var request identifierLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.LoginWithIdentifier(c.Request.Context(), service.IdentifierLoginParams{
		Identifier: strings.TrimSpace(request.Identifier),
		Password:   request.Password,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 21. Logout handles logout calls.
func (h *AuthHandler) Logout(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), user.UserID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"logged_out": true})
}
