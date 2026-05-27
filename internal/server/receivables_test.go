package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/account"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// makeReceivablePosting builds a posting that simulates lending money out
// of an Assets:* account: the receivable account itself gets a positive
// posting equal to the amount lent. MarketAmount is pre-populated so the
// handler does not need a DB-backed price lookup in tests.
func makeReceivablePosting(accountName string, marketAmount int64, date time.Time) posting.Posting {
	return posting.Posting{
		Account:      accountName,
		Amount:       decimal.NewFromInt(marketAmount),
		MarketAmount: decimal.NewFromInt(marketAmount),
		Quantity:     decimal.NewFromInt(marketAmount),
		Commodity:    "CNY",
		Date:         date,
	}
}

// TestComputeReceivables_BothWithAndWithoutMetadata is the core M2-G test.
//
// Two accounts have `kind: receivable`:
//   - Assets:YangLiu:LiQuanRong has a matching receivables[] entry in
//     config (borrower / dates / interest rate populated).
//   - Assets:YangLiu:LiuChunliang has NO matching receivables[] entry;
//     the handler must still surface it with the account leaf name as
//     borrower and the dates left empty.
func TestComputeReceivables_BothWithAndWithoutMetadata(t *testing.T) {
	yaml := []byte(
		"journal_path: /tmp/x.ledger\n" +
			"db_path: /tmp/x.db\n" +
			"default_currency: CNY\n" +
			"accounts:\n" +
			"  - name: Assets:YangLiu:LiQuanRong\n" +
			"    kind: receivable\n" +
			"  - name: Assets:YangLiu:LiuChunliang\n" +
			"    kind: receivable\n" +
			"receivables:\n" +
			"  - name: Assets:YangLiu:LiQuanRong\n" +
			"    borrower: 李泉荣\n" +
			"    lend_date: \"2024-06-15\"\n" +
			"    due_date: \"2025-06-15\"\n" +
			"    interest_rate: 0\n" +
			"    note: car purchase\n",
	)
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
	defer func() {
		_ = config.LoadConfig([]byte("journal_path: /tmp/x.ledger\ndb_path: /tmp/x.db\n"), "")
	}()

	postings := []posting.Posting{
		makeReceivablePosting("Assets:YangLiu:LiQuanRong", 50_000, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
		makeReceivablePosting("Assets:YangLiu:LiuChunliang", 12_000, time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)),
		// Unrelated bank account that must not appear.
		makeReceivablePosting("Assets:Saving:CMB", 1000, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	rs := computeReceivables(postings, config.GetConfig().Receivables, toAccountLookup(config.GetConfig().Accounts))

	assert.Len(t, rs, 2, "expected two receivables; got %d", len(rs))

	byAccount := make(map[string]Receivable, len(rs))
	for _, r := range rs {
		byAccount[r.Account] = r
	}

	li, ok := byAccount["Assets:YangLiu:LiQuanRong"]
	assert.True(t, ok, "李泉荣 receivable missing")
	assert.Equal(t, "李泉荣", li.Borrower)
	assert.True(t, li.Outstanding.Equal(decimal.NewFromInt(50_000)),
		"李泉荣 outstanding: got %s want 50000", li.Outstanding.String())
	assert.NotNil(t, li.LendDate, "李泉荣 lend_date must be set")
	assert.Equal(t, 2024, li.LendDate.Year())
	assert.NotNil(t, li.DueDate, "李泉荣 due_date must be set")
	assert.Equal(t, 2025, li.DueDate.Year())
	assert.True(t, li.InterestRate.Equal(decimal.Zero))
	assert.Equal(t, "car purchase", li.Note)
	assert.Equal(t, account.Receivable, li.Kind)

	liu, ok := byAccount["Assets:YangLiu:LiuChunliang"]
	assert.True(t, ok, "LiuChunliang receivable missing")
	// With no config entry, borrower falls back to the leaf account name.
	assert.Equal(t, "LiuChunliang", liu.Borrower)
	assert.True(t, liu.Outstanding.Equal(decimal.NewFromInt(12_000)))
	// Dates must be nil so the frontend can render an empty cell rather
	// than "Invalid Date" / "1970/01/01".
	assert.Nil(t, liu.LendDate, "LiuChunliang lend_date must be nil")
	assert.Nil(t, liu.DueDate, "LiuChunliang due_date must be nil")
	assert.True(t, liu.InterestRate.Equal(decimal.Zero))
	assert.Equal(t, "", liu.Note)
}

// TestComputeReceivables_SortedByOutstandingDesc verifies the deterministic
// ordering documented in the issue (default sort = outstanding desc).
func TestComputeReceivables_SortedByOutstandingDesc(t *testing.T) {
	yaml := []byte(
		"journal_path: /tmp/x.ledger\n" +
			"db_path: /tmp/x.db\n" +
			"default_currency: CNY\n" +
			"accounts:\n" +
			"  - name: Assets:Loans:Small\n" +
			"    kind: receivable\n" +
			"  - name: Assets:Loans:Big\n" +
			"    kind: receivable\n" +
			"  - name: Assets:Loans:Mid\n" +
			"    kind: receivable\n",
	)
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
	defer func() {
		_ = config.LoadConfig([]byte("journal_path: /tmp/x.ledger\ndb_path: /tmp/x.db\n"), "")
	}()

	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makeReceivablePosting("Assets:Loans:Small", 100, d),
		makeReceivablePosting("Assets:Loans:Big", 100_000, d),
		makeReceivablePosting("Assets:Loans:Mid", 5_000, d),
	}

	rs := computeReceivables(postings, config.GetConfig().Receivables, toAccountLookup(config.GetConfig().Accounts))

	assert.Len(t, rs, 3)
	assert.Equal(t, "Assets:Loans:Big", rs[0].Account)
	assert.Equal(t, "Assets:Loans:Mid", rs[1].Account)
	assert.Equal(t, "Assets:Loans:Small", rs[2].Account)
}

// TestComputeReceivables_ZeroBalanceFiltered guards that fully-repaid
// receivables (sum of postings == 0) are dropped from the page. They
// don't represent an outstanding loan anymore.
func TestComputeReceivables_ZeroBalanceFiltered(t *testing.T) {
	yaml := []byte(
		"journal_path: /tmp/x.ledger\n" +
			"db_path: /tmp/x.db\n" +
			"default_currency: CNY\n" +
			"accounts:\n" +
			"  - name: Assets:Loans:Paid\n" +
			"    kind: receivable\n",
	)
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
	defer func() {
		_ = config.LoadConfig([]byte("journal_path: /tmp/x.ledger\ndb_path: /tmp/x.db\n"), "")
	}()

	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makeReceivablePosting("Assets:Loans:Paid", 5_000, d),
		makeReceivablePosting("Assets:Loans:Paid", -5_000, d.AddDate(0, 6, 0)),
	}

	rs := computeReceivables(postings, config.GetConfig().Receivables, toAccountLookup(config.GetConfig().Accounts))
	assert.Empty(t, rs, "fully-repaid receivable must not appear")
}

// TestComputeReceivables_PrefixFallbackKind exercises the path-prefix
// fallback in account.GetKind: an account under Assets:Loans:* without an
// explicit `accounts[].kind` still resolves to Receivable and must appear.
func TestComputeReceivables_PrefixFallbackKind(t *testing.T) {
	yaml := []byte(
		"journal_path: /tmp/x.ledger\n" +
			"db_path: /tmp/x.db\n" +
			"default_currency: CNY\n",
	)
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
	defer func() {
		_ = config.LoadConfig([]byte("journal_path: /tmp/x.ledger\ndb_path: /tmp/x.db\n"), "")
	}()

	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makeReceivablePosting("Assets:Loans:Alice", 3_000, d),
	}

	rs := computeReceivables(postings, config.GetConfig().Receivables, toAccountLookup(config.GetConfig().Accounts))
	assert.Len(t, rs, 1)
	assert.Equal(t, "Assets:Loans:Alice", rs[0].Account)
	assert.Equal(t, "Alice", rs[0].Borrower)
	assert.True(t, rs[0].Outstanding.Equal(decimal.NewFromInt(3_000)))
}
