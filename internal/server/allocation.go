package server

import (
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/account"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// allocationRootKey is the synthetic top-level node under which all
// kind-grouped buckets live. Using a fixed key (instead of "Assets") makes
// the API shape independent of the user's ledger account naming and keeps
// the d3 hierarchy stable.
const allocationRootKey = "Allocation"

// Aggregate represents a single node in the kind-grouped allocation tree.
//
// As of M1-E the tree is layered as:
//
//	Allocation                                   // root (always present)
//	  Allocation:<account_kind>                  // one per kind that has children
//	    Allocation:<account_kind>:<ledger:path>  // leaf, one per posting account
//
// The `Account` field is the synthetic path (used by d3 stratify on the
// frontend); `OriginalAccount` carries the real ledger account name for
// links / table display; `Kind` is machine-readable for the UI to look up
// a localized label.
type Aggregate struct {
	Date            time.Time           `json:"date"`
	Account         string              `json:"account"`
	OriginalAccount string              `json:"original_account,omitempty"`
	Kind            account.AccountKind `json:"kind,omitempty"`
	MarketAmount    decimal.Decimal     `json:"market_amount"`
}

type AllocationTargetConfig struct {
	Name     string
	Target   decimal.Decimal
	Accounts []string
}

type AllocationTarget struct {
	Name       string               `json:"name"`
	Target     decimal.Decimal      `json:"target"`
	Current    decimal.Decimal      `json:"current"`
	Aggregates map[string]Aggregate `json:"aggregates"`
}

func GetAllocation(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").All()

	now := utils.EndOfToday()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, now)
		return p
	})
	aggregates := computeAggregate(db, postings, now)
	aggregates_timeline := computeAggregateTimeline(db, postings)
	allocation_targets := computeAllocationTargets(db, postings)
	return gin.H{"aggregates": aggregates, "aggregates_timeline": aggregates_timeline, "allocation_targets": allocation_targets}
}

func computeAggregateTimeline(db *gorm.DB, postings []posting.Posting) []map[string]Aggregate {
	var timeline []map[string]Aggregate

	var p posting.Posting

	type RunningSum struct {
		units decimal.Decimal
		cost  decimal.Decimal
	}

	accumulator := make(map[string]map[string]RunningSum)

	if len(postings) == 0 {
		return timeline
	}

	configAccounts := toAccountLookup(config.GetConfig().Accounts)

	end := utils.EndOfToday()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			rsByAccount := accumulator[p.Account]
			if rsByAccount == nil {
				rsByAccount = make(map[string]RunningSum)
			}

			rs := rsByAccount[p.Commodity]
			rs.units = rs.units.Add(p.Quantity)
			rs.cost = rs.cost.Add(p.Amount)

			rsByAccount[p.Commodity] = rs
			accumulator[p.Account] = rsByAccount

		}

		result := make(map[string]Aggregate)

		for accountName, rsByAccount := range accumulator {
			marketAmount := decimal.Zero

			for commodity, rs := range rsByAccount {
				if utils.IsCurrency(commodity) {
					marketAmount = marketAmount.Add(rs.cost)
				} else {
					price := service.GetUnitPrice(db, commodity, start)
					if !price.Value.Equal(decimal.Zero) {
						marketAmount = marketAmount.Add(rs.units.Mul(price.Value))
					} else {
						marketAmount = marketAmount.Add(rs.cost)
					}
				}
			}

			kind := account.GetKind(accountName, configAccounts)
			synthetic := buildSyntheticPath(kind, accountName)
			result[synthetic] = Aggregate{
				Date:            start,
				Account:         synthetic,
				OriginalAccount: accountName,
				Kind:            kind,
				MarketAmount:    marketAmount,
			}

		}

		timeline = append(timeline, result)

	}
	return timeline
}

func computeAllocationTargets(db *gorm.DB, postings []posting.Posting) []AllocationTarget {
	var targetAllocations []AllocationTarget
	allocationTargetConfigs := config.GetConfig().AllocationTargets

	if len(postings) == 0 {
		return targetAllocations
	}

	totalMarketAmount := accounting.CurrentBalance(postings)

	for _, allocationTargetConfig := range allocationTargetConfigs {
		targetAllocations = append(targetAllocations, computeAllocationTarget(db, postings, allocationTargetConfig, totalMarketAmount))
	}

	return targetAllocations
}

func computeAllocationTarget(db *gorm.DB, postings []posting.Posting, allocationTargetConfig config.AllocationTarget, total decimal.Decimal) AllocationTarget {
	date := utils.EndOfToday()
	postings = accounting.FilterByGlob(postings, allocationTargetConfig.Accounts)
	aggregates := computeAggregate(db, postings, date)
	currentTotal := accounting.CurrentBalance(postings)
	return AllocationTarget{Name: allocationTargetConfig.Name, Target: decimal.NewFromFloat(allocationTargetConfig.Target), Current: (currentTotal.Div(total)).Mul(decimal.NewFromInt(100)), Aggregates: aggregates}
}

// computeAggregate resolves each posting's market amount as of `date` and
// then defers to aggregateByKind. Split from aggregateByKind so the latter
// stays unit-testable without a *gorm.DB.
func computeAggregate(db *gorm.DB, postings []posting.Posting, date time.Time) map[string]Aggregate {
	priced := lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, date)
		return p
	})
	return aggregateByKindFromConfigAt(priced, date)
}

// aggregateByKindFromConfig is a convenience wrapper that reads the
// account list from the loaded user config. Used directly by tests that
// want the production code path with config-driven kind overrides.
func aggregateByKindFromConfig(postings []posting.Posting, date time.Time) map[string]Aggregate {
	return aggregateByKindFromConfigAt(postings, date)
}

func aggregateByKindFromConfigAt(postings []posting.Posting, date time.Time) map[string]Aggregate {
	return aggregateByKind(postings, toAccountLookup(config.GetConfig().Accounts), date)
}

// toAccountLookup adapts the rich config.Account slice down to the minimal
// shape expected by account.GetKind (avoids an import cycle).
func toAccountLookup(in []config.Account) []account.Account {
	out := make([]account.Account, 0, len(in))
	for _, a := range in {
		out = append(out, account.Account{Name: a.Name, Kind: a.Kind})
	}
	return out
}

// aggregateByKind groups postings into a kind-bucketed tree.
//
// The returned map always contains a root entry under `allocationRootKey`;
// for every distinct AccountKind seen across `postings` it adds an
// intermediate bucket node and one leaf node per ledger account.
//
// Each posting must already carry its market amount in `MarketAmount`;
// callers using a DB-backed posting source should pre-populate that
// (see computeAggregate).
//
// `accounts` is the user-configured override list (M1-D); pass nil to
// rely entirely on path-prefix fallback rules.
func aggregateByKind(postings []posting.Posting, accounts []account.Account, date time.Time) map[string]Aggregate {
	result := make(map[string]Aggregate)
	result[allocationRootKey] = Aggregate{Account: allocationRootKey}

	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })

	// Track which kind buckets we've created so the bucket node is added
	// exactly once.
	bucketSeen := make(map[account.AccountKind]bool)

	for accountName, ps := range byAccount {
		kind := account.GetKind(accountName, accounts)
		bucketKey := bucketKeyFor(kind)
		if !bucketSeen[kind] {
			result[bucketKey] = Aggregate{Account: bucketKey, Kind: kind}
			bucketSeen[kind] = true
		}

		leafKey := buildSyntheticPath(kind, accountName)
		marketAmount := accounting.CurrentBalance(ps)
		result[leafKey] = Aggregate{
			Date:            date,
			Account:         leafKey,
			OriginalAccount: accountName,
			Kind:            kind,
			MarketAmount:    marketAmount,
		}
	}

	return result
}

// bucketKeyFor returns the synthetic path of the kind-level bucket node.
func bucketKeyFor(kind account.AccountKind) string {
	return allocationRootKey + ":" + string(kind)
}

// buildSyntheticPath joins the root, kind, and (sanitized) original ledger
// account into the synthetic colon path used by the frontend's d3
// hierarchy. The original ledger account is collapsed to a single segment
// via sanitizeAccountSegment so that d3 stratify resolves each leaf's
// parent to the kind bucket — not to some intermediate ":Assets:Saving:..."
// node that doesn't exist in the map.
func buildSyntheticPath(kind account.AccountKind, accountName string) string {
	return bucketKeyFor(kind) + ":" + sanitizeAccountSegment(accountName)
}

// sanitizeAccountSegment replaces the ledger account separator (`:`) with a
// safe character so the entire account name becomes a single segment in
// the synthetic colon-delimited path. The frontend keeps the real account
// name in Aggregate.OriginalAccount for display / linking.
func sanitizeAccountSegment(accountName string) string {
	out := make([]byte, len(accountName))
	for i := 0; i < len(accountName); i++ {
		if accountName[i] == ':' {
			out[i] = '/'
		} else {
			out[i] = accountName[i]
		}
	}
	return string(out)
}
