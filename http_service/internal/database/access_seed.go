/*
 * Access control seed utilities.
 * 1. Insert the baseline RBAC role and permission catalog.
 * 2. Keep system role-permission bindings synchronized at startup.
 */
package database

import (
	"context"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. SeedAccessControl inserts the baseline RBAC catalog and bindings.
func SeedAccessControl(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		permissionsByCode, err := seedPermissions(tx)
		if err != nil {
			return err
		}

		rolesByCode, err := seedRoles(tx)
		if err != nil {
			return err
		}
		if err := pruneDeprecatedRoles(tx, rolesByCode); err != nil {
			return err
		}

		return seedRolePermissions(tx, rolesByCode, permissionsByCode)
	})
}

// 2. seedPermissions upserts the default permission catalog.
func seedPermissions(tx *gorm.DB) (map[string]model.Permission, error) {
	items := model.DefaultPermissionDefinitions()
	result := make(map[string]model.Permission, len(items))

	for _, definition := range items {
		record := model.Permission{}
		if err := tx.Where("code = ?", definition.Code).Limit(1).Find(&record).Error; err != nil {
			return nil, err
		}
		if record.ID == 0 {
			record = model.Permission{
				PublicID:    utils.NewPublicID(),
				Code:        definition.Code,
				Scope:       definition.Scope,
				Name:        definition.Name,
				Description: definition.Description,
				IsSystem:    true,
			}
			if createErr := tx.Create(&record).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			if updateErr := tx.Model(&record).Updates(map[string]any{
				"scope":       definition.Scope,
				"name":        definition.Name,
				"description": definition.Description,
				"is_system":   true,
			}).Error; updateErr != nil {
				return nil, updateErr
			}
		}

		result[definition.Code] = record
	}

	return result, nil
}

// 3. seedRoles upserts the default role catalog.
func seedRoles(tx *gorm.DB) (map[string]model.Role, error) {
	items := model.DefaultRoleDefinitions()
	result := make(map[string]model.Role, len(items))

	for _, definition := range items {
		record := model.Role{}
		if err := tx.Where("code = ?", definition.Code).Limit(1).Find(&record).Error; err != nil {
			return nil, err
		}
		if record.ID == 0 {
			record = model.Role{
				PublicID:    utils.NewPublicID(),
				Code:        definition.Code,
				Scope:       definition.Scope,
				Name:        definition.Name,
				Description: definition.Description,
				IsSystem:    true,
			}
			if createErr := tx.Create(&record).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			if updateErr := tx.Model(&record).Updates(map[string]any{
				"scope":       definition.Scope,
				"name":        definition.Name,
				"description": definition.Description,
				"is_system":   true,
			}).Error; updateErr != nil {
				return nil, updateErr
			}
		}

		result[definition.Code] = record
	}

	return result, nil
}

// 4. seedRolePermissions upserts the default role-permission matrix.
func seedRolePermissions(tx *gorm.DB, rolesByCode map[string]model.Role, permissionsByCode map[string]model.Permission) error {
	for roleCode, permissionCodes := range model.DefaultRolePermissionMatrix() {
		role, ok := rolesByCode[roleCode]
		if !ok {
			continue
		}

		permissionIDs := make([]int64, 0, len(permissionCodes))
		for _, permissionCode := range permissionCodes {
			permission, exists := permissionsByCode[permissionCode]
			if !exists {
				continue
			}
			permissionIDs = append(permissionIDs, permission.ID)
		}

		if len(permissionIDs) > 0 {
			if err := tx.
				Where("role_id = ? AND permission_id NOT IN ?", role.ID, permissionIDs).
				Delete(&model.RolePermission{}).
				Error; err != nil {
				return err
			}
		}

		for _, permissionCode := range permissionCodes {
			permission, exists := permissionsByCode[permissionCode]
			if !exists {
				continue
			}

			binding := model.RolePermission{}
			if err := tx.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).Limit(1).Find(&binding).Error; err != nil {
				return err
			}
			if binding.RoleID > 0 {
				continue
			}

			if createErr := tx.Create(&model.RolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
			}).Error; createErr != nil {
				return createErr
			}
		}
	}

	return nil
}

// 5. pruneDeprecatedRoles removes obsolete system role records and bindings.
func pruneDeprecatedRoles(tx *gorm.DB, rolesByCode map[string]model.Role) error {
	activeCodes := make([]string, 0, len(rolesByCode))
	for code := range rolesByCode {
		activeCodes = append(activeCodes, code)
	}

	var deprecatedRoles []model.Role
	if err := tx.Where("is_system = ? AND code NOT IN ?", true, activeCodes).Find(&deprecatedRoles).Error; err != nil {
		return err
	}
	if len(deprecatedRoles) == 0 {
		return nil
	}

	roleIDs := make([]int64, 0, len(deprecatedRoles))
	for _, role := range deprecatedRoles {
		roleIDs = append(roleIDs, role.ID)
	}

	if err := tx.Where("role_id IN ?", roleIDs).Delete(&model.UserRoleBinding{}).Error; err != nil {
		return err
	}
	if err := tx.Where("role_id IN ?", roleIDs).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}

	return tx.Where("id IN ?", roleIDs).Delete(&model.Role{}).Error
}
