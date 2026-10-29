/*
 * iSmart 單位級副戶授權服務。
 * 1. 主戶對指定單位發起副戶授權：已註冊對象發授權確認通知，未註冊對象發註冊邀請。
 * 2. 被綁定人完成註冊後由註冊流程命中待授權記錄，記錄由待註冊轉為等待確認。
 * 3. 被綁定人確認後才調用 iSmart 授予 API，逾期 2 天未完成則記錄失效。
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"

	"gorm.io/gorm"
)

// 1. 副戶授權狀態與期限常量。
const (
	IsmartSubaccountAuthStatusPendingRegistration  = "pending_registration"
	IsmartSubaccountAuthStatusAwaitingConfirmation = "awaiting_confirmation"
	IsmartSubaccountAuthStatusActive               = "active"
	IsmartSubaccountAuthStatusExpired              = "expired"
	IsmartSubaccountAuthStatusRevoked              = "revoked"

	// IsmartSubaccountAuthNotificationCategory marks in-app notifications for subaccount authorizations.
	IsmartSubaccountAuthNotificationCategory = "ismart_subaccount_authorization"

	// IsmartSubaccountAuthWindow is the registration and confirmation validity window.
	IsmartSubaccountAuthWindow = 48 * time.Hour
)

// 2. IsmartSubaccountAuthorizationParams defines one owner invitation request.
type IsmartSubaccountAuthorizationParams struct {
	UnitID    string
	Phone     string
	Email     string
	SendEmail bool
	Remark    string
}

// 3. IsmartSubaccountAuthorizationResult reports the created authorization state.
type IsmartSubaccountAuthorizationResult struct {
	AuthorizationID      string `json:"authorization_id"`
	Status               string `json:"status"`
	RequiresRegistration bool   `json:"requires_registration"`
	EmailSent            bool   `json:"email_sent"`
	TargetUserID         int64  `json:"target_user_id,omitempty"`
	UpstreamMessage      string `json:"upstream_message,omitempty"`
}

// 4. IsmartSubaccountAuthorizationItem is one authorization row for the owner panel.
type IsmartSubaccountAuthorizationItem struct {
	AuthorizationID     string     `json:"authorization_id"`
	UnitID              string     `json:"unit_id"`
	BuildingID          string     `json:"building_id,omitempty"`
	BuildingName        string     `json:"building_name,omitempty"`
	UnitDisplay         string     `json:"unit_display,omitempty"`
	Phone               string     `json:"phone,omitempty"`
	Email               string     `json:"email,omitempty"`
	TargetUserID        int64      `json:"target_user_id,omitempty"`
	TargetUsername      string     `json:"target_username,omitempty"`
	Status              string     `json:"status"`
	Source              string     `json:"source,omitempty"`
	InvitationExpiresAt *time.Time `json:"invitation_expires_at,omitempty"`
	ConfirmExpiresAt    *time.Time `json:"confirm_expires_at,omitempty"`
	ResendCount         int        `json:"resend_count"`
	Remark              string     `json:"remark,omitempty"`
}

// 5. IsmartSubaccountConfirmResult reports the confirmation outcome.
type IsmartSubaccountConfirmResult struct {
	Status  string `json:"status"`
	Granted bool   `json:"granted"`
}

// 6. IsmartSubaccountAuthorizationService manages unit-level subaccount invitations and confirmations.
type IsmartSubaccountAuthorizationService struct {
	runtime *Runtime
}

// 7. NewIsmartSubaccountAuthorizationService creates the authorization service.
func NewIsmartSubaccountAuthorizationService(runtime *Runtime) *IsmartSubaccountAuthorizationService {
	return &IsmartSubaccountAuthorizationService{runtime: runtime}
}

// 8. CreateInvitation validates the target contacts and records the invitation or confirmation.
func (s *IsmartSubaccountAuthorizationService) CreateInvitation(ctx context.Context, ownerUserID int64, params IsmartSubaccountAuthorizationParams) (*IsmartSubaccountAuthorizationResult, error) {
	external := NewIsmartExternalService(s.runtime)
	unitID := strings.TrimSpace(params.UnitID)
	phone := normalizeSubaccountRegistrationPhone(params.Phone)
	email := normalizeEmail(params.Email)
	if unitID == "" {
		return nil, errcode.New(errcode.CodeValidationError, "unit_id is required")
	}
	if phone == "" && email == "" {
		return nil, errcode.New(errcode.CodeValidationError, "phone or email is required")
	}
	if email != "" && !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}
	if err := external.validateSubaccountOwnerUnit(ctx, ownerUserID, unitID); err != nil {
		return nil, err
	}

	contacts, err := external.CheckSubaccountContacts(ctx, IsmartSubaccountContactParams{Phone: phone, Email: email})
	if err != nil {
		return nil, err
	}
	// 兩項都填且至少一項已註冊時，必須落在同一個既有 iSmart 帳戶，否則拒絕提交。
	// 兩項均未註冊代表要走註冊邀請流程，不屬於衝突。
	if phone != "" && email != "" && (contacts.EmailExists || contacts.PhoneExists) && !contacts.SameUser {
		return nil, errcode.New(errcode.CodeValidationError, "target phone and email do not match the same ismart account")
	}

	countryCode, phoneNumber := subaccountAuthPhoneParts(phone)
	record := model.IsmartSubaccountAuthorization{
		PublicID:         utils.NewPublicID(),
		OwnerUserID:      ownerUserID,
		BuildingID:       posUnitBuildingID(unitID),
		UnitID:           unitID,
		PhoneCountryCode: countryCode,
		PhoneNumber:      phoneNumber,
		Email:            email,
		Remark:           strings.TrimSpace(params.Remark),
	}
	result := &IsmartSubaccountAuthorizationResult{AuthorizationID: record.PublicID}

	existing := contacts.EmailExists || contacts.PhoneExists
	if !existing {
		// 未註冊對象：不再代填姓名與密碼，改以電郵邀請對方自行註冊，電郵為必要條件。
		if email == "" {
			return nil, errcode.New(errcode.CodeValidationError, "email is required when the target has no ismart account")
		}
		record.Status = IsmartSubaccountAuthStatusPendingRegistration
		record.InvitationExpiresAt = subaccountAuthPtrTime(s.runtime.Now().Add(IsmartSubaccountAuthWindow))
		record.LastSentAt = subaccountAuthPtrTime(s.runtime.Now())
		if err := s.runtime.DB.WithContext(ctx).Create(&record).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create subaccount invitation")
		}
		result.Status = record.Status
		result.RequiresRegistration = true
		if params.SendEmail {
			result.EmailSent = s.sendInvitationEmail(ctx, record.Email, &record)
		}
		return result, nil
	}

	// 已註冊對象：直接進入授權確認流程，對方確認後才調用 iSmart。
	targetUserID := contacts.EmailUserID
	if targetUserID <= 0 {
		targetUserID = contacts.PhoneUserID
	}
	record.Status = IsmartSubaccountAuthStatusAwaitingConfirmation
	record.TargetUserID = targetUserID
	if localUserID := s.resolveLocalUserID(ctx, countryCode, phoneNumber, email); localUserID > 0 {
		record.GranteeUserID = subaccountAuthPtrInt64(localUserID)
	}
	token := s.issueConfirmToken(&record)
	record.LastSentAt = subaccountAuthPtrTime(s.runtime.Now())
	if err := s.runtime.DB.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create subaccount authorization")
	}
	result.Status = record.Status
	result.TargetUserID = targetUserID
	if params.SendEmail {
		result.EmailSent = s.sendConfirmationEmail(ctx, record.Email, token, &record)
	}
	s.notifyGrantee(ctx, &record)
	return result, nil
}

// 9. ListAuthorizations lists non-terminal authorizations owned by the current member.
func (s *IsmartSubaccountAuthorizationService) ListAuthorizations(ctx context.Context, ownerUserID int64, unitID string) ([]IsmartSubaccountAuthorizationItem, error) {
	query := s.runtime.DB.WithContext(ctx).
		Where("owner_user_id = ? AND status IN ?", ownerUserID, []string{
			IsmartSubaccountAuthStatusPendingRegistration,
			IsmartSubaccountAuthStatusAwaitingConfirmation,
		})
	if strings.TrimSpace(unitID) != "" {
		query = query.Where("unit_id = ?", strings.TrimSpace(unitID))
	}
	var records []model.IsmartSubaccountAuthorization
	if err := query.Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load subaccount authorizations")
	}
	items := make([]IsmartSubaccountAuthorizationItem, 0, len(records))
	for index := range records {
		items = append(items, *s.authorizationItem(&records[index]))
	}
	return items, nil
}

// 10. ResendNotification resends the invitation or confirmation notice with a fresh window.
func (s *IsmartSubaccountAuthorizationService) ResendNotification(ctx context.Context, ownerUserID int64, publicID string, sendEmail bool) (*IsmartSubaccountAuthorizationItem, error) {
	record, err := s.loadOwnedRecord(ctx, ownerUserID, publicID)
	if err != nil {
		return nil, err
	}
	if record.Status != IsmartSubaccountAuthStatusPendingRegistration && record.Status != IsmartSubaccountAuthStatusAwaitingConfirmation {
		return nil, errcode.New(errcode.CodeValidationError, "subaccount authorization is not pending")
	}
	now := s.runtime.Now()
	updates := map[string]any{
		"resend_count": gorm.Expr("resend_count + 1"),
		"last_sent_at": now,
		"updated_at":   now,
	}
	token := ""
	if record.Status == IsmartSubaccountAuthStatusPendingRegistration {
		updates["invitation_expires_at"] = now.Add(IsmartSubaccountAuthWindow)
	} else {
		token = utils.NewPublicID() + utils.NewPublicID()
		updates["confirm_token_hash"] = hashAuthorizationToken(token)
		updates["confirm_expires_at"] = now.Add(IsmartSubaccountAuthWindow)
	}
	if err := s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to resend subaccount notification")
	}
	record.ResendCount++
	record.LastSentAt = subaccountAuthPtrTime(now)
	if record.Status == IsmartSubaccountAuthStatusPendingRegistration {
		record.InvitationExpiresAt = subaccountAuthPtrTime(now.Add(IsmartSubaccountAuthWindow))
		if sendEmail {
			s.sendInvitationEmail(ctx, record.Email, record)
		}
	} else {
		record.ConfirmExpiresAt = subaccountAuthPtrTime(now.Add(IsmartSubaccountAuthWindow))
		if sendEmail {
			s.sendConfirmationEmail(ctx, record.Email, token, record)
		}
		s.notifyGrantee(ctx, record)
	}
	return s.authorizationItem(record), nil
}

// 11. Revoke cancels one pending authorization owned by the current member.
func (s *IsmartSubaccountAuthorizationService) Revoke(ctx context.Context, ownerUserID int64, publicID string) error {
	record, err := s.loadOwnedRecord(ctx, ownerUserID, publicID)
	if err != nil {
		return err
	}
	if record.Status == IsmartSubaccountAuthStatusActive {
		return errcode.New(errcode.CodeValidationError, "active subaccount authorization must be revoked through ismart")
	}
	result := s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
		Where("id = ? AND status <> ?", record.ID, IsmartSubaccountAuthStatusActive).
		Update("status", IsmartSubaccountAuthStatusRevoked)
	if result.Error != nil {
		return errcode.New(errcode.CodeInternalError, "failed to revoke subaccount authorization")
	}
	return nil
}

// 12. ResolveToken resolves one confirmation token for the public confirmation page.
func (s *IsmartSubaccountAuthorizationService) ResolveToken(ctx context.Context, token string) (*IsmartSubaccountAuthorizationItem, error) {
	record, err := s.loadByConfirmToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.authorizationItem(record), nil
}

// 13. ConfirmByToken confirms one authorization from the email link.
func (s *IsmartSubaccountAuthorizationService) ConfirmByToken(ctx context.Context, token string) (*IsmartSubaccountConfirmResult, error) {
	record, err := s.loadByConfirmToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.grantOnConfirm(ctx, record)
}

// 14. ConfirmByID confirms one authorization from the in-app notification entry.
func (s *IsmartSubaccountAuthorizationService) ConfirmByID(ctx context.Context, granteeUserID int64, publicID string) (*IsmartSubaccountConfirmResult, error) {
	var record model.IsmartSubaccountAuthorization
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(publicID)).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "subaccount authorization not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load subaccount authorization")
	}
	if record.GranteeUserID == nil || *record.GranteeUserID != granteeUserID {
		return nil, errcode.New(errcode.CodeAuthForbidden, "subaccount authorization belongs to another member")
	}
	if err := s.ensureNotExpired(ctx, &record); err != nil {
		return nil, err
	}
	return s.grantOnConfirm(ctx, &record)
}

// 15. ProcessRegistrationForPendingAuthorizations matches a fresh registration against pending invitations.
func (s *IsmartSubaccountAuthorizationService) ProcessRegistrationForPendingAuthorizations(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return nil
	}
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load registered member")
	}
	email := ""
	var credential model.UserCredential
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&credential).Error; err == nil && credential.Email != nil {
		email = normalizeEmail(*credential.Email)
	}
	phones := subaccountAuthPhoneCandidates(user.PhoneCountryCode, user.PhoneNumber)

	query := s.runtime.DB.WithContext(ctx).
		Where("status = ?", IsmartSubaccountAuthStatusPendingRegistration).
		Where("invitation_expires_at IS NULL OR invitation_expires_at > ?", s.runtime.Now())
	if email != "" && len(phones) > 0 {
		query = query.Where("email = ? OR phone_number IN ?", email, phones)
	} else if email != "" {
		query = query.Where("email = ?", email)
	} else if len(phones) > 0 {
		query = query.Where("phone_number IN ?", phones)
	} else {
		return nil
	}

	var records []model.IsmartSubaccountAuthorization
	if err := query.Find(&records).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load pending subaccount authorizations")
	}
	if len(records) == 0 {
		return nil
	}

	targetUserID, targetUsername := int64(0), ""
	if account, err := NewIsmartExternalService(s.runtime).loadIsmartAccount(ctx, userID); err == nil && account != nil {
		targetUserID = account.IsmartUserID
	}
	for index := range records {
		record := &records[index]
		token := utils.NewPublicID() + utils.NewPublicID()
		now := s.runtime.Now()
		updates := map[string]any{
			"status":             IsmartSubaccountAuthStatusAwaitingConfirmation,
			"grantee_user_id":    userID,
			"target_user_id":     targetUserID,
			"target_username":    targetUsername,
			"confirm_token_hash": hashAuthorizationToken(token),
			"confirm_expires_at": now.Add(IsmartSubaccountAuthWindow),
			"last_sent_at":       now,
			"updated_at":         now,
		}
		// 條件更新保證同一記錄只會被一次註冊流程接手。
		result := s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
			Where("id = ? AND status = ?", record.ID, IsmartSubaccountAuthStatusPendingRegistration).
			Updates(updates)
		if result.Error != nil {
			return errcode.New(errcode.CodeInternalError, "failed to promote subaccount authorization")
		}
		if result.RowsAffected == 0 {
			continue
		}
		record.Status = IsmartSubaccountAuthStatusAwaitingConfirmation
		record.GranteeUserID = subaccountAuthPtrInt64(userID)
		record.TargetUserID = targetUserID
		record.TargetUsername = targetUsername
		record.ConfirmTokenHash = hashAuthorizationToken(token)
		record.ConfirmExpiresAt = subaccountAuthPtrTime(now.Add(IsmartSubaccountAuthWindow))
		s.sendConfirmationEmail(ctx, record.Email, token, record)
		s.notifyGrantee(ctx, record)
	}
	return nil
}

// 16. ExpireOverdueAuthorizations marks overdue invitations and confirmations as expired.
func (s *IsmartSubaccountAuthorizationService) ExpireOverdueAuthorizations(ctx context.Context) (int64, error) {
	now := s.runtime.Now()
	result := s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
		Where("status IN ?", []string{IsmartSubaccountAuthStatusPendingRegistration, IsmartSubaccountAuthStatusAwaitingConfirmation}).
		Where("(invitation_expires_at IS NOT NULL AND invitation_expires_at < ?) OR (confirm_expires_at IS NOT NULL AND confirm_expires_at < ?)", now, now).
		Update("status", IsmartSubaccountAuthStatusExpired)
	if result.Error != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to expire subaccount authorizations")
	}
	return result.RowsAffected, nil
}

// 17. grantOnConfirm claims the record then calls the iSmart grant API exactly once.
func (s *IsmartSubaccountAuthorizationService) grantOnConfirm(ctx context.Context, record *model.IsmartSubaccountAuthorization) (*IsmartSubaccountConfirmResult, error) {
	if record.Status == IsmartSubaccountAuthStatusActive {
		return &IsmartSubaccountConfirmResult{Status: IsmartSubaccountAuthStatusActive, Granted: false}, nil
	}
	if record.Status != IsmartSubaccountAuthStatusAwaitingConfirmation {
		return nil, errcode.New(errcode.CodeValidationError, "subaccount authorization is not awaiting confirmation")
	}
	if err := s.ensureNotExpired(ctx, record); err != nil {
		return nil, err
	}

	now := s.runtime.Now()
	claim := s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
		Where("id = ? AND status = ?", record.ID, IsmartSubaccountAuthStatusAwaitingConfirmation).
		Updates(map[string]any{
			"status":             IsmartSubaccountAuthStatusActive,
			"confirmed_at":       now,
			"granted_at":         now,
			"confirm_token_hash": "",
			"updated_at":         now,
		})
	if claim.Error != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to confirm subaccount authorization")
	}
	if claim.RowsAffected == 0 {
		return &IsmartSubaccountConfirmResult{Status: IsmartSubaccountAuthStatusActive, Granted: false}, nil
	}

	_, err := NewIsmartExternalService(s.runtime).GrantSubaccount(ctx, record.OwnerUserID, IsmartSubaccountMutationParams{
		UnitID:       record.UnitID,
		TargetUserID: record.TargetUserID,
		TargetPhone:  strings.TrimSpace(record.PhoneCountryCode + record.PhoneNumber),
		TargetEmail:  record.Email,
		Remark:       record.Remark,
	})
	if err != nil {
		// 上游失敗時回退為待確認，讓被綁人或主戶可重試，不留在假的已生效狀態。
		s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
			Where("id = ?", record.ID).
			Updates(map[string]any{
				"status":       IsmartSubaccountAuthStatusAwaitingConfirmation,
				"confirmed_at": nil,
				"granted_at":   nil,
				"updated_at":   s.runtime.Now(),
			})
		return nil, err
	}

	record.Status = IsmartSubaccountAuthStatusActive
	record.ConfirmedAt = subaccountAuthPtrTime(now)
	record.GrantedAt = subaccountAuthPtrTime(now)
	s.sendGrantedEmail(ctx, record)
	s.notifyGranted(ctx, record)
	return &IsmartSubaccountConfirmResult{Status: IsmartSubaccountAuthStatusActive, Granted: true}, nil
}

// 18. ensureNotExpired lazily expires one overdue record before confirmation.
func (s *IsmartSubaccountAuthorizationService) ensureNotExpired(ctx context.Context, record *model.IsmartSubaccountAuthorization) error {
	now := s.runtime.Now()
	deadline := record.ConfirmExpiresAt
	if deadline == nil {
		deadline = record.InvitationExpiresAt
	}
	if deadline != nil && now.After(*deadline) {
		s.runtime.DB.WithContext(ctx).Model(&model.IsmartSubaccountAuthorization{}).
			Where("id = ? AND status IN ?", record.ID, []string{IsmartSubaccountAuthStatusPendingRegistration, IsmartSubaccountAuthStatusAwaitingConfirmation}).
			Update("status", IsmartSubaccountAuthStatusExpired)
		record.Status = IsmartSubaccountAuthStatusExpired
		return errcode.New(errcode.CodeExpired, "subaccount authorization is expired")
	}
	return nil
}

// 19. loadOwnedRecord loads one authorization owned by the current member.
func (s *IsmartSubaccountAuthorizationService) loadOwnedRecord(ctx context.Context, ownerUserID int64, publicID string) (*model.IsmartSubaccountAuthorization, error) {
	var record model.IsmartSubaccountAuthorization
	err := s.runtime.DB.WithContext(ctx).
		Where("public_id = ? AND owner_user_id = ?", strings.TrimSpace(publicID), ownerUserID).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "subaccount authorization not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load subaccount authorization")
	}
	return &record, nil
}

// 20. loadByConfirmToken loads one authorization by confirmation token hash.
func (s *IsmartSubaccountAuthorizationService) loadByConfirmToken(ctx context.Context, token string) (*model.IsmartSubaccountAuthorization, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil, errcode.New(errcode.CodeValidationError, "confirmation token is required")
	}
	var record model.IsmartSubaccountAuthorization
	err := s.runtime.DB.WithContext(ctx).Where("confirm_token_hash = ?", hashAuthorizationToken(trimmed)).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "subaccount authorization not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load subaccount authorization")
	}
	return &record, nil
}

// 21. issueConfirmToken rotates the confirmation token and its expiry on one record.
func (s *IsmartSubaccountAuthorizationService) issueConfirmToken(record *model.IsmartSubaccountAuthorization) string {
	token := utils.NewPublicID() + utils.NewPublicID()
	record.ConfirmTokenHash = hashAuthorizationToken(token)
	record.ConfirmExpiresAt = subaccountAuthPtrTime(s.runtime.Now().Add(IsmartSubaccountAuthWindow))
	return token
}

// 22. authorizationItem maps one record to the owner-facing row.
func (s *IsmartSubaccountAuthorizationService) authorizationItem(record *model.IsmartSubaccountAuthorization) *IsmartSubaccountAuthorizationItem {
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(context.Background(), record.UnitID)
	return &IsmartSubaccountAuthorizationItem{
		AuthorizationID:     record.PublicID,
		UnitID:              record.UnitID,
		BuildingID:          record.BuildingID,
		UnitDisplay:         unitDisplay,
		Phone:               strings.TrimSpace(record.PhoneCountryCode + record.PhoneNumber),
		Email:               record.Email,
		TargetUserID:        record.TargetUserID,
		TargetUsername:      record.TargetUsername,
		Status:              record.Status,
		Source:              "local",
		InvitationExpiresAt: record.InvitationExpiresAt,
		ConfirmExpiresAt:    record.ConfirmExpiresAt,
		ResendCount:         record.ResendCount,
		Remark:              record.Remark,
	}
}

// 23. resolveLocalUserID resolves the AJO member matching the invited contacts, if any.
func (s *IsmartSubaccountAuthorizationService) resolveLocalUserID(ctx context.Context, countryCode string, phoneNumber string, email string) int64 {
	if phoneNumber != "" {
		var user model.User
		query := s.runtime.DB.WithContext(ctx).Model(&model.User{}).Where("phone_number = ?", phoneNumber)
		if countryCode != "" {
			query = query.Where("phone_country_code IN ?", []string{countryCode, strings.TrimPrefix(countryCode, "+")})
		}
		if err := query.First(&user).Error; err == nil {
			return user.ID
		}
	}
	if email != "" {
		var credential model.UserCredential
		if err := s.runtime.DB.WithContext(ctx).Where("email = ?", email).First(&credential).Error; err == nil {
			return credential.UserID
		}
	}
	return 0
}

// 24. notifyGrantee writes the in-app confirmation entry for the invited member.
func (s *IsmartSubaccountAuthorizationService) notifyGrantee(ctx context.Context, record *model.IsmartSubaccountAuthorization) {
	if record.GranteeUserID == nil || *record.GranteeUserID <= 0 {
		return
	}
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(ctx, record.UnitID)
	body := fmt.Sprintf("單位 %s 的副戶授權等待您確認，請於 2 天內確認。", paymentFirstNonEmpty(unitDisplay, record.UnitID))
	s.createNotification(ctx, *record.GranteeUserID, "待確認的 iSmart 副戶授權", body, record.PublicID)
}

// 25. notifyGranted writes the in-app entry after the authorization becomes effective.
func (s *IsmartSubaccountAuthorizationService) notifyGranted(ctx context.Context, record *model.IsmartSubaccountAuthorization) {
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(ctx, record.UnitID)
	body := fmt.Sprintf("單位 %s 的副戶授權已生效。", paymentFirstNonEmpty(unitDisplay, record.UnitID))
	if record.GranteeUserID != nil && *record.GranteeUserID > 0 {
		s.createNotification(ctx, *record.GranteeUserID, "iSmart 副戶授權已生效", body, record.PublicID)
	}
	s.createNotification(ctx, record.OwnerUserID, "副戶授權已生效", body, record.PublicID)
}

// 26. createNotification stores one in-app notification without blocking the business flow.
func (s *IsmartSubaccountAuthorizationService) createNotification(ctx context.Context, userID int64, title string, body string, relatedPublicID string) {
	if userID <= 0 {
		return
	}
	err := NewNotificationService(s.runtime).CreateNotification(ctx, s.runtime.DB, CreateNotificationParams{
		UserID:          userID,
		Category:        IsmartSubaccountAuthNotificationCategory,
		Title:           title,
		Body:            body,
		RelatedType:     IsmartSubaccountAuthNotificationCategory,
		RelatedPublicID: relatedPublicID,
	})
	if err != nil && s.runtime.Logger != nil {
		s.runtime.Logger.Warn("failed to create subaccount notification", "user_id", userID, "error", err)
	}
}

// 27. sendInvitationEmail sends the registration invitation to an unregistered target.
func (s *IsmartSubaccountAuthorizationService) sendInvitationEmail(ctx context.Context, recipient string, record *model.IsmartSubaccountAuthorization) bool {
	if strings.TrimSpace(recipient) == "" || s.runtime.MailSender == nil {
		return false
	}
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(ctx, record.UnitID)
	link := s.registrationLink(record)
	htmlBody := strings.Join([]string{
		"<p>您好，</p>",
		"<p>您獲邀請使用以下 iSmart 單位功能：</p>",
		fmt.Sprintf(
			"<p>大廈：%s<br>單位：%s<br>授權範圍：遙距開門及查看大廈通告。</p>",
			html.EscapeString(paymentFirstNonEmpty(record.BuildingID, "目前單位")),
			html.EscapeString(paymentFirstNonEmpty(unitDisplay, record.UnitID)),
		),
		fmt.Sprintf(
			"<p>請於 2 天內點擊以下連結，自行填寫資料及設定密碼完成註冊：<br><a href=\"%s\">%s</a></p>",
			html.EscapeString(link),
			html.EscapeString(link),
		),
		"<p>完成註冊後，系統會再發出授權確認通知，確認後授權方會生效。</p>",
		"<p>AJO Living</p>",
	}, "\n")
	if err := s.runtime.MailSender.SendHTML(ctx, recipient, "[AJO Living] iSmart 副戶授權邀請", htmlBody); err != nil {
		if s.runtime.Logger != nil {
			s.runtime.Logger.Warn("failed to send subaccount invitation email", "recipient", recipient, "error", err)
		}
		return false
	}
	return true
}

// 28. sendConfirmationEmail sends the authorization confirmation link.
func (s *IsmartSubaccountAuthorizationService) sendConfirmationEmail(ctx context.Context, recipient string, token string, record *model.IsmartSubaccountAuthorization) bool {
	if strings.TrimSpace(recipient) == "" || s.runtime.MailSender == nil {
		return false
	}
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(ctx, record.UnitID)
	confirmationLink := s.confirmationLink(token)
	htmlBody := strings.Join([]string{
		"<p>您好，</p>",
		"<p>以下單位正等待您確認副戶授權：</p>",
		fmt.Sprintf(
			"<p>大廈：%s<br>單位：%s<br>授權範圍：遙距開門及查看大廈通告。</p>",
			html.EscapeString(paymentFirstNonEmpty(record.BuildingID, "目前單位")),
			html.EscapeString(paymentFirstNonEmpty(unitDisplay, record.UnitID)),
		),
		fmt.Sprintf(
			"<p>請於 2 天內點擊以下連結並確認接受，確認後授權才會生效：<br><a href=\"%s\">%s</a></p>",
			html.EscapeString(confirmationLink),
			html.EscapeString(confirmationLink),
		),
		"<p>若逾期未確認，授權將失效，需由主戶重新邀請。</p>",
		"<p>AJO Living</p>",
	}, "\n")
	if err := s.runtime.MailSender.SendHTML(ctx, recipient, "[AJO Living] iSmart 副戶授權確認", htmlBody); err != nil {
		if s.runtime.Logger != nil {
			s.runtime.Logger.Warn("failed to send subaccount confirmation email", "recipient", recipient, "error", err)
		}
		return false
	}
	return true
}

// 29. sendGrantedEmail notifies both sides after the authorization takes effect.
func (s *IsmartSubaccountAuthorizationService) sendGrantedEmail(ctx context.Context, record *model.IsmartSubaccountAuthorization) {
	if s.runtime.MailSender == nil {
		return
	}
	_, unitDisplay := NewIsmartExternalService(s.runtime).resolveSubaccountUnitDisplay(ctx, record.UnitID)
	body := strings.Join([]string{
		"您好，",
		"",
		"以下 iSmart 副戶授權已生效：",
		fmt.Sprintf("大廈：%s", paymentFirstNonEmpty(record.BuildingID, "目前單位")),
		fmt.Sprintf("單位：%s", paymentFirstNonEmpty(unitDisplay, record.UnitID)),
		"授權範圍：遙距開門及查看大廈通告。",
		"",
		fmt.Sprintf("登入入口：%s/login", strings.TrimRight(strings.TrimSpace(s.runtime.Config.AppPublicBaseURL), "/")),
		"",
		"AJO Living",
	}, "\n")
	if strings.TrimSpace(record.Email) != "" {
		if err := s.runtime.MailSender.Send(ctx, record.Email, "[AJO Living] iSmart 副戶授權已生效", body); err != nil && s.runtime.Logger != nil {
			s.runtime.Logger.Warn("failed to send subaccount granted email", "recipient", record.Email, "error", err)
		}
	}
}

// 30. registrationLink builds the prefilled registration entry for one invitation.
func (s *IsmartSubaccountAuthorizationService) registrationLink(record *model.IsmartSubaccountAuthorization) string {
	base := strings.TrimRight(strings.TrimSpace(s.runtime.Config.AppPublicBaseURL), "/")
	query := url.Values{}
	query.Set("action", "register")
	if record.Email != "" {
		query.Set("email", record.Email)
	}
	if phone := strings.TrimSpace(record.PhoneCountryCode + record.PhoneNumber); phone != "" {
		query.Set("phone", phone)
	}
	return fmt.Sprintf("%s/login?%s", base, query.Encode())
}

// 31. confirmationLink builds the public confirmation link for one authorization.
func (s *IsmartSubaccountAuthorizationService) confirmationLink(token string) string {
	base := strings.TrimRight(strings.TrimSpace(s.runtime.Config.AppPublicBaseURL), "/")
	return fmt.Sprintf("%s/auth/ismart-subaccount-authorizations/confirm?token=%s", base, url.QueryEscape(token))
}

// 32. subaccountAuthPhoneParts splits one normalized E.164 value into country code and local number.
func subaccountAuthPhoneParts(phone string) (string, string) {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return "", ""
	}
	digits := digitsOnly(trimmed)
	// 本地號碼必須去掉區號，才能與會員資料的 phone_number 及 iSmart 上報的 E.164 保持一致。
	switch {
	case strings.HasPrefix(digits, "852") && len(digits) > 8:
		return "+852", strings.TrimPrefix(digits, "852")
	case strings.HasPrefix(digits, "86") && len(digits) > 11:
		return "+86", strings.TrimPrefix(digits, "86")
	}
	return "", digits
}

// 33. subaccountAuthPhoneCandidates returns the normalized phone forms used for matching.
func subaccountAuthPhoneCandidates(countryCode string, phoneNumber string) []string {
	raw := strings.TrimSpace(phoneNumber)
	if raw == "" {
		return nil
	}
	candidates := []string{normalizeSubaccountRegistrationPhone(raw)}
	country := strings.TrimSpace(countryCode)
	if country != "" {
		candidates = append(candidates, normalizeSubaccountRegistrationPhone(country+raw))
	}
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate != "" && !containsString(result, candidate) {
			result = append(result, candidate)
		}
	}
	return result
}

// 34. subaccountAuthPtrTime returns a pointer to one timestamp.
func subaccountAuthPtrTime(value time.Time) *time.Time { return &value }

// 35. subaccountAuthPtrInt64 returns a pointer to one identifier.
func subaccountAuthPtrInt64(value int64) *int64 { return &value }
