/*
 * Staff wallet ledger business logic.
 * 1. Credit operator grants into the shared AJO Point ledger.
 * 2. Search immutable wallet transactions for staff review.
 * 3. Attach target and operator account metadata to staff responses.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. GrantOperatorPointsByUserID credits points to a user selected by public id.
func (s *WalletService) GrantOperatorPointsByUserID(ctx context.Context, operatorUserID int64, targetPublicID string, amount int64, note string) (*StaffWalletGrantResponse, error) {
	note = strings.TrimSpace(note)
	if amount <= 0 || amount > maxOperatorGrantPoints {
		return nil, errcode.New(errcode.CodeValidationError, "points amount is invalid")
	}
	if note == "" {
		return nil, errcode.New(errcode.CodeValidationError, "grant note is required")
	}

	target, err := s.loadWalletUserByPublicID(ctx, targetPublicID)
	if err != nil {
		return nil, err
	}
	operator, err := s.loadWalletUserByID(ctx, operatorUserID)
	if err != nil {
		return nil, err
	}
	charge, err := s.GrantOperatorPoints(ctx, operatorUserID, target.ID, amount, note)
	if err != nil {
		return nil, err
	}

	users, err := s.walletUsersByIDs(ctx, []int64{target.ID, operator.ID})
	if err != nil {
		return nil, err
	}
	return &StaffWalletGrantResponse{
		Charge:     *charge,
		TargetUser: users[target.ID],
		Operator:   users[operator.ID],
	}, nil
}

// 2. ListStaffWalletTransactions returns paged wallet ledger rows for staff.
func (s *WalletService) ListStaffWalletTransactions(ctx context.Context, filters StaffWalletTransactionFilters) ([]StaffWalletTransactionResponse, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.WalletTransaction{})

	if userID := strings.TrimSpace(filters.UserID); userID != "" {
		user, err := s.loadWalletUserByPublicID(ctx, userID)
		if err != nil {
			return nil, nil, err
		}
		query = query.Where("user_id = ?", user.ID)
	}
	query = applyWalletTransactionFilters(query, filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count wallet transactions")
	}

	var rows []model.WalletTransaction
	if err := query.Order("created_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load wallet transactions")
	}

	items, err := s.toStaffWalletTransactions(ctx, rows)
	if err != nil {
		return nil, nil, err
	}
	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 3. applyWalletTransactionFilters applies safe ledger filters.
func applyWalletTransactionFilters(query *gorm.DB, filters StaffWalletTransactionFilters) *gorm.DB {
	if direction := strings.TrimSpace(filters.Direction); direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if sourceType := strings.TrimSpace(filters.SourceType); sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if bizModule := strings.TrimSpace(filters.BizModule); bizModule != "" {
		query = query.Where("biz_module = ?", bizModule)
	}
	return query
}

// 4. toStaffWalletTransactions maps ledger rows with user metadata.
func (s *WalletService) toStaffWalletTransactions(ctx context.Context, rows []model.WalletTransaction) ([]StaffWalletTransactionResponse, error) {
	userIDs := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
		if row.OperatorUserID != nil {
			userIDs = append(userIDs, *row.OperatorUserID)
		}
	}
	users, err := s.walletUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]StaffWalletTransactionResponse, 0, len(rows))
	for _, row := range rows {
		item := StaffWalletTransactionResponse{
			TransactionID: row.PublicID,
			Direction:     row.Direction,
			Amount:        row.Amount,
			BalanceBefore: row.BalanceBefore,
			BalanceAfter:  row.BalanceAfter,
			SourceType:    row.SourceType,
			BizModule:     row.BizModule,
			ActionType:    row.ActionType,
			Note:          row.Note,
			TargetUser:    users[row.UserID],
			CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		}
		if row.OperatorUserID != nil {
			operator := users[*row.OperatorUserID]
			item.OperatorUser = &operator
		}
		items = append(items, item)
	}
	return items, nil
}

// 5. loadWalletUserByPublicID loads a user by public id.
func (s *WalletService) loadWalletUserByPublicID(ctx context.Context, publicID string) (*model.User, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(publicID)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "user not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user")
	}
	return &user, nil
}

// 6. loadWalletUserByID loads a user by internal id.
func (s *WalletService) loadWalletUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "user not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user")
	}
	return &user, nil
}

// 7. walletUsersByIDs loads user metadata for staff wallet responses.
func (s *WalletService) walletUsersByIDs(ctx context.Context, userIDs []int64) (map[int64]StaffWalletUserResponse, error) {
	uniqueIDs := uniqueInt64s(userIDs)
	if len(uniqueIDs) == 0 {
		return map[int64]StaffWalletUserResponse{}, nil
	}

	var users []model.User
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", uniqueIDs).Find(&users).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load wallet users")
	}

	var profiles []model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Where("user_id IN ?", uniqueIDs).Find(&profiles).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load wallet user profiles")
	}
	displayNames := make(map[int64]string, len(profiles))
	for _, profile := range profiles {
		displayNames[profile.UserID] = profile.DisplayName
	}

	result := make(map[int64]StaffWalletUserResponse, len(users))
	for _, user := range users {
		result[user.ID] = StaffWalletUserResponse{
			UserID:           user.PublicID,
			PhoneCountryCode: user.PhoneCountryCode,
			PhoneNumber:      user.PhoneNumber,
			DisplayName:      displayNames[user.ID],
		}
	}
	return result, nil
}

// 8. uniqueInt64s removes duplicate and zero ids.
func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
