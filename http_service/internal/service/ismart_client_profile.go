/*
 * iSmart ClientTbl profile synchronization.
 * 1. Read the linked member's ClientTbl profile through the integration API.
 * 2. Validate the upstream identity before replacing the local display snapshot.
 * 3. Keep the complete sanitized ClientTbl payload scoped to the linked AJO member.
 */
package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. RefreshClientProfile updates the linked member's read-only ClientTbl snapshot.
func (s *IsmartExternalService) RefreshClientProfile(ctx context.Context, userID int64) error {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("user_id", strconv.FormatInt(account.IsmartUserID, 10))
	result, err := s.getIntegration(ctx, "/auth/client/", query)
	if err != nil {
		return err
	}
	data := paymentMapValue(result.Payload)
	if paymentInt64Value(data["user_id"]) != account.IsmartUserID {
		return errcode.New(errcode.CodeInternalError, "ismart client profile identity mismatch")
	}
	client := paymentMapValue(data["client"])
	if len(client) == 0 {
		return errcode.New(errcode.CodeNotFound, "ismart client profile not found")
	}

	profileSnapshot, err := marshalJSON(ismartClientProfileSnapshot(data, client, account))
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to save ismart client profile")
	}
	raw := ismartRawMessage(account.RawMessage)
	raw["client"] = sanitizeIsmartRawMessage(client)
	rawMessage, err := marshalJSON(raw)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to save ismart client profile")
	}

	updates := map[string]any{
		"email":                    normalizeEmail(paymentStringValue(data["email"])),
		"phone":                    strings.TrimSpace(paymentStringValue(data["phone"])),
		"profile_snapshot":         profileSnapshot,
		"raw_message":              rawMessage,
		"client_profile_synced_at": s.runtime.Now(),
	}
	if username := strings.TrimSpace(paymentStringValue(data["username"])); username != "" {
		updates["username"] = username
	}
	if err := s.runtime.DB.WithContext(ctx).Model(&model.UserIsmartAccount{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to save ismart client profile")
	}

	return nil
}

// 2. ismartClientProfileSnapshot maps ClientTbl fields to AJO's stable member-center contract.
func ismartClientProfileSnapshot(data map[string]any, client map[string]any, account *model.UserIsmartAccount) map[string]any {
	snapshot := sanitizeIsmartRawMessage(client)
	snapshot["account_code"] = firstIsmartProfileText(client, data, account.Username, "cli_id", "username")
	snapshot["account_phone"] = firstIsmartProfileText(data, client, account.Phone, "phone", "account_phone")
	snapshot["account_email"] = firstIsmartProfileText(data, client, account.Email, "email", "account_email")
	snapshot["owner_name_en"] = firstIsmartProfileText(client, nil, "", "cli_name")
	snapshot["owner_name_zh"] = firstIsmartProfileText(client, nil, "", "cli_chi_name")
	snapshot["account_name"] = firstIsmartProfileText(client, nil, "", "cli_acname")
	snapshot["identity_number"] = firstIsmartProfileText(client, nil, "", "cli_id_card")
	snapshot["legal_entity"] = firstIsmartProfileText(client, nil, "", "cli_legalentity")
	snapshot["client_type"] = firstIsmartProfileText(client, nil, "", "cli_type")
	snapshot["gender"] = firstIsmartProfileText(client, nil, "", "cli_sex")
	snapshot["birth_date"] = firstIsmartProfileText(client, nil, "", "cli_birthday")
	snapshot["contact_name"] = firstIsmartProfileText(client, nil, "", "cli_contact_person")
	snapshot["contact_phone"] = firstIsmartProfileText(client, nil, account.Phone, "cli_tel")
	snapshot["emergency_contact_name"] = firstIsmartProfileText(client, nil, "", "cli_urgent_contact_person")
	snapshot["emergency_contact_phone"] = firstIsmartProfileText(client, nil, "", "cli_tel2")
	snapshot["billing_phone"] = firstIsmartProfileText(client, nil, account.Phone, "cli_tel")
	snapshot["billing_email"] = firstIsmartProfileText(client, nil, account.Email, "cli_email")
	snapshot["billing_address"] = firstIsmartProfileText(client, nil, "", "cli_addr", "cli_chiadd")
	snapshot["billing_address_en"] = firstIsmartProfileText(client, nil, "", "cli_addr")
	snapshot["billing_address_zh"] = firstIsmartProfileText(client, nil, "", "cli_chiadd")

	return snapshot
}

// 3. firstIsmartProfileText resolves ClientTbl, envelope, and local fallback values in order.
func firstIsmartProfileText(primary map[string]any, secondary map[string]any, fallback string, keys ...string) string {
	for _, source := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			if value := strings.TrimSpace(paymentStringValue(source[key])); value != "" {
				return value
			}
		}
	}

	return strings.TrimSpace(fallback)
}
