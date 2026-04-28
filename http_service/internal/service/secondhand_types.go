/*
 * Secondhand service DTOs.
 * 1. Define input and output structures for secondhand listing workflows.
 * 2. Keep handler payload mapping separate from database models.
 */
package service

import "ajoliving_web/http_service/internal/model"

// 1. ListingImageInput defines image binding payload.
type ListingImageInput struct {
	MediaAssetID string `json:"media_asset_id"`
	SortOrder    int    `json:"sort_order"`
	IsCover      bool   `json:"is_cover"`
}

// 2. ListingContactInput defines contact binding payload.
type ListingContactInput struct {
	Phone           string
	WhatsApp        string
	Email           string
	ShowPhone       bool
	ShowWhatsApp    bool
	ShowChat        bool
	ShowInquiryForm bool
}

// 3. UpsertSecondhandParams defines create and update listing input.
type UpsertSecondhandParams struct {
	OwnerUserID           int64
	ListingPublicID       string
	Title                 string
	Summary               string
	Description           string
	DistrictCode          string
	CommunityID           string
	PublisherIdentityType string
	CategoryCode          string
	PriceMode             string
	PriceHKD              *float64
	ConditionLevel        string
	DimensionText         string
	PickupRegionCode      string
	PickupLocationText    string
	DeliveryTags          []string
	VisibilityScope       string
	ContactMethod         string
	BusinessStatus        string
	Images                []ListingImageInput
	Contact               ListingContactInput
}

// 4. SecondhandListFilters defines public list query filters.
type SecondhandListFilters struct {
	Page            int
	PageSize        int
	Keyword         string
	CategoryCode    string
	RegionCode      string
	DistrictCode    string
	PriceMode       string
	ConditionLevel  string
	VisibilityScope string
	MinPriceHKD     *float64
	MaxPriceHKD     *float64
	OnlyBuilding    bool
	OnlyFree        bool
	ExcludeFree     bool
	SortBy          string
	CommunityID     *int64
}

// 5. MySecondhandFilters defines owner list filters.
type MySecondhandFilters struct {
	Page     int
	PageSize int
	Status   string
}

// 6. ListingImageResponse defines image response payload.
type ListingImageResponse struct {
	MediaAssetID string `json:"media_asset_id"`
	URL          string `json:"url"`
	SortOrder    int    `json:"sort_order"`
	IsCover      bool   `json:"is_cover"`
}

// 7. ListingContactSummary defines contact channel summary.
type ListingContactSummary struct {
	ShowPhone    bool `json:"show_phone"`
	ShowWhatsApp bool `json:"show_whatsapp"`
	ShowChat     bool `json:"show_chat"`
	ShowInquiry  bool `json:"show_inquiry_form"`
}

// 8. SecondhandListingSummary defines public list payload.
type SecondhandListingSummary struct {
	ListingID             string                `json:"listing_id"`
	Title                 string                `json:"title"`
	Summary               string                `json:"summary"`
	DistrictCode          string                `json:"district_code"`
	PublishedAt           *string               `json:"published_at,omitempty"`
	CategoryCode          string                `json:"category_code"`
	PriceMode             string                `json:"price_mode"`
	PriceHKD              *float64              `json:"price_hkd,omitempty"`
	ConditionLevel        string                `json:"condition_level"`
	VisibilityScope       string                `json:"visibility_scope"`
	ContactMethod         string                `json:"contact_method"`
	IsFreeGiveaway        bool                  `json:"is_free_giveaway"`
	PublisherIdentityType string                `json:"publisher_identity_type"`
	PublicationStatus     string                `json:"publication_status"`
	BusinessStatus        string                `json:"business_status"`
	ExpireAt              *string               `json:"expire_at,omitempty"`
	UpdatedAt             string                `json:"updated_at"`
	Community             *CommunityResponse    `json:"community,omitempty"`
	Owner                 *UserPreviewResponse  `json:"owner,omitempty"`
	CoverImage            *ListingImageResponse `json:"cover_image,omitempty"`
}

// 9. UserPreviewResponse defines reusable member display payload.
type UserPreviewResponse struct {
	UserID                string `json:"user_id"`
	PublicID              string `json:"public_id"`
	DisplayName           string `json:"display_name"`
	PublisherIdentityType string `json:"publisher_identity_type"`
}

// 10. SecondhandListingDetail defines detail payload.
type SecondhandListingDetail struct {
	SecondhandListingSummary
	Description        string                 `json:"description"`
	DimensionText      string                 `json:"dimension_text"`
	PickupRegionCode   string                 `json:"pickup_region_code"`
	PickupLocationText string                 `json:"pickup_location_text"`
	DeliveryTags       []string               `json:"delivery_tags"`
	Images             []ListingImageResponse `json:"images"`
	ContactSummary     ListingContactSummary  `json:"contact_summary"`
}

// 11. ContactAccessResult defines contact access payload.
type ContactAccessResult struct {
	ListingID       string            `json:"listing_id"`
	AllowedChannels map[string]bool   `json:"allowed_channels"`
	ContactPayload  map[string]string `json:"contact_payload,omitempty"`
}

// 12. toISOTime converts a time pointer to RFC3339 string pointer.
func toISOTime(value *model.Listing) *string {
	if value == nil || value.ExpireAt == nil {
		return nil
	}

	formatted := value.ExpireAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	return &formatted
}
