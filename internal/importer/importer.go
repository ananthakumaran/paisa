// Package importer defines the plugin framework for bank/payment statement
// importers. Each importer implements [Importer] and is registered into the
// process-wide registry (see registry.go). Specific importers (支付宝, 微信,
// 招商银行, IBKR, …) live in their own subpackages and register themselves via
// an init() function. The framework itself is format-agnostic.
//
// This package is intentionally tiny: it only describes WHAT an importer must
// do. HOW each importer parses its specific format is the importer's concern.
// The HTTP handlers in internal/server/import.go convert detected importers
// and parsed transactions to/from JSON for the preview UI.
package importer

import (
	"time"

	"github.com/shopspring/decimal"
)

// ParsedTxn is a single transaction extracted from a statement file. It is
// the lingua franca between an importer and the preview UI: the UI lets the
// user edit any of these fields before the final ledger commit, so importers
// should populate as many of them as the source data allows.
//
// Amount sign convention: a POSITIVE amount represents money LEAVING the
// source account (expense, transfer out, withdrawal). A NEGATIVE amount
// represents money ENTERING the source account (income, refund, deposit).
// The commit step turns each ParsedTxn into two postings: the source
// account, and SuggestedAccount as the counterpart.
type ParsedTxn struct {
	Date     time.Time       `json:"date"`
	Payee    string          `json:"payee"`
	Note     string          `json:"note"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	// SuggestedAccount is the importer's best guess for the OTHER leg of
	// the transaction (i.e. not the source account). M3-F's TF-IDF
	// predictor may further refine this; leaving it empty signals "unknown,
	// let the UI pick".
	SuggestedAccount string `json:"suggested_account"`
	// RawText is the original record (e.g. one CSV row, or a paragraph of
	// PDF text). Kept for user reference in the preview UI and for
	// round-trip debugging — never written to the ledger.
	RawText string `json:"raw_text"`
}

// Importer is the contract every statement parser must satisfy. Implementations
// must be stateless and safe for concurrent calls; the registry hands the same
// instance to every request.
type Importer interface {
	// Code is a stable, machine-readable identifier (e.g. "alipay",
	// "wechat", "cmb-debit"). Used in HTTP requests and persisted in
	// preferences. MUST NOT change once shipped — clients pin to it.
	Code() string

	// Name is a human-readable label shown in the UI (e.g. "支付宝月账单").
	// Translation is the UI's responsibility; importers return their
	// canonical name.
	Name() string

	// Detect returns true if `filename` and the first chunk of `content`
	// match this importer's expected format. Implementations should be
	// cheap (look at magic bytes / header line / filename suffix) — they
	// run for every uploaded file. False positives are tolerable (the user
	// can override); false negatives hide importers entirely.
	Detect(filename string, content []byte) bool

	// Parse converts the full file into a slice of transactions. Errors
	// from Parse are surfaced to the user verbatim — return descriptive,
	// localizable messages.
	Parse(content []byte) ([]ParsedTxn, error)
}
