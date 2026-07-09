/*
 * POS 物業繳費安全回歸測試。
 * 1. 驗證線下繳費 payload 必須與 AJO 單位上下文一致。
 * 2. 驗證 H5 訂單可見性缺少單位時拒絕通過。
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestEnsureReportPayloadContextNormalizesContextAndAmount verifies offline report payload normalization.
func TestEnsureReportPayloadContextNormalizesContextAndAmount(t *testing.T) {
	payload := map[string]any{
		"FINAL_AMOUNT":   12345,
		"ENTRY_DATETIME": "2026-06-02 10:00:00",
		"TRAN_DATETIME":  "2026-06-02 10:01:00",
		"TRAN_REF_NO":    "REF-001",
		"PAY_METHOD":     "POS_CASH",
		"BILL_OBJS": []any{
			map[string]any{"invoice_no": "INV-001", "net_amount": 12345},
		},
		"BLG_ID":  "BLG-001",
		"UNIT_ID": "UNIT-001",
	}
	contextValue := &POSUnitContext{BuildingID: "BLG-001", UnitID: "UNIT-001"}

	if err := ensureReportPayloadContext(payload, contextValue); err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
	if payload["FINAL_AMOUNT"] != "12345" {
		t.Fatalf("expected FINAL_AMOUNT string, got %#v", payload["FINAL_AMOUNT"])
	}
	if payload["BLG_ID"] != "BLG-001" || payload["UNIT_ID"] != "UNIT-001" {
		t.Fatalf("expected normalized POS context, got %#v", payload)
	}
}

// 2. TestEnsureReportPayloadContextRejectsConflictingUnit verifies mismatched payload context is blocked.
func TestEnsureReportPayloadContextRejectsConflictingUnit(t *testing.T) {
	payload := map[string]any{
		"FINAL_AMOUNT":   "1000",
		"ENTRY_DATETIME": "2026-06-02 10:00:00",
		"TRAN_DATETIME":  "2026-06-02 10:01:00",
		"TRAN_REF_NO":    "REF-002",
		"PAY_METHOD":     "POS_BANK",
		"BILL_OBJS": []map[string]any{
			{"invoice_no": "INV-002", "net_amount": 1000},
		},
		"BLG_ID":  "BLG-001",
		"UNIT_ID": "UNIT-OTHER",
	}
	contextValue := &POSUnitContext{BuildingID: "BLG-001", UnitID: "UNIT-001"}

	if err := ensureReportPayloadContext(payload, contextValue); err == nil {
		t.Fatal("expected conflicting unit to be rejected")
	}
}

// 3. TestH5OrderMatchesContextRequiresUnit verifies H5 visibility fails closed without unit context.
func TestH5OrderMatchesContextRequiresUnit(t *testing.T) {
	contextValue := &POSUnitContext{BuildingID: "BLG-001", UnitID: "UNIT-001"}
	matchingPayload := map[string]any{
		"data": map[string]any{
			"request_payload": map[string]any{
				"blg_id":  "BLG-001",
				"unit_id": "UNIT-001",
			},
		},
	}
	missingContextPayload := map[string]any{
		"data": map[string]any{
			"mch_order_no": "ORDER-001",
		},
	}
	mismatchedPayload := map[string]any{
		"data": map[string]any{
			"request_payload": map[string]any{
				"report_payment_data": map[string]any{
					"BLG_ID":  "BLG-001",
					"UNIT_ID": "UNIT-OTHER",
				},
			},
		},
	}

	if !h5OrderMatchesContext(matchingPayload, contextValue) {
		t.Fatal("expected matching order to be visible")
	}
	if h5OrderMatchesContext(missingContextPayload, contextValue) {
		t.Fatal("expected missing unit context to be rejected")
	}
	if h5OrderMatchesContext(mismatchedPayload, contextValue) {
		t.Fatal("expected mismatched unit to be rejected")
	}
}

// 4. TestPOSTerminalHelpers verifies terminal proxy safety helpers.
func TestPOSTerminalHelpers(t *testing.T) {
	if formatPOSTerminalAmount(12345) != "000000012345" {
		t.Fatalf("expected padded terminal amount, got %s", formatPOSTerminalAmount(12345))
	}
	if posTerminalBusinessID("POS_CARD") != "100100001" {
		t.Fatal("expected card business id")
	}
	if posTerminalBusinessID("POS_BANK") != "" {
		t.Fatal("expected unsupported terminal pay type to be rejected")
	}
	if !posTerminalBuildingAllowed("BLG-001,BLG-002", "BLG-002") {
		t.Fatal("expected listed building to be allowed")
	}
	if posTerminalBuildingAllowed("BLG-001", "BLG-002") {
		t.Fatal("expected unlisted building to be rejected")
	}
	if posH5SimulationAllowed("production") {
		t.Fatal("expected production simulation to be disabled")
	}
	if !posH5SimulationAllowed("development") {
		t.Fatal("expected development simulation to be enabled")
	}
}

// 5. TestPOSPaymentClientType verifies Staff and member POS fee or bill scopes.
func TestPOSPaymentClientType(t *testing.T) {
	if posPaymentClientType(nil) != "web_client" {
		t.Fatal("expected nil account to use web_client")
	}
	if posPaymentClientType(&model.UserIsmartAccount{}) != "web_client" {
		t.Fatal("expected member account to use web_client")
	}
	if posPaymentClientType(&model.UserIsmartAccount{IsStaff: true}) != "web_staff" {
		t.Fatal("expected staff account to use web_staff")
	}
}

// 6. TestSanitizePOSGatewayRequestOverrides verifies only Alipay wallet type is accepted.
func TestSanitizePOSGatewayRequestOverrides(t *testing.T) {
	aliOverrides := sanitizePOSGatewayRequestOverrides("ALI_H5", map[string]any{
		"walletType": "hk",
		"clientIp":   "1.2.3.4",
	})
	if aliOverrides["walletType"] != "HK" {
		t.Fatalf("expected normalized wallet type, got %#v", aliOverrides)
	}
	if _, ok := aliOverrides["clientIp"]; ok {
		t.Fatal("expected client supplied clientIp to be ignored")
	}

	wxOverrides := sanitizePOSGatewayRequestOverrides("WX_H5", map[string]any{"walletType": "CN"})
	if len(wxOverrides) != 0 {
		t.Fatalf("expected non-Alipay overrides to be empty, got %#v", wxOverrides)
	}
}

// 7. TestNormalizePOSPaymentScene verifies old cart and billing order scenes are preserved.
func TestNormalizePOSPaymentScene(t *testing.T) {
	if normalizePOSPaymentScene("billing") != "billing" {
		t.Fatal("expected billing scene")
	}
	if normalizePOSPaymentScene("cart") != "cart" {
		t.Fatal("expected cart scene")
	}
	if normalizePOSPaymentScene("") != "cart" {
		t.Fatal("expected blank scene to default to cart")
	}
}

// 8. TestBuildH5OrderPayloadKeepsOldReportFields verifies old POS post-process fields.
func TestBuildH5OrderPayloadKeepsOldReportFields(t *testing.T) {
	serviceValue := &POSPaymentService{}
	payload := serviceValue.buildH5OrderPayload(
		&POSUnitContext{BuildingID: "BLG-001", UnitID: "UNIT-001", UnitLabel: "1 / A"},
		&model.UserIsmartAccount{Username: "staff", IsmartUserID: 88},
		POSPaymentOrderCreateParams{
			Scene:                   "cart",
			PayChannel:              "ALI_H5",
			ExpireSeconds:           180,
			FinalAmount:             12345,
			HandleFeeAmount:         100,
			BillObjs:                []map[string]any{{"invoice_no": "INV-001"}},
			HandleFeeObj:            []map[string]any{{"handling_fee": 100}},
			GatewayRequestOverrides: map[string]any{"walletType": "CN"},
			ClientIP:                "127.0.0.1",
		},
	)

	if payload["scene"] != "cart" {
		t.Fatalf("expected cart scene, got %#v", payload["scene"])
	}
	if paymentInt64Value(payload["expire_seconds"]) != 180 {
		t.Fatalf("expected 180 second payment window, got %#v", payload["expire_seconds"])
	}
	reportData := paymentMapValue(payload["report_payment_data"])
	if paymentInt64Value(reportData["FINAL_AMOUNT"]) != 12345 || reportData["CURRENCY"] != posPaymentDefaultCurrency {
		t.Fatalf("expected report amount and currency, got %#v", reportData)
	}
	if reportData["bank_account_received"] != "" {
		t.Fatalf("expected blank bank account for H5 report, got %#v", reportData)
	}
	overrides := paymentMapValue(payload["gateway_request_overrides"])
	if overrides["walletType"] != "CN" || overrides["clientIp"] != "127.0.0.1" {
		t.Fatalf("expected wallet type and server client ip, got %#v", overrides)
	}
}

// 8.1 TestPOSPaymentRewardPointsUsesOnePercent verifies AJO Coin reward ratio.
func TestPOSPaymentRewardPointsUsesOnePercent(t *testing.T) {
	if posPaymentRewardPoints(map[string]any{"amount_hkd": "20850"}) != 208 {
		t.Fatal("expected one percent reward from HKD amount")
	}
	if posPaymentRewardPoints(map[string]any{"FINAL_AMOUNT": "2085000"}) != 208 {
		t.Fatal("expected one percent reward from cents amount")
	}
}

// 8.2 TestPOSReportPaymentRewardKeyPrefersReceipt verifies offline reward idempotency.
func TestPOSReportPaymentRewardKeyPrefersReceipt(t *testing.T) {
	key := posReportPaymentRewardKey(
		map[string]any{"TRAN_REF_NO": "REF-001", "FINAL_AMOUNT": "2085000"},
		map[string]any{"ismart_receipt_no": map[string]any{"receipt_id": "RCPT-001"}},
	)
	if key != "RCPT-001" {
		t.Fatalf("expected receipt id key, got %s", key)
	}
}

// 9. TestChargeAllinpayTerminalPostsExpectedForm verifies terminal request format.
func TestChargeAllinpayTerminalPostsExpectedForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", request.Method)
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("expected form request, got %v", err)
		}
		if request.Form.Get("BUSINESS_ID") != "100100001" {
			t.Fatalf("expected card business id, got %s", request.Form.Get("BUSINESS_ID"))
		}
		if request.Form.Get("AMOUNT") != "000000012345" {
			t.Fatalf("expected padded amount, got %s", request.Form.Get("AMOUNT"))
		}
		if request.Form.Get("CURRENCY") != posPaymentDefaultCurrency {
			t.Fatalf("expected currency, got %s", request.Form.Get("CURRENCY"))
		}
		response.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = response.Write([]byte("TRAN_REF_NO=TERM-001&ENTRY_DATETIME=2026-06-02+10%3A00%3A00"))
	}))
	defer server.Close()

	serviceValue := &POSPaymentService{
		runtime: &Runtime{
			Config: &config.Config{POSLoginTimeout: time.Second},
			Now: func() time.Time {
				return time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
			},
		},
	}
	payload, rawResponse, err := serviceValue.chargeAllinpayTerminal(
		context.Background(),
		server.URL,
		&POSUnitContext{BuildingID: "BLG-001"},
		POSTerminalPaymentParams{PayType: "POS_CARD", FinalAmount: 12345},
	)
	if err != nil {
		t.Fatalf("expected terminal charge to pass, got %v", err)
	}
	if rawResponse == "" || payload["TRAN_REF_NO"] != "TERM-001" {
		t.Fatalf("expected parsed terminal response, got raw=%q payload=%#v", rawResponse, payload)
	}
}

// 10. TestPOSPaymentOverviewUsesStoredPasswordForLoginReadiness verifies token refresh readiness.
func TestPOSPaymentOverviewUsesStoredPasswordForLoginReadiness(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	passwordEncrypted, err := utils.EncryptString(runtimeValue.Config.EncryptionKey, "s61980774")
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234567",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	building := model.Community{
		PublicID:      "BLG-001",
		CommunityType: "building",
		NameZH:        "BLG-001",
		NameEN:        "BLG-001",
		DistrictCode:  "unknown",
		AddressText:   "BLG-001",
	}
	if err := runtimeValue.DB.Create(&building).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	buildingJSON, err := marshalJSON([]string{"BLG-001"})
	if err != nil {
		t.Fatalf("marshal building json: %v", err)
	}
	unitJSON, err := marshalJSON([]string{"BLG-0010000101"})
	if err != nil {
		t.Fatalf("marshal unit json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &building.ID,
		BoundBuildingIDs:   buildingJSON,
		BoundFlatUnitIDs:   unitJSON,
		ResidenceFloor:     "1",
		ResidenceUnit:      "01",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       88,
		Username:                           "patrick",
		Email:                              "patrick@example.com",
		Phone:                              "+85261234567",
		ClientBuildingPermissions:          buildingJSON,
		ClientBuildingFlatUnitsPermissions: unitJSON,
		Building:                           []byte("[]"),
		StaffBuildingPermissions:           []byte("[]"),
		RawMessage:                         []byte("{}"),
		PasswordEncrypted:                  passwordEncrypted,
	}).Error; err != nil {
		t.Fatalf("create ismart account: %v", err)
	}

	overview, err := NewPOSPaymentService(runtimeValue).Overview(context.Background(), user.ID, POSPaymentSelection{}, false)
	if err != nil {
		t.Fatalf("load overview: %v", err)
	}
	if overview.POSLoginRequired {
		t.Fatalf("expected stored password to satisfy POS login readiness")
	}
	if overview.Context == nil || overview.Context.BuildingID != "BLG-001" || overview.Context.UnitID != "BLG-0010000101" {
		t.Fatalf("expected synced POS context, got %+v", overview.Context)
	}
}

// 11. TestPOSPaymentHistoryCanonicalizesMemberUnitID verifies POS history uses the real relay unit id.
func TestPOSPaymentHistoryCanonicalizesMemberUnitID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/building/0241100/units":
			_, _ = response.Write([]byte(`[{"unit_id":"02411000313","floor":"03","unit":"C"}]`))
		case request.Method == http.MethodPost && request.URL.Path == "/transactions/flat_units":
			var payload map[string][]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode history payload: %v", err)
			}
			unitIDs := payload["unit_id_list"]
			if len(unitIDs) != 1 || unitIDs[0] != "02411000313" {
				t.Fatalf("expected canonical POS unit id, got %#v", payload)
			}
			_, _ = response.Write([]byte(`{"payment_objs":{"payment_objs":[{"payment_id":"PAY-001","receipt_id":"20135605","input_time":"2026-04-21 11:35:38","tran_time":"2026-04-16 03:35:00","trs_val":6200,"status":"confirmed"}]}}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{POSAPIBaseURL: server.URL, POSLoginTimeout: time.Second},
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	tokenEncrypted, err := utils.EncryptString(runtimeValue.Config.EncryptionKey, "relay-token")
	if err != nil {
		t.Fatalf("encrypt relay token: %v", err)
	}
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234568",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	building := model.Community{
		PublicID:      "0241100",
		CommunityType: "building",
		NameZH:        "協和大廈",
		DistrictCode:  "unknown",
		AddressText:   "協和大廈",
	}
	if err := runtimeValue.DB.Create(&building).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	buildingJSON, err := marshalJSON([]string{"0241100"})
	if err != nil {
		t.Fatalf("marshal building json: %v", err)
	}
	unitJSON, err := marshalJSON([]string{"03/C"})
	if err != nil {
		t.Fatalf("marshal unit json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &building.ID,
		BoundBuildingIDs:   buildingJSON,
		BoundFlatUnitIDs:   unitJSON,
		ResidenceFloor:     "03",
		ResidenceUnit:      "C",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       89,
		Username:                           "owner-c",
		ClientBuildingPermissions:          buildingJSON,
		ClientBuildingFlatUnitsPermissions: unitJSON,
		Building:                           []byte("[]"),
		StaffBuildingPermissions:           []byte("[]"),
		RawMessage:                         []byte("{}"),
		RelayTokenEncrypted:                tokenEncrypted,
	}).Error; err != nil {
		t.Fatalf("create ismart account: %v", err)
	}

	result, err := NewPOSPaymentService(runtimeValue).ListHistoryByQuery(context.Background(), user.ID, POSPaymentHistoryQuery{
		Selection: POSPaymentSelection{UnitID: "02411000313"},
	})
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if result.Context == nil || result.Context.UnitID != "02411000313" {
		t.Fatalf("expected canonical context, got %+v", result.Context)
	}
	if len(result.Items) != 1 || result.Items[0]["payment_id"] != "PAY-001" {
		t.Fatalf("expected nested history row, got %#v", result.Items)
	}
}

// 12. TestPOSPaymentHistoryInfersStaffBuildingFromUnitID verifies Staff unit history does not use the first building.
func TestPOSPaymentHistoryInfersStaffBuildingFromUnitID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/building/0999900/units":
			_, _ = response.Write([]byte(`[{"unit_id":"09999000211","floor":"02","unit":"A"}]`))
		case request.Method == http.MethodPost && request.URL.Path == "/transactions/flat_units":
			var payload map[string][]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode history payload: %v", err)
			}
			unitIDs := payload["unit_id_list"]
			if len(unitIDs) != 1 || unitIDs[0] != "09999000211" {
				t.Fatalf("expected requested Staff unit id, got %#v", payload)
			}
			_, _ = response.Write([]byte(`{"payment_objs":{"payment_objs":[{"payment_id":"PAY-STAFF-001","receipt_id":"RCPT-STAFF-001"}]}}`))
		default:
			t.Fatalf("unexpected POS request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{POSAPIBaseURL: server.URL, POSLoginTimeout: time.Second},
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	tokenEncrypted, err := utils.EncryptString(runtimeValue.Config.EncryptionKey, "relay-token")
	if err != nil {
		t.Fatalf("encrypt relay token: %v", err)
	}
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234569",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	building := model.Community{
		PublicID:      "0999900",
		CommunityType: "building",
		NameZH:        "測試1大廈",
		DistrictCode:  "unknown",
		AddressText:   "測試1大廈",
	}
	if err := runtimeValue.DB.Create(&building).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	buildingJSON, err := marshalJSON([]string{"0031300", "0999900"})
	if err != nil {
		t.Fatalf("marshal building json: %v", err)
	}
	unitJSON, err := marshalJSON([]string{"09999000211"})
	if err != nil {
		t.Fatalf("marshal unit json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &building.ID,
		BoundBuildingIDs:   buildingJSON,
		BoundFlatUnitIDs:   unitJSON,
		ResidenceFloor:     "02",
		ResidenceUnit:      "A",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       90,
		Username:                           "staff-owner-a",
		IsStaff:                            true,
		Building:                           buildingJSON,
		StaffBuildingPermissions:           buildingJSON,
		ClientBuildingPermissions:          []byte("[]"),
		ClientBuildingFlatUnitsPermissions: []byte("[]"),
		RawMessage:                         []byte("{}"),
		RelayTokenEncrypted:                tokenEncrypted,
	}).Error; err != nil {
		t.Fatalf("create ismart account: %v", err)
	}

	result, err := NewPOSPaymentService(runtimeValue).ListHistoryByQuery(context.Background(), user.ID, POSPaymentHistoryQuery{
		Selection: POSPaymentSelection{UnitID: "09999000211"},
	})
	if err != nil {
		t.Fatalf("load staff unit history: %v", err)
	}
	if result.Context == nil || result.Context.BuildingID != "0999900" || result.Context.UnitID != "09999000211" {
		t.Fatalf("expected unit-derived Staff context, got %+v", result.Context)
	}
	if len(result.Items) != 1 || result.Items[0]["payment_id"] != "PAY-STAFF-001" {
		t.Fatalf("expected Staff unit history row, got %#v", result.Items)
	}
}
