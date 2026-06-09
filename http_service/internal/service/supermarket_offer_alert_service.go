/*
 * Supermarket price alert service.
 * 1. Manage AJO-owned price alert rules.
 * 2. Evaluate rules against deployed good-price product data.
 * 3. Send matched reminders through the AJO mail provider.
 */
package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. ListPriceAlerts returns current member price alert rules.
func (s *SupermarketOfferService) ListPriceAlerts(ctx context.Context, userID int64) ([]SupermarketPriceAlertView, error) {
	var rules []model.SupermarketPriceAlert
	if err := s.runtime.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at desc").
		Find(&rules).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load supermarket price alerts")
	}

	items := make([]SupermarketPriceAlertView, 0, len(rules))
	for _, rule := range rules {
		items = append(items, supermarketAlertView(rule))
	}
	return items, nil
}

// 2. CreatePriceAlert creates or replaces the member's current rule for one product.
func (s *SupermarketOfferService) CreatePriceAlert(ctx context.Context, userID int64, input SupermarketPriceAlertInput) (*SupermarketPriceAlertView, error) {
	input.ProductCode = strings.TrimSpace(input.ProductCode)
	if input.ProductCode == "" {
		return nil, errcode.New(errcode.CodeValidationError, "product code is required")
	}
	if input.PriceMode == "" {
		input.PriceMode = "effective"
	}
	if input.PriceMode != "effective" && input.PriceMode != "list" {
		return nil, errcode.New(errcode.CodeValidationError, "price mode must be effective or list")
	}
	if input.TargetPrice != nil && *input.TargetPrice < 0 {
		return nil, errcode.New(errcode.CodeValidationError, "target price must not be negative")
	}

	detail, err := s.ProductDetail(ctx, input.ProductCode, 1, &userID)
	if err != nil {
		return nil, err
	}
	product, _ := detail["product"].(map[string]any)
	if product == nil {
		return nil, errcode.New(errcode.CodeNotFound, "product not found")
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	rule := model.SupermarketPriceAlert{
		UserID:        userID,
		ProductCode:   stringFromMap(product, "code"),
		ProductName:   stringFromMap(product, "name"),
		TargetPrice:   input.TargetPrice,
		PriceMode:     input.PriceMode,
		OfferRequired: input.OfferRequired,
		Enabled:       enabled,
	}
	if err := s.runtime.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "product_code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"product_name",
			"target_price",
			"price_mode",
			"offer_required",
			"enabled",
			"updated_at",
		}),
	}).Create(&rule).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save supermarket price alert")
	}

	current, err := s.latestAlertRule(ctx, userID, rule.ProductCode)
	if err != nil {
		return nil, err
	}
	view := supermarketAlertView(*current)
	return &view, nil
}

// 3. UpdatePriceAlert updates one member alert rule.
func (s *SupermarketOfferService) UpdatePriceAlert(ctx context.Context, userID int64, ruleID int64, input SupermarketPriceAlertUpdateInput) (*SupermarketPriceAlertView, error) {
	var rule model.SupermarketPriceAlert
	if err := s.runtime.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", ruleID, userID).
		First(&rule).Error; err != nil {
		return nil, supermarketAlertNotFound(err)
	}
	if input.PriceMode != "" {
		if input.PriceMode != "effective" && input.PriceMode != "list" {
			return nil, errcode.New(errcode.CodeValidationError, "price mode must be effective or list")
		}
		rule.PriceMode = input.PriceMode
	}
	if input.TargetPrice != nil && *input.TargetPrice < 0 {
		return nil, errcode.New(errcode.CodeValidationError, "target price must not be negative")
	}
	rule.TargetPrice = input.TargetPrice
	if input.OfferRequired != nil {
		rule.OfferRequired = *input.OfferRequired
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}

	if err := s.runtime.DB.WithContext(ctx).Save(&rule).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update supermarket price alert")
	}

	view := supermarketAlertView(rule)
	return &view, nil
}

// 4. DeletePriceAlert deletes one member alert rule.
func (s *SupermarketOfferService) DeletePriceAlert(ctx context.Context, userID int64, ruleID int64) error {
	if ruleID <= 0 {
		return errcode.New(errcode.CodeValidationError, "alert id is required")
	}
	if err := s.runtime.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", ruleID, userID).
		Delete(&model.SupermarketPriceAlert{}).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to delete supermarket price alert")
	}

	return nil
}

// 5. EvaluatePriceAlerts checks all enabled rules and sends matched email reminders.
func (s *SupermarketOfferService) EvaluatePriceAlerts(ctx context.Context) (int, error) {
	var rules []model.SupermarketPriceAlert
	if err := s.runtime.DB.WithContext(ctx).Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to load supermarket price alerts")
	}

	sent := 0
	for _, rule := range rules {
		matched, err := s.evaluateSinglePriceAlert(ctx, rule)
		if err != nil && s.runtime.Logger != nil {
			s.runtime.Logger.Warn("supermarket price alert failed", "rule_id", rule.ID, "error", err)
		}
		if matched {
			sent++
		}
	}

	return sent, nil
}

// 6. latestAlertRule loads the newest alert rule for one user and product.
func (s *SupermarketOfferService) latestAlertRule(ctx context.Context, userID int64, productCode string) (*model.SupermarketPriceAlert, error) {
	var rule model.SupermarketPriceAlert
	err := s.runtime.DB.WithContext(ctx).
		Where("user_id = ? AND product_code = ?", userID, productCode).
		Order("updated_at desc").
		First(&rule).Error
	if err != nil {
		return nil, supermarketAlertNotFound(err)
	}

	return &rule, nil
}

// 7. evaluateSinglePriceAlert evaluates one rule and sends one deduplicated email.
func (s *SupermarketOfferService) evaluateSinglePriceAlert(ctx context.Context, rule model.SupermarketPriceAlert) (bool, error) {
	detail, err := s.ProductDetail(ctx, rule.ProductCode, 1, nil)
	if err != nil {
		return false, err
	}
	stores := supermarketStoreSnapshots(detail)
	match := supermarketBestAlertMatch(rule, stores)
	if match == nil {
		return false, nil
	}

	email, err := s.userEmail(ctx, rule.UserID)
	if err != nil {
		return false, err
	}
	message := supermarketAlertMessage(rule, *match)
	event := model.SupermarketPriceAlertEvent{
		RuleID:       rule.ID,
		UserID:       rule.UserID,
		ProductCode:  rule.ProductCode,
		StoreCode:    match.Store,
		SnapshotDate: match.SnapshotDate,
		MatchedPrice: match.Price,
		OfferText:    match.Offer,
		Message:      message,
	}
	result := s.runtime.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if result.Error != nil {
		return false, errcode.New(errcode.CodeInternalError, "failed to save supermarket alert event")
	}
	if result.RowsAffected == 0 {
		return false, nil
	}

	if err := s.runtime.MailSender.Send(ctx, email, "[AJO Living] 超市價格提示", message); err != nil {
		return false, errcode.New(errcode.CodeInternalError, "failed to send supermarket price alert")
	}
	now := s.runtime.Now()
	_ = s.runtime.DB.WithContext(ctx).Model(&model.SupermarketPriceAlertEvent{}).
		Where("id = ?", event.ID).
		Update("notified_at", now).Error
	_ = s.runtime.DB.WithContext(ctx).Model(&model.SupermarketPriceAlert{}).
		Where("id = ?", rule.ID).
		Updates(map[string]any{"last_triggered_at": now, "updated_at": now}).Error

	return true, nil
}

// 8. supermarketStoreSnapshots extracts current store prices from product detail.
func supermarketStoreSnapshots(detail map[string]any) []supermarketStoreSnapshot {
	rawStores, _ := detail["stores"].([]any)
	stores := make([]supermarketStoreSnapshot, 0, len(rawStores))
	for _, raw := range rawStores {
		item, _ := raw.(map[string]any)
		if item == nil {
			continue
		}
		stores = append(stores, supermarketStoreSnapshot{
			Store:              stringFromMap(item, "store"),
			ListPrice:          floatFromMap(item, "listPrice"),
			EffectiveUnitPrice: floatFromMap(item, "effectiveUnitPrice"),
			Offer:              stringFromMap(item, "offer"),
			SnapshotDate:       stringFromMap(item, "snapshotDate"),
		})
	}

	return stores
}

// 9. supermarketBestAlertMatch returns the cheapest store row satisfying one rule.
func supermarketBestAlertMatch(rule model.SupermarketPriceAlert, stores []supermarketStoreSnapshot) *supermarketAlertMatch {
	var best *supermarketAlertMatch
	bestPrice := math.Inf(1)
	for _, store := range stores {
		if rule.OfferRequired && strings.TrimSpace(store.Offer) == "" {
			continue
		}
		price := store.EffectiveUnitPrice
		if rule.PriceMode == "list" {
			price = store.ListPrice
		}
		if rule.TargetPrice != nil && price > *rule.TargetPrice {
			continue
		}
		if price >= bestPrice {
			continue
		}
		bestPrice = price
		best = &supermarketAlertMatch{
			Store:        store.Store,
			Price:        price,
			Offer:        store.Offer,
			SnapshotDate: store.SnapshotDate,
		}
	}

	return best
}

// 10. userEmail loads the current AJO member email for alert delivery.
func (s *SupermarketOfferService) userEmail(ctx context.Context, userID int64) (string, error) {
	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error; err != nil {
		return "", errcode.New(errcode.CodeValidationError, "member email is required for price alerts")
	}
	if credential.Email == nil || strings.TrimSpace(*credential.Email) == "" {
		return "", errcode.New(errcode.CodeValidationError, "member email is required for price alerts")
	}

	return strings.TrimSpace(*credential.Email), nil
}

// 11. supermarketAlertMessage builds the plain text price alert email.
func supermarketAlertMessage(rule model.SupermarketPriceAlert, match supermarketAlertMatch) string {
	target := "任何價格"
	if rule.TargetPrice != nil {
		target = formatHKPriceForMail(*rule.TargetPrice)
	}
	priceLabel := "優惠後等效價"
	if rule.PriceMode == "list" {
		priceLabel = "原價"
	}
	offer := strings.TrimSpace(match.Offer)
	if offer == "" {
		offer = "-"
	}

	return fmt.Sprintf(
		"你的 AJO Living 超市價格提示已觸發。\n\n商品：%s (%s)\n超市：%s\n日期：%s\n%s：%s\n目標：%s\n優惠：%s\n\n價格及優惠只供參考，實際售價以商戶公布為準。",
		rule.ProductName,
		rule.ProductCode,
		match.Store,
		match.SnapshotDate,
		priceLabel,
		formatHKPriceForMail(match.Price),
		target,
		offer,
	)
}

// 12. supermarketAlertNotFound converts GORM not found errors.
func supermarketAlertNotFound(err error) error {
	if err == nil {
		return nil
	}
	if err == gorm.ErrRecordNotFound {
		return errcode.New(errcode.CodeNotFound, "price alert not found")
	}

	return errcode.New(errcode.CodeInternalError, "failed to load supermarket price alert")
}
