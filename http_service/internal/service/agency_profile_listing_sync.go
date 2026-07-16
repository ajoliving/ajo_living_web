/*
 * Approved agency profile listing synchronization.
 * 1. Replace property-sale contact snapshots only after profile approval.
 * 2. Include every linked company child while preserving unrelated contact metadata.
 */
package service

import (
	"context"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. syncApprovedAgencyProfileListings refreshes all non-deleted sale listing snapshots.
func (s *AgencyProfileService) syncApprovedAgencyProfileListings(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile) error {
	ownerIDs, err := s.agencyProfileListingOwnerIDs(ctx, tx, profile)
	if err != nil {
		return err
	}
	var listings []model.Listing
	if err := tx.WithContext(ctx).Where("module = ? AND owner_user_id IN ? AND is_deleted = ?", string(PropertyChannelSale), ownerIDs, false).Find(&listings).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load agency property listings")
	}
	if len(listings) == 0 {
		return nil
	}
	listingIDs := make([]int64, 0, len(listings))
	for _, listing := range listings {
		listingIDs = append(listingIDs, listing.ID)
	}
	var contacts []model.ListingContact
	if err := tx.WithContext(ctx).Where("listing_id IN ?", listingIDs).Find(&contacts).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load agency listing contacts")
	}
	contactMap := make(map[int64]model.ListingContact, len(contacts))
	for _, contact := range contacts {
		contactMap[contact.ListingID] = contact
	}
	propertyService := NewPropertyService(s.runtime)
	publisher := propertyService.toPropertySalePublisher(ctx, tx, profile)
	for _, listing := range listings {
		if err := propertyService.replaceAgencyListingSnapshot(ctx, tx, &listing, contactMap[listing.ID], publisher); err != nil {
			return err
		}
	}
	return nil
}

// 2. agencyProfileListingOwnerIDs returns the profile owner and linked company children.
func (s *AgencyProfileService) agencyProfileListingOwnerIDs(ctx context.Context, tx *gorm.DB, profile *model.AgencyProfile) ([]int64, error) {
	result := []int64{profile.UserID}
	if profile.ProfileType != AgencyProfileTypeCompany {
		return result, nil
	}
	var links []model.AgencyCompanySubaccount
	if err := tx.WithContext(ctx).Where("company_owner_user_id = ?", profile.UserID).Find(&links).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load agency company listing members")
	}
	for _, link := range links {
		result = append(result, link.ChildUserID)
	}
	return result, nil
}

// 3. replaceAgencyListingSnapshot writes one approved public-safe contact snapshot.
func (s *PropertyService) replaceAgencyListingSnapshot(ctx context.Context, tx *gorm.DB, listing *model.Listing, current model.ListingContact, publisher propertySalePublisher) error {
	params := UpsertPropertySaleParams{Contact: PropertyContactInput{
		ContactAttributes: decodeStringMapBytes(current.ContactAttributes),
		ShowPhone:         current.ShowPhone, ShowWhatsApp: current.ShowWhatsApp,
		ShowChat: current.ShowChat, ShowInquiryForm: current.ShowInquiryForm,
	}}
	applyPropertySalePublisher(&params, publisher)
	next, err := s.buildPropertyContact(listing.ID, params.Contact, current.ContactMode)
	if err != nil {
		return err
	}
	next.CreatedAt = current.CreatedAt
	if err := tx.WithContext(ctx).Save(next).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to update agency listing contact")
	}
	if err := tx.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Update("publisher_identity_type", "agent").Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to update agency listing identity")
	}
	updates := map[string]any{"agency_company_name": publisher.AgencyCompanyName, "publisher_role_label": publisherRoleLabel("agent")}
	if err := tx.WithContext(ctx).Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Updates(updates).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to update agency listing profile")
	}
	return nil
}
