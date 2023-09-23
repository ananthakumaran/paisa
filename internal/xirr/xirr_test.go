package xirr

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var XIRR_EXPECTED_VALUES = map[string]float64{
	"unordered.csv":         16.35,
	"single_redemption.csv": 13.62,
	"random.csv":            69.25,
	"random_100.csv":        2982.94,
	"random_1000.csv":       550.89,
	"30-0.csv":              16.60,
	"30-1.csv":              18.18,
	"30-2.csv":              -0.27,
	"30-3.csv":              585.25,
	"30-4.csv":              16.10,
	"30-5.csv":              0.90,
	"30-6.csv":              32.55,
	"30-7.csv":              35.01,
	"30-8.csv":              335.30,
	"30-9.csv":              287.80,
	"30-10.csv":             11.14,
	"30-11.csv":             151.22,
	"30-12.csv":             -2.58,
	"30-13.csv":             -65.91,
	"30-14.csv":             69.97,
	"30-15.csv":             2.98,
	"30-16.csv":             44.20,
	"30-17.csv":             279.56,
	"30-18.csv":             25.94,
	"30-19.csv":             -0.16,
	"30-20.csv":             0.00,
	"30-21.csv":             5.90,
	"30-22.csv":             -31.54,
	"30-23.csv":             112.77,
	"30-24.csv":             3290.89,
	"30-25.csv":             -0.12,
	"30-26.csv":             -33.23,
	"30-27.csv":             0.02,
	"30-28.csv":             112.58,
	"30-29.csv":             0.00,
	"30-30.csv":             8.12,
	"30-31.csv":             0.00,
	"30-32.csv":             -86.02,
	"30-33.csv":             0.00,
	"30-34.csv":             0.00,
	"30-35.csv":             -89.57,
	"30-36.csv":             -9.61,
	"30-37.csv":             -65.19,
	"30-38.csv":             0.00,
	"30-39.csv":             -20.27,
	"30-40.csv":             0.00,
	"30-41.csv":             -11.64,
	"30-42.csv":             0.00,
	"30-43.csv":             -12.84,
	"30-44.csv":             0.00,
	"30-45.csv":             -84.08,
	"30-46.csv":             -4.74,
	"30-47.csv":             -61.03,
	"30-48.csv":             -7.53,
	// "minus_0_99.csv":        -99.90,
	// "minus_0_99999.csv":     -99.99,
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func readCSV(filename string) []Cashflow {
	f := lo.Must(os.Open(filename))
	defer f.Close()

	csvReader := csv.NewReader(f)
	records := lo.Must(csvReader.ReadAll())
	var cashflows = []Cashflow{}
	for _, record := range records {
		cashflow := Cashflow{}
		cashflow.Date = lo.Must(time.Parse("2006-01-02", record[0]))
		cashflow.Amount = lo.Must(decimal.NewFromString(strings.Trim(record[1], " "))).InexactFloat64()
		cashflows = append(cashflows, cashflow)
	}
	return cashflows
}

func TestXIRR(t *testing.T) {
	// Example from Microsoft Excel documentation:
	cashflows := []Cashflow{
		{Date: date(2008, 01, 01), Amount: -10000.00},
		{Date: date(2008, 03, 01), Amount: 2750.00},
		{Date: date(2008, 10, 30), Amount: 4250.00},
		{Date: date(2009, 02, 15), Amount: 3250.00},
		{Date: date(2009, 04, 01), Amount: 2750.00},
	}
	assert.Equal(t, decimal.NewFromFloat(37.34), XIRR(cashflows))

	_, filename, _, _ := runtime.Caller(0)
	dirname := filepath.Join(filepath.Dir(filename), "samples")

	t.Logf("dirname : %s", dirname)
	files := lo.Must(os.ReadDir(dirname))

	for _, f := range files {
		if value, exist := XIRR_EXPECTED_VALUES[f.Name()]; exist {
			cashflows := readCSV(filepath.Join(dirname, f.Name()))
			expected := decimal.NewFromFloat(value).Round(2)
			if !expected.Equal(XIRR(cashflows)) {
				t.Logf("XIRR(%s) : %s %s", f.Name(), XIRR(cashflows), expected)
				break
			}
			assert.Equal(t, expected, XIRR(cashflows))
		}
	}

}
