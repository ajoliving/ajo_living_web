/*
 * Agency company subaccount workflows.
 * 1. Create child logins under one approved company owner.
 * 2. List, disable, and remove child access.
 * 3. Revalidate parent approval for every child authentication.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

var agencySubaccountPermissionSet = map[string]struct{}{"property_publish": {}, "property_manage": {}}

// 1. CreateSubaccount creates an active child under an approved company.
func (s *AgencyProfileService) CreateSubaccount(ctx context.Context, ownerUserID int64, params AgencySubaccountCreateParams) (*AgencySubaccountResponse, error) {
	permissions, err := normalizeAgencySubaccountPermissions(params.Permissions)
	if err != nil {
		return nil, err
	}
	phoneCode, phone := normalizePhoneCountryCode(params.PhoneCountryCode), normalizePhoneNumber(params.PhoneNumber)
	email, displayName, password := normalizeEmail(params.Email), strings.TrimSpace(params.DisplayName), strings.TrimSpace(params.Password)
	if displayName == "" || !isValidPhone(phoneCode, phone) || len(password) < 8 || (email != "" && !isValidEmail(email)) {
		return nil, errcode.New(errcode.CodeValidationError, "valid name, phone, email, and password are required")
	}
	if err := s.requireApprovedCompanyOwner(ctx, s.runtime.DB, ownerUserID); err != nil {
		return nil, err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	var link model.AgencyCompanySubaccount
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Where("phone_country_code = ? AND phone_number = ?", phoneCode, phone).Count(&count).Error; err != nil || count > 0 {
			return errcode.New(errcode.CodeValidationError, "phone number is already registered")
		}
		if email != "" {
			if err := tx.Model(&model.UserCredential{}).Where("email = ?", email).Count(&count).Error; err != nil || count > 0 {
				return errcode.New(errcode.CodeValidationError, "email is already registered")
			}
		}
		user := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: phoneCode, PhoneNumber: phone, MemberStatus: "active", MemberType: MemberTypeUser, IsVerifiedPhone: false}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		credential := model.UserCredential{UserID: user.ID, PasswordHash: hash, PasswordEncrypted: encrypted, IsVerified: email != ""}
		if email != "" {
			credential.Email = &email
		}
		if err := tx.Create(&credential).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.UserProfile{UserID: user.ID, DisplayName: displayName, AccountType: AccountTypeAgencyCompanySubaccount, PublisherIdentityType: "agent"}).Error; err != nil {
			return err
		}
		permissionJSON, _ := json.Marshal(permissions)
		link = model.AgencyCompanySubaccount{PublicID: utils.NewPublicID(), CompanyOwnerUserID: ownerUserID, ChildUserID: user.ID, DisplayName: displayName, Status: "active", Permissions: permissionJSON}
		return tx.Create(&link).Error
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to create company subaccount")
	}
	return s.getSubaccountResponse(ctx, ownerUserID, link.PublicID)
}

// 2. ListSubaccounts returns company-owned child logins.
func (s *AgencyProfileService) ListSubaccounts(ctx context.Context, ownerUserID int64) ([]AgencySubaccountResponse, error) {
	if err := s.requireApprovedCompanyOwner(ctx, s.runtime.DB, ownerUserID); err != nil {
		return nil, err
	}
	var links []model.AgencyCompanySubaccount
	if err := s.runtime.DB.WithContext(ctx).Where("company_owner_user_id = ?", ownerUserID).Order("id desc").Find(&links).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load company subaccounts")
	}
	items := make([]AgencySubaccountResponse, 0, len(links))
	for _, link := range links {
		item, err := s.buildSubaccountResponse(ctx, &link)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

// 3. UpdateSubaccountStatus enables or disables one company child.
func (s *AgencyProfileService) UpdateSubaccountStatus(ctx context.Context, ownerUserID int64, publicID string, status string) (*AgencySubaccountResponse, error) {
	if status != "active" && status != "disabled" {
		return nil, errcode.New(errcode.CodeValidationError, "subaccount status is invalid")
	}
	if err := s.requireApprovedCompanyOwner(ctx, s.runtime.DB, ownerUserID); err != nil {
		return nil, err
	}
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var link model.AgencyCompanySubaccount
		if err := tx.Where("public_id = ? AND company_owner_user_id = ?", strings.TrimSpace(publicID), ownerUserID).First(&link).Error; err != nil {
			return errcode.New(errcode.CodeNotFound, "company subaccount not found")
		}
		if err := tx.Model(&link).Update("status", status).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Where("id = ?", link.ChildUserID).Update("member_status", status).Error
	})
	if err != nil {
		return nil, err
	}
	return s.getSubaccountResponse(ctx, ownerUserID, publicID)
}

// 4. DeleteSubaccount revokes a child while retaining its company and listing audit link.
func (s *AgencyProfileService) DeleteSubaccount(ctx context.Context, ownerUserID int64, publicID string) error {
	if err := s.requireApprovedCompanyOwner(ctx, s.runtime.DB, ownerUserID); err != nil {
		return err
	}
	return s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var link model.AgencyCompanySubaccount
		if err := tx.Where("public_id = ? AND company_owner_user_id = ?", strings.TrimSpace(publicID), ownerUserID).First(&link).Error; err != nil {
			return errcode.New(errcode.CodeNotFound, "company subaccount not found")
		}
		if err := tx.Model(&model.User{}).Where("id = ?", link.ChildUserID).Update("member_status", "disabled").Error; err != nil {
			return err
		}
		return tx.Model(&link).Update("status", "disabled").Error
	})
}

// 5. ValidateSubaccountAccess invalidates existing child tokens when parent or link is unavailable.
func (s *AgencyProfileService) ValidateSubaccountAccess(ctx context.Context, childUserID int64) error {
	var link model.AgencyCompanySubaccount
	if err := s.runtime.DB.WithContext(ctx).Where("child_user_id = ? AND status = ?", childUserID, "active").First(&link).Error; err != nil {
		return errcode.New(errcode.CodeAuthForbidden, "company subaccount is disabled")
	}
	return s.requireApprovedCompanyOwner(ctx, s.runtime.DB, link.CompanyOwnerUserID)
}

// 5.1 LoadSubaccountPermissions returns the current child permission set.
func (s *AgencyProfileService) LoadSubaccountPermissions(ctx context.Context, childUserID int64) ([]string, error) {
	if err := s.ValidateSubaccountAccess(ctx, childUserID); err != nil {
		return nil, err
	}
	var link model.AgencyCompanySubaccount
	if err := s.runtime.DB.WithContext(ctx).Where("child_user_id = ?", childUserID).First(&link).Error; err != nil {
		return nil, errcode.New(errcode.CodeAuthForbidden, "company subaccount is disabled")
	}
	permissions := []string{}
	_ = json.Unmarshal(link.Permissions, &permissions)
	return permissions, nil
}

// 6. requireApprovedCompanyOwner verifies an active company account and profile.
func (s *AgencyProfileService) requireApprovedCompanyOwner(ctx context.Context, db *gorm.DB, ownerUserID int64) error {
	var count int64
	err := db.WithContext(ctx).Table("users").Joins("JOIN user_profiles ON user_profiles.user_id = users.id").Joins("JOIN agency_profile_bindings ON agency_profile_bindings.user_id = users.id").Joins("JOIN agency_profiles ON agency_profiles.id = agency_profile_bindings.active_profile_id").Where("users.id = ? AND users.member_status = ? AND user_profiles.account_type = ? AND agency_profiles.profile_type = ? AND agency_profiles.status = ?", ownerUserID, "active", AccountTypeAgencyCompany, AgencyProfileTypeCompany, AgencyProfileStatusApproved).Count(&count).Error
	if err != nil || count != 1 {
		return errcode.New(errcode.CodeAuthForbidden, "approved agency company is required")
	}
	return nil
}

// 7. normalizeAgencySubaccountPermissions validates the bounded permission set.
func normalizeAgencySubaccountPermissions(values []string) ([]string, error) {
	result, seen := make([]string, 0, len(values)), map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, ok := agencySubaccountPermissionSet[value]; !ok {
			return nil, errcode.New(errcode.CodeValidationError, "subaccount permission is invalid")
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result, nil
}

// 8. getSubaccountResponse loads one owner-controlled child response.
func (s *AgencyProfileService) getSubaccountResponse(ctx context.Context, ownerUserID int64, publicID string) (*AgencySubaccountResponse, error) {
	var link model.AgencyCompanySubaccount
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ? AND company_owner_user_id = ?", strings.TrimSpace(publicID), ownerUserID).First(&link).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "company subaccount not found")
	}
	return s.buildSubaccountResponse(ctx, &link)
}

// 9. buildSubaccountResponse joins child account and credential data.
func (s *AgencyProfileService) buildSubaccountResponse(ctx context.Context, link *model.AgencyCompanySubaccount) (*AgencySubaccountResponse, error) {
	type row struct {
		PublicID, PhoneCountryCode, PhoneNumber string
		Email                                   *string
	}
	var value row
	if err := s.runtime.DB.WithContext(ctx).Table("users").Select("users.public_id, users.phone_country_code, users.phone_number, user_credentials.email").Joins("LEFT JOIN user_credentials ON user_credentials.user_id = users.id").Where("users.id = ?", link.ChildUserID).Scan(&value).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load company subaccount")
	}
	permissions := []string{}
	_ = json.Unmarshal(link.Permissions, &permissions)
	email := ""
	if value.Email != nil {
		email = *value.Email
	}
	return &AgencySubaccountResponse{PublicID: link.PublicID, UserPublicID: value.PublicID, DisplayName: link.DisplayName, PhoneCountryCode: value.PhoneCountryCode, PhoneNumber: value.PhoneNumber, Email: email, Status: link.Status, Permissions: permissions, CreatedAt: link.CreatedAt.UTC().Format(time.RFC3339)}, nil
}
