/*
 * Staff HTTP handlers.
 * 1. Bind staff-only user query, account creation, and staff flag update requests.
 * 2. Delegate staff business rules to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. StaffHandler handles staff-only endpoints.
type StaffHandler struct {
	staffService *service.StaffService
}

// 2. updateStaffUserRoleRequest defines the staff role update payload.
type updateStaffUserRoleRequest struct {
	MemberType string   `json:"member_type"`
	RoleCodes  []string `json:"role_codes"`
	IsStaff    *bool    `json:"is_staff"`
}

// 3. createStaffUserRequest defines the staff account creation payload.
type createStaffUserRequest struct {
	Email                 string   `json:"email"`
	Password              string   `json:"password"`
	DisplayName           string   `json:"display_name"`
	PhoneCountryCode      string   `json:"phone_country_code"`
	PhoneNumber           string   `json:"phone_number"`
	PublisherIdentityType string   `json:"publisher_identity_type"`
	PrimaryCommunityID    string   `json:"primary_community_id"`
	PrimaryCommunityName  string   `json:"primary_community_name"`
	ResidenceFloor        string   `json:"residence_floor"`
	ResidenceUnit         string   `json:"residence_unit"`
	DistrictCode          string   `json:"district_code"`
	MemberType            string   `json:"member_type"`
	RoleCodes             []string `json:"role_codes"`
	IsStaff               *bool    `json:"is_staff"`
}

// 4. NewStaffHandler creates a staff handler instance.
func NewStaffHandler(staffService *service.StaffService) *StaffHandler {
	return &StaffHandler{staffService: staffService}
}

// 5. GetMe returns the current staff account summary.
func (h *StaffHandler) GetMe(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.staffService.GetStaffMe(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. ListUsers returns staff-visible user records.
func (h *StaffHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePagination(c)
	result, pagination, err := h.staffService.ListUsers(c.Request.Context(), service.StaffUserListFilters{
		Page:       page,
		PageSize:   pageSize,
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Status:     strings.TrimSpace(c.Query("status")),
		MemberType: strings.TrimSpace(c.Query("member_type")),
		RoleCode:   strings.TrimSpace(c.Query("role_code")),
		IsStaff:    parseOptionalBoolQuery(c, "is_staff"),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result, "pagination": pagination})
}

// 7. ListRoles returns staff role catalog items.
func (h *StaffHandler) ListRoles(c *gin.Context) {
	result, err := h.staffService.ListRoles(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 8. CreateUser creates a staff-managed account.
func (h *StaffHandler) CreateUser(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request createStaffUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.staffService.CreateUser(c.Request.Context(), user.UserID, service.StaffUserCreateParams{
		Email:                 strings.TrimSpace(request.Email),
		Password:              request.Password,
		DisplayName:           strings.TrimSpace(request.DisplayName),
		PhoneCountryCode:      strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber:           strings.TrimSpace(request.PhoneNumber),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		PrimaryCommunityID:    strings.TrimSpace(request.PrimaryCommunityID),
		PrimaryCommunityName:  strings.TrimSpace(request.PrimaryCommunityName),
		ResidenceFloor:        strings.TrimSpace(request.ResidenceFloor),
		ResidenceUnit:         strings.TrimSpace(request.ResidenceUnit),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		MemberType:            strings.TrimSpace(request.MemberType),
		RoleCodes:             trimStringSlice(request.RoleCodes),
		IsStaff:               request.IsStaff,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 9. UpdateUserRole updates the target user's role bindings.
func (h *StaffHandler) UpdateUserRole(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request updateStaffUserRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.staffService.UpdateUserRole(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("userId")), service.StaffUserRoleUpdateParams{
		MemberType: strings.TrimSpace(request.MemberType),
		RoleCodes:  trimStringSlice(request.RoleCodes),
		IsStaff:    request.IsStaff,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 10. trimStringSlice normalizes JSON string arrays.
func trimStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
