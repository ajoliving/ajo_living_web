/*
 * Supermarket offer business service.
 * 1. Proxy public good-price summary, search, and product detail data.
 * 2. Store AJO member favorites without using the good-price account system.
 * 3. Attach member favorite and alert state to product detail responses.
 */
package service

import (
	"context"
	"net/url"
	"strings"

	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. Summary returns cached public good-price summary data.
func (s *SupermarketOfferService) Summary(ctx context.Context) (map[string]any, error) {
	result, err := s.fetchGoodPriceJSON(ctx, "/summary", nil, true)
	if err != nil {
		return supermarketUnavailableSummary(), nil
	}

	s.attachSupermarketImages(result)
	return result, nil
}

// 2. Search returns cached public good-price product search data.
func (s *SupermarketOfferService) Search(ctx context.Context, filters SupermarketSearchFilters) (map[string]any, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := url.Values{}
	query.Set("q", strings.TrimSpace(filters.Query))
	query.Set("category", strings.TrimSpace(filters.Category))
	query.Set("brand", strings.TrimSpace(filters.Brand))
	query.Set("store", strings.TrimSpace(filters.Store))
	query.Set("offerOnly", boolString(filters.OfferOnly))
	query.Set("sort", normalizeSupermarketSort(filters.Sort))
	query.Set("page", intString(page))
	query.Set("pageSize", intString(pageSize))

	result, err := s.fetchGoodPriceJSON(ctx, "/search", query, true)
	if err != nil {
		return supermarketUnavailableSearch(page, pageSize), nil
	}

	s.attachSupermarketImages(result)
	return result, nil
}

// 3.1 attachSupermarketImages injects canonical image URLs into product payloads.
func (s *SupermarketOfferService) attachSupermarketImages(payload map[string]any) {
	if payload == nil {
		return
	}

	baseURL := strings.TrimSpace(s.runtime.Config.MediaBaseURL)
	attachProductImage := func(product map[string]any) {
		if product == nil {
			return
		}
		code := stringFromMap(product, "code")
		if code == "" {
			return
		}

		imageURL := buildMediaURL(baseURL, supermarketImageObjectKey(code))
		if imageURL == "" {
			return
		}
		product["image_url"] = imageURL
		product["imageUrl"] = imageURL
	}

	if product, _ := payload["product"].(map[string]any); product != nil {
		attachProductImage(product)
	}
	attachProductSlice := func(key string) {
		items, _ := payload[key].([]any)
		if len(items) == 0 {
			return
		}
		for _, raw := range items {
			item, _ := raw.(map[string]any)
			attachProductImage(item)
		}
	}

	attachProductSlice("sameBrand")
	attachProductSlice("sameCategory")
	attachProductSlice("items")
	attachProductSlice("favorites")
	attachProductSlice("cheapest")
	attachProductSlice("bestDiscounts")
	attachProductSlice("biggestDiffs")
	attachProductSlice("offers")
}

// 3. ProductDetail returns one good-price product detail with optional AJO member state.
func (s *SupermarketOfferService) ProductDetail(ctx context.Context, code string, days int, userID *int64) (map[string]any, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errcode.New(errcode.CodeValidationError, "product code is required")
	}
	if days <= 0 || days > 365 {
		days = 90
	}

	query := url.Values{}
	query.Set("days", intString(days))
	detail, err := s.fetchGoodPriceJSON(ctx, "/products/"+url.PathEscape(code), query, true)
	if err != nil {
		return nil, err
	}
	s.attachSupermarketImages(detail)

	isFavorite := false
	var rule *SupermarketPriceAlertView
	if userID != nil && *userID > 0 {
		isFavorite, _ = s.isFavorite(ctx, *userID, code)
		if currentRule, findErr := s.latestAlertRule(ctx, *userID, code); findErr == nil && currentRule != nil {
			view := supermarketAlertView(*currentRule)
			rule = &view
		}
	}
	detail["isFavorite"] = isFavorite
	if rule != nil {
		detail["alertRule"] = rule
	} else {
		detail["alertRule"] = nil
	}

	return detail, nil
}

// 4. ListFavorites returns AJO member favorites with latest good-price product snapshots.
func (s *SupermarketOfferService) ListFavorites(ctx context.Context, userID int64, page int, pageSize int) ([]any, *model.Pagination, error) {
	page, pageSize = normalizePagination(page, pageSize)
	db := s.runtime.DB.WithContext(ctx).Model(&model.SupermarketFavorite{}).Where("user_id = ?", userID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count supermarket favorites")
	}

	var favorites []model.SupermarketFavorite
	if err := db.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&favorites).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load supermarket favorites")
	}

	items := make([]any, 0, len(favorites))
	for _, favorite := range favorites {
		detail, err := s.ProductDetail(ctx, favorite.ProductCode, 1, &userID)
		if err != nil {
			items = append(items, fallbackFavoriteProduct(favorite, s.runtime.Config.MediaBaseURL))
			continue
		}
		product, _ := detail["product"].(map[string]any)
		if product == nil {
			items = append(items, fallbackFavoriteProduct(favorite, s.runtime.Config.MediaBaseURL))
			continue
		}
		product["isFavorite"] = true
		items = append(items, product)
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 5. AddFavorite saves one good-price product for an AJO member.
func (s *SupermarketOfferService) AddFavorite(ctx context.Context, userID int64, productCode string) (map[string]any, error) {
	detail, err := s.ProductDetail(ctx, productCode, 1, &userID)
	if err != nil {
		return nil, err
	}
	product, _ := detail["product"].(map[string]any)
	if product == nil {
		return nil, errcode.New(errcode.CodeNotFound, "product not found")
	}

	code := stringFromMap(product, "code")
	favorite := model.SupermarketFavorite{
		UserID:      userID,
		ProductCode: code,
		ProductName: stringFromMap(product, "name"),
		Brand:       stringFromMap(product, "brand"),
	}
	if err := s.runtime.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"product_name", "brand", "updated_at"}),
	}).Create(&favorite).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save supermarket favorite")
	}

	product["isFavorite"] = true
	return product, nil
}

// 6. RemoveFavorite deletes one AJO member supermarket favorite.
func (s *SupermarketOfferService) RemoveFavorite(ctx context.Context, userID int64, productCode string) error {
	productCode = strings.TrimSpace(productCode)
	if productCode == "" {
		return errcode.New(errcode.CodeValidationError, "product code is required")
	}

	if err := s.runtime.DB.WithContext(ctx).
		Where("user_id = ? AND product_code = ?", userID, productCode).
		Delete(&model.SupermarketFavorite{}).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to remove supermarket favorite")
	}

	return nil
}

// 7. isFavorite checks whether a product is saved by the user.
func (s *SupermarketOfferService) isFavorite(ctx context.Context, userID int64, productCode string) (bool, error) {
	var count int64
	err := s.runtime.DB.WithContext(ctx).Model(&model.SupermarketFavorite{}).
		Where("user_id = ? AND product_code = ?", userID, productCode).
		Count(&count).Error
	return count > 0, err
}

// 8. fallbackFavoriteProduct returns a minimal product payload when good-price detail is temporarily unavailable.
func fallbackFavoriteProduct(favorite model.SupermarketFavorite, baseURL string) map[string]any {
	imageURL := buildMediaURL(baseURL, supermarketImageObjectKey(favorite.ProductCode))
	return map[string]any{
		"code":       favorite.ProductCode,
		"name":       favorite.ProductName,
		"brand":      favorite.Brand,
		"image_url":  imageURL,
		"imageUrl":   imageURL,
		"isFavorite": true,
	}
}
