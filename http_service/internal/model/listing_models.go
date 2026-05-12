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
	ListingID          int64          `gorm:"primaryKey" json:"listing_id"`
	PropertyType       string         `gorm:"type:varchar(64);not null" json:"property_type"`
	EstateName         string         `gorm:"type:varchar(200)" json:"estate_name"`
	AddressText        string         `gorm:"type:varchar(500);not null" json:"address_text"`
	AskingPriceHKD     float64        `gorm:"type:numeric(14,2);not null" json:"asking_price_hkd"`
	UsableAreaSqft     int            `gorm:"not null" json:"usable_area_sqft"`
	GrossAreaSqft      *int           `json:"gross_area_sqft"`
	BedroomCount       int            `gorm:"not null;default:0" json:"bedroom_count"`
	LivingRoomCount    int            `gorm:"not null;default:0" json:"living_room_count"`
	BathroomCount      int            `gorm:"not null;default:0" json:"bathroom_count"`
	FloorLevel         string         `gorm:"type:varchar(64)" json:"floor_level"`
	Direction          string         `gorm:"type:varchar(64)" json:"direction"`
	BuildingAge        string         `gorm:"type:varchar(64)" json:"building_age"`
	FeatureTags        datatypes.JSON `gorm:"type:jsonb" json:"feature_tags"`
	ContactMethod      string         `gorm:"type:varchar(32);not null" json:"contact_method"`
	PublisherRoleLabel string         `gorm:"type:varchar(64)" json:"publisher_role_label"`
}

// 10. ServicedApartmentProject stores serviced apartment project fields.
type ServicedApartmentProject struct {
	ListingID            int64          `gorm:"primaryKey" json:"listing_id"`
	ProjectName          string         `gorm:"type:varchar(200);not null" json:"project_name"`
	AddressText          string         `gorm:"type:varchar(500);not null" json:"address_text"`
	LowestMonthlyRentHKD float64        `gorm:"type:numeric(12,2);not null" json:"lowest_monthly_rent_hkd"`
	MinLeaseMonths       int            `gorm:"not null;default:1" json:"min_lease_months"`
	FacilityTags         datatypes.JSON `gorm:"type:jsonb" json:"facility_tags"`
	ServiceTags          datatypes.JSON `gorm:"type:jsonb" json:"service_tags"`
	RoomTypes            datatypes.JSON `gorm:"type:jsonb" json:"room_types"`
	ContactMethod        string         `gorm:"type:varchar(32);not null" json:"contact_method"`
	PublisherRoleLabel   string         `gorm:"type:varchar(64)" json:"publisher_role_label"`
}
