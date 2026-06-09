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
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRoleBinding{},
		&model.Listing{},
		&model.ListingContact{},
		&model.MediaAsset{},
		&model.ListingImage{},
		&model.ContactAccessLog{},
		&model.DiscoverPlacement{},
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
		&model.ServicedApartmentProject{},
		&model.SupermarketFavorite{},
		&model.SupermarketPriceAlert{},
		&model.SupermarketPriceAlertEvent{},
	); err != nil {
		return err
	}

	if err := migrateDistrictCodes(db); err != nil {
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

// 4. migrateDistrictCodes folds legacy district codes into the current area tags.
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
		if err := db.Model(&model.Listing{}).Where("district_code IN ?", legacyCodes).Update("district_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate listing district codes: %w", err)
		}
		if err := db.Model(&model.Community{}).Where("district_code IN ?", legacyCodes).Update("district_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate community district codes: %w", err)
		}
		if err := db.Model(&model.UserProfile{}).Where("district_code IN ?", legacyCodes).Update("district_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate user profile district codes: %w", err)
		}
		if err := db.Model(&model.SecondhandListing{}).Where("pickup_region_code IN ?", legacyCodes).Update("pickup_region_code", currentCode).Error; err != nil {
			return fmt.Errorf("migrate secondhand pickup region codes: %w", err)
		}
	}

	return nil
}
