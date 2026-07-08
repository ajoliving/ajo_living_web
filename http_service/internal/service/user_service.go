/*
 * User and member profile business logic.
 * 1. Load and update current member profile data.
 * 2. Provide community metadata and simple channel home content.
 */
package service

import (
	"context"
	"encoding/json"
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
	PublicID          string                        `json:"public_id"`
	Email             string                        `json:"email"`
	PhoneCountryCode  string                        `json:"phone_country_code"`
	PhoneNumber       string                        `json:"phone_number"`
	MemberStatus      string                        `json:"member_status"`
	MemberType        string                        `json:"member_type"`
	IsStaff           bool                          `json:"is_staff"`
	Role              string                        `json:"role"`
	Roles             []string                      `json:"roles"`
	Permissions       []string                      `json:"permissions"`
	DisplayName       string                        `json:"display_name"`
	AvatarURL         string                        `json:"avatar_url"`
	PublisherIdentity string                        `json:"publisher_identity_type"`
	DistrictCode      string                        `json:"district_code"`
	ResidenceFloor    string                        `json:"residence_floor"`
	ResidenceUnit     string                        `json:"residence_unit"`
	BoundBuildingIDs  []string                      `json:"bound_building_ids"`
	BoundFlatUnitIDs  []string                      `json:"bound_flat_unit_ids"`
	PrimaryCommunity  *CommunityResponse            `json:"primary_community,omitempty"`
	ProfileCompleted  bool                          `json:"profile_completed"`
	AJOBalance        int64                         `json:"ajo_balance"`
	IsmartLinked      bool                          `json:"ismart_linked"`
	IsmartUsername    string                        `json:"ismart_username"`
	IsmartBoundPhone  string                        `json:"ismart_bound_phone"`
	IsmartMsg         *IsmartMessage                `json:"ismart_msg,omitempty"`
	IsmartAccount     *IsmartAccountProfileResponse `json:"ismart_account_profile,omitempty"`
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

// 4. IsmartRelatedPropertyResponse defines a read-only legacy property row.
type IsmartRelatedPropertyResponse struct {
	PropertyName string `json:"property_name"`
	Status       string `json:"status"`
}

// 5. IsmartAccountProfileResponse defines read-only legacy iSmart account data.
type IsmartAccountProfileResponse struct {
	AccountCode    string                          `json:"account_code"`
	AccountPhone   string                          `json:"account_phone"`
	AccountEmail   string                          `json:"account_email"`
	OwnerNameEN    string                          `json:"owner_name_en"`
	OwnerNameZH    string                          `json:"owner_name_zh"`
	IdentityNumber string                          `json:"identity_number"`
	LegalEntity    string                          `json:"legal_entity"`
	Gender         string                          `json:"gender"`
	BirthDate      string                          `json:"birth_date"`
	ContactName    string                          `json:"contact_name"`
	ContactPhone   string                          `json:"contact_phone"`
	BillingEmail   string                          `json:"billing_email"`
	BillingAddress string                          `json:"billing_address"`
	Properties     []IsmartRelatedPropertyResponse `json:"properties"`
}

// 4. UpdateProfileParams defines profile update input.
type UpdateProfileParams struct {
	DisplayName           string
	Email                 string
	EmailOTPCode          string
	PhoneCountryCode      string
	PhoneNumber           string
	Password              string
	PublisherIdentityType string
	PrimaryCommunityID    string
	PrimaryCommunityName  string
	BoundBuildingIDs      []string
	BoundFlatUnitIDs      []string
	ResidenceFloor        string
	ResidenceUnit         string
	DistrictCode          string
	AvatarAssetID         string
}

const profileAvatarUpdateCost = int64(50)
const profileEmailUpdateScene = "profile_email_update"

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

	access := buildAccessSnapshot(user.IsStaff)

	var profile model.UserProfile
	err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", userID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load profile")
	}

	ismartMsg := NewAuthService(s.runtime).loadIsmartMessage(ctx, user.ID)
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
		ResidenceFloor:    profile.ResidenceFloor,
		ResidenceUnit:     profile.ResidenceUnit,
		BoundBuildingIDs:  s.profileBoundBuildings(&profile, ismartMsg),
		BoundFlatUnitIDs:  s.profileBoundFlatUnits(&profile, ismartMsg),
		ProfileCompleted:  isProfileCompleted(&profile),
		IsmartLinked:      ismartMsg != nil,
		IsmartMsg:         ismartMsg,
		IsmartAccount:     s.loadIsmartAccountProfile(ctx, user.ID),
	}
	if ismartMsg != nil {
		response.IsmartUsername = ismartMsg.Username
		response.IsmartBoundPhone = ismartMsg.Phone
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
		"residence_floor":         params.ResidenceFloor,
		"residence_unit":          params.ResidenceUnit,
		"district_code":           params.DistrictCode,
	}
	userUpdates := map[string]any{}
	credentialUpdates := map[string]any{}
	phoneCountryCode, phoneNumber, shouldUpdatePhone, err := profilePhoneUpdate(user, params)
	if err != nil {
		return nil, err
	}
	if shouldUpdatePhone {
		userUpdates["phone_country_code"] = phoneCountryCode
		userUpdates["phone_number"] = phoneNumber
		userUpdates["is_verified_phone"] = false
	}
	if strings.TrimSpace(params.Password) != "" {
		if len(strings.TrimSpace(params.Password)) < 8 {
			return nil, errcode.New(errcode.CodeValidationError, "password must be at least 8 characters")
		}
		passwordHash, err := utils.HashPassword(strings.TrimSpace(params.Password))
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
		}
		passwordEncrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, strings.TrimSpace(params.Password))
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
		}
		credentialUpdates["password_hash"] = passwordHash
		credentialUpdates["password_encrypted"] = passwordEncrypted
	}

	email, shouldUpdateEmail, err := s.profileEmailUpdate(ctx, user.ID, params)
	if err != nil {
		return nil, err
	}

	if params.PrimaryCommunityID != "" {
		community, err := s.resolveProfileCommunity(ctx, params.PrimaryCommunityID, params.PrimaryCommunityName)
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
	if params.BoundBuildingIDs != nil {
		boundBuildingIDs, err := marshalJSON(normalizeStringSlice(params.BoundBuildingIDs))
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to prepare building bindings")
		}
		updates["bound_building_ids"] = boundBuildingIDs
	}
	if params.BoundFlatUnitIDs != nil {
		boundFlatUnitIDs, err := marshalJSON(normalizeStringSlice(params.BoundFlatUnitIDs))
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to prepare unit bindings")
		}
		updates["bound_flat_unit_ids"] = boundFlatUnitIDs
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

		if shouldUpdateEmail {
			if err := s.saveProfileEmailCredential(ctx, tx, userID, email); err != nil {
				return err
			}
		}
		if len(credentialUpdates) > 0 {
			if err := s.updateProfilePassword(ctx, tx, userID, credentialUpdates); err != nil {
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

	if shouldUpdateEmail {
		s.runtime.OTPStore.Delete(emailOTPKey(email, profileEmailUpdateScene))
	}

	return s.GetMe(ctx, userID)
}

// 9. BindIsmart links the current member profile to an ismart account.
func (s *UserService) BindIsmart(ctx context.Context, userID int64, params IsmartLoginParams) (*MeResponse, error) {
	if _, err := NewAuthService(s.runtime).BindIsmartAccount(ctx, userID, params); err != nil {
		return nil, err
	}

	return s.GetMe(ctx, userID)
}

// 10. ListCommunities returns the seeded communities.
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

// 11. GetChannelHomeOverview returns channel cards and featured secondhand content.
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

// 12. resolveProfileCommunity loads or mirrors a POS building as local community.
func (s *UserService) resolveProfileCommunity(ctx context.Context, publicID string, name string) (*model.Community, error) {
	community, err := s.findCommunityByPublicID(ctx, publicID)
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
	if err := s.runtime.DB.WithContext(ctx).Create(community).Error; err != nil {
		found, findErr := s.findCommunityByPublicID(ctx, publicID)
		if findErr == nil {
			return found, nil
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to create profile community")
	}

	return community, nil
}

// 13. findCommunityByPublicID resolves a community by public ID.
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

// 14. ensurePhoneAvailable validates a changed phone number is not used by another account.
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

// 15. loadOwnedAvatarAsset resolves an uploaded account avatar media asset.
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

// 16. userEmail returns the verified email credential for the current member.
func (s *UserService) userEmail(ctx context.Context, userID int64) string {
	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error; err != nil {
		return ""
	}
	if credential.Email == nil {
		return ""
	}

	return *credential.Email
}

// 17. profileEmailUpdate validates email OTP when the email is changed.
func (s *UserService) profileEmailUpdate(ctx context.Context, userID int64, params UpdateProfileParams) (string, bool, error) {
	email := normalizeEmail(params.Email)
	if email == "" {
		return "", false, nil
	}
	if !isValidEmail(email) {
		return "", false, errcode.New(errcode.CodeValidationError, "valid email is required")
	}

	currentEmail := normalizeEmail(s.userEmail(ctx, userID))
	if currentEmail == email {
		return email, false, nil
	}
	if strings.TrimSpace(params.EmailOTPCode) == "" {
		return "", false, errcode.New(errcode.CodeValidationError, "email verification code is required")
	}

	key := emailOTPKey(email, profileEmailUpdateScene)
	record, ok := s.runtime.OTPStore.Get(key)
	if !ok || s.runtime.Now().After(record.ExpiresAt) {
		return "", false, errcode.New(errcode.CodeValidationError, "email otp is invalid or expired")
	}
	if strings.TrimSpace(params.EmailOTPCode) != record.Code {
		return "", false, errcode.New(errcode.CodeValidationError, "email otp is invalid or expired")
	}

	return email, true, nil
}

// 19. saveProfileEmailCredential updates or creates the verified email credential.
func (s *UserService) saveProfileEmailCredential(ctx context.Context, tx *gorm.DB, userID int64, email string) error {
	var credential model.UserCredential
	findErr := tx.WithContext(ctx).Where("email = ?", email).First(&credential).Error
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load email credential")
	}
	if credential.UserID > 0 && credential.UserID != userID {
		return errcode.New(errcode.CodeValidationError, "email is already registered")
	}

	var current model.UserCredential
	currentErr := tx.WithContext(ctx).Where("user_id = ?", userID).First(&current).Error
	if currentErr != nil && !errors.Is(currentErr, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load current email credential")
	}
	if current.UserID > 0 {
		return tx.WithContext(ctx).Model(&current).Updates(map[string]any{
			"email":       email,
			"is_verified": true,
		}).Error
	}

	return tx.WithContext(ctx).Create(&model.UserCredential{
		UserID:     userID,
		Email:      &email,
		IsVerified: true,
	}).Error
}

// 20. updateProfilePassword updates or creates the local password credential.
func (s *UserService) updateProfilePassword(ctx context.Context, tx *gorm.DB, userID int64, updates map[string]any) error {
	var credential model.UserCredential
	currentErr := tx.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error
	if currentErr != nil && !errors.Is(currentErr, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load current credential")
	}
	if credential.UserID > 0 {
		return tx.WithContext(ctx).Model(&credential).Updates(updates).Error
	}

	return tx.WithContext(ctx).Create(&model.UserCredential{
		UserID:            userID,
		PasswordHash:      updates["password_hash"].(string),
		PasswordEncrypted: updates["password_encrypted"].(string),
		IsVerified:        false,
	}).Error
}

// 21. avatarURL builds the public URL for a profile avatar asset.
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

// 22. profilePhoneUpdate normalizes an optional profile phone update.
func profilePhoneUpdate(user model.User, params UpdateProfileParams) (string, string, bool, error) {
	if strings.TrimSpace(params.PhoneCountryCode) == "" && strings.TrimSpace(params.PhoneNumber) == "" {
		return "", "", false, nil
	}
	if strings.EqualFold(strings.TrimSpace(params.PhoneCountryCode), "email") || strings.EqualFold(strings.TrimSpace(params.PhoneCountryCode), "ismart") {
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

// 23. profileBoundBuildings returns member center building bindings or POS fallback.
func (s *UserService) profileBoundBuildings(profile *model.UserProfile, ismartMsg *IsmartMessage) []string {
	if profile != nil {
		if values := normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs)); len(values) > 0 {
			return values
		}
		if profile.PrimaryCommunity != nil && strings.TrimSpace(profile.PrimaryCommunity.PublicID) != "" {
			return []string{strings.TrimSpace(profile.PrimaryCommunity.PublicID)}
		}
		if hasLocalProfileBinding(profile) {
			return []string{}
		}
	}

	return resolveIsmartBoundBuildings(ismartMsg)
}

// 24. profileBoundFlatUnits returns member center unit bindings or POS fallback.
func (s *UserService) profileBoundFlatUnits(profile *model.UserProfile, ismartMsg *IsmartMessage) []string {
	if profile != nil {
		if values := normalizeStringSlice(unmarshalStringSlice(profile.BoundFlatUnitIDs)); len(values) > 0 {
			return values
		}
		if hasLocalProfileBinding(profile) {
			return []string{}
		}
	}

	return resolveIsmartBoundUnits(ismartMsg)
}

// 25. shouldChargeAvatarUpdate returns whether the avatar update needs a point charge.
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

// 26. loadIsmartAccountProfile returns the read-only legacy account snapshot.
func (s *UserService) loadIsmartAccountProfile(ctx context.Context, userID int64) *IsmartAccountProfileResponse {
	var account model.UserIsmartAccount
	result := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).Limit(1).Find(&account)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil
	}

	raw := map[string]any{}
	_ = json.Unmarshal(account.RawMessage, &raw)
	properties := make([]IsmartRelatedPropertyResponse, 0)
	for _, propertyName := range normalizeStringSlice(append(
		unmarshalStringSlice(account.ClientBuildingPermissions),
		unmarshalStringSlice(account.StaffBuildingPermissions)...,
	)) {
		properties = append(properties, IsmartRelatedPropertyResponse{
			PropertyName: propertyName,
			Status:       "",
		})
	}

	return &IsmartAccountProfileResponse{
		AccountCode:    firstLegacyText(raw, account.Username, "account_code", "account_no", "account_number", "username", "memberno"),
		AccountPhone:   firstLegacyText(raw, account.Phone, "account_phone", "memberphone", "phone", "tel"),
		AccountEmail:   firstLegacyText(raw, account.Email, "account_email", "memberemail", "email", "billing_email"),
		OwnerNameEN:    firstLegacyText(raw, "", "owner_name_en", "memberengname", "eng_name", "english_name"),
		OwnerNameZH:    firstLegacyText(raw, "", "owner_name_zh", "memberchiname", "chi_name", "chinese_name"),
		IdentityNumber: firstLegacyText(raw, "", "identity_number", "memberid", "id_number", "hkid"),
		LegalEntity:    firstLegacyText(raw, "", "legal_entity", "member_legalentity", "legalentity"),
		Gender:         firstLegacyText(raw, "", "gender", "membergender"),
		BirthDate:      firstLegacyText(raw, "", "birth_date", "birthday", "date_of_birth", "dob"),
		ContactName:    firstLegacyText(raw, "", "contact_name", "contact_person", "contactperson"),
		ContactPhone:   firstLegacyText(raw, account.Phone, "contact_phone", "contact_tel", "contactphone"),
		BillingEmail:   firstLegacyText(raw, account.Email, "billing_email", "bill_email", "memberemail"),
		BillingAddress: firstLegacyText(raw, "", "billing_address", "bill_address", "address"),
		Properties:     properties,
	}
}

// 27. firstLegacyText reads a stable text value from possible legacy keys.
func firstLegacyText(raw map[string]any, fallback string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(paymentStringValue(raw[key])); value != "" {
			return value
		}
	}

	return strings.TrimSpace(fallback)
}

// 28. toCommunityResponse maps a community model to response data.
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
