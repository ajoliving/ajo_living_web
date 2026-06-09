/*
 * Property address reference service.
 * 1. Search imported building address records for property publishing autocomplete.
 * 2. Import cleaned address rows from public source spreadsheets.
 * 3. Keep address data separate from listing business records.
 */
package service

import (
	"context"
	"path/filepath"
	"strings"

	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. SearchPropertyAddresses returns building address suggestions.
func (s *PropertyService) SearchPropertyAddresses(ctx context.Context, keyword string, districtCode string, limit int) ([]PropertyAddressSuggestion, error) {
	normalizedKeyword := strings.TrimSpace(keyword)
	if normalizedKeyword == "" {
		return []PropertyAddressSuggestion{}, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	query := s.runtime.DB.WithContext(ctx).Model(&model.PropertyAddress{})
	if strings.TrimSpace(districtCode) != "" {
		query = query.Where("district_code = ?", strings.TrimSpace(districtCode))
	}
	likeKeyword := "%" + normalizedKeyword + "%"
	query = query.Where("(estate_name LIKE ? OR estate_name_en ILIKE ? OR address_text LIKE ? OR address_text_en ILIKE ? OR search_text ILIKE ?)", likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword)

	var rows []model.PropertyAddress
	if err := query.Order("estate_name asc, id asc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to search property addresses")
	}

	result := make([]PropertyAddressSuggestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, toPropertyAddressSuggestion(row))
	}
	return result, nil
}

// 2. ImportPropertyAddressRows upserts cleaned address rows.
func (s *PropertyService) ImportPropertyAddressRows(ctx context.Context, rows []PropertyAddressSuggestion, sourceFile string) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	imported := 0
	for _, row := range rows {
		if strings.TrimSpace(row.EstateName) == "" || strings.TrimSpace(row.AddressText) == "" {
			continue
		}
		blockNames, err := marshalJSON(row.BlockNames)
		if err != nil {
			return imported, err
		}
		record := model.PropertyAddress{
			PublicID:       utils.NewPublicID(),
			SourceFile:     filepath.Base(sourceFile),
			RegionCode:     strings.TrimSpace(row.RegionCode),
			DistrictCode:   strings.TrimSpace(row.DistrictCode),
			EstateName:     strings.TrimSpace(row.EstateName),
			EstateNameEn:   strings.TrimSpace(row.EstateNameEn),
			DisplayName:    strings.TrimSpace(row.DisplayName),
			AddressText:    strings.TrimSpace(row.AddressText),
			AddressTextEn:  strings.TrimSpace(row.AddressTextEn),
			DistrictLabel:  strings.TrimSpace(row.DistrictLabel),
			BlockNames:     blockNames,
			CompletionYear: row.CompletionYear,
			Remark:         strings.TrimSpace(row.Remark),
			SearchText: strings.Join([]string{
				strings.TrimSpace(row.EstateName),
				strings.TrimSpace(row.EstateNameEn),
				strings.TrimSpace(row.DisplayName),
				strings.TrimSpace(row.AddressText),
				strings.TrimSpace(row.AddressTextEn),
				strings.TrimSpace(row.DistrictLabel),
				strings.Join(row.BlockNames, " "),
			}, " "),
		}
		if err := s.runtime.DB.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "source_file"}, {Name: "estate_name"}, {Name: "address_text"}},
			DoUpdates: clause.AssignmentColumns([]string{"estate_name_en", "display_name", "address_text_en", "district_label", "block_names", "completion_year", "remark", "search_text", "updated_at"}),
		}).Create(&record).Error; err != nil {
			return imported, errcode.New(errcode.CodeInternalError, "failed to import property address")
		}
		imported++
	}

	return imported, nil
}

// 3. toPropertyAddressSuggestion maps a model row to API output.
func toPropertyAddressSuggestion(row model.PropertyAddress) PropertyAddressSuggestion {
	return PropertyAddressSuggestion{
		AddressID:      row.PublicID,
		EstateName:     row.EstateName,
		EstateNameEn:   row.EstateNameEn,
		DisplayName:    row.DisplayName,
		AddressText:    row.AddressText,
		AddressTextEn:  row.AddressTextEn,
		DistrictCode:   row.DistrictCode,
		DistrictLabel:  row.DistrictLabel,
		RegionCode:     row.RegionCode,
		BlockNames:     decodeStringSliceBytes([]byte(row.BlockNames)),
		CompletionYear: row.CompletionYear,
		Remark:         row.Remark,
	}
}
