package accounting

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func makePosting(txID, account string, amount int64) posting.Posting {
	return posting.Posting{
		TransactionID: txID,
		Account:       account,
		Amount:        decimal.NewFromInt(amount),
		Quantity:      decimal.NewFromInt(amount),
		Commodity:     "INR",
	}
}

func TestIsInternalTransfer_NoGlobs(t *testing.T) {
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:IBKR", 1000),
		makePosting("tx1", "Assets:Saving:CMB", -1000),
	}
	assert.False(t, IsInternalTransfer(ps, []string{}))
	assert.False(t, IsInternalTransfer(ps, nil))
}

func TestIsInternalTransfer_BothAssetsAndMatch(t *testing.T) {
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:IBKR", 1000),
		makePosting("tx1", "Assets:Saving:CMB", -1000),
	}
	assert.True(t, IsInternalTransfer(ps, []string{"Assets:Bridge:*"}))
}

func TestIsInternalTransfer_MatchOnOtherLeg(t *testing.T) {
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:IBKR", 1000),
		makePosting("tx1", "Assets:Saving:CMB:3173", -1000),
	}
	assert.True(t, IsInternalTransfer(ps, []string{"Assets:Saving:CMB:3173"}))
}

func TestIsInternalTransfer_NoLegMatchesGlob(t *testing.T) {
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Checking:HDFC", 1000),
		makePosting("tx1", "Assets:Saving:CMB", -1000),
	}
	assert.False(t, IsInternalTransfer(ps, []string{"Assets:Bridge:*"}))
}

func TestIsInternalTransfer_OneLegNotAsset(t *testing.T) {
	// salary into bridge — must NOT be considered internal transfer
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:IBKR", 1000),
		makePosting("tx1", "Income:Salary", -1000),
	}
	assert.False(t, IsInternalTransfer(ps, []string{"Assets:Bridge:*"}))
}

func TestIsInternalTransfer_ExpenseOnTransferAccountNotCounted(t *testing.T) {
	// buying from a "bridge" account into expenses — not an internal transfer
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:IBKR", -1000),
		makePosting("tx1", "Expenses:Food", 1000),
	}
	assert.False(t, IsInternalTransfer(ps, []string{"Assets:Bridge:*"}))
}

func TestIsInternalTransfer_StarMatchesNested(t *testing.T) {
	ps := []posting.Posting{
		makePosting("tx1", "Assets:Bridge:Brokerage:DanJuan", 1000),
		makePosting("tx1", "Assets:Saving:CMB", -1000),
	}
	// Because account names use ':' (not '/') as the separator,
	// 'Assets:Bridge:*' matches arbitrarily deep accounts under
	// Assets:Bridge — consistent with paisa's existing FilterByGlob.
	assert.True(t, IsInternalTransfer(ps, []string{"Assets:Bridge:*"}))
}

func TestFilterOutInternalTransfers(t *testing.T) {
	all := []posting.Posting{
		// transfer: both legs assets, one matches glob — excluded
		makePosting("tx1", "Assets:Bridge:IBKR", 1000),
		makePosting("tx1", "Assets:Saving:CMB", -1000),
		// real investment: from Income to Assets — kept
		makePosting("tx2", "Assets:Equity:VOO", 5000),
		makePosting("tx2", "Income:Salary", -5000),
		// transfer to a non-glob asset — kept (no leg matches transfer glob)
		makePosting("tx3", "Assets:Checking:HDFC", 200),
		makePosting("tx3", "Assets:Saving:CMB", -200),
	}

	filtered := FilterOutInternalTransfers(all, []string{"Assets:Bridge:*"})

	// Should remove tx1 (both postings), keep tx2 and tx3
	ids := make(map[string]bool)
	for _, p := range filtered {
		ids[p.TransactionID] = true
	}
	assert.False(t, ids["tx1"], "tx1 should be excluded as internal transfer")
	assert.True(t, ids["tx2"], "tx2 should be kept")
	assert.True(t, ids["tx3"], "tx3 should be kept")
	assert.Equal(t, 4, len(filtered))
}
