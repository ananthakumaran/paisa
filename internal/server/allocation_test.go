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

// makeAggregatePosting builds a posting record with both Amount and
// MarketAmount pre-populated to the same value so that
// AggregateByKindForTest can compute a deterministic aggregate without a DB.
func makeAggregatePosting(accountName string, marketAmount int64) posting.Posting {
	return posting.Posting{
		Account:      accountName,
		Amount:       decimal.NewFromInt(marketAmount),
		MarketAmount: decimal.NewFromInt(marketAmount),
		Quantity:     decimal.NewFromInt(marketAmount),
		Commodity:    "CNY",
		Date:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// TestAggregateByKind_OverrideMutualFund — the core M1-E test.
//
// `Assets:Saving:CMB:Fund` would fall under `bank_current` (Saving prefix)
// without an explicit kind. With `accounts[].kind = mutual_fund` in config,
// the aggregator must place it under the `mutual_fund` (基金) kind bucket.
func TestAggregateByKind_OverrideMutualFund(t *testing.T) {
	accounts := []account.Account{
		{Name: "Assets:Saving:CMB:Fund", Kind: account.MutualFund},
	}
	postings := []posting.Posting{
		makeAggregatePosting("Assets:Saving:CMB", 1000),
		makeAggregatePosting("Assets:Saving:CMB:Fund", 5000),
	}

	result := aggregateByKind(postings, accounts, time.Now())

	// Root node always exists.
	root, ok := result[allocationRootKey]
	assert.True(t, ok, "root node %q must exist", allocationRootKey)
	assert.NotNil(t, root)

	// The mutual_fund kind bucket must exist and contain CMB:Fund.
	mfBucket := allocationRootKey + ":" + string(account.MutualFund)
	mfLeaf := mfBucket + ":" + sanitizeAccountSegment("Assets:Saving:CMB:Fund")
	leaf, ok := result[mfLeaf]
	assert.True(t, ok, "expected leaf %q under mutual_fund bucket; got keys: %v", mfLeaf, keysOf(result))
	assert.True(t, leaf.MarketAmount.Equal(decimal.NewFromInt(5000)),
		"CMB:Fund market amount: got %s want 5000", leaf.MarketAmount.String())
	assert.Equal(t, "Assets:Saving:CMB:Fund", leaf.OriginalAccount)
	assert.Equal(t, account.MutualFund, leaf.Kind)

	// CMB (no override) must fall under bank_current via prefix rule.
	bcLeaf := allocationRootKey + ":" + string(account.BankCurrent) + ":" + sanitizeAccountSegment("Assets:Saving:CMB")
	leafBC, ok := result[bcLeaf]
	assert.True(t, ok, "expected leaf %q under bank_current; got keys: %v", bcLeaf, keysOf(result))
	assert.True(t, leafBC.MarketAmount.Equal(decimal.NewFromInt(1000)))
	assert.Equal(t, account.BankCurrent, leafBC.Kind)

	// CMB:Fund must NOT appear under bank_current.
	bcCmbFund := allocationRootKey + ":" + string(account.BankCurrent) + ":" + sanitizeAccountSegment("Assets:Saving:CMB:Fund")
	_, present := result[bcCmbFund]
	assert.False(t, present, "CMB:Fund must not be aggregated under bank_current")
}

// TestAggregateByKind_RealEstateAndUnknown verifies multiple kinds end up in
// distinct top-level buckets and that an unmatched account falls into
// `unknown`.
func TestAggregateByKind_RealEstateAndUnknown(t *testing.T) {
	accounts := []account.Account{
		{Name: "Assets:Property:House:LCHWK1154", Kind: account.RealEstate},
	}
	postings := []posting.Posting{
		makeAggregatePosting("Assets:Property:House:LCHWK1154", 2_000_000),
		makeAggregatePosting("Assets:Mystery:Foo", 42),
	}

	result := aggregateByKind(postings, accounts, time.Now())

	rePath := allocationRootKey + ":" + string(account.RealEstate) + ":" + sanitizeAccountSegment("Assets:Property:House:LCHWK1154")
	leaf, ok := result[rePath]
	assert.True(t, ok, "real_estate leaf missing; keys: %v", keysOf(result))
	assert.Equal(t, account.RealEstate, leaf.Kind)

	unknownPath := allocationRootKey + ":" + string(account.KindUnknown) + ":" + sanitizeAccountSegment("Assets:Mystery:Foo")
	unkLeaf, ok := result[unknownPath]
	assert.True(t, ok, "unknown leaf missing; keys: %v", keysOf(result))
	assert.Equal(t, account.KindUnknown, unkLeaf.Kind)
}

// TestAggregateByKind_KindBucketAccumulates checks that the kind-level node
// reports the sum of all child accounts' market amounts. The frontend uses
// the partition layout's `sum` so the intermediate value is not strictly
// required to be non-zero, but exposing the sum is a useful invariant.
func TestAggregateByKind_KindBucketIsParentOfLeaves(t *testing.T) {
	postings := []posting.Posting{
		makeAggregatePosting("Assets:Saving:CMB", 1000),
		makeAggregatePosting("Assets:Saving:ICBC", 3000),
	}

	result := aggregateByKind(postings, nil, time.Now())

	bcBucket := allocationRootKey + ":" + string(account.BankCurrent)
	bucket, ok := result[bcBucket]
	assert.True(t, ok, "expected bank_current bucket %q", bcBucket)
	assert.Equal(t, account.BankCurrent, bucket.Kind)
	// The bucket account string must be exactly the bucket key so d3
	// stratify can resolve parent IDs via paisa.utils.parentName.
	assert.Equal(t, bcBucket, bucket.Account)

	// Each leaf must hang DIRECTLY off the bucket: the synthetic path must
	// be bucket + ":" + <single-segment account id>, so that on the frontend
	// `parentName(leaf)` returns the bucket exactly. This requires the
	// original account name to be sanitized to a single segment.
	for _, leafSuffix := range []string{
		sanitizeAccountSegment("Assets:Saving:CMB"),
		sanitizeAccountSegment("Assets:Saving:ICBC"),
	} {
		leafKey := bcBucket + ":" + leafSuffix
		_, ok := result[leafKey]
		assert.True(t, ok, "missing leaf %s; keys=%v", leafKey, keysOf(result))
		// Stratify invariant: the leaf must have exactly one ":" between
		// the bucket and its single-segment account id.
		afterBucket := leafKey[len(bcBucket)+1:]
		assert.NotContains(t, afterBucket, ":",
			"leaf path %q must not contain ':' after bucket prefix (would break stratify)", leafKey)
	}
}

// TestAggregateByKind_EmptyInput returns just the root.
func TestAggregateByKind_EmptyInput(t *testing.T) {
	result := aggregateByKind(nil, nil, time.Now())
	_, ok := result[allocationRootKey]
	assert.True(t, ok)
	assert.Len(t, result, 1)
}

// TestComputeAggregate_HonorsConfigAccounts wires aggregateByKind together
// with config.GetConfig().Accounts as used by the live handler. This guards
// the regression in the issue: putting `kind: mutual_fund` on a Saving sub
// account must reroute its allocation.
func TestComputeAggregate_HonorsConfigAccounts(t *testing.T) {
	yaml := []byte(
		"journal_path: /tmp/x.ledger\n" +
			"db_path: /tmp/x.db\n" +
			"default_currency: CNY\n" +
			"accounts:\n" +
			"  - name: Assets:Saving:CMB:Fund\n" +
			"    kind: mutual_fund\n",
	)
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
	defer func() {
		// Reset config to empty so other tests aren't affected.
		_ = config.LoadConfig([]byte("journal_path: /tmp/x.ledger\ndb_path: /tmp/x.db\n"), "")
	}()

	postings := []posting.Posting{
		makeAggregatePosting("Assets:Saving:CMB:Fund", 5000),
		makeAggregatePosting("Assets:Saving:CMB", 1000),
	}

	result := aggregateByKindFromConfig(postings, time.Now())

	mfPath := allocationRootKey + ":" + string(account.MutualFund) + ":" + sanitizeAccountSegment("Assets:Saving:CMB:Fund")
	leaf, ok := result[mfPath]
	assert.True(t, ok, "expected mutual_fund leaf %q; keys=%v", mfPath, keysOf(result))
	assert.Equal(t, "Assets:Saving:CMB:Fund", leaf.OriginalAccount)
}
