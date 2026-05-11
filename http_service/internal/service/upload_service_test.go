/*
 * Upload service tests.
 * 1. Verify media assets can be completed, listed, queried, and deleted.
 * 2. Verify complete upload rejects missing storage objects.
 * 3. Verify referenced media assets cannot be deleted.
 */
package service

import (
	"context"
	"errors"
	"testing"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. uploadServiceStorageStub controls storage provider responses for upload service tests.
type uploadServiceStorageStub struct {
	headObjectInfo  *StorageObjectInfo
	headObjectErr   error
	deleteObjectErr error
}

// 2. PresignUpload returns a deterministic presign payload for tests.
func (s *uploadServiceStorageStub) PresignUpload(_ context.Context, input PresignUploadInput) (*PresignUploadResult, error) {
	return &PresignUploadResult{
		UploadURL: "http://example.test/upload/" + input.FileName,
		ObjectKey: mediaObjectPrefix + input.FileName,
		Headers: map[string]string{
			"Content-Type": input.MimeType,
		},
	}, nil
}

// 3. PutObject accepts direct upload in tests.
func (s *uploadServiceStorageStub) PutObject(_ context.Context, input PutObjectInput) (*StorageObjectInfo, error) {
	return &StorageObjectInfo{
		ObjectKey:     input.ObjectKey,
		ContentType:   input.MimeType,
		ContentLength: int64(len(input.Body)),
	}, nil
}

// 4. HeadObject returns the configured object metadata for tests.
func (s *uploadServiceStorageStub) HeadObject(_ context.Context, _ string) (*StorageObjectInfo, error) {
	if s.headObjectErr != nil {
		return nil, s.headObjectErr
	}
	if s.headObjectInfo != nil {
		return s.headObjectInfo, nil
	}

	return &StorageObjectInfo{}, nil
}

// 5. DeleteObject returns the configured delete result for tests.
func (s *uploadServiceStorageStub) DeleteObject(_ context.Context, _ string) error {
	return s.deleteObjectErr
}

// 6. TestCompleteListGetAndDeleteMediaAsset verifies the member media asset lifecycle.
func TestCompleteListGetAndDeleteMediaAsset(t *testing.T) {
	runtime := newTestRuntime(t)
	uploadService := NewUploadService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91230021", &communityA.ID)

	created, err := uploadService.CompleteUpload(context.Background(), CompleteUploadParams{
		UserID:         owner.ID,
		ObjectKey:      mediaObjectPrefix + "test-image.webp",
		MimeType:       "image/webp",
		FileSize:       2048,
		Width:          ptrInt(1200),
		Height:         ptrInt(900),
		ChecksumSHA256: "asset-checksum",
	})
	if err != nil {
		t.Fatalf("complete upload: %v", err)
	}

	items, pagination, err := uploadService.ListMediaAssets(context.Background(), owner.ID, MediaAssetListFilters{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list media assets: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 {
		t.Fatalf("expected one media asset, total=%d len=%d", pagination.Total, len(items))
	}
	if items[0].MediaAssetID != created.MediaAssetID || items[0].InUse {
		t.Fatalf("unexpected listed asset: %#v", items[0])
	}

	detail, err := uploadService.GetMediaAsset(context.Background(), owner.ID, created.MediaAssetID)
	if err != nil {
		t.Fatalf("get media asset: %v", err)
	}
	if detail.ObjectKey != mediaObjectPrefix+"test-image.webp" || detail.URL == "" {
		t.Fatalf("unexpected media asset detail: %#v", detail)
	}

	deleted, err := uploadService.DeleteMediaAsset(context.Background(), owner.ID, created.MediaAssetID)
	if err != nil {
		t.Fatalf("delete media asset: %v", err)
	}
	if deleted.MediaAssetID != created.MediaAssetID {
		t.Fatalf("unexpected deleted asset payload: %#v", deleted)
	}

	_, err = uploadService.GetMediaAsset(context.Background(), owner.ID, created.MediaAssetID)
	if err == nil {
		t.Fatalf("expected deleted media asset to be missing")
	}
	assertAppErrorCode(t, err, errcode.CodeNotFound)
}

// 7. TestCompleteUploadRejectsMissingStorageObject verifies storage object validation.
func TestCompleteUploadRejectsMissingStorageObject(t *testing.T) {
	runtime := newTestRuntime(t)
	runtime.StorageProvider = &uploadServiceStorageStub{
		headObjectErr: errStorageObjectNotFound,
	}
	uploadService := NewUploadService(runtime)

	_, err := uploadService.CompleteUpload(context.Background(), CompleteUploadParams{
		UserID:    1,
		ObjectKey: mediaObjectPrefix + "missing.webp",
		MimeType:  "image/webp",
		FileSize:  2048,
	})
	if err == nil {
		t.Fatalf("expected missing storage object to fail")
	}
	assertAppErrorCode(t, err, errcode.CodeValidationError)
}

// 8. TestDeleteMediaAssetRejectsReferencedAsset verifies listing references block deletion.
func TestDeleteMediaAssetRejectsReferencedAsset(t *testing.T) {
	runtime := newTestRuntime(t)
	uploadService := NewUploadService(runtime)
	secondhandService := NewSecondhandService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91230031", &communityA.ID)

	asset, err := uploadService.CompleteUpload(context.Background(), CompleteUploadParams{
		UserID:         owner.ID,
		ObjectKey:      mediaObjectPrefix + "referenced.webp",
		MimeType:       "image/webp",
		FileSize:       2048,
		ChecksumSHA256: "referenced-checksum",
	})
	if err != nil {
		t.Fatalf("complete upload: %v", err)
	}

	_, err = secondhandService.CreateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           owner.ID,
		Title:                 "Reference Asset Listing",
		Summary:               "created for media delete test",
		Description:           "listing holds a media reference",
		DistrictCode:          "kwun_tong",
		CommunityID:           communityA.PublicID,
		PublisherIdentityType: "owner",
		CategoryCode:          "home_appliance",
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(500),
		ConditionLevel:        "used_good",
		DimensionText:         "50 x 50 x 50 cm",
		PickupRegionCode:      "kwun_tong",
		PickupLocationText:    "Lobby pickup",
		DeliveryTags:          []string{"self_pickup"},
		VisibilityScope:       "public",
		ContactMethod:         "both",
		Images: []ListingImageInput{
			{MediaAssetID: asset.MediaAssetID, SortOrder: 1, IsCover: true},
		},
		Contact: ListingContactInput{
			WhatsApp:     "+85291234567",
			ShowWhatsApp: true,
			ShowChat:     true,
		},
	})
	if err != nil {
		t.Fatalf("create secondhand listing: %v", err)
	}

	_, err = uploadService.DeleteMediaAsset(context.Background(), owner.ID, asset.MediaAssetID)
	if err == nil {
		t.Fatalf("expected referenced media asset delete to fail")
	}
	assertAppErrorCode(t, err, errcode.CodeValidationError)
}

// 9. assertAppErrorCode validates the business error code.
func assertAppErrorCode(t *testing.T, err error, expectedCode string) {
	t.Helper()

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	if appErr.Code != expectedCode {
		t.Fatalf("expected error code %s, got %s", expectedCode, appErr.Code)
	}
}

// 10. ptrInt returns an int pointer.
func ptrInt(value int) *int {
	return &value
}
