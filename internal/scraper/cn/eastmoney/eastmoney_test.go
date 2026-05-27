package eastmoney

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/shopspring/decimal"
)

func TestResolveSecid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// User-friendly codes get auto-prefixed.
		{"Shanghai 6xxxxx", "600000", "1.600000"},
		{"Shanghai 6xxxxx alt", "601318", "1.601318"},
		{"Shenzhen 0xxxxx", "000001", "0.000001"},
		{"Shenzhen 3xxxxx (ChiNext)", "300750", "0.300750"},
		// Explicit market.code is passed through untouched.
		{"Explicit Shanghai", "1.600000", "1.600000"},
		{"Explicit Shenzhen", "0.000001", "0.000001"},
		{"Explicit Hong Kong", "116.00700", "116.00700"},
		{"Explicit Nasdaq", "105.AAPL", "105.AAPL"},
		{"Explicit NYSE", "106.BABA", "106.BABA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSecid(tt.in)
			if err != nil {
				t.Fatalf("resolveSecid(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("resolveSecid(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveSecidInvalid(t *testing.T) {
	cases := []string{
		"",
		"abc",
		"12345",   // 5 digits is not valid
		"9999999", // 7 digits is not valid
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := resolveSecid(c); err == nil {
				t.Errorf("expected error for invalid code %q", c)
			}
		})
	}
}

func TestParseKlinesFixture(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "600000.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	prices, err := parseKlines(body, "600000", "PUFA")
	if err != nil {
		t.Fatalf("parseKlines returned error: %v", err)
	}

	if len(prices) != 3 {
		t.Fatalf("expected 3 prices, got %d", len(prices))
	}

	// kline string: date,open,close,high,low,...   close is column 2.
	if !prices[0].Value.Equal(decimal.NewFromFloat(11.30)) {
		t.Errorf("first close = %s, want 11.30", prices[0].Value.String())
	}
	if !prices[1].Value.Equal(decimal.NewFromFloat(11.25)) {
		t.Errorf("second close = %s, want 11.25", prices[1].Value.String())
	}
	if !prices[2].Value.Equal(decimal.NewFromFloat(11.40)) {
		t.Errorf("third close = %s, want 11.40", prices[2].Value.String())
	}

	if prices[0].CommodityType != config.Stock {
		t.Errorf("CommodityType = %s, want stock", prices[0].CommodityType)
	}
	if prices[0].CommodityID != "600000" {
		t.Errorf("CommodityID = %s, want 600000", prices[0].CommodityID)
	}
	if prices[0].CommodityName != "PUFA" {
		t.Errorf("CommodityName = %s, want PUFA", prices[0].CommodityName)
	}

	if y, m, d := prices[0].Date.Date(); y != 2024 || m != 1 || d != 2 {
		t.Errorf("first date = %s, want 2024-01-02", prices[0].Date)
	}
}

func TestParseKlinesEmpty(t *testing.T) {
	body := []byte(`{"data":{"code":"600000","name":"","klines":[]}}`)
	prices, err := parseKlines(body, "600000", "PUFA")
	if err != nil {
		t.Fatalf("parseKlines returned error for empty klines: %v", err)
	}
	if len(prices) != 0 {
		t.Errorf("expected 0 prices for empty klines, got %d", len(prices))
	}
}

func TestParseKlinesNullData(t *testing.T) {
	// eastmoney returns "data": null for delisted / unknown tickers.
	body := []byte(`{"data":null,"rc":0}`)
	prices, err := parseKlines(body, "999999", "X")
	if err != nil {
		t.Fatalf("parseKlines should tolerate null data, got error: %v", err)
	}
	if len(prices) != 0 {
		t.Errorf("expected 0 prices for null data, got %d", len(prices))
	}
}

func TestParseKlinesSkipsMalformedRows(t *testing.T) {
	body := []byte(`{"data":{"code":"600000","name":"","klines":[
		"2024-01-02,11.20,11.30,11.35,11.15,1,1,1,1,1,1",
		"this is garbage",
		"2024-01-03,bad-close-not-a-number,abc,1,1,1,1,1,1,1,1",
		"2024-01-04,11.25,11.40,11.45,11.20,1,1,1,1,1,1"
	]}}`)
	prices, err := parseKlines(body, "600000", "X")
	if err != nil {
		t.Fatalf("parseKlines returned error: %v", err)
	}
	// Should keep 2 well-formed rows and skip 2 malformed ones.
	if len(prices) != 2 {
		t.Fatalf("expected 2 prices, got %d", len(prices))
	}
}

func TestGetPricesUsesHTTP(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "600000.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	// Swap base URL during the test.
	prev := baseURL
	baseURL = srv.URL + "/api/qt/stock/kline/get"
	defer func() { baseURL = prev }()
	// Also clear cache so the request actually hits the server.
	clearCache()

	provider := &PriceProvider{}
	prices, err := provider.GetPrices("600000", "PUFA")
	if err != nil {
		t.Fatalf("GetPrices returned error: %v", err)
	}
	if len(prices) != 3 {
		t.Fatalf("expected 3 prices, got %d", len(prices))
	}
	if !strings.Contains(gotURL, "secid=1.600000") {
		t.Errorf("expected secid=1.600000 in URL, got %s", gotURL)
	}
	if !strings.Contains(gotURL, "klt=101") {
		t.Errorf("expected klt=101 (daily) in URL, got %s", gotURL)
	}
}

func TestProviderMetadata(t *testing.T) {
	p := &PriceProvider{}
	if p.Code() != "cn-eastmoney" {
		t.Errorf("Code() = %q, want cn-eastmoney", p.Code())
	}
	if p.Label() == "" {
		t.Error("Label() is empty")
	}
	if p.Description() == "" {
		t.Error("Description() is empty")
	}
	if len(p.AutoCompleteFields()) == 0 {
		t.Error("AutoCompleteFields() returned no fields")
	}
}

func TestCachingDoesNotHammerServer(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "600000.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	prev := baseURL
	baseURL = srv.URL + "/api/qt/stock/kline/get"
	defer func() { baseURL = prev }()
	clearCache()

	provider := &PriceProvider{}
	for i := 0; i < 3; i++ {
		if _, err := provider.GetPrices("600000", "PUFA"); err != nil {
			t.Fatalf("call %d returned error: %v", i, err)
		}
	}
	if hits != 1 {
		t.Errorf("expected exactly 1 HTTP hit due to caching, got %d", hits)
	}
}
