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
	return db.AutoMigrate(
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
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
		&model.Order{},
		&model.OrderLog{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
		&model.ModerationAction{},
		&model.Notification{},
		&model.SecondhandListing{},
		&model.PropertySaleListing{},
		&model.ServicedApartmentProject{},
	)
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
			DistrictCode:  "hk_east",
			AddressText:   "Quarry Bay",
		},
		{
			PublicID:      utils.NewPublicID(),
			CommunityType: "building",
			NameZH:        "太古城金星閣",
			NameEN:        "Taikoo Shing Venus Mansion",
			DistrictCode:  "hk_east",
			AddressText:   "Taikoo Shing",
		},
		{
			PublicID:      utils.NewPublicID(),
			CommunityType: "estate",
			NameZH:        "麗港城",
			NameEN:        "Laguna City",
			DistrictCode:  "kwun_tong",
			AddressText:   "Lam Tin",
		},
	}

	if err := db.WithContext(ctx).Create(&communities).Error; err != nil {
		return fmt.Errorf("seed communities: %w", err)
	}

	return nil
}
