// User-feedback layer of the importer suggestion stack. When a user accepts
// (or edits) the importer preview's SuggestedAccount and commits the row,
// the chosen (payee → account) pair is recorded here. The next time the
// same payee shows up, LookupLearned returns the most-recently-confirmed
// account, which SuggestForPayee uses to overlay the static seed
// dictionary.
//
// Design notes:
//   - Composite primary key (payee, account). One payee can map to MULTIPLE
//     accounts (think 京东 → Shopping vs 京东金融 → Investment); we keep
//     every observation and let LookupLearned pick the one with the highest
//     count, breaking ties on UpdatedAt. This lets users "undo" a mis-edit
//     by recording the correct choice a few times rather than needing a
//     dedicated delete API.
//   - Opt-in: writes only happen from explicit calls to RecordUserChoice.
//     The importer Parse step never writes; the server's /api/import/commit
//     handler is the single caller in production.
//   - Schema lives in this file (rather than the model/ tree) because it is
//     conceptually a cache of learned suggestions, not a journal-derived
//     entity. The prediction package owns it end-to-end.
package prediction

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// AccountLearning is the GORM model backing the account_learning table.
// One row per (payee, account) observation. Count is the running tally of
// times the user confirmed this pairing; UpdatedAt is the last touch and
// breaks ties when two accounts have the same count.
//
// The Payee column is indexed because LookupLearned filters by it on every
// hot-path call (one query per importer-preview row).
type AccountLearning struct {
	// ID is a synthetic primary key. We do NOT use (Payee, Account) as a
	// composite PK because gorm's upsert on composite keys is brittle
	// across the sqlite / postgres drivers; an integer PK plus an explicit
	// unique index gives the same uniqueness guarantee with simpler SQL.
	ID uint `gorm:"primaryKey" json:"id"`

	// Payee is the exact 交易对方 / 商户名 string the importer captured.
	// We deliberately store it verbatim rather than normalising — the
	// downstream matcher uses substring lookup, so any kept characters
	// only make the match MORE specific.
	Payee string `gorm:"index:idx_account_learning_payee;uniqueIndex:idx_account_learning_payee_account,priority:1" json:"payee"`

	// Account is the ledger account the user picked for this payee.
	Account string `gorm:"uniqueIndex:idx_account_learning_payee_account,priority:2" json:"account"`

	// Count is how many times the user has confirmed this pairing.
	// RecordUserChoice increments on conflict.
	Count uint `gorm:"default:1" json:"count"`

	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins the table to a stable name independent of the package
// rename history; the implicit gorm default would be "account_learnings"
// which reads awkwardly in SQL.
func (AccountLearning) TableName() string { return "account_learning" }

// AutoMigrateLearning creates / updates the account_learning table. Wired
// into model.AutoMigrate so it runs on every serve / update — same
// lifecycle as posting and price tables.
//
// Returns the error verbatim so callers (notably model.AutoMigrate) can
// decide whether to bail or log-and-continue. Today nothing in the call
// chain treats migration errors as fatal; we surface the error to keep
// future work honest.
func AutoMigrateLearning(db *gorm.DB) error {
	return db.AutoMigrate(&AccountLearning{})
}

// RecordUserChoice persists a (payee → account) observation. Empty payee
// or account are silently ignored (the UI sometimes ships placeholder rows
// when the user did not edit anything; we do not want those polluting the
// table). A nil db is also a no-op so unit tests that don't care about
// learning can pass nil.
//
// Semantics on conflict: if a row already exists for the same (payee,
// account), Count is incremented by 1 and UpdatedAt bumped. We do this
// inside a transaction so two concurrent commits on the same payee /
// account don't race on the count.
func RecordUserChoice(db *gorm.DB, payee, account string) error {
	if db == nil {
		return nil
	}
	payee = strings.TrimSpace(payee)
	account = strings.TrimSpace(account)
	if payee == "" || account == "" {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var existing AccountLearning
		err := tx.Where("payee = ? AND account = ?", payee, account).First(&existing).Error
		now := time.Now()
		if err == nil {
			// Upsert path. Bump count + touch UpdatedAt.
			existing.Count++
			existing.UpdatedAt = now
			return tx.Save(&existing).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		row := AccountLearning{
			Payee:     payee,
			Account:   account,
			Count:     1,
			UpdatedAt: now,
		}
		return tx.Create(&row).Error
	})
}

// LookupLearned returns the most-confirmed account for payee, or "" if the
// learning table has never seen this payee (or if db is nil). Ranking:
//
//  1. Highest Count wins.
//  2. Ties broken by most recent UpdatedAt.
//
// The query is keyed by an exact payee match — substring matching is the
// seed dictionary's job. We deliberately keep this layer literal so a user
// who teaches the system "支付宝-星巴克(上海店) → Expenses:Dining" gets
// EXACTLY that mapping back; the seed dictionary already covers the looser
// substring case.
func LookupLearned(db *gorm.DB, payee string) string {
	if db == nil {
		return ""
	}
	payee = strings.TrimSpace(payee)
	if payee == "" {
		return ""
	}
	var row AccountLearning
	err := db.
		Where("payee = ?", payee).
		Order("count desc, updated_at desc").
		Limit(1).
		First(&row).Error
	if err != nil {
		return ""
	}
	return row.Account
}
