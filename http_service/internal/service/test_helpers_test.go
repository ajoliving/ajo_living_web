/*
 * Service test helpers.
 * 1. Build isolated sqlite-backed runtime dependencies for tests.
 * 2. Provide reusable fixtures for members, media, and listings.
 */
package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

var fixedNow = time.Date(2036, 4, 17, 12, 0, 0, 0, time.UTC)

// 1. newTestRuntime creates an isolated sqlite-backed runtime for tests.
func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()

	cfg := &config.Config{
		AppEnv:               "test",
		AppPort:              "8080",
		AppPublicBaseURL:     "http://localhost:5173",
		DBDriver:             "sqlite",
		DBDSN:                fmt.Sprintf("file:%s?mode=memory&cache=shared", utils.NewPublicID()),
		JWTSecret:            "test-secret",
		EncryptionKey:        "test-encryption-key",
		BootstrapStaffPhones: "",
		OTPProvider:          "mock",
		OTPMockCode:          "123456",
		StorageProvider:      "mock",
		StorageBucket:        "test-bucket",
		StorageEndpoint:      "http://storage.local",
		MediaBaseURL:         "http://storage.local/test-bucket",
	}

	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if err := database.SeedAccessControl(context.Background(), db); err != nil {
		t.Fatalf("seed access control: %v", err)
	}

	if err := database.SeedSystemNotificationAccount(context.Background(), db); err != nil {
		t.Fatalf("seed system notification account: %v", err)
	}

	if err := database.SeedCommunities(context.Background(), db); err != nil {
		t.Fatalf("seed communities: %v", err)
	}

	storageProvider, err := NewStorageProvider(cfg)
	if err != nil {
		t.Fatalf("create storage provider: %v", err)
	}

	return &Runtime{
		Config:          cfg,
		DB:              db,
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		OTPProvider:     NewOTPProvider(cfg),
		MailSender:      NewMailSender(cfg),
		StorageProvider: storageProvider,
		OTPStore:        NewOTPStore(),
		Now:             func() time.Time { return fixedNow },
	}
}

// 2. mustGetCommunities returns at least two seeded communities.
func mustGetCommunities(t *testing.T, runtime *Runtime) (model.Community, model.Community) {
	t.Helper()

	var communities []model.Community
	if err := runtime.DB.Order("id asc").Find(&communities).Error; err != nil {
		t.Fatalf("load communities: %v", err)
	}
	if len(communities) < 2 {
		t.Fatalf("expected at least two communities, got %d", len(communities))
	}

	return communities[0], communities[1]
}

// 3. mustCreateUser creates a member with a selected primary community.
func mustCreateUser(t *testing.T, runtime *Runtime, countryCode string, phone string, communityID *int64) model.User {
	t.Helper()

	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: countryCode,
		PhoneNumber:      phone,
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsStaff:          false,
		IsVerifiedPhone:  true,
	}
	if err := runtime.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	profile := model.UserProfile{
		UserID:                user.ID,
		DisplayName:           "Tester " + phone,
		PublisherIdentityType: "owner",
		PrimaryCommunityID:    communityID,
	}
	if err := runtime.DB.Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	accessService := NewAccessService(runtime)
	if err := accessService.EnsureUserRoles(context.Background(), runtime.DB, &user, nil, ""); err != nil {
		t.Fatalf("ensure user roles: %v", err)
	}

	return user
}

// 4. mustCreateMediaAsset creates a media asset owned by the target user.
func mustCreateMediaAsset(t *testing.T, runtime *Runtime, userID int64) string {
	t.Helper()

	uploadService := NewUploadService(runtime)
	result, err := uploadService.CompleteUpload(context.Background(), CompleteUploadParams{
		UserID:         userID,
		ObjectKey:      mediaObjectPrefix + "test-cover.webp",
		MimeType:       "image/webp",
		FileSize:       1024,
		ChecksumSHA256: "checksum",
	})
	if err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	return result.MediaAssetID
}

// 5. mustCreatePublishedListing creates and publishes a secondhand listing.
func mustCreatePublishedListing(t *testing.T, runtime *Runtime, owner model.User, visibilityScope string, communityPublicID string) string {
	t.Helper()

	return mustCreatePublishedListingWithCategory(t, runtime, owner, "home_appliance", visibilityScope, communityPublicID)
}

// 6. mustCreatePublishedListingWithCategory creates a published secondhand listing with a category.
func mustCreatePublishedListingWithCategory(t *testing.T, runtime *Runtime, owner model.User, categoryCode string, visibilityScope string, communityPublicID string) string {
	t.Helper()

	secondhandService := NewSecondhandService(runtime)
	assetID := mustCreateMediaAsset(t, runtime, owner.ID)

	created, err := secondhandService.CreateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		Title:                 "Ninety Percent New Washer",
		Summary:               "Clean and ready",
		Description:           "Well maintained home appliance.",
		DistrictCode:          "kwun_tong",
		CommunityID:           communityPublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          categoryCode,
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(1200),
		ConditionLevel:        "used_good",
		DimensionText:         "60 x 60 x 85 cm",
		PickupRegionCode:      "kwun_tong",
		PickupLocationText:    "Lobby pickup",
		DeliveryTags:          []string{"self_pickup", "elevator"},
		VisibilityScope:       visibilityScope,
		ContactMethod:         "both",
		Images: []ListingImageInput{
			{MediaAssetID: assetID, SortOrder: 1, IsCover: true},
		},
		Contact: ListingContactInput{
			WhatsApp:     "+85291234567",
			ShowWhatsApp: true,
			ShowChat:     true,
		},
	})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}

	published, err := secondhandService.PublishSecondhandListing(context.Background(), owner.ID, created.ListingID)
	if err != nil {
		t.Fatalf("publish listing: %v", err)
	}

	return published.ListingID
}

// 7. ptrFloat64 returns a float64 pointer.
func ptrFloat64(value float64) *float64 {
	return &value
}
