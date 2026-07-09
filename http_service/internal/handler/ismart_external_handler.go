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

// 10. RegisterAccount proxies one iSmart account registration request.
func (h *IsmartExternalHandler) RegisterAccount(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid ismart registration payload"))
		return
	}

	result, err := h.ismartService.RegisterAccount(c.Request.Context(), user.UserID, request)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 11. ListManagementFees returns visible building management-fee receivables.
func (h *IsmartExternalHandler) ListManagementFees(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListManagementFees(c.Request.Context(), user.UserID, service.IsmartReceivableParams{
		BuildingID:    strings.TrimSpace(c.Query("building_id")),
		DataStructure: strings.TrimSpace(c.Query("data_structure")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 12. ListOtherFees returns visible building other-fee receivables.
func (h *IsmartExternalHandler) ListOtherFees(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListOtherFees(c.Request.Context(), user.UserID, service.IsmartReceivableParams{
		BuildingID:    strings.TrimSpace(c.Query("building_id")),
		DataStructure: strings.TrimSpace(c.Query("data_structure")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 13. ListBuildingNotices returns visible building notices.
func (h *IsmartExternalHandler) ListBuildingNotices(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListBuildingNotices(c.Request.Context(), user.UserID, service.IsmartBuildingParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 14. SubmitOwnerBindingRequest submits an owner-role binding request.
func (h *IsmartExternalHandler) SubmitOwnerBindingRequest(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID        string   `json:"building_id"`
		OwnedFlat         []string `json:"ownedflat"`
		Role              string   `json:"cli_role"`
		OwnerNote         string   `json:"ownernote"`
		IsReceiveEmail    *bool    `json:"is_receive_email"`
		RegistrationTel   string   `json:"reg_tel"`
		RegistrationEmail string   `json:"reg_email"`
		ClientName        string   `json:"cli_name"`
		ClientIDCard      string   `json:"cli_id_card"`
		ClientTel         string   `json:"cli_tel"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid owner binding payload"))
		return
	}

	result, err := h.ismartService.SubmitOwnerBindingRequest(c.Request.Context(), user.UserID, service.IsmartOwnerBindingParams{
		BuildingID:        strings.TrimSpace(request.BuildingID),
		OwnedFlat:         request.OwnedFlat,
		Role:              strings.TrimSpace(request.Role),
		OwnerNote:         strings.TrimSpace(request.OwnerNote),
		IsReceiveEmail:    request.IsReceiveEmail,
		RegistrationTel:   strings.TrimSpace(request.RegistrationTel),
		RegistrationEmail: strings.TrimSpace(request.RegistrationEmail),
		ClientName:        strings.TrimSpace(request.ClientName),
		ClientIDCard:      strings.TrimSpace(request.ClientIDCard),
		ClientTel:         strings.TrimSpace(request.ClientTel),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 15. ListSubaccounts returns current owner-controlled authorized users.
func (h *IsmartExternalHandler) ListSubaccounts(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListSubaccounts(c.Request.Context(), user.UserID, service.IsmartSubaccountQuery{
		UnitID: strings.TrimSpace(c.Query("unit_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 16. GrantSubaccount grants one iSmart authorized user.
func (h *IsmartExternalHandler) GrantSubaccount(c *gin.Context) {
	h.handleSubaccountMutation(c, "grant")
}

// 17. RevokeSubaccount revokes one iSmart authorized user.
func (h *IsmartExternalHandler) RevokeSubaccount(c *gin.Context) {
	h.handleSubaccountMutation(c, "revoke")
}

// 18. handleSubaccountMutation handles grant and revoke requests.
func (h *IsmartExternalHandler) handleSubaccountMutation(c *gin.Context, action string) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		UnitID       string `json:"unit_id"`
		TargetUserID int64  `json:"target_user_id"`
		Remark       string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid subaccount payload"))
		return
	}

	params := service.IsmartSubaccountMutationParams{
		UnitID:       strings.TrimSpace(request.UnitID),
		TargetUserID: request.TargetUserID,
		Remark:       strings.TrimSpace(request.Remark),
	}
	var (
		result map[string]any
		err    error
	)
	if action == "grant" {
		result, err = h.ismartService.GrantSubaccount(c.Request.Context(), user.UserID, params)
	} else {
		result, err = h.ismartService.RevokeSubaccount(c.Request.Context(), user.UserID, params)
	}
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}
