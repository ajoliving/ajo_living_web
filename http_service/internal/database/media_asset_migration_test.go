/*
 * Media asset migration tests.
 * 1. Validate legacy duplicate media rows are merged before unique index creation.
 * 2. Validate existing media references are preserved after the merge.
 */
package database

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

type legacyMediaAsset struct {
	ID              int64  `gorm:"primaryKey;autoIncrement"`
	PublicID        string `gorm:"type:varchar(26);not null"`
	StorageProvider string `gorm:"type:varchar(32);not null"`
	BucketName      string `gorm:"type:varchar(120);not null"`
	ObjectKey       string `gorm:"type:varchar(500);not null"`
	MimeType        string `gorm:"type:varchar(100);not null"`
	Width           *int
	Height          *int
	FileSize        int64
	ChecksumSHA256  string `gorm:"type:varchar(128)"`
	CreatedBy       *int64 `gorm:"index"`
	CreatedAt       time.Time
}

// 1. TableName maps the legacy test model to media_assets.
func (legacyMediaAsset) TableName() string {
	return "media_assets"
}

// 2. TestMigrateMergesDuplicateMediaAssetsBeforeUniqueIndex validates legacy media cleanup.
func TestMigrateMergesDuplicateMediaAssetsBeforeUniqueIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	seedLegacyDuplicateMediaAssetTables(t, db)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	var count int64
	if err := db.Model(&model.MediaAsset{}).
		Where("bucket_name = ? AND object_key = ?", "ajo-living", "uploads/sample.webp").
		Count(&count).Error; err != nil {
		t.Fatalf("count media assets: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one merged media asset, got %d", count)
	}

	assertMediaAssetReference(t, db, "listing_images", "media_asset_id", 1)
	assertMediaAssetReference(t, db, "home_content_placements", "media_asset_id", 1)
	assertMediaAssetReference(t, db, "user_profiles", "avatar_asset_id", 1)

	duplicate := model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: "oss",
		BucketName:      "ajo-living",
		ObjectKey:       "uploads/sample.webp",
		MimeType:        "image/webp",
	}
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("expected duplicate media asset insert to fail")
	}
}

// 3. seedLegacyDuplicateMediaAssetTables creates a pre-index schema with duplicate media rows.
func seedLegacyDuplicateMediaAssetTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.AutoMigrate(
		&legacyMediaAsset{},
		&model.ListingImage{},
		&model.HomeContentPlacement{},
		&model.UserProfile{},
	); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}

	insertLegacyMediaAsset(t, db, 1, "uploads/sample.webp")
	insertLegacyMediaAsset(t, db, 2, "uploads/sample.webp")
	insertLegacyMediaAsset(t, db, 3, "uploads/sample.webp")
	insertLegacyMediaAsset(t, db, 4, "uploads/other.webp")

	if err := db.Create(&model.ListingImage{ListingID: 11, MediaAssetID: 2, SortOrder: 1, IsCover: true}).Error; err != nil {
		t.Fatalf("insert listing image: %v", err)
	}
	if err := db.Create(&model.HomeContentPlacement{PlacementType: "hero", SlotIndex: 1, MediaAssetID: 3}).Error; err != nil {
		t.Fatalf("insert home content placement: %v", err)
	}
	if err := db.Create(&model.UserProfile{UserID: 21, AvatarAssetID: int64Ptr(2)}).Error; err != nil {
		t.Fatalf("insert user profile: %v", err)
	}
}

// 4. insertLegacyMediaAsset inserts one legacy media asset row.
func insertLegacyMediaAsset(t *testing.T, db *gorm.DB, id int64, objectKey string) {
	t.Helper()

	asset := legacyMediaAsset{
		ID:              id,
		PublicID:        utils.NewPublicID(),
		StorageProvider: "oss",
		BucketName:      "ajo-living",
		ObjectKey:       objectKey,
		MimeType:        "image/webp",
		FileSize:        100,
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("insert legacy media asset: %v", err)
	}
}

// 5. int64Ptr returns a pointer for optional legacy references.
func int64Ptr(value int64) *int64 {
	return &value
}

// 6. assertMediaAssetReference validates duplicate references now point at the kept asset ID.
func assertMediaAssetReference(t *testing.T, db *gorm.DB, table string, column string, expectedID int64) {
	t.Helper()

	var keptCount int64
	if err := db.Table(table).Where(column+" = ?", expectedID).Count(&keptCount).Error; err != nil {
		t.Fatalf("count kept references in %s: %v", table, err)
	}
	if keptCount != 1 {
		t.Fatalf("expected one kept reference in %s, got %d", table, keptCount)
	}

	var duplicateCount int64
	if err := db.Table(table).Where(column+" IN ?", []int64{2, 3}).Count(&duplicateCount).Error; err != nil {
		t.Fatalf("count duplicate references in %s: %v", table, err)
	}
	if duplicateCount != 0 {
		t.Fatalf("expected no duplicate references in %s, got %d", table, duplicateCount)
	}
}
