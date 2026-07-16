/*
 * Agency profile service contracts.
 * 1. Define individual and company profile inputs and versioned responses.
 * 2. Define staff review and company subaccount contracts.
 */
package service

const (
	AgencyProfileTypeIndividual = "individual"
	AgencyProfileTypeCompany    = "company"
	AgencyProfileStatusDraft    = "draft"
	AgencyProfileStatusPending  = "pending"
	AgencyProfileStatusApproved = "approved"
	AgencyProfileStatusRejected = "rejected"
)

const (
	AccountTypePersonal                = "personal"
	AccountTypeIndividualAgent         = "individual_agent"
	AccountTypeAgencyCompany           = "agency_company"
	AccountTypeAgencyCompanySubaccount = "agency_company_subaccount"
)

// 1. AgencyProfileUpsertParams defines fields shared by both profile types.
type AgencyProfileUpsertParams struct {
	ProfileType                 string
	NameZH                      string
	NameEN                      string
	AddressZH                   string
	AddressEN                   string
	LicenseNumber               string
	IsOverseas                  bool
	IsBigFour                   bool
	Phone1CountryCode           string
	Phone1Number                string
	Phone1WhatsApp              bool
	Phone2CountryCode           string
	Phone2Number                string
	Phone2WhatsApp              bool
	WechatID                    string
	WechatURL                   string
	SignatureZH                 string
	SignatureEN                 string
	DefaultAvatar               string
	AvatarAssetID               string
	WechatQRAssetID             string
	LogoAssetID                 string
	EAALicenseAssetID           string
	BusinessRegistrationAssetID string
	CompanyCardAssetID          string
}

// 2. AgencyProfileAssetResponse defines one authorized media reference.
type AgencyProfileAssetResponse struct {
	MediaAssetID string `json:"media_asset_id"`
	URL          string `json:"url"`
	MimeType     string `json:"mime_type"`
}

// 3. AgencyProfileResponse defines one approved or working profile version.
type AgencyProfileResponse struct {
	ProfileID                 string                      `json:"profile_id"`
	ProfileType               string                      `json:"profile_type"`
	NameZH                    string                      `json:"name_zh"`
	NameEN                    string                      `json:"name_en"`
	AddressZH                 string                      `json:"address_zh"`
	AddressEN                 string                      `json:"address_en"`
	LicenseNumber             string                      `json:"license_number"`
	IsOverseas                bool                        `json:"is_overseas"`
	IsBigFour                 bool                        `json:"is_big_four"`
	Phone1CountryCode         string                      `json:"phone_1_country_code"`
	Phone1Number              string                      `json:"phone_1_number"`
	Phone1WhatsApp            bool                        `json:"phone_1_whatsapp"`
	Phone2CountryCode         string                      `json:"phone_2_country_code"`
	Phone2Number              string                      `json:"phone_2_number"`
	Phone2WhatsApp            bool                        `json:"phone_2_whatsapp"`
	WechatID                  string                      `json:"wechat_id"`
	WechatURL                 string                      `json:"wechat_url"`
	SignatureZH               string                      `json:"signature_zh"`
	SignatureEN               string                      `json:"signature_en"`
	DefaultAvatar             string                      `json:"default_avatar"`
	AvatarAsset               *AgencyProfileAssetResponse `json:"avatar_asset"`
	WechatQRAsset             *AgencyProfileAssetResponse `json:"wechat_qr_asset"`
	LogoAsset                 *AgencyProfileAssetResponse `json:"logo_asset"`
	EAALicenseAsset           *AgencyProfileAssetResponse `json:"eaa_license_asset"`
	BusinessRegistrationAsset *AgencyProfileAssetResponse `json:"business_registration_asset"`
	CompanyCardAsset          *AgencyProfileAssetResponse `json:"company_card_asset"`
	Status                    string                      `json:"status"`
	ReviewNote                string                      `json:"review_note"`
	SubmittedAt               *string                     `json:"submitted_at"`
	ReviewedAt                *string                     `json:"reviewed_at"`
	UpdatedAt                 string                      `json:"updated_at"`
	NextEditableAt            *string                     `json:"next_editable_at"`
}

// 4. AgencyProfileMemberResponse returns active data and one working revision.
type AgencyProfileMemberResponse struct {
	AccountType    string                 `json:"account_type"`
	MemberStatus   string                 `json:"member_status"`
	ActiveProfile  *AgencyProfileResponse `json:"active_profile"`
	Revision       *AgencyProfileResponse `json:"revision"`
	NextEditableAt *string                `json:"next_editable_at"`
}

// 5. AgencyProfileOwnerResponse defines the profile owner shown to staff.
type AgencyProfileOwnerResponse struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	PublicID    string `json:"public_id"`
}

// 6. AgencyProfileStaffResponse defines a staff review payload.
type AgencyProfileStaffResponse struct {
	AgencyProfileResponse
	Owner AgencyProfileOwnerResponse `json:"owner"`
}

// 7. AgencyProfileListFilters defines staff list filters.
type AgencyProfileListFilters struct {
	Page        int
	PageSize    int
	Status      string
	ProfileType string
	Keyword     string
}

// 8. AgencyProfileReviewParams defines a staff approval decision.
type AgencyProfileReviewParams struct {
	ReviewerUserID int64
	Approved       bool
	ReviewNote     string
}

// 9. AgencySubaccountCreateParams defines a company-created child login.
type AgencySubaccountCreateParams struct {
	DisplayName      string
	PhoneCountryCode string
	PhoneNumber      string
	Email            string
	Password         string
	Permissions      []string
}

// 10. AgencySubaccountResponse defines a company child account.
type AgencySubaccountResponse struct {
	PublicID         string   `json:"public_id"`
	UserPublicID     string   `json:"user_public_id"`
	DisplayName      string   `json:"display_name"`
	PhoneCountryCode string   `json:"phone_country_code"`
	PhoneNumber      string   `json:"phone_number"`
	Email            string   `json:"email"`
	Status           string   `json:"status"`
	Permissions      []string `json:"permissions"`
	CreatedAt        string   `json:"created_at"`
}
