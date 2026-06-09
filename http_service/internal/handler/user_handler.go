/*
 * User and metadata HTTP handlers.
 * 1. Bind current user profile requests.
 * 2. Return communities and channel home data.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. UserHandler handles me, communities, and overview endpoints.
type UserHandler struct {
	userService *service.UserService
}

// 2. updateProfileRequest defines the profile update payload.
type updateProfileRequest struct {
	DisplayName           string   `json:"display_name"`
	Email                 string   `json:"email"`
	EmailOTPCode          string   `json:"email_otp_code"`
	PhoneCountryCode      string   `json:"phone_country_code"`
	PhoneNumber           string   `json:"phone_number"`
	Password              string   `json:"password"`
	PublisherIdentityType string   `json:"publisher_identity_type"`
	PrimaryCommunityID    string   `json:"primary_community_id"`
	PrimaryCommunityName  string   `json:"primary_community_name"`
	BoundBuildingIDs      []string `json:"bound_building_ids"`
	BoundFlatUnitIDs      []string `json:"bound_flat_unit_ids"`
	ResidenceFloor        string   `json:"residence_floor"`
	ResidenceUnit         string   `json:"residence_unit"`
	DistrictCode          string   `json:"district_code"`
	AvatarAssetID         string   `json:"avatar_asset_id"`
}

// 3. bindIsmartRequest defines current user ismart binding payload.
type bindIsmartRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// 4. NewUserHandler creates a user handler instance.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// 5. GetMe returns the current user payload.
func (h *UserHandler) GetMe(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.userService.GetMe(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. UpdateProfile updates the current user profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request updateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.userService.UpdateProfile(c.Request.Context(), user.UserID, service.UpdateProfileParams{
		DisplayName:           strings.TrimSpace(request.DisplayName),
		Email:                 strings.TrimSpace(request.Email),
		EmailOTPCode:          strings.TrimSpace(request.EmailOTPCode),
		PhoneCountryCode:      strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:           strings.TrimSpace(request.PhoneNumber),
		Password:              request.Password,
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		PrimaryCommunityID:    strings.TrimSpace(request.PrimaryCommunityID),
		PrimaryCommunityName:  strings.TrimSpace(request.PrimaryCommunityName),
		BoundBuildingIDs:      request.BoundBuildingIDs,
		BoundFlatUnitIDs:      request.BoundFlatUnitIDs,
		ResidenceFloor:        strings.TrimSpace(request.ResidenceFloor),
		ResidenceUnit:         strings.TrimSpace(request.ResidenceUnit),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		AvatarAssetID:         strings.TrimSpace(request.AvatarAssetID),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 7. BindIsmart links the current user to an ismart account.
func (h *UserHandler) BindIsmart(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request bindIsmartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.userService.BindIsmart(c.Request.Context(), user.UserID, service.IsmartLoginParams{
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

// 8. ListCommunities returns community options.
func (h *UserHandler) ListCommunities(c *gin.Context) {
	result, err := h.userService.ListCommunities(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 9. ChannelHomeOverview returns the channel home payload.
func (h *UserHandler) ChannelHomeOverview(c *gin.Context) {
	result, err := h.userService.GetChannelHomeOverview(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}
