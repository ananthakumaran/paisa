package okx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// helper to read a fixture file from testdata
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("could not read fixture %s: %v", name, err)
	}
	return data
}

// TestParseCandles verifies that parsing a single OKX response page produces
// price records with the close price (index 4) and the date derived from the
// timestamp in milliseconds.
func TestParseCandles(t *testing.T) {
	body := readFixture(t, "BTC-USDT.json")
	prices, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.Len(t, prices, 4)

	// OKX returns descending order, but the parser should preserve the
	// raw order. The first row in the fixture has close=42600 at
	// 1706486400000 (2024-01-29 UTC).
	first := prices[0]
	assert.Equal(t, "BTC-USDT", first.CommodityID)
	assert.Equal(t, "BITCOIN", first.CommodityName)
	assert.True(t, decimal.NewFromInt(42600).Equal(first.Value),
		"expected close=42600, got %s", first.Value.String())
	assert.Equal(t, int64(1706486400), first.Date.Unix())

	last := prices[3]
	assert.True(t, decimal.NewFromInt(42000).Equal(last.Value),
		"expected close=42000, got %s", last.Value.String())
}

// TestParseCandlesEmpty verifies that an empty data array yields no prices
// and is not treated as an error.
func TestParseCandlesEmpty(t *testing.T) {
	body := readFixture(t, "BTC-USDT-empty.json")
	prices, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.Len(t, prices, 0)
}

// TestParseCandlesAPIError verifies that a non-zero code in the response is
// surfaced as an error rather than silently returning empty data.
func TestParseCandlesAPIError(t *testing.T) {
	body := []byte(`{"code":"50011","msg":"Too Many Requests","data":[]}`)
	_, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.Error(t, err)
}

// TestFetchAllPaginates verifies pagination terminates when the server
// returns an empty data page, and that all collected pages are concatenated.
func TestFetchAllPaginates(t *testing.T) {
	var calls atomic.Int32
	page1 := readFixture(t, "BTC-USDT.json")
	page2 := readFixture(t, "BTC-USDT-page2.json")
	empty := readFixture(t, "BTC-USDT-empty.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch calls.Add(1) {
		case 1:
			_, _ = w.Write(page1)
		case 2:
			_, _ = w.Write(page2)
		default:
			_, _ = w.Write(empty)
		}
	}))
	defer srv.Close()

	prices, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	// 4 from page 1 + 2 from page 2.
	assert.Len(t, prices, 6)
	// Three calls: two with data + one empty terminator.
	assert.Equal(t, int32(3), calls.Load())
}

// TestFetchAllTerminatesOnDegenerateResponse guards against infinite loops
// when the server keeps returning the same oldest timestamp (no progress).
func TestFetchAllTerminatesOnDegenerateResponse(t *testing.T) {
	var calls atomic.Int32
	body := readFixture(t, "BTC-USDT.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	prices, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.NotEmpty(t, prices)
	// Must not infinite-loop on degenerate (non-empty, non-advancing) responses.
	// fetchAll has an upper page bound; we just assert it terminated in a
	// reasonable number of calls.
	assert.Less(t, int(calls.Load()), 200, "pagination must terminate")
}

// TestFetchAllSendsBeforeCursor verifies that subsequent requests carry a
// `before` query param derived from the oldest timestamp of the previous page,
// matching OKX's pagination contract.
func TestFetchAllSendsBeforeCursor(t *testing.T) {
	var seenBefore []string
	page1 := readFixture(t, "BTC-USDT.json")
	empty := readFixture(t, "BTC-USDT-empty.json")
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenBefore = append(seenBefore, r.URL.Query().Get("before"))
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) == 1 {
			_, _ = w.Write(page1)
		} else {
			_, _ = w.Write(empty)
		}
	}))
	defer srv.Close()

	_, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)

	assert.Equal(t, "", seenBefore[0], "first call should not include before")
	// Oldest ts on page 1 is 1706227200000.
	assert.Equal(t, "1706227200000", seenBefore[1],
		"second call should use the oldest ts of page 1 as `before` cursor")
}

// TestProviderMetadata pins the registered code and basic descriptors.
func TestProviderMetadata(t *testing.T) {
	p := &PriceProvider{}
	assert.Equal(t, "cn-okx", p.Code())
	assert.NotEmpty(t, p.Label())
	assert.NotEmpty(t, p.Description())
	fields := p.AutoCompleteFields()
	assert.Greater(t, len(fields), 0)
}

// helper to format ms as int string
func _msStr(ms int64) string { return strconv.FormatInt(ms, 10) }

var _ = fmt.Sprintf
