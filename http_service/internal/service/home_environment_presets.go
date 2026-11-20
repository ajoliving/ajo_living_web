/*
 * 首頁公開參考屋苑。
 * 1. 按首頁規格提供固定資料站點，不代表屋苑真實座標。
 * 2. 不修改會員已保存的目前大廈或用戶權限。
 */
package service

type homeEstate struct {
	Location                 environmentLocation
	TemperatureStation       string
	AirStation, AirStationZH string
	Hospital, ParkingID      string
}

// 1. Public references follow the approved display mapping, not a geocoding result.
var homeEstates = map[string]homeEstate{
	"seaview": {
		environmentLocation{Preset: "seaview", ZH: "海景灣", EN: "Seaview Bay", District: "南區"},
		"Stanley", "Southern", "南區", "Queen Mary Hospital", "tdc14p6401",
	},
	"summit": {
		environmentLocation{Preset: "summit", ZH: "尚珀·天際", EN: "Summit Sky", District: "油尖旺"},
		"King's Park", "Mong Kok", "旺角", "Queen Elizabeth Hospital", "tdc368p1",
	},
	"fleur": {
		environmentLocation{Preset: "fleur", ZH: "柏蔚山", EN: "Fleur Pavilia", District: "西貢"},
		"Tseung Kwan O", "Tseung Kwan O", "將軍澳", "Tseung Kwan O Hospital", "tdc14p1008",
	},
}

// 2. ValidHomeEstate prevents arbitrary upstream keys from public requests.
func ValidHomeEstate(key string) bool {
	_, ok := homeEstates[key]
	return ok
}
