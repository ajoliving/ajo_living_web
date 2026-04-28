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
	// 2. MemberTypeProUser is the elevated front-end member type.
	MemberTypeProUser = "pro_user"
	// 3. RoleStaff is the back-office staff role.
	RoleStaff = "staff"
)

// 4. normalizeMemberType returns a stable supported member type.
func normalizeMemberType(memberType string) string {
	switch strings.TrimSpace(memberType) {
	case MemberTypeProUser:
		return MemberTypeProUser
	default:
		return MemberTypeUser
	}
}

// 5. resolveUserRole derives the effective role for a user account.
func resolveUserRole(memberType string, isStaff bool) string {
	if isStaff {
		return RoleStaff
	}

	return normalizeMemberType(memberType)
}

// 6. isValidMemberType reports whether the input is a supported member type.
func isValidMemberType(memberType string) bool {
	switch strings.TrimSpace(memberType) {
	case MemberTypeUser, MemberTypeProUser:
		return true
	default:
		return false
	}
}
