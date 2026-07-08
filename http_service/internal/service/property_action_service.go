/*
 * Property sale action service.
 * 1. Manage favorites, similar listings, appointments, reports, and counters.
 * 2. Keep public interaction writes scoped to the property sale channel.
 */
package service

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. AddFavoriteProperty saves one active property sale listing.
func (s *PropertyService) AddFavoriteProperty(ctx context.Context, userID int64, listingPublicID string) (*PropertyActionResult, error) {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, PropertyChannelSale, strings.TrimSpace(listingPublicID))
	if err != nil {
		return nil, err
	}
	if err := validatePublicPropertySale(listing); err != nil {
		return nil, err
	}

	favorite := model.ListingFavorite{UserID: userID, ListingID: listing.ID}
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ? AND listing_id = ?", userID, listing.ID).FirstOrCreate(&favorite).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save favorite listing")
	}

	return &PropertyActionResult{ListingID: listing.PublicID, IsFavorite: true}, nil
}

// 2. RemoveFavoriteProperty removes one saved property sale listing.
func (s *PropertyService) RemoveFavoriteProperty(ctx context.Context, userID int64, listingPublicID string) (*PropertyActionResult, error) {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, PropertyChannelSale, strings.TrimSpace(listingPublicID))
	if err != nil {
		return nil, err
	}
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ? AND listing_id = ?", userID, listing.ID).Delete(&model.ListingFavorite{}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to remove favorite listing")
	}

	return &PropertyActionResult{ListingID: listing.PublicID, IsFavorite: false}, nil
}

// 3. ListFavoriteProperties returns member favorite property sale listings.
func (s *PropertyService) ListFavoriteProperties(ctx context.Context, userID int64, filters PropertyListFilters) ([]PropertyListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.basePropertyListQuery(ctx, PropertyChannelSale).
		Joins("JOIN listing_favorites ON listing_favorites.listing_id = listings.id").
		Where("listing_favorites.user_id = ? AND listings.module = ? AND listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND listings.is_deleted = ?", userID, string(PropertyChannelSale), "active", "approved", "available", false)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count favorite properties")
	}

	var rows []propertyListingRow
	if err := baseQuery.Order("listing_favorites.created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load favorite properties")
	}
	items, err := s.buildPropertySummaries(ctx, PropertyChannelSale, rows)
	if err != nil {
		return nil, nil, err
	}
	for index := range items {
		items[index].IsFavorite = true
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 4. ListSimilarProperties returns related public sale listings.
func (s *PropertyService) ListSimilarProperties(ctx context.Context, listingPublicID string, limit int) ([]PropertyListingSummary, error) {
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	listing, _, err := s.loadPropertyListingByPublicID(ctx, PropertyChannelSale, strings.TrimSpace(listingPublicID))
	if err != nil {
		return nil, err
	}
	if err := validatePublicPropertySale(listing); err != nil {
		return nil, err
	}
	rows, err := s.loadPropertyRowsByIDs(ctx, PropertyChannelSale, []int64{listing.ID})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []PropertyListingSummary{}, nil
	}

	baseQuery := s.basePropertyListQuery(ctx, PropertyChannelSale).
		Where("listings.id <> ? AND listings.module = ? AND listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND listings.is_deleted = ?", listing.ID, string(PropertyChannelSale), "active", "approved", "available", false).
		Where("(listings.district_code = ? OR property_sale_listings.property_type = ?)", rows[0].DistrictCode, rows[0].SalePropertyType)

	var resultRows []propertyListingRow
	if err := baseQuery.Order("property_sale_listings.ad_weight desc, listings.sort_refreshed_at desc NULLS LAST, listings.created_at desc").Limit(limit).Scan(&resultRows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load similar properties")
	}

	return s.buildPropertySummaries(ctx, PropertyChannelSale, resultRows)
}

// 5. CreatePropertyAppointment stores one viewing request.
func (s *PropertyService) CreatePropertyAppointment(ctx context.Context, params PropertyAppointmentParams) (*PropertyAppointmentResponse, error) {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, PropertyChannelSale, strings.TrimSpace(params.ListingPublicID))
	if err != nil {
		return nil, err
	}
	if err := validatePublicPropertySale(listing); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.ContactName) == "" || strings.TrimSpace(params.ContactPhone) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "contact name and phone are required")
	}
	appointmentType := strings.TrimSpace(params.AppointmentType)
	if appointmentType == "" {
		appointmentType = "viewing"
	}
	appointment := model.PropertyViewingAppointment{
		PublicID:        utils.NewPublicID(),
		ListingID:       listing.ID,
		RequestUserID:   params.UserID,
		ContactName:     strings.TrimSpace(params.ContactName),
		ContactPhone:    strings.TrimSpace(params.ContactPhone),
		PreferredTime:   strings.TrimSpace(params.PreferredTime),
		Message:         strings.TrimSpace(params.Message),
		AppointmentType: appointmentType,
		Status:          "pending",
	}
	if err := s.runtime.DB.WithContext(ctx).Create(&appointment).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create viewing appointment")
	}
	_ = s.incrementPropertySaleCounter(ctx, listing.ID, "inquiry_count")

	return &PropertyAppointmentResponse{AppointmentID: appointment.PublicID, ListingID: listing.PublicID, Status: appointment.Status}, nil
}

// 6. ReportProperty stores one public listing report.
func (s *PropertyService) ReportProperty(ctx context.Context, params PropertyReportParams) (*PropertyReportResponse, error) {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, PropertyChannelSale, strings.TrimSpace(params.ListingPublicID))
	if err != nil {
		return nil, err
	}
	if err := validatePublicPropertySale(listing); err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(params.Reason)
	if reason == "" {
		return nil, errcode.New(errcode.CodeValidationError, "report reason is required")
	}
	report := model.PropertyReport{
		PublicID:     utils.NewPublicID(),
		ListingID:    listing.ID,
		ReporterID:   params.UserID,
		Reason:       reason,
		Message:      strings.TrimSpace(params.Message),
		ReviewStatus: "pending",
	}
	if err := s.runtime.DB.WithContext(ctx).Create(&report).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to submit property report")
	}

	return &PropertyReportResponse{ReportID: report.PublicID, ListingID: listing.PublicID, ReviewStatus: report.ReviewStatus}, nil
}

// 7. incrementPropertySaleCounter increments a public property counter.
func (s *PropertyService) incrementPropertySaleCounter(ctx context.Context, listingID int64, column string) error {
	if column != "view_count" && column != "inquiry_count" {
		return nil
	}
	return s.runtime.DB.WithContext(ctx).Model(&model.PropertySaleListing{}).
		Where("listing_id = ?", listingID).
		UpdateColumn(column, gorm.Expr(column+" + ?", 1)).
		Error
}

// 8. isFavoriteProperty checks whether the member saved a listing.
func (s *PropertyService) isFavoriteProperty(ctx context.Context, userID int64, listingID int64) bool {
	var favorite model.ListingFavorite
	err := s.runtime.DB.WithContext(ctx).Where("user_id = ? AND listing_id = ?", userID, listingID).First(&favorite).Error
	return err == nil
}

// 9. validatePublicPropertySale checks public listing availability.
func validatePublicPropertySale(listing *model.Listing) error {
	if listing == nil {
		return errcode.New(errcode.CodeNotFound, "listing not found")
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" || listing.IsDeleted {
		return errcode.New(errcode.CodeNotFound, "listing not found")
	}
	return nil
}

// 10. normalizePropertyRenovationType keeps supported renovation values.
func normalizePropertyRenovationType(value string, featureTags []string) string {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case "brand_new", "renovated", "simple", "special":
		return normalized
	}
	for _, tag := range featureTags {
		switch strings.TrimSpace(tag) {
		case "brand_new", "renovated", "special":
			return strings.TrimSpace(tag)
		}
	}
	return ""
}

// 11. normalizeCoordinate keeps coordinates within valid range.
func normalizeCoordinate(value *float64, min float64, max float64) *float64 {
	if value == nil {
		return nil
	}
	if *value < min || *value > max {
		return nil
	}
	return value
}
