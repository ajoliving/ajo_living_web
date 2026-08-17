/*
 * Database bootstrap utilities.
 * 1. Open the configured database connection.
 * 2. Run schema migration and seed required baseline data.
 * 3. Keep startup database logic isolated from handlers and services.
 */
package database

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

type duplicateMediaAssetGroup struct {
	BucketName string
	ObjectKey  string
	KeepID     int64
}

// 1. Open creates a gorm database instance for the configured driver.
func Open(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBDSN)
	default:
		dialector = postgres.Open(cfg.DBDSN)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

// 2. Migrate runs the required schema migration set.
func Migrate(db *gorm.DB) error {
	if err := mergeDuplicateMediaAssets(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.BuildingAuthorization{},
		&model.Community{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRoleBinding{},
		&model.Listing{},
		&model.ListingContact{},
		&model.MediaAsset{},
		&model.AgencyProfileBinding{},
		&model.AgencyProfile{},
		&model.AgencyCompanySubaccount{},
		&model.ListingImage{},
		&model.ContactAccessLog{},
		&model.ListingFavorite{},
		&model.HomeContentPlacement{},
		&model.Order{},
		&model.OrderLog{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
		&model.ModerationAction{},
		&model.Notification{},
		&model.WalletAccount{},
		&model.WalletTransaction{},
		&model.RewardAd{},
		&model.RewardAdClaim{},
		&model.DisplayAdSlotAssignment{},
		&model.WalletRechargeOrder{},
		&model.SecondhandListing{},
		&model.PropertySaleListing{},
		&model.PropertyAddress{},
		&model.PropertyViewingAppointment{},
		&model.PropertyReport{},
		&model.ServicedApartmentProject{},
		&model.SupermarketFavorite{},
		&model.SupermarketPriceAlert{},
		&model.SupermarketPriceAlertEvent{},
		&model.SupermarketImageReport{},
	); err != nil {
		return err
	}

	if err := migrateDistrictCodeColumnSize(db); err != nil {
		return err
	}

	if err := migrateDistrictCodes(db); err != nil {
		return err
	}

	if err := migrateSecondhandOptionCodes(db); err != nil {
		return err
	}

	return nil
}

// 3. SeedCommunities inserts baseline communities when the table is empty.
func SeedCommunities(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.Community{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	communities := []model.Community{
		{
			PublicID:      utils.NewPublicID(),
			CommunityType: "estate",
			NameZH:        "康怡花園",
			NameEN:        "Kornhill",
			DistrictCode:  "hong_kong_island",
			AddressText:   "Quarry Bay",
		},
		{
			PublicID:      utils.NewPublicID(),
			CommunityType: "building",
			NameZH:        "太古城金星閣",
			NameEN:        "Taikoo Shing Venus Mansion",
			DistrictCode:  "hong_kong_island",
			AddressText:   "Taikoo Shing",
		},
		{
			PublicID:      utils.NewPublicID(),
			CommunityType: "estate",
			NameZH:        "麗港城",
			NameEN:        "Laguna City",
			DistrictCode:  "kowloon",
			AddressText:   "Lam Tin",
		},
	}

	if err := db.WithContext(ctx).Create(&communities).Error; err != nil {
		return fmt.Errorf("seed communities: %w", err)
	}

	return nil
}

// 4. mergeDuplicateMediaAssets folds legacy duplicate media rows before unique index migration.
func mergeDuplicateMediaAssets(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.MediaAsset{}) {
		return nil
	}

	var groups []duplicateMediaAssetGroup
	if err := db.Table("media_assets").
		Select("bucket_name, object_key, MIN(id) AS keep_id").
		Group("bucket_name, object_key").
		Having("COUNT(*) > 1").
		Scan(&groups).Error; err != nil {
		return fmt.Errorf("find duplicate media assets: %w", err)
	}
	if len(groups) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, group := range groups {
			var duplicateIDs []int64
			if err := tx.Table("media_assets").
				Where("bucket_name = ? AND object_key = ? AND id <> ?", group.BucketName, group.ObjectKey, group.KeepID).
				Order("id").
				Pluck("id", &duplicateIDs).Error; err != nil {
				return fmt.Errorf("load duplicate media asset ids: %w", err)
			}
			if len(duplicateIDs) == 0 {
				continue
			}
			if err := reassignMediaAssetReferences(tx, duplicateIDs, group.KeepID); err != nil {
				return err
			}
			if err := tx.Where("id IN ?", duplicateIDs).Delete(&model.MediaAsset{}).Error; err != nil {
				return fmt.Errorf("delete duplicate media assets: %w", err)
			}
		}

		return nil
	})
}

// 5. reassignMediaAssetReferences points legacy duplicate references at the kept asset row.
func reassignMediaAssetReferences(tx *gorm.DB, duplicateIDs []int64, keepID int64) error {
	references := []struct {
		model  any
		table  string
		column string
	}{
		{model: &model.ListingImage{}, table: "listing_images", column: "media_asset_id"},
		{model: &model.HomeContentPlacement{}, table: "home_content_placements", column: "media_asset_id"},
		{model: &model.UserProfile{}, table: "user_profiles", column: "avatar_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "logo_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "avatar_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "wechat_qr_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "eaa_license_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "business_registration_asset_id"},
		{model: &model.AgencyProfile{}, table: "agency_profiles", column: "company_card_asset_id"},
	}

	for _, reference := range references {
		if !tx.Migrator().HasTable(reference.model) || !tx.Migrator().HasColumn(reference.model, reference.column) {
			continue
		}
		if err := tx.Table(reference.table).
			Where(reference.column+" IN ?", duplicateIDs).
			Update(reference.column, keepID).Error; err != nil {
			return fmt.Errorf("reassign %s.%s media asset references: %w", reference.table, reference.column, err)
		}
	}

	return nil
}

// 6. migrateDistrictCodeColumnSize expands location code columns for subdistrict codes.
func migrateDistrictCodeColumnSize(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}

	statements := []string{
		"ALTER TABLE listings ALTER COLUMN district_code TYPE varchar(64)",
		"ALTER TABLE communities ALTER COLUMN district_code TYPE varchar(64)",
		"ALTER TABLE user_profiles ALTER COLUMN district_code TYPE varchar(64)",
		"ALTER TABLE property_addresses ALTER COLUMN region_code TYPE varchar(64)",
		"ALTER TABLE property_addresses ALTER COLUMN district_code TYPE varchar(64)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("expand district code column size: %w", err)
		}
	}

	return nil
}

// 7. migrateDistrictCodes folds legacy secondhand pickup district codes into area tags.
func migrateDistrictCodes(db *gorm.DB) error {
	districtGroups := map[string][]string{
		"hong_kong_island": {
			"central_western",
			"wan_chai",
			"eastern",
			"southern",
		},
		"kowloon": {
			"yau_tsim_mong",
			"sham_shui_po",
			"kowloon_city",
			"wong_tai_sin",
			"kwun_tong",
		},
		"new_territories": {
			"kwai_tsing",
			"tsuen_wan",
			"tuen_mun",
			"yuen_long",
			"north",
			"tai_po",
			"sha_tin",
			"sai_kung",
		},
		"outlying_islands": {
			"islands",
		},
	}

	for currentCode, legacyCodes := range districtGroups {
		if err := db.Model(&model.SecondhandListing{}).Where("pickup_region_code IN ?", legacyCodes).Update("pickup_region_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate secondhand pickup region codes: %w", err)
		}
	}

	return nil
}

// 8. migrateSecondhandOptionCodes folds legacy furniture option codes into current enums.
func migrateSecondhandOptionCodes(db *gorm.DB) error {
	categoryGroups := map[string][]string{
		"home_furniture": {
			"office_furniture",
			"home_decor",
		},
		"other": {
			"music",
		},
	}

	for currentCode, legacyCodes := range categoryGroups {
		if err := db.Model(&model.SecondhandListing{}).Where("category_code IN ?", legacyCodes).Update("category_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate secondhand category codes: %w", err)
		}
	}

	conditionGroups := map[string][]string{
		"used_excellent": {
			"brand_new",
		},
	}

	for currentCode, legacyCodes := range conditionGroups {
		if err := db.Model(&model.SecondhandListing{}).Where("condition_level IN ?", legacyCodes).Update("condition_level", currentCode).Error; err != nil {
			return fmt.Errorf("migrate secondhand condition codes: %w", err)
		}
	}

	return nil
}
