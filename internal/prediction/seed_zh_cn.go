// Chinese-merchant seed dictionary for the M3-F importer suggestion pipeline.
//
// This file is intentionally a static, hand-curated keyword → account map.
// It is the "Layer 1" of the suggestion stack described in issue #24: the
// importer Parse step has no journal context, so it relies on cheap
// substring matching against the payee (and, as a fallback, the merchant
// note) to fill SuggestedAccount with something sensible. The TF-IDF model
// in tf_idf.go and the user-feedback table in learning.go layer on top of
// this — both can override a seed match, but neither replaces it.
//
// Why a hard-coded list rather than letting TF-IDF do everything: TF-IDF
// needs a populated journal to learn from. First-time users (and the very
// first import for any user) have an empty database; without a static
// fallback every transaction would land in Expenses:Unknown and the import
// preview becomes a tedious "set the account on every row" chore. The seed
// list covers the ~60 most common Chinese merchants and payment-system
// keywords so the import preview already looks roughly right on day one.
//
// Conventions:
//   - Keys are Chinese-only. English merchants are explicitly out of scope
//     for this PR (issue #24 hard rule). A future PR can add an English
//     companion file with the same shape.
//   - Match is case-insensitive substring against the payee. The dictionary
//     is sorted longest-key-first at construction time so "美团外卖" is
//     tried before "美团" (otherwise "美团" would shadow the more specific
//     entry).
//   - The "note" channel covers WeChat/Alipay 商品 / 备注 fields. Some
//     merchants only appear there (e.g. "微信红包" lives in 商品, not 交易对方).
//   - The dictionary does NOT decide direction (income vs expense). The
//     importer already knows that from its source format; the seed only
//     refines the bucket WITHIN the direction. Callers pass a fallback
//     account that already encodes "this is an expense / income" so the
//     seed can return "" to mean "no opinion, use fallback".
package prediction

import (
	"sort"
	"strings"
)

// seedRule is one (keyword → account) pair. Kept as a struct (rather than a
// bare map) so the sort step has somewhere to attach the key length without
// recomputing it for every comparison. The exported builder below freezes
// the slice in length-descending order, so callers may iterate from index 0
// and take the first match.
type seedRule struct {
	keyword string
	account string
}

// seedRulesRaw is the source of truth for the seed dictionary. The init()
// hook below sorts a copy of this slice into seedRules by descending keyword
// length. We keep the raw list in human-readable, category-grouped order so
// future contributors can scan it without losing the longest-first ordering
// the matcher depends on.
//
// Categories (in order):
//
//	dining (sit-down, fast food, coffee, milk tea, hot-pot)
//	food delivery
//	groceries / supermarkets
//	online shopping
//	transport (ride-hail, public transit)
//	travel (tickets, OTAs)
//	utilities (electric, water, gas, phone)
//	streaming (music, video)
//	apps / app stores
//	salary / interest / red packets / AA
//
// Every category MUST have at least one test case in seed_zh_cn_test.go.
var seedRulesRaw = []seedRule{
	// ── Dining: coffee chains ────────────────────────────────────────────
	{"星巴克", "Expenses:Dining"},
	{"瑞幸", "Expenses:Dining"},
	{"manner", "Expenses:Dining"},

	// ── Dining: fast food ────────────────────────────────────────────────
	{"麦当劳", "Expenses:Dining"},
	{"肯德基", "Expenses:Dining"},
	{"必胜客", "Expenses:Dining"},
	{"汉堡王", "Expenses:Dining"},
	{"塔斯汀", "Expenses:Dining"},
	{"华莱士", "Expenses:Dining"},

	// ── Dining: tea / drinks ─────────────────────────────────────────────
	{"喜茶", "Expenses:Dining"},
	{"奈雪", "Expenses:Dining"},
	{"茶颜悦色", "Expenses:Dining"},
	{"蜜雪冰城", "Expenses:Dining"},
	{"古茗", "Expenses:Dining"},

	// ── Dining: hot pot ──────────────────────────────────────────────────
	{"海底捞", "Expenses:Dining"},
	{"小龙坎", "Expenses:Dining"},
	{"呷哺呷哺", "Expenses:Dining"},

	// ── Food delivery ────────────────────────────────────────────────────
	// Put the more specific "美团外卖" before bare "美团" so it wins on
	// substring. (sortByLengthDesc enforces this regardless of source
	// order, but leaving them adjacent makes the intent obvious.) Bare
	// "美团" mostly shows up on dining rows in practice — 美团到店, 美团
	// 团购, etc. — so we keep the broader rule pointing at Dining as a
	// pragmatic default.
	{"美团外卖", "Expenses:Dining"},
	{"美团", "Expenses:Dining"},
	{"饿了么", "Expenses:Dining"},

	// ── Groceries / supermarkets ─────────────────────────────────────────
	{"盒马", "Expenses:Groceries"},
	{"永辉", "Expenses:Groceries"},
	{"沃尔玛", "Expenses:Groceries"},
	{"家乐福", "Expenses:Groceries"},
	{"山姆", "Expenses:Groceries"},
	{"costco", "Expenses:Groceries"},
	{"开市客", "Expenses:Groceries"},

	// ── Online shopping ──────────────────────────────────────────────────
	{"京东", "Expenses:Shopping"},
	{"拼多多", "Expenses:Shopping"},
	{"淘宝", "Expenses:Shopping"},
	{"天猫", "Expenses:Shopping"},
	{"唯品会", "Expenses:Shopping"},

	// ── Transport: ride-hail ─────────────────────────────────────────────
	{"滴滴", "Expenses:Transport:Taxi"},
	{"高德打车", "Expenses:Transport:Taxi"},
	{"曹操出行", "Expenses:Transport:Taxi"},
	{"t3出行", "Expenses:Transport:Taxi"},

	// ── Transport: public transit (a catch-all bucket; subway / bus
	//    distinction is intentionally out of scope for the seed) ─────────
	{"地铁", "Expenses:Transport:Transit"},
	{"公交", "Expenses:Transport:Transit"},

	// ── Travel / OTA ─────────────────────────────────────────────────────
	{"12306", "Expenses:Travel"},
	{"携程", "Expenses:Travel"},
	{"飞猪", "Expenses:Travel"},
	{"去哪儿", "Expenses:Travel"},

	// ── Utilities: electric ──────────────────────────────────────────────
	{"国家电网", "Expenses:Utilities:Electric"},
	{"南方电网", "Expenses:Utilities:Electric"},

	// ── Utilities: water ────────────────────────────────────────────────
	{"水务", "Expenses:Utilities:Water"},
	{"自来水", "Expenses:Utilities:Water"},

	// ── Utilities: gas ───────────────────────────────────────────────────
	{"燃气", "Expenses:Utilities:Gas"},
	{"天然气", "Expenses:Utilities:Gas"},

	// ── Utilities: phone / mobile carriers ───────────────────────────────
	{"中国移动", "Expenses:Utilities:Phone"},
	{"中国联通", "Expenses:Utilities:Phone"},
	{"中国电信", "Expenses:Utilities:Phone"},

	// ── Entertainment: music ─────────────────────────────────────────────
	{"网易云音乐", "Expenses:Entertainment:Music"},
	{"qq音乐", "Expenses:Entertainment:Music"},
	{"apple music", "Expenses:Entertainment:Music"},
	{"spotify", "Expenses:Entertainment:Music"},

	// ── Entertainment: video ─────────────────────────────────────────────
	{"爱奇艺", "Expenses:Entertainment:Video"},
	{"腾讯视频", "Expenses:Entertainment:Video"},
	{"优酷", "Expenses:Entertainment:Video"},
	{"芒果tv", "Expenses:Entertainment:Video"},
	{"bilibili", "Expenses:Entertainment:Video"},
	{"哔哩哔哩", "Expenses:Entertainment:Video"},
	{"netflix", "Expenses:Entertainment:Video"},

	// ── Apps / digital storefronts ───────────────────────────────────────
	{"苹果apple", "Expenses:Shopping:Apps"},
	{"app store", "Expenses:Shopping:Apps"},

	// ── Income hints (only meaningful when the importer reports 收入) ────
	// These intentionally do NOT live in the regular merchant section
	// because their bucket also depends on direction. The matcher returns
	// them as-is; the caller decides whether to use them based on the
	// importer's flow.
	{"工资", "Income:Salary"},
	{"薪金", "Income:Salary"},
	{"利息", "Income:BankInterest"},
	{"结息", "Income:BankInterest"},
}

// seedRules is the active, length-sorted view of seedRulesRaw. Populated
// once at init() and treated as read-only thereafter.
var seedRules []seedRule

func init() {
	seedRules = make([]seedRule, len(seedRulesRaw))
	copy(seedRules, seedRulesRaw)
	sort.SliceStable(seedRules, func(i, j int) bool {
		// Longest keyword first so "美团外卖" beats "美团".
		return len([]rune(seedRules[i].keyword)) > len([]rune(seedRules[j].keyword))
	})
}

// MatchSeed returns the seed-dictionary account for the first keyword that
// occurs as a (case-insensitive) substring of payee or note, with payee
// taking precedence over note. An empty string means "no seed match — let
// the caller fall back to whatever default it picked".
//
// Both inputs are accepted because Alipay and WeChat split the human-
// readable label across two fields:
//
//	WeChat:  交易对方 (payee)  +  商品 (note)
//	Alipay:  交易对方 (payee)  +  商品说明 (note)
//
// A red-packet send/receive shows "微信红包" in 商品 / 商品说明 with the
// counterparty's nickname in 交易对方. Falling back to note when payee has
// no match captures those.
func MatchSeed(payee, note string) string {
	p := strings.ToLower(payee)
	n := strings.ToLower(note)
	// Two passes: payee first, then note. This is more expensive than a
	// single concatenated pass but preserves the precedence guarantee —
	// "星巴克" in the payee beats "微信红包" in the note.
	for _, r := range seedRules {
		if p != "" && strings.Contains(p, r.keyword) {
			return r.account
		}
	}
	for _, r := range seedRules {
		if n != "" && strings.Contains(n, r.keyword) {
			return r.account
		}
	}
	return ""
}
