/*
 * 首頁社區環境。
 * 1. 經已授權大廈位置選取實測與分區資料。
 * 2. 公開天氣與已核實位置使用 Redis 快取，不快取會員權限。
 * 3. 各氣象來源獨立，失效時不冒充即時資料。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	environmentWeatherURL = "https://data.weather.gov.hk/weatherAPI/opendata/weather.php"
	environmentAirURL     = "https://dashboard.data.gov.hk/api/aqhi-individual?format=json"
)

type HomeEnvironmentService struct {
	runtime *Runtime
	ismart  *IsmartExternalService
	client  *http.Client
	fetches singleflight.Group
	slots   chan struct{}
}

type HomeEnvironment struct {
	Scope        string                  `json:"scope"`
	BuildingName string                  `json:"building_name,omitempty"`
	Temperature  *environmentTemperature `json:"temperature,omitempty"`
	Forecast     *environmentForecast    `json:"forecast,omitempty"`
	ForecastDays []environmentForecast   `json:"forecast_days,omitempty"`
	Humidity     *float64                `json:"humidity,omitempty"`
	UV           *float64                `json:"uv,omitempty"`
	Rainfall     *environmentRainfall    `json:"rainfall,omitempty"`
	AirQuality   *environmentAir         `json:"air_quality,omitempty"`
	Transport    *homeTransport          `json:"transport,omitempty"`
	Holiday      *homeHoliday            `json:"holiday,omitempty"`
	Parking      *homeParking            `json:"parking,omitempty"`
	Emergency    *homeEmergency          `json:"emergency,omitempty"`
}

type environmentTemperature struct {
	Value      float64 `json:"value"`
	Station    string  `json:"station"`
	ObservedAt string  `json:"observed_at"`
}

type environmentForecast struct {
	Date        string  `json:"date"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Description string  `json:"description"`
}

type environmentRainfall struct {
	Value      float64 `json:"value"`
	District   string  `json:"district"`
	ObservedAt string  `json:"observed_at"`
}

type environmentAir struct {
	Value      string `json:"value"`
	Station    string `json:"station"`
	ObservedAt string `json:"observed_at"`
}

type environmentWeatherData struct {
	Humidity struct {
		RecordTime string `json:"recordTime"`
		Data       []struct {
			Value *float64 `json:"value"`
		} `json:"data"`
	} `json:"humidity"`
	UV struct {
		Data []struct {
			Value *float64 `json:"value"`
		} `json:"data"`
	} `json:"uvindex"`
	Temperature struct {
		RecordTime string `json:"recordTime"`
		Data       []struct {
			Place string   `json:"place"`
			Value *float64 `json:"value"`
		} `json:"data"`
	} `json:"temperature"`
	Rainfall struct {
		EndTime string `json:"endTime"`
		Data    []struct {
			Place string   `json:"place"`
			Max   *float64 `json:"max"`
		} `json:"data"`
	} `json:"rainfall"`
}

// 1. NewHomeEnvironmentService wires the existing building proxy and fixed public feeds.
func NewHomeEnvironmentService(runtime *Runtime, ismart *IsmartExternalService) *HomeEnvironmentService {
	return &HomeEnvironmentService{
		runtime: runtime, ismart: ismart,
		slots: make(chan struct{}, 12),
		client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}
}

// 2. Read loads both town-wide resources and authorized building-scoped data.
func (s *HomeEnvironmentService) Read(ctx context.Context, userID int64, buildingID, lang, estate string, detail bool) HomeEnvironment {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	result := HomeEnvironment{Scope: "hong_kong"}
	var weather environmentWeatherData
	reference := homeEstates[estate]
	location := reference.Location
	var air []environmentAirReading
	var emergency emergencyFeed
	var tasks sync.WaitGroup
	tasks.Add(4)
	go func() {
		defer tasks.Done()
		_ = s.readJSON(ctx, "weather:"+lang, environmentWeatherURL+"?dataType=rhrread&lang="+lang, 3*time.Minute, &weather)
	}()
	go func() {
		defer tasks.Done()
		days := s.readForecast(ctx, lang)
		tomorrow := s.runtime.Now().In(environmentHKTime).AddDate(0, 0, 1).Format(time.DateOnly)
		for _, day := range days {
			if day.Date == tomorrow {
				result.Forecast = &day
				break
			}
		}
		if detail {
			result.ForecastDays = days
		}
	}()
	go func() {
		defer tasks.Done()
		result.Holiday = s.readHoliday(ctx, lang)
	}()
	go func() {
		defer tasks.Done()
		_ = s.readJSON(ctx, "emergency:"+lang, emergencyURL+lang+".json", 5*time.Minute, &emergency)
	}()
	var local sync.WaitGroup
	local.Add(1)
	go func() {
		defer local.Done()
		if userID > 0 && buildingID != "" {
			location = s.buildingLocation(ctx, userID, buildingID)
		}
		var transport sync.WaitGroup
		transport.Add(1)
		go func() {
			defer transport.Done()
			result.Transport = s.readTransport(ctx, location, lang)
		}()
		if location.pointValid() || location.Preset != "" {
			result.Parking = s.readParking(ctx, location, lang)
		}
		transport.Wait()
	}()
	if (userID > 0 && buildingID != "") || estate != "" {
		tasks.Add(1)
		go func() {
			defer tasks.Done()
			_ = s.readJSON(ctx, "aqhi", environmentAirURL, 5*time.Minute, &air)
		}()
	}
	tasks.Wait()
	local.Wait()

	now := s.runtime.Now()
	if detail {
		if environmentObservation(weather.Humidity.RecordTime, now, 2*time.Hour) != "" && len(weather.Humidity.Data) > 0 {
			result.Humidity = weather.Humidity.Data[0].Value
		}
		if environmentObservation(weather.Temperature.RecordTime, now, 2*time.Hour) != "" && len(weather.UV.Data) > 0 {
			result.UV = weather.UV.Data[0].Value
		}
	}
	result.Emergency = selectHomeEmergency(emergency, location, lang, now)
	result.Temperature = selectEnvironmentTemperature(weather, location, lang, now)
	if location.pointValid() || location.Preset != "" {
		result.Scope = "building"
		if location.Preset != "" {
			result.Scope = "reference"
		}
		result.BuildingName = location.name(lang)
		if district := environmentDistrictFor(location.District); district != nil {
			source := *district
			if location.Preset != "" {
				source.Air, source.AirZH = reference.AirStation, reference.AirStationZH
			}
			result.Rainfall = selectEnvironmentRainfall(weather, source.label(lang), now)
			result.AirQuality = selectEnvironmentAir(air, source, lang, now)
		}
	}
	return result
}

// 3. readForecast validates Hong Kong calendar days before returning outlooks.
func (s *HomeEnvironmentService) readForecast(ctx context.Context, lang string) []environmentForecast {
	var response struct {
		Weather []struct {
			Date string `json:"forecastDate"`
			Min  struct {
				Value *float64 `json:"value"`
			} `json:"forecastMintemp"`
			Max struct {
				Value *float64 `json:"value"`
			} `json:"forecastMaxtemp"`
			Description string `json:"forecastWeather"`
		} `json:"weatherForecast"`
	}
	if s.readJSON(ctx, "forecast:"+lang, environmentWeatherURL+"?dataType=fnd&lang="+lang, 15*time.Minute, &response) != nil {
		return nil
	}
	today := s.runtime.Now().In(environmentHKTime).Format(time.DateOnly)
	var results []environmentForecast
	for _, day := range response.Weather {
		date, err := time.Parse("20060102", day.Date)
		if err == nil && date.Format(time.DateOnly) >= today && day.Min.Value != nil && day.Max.Value != nil {
			results = append(results, environmentForecast{date.Format(time.DateOnly), *day.Min.Value, *day.Max.Value, strings.TrimSpace(day.Description)})
			if len(results) == 9 {
				break
			}
		}
	}
	return results
}

// 4. readJSON bounds public responses and shares reads through the existing cache.
func (s *HomeEnvironmentService) readJSON(ctx context.Context, key, endpoint string, ttl time.Duration, dest any) error {
	key = "home:environment:v1:" + key
	fetch := s.fetches.DoChan(key, func() (any, error) {
		// 4.1 A disconnected caller must not cancel a fetch shared by other visitors.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		cache := s.runtime.CacheStore
		if cache != nil {
			if data, ok, err := cache.Get(ctx, key); err == nil && ok {
				if string(data) == "-" {
					return nil, fmt.Errorf("environment upstream cooling down")
				}
				if json.Valid(data) {
					return data, nil
				}
			}
		}
		select {
		case s.slots <- struct{}{}:
			defer func() { <-s.slots }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		data, err := s.fetchEnvironmentJSON(ctx, endpoint)
		if cache != nil && err == nil {
			_ = cache.Set(ctx, key, data, ttl)
		} else if cache != nil {
			writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 400*time.Millisecond)
			defer cancel()
			_ = cache.Set(writeCtx, key, []byte("-"), 10*time.Second)
		}
		return data, err
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-fetch:
		if result.Err != nil {
			return result.Err
		}
		return json.Unmarshal(result.Val.([]byte), dest)
	}
}

// 5. fetchEnvironmentJSON only accepts bounded JSON from fixed upstream endpoints.
func (s *HomeEnvironmentService) fetchEnvironmentJSON(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("environment upstream HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	if err != nil || len(data) > 1<<20 || !json.Valid(data) {
		return nil, fmt.Errorf("invalid environment upstream response")
	}
	return data, nil
}

var environmentHKTime = time.FixedZone("HKT", 8*60*60)

type environmentAirReading struct {
	Station   string          `json:"station"`
	AQHI      json.RawMessage `json:"aqhi"`
	Published string          `json:"publish_date"`
}

// 6. environmentObservation rejects future, stale, and missing source timestamps.
func environmentObservation(raw string, now time.Time, maxAge time.Duration) string {
	stamp, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		stamp, err = time.ParseInLocation("2006-01-02T15:04:05", raw, environmentHKTime)
	}
	if err != nil || stamp.After(now.Add(5*time.Minute)) || now.Sub(stamp) > maxAge {
		return ""
	}
	return stamp.Format(time.RFC3339)
}

// 7. selectEnvironmentTemperature picks the reference or closest available station.
func selectEnvironmentTemperature(weather environmentWeatherData, loc environmentLocation, lang string, now time.Time) *environmentTemperature {
	observed := environmentObservation(weather.Temperature.RecordTime, now, 2*time.Hour)
	if observed == "" {
		return nil
	}
	values := make(map[string]float64, len(weather.Temperature.Data))
	for _, reading := range weather.Temperature.Data {
		if reading.Value != nil {
			values[reading.Place] = *reading.Value
		}
	}
	best := -1.0
	var result *environmentTemperature
	for _, station := range environmentStations {
		label := station.EN
		if lang == "tc" {
			label = station.ZH
		}
		value, present := values[label]
		if !present {
			continue
		}
		if loc.Preset != "" {
			if station.EN == homeEstates[loc.Preset].TemperatureStation {
				return &environmentTemperature{value, label, observed}
			}
			continue
		}
		if !loc.pointValid() && station.EN != "Hong Kong Observatory" {
			continue
		}
		// 7.1 Local distance ranking adjusts longitude for Hong Kong's latitude.
		dy, dx := station.Lat-loc.Lat, 0.925*(station.Lon-loc.Lon)
		if distance := dy*dy + dx*dx; best < 0 || distance < best {
			best = distance
			result = &environmentTemperature{value, label, observed}
		}
	}
	return result
}

// 8. selectEnvironmentRainfall preserves a genuine zero and uses the district maximum.
func selectEnvironmentRainfall(weather environmentWeatherData, district string, now time.Time) *environmentRainfall {
	observed := environmentObservation(weather.Rainfall.EndTime, now, 2*time.Hour)
	if observed == "" {
		return nil
	}
	for _, reading := range weather.Rainfall.Data {
		if reading.Place == district && reading.Max != nil && *reading.Max >= 0 {
			return &environmentRainfall{*reading.Max, district, observed}
		}
	}
	return nil
}

// 9. selectEnvironmentAir uses only the district's observation station when present.
func selectEnvironmentAir(readings []environmentAirReading, district environmentDistrict, lang string, now time.Time) *environmentAir {
	for _, reading := range readings {
		if district.Air == "" || reading.Station != district.Air {
			continue
		}
		observed := environmentObservation(reading.Published, now, 3*time.Hour)
		value := strings.Trim(string(reading.AQHI), `"`)
		switch value {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "10+":
			if observed != "" {
				station := district.Air
				if lang == "tc" {
					station = district.AirZH
				}
				return &environmentAir{value, station, observed}
			}
		}
	}
	return nil
}
