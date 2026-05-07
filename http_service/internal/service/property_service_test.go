/*
 * Property service tests.
 * 1. Validate sale listing publish, discovery, and contact access flows.
 * 2. Validate serviced apartment publish and discovery flows.
 */
package service

import (
	"context"
	"testing"
)

// 1. TestPropertySalePublishListAndContactAccess validates the sale channel happy path.
func TestPropertySalePublishListAndContactAccess(t *testing.T) {
	runtime := newTestRuntime(t)
	propertyService := NewPropertyService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91000001", &communityA.ID)
	viewer := mustCreateUser(t, runtime, "+852", "91000002", &communityA.ID)
	assetID := mustCreateMediaAsset(t, runtime, owner.ID)

	created, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           owner.ID,
		Title:                 "Harbour View Two Bedroom",
		Summary:               "Bright flat near MTR.",
		Description:           "Ready for viewing with practical layout.",
		DistrictCode:          "kwun_tong",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		PropertyType:          "private_flat",
		EstateName:            "AJO Residence",
		AddressText:           "18 Hoi Bun Road",
		AskingPriceHKD:        7800000,
		UsableAreaSqft:        520,
		BedroomCount:          2,
		LivingRoomCount:       1,
		BathroomCount:         1,
		FeatureTags:           []string{"near_mtr", "sea_view"},
		ContactMethod:         "both",
		Images: []ListingImageInput{
			{MediaAssetID: assetID, SortOrder: 1, IsCover: true},
		},
		Contact: PropertyContactInput{
			Phone:        "+85291234567",
			WhatsApp:     "+85291234567",
			ShowPhone:    true,
			ShowWhatsApp: true,
		},
	})
	if err != nil {
		t.Fatalf("create sale listing: %v", err)
	}
	if created.PublicationStatus != "draft" || created.PropertySale == nil {
		t.Fatalf("expected sale draft response, got %#v", created)
	}

	published, err := propertyService.PublishProperty(context.Background(), PropertyChannelSale, owner.ID, created.ListingID)
	if err != nil {
		t.Fatalf("publish sale listing: %v", err)
	}
	if published.PublicationStatus != "active" || published.ExpireAt == nil {
		t.Fatalf("expected active sale listing with expiry, got %#v", published)
	}

	items, pagination, err := propertyService.ListPublicProperties(context.Background(), PropertyChannelSale, PropertyListFilters{
		Page:         1,
		PageSize:     20,
		DistrictCode: "kwun_tong",
		SortBy:       "price_asc",
	})
	if err != nil {
		t.Fatalf("list public sale listings: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 || items[0].PropertySale == nil || items[0].CoverImage == nil {
		t.Fatalf("expected one sale listing with cover, got total=%d items=%#v", pagination.Total, items)
	}

	access, err := propertyService.GrantPropertyContactAccess(context.Background(), PropertyChannelSale, viewer.ID, created.ListingID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("grant sale contact access: %v", err)
	}
	if !access.AllowedChannels["phone"] || !access.AllowedChannels["whatsapp"] {
		t.Fatalf("expected phone and whatsapp access, got %#v", access.AllowedChannels)
	}
	if access.ContactPayload["phone"] != "+85291234567" || access.ContactPayload["whatsapp_url"] == "" {
		t.Fatalf("expected contact payload, got %#v", access.ContactPayload)
	}
}

// 2. TestServicedApartmentPublishAndList validates the serviced apartment happy path.
func TestServicedApartmentPublishAndList(t *testing.T) {
	runtime := newTestRuntime(t)
	propertyService := NewPropertyService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91000011", &communityA.ID)
	assetID := mustCreateMediaAsset(t, runtime, owner.ID)

	created, err := propertyService.CreateServicedApartment(context.Background(), UpsertServicedApartmentParams{
		OwnerUserID:           owner.ID,
		Title:                 "Central Serviced Residence",
		Summary:               "Flexible monthly stay with housekeeping.",
		Description:           "Suitable for relocation and business stays.",
		DistrictCode:          "central_western",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "agent",
		ProjectName:           "AJO Suites Central",
		AddressText:           "8 Queen's Road Central",
		LowestMonthlyRentHKD:  23800,
		MinLeaseMonths:        1,
		FacilityTags:          []string{"gym", "laundry"},
		ServiceTags:           []string{"housekeeping", "wifi"},
		RoomTypes: []ServicedApartmentRoomTypeInput{
			{
				Name:           "Studio",
				UsableAreaSqft: 260,
				MonthlyRentHKD: 23800,
				IncludedFees:   true,
				MinLeaseMonths: 1,
				FeatureTags:    []string{"city_view"},
			},
		},
		ContactMethod: "whatsapp",
		Images: []ListingImageInput{
			{MediaAssetID: assetID, SortOrder: 1, IsCover: true},
		},
		Contact: PropertyContactInput{
			WhatsApp:     "+85298765432",
			ShowWhatsApp: true,
		},
	})
	if err != nil {
		t.Fatalf("create serviced apartment: %v", err)
	}

	published, err := propertyService.PublishProperty(context.Background(), PropertyChannelServiced, owner.ID, created.ListingID)
	if err != nil {
		t.Fatalf("publish serviced apartment: %v", err)
	}
	if published.PublicationStatus != "active" || published.ServicedApartment == nil || published.ServicedApartment.PublisherRoleLabel != "代理人" {
		t.Fatalf("expected active serviced apartment response, got %#v", published)
	}

	items, pagination, err := propertyService.ListPublicProperties(context.Background(), PropertyChannelServiced, PropertyListFilters{
		Page:        1,
		PageSize:    20,
		MinPriceHKD: ptrFloat64(20000),
		MaxPriceHKD: ptrFloat64(30000),
		SortBy:      "price_desc",
	})
	if err != nil {
		t.Fatalf("list public serviced apartments: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 || items[0].ServicedApartment == nil {
		t.Fatalf("expected one serviced apartment, got total=%d items=%#v", pagination.Total, items)
	}
	if len(items[0].ServicedApartment.RoomTypes) != 1 || items[0].ServicedApartment.RoomTypes[0].Name != "Studio" {
		t.Fatalf("expected room type payload, got %#v", items[0].ServicedApartment.RoomTypes)
	}
}
