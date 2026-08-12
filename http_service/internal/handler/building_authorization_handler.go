/*
 * 大廈副戶授權 HTTP 介面。
 * 1. 提供主戶按大廈管理副戶授權的入口。
 * 2. 提供受邀帳戶首次設定用戶 ID 及密碼的入口。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. BuildingAuthorizationHandler handles building authorization requests.
type BuildingAuthorizationHandler struct {
	authorizationService *service.BuildingAuthorizationService
}

// 2. NewBuildingAuthorizationHandler creates a building authorization handler.
func NewBuildingAuthorizationHandler(authorizationService *service.BuildingAuthorizationService) *BuildingAuthorizationHandler {
	return &BuildingAuthorizationHandler{authorizationService: authorizationService}
}

// 3. AvailablePermissions returns the selectable building menu functions.
func (h *BuildingAuthorizationHandler) AvailablePermissions(c *gin.Context) {
	errcode.Success(c, map[string]any{"items": h.authorizationService.AvailablePermissions()})
}

// 4. List returns owner-created authorizations for the requested building.
func (h *BuildingAuthorizationHandler) List(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.authorizationService.List(c.Request.Context(), user.UserID, strings.TrimSpace(c.Query("building_id")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, map[string]any{"items": result})
}

// 5. Create grants selected functions to one phone and email identity.
func (h *BuildingAuthorizationHandler) Create(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request struct {
		BuildingID       string   `json:"building_id"`
		PhoneCountryCode string   `json:"phone_country_code"`
		PhoneNumber      string   `json:"phone_number"`
		Email            string   `json:"email"`
		Permissions      []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid building authorization payload"))
		return
	}
	result, err := h.authorizationService.Create(c.Request.Context(), user.UserID, service.BuildingAuthorizationCreateParams{
		BuildingID: strings.TrimSpace(request.BuildingID), PhoneCountryCode: strings.TrimSpace(request.PhoneCountryCode),
		PhoneNumber: strings.TrimSpace(request.PhoneNumber), Email: strings.TrimSpace(request.Email), Permissions: request.Permissions,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 6. Revoke removes one owner-created authorization.
func (h *BuildingAuthorizationHandler) Revoke(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	if err := h.authorizationService.Revoke(c.Request.Context(), user.UserID, c.Param("authorizationId")); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, map[string]any{"revoked": true})
}

// 7. Accept activates a pending invitation with the first local credentials.
func (h *BuildingAuthorizationHandler) Accept(c *gin.Context) {
	var request struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid building invitation payload"))
		return
	}
	if err := h.authorizationService.Accept(c.Request.Context(), service.BuildingAuthorizationAcceptParams{Token: request.Token, Username: request.Username, Password: request.Password}); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, map[string]any{"activated": true})
}
