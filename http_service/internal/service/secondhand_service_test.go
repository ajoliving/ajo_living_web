/*
 * Secondhand service tests.
 * 1. Validate building-only visibility rules and contact access.
 * 2. Verify lifecycle expiration updates the listing status.
 */
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. TestBuildingOnlyVisibilityAndContactAccess validates visibility enforcement.
func TestBuildingOnlyVisibilityAndContactAccess(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, communityB := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000001", &communityA.ID)
	viewer := mustCreateUser(t, runtime, "+852", "90000002", &communityA.ID)
	outsider := mustCreateUser(t, runtime, "+852", "90000003", &communityB.ID)

	listingID := mustCreatePublishedListing(t, runtime, owner, "building_only", communityA.PublicID)

	if _, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, &viewer.ID, viewerProfileCommunityID(&viewer, runtime)); err != nil {
		t.Fatalf("expected same-community viewer to access detail: %v", err)
	}

	_, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, &outsider.ID, viewerProfileCommunityID(&outsider, runtime))
	if err == nil {
		t.Fatalf("expected outsider detail access to fail")
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeVisibilityForbidden {
		t.Fatalf("expected visibility forbidden, got %#v", err)
	}

	access, err := secondhandService.GrantContactAccess(context.Background(), viewer.ID, viewerProfileCommunityID(&viewer, runtime), listingID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("contact access: %v", err)
	}
	if !access.AllowedChannels["whatsapp"] || !access.AllowedChannels["chat"] {
		t.Fatalf("expected whatsapp and chat channels to be granted")
	}
	if access.ContactPayload["whatsapp_url"] == "" {
		t.Fatalf("expected whatsapp url to be returned")
	}
}

// 2. TestSecondhandResponseIncludesDisplayFields validates list and detail display fields.
func TestSecondhandResponseIncludesDisplayFields(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000021", &communityA.ID)
	avatarAssetPublicID := mustCreateMediaAsset(t, runtime, owner.ID)
	var avatarAsset model.MediaAsset
	if err := runtime.DB.Where("public_id = ?", avatarAssetPublicID).First(&avatarAsset).Error; err != nil {
		t.Fatalf("load avatar asset: %v", err)
	}
	if err := runtime.DB.Model(&model.UserProfile{}).
		Where("user_id = ?", owner.ID).
		Update("avatar_asset_id", avatarAsset.ID).Error; err != nil {
		t.Fatalf("update owner avatar: %v", err)
	}
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	items, _, err := secondhandService.ListPublicSecondhand(context.Background(), SecondhandListFilters{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public secondhand: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one listing, got %d", len(items))
	}
	if items[0].PublishedAt == nil || *items[0].PublishedAt == "" {
		t.Fatalf("expected published_at in list response")
	}
	if items[0].Community == nil || items[0].Community.PublicID != communityA.PublicID {
		t.Fatalf("expected community info in list response")
	}
	if items[0].Owner == nil || items[0].Owner.PublicID != owner.PublicID || items[0].Owner.DisplayName == "" || items[0].Owner.AvatarURL == "" {
		t.Fatalf("expected owner display info in list response")
	}

	detail, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, nil, nil)
	if err != nil {
		t.Fatalf("get secondhand detail: %v", err)
	}
	if detail.PublishedAt == nil || *detail.PublishedAt == "" {
		t.Fatalf("expected published_at in detail response")
	}
	if detail.Community == nil || detail.Community.PublicID != communityA.PublicID {
		t.Fatalf("expected community info in detail response")
	}
	if detail.Owner == nil || detail.Owner.PublicID != owner.PublicID || detail.Owner.DisplayName == "" || detail.Owner.AvatarURL == "" {
		t.Fatalf("expected owner display info in detail response")
	}
}

// 3. TestExpireListings updates overdue listings to expired.
func TestExpireListings(t *testing.T) {
	runtime := newTestRuntime(t)
	lifecycleService := NewLifecycleService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000011", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	secondhandService := NewSecondhandService(runtime)
	listing, _, _, err := secondhandService.loadListingByPublicID(context.Background(), listingID)
	if err != nil {
		t.Fatalf("load listing: %v", err)
	}

	past := fixedNow.Add(-2 * time.Hour)
	if err := runtime.DB.Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(map[string]any{
		"expire_at": past,
	}).Error; err != nil {
		t.Fatalf("update expire_at: %v", err)
	}

	rows, err := lifecycleService.ExpireListings(context.Background())
	if err != nil {
		t.Fatalf("expire listings: %v", err)
	}
	if rows < 1 {
		t.Fatalf("expected at least one expired listing")
	}

	var refreshed model.Listing
	if err := runtime.DB.First(&refreshed, listing.ID).Error; err != nil {
		t.Fatalf("reload listing: %v", err)
	}
	if refreshed.PublicationStatus != "expired" {
		t.Fatalf("expected publication_status=expired, got %s", refreshed.PublicationStatus)
	}
}

// 4. TestOwnerLifecycleActions validates owner status management flows.
func TestOwnerLifecycleActions(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000031", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	if err := secondhandService.MarkSold(context.Background(), owner.ID, listingID); err != nil {
		t.Fatalf("mark sold: %v", err)
	}

	soldItems, _, err := secondhandService.ListMySecondhand(context.Background(), owner.ID, MySecondhandFilters{Page: 1, PageSize: 20, Status: "sold"})
	if err != nil {
		t.Fatalf("list sold listings: %v", err)
	}
	if len(soldItems) != 1 || soldItems[0].BusinessStatus != "sold" {
		t.Fatalf("expected one sold listing, got %#v", soldItems)
	}

	publicItems, _, err := secondhandService.ListPublicSecondhand(context.Background(), SecondhandListFilters{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public listings: %v", err)
	}
	if len(publicItems) != 0 {
		t.Fatalf("expected sold listing hidden from public discovery, got %d", len(publicItems))
	}
}

// 5. TestSecondhandFavorites validates saved listing lifecycle.
func TestSecondhandFavorites(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000041", &communityA.ID)
	viewer := mustCreateUser(t, runtime, "+852", "90000042", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	favorite, err := secondhandService.AddFavoriteSecondhand(context.Background(), viewer.ID, viewerProfileCommunityID(&viewer, runtime), listingID)
	if err != nil {
		t.Fatalf("add favorite: %v", err)
	}
	if favorite.ListingID != listingID || !favorite.IsFavorited {
		t.Fatalf("expected favorite result, got %#v", favorite)
	}

	items, pagination, err := secondhandService.ListFavoriteSecondhand(context.Background(), viewer.ID, viewerProfileCommunityID(&viewer, runtime), MySecondhandFilters{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list favorites: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 || !items[0].IsFavorited {
		t.Fatalf("expected one favorited listing, got items=%#v pagination=%#v", items, pagination)
	}

	publicItems, _, err := secondhandService.ListPublicSecondhand(context.Background(), SecondhandListFilters{
		Page:         1,
		PageSize:     20,
		ViewerUserID: &viewer.ID,
	})
	if err != nil {
		t.Fatalf("list public secondhand: %v", err)
	}
	if len(publicItems) != 1 || !publicItems[0].IsFavorited {
		t.Fatalf("expected public item to carry favorite state, got %#v", publicItems)
	}

	removed, err := secondhandService.RemoveFavoriteSecondhand(context.Background(), viewer.ID, listingID)
	if err != nil {
		t.Fatalf("remove favorite: %v", err)
	}
	if removed.IsFavorited {
		t.Fatalf("expected removed favorite state, got %#v", removed)
	}
}

// 6. TestRepublishExpiredListing validates owner renewal for expired listings.
func TestRepublishExpiredListing(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000032", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	mustGrantPoints(t, runtime, owner.ID, 100)

	listing, _, _, err := secondhandService.loadListingByPublicID(context.Background(), listingID)
	if err != nil {
		t.Fatalf("load listing: %v", err)
	}
	if err := runtime.DB.Model(&model.Listing{}).Where("id = ?", listing.ID).Update("publication_status", "expired").Error; err != nil {
		t.Fatalf("expire listing: %v", err)
	}

	republished, err := secondhandService.RepublishSecondhandListing(context.Background(), owner.ID, listingID)
	if err != nil {
		t.Fatalf("republish listing: %v", err)
	}
	if republished.PublicationStatus != "active" || republished.ExpireAt == nil {
		t.Fatalf("expected active republished listing, got %#v", republished)
	}
}

// 7. TestUpdateSecondhandListingRefreshesEditableFields validates edit flow persistence.
func TestUpdateSecondhandListingRefreshesEditableFields(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000033", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	mustGrantPoints(t, runtime, owner.ID, 100)
	assetID := mustCreateMediaAsset(t, runtime, owner.ID)

	if err := secondhandService.MarkSold(context.Background(), owner.ID, listingID); err != nil {
		t.Fatalf("mark listing sold before update: %v", err)
	}

	updated, err := secondhandService.UpdateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		ListingPublicID:       listingID,
		Title:                 "Updated Dining Table",
		Summary:               "Updated summary",
		Description:           "Updated detail description.",
		DistrictCode:          "kwun_tong",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          "home_furniture",
		PriceMode:             "negotiable",
		PriceHKD:              ptrFloat64(800),
		ConditionLevel:        "used_excellent",
		DimensionText:         "120 x 80 cm",
		PickupRegionCode:      "kwun_tong",
		PickupLocationText:    "Clubhouse entrance",
		DeliveryTags:          []string{"self_pickup"},
		VisibilityScope:       "public",
		ContactMethod:         "chat",
		BusinessStatus:        "available",
		Images: []ListingImageInput{
			{MediaAssetID: assetID, SortOrder: 1, IsCover: true},
		},
		Contact: ListingContactInput{
			ShowChat: true,
		},
	})
	if err != nil {
		t.Fatalf("update listing: %v", err)
	}
	if updated.Title != "Updated Dining Table" || updated.CategoryCode != "home_furniture" || updated.PriceMode != "negotiable" {
		t.Fatalf("expected updated fields, got %#v", updated)
	}
	if updated.BusinessStatus != "available" {
		t.Fatalf("expected listing to be available after update, got %s", updated.BusinessStatus)
	}
	if len(updated.Images) != 1 || updated.Images[0].MediaAssetID != assetID {
		t.Fatalf("expected updated image, got %#v", updated.Images)
	}
}

// 8. TestCreateSecondhandDraftChargesHalfPublishCost validates paid draft creation.
func TestCreateSecondhandDraftChargesHalfPublishCost(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000035", &communityA.ID)

	mustGrantPoints(t, runtime, owner.ID, 50)
	created, err := secondhandService.CreateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		Title:                 "Draft Chair",
		Summary:               "Draft summary",
		Description:           "Draft detail description.",
		DistrictCode:          "eastern",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          "home_furniture",
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(100),
		ConditionLevel:        "used_good",
		PickupRegionCode:      "eastern",
		PickupLocationText:    "Lobby",
		DeliveryTags:          []string{"self_pickup"},
		VisibilityScope:       "public",
		ContactMethod:         "chat",
		ChargeDraftSave:       true,
		Contact: ListingContactInput{
			ShowChat: true,
		},
	})
	if err != nil {
		t.Fatalf("create draft listing: %v", err)
	}
	if created.PointsCharged != 50 || created.PointsBalanceAfter == nil || *created.PointsBalanceAfter != 0 {
		t.Fatalf("expected half publish draft creation charge, got %#v", created)
	}
}

// 9. TestDraftSecondhandSaveChargesHalfPublishCost validates paid draft saves.
func TestDraftSecondhandSaveChargesHalfPublishCost(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000036", &communityA.ID)
	created, err := secondhandService.CreateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		Title:                 "Draft Chair",
		Summary:               "Draft summary",
		Description:           "Draft detail description.",
		DistrictCode:          "eastern",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          "home_furniture",
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(100),
		ConditionLevel:        "used_good",
		PickupRegionCode:      "eastern",
		PickupLocationText:    "Lobby",
		DeliveryTags:          []string{"self_pickup"},
		VisibilityScope:       "public",
		ContactMethod:         "chat",
		Contact: ListingContactInput{
			ShowChat: true,
		},
	})
	if err != nil {
		t.Fatalf("create draft listing: %v", err)
	}
	mustGrantPoints(t, runtime, owner.ID, 50)

	updated, err := secondhandService.UpdateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		ListingPublicID:       created.ListingID,
		Title:                 "Draft Chair Updated",
		Summary:               "Draft summary",
		Description:           "Draft detail description.",
		DistrictCode:          "eastern",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          "home_furniture",
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(100),
		ConditionLevel:        "used_good",
		PickupRegionCode:      "eastern",
		PickupLocationText:    "Lobby",
		DeliveryTags:          []string{"self_pickup"},
		VisibilityScope:       "public",
		ContactMethod:         "chat",
		ChargeDraftSave:       true,
		Contact: ListingContactInput{
			ShowChat: true,
		},
	})
	if err != nil {
		t.Fatalf("update draft listing: %v", err)
	}
	if updated.PointsCharged != 50 || updated.PointsBalanceAfter == nil || *updated.PointsBalanceAfter != 0 {
		t.Fatalf("expected half publish draft charge, got %#v", updated)
	}
}

// 10. TestRenewSecondhandListingChargesHalfPublishCost validates member renewal charge.
func TestRenewSecondhandListingChargesHalfPublishCost(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000037", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	mustGrantPoints(t, runtime, owner.ID, 50)

	renewed, err := secondhandService.RenewSecondhandListing(context.Background(), owner.ID, listingID)
	if err != nil {
		t.Fatalf("renew listing: %v", err)
	}
	if renewed.PointsCharged != 50 || renewed.PointsBalanceAfter == nil || *renewed.PointsBalanceAfter != 900 {
		t.Fatalf("expected renewal charge, got %#v", renewed)
	}
	if renewed.PublicationStatus != "active" || renewed.BusinessStatus != "available" {
		t.Fatalf("expected active available listing, got %#v", renewed)
	}

	listing, _, _, err := secondhandService.loadListingByPublicID(context.Background(), listingID)
	if err != nil {
		t.Fatalf("load renewed listing: %v", err)
	}
	expectedExpireAt := fixedNow.Add(14 * 24 * time.Hour)
	if listing.ExpireAt == nil || !listing.ExpireAt.Equal(expectedExpireAt) {
		t.Fatalf("expected expire_at %s, got %#v", expectedExpireAt, listing.ExpireAt)
	}
}

// 11. TestRenewSecondhandListingRejectsExpired validates member renewal state guard.
func TestRenewSecondhandListingRejectsExpired(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000038", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	mustGrantPoints(t, runtime, owner.ID, 50)

	if err := runtime.DB.Model(&model.Listing{}).Where("public_id = ?", listingID).Update("publication_status", "expired").Error; err != nil {
		t.Fatalf("expire listing: %v", err)
	}

	if _, err := secondhandService.RenewSecondhandListing(context.Background(), owner.ID, listingID); err == nil {
		t.Fatalf("expected renewal to reject expired listing")
	}
}

// 12. TestStaffRenewSecondhandListingUsesCustomDays validates staff renewal duration.
func TestStaffRenewSecondhandListingUsesCustomDays(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000034", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	if err := secondhandService.RenewForStaff(context.Background(), listingID, 21); err != nil {
		t.Fatalf("staff renew listing: %v", err)
	}

	listing, _, _, err := secondhandService.loadListingByPublicID(context.Background(), listingID)
	if err != nil {
		t.Fatalf("load renewed listing: %v", err)
	}
	expectedExpireAt := fixedNow.Add(21 * 24 * time.Hour)
	if listing.ExpireAt == nil || !listing.ExpireAt.Equal(expectedExpireAt) {
		t.Fatalf("expected expire_at %s, got %#v", expectedExpireAt, listing.ExpireAt)
	}
	if listing.PublicationStatus != "active" || listing.BusinessStatus != "available" {
		t.Fatalf("expected active available listing, got %#v", listing)
	}
}

// 10. viewerProfileCommunityID returns the viewer primary community ID.
func viewerProfileCommunityID(user *model.User, runtime *Runtime) *int64 {
	var profile model.UserProfile
	if err := runtime.DB.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		return nil
	}
	return profile.PrimaryCommunityID
}
