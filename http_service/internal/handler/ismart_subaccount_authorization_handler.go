/*
 * iSmart 單位級副戶授權 HTTP 介面。
 * 1. 主戶發起、列出、重發、撤銷副戶授權。
 * 2. 被綁定人以登入態或公開令牌完成授權確認。
 * 3. 請求參數綁定、身分校驗與統一回應輸出。
 */
package handler

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. IsmartSubaccountAuthorizationHandler handles unit-level subaccount authorization endpoints.
type IsmartSubaccountAuthorizationHandler struct {
	authorizationService *service.IsmartSubaccountAuthorizationService
}

// 2. NewIsmartSubaccountAuthorizationHandler creates the authorization handler instance.
func NewIsmartSubaccountAuthorizationHandler(authorizationService *service.IsmartSubaccountAuthorizationService) *IsmartSubaccountAuthorizationHandler {
	return &IsmartSubaccountAuthorizationHandler{authorizationService: authorizationService}
}

// 3. CreateAuthorization records one invitation or confirmation request for the current unit.
func (h *IsmartSubaccountAuthorizationHandler) CreateAuthorization(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request struct {
		UnitID    string `json:"unit_id"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		SendEmail *bool  `json:"send_email"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid subaccount authorization payload"))
		return
	}
	sendEmail := true
	if request.SendEmail != nil {
		sendEmail = *request.SendEmail
	}
	result, err := h.authorizationService.CreateInvitation(c.Request.Context(), user.UserID, service.IsmartSubaccountAuthorizationParams{
		UnitID:    request.UnitID,
		Phone:     request.Phone,
		Email:     request.Email,
		SendEmail: sendEmail,
		Remark:    request.Remark,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 4. ListAuthorizations returns non-terminal authorizations for one unit.
func (h *IsmartSubaccountAuthorizationHandler) ListAuthorizations(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	items, err := h.authorizationService.ListAuthorizations(c.Request.Context(), user.UserID, c.Query("unit_id"))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items, "count": len(items)})
}

// 5. ResendNotification resends the invitation or confirmation notice.
func (h *IsmartSubaccountAuthorizationHandler) ResendNotification(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request struct {
		SendEmail *bool `json:"send_email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid resend payload"))
		return
	}
	sendEmail := true
	if request.SendEmail != nil {
		sendEmail = *request.SendEmail
	}
	item, err := h.authorizationService.ResendNotification(c.Request.Context(), user.UserID, c.Param("authorizationId"), sendEmail)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, item)
}

// 6. ConfirmAuthorization confirms one authorization for the signed-in grantee.
func (h *IsmartSubaccountAuthorizationHandler) ConfirmAuthorization(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.authorizationService.ConfirmByID(c.Request.Context(), user.UserID, c.Param("authorizationId"))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 7. RevokeAuthorization cancels one pending authorization owned by the current member.
func (h *IsmartSubaccountAuthorizationHandler) RevokeAuthorization(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	if err := h.authorizationService.Revoke(c.Request.Context(), user.UserID, c.Param("authorizationId")); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"status": service.IsmartSubaccountAuthStatusRevoked})
}

// 8. ResolveAuthorizationToken resolves one public confirmation token.
func (h *IsmartSubaccountAuthorizationHandler) ResolveAuthorizationToken(c *gin.Context) {
	item, err := h.authorizationService.ResolveToken(c.Request.Context(), c.Param("token"))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, item)
}

// 9. ConfirmAuthorizationByToken confirms one authorization from the public email link.
func (h *IsmartSubaccountAuthorizationHandler) ConfirmAuthorizationByToken(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid confirmation payload"))
		return
	}
	result, err := h.authorizationService.ConfirmByToken(c.Request.Context(), request.Token)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}
