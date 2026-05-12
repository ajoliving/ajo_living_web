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
	"ajoliving_web/http_service/internal/utils"
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
	AJOBalance        int64              `json:"ajo_balance"`
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
	PhoneCountryCode      string
	PhoneNumber           string
	PublisherIdentityType string
	PrimaryCommunityID    string
	DistrictCode          string
	AvatarAssetID         string
}

const profileAvatarUpdateCost = int64(50)

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
	if s.runtime.WalletService != nil {
		overview, err := s.runtime.WalletService.GetWalletOverview(ctx, userID)
		if err != nil {
			return nil, err
		}
		response.AJOBalance = overview.Account.Balance
	}

	return response, nil
}

// 8. UpdateProfile updates the authenticated profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, params UpdateProfileParams) (*MeResponse, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "user not found")
	}

	updates := map[string]any{
		"display_name":            params.DisplayName,
		"publisher_identity_type": params.PublisherIdentityType,
		"district_code":           params.DistrictCode,
	}
	userUpdates := map[string]any{}
	phoneCountryCode, phoneNumber, shouldUpdatePhone, err := profilePhoneUpdate(user, params)
	if err != nil {
		return nil, err
	}
	if shouldUpdatePhone {
		userUpdates["phone_country_code"] = phoneCountryCode
		userUpdates["phone_number"] = phoneNumber
		userUpdates["is_verified_phone"] = false
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

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var currentProfile model.UserProfile
		profileErr := tx.Where("user_id = ?", userID).First(&currentProfile).Error
		if profileErr != nil && !errors.Is(profileErr, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.CodeInternalError, "failed to load profile")
		}

		if len(userUpdates) > 0 {
			if err := s.ensurePhoneAvailable(ctx, tx, userID, phoneCountryCode, phoneNumber); err != nil {
				return err
			}
		}

		if params.AvatarAssetID != "" && shouldChargeAvatarUpdate(&currentProfile, updates["avatar_asset_id"]) {
			if s.runtime.WalletService == nil {
				return errcode.New(errcode.CodeInternalError, "wallet service is not configured")
			}
			if _, err := s.runtime.WalletService.SpendPointsWithTx(ctx, tx, WalletSpendParams{
				UserID:         userID,
				Amount:         profileAvatarUpdateCost,
				SourceType:     WalletSourceProfileCharge,
				BizModule:      "profile",
				ActionType:     WalletActionAvatar,
				IdempotencyKey: "profile:avatar:" + utils.NewPublicID(),
				Note:           "profile avatar update",
			}); err != nil {
				return err
			}
		}

		if len(userUpdates) > 0 {
			if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to update user phone")
			}
		}

		if err := tx.Where("user_id = ?", userID).Assign(updates).FirstOrCreate(&model.UserProfile{UserID: userID}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update profile")
		}

		return nil
	})
	if err != nil {
		return nil, err
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

// 12. ensurePhoneAvailable validates a changed phone number is not used by another account.
func (s *UserService) ensurePhoneAvailable(ctx context.Context, tx *gorm.DB, userID int64, countryCode string, phoneNumber string) error {
	var count int64
	if err := tx.WithContext(ctx).Model(&model.User{}).
		Where("id <> ? AND phone_country_code = ? AND phone_number = ?", userID, countryCode, phoneNumber).
		Count(&count).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate phone number")
	}
	if count > 0 {
		return errcode.New(errcode.CodeValidationError, "phone number is already registered")
	}

	return nil
}

// 13. loadOwnedAvatarAsset resolves an uploaded account avatar media asset.
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

// 14. userEmail returns the verified email credential for the current member.
func (s *UserService) userEmail(ctx context.Context, userID int64) string {
	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error; err != nil {
		return ""
	}

	return credential.Email
}

// 15. avatarURL builds the public URL for a profile avatar asset.
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

// 16. profilePhoneUpdate normalizes an optional profile phone update.
func profilePhoneUpdate(user model.User, params UpdateProfileParams) (string, string, bool, error) {
	if strings.TrimSpace(params.PhoneCountryCode) == "" && strings.TrimSpace(params.PhoneNumber) == "" {
		return "", "", false, nil
	}

	countryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	if !isValidPhone(countryCode, phoneNumber) {
		return "", "", false, errcode.New(errcode.CodeValidationError, "valid phone number is required")
	}

	if user.PhoneCountryCode == countryCode && user.PhoneNumber == phoneNumber {
		return countryCode, phoneNumber, false, nil
	}

	return countryCode, phoneNumber, true, nil
}

// 17. shouldChargeAvatarUpdate returns whether the avatar update needs a point charge.
func shouldChargeAvatarUpdate(profile *model.UserProfile, nextAssetID any) bool {
	assetID, ok := nextAssetID.(int64)
	if !ok || assetID <= 0 {
		return false
	}
	if profile == nil || profile.UserID == 0 || profile.AvatarAssetID == nil {
		return true
	}

	return *profile.AvatarAssetID != assetID
}

// 18. toCommunityResponse maps a community model to response data.
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
