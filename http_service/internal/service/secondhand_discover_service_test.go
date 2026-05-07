/*
 * Secondhand discover placement tests.
 * 1. Validate settings placement persistence and public discover output.
 * 2. Verify invalid slots and settings-level listing actions.
 */
package service

import (
	"context"
	"errors"
	"testing"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. TestDiscoverPlacementsSaveAndPublicOutput validates slot persistence and public output.
func TestDiscoverPlacementsSaveAndPublicOutput(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90001001", &communityA.ID)
	heroListingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	categoryListingID := mustCreatePublishedListingWithCategory(t, runtime, owner, "home_furniture", "public", communityA.PublicID)

	settings, err := secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneHero, SlotIndex: 1, ListingID: heroListingID},
		{Scene: discoverSceneCategory, CategoryCode: "home_furniture", SlotIndex: 1, ListingID: categoryListingID},
	})
	if err != nil {
		t.Fatalf("save discover placements: %v", err)
	}
	if len(settings.Hero) != discoverHeroSlotLimit || settings.Hero[0].Listing == nil || settings.Hero[0].Listing.ListingID != heroListingID {
		t.Fatalf("expected hero slot to include saved listing, got %#v", settings.Hero)
	}
	if len(settings.Categories["home_furniture"]) != discoverCategorySlotLimit || settings.Categories["home_furniture"][0].Listing == nil {
		t.Fatalf("expected category slot matrix with saved listing")
	}

	discover, err := secondhandService.GetPublicDiscover(context.Background())
	if err != nil {
		t.Fatalf("get public discover: %v", err)
	}
	if len(discover.Hero) != 1 || discover.Hero[0].Listing == nil || discover.Hero[0].Listing.ListingID != heroListingID {
		t.Fatalf("expected one public hero placement, got %#v", discover.Hero)
	}
	if len(discover.Categories["home_furniture"]) != 1 || discover.Categories["home_furniture"][0].Listing.ListingID != categoryListingID {
		t.Fatalf("expected one public category placement, got %#v", discover.Categories)
	}
}

// 2. TestDiscoverPlacementsClearAndInvalidInput validates clearing and validation.
func TestDiscoverPlacementsClearAndInvalidInput(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90001002", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	if _, err := secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneHero, SlotIndex: 1, ListingID: listingID},
	}); err != nil {
		t.Fatalf("save hero placement: %v", err)
	}
	settings, err := secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneHero, SlotIndex: 1, ListingID: ""},
	})
	if err != nil {
		t.Fatalf("clear hero placement: %v", err)
	}
	if settings.Hero[0].Listing != nil {
		t.Fatalf("expected cleared hero slot, got %#v", settings.Hero[0])
	}

	_, err = secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneCategory, CategoryCode: "bad_category", SlotIndex: 1, ListingID: listingID},
	})
	assertDiscoverValidationError(t, err)

	_, err = secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneCategory, CategoryCode: "home_appliance", SlotIndex: 11, ListingID: listingID},
	})
	assertDiscoverValidationError(t, err)
}

// 3. TestPublicDiscoverSkipsUnavailableListings validates public visibility filtering.
func TestPublicDiscoverSkipsUnavailableListings(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90001003", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	if _, err := secondhandService.SaveDiscoverPlacements(context.Background(), owner.ID, []DiscoverPlacementInput{
		{Scene: discoverSceneHero, SlotIndex: 1, ListingID: listingID},
	}); err != nil {
		t.Fatalf("save hero placement: %v", err)
	}
	if err := secondhandService.MarkSoldForSettings(context.Background(), listingID); err != nil {
		t.Fatalf("mark sold for settings: %v", err)
	}

	discover, err := secondhandService.GetPublicDiscover(context.Background())
	if err != nil {
		t.Fatalf("get public discover: %v", err)
	}
	if len(discover.Hero) != 0 {
		t.Fatalf("expected sold listing to be skipped, got %#v", discover.Hero)
	}
}

// 4. TestSettingsListingActions validates settings list and status actions.
func TestSettingsListingActions(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90001004", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	items, pagination, err := secondhandService.ListSettingsSecondhand(context.Background(), SecondhandSettingsListFilters{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	})
	if err != nil {
		t.Fatalf("list settings secondhand: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 || items[0].ListingID != listingID {
		t.Fatalf("expected active listing in settings list, got %#v %#v", pagination, items)
	}

	if err := secondhandService.DeactivateForSettings(context.Background(), listingID); err != nil {
		t.Fatalf("deactivate for settings: %v", err)
	}
	hiddenItems, _, err := secondhandService.ListSettingsSecondhand(context.Background(), SecondhandSettingsListFilters{
		Page:     1,
		PageSize: 20,
		Status:   "hidden",
	})
	if err != nil {
		t.Fatalf("list hidden settings secondhand: %v", err)
	}
	if len(hiddenItems) != 1 || hiddenItems[0].PublicationStatus != "hidden" {
		t.Fatalf("expected hidden listing, got %#v", hiddenItems)
	}
}

// 5. assertDiscoverValidationError validates discover placement validation errors.
func assertDiscoverValidationError(t *testing.T, err error) {
	t.Helper()

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeValidationError {
		t.Fatalf("expected validation error, got %#v", err)
	}
}
