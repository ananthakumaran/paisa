package importer_test

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/importer/stub"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// resetRegistry uses the in-package export_test.go helper to drop any
// importers registered by previous test cases. The framework registry is
// process-global by design, so tests that need a clean slate must clear it.
func resetRegistry(t *testing.T) {
	t.Helper()
	importer.ResetForTesting()
}

var _ = resetRegistry // referenced by every test below

// TestStubRegistersAndIsDiscoverable confirms the wiring contract: after
// calling stub.Register(), the stub appears in importer.All() and can be
// looked up by its Code(). This is what every real importer (M3-B…E) will
// rely on at init() time.
func TestStubRegistersAndIsDiscoverable(t *testing.T) {
	resetRegistry(t)
	stub.Register()

	all := importer.All()
	assert.Len(t, all, 1, "expected stub to be the only registered importer")
	assert.Equal(t, stub.Code, all[0].Code())
	assert.NotEmpty(t, all[0].Name(), "Name() must be non-empty for the UI")

	found := importer.ByCode(stub.Code)
	if assert.NotNil(t, found, "ByCode must return the registered stub") {
		assert.Equal(t, stub.Code, found.Code())
	}
}

// TestRegisterIsIdempotent guards against double-registration of the same
// importer (e.g. when test setup runs twice). The second call must not add
// a duplicate or panic.
func TestRegisterIsIdempotent(t *testing.T) {
	resetRegistry(t)
	stub.Register()
	stub.Register()
	assert.Len(t, importer.All(), 1)
}

// TestDetectByHeader exercises the content-based detection branch: a CSV
// whose first line is exactly the stub's header line must match, regardless
// of the filename.
func TestDetectByHeader(t *testing.T) {
	resetRegistry(t)
	stub.Register()

	csv := []byte("date,payee,amount\n2024-01-02,Coffee,4.50\n")
	matches := importer.Detect("anything.csv", csv)
	if assert.Len(t, matches, 1) {
		assert.Equal(t, stub.Code, matches[0].Code())
	}
}

// TestDetectByFilename exercises the filename hint branch: a file with the
// magic suffix `.stub.csv` should match even if the content header is
// missing. False positives are acceptable for stubs; specific importers
// will tighten their own heuristics.
func TestDetectByFilename(t *testing.T) {
	resetRegistry(t)
	stub.Register()

	matches := importer.Detect("transactions.stub.csv", []byte("garbage"))
	if assert.Len(t, matches, 1) {
		assert.Equal(t, stub.Code, matches[0].Code())
	}
}

// TestDetectNoMatch confirms Detect returns an empty slice (not nil-panic)
// when nothing matches. The HTTP handler relies on this.
func TestDetectNoMatch(t *testing.T) {
	resetRegistry(t)
	stub.Register()

	matches := importer.Detect("bank.csv", []byte("col1,col2\n1,2\n"))
	assert.Empty(t, matches)
}

// TestParseStubCSV is the heart of the framework contract: an importer's
// Parse() must return a slice of ParsedTxn with date, payee, amount, and
// currency populated. Two-row fixture exercises a typical case.
func TestParseStubCSV(t *testing.T) {
	resetRegistry(t)
	stub.Register()

	content := []byte("date,payee,amount\n2024-01-02,Coffee,4.50\n2024-01-03,Lunch,12.00\n")
	txns, err := importer.ByCode(stub.Code).Parse(content)
	assert.NoError(t, err)
	if assert.Len(t, txns, 2) {
		assert.Equal(t, "Coffee", txns[0].Payee)
		assert.True(t, txns[0].Amount.Equal(decimal.NewFromFloat(4.50)), "amount mismatch: %s", txns[0].Amount)
		assert.Equal(t, 2024, txns[0].Date.Year())
		assert.Equal(t, "INR", txns[0].Currency)
		assert.NotEmpty(t, txns[0].RawText, "RawText must be preserved for round-trip display")

		assert.Equal(t, "Lunch", txns[1].Payee)
		assert.True(t, txns[1].Amount.Equal(decimal.NewFromFloat(12.00)))
	}
}

// TestParseStubCSVInvalidHeader makes sure a wrong header produces a useful
// error rather than a nil slice or a panic.
func TestParseStubCSVInvalidHeader(t *testing.T) {
	resetRegistry(t)
	stub.Register()
	_, err := importer.ByCode(stub.Code).Parse([]byte("a,b,c\n"))
	assert.Error(t, err)
}
