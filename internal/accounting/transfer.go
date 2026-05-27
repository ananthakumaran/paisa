package accounting

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

// matchAccountGlob matches an account name against a glob pattern.
// Account names use ':' as the separator (not '/'), so we use
// doublestar.Match with '/' as the path separator while leaving the
// account string untouched — this means '*' matches across the rest of
// the account hierarchy (e.g. 'Assets:Bridge:*' matches both
// 'Assets:Bridge:IBKR' and 'Assets:Bridge:Brokerage:DanJuan').
// This mirrors the convention used elsewhere in paisa (see
// accounting.FilterByGlob, which uses filepath.Match in the same way).
func matchAccountGlob(pattern, account string) bool {
	matched, err := doublestar.Match(pattern, account)
	if err != nil {
		log.Warnf("Invalid transfer_accounts glob %q: %v", pattern, err)
		return false
	}
	return matched
}

// IsInternalTransfer reports whether the given set of postings (which
// must all share the same TransactionID) represents an internal
// transfer that should be excluded from Investment Timeline / Savings
// Rate.
//
// A transaction qualifies as an internal transfer when:
//  1. Every posting in the transaction is under Assets:* (i.e. there is
//     no leg in Income / Expenses / Liabilities / Equity).
//  2. At least one leg's account matches one of the supplied globs.
//
// This way salaries and expenses are never accidentally treated as
// transfers, even if a leg happens to be one of the transfer accounts.
func IsInternalTransfer(postings []posting.Posting, globs []string) bool {
	if len(globs) == 0 || len(postings) < 2 {
		return false
	}

	allAssets := lo.EveryBy(postings, func(p posting.Posting) bool {
		return utils.IsSameOrParent(p.Account, "Assets")
	})
	if !allAssets {
		return false
	}

	for _, p := range postings {
		for _, g := range globs {
			if matchAccountGlob(g, p.Account) {
				return true
			}
		}
	}
	return false
}

// FilterOutInternalTransfers groups postings by transaction id and
// removes any posting whose transaction qualifies as an internal
// transfer (see IsInternalTransfer).
func FilterOutInternalTransfers(postings []posting.Posting, globs []string) []posting.Posting {
	if len(globs) == 0 || len(postings) == 0 {
		return postings
	}

	byTx := lo.GroupBy(postings, func(p posting.Posting) string { return p.TransactionID })
	skip := make(map[string]bool, len(byTx))
	for id, ps := range byTx {
		if IsInternalTransfer(ps, globs) {
			skip[id] = true
		}
	}

	return lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return !skip[p.TransactionID]
	})
}
