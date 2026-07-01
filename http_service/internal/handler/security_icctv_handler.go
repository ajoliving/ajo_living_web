/*
 * iCCTV 視像監控 HTTP 介面。
 * 1. 讀取目前會員可見大廈的視像監控鏡頭。
 * 2. 綁定 AJO 登入身份與大廈可見範圍。
 * 3. 統一輸出 AJO API 回應格式。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. SecurityICCTVHandler handles iCCTV endpoints.
type SecurityICCTVHandler struct {
	securityICCTVService *service.SecurityICCTVService
}

// 2. NewSecurityICCTVHandler creates an iCCTV handler instance.
func NewSecurityICCTVHandler(securityICCTVService *service.SecurityICCTVService) *SecurityICCTVHandler {
	return &SecurityICCTVHandler{securityICCTVService: securityICCTVService}
}

// 3. GetPublicCameras returns visible iCCTV camera URLs for the current member.
func (h *SecurityICCTVHandler) GetPublicCameras(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login is required"))
		return
	}

	result, err := h.securityICCTVService.GetPublicCameras(c.Request.Context(), user.UserID, service.ICCTVBuildingParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}
