/*
 * Agency profile write workflows.
 * 1. Create and edit one account-owned working revision.
 * 2. Submit revisions without replacing approved data.
 * 3. Promote or reject revisions through staff review.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const agencyProfileEditCooldown = 6 * time.Hour

// 1. AgencyProfileService manages versioned agency profiles and company children.
type AgencyProfileService struct {
	runtime *Runtime
}

// AgencyCompanyService remains a source-compatible alias for server wiring.
type AgencyCompanyService = AgencyProfileService

// 2. NewAgencyCompanyService creates the agency profile service.
func NewAgencyCompanyService(runtime *Runtime) *AgencyProfileService {
	return &AgencyProfileService{runtime: runtime}
}

// 3. CreateMemberAgencyProfile creates the only working revision for an agency account.
func (s *AgencyProfileService) CreateMemberAgencyProfile(ctx context.Context, userID int64, params AgencyProfileUpsertParams) (*AgencyProfileMemberResponse, error) {
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.validateAgencyAccountType(ctx, tx, userID, params.ProfileType); err != nil {
			return err
		}
		binding, err := s.loadAgencyProfileBindingForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if binding.RevisionProfileID != nil {
			return errcode.New(errcode.CodeValidationError, "agency profile revision already exists")
		}
		if !s.canEditAgencyProfile(binding, s.agencyProfileNow()) {
			return errcode.New(errcode.CodeRateLimited, "您的修改過於頻繁")
		}
		profile := model.AgencyProfile{PublicID: utils.NewPublicID(), UserID: userID, Status: AgencyProfileStatusDraft}
		if err := s.applyAgencyProfileDraft(ctx, tx, &profile, userID, params); err != nil {
			return err
		}
		if err := tx.Create(&profile).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to create agency profile")
		}
		binding.RevisionProfileID = &profile.ID
		now := s.agencyProfileNow()
		binding.LastEditedAt = &now
		return tx.Save(binding).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetMemberAgencyProfile(ctx, userID)
}

// 4. UpdateMemberAgencyProfile updates a draft or rejected revision.
func (s *AgencyProfileService) UpdateMemberAgencyProfile(ctx context.Context, userID int64, params AgencyProfileUpsertParams) (*AgencyProfileMemberResponse, error) {
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.validateAgencyAccountType(ctx, tx, userID, params.ProfileType); err != nil {
			return err
		}
		binding, err := s.loadAgencyProfileBindingForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if binding.RevisionProfileID == nil {
			return errcode.New(errcode.CodeNotFound, "agency profile revision not found")
		}
		if !s.canEditAgencyProfile(binding, s.agencyProfileNow()) {
			return errcode.New(errcode.CodeRateLimited, "您的修改過於頻繁")
		}
		profile, err := s.loadAgencyProfileForUpdate(ctx, tx, *binding.RevisionProfileID, userID)
		if err != nil {
			return err
		}
		if profile.Status != AgencyProfileStatusDraft && profile.Status != AgencyProfileStatusRejected {
			return errcode.New(errcode.CodeValidationError, "agency profile revision cannot be edited")
		}
		if err := s.applyAgencyProfileDraft(ctx, tx, profile, userID, params); err != nil {
			return err
		}
		profile.Status, profile.ReviewNote = AgencyProfileStatusDraft, ""
		profile.SubmittedAt, profile.ReviewedAt, profile.ReviewedByUserID = nil, nil, nil
		if err := tx.Save(profile).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update agency profile")
		}
		now := s.agencyProfileNow()
		binding.LastEditedAt = &now
		return tx.Save(binding).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetMemberAgencyProfile(ctx, userID)
}

// 5. SubmitMemberAgencyProfile submits a complete revision for staff review.
func (s *AgencyProfileService) SubmitMemberAgencyProfile(ctx context.Context, userID int64) (*AgencyProfileMemberResponse, error) {
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		binding, err := s.loadAgencyProfileBindingForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if binding.RevisionProfileID == nil {
			return errcode.New(errcode.CodeNotFound, "agency profile revision not found")
		}
		profile, err := s.loadAgencyProfileForUpdate(ctx, tx, *binding.RevisionProfileID, userID)
		if err != nil {
			return err
		}
		if profile.Status != AgencyProfileStatusDraft && profile.Status != AgencyProfileStatusRejected {
			return errcode.New(errcode.CodeValidationError, "agency profile revision cannot be submitted")
		}
		if err := s.validateAgencyProfileSubmission(ctx, tx, profile, userID); err != nil {
			return err
		}
		now := s.agencyProfileNow()
		profile.Status, profile.ReviewNote, profile.SubmittedAt = AgencyProfileStatusPending, "", &now
		profile.ReviewedAt, profile.ReviewedByUserID = nil, nil
		if err := tx.Save(profile).Error; err != nil {
			return err
		}
		if binding.ActiveProfileID == nil {
			return tx.Model(&model.User{}).Where("id = ?", userID).Update("member_status", "pending_review").Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetMemberAgencyProfile(ctx, userID)
}

// 6. ReviewAgencyProfile approves or rejects one pending revision.
func (s *AgencyProfileService) ReviewAgencyProfile(ctx context.Context, publicID string, params AgencyProfileReviewParams) (*AgencyProfileStaffResponse, error) {
	publicID = strings.TrimSpace(publicID)
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		profile, err := s.loadAgencyProfileByPublicIDForUpdate(ctx, tx, publicID)
		if err != nil {
			return err
		}
		if profile.Status != AgencyProfileStatusPending {
			return errcode.New(errcode.CodeValidationError, "agency profile is not pending review")
		}
		binding, err := s.loadAgencyProfileBindingForUpdate(ctx, tx, profile.UserID)
		if err != nil {
			return err
		}
		if binding.RevisionProfileID == nil || *binding.RevisionProfileID != profile.ID {
			return errcode.New(errcode.CodeValidationError, "agency profile revision is no longer current")
		}
		now, reviewerID := s.agencyProfileNow(), params.ReviewerUserID
		profile.ReviewedAt, profile.ReviewedByUserID = &now, &reviewerID
		profile.ReviewNote = strings.TrimSpace(params.ReviewNote)
		if params.Approved {
			var duplicateCount int64
			if strings.TrimSpace(profile.LicenseNumber) != "" {
				if err := tx.Model(&model.AgencyProfile{}).Joins("JOIN agency_profile_bindings ON agency_profile_bindings.active_profile_id = agency_profiles.id").Where("agency_profiles.id <> ? AND agency_profiles.user_id <> ? AND agency_profiles.profile_type = ? AND agency_profiles.license_number = ? AND agency_profiles.status = ?", profile.ID, profile.UserID, profile.ProfileType, profile.LicenseNumber, AgencyProfileStatusApproved).Count(&duplicateCount).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to validate agency licence")
				}
				if duplicateCount > 0 {
					return errcode.New(errcode.CodeValidationError, "agency licence is already approved for another account")
				}
			}
			profile.Status = AgencyProfileStatusApproved
			binding.ActiveProfileID, binding.RevisionProfileID = &profile.ID, nil
			if err := tx.Model(&model.User{}).Where("id = ?", profile.UserID).Update("member_status", "active").Error; err != nil {
				return err
			}
		} else {
			if profile.ReviewNote == "" {
				return errcode.New(errcode.CodeValidationError, "review note is required when rejecting")
			}
			profile.Status = AgencyProfileStatusRejected
			if binding.ActiveProfileID == nil {
				if err := tx.Model(&model.User{}).Where("id = ?", profile.UserID).Update("member_status", "rejected").Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Save(profile).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to save agency profile review")
		}
		if err := tx.Save(binding).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update agency profile binding")
		}
		if params.Approved {
			if err := s.syncApprovedAgencyProfileListings(ctx, tx, profile); err != nil {
				return err
			}
		}
		return s.createAgencyProfileReviewNotification(ctx, tx, profile, params.Approved)
	})
	if err != nil {
		return nil, err
	}
	result, err := s.GetAgencyProfileForStaff(ctx, publicID)
	if err != nil {
		return nil, err
	}
	s.sendAgencyProfileReviewEmail(ctx, publicID, params.Approved, strings.TrimSpace(params.ReviewNote))
	return result, nil
}

// 7. loadAgencyProfileBindingForUpdate locks or creates the account binding.
func (s *AgencyProfileService) loadAgencyProfileBindingForUpdate(ctx context.Context, tx *gorm.DB, userID int64) (*model.AgencyProfileBinding, error) {
	var binding model.AgencyProfileBinding
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&binding).Error
	if err == nil {
		return &binding, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile binding")
	}
	binding = model.AgencyProfileBinding{UserID: userID}
	if err := tx.Create(&binding).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create agency profile binding")
	}
	return &binding, nil
}

// 8. loadAgencyProfileForUpdate loads one account-owned revision under lock.
func (s *AgencyProfileService) loadAgencyProfileForUpdate(ctx context.Context, tx *gorm.DB, profileID int64, userID int64) (*model.AgencyProfile, error) {
	var profile model.AgencyProfile
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", profileID, userID).First(&profile).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "agency profile revision not found")
	}
	return &profile, nil
}

// 9. loadAgencyProfileByPublicIDForUpdate loads a staff target under lock.
func (s *AgencyProfileService) loadAgencyProfileByPublicIDForUpdate(ctx context.Context, tx *gorm.DB, publicID string) (*model.AgencyProfile, error) {
	var profile model.AgencyProfile
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("public_id = ?", publicID).First(&profile).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "agency profile not found")
	}
	return &profile, nil
}

// 10. canEditAgencyProfile checks the six-hour revision cooldown.
func (s *AgencyProfileService) canEditAgencyProfile(binding *model.AgencyProfileBinding, now time.Time) bool {
	return binding.LastEditedAt == nil || !now.Before(binding.LastEditedAt.Add(agencyProfileEditCooldown))
}

// 11. agencyProfileNow returns the injected clock with production fallback.
func (s *AgencyProfileService) agencyProfileNow() time.Time {
	if s.runtime.Now != nil {
		return s.runtime.Now().UTC()
	}
	return time.Now().UTC()
}
