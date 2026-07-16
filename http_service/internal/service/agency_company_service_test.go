/*
 * Agency profile service tests.
 * 1. Verify individual and company review activation.
 * 2. Verify approved revisions remain active while a replacement is reviewed.
 */
package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestAgencyProfileReviewWorkflow verifies company approval and active revision preservation.
func TestAgencyProfileReviewWorkflow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	cfg := &config.Config{MediaBaseURL: "https://cdn.test", StorageEndpoint: "https://storage.test", EncryptionKey: "agency-profile-test-key"}
	runtime := &Runtime{Config: cfg, DB: db, Now: func() time.Time { return now }, StorageProvider: &MockStorageProvider{config: cfg}}
	presignedEvidence, err := runtime.StorageProvider.PresignUpload(context.Background(), PresignUploadInput{FileName: "eaa.jpg", MimeType: "image/jpeg", FileSize: 10, ObjectPrefix: agencyCompanyEAAObjectPrefix})
	if err != nil || presignedEvidence.Headers["x-oss-object-acl"] != "private" {
		t.Fatalf("expected private evidence upload: %#v %v", presignedEvidence, err)
	}
	service := NewAgencyCompanyService(runtime)
	user := createAgencyProfileTestUser(t, db, AccountTypeAgencyCompany, "pending_profile", "61234567")
	eaa := createAgencyProfileAsset(t, db, user.ID, agencyCompanyEAAObjectPrefix+"eaa.jpg")
	registration := createAgencyProfileAsset(t, db, user.ID, agencyCompanyBusinessRegistrationObjectPrefix+"br.jpg")
	params := AgencyProfileUpsertParams{ProfileType: AgencyProfileTypeCompany, NameZH: "測試地產", NameEN: "Test Agency", AddressZH: "香港中環", AddressEN: "Central Hong Kong", LicenseNumber: "C-123456", Phone1CountryCode: "+852", Phone1Number: "21234567", EAALicenseAssetID: eaa.PublicID, BusinessRegistrationAssetID: registration.PublicID}
	draft, err := service.CreateMemberAgencyProfile(context.Background(), user.ID, params)
	if err != nil || draft.Revision == nil {
		t.Fatalf("create draft: %#v %v", draft, err)
	}
	pending, err := service.SubmitMemberAgencyProfile(context.Background(), user.ID)
	if err != nil || pending.Revision.Status != AgencyProfileStatusPending {
		t.Fatalf("submit: %#v %v", pending, err)
	}
	approved, err := service.ReviewAgencyProfile(context.Background(), pending.Revision.ProfileID, AgencyProfileReviewParams{ReviewerUserID: user.ID, Approved: true})
	if err != nil || approved.Status != AgencyProfileStatusApproved {
		t.Fatalf("approve: %#v %v", approved, err)
	}
	if approved.EAALicenseAsset == nil || !strings.Contains(approved.EAALicenseAsset.URL, "/download/") || strings.Contains(approved.EAALicenseAsset.URL, "https://cdn.test/") {
		t.Fatalf("expected private signed EAA URL, got %#v", approved.EAALicenseAsset)
	}
	var saved model.User
	db.First(&saved, user.ID)
	if saved.MemberStatus != "active" {
		t.Fatalf("expected active, got %s", saved.MemberStatus)
	}
	child, err := service.CreateSubaccount(context.Background(), user.ID, AgencySubaccountCreateParams{DisplayName: "公司職員", PhoneCountryCode: "+852", PhoneNumber: "63334444", Password: "password123", Permissions: []string{"property_publish"}})
	if err != nil || child.Status != "active" {
		t.Fatalf("create subaccount: %#v %v", child, err)
	}
	var childLink model.AgencyCompanySubaccount
	if err := db.Where("public_id = ?", child.PublicID).First(&childLink).Error; err != nil {
		t.Fatal(err)
	}
	propertyService := NewPropertyService(runtime)
	if err := propertyService.requireSubaccountPropertyPermission(context.Background(), db, childLink.ChildUserID, "property_publish"); err != nil {
		t.Fatalf("publish permission should be allowed: %v", err)
	}
	if err := propertyService.requireSubaccountPropertyPermission(context.Background(), db, childLink.ChildUserID, "property_manage"); err == nil {
		t.Fatal("manage permission should be denied")
	}
	listing, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: childLink.ChildUserID, PublisherIdentityType: "owner", AgencyCompanyName: "untrusted", LocationScope: "local"})
	if err != nil {
		t.Fatalf("create child company listing: %v", err)
	}
	if listing.PublisherIdentityType != "agent" || listing.AgentSnapshot == nil || listing.AgentSnapshot.LicenseNumber != params.LicenseNumber {
		t.Fatalf("unexpected authoritative publisher snapshot: %#v", listing)
	}
	if _, err := NewPropertyService(runtime).UpdatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: childLink.ChildUserID, ListingPublicID: listing.ListingID, LocationScope: "local"}); err == nil {
		t.Fatal("expected publish-only child update to fail")
	}
	if _, err := NewPropertyService(runtime).UpdatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: user.ID, ListingPublicID: listing.ListingID, LocationScope: "local"}); err != nil {
		t.Fatalf("company owner update child listing: %v", err)
	}
	managed, _, err := NewPropertyService(runtime).ListMyProperties(context.Background(), PropertyChannelSale, user.ID, PropertyListFilters{Page: 1, PageSize: 10})
	if err != nil || len(managed) != 1 {
		t.Fatalf("company owner child listings: %#v %v", managed, err)
	}
	manager, err := service.CreateSubaccount(context.Background(), user.ID, AgencySubaccountCreateParams{DisplayName: "管理職員", PhoneCountryCode: "+852", PhoneNumber: "64445555", Password: "password123", Permissions: []string{"property_publish", "property_manage"}})
	if err != nil {
		t.Fatalf("create manager subaccount: %v", err)
	}
	var managerLink model.AgencyCompanySubaccount
	if err := db.Where("public_id = ?", manager.PublicID).First(&managerLink).Error; err != nil {
		t.Fatal(err)
	}
	managerListing, err := NewPropertyService(runtime).CreatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: managerLink.ChildUserID, LocationScope: "local"})
	if err != nil {
		t.Fatalf("create manager listing: %v", err)
	}
	if _, err := NewPropertyService(runtime).UpdatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: managerLink.ChildUserID, ListingPublicID: managerListing.ListingID, LocationScope: "local"}); err != nil {
		t.Fatalf("manager update listing: %v", err)
	}
	if err := NewPropertyService(runtime).DeactivateProperty(context.Background(), PropertyChannelSale, managerLink.ChildUserID, managerListing.ListingID); err != nil {
		t.Fatalf("manager deactivate listing: %v", err)
	}
	if _, err := service.UpdateSubaccountStatus(context.Background(), user.ID, child.PublicID, "disabled"); err != nil {
		t.Fatalf("disable subaccount: %v", err)
	}
	if err := service.ValidateSubaccountAccess(context.Background(), childLink.ChildUserID); err == nil {
		t.Fatal("expected disabled child access to fail")
	}
	var sourceListing model.Listing
	if err := db.Where("public_id = ?", listing.ListingID).First(&sourceListing).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.Listing{}).Where("id = ?", sourceListing.ID).Update("publication_status", "active").Error; err != nil {
		t.Fatal(err)
	}
	sourceListing.PublicationStatus = "active"
	var sourceContact model.ListingContact
	if err := db.Where("listing_id = ?", sourceListing.ID).First(&sourceContact).Error; err != nil {
		t.Fatal(err)
	}
	sourceAttrs := decodeStringMapBytes(sourceContact.ContactAttributes)
	sourceAttrs["business_source"] = "member_center"
	encodedAttrs, _ := marshalJSON(sourceAttrs)
	if err := db.Model(&model.ListingContact{}).Where("listing_id = ?", sourceListing.ID).Update("contact_attributes", encodedAttrs).Error; err != nil {
		t.Fatal(err)
	}
	now = now.Add(agencyProfileEditCooldown)
	revisedParams := params
	revisedParams.NameZH = "更新地產"
	revisedParams.NameEN = "Updated Agency"
	revisedParams.Phone1Number = "29876543"
	revisedParams.DefaultAvatar = "male"
	revision, err := service.CreateMemberAgencyProfile(context.Background(), user.ID, revisedParams)
	if err != nil || revision.ActiveProfile == nil || revision.Revision == nil {
		t.Fatalf("revision: %#v %v", revision, err)
	}
	if _, err := service.SubmitMemberAgencyProfile(context.Background(), user.ID); err != nil {
		t.Fatal(err)
	}
	current, _ := service.GetMemberAgencyProfile(context.Background(), user.ID)
	if current.ActiveProfile == nil || current.ActiveProfile.ProfileID != approved.ProfileID || current.Revision.Status != AgencyProfileStatusPending {
		t.Fatalf("active profile not preserved: %#v", current)
	}
	listingRow := sourceListing
	var pendingContact model.ListingContact
	if err := db.Where("listing_id = ?", listingRow.ID).First(&pendingContact).Error; err != nil {
		t.Fatal(err)
	}
	pendingAttrs := decodeStringMapBytes(pendingContact.ContactAttributes)
	if pendingAttrs["agency_name_zh"] != params.NameZH {
		t.Fatalf("pending revision changed listing snapshot: %#v", pendingAttrs)
	}
	newActive, err := service.ReviewAgencyProfile(context.Background(), current.Revision.ProfileID, AgencyProfileReviewParams{ReviewerUserID: user.ID, Approved: true})
	if err != nil {
		t.Fatalf("approve replacement profile: %v", err)
	}
	staffProfiles, _, err := service.ListAgencyProfilesForStaff(context.Background(), AgencyProfileListFilters{Page: 1, PageSize: 10, Status: AgencyProfileStatusApproved})
	if err != nil || len(staffProfiles) != 1 || staffProfiles[0].ProfileID != newActive.ProfileID {
		t.Fatalf("expected only current approved profile: %#v %v", staffProfiles, err)
	}
	var approvedContact model.ListingContact
	if err := db.Where("listing_id = ?", listingRow.ID).First(&approvedContact).Error; err != nil {
		t.Fatal(err)
	}
	approvedAttrs := decodeStringMapBytes(approvedContact.ContactAttributes)
	phone, err := utils.DecryptString(runtime.Config.EncryptionKey, approvedContact.PhoneEncrypted)
	if err != nil || approvedAttrs["agency_name_zh"] != revisedParams.NameZH || phone != revisedParams.Phone1Number || approvedAttrs["business_source"] != "member_center" {
		t.Fatalf("approved revision did not refresh child listing: attrs=%#v phone=%q err=%v", approvedAttrs, phone, err)
	}
	if newActive.DefaultAvatar != "" {
		t.Fatalf("company default avatar must be empty, got %q", newActive.DefaultAvatar)
	}
}

// 2. TestIndividualOverseasValidation verifies overseas agents may submit without a Hong Kong licence.
func TestIndividualOverseasValidation(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = database.Migrate(db)
	runtime := &Runtime{Config: &config.Config{EncryptionKey: "individual-test-key"}, DB: db, Now: time.Now}
	user := createAgencyProfileTestUser(t, db, AccountTypeIndividualAgent, "pending_profile", "62345678")
	params := AgencyProfileUpsertParams{ProfileType: AgencyProfileTypeIndividual, NameZH: "測試代理", NameEN: "Test Agent", Phone1CountryCode: "+852", Phone1Number: "62345678", Phone1WhatsApp: true, DefaultAvatar: "male", IsOverseas: true}
	service := NewAgencyCompanyService(runtime)
	if _, err := service.CreateMemberAgencyProfile(context.Background(), user.ID, params); err != nil {
		t.Fatal(err)
	}
	pending, err := service.SubmitMemberAgencyProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("overseas submit: %v", err)
	}
	if _, err := service.ReviewAgencyProfile(context.Background(), pending.Revision.ProfileID, AgencyProfileReviewParams{ReviewerUserID: user.ID, Approved: true}); err != nil {
		t.Fatal(err)
	}
	_, err = NewPropertyService(runtime).CreatePropertySale(context.Background(), UpsertPropertySaleParams{OwnerUserID: user.ID, LocationScope: "local"})
	if err == nil {
		t.Fatal("expected unlicensed overseas local listing to fail")
	}
}

// 3. TestAgencyProfileRejectsUnsafeWechatURL blocks executable public links.
func TestAgencyProfileRejectsUnsafeWechatURL(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = database.Migrate(db)
	runtime := &Runtime{Config: &config.Config{EncryptionKey: "wechat-url-test-key"}, DB: db, Now: time.Now}
	user := createAgencyProfileTestUser(t, db, AccountTypeIndividualAgent, "pending_profile", "63456789")
	service := NewAgencyCompanyService(runtime)
	params := AgencyProfileUpsertParams{ProfileType: AgencyProfileTypeIndividual, NameZH: "測試代理", NameEN: "Test Agent", Phone1CountryCode: "+852", Phone1Number: "63456789", DefaultAvatar: "male", IsOverseas: true, WechatURL: "javascript:alert(1)"}
	if _, err := service.CreateMemberAgencyProfile(context.Background(), user.ID, params); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitMemberAgencyProfile(context.Background(), user.ID); err == nil {
		t.Fatal("expected unsafe WeChat URL to fail")
	}
}

// 4. createAgencyProfileTestUser inserts one agency account.
func createAgencyProfileTestUser(t *testing.T, db *gorm.DB, accountType string, status string, phone string) model.User {
	t.Helper()
	user := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: "+852", PhoneNumber: phone, MemberStatus: status, MemberType: "user"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserProfile{UserID: user.ID, DisplayName: "測試", AccountType: accountType, PublisherIdentityType: derivedPublisherIdentity(accountType)}).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

// 5. createAgencyProfileAsset inserts one owned image under a semantic OSS root.
func createAgencyProfileAsset(t *testing.T, db *gorm.DB, userID int64, key string) model.MediaAsset {
	t.Helper()
	asset := model.MediaAsset{PublicID: utils.NewPublicID(), StorageProvider: "oss", BucketName: "test", ObjectKey: key, MimeType: "image/jpeg", FileSize: 10, CreatedBy: &userID}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	return asset
}
