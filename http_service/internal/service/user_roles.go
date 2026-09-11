/*
 * Shared user role helpers.
 * 1. Keep member type and staff role normalization consistent.
 * 2. Resolve compact staff or member access snapshots from is_staff.
 */
package service

import "strings"

const (
	// 1. MemberTypeUser is the default front-end member type.
	MemberTypeUser = "user"
	// 2. RoleStaff is the back-office staff role.
	RoleStaff = "staff"
)

// 3. normalizeMemberType returns the single ordinary user member type.
func normalizeMemberType(memberType string) string {
	return MemberTypeUser
}

// 4. resolveUserRole derives the effective role for a user account.
func resolveUserRole(memberType string, isStaff bool) string {
	if isStaff {
		return RoleStaff
	}

	return normalizeMemberType(memberType)
}

// 5. isValidMemberType reports whether the input is the ordinary user type.
func isValidMemberType(memberType string) bool {
	return strings.TrimSpace(memberType) == MemberTypeUser
}

// 1. AccessSnapshot defines resolved access data for one user.
type AccessSnapshot struct {
	RoleCodes   []string
	Permissions []string
	IsStaff     bool
}

// 2. buildAccessSnapshot returns the compact user or staff access view.
func buildAccessSnapshot(isStaff bool) *AccessSnapshot {
	if isStaff {
		return &AccessSnapshot{
			RoleCodes:   []string{RoleStaff},
			Permissions: []string{},
			IsStaff:     true,
		}
	}

	return &AccessSnapshot{
		RoleCodes:   []string{MemberTypeUser},
		Permissions: []string{},
		IsStaff:     false,
	}
}
