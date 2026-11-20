/*
 * 首頁社區環境回歸驗證。
 * 1. 核對標點、同名門牌及分區讀數。
 * 2. 驗證訪客的全港天氣與觀測有效期。
 */
package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 1. TestEnvironmentMapPoint distinguishes pins from viewport centers.
func TestEnvironmentMapPoint(t *testing.T) {
	for _, tc := range []struct {
		link string
		want bool
	}{
		{`<iframe src="https://www.google.com/maps/embed?pb=!2d114.1!3d22.1!3d22.28997!4d114.20373"></iframe>`, true},
		{"https://www.google.com/maps?q=22.28997,114.20373", true},
		{"https://www.google.com/maps/@22.28997,114.20373,18z", false},
		{"https://www.google.com/maps/embed?pb=!2d114.20373!3d22.28997", false},
		{"https://www.google.com.example/maps?q=22.28997,114.20373", false},
		{"https://maps.app.goo.gl/example", false},
		{"https://www.google.com/maps?q=0,0", false},
	} {
		lat, lon := environmentMapPoint(tc.link)
		if got := (environmentLocation{Lat: lat, Lon: lon}).pointValid(); got != tc.want {
			t.Errorf("point validity: %q = %t, want %t", tc.link, got, tc.want)
		}
	}
}

// 6. TestEnvironmentSharedFetch verifies a cold burst makes one request even if its first caller leaves.
func TestEnvironmentSharedFetch(t *testing.T) {
	s := NewHomeEnvironmentService(&Runtime{}, nil)
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	s.client.Transport = environmentTransport(func(req *http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		select {
		case <-release:
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("\ufeff" + `{"value":13}`))}, nil
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() {
		var data map[string]int
		first <- s.readJSON(ctx, "burst", environmentWeatherURL, time.Minute, &data)
	}()
	<-started
	var callers sync.WaitGroup
	for i := 0; i < 100; i++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			var data map[string]int
			if err := s.readJSON(context.Background(), "burst", environmentWeatherURL, time.Minute, &data); err != nil || data["value"] != 13 {
				t.Errorf("shared fetch: data=%v err=%v", data, err)
			}
		}()
	}
	cancel()
	if err := <-first; err != context.Canceled {
		t.Errorf("first caller should cancel independently: %v", err)
	}
	time.Sleep(75 * time.Millisecond)
	close(release)
	callers.Wait()
	if calls.Load() != 1 {
		t.Fatalf("expected one shared fetch for 101 clients, got %d", calls.Load())
	}
}

// 7. TestEnvironmentPresetLife verifies the named hospital and the correct combined category.
func TestEnvironmentPresetLife(t *testing.T) {
	var feed emergencyFeed
	if err := json.Unmarshal([]byte(`{"updateTime":"2026年11月20日 下午5時15分","waitTime":[
		{"hospName":"伊利沙伯醫院","t45p50":"1 小時"},
		{"hospName":"將軍澳醫院","t45p50":"2.5 小時"}]}`), &feed); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 20, 17, 18, 0, 0, environmentHKTime)
	got := selectHomeEmergency(feed, homeEstates["fleur"].Location, "tc", now)
	if got == nil || got.Hospital != "將軍澳醫院" || got.Category != "45" || got.Nearby {
		t.Fatalf("wrong reference hospital or category: %+v", got)
	}
	if got := selectHomeEmergency(feed, homeEstates["fleur"].Location, "tc", now.Add(2*time.Hour)); got != nil {
		t.Fatalf("stale emergency wait: %+v", got)
	}
}

// 2. TestEnvironmentAddressMatch requires an exact name and door number.
func TestEnvironmentAddressMatch(t *testing.T) {
	loc := environmentLocation{ZH: "柏蔚山", EN: "Fleur Pavilia"}
	matches := []environmentAddressMatch{
		{NameZH: "柏蔚山第1座", AddressZH: "繼園街1號", X: 1, Y: 2},
		{NameZH: "柏蔚山", AddressZH: "繼園街11號", X: 1, Y: 3},
		{NameZH: "柏蔚山", AddressZH: "繼園街   1號", NameEN: "Fleur Pavilia", AddressEN: "1 KAI YUEN STREET", DistrictZH: "東區", X: 839035, Y: 816785},
	}
	for _, street := range []string{"繼園街", "Kai Yuen St."} {
		match := verifiedEnvironmentAddress(matches, loc, street, "1")
		if match == nil || match.X != 839035 {
			t.Fatalf("address not resolved: %q: %+v", street, match)
		}
	}
	if got := verifiedEnvironmentAddress(matches[:2], loc, "繼園街", "1"); got != nil {
		t.Fatalf("incorrect match: %+v", got)
	}
}

// 3. TestEnvironmentReadings verifies nearest station, zero rainfall, and stale AQHI.
func TestEnvironmentReadings(t *testing.T) {
	now := time.Date(2026, 11, 20, 17, 0, 0, 0, environmentHKTime)
	var weather environmentWeatherData
	if err := json.Unmarshal([]byte(`{
		"temperature":{"recordTime":"2026-11-20T16:45:00+08:00","data":[
			{"place":"香港天文台","value":13},{"place":"啟德跑道公園","value":14},{"place":"跑馬地","value":15}]},
		"rainfall":{"endTime":"2026-11-20T16:45:00+08:00","data":[{"place":"東區","max":0}]}
	}`), &weather); err != nil {
		t.Fatal(err)
	}
	loc := environmentLocation{Lat: 22.28997, Lon: 114.20373, District: "東區"}
	if temp := selectEnvironmentTemperature(weather, loc, "tc", now); temp == nil || temp.Station != "啟德跑道公園" {
		t.Fatalf("wrong regional temperature: %+v", temp)
	}
	if temp := selectEnvironmentTemperature(weather, environmentLocation{}, "tc", now); temp == nil || temp.Station != "香港天文台" {
		t.Fatalf("wrong Hong Kong temperature: %+v", temp)
	}
	district := environmentDistrictFor("Eastern District")
	if rain := selectEnvironmentRainfall(weather, district.label("tc"), now); rain == nil || rain.Value != 0 {
		t.Fatalf("zero rainfall lost: %+v", rain)
	}
	readings := []environmentAirReading{
		{Station: "Central", AQHI: json.RawMessage(`2`), Published: "2026-11-20T16:30:00"},
		{Station: "Eastern", AQHI: json.RawMessage(`"10+"`), Published: "2026-11-20T16:30:00"},
	}
	if air := selectEnvironmentAir(readings, *district, "tc", now); air == nil || air.Value != "10+" || air.ObservedAt != "2026-11-20T16:30:00+08:00" {
		t.Fatalf("invalid regional AQHI: %+v", air)
	}
	if air := selectEnvironmentAir(readings, *district, "tc", now.Add(4*time.Hour)); air != nil {
		t.Fatalf("stale reading returned: %+v", air)
	}
}

type environmentTransport func(*http.Request) (*http.Response, error)

// 4. RoundTrip provides deterministic official feed fixtures.
func (f environmentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// 5. TestEnvironmentGuestRead ignores a guest building ID and uses Hong Kong calendar dates.
func TestEnvironmentGuestRead(t *testing.T) {
	now := time.Date(2026, 11, 20, 17, 30, 0, 0, time.UTC) // Nov 21 in Hong Kong.
	s := NewHomeEnvironmentService(&Runtime{Now: func() time.Time { return now }}, nil)
	s.client.Transport = environmentTransport(func(req *http.Request) (*http.Response, error) {
		var data string
		switch req.URL.Query().Get("dataType") {
		case "rhrread":
			data = `{"temperature":{"recordTime":"2026-11-21T01:15:00+08:00","data":[{"place":"Hong Kong Observatory","value":15}]}}`
		case "fnd":
			data = `{"weatherForecast":[{"forecastDate":"20261121","forecastMintemp":{"value":15},"forecastMaxtemp":{"value":19}},{"forecastDate":"20261122","forecastMintemp":{"value":17},"forecastMaxtemp":{"value":22}}]}`
		default:
			switch req.URL.Host {
			case "www.1823.gov.hk", "www.ha.org.hk", "rt.data.gov.hk":
				data = `{}`
			default:
				t.Errorf("unexpected upstream for guest: %s", req.URL)
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	got := s.Read(context.Background(), 0, "sample", "en", "", false)
	if got.Scope != "hong_kong" || got.BuildingName != "" || got.Rainfall != nil || got.AirQuality != nil {
		t.Fatalf("unexpected guest scope: %+v", got)
	}
	if got.Temperature == nil || got.Forecast == nil || got.Forecast.Date != "2026-11-22" || got.Forecast.Min != 17 {
		t.Fatalf("wrong guest feed: %+v", got)
	}
}
