/*
 * Staff management business logic.
 * 1. Provide staff-only user listing and role management APIs.
 * 2. Keep role promotion and staff access rules centralized.
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

// 1. StaffService handles staff-only account management flows.
type StaffService struct {
	runtime *Runtime
}

// 2. StaffUserSummary defines the staff-facing user payload.
type StaffUserSummary struct {
	PublicID          string             `json:"public_id"`
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
	PrimaryCommunity  *CommunityResponse `json:"primary_community,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// 3. StaffUserListFilters defines staff user list filters.
type StaffUserListFilters struct {
	Page       int
	PageSize   int
	Keyword    string
	MemberType string
	RoleCode   string
	IsStaff    *bool
}

// 4. StaffUserRoleUpdateParams defines staff role update input.
type StaffUserRoleUpdateParams struct {
	MemberType *string
	RoleCodes  *[]string
	IsStaff    *bool
}

// 5. NewStaffService creates a staff service instance.
func NewStaffService(runtime *Runtime) *StaffService {
	return &StaffService{runtime: runtime}
}

// 6. GetStaffMe returns the current staff account payload.
func (s *StaffService) GetStaffMe(ctx context.Context, userID int64) (*StaffUserSummary, error) {
	user, profile, err := s.loadUserWithProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	access, err := NewAccessService(s.runtime).ResolveUserAccess(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.toStaffUserSummary(user, profile, access), nil
}

// 7. ListUsers returns paginated user records for staff operations.
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

	if memberType := strings.TrimSpace(filters.MemberType); memberType != "" {
		if !isValidMemberType(memberType) {
			return nil, nil, errcode.New(errcode.CodeValidationError, "member_type is invalid")
		}
		query = query.Where("member_type = ?", normalizeMemberType(memberType))
	}

	if roleCode := strings.TrimSpace(filters.RoleCode); roleCode != "" {
		query = query.
			Joins("JOIN user_role_bindings ON user_role_bindings.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_role_bindings.role_id").
			Where("roles.code = ?", roleCode).
			Distinct("users.id")
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

	accessService := NewAccessService(s.runtime)
	items := make([]StaffUserSummary, 0, len(users))
	for _, user := range users {
		profile := profiles[user.ID]
		access, err := accessService.ResolveUserAccess(ctx, &user)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *s.toStaffUserSummary(&user, profile, access))
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 8. UpdateUserRole updates a target user's front-end member type and staff flag.
func (s *StaffService) UpdateUserRole(ctx context.Context, operatorUserID int64, targetPublicID string, params StaffUserRoleUpdateParams) (*StaffUserSummary, error) {
	if params.MemberType == nil && params.RoleCodes == nil && params.IsStaff == nil {
		return nil, errcode.New(errcode.CodeValidationError, "at least one role field is required")
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(targetPublicID)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "user not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user")
	}

	accessService := NewAccessService(s.runtime)
	currentAccess, err := accessService.ResolveUserAccess(ctx, &user)
	if err != nil {
		return nil, err
	}

	roleCodes := append([]string{}, currentAccess.RoleCodes...)
	if params.RoleCodes != nil {
		roleCodes = append([]string{}, *params.RoleCodes...)
	}
	if params.IsStaff != nil && !*params.IsStaff && containsStaffRole(roleCodes) && params.RoleCodes != nil {
		return nil, errcode.New(errcode.CodeValidationError, "role_codes contains staff roles while is_staff is false")
	}
	if params.IsStaff != nil && !*params.IsStaff {
		roleCodes = dropStaffRoles(roleCodes)
	}
	if params.IsStaff != nil && *params.IsStaff && !containsStaffRole(roleCodes) {
		roleCodes = append(roleCodes, model.RoleCodeStaff)
	}

	if params.MemberType != nil {
		memberType := strings.TrimSpace(*params.MemberType)
		if !isValidMemberType(memberType) {
			return nil, errcode.New(errcode.CodeValidationError, "member_type is invalid")
		}
		user.MemberType = normalizeMemberType(memberType)
	}

	user.IsStaff = containsStaffRole(roleCodes)
	if params.IsStaff != nil {
		user.IsStaff = *params.IsStaff
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"member_type": user.MemberType,
			"is_staff":    user.IsStaff,
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update user role")
		}

		return accessService.SetUserRoleCodes(ctx, tx, &user, roleCodes, &operatorUserID, model.RoleCodeStaff)
	}); err != nil {
		return nil, err
	}

	return s.GetStaffMe(ctx, user.ID)
}

// 9. loadUserWithProfile loads a user record with profile and community data.
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

// 10. loadProfilesByUserIDs loads profile data for a batch of users.
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

// 11. toStaffUserSummary maps a user model to the staff-facing payload.
func (s *StaffService) toStaffUserSummary(user *model.User, profile *model.UserProfile, access *AccessSnapshot) *StaffUserSummary {
	if user == nil {
		return nil
	}

	roleCodes := []string{}
	permissionCodes := []string{}
	isStaff := false
	if access != nil {
		roleCodes = access.RoleCodes
		permissionCodes = access.Permissions
		isStaff = access.IsStaff
	}

	summary := &StaffUserSummary{
		PublicID:         user.PublicID,
		PhoneCountryCode: user.PhoneCountryCode,
		PhoneNumber:      user.PhoneNumber,
		MemberStatus:     user.MemberStatus,
		MemberType:       normalizeMemberType(user.MemberType),
		IsStaff:          isStaff,
		Role:             resolveUserRole(user.MemberType, isStaff),
		Roles:            roleCodes,
		Permissions:      permissionCodes,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if profile != nil {
		summary.DisplayName = profile.DisplayName
		summary.PublisherIdentity = profile.PublisherIdentityType
		summary.DistrictCode = profile.DistrictCode
		summary.PrimaryCommunity = toCommunityResponse(profile.PrimaryCommunity)
	}

	return summary
}

// 12. ListRoles returns the available role catalog for staff tooling.
func (s *StaffService) ListRoles(ctx context.Context) ([]RoleCatalogItem, error) {
	return NewAccessService(s.runtime).ListRoles(ctx)
}
