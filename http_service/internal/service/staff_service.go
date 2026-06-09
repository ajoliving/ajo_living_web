/*
 * Staff management business logic.
 * 1. Provide staff-only user listing and account creation APIs.
 * 2. Keep staff access controlled by the is_staff flag only.
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
	"ajoliving_web/http_service/internal/utils"
)

// 1. StaffService handles staff-only account management flows.
type StaffService struct {
	runtime *Runtime
}

// 2. StaffUserSummary defines the staff-facing user payload.
type StaffUserSummary struct {
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
	PublisherIdentity string             `json:"publisher_identity_type"`
	DistrictCode      string             `json:"district_code"`
	ResidenceFloor    string             `json:"residence_floor"`
	ResidenceUnit     string             `json:"residence_unit"`
	PrimaryCommunity  *CommunityResponse `json:"primary_community,omitempty"`
	AJOBalance        int64              `json:"ajo_balance"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// 3. StaffUserListFilters defines staff user list filters.
type StaffUserListFilters struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	IsStaff  *bool
}

// 4. StaffUserRoleUpdateParams defines staff flag update input.
type StaffUserRoleUpdateParams struct {
	IsStaff *bool
}

// 5. StaffUserCreateParams defines staff-created account input.
type StaffUserCreateParams struct {
	Email                 string
	Password              string
	DisplayName           string
	PhoneCountryCode      string
	PhoneNumber           string
	PublisherIdentityType string
	PrimaryCommunityID    string
	PrimaryCommunityName  string
	ResidenceFloor        string
	ResidenceUnit         string
	DistrictCode          string
	IsStaff               bool
}

// 6. NewStaffService creates a staff service instance.
func NewStaffService(runtime *Runtime) *StaffService {
	return &StaffService{runtime: runtime}
}

// 7. GetStaffMe returns the current staff account payload.
func (s *StaffService) GetStaffMe(ctx context.Context, userID int64) (*StaffUserSummary, error) {
	user, profile, err := s.loadUserWithProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	emails, err := s.loadCredentialEmailsByUserIDs(ctx, []model.User{*user})
	if err != nil {
		return nil, err
	}

	balances, err := s.loadWalletBalancesByUserIDs(ctx, []model.User{*user})
	if err != nil {
		return nil, err
	}

	return s.toStaffUserSummary(user, profile, emails[user.ID], balances[user.ID]), nil
}

// 8. ListUsers returns paginated user records for staff operations.
func (s *StaffService) ListUsers(ctx context.Context, filters StaffUserListFilters) ([]StaffUserSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.User{})

	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where(
			"phone_country_code LIKE ? OR phone_number LIKE ? OR public_id LIKE ?",
			pattern,
			pattern,
			pattern,
		)
	}

	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("member_status = ?", status)
	}

	if filters.IsStaff != nil {
		query = query.Where("is_staff = ?", *filters.IsStaff)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count users")
	}

	var users []model.User
	if err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load users")
	}

	profiles, err := s.loadProfilesByUserIDs(ctx, users)
	if err != nil {
		return nil, nil, err
	}

	emails, err := s.loadCredentialEmailsByUserIDs(ctx, users)
	if err != nil {
		return nil, nil, err
	}

	balances, err := s.loadWalletBalancesByUserIDs(ctx, users)
	if err != nil {
		return nil, nil, err
	}

	items := make([]StaffUserSummary, 0, len(users))
	for _, user := range users {
		profile := profiles[user.ID]
		items = append(items, *s.toStaffUserSummary(&user, profile, emails[user.ID], balances[user.ID]))
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 9. CreateUser creates a staff-managed email password account.
func (s *StaffService) CreateUser(ctx context.Context, operatorUserID int64, params StaffUserCreateParams) (*StaffUserSummary, error) {
	_ = operatorUserID

	email := normalizeEmail(params.Email)
	password := strings.TrimSpace(params.Password)
	phoneCountryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	if !isValidEmail(email) || !isValidPhone(phoneCountryCode, phoneNumber) || len(password) < 8 {
		return nil, errcode.New(errcode.CodeValidationError, "valid email, phone number, and password are required")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}

	var user model.User
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.UserCredential{}).Where("email = ?", email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "email is already registered")
		}
		if err := tx.Model(&model.User{}).Where("phone_country_code = ? AND phone_number = ?", phoneCountryCode, phoneNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "phone number is already registered")
		}

		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: phoneCountryCode,
			PhoneNumber:      phoneNumber,
			MemberStatus:     "active",
			MemberType:       MemberTypeUser,
			IsStaff:          params.IsStaff,
			IsVerifiedPhone:  false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.UserCredential{
			UserID:       user.ID,
			Email:        &email,
			PasswordHash: passwordHash,
			IsVerified:   true,
		}).Error; err != nil {
			return err
		}

		displayName := strings.TrimSpace(params.DisplayName)
		if displayName == "" {
			displayName = email
		}
		profile := model.UserProfile{
			UserID:                user.ID,
			DisplayName:           displayName,
			PublisherIdentityType: strings.TrimSpace(params.PublisherIdentityType),
			ResidenceFloor:        strings.TrimSpace(params.ResidenceFloor),
			ResidenceUnit:         strings.TrimSpace(params.ResidenceUnit),
			DistrictCode:          strings.TrimSpace(params.DistrictCode),
		}
		if strings.TrimSpace(params.PrimaryCommunityID) != "" {
			community, err := s.resolveStaffCreateCommunity(ctx, tx, params.PrimaryCommunityID, params.PrimaryCommunityName)
			if err != nil {
				return err
			}
			profile.PrimaryCommunityID = &community.ID
			if profile.DistrictCode == "" {
				profile.DistrictCode = community.DistrictCode
			}
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to create user")
	}

	return s.GetStaffMe(ctx, user.ID)
}

// 9.1 resolveStaffCreateCommunity loads or mirrors a POS building for staff-created users.
func (s *StaffService) resolveStaffCreateCommunity(ctx context.Context, tx *gorm.DB, publicID string, name string) (*model.Community, error) {
	trimmedPublicID := strings.TrimSpace(publicID)
	var community model.Community
	err := tx.WithContext(ctx).Where("public_id = ?", trimmedPublicID).First(&community).Error
	if err == nil {
		return &community, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load community")
	}
	if len(trimmedPublicID) == 0 || len(trimmedPublicID) > 26 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid primary community")
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = trimmedPublicID
	}

	community = model.Community{
		PublicID:      trimmedPublicID,
		CommunityType: "building",
		NameZH:        displayName,
		NameEN:        displayName,
		DistrictCode:  "",
		AddressText:   displayName,
	}
	if err := tx.WithContext(ctx).Create(&community).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create community")
	}

	return &community, nil
}

// 10. UpdateUserRole updates a target user's staff flag.
func (s *StaffService) UpdateUserRole(ctx context.Context, operatorUserID int64, targetPublicID string, params StaffUserRoleUpdateParams) (*StaffUserSummary, error) {
	if params.IsStaff == nil {
		return nil, errcode.New(errcode.CodeValidationError, "is_staff is required")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(targetPublicID)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "user not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user")
	}

	_ = operatorUserID
	user.MemberType = MemberTypeUser
	user.IsStaff = *params.IsStaff

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"member_type": user.MemberType,
			"is_staff":    user.IsStaff,
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update user role")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetStaffMe(ctx, user.ID)
}

// 11. loadUserWithProfile loads a user record with profile and community data.
func (s *StaffService) loadUserWithProfile(ctx context.Context, userID int64) (*model.User, *model.UserProfile, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeNotFound, "user not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load user")
	}

	var profile model.UserProfile
	err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", user.ID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load user profile")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &user, nil, nil
	}

	return &user, &profile, nil
}

// 12. loadProfilesByUserIDs loads profile data for a batch of users.
func (s *StaffService) loadProfilesByUserIDs(ctx context.Context, users []model.User) (map[int64]*model.UserProfile, error) {
	if len(users) == 0 {
		return map[int64]*model.UserProfile{}, nil
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var profiles []model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id IN ?", userIDs).Find(&profiles).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user profiles")
	}

	result := make(map[int64]*model.UserProfile, len(profiles))
	for i := range profiles {
		profile := profiles[i]
		result[profile.UserID] = &profile
	}

	return result, nil
}

// 13. loadCredentialEmailsByUserIDs loads login email data for a batch of users.
func (s *StaffService) loadCredentialEmailsByUserIDs(ctx context.Context, users []model.User) (map[int64]string, error) {
	if len(users) == 0 {
		return map[int64]string{}, nil
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var credentials []model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Select("user_id", "email").Where("user_id IN ?", userIDs).Find(&credentials).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user credentials")
	}

	result := make(map[int64]string, len(credentials))
	for _, credential := range credentials {
		if credential.Email != nil {
			result[credential.UserID] = *credential.Email
		}
	}

	return result, nil
}

// 14. loadWalletBalancesByUserIDs loads current AJO Point balances for staff lists.
func (s *StaffService) loadWalletBalancesByUserIDs(ctx context.Context, users []model.User) (map[int64]int64, error) {
	if len(users) == 0 {
		return map[int64]int64{}, nil
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var accounts []model.WalletAccount
	if err := s.runtime.DB.WithContext(ctx).Select("user_id", "balance").Where("user_id IN ?", userIDs).Find(&accounts).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load wallet balances")
	}

	result := make(map[int64]int64, len(accounts))
	for _, account := range accounts {
		result[account.UserID] = account.Balance
	}

	return result, nil
}

// 15. toStaffUserSummary maps a user model to the staff-facing payload.
func (s *StaffService) toStaffUserSummary(user *model.User, profile *model.UserProfile, email string, ajoBalance int64) *StaffUserSummary {
	if user == nil {
		return nil
	}

	access := buildAccessSnapshot(user.IsStaff)

	summary := &StaffUserSummary{
		PublicID:         user.PublicID,
		Email:            email,
		PhoneCountryCode: user.PhoneCountryCode,
		PhoneNumber:      user.PhoneNumber,
		MemberStatus:     user.MemberStatus,
		MemberType:       normalizeMemberType(user.MemberType),
		IsStaff:          access.IsStaff,
		Role:             resolveUserRole(user.MemberType, access.IsStaff),
		Roles:            access.RoleCodes,
		Permissions:      access.Permissions,
		AJOBalance:       ajoBalance,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if profile != nil {
		summary.DisplayName = profile.DisplayName
		summary.PublisherIdentity = profile.PublisherIdentityType
		summary.DistrictCode = profile.DistrictCode
		summary.ResidenceFloor = profile.ResidenceFloor
		summary.ResidenceUnit = profile.ResidenceUnit
		summary.PrimaryCommunity = toCommunityResponse(profile.PrimaryCommunity)
	}

	return summary
}
