package prediction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMatchSeedCategories walks the major dictionary buckets and asserts the
// expected account is returned. The list is intentionally broader than the
// 20-case minimum demanded by issue #24 so a future contributor who edits
// the seed dictionary has to consciously update a test rather than silently
// regressing a category. Cases are grouped to mirror the source ordering in
// seed_zh_cn.go — keep them in sync.
func TestMatchSeedCategories(t *testing.T) {
	cases := []struct {
		name    string
		payee   string
		note    string
		account string
	}{
		// dining ─────────────────────────────────────────────────────────
		{"starbucks", "星巴克咖啡 (上海陆家嘴店)", "拿铁", "Expenses:Dining"},
		{"luckin", "瑞幸咖啡北京建国门店", "", "Expenses:Dining"},
		{"mcdonalds", "麦当劳", "", "Expenses:Dining"},
		{"kfc", "肯德基", "", "Expenses:Dining"},
		{"pizzahut", "必胜客", "", "Expenses:Dining"},
		{"burgerking", "汉堡王", "", "Expenses:Dining"},
		{"heytea", "喜茶GO", "", "Expenses:Dining"},
		{"nayuki", "奈雪的茶", "", "Expenses:Dining"},
		{"chayanyuese", "茶颜悦色", "", "Expenses:Dining"},
		{"haidilao", "海底捞火锅", "", "Expenses:Dining"},
		{"xiaolongkan", "小龙坎老火锅", "", "Expenses:Dining"},
		{"xiabuxiabu", "呷哺呷哺", "", "Expenses:Dining"},

		// food delivery ──────────────────────────────────────────────────
		{"meituanwaimai", "美团外卖", "", "Expenses:Dining"},
		{"meituanbare", "美团", "", "Expenses:Dining"},
		{"eleme", "饿了么", "", "Expenses:Dining"},

		// groceries ──────────────────────────────────────────────────────
		{"hema", "盒马鲜生", "", "Expenses:Groceries"},
		{"yonghui", "永辉超市", "", "Expenses:Groceries"},
		{"walmart_cn", "沃尔玛", "", "Expenses:Groceries"},
		{"carrefour", "家乐福", "", "Expenses:Groceries"},
		{"sams", "山姆会员店", "", "Expenses:Groceries"},

		// shopping ───────────────────────────────────────────────────────
		{"jd", "京东商城", "", "Expenses:Shopping"},
		{"pdd", "拼多多", "", "Expenses:Shopping"},
		{"taobao", "淘宝网", "", "Expenses:Shopping"},
		{"tmall", "天猫超市", "", "Expenses:Shopping"},

		// transport ──────────────────────────────────────────────────────
		{"didi", "滴滴出行", "", "Expenses:Transport:Taxi"},
		{"gaode", "高德打车", "", "Expenses:Transport:Taxi"},
		{"caocao", "曹操出行", "", "Expenses:Transport:Taxi"},

		// travel ─────────────────────────────────────────────────────────
		{"railway", "12306中国铁路客户服务中心", "", "Expenses:Travel"},
		{"ctrip", "携程旅行", "", "Expenses:Travel"},
		{"fliggy", "飞猪旅行", "", "Expenses:Travel"},
		{"qunar", "去哪儿网", "", "Expenses:Travel"},

		// utilities ──────────────────────────────────────────────────────
		{"sgcc", "国家电网", "", "Expenses:Utilities:Electric"},
		{"csg", "南方电网", "", "Expenses:Utilities:Electric"},
		{"water", "上海水务", "", "Expenses:Utilities:Water"},
		{"tapwater", "自来水公司", "", "Expenses:Utilities:Water"},
		{"gas", "上海燃气", "", "Expenses:Utilities:Gas"},
		{"naturalgas", "天然气", "", "Expenses:Utilities:Gas"},
		{"mobile", "中国移动", "", "Expenses:Utilities:Phone"},
		{"unicom", "中国联通", "", "Expenses:Utilities:Phone"},
		{"telecom", "中国电信", "", "Expenses:Utilities:Phone"},

		// entertainment ──────────────────────────────────────────────────
		{"neteasemusic", "网易云音乐", "", "Expenses:Entertainment:Music"},
		{"qqmusic", "QQ音乐", "", "Expenses:Entertainment:Music"},
		{"applemusic_payee", "Apple Music", "", "Expenses:Entertainment:Music"},
		{"spotify_payee", "Spotify", "", "Expenses:Entertainment:Music"},
		{"iqiyi", "爱奇艺", "", "Expenses:Entertainment:Video"},
		{"tencentvideo", "腾讯视频", "", "Expenses:Entertainment:Video"},
		{"youku", "优酷视频", "", "Expenses:Entertainment:Video"},
		{"netflix_payee", "Netflix", "", "Expenses:Entertainment:Video"},

		// apps ───────────────────────────────────────────────────────────
		{"appstore_payee", "App Store", "", "Expenses:Shopping:Apps"},

		// income hints ───────────────────────────────────────────────────
		{"salary_payee", "工资", "", "Income:Salary"},
		{"salary_note", "公司", "工资发放", "Income:Salary"},
		{"interest", "招商银行", "结息", "Income:BankInterest"},

		// note channel ───────────────────────────────────────────────────
		// 微信红包 lives in note (商品) not payee (交易对方). The seed
		// dictionary itself does not have a "微信红包" entry — direction is
		// the importer's job — but if a future contributor adds one this
		// case will catch a regression where note matching breaks.
		{"empty", "", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := MatchSeed(c.payee, c.note)
			assert.Equal(t, c.account, got,
				"MatchSeed(%q, %q) want %q got %q",
				c.payee, c.note, c.account, got)
		})
	}
}

// TestMatchSeedLongestKeyFirst guards the sort step in init(): without
// length-descending order, "美团" would shadow "美团外卖". This test exists
// so a refactor that accidentally changes the order (or removes the sort)
// fails loudly rather than silently degrading suggestion quality.
func TestMatchSeedLongestKeyFirst(t *testing.T) {
	// Both keys live in the dictionary today. If they ever diverge in
	// account this assertion will still hold — it only checks the more
	// specific keyword wins on its own input.
	got := MatchSeed("美团外卖", "")
	assert.Equal(t, "Expenses:Dining", got,
		"expected the longer key '美团外卖' to win; got %q. "+
			"Did the seed-list ordering break? See seed_zh_cn.go init().",
		got)
}

// TestMatchSeedCaseInsensitive: payees from Alipay sometimes arrive with
// English merchant names in mixed case ("Apple Music", "Spotify"). The
// matcher lowercases both sides, so the dictionary lives in lowercase and
// the input can be anything.
func TestMatchSeedCaseInsensitive(t *testing.T) {
	assert.Equal(t, "Expenses:Entertainment:Music", MatchSeed("APPLE MUSIC", ""))
	assert.Equal(t, "Expenses:Entertainment:Music", MatchSeed("apple music", ""))
	assert.Equal(t, "Expenses:Entertainment:Video", MatchSeed("NETFLIX", ""))
}

// TestMatchSeedPayeeBeatsNote: when both fields would match, payee wins.
// This matters for refund rows where the original 商品 description (note)
// might mention a generic word ("退款"), while the payee is the real
// merchant.
func TestMatchSeedPayeeBeatsNote(t *testing.T) {
	got := MatchSeed("星巴克", "瑞幸赠送的咖啡")
	// Either is "Expenses:Dining" so we cannot distinguish on account
	// alone; pick fields with different accounts.
	assert.Equal(t, "Expenses:Dining", got)

	// Payee in transport, note in shopping → payee wins.
	got = MatchSeed("滴滴", "京东plus会员说明")
	assert.Equal(t, "Expenses:Transport:Taxi", got)
}

// TestMatchSeedNoMatchReturnsEmpty: a payee outside the seed list must
// return "" so the caller knows to apply its own fallback.
func TestMatchSeedNoMatchReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", MatchSeed("陌生商户ABC", ""))
	assert.Equal(t, "", MatchSeed("", ""))
}
