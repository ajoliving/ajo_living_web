/*
 * 首頁大廈定位。
 * 1. 只讀取通過既有 iSmart 可見範圍校驗的大廈資料。
 * 2. 採用明確地圖標點，或以名稱及門牌核對地政總署的座標。
 * 3. 不跟隨地圖短網址，不採用地圖中心或模糊搜尋結果。
 */
package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const environmentSearchURL = "https://www.map.gov.hk/gs/api/v1.0.0/locationSearch"
const environmentTransformURL = "https://www.geodetic.gov.hk/transform/v2/"

type environmentLocation struct {
	Lat, Lon float64
	ZH, EN   string
	District string
	Preset   string
}

type environmentAddressMatch struct {
	NameZH     string  `json:"nameZH"`
	NameEN     string  `json:"nameEN"`
	AddressZH  string  `json:"addressZH"`
	AddressEN  string  `json:"addressEN"`
	DistrictZH string  `json:"districtZH"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
}

// 1. pointValid prevents coordinates outside the Hong Kong area from qualifying.
func (l environmentLocation) pointValid() bool {
	return l.Lat >= 22.12 && l.Lat <= 22.58 && l.Lon >= 113.82 && l.Lon <= 114.44
}

// 2. name returns the localized building name without substituting an ID.
func (l environmentLocation) name(lang string) string {
	if lang == "en" && l.EN != "" || l.ZH == "" {
		return l.EN
	}
	return l.ZH
}

// 3. buildingLocation checks member access before using any cached position.
func (s *HomeEnvironmentService) buildingLocation(ctx context.Context, userID int64, buildingID string) environmentLocation {
	if s.ismart == nil {
		return environmentLocation{}
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	_, buildingID, _, err := s.ismart.resolveBuildingAccess(ctx, userID, buildingID)
	if err != nil {
		return environmentLocation{}
	}
	cacheKey := "home:building-position:v2:" + buildingID
	read := s.fetches.DoChan(cacheKey, func() (any, error) {
		sourceCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 7*time.Second)
		defer done()
		cache := s.runtime.CacheStore
		if cache != nil {
			var loc environmentLocation
			if data, ok, err := cache.Get(sourceCtx, cacheKey); err == nil && ok && json.Unmarshal(data, &loc) == nil {
				return loc, nil
			}
		}
		select {
		case s.slots <- struct{}{}:
		case <-sourceCtx.Done():
			return environmentLocation{}, nil
		}
		result, err := s.ismart.loadBuildingInfo(sourceCtx, buildingID)
		<-s.slots
		if err != nil {
			return environmentLocation{}, nil
		}
		loc := s.parseBuildingLocation(sourceCtx, buildingID, paymentMapValue(result.Payload))
		if cache != nil {
			ttl := 30 * time.Second
			if loc.pointValid() {
				ttl = 10 * time.Minute
			}
			if data, err := json.Marshal(loc); err == nil {
				_ = cache.Set(sourceCtx, cacheKey, data, ttl)
			}
		}
		return loc, nil
	})
	select {
	case <-ctx.Done():
		return environmentLocation{}
	case result := <-read:
		loc, _ := result.Val.(environmentLocation)
		return loc
	}
}

// 4. parseBuildingLocation retains only the names and location of a shared building.
func (s *HomeEnvironmentService) parseBuildingLocation(ctx context.Context, buildingID string, info map[string]any) environmentLocation {
	building := paymentMapValue(info["building"])
	details := paymentMapValue(info["building_info"])
	value := func(fields map[string]any, key string) string {
		v, _ := fields[key].(string)
		return strings.TrimSpace(v)
	}
	loc := environmentLocation{ZH: value(building, "buildname_chi"), EN: value(building, "buildname"), District: value(building, "district")}
	if strings.EqualFold(loc.ZH, buildingID) {
		loc.ZH = ""
	}
	if strings.EqualFold(loc.EN, buildingID) {
		loc.EN = ""
	}
	mapURL := value(details, "google_map_url")
	street, number := value(building, "street"), paymentStringValue(building["street_no"])
	fingerprint := sha256.Sum256([]byte(strings.Join([]string{buildingID, loc.ZH, loc.EN, loc.District, street, number, mapURL}, "\x00")))
	key := fmt.Sprintf("home:building-location:v1:%x", fingerprint)
	if cache := s.runtime.CacheStore; cache != nil {
		var cached environmentLocation
		if data, ok, err := cache.Get(ctx, key); err == nil && ok && json.Unmarshal(data, &cached) == nil && cached.pointValid() {
			return cached
		}
	}
	loc = s.resolveEnvironmentLocation(ctx, loc, mapURL, street, number)
	if cache := s.runtime.CacheStore; cache != nil && loc.pointValid() {
		if data, err := json.Marshal(loc); err == nil {
			_ = cache.Set(ctx, key, data, 7*24*time.Hour)
		}
	}
	return loc
}

// 4. resolveEnvironmentLocation prefers a map marker; geocoding requires both name and address.
func (s *HomeEnvironmentService) resolveEnvironmentLocation(ctx context.Context, loc environmentLocation, mapURL, street, number string) environmentLocation {
	loc.Lat, loc.Lon = environmentMapPoint(mapURL)
	if loc.pointValid() && environmentDistrictFor(loc.District) != nil {
		return loc
	}
	if street == "" || number == "" || loc.name("tc") == "" {
		return loc
	}
	name := loc.name("tc")
	var matches []environmentAddressMatch
	endpoint := environmentSearchURL + "?" + url.Values{"q": {name}}.Encode()
	if s.readJSON(ctx, "address:"+name, endpoint, 24*time.Hour, &matches) != nil {
		return loc
	}
	match := verifiedEnvironmentAddress(matches, loc, street, number)
	if match == nil {
		return loc
	}
	if !loc.pointValid() {
		query := url.Values{
			"inSys": {"hkgrid"}, "outSys": {"wgsgeog"},
			"e": {strconv.FormatFloat(match.X, 'f', -1, 64)},
			"n": {strconv.FormatFloat(match.Y, 'f', -1, 64)},
		}.Encode()
		var result struct {
			Lat float64 `json:"wgsLat"`
			Lon float64 `json:"wgsLong"`
		}
		if s.readJSON(ctx, "point:"+query, environmentTransformURL+"?"+query, 7*24*time.Hour, &result) != nil {
			return loc
		}
		loc.Lat, loc.Lon = result.Lat, result.Lon
	}
	loc.District = match.DistrictZH
	return loc
}

var (
	environmentIframeSrc = regexp.MustCompile(`(?i)\bsrc\s*=\s*["']([^"']+)["']`)
	environmentCoordPair = regexp.MustCompile(`^\s*(-?[0-9]+(?:\.[0-9]+)?)\s*,\s*(-?[0-9]+(?:\.[0-9]+)?)\s*$`)
	environmentPlacePair = regexp.MustCompile(`!3d(-?[0-9]+(?:\.[0-9]+)?)!4d(-?[0-9]+(?:\.[0-9]+)?)`)
)

// 5. environmentMapPoint extracts an explicit Google place pin, never an @ camera centre.
func environmentMapPoint(raw string) (float64, float64) {
	if match := environmentIframeSrc.FindStringSubmatch(raw); match != nil {
		raw = strings.ReplaceAll(match[1], "&amp;", "&")
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || !strings.HasPrefix(u.Path, "/maps") {
		return 0, 0
	}
	switch strings.ToLower(u.Hostname()) {
	case "www.google.com", "google.com", "maps.google.com", "www.google.com.hk", "google.com.hk", "maps.google.com.hk":
	default:
		return 0, 0
	}
	query := u.Query()
	pair := environmentCoordPair.FindStringSubmatch(query.Get("q"))
	if pair == nil {
		pair = environmentCoordPair.FindStringSubmatch(query.Get("query"))
	}
	if pair == nil && (strings.HasPrefix(u.Path, "/maps/place/") || u.Path == "/maps/embed") {
		matches := environmentPlacePair.FindAllStringSubmatch(u.Path+" "+query.Get("pb"), 2)
		if len(matches) == 1 {
			pair = matches[0]
		}
	}
	if len(pair) != 3 {
		return 0, 0
	}
	lat, _ := strconv.ParseFloat(pair[1], 64)
	lon, _ := strconv.ParseFloat(pair[2], 64)
	return lat, lon
}

// 6. verifiedEnvironmentAddress rejects fuzzy matches and ambiguous coordinates.
func verifiedEnvironmentAddress(matches []environmentAddressMatch, loc environmentLocation, street, number string) *environmentAddressMatch {
	var result *environmentAddressMatch
	zhAddress, enAddress := environmentNormalize(street+number+"號"), environmentNormalize(number+" "+street)
	for _, match := range matches {
		named := loc.ZH != "" && environmentNormalize(loc.ZH) == environmentNormalize(match.NameZH) ||
			loc.EN != "" && environmentNormalize(loc.EN) == environmentNormalize(match.NameEN)
		if !named || (zhAddress != environmentNormalize(match.AddressZH) && enAddress != environmentNormalize(match.AddressEN)) ||
			match.X <= 0 || match.Y <= 0 {
			continue
		}
		if result != nil && (result.X != match.X || result.Y != match.Y) {
			return nil
		}
		candidate := match
		result = &candidate
	}
	return result
}

// 7. environmentNormalize tolerates spacing, punctuation and common street abbreviations.
func environmentNormalize(value string) string {
	fields := strings.Fields(strings.ToLower(value))
	for i, word := range fields {
		switch strings.TrimRight(word, ".") {
		case "st":
			fields[i] = "street"
		case "rd":
			fields[i] = "road"
		}
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, strings.Join(fields, ""))
}
