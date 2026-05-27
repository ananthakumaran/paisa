package prediction_test

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB returns an in-memory SQLite gorm.DB with the account_learning
// table already migrated. Kept inline (rather than reusing inMemoryDB from
// the server package) so this package's tests don't acquire a circular
// import on /internal/server.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	if err := prediction.AutoMigrateLearning(db); err != nil {
		t.Fatalf("migrate account_learning: %v", err)
	}
	return db
}

// TestRecordAndLookup is the happy-path smoke test: write one observation,
// read it back. Confirms migration ran, the column names line up, and the
// trim-on-write logic preserves the original payee.
func TestRecordAndLookup(t *testing.T) {
	db := newTestDB(t)
	assert.NoError(t, prediction.RecordUserChoice(db, "星巴克咖啡(陆家嘴)", "Expenses:Dining"))
	got := prediction.LookupLearned(db, "星巴克咖啡(陆家嘴)")
	assert.Equal(t, "Expenses:Dining", got)
}

// TestRecordIncrementsCount: repeated confirmations bump Count rather than
// creating duplicate rows. This is what makes LookupLearned's "highest
// count wins" tiebreak meaningful.
func TestRecordIncrementsCount(t *testing.T) {
	db := newTestDB(t)
	for i := 0; i < 3; i++ {
		assert.NoError(t, prediction.RecordUserChoice(db, "京东", "Expenses:Shopping"))
	}

	var rows []prediction.AccountLearning
	assert.NoError(t, db.Where("payee = ?", "京东").Find(&rows).Error)
	if assert.Len(t, rows, 1, "expected a single row with count=3, got %d rows", len(rows)) {
		assert.Equal(t, uint(3), rows[0].Count)
	}
}

// TestLookupPicksHighestCount: when a payee has multiple recorded accounts
// (because the user occasionally categorises 京东 → Investment for stock
// tickets), the one with the highest confirmation count wins.
func TestLookupPicksHighestCount(t *testing.T) {
	db := newTestDB(t)
	// Three confirmations for Shopping, one for Investment.
	for i := 0; i < 3; i++ {
		assert.NoError(t, prediction.RecordUserChoice(db, "京东", "Expenses:Shopping"))
	}
	assert.NoError(t, prediction.RecordUserChoice(db, "京东", "Assets:Investment:JDStock"))

	got := prediction.LookupLearned(db, "京东")
	assert.Equal(t, "Expenses:Shopping", got)
}

// TestLookupUnknownPayeeReturnsEmpty: payees the user has never confirmed
// must return "" so SuggestForPayee falls through to the seed dictionary.
func TestLookupUnknownPayeeReturnsEmpty(t *testing.T) {
	db := newTestDB(t)
	assert.Equal(t, "", prediction.LookupLearned(db, "陌生商户"))
}

// TestRecordIgnoresEmpty: the UI sometimes ships placeholder rows with an
// empty payee or account. Those must NOT pollute the learning table — they
// would shadow real entries on lookup.
func TestRecordIgnoresEmpty(t *testing.T) {
	db := newTestDB(t)
	assert.NoError(t, prediction.RecordUserChoice(db, "", "Expenses:Dining"))
	assert.NoError(t, prediction.RecordUserChoice(db, "星巴克", ""))
	assert.NoError(t, prediction.RecordUserChoice(db, "   ", "   "))

	var count int64
	assert.NoError(t, db.Model(&prediction.AccountLearning{}).Count(&count).Error)
	assert.Equal(t, int64(0), count, "empty-input records must not be persisted")
}

// TestRecordAndLookupHandleNilDB: a nil db is the explicit signal that the
// caller does not want persistence (used by importer Parse, which has no
// db handle). Must NOT panic — must return nil / "".
func TestRecordAndLookupHandleNilDB(t *testing.T) {
	assert.NoError(t, prediction.RecordUserChoice(nil, "京东", "Expenses:Shopping"))
	assert.Equal(t, "", prediction.LookupLearned(nil, "京东"))
}

// TestSuggestForPayeeLayering verifies the three-layer fallback chain
// described in issue #24:
//  1. Learned mapping (DB) — most specific.
//  2. Seed dictionary — keyword match against payee / note.
//  3. Caller's fallback — usually Expenses:Unknown / Income:Unknown.
func TestSuggestForPayeeLayering(t *testing.T) {
	db := newTestDB(t)

	// Layer 1: learning beats seed. We teach the system that "星巴克咖啡"
	// goes to a custom account; seed would say plain Expenses:Dining.
	assert.NoError(t, prediction.RecordUserChoice(db, "星巴克咖啡", "Expenses:Coffee:Starbucks"))
	got := prediction.SuggestForPayee(db, "星巴克咖啡", "", "Expenses:Unknown")
	assert.Equal(t, "Expenses:Coffee:Starbucks", got,
		"learned mapping should override seed dictionary")

	// Layer 2: no learned entry, but seed matches.
	got = prediction.SuggestForPayee(db, "瑞幸咖啡 (上海)", "", "Expenses:Unknown")
	assert.Equal(t, "Expenses:Dining", got)

	// Layer 3: nothing matches → return fallback.
	got = prediction.SuggestForPayee(db, "陌生商户", "", "Expenses:Unknown")
	assert.Equal(t, "Expenses:Unknown", got)

	// nil db skips Layer 1 but still uses Layer 2 → 3.
	got = prediction.SuggestForPayee(nil, "瑞幸咖啡", "", "Expenses:Unknown")
	assert.Equal(t, "Expenses:Dining", got)
}
