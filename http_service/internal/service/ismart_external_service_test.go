/*
 * iSmart 對外介面代理服務測試。
 * 1. 驗證會員中心綁定大廈優先於 iSmart 可見大廈預設順序。
 * 2. 驗證未指定大廈時仍保留 iSmart 可見範圍校驗。
 */
package service

import (
	"context"
	"testing"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestIsmartSelectVisibleBuildingPrefersProfileBuilding verifies member-center binding priority.
func TestIsmartSelectVisibleBuildingPrefersProfileBuilding(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
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
	boundBuilding := model.Community{
		PublicID:      "BLG-002",
		CommunityType: "building",
		NameZH:        "仁英大廈",
		DistrictCode:  "unknown",
	}
	if err := runtimeValue.DB.Create(&boundBuilding).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	profileBuildingJSON, err := marshalJSON([]string{"BLG-002"})
	if err != nil {
		t.Fatalf("marshal profile building json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &boundBuilding.ID,
		BoundBuildingIDs:   profileBuildingJSON,
		ResidenceFloor:     "01",
		ResidenceUnit:      "B",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	visibleBuildingJSON, err := marshalJSON([]string{"BLG-001", "BLG-002"})
	if err != nil {
		t.Fatalf("marshal visible building json: %v", err)
	}
	account := &model.UserIsmartAccount{
		UserID:                    user.ID,
		IsmartUserID:              88,
		Username:                  "patrick",
		ClientBuildingPermissions: visibleBuildingJSON,
		Building:                  []byte("[]"),
		StaffBuildingPermissions:  []byte("[]"),
		RawMessage:                []byte("{}"),
	}

	buildingID, buildingOptions, err := NewIsmartExternalService(runtimeValue).selectVisibleBuilding(context.Background(), account, "")
	if err != nil {
		t.Fatalf("select building: %v", err)
	}
	if buildingID != "BLG-002" {
		t.Fatalf("expected profile building BLG-002, got %s", buildingID)
	}
	if len(buildingOptions) != 2 || buildingOptions[0] != "BLG-001" || buildingOptions[1] != "BLG-002" {
		t.Fatalf("expected original visible building options, got %#v", buildingOptions)
	}
}

// 2. TestIsmartSelectVisibleBuildingRejectsInvisibleProfileBuilding verifies mismatched bindings fail closed.
func TestIsmartSelectVisibleBuildingRejectsInvisibleProfileBuilding(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
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
	boundBuilding := model.Community{
		PublicID:      "BLG-002",
		CommunityType: "building",
		NameZH:        "仁英大廈",
		DistrictCode:  "unknown",
	}
	if err := runtimeValue.DB.Create(&boundBuilding).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	profileBuildingJSON, err := marshalJSON([]string{"BLG-002"})
	if err != nil {
		t.Fatalf("marshal profile building json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &boundBuilding.ID,
		BoundBuildingIDs:   profileBuildingJSON,
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	visibleBuildingJSON, err := marshalJSON([]string{"BLG-001"})
	if err != nil {
		t.Fatalf("marshal visible building json: %v", err)
	}
	account := &model.UserIsmartAccount{
		UserID:                    user.ID,
		IsmartUserID:              89,
		Username:                  "patrick",
		ClientBuildingPermissions: visibleBuildingJSON,
		Building:                  []byte("[]"),
		StaffBuildingPermissions:  []byte("[]"),
		RawMessage:                []byte("{}"),
	}

	_, _, err = NewIsmartExternalService(runtimeValue).selectVisibleBuilding(context.Background(), account, "")
	if err == nil {
		t.Fatal("expected invisible profile building to be rejected")
	}
}
