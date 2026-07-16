/*
 * 會員資料服務測試。
 * 1. 驗證切換已授權大廈後回應使用最新綁定。
 * 2. 驗證既有 ID 佔位社區會更新為正式大廈名稱。
 */
package service

import (
	"context"
	"testing"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestUserServiceUpdatesBuildingBindingAndPlaceholderName verifies profile binding response parity.
func TestUserServiceUpdatesBuildingBindingAndPlaceholderName(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234572",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	oldCommunity := model.Community{
		PublicID:      "0419900",
		CommunityType: "building",
		NameZH:        "舊大廈",
		NameEN:        "Old Building",
		DistrictCode:  "unknown",
		AddressText:   "舊大廈",
	}
	if err := runtimeValue.DB.Create(&oldCommunity).Error; err != nil {
		t.Fatalf("create old community: %v", err)
	}
	placeholderCommunity := model.Community{
		PublicID:      "0999900",
		CommunityType: "building",
		NameZH:        "0999900",
		NameEN:        "0999900",
		DistrictCode:  "unknown",
		AddressText:   "0999900",
	}
	if err := runtimeValue.DB.Create(&placeholderCommunity).Error; err != nil {
		t.Fatalf("create placeholder community: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:                 user.ID,
		DisplayName:            "Building Member",
		PrimaryCommunityID:     &oldCommunity.ID,
		ResidenceBindingStatus: residenceBindingStatusPending,
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	result, err := NewUserService(runtimeValue).UpdateProfile(context.Background(), user.ID, UpdateProfileParams{
		DisplayName:          "Building Member",
		PrimaryCommunityID:   "0999900",
		PrimaryCommunityName: "測試1大廈",
		BoundBuildingIDs:     []string{"0999900"},
		BoundFlatUnitIDs:     []string{"09999000012"},
		ResidenceFloor:       "G",
		ResidenceUnit:        "B",
	})
	if err != nil {
		t.Fatalf("update profile binding: %v", err)
	}
	if result.PrimaryCommunity == nil || result.PrimaryCommunity.PublicID != "0999900" || result.PrimaryCommunity.NameZH != "測試1大廈" {
		t.Fatalf("expected updated community response, got %#v", result.PrimaryCommunity)
	}
	if len(result.BoundBuildingIDs) != 1 || result.BoundBuildingIDs[0] != "0999900" {
		t.Fatalf("expected latest building binding, got %v", result.BoundBuildingIDs)
	}
	if len(result.BoundFlatUnitIDs) != 1 || result.BoundFlatUnitIDs[0] != "09999000012" {
		t.Fatalf("expected latest unit binding, got %v", result.BoundFlatUnitIDs)
	}

	var community model.Community
	if err := runtimeValue.DB.Where("public_id = ?", "0999900").First(&community).Error; err != nil {
		t.Fatalf("load updated community: %v", err)
	}
	if community.NameZH != "測試1大廈" || community.NameEN != "測試1大廈" || community.AddressText != "測試1大廈" {
		t.Fatalf("expected placeholder name refresh, got %#v", community)
	}
}
