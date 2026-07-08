/*
 * 物業地區選項規則。
 * 1. 維護樓盤發布用區域、地區、分區代碼白名單。
 * 2. 提供區域篩選展開與發布校驗。
 */
package service

import "strings"

// 1. 物業區域對應可接受地區與分區代碼
var propertyDistrictCodesByRegion = map[string][]string{
	"hong_kong_island": {
		"hong_kong_island",
		"central_western",
		"wan_chai",
		"eastern",
		"southern",
		"kennedy_town",
		"shek_tong_tsui",
		"sai_ying_pun",
		"sheung_wan",
		"central",
		"admiralty",
		"mid_levels",
		"peak",
		"causeway_bay",
		"happy_valley",
		"tai_hang",
		"so_kon_po",
		"jardines_lookout",
		"tin_hau",
		"braemar_hill",
		"north_point",
		"quarry_bay",
		"sai_wan_ho",
		"shau_kei_wan",
		"chai_wan",
		"siu_sai_wan",
		"pok_fu_lam",
		"aberdeen",
		"ap_lei_chau",
		"wong_chuk_hang",
		"shouson_hill",
		"repulse_bay",
		"chung_hom_kok",
		"stanley",
		"tai_tam",
		"shek_o",
	},
	"kowloon": {
		"kowloon",
		"yau_tsim_mong",
		"sham_shui_po",
		"kowloon_city",
		"wong_tai_sin",
		"kwun_tong",
		"tsim_sha_tsui",
		"yau_ma_tei",
		"west_kowloon_reclamation",
		"kings_park",
		"mong_kok",
		"tai_kok_tsui",
		"mei_foo",
		"lai_chi_kok",
		"cheung_sha_wan",
		"shek_kip_mei",
		"yau_yat_tsuen",
		"tai_wo_ping",
		"stonecutters_island",
		"hung_hom",
		"to_kwa_wan",
		"ma_tau_kok",
		"ma_tau_wai",
		"kai_tak",
		"ho_man_tin",
		"kowloon_tong",
		"beacon_hill",
		"san_po_kong",
		"tung_tau",
		"wang_tau_hom",
		"lok_fu",
		"diamond_hill",
		"tsz_wan_shan",
		"ngau_chi_wan",
		"ping_shek",
		"kowloon_bay",
		"ngau_tau_kok",
		"jordan_valley",
		"sau_mau_ping",
		"lam_tin",
		"yau_tong",
		"lei_yue_mun",
	},
	"new_territories": {
		"new_territories",
		"kwai_tsing",
		"tsuen_wan",
		"tuen_mun",
		"yuen_long",
		"north",
		"tai_po",
		"sha_tin",
		"sai_kung",
		"islands",
		"kwai_chung",
		"tsing_yi",
		"lei_muk_shue",
		"ting_kau",
		"sham_tseng",
		"tsing_lung_tau",
		"ma_wan",
		"sunny_bay",
		"tai_lam_chung",
		"so_kwun_wat",
		"lam_tei",
		"hung_shui_kiu",
		"ha_tsuen",
		"lau_fau_shan",
		"tin_shui_wai",
		"san_tin",
		"lok_ma_chau",
		"kam_tin",
		"shek_kong",
		"pat_heung",
		"fanling",
		"luen_wo_hui",
		"sheung_shui",
		"shek_wu_hui",
		"sha_tau_kok",
		"luk_keng",
		"wu_kau_tang",
		"tai_po_market",
		"tai_po_kau",
		"tai_mei_tuk",
		"shuen_wan",
		"cheung_muk_tau",
		"kei_ling_ha",
		"tai_wai",
		"fo_tan",
		"ma_liu_shui",
		"wu_kai_sha",
		"ma_on_shan",
		"clear_water_bay",
		"tai_mong_tsai",
		"tseung_kwan_o",
		"hang_hau",
		"tiu_keng_leng",
		"ma_yau_tong",
		"cheung_chau",
		"peng_chau",
		"lantau_island_including_tung_chung",
		"lamma_island",
	},
	"outlying_islands": {
		"outlying_islands",
		"islands",
		"cheung_chau",
		"peng_chau",
		"lantau_island_including_tung_chung",
		"lamma_island",
	},
}

// 2. propertyLocationDistrictsForRegion returns accepted district codes for property regions.
func propertyLocationDistrictsForRegion(value string) ([]string, bool) {
	districts, ok := propertyDistrictCodesByRegion[value]
	return districts, ok
}

// 3. propertyDistrictSearchCodes returns compatible codes for legacy address data.
func propertyDistrictSearchCodes(value string) []string {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return nil
	}

	result := []string{normalizedValue}
	for region, districts := range propertyDistrictCodesByRegion {
		if normalizedValue == region {
			return result
		}
		if isAllowedPropertyValue(normalizedValue, districts) {
			result = append(result, region)
		}
	}

	uniqueCodes := make([]string, 0, len(result))
	seen := map[string]bool{}
	for _, code := range result {
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		uniqueCodes = append(uniqueCodes, code)
	}

	return uniqueCodes
}

// 4. isAllowedPropertyLocationCode checks whether a location code is accepted.
func isAllowedPropertyLocationCode(value string) bool {
	for _, districts := range propertyDistrictCodesByRegion {
		if isAllowedPropertyValue(value, districts) {
			return true
		}
	}

	return false
}
