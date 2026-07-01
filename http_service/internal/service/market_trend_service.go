/*
 * 市場走勢服務。
 * 1. 從香港公開數據讀取私人住宅平均租金 CSV。
 * 2. 解析最近月份並換算為每平方呎月租。
 * 3. 以短期記憶體快取降低外部資料源請求。
 */
package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

const (
	marketRentTrendSourceURL     = "https://www.rvd.gov.hk/datagovhk/1.1M.csv"
	marketRentTrendSourceName    = "Rating and Valuation Department, DATA.GOV.HK"
	marketRentTrendDatasetName   = "Private Domestic - Average Rents by Class - Monthly"
	marketRentTrendCacheTTL      = 6 * time.Hour
	marketRentTrendMonthLimit    = 6
	squareMetreToSquareFootRatio = 10.76391041671
)

// 1. MarketTrendService handles public market data integrations.
type MarketTrendService struct {
	runtime *Runtime
	cacheMu sync.Mutex
	cache   *marketRentTrendCache
}

// 2. MarketRentTrendResponse defines the rent trend payload.
type MarketRentTrendResponse struct {
	Source       string                  `json:"source"`
	SourceURL    string                  `json:"source_url"`
	Dataset      string                  `json:"dataset"`
	Unit         string                  `json:"unit"`
	DisplayUnit  string                  `json:"display_unit"`
	UpdatedMonth string                  `json:"updated_month"`
	Method       string                  `json:"method"`
	Months       []string                `json:"months"`
	Regions      []MarketRentTrendRegion `json:"regions"`
}

// 3. MarketRentTrendRegion defines one region series.
type MarketRentTrendRegion struct {
	Key                  string                 `json:"key"`
	Label                string                 `json:"label"`
	LatestHKDPerSqft     float64                `json:"latest_hkd_per_sqft"`
	MonthlyChangePercent float64                `json:"monthly_change_percent"`
	Direction            string                 `json:"direction"`
	Points               []MarketRentTrendPoint `json:"points"`
}

// 4. MarketRentTrendPoint defines one monthly data point.
type MarketRentTrendPoint struct {
	Month                string  `json:"month"`
	Label                string  `json:"label"`
	ValueHKDPerSqft      float64 `json:"value_hkd_per_sqft"`
	SourceValueHKDPerSqm float64 `json:"source_value_hkd_per_sqm"`
}

// 5. marketRentRegionColumns stores source columns for one region.
type marketRentRegionColumns struct {
	key     string
	label   string
	columns []string
}

// 6. marketRentTrendMonth stores parsed region values for one month.
type marketRentTrendMonth struct {
	month        string
	label        string
	regionValues map[string]marketRentRegionValue
}

// 7. marketRentRegionValue stores one calculated region value.
type marketRentRegionValue struct {
	hkdPerSqm  float64
	hkdPerSqft float64
}

// 8. marketRentTrendCache stores cached trend output.
type marketRentTrendCache struct {
	expiresAt time.Time
	payload   *MarketRentTrendResponse
}

var marketRentTrendRegions = []marketRentRegionColumns{
	{
		key:   "hk",
		label: "香港島",
		columns: []string{
			"Class A Hong Kong",
			"Class B Hong Kong",
			"Class C Hong Kong",
		},
	},
	{
		key:   "kln",
		label: "九龍",
		columns: []string{
			"Class A Kowloon",
			"Class B Kowloon",
			"Class C Kowloon",
		},
	},
	{
		key:   "nt",
		label: "新界",
		columns: []string{
			"Class A New Territories",
			"Class B New Territories",
			"Class C New Territories",
		},
	},
}

// 9. NewMarketTrendService creates a market trend service.
func NewMarketTrendService(runtime *Runtime) *MarketTrendService {
	return &MarketTrendService{runtime: runtime}
}

// 10. GetRentTrend returns recent Hong Kong private domestic rent trend data.
func (s *MarketTrendService) GetRentTrend(ctx context.Context) (*MarketRentTrendResponse, error) {
	if payload, ok := s.cachedRentTrend(); ok {
		return payload, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, marketRentTrendSourceURL, nil)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare market trend request")
	}

	response, err := s.marketTrendHTTPClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load market trend data")
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errcode.New(errcode.CodeInternalError, "market trend source request failed")
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read market trend data")
	}

	payload, err := parseMarketRentTrendCSV(body)
	if err != nil {
		return nil, err
	}
	s.saveRentTrendCache(payload)

	return payload, nil
}

// 11. cachedRentTrend returns the cached payload when it is still valid.
func (s *MarketTrendService) cachedRentTrend() (*MarketRentTrendResponse, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if s.cache == nil || s.now().After(s.cache.expiresAt) {
		s.cache = nil
		return nil, false
	}

	return s.cache.payload, true
}

// 12. saveRentTrendCache stores the latest parsed payload.
func (s *MarketTrendService) saveRentTrendCache(payload *MarketRentTrendResponse) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache = &marketRentTrendCache{
		expiresAt: s.now().Add(marketRentTrendCacheTTL),
		payload:   payload,
	}
}

// 13. marketTrendHTTPClient returns an HTTP client for public open data.
func (s *MarketTrendService) marketTrendHTTPClient() *http.Client {
	timeout := s.runtime.Config.GoodPriceRequestTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 14. now returns the configured clock.
func (s *MarketTrendService) now() time.Time {
	if s.runtime.Now != nil {
		return s.runtime.Now()
	}

	return time.Now()
}

// 15. parseMarketRentTrendCSV parses RVD monthly average rent CSV.
func parseMarketRentTrendCSV(payload []byte) (*MarketRentTrendResponse, error) {
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "invalid market trend csv")
	}

	headerIndex := marketRentTrendHeaderIndex(records)
	if headerIndex < 0 {
		return nil, errcode.New(errcode.CodeInternalError, "market trend csv header not found")
	}

	header := marketRentTrendHeaderMap(records[headerIndex])
	months := make([]marketRentTrendMonth, 0, len(records)-headerIndex-1)
	for _, record := range records[headerIndex+1:] {
		month, ok := parseMarketRentTrendRecord(record, header)
		if ok {
			months = append(months, month)
		}
	}

	if len(months) < 2 {
		return nil, errcode.New(errcode.CodeInternalError, "market trend data is not enough")
	}

	sort.Slice(months, func(i, j int) bool {
		return months[i].month < months[j].month
	})

	return buildMarketRentTrendResponse(lastMarketRentTrendMonths(months)), nil
}

// 16. marketRentTrendHeaderIndex locates the CSV header row.
func marketRentTrendHeaderIndex(records [][]string) int {
	for index, record := range records {
		if len(record) > 0 && strings.EqualFold(strings.TrimSpace(record[0]), "Month") {
			return index
		}
	}

	return -1
}

// 17. marketRentTrendHeaderMap maps CSV column names to indexes.
func marketRentTrendHeaderMap(header []string) map[string]int {
	result := make(map[string]int, len(header))
	for index, value := range header {
		result[strings.TrimSpace(value)] = index
	}

	return result
}

// 18. parseMarketRentTrendRecord parses one month row.
func parseMarketRentTrendRecord(record []string, header map[string]int) (marketRentTrendMonth, bool) {
	if len(record) == 0 {
		return marketRentTrendMonth{}, false
	}

	month, label, ok := normalizeMarketRentTrendMonth(record[0])
	if !ok {
		return marketRentTrendMonth{}, false
	}

	values := make(map[string]marketRentRegionValue, len(marketRentTrendRegions))
	for _, region := range marketRentTrendRegions {
		average, ok := averageMarketRentColumns(record, header, region.columns)
		if !ok {
			return marketRentTrendMonth{}, false
		}
		values[region.key] = marketRentRegionValue{
			hkdPerSqm:  roundMarketTrendValue(average),
			hkdPerSqft: roundMarketTrendValue(average / squareMetreToSquareFootRatio),
		}
	}

	return marketRentTrendMonth{
		month:        month,
		label:        label,
		regionValues: values,
	}, true
}

// 19. averageMarketRentColumns averages the configured source columns.
func averageMarketRentColumns(record []string, header map[string]int, columns []string) (float64, bool) {
	total := 0.0
	count := 0

	for _, column := range columns {
		index, exists := header[column]
		if !exists || index >= len(record) {
			return 0, false
		}
		value, ok := parseMarketRentNumber(record[index])
		if !ok {
			return 0, false
		}
		total += value
		count++
	}

	if count == 0 {
		return 0, false
	}

	return total / float64(count), true
}

// 20. parseMarketRentNumber parses one numeric CSV value.
func parseMarketRentNumber(value string) (float64, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" || normalized == "-" {
		return 0, false
	}

	result, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, false
	}

	return result, true
}

// 21. normalizeMarketRentTrendMonth converts MM-YYYY to YYYY-MM.
func normalizeMarketRentTrendMonth(value string) (string, string, bool) {
	parsed, err := time.Parse("01-2006", strings.TrimSpace(value))
	if err != nil {
		return "", "", false
	}

	return parsed.Format("2006-01"), parsed.Format("2006年1月"), true
}

// 22. lastMarketRentTrendMonths returns the recent public months.
func lastMarketRentTrendMonths(months []marketRentTrendMonth) []marketRentTrendMonth {
	if len(months) <= marketRentTrendMonthLimit {
		return months
	}

	return months[len(months)-marketRentTrendMonthLimit:]
}

// 23. buildMarketRentTrendResponse maps parsed rows to API payload.
func buildMarketRentTrendResponse(months []marketRentTrendMonth) *MarketRentTrendResponse {
	monthKeys := make([]string, 0, len(months))
	for _, month := range months {
		monthKeys = append(monthKeys, month.month)
	}

	regions := make([]MarketRentTrendRegion, 0, len(marketRentTrendRegions))
	for _, region := range marketRentTrendRegions {
		regions = append(regions, buildMarketRentTrendRegion(region, months))
	}

	return &MarketRentTrendResponse{
		Source:       marketRentTrendSourceName,
		SourceURL:    marketRentTrendSourceURL,
		Dataset:      marketRentTrendDatasetName,
		Unit:         "HKD per square metre per month",
		DisplayUnit:  "HKD per square foot per month",
		UpdatedMonth: months[len(months)-1].month,
		Method:       "A 至 C 類私人住宅各區平均租金簡單平均，並由每平方米月租換算為每平方呎月租。",
		Months:       monthKeys,
		Regions:      regions,
	}
}

// 24. buildMarketRentTrendRegion builds one region series.
func buildMarketRentTrendRegion(region marketRentRegionColumns, months []marketRentTrendMonth) MarketRentTrendRegion {
	points := make([]MarketRentTrendPoint, 0, len(months))
	for _, month := range months {
		value := month.regionValues[region.key]
		points = append(points, MarketRentTrendPoint{
			Month:                month.month,
			Label:                month.label,
			ValueHKDPerSqft:      value.hkdPerSqft,
			SourceValueHKDPerSqm: value.hkdPerSqm,
		})
	}

	latest := points[len(points)-1].ValueHKDPerSqft
	previous := points[len(points)-2].ValueHKDPerSqft
	change := 0.0
	if previous > 0 {
		change = roundMarketTrendValue((latest - previous) / previous * 100)
	}

	return MarketRentTrendRegion{
		Key:                  region.key,
		Label:                region.label,
		LatestHKDPerSqft:     latest,
		MonthlyChangePercent: change,
		Direction:            marketTrendDirection(change),
		Points:               points,
	}
}

// 25. marketTrendDirection maps numeric change to display direction.
func marketTrendDirection(change float64) string {
	if change > 0 {
		return "up"
	}
	if change < 0 {
		return "down"
	}

	return "flat"
}

// 26. roundMarketTrendValue rounds market values to one decimal place.
func roundMarketTrendValue(value float64) float64 {
	return math.Round(value*10) / 10
}
