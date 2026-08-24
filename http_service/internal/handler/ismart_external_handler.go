/*
 * iSmart 對外介面代理 HTTP 介面。
 * 1. 接收 AJO 前端請求並使用目前登入態解析 iSmart 帳戶。
 * 2. 代理大廈資料、意見、門禁、二維碼與 POS payment to iSmart。
 * 3. 不接受前端提交舊系統 user_id 作為權限依據。
 * 4. 代理支付 integration 查詢介面並保留上游 raw response。
 */
package handler

import (
	"net/http"
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

// 3.0 ChangePassword changes the current member's linked iSmart password.
func (h *IsmartExternalHandler) ChangePassword(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request struct {
		OldPassword        string `json:"old_password"`
		NewPassword        string `json:"new_password"`
		NewPasswordConfirm string `json:"new_password_confirm"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid password payload"))
		return
	}
	result, err := h.ismartService.ChangeIsmartPassword(c.Request.Context(), user.UserID, service.IsmartPasswordChangeParams{
		OldPassword: request.OldPassword, NewPassword: request.NewPassword, NewPasswordConfirm: request.NewPasswordConfirm,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 3.0.1 GetNotificationSettings reads the current member's iSmart email preference.
func (h *IsmartExternalHandler) GetNotificationSettings(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.ismartService.GetIsmartNotificationSettings(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 3.0.2 UpdateNotificationSettings updates the current member's iSmart email preference.
func (h *IsmartExternalHandler) UpdateNotificationSettings(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request struct {
		ReceiveEmail *bool `json:"is_receive_email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.ReceiveEmail == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "is_receive_email is required"))
		return
	}
	result, err := h.ismartService.UpdateIsmartNotificationSettings(c.Request.Context(), user.UserID, *request.ReceiveEmail)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 3.1 UpdateClientProfile updates the current member's documented iSmart profile fields.
func (h *IsmartExternalHandler) UpdateClientProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		AccountEmail          *string `json:"account_email"`
		AccountPhone          *string `json:"account_phone"`
		ContactName           *string `json:"contact_name"`
		EmergencyContactName  *string `json:"emergency_contact_name"`
		OwnerNameEN           *string `json:"owner_name_en"`
		OwnerNameZH           *string `json:"owner_name_zh"`
		AccountName           *string `json:"account_name"`
		IdentityNumber        *string `json:"identity_number"`
		ContactPhone          *string `json:"contact_phone"`
		EmergencyContactPhone *string `json:"emergency_contact_phone"`
		ContactEmail          *string `json:"contact_email"`
		BirthDate             *string `json:"birth_date"`
		Gender                *string `json:"gender"`
		AddressEN             *string `json:"billing_address_en"`
		AddressZH             *string `json:"billing_address_zh"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid ismart profile payload"))
		return
	}

	result, err := h.ismartService.UpdateClientProfile(c.Request.Context(), user.UserID, service.IsmartClientProfileUpdate{
		AccountEmail:          request.AccountEmail,
		AccountPhone:          request.AccountPhone,
		ContactName:           request.ContactName,
		EmergencyContactName:  request.EmergencyContactName,
		OwnerNameEN:           request.OwnerNameEN,
		OwnerNameZH:           request.OwnerNameZH,
		AccountName:           request.AccountName,
		IdentityNumber:        request.IdentityNumber,
		ContactPhone:          request.ContactPhone,
		EmergencyContactPhone: request.EmergencyContactPhone,
		ContactEmail:          request.ContactEmail,
		BirthDate:             request.BirthDate,
		Gender:                request.Gender,
		AddressEN:             request.AddressEN,
		AddressZH:             request.AddressZH,
	})
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
		BuildingID   string `json:"building_id"`
		RequestType  string `json:"request_type"`
		Category     string `json:"category"`
		Subcategory  string `json:"subcategory"`
		Subject      string `json:"subject"`
		Content      string `json:"content"`
		LocationText string `json:"location_text"`
		UnitID       string `json:"unit_id"`
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		CommentType  string `json:"comment_type"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid building comment payload"))
		return
	}

	result, err := h.ismartService.SubmitBuildingComment(c.Request.Context(), user.UserID, service.IsmartBuildingCommentParams{
		BuildingID:   strings.TrimSpace(request.BuildingID),
		RequestType:  strings.TrimSpace(request.RequestType),
		Category:     strings.TrimSpace(request.Category),
		Subcategory:  strings.TrimSpace(request.Subcategory),
		Subject:      strings.TrimSpace(request.Subject),
		Content:      strings.TrimSpace(request.Content),
		LocationText: strings.TrimSpace(request.LocationText),
		UnitID:       strings.TrimSpace(request.UnitID),
		ContactName:  strings.TrimSpace(request.ContactName),
		ContactPhone: strings.TrimSpace(request.ContactPhone),
		CommentType:  strings.TrimSpace(request.CommentType),
		Comment:      strings.TrimSpace(request.Comment),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 5.1 ListBuildingServiceCases returns current member-visible service cases.
func (h *IsmartExternalHandler) ListBuildingServiceCases(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.ListBuildingServiceCases(c.Request.Context(), user.UserID, service.IsmartServiceCaseListParams{
		BuildingID:  strings.TrimSpace(c.Query("building_id")),
		Status:      strings.TrimSpace(c.Query("status")),
		RequestType: strings.TrimSpace(c.Query("request_type")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 5.2 GetBuildingServiceCase returns one service-case detail and message thread.
func (h *IsmartExternalHandler) GetBuildingServiceCase(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.ismartService.GetBuildingServiceCase(c.Request.Context(), user.UserID, service.IsmartServiceCaseDetailParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
		CaseID:     strings.TrimSpace(c.Param("caseId")),
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

// 10. ListPaymentUnpaidInvoices returns raw unpaid invoice rows.
func (h *IsmartExternalHandler) ListPaymentUnpaidInvoices(c *gin.Context) {
	result, err := h.ismartService.ListPaymentUnpaidInvoices(c.Request.Context(), service.IsmartPaymentUnpaidInvoiceParams{
		UnitID: strings.TrimSpace(c.Query("unit_id")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// 11. ListPaymentTransactionsByUnit returns raw unit payment history rows.
func (h *IsmartExternalHandler) ListPaymentTransactionsByUnit(c *gin.Context) {
	result, err := h.ismartService.ListPaymentTransactionsByUnit(c.Request.Context(), service.IsmartPaymentTransactionsByUnitParams{
		UnitIDList: c.QueryArray("unit_id_list"),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// 12. ListPaymentTransactionsByDate returns raw date payment history rows.
func (h *IsmartExternalHandler) ListPaymentTransactionsByDate(c *gin.Context) {
	result, err := h.ismartService.ListPaymentTransactionsByDate(c.Request.Context(), service.IsmartPaymentTransactionsByDateParams{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
		FromDate:   strings.TrimSpace(c.Query("from_date")),
		ToDate:     strings.TrimSpace(c.Query("to_date")),
		DateType:   strings.TrimSpace(c.Query("date_type")),
		PayMethod:  strings.TrimSpace(c.Query("pay_method")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// 13. RegisterAccount proxies one iSmart account registration request.
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

// 14. ListManagementFees returns visible building management-fee receivables.
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

// 15. ListOtherFees returns visible building other-fee receivables.
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

// 16. ListBuildingNotices returns visible building notices.
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

// 17. SubmitOwnerBindingRequest submits an owner-role binding request.
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

// 18. ListSubaccounts returns current owner-controlled authorized users.
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

// 19. GrantSubaccount grants one iSmart authorized user.
func (h *IsmartExternalHandler) GrantSubaccount(c *gin.Context) {
	h.handleSubaccountMutation(c, "grant")
}

// 20. RevokeSubaccount revokes one iSmart authorized user.
func (h *IsmartExternalHandler) RevokeSubaccount(c *gin.Context) {
	h.handleSubaccountMutation(c, "revoke")
}

// 21. handleSubaccountMutation handles grant and revoke requests.
func (h *IsmartExternalHandler) handleSubaccountMutation(c *gin.Context, action string) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		UnitID       string `json:"unit_id"`
		TargetUserID int64  `json:"target_user_id"`
		TargetPhone  string `json:"target_phone"`
		TargetEmail  string `json:"target_email"`
		Remark       string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid subaccount payload"))
		return
	}

	params := service.IsmartSubaccountMutationParams{
		UnitID:       strings.TrimSpace(request.UnitID),
		TargetUserID: request.TargetUserID,
		TargetPhone:  strings.TrimSpace(request.TargetPhone),
		TargetEmail:  strings.TrimSpace(request.TargetEmail),
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
