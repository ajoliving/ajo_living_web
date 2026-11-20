/*
 * 首頁港鐵到站資訊。
 * 1. 已核實的大廈位置以直線距離選站；公開示例與預設站不標為附近。
 * 2. 讀取港鐵公開即時班次，只顯示各方向最近一班未過期的到站。
 */
package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const homeMTRURL = "https://rt.data.gov.hk/v1/transport/mtr/getSchedule.php"
const homeMTRTimeLayout = "2006-01-02 15:04:05"

type homeTransport struct {
	Station    string           `json:"station"`
	Line       string           `json:"line"`
	Nearby     bool             `json:"nearby"`
	ObservedAt string           `json:"observed_at,omitempty"`
	Arrivals   []homeMTRArrival `json:"arrivals"`
	Delayed    bool             `json:"delayed,omitempty"`
}

type homeMTRArrival struct {
	Destination string `json:"destination"`
	ArrivalAt   string `json:"arrival_at"`
}

type homeMTRTrain struct {
	Destination string `json:"dest"`
	Time        string `json:"time"`
	Valid       string `json:"valid"`
	TimeType    string `json:"timeType"`
}

// 1. readTransport requests one trusted station and returns up to two directions.
func (s *HomeEnvironmentService) readTransport(ctx context.Context, loc environmentLocation, lang string) *homeTransport {
	station, nearby := selectHomeMTRStation(loc)
	query := url.Values{"line": {station.Line}, "sta": {station.Code}, "lang": {lang}}.Encode()
	var response struct {
		Status json.RawMessage `json:"status"`
		Curr   string          `json:"curr_time"`
		Delay  string          `json:"isdelay"`
		Data   map[string]struct {
			Curr string         `json:"curr_time"`
			Up   []homeMTRTrain `json:"UP"`
			Down []homeMTRTrain `json:"DOWN"`
		} `json:"data"`
	}
	key := station.Line + "-" + station.Code
	if s.readJSON(ctx, "mtr:"+key+":"+lang, homeMTRURL+"?"+query, 20*time.Second, &response) != nil ||
		strings.Trim(string(response.Status), `"`) != "1" {
		return nil
	}
	schedule, found := response.Data[key]
	if !found {
		return nil
	}
	result := &homeTransport{
		Station: homeMTRStationName(station, lang), Nearby: nearby,
		Delayed: strings.EqualFold(response.Delay, "Y"), Arrivals: []homeMTRArrival{},
	}
	line := homeMTRLineNames[station.Line]
	result.Line = line[0]
	if lang == "en" {
		result.Line = line[1]
	}
	if len(schedule.Up)+len(schedule.Down) == 0 {
		return result
	}
	now := s.runtime.Now()
	sourceTime := schedule.Curr
	if sourceTime == "" || sourceTime == "-" {
		sourceTime = response.Curr
	}
	observed, err := time.ParseInLocation(homeMTRTimeLayout, sourceTime, environmentHKTime)
	if err != nil || now.Sub(observed) > 2*time.Minute || observed.After(now.Add(30*time.Second)) {
		return nil
	}
	result.ObservedAt = observed.Format(time.RFC3339)
	for _, trains := range [][]homeMTRTrain{schedule.Up, schedule.Down} {
		if candidate := nextHomeMTRArrival(trains, now, lang); candidate != nil {
			result.Arrivals = append(result.Arrivals, *candidate)
		}
	}
	return result
}

// 2. selectHomeMTRStation keeps published reference stations separate from nearby matches.
func selectHomeMTRStation(loc environmentLocation) (homeMTRStation, bool) {
	switch loc.Preset {
	case "seaview":
		return homeMTRStationByCode("OCP"), false
	case "summit":
		return homeMTRStationByCode("SUW"), false
	case "fleur":
		return homeMTRStationByCode("TKO"), false
	}
	selected := homeMTRStationByCode("NOP")
	if !loc.pointValid() {
		return selected, false
	}
	best := 1.5 * 1.5
	nearby := false
	for _, station := range homeMTRStations {
		if station.Code == "RAC" {
			continue
		}
		// 2.1 A straight-line radius does not imply a walking route.
		dy, dx := 111.2*(station.Lat-loc.Lat), 102.9*(station.Lon-loc.Lon)
		if distance := dy*dy + dx*dx; distance < best {
			best, selected, nearby = distance, station, true
		}
	}
	return selected, nearby
}

// 3. nextHomeMTRArrival excludes explicit departures and outdated records.
func nextHomeMTRArrival(trains []homeMTRTrain, now time.Time, lang string) *homeMTRArrival {
	var best time.Time
	var result *homeMTRArrival
	for _, train := range trains {
		if strings.EqualFold(train.Valid, "N") || strings.EqualFold(train.TimeType, "D") {
			continue
		}
		arrival, err := time.ParseInLocation(homeMTRTimeLayout, train.Time, environmentHKTime)
		if err != nil || arrival.Before(now) || arrival.After(now.Add(2*time.Hour)) || !best.IsZero() && !arrival.Before(best) {
			continue
		}
		destination := homeMTRStationByCode(train.Destination)
		if destination.Code != "" {
			best = arrival
			result = &homeMTRArrival{homeMTRStationName(destination, lang), arrival.Format(time.RFC3339)}
		}
	}
	return result
}

// 4. homeMTRStationByCode resolves destination labels without guessing names.
func homeMTRStationByCode(code string) homeMTRStation {
	for _, station := range homeMTRStations {
		if station.Code == code {
			return station
		}
	}
	return homeMTRStation{}
}

// 5. homeMTRStationName returns the label in the current homepage language.
func homeMTRStationName(station homeMTRStation, lang string) string {
	if lang == "en" {
		return station.EN
	}
	return station.ZH
}
