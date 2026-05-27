// Package stub provides a minimal CSV importer used ONLY by the framework
// tests in internal/importer and internal/server. It deliberately does NOT
// register itself via init() — tests must call Register() explicitly so the
// stub never leaks into production builds. The format is the fixed three-
// column CSV `date,payee,amount` with a literal header row of the same
// names.
package stub

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// Code is exported so tests can assert against a single constant rather than
// hard-coding the string.
const Code = "stub-csv"

type Stub struct{}

func (Stub) Code() string { return Code }
func (Stub) Name() string { return "Stub CSV (test only)" }

// Detect matches files whose first line is exactly "date,payee,amount" OR
// whose filename ends in ".stub.csv". The double match lets the http
// detection test exercise both code paths.
func (Stub) Detect(filename string, content []byte) bool {
	if strings.HasSuffix(strings.ToLower(filename), ".stub.csv") {
		return true
	}
	firstLine := content
	if idx := bytes.IndexByte(content, '\n'); idx >= 0 {
		firstLine = content[:idx]
	}
	return strings.TrimSpace(string(firstLine)) == "date,payee,amount"
}

func (Stub) Parse(content []byte) ([]importer.ParsedTxn, error) {
	r := csv.NewReader(bytes.NewReader(content))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, errors.New("stub: empty file")
	}
	header := rows[0]
	if len(header) != 3 || header[0] != "date" || header[1] != "payee" || header[2] != "amount" {
		return nil, errors.New("stub: expected header date,payee,amount")
	}

	var out []importer.ParsedTxn
	for i, row := range rows[1:] {
		if len(row) != 3 {
			return nil, errors.New("stub: malformed row " + itoa(i+2))
		}
		date, err := time.Parse("2006-01-02", strings.TrimSpace(row[0]))
		if err != nil {
			return nil, errors.New("stub: bad date on row " + itoa(i+2) + ": " + err.Error())
		}
		amount, err := decimal.NewFromString(strings.TrimSpace(row[2]))
		if err != nil {
			return nil, errors.New("stub: bad amount on row " + itoa(i+2) + ": " + err.Error())
		}
		out = append(out, importer.ParsedTxn{
			Date:     date,
			Payee:    strings.TrimSpace(row[1]),
			Amount:   amount,
			Currency: "INR",
			RawText:  strings.Join(row, ","),
		})
	}
	return out, nil
}

// Register installs the stub importer into the global registry. Call from a
// test's setup; production binaries must NOT call this function.
func Register() {
	importer.Register(Stub{})
}

// itoa is a tiny strconv.Itoa to avoid the strconv import (keeps the stub
// package's footprint small and easy to audit).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
