/*
 * Listing and media data models.
 * 1. Define shared listing entities, secondhand module extension, and discover placements.
 * 2. Keep sensitive contact and visibility fields isolated for service access.
 */
package model

import (
	"time"

	"gorm.io/datatypes"
)

// 1. Listing stores the shared listing root entity.
type Listing struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID              string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	Module                string     `gorm:"type:varchar(32);not null;index:idx_listings_module_status,priority:1" json:"module"`
	OwnerUserID           int64      `gorm:"not null;index:idx_listings_owner_module_status,priority:1" json:"owner_user_id"`
	Title                 string     `gorm:"type:varchar(300);not null" json:"title"`
	Summary               string     `gorm:"type:varchar(500)" json:"summary"`
	Description           string     `gorm:"type:text" json:"description"`
	DistrictCode          string     `gorm:"type:varchar(32);not null" json:"district_code"`
	CommunityID           *int64     `gorm:"index" json:"community_id"`
	PublisherIdentityType string     `gorm:"type:varchar(32);not null" json:"publisher_identity_type"`
	PublicationStatus     string     `gorm:"type:varchar(32);not null;index:idx_listings_module_status,priority:2;index:idx_listings_expire_at,priority:1;index:idx_listings_owner_module_status,priority:3" json:"publication_status"`
	ModerationStatus      string     `gorm:"type:varchar(32);not null;index:idx_listings_module_status,priority:3" json:"moderation_status"`
	BusinessStatus        string     `gorm:"type:varchar(32);not null" json:"business_status"`
	PublishedAt           *time.Time `json:"published_at"`
	SortRefreshedAt       *time.Time `gorm:"index:idx_listings_module_status,priority:4" json:"sort_refreshed_at"`
	ExpireAt              *time.Time `gorm:"index:idx_listings_expire_at,priority:2" json:"expire_at"`
	IsDeleted             bool       `gorm:"not null;index" json:"is_deleted"`
	DeletedAt             *time.Time `json:"deleted_at"`
	TimestampModel
}

// 2. ListingContact stores encrypted contact values.
type ListingContact struct {
	ListingID         int64     `gorm:"primaryKey" json:"listing_id"`
	PhoneEncrypted    string    `gorm:"type:text" json:"phone_encrypted"`
	PhoneMasked       string    `gorm:"type:varchar(64)" json:"phone_masked"`
	WhatsAppEncrypted string    `gorm:"type:text" json:"whatsapp_encrypted"`
	WhatsAppMasked    string    `gorm:"type:varchar(64)" json:"whatsapp_masked"`
	EmailEncrypted    string    `gorm:"type:text" json:"email_encrypted"`
	ShowPhone         bool      `gorm:"not null" json:"show_phone"`
	ShowWhatsApp      bool      `gorm:"not null" json:"show_whatsapp"`
	ShowChat          bool      `gorm:"not null" json:"show_chat"`
	ShowInquiryForm   bool      `gorm:"not null" json:"show_inquiry_form"`
	ContactMode       string    `gorm:"type:varchar(32);not null" json:"contact_mode"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// 3. MediaAsset stores uploaded media metadata.
type MediaAsset struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID        string    `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	StorageProvider string    `gorm:"type:varchar(32);not null" json:"storage_provider"`
	BucketName      string    `gorm:"type:varchar(120);not null" json:"bucket_name"`
	ObjectKey       string    `gorm:"type:varchar(500);not null;index" json:"object_key"`
	MimeType        string    `gorm:"type:varchar(100);not null" json:"mime_type"`
	Width           *int      `json:"width"`
	Height          *int      `json:"height"`
	FileSize        int64     `json:"file_size"`
	ChecksumSHA256  string    `gorm:"type:varchar(128)" json:"checksum_sha256"`
	CreatedBy       *int64    `gorm:"index" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// 4. ListingImage stores listing image order and cover state.
type ListingImage struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ListingID    int64     `gorm:"not null;index" json:"listing_id"`
	MediaAssetID int64     `gorm:"not null;index" json:"media_asset_id"`
	SortOrder    int       `gorm:"not null" json:"sort_order"`
	IsCover      bool      `gorm:"not null" json:"is_cover"`
	CreatedAt    time.Time `json:"created_at"`
}

// 5. ContactAccessLog stores contact reveal audit records.
type ContactAccessLog struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ListingID       int64          `gorm:"not null;index" json:"listing_id"`
	RequestUserID   int64          `gorm:"not null;index" json:"request_user_id"`
	GrantedChannels datatypes.JSON `gorm:"type:jsonb;not null" json:"granted_channels"`
	RequestIP       string         `gorm:"type:varchar(64)" json:"request_ip"`
	UserAgent       string         `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt       time.Time      `json:"created_at"`
}

// 6. DiscoverPlacement stores manually curated discover page slots.
type DiscoverPlacement struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Scene        string `gorm:"type:varchar(64);not null;uniqueIndex:uk_discover_placements_position,priority:1;index" json:"scene"`
	CategoryCode string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_discover_placements_position,priority:2;index" json:"category_code"`
	SlotIndex    int    `gorm:"not null;uniqueIndex:uk_discover_placements_position,priority:3" json:"slot_index"`
	ListingID    int64  `gorm:"not null;index" json:"listing_id"`
	CreatedBy    *int64 `gorm:"index" json:"created_by"`
	UpdatedBy    *int64 `gorm:"index" json:"updated_by"`
	TimestampModel
}

// 7. ListingFavorite stores member saved listing relationships.
type ListingFavorite struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:uk_listing_favorite_user_listing,priority:1;index" json:"user_id"`
	ListingID int64     `gorm:"not null;uniqueIndex:uk_listing_favorite_user_listing,priority:2;index" json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Listing   *Listing  `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
}

// 8. SecondhandListing stores secondhand-only listing fields.
type SecondhandListing struct {
	ListingID          int64          `gorm:"primaryKey" json:"listing_id"`
	CategoryCode       string         `gorm:"type:varchar(64);not null" json:"category_code"`
	PriceMode          string         `gorm:"type:varchar(32);not null" json:"price_mode"`
	PriceHKD           *float64       `gorm:"type:numeric(12,2)" json:"price_hkd"`
	ConditionLevel     string         `gorm:"type:varchar(32);not null" json:"condition_level"`
	DimensionText      string         `gorm:"type:varchar(300)" json:"dimension_text"`
	PickupRegionCode   string         `gorm:"type:varchar(32);not null" json:"pickup_region_code"`
	PickupLocationText string         `gorm:"type:varchar(300);not null" json:"pickup_location_text"`
	DeliveryTags       datatypes.JSON `gorm:"type:jsonb" json:"delivery_tags"`
	VisibilityScope    string         `gorm:"type:varchar(32);not null;index:idx_secondhand_visibility_community,priority:1" json:"visibility_scope"`
	VisibleCommunityID *int64         `gorm:"index:idx_secondhand_visibility_community,priority:2" json:"visible_community_id"`
	ContactMethod      string         `gorm:"type:varchar(32);not null" json:"contact_method"`
	IsFreeGiveaway     bool           `gorm:"not null" json:"is_free_giveaway"`
}

// 9. PropertySaleListing stores property sale listing fields.
type PropertySaleListing struct {
	ListingID            int64          `gorm:"primaryKey" json:"listing_id"`
	PropertyNo           string         `gorm:"type:varchar(80);index" json:"property_no"`
	TransactionType      string         `gorm:"type:varchar(16);not null;default:'sale';index" json:"transaction_type"`
	LocationScope        string         `gorm:"type:varchar(32);not null;default:'local'" json:"location_scope"`
	ListingCategory      string         `gorm:"type:varchar(64);not null;default:'standard'" json:"listing_category"`
	MultiUnitProject     bool           `gorm:"not null;default:false" json:"multi_unit_project"`
	PropertyType         string         `gorm:"type:varchar(64);not null" json:"property_type"`
	RentalType           string         `gorm:"type:varchar(64)" json:"rental_type"`
	EstateName           string         `gorm:"type:varchar(200)" json:"estate_name"`
	AddressText          string         `gorm:"type:varchar(500);not null" json:"address_text"`
	AddressTextEn        string         `gorm:"type:varchar(500)" json:"address_text_en"`
	BlockName            string         `gorm:"type:varchar(120)" json:"block_name"`
	UnitName             string         `gorm:"type:varchar(80)" json:"unit_name"`
	ShowUnit             bool           `gorm:"not null;default:true" json:"show_unit"`
	AskingPriceHKD       float64        `gorm:"type:numeric(14,2);not null" json:"asking_price_hkd"`
	MonthlyRentHKD       float64        `gorm:"type:numeric(12,2);not null;default:0" json:"monthly_rent_hkd"`
	PriceReferenceOnly   bool           `gorm:"not null;default:false" json:"price_reference_only"`
	PriceNegotiable      bool           `gorm:"not null;default:false" json:"price_negotiable"`
	AnnualPrepayDiscount bool           `gorm:"not null;default:false" json:"annual_prepay_discount"`
	AnnualPrepayOption   string         `gorm:"type:varchar(16);not null;default:'none'" json:"annual_prepay_option"`
	LeaseStartDate       string         `gorm:"type:varchar(20)" json:"lease_start_date"`
	RentIncluded         string         `gorm:"type:varchar(300)" json:"rent_included"`
	AreaMode             string         `gorm:"type:varchar(32);not null;default:'usable'" json:"area_mode"`
	UsableAreaSqft       int            `gorm:"not null" json:"usable_area_sqft"`
	GrossAreaSqft        *int           `json:"gross_area_sqft"`
	BedroomCount         int            `gorm:"not null;default:0" json:"bedroom_count"`
	LivingRoomCount      int            `gorm:"not null;default:0" json:"living_room_count"`
	BathroomCount        int            `gorm:"not null;default:0" json:"bathroom_count"`
	FloorLevel           string         `gorm:"type:varchar(64)" json:"floor_level"`
	FloorRaw             string         `gorm:"type:varchar(64)" json:"floor_raw"`
	FloorZone            string         `gorm:"type:varchar(16)" json:"floor_zone"`
	FloorDisplayRange    string         `gorm:"type:varchar(80)" json:"floor_display_range"`
	TotalFloors          int            `gorm:"not null;default:0" json:"total_floors"`
	PublicLocationText   string         `gorm:"type:varchar(300)" json:"public_location_text"`
	Direction            string         `gorm:"type:varchar(64)" json:"direction"`
	BuildingAge          string         `gorm:"type:varchar(64)" json:"building_age"`
	KitchenType          string         `gorm:"type:varchar(64)" json:"kitchen_type"`
	CookingMode          string         `gorm:"type:varchar(64)" json:"cooking_mode"`
	ManagementFeeHKD     float64        `gorm:"type:numeric(12,2);not null;default:0" json:"management_fee_hkd"`
	VideoURL             string         `gorm:"type:varchar(500)" json:"video_url"`
	VRURL                string         `gorm:"type:varchar(500)" json:"vr_url"`
	PrivateNote          string         `gorm:"type:text" json:"private_note"`
	TitleEn              string         `gorm:"type:varchar(300)" json:"title_en"`
	DescriptionEn        string         `gorm:"type:text" json:"description_en"`
	AdPackageCode        string         `gorm:"type:varchar(32);not null;default:'basic';index" json:"ad_package_code"`
	AdWeight             int            `gorm:"not null;default:0;index" json:"ad_weight"`
	AdPriceHKD           float64        `gorm:"type:numeric(12,2);not null;default:600" json:"ad_price_hkd"`
	AdPricePoints        int64          `gorm:"not null;default:600" json:"ad_price_points"`
	AdDurationDays       int            `gorm:"not null;default:30" json:"ad_duration_days"`
	AdExpiresAt          *time.Time     `json:"ad_expires_at"`
	FeatureTags          datatypes.JSON `gorm:"type:jsonb" json:"feature_tags"`
	ContactMethod        string         `gorm:"type:varchar(32);not null" json:"contact_method"`
	PublisherRoleLabel   string         `gorm:"type:varchar(64)" json:"publisher_role_label"`
}

// 10. PropertyAddress stores searchable building and address reference data.
type PropertyAddress struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID       string         `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	SourceFile     string         `gorm:"type:varchar(120);not null;uniqueIndex:uk_property_address_source_name_address,priority:1" json:"source_file"`
	RegionCode     string         `gorm:"type:varchar(32);index" json:"region_code"`
	DistrictCode   string         `gorm:"type:varchar(32);index" json:"district_code"`
	EstateName     string         `gorm:"type:varchar(200);not null;index;uniqueIndex:uk_property_address_source_name_address,priority:2" json:"estate_name"`
	EstateNameEn   string         `gorm:"type:varchar(200)" json:"estate_name_en"`
	DisplayName    string         `gorm:"type:varchar(300)" json:"display_name"`
	AddressText    string         `gorm:"type:varchar(500);not null;uniqueIndex:uk_property_address_source_name_address,priority:3" json:"address_text"`
	AddressTextEn  string         `gorm:"type:varchar(500)" json:"address_text_en"`
	DistrictLabel  string         `gorm:"type:varchar(120)" json:"district_label"`
	BlockNames     datatypes.JSON `gorm:"type:jsonb" json:"block_names"`
	CompletionYear int            `gorm:"not null;default:0" json:"completion_year"`
	Remark         string         `gorm:"type:varchar(500)" json:"remark"`
	SearchText     string         `gorm:"type:text;index" json:"search_text"`
	TimestampModel
}

// 11. ServicedApartmentProject stores serviced apartment project fields.
type ServicedApartmentProject struct {
	ListingID            int64          `gorm:"primaryKey" json:"listing_id"`
	ProjectName          string         `gorm:"type:varchar(200);not null" json:"project_name"`
	ProjectNameEn        string         `gorm:"type:varchar(200)" json:"project_name_en"`
	AddressText          string         `gorm:"type:varchar(500);not null" json:"address_text"`
	AddressTextEn        string         `gorm:"type:varchar(500)" json:"address_text_en"`
	WebsiteURL           string         `gorm:"type:varchar(500)" json:"website_url"`
	WhatsApp             string         `gorm:"column:whatsapp;type:varchar(80)" json:"whatsapp"`
	Fax                  string         `gorm:"type:varchar(80)" json:"fax"`
	DescriptionEn        string         `gorm:"type:text" json:"description_en"`
	ServiceIntro         string         `gorm:"type:text" json:"service_intro"`
	BenefitsText         string         `gorm:"type:text" json:"benefits_text"`
	ExtraChargesText     string         `gorm:"type:text" json:"extra_charges_text"`
	LowestMonthlyRentHKD float64        `gorm:"type:numeric(12,2);not null" json:"lowest_monthly_rent_hkd"`
	LowestDailyRentHKD   float64        `gorm:"type:numeric(12,2);not null;default:0" json:"lowest_daily_rent_hkd"`
	PriceReferenceOnly   bool           `gorm:"not null;default:false" json:"price_reference_only"`
	PriceNegotiable      bool           `gorm:"not null;default:false" json:"price_negotiable"`
	MinUsableAreaSqft    int            `gorm:"not null;default:0;index" json:"min_usable_area_sqft"`
	MinLeaseMonths       int            `gorm:"not null;default:1" json:"min_lease_months"`
	MinStayValue         int            `gorm:"not null;default:1" json:"min_stay_value"`
	MinStayUnit          string         `gorm:"type:varchar(16);not null;default:'month'" json:"min_stay_unit"`
	LocationScope        string         `gorm:"type:varchar(32);not null;default:'local'" json:"location_scope"`
	ListingCategory      string         `gorm:"type:varchar(64);not null;default:'standard'" json:"listing_category"`
	MultiUnitProject     bool           `gorm:"not null;default:false" json:"multi_unit_project"`
	FacilityTags         datatypes.JSON `gorm:"type:jsonb" json:"facility_tags"`
	ServiceTags          datatypes.JSON `gorm:"type:jsonb" json:"service_tags"`
	RoomTypes            datatypes.JSON `gorm:"type:jsonb" json:"room_types"`
	AdPackageCode        string         `gorm:"type:varchar(32);not null;default:'basic';index" json:"ad_package_code"`
	AdWeight             int            `gorm:"not null;default:0;index" json:"ad_weight"`
	AdPriceHKD           float64        `gorm:"type:numeric(12,2);not null;default:600" json:"ad_price_hkd"`
	AdPricePoints        int64          `gorm:"not null;default:800" json:"ad_price_points"`
	AdDurationDays       int            `gorm:"not null;default:30" json:"ad_duration_days"`
	AdExpiresAt          *time.Time     `json:"ad_expires_at"`
	ContactMethod        string         `gorm:"type:varchar(32);not null" json:"contact_method"`
	PublisherRoleLabel   string         `gorm:"type:varchar(64)" json:"publisher_role_label"`
}
