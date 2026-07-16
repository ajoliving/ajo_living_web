/*
 * Agency profile HTTP handlers.
 * 1. Bind individual and company profile revisions.
 * 2. Bind staff review and company subaccount operations.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. AgencyCompanyHandler handles unified agency profile endpoints.
type AgencyCompanyHandler struct{ agencyProfileService *service.AgencyProfileService }

// 2. agencyProfileRequest defines member profile fields.
type agencyProfileRequest struct {
	ProfileType                 string `json:"profile_type"`
	NameZH                      string `json:"name_zh"`
	NameEN                      string `json:"name_en"`
	AddressZH                   string `json:"address_zh"`
	AddressEN                   string `json:"address_en"`
	LicenseNumber               string `json:"license_number"`
	IsOverseas                  bool   `json:"is_overseas"`
	IsBigFour                   bool   `json:"is_big_four"`
	Phone1CountryCode           string `json:"phone_1_country_code"`
	Phone1Number                string `json:"phone_1_number"`
	Phone1WhatsApp              bool   `json:"phone_1_whatsapp"`
	Phone2CountryCode           string `json:"phone_2_country_code"`
	Phone2Number                string `json:"phone_2_number"`
	Phone2WhatsApp              bool   `json:"phone_2_whatsapp"`
	WechatID                    string `json:"wechat_id"`
	WechatURL                   string `json:"wechat_url"`
	SignatureZH                 string `json:"signature_zh"`
	SignatureEN                 string `json:"signature_en"`
	DefaultAvatar               string `json:"default_avatar"`
	AvatarAssetID               string `json:"avatar_asset_id"`
	WechatQRAssetID             string `json:"wechat_qr_asset_id"`
	LogoAssetID                 string `json:"logo_asset_id"`
	EAALicenseAssetID           string `json:"eaa_license_asset_id"`
	BusinessRegistrationAssetID string `json:"business_registration_asset_id"`
	CompanyCardAssetID          string `json:"company_card_asset_id"`
}

// 3. agencyProfileReviewRequest defines a staff decision.
type agencyProfileReviewRequest struct {
	Approved   bool   `json:"approved"`
	ReviewNote string `json:"review_note"`
}

// 4. agencySubaccountRequest defines a company child login.
type agencySubaccountRequest struct {
	DisplayName      string   `json:"display_name"`
	PhoneCountryCode string   `json:"phone_country_code"`
	PhoneNumber      string   `json:"phone_number"`
	Email            string   `json:"email"`
	Password         string   `json:"password"`
	Permissions      []string `json:"permissions"`
	Status           string   `json:"status"`
}

// 4. NewAgencyCompanyHandler creates the unified agency handler.
func NewAgencyCompanyHandler(profileService *service.AgencyProfileService) *AgencyCompanyHandler {
	return &AgencyCompanyHandler{agencyProfileService: profileService}
}

// 5. GetMemberAgencyProfile returns active and working versions.
func (h *AgencyCompanyHandler) GetMemberAgencyProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.agencyProfileService.GetMemberAgencyProfile(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 6. CreateMemberAgencyProfile creates a draft.
func (h *AgencyCompanyHandler) CreateMemberAgencyProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	params, ok := bindAgencyProfileRequest(c)
	if !ok {
		return
	}
	result, err := h.agencyProfileService.CreateMemberAgencyProfile(c.Request.Context(), user.UserID, params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 7. UpdateMemberAgencyProfile updates a draft or rejected revision.
func (h *AgencyCompanyHandler) UpdateMemberAgencyProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	params, ok := bindAgencyProfileRequest(c)
	if !ok {
		return
	}
	result, err := h.agencyProfileService.UpdateMemberAgencyProfile(c.Request.Context(), user.UserID, params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. SubmitMemberAgencyProfile submits the revision.
func (h *AgencyCompanyHandler) SubmitMemberAgencyProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.agencyProfileService.SubmitMemberAgencyProfile(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. ListAgencyProfilesForStaff returns filtered review rows.
func (h *AgencyCompanyHandler) ListAgencyProfilesForStaff(c *gin.Context) {
	page, pageSize := parsePagination(c)
	items, pagination, err := h.agencyProfileService.ListAgencyProfilesForStaff(c.Request.Context(), service.AgencyProfileListFilters{Page: page, PageSize: pageSize, Status: strings.TrimSpace(c.DefaultQuery("status", service.AgencyProfileStatusPending)), ProfileType: strings.TrimSpace(c.Query("profile_type")), Keyword: strings.TrimSpace(c.Query("keyword"))})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 10. GetAgencyProfileForStaff returns one review detail.
func (h *AgencyCompanyHandler) GetAgencyProfileForStaff(c *gin.Context) {
	result, err := h.agencyProfileService.GetAgencyProfileForStaff(c.Request.Context(), strings.TrimSpace(c.Param("profileId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 11. ReviewAgencyProfile approves or rejects a revision.
func (h *AgencyCompanyHandler) ReviewAgencyProfile(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request agencyProfileReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.agencyProfileService.ReviewAgencyProfile(c.Request.Context(), strings.TrimSpace(c.Param("profileId")), service.AgencyProfileReviewParams{ReviewerUserID: user.UserID, Approved: request.Approved, ReviewNote: strings.TrimSpace(request.ReviewNote)})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 12. ListSubaccounts returns company child accounts.
func (h *AgencyCompanyHandler) ListSubaccounts(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	items, err := h.agencyProfileService.ListSubaccounts(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 13. CreateSubaccount creates a company child login.
func (h *AgencyCompanyHandler) CreateSubaccount(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var r agencySubaccountRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.agencyProfileService.CreateSubaccount(c.Request.Context(), user.UserID, service.AgencySubaccountCreateParams{DisplayName: strings.TrimSpace(r.DisplayName), PhoneCountryCode: strings.TrimSpace(r.PhoneCountryCode), PhoneNumber: strings.TrimSpace(r.PhoneNumber), Email: strings.TrimSpace(r.Email), Password: r.Password, Permissions: r.Permissions})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 14. UpdateSubaccountStatus changes one child status.
func (h *AgencyCompanyHandler) UpdateSubaccountStatus(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var r agencySubaccountRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.agencyProfileService.UpdateSubaccountStatus(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("subaccountId")), strings.TrimSpace(r.Status))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 15. DeleteSubaccount removes one child link and disables its login.
func (h *AgencyCompanyHandler) DeleteSubaccount(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	if err := h.agencyProfileService.DeleteSubaccount(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("subaccountId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"deleted": true})
}

// 16. bindAgencyProfileRequest maps an HTTP payload into service input.
func bindAgencyProfileRequest(c *gin.Context) (service.AgencyProfileUpsertParams, bool) {
	var r agencyProfileRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return service.AgencyProfileUpsertParams{}, false
	}
	return service.AgencyProfileUpsertParams{ProfileType: strings.TrimSpace(r.ProfileType), NameZH: strings.TrimSpace(r.NameZH), NameEN: strings.TrimSpace(r.NameEN), AddressZH: strings.TrimSpace(r.AddressZH), AddressEN: strings.TrimSpace(r.AddressEN), LicenseNumber: strings.TrimSpace(r.LicenseNumber), IsOverseas: r.IsOverseas, IsBigFour: r.IsBigFour, Phone1CountryCode: strings.TrimSpace(r.Phone1CountryCode), Phone1Number: strings.TrimSpace(r.Phone1Number), Phone1WhatsApp: r.Phone1WhatsApp, Phone2CountryCode: strings.TrimSpace(r.Phone2CountryCode), Phone2Number: strings.TrimSpace(r.Phone2Number), Phone2WhatsApp: r.Phone2WhatsApp, WechatID: strings.TrimSpace(r.WechatID), WechatURL: strings.TrimSpace(r.WechatURL), SignatureZH: strings.TrimSpace(r.SignatureZH), SignatureEN: strings.TrimSpace(r.SignatureEN), DefaultAvatar: strings.TrimSpace(r.DefaultAvatar), AvatarAssetID: strings.TrimSpace(r.AvatarAssetID), WechatQRAssetID: strings.TrimSpace(r.WechatQRAssetID), LogoAssetID: strings.TrimSpace(r.LogoAssetID), EAALicenseAssetID: strings.TrimSpace(r.EAALicenseAssetID), BusinessRegistrationAssetID: strings.TrimSpace(r.BusinessRegistrationAssetID), CompanyCardAssetID: strings.TrimSpace(r.CompanyCardAssetID)}, true
}
