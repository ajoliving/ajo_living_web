/*
 * Access control business logic.
 * 1. Resolve role codes and permission codes for one user account.
 * 2. Keep implicit member roles and explicit staff roles synchronized.
 * 3. Expose staff-facing role catalog payloads.
 */
package service

import (
	"context"
	"slices"
	"sort"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. AccessService resolves user access and manages role bindings.
type AccessService struct {
	runtime *Runtime
}

// 2. AccessSnapshot defines resolved role and permission data for one user.
type AccessSnapshot struct {
	RoleCodes   []string
	Permissions []string
	IsStaff     bool
}

// 3. RoleCatalogItem defines one staff-visible role catalog item.
type RoleCatalogItem struct {
	Code        string   `json:"code"`
	Scope       string   `json:"scope"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

var rolePriority = map[string]int{
	model.RoleCodeSuperAdmin: 10,
	model.RoleCodeStaff:      20,
	model.RoleCodeMember:     30,
}

// 4. NewAccessService creates an access service instance.
func NewAccessService(runtime *Runtime) *AccessService {
	return &AccessService{runtime: runtime}
}

// 5. ResolveUserAccess loads one user's role codes and permission codes.
func (s *AccessService) ResolveUserAccess(ctx context.Context, user *model.User) (*AccessSnapshot, error) {
	if user == nil {
		return nil, errcode.New(errcode.CodeValidationError, "user is required")
	}

	roleCodes, err := s.loadUserRoleCodes(ctx, s.runtime.DB, user.ID)
	if err != nil {
		return nil, err
	}

	finalRoleCodes := s.mergeImplicitRoleCodes(user, roleCodes, "")
	return &AccessSnapshot{
		RoleCodes:   finalRoleCodes,
		Permissions: permissionsForRoleCodes(finalRoleCodes),
		IsStaff:     containsStaffRole(finalRoleCodes),
	}, nil
}

// 6. EnsureUserRoles syncs the user's existing explicit roles with implicit defaults.
func (s *AccessService) EnsureUserRoles(ctx context.Context, tx *gorm.DB, user *model.User, assignedBy *int64, preferredStaffRole string) error {
	roleCodes, err := s.loadUserRoleCodes(ctx, tx, user.ID)
	if err != nil {
		return err
	}

	return s.SetUserRoleCodes(ctx, tx, user, roleCodes, assignedBy, preferredStaffRole)
}

// 7. SetUserRoleCodes replaces one user's role bindings while keeping implicit member roles.
func (s *AccessService) SetUserRoleCodes(ctx context.Context, tx *gorm.DB, user *model.User, roleCodes []string, assignedBy *int64, preferredStaffRole string) error {
	if user == nil {
		return errcode.New(errcode.CodeValidationError, "user is required")
	}

	finalRoleCodes := s.mergeImplicitRoleCodes(user, roleCodes, preferredStaffRole)
	if err := s.replaceUserRoleCodes(ctx, tx, user.ID, finalRoleCodes, assignedBy); err != nil {
		return err
	}

	user.IsStaff = containsStaffRole(finalRoleCodes)
	if err := tx.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Update("is_staff", user.IsStaff).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to update user staff flag")
	}

	return nil
}

// 8. ListRoles returns the seeded role catalog with permission codes.
func (s *AccessService) ListRoles(context.Context) ([]RoleCatalogItem, error) {
	definitions := model.DefaultRoleDefinitions()
	result := make([]RoleCatalogItem, 0, len(definitions))

	for _, item := range definitions {
		result = append(result, RoleCatalogItem{
			Code:        item.Code,
			Scope:       item.Scope,
			Name:        item.Name,
			Description: item.Description,
			Permissions: permissionsForRoleCodes([]string{item.Code}),
		})
	}

	sort.Slice(result, func(left int, right int) bool {
		return compareRoleCode(result[left].Code, result[right].Code) < 0
	})

	return result, nil
}

// 9. loadUserRoleCodes reads persisted role bindings for one user.
func (s *AccessService) loadUserRoleCodes(ctx context.Context, tx *gorm.DB, userID int64) ([]string, error) {
	type roleRow struct {
		Code string
	}

	var rows []roleRow
	if err := tx.WithContext(ctx).
		Table("user_role_bindings").
		Select("roles.code").
		Joins("JOIN roles ON roles.id = user_role_bindings.role_id").
		Where("user_role_bindings.user_id = ?", userID).
		Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user roles")
	}

	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.Code) == "" {
			continue
		}
		result = append(result, strings.TrimSpace(row.Code))
	}

	return normalizeRoleCodes(result), nil
}

// 10. mergeImplicitRoleCodes adds implicit member roles and optional default staff role.
func (s *AccessService) mergeImplicitRoleCodes(user *model.User, roleCodes []string, preferredStaffRole string) []string {
	explicitRoleCodes, _ := normalizeExplicitRoleCodes(roleCodes)
	if containsStaffRole(explicitRoleCodes) {
		preferredStaffRole = ""
	}

	result := append([]string{}, explicitRoleCodes...)
	result = append(result, model.RoleCodeMember)
	if user.IsStaff {
		if !containsStaffRole(result) {
			roleCode := strings.TrimSpace(preferredStaffRole)
			if roleCode == "" {
				roleCode = model.RoleCodeStaff
			}
			result = append(result, roleCode)
		}
	}

	return normalizeRoleCodes(result)
}

// 11. replaceUserRoleCodes writes one user's final role binding set.
func (s *AccessService) replaceUserRoleCodes(ctx context.Context, tx *gorm.DB, userID int64, roleCodes []string, assignedBy *int64) error {
	normalizedRoleCodes, err := normalizeExplicitRoleCodes(roleCodes)
	if err != nil {
		return err
	}

	var roles []model.Role
	if err := tx.WithContext(ctx).Where("code IN ?", normalizedRoleCodes).Find(&roles).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load role catalog")
	}
	if len(roles) != len(normalizedRoleCodes) {
		return errcode.New(errcode.CodeValidationError, "role_codes contains an unsupported role")
	}

	roleIDs := make(map[int64]struct{}, len(roles))
	for _, role := range roles {
		roleIDs[role.ID] = struct{}{}
	}

	var bindings []model.UserRoleBinding
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Find(&bindings).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load current user role bindings")
	}

	for _, binding := range bindings {
		if _, exists := roleIDs[binding.RoleID]; exists {
			delete(roleIDs, binding.RoleID)
			continue
		}

		if err := tx.WithContext(ctx).Delete(&model.UserRoleBinding{}, binding.ID).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to remove user role binding")
		}
	}

	roleByID := make(map[int64]model.Role, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}

	for roleID := range roleIDs {
		record := model.UserRoleBinding{
			UserID:     userID,
			RoleID:     roleID,
			AssignedBy: assignedBy,
			AssignedAt: s.runtime.Now(),
		}
		if err := tx.WithContext(ctx).Create(&record).Error; err != nil {
			_ = roleByID
			return errcode.New(errcode.CodeInternalError, "failed to create user role binding")
		}
	}

	return nil
}

// 12. normalizeExplicitRoleCodes validates and deduplicates requested role codes.
func normalizeExplicitRoleCodes(roleCodes []string) ([]string, error) {
	allowedRoleCodes := roleCatalog()
	seen := make(map[string]struct{}, len(roleCodes))
	result := make([]string, 0, len(roleCodes))

	for _, roleCode := range roleCodes {
		code := strings.TrimSpace(roleCode)
		if code == "" {
			continue
		}
		if _, exists := allowedRoleCodes[code]; !exists {
			return nil, errcode.New(errcode.CodeValidationError, "role_codes contains an unsupported role")
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}

	sort.Slice(result, func(left int, right int) bool {
		return compareRoleCode(result[left], result[right]) < 0
	})

	return result, nil
}

// 13. normalizeRoleCodes keeps the provided role codes unique and stable.
func normalizeRoleCodes(roleCodes []string) []string {
	result, err := normalizeExplicitRoleCodes(roleCodes)
	if err != nil {
		return []string{}
	}

	return result
}

// 14. permissionsForRoleCodes resolves permission codes from the default role matrix.
func permissionsForRoleCodes(roleCodes []string) []string {
	matrix := model.DefaultRolePermissionMatrix()
	seen := make(map[string]struct{})
	result := make([]string, 0, len(roleCodes))

	for _, roleCode := range roleCodes {
		for _, permissionCode := range matrix[roleCode] {
			if _, exists := seen[permissionCode]; exists {
				continue
			}
			seen[permissionCode] = struct{}{}
			result = append(result, permissionCode)
		}
	}

	sort.Strings(result)
	return result
}

// 15. containsStaffRole reports whether the role list contains any staff role.
func containsStaffRole(roleCodes []string) bool {
	catalog := roleCatalog()
	for _, roleCode := range roleCodes {
		definition, exists := catalog[roleCode]
		if exists && definition.Scope == model.RoleScopeStaff {
			return true
		}
	}

	return false
}

// 16. dropStaffRoles removes staff role codes from the provided list.
func dropStaffRoles(roleCodes []string) []string {
	result := make([]string, 0, len(roleCodes))
	for _, roleCode := range roleCodes {
		definition, exists := roleCatalog()[roleCode]
		if exists && definition.Scope == model.RoleScopeStaff {
			continue
		}
		result = append(result, roleCode)
	}

	return normalizeRoleCodes(result)
}

// 17. roleCatalog builds a code-keyed role catalog map.
func roleCatalog() map[string]model.SystemRoleDefinition {
	result := make(map[string]model.SystemRoleDefinition)
	for _, definition := range model.DefaultRoleDefinitions() {
		result[definition.Code] = definition
	}

	return result
}

// 18. compareRoleCode keeps role code ordering stable and predictable.
func compareRoleCode(left string, right string) int {
	leftPriority, leftExists := rolePriority[left]
	rightPriority, rightExists := rolePriority[right]
	if leftExists && rightExists && leftPriority != rightPriority {
		return leftPriority - rightPriority
	}
	if leftExists && !rightExists {
		return -1
	}
	if !leftExists && rightExists {
		return 1
	}
	return strings.Compare(left, right)
}

// 19. hasPermission reports whether the access snapshot includes the target permission.
func hasPermission(access *AccessSnapshot, permissionCode string) bool {
	if access == nil {
		return false
	}

	return slices.Contains(access.Permissions, strings.TrimSpace(permissionCode))
}
