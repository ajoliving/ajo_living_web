/*
 * 樓盤草稿扣費更正服務。
 * 1. 為舊版重複草稿扣費建立可追溯的退款流水。
 * 2. 只允許更正仍處於草稿狀態的指定放售樓盤。
 * 3. 以穩定幂等鍵確保重複執行不會重複退款。
 */
package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. PropertyDraftChargeCorrectionResult reports the effective correction result for one draft.
type PropertyDraftChargeCorrectionResult struct {
	ListingID        string
	DraftPointsPaid  int64
	RefundPoints     int64
	CorrectionNeeded bool
}

// 2. CorrectPropertySaleDraftCharge previews or applies one legacy draft-charge correction.
func (s *WalletService) CorrectPropertySaleDraftCharge(ctx context.Context, listingPublicID string, apply bool) (*PropertyDraftChargeCorrectionResult, error) {
	if s == nil || s.runtime == nil || s.runtime.DB == nil {
		return nil, errcode.New(errcode.CodeInternalError, "wallet service is not configured")
	}

	result := &PropertyDraftChargeCorrectionResult{}
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var listing model.Listing
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("public_id = ? AND module = ? AND publication_status = ? AND is_deleted = ?", listingPublicID, string(PropertyChannelSale), "draft", false).
			First(&listing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errcode.New(errcode.CodeNotFound, "property sale draft not found")
			}
			return errcode.New(errcode.CodeInternalError, "failed to load property sale draft")
		}

		paid, err := propertySaleDraftDebitTotal(ctx, tx, listing.ID)
		if err != nil {
			return err
		}
		corrected, err := propertySaleDraftCorrectionTotal(ctx, tx, listing.ID)
		if err != nil {
			return err
		}
		netPaid := max(paid-corrected, 0)
		result.ListingID = listing.PublicID
		result.DraftPointsPaid = netPaid
		if netPaid <= PropertySaleDraftCost {
			return nil
		}

		result.RefundPoints = netPaid - PropertySaleDraftCost
		result.CorrectionNeeded = true
		if !apply {
			return nil
		}
		_, err = s.CreditPointsWithTx(ctx, tx, WalletCreditParams{
			UserID:         listing.OwnerUserID,
			Amount:         result.RefundPoints,
			SourceType:     WalletSourceListingRefund,
			BizModule:      string(PropertyChannelSale),
			ActionType:     WalletActionDraftChargeCorrection,
			ListingID:      &listing.ID,
			IdempotencyKey: fmt.Sprintf("property_sale:draft_charge_correction:%d:v1", listing.ID),
			Note:           "樓盤草稿重複扣費更正",
		})
		if err != nil {
			return err
		}
		result.DraftPointsPaid = PropertySaleDraftCost
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 3. propertySaleDraftDebitTotal returns all historical draft debits for one listing.
func propertySaleDraftDebitTotal(ctx context.Context, tx *gorm.DB, listingID int64) (int64, error) {
	var paid int64
	if err := tx.WithContext(ctx).Model(&model.WalletTransaction{}).
		Where("listing_id = ? AND biz_module = ? AND action_type = ? AND direction = ?", listingID, string(PropertyChannelSale), WalletActionSaveDraft, WalletDirectionDebit).
		Select("COALESCE(SUM(amount), 0)").Scan(&paid).Error; err != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to load property draft debits")
	}
	return paid, nil
}

// 4. propertySaleDraftCorrectionTotal returns all recorded draft-charge corrections for one listing.
func propertySaleDraftCorrectionTotal(ctx context.Context, tx *gorm.DB, listingID int64) (int64, error) {
	var corrected int64
	if err := tx.WithContext(ctx).Model(&model.WalletTransaction{}).
		Where("listing_id = ? AND biz_module = ? AND action_type = ? AND direction = ?", listingID, string(PropertyChannelSale), WalletActionDraftChargeCorrection, WalletDirectionCredit).
		Select("COALESCE(SUM(amount), 0)").Scan(&corrected).Error; err != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to load property draft corrections")
	}
	return corrected, nil
}
