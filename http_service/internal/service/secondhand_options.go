/*
 * Secondhand listing option rules.
 * 1. Keep category, district, and region codes consistent for list filters and publishing.
 * 2. Provide small validation helpers used by secondhand service logic.
 */
package service

// 1. Secondhand category codes accepted by the marketplace.
var allowedSecondhandCategoryCodes = []string{
	"home_furniture",
	"home_appliance",
	"electronics",
	"music",
	"baby_goods",
	"office_furniture",
	"home_decor",
	"other",
}

// 2. Secondhand district codes grouped by top-level Hong Kong region.
var secondhandDistrictCodesByRegion = map[string][]string{
	"hong_kong_island": {
		"hong_kong_island",
		"central_western",
		"wan_chai",
		"eastern",
		"southern",
	},
	"kowloon": {
		"kowloon",
		"yau_tsim_mong",
		"sham_shui_po",
		"kowloon_city",
		"wong_tai_sin",
		"kwun_tong",
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
	},
	"outlying_islands": {
		"outlying_islands",
		"islands",
	},
	"primary_school_net": {
		"primary_school_net",
	},
	"secondary_school_net": {
		"secondary_school_net",
	},
	"tertiary_institution": {
		"tertiary_institution",
	},
}

// 3. isAllowedSecondhandDistrict checks whether a district code is available.
func isAllowedSecondhandDistrict(value string) bool {
	for _, districts := range secondhandDistrictCodesByRegion {
		if isAllowedSecondhandValue(value, districts) {
			return true
		}
	}

	return false
}

// 4. secondhandDistrictsForRegion returns district codes for a region code.
func secondhandDistrictsForRegion(value string) ([]string, bool) {
	districts, ok := secondhandDistrictCodesByRegion[value]
	return districts, ok
}
