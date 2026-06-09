/*
 * Property channel DTOs.
 * 1. Define property sale and serviced apartment input payloads.
 * 2. Define shared listing responses for the two property channels.
 */
package service

// 1. PropertyChannel identifies a supported property listing module.
type PropertyChannel string

const (
	// 2. PropertyChannelSale identifies sale listings.
	PropertyChannelSale PropertyChannel = "property_sale"
	// 3. PropertyChannelServiced identifies serviced apartment listings.
	PropertyChannelServiced PropertyChannel = "serviced_apartment"
)

// 4. PropertyListFilters defines public and owner list filters.
type PropertyListFilters struct {
	Page                  int
	PageSize              int
	Keyword               string
	DistrictCode          string
	Status                string
	TransactionType       string
	PropertyType          string
	RentalType            string
	AreaMode              string
	BedroomCount          *int
	MinPriceHKD           *float64
	MaxPriceHKD           *float64
	MinAreaSqft           *int
	MaxAreaSqft           *int
	PriceMode             string
	FeatureTags           []string
	FacilityTags          []string
	ServiceTags           []string
	PublisherIdentityType string
	HasMedia              bool
	IsNew                 *bool
	SortBy                string
}

// 5. PropertyContactInput defines contact binding payload.
type PropertyContactInput struct {
	Phone           string
	WhatsApp        string
	Email           string
	ShowPhone       bool
	ShowWhatsApp    bool
	ShowChat        bool
	ShowInquiryForm bool
}

// 6. UpsertPropertySaleParams defines create and update input for sale listings.
type UpsertPropertySaleParams struct {
	OwnerUserID           int64
	ListingPublicID       string
	PropertyNo            string
	Title                 string
	TitleEn               string
	Summary               string
	Description           string
	DescriptionEn         string
	DistrictCode          string
	CommunityID           string
	PublisherIdentityType string
	TransactionType       string
	LocationScope         string
	ListingCategory       string
	MultiUnitProject      bool
	PropertyType          string
	RentalType            string
	EstateName            string
	AddressText           string
	AddressTextEn         string
	BlockName             string
	UnitName              string
	ShowUnit              bool
	AskingPriceHKD        float64
	MonthlyRentHKD        float64
	PriceReferenceOnly    bool
	PriceNegotiable       bool
	AnnualPrepayDiscount  bool
	AnnualPrepayOption    string
	LeaseStartDate        string
	RentIncluded          string
	AreaMode              string
	UsableAreaSqft        int
	GrossAreaSqft         *int
	BedroomCount          int
	LivingRoomCount       int
	BathroomCount         int
	FloorLevel            string
	FloorRaw              string
	FloorZone             string
	TotalFloors           int
	Direction             string
	BuildingAge           string
	KitchenType           string
	CookingMode           string
	ManagementFeeHKD      float64
	VideoURL              string
	VRURL                 string
	PrivateNote           string
	AdPackageCode         string
	FeatureTags           []string
	ContactMethod         string
	BusinessStatus        string
	Images                []ListingImageInput
	Contact               PropertyContactInput
}

// 7. ServicedApartmentRoomTypeInput defines one room type input.
type ServicedApartmentRoomTypeInput struct {
	Name              string   `json:"name"`
	RoomCategory      string   `json:"room_category"`
	UsableAreaSqft    int      `json:"usable_area_sqft"`
	MonthlyRentHKD    float64  `json:"monthly_rent_hkd"`
	MonthlyRentMinHKD float64  `json:"monthly_rent_min_hkd"`
	MonthlyRentMaxHKD float64  `json:"monthly_rent_max_hkd"`
	DailyRentMinHKD   float64  `json:"daily_rent_min_hkd"`
	DailyRentMaxHKD   float64  `json:"daily_rent_max_hkd"`
	IncludedFees      bool     `json:"included_fees"`
	IncludedFeeItems  []string `json:"included_fee_items"`
	MinLeaseMonths    int      `json:"min_lease_months"`
	MinStayValue      int      `json:"min_stay_value"`
	MinStayUnit       string   `json:"min_stay_unit"`
	FeatureTags       []string `json:"feature_tags"`
	ImageMediaAssetID string   `json:"image_media_asset_id"`
}

// 8. UpsertServicedApartmentParams defines create and update input for serviced apartments.
type UpsertServicedApartmentParams struct {
	OwnerUserID           int64
	ListingPublicID       string
	Title                 string
	Summary               string
	Description           string
	DistrictCode          string
	CommunityID           string
	PublisherIdentityType string
	ProjectName           string
	ProjectNameEn         string
	AddressText           string
	AddressTextEn         string
	WebsiteURL            string
	WhatsApp              string
	Fax                   string
	DescriptionEn         string
	ServiceIntro          string
	BenefitsText          string
	ExtraChargesText      string
	LowestMonthlyRentHKD  float64
	LowestDailyRentHKD    float64
	PriceReferenceOnly    bool
	PriceNegotiable       bool
	MinUsableAreaSqft     int
	MinLeaseMonths        int
	MinStayValue          int
	MinStayUnit           string
	LocationScope         string
	ListingCategory       string
	MultiUnitProject      bool
	AdPackageCode         string
	FacilityTags          []string
	ServiceTags           []string
	RoomTypes             []ServicedApartmentRoomTypeInput
	ContactMethod         string
	BusinessStatus        string
	Images                []ListingImageInput
	Contact               PropertyContactInput
}

// 9. PropertySalePayload defines sale-specific response fields.
type PropertySalePayload struct {
	PropertyNo           string   `json:"property_no"`
	TransactionType      string   `json:"transaction_type"`
	LocationScope        string   `json:"location_scope"`
	ListingCategory      string   `json:"listing_category"`
	MultiUnitProject     bool     `json:"multi_unit_project"`
	PropertyType         string   `json:"property_type"`
	RentalType           string   `json:"rental_type"`
	EstateName           string   `json:"estate_name"`
	AddressText          string   `json:"address_text"`
	AddressTextEn        string   `json:"address_text_en"`
	BlockName            string   `json:"block_name"`
	UnitName             string   `json:"unit_name,omitempty"`
	ShowUnit             bool     `json:"show_unit"`
	AskingPriceHKD       float64  `json:"asking_price_hkd"`
	MonthlyRentHKD       float64  `json:"monthly_rent_hkd"`
	PriceReferenceOnly   bool     `json:"price_reference_only"`
	PriceNegotiable      bool     `json:"price_negotiable"`
	AnnualPrepayDiscount bool     `json:"annual_prepay_discount"`
	AnnualPrepayOption   string   `json:"annual_prepay_option"`
	LeaseStartDate       string   `json:"lease_start_date"`
	RentIncluded         string   `json:"rent_included"`
	AreaMode             string   `json:"area_mode"`
	UsableAreaSqft       int      `json:"usable_area_sqft"`
	GrossAreaSqft        *int     `json:"gross_area_sqft,omitempty"`
	BedroomCount         int      `json:"bedroom_count"`
	LivingRoomCount      int      `json:"living_room_count"`
	BathroomCount        int      `json:"bathroom_count"`
	FloorLevel           string   `json:"floor_level"`
	FloorRaw             string   `json:"floor_raw,omitempty"`
	FloorZone            string   `json:"floor_zone"`
	FloorDisplayRange    string   `json:"floor_display_range"`
	TotalFloors          int      `json:"total_floors"`
	PublicLocationText   string   `json:"public_location_text"`
	Direction            string   `json:"direction"`
	BuildingAge          string   `json:"building_age"`
	KitchenType          string   `json:"kitchen_type"`
	CookingMode          string   `json:"cooking_mode"`
	ManagementFeeHKD     float64  `json:"management_fee_hkd"`
	VideoURL             string   `json:"video_url"`
	VRURL                string   `json:"vr_url"`
	PrivateNote          string   `json:"private_note"`
	TitleEn              string   `json:"title_en"`
	DescriptionEn        string   `json:"description_en"`
	AdPackageCode        string   `json:"ad_package_code"`
	AdWeight             int      `json:"ad_weight"`
	AdPriceHKD           float64  `json:"ad_price_hkd"`
	AdPricePoints        int64    `json:"ad_price_points"`
	AdDurationDays       int      `json:"ad_duration_days"`
	AdExpiresAt          *string  `json:"ad_expires_at,omitempty"`
	FeatureTags          []string `json:"feature_tags"`
	ContactMethod        string   `json:"contact_method"`
	PublisherRoleLabel   string   `json:"publisher_role_label"`
}

// 10. PropertyAddressSuggestion defines one building address autocomplete result.
type PropertyAddressSuggestion struct {
	AddressID      string   `json:"address_id"`
	EstateName     string   `json:"estate_name"`
	EstateNameEn   string   `json:"estate_name_en"`
	DisplayName    string   `json:"display_name"`
	AddressText    string   `json:"address_text"`
	AddressTextEn  string   `json:"address_text_en"`
	DistrictCode   string   `json:"district_code"`
	DistrictLabel  string   `json:"district_label"`
	RegionCode     string   `json:"region_code"`
	BlockNames     []string `json:"block_names"`
	CompletionYear int      `json:"completion_year"`
	Remark         string   `json:"remark"`
}

// 11. ServicedApartmentPayload defines serviced-specific response fields.
type ServicedApartmentPayload struct {
	ProjectName          string                           `json:"project_name"`
	ProjectNameEn        string                           `json:"project_name_en"`
	AddressText          string                           `json:"address_text"`
	AddressTextEn        string                           `json:"address_text_en"`
	WebsiteURL           string                           `json:"website_url"`
	WhatsApp             string                           `json:"whatsapp"`
	Fax                  string                           `json:"fax"`
	DescriptionEn        string                           `json:"description_en"`
	ServiceIntro         string                           `json:"service_intro"`
	BenefitsText         string                           `json:"benefits_text"`
	ExtraChargesText     string                           `json:"extra_charges_text"`
	LowestMonthlyRentHKD float64                          `json:"lowest_monthly_rent_hkd"`
	LowestDailyRentHKD   float64                          `json:"lowest_daily_rent_hkd"`
	PriceReferenceOnly   bool                             `json:"price_reference_only"`
	PriceNegotiable      bool                             `json:"price_negotiable"`
	MinUsableAreaSqft    int                              `json:"min_usable_area_sqft"`
	MinLeaseMonths       int                              `json:"min_lease_months"`
	MinStayValue         int                              `json:"min_stay_value"`
	MinStayUnit          string                           `json:"min_stay_unit"`
	LocationScope        string                           `json:"location_scope"`
	ListingCategory      string                           `json:"listing_category"`
	MultiUnitProject     bool                             `json:"multi_unit_project"`
	AdPackageCode        string                           `json:"ad_package_code"`
	AdWeight             int                              `json:"ad_weight"`
	AdPriceHKD           float64                          `json:"ad_price_hkd"`
	AdPricePoints        int64                            `json:"ad_price_points"`
	AdDurationDays       int                              `json:"ad_duration_days"`
	AdExpiresAt          *string                          `json:"ad_expires_at,omitempty"`
	FacilityTags         []string                         `json:"facility_tags"`
	ServiceTags          []string                         `json:"service_tags"`
	RoomTypes            []ServicedApartmentRoomTypeInput `json:"room_types"`
	ContactMethod        string                           `json:"contact_method"`
	PublisherRoleLabel   string                           `json:"publisher_role_label"`
}

// 12. PropertyListingSummary defines public list payload for property channels.
type PropertyListingSummary struct {
	ListingID             string                    `json:"listing_id"`
	Module                string                    `json:"module"`
	Title                 string                    `json:"title"`
	Summary               string                    `json:"summary"`
	DistrictCode          string                    `json:"district_code"`
	PublisherIdentityType string                    `json:"publisher_identity_type"`
	PublicationStatus     string                    `json:"publication_status"`
	BusinessStatus        string                    `json:"business_status"`
	ExpireAt              *string                   `json:"expire_at,omitempty"`
	PublishedAt           *string                   `json:"published_at,omitempty"`
	UpdatedAt             string                    `json:"updated_at"`
	Community             *CommunityResponse        `json:"community,omitempty"`
	Owner                 *UserPreviewResponse      `json:"owner,omitempty"`
	CoverImage            *ListingImageResponse     `json:"cover_image,omitempty"`
	PropertySale          *PropertySalePayload      `json:"property_sale,omitempty"`
	ServicedApartment     *ServicedApartmentPayload `json:"serviced_apartment,omitempty"`
}

// 13. PropertyListingDetail defines detail payload for property channels.
type PropertyListingDetail struct {
	PropertyListingSummary
	Description         string                 `json:"description"`
	Images              []ListingImageResponse `json:"images"`
	ContactSummary      ListingContactSummary  `json:"contact_summary"`
	PointsCharged       int64                  `json:"points_charged,omitempty"`
	PointsBalanceAfter  *int64                 `json:"points_balance_after,omitempty"`
	PointsTransactionID string                 `json:"points_transaction_id,omitempty"`
}
