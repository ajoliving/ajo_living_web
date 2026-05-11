/*
 * Authentication HTTP handlers.
 * 1. Bind OTP request and verify payloads.
 * 2. Delegate auth flows to the service layer.
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
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required"`
	DisplayName      string `json:"display_name"`
	PhoneCountryCode string `json:"phone_country_code"`
	PhoneNumber      string `json:"phone_number"`
}

// 6. phonePasswordRequest defines phone password auth payload.
type phonePasswordRequest struct {
	PhoneCountryCode string `json:"phone_country_code" binding:"required"`
	PhoneNumber      string `json:"phone_number" binding:"required"`
	Password         string `json:"password" binding:"required"`
}

// 7. NewAuthHandler creates an auth handler instance.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 8. RequestOTP handles OTP request calls.
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

// 9. VerifyOTP handles OTP verify calls.
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

// 10. RequestEmailOTP handles email OTP request calls.
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

// 11. VerifyEmailOTP handles email OTP verify calls.
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

// 12. RegisterEmail handles email and phone password account creation.
func (h *AuthHandler) RegisterEmail(c *gin.Context) {
	var request emailPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.authService.RegisterWithEmail(c.Request.Context(), service.EmailPasswordParams{
		Email:            strings.TrimSpace(request.Email),
		Password:         request.Password,
		DisplayName:      strings.TrimSpace(request.DisplayName),
		PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:      strings.TrimSpace(request.PhoneNumber),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 13. LoginEmail handles email password sign-in.
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

// 14. LoginPhone handles phone password sign-in.
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

// 15. Logout handles logout calls.
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
