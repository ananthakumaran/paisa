// Three-layer suggestion entry point for importer SuggestedAccount.
//
// SuggestForPayee is the single facade callers should use; the layered
// implementation (learned → seed → fallback) is an internal detail. The
// importer Parse step calls this with db=nil to get seed-only suggestions
// (Parse has no DB handle by design — see internal/importer/importer.go);
// the /api/import/parse server handler may re-run with a real db to
// upgrade the suggestion using the learning table before sending the
// preview to the UI.
package prediction

import "gorm.io/gorm"

// SuggestForPayee picks the best counterpart account for a payee.
//
// Layering (most specific first):
//
//  1. LookupLearned(db, payee) — exact payee match against the
//     account_learning table. Skipped when db is nil.
//  2. MatchSeed(payee, note) — case-insensitive substring match against
//     the hand-curated Chinese-merchant dictionary.
//  3. fallback — the account the caller would use if nothing matched,
//     typically Expenses:Unknown or Income:Unknown.
//
// A non-empty result from any layer short-circuits the rest. The function
// is safe to call from concurrent goroutines: seed matching is pure, and
// gorm's *DB is safe for concurrent reads.
func SuggestForPayee(db *gorm.DB, payee, note, fallback string) string {
	if learned := LookupLearned(db, payee); learned != "" {
		return learned
	}
	if seed := MatchSeed(payee, note); seed != "" {
		return seed
	}
	return fallback
}
