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
	Page         int
	PageSize     int
	Keyword      string
	DistrictCode string
	Status       string
	MinPriceHKD  *float64
	MaxPriceHKD  *float64
	SortBy       string
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
	Title                 string
	Summary               string
	Description           string
	DistrictCode          string
	CommunityID           string
	PublisherIdentityType string
	PropertyType          string
	EstateName            string
	AddressText           string
	AskingPriceHKD        float64
	UsableAreaSqft        int
	GrossAreaSqft         *int
	BedroomCount          int
	LivingRoomCount       int
	BathroomCount         int
	FloorLevel            string
	Direction             string
	BuildingAge           string
	FeatureTags           []string
	ContactMethod         string
	BusinessStatus        string
	Images                []ListingImageInput
	Contact               PropertyContactInput
}

// 7. ServicedApartmentRoomTypeInput defines one room type input.
type ServicedApartmentRoomTypeInput struct {
	Name              string   `json:"name"`
	UsableAreaSqft    int      `json:"usable_area_sqft"`
	MonthlyRentHKD    float64  `json:"monthly_rent_hkd"`
	IncludedFees      bool     `json:"included_fees"`
	MinLeaseMonths    int      `json:"min_lease_months"`
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
	AddressText           string
	LowestMonthlyRentHKD  float64
	MinLeaseMonths        int
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
	PropertyType       string   `json:"property_type"`
	EstateName         string   `json:"estate_name"`
	AddressText        string   `json:"address_text"`
	AskingPriceHKD     float64  `json:"asking_price_hkd"`
	UsableAreaSqft     int      `json:"usable_area_sqft"`
	GrossAreaSqft      *int     `json:"gross_area_sqft,omitempty"`
	BedroomCount       int      `json:"bedroom_count"`
	LivingRoomCount    int      `json:"living_room_count"`
	BathroomCount      int      `json:"bathroom_count"`
	FloorLevel         string   `json:"floor_level"`
	Direction          string   `json:"direction"`
	BuildingAge        string   `json:"building_age"`
	FeatureTags        []string `json:"feature_tags"`
	ContactMethod      string   `json:"contact_method"`
	PublisherRoleLabel string   `json:"publisher_role_label"`
}

// 10. ServicedApartmentPayload defines serviced-specific response fields.
type ServicedApartmentPayload struct {
	ProjectName          string                           `json:"project_name"`
	AddressText          string                           `json:"address_text"`
	LowestMonthlyRentHKD float64                          `json:"lowest_monthly_rent_hkd"`
	MinLeaseMonths       int                              `json:"min_lease_months"`
	FacilityTags         []string                         `json:"facility_tags"`
	ServiceTags          []string                         `json:"service_tags"`
	RoomTypes            []ServicedApartmentRoomTypeInput `json:"room_types"`
	ContactMethod        string                           `json:"contact_method"`
	PublisherRoleLabel   string                           `json:"publisher_role_label"`
}

// 11. PropertyListingSummary defines public list payload for property channels.
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

// 12. PropertyListingDetail defines detail payload for property channels.
type PropertyListingDetail struct {
	PropertyListingSummary
	Description    string                 `json:"description"`
	Images         []ListingImageResponse `json:"images"`
	ContactSummary ListingContactSummary  `json:"contact_summary"`
}
