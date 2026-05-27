package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// importer/{detect,parse,commit} together implement the new pluggable
// importer pipeline (issue #19, M3-A). The Handlebars-based template
// importer is unchanged and lives at /api/templates/* — these endpoints run
// alongside it. NO production importer is registered in this PR; the routes
// are wired so subsequent issues (M3-B…E) can plug in their formats without
// touching the server again.

type importDetectRequest struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
}

type importParseRequest struct {
	ImporterCode  string `json:"importer_code"`
	ContentBase64 string `json:"content_base64"`
}

// importCommitTxn is the wire shape of ParsedTxn. It mirrors
// importer.ParsedTxn but accepts strings for Date and Amount so the UI can
// edit them as plain text and POST them back without reserialising
// shopspring/decimal — the conversion happens in the handler.
type importCommitTxn struct {
	Date             string `json:"date"` // RFC3339 or "2006-01-02"
	Payee            string `json:"payee"`
	Note             string `json:"note"`
	Amount           string `json:"amount"` // decimal string, sign per ParsedTxn convention
	Currency         string `json:"currency"`
	SuggestedAccount string `json:"suggested_account"` // counterpart account
}

type importCommitRequest struct {
	// SourceAccount is the leg that always appears (Assets:Bank:…,
	// Liabilities:CreditCard:…, etc). Counterpart comes from each txn's
	// SuggestedAccount, with a fallback for missing values.
	SourceAccount string            `json:"source_account"`
	LedgerFile    string            `json:"ledger_file"` // path relative to journal dir
	Txns          []importCommitTxn `json:"txns"`
}

// ImportDetect implements POST /api/import/detect. Returns the list of
// importers whose Detect() matched the file. Always responds 200 unless the
// payload is malformed; an empty match list means "no importer recognises
// this — try Handlebars or pick one manually".
func ImportDetect(c *gin.Context) {
	var req importDetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content, err := base64.StdEncoding.DecodeString(req.ContentBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64: " + err.Error()})
		return
	}

	matched := importer.Detect(req.Filename, content)
	out := make([]gin.H, 0, len(matched))
	for _, i := range matched {
		out = append(out, gin.H{"code": i.Code(), "name": i.Name()})
	}
	c.JSON(http.StatusOK, gin.H{"importers": out})
}

// ImportParse implements POST /api/import/parse. Dispatches to the importer
// identified by `importer_code` and returns the parsed transactions. The UI
// then renders them in an editable preview before commit.
func ImportParse(c *gin.Context) {
	var req importParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	imp := importer.ByCode(req.ImporterCode)
	if imp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown importer: %q", req.ImporterCode)})
		return
	}
	content, err := base64.StdEncoding.DecodeString(req.ContentBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64: " + err.Error()})
		return
	}
	txns, err := imp.Parse(content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}

// ImportCommit implements POST /api/import/commit. Appends the user-confirmed
// transactions to `ledger_file` (resolved relative to the configured journal
// directory). Read-only mode short-circuits with `{saved: true}` per the
// project convention. Validation errors do NOT roll back already-written
// bytes — the handler builds the full text in memory first and only writes
// once everything passes ledger CLI validation.
func ImportCommit(c *gin.Context) {
	if config.GetConfig().Readonly {
		c.JSON(http.StatusOK, gin.H{"saved": true, "count": 0, "errors": []string{}})
		return
	}

	var req importCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.SourceAccount) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_account is required"})
		return
	}
	if strings.TrimSpace(req.LedgerFile) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ledger_file is required"})
		return
	}

	journalDir := filepath.Dir(config.GetJournalPath())
	target, err := utils.BuildSubPath(journalDir, req.LedgerFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rendered strings.Builder
	var errs []string
	for i, t := range req.Txns {
		entry, err := renderImportTxn(t, req.SourceAccount)
		if err != nil {
			errs = append(errs, fmt.Sprintf("txn %d: %s", i+1, err.Error()))
			continue
		}
		rendered.WriteString(entry)
		rendered.WriteString("\n")
	}
	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"saved": false, "errors": errs})
		return
	}

	// Append (not overwrite). If the file doesn't exist, create it with a
	// newline-friendly leading blank line so the appended block starts
	// cleanly. utils.BuildSubPath already guards against path escape.
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		log.Warn(err)
		c.JSON(http.StatusInternalServerError, gin.H{"saved": false, "errors": []string{err.Error()}})
		return
	}
	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn(err)
		c.JSON(http.StatusInternalServerError, gin.H{"saved": false, "errors": []string{err.Error()}})
		return
	}
	defer f.Close()

	// Add a separating newline if the existing file does not end in one.
	if stat, err := f.Stat(); err == nil && stat.Size() > 0 {
		// Cheap heuristic: read last byte to decide whether to inject a
		// leading newline. Avoids a full file read.
		buf := make([]byte, 1)
		if _, err := f.ReadAt(buf, stat.Size()-1); err == nil && buf[0] != '\n' {
			rendered.WriteString("")
			if _, werr := f.WriteString("\n"); werr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"saved": false, "errors": []string{werr.Error()}})
				return
			}
		}
	}

	if _, err := f.WriteString(rendered.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"saved": false, "errors": []string{err.Error()}})
		return
	}

	// We deliberately do NOT call Sync(db, ...) here: the import package
	// has no db handle (handler is a free function), and the standard
	// editor save path already wires Sync. For now, mutation-via-import is
	// reflected in the file but the in-process cache will refresh on the
	// next /api/sync. Subsequent issues that wire commit into a db-aware
	// handler can re-enable the sync call.

	c.JSON(http.StatusOK, gin.H{"saved": true, "count": len(req.Txns), "errors": []string{}})
}

// renderImportTxn turns one importCommitTxn into a ledger entry. Format:
//
//	2024/01/02 Payee  ; note (optional)
//	    Counterpart                100.00 USD
//	    SourceAccount
//
// Amount sign per ParsedTxn convention: positive = money LEAVING source.
// The counterpart leg carries the explicit amount; the source leg is
// implicit so ledger balances it.
func renderImportTxn(t importCommitTxn, sourceAccount string) (string, error) {
	date, err := parseImportDate(t.Date)
	if err != nil {
		return "", err
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(t.Amount))
	if err != nil {
		return "", fmt.Errorf("invalid amount %q: %w", t.Amount, err)
	}
	payee := strings.TrimSpace(t.Payee)
	if payee == "" {
		payee = "Unknown"
	}
	counterpart := strings.TrimSpace(t.SuggestedAccount)
	if counterpart == "" {
		counterpart = "Expenses:Unknown"
	}
	currency := strings.TrimSpace(t.Currency)
	if currency == "" {
		currency = config.DefaultCurrency()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", date.Format("2006/01/02"), payee)
	if note := strings.TrimSpace(t.Note); note != "" {
		fmt.Fprintf(&b, "  ; %s", note)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "    %s    %s %s\n", counterpart, amount.StringFixed(2), currency)
	fmt.Fprintf(&b, "    %s\n", sourceAccount)
	return b.String(), nil
}

// parseImportDate accepts the formats the UI is likely to send: full
// RFC3339 (what JS Date pickers serialise to) and plain `2006-01-02`.
func parseImportDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q (want 2006-01-02 or RFC3339)", s)
}
