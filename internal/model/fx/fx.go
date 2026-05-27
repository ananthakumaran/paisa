// Package fx provides a base-currency-aware FX rate store and helpers for
// converting amounts denominated in one currency to a configured base
// currency (e.g. CNY).
//
// The store is intentionally process-local and seeded by the scrapers in
// `internal/scraper/cn/boc` and `internal/scraper/yahoo`. Rates are looked up
// by "most recent rate on or before the requested date" which mirrors how
// `internal/service/market.go` handles commodity prices.
//
// When a direct base/target pair isn't available, the store falls back to
// (1) inverting a known reverse rate, then (2) pivoting through USD. This
// matches the FX provider chain expectation from the M1-F design: the
// frankfurter.app endpoint sometimes lacks HKD->CNY historicals, so we use
// HKD->USD * USD->CNY as the fallback path.
package fx

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// RateStore caches FX rates keyed by (from, to). It is safe for concurrent
// reads after rates are loaded.
type RateStore struct {
	mu    sync.RWMutex
	rates map[string]map[string][]ratePoint // from -> to -> sorted-by-date asc
}

type ratePoint struct {
	date  time.Time
	value decimal.Decimal
}

// NewRateStore returns an empty store.
func NewRateStore() *RateStore {
	return &RateStore{rates: make(map[string]map[string][]ratePoint)}
}

// Put inserts a rate point into the store. Subsequent inserts re-sort the
// time series so lookups stay O(log n).
//
// Same-day duplicates are coalesced: if an entry already exists for `date`,
// the value is overwritten rather than appended. This keeps the store stable
// across repeat seeder invocations (sync runs every time `/api/update` is
// called, and the store is process-cached — see Store/ClearStore).
func (s *RateStore) Put(from, to string, date time.Time, value decimal.Decimal) {
	if from == to {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rates[from] == nil {
		s.rates[from] = make(map[string][]ratePoint)
	}
	series := s.rates[from][to]
	// Dedup-by-date: overwrite an existing same-day point if present.
	day := dayKey(date)
	for i := range series {
		if dayKey(series[i].date) == day {
			series[i].value = value
			s.rates[from][to] = series
			return
		}
	}
	series = append(series, ratePoint{date: date, value: value})
	sort.Slice(series, func(i, j int) bool {
		return series[i].date.Before(series[j].date)
	})
	s.rates[from][to] = series
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

// Lookup returns the rate from -> to on or before `asOf`. If no exact direct
// pair is found it tries inversion, then pivots through USD. Returns
// (rate, true) on success, (Zero, false) otherwise.
func (s *RateStore) Lookup(from, to string, asOf time.Time) (decimal.Decimal, bool) {
	if from == to {
		return decimal.NewFromInt(1), true
	}
	if r, ok := s.directLookup(from, to, asOf); ok {
		return r, true
	}
	// Inversion: 1 / (to -> from)
	if r, ok := s.directLookup(to, from, asOf); ok && !r.IsZero() {
		return decimal.NewFromInt(1).Div(r), true
	}
	// USD pivot.
	const pivot = "USD"
	if from != pivot && to != pivot {
		left, leftOK := s.Lookup(from, pivot, asOf)
		right, rightOK := s.Lookup(pivot, to, asOf)
		if leftOK && rightOK {
			return left.Mul(right), true
		}
	}
	return decimal.Zero, false
}

func (s *RateStore) directLookup(from, to string, asOf time.Time) (decimal.Decimal, bool) {
	v, ok, _ := s.directLookupWithStale(from, to, asOf)
	return v, ok
}

// directLookupWithStale returns the direct rate (if any) plus a `stale` flag
// indicating that we had to extrapolate backwards (asOf is earlier than the
// first known data point). Callers that care can warn; today both callers
// just use the value because the alternative — erroring an entire timeline
// because the user's first transaction predates frankfurter's 1999-01-04
// epoch — would be worse.
func (s *RateStore) directLookupWithStale(from, to string, asOf time.Time) (decimal.Decimal, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.rates[from] == nil {
		return decimal.Zero, false, false
	}
	series := s.rates[from][to]
	if len(series) == 0 {
		return decimal.Zero, false, false
	}
	idx := sort.Search(len(series), func(i int) bool {
		return series[i].date.After(asOf)
	})
	if idx == 0 {
		// No rate is on or before asOf; extrapolate backwards with the
		// earliest known rate (stale).
		return series[0].value, true, true
	}
	return series[idx-1].value, true, false
}

// ConvertToBase converts amount denominated in `from` to `base` as of
// `asOfDate`. Returns the converted amount or an error if no rate can be
// resolved (directly, via inversion, or via USD pivot).
func (s *RateStore) ConvertToBase(amount decimal.Decimal, from, base string, asOfDate time.Time) (decimal.Decimal, error) {
	if from == base {
		return amount, nil
	}
	rate, ok := s.Lookup(from, base, asOfDate)
	if !ok {
		return decimal.Zero, fmt.Errorf("fx: no rate available for %s->%s as of %s", from, base, asOfDate.Format("2006-01-02"))
	}
	return amount.Mul(rate), nil
}

// DatedAmount pairs a date with a non-base amount, used by HistoricalConvert.
type DatedAmount struct {
	Date   time.Time
	Amount decimal.Decimal
}

// HistoricalConvert converts a series of dated amounts in `from` to `base`,
// using the rate applicable on each individual date.
func (s *RateStore) HistoricalConvert(timeline []DatedAmount, from, base string) ([]DatedAmount, error) {
	out := make([]DatedAmount, 0, len(timeline))
	for _, p := range timeline {
		v, err := s.ConvertToBase(p.Amount, from, base, p.Date)
		if err != nil {
			return nil, err
		}
		out = append(out, DatedAmount{Date: p.Date, Amount: v})
	}
	return out, nil
}

// IsKnownCurrency reports whether `c` looks like an ISO-4217-style currency
// code: exactly 3 uppercase ASCII letters. This is intentionally permissive —
// we don't want to maintain a hard-coded whitelist that future scrapers would
// have to update. The combination "looks like a currency" + "has a rate in
// the store" is what `RateStore.Lookup` ultimately enforces. Used to gate
// FX-attribution branches when iterating ledger postings whose `Commodity`
// field may be a ticker (e.g. "AAPL"), an empty string, or a real currency
// like "USD"/"HKD".
func IsKnownCurrency(c string) bool {
	if len(c) != 3 {
		return false
	}
	for i := 0; i < len(c); i++ {
		ch := c[i]
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}

// processStore is the singleton FX rate store consulted by server handlers.
// It's populated lazily from the prices table via the seeder hook, and
// cleared via ClearStore() on every sync (see internal/cache.Clear).
var processStore *RateStore
var processStoreMu sync.Mutex

// Store returns the process-wide singleton, creating it if necessary.
func Store() *RateStore {
	processStoreMu.Lock()
	defer processStoreMu.Unlock()
	if processStore == nil {
		processStore = NewRateStore()
	}
	return processStore
}

// ClearStore drops the singleton so the next caller rebuilds from the DB.
func ClearStore() {
	processStoreMu.Lock()
	defer processStoreMu.Unlock()
	processStore = nil
}
