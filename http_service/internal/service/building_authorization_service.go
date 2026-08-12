/*
 * 大廈副戶授權服務。
 * 1. 以電話及電郵建立或識別副戶帳戶。
 * 2. 保存按大廈劃分的功能授權，並發送邀請及通知郵件。
 * 3. 在授權接受及撤銷時維持一次性令牌與狀態。
 */
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	BuildingAuthorizationStatusPending = "pending"
	BuildingAuthorizationStatusActive  = "active"
	BuildingAuthorizationStatusRevoked = "revoked"
	BuildingPermissionRemoteDoorOpen   = "remote_door_open"
	BuildingPermissionBuildingNotices  = "building_notices"
	BuildingPermissionBuildingInfo     = "building_info"
	BuildingPermissionManagementFees   = "management_fees"
	BuildingPermissionOtherFees        = "other_fees"
	BuildingPermissionServiceCases     = "service_cases"
	BuildingPermissionBuildingComments = "building_comments"
	BuildingPermissionQRCode           = "building_qrcode"
)

// 1. BuildingAuthorizationCreateParams defines owner grant input.
type BuildingAuthorizationCreateParams struct {
	BuildingID       string
	PhoneCountryCode string
	PhoneNumber      string
	Email            string
	Permissions      []string
}

// 2. BuildingAuthorizationAcceptParams defines invitation acceptance input.
type BuildingAuthorizationAcceptParams struct {
	Token    string
	Username string
	Password string
}

// 3. BuildingAuthorizationService handles local building capability grants.
type BuildingAuthorizationService struct {
	runtime *Runtime
}

// 4. NewBuildingAuthorizationService creates the authorization service.
func NewBuildingAuthorizationService(runtime *Runtime) *BuildingAuthorizationService {
	return &BuildingAuthorizationService{runtime: runtime}
}

// 5. AvailablePermissions returns the member-facing building menu capabilities.
func (s *BuildingAuthorizationService) AvailablePermissions() []map[string]string {
	return []map[string]string{
		{"code": BuildingPermissionRemoteDoorOpen, "name": "遙距開大廈公門"},
		{"code": BuildingPermissionBuildingNotices, "name": "查看通告"},
	}
}

// 6. Create grants building capabilities and sends the appropriate email.
func (s *BuildingAuthorizationService) Create(ctx context.Context, ownerUserID int64, params BuildingAuthorizationCreateParams) (map[string]any, error) {
	buildingID := strings.TrimSpace(params.BuildingID)
	phoneCountryCode := normalizePhoneCountryCode(params.PhoneCountryCode)
	phoneNumber := normalizePhoneNumber(params.PhoneNumber)
	email := normalizeEmail(params.Email)
	if buildingID == "" || !isValidPhone(phoneCountryCode, phoneNumber) || !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "building id, valid phone number and email are required")
	}
	permissionValues, err := normalizeBuildingPermissions(params.Permissions)
	if err != nil {
		return nil, err
	}
	permissions, _ := json.Marshal(permissionValues)
	if err := s.ownerCanManageBuilding(ctx, ownerUserID, buildingID); err != nil {
		return nil, err
	}

	var existingByPhone, existingByEmail model.User
	phoneErr := s.runtime.DB.WithContext(ctx).Where("phone_country_code = ? AND phone_number = ?", phoneCountryCode, phoneNumber).First(&existingByPhone).Error
	emailErr := s.runtime.DB.WithContext(ctx).Table("users").Joins("JOIN user_credentials ON user_credentials.user_id = users.id").Where("user_credentials.email = ?", email).First(&existingByEmail).Error
	if phoneErr != nil && !errors.Is(phoneErr, gorm.ErrRecordNotFound) || emailErr != nil && !errors.Is(emailErr, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to locate invited account")
	}
	if phoneErr == nil && emailErr == nil && existingByPhone.ID != existingByEmail.ID {
		return nil, errcode.New(errcode.CodeValidationError, "phone and email belong to different accounts")
	}
	if phoneErr == nil && emailErr != nil || phoneErr != nil && emailErr == nil {
		return nil, errcode.New(errcode.CodeValidationError, "phone and email must belong to the same account")
	}

	grant := model.BuildingAuthorization{
		PublicID: utils.NewPublicID(), BuildingID: buildingID, OwnerUserID: ownerUserID,
		PhoneCountryCode: phoneCountryCode, PhoneNumber: phoneNumber, Email: email,
		Permissions: permissions, Status: BuildingAuthorizationStatusActive,
	}
	result := map[string]any{"authorization_id": grant.PublicID, "building_id": buildingID, "permissions": permissionValues, "status": grant.Status, "account_created": false}
	if phoneErr != nil && emailErr != nil {
		grant.Status = BuildingAuthorizationStatusPending
		grant.InvitationExpiresAt = ptrTime(s.runtime.Now().Add(72 * time.Hour))
		token := utils.NewPublicID() + utils.NewPublicID()
		grant.InvitationTokenHash = hashAuthorizationToken(token)
		if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			user := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: phoneCountryCode, PhoneNumber: phoneNumber, MemberStatus: "pending_activation", MemberType: MemberTypeUser, IsVerifiedPhone: false}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			credential := model.UserCredential{UserID: user.ID, Email: &email, IsVerified: false}
			if err := tx.Create(&credential).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.UserProfile{UserID: user.ID, DisplayName: email}).Error; err != nil {
				return err
			}
			grant.GranteeUserID = &user.ID
			if err := tx.Create(&grant).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create building invitation")
		}
		result["status"] = grant.Status
		result["account_created"] = true
		s.sendAuthorizationEmail(ctx, email, "[AJO Living] 大廈副戶設定邀請", fmt.Sprintf("您已獲邀請使用 AJO Living 大廈功能。請於 72 小時內開啟以下連結設定用戶 ID 及密碼：\n\n%s/auth/building-authorizations/accept?token=%s\n\nAJO Living", strings.TrimRight(s.runtime.Config.AppPublicBaseURL, "/"), token))
		return result, nil
	}

	grant.GranteeUserID = &existingByPhone.ID
	if err := s.runtime.DB.WithContext(ctx).Where("owner_user_id = ? AND building_id = ? AND grantee_user_id = ? AND status = ?", ownerUserID, buildingID, existingByPhone.ID, BuildingAuthorizationStatusActive).FirstOrCreate(&grant).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save building authorization")
	}
	if err := s.runtime.DB.WithContext(ctx).Model(&grant).Updates(map[string]any{"permissions": permissions, "status": BuildingAuthorizationStatusActive, "email": email, "phone_country_code": phoneCountryCode, "phone_number": phoneNumber}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update building authorization")
	}
	result["authorization_id"] = grant.PublicID
	result["status"] = grant.Status
	s.sendAuthorizationEmail(ctx, email, "[AJO Living] 大廈副戶授權通知", fmt.Sprintf("您已獲授權查看或使用大廈（%s）的指定功能：%s。請登入 AJO Living 查看。\n\nAJO Living", buildingID, strings.Join(permissionValues, ", ")))
	return result, nil
}

// 7. List returns authorizations owned by the current member for one building.
func (s *BuildingAuthorizationService) List(ctx context.Context, ownerUserID int64, buildingID string) ([]map[string]any, error) {
	var records []model.BuildingAuthorization
	query := s.runtime.DB.WithContext(ctx).Where("owner_user_id = ? AND status <> ?", ownerUserID, BuildingAuthorizationStatusRevoked)
	if strings.TrimSpace(buildingID) != "" {
		query = query.Where("building_id = ?", strings.TrimSpace(buildingID))
	}
	if err := query.Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load building authorizations")
	}
	result := make([]map[string]any, 0, len(records))
	for _, record := range records {
		result = append(result, authorizationResponse(&record))
	}
	return result, nil
}

// 8. Revoke revokes one owner-created authorization.
func (s *BuildingAuthorizationService) Revoke(ctx context.Context, ownerUserID int64, publicID string) error {
	result := s.runtime.DB.WithContext(ctx).Model(&model.BuildingAuthorization{}).Where("public_id = ? AND owner_user_id = ? AND status <> ?", strings.TrimSpace(publicID), ownerUserID, BuildingAuthorizationStatusRevoked).Updates(map[string]any{"status": BuildingAuthorizationStatusRevoked})
	if result.Error != nil {
		return errcode.New(errcode.CodeInternalError, "failed to revoke building authorization")
	}
	if result.RowsAffected == 0 {
		return errcode.New(errcode.CodeNotFound, "building authorization not found")
	}
	return nil
}

// 9. Accept activates an invited account and sets its first local credentials.
func (s *BuildingAuthorizationService) Accept(ctx context.Context, params BuildingAuthorizationAcceptParams) error {
	token := strings.TrimSpace(params.Token)
	username := normalizeUsername(params.Username)
	password := strings.TrimSpace(params.Password)
	if token == "" || !isValidUsername(username) || len(password) < 8 {
		return errcode.New(errcode.CodeValidationError, "valid invitation token, username and password are required")
	}
	hash := hashAuthorizationToken(token)
	var grant model.BuildingAuthorization
	if err := s.runtime.DB.WithContext(ctx).Where("invitation_token_hash = ?", hash).First(&grant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.CodeNotFound, "invitation not found")
		}
		return errcode.New(errcode.CodeInternalError, "failed to load invitation")
	}
	if grant.Status != BuildingAuthorizationStatusPending || grant.InvitationExpiresAt == nil || s.runtime.Now().After(*grant.InvitationExpiresAt) {
		return errcode.New(errcode.CodeExpired, "invitation is expired")
	}
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	passwordEncrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, password)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare password")
	}
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var credential model.UserCredential
		if err := tx.Where("user_id = ?", *grant.GranteeUserID).First(&credential).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.UserCredential{}).Where("username = ? AND user_id <> ?", username, *grant.GranteeUserID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.New(errcode.CodeValidationError, "username is already registered")
		}
		if err := tx.Model(&credential).Updates(map[string]any{"username": username, "password_hash": passwordHash, "password_encrypted": passwordEncrypted, "is_verified": true}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.User{}).Where("id = ?", *grant.GranteeUserID).Update("member_status", "active").Error; err != nil {
			return err
		}
		now := s.runtime.Now()
		return tx.Model(&grant).Updates(map[string]any{"status": BuildingAuthorizationStatusActive, "accepted_at": now, "invitation_token_hash": ""}).Error
	}); err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return errcode.New(errcode.CodeInternalError, "failed to accept invitation")
	}
	return nil
}

// 10. RequirePermission checks local grantee capability; owners retain existing access.
func (s *BuildingAuthorizationService) RequirePermission(ctx context.Context, userID int64, buildingID string, permission string) error {
	var records []model.BuildingAuthorization
	if err := s.runtime.DB.WithContext(ctx).Where("grantee_user_id = ? AND building_id = ? AND status = ?", userID, strings.TrimSpace(buildingID), BuildingAuthorizationStatusActive).Find(&records).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load building authorization")
	}
	for _, record := range records {
		var permissions []string
		_ = json.Unmarshal(record.Permissions, &permissions)
		for _, value := range permissions {
			if value == permission {
				return nil
			}
		}
	}
	return errcode.New(errcode.CodeAuthForbidden, "building feature is not authorized")
}

// 11. GrantedBuildingIDs returns active building IDs granted to one member.
func (s *BuildingAuthorizationService) GrantedBuildingIDs(ctx context.Context, userID int64) ([]string, error) {
	var buildingIDs []string
	if err := s.runtime.DB.WithContext(ctx).Model(&model.BuildingAuthorization{}).
		Where("grantee_user_id = ? AND status = ?", userID, BuildingAuthorizationStatusActive).
		Distinct("building_id").Pluck("building_id", &buildingIDs).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load building authorizations")
	}
	return normalizeStringSlice(buildingIDs), nil
}

// 12. ownerCanManageBuilding ensures the owner can see the requested iSmart building.
func (s *BuildingAuthorizationService) ownerCanManageBuilding(ctx context.Context, userID int64, buildingID string) error {
	account, err := NewIsmartExternalService(s.runtime).loadIsmartAccount(ctx, userID)
	if err != nil {
		return err
	}
	if !containsString(NewIsmartExternalService(s.runtime).visibleBuildingIDs(account), buildingID) {
		return errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}
	return nil
}

// 13. normalizeBuildingPermissions validates the member-facing building menu codes.
func normalizeBuildingPermissions(values []string) ([]string, error) {
	allowed := map[string]bool{}
	for _, item := range NewBuildingAuthorizationService(nil).AvailablePermissions() {
		allowed[item["code"]] = true
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || !allowed[value] || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "at least one building permission is required")
	}
	return normalized, nil
}

// 14. authorizationResponse maps private authorization data to a stable API shape.
func authorizationResponse(record *model.BuildingAuthorization) map[string]any {
	var permissions []string
	_ = json.Unmarshal(record.Permissions, &permissions)
	return map[string]any{"authorization_id": record.PublicID, "building_id": record.BuildingID, "phone_country_code": record.PhoneCountryCode, "phone_number": record.PhoneNumber, "email": record.Email, "permissions": permissions, "status": record.Status, "grantee_user_id": record.GranteeUserID, "invitation_expires_at": record.InvitationExpiresAt}
}

// 15. sendAuthorizationEmail keeps email delivery non-blocking for committed grants.
func (s *BuildingAuthorizationService) sendAuthorizationEmail(ctx context.Context, recipient, subject, body string) {
	if s.runtime.MailSender == nil {
		return
	}
	if err := s.runtime.MailSender.Send(ctx, recipient, subject, body); err != nil && s.runtime.Logger != nil {
		s.runtime.Logger.Warn("failed to send building authorization email", "recipient", recipient, "error", err)
	}
}

// 16. hashAuthorizationToken stores only a digest of invitation credentials.
func hashAuthorizationToken(token string) string {
	value := sha256.Sum256([]byte(token))
	return hex.EncodeToString(value[:])
}

// 17. ptrTime returns a pointer to one timestamp.
func ptrTime(value time.Time) *time.Time { return &value }
