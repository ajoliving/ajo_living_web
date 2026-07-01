/*
 * 市場走勢服務測試。
 * 1. 驗證 RVD 私人住宅平均租金 CSV 解析。
 * 2. 驗證最近月份、區域序列與每平方呎換算。
 */
package service

import "testing"

// 1. TestParseMarketRentTrendCSV verifies public rent trend parsing.
func TestParseMarketRentTrendCSV(t *testing.T) {
	payload := []byte(`PRIVATE  DOMESTIC  -  AVERAGE  RENTS  BY  CLASS [MONTHLY],,,,,,,,,,
Month,Class A Hong Kong,Class A Hong Kong - Remarks,Class A Kowloon,Class A Kowloon - Remarks,Class A New Territories,Class A New Territories - Remarks,Class B Hong Kong,Class B Hong Kong - Remarks,Class B Kowloon,Class B Kowloon - Remarks,Class B New Territories,Class B New Territories - Remarks,Class C Hong Kong,Class C Hong Kong - Remarks,Class C Kowloon,Class C Kowloon - Remarks,Class C New Territories,Class C New Territories - Remarks
12-2025,509,,450,,351,,413,,375,,286,,454,,407,,278
01-2026,513,,451,,347,,444,,384,,278,,483,,387,,277
02-2026,507,,467,,343,,427,,396,,280,,460,,395,,291
03-2026,528,P,459,P,343,P,422,P,402,P,289,P,482,P,421,P,295,P
04-2026,518,P,444,P,354,P,428,P,396,P,284,P,472,P,424,P,287,P
05-2026,527,P,452,P,366,P,438,P,405,P,297,P,484,P,433,P,291,P
`)

	result, err := parseMarketRentTrendCSV(payload)
	if err != nil {
		t.Fatalf("parse market rent trend: %v", err)
	}

	if result.UpdatedMonth != "2026-05" {
		t.Fatalf("expected updated month 2026-05, got %s", result.UpdatedMonth)
	}
	if len(result.Months) != 6 || len(result.Regions) != 3 {
		t.Fatalf("expected 6 months and 3 regions, got %d months and %d regions", len(result.Months), len(result.Regions))
	}
	if result.Regions[0].Key != "hk" || result.Regions[0].LatestHKDPerSqft != 44.9 {
		t.Fatalf("unexpected Hong Kong latest value: %+v", result.Regions[0])
	}
	if result.Regions[1].Direction != "up" {
		t.Fatalf("expected Kowloon direction up, got %s", result.Regions[1].Direction)
	}
	if result.Regions[2].Points[0].Month != "2025-12" {
		t.Fatalf("expected first point 2025-12, got %s", result.Regions[2].Points[0].Month)
	}
}
