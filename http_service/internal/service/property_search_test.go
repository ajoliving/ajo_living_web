/*
 * 樓盤搜尋條件回歸測試。
 * 1. 驗證樓盤租售搜尋覆蓋公開可搜尋欄位。
 * 2. 驗證服務式住宅搜尋使用大小寫不敏感匹配。
 * 3. 驗證服務式住宅房型欄位可正確派生公開排序與詳情值。
 * 4. 驗證發布身份可輸出正式角色標籤。
 */
package service

import (
	"strings"
	"testing"
)

// 1. TestPropertySaleKeywordSearchCondition verifies sale keyword search fields.
func TestPropertySaleKeywordSearchCondition(t *testing.T) {
	condition, argCount := propertyKeywordSearchCondition(PropertyChannelSale)

	if argCount != 10 {
		t.Fatalf("expected 10 keyword args, got %d", argCount)
	}
	for _, fragment := range []string{
		"listings.title ILIKE ?",
		"property_sale_listings.property_no ILIKE ?",
		"property_sale_listings.estate_name ILIKE ?",
		"property_sale_listings.address_text ILIKE ?",
		"property_sale_listings.block_name ILIKE ?",
		"property_sale_listings.unit_name ILIKE ?",
		"property_sale_listings.public_location_text ILIKE ?",
	} {
		if !strings.Contains(condition, fragment) {
			t.Fatalf("expected sale keyword condition to contain %q, got %s", fragment, condition)
		}
	}
	if strings.Contains(condition, "community_name") {
		t.Fatalf("sale keyword condition should not reference missing community_name column: %s", condition)
	}
}

// 2. TestServicedKeywordSearchCondition verifies serviced keyword search fields.
func TestServicedKeywordSearchCondition(t *testing.T) {
	condition, argCount := propertyKeywordSearchCondition(PropertyChannelServiced)

	if argCount != 7 {
		t.Fatalf("expected 7 keyword args, got %d", argCount)
	}
	for _, fragment := range []string{
		"listings.title ILIKE ?",
		"serviced_apartment_projects.project_name ILIKE ?",
		"serviced_apartment_projects.project_name_en ILIKE ?",
		"serviced_apartment_projects.address_text ILIKE ?",
		"serviced_apartment_projects.address_text_en ILIKE ?",
	} {
		if !strings.Contains(condition, fragment) {
			t.Fatalf("expected serviced keyword condition to contain %q, got %s", fragment, condition)
		}
	}
}

// 3. TestPropertyDistrictsForRegion verifies region filter expansion.
func TestPropertyDistrictsForRegion(t *testing.T) {
	districts, ok := propertyDistrictsForRegion("kowloon")
	if !ok {
		t.Fatal("expected kowloon region to be supported")
	}
	for _, expected := range []string{"kowloon", "yau_tsim_mong", "kwun_tong", "tsim_sha_tsui", "kowloon_bay"} {
		found := false
		for _, district := range districts {
			if district == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected district %s in kowloon region: %#v", expected, districts)
		}
	}
	if _, ok := propertyDistrictsForRegion("unknown"); ok {
		t.Fatal("unknown region should not be supported")
	}
	if !isAllowedPropertyDistrict("tseung_kwan_o") {
		t.Fatal("expected property subdistrict code to be accepted")
	}
	if isAllowedPropertyDistrict("unknown") {
		t.Fatal("unknown property district should not be accepted")
	}
	for _, testcase := range []struct {
		value    string
		expected []string
	}{
		{value: "kwun_tong", expected: []string{"kwun_tong", "kowloon"}},
		{value: "kowloon_bay", expected: []string{"kowloon_bay", "kowloon"}},
	} {
		codes := propertyDistrictSearchCodes(testcase.value)
		for _, expected := range testcase.expected {
			found := false
			for _, code := range codes {
				if code == expected {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected search code %s for %s, got %#v", expected, testcase.value, codes)
			}
		}
	}
}

// 4. TestDeriveServicedApartmentFieldsFromRoomTypes verifies serviced room field closure.
func TestDeriveServicedApartmentFieldsFromRoomTypes(t *testing.T) {
	params := UpsertServicedApartmentParams{
		LowestMonthlyRentHKD: 32000,
		LowestDailyRentHKD:   1500,
		MinLeaseMonths:       6,
		MinStayValue:         6,
		MinStayUnit:          "month",
		RoomTypes: []ServicedApartmentRoomTypeInput{
			{
				Name:              "Long stay suite",
				UsableAreaSqft:    520,
				MonthlyRentMinHKD: 28800,
				MonthlyRentMaxHKD: 36000,
				DailyRentMinHKD:   1200,
				DailyRentMaxHKD:   1800,
				IncludedFees:      true,
				IncludedFeeItems:  []string{"housekeeping", "linen"},
				MinLeaseMonths:    1,
				MinStayValue:      1,
				MinStayUnit:       "month",
			},
		},
	}

	derived := deriveServicedApartmentFields(params)
	if derived.LowestMonthlyRentHKD != 28800 {
		t.Fatalf("expected lowest monthly rent 28800, got %v", derived.LowestMonthlyRentHKD)
	}
	if derived.LowestDailyRentHKD != 1200 {
		t.Fatalf("expected lowest daily rent 1200, got %v", derived.LowestDailyRentHKD)
	}
	if derived.MinUsableAreaSqft != 520 {
		t.Fatalf("expected min usable area 520, got %d", derived.MinUsableAreaSqft)
	}
	if derived.MinLeaseMonths != 1 {
		t.Fatalf("expected min lease months 1, got %d", derived.MinLeaseMonths)
	}
	if derived.MinStayValue != 1 || derived.MinStayUnit != "month" {
		t.Fatalf("expected min stay 1 month, got %d %s", derived.MinStayValue, derived.MinStayUnit)
	}
}

// 5. TestPublisherRoleLabel verifies owner and agent display labels.
func TestPublisherRoleLabel(t *testing.T) {
	cases := map[string]string{
		"owner":               "業主",
		"agent":               "代理人",
		"professional_seller": "代理人",
		"":                    "業主",
	}

	for input, expected := range cases {
		if actual := publisherRoleLabel(input); actual != expected {
			t.Fatalf("expected publisher role %q for %q, got %q", expected, input, actual)
		}
	}
}
