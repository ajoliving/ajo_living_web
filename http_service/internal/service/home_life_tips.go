/*
 * 首頁生活提示。
 * 1. 按香港日期選取 1823 公眾假期。
 * 2. 保留運輸署私家車實時數量與更新時間。
 * 3. 按距離選擇醫管局的輪候讀數，不按輪候長短推薦醫院。
 */
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	holidayURL       = "https://www.1823.gov.hk/common/ical/"
	parkingURL       = "https://api.data.gov.hk/v1/carpark-info-vacancy?"
	emergencyURL     = "https://www.ha.org.hk/opendata/aed/aedwtdata2-"
	parkingRadiusKM  = 1.5
	hospitalRadiusKM = 10.0
)

type homeHoliday struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	IsToday bool   `json:"is_today"`
}

type homeParking struct {
	Name       string `json:"name"`
	Spaces     *int   `json:"spaces,omitempty"`
	ObservedAt string `json:"observed_at,omitempty"`
	Closed     bool   `json:"closed,omitempty"`
}

type homeEmergency struct {
	Hospital   string `json:"hospital"`
	Wait       string `json:"wait"`
	ObservedAt string `json:"observed_at"`
	Category   string `json:"category"`
	Nearby     bool   `json:"nearby"`
}

// 1. readHoliday selects today or the next listed general holiday.
func (s *HomeEnvironmentService) readHoliday(ctx context.Context, lang string) *homeHoliday {
	var feed struct {
		Calendars []struct {
			Events []struct {
				Start []json.RawMessage `json:"dtstart"`
				Name  string            `json:"summary"`
			} `json:"vevent"`
		} `json:"vcalendar"`
	}
	if s.readJSON(ctx, "holiday:"+lang, holidayURL+lang+".json", 12*time.Hour, &feed) != nil {
		return nil
	}
	today := s.runtime.Now().In(environmentHKTime).Format(time.DateOnly)
	var result *homeHoliday
	for _, calendar := range feed.Calendars {
		for _, event := range calendar.Events {
			var raw string
			if len(event.Start) == 0 || json.Unmarshal(event.Start[0], &raw) != nil || strings.TrimSpace(event.Name) == "" {
				continue
			}
			date, err := time.Parse("20060102", raw)
			if err != nil {
				continue
			}
			day := date.Format(time.DateOnly)
			if day >= today && (result == nil || day < result.Date) {
				result = &homeHoliday{event.Name, day, day == today}
			}
		}
	}
	return result
}

type parkingDetailsFeed struct {
	Results []struct {
		ID        string  `json:"park_Id"`
		Name      string  `json:"name"`
		Status    string  `json:"opening_status"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type parkingVacancyFeed struct {
	Results []struct {
		ID      string `json:"park_Id"`
		Private []struct {
			Type     string `json:"vacancy_type"`
			Count    *int   `json:"vacancy"`
			Update   string `json:"lastupdate"`
			Category string `json:"category"`
		} `json:"privateCar"`
	} `json:"results"`
}

// 2. readParking gets the catalogue and live counts independently, sharing district-sized cache cells.
func (s *HomeEnvironmentService) readParking(ctx context.Context, loc environmentLocation, lang string) *homeParking {
	preset := homeEstates[loc.Preset]
	if !loc.pointValid() && preset.ParkingID == "" {
		return nil
	}
	lat, lon := math.Round(loc.Lat*50)/50, math.Round(loc.Lon*50)/50
	extent := fmt.Sprintf("%.3f,%.3f,%.3f,%.3f", lon-0.027, lat-0.027, lon+0.027, lat+0.027)
	locale := "zh_TW"
	if lang == "en" {
		locale = "en_US"
	}
	locationQuery := "extent"
	if preset.ParkingID != "" {
		locationQuery, extent = "carparkIds", preset.ParkingID
	}
	infoQuery := url.Values{"data": {"info"}, "vehicleTypes": {"privateCar"}, locationQuery: {extent}, "lang": {locale}}.Encode()
	vacQuery := url.Values{"data": {"vacancy"}, "vehicleTypes": {"privateCar"}, locationQuery: {extent}}.Encode()
	var info parkingDetailsFeed
	var vacancy parkingVacancyFeed
	var details sync.WaitGroup
	details.Add(1)
	go func() {
		defer details.Done()
		_ = s.readJSON(ctx, "parking:info:"+infoQuery, parkingURL+infoQuery, 30*time.Minute, &info)
	}()
	_ = s.readJSON(ctx, "parking:vacancy:"+vacQuery, parkingURL+vacQuery, time.Minute, &vacancy)
	details.Wait()
	return selectHomeParking(info, vacancy, loc, s.runtime.Now())
}

// 3. selectHomeParking keeps precise, recent hourly counts, including zero.
func selectHomeParking(info parkingDetailsFeed, vacancy parkingVacancyFeed, loc environmentLocation, now time.Time) *homeParking {
	counts := make(map[string]homeParking)
	for _, park := range vacancy.Results {
		for _, count := range park.Private {
			if count.Type != "A" || count.Count == nil || *count.Count < 0 ||
				(count.Category != "" && count.Category != "HOURLY") {
				continue
			}
			if ts := environmentObservation(strings.Replace(count.Update, " ", "T", 1), now, 15*time.Minute); ts != "" {
				if existing := counts[park.ID]; existing.ObservedAt < ts {
					counts[park.ID] = homeParking{Spaces: count.Count, ObservedAt: ts}
				}
			}
		}
	}
	best := parkingRadiusKM
	var result *homeParking
	for _, park := range info.Results {
		row, ok := counts[park.ID]
		if park.Name == "" {
			continue
		}
		if loc.Preset != "" {
			if park.ID == homeEstates[loc.Preset].ParkingID {
				if park.Status == "CLOSED" {
					return &homeParking{Name: park.Name, Closed: true}
				}
				if !ok || (park.Status != "" && park.Status != "OPEN") {
					return nil
				}
				row.Name = park.Name
				return &row
			}
			continue
		}
		if !ok || (park.Status != "" && park.Status != "OPEN") {
			continue
		}
		if distance := homeDistanceKM(loc, park.Latitude, park.Longitude); distance < best {
			best, row.Name, result = distance, park.Name, &row
		}
	}
	return result
}

type emergencyFeed struct {
	Update string `json:"updateTime"`
	Waits  []struct {
		Name string `json:"hospName"`
		Wait string `json:"t45p50"`
	} `json:"waitTime"`
}

// 4. Hospital locations match the HA feed names to official GeoInfo Map locations.
var homeHospitals = []environmentStation{
	{"Alice Ho Miu Ling Nethersole Hospital", "雅麗氏何妙齡那打素醫院", 22.45861, 114.17481},
	{"Caritas Medical Centre", "明愛醫院", 22.34178, 114.15402},
	{"Kwong Wah Hospital", "廣華醫院", 22.31479, 114.17194},
	{"North District Hospital", "北區醫院", 22.49685, 114.12416},
	{"North Lantau Hospital", "北大嶼山醫院", 22.28216, 113.93925},
	{"Pamela Youde Nethersole Eastern Hospital", "東區尤德夫人那打素醫院", 22.27012, 114.23661},
	{"Pok Oi Hospital", "博愛醫院", 22.44540, 114.04139},
	{"Prince of Wales Hospital", "威爾斯親王醫院", 22.37970, 114.20193},
	{"Princess Margaret Hospital", "瑪嘉烈醫院", 22.34090, 114.13471},
	{"Queen Elizabeth Hospital", "伊利沙伯醫院", 22.30893, 114.17516},
	{"Queen Mary Hospital", "瑪麗醫院", 22.27043, 114.13049},
	{"Ruttonjee Hospital", "律敦治醫院", 22.27550, 114.17503},
	{"St John Hospital", "長洲醫院", 22.20801, 114.03152},
	{"Tin Shui Wai Hospital", "天水圍醫院", 22.45841, 113.99568},
	{"Tseung Kwan O Hospital", "將軍澳醫院", 22.31826, 114.27020},
	{"Tuen Mun Hospital", "屯門醫院", 22.40931, 113.97647},
	{"United Christian Hospital", "基督教聯合醫院", 22.32225, 114.22790},
	{"Yan Chai Hospital", "仁濟醫院", 22.36963, 114.11975},
}

// 5. selectHomeEmergency shows the nearest supported hospital or a named public reference.
func selectHomeEmergency(feed emergencyFeed, loc environmentLocation, lang string, now time.Time) *homeEmergency {
	updated := strings.TrimSpace(feed.Update)
	if lang != "en" {
		updated = strings.NewReplacer("年", "/", "月", "/", "日", " ", "上午", "", "下午", "", "時", ":", "分", "").Replace(updated)
		meridiem := "AM"
		if strings.Contains(feed.Update, "下午") {
			meridiem = "PM"
		}
		chunks := strings.Fields(updated)
		if len(chunks) != 2 {
			return nil
		}
		date, err := time.Parse("2006/1/2", chunks[0])
		if err != nil {
			return nil
		}
		updated = date.Format("02/01/2006") + " " + chunks[1] + meridiem
	}
	stamp, err := time.ParseInLocation("02/01/2006 3:04PM", updated, environmentHKTime)
	if err != nil {
		return nil
	}
	observed := environmentObservation(stamp.Format(time.RFC3339), now, time.Hour)
	if observed == "" {
		return nil
	}
	waits := make(map[string]string)
	for _, row := range feed.Waits {
		wait := strings.TrimSpace(row.Wait)
		if wait != "" && wait != "-" && wait != "--" && wait != "N/A" {
			waits[row.Name] = wait
		}
	}
	best := hospitalRadiusKM
	var nearest, reference *homeEmergency
	for _, hospital := range homeHospitals {
		name := hospital.ZH
		if lang == "en" {
			name = hospital.EN
		}
		wait := waits[name]
		if wait == "" {
			continue
		}
		if loc.Preset != "" {
			if hospital.EN == homeEstates[loc.Preset].Hospital {
				return &homeEmergency{name, wait, observed, "45", false}
			}
			continue
		}
		if hospital.EN == "Queen Elizabeth Hospital" {
			reference = &homeEmergency{name, wait, observed, "45", false}
		}
		if distance := homeDistanceKM(loc, hospital.Lat, hospital.Lon); loc.pointValid() && distance < best {
			best, nearest = distance, &homeEmergency{name, wait, observed, "45", true}
		}
	}
	if nearest != nil {
		return nearest
	}
	return reference
}

// 6. homeDistanceKM ranks local places using a Hong Kong latitude adjustment.
func homeDistanceKM(loc environmentLocation, lat, lon float64) float64 {
	return 111.2 * math.Hypot(lat-loc.Lat, 0.925*(lon-loc.Lon))
}
