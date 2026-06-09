/*
 * Supermarket offer service types.
 * 1. Define query, favorite, and price alert payloads.
 * 2. Keep AJO-owned member payloads separate from good-price raw data.
 * 3. Provide small helpers for normalizing product and alert fields.
 */
package service

import (
	"time"

	"ajoliving_web/http_service/internal/model"
)

// 1. SupermarketSearchFilters defines public product search filters.
type SupermarketSearchFilters struct {
	Query     string
	Category  string
	Brand     string
	Store     string
	OfferOnly bool
	Sort      string
	Page      int
	PageSize  int
}

// 2. SupermarketPriceAlertInput defines a create or replace alert request.
type SupermarketPriceAlertInput struct {
	ProductCode   string
	TargetPrice   *float64
	PriceMode     string
	OfferRequired bool
	Enabled       *bool
}

// 3. SupermarketPriceAlertUpdateInput defines a partial alert update request.
type SupermarketPriceAlertUpdateInput struct {
	TargetPrice   *float64
	PriceMode     string
	OfferRequired *bool
	Enabled       *bool
}

// 4. SupermarketPriceAlertView is the member-facing alert payload.
type SupermarketPriceAlertView struct {
	ID              int64      `json:"id"`
	ProductCode     string     `json:"productCode"`
	ProductName     string     `json:"productName"`
	TargetPrice     *float64   `json:"targetPrice,omitempty"`
	PriceMode       string     `json:"priceMode"`
	OfferRequired   bool       `json:"offerRequired"`
	Enabled         bool       `json:"enabled"`
	LastTriggeredAt *time.Time `json:"lastTriggeredAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// 5. supermarketStoreSnapshot stores one current store price row from good-price.
type supermarketStoreSnapshot struct {
	Store              string
	ListPrice          float64
	EffectiveUnitPrice float64
	Offer              string
	SnapshotDate       string
}

// 6. supermarketAlertMatch stores one matched alert candidate.
type supermarketAlertMatch struct {
	Store        string
	Price        float64
	Offer        string
	SnapshotDate string
}

// 7. supermarketCacheEntry stores one cached good-price JSON response.
type supermarketCacheEntry struct {
	expiresAt time.Time
	payload   []byte
}

// 8. supermarketAlertView maps a database rule to API payload.
func supermarketAlertView(rule model.SupermarketPriceAlert) SupermarketPriceAlertView {
	return SupermarketPriceAlertView{
		ID:              rule.ID,
		ProductCode:     rule.ProductCode,
		ProductName:     rule.ProductName,
		TargetPrice:     rule.TargetPrice,
		PriceMode:       rule.PriceMode,
		OfferRequired:   rule.OfferRequired,
		Enabled:         rule.Enabled,
		LastTriggeredAt: rule.LastTriggeredAt,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       rule.UpdatedAt,
	}
}
