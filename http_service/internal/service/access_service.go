/*
 * Access control compatibility helpers.
 * 1. Resolve account access from the user is_staff flag.
 * 2. Keep legacy role and permission response fields stable.
 * 3. Avoid RBAC role bindings in runtime permission decisions.
 */
package service

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
