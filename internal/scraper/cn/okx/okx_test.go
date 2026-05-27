package okx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

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
// timestamp in milliseconds, anchored in UTC.
func TestParseCandles(t *testing.T) {
	body := readFixture(t, "BTC-USDT.json")
	page, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.Len(t, page.prices, 4)

	// OKX returns descending order; the parser preserves it. The first
	// row in the fixture has close=42600 at 1706486400000
	// (2024-01-29 00:00 UTC).
	first := page.prices[0]
	assert.Equal(t, "BTC-USDT", first.CommodityID)
	assert.Equal(t, "BITCOIN", first.CommodityName)
	assert.True(t, decimal.NewFromInt(42600).Equal(first.Value),
		"expected close=42600, got %s", first.Value.String())
	assert.Equal(t, time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), first.Date)
	assert.Equal(t, time.UTC, first.Date.Location(),
		"Price.Date must be UTC so the day doesn't shift under negative TZs")

	last := page.prices[3]
	assert.True(t, decimal.NewFromInt(42000).Equal(last.Value),
		"expected close=42000, got %s", last.Value.String())

	// oldestMs must equal the raw ts of the LAST row (OKX returns descending).
	assert.Equal(t, "1706227200000", page.oldestMs)
}

// TestParseCandlesEmpty verifies that an empty data array yields no prices
// and is not treated as an error. oldestMs must be empty so fetchAll can
// detect terminator.
func TestParseCandlesEmpty(t *testing.T) {
	body := readFixture(t, "BTC-USDT-empty.json")
	page, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.Len(t, page.prices, 0)
	assert.Equal(t, "", page.oldestMs)
}

// TestParseCandlesAPIError verifies that a non-zero code in the response is
// surfaced as an error rather than silently returning empty data.
func TestParseCandlesAPIError(t *testing.T) {
	body := []byte(`{"code":"50011","msg":"Too Many Requests","data":[]}`)
	_, err := parseCandles(body, "BTC-USDT", "BITCOIN")
	assert.Error(t, err)
}

// newOKXMock returns an httptest server that dispatches by the `after`
// query param to match OKX's real pagination contract:
//   - no `after`                    -> the latest page
//   - after == latest page's oldest -> the next-older page
//   - anything else                 -> empty
//
// This guarantees that any future regression of the cursor direction
// (e.g. accidentally going back to `before`, or sending the wrong
// timestamp) fails fast, instead of the test ratcheting through a
// call-counter.
func newOKXMock(t *testing.T, latest, older, empty []byte, latestOldestMs string, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls != nil {
			calls.Add(1)
		}
		w.Header().Set("Content-Type", "application/json")

		// Sanity-check the mandatory params so the test also pins them.
		q := r.URL.Query()
		assert.Equal(t, "BTC-USDT", q.Get("instId"))
		assert.Equal(t, "1D", q.Get("bar"))
		assert.Equal(t, "100", q.Get("limit"))
		assert.Equal(t, "", q.Get("before"),
			"must never send `before`: that walks forward, not backward")

		switch q.Get("after") {
		case "":
			_, _ = w.Write(latest)
		case latestOldestMs:
			_, _ = w.Write(older)
		default:
			_, _ = w.Write(empty)
		}
	}))
}

// TestFetchAllPaginatesBackward is the critical directional-contract test.
// It mocks OKX's `after` semantics and asserts that the second request
// carries the oldest ts of the first page, producing strictly older
// candles, and that the combined result spans both pages.
func TestFetchAllPaginatesBackward(t *testing.T) {
	page1 := readFixture(t, "BTC-USDT.json")       // 4 candles, newest
	page2 := readFixture(t, "BTC-USDT-page2.json") // 4 candles, older
	empty := readFixture(t, "BTC-USDT-empty.json") // terminator
	const page1OldestMs = "1706227200000"          // ts of page1's oldest row

	var calls atomic.Int32
	srv := newOKXMock(t, page1, page2, empty, page1OldestMs, &calls)
	defer srv.Close()

	prices, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)

	// 4 from latest page + 4 from older page.
	assert.Len(t, prices, 8)
	// Three calls: latest + older + empty terminator.
	assert.Equal(t, int32(3), calls.Load())

	// All page2 candles must be STRICTLY older than all page1 candles —
	// that's the whole point of `after` walking backward.
	page1Oldest := prices[3].Date.Unix() // last of page 1 region
	page2Newest := prices[4].Date.Unix() // first of page 2 region
	assert.Greater(t, page1Oldest, page2Newest,
		"page2 must contain candles strictly older than page1")

	// Spot check the oldest candle (last of page 2 fixture).
	assert.True(t, decimal.NewFromInt(40700).Equal(prices[7].Value),
		"expected oldest close=40700, got %s", prices[7].Value.String())
}

// TestFetchAllUsesAfterNotBefore is a narrow regression test for the
// cursor-direction contract. Even if the mock in newOKXMock changes
// shape, this test alone documents the rule.
func TestFetchAllUsesAfterNotBefore(t *testing.T) {
	page1 := readFixture(t, "BTC-USDT.json")
	empty := readFixture(t, "BTC-USDT-empty.json")
	const page1OldestMs = "1706227200000"

	var seenAfter, seenBefore []string
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAfter = append(seenAfter, r.URL.Query().Get("after"))
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

	assert.Equal(t, []string{"", ""}, seenBefore,
		"`before` must never be set: it walks forward, not backward")
	assert.Equal(t, []string{"", page1OldestMs}, seenAfter,
		"first call sends no cursor; second sends `after=<page1 oldest ts>`")
}

// TestFetchAllTerminatesOnDegenerateResponse guards against infinite loops
// when the server keeps returning the same oldest timestamp (no progress).
// Modeled by an always-the-same response regardless of `after`.
func TestFetchAllTerminatesOnDegenerateResponse(t *testing.T) {
	body := readFixture(t, "BTC-USDT.json")
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	prices, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.NoError(t, err)
	assert.NotEmpty(t, prices)
	// Must not infinite-loop: stalled-cursor guard kicks in on call 2.
	assert.Equal(t, int32(2), calls.Load(),
		"degenerate (non-advancing) response must terminate after one repeat")
}

// TestFetchAllStopsOnAPIError verifies that an upstream error code aborts
// pagination instead of being swallowed page-by-page.
func TestFetchAllStopsOnAPIError(t *testing.T) {
	page1 := readFixture(t, "BTC-USDT.json")
	const errBody = `{"code":"50011","msg":"Too Many Requests","data":[]}`
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) == 1 {
			_, _ = w.Write(page1)
		} else {
			_, _ = w.Write([]byte(errBody))
		}
	}))
	defer srv.Close()

	_, err := fetchAll(srv.URL, "BTC-USDT", "BITCOIN")
	assert.Error(t, err)
	assert.Equal(t, int32(2), calls.Load(),
		"should stop on the page where the API returned an error")
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
