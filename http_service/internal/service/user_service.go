/*
 * User and discovery business logic.
 * 1. Load and update current member profile data.
 * 2. Provide community metadata and simple channel home content.
 */
package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. UserService handles member profile and community reads.
type UserService struct {
	runtime *Runtime
}

// 2. MeResponse defines the current user response shape.
type MeResponse struct {
	PublicID          string             `json:"public_id"`
	Email             string             `json:"email"`
	PhoneCountryCode  string             `json:"phone_country_code"`
	PhoneNumber       string             `json:"phone_number"`
	MemberStatus      string             `json:"member_status"`
	MemberType        string             `json:"member_type"`
	IsStaff           bool               `json:"is_staff"`
	Role              string             `json:"role"`
	Roles             []string           `json:"roles"`
	Permissions       []string           `json:"permissions"`
	DisplayName       string             `json:"display_name"`
	AvatarURL         string             `json:"avatar_url"`
	PublisherIdentity string             `json:"publisher_identity_type"`
	DistrictCode      string             `json:"district_code"`
	PrimaryCommunity  *CommunityResponse `json:"primary_community,omitempty"`
	ProfileCompleted  bool               `json:"profile_completed"`
}

// 3. CommunityResponse defines a lightweight community payload.
type CommunityResponse struct {
	PublicID      string `json:"public_id"`
	CommunityType string `json:"community_type"`
	NameZH        string `json:"name_zh"`
	NameEN        string `json:"name_en"`
	DistrictCode  string `json:"district_code"`
	AddressText   string `json:"address_text"`
}

// 4. UpdateProfileParams defines profile update input.
type UpdateProfileParams struct {
	DisplayName           string
	PublisherIdentityType string
	PrimaryCommunityID    string
	DistrictCode          string
	AvatarAssetID         string
}

// 5. ChannelHomeOverview defines the channel home payload.
type ChannelHomeOverview struct {
	Channels []map[string]string        `json:"channels"`
	Featured []SecondhandListingSummary `json:"featured_secondhand"`
}

// 6. NewUserService creates a user service instance.
func NewUserService(runtime *Runtime) *UserService {
	return &UserService{runtime: runtime}
}

// 7. GetMe loads the authenticated user and profile data.
func (s *UserService) GetMe(ctx context.Context, userID int64) (*MeResponse, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "user not found")
	}

	access, err := NewAccessService(s.runtime).ResolveUserAccess(ctx, &user)
	if err != nil {
		return nil, err
	}

	var profile model.UserProfile
	err = s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", userID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load profile")
	}

	response := &MeResponse{
		PublicID:          user.PublicID,
		Email:             s.userEmail(ctx, user.ID),
		PhoneCountryCode:  user.PhoneCountryCode,
		PhoneNumber:       user.PhoneNumber,
		MemberStatus:      user.MemberStatus,
		MemberType:        normalizeMemberType(user.MemberType),
		IsStaff:           access.IsStaff,
		Role:              resolveUserRole(user.MemberType, access.IsStaff),
		Roles:             access.RoleCodes,
		Permissions:       access.Permissions,
		DisplayName:       profile.DisplayName,
		AvatarURL:         s.avatarURL(ctx, profile.AvatarAssetID),
		PublisherIdentity: profile.PublisherIdentityType,
		DistrictCode:      profile.DistrictCode,
		ProfileCompleted:  profile.PrimaryCommunityID != nil,
	}

	if profile.PrimaryCommunity != nil {
		response.PrimaryCommunity = toCommunityResponse(profile.PrimaryCommunity)
	}

	return response, nil
}

// 8. UpdateProfile updates the authenticated profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, params UpdateProfileParams) (*MeResponse, error) {
	updates := map[string]any{
		"display_name":            params.DisplayName,
		"publisher_identity_type": params.PublisherIdentityType,
		"district_code":           params.DistrictCode,
	}

	if params.PrimaryCommunityID != "" {
		community, err := s.findCommunityByPublicID(ctx, params.PrimaryCommunityID)
		if err != nil {
			return nil, err
		}

		updates["primary_community_id"] = community.ID
		if params.DistrictCode == "" {
			updates["district_code"] = community.DistrictCode
		}
	}

	if params.AvatarAssetID != "" {
		asset, err := s.loadOwnedAvatarAsset(ctx, userID, params.AvatarAssetID)
		if err != nil {
			return nil, err
		}

		updates["avatar_asset_id"] = asset.ID
	}

	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).Assign(updates).FirstOrCreate(&model.UserProfile{UserID: userID}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update profile")
	}

	return s.GetMe(ctx, userID)
}

// 9. ListCommunities returns the seeded communities.
func (s *UserService) ListCommunities(ctx context.Context) ([]CommunityResponse, error) {
	var communities []model.Community
	if err := s.runtime.DB.WithContext(ctx).Order("district_code asc, name_zh asc").Find(&communities).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load communities")
	}

	result := make([]CommunityResponse, 0, len(communities))
	for _, item := range communities {
		community := item
		result = append(result, *toCommunityResponse(&community))
	}

	return result, nil
}

// 10. GetChannelHomeOverview returns channel cards and featured secondhand content.
func (s *UserService) GetChannelHomeOverview(ctx context.Context) (*ChannelHomeOverview, error) {
	secondhandService := NewSecondhandService(s.runtime)
	featured, _, err := secondhandService.ListPublicSecondhand(ctx, SecondhandListFilters{
		Page:     1,
		PageSize: 6,
	})
	if err != nil {
		return nil, err
	}

	return &ChannelHomeOverview{
		Channels: []map[string]string{
			{"code": "property_sale", "title": "Property Sale", "description": "Future module placeholder"},
			{"code": "serviced_apartment", "title": "Serviced Apartment", "description": "Future module placeholder"},
			{"code": "secondhand", "title": "Secondhand", "description": "Neighbour-first secondhand marketplace"},
		},
		Featured: featured,
	}, nil
}

// 11. findCommunityByPublicID resolves a community by public ID.
func (s *UserService) findCommunityByPublicID(ctx context.Context, publicID string) (*model.Community, error) {
	var community model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", publicID).First(&community).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "community not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load community")
	}

	return &community, nil
}

// 12. loadOwnedAvatarAsset resolves an uploaded account avatar media asset.
func (s *UserService) loadOwnedAvatarAsset(ctx context.Context, userID int64, mediaAssetID string) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).
		Where("public_id = ? AND created_by = ?", mediaAssetID, userID).
		First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "avatar media asset not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load avatar media asset")
	}

	if !strings.HasPrefix(strings.ToLower(asset.MimeType), "image/") {
		return nil, errcode.New(errcode.CodeValidationError, "avatar media asset must be an image")
	}
	if !strings.HasPrefix(strings.TrimSpace(asset.ObjectKey), accountMediaObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "avatar media asset must be uploaded under account directory")
	}

	return &asset, nil
}

// 13. userEmail returns the verified email credential for the current member.
func (s *UserService) userEmail(ctx context.Context, userID int64) string {
	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error; err != nil {
		return ""
	}

	return credential.Email
}

// 14. avatarURL builds the public URL for a profile avatar asset.
func (s *UserService) avatarURL(ctx context.Context, avatarAssetID *int64) string {
	if avatarAssetID == nil {
		return ""
	}

	var asset model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).First(&asset, *avatarAssetID).Error; err != nil {
		return ""
	}

	return buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey)
}

// 15. toCommunityResponse maps a community model to response data.
func toCommunityResponse(community *model.Community) *CommunityResponse {
	if community == nil {
		return nil
	}

	return &CommunityResponse{
		PublicID:      community.PublicID,
		CommunityType: community.CommunityType,
		NameZH:        community.NameZH,
		NameEN:        community.NameEN,
		DistrictCode:  community.DistrictCode,
		AddressText:   community.AddressText,
	}
}
