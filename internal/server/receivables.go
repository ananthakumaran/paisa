package server

import (
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/account"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Receivable is the API row for /api/receivables.
//
// One row per Asset account classified as `kind: receivable` (M1-D) with
// a non-zero outstanding balance. Metadata fields (borrower / dates /
// interest rate / note) are joined in from config.Receivables when an
// entry with a matching `Name` exists; otherwise reasonable defaults are
// used (leaf account name as borrower, dates left as nil).
//
// Dates are emitted as *time.Time so the JSON serializer can produce
// `null` when not configured — the frontend renders an empty cell rather
// than 1970-01-01.
type Receivable struct {
	Account      string              `json:"account"`
	Borrower     string              `json:"borrower"`
	Outstanding  decimal.Decimal     `json:"outstanding"`
	LendDate     *time.Time          `json:"lend_date"`
	DueDate      *time.Time          `json:"due_date"`
	InterestRate decimal.Decimal     `json:"interest_rate"`
	Note         string              `json:"note"`
	Kind         account.AccountKind `json:"kind"`
}

// GetReceivables is the /api/receivables handler. Fetches every Assets:*
// posting, joins each receivable-kind account with its config metadata,
// and returns the sorted list plus a precomputed total for the summary
// card.
func GetReceivables(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").All()
	postings = service.PopulateMarketPrice(db, postings)

	cfg := config.GetConfig()
	rs := computeReceivables(postings, cfg.Receivables, toAccountLookup(cfg.Accounts))

	total := decimal.Zero
	for _, r := range rs {
		total = total.Add(r.Outstanding)
	}

	// Always emit a non-nil slice so the JSON response shape is stable
	// (`[]` instead of `null`) which simplifies frontend consumption.
	if rs == nil {
		rs = []Receivable{}
	}

	return gin.H{
		"receivables":       rs,
		"total_outstanding": total,
	}
}

// computeReceivables is the testable core that aggregates postings by
// account, keeps only those whose AccountKind == Receivable, and joins
// each with its (optional) per-loan config metadata.
//
// Returned slice is sorted by outstanding amount in descending order so
// the frontend's default view shows the largest balances first.
func computeReceivables(postings []posting.Posting, configReceivables []config.Receivable, configAccounts []account.Account) []Receivable {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })

	metaByName := make(map[string]config.Receivable, len(configReceivables))
	for _, r := range configReceivables {
		metaByName[r.Name] = r
	}

	out := make([]Receivable, 0)
	for accountName, ps := range byAccount {
		kind := account.GetKind(accountName, configAccounts)
		if kind != account.Receivable {
			continue
		}

		outstanding := accounting.CurrentBalance(ps)
		if outstanding.IsZero() {
			// Fully-repaid: don't surface a row.
			continue
		}

		row := Receivable{
			Account:     accountName,
			Outstanding: outstanding,
			Kind:        kind,
		}

		if meta, ok := metaByName[accountName]; ok {
			row.Borrower = strings.TrimSpace(meta.Borrower)
			row.LendDate = parseOptionalDate(meta.LendDate)
			row.DueDate = parseOptionalDate(meta.DueDate)
			row.InterestRate = decimal.NewFromFloat(meta.InterestRate)
			row.Note = meta.Note
		}

		if row.Borrower == "" {
			row.Borrower = leafName(accountName)
		}

		out = append(out, row)
	}

	sort.SliceStable(out, func(i, j int) bool {
		// Outstanding desc; on equal balances fall back to account name
		// ascending for a deterministic order in tests.
		if !out[i].Outstanding.Equal(out[j].Outstanding) {
			return out[i].Outstanding.GreaterThan(out[j].Outstanding)
		}
		return out[i].Account < out[j].Account
	})
	return out
}

// parseOptionalDate accepts a YYYY-MM-DD string and returns *time.Time, or
// nil for empty / unparseable input. We intentionally return nil rather
// than the zero Time so the JSON output is `null` — the frontend uses
// that to render a blank cell instead of "Invalid Date".
func parseOptionalDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// leafName returns the final colon-separated segment of a ledger account
// name. Used as the fallback borrower when no config entry exists for an
// account.
func leafName(accountName string) string {
	idx := strings.LastIndex(accountName, ":")
	if idx == -1 {
		return accountName
	}
	return accountName[idx+1:]
}
