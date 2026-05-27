package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/importer/stub"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// helper: build a gin engine with just the three import routes wired against
// a fresh registry that has the stub importer registered. We avoid spinning
// the full server.Build(db,...) because that needs a config + sqlite. The
// import endpoints are stateless w.r.t. db, so the slim router faithfully
// exercises the handler contract.
func newImportTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	importer.ResetForTesting()
	stub.Register()

	r := gin.New()
	r.POST("/api/import/detect", ImportDetect)
	// nil db: the parse handler's learning-overlay step (M3-F, #24) is
	// skipped, which is what we want — these unit tests verify the
	// dispatch + JSON contract without a real sqlite in the loop.
	r.POST("/api/import/parse", ImportParse(nil))
	// nil db: the auto-sync branch is skipped, which is what we want — these
	// unit tests verify the file-write contract without a real sqlite +
	// ledger CLI in the loop. Integration coverage lives in tests/.
	r.POST("/api/import/commit", ImportCommit(nil))
	return r
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func doJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// TestImportDetectReturnsStubCode posts a stub-compatible CSV and expects
// the response to list the stub's code in `importers`.
func TestImportDetectReturnsStubCode(t *testing.T) {
	r := newImportTestRouter(t)

	body := map[string]string{
		"filename":       "any.csv",
		"content_base64": b64("date,payee,amount\n2024-01-02,Coffee,4.50\n"),
	}
	w := doJSON(r, "/api/import/detect", body)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Importers []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"importers"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Importers, 1) {
		assert.Equal(t, stub.Code, resp.Importers[0].Code)
		assert.NotEmpty(t, resp.Importers[0].Name)
	}
}

// TestImportDetectNoMatch confirms an empty (but non-null) array is returned
// when nothing matches. The frontend uses `length === 0` to drive the empty
// state, so `null` would break it.
func TestImportDetectNoMatch(t *testing.T) {
	r := newImportTestRouter(t)
	body := map[string]string{
		"filename":       "unknown.csv",
		"content_base64": b64("col1,col2\n1,2\n"),
	}
	w := doJSON(r, "/api/import/detect", body)
	assert.Equal(t, 200, w.Code)

	// Must contain the JSON `"importers": []` token explicitly, not `null`.
	assert.Contains(t, w.Body.String(), `"importers":[]`)
}

// TestImportParseReturnsTransactions exercises the dispatch-by-code path:
// post `importer_code: "stub-csv"` plus a base64 body, get back the parsed
// transactions as JSON.
func TestImportParseReturnsTransactions(t *testing.T) {
	r := newImportTestRouter(t)

	body := map[string]string{
		"importer_code":  stub.Code,
		"content_base64": b64("date,payee,amount\n2024-01-02,Coffee,4.50\n2024-01-03,Lunch,12.00\n"),
	}
	w := doJSON(r, "/api/import/parse", body)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Transactions []struct {
			Payee    string `json:"payee"`
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"transactions"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Transactions, 2) {
		assert.Equal(t, "Coffee", resp.Transactions[0].Payee)
		assert.Equal(t, "INR", resp.Transactions[0].Currency)
		// shopspring/decimal marshals as a JSON number (not quoted) thanks
		// to MarshalJSONWithoutQuotes. We only check it parses; exact
		// representation may vary ("4.5" vs "4.50").
		assert.NotEmpty(t, resp.Transactions[0].Amount)
	}
}

// TestImportParseUnknownCode confirms a 400 + error when the importer code
// is not registered. Catches typos in the frontend.
func TestImportParseUnknownCode(t *testing.T) {
	r := newImportTestRouter(t)
	body := map[string]string{
		"importer_code":  "no-such-importer",
		"content_base64": b64("anything"),
	}
	w := doJSON(r, "/api/import/parse", body)
	assert.Equal(t, 400, w.Code)
}

// loadConfigInDir loads a minimal paisa config with an absolute journal_path
// pointing at `dir/main.ledger`. We use an absolute path because the global
// `configPath` is set once at process start (LoadConfig only assigns it if
// empty) so we can't rely on GetConfigDir to point inside a t.TempDir.
func loadConfigInDir(t *testing.T, dir string, readonly bool) string {
	t.Helper()
	journalPath := filepath.Join(dir, "main.ledger")
	dbPath := filepath.Join(dir, "paisa.db")
	yaml := "journal_path: " + journalPath + "\ndb_path: " + dbPath + "\ndefault_currency: USD\n"
	if readonly {
		yaml += "readonly: true\n"
	}
	assert.NoError(t, config.LoadConfig([]byte(yaml), ""))
	return journalPath
}

// TestImportCommitAppendsToFile drives the full commit pipeline against a
// fresh temp journal: a tmp dir replaces the config's journal location, the
// handler appends two postings per ParsedTxn, and the file content matches.
func TestImportCommitAppendsToFile(t *testing.T) {
	r := newImportTestRouter(t)

	tmpDir := t.TempDir()
	journalPath := loadConfigInDir(t, tmpDir, false)
	assert.NoError(t, os.WriteFile(journalPath, []byte(""), 0644))

	body := map[string]any{
		"source_account": "Assets:Bank:Chase",
		"ledger_file":    "main.ledger",
		"txns": []map[string]any{
			{
				"date":              "2024-01-02T00:00:00Z",
				"payee":             "Coffee",
				"amount":            "4.50",
				"currency":          "USD",
				"suggested_account": "Expenses:Food",
				"note":              "morning cup",
			},
			{
				"date":              "2024-01-03T00:00:00Z",
				"payee":             "Lunch",
				"amount":            "12.00",
				"currency":          "USD",
				"suggested_account": "Expenses:Food",
			},
		},
	}
	w := doJSON(r, "/api/import/commit", body)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Saved  bool     `json:"saved"`
		Count  int      `json:"count"`
		Errors []string `json:"errors"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Saved, "expected saved=true, errors=%v", resp.Errors)
	assert.Equal(t, 2, resp.Count)

	written, err := os.ReadFile(journalPath)
	assert.NoError(t, err)
	content := string(written)
	assert.Contains(t, content, "2024/01/02 Coffee")
	assert.Contains(t, content, "Assets:Bank:Chase")
	assert.Contains(t, content, "Expenses:Food")
	assert.Contains(t, content, "2024/01/03 Lunch")
}

// TestImportCommitReadonly verifies the readonly short-circuit matches the
// convention used by every other mutating endpoint: returns success: true
// without touching disk.
func TestImportCommitReadonly(t *testing.T) {
	r := newImportTestRouter(t)

	tmpDir := t.TempDir()
	journalPath := loadConfigInDir(t, tmpDir, true)
	assert.NoError(t, os.WriteFile(journalPath, []byte("untouched"), 0644))

	body := map[string]any{
		"source_account": "Assets:Bank",
		"ledger_file":    "main.ledger",
		"txns":           []map[string]any{},
	}
	w := doJSON(r, "/api/import/commit", body)
	assert.Equal(t, 200, w.Code)

	written, _ := os.ReadFile(journalPath)
	assert.Equal(t, "untouched", string(written), "readonly mode must not modify the journal")
}

// TestImportDetectBadBase64 — defensive: a malformed base64 body must yield
// a 400, not a 500.
func TestImportDetectBadBase64(t *testing.T) {
	r := newImportTestRouter(t)
	body := map[string]string{
		"filename":       "x.csv",
		"content_base64": "!!!not base64!!!",
	}
	w := doJSON(r, "/api/import/detect", body)
	assert.Equal(t, 400, w.Code)
}

// TestImportParseAppliesLearningOverlay drives the M3-F (#24) Layer-1
// behaviour end-to-end through the HTTP handler. We seed the
// account_learning table with a (payee → account) row, then ask the parse
// handler to re-parse a stub CSV containing the same payee. The handler
// should overlay the learned mapping on top of whatever the importer
// suggested.
func TestImportParseAppliesLearningOverlay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	importer.ResetForTesting()
	stub.Register()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, prediction.AutoMigrateLearning(db))
	// Teach the system that "Coffee" goes to a custom account. The stub
	// importer leaves SuggestedAccount blank, so without the overlay the
	// response would carry an empty string.
	assert.NoError(t, prediction.RecordUserChoice(db, "Coffee", "Expenses:Coffee:Daily"))

	r := gin.New()
	r.POST("/api/import/parse", ImportParse(db))

	body := map[string]string{
		"importer_code":  stub.Code,
		"content_base64": b64("date,payee,amount\n2024-01-02,Coffee,4.50\n"),
	}
	w := doJSON(r, "/api/import/parse", body)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Transactions []struct {
			Payee            string `json:"payee"`
			SuggestedAccount string `json:"suggested_account"`
		} `json:"transactions"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Transactions, 1) {
		assert.Equal(t, "Coffee", resp.Transactions[0].Payee)
		assert.Equal(t, "Expenses:Coffee:Daily", resp.Transactions[0].SuggestedAccount,
			"learning overlay must replace the importer's suggestion")
	}
}

// TestImportCommitRecordsLearning drives the M3-F (#24) Layer-2 behaviour
// end-to-end: committing a transaction with a non-empty SuggestedAccount
// must persist a (payee → account) row in account_learning so the NEXT
// parse for the same payee gets the same suggestion.
func TestImportCommitRecordsLearning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	importer.ResetForTesting()
	stub.Register()

	tmpDir := t.TempDir()
	journalPath := loadConfigInDir(t, tmpDir, false)
	assert.NoError(t, os.WriteFile(journalPath, []byte(""), 0644))

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, prediction.AutoMigrateLearning(db))

	r := gin.New()
	r.POST("/api/import/commit", ImportCommit(db))

	body := map[string]any{
		"source_account": "Assets:Bank:Chase",
		"ledger_file":    "main.ledger",
		"txns": []map[string]any{
			{
				"date":              "2024-01-02T00:00:00Z",
				"payee":             "星巴克咖啡",
				"amount":            "38.00",
				"currency":          "CNY",
				"suggested_account": "Expenses:Coffee:Starbucks",
			},
		},
	}
	w := doJSON(r, "/api/import/commit", body)
	// Without a real ledger CLI / posting model the auto-sync branch
	// surfaces a warning but the write itself still succeeds — we only
	// care that the learning row landed.
	assert.Contains(t, []int{200, 500}, w.Code, "body: %s", w.Body.String())

	got := prediction.LookupLearned(db, "星巴克咖啡")
	assert.Equal(t, "Expenses:Coffee:Starbucks", got,
		"commit must persist (payee → account) into account_learning")
}
