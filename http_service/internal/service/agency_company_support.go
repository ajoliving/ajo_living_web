/*
 * Agency profile validation helpers.
 * 1. Enforce account type and profile-specific required fields.
 * 2. Validate member-owned media under dedicated OSS directories.
 * 3. Send committed review result emails without rolling back review.
 */
package service

import (
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"strings"
	"unicode/utf8"
)

// 1. applyAgencyProfileDraft maps member input and resolves semantic assets.
func (s *AgencyProfileService) applyAgencyProfileDraft(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, userID int64, params AgencyProfileUpsertParams) error {
	prefixes := agencyProfileAssetPrefixes(params.ProfileType)
	assetInputs := []struct {
		publicID string
		prefix   string
		target   **int64
	}{
		{params.AvatarAssetID, prefixes["avatar"], &profile.AvatarAssetID},
		{params.WechatQRAssetID, prefixes["wechat_qr"], &profile.WechatQRAssetID},
		{params.LogoAssetID, prefixes["logo"], &profile.LogoAssetID},
		{params.EAALicenseAssetID, prefixes["eaa"], &profile.EAALicenseAssetID},
		{params.BusinessRegistrationAssetID, prefixes["business_registration"], &profile.BusinessRegistrationAssetID},
		{params.CompanyCardAssetID, prefixes["company_card"], &profile.CompanyCardAssetID},
	}
	for _, input := range assetInputs {
		assetID, err := s.resolveAgencyProfileAssetID(ctx, tx, userID, input.publicID, input.prefix)
		if err != nil {
			return err
		}
		*input.target = assetID
	}
	profile.ProfileType = strings.TrimSpace(params.ProfileType)
	profile.NameZH, profile.NameEN = strings.TrimSpace(params.NameZH), strings.TrimSpace(params.NameEN)
	profile.AddressZH, profile.AddressEN = strings.TrimSpace(params.AddressZH), strings.TrimSpace(params.AddressEN)
	profile.LicenseNumber, profile.IsOverseas, profile.IsBigFour = strings.TrimSpace(params.LicenseNumber), params.IsOverseas, params.IsBigFour
	profile.Phone1CountryCode, profile.Phone1Number, profile.Phone1WhatsApp = strings.TrimSpace(params.Phone1CountryCode), strings.TrimSpace(params.Phone1Number), params.Phone1WhatsApp
	profile.Phone2CountryCode, profile.Phone2Number, profile.Phone2WhatsApp = strings.TrimSpace(params.Phone2CountryCode), strings.TrimSpace(params.Phone2Number), params.Phone2WhatsApp
	profile.WechatID, profile.WechatURL = strings.TrimSpace(params.WechatID), strings.TrimSpace(params.WechatURL)
	profile.SignatureZH, profile.SignatureEN = strings.TrimSpace(params.SignatureZH), strings.TrimSpace(params.SignatureEN)
	profile.DefaultAvatar = strings.TrimSpace(params.DefaultAvatar)
	if profile.ProfileType == AgencyProfileTypeCompany {
		profile.DefaultAvatar = ""
		profile.AvatarAssetID = nil
	}
	return nil
}

// 2. validateAgencyAccountType ensures profile type cannot be selected independently.
func (s *AgencyProfileService) validateAgencyAccountType(ctx context.Context, tx *gorm.DB, userID int64, profileType string) error {
	var profile model.UserProfile
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return errcode.New(errcode.CodeAuthForbidden, "agency account is required")
	}
	expected := map[string]string{AccountTypeIndividualAgent: AgencyProfileTypeIndividual, AccountTypeAgencyCompany: AgencyProfileTypeCompany}[profile.AccountType]
	if expected == "" || expected != strings.TrimSpace(profileType) {
		return errcode.New(errcode.CodeAuthForbidden, "agency profile type does not match account")
	}
	return nil
}

// 3. validateAgencyProfileSubmission validates required fields and evidence.
func (s *AgencyProfileService) validateAgencyProfileSubmission(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, userID int64) error {
	if !validRequiredText(profile.NameZH, 120) || !validRequiredText(profile.NameEN, 200) ||
		!validRequiredText(profile.Phone1CountryCode, 16) || !validRequiredText(profile.Phone1Number, 32) {
		return errcode.New(errcode.CodeValidationError, "complete agency profile information is required")
	}
	if utf8.RuneCountInString(profile.SignatureZH) > 300 || utf8.RuneCountInString(profile.SignatureEN) > 1000 {
		return errcode.New(errcode.CodeValidationError, "agency profile signature is too long")
	}
	if !validOptionalAgencyHTTPSURL(profile.WechatURL) {
		return errcode.New(errcode.CodeValidationError, "WeChat URL must be a valid HTTPS URL")
	}
	if profile.ProfileType == AgencyProfileTypeIndividual {
		return s.validateIndividualAgencySubmission(ctx, tx, profile, userID)
	}
	return s.validateCompanyAgencySubmission(ctx, tx, profile, userID)
}

// 4. validateIndividualAgencySubmission validates personal agent fields.
func (s *AgencyProfileService) validateIndividualAgencySubmission(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, userID int64) error {
	if profile.DefaultAvatar != "male" && profile.DefaultAvatar != "female" && profile.DefaultAvatar != "custom" {
		return errcode.New(errcode.CodeValidationError, "default avatar is required")
	}
	if profile.DefaultAvatar == "custom" && profile.AvatarAssetID == nil {
		return errcode.New(errcode.CodeValidationError, "custom avatar is required")
	}
	if !profile.IsOverseas && (!validRequiredText(profile.LicenseNumber, 120) || profile.EAALicenseAssetID == nil) {
		return errcode.New(errcode.CodeValidationError, "Hong Kong agent licence and EAA image are required")
	}
	return s.validateAgencyProfileAssets(ctx, tx, profile, userID)
}

// 5. validateCompanyAgencySubmission validates company identity and licence evidence.
func (s *AgencyProfileService) validateCompanyAgencySubmission(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, userID int64) error {
	if !validRequiredText(profile.LicenseNumber, 120) || profile.EAALicenseAssetID == nil {
		return errcode.New(errcode.CodeValidationError, "company licence number and licence image are required")
	}
	return s.validateAgencyProfileAssets(ctx, tx, profile, userID)
}

// 6. validateAgencyProfileAssets revalidates all persisted semantic references.
func (s *AgencyProfileService) validateAgencyProfileAssets(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, userID int64) error {
	prefixes := agencyProfileAssetPrefixes(profile.ProfileType)
	assets := []struct {
		id     *int64
		prefix string
	}{
		{profile.AvatarAssetID, prefixes["avatar"]}, {profile.WechatQRAssetID, prefixes["wechat_qr"]},
		{profile.LogoAssetID, prefixes["logo"]}, {profile.EAALicenseAssetID, prefixes["eaa"]},
		{profile.BusinessRegistrationAssetID, prefixes["business_registration"]}, {profile.CompanyCardAssetID, prefixes["company_card"]},
	}
	for _, asset := range assets {
		if asset.id != nil {
			if err := s.validateAgencyProfileAssetID(ctx, tx, userID, *asset.id, asset.prefix); err != nil {
				return err
			}
		}
	}
	return nil
}

// 7. resolveAgencyProfileAssetID validates ownership, image type, and OSS root.
func (s *AgencyProfileService) resolveAgencyProfileAssetID(ctx context.Context, tx *gorm.DB, userID int64, publicID string, prefix string) (*int64, error) {
	if strings.TrimSpace(publicID) == "" {
		return nil, nil
	}
	if prefix == "" {
		return nil, errcode.New(errcode.CodeValidationError, "media asset is not allowed for this profile type")
	}
	var asset model.MediaAsset
	if err := tx.WithContext(ctx).Where("public_id = ? AND created_by = ?", strings.TrimSpace(publicID), userID).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "agency profile media asset not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile media asset")
	}
	if !strings.HasPrefix(strings.ToLower(asset.MimeType), "image/") || !strings.HasPrefix(asset.ObjectKey, prefix) {
		return nil, errcode.New(errcode.CodeValidationError, "agency profile media asset is invalid")
	}
	return &asset.ID, nil
}

// 8. validateAgencyProfileAssetID validates a persisted asset reference.
func (s *AgencyProfileService) validateAgencyProfileAssetID(ctx context.Context, tx *gorm.DB, userID int64, assetID int64, prefix string) error {
	var asset model.MediaAsset
	if prefix == "" || tx.WithContext(ctx).Where("id = ? AND created_by = ?", assetID, userID).First(&asset).Error != nil ||
		!strings.HasPrefix(strings.ToLower(asset.MimeType), "image/") || !strings.HasPrefix(asset.ObjectKey, prefix) {
		return errcode.New(errcode.CodeValidationError, "agency profile media asset is invalid")
	}
	return nil
}

// 9. createAgencyProfileReviewNotification sends a review result to the member inbox.
func (s *AgencyProfileService) createAgencyProfileReviewNotification(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile, approved bool) error {
	body := "代理資料未獲批准，請修改後重新提交。"
	if approved {
		body = "代理資料已獲批准，現可使用代理身份發布樓盤。"
	} else if note := strings.TrimSpace(profile.ReviewNote); note != "" {
		body += " 原因：" + note
	}
	return NewNotificationService(s.runtime).CreateNotification(ctx, tx, CreateNotificationParams{
		UserID: profile.UserID, Category: "agency_profile_review", Title: "代理資料審核結果", Body: body,
		RelatedType: "agency_profile", RelatedPublicID: profile.PublicID,
	})
}

// 10. validRequiredText validates a required trimmed string.
func validRequiredText(value string, maxLength int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && utf8.RuneCountInString(trimmed) <= maxLength
}

// 11. validOptionalAgencyHTTPSURL allows only empty or absolute HTTPS links.
func validOptionalAgencyHTTPSURL(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && strings.EqualFold(parsed.Scheme, "https") && strings.TrimSpace(parsed.Host) != ""
}

// 1. sendAgencyProfileReviewEmail sends the committed review result to the registered email.
func (s *AgencyProfileService) sendAgencyProfileReviewEmail(ctx context.Context, profilePublicID string, approved bool, reviewNote string) {
	if s.runtime.MailSender == nil {
		return
	}

	var recipient struct {
		Email  *string
		NameZH string
		NameEN string
	}
	err := s.runtime.DB.WithContext(ctx).
		Table("agency_profiles").
		Select("user_credentials.email, agency_profiles.name_zh, agency_profiles.name_en").
		Joins("JOIN user_credentials ON user_credentials.user_id = agency_profiles.user_id").
		Where("agency_profiles.public_id = ?", strings.TrimSpace(profilePublicID)).
		Take(&recipient).Error
	if err != nil || recipient.Email == nil || strings.TrimSpace(*recipient.Email) == "" {
		return
	}

	subject, body := buildAgencyProfileReviewEmailContent(
		firstNonBlank(recipient.NameZH, recipient.NameEN),
		approved,
		strings.TrimSpace(reviewNote),
		strings.TrimSpace(s.runtime.Config.AppPublicBaseURL),
	)
	if err := s.runtime.MailSender.Send(ctx, strings.TrimSpace(*recipient.Email), subject, body); err != nil && s.runtime.Logger != nil {
		s.runtime.Logger.Warn("failed to send agency profile review email", "profile_public_id", profilePublicID, "error", err)
	}
}

// 2. buildAgencyProfileReviewEmailContent builds the Traditional Chinese review result email.
func buildAgencyProfileReviewEmailContent(name string, approved bool, reviewNote string, publicBaseURL string) (string, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "會員"
	}

	subject := "[AJO Living] 代理資料審核已通過"
	body := fmt.Sprintf("%s：\n\n您的代理資料已通過審核，現可登入 AJO Living 並使用已批准的代理資料刊登樓盤。", name)
	if !approved {
		subject = "[AJO Living] 代理資料審核未通過"
		body = fmt.Sprintf("%s：\n\n您的代理資料未通過審核，請登入 AJO Living 查看並修改資料後重新提交。\n\n拒絕原因：%s", name, firstNonBlank(reviewNote, "請登入會員中心查看審核意見。"))
	}

	if baseURL := strings.TrimRight(publicBaseURL, "/"); baseURL != "" {
		body += "\n\n查看代理資料：" + baseURL + "/account/profile/agency-profile"
	}
	body += "\n\nAJO Living"
	return subject, body
}
