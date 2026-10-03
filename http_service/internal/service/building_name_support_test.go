/*
 * 大廈顯示名稱解析測試。
 * 1. 驗證大廈 ID 補零查找鍵。
 * 2. 驗證佔位名稱排除與索引回退。
 */
package service

import "testing"

// 1. buildingIDLookupKeys covers padded and trimmed digit variants.
func TestBuildingIDLookupKeysCoversPaddedVariants(t *testing.T) {
	got := buildingIDLookupKeys("419900")
	want := []string{"419900", "0419900"}
	if len(got) != len(want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	for index, item := range want {
		if got[index] != item {
			t.Fatalf("keys[%d] = %q, want %q", index, got[index], item)
		}
	}

	if keys := buildingIDLookupKeys("0419900"); len(keys) != 2 || keys[0] != "0419900" || keys[1] != "419900" {
		t.Fatalf("padded id keys = %v", keys)
	}
	if keys := buildingIDLookupKeys("TWR-A"); len(keys) != 1 || keys[0] != "TWR-A" {
		t.Fatalf("non numeric keys = %v", keys)
	}
	if keys := buildingIDLookupKeys("   "); keys != nil {
		t.Fatalf("blank keys = %v, want nil", keys)
	}
}

// 2. usableBuildingName rejects placeholder names equal to the building id.
func TestUsableBuildingNameRejectsIDPlaceholder(t *testing.T) {
	keys := buildingIDLookupKeys("0031300")
	if got := usableBuildingName("時安大廈", keys); got != "時安大廈" {
		t.Fatalf("real name = %q", got)
	}
	if got := usableBuildingName("0031300", keys); got != "" {
		t.Fatalf("placeholder name = %q, want empty", got)
	}
	if got := usableBuildingName("  ", keys); got != "" {
		t.Fatalf("blank name = %q, want empty", got)
	}
}

// 3. lookupBuildingName resolves a padded stored id from the index.
func TestLookupBuildingNameResolvesPaddedKey(t *testing.T) {
	names := map[string]buildingDisplayName{
		"0999900": {ZH: "測試1大廈", EN: "DEMO1 BUILDING"},
	}

	name, ok := lookupBuildingName(names, "999900")
	if !ok || name.ZH != "測試1大廈" || name.EN != "DEMO1 BUILDING" {
		t.Fatalf("lookup = %+v ok=%v", name, ok)
	}
	if _, ok := lookupBuildingName(names, "0419900"); ok {
		t.Fatalf("unexpected lookup hit for an unrelated building")
	}
}
