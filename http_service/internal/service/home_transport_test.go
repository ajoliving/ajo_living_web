/*
 * 首頁港鐵到站回歸測試。
 * 1. 驗證附近車站與公開參考站不混用。
 * 2. 驗證班次時間、線路、有效性及雙向資料。
 */
package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// 1. TestHomeMTRStationSelection respects verified proximity and fixed public presets.
func TestHomeMTRStationSelection(t *testing.T) {
	cases := []struct {
		name   string
		loc    environmentLocation
		code   string
		line   string
		nearby bool
	}{
		{"no location", environmentLocation{}, "NOP", "ISL", false},
		{"outside rail radius", environmentLocation{Lat: 22.15, Lon: 114.37}, "NOP", "ISL", false},
		{"Fleur verified address", environmentLocation{Lat: 22.289967, Lon: 114.203725}, "NOP", "ISL", true},
		{"Sha Tin verified address", environmentLocation{Lat: 22.3824, Lon: 114.1870}, "SHT", "EAL", true},
		{"Seaview reference", environmentLocation{Preset: "seaview"}, "OCP", "SIL", false},
		{"Summit reference", environmentLocation{Preset: "summit"}, "SUW", "TML", false},
		{"Fleur reference", environmentLocation{Preset: "fleur", Lat: 22.289967, Lon: 114.203725}, "TKO", "TKL", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			station, nearby := selectHomeMTRStation(tc.loc)
			if nearby != tc.nearby || station.Code != tc.code || station.Line != tc.line {
				t.Fatalf("station=%+v nearby=%v", station, nearby)
			}
		})
	}
}

// 2. TestHomeMTRSchedule verifies that arrivals are localized, fresh and capped per direction.
func TestHomeMTRSchedule(t *testing.T) {
	now := time.Date(2026, 11, 20, 17, 30, 20, 0, environmentHKTime)
	s := NewHomeEnvironmentService(&Runtime{Now: func() time.Time { return now }}, nil)
	s.client.Transport = environmentTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "rt.data.gov.hk" || req.URL.Query().Get("line") != "ISL" ||
			req.URL.Query().Get("sta") != "NOP" || req.URL.Query().Get("lang") != "en" {
			t.Fatalf("wrong MTR request: %v", req.URL)
		}
		data := `{"status":"1","curr_time":"2026-11-20 17:29:55","isdelay":"Y","data":{"ISL-NOP":{
			"UP":[
				{"dest":"CHW","time":"2026-11-20 17:30:00","valid":"Y"},
				{"dest":"CHW","time":"2026-11-20 17:31:00","valid":"N"},
				{"dest":"CHW","time":"2026-11-20 17:32:00","timeType":"D"},
				{"dest":"CHW","time":"2026-11-20 17:36:00"},
				{"dest":"CHW","time":"2026-11-20 17:33:00","valid":"Y"}],
			"DOWN":[
				{"dest":"UNK","time":"2026-11-20 17:31:00"},
				{"dest":"KET","time":"2026-11-20 17:34:00","valid":"Y"}]
		}}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	got := s.readTransport(context.Background(), environmentLocation{}, "en")
	if got == nil || got.Station != "North Point" || got.Line != "Island Line" || got.Nearby ||
		!got.Delayed || got.ObservedAt != "2026-11-20T17:29:55+08:00" || len(got.Arrivals) != 2 {
		t.Fatalf("unexpected schedule: %+v", got)
	}
	if got.Arrivals[0] != (homeMTRArrival{"Chai Wan", "2026-11-20T17:33:00+08:00"}) ||
		got.Arrivals[1] != (homeMTRArrival{"Kennedy Town", "2026-11-20T17:34:00+08:00"}) {
		t.Fatalf("unexpected arrivals: %+v", got.Arrivals)
	}
}

// 3. TestHomeMTRScheduleAvailability distinguishes a successful empty timetable from rejected feeds.
func TestHomeMTRScheduleAvailability(t *testing.T) {
	now := time.Date(2026, 11, 20, 17, 30, 20, 0, environmentHKTime)
	for _, tc := range []struct {
		name  string
		json  string
		valid bool
	}{
		{"empty timetable", `{"status":1,"data":{"ISL-NOP":{"curr_time":"-"}}}`, true},
		{"API error", `{"status":0}`, false},
		{"wrong station", `{"status":1,"data":{"ISL-KET":{}}}`, false},
		{"stale source", `{"status":1,"curr_time":"2026-11-20 17:00:00","data":{"ISL-NOP":{"UP":[{"dest":"CHW","time":"2026-11-20 17:31:00"}]}}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewHomeEnvironmentService(&Runtime{Now: func() time.Time { return now }}, nil)
			s.client.Transport = environmentTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tc.json))}, nil
			})
			got := s.readTransport(context.Background(), environmentLocation{}, "tc")
			if (got != nil) != tc.valid || got != nil && (got.Arrivals == nil || len(got.Arrivals) != 0) {
				t.Fatalf("unexpected availability: %+v", got)
			}
		})
	}
}
