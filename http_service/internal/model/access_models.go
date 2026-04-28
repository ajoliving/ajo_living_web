/*
 * Access control data models.
 * 1. Define RBAC roles, permissions, and user role bindings.
 * 2. Provide the default system catalog for role and permission seeding.
 */
package model

import "time"

const (
	// 1. RoleScopeMember marks front-end member roles.
	RoleScopeMember = "member"
	// 2. RoleScopeStaff marks back-office staff roles.
	RoleScopeStaff = "staff"
)

const (
	// 3. RoleCodeMemberBasic grants baseline member access.
	RoleCodeMemberBasic = "member_basic"
	// 4. RoleCodeMemberVerified grants verified-member access.
	RoleCodeMemberVerified = "member_verified"
	// 5. RoleCodeMemberPro grants pro member access.
	RoleCodeMemberPro = "member_pro"
	// 6. RoleCodeSuperAdmin grants full staff access.
	RoleCodeSuperAdmin = "super_admin"
	// 7. RoleCodeOpsAdmin grants day-to-day operations access.
	RoleCodeOpsAdmin = "ops_admin"
	// 8. RoleCodeContentModerator grants review access.
	RoleCodeContentModerator = "content_moderator"
	// 9. RoleCodeCustomerService grants support access.
	RoleCodeCustomerService = "customer_service"
)

const (
	// 10. PermissionCodeAccountProfileRead grants profile read access.
	PermissionCodeAccountProfileRead = "account.profile.read"
	// 11. PermissionCodeAccountProfileWrite grants profile write access.
	PermissionCodeAccountProfileWrite = "account.profile.write"
	// 12. PermissionCodeListingOwnManage grants self listing management.
	PermissionCodeListingOwnManage = "listing.own.manage"
	// 13. PermissionCodeChatUse grants in-app chat access.
	PermissionCodeChatUse = "chat.use"
	// 14. PermissionCodeOrderCreate grants order creation access.
	PermissionCodeOrderCreate = "order.create"
	// 15. PermissionCodeOrderOwnManage grants self order management.
	PermissionCodeOrderOwnManage = "order.own.manage"
	// 16. PermissionCodeNotificationRead grants notification access.
	PermissionCodeNotificationRead = "notification.read"
	// 17. PermissionCodeStaffConsoleAccess grants staff console access.
	PermissionCodeStaffConsoleAccess = "staff.console.access"
	// 18. PermissionCodeStaffUserRead grants staff user read access.
	PermissionCodeStaffUserRead = "staff.user.read"
	// 19. PermissionCodeStaffUserManage grants staff user management access.
	PermissionCodeStaffUserManage = "staff.user.manage"
	// 20. PermissionCodeStaffRoleRead grants staff role read access.
	PermissionCodeStaffRoleRead = "staff.role.read"
	// 21. PermissionCodeStaffRoleManage grants staff role management access.
	PermissionCodeStaffRoleManage = "staff.role.manage"
	// 22. PermissionCodeStaffReviewManage grants moderation access.
	PermissionCodeStaffReviewManage = "staff.review.manage"
	// 23. PermissionCodeStaffSupportManage grants customer support access.
	PermissionCodeStaffSupportManage = "staff.support.manage"
)

// 24. Role stores a reusable RBAC role definition.
type Role struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID    string `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	Code        string `gorm:"type:varchar(64);not null;uniqueIndex" json:"code"`
	Scope       string `gorm:"type:varchar(32);not null;index" json:"scope"`
	Name        string `gorm:"type:varchar(120);not null" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	IsSystem    bool   `gorm:"not null;default:true;index" json:"is_system"`
	TimestampModel
}

// 25. Permission stores a reusable RBAC permission definition.
type Permission struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID    string `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	Code        string `gorm:"type:varchar(96);not null;uniqueIndex" json:"code"`
	Scope       string `gorm:"type:varchar(32);not null;index" json:"scope"`
	Name        string `gorm:"type:varchar(160);not null" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	IsSystem    bool   `gorm:"not null;default:true;index" json:"is_system"`
	TimestampModel
}

// 26. RolePermission stores the role-to-permission binding table.
type RolePermission struct {
	RoleID       int64     `gorm:"primaryKey" json:"role_id"`
	PermissionID int64     `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// 27. UserRoleBinding stores the user-to-role binding table.
type UserRoleBinding struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;uniqueIndex:uk_user_role_bindings_user_role,priority:1;index" json:"user_id"`
	RoleID     int64     `gorm:"not null;uniqueIndex:uk_user_role_bindings_user_role,priority:2;index" json:"role_id"`
	AssignedBy *int64    `gorm:"index" json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
}

// 28. SystemRoleDefinition defines one seedable role catalog item.
type SystemRoleDefinition struct {
	Code        string
	Scope       string
	Name        string
	Description string
}

// 29. SystemPermissionDefinition defines one seedable permission catalog item.
type SystemPermissionDefinition struct {
	Code        string
	Scope       string
	Name        string
	Description string
}

// 30. DefaultRoleDefinitions returns the baseline role catalog.
func DefaultRoleDefinitions() []SystemRoleDefinition {
	return []SystemRoleDefinition{
		{Code: RoleCodeMemberBasic, Scope: RoleScopeMember, Name: "Member Basic", Description: "Baseline member access."},
		{Code: RoleCodeMemberVerified, Scope: RoleScopeMember, Name: "Member Verified", Description: "Verified member access."},
		{Code: RoleCodeMemberPro, Scope: RoleScopeMember, Name: "Member Pro", Description: "Upgraded member access."},
		{Code: RoleCodeSuperAdmin, Scope: RoleScopeStaff, Name: "Super Admin", Description: "Full back-office access."},
		{Code: RoleCodeOpsAdmin, Scope: RoleScopeStaff, Name: "Operations Admin", Description: "Daily operation access."},
		{Code: RoleCodeContentModerator, Scope: RoleScopeStaff, Name: "Content Moderator", Description: "Review and moderation access."},
		{Code: RoleCodeCustomerService, Scope: RoleScopeStaff, Name: "Customer Service", Description: "Support and order follow-up access."},
	}
}

// 31. DefaultPermissionDefinitions returns the baseline permission catalog.
func DefaultPermissionDefinitions() []SystemPermissionDefinition {
	return []SystemPermissionDefinition{
		{Code: PermissionCodeAccountProfileRead, Scope: RoleScopeMember, Name: "Account Profile Read", Description: "Read current member profile."},
		{Code: PermissionCodeAccountProfileWrite, Scope: RoleScopeMember, Name: "Account Profile Write", Description: "Update current member profile."},
		{Code: PermissionCodeListingOwnManage, Scope: RoleScopeMember, Name: "Listing Own Manage", Description: "Manage owned secondhand listings."},
		{Code: PermissionCodeChatUse, Scope: RoleScopeMember, Name: "Chat Use", Description: "Use in-app listing chat."},
		{Code: PermissionCodeOrderCreate, Scope: RoleScopeMember, Name: "Order Create", Description: "Create an order from a listing."},
		{Code: PermissionCodeOrderOwnManage, Scope: RoleScopeMember, Name: "Order Own Manage", Description: "Manage owned buyer or seller orders."},
		{Code: PermissionCodeNotificationRead, Scope: RoleScopeMember, Name: "Notification Read", Description: "Read in-app notifications."},
		{Code: PermissionCodeStaffConsoleAccess, Scope: RoleScopeStaff, Name: "Staff Console Access", Description: "Access staff-only endpoints."},
		{Code: PermissionCodeStaffUserRead, Scope: RoleScopeStaff, Name: "Staff User Read", Description: "Read staff user list and summaries."},
		{Code: PermissionCodeStaffUserManage, Scope: RoleScopeStaff, Name: "Staff User Manage", Description: "Manage member and staff access."},
		{Code: PermissionCodeStaffRoleRead, Scope: RoleScopeStaff, Name: "Staff Role Read", Description: "Read available roles and permission matrices."},
		{Code: PermissionCodeStaffRoleManage, Scope: RoleScopeStaff, Name: "Staff Role Manage", Description: "Manage user role bindings."},
		{Code: PermissionCodeStaffReviewManage, Scope: RoleScopeStaff, Name: "Staff Review Manage", Description: "Handle moderation and review operations."},
		{Code: PermissionCodeStaffSupportManage, Scope: RoleScopeStaff, Name: "Staff Support Manage", Description: "Handle customer support and disputes."},
	}
}

// 32. DefaultRolePermissionMatrix returns the baseline role-permission matrix.
func DefaultRolePermissionMatrix() map[string][]string {
	return map[string][]string{
		RoleCodeMemberBasic: {
			PermissionCodeAccountProfileRead,
			PermissionCodeAccountProfileWrite,
			PermissionCodeListingOwnManage,
			PermissionCodeChatUse,
			PermissionCodeOrderCreate,
			PermissionCodeOrderOwnManage,
			PermissionCodeNotificationRead,
		},
		RoleCodeMemberVerified: {
			PermissionCodeAccountProfileRead,
			PermissionCodeAccountProfileWrite,
			PermissionCodeListingOwnManage,
			PermissionCodeChatUse,
			PermissionCodeOrderCreate,
			PermissionCodeOrderOwnManage,
			PermissionCodeNotificationRead,
		},
		RoleCodeMemberPro: {
			PermissionCodeAccountProfileRead,
			PermissionCodeAccountProfileWrite,
			PermissionCodeListingOwnManage,
			PermissionCodeChatUse,
			PermissionCodeOrderCreate,
			PermissionCodeOrderOwnManage,
			PermissionCodeNotificationRead,
		},
		RoleCodeSuperAdmin: {
			PermissionCodeStaffConsoleAccess,
			PermissionCodeStaffUserRead,
			PermissionCodeStaffUserManage,
			PermissionCodeStaffRoleRead,
			PermissionCodeStaffRoleManage,
			PermissionCodeStaffReviewManage,
			PermissionCodeStaffSupportManage,
		},
		RoleCodeOpsAdmin: {
			PermissionCodeStaffConsoleAccess,
			PermissionCodeStaffUserRead,
			PermissionCodeStaffUserManage,
			PermissionCodeStaffRoleRead,
		},
		RoleCodeContentModerator: {
			PermissionCodeStaffConsoleAccess,
			PermissionCodeStaffReviewManage,
			PermissionCodeStaffUserRead,
		},
		RoleCodeCustomerService: {
			PermissionCodeStaffConsoleAccess,
			PermissionCodeStaffSupportManage,
			PermissionCodeStaffUserRead,
		},
	}
}
