/*
 * Agency profile query workflows.
 * 1. Return active and working versions to members.
 * 2. Return searchable review records and evidence to staff.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. GetMemberAgencyProfile returns the approved profile and working revision.
func (s *AgencyProfileService) GetMemberAgencyProfile(ctx context.Context, userID int64) (*AgencyProfileMemberResponse, error) {
	result := &AgencyProfileMemberResponse{}
	var user model.User
	var userProfile model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "user not found")
	}
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "user profile not found")
	}
	result.AccountType, result.MemberStatus = normalizeAccountType(userProfile.AccountType), user.MemberStatus
	bindingUserID := userID
	if result.AccountType == AccountTypeAgencyCompanySubaccount {
		var link model.AgencyCompanySubaccount
		if err := s.runtime.DB.WithContext(ctx).Where("child_user_id = ? AND status = ?", userID, "active").First(&link).Error; err != nil {
			return result, nil
		}
		bindingUserID = link.CompanyOwnerUserID
	}
	var binding model.AgencyProfileBinding
	err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", bindingUserID).First(&binding).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile binding")
	}
	result.NextEditableAt = s.nextAgencyProfileEditableAt(&binding)
	if binding.ActiveProfileID != nil {
		profile, err := s.loadAgencyProfileByID(ctx, *binding.ActiveProfileID)
		if err != nil {
			return nil, err
		}
		result.ActiveProfile, err = s.buildAgencyProfileResponse(ctx, profile, result.NextEditableAt)
		if err != nil {
			return nil, err
		}
	}
	if binding.RevisionProfileID != nil {
		if result.AccountType == AccountTypeAgencyCompanySubaccount {
			return result, nil
		}
		profile, err := s.loadAgencyProfileByID(ctx, *binding.RevisionProfileID)
		if err != nil {
			return nil, err
		}
		result.Revision, err = s.buildAgencyProfileResponse(ctx, profile, result.NextEditableAt)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// 2. ListAgencyProfilesForStaff returns paginated review records.
func (s *AgencyProfileService) ListAgencyProfilesForStaff(ctx context.Context, filters AgencyProfileListFilters) ([]AgencyProfileStaffResponse, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.AgencyProfile{}).
		Joins("JOIN agency_profile_bindings ON agency_profile_bindings.user_id = agency_profiles.user_id AND (agency_profile_bindings.active_profile_id = agency_profiles.id OR agency_profile_bindings.revision_profile_id = agency_profiles.id)").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = agency_profiles.user_id").
		Joins("LEFT JOIN user_credentials ON user_credentials.user_id = agency_profiles.user_id")
	if status := strings.TrimSpace(filters.Status); status != "" && status != "all" {
		query = query.Where("agency_profiles.status = ?", status)
	}
	if profileType := strings.TrimSpace(filters.ProfileType); profileType != "" {
		query = query.Where("agency_profiles.profile_type = ?", profileType)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("LOWER(agency_profiles.name_zh) LIKE LOWER(?) OR LOWER(agency_profiles.name_en) LIKE LOWER(?) OR LOWER(agency_profiles.license_number) LIKE LOWER(?) OR LOWER(user_profiles.display_name) LIKE LOWER(?) OR LOWER(user_credentials.email) LIKE LOWER(?)", like, like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count agency profiles")
	}
	var profiles []model.AgencyProfile
	if err := query.Order("agency_profiles.updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&profiles).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load agency profiles")
	}
	items := make([]AgencyProfileStaffResponse, 0, len(profiles))
	for index := range profiles {
		item, err := s.buildAgencyProfileStaffResponse(ctx, &profiles[index])
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 3. GetAgencyProfileForStaff returns one review detail.
func (s *AgencyProfileService) GetAgencyProfileForStaff(ctx context.Context, publicID string) (*AgencyProfileStaffResponse, error) {
	var profile model.AgencyProfile
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(publicID)).First(&profile).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "agency profile not found")
	}
	return s.buildAgencyProfileStaffResponse(ctx, &profile)
}

// 4. loadAgencyProfileByID loads one profile version.
func (s *AgencyProfileService) loadAgencyProfileByID(ctx context.Context, profileID int64) (*model.AgencyProfile, error) {
	var profile model.AgencyProfile
	if err := s.runtime.DB.WithContext(ctx).First(&profile, profileID).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "agency profile not found")
	}
	return &profile, nil
}

// 5. buildAgencyProfileStaffResponse joins a profile with account identity.
func (s *AgencyProfileService) buildAgencyProfileStaffResponse(ctx context.Context, profile *model.AgencyProfile) (*AgencyProfileStaffResponse, error) {
	response, err := s.buildAgencyProfileResponse(ctx, profile, nil)
	if err != nil {
		return nil, err
	}
	owner, err := s.loadAgencyProfileOwner(ctx, profile.UserID)
	if err != nil {
		return nil, err
	}
	return &AgencyProfileStaffResponse{AgencyProfileResponse: *response, Owner: *owner}, nil
}

// 6. buildAgencyProfileResponse maps one profile and semantic assets.
func (s *AgencyProfileService) buildAgencyProfileResponse(ctx context.Context, profile *model.AgencyProfile, nextEditableAt *string) (*AgencyProfileResponse, error) {
	response := &AgencyProfileResponse{
		ProfileID: profile.PublicID, ProfileType: profile.ProfileType, NameZH: profile.NameZH, NameEN: profile.NameEN,
		AddressZH: profile.AddressZH, AddressEN: profile.AddressEN, LicenseNumber: profile.LicenseNumber,
		IsOverseas: profile.IsOverseas, IsBigFour: profile.IsBigFour, Phone1CountryCode: profile.Phone1CountryCode,
		Phone1Number: profile.Phone1Number, Phone1WhatsApp: profile.Phone1WhatsApp, Phone2CountryCode: profile.Phone2CountryCode,
		Phone2Number: profile.Phone2Number, Phone2WhatsApp: profile.Phone2WhatsApp, WechatID: profile.WechatID,
		WechatURL: profile.WechatURL, SignatureZH: profile.SignatureZH, SignatureEN: profile.SignatureEN,
		DefaultAvatar: profile.DefaultAvatar, Status: profile.Status, ReviewNote: profile.ReviewNote,
		SubmittedAt: formatAgencyProfileTime(profile.SubmittedAt), ReviewedAt: formatAgencyProfileTime(profile.ReviewedAt),
		UpdatedAt: profile.UpdatedAt.UTC().Format(time.RFC3339), NextEditableAt: nextEditableAt,
	}
	assets, err := s.loadAgencyProfileAssets(ctx, profile)
	if err != nil {
		return nil, err
	}
	response.AvatarAsset = agencyAssetResponse(s, assets, profile.AvatarAssetID)
	response.WechatQRAsset = agencyAssetResponse(s, assets, profile.WechatQRAssetID)
	response.LogoAsset = agencyAssetResponse(s, assets, profile.LogoAssetID)
	response.EAALicenseAsset, err = agencyPrivateAssetResponse(ctx, s, assets, profile.EAALicenseAssetID)
	if err != nil {
		return nil, err
	}
	response.BusinessRegistrationAsset, err = agencyPrivateAssetResponse(ctx, s, assets, profile.BusinessRegistrationAssetID)
	if err != nil {
		return nil, err
	}
	response.CompanyCardAsset = agencyAssetResponse(s, assets, profile.CompanyCardAssetID)
	return response, nil
}

// 7. loadAgencyProfileAssets loads semantic media references.
func (s *AgencyProfileService) loadAgencyProfileAssets(ctx context.Context, profile *model.AgencyProfile) (map[int64]model.MediaAsset, error) {
	ids := make([]int64, 0, 6)
	for _, id := range []*int64{profile.AvatarAssetID, profile.WechatQRAssetID, profile.LogoAssetID, profile.EAALicenseAssetID, profile.BusinessRegistrationAssetID, profile.CompanyCardAssetID} {
		if id != nil {
			ids = append(ids, *id)
		}
	}
	result := make(map[int64]model.MediaAsset, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var assets []model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", ids).Find(&assets).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile assets")
	}
	for _, asset := range assets {
		result[asset.ID] = asset
	}
	return result, nil
}

// 8. loadAgencyProfileOwner loads staff-visible member identity.
func (s *AgencyProfileService) loadAgencyProfileOwner(ctx context.Context, userID int64) (*AgencyProfileOwnerResponse, error) {
	type ownerRow struct {
		DisplayName, PublicID string
		Email                 *string
	}
	var row ownerRow
	if err := s.runtime.DB.WithContext(ctx).Table("users").Select("user_profiles.display_name, user_credentials.email, users.public_id").Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.id").Joins("LEFT JOIN user_credentials ON user_credentials.user_id = users.id").Where("users.id = ?", userID).Scan(&row).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile owner")
	}
	email := ""
	if row.Email != nil {
		email = *row.Email
	}
	return &AgencyProfileOwnerResponse{DisplayName: row.DisplayName, Email: email, PublicID: row.PublicID}, nil
}

// 9. nextAgencyProfileEditableAt returns the edit cooldown deadline.
func (s *AgencyProfileService) nextAgencyProfileEditableAt(binding *model.AgencyProfileBinding) *string {
	if binding.LastEditedAt == nil {
		return nil
	}
	next := binding.LastEditedAt.Add(agencyProfileEditCooldown)
	if !s.agencyProfileNow().Before(next) {
		return nil
	}
	formatted := next.UTC().Format(time.RFC3339)
	return &formatted
}

// 10. agencyAssetResponse maps a referenced media asset.
func agencyAssetResponse(service *AgencyProfileService, assets map[int64]model.MediaAsset, id *int64) *AgencyProfileAssetResponse {
	if id == nil {
		return nil
	}
	asset, ok := assets[*id]
	if !ok {
		return nil
	}
	return &AgencyProfileAssetResponse{MediaAssetID: asset.PublicID, URL: buildMediaURL(service.runtime.Config.MediaBaseURL, asset.ObjectKey), MimeType: asset.MimeType}
}

// 11. agencyPrivateAssetResponse maps evidence to a ten-minute signed download URL.
func agencyPrivateAssetResponse(ctx context.Context, service *AgencyProfileService, assets map[int64]model.MediaAsset, id *int64) (*AgencyProfileAssetResponse, error) {
	if id == nil {
		return nil, nil
	}
	asset, ok := assets[*id]
	if !ok {
		return nil, nil
	}
	if service.runtime.StorageProvider == nil {
		return nil, errcode.New(errcode.CodeInternalError, "private agency document storage is not configured")
	}
	signedURL, err := service.runtime.StorageProvider.PresignDownload(ctx, asset.ObjectKey, 10*time.Minute)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to authorize private agency document")
	}
	return &AgencyProfileAssetResponse{MediaAssetID: asset.PublicID, URL: signedURL, MimeType: asset.MimeType}, nil
}

// 12. formatAgencyProfileTime formats nullable timestamps.
func formatAgencyProfileTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

// 13. normalizeAccountType keeps legacy empty profiles personal.
func normalizeAccountType(value string) string {
	switch strings.TrimSpace(value) {
	case AccountTypeIndividualAgent, AccountTypeAgencyCompany, AccountTypeAgencyCompanySubaccount:
		return strings.TrimSpace(value)
	default:
		return AccountTypePersonal
	}
}
