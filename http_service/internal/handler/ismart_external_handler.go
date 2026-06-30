/*
 * iSmart 對外介面代理 HTTP 介面。
 * 1. 接收 AJO 前端請求並使用目前登入態解析 iSmart 帳戶。
 * 2. 代理大廈資料、意見、門禁、二維碼與 POS payment to iSmart。
 * 3. 不接受前端提交舊系統 user_id 作為權限依據。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. IsmartExternalHandler handles iSmart external proxy endpoints.
type IsmartExternalHandler struct {
	ismartService *service.IsmartExternalService
}

// 2. NewIsmartExternalHandler creates an iSmart external handler instance.
func NewIsmartExternalHandler(ismartService *service.IsmartExternalService) *IsmartExternalHandler {
	return &IsmartExternalHandler{ismartService: ismartService}
}

// 3. ListBuildings returns current member iSmart building options.
func (h *IsmartExternalHandler) ListBuildings(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListBuildings(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 4. GetBuildingInfo returns the selected iSmart building profile.
func (h *IsmartExternalHandler) GetBuildingInfo(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.GetBuildingInfo(c.Request.Context(), user.UserID, service.IsmartBuildingParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 5. SubmitBuildingComment submits one building comment to iSmart.
func (h *IsmartExternalHandler) SubmitBuildingComment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID  string `json:"building_id"`
		CommentType string `json:"comment_type"`
		Comment     string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid building comment payload"))
		return
	}

	result, err := h.ismartService.SubmitBuildingComment(c.Request.Context(), user.UserID, service.IsmartBuildingCommentParams{
		BuildingID:  strings.TrimSpace(request.BuildingID),
		CommentType: strings.TrimSpace(request.CommentType),
		Comment:     strings.TrimSpace(request.Comment),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 6. GetBuildingAccess returns iSmart building access summary.
func (h *IsmartExternalHandler) GetBuildingAccess(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.GetBuildingAccess(c.Request.Context(), user.UserID, service.IsmartBuildingParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 7. OpenDoor opens one visible iSmart door.
func (h *IsmartExternalHandler) OpenDoor(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID string `json:"building_id"`
		DoorID     int64  `json:"door_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid door open payload"))
		return
	}

	result, err := h.ismartService.OpenDoor(c.Request.Context(), user.UserID, service.IsmartDoorOpenParams{
		BuildingID: strings.TrimSpace(request.BuildingID),
		DoorID:     request.DoorID,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. GenerateQRCode generates one iSmart door QR payload.
func (h *IsmartExternalHandler) GenerateQRCode(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID     string `json:"building_id"`
		QRCodeRecordID int64  `json:"qrcode_record_id"`
		Term           string `json:"term"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid qrcode payload"))
		return
	}

	result, err := h.ismartService.GenerateQRCode(c.Request.Context(), user.UserID, service.IsmartQRCodeParams{
		BuildingID:     strings.TrimSpace(request.BuildingID),
		QRCodeRecordID: request.QRCodeRecordID,
		Term:           strings.TrimSpace(request.Term),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. SubmitPOSPayment submits one POS payment to iSmart.
func (h *IsmartExternalHandler) SubmitPOSPayment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid pos payment payload"))
		return
	}

	result, err := h.ismartService.SubmitPOSPayment(c.Request.Context(), user.UserID, request)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}
