/*
 * Secondhand access and helper logic.
 * 1. Build listing detail payloads and contact access responses.
 * 2. Resolve media, contacts, communities, and visibility checks.
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. secondhandListingRow defines the common secondhand list query row.
type secondhandListingRow struct {
	model.Listing
	CategoryCode    string
	PriceMode       string
	PriceHKD        *float64
	ConditionLevel  string
	VisibilityScope string
	ContactMethod   string
	IsFreeGiveaway  bool
}

// 2. secondhandUserPreviewRow defines batched user preview query fields.
type secondhandUserPreviewRow struct {
	ID                    int64
	PublicID              string
	DisplayName           string
	PublisherIdentityType string
	AvatarObjectKey       string
}

// 3. GrantContactAccess validates listing visibility and returns allowed contact payload.
func (s *SecondhandService) GrantContactAccess(ctx context.Context, userID int64, communityID *int64, listingPublicID string, requestIP string, userAgent string) (*ContactAccessResult, error) {
	listing, secondhand, contact, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}

	if listing.PublicationStatus != "active" {
		return nil, errcode.New(errcode.CodeExpired, "listing is not active")
	}
	if listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available")
	}
	if !s.canViewListing(listing, secondhand, communityID) {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}

	payload := map[string]string{}
	channels := map[string]bool{
		"phone":        false,
		"whatsapp":     false,
		"chat":         false,
		"inquiry_form": false,
	}

	if contact.ShowPhone && contact.PhoneEncrypted != "" {
		phone, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.PhoneEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt phone")
		}
		payload["phone"] = phone
		channels["phone"] = true
	}

	if contact.ShowWhatsApp && contact.WhatsAppEncrypted != "" {
		whatsApp, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.WhatsAppEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt whatsapp")
		}
		payload["whatsapp_url"] = buildWhatsAppURL(whatsApp, listing.Title, listing.PublicID, s.runtime.Config.AppPublicBaseURL)
		channels["whatsapp"] = true
	}

	if contact.ShowChat {
		channels["chat"] = true
	}

	grantedChannels, err := marshalJSON(channels)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to encode access audit")
	}

	logRecord := &model.ContactAccessLog{
		ListingID:       listing.ID,
		RequestUserID:   userID,
		GrantedChannels: grantedChannels,
		RequestIP:       requestIP,
		UserAgent:       userAgent,
	}
	if err := s.runtime.DB.WithContext(ctx).Create(logRecord).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to store contact access log")
	}

	return &ContactAccessResult{
		ListingID:       listing.PublicID,
		AllowedChannels: channels,
		ContactPayload:  payload,
	}, nil
}

// 4. loadListingByPublicID loads listing aggregate by public ID.
func (s *SecondhandService) loadListingByPublicID(ctx context.Context, listingPublicID string) (*model.Listing, *model.SecondhandListing, *model.ListingContact, error) {
	var listing model.Listing
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ? AND module = ? AND is_deleted = ?", listingPublicID, "secondhand", false).First(&listing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing")
	}

	var secondhand model.SecondhandListing
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&secondhand).Error; err != nil {
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing extension")
	}

	var contact model.ListingContact
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&contact).Error; err != nil {
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing contact")
	}

	return &listing, &secondhand, &contact, nil
}

// 5. loadOwnedListing loads a listing aggregate owned by the current user.
func (s *SecondhandService) loadOwnedListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*model.Listing, *model.SecondhandListing, *model.ListingContact, error) {
	return s.loadOwnedListingWithTx(ctx, s.runtime.DB, ownerUserID, listingPublicID)
}

// 6. loadOwnedListingWithTx loads a listing aggregate owned by the current user inside a transaction.
func (s *SecondhandService) loadOwnedListingWithTx(ctx context.Context, tx *gorm.DB, ownerUserID int64, listingPublicID string) (*model.Listing, *model.SecondhandListing, *model.ListingContact, error) {
	listing, secondhand, contact, err := s.loadListingByPublicIDWithDB(ctx, tx, listingPublicID)
	if err != nil {
		return nil, nil, nil, err
	}
	if listing.OwnerUserID != ownerUserID {
		return nil, nil, nil, errcode.New(errcode.CodeAuthForbidden, "listing does not belong to the current user")
	}

	return listing, secondhand, contact, nil
}

// 7. loadListingByPublicIDWithDB loads listing aggregate with the provided DB handle.
func (s *SecondhandService) loadListingByPublicIDWithDB(ctx context.Context, db *gorm.DB, listingPublicID string) (*model.Listing, *model.SecondhandListing, *model.ListingContact, error) {
	var listing model.Listing
	if err := db.WithContext(ctx).Where("public_id = ? AND module = ? AND is_deleted = ?", listingPublicID, "secondhand", false).First(&listing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing")
	}

	var secondhand model.SecondhandListing
	if err := db.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&secondhand).Error; err != nil {
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing extension")
	}

	var contact model.ListingContact
	if err := db.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&contact).Error; err != nil {
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing contact")
	}

	return &listing, &secondhand, &contact, nil
}

// 8. buildListingDetail builds a detail response from listing aggregate models.
func (s *SecondhandService) buildListingDetail(ctx context.Context, listing *model.Listing, secondhand *model.SecondhandListing, contact *model.ListingContact) (*SecondhandListingDetail, error) {
	summaries, err := s.buildListingSummaries(ctx, []secondhandListingRow{
		{
			Listing:         *listing,
			CategoryCode:    secondhand.CategoryCode,
			PriceMode:       secondhand.PriceMode,
			PriceHKD:        secondhand.PriceHKD,
			ConditionLevel:  secondhand.ConditionLevel,
			VisibilityScope: secondhand.VisibilityScope,
			ContactMethod:   secondhand.ContactMethod,
			IsFreeGiveaway:  secondhand.IsFreeGiveaway,
		},
	})
	if err != nil {
		return nil, err
	}

	detail := &SecondhandListingDetail{
		SecondhandListingSummary: summaries[0],
		Description:              listing.Description,
		DimensionText:            secondhand.DimensionText,
		PickupRegionCode:         secondhand.PickupRegionCode,
		PickupLocationText:       secondhand.PickupLocationText,
		DeliveryTags:             unmarshalStringSlice(secondhand.DeliveryTags),
		ContactSummary: ListingContactSummary{
			ShowPhone:    contact.ShowPhone,
			ShowWhatsApp: contact.ShowWhatsApp,
			ShowChat:     contact.ShowChat,
			ShowInquiry:  contact.ShowInquiryForm,
		},
	}

	images, err := s.loadListingImages(ctx, []int64{listing.ID})
	if err != nil {
		return nil, err
	}
	detail.Images = images[listing.ID]
	return detail, nil
}

// 9. buildListingSummaries maps list query rows into response payloads.
func (s *SecondhandService) buildListingSummaries(ctx context.Context, rows []secondhandListingRow) ([]SecondhandListingSummary, error) {
	listingIDs := make([]int64, 0, len(rows))
	ownerUserIDs := make([]int64, 0, len(rows))
	communityIDs := make([]int64, 0, len(rows))
	for _, item := range rows {
		listingIDs = append(listingIDs, item.ID)
		ownerUserIDs = append(ownerUserIDs, item.OwnerUserID)
		if item.CommunityID != nil {
			communityIDs = append(communityIDs, *item.CommunityID)
		}
	}

	imageMap, err := s.loadListingImages(ctx, listingIDs)
	if err != nil {
		return nil, err
	}
	ownerMap, err := s.loadUserPreviewMap(ctx, ownerUserIDs)
	if err != nil {
		return nil, err
	}
	communityMap, err := s.loadCommunityMap(ctx, communityIDs)
	if err != nil {
		return nil, err
	}

	result := make([]SecondhandListingSummary, 0, len(rows))
	for _, item := range rows {
		expireAt := ""
		expireAtPtr := &expireAt
		if item.ExpireAt != nil {
			expireAt = item.ExpireAt.UTC().Format(time.RFC3339)
		} else {
			expireAtPtr = nil
		}
		var publishedAtPtr *string
		if item.PublishedAt != nil {
			publishedAt := item.PublishedAt.UTC().Format(time.RFC3339)
			publishedAtPtr = &publishedAt
		}

		var cover *ListingImageResponse
		if images := imageMap[item.ID]; len(images) > 0 {
			cover = &images[0]
		}
		owner := ownerMap[item.OwnerUserID]
		if owner != nil && owner.PublisherIdentityType == "" {
			owner.PublisherIdentityType = item.PublisherIdentityType
		}
		var community *CommunityResponse
		if item.CommunityID != nil {
			community = communityMap[*item.CommunityID]
		}

		result = append(result, SecondhandListingSummary{
			ListingID:             item.PublicID,
			Title:                 item.Title,
			Summary:               item.Summary,
			DistrictCode:          item.DistrictCode,
			PublishedAt:           publishedAtPtr,
			CategoryCode:          item.CategoryCode,
			PriceMode:             item.PriceMode,
			PriceHKD:              item.PriceHKD,
			ConditionLevel:        item.ConditionLevel,
			VisibilityScope:       item.VisibilityScope,
			ContactMethod:         item.ContactMethod,
			IsFreeGiveaway:        item.IsFreeGiveaway,
			PublisherIdentityType: item.PublisherIdentityType,
			PublicationStatus:     item.PublicationStatus,
			BusinessStatus:        item.BusinessStatus,
			ExpireAt:              expireAtPtr,
			UpdatedAt:             item.UpdatedAt.UTC().Format(time.RFC3339),
			Community:             community,
			Owner:                 owner,
			CoverImage:            cover,
		})
	}

	return result, nil
}

// 10. loadListingImages loads listing images and media assets for listing IDs.
func (s *SecondhandService) loadListingImages(ctx context.Context, listingIDs []int64) (map[int64][]ListingImageResponse, error) {
	result := make(map[int64][]ListingImageResponse)
	if len(listingIDs) == 0 {
		return result, nil
	}

	var images []model.ListingImage
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id IN ?", listingIDs).Order("sort_order asc").Find(&images).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing images")
	}

	assetIDs := make([]int64, 0, len(images))
	for _, image := range images {
		assetIDs = append(assetIDs, image.MediaAssetID)
	}

	var assets []model.MediaAsset
	if len(assetIDs) > 0 {
		if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", assetIDs).Find(&assets).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
		}
	}

	assetMap := make(map[int64]model.MediaAsset, len(assets))
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}

	for _, image := range images {
		asset := assetMap[image.MediaAssetID]
		result[image.ListingID] = append(result[image.ListingID], ListingImageResponse{
			MediaAssetID: asset.PublicID,
			URL:          s.mediaURL(&asset),
			SortOrder:    image.SortOrder,
			IsCover:      image.IsCover,
		})
	}

	return result, nil
}

// 11. loadUserPreviewMap loads lightweight user display data for response payloads.
func (s *SecondhandService) loadUserPreviewMap(ctx context.Context, userIDs []int64) (map[int64]*UserPreviewResponse, error) {
	result := make(map[int64]*UserPreviewResponse)
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []secondhandUserPreviewRow
	if err := s.runtime.DB.WithContext(ctx).Table("users").
		Select("users.id, users.public_id, user_profiles.display_name, user_profiles.publisher_identity_type, media_assets.object_key AS avatar_object_key").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.id").
		Joins("LEFT JOIN media_assets ON media_assets.id = user_profiles.avatar_asset_id").
		Where("users.id IN ?", userIDs).
		Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user previews")
	}

	for _, item := range rows {
		displayName := strings.TrimSpace(item.DisplayName)
		if displayName == "" {
			displayName = item.PublicID
		}
		result[item.ID] = &UserPreviewResponse{
			UserID:                fmt.Sprintf("%d", item.ID),
			PublicID:              item.PublicID,
			DisplayName:           displayName,
			AvatarURL:             buildMediaURL(s.runtime.Config.MediaBaseURL, item.AvatarObjectKey),
			PublisherIdentityType: item.PublisherIdentityType,
		}
	}

	return result, nil
}

// 12. loadCommunityMap loads lightweight communities by ID.
func (s *SecondhandService) loadCommunityMap(ctx context.Context, communityIDs []int64) (map[int64]*CommunityResponse, error) {
	result := make(map[int64]*CommunityResponse)
	if len(communityIDs) == 0 {
		return result, nil
	}

	var communities []model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", communityIDs).Find(&communities).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing communities")
	}

	for i := range communities {
		community := communities[i]
		result[community.ID] = toCommunityResponse(&community)
	}

	return result, nil
}

// 13. buildListingContact builds the encrypted listing contact record.
func (s *SecondhandService) buildListingContact(listingID int64, input ListingContactInput, contactMethod string) (*model.ListingContact, error) {
	contact := &model.ListingContact{
		ListingID:       listingID,
		ShowPhone:       input.ShowPhone,
		ShowWhatsApp:    input.ShowWhatsApp,
		ShowChat:        input.ShowChat,
		ShowInquiryForm: input.ShowInquiryForm,
		ContactMode:     contactMethod,
	}

	if input.Phone != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, input.Phone)
		if err != nil {
			return nil, err
		}
		contact.PhoneEncrypted = encrypted
		contact.PhoneMasked = utils.MaskPhone(input.Phone)
	}

	if input.WhatsApp != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, input.WhatsApp)
		if err != nil {
			return nil, err
		}
		contact.WhatsAppEncrypted = encrypted
		contact.WhatsAppMasked = utils.MaskPhone(input.WhatsApp)
	}

	return contact, nil
}

// 14. resolveCommunityID resolves explicit or profile-based community selection.
func (s *SecondhandService) resolveCommunityID(ctx context.Context, ownerUserID int64, communityPublicID string, visibilityScope string) (*int64, error) {
	if strings.TrimSpace(communityPublicID) != "" {
		community, err := s.findCommunityByPublicID(ctx, communityPublicID)
		if err != nil {
			return nil, err
		}
		return &community.ID, nil
	}

	var profile model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", ownerUserID).First(&profile).Error; err == nil && profile.PrimaryCommunityID != nil {
		return profile.PrimaryCommunityID, nil
	}

	if visibilityScope == "building_only" {
		return nil, errcode.New(errcode.CodeValidationError, "community is required for building_only visibility")
	}

	return nil, nil
}

// 15. findCommunityByPublicID loads a community by public ID.
func (s *SecondhandService) findCommunityByPublicID(ctx context.Context, publicID string) (*model.Community, error) {
	var community model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", publicID).First(&community).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "community not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load community")
	}

	return &community, nil
}

// 16. resolveListingImages validates referenced media assets and builds listing image records.
func (s *SecondhandService) resolveListingImages(ctx context.Context, tx *gorm.DB, listingID int64, ownerUserID int64, inputs []ListingImageInput) ([]model.ListingImage, error) {
	if len(inputs) == 0 {
		return []model.ListingImage{}, nil
	}

	assetPublicIDs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		assetPublicIDs = append(assetPublicIDs, input.MediaAssetID)
	}

	var assets []model.MediaAsset
	if err := tx.WithContext(ctx).Where("public_id IN ? AND created_by = ?", assetPublicIDs, ownerUserID).Find(&assets).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
	}

	assetMap := make(map[string]model.MediaAsset, len(assets))
	for _, asset := range assets {
		assetMap[asset.PublicID] = asset
	}

	images := make([]model.ListingImage, 0, len(inputs))
	for index, input := range inputs {
		asset, exists := assetMap[input.MediaAssetID]
		if !exists {
			return nil, errcode.New(errcode.CodeValidationError, "media asset not found")
		}

		sortOrder := input.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}
		images = append(images, model.ListingImage{
			ListingID:    listingID,
			MediaAssetID: asset.ID,
			SortOrder:    sortOrder,
			IsCover:      input.IsCover || index == 0,
		})
	}

	return images, nil
}

// 17. validateListingReady ensures draft listing has publishable content.
func (s *SecondhandService) validateListingReady(ctx context.Context, ownerUserID int64, listingID int64) error {
	var imageCount int64
	if err := s.runtime.DB.WithContext(ctx).Model(&model.ListingImage{}).Where("listing_id = ?", listingID).Count(&imageCount).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate listing images")
	}
	if imageCount == 0 {
		return errcode.New(errcode.CodeValidationError, "at least one image is required before publish")
	}

	var contact model.ListingContact
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id = ?", listingID).First(&contact).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate listing contacts")
	}
	if !contact.ShowPhone && !contact.ShowWhatsApp && !contact.ShowChat {
		return errcode.New(errcode.CodeValidationError, "at least one contact channel must be enabled")
	}

	return nil
}

// 18. canViewListing checks secondhand visibility rules for the current viewer.
func (s *SecondhandService) canViewListing(listing *model.Listing, secondhand *model.SecondhandListing, viewerCommunityID *int64) bool {
	if secondhand.VisibilityScope == "public" {
		return true
	}
	return viewerCommunityID != nil && secondhand.VisibleCommunityID != nil && *viewerCommunityID == *secondhand.VisibleCommunityID
}

// 19. findCommunityByID loads a community by numeric ID.
func (s *SecondhandService) findCommunityByID(ctx context.Context, communityID *int64) (*model.Community, error) {
	if communityID == nil {
		return nil, nil
	}

	var community model.Community
	if err := s.runtime.DB.WithContext(ctx).First(&community, *communityID).Error; err != nil {
		return nil, err
	}

	return &community, nil
}

// 20. buildWhatsAppURL creates a prefilled WhatsApp deep link.
func buildWhatsAppURL(phone string, title string, listingPublicID string, baseURL string) string {
	digits := strings.NewReplacer("+", "", " ", "", "-", "", "(", "", ")", "").Replace(phone)
	message := fmt.Sprintf("I am interested in %s %s/secondhand/%s", title, strings.TrimRight(baseURL, "/"), listingPublicID)
	return "https://wa.me/" + digits + "?text=" + url.QueryEscape(message)
}

// 21. normalizePrice normalizes nullable price for free listings.
func normalizePrice(priceMode string, price *float64) *float64 {
	if priceMode == "free" {
		return nil
	}
	return price
}

// 22. visibleCommunityID returns the visible community when required.
func visibleCommunityID(visibilityScope string, communityID *int64) *int64 {
	if visibilityScope != "building_only" {
		return nil
	}
	return communityID
}
