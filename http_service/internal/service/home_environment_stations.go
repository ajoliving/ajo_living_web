/*
 * 香港分區及氣象站資料。
 * 1. 觀測站座標取自天文台氣象站表，只匹配即時天氣回應已有的測站。
 * 2. 分區雨量與一般空氣監測站使用各自的官方名稱。
 */
package service

type environmentStation struct {
	EN, ZH   string
	Lat, Lon float64
}

// 1. Published station coordinates, rounded to five decimal places.
// Source: https://www.hko.gov.hk/en/cis/stn.htm (2026-11-20).
var environmentStations = []environmentStation{
	{"Hong Kong Observatory", "香港天文台", 22.30194, 114.17417},
	{"King's Park", "京士柏", 22.31194, 114.17278},
	{"Wong Chuk Hang", "黃竹坑", 22.24778, 114.17361},
	{"Ta Kwu Ling", "打鼓嶺", 22.52861, 114.15667},
	{"Lau Fau Shan", "流浮山", 22.46889, 113.98361},
	{"Tai Po", "大埔", 22.44833, 114.17722},
	{"Sha Tin", "沙田", 22.40250, 114.21000},
	{"Tuen Mun", "屯門", 22.38583, 113.96417},
	{"Tseung Kwan O", "將軍澳", 22.31583, 114.25556},
	{"Sai Kung", "西貢", 22.37556, 114.27444},
	{"Cheung Chau", "長洲", 22.20111, 114.02667},
	{"Chek Lap Kok", "赤鱲角", 22.30944, 113.92194},
	{"Tsing Yi", "青衣", 22.34417, 114.11000},
	{"Shek Kong", "石崗", 22.43611, 114.08472},
	{"Tsuen Wan Ho Koon", "荃灣可觀", 22.38361, 114.10778},
	{"Tsuen Wan Shing Mun Valley", "荃灣城門谷", 22.37556, 114.12667},
	{"Hong Kong Park", "香港公園", 22.27833, 114.16222},
	{"Shau Kei Wan", "筲箕灣", 22.28167, 114.23611},
	{"Kowloon City", "九龍城", 22.33500, 114.18472},
	{"Happy Valley", "跑馬地", 22.27056, 114.18361},
	{"Wong Tai Sin", "黃大仙", 22.33944, 114.20528},
	{"Stanley", "赤柱", 22.21417, 114.21861},
	{"Kwun Tong", "觀塘", 22.31861, 114.22472},
	{"Sham Shui Po", "深水埗", 22.33583, 114.13694},
	{"Kai Tak Runway Park", "啟德跑道公園", 22.30500, 114.21694},
	{"Yuen Long Park", "元朗公園", 22.44083, 114.01833},
	{"Tai Mei Tuk", "大美督", 22.47528, 114.23750},
}

type environmentDistrict struct {
	ZH, EN     string
	Air, AirZH string
}

var environmentDistricts = []environmentDistrict{
	{"中西區", "Central & Western District", "Central/Western", "中西區"},
	{"東區", "Eastern District", "Eastern", "東區"},
	{"南區", "Southern District", "Southern", "南區"},
	{"灣仔", "Wan Chai", "", ""},
	{"九龍城", "Kowloon City", "", ""},
	{"觀塘", "Kwun Tong", "Kwun Tong", "觀塘"},
	{"深水埗", "Sham Shui Po", "Sham Shui Po", "深水埗"},
	{"黃大仙", "Wong Tai Sin", "", ""},
	{"油尖旺", "Yau Tsim Mong", "", ""},
	{"葵青", "Kwai Tsing", "Kwai Chung", "葵涌"},
	{"北區", "North District", "North", "北區"},
	{"離島區", "Islands District", "", ""},
	{"西貢", "Sai Kung", "", ""},
	{"沙田", "Sha Tin", "Sha Tin", "沙田"},
	{"大埔", "Tai Po", "Tai Po", "大埔"},
	{"荃灣", "Tsuen Wan", "Tsuen Wan", "荃灣"},
	{"屯門", "Tuen Mun", "Tuen Mun", "屯門"},
	{"元朗", "Yuen Long", "Yuen Long", "元朗"},
}

// 2. label returns the weather feed's district label for its requested language.
func (d environmentDistrict) label(lang string) string {
	if lang == "en" {
		return d.EN
	}
	return d.ZH
}

// 3. environmentDistrictFor accepts the same district with or without its suffix.
func environmentDistrictFor(name string) *environmentDistrict {
	normalized := environmentNormalize(name)
	for _, district := range environmentDistricts {
		zh, en := environmentNormalize(district.ZH), environmentNormalize(district.EN)
		if normalized == zh || normalized == en ||
			normalized == zh+"區" || normalized == en+"district" {
			return &district
		}
	}
	return nil
}
