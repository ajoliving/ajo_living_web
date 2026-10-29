/*
 * iSmart 副戶註冊與授權服務。
 * 1. 解析目標電話及電郵是否屬於同一個既有 iSmart 帳戶。
 * 2. 沒有既有帳戶時使用手動輸入的密碼，建立 iSmart 帳戶後再授權單位。
 * 3. 授權完成後以 Redis/POS 目錄解析大廈及單位名稱，發送正式通知電郵。
 */
package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. IsmartSubaccountRegisterGrantParams defines the register-and-grant request.
type IsmartSubaccountRegisterGrantParams struct {
	UnitID    string
	Phone     string
	Email     string
	EngName   string
	Password  string
	SendEmail bool
	Remark    string
}

// 2. IsmartSubaccountRegisterGrantResult reports the committed grant and notification state.
type IsmartSubaccountRegisterGrantResult struct {
	AccountCreated  bool   `json:"account_created"`
	TargetUserID    int64  `json:"target_user_id"`
	TargetUsername  string `json:"target_username,omitempty"`
	BuildingName    string `json:"building_name,omitempty"`
	UnitDisplay     string `json:"unit_display,omitempty"`
	EmailSent       bool   `json:"email_sent"`
	UpstreamMessage string `json:"upstream_message,omitempty"`
}

// 3. RegisterAndGrantSubaccount creates an iSmart account when needed and grants the current unit.
func (s *IsmartExternalService) RegisterAndGrantSubaccount(ctx context.Context, ownerUserID int64, params IsmartSubaccountRegisterGrantParams) (*IsmartSubaccountRegisterGrantResult, error) {
	unitID := strings.TrimSpace(params.UnitID)
	phone := normalizeSubaccountRegistrationPhone(params.Phone)
	email := normalizeEmail(params.Email)
	if unitID == "" || phone == "" || email == "" {
		return nil, errcode.New(errcode.CodeValidationError, "unit_id, phone and email are required")
	}
	if !isValidEmail(email) {
		return nil, errcode.New(errcode.CodeValidationError, "valid email is required")
	}
	if err := s.validateSubaccountOwnerUnit(ctx, ownerUserID, unitID); err != nil {
		return nil, err
	}

	contacts, err := s.CheckSubaccountContacts(ctx, IsmartSubaccountContactParams{Phone: phone, Email: email})
	if err != nil {
		return nil, err
	}

	result := &IsmartSubaccountRegisterGrantResult{}
	targetUserID := int64(0)
	switch {
	case contacts.CanAuthorize:
		targetUserID = contacts.EmailUserID
	case contacts.EmailExists || contacts.PhoneExists:
		return nil, errcode.New(errcode.CodeValidationError, "target phone and email must belong to the same existing ismart account")
	default:
		engName := strings.TrimSpace(params.EngName)
		if engName == "" {
			return nil, errcode.New(errcode.CodeValidationError, "eng_name is required when registering a new account")
		}
		if params.Password != strings.TrimSpace(params.Password) || utf8.RuneCountInString(params.Password) < 8 {
			return nil, errcode.New(errcode.CodeValidationError, "password must contain at least 8 characters without leading or trailing whitespace")
		}
		registered, registerErr := s.RegisterDirectAccountOnSubaccountAPI(ctx, IsmartDirectRegistrationParams{
			Phone:          phone,
			Email:          email,
			EngName:        engName,
			LegalEntity:    "NA",
			Remark:         strings.TrimSpace(params.Remark),
			IsReceiveEmail: true,
			Password:       params.Password,
		})
		if registerErr != nil {
			return nil, registerErr
		}
		targetUserID = registered.UserID
		result.AccountCreated = true
		result.TargetUsername = registered.Username
	}

	grant, err := s.GrantSubaccount(ctx, ownerUserID, IsmartSubaccountMutationParams{
		UnitID:       unitID,
		TargetUserID: targetUserID,
		TargetPhone:  phone,
		TargetEmail:  email,
		Remark:       strings.TrimSpace(params.Remark),
	})
	if err != nil {
		return nil, err
	}
	result.TargetUserID = targetUserID
	if result.TargetUsername == "" {
		result.TargetUsername = strings.TrimSpace(paymentStringValue(grant["target_username"]))
	}
	result.UpstreamMessage = strings.TrimSpace(paymentStringValue(grant["upstream_message"]))
	result.BuildingName, result.UnitDisplay = s.resolveSubaccountUnitDisplay(ctx, unitID)

	if params.SendEmail {
		result.EmailSent = s.sendSubaccountEmail(ctx, email, result)
	}
	return result, nil
}

// 4. validateSubaccountOwnerUnit checks the current owner and unit before account creation.
func (s *IsmartExternalService) validateSubaccountOwnerUnit(ctx context.Context, ownerUserID int64, unitID string) error {
	account, err := s.loadIsmartAccount(ctx, ownerUserID)
	if err != nil {
		return err
	}
	if !s.unitVisible(account, unitID) {
		return errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}
	return nil
}

// 5. sendSubaccountEmail sends the post-grant notice without rolling back iSmart changes.
func (s *IsmartExternalService) sendSubaccountEmail(ctx context.Context, recipient string, result *IsmartSubaccountRegisterGrantResult) bool {
	if s.runtime == nil || s.runtime.MailSender == nil {
		return false
	}
	subject := "[AJO Living] iSmart 副戶授權通知"
	body := buildSubaccountEmailBody(result, strings.TrimSpace(recipient))
	if s.runtime.Config != nil {
		loginURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.AppPublicBaseURL), "/") + "/account/login"
		body += fmt.Sprintf("\n登入入口：%s\n", loginURL)
	}
	if result.AccountCreated {
		subject = "[AJO Living] iSmart 帳戶及副戶授權通知"
	}
	if err := s.runtime.MailSender.Send(ctx, recipient, subject, body); err != nil {
		if s.runtime.Logger != nil {
			s.runtime.Logger.Warn("failed to send iSmart subaccount email", "recipient", recipient, "error", err)
		}
		return false
	}
	return true
}

// 6. resolveSubaccountUnitDisplay resolves building and unit labels from the shared POS directory.
func (s *IsmartExternalService) resolveSubaccountUnitDisplay(ctx context.Context, unitID string) (string, string) {
	buildingID := posUnitBuildingID(unitID)
	buildingName := buildingID
	unitDisplay := strings.TrimSpace(unitID)
	if buildingID == "" {
		return buildingName, unitDisplay
	}

	directory := NewPOSBuildingService(s.runtime)
	if buildings, err := directory.ListBuildings(ctx); err == nil {
		for _, item := range buildings {
			candidateID := posBuildingID(item)
			if !sameSubaccountBuildingID(candidateID, buildingID) {
				continue
			}
			buildingName = paymentFirstNonEmpty(item.BuildnameCHI, item.Buildname, item.Name, candidateID)
			break
		}
	}
	if units, err := directory.ListUnits(ctx, buildingID); err == nil {
		for _, item := range units {
			if posUnitID(item) != strings.TrimSpace(unitID) {
				continue
			}
			unitDisplay = paymentFirstNonEmpty(posUnitLabel(item.Floor, posUnitName(item)), unitDisplay)
			break
		}
	}
	return buildingName, unitDisplay
}

// 7. sameSubaccountBuildingID compares padded and unpadded POS building identifiers.
func sameSubaccountBuildingID(left string, right string) bool {
	leftKeys := buildingIDLookupKeys(left)
	rightKeys := buildingIDLookupKeys(right)
	for _, leftKey := range leftKeys {
		if containsString(rightKeys, leftKey) {
			return true
		}
	}
	return false
}

// 8. buildSubaccountEmailBody builds the account or authorization notice.
func buildSubaccountEmailBody(result *IsmartSubaccountRegisterGrantResult, recipient string) string {
	lines := []string{
		"您好，",
		"",
		"您已獲授權使用以下 iSmart 單位功能：",
		fmt.Sprintf("大廈：%s", paymentFirstNonEmpty(result.BuildingName, "目前單位")),
		fmt.Sprintf("單位：%s", paymentFirstNonEmpty(result.UnitDisplay, "已授權單位")),
		"授權範圍：遙距開門及查看大廈通告。",
		"",
		fmt.Sprintf("登入電郵：%s", recipient),
	}
	if result.AccountCreated {
		lines = append(lines,
			fmt.Sprintf("登入用戶 ID：%s", result.TargetUsername),
			"",
			"此為新建立的 iSmart 帳戶。請使用註冊時自行設定的密碼登入。",
		)
	} else {
		lines = append(lines, "請使用您現有的 iSmart 帳戶登入。")
	}
	lines = append(lines, "", "AJO Living")
	return strings.Join(lines, "\n")
}

// 9. normalizeSubaccountRegistrationPhone normalizes common HK input before iSmart registration.
func normalizeSubaccountRegistrationPhone(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return raw
	}
	if strings.HasPrefix(raw, "+") {
		return "+" + digitsOnly(raw)
	}
	digits := digitsOnly(raw)
	if len(digits) == 8 {
		return "+852" + digits
	}
	return raw
}
