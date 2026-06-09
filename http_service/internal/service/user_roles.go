/*
 * Shared user role helpers.
 * 1. Keep member type and staff role normalization consistent.
 * 2. Expose small helpers for auth, profile, and staff services.
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
