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
	DisplayName           string `json:"display_name"`
	PhoneCountryCode      string `json:"phone_country_code"`
	PhoneNumber           string `json:"phone_number"`
	PublisherIdentityType string `json:"publisher_identity_type"`
	PrimaryCommunityID    string `json:"primary_community_id"`
	DistrictCode          string `json:"district_code"`
	AvatarAssetID         string `json:"avatar_asset_id"`
}

// 3. NewUserHandler creates a user handler instance.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// 4. GetMe returns the current user payload.
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

// 5. UpdateProfile updates the current user profile.
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
		PhoneCountryCode:      strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:           strings.TrimSpace(request.PhoneNumber),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		PrimaryCommunityID:    strings.TrimSpace(request.PrimaryCommunityID),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		AvatarAssetID:         strings.TrimSpace(request.AvatarAssetID),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. ListCommunities returns community options.
func (h *UserHandler) ListCommunities(c *gin.Context) {
	result, err := h.userService.ListCommunities(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 7. ChannelHomeOverview returns the channel home payload.
func (h *UserHandler) ChannelHomeOverview(c *gin.Context) {
	result, err := h.userService.GetChannelHomeOverview(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}
