/*
 * Shared service helpers.
 * 1. Marshal and unmarshal JSON fields used by models.
 * 2. Normalize pagination values and media URLs.
 */
package service

import (
	"encoding/json"
	"strings"

	"gorm.io/datatypes"

	"ajoliving_web/http_service/internal/model"
)

// 1. normalizePagination keeps page and page size within safe bounds.
func normalizePagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// 2. marshalJSON converts a value to datatypes.JSON.
func marshalJSON(value any) (datatypes.JSON, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(bytes), nil
}

// 3. unmarshalStringSlice decodes a JSON string slice.
func unmarshalStringSlice(value datatypes.JSON) []string {
	if len(value) == 0 {
		return []string{}
	}

	var result []string
	if err := json.Unmarshal(value, &result); err != nil {
		return []string{}
	}

	return result
}

// 4. unmarshalRoomTypeSlice decodes serviced apartment room types.
func unmarshalRoomTypeSlice(value datatypes.JSON) []ServicedApartmentRoomTypeInput {
	if len(value) == 0 {
		return []ServicedApartmentRoomTypeInput{}
	}

	var result []ServicedApartmentRoomTypeInput
	if err := json.Unmarshal(value, &result); err != nil {
		return []ServicedApartmentRoomTypeInput{}
	}

	return result
}

// 5. mediaURL builds the public media URL from config and object key.
func (s *SecondhandService) mediaURL(asset *model.MediaAsset) string {
	return buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey)
}

// 6. propertyMediaURL builds the public media URL from config and object key.
func (s *PropertyService) mediaURL(asset *model.MediaAsset) string {
	return buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey)
}

// 7. buildMediaURL builds the public media URL from base URL and object key.
func buildMediaURL(baseURL string, objectKey string) string {
	if strings.TrimSpace(objectKey) == "" {
		return ""
	}

	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

// 8. firstListingImage returns the first image or nil.
func firstListingImage(images []ListingImageResponse) *ListingImageResponse {
	if len(images) == 0 {
		return nil
	}

	image := images[0]
	return &image
}
