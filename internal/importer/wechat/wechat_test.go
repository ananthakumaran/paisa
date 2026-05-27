package wechat_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/importer/wechat"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// loadSample reads testdata/sample.csv. Kept as a helper so individual tests
// stay focused on assertions; the fixture is shared between Detect and Parse
// cases because a real importer must agree with itself: anything Detect()
// matches, Parse() should handle.
func loadSample(t *testing.T) []byte {
	t.Helper()
	bs, err := os.ReadFile(filepath.Join("testdata", "sample.csv"))
	if err != nil {
		t.Fatalf("read sample.csv: %v", err)
	}
	return bs
}

func TestCodeAndName(t *testing.T) {
	w := wechat.WeChat{}
	assert.Equal(t, "wechat", w.Code())
	assert.NotEmpty(t, w.Name())
}

// TestDetectByHeader: real WeChat exports always begin with the literal
// "微信支付账单明细" line. Detect must recognise it regardless of filename.
func TestDetectByHeader(t *testing.T) {
	w := wechat.WeChat{}
	content := loadSample(t)
	assert.True(t, w.Detect("random.csv", content), "expected detect-by-header")
}

// TestDetectByHeaderHandlesBOM: UTF-8 BOM (EF BB BF) is present at the very
// top of every WeChat export. Detection must not be confused by it.
func TestDetectByHeaderHandlesBOM(t *testing.T) {
	w := wechat.WeChat{}
	bomContent := append([]byte{0xEF, 0xBB, 0xBF}, []byte("微信支付账单明细\n")...)
	assert.True(t, w.Detect("anything.csv", bomContent))
}

// TestDetectByFilename: filename heuristic kicks in when content is unknown
// (e.g. the user renamed the export). Both English "wechat" and Chinese "微信"
// must match.
func TestDetectByFilename(t *testing.T) {
	w := wechat.WeChat{}
	assert.True(t, w.Detect("wechat-2024-01.csv", []byte("garbage")))
	assert.True(t, w.Detect("微信账单.csv", []byte("garbage")))
}

// TestDetectNoMatch: unrelated CSV must not match — false positives crowd
// the picker UI and confuse users.
func TestDetectNoMatch(t *testing.T) {
	w := wechat.WeChat{}
	assert.False(t, w.Detect("bank.csv", []byte("date,amount\n2024-01-01,10\n")))
}

// TestParseFromSample is the end-to-end happy-path: the fixture has a mix of
// merchant payments, red packets (sent + received), an internal-transfer, a
// withdrawal, a failed transaction, and a refund. We assert the parser drops
// only the failed row and the neutral-transfer row, and produces the right
// signs and counterpart guesses for the rest.
func TestParseFromSample(t *testing.T) {
	w := wechat.WeChat{}
	content := loadSample(t)
	txns, err := w.Parse(content)
	assert.NoError(t, err)

	// Drops: 支付失败 row (#5) and the neutral 零钱提现 row (#8) with 收/支="/".
	// Keeps: 1,2,3,4,6,7,9,10 → 8 rows.
	if !assert.Len(t, txns, 8) {
		for i, tx := range txns {
			t.Logf("txn[%d] payee=%s amount=%s suggested=%s", i, tx.Payee, tx.Amount, tx.SuggestedAccount)
		}
		return
	}

	// Row 1: 星巴克 38.00 支出 → Expenses:Dining, amount positive
	assert.Equal(t, "星巴克", txns[0].Payee)
	assert.Equal(t, "星巴克拿铁", txns[0].Note)
	assert.True(t, txns[0].Amount.Equal(decimal.NewFromFloat(38.00)), "want 38.00, got %s", txns[0].Amount)
	assert.Equal(t, "Expenses:Dining", txns[0].SuggestedAccount)
	assert.Equal(t, "CNY", txns[0].Currency)
	assert.Equal(t, 2024, txns[0].Date.Year())
	assert.Equal(t, 1, int(txns[0].Date.Month()))
	assert.Equal(t, 2, txns[0].Date.Day())
	assert.NotEmpty(t, txns[0].RawText)

	// Row 2: 滴滴出行 25.50 支出 → Expenses:Transport:Taxi
	assert.Equal(t, "滴滴出行", txns[1].Payee)
	assert.True(t, txns[1].Amount.Equal(decimal.NewFromFloat(25.50)))
	assert.Equal(t, "Expenses:Transport:Taxi", txns[1].SuggestedAccount)

	// Row 3: 微信红包 sent 88.00 → Expenses:Social:RedPacket
	assert.Equal(t, "微信红包", txns[2].Note)
	assert.True(t, txns[2].Amount.Equal(decimal.NewFromFloat(88.00)))
	assert.Equal(t, "Expenses:Social:RedPacket", txns[2].SuggestedAccount)

	// Row 4: 微信红包 received 66.00 → Income:RedPacket, amount NEGATIVE
	// (per ParsedTxn convention: negative = money entering source account).
	assert.Equal(t, "微信红包", txns[3].Note)
	assert.True(t, txns[3].Amount.Equal(decimal.NewFromFloat(-66.00)), "want -66.00, got %s", txns[3].Amount)
	assert.Equal(t, "Income:RedPacket", txns[3].SuggestedAccount)

	// Row 5 (raw) was 支付失败 → dropped. So txns[4] is the original row 6 (瑞幸).
	assert.Equal(t, "瑞幸咖啡", txns[4].Payee)
	assert.True(t, txns[4].Amount.Equal(decimal.NewFromFloat(19.00)))
	assert.Equal(t, "Expenses:Dining", txns[4].SuggestedAccount)

	// txns[5] is original row 7: 转账 to 匿名好友C — internal transfer hint.
	assert.Equal(t, "匿名好友C", txns[5].Payee)
	assert.True(t, txns[5].Amount.Equal(decimal.NewFromFloat(50.00)))
	assert.Equal(t, "Assets:Checking", txns[5].SuggestedAccount)

	// Row 8 (零钱提现, 收/支 "/") → dropped.

	// txns[6] is original row 9: 拼多多 → Expenses:Shopping
	assert.Equal(t, "拼多多", txns[6].Payee)
	assert.True(t, txns[6].Amount.Equal(decimal.NewFromFloat(99.99)))
	assert.Equal(t, "Expenses:Shopping", txns[6].SuggestedAccount)

	// txns[7] is original row 10: 二维码收款 from 匿名好友D — default income.
	assert.Equal(t, "匿名好友D", txns[7].Payee)
	assert.True(t, txns[7].Amount.Equal(decimal.NewFromFloat(-34.00)), "want -34.00, got %s", txns[7].Amount)
	assert.Equal(t, "Income:Unknown", txns[7].SuggestedAccount)
}

// TestParseStripsBOM: the parser must not choke on the UTF-8 BOM at offset 0,
// nor leak the BOM bytes into the first transaction's RawText / fields.
func TestParseStripsBOM(t *testing.T) {
	w := wechat.WeChat{}
	content := loadSample(t)
	// Sample already has BOM; sanity check the test data first.
	if len(content) < 3 || content[0] != 0xEF || content[1] != 0xBB || content[2] != 0xBF {
		t.Fatal("sample.csv must start with UTF-8 BOM for this test to be meaningful")
	}
	txns, err := w.Parse(content)
	assert.NoError(t, err)
	assert.NotEmpty(t, txns)
}

// TestParseStripsCurrencySymbol: every amount in the source is prefixed with
// "¥". The decimal parser must not see that character.
func TestParseStripsCurrencySymbol(t *testing.T) {
	w := wechat.WeChat{}
	content := loadSample(t)
	txns, err := w.Parse(content)
	assert.NoError(t, err)
	for i, tx := range txns {
		assert.False(t, tx.Amount.IsZero(), "txn %d has zero amount; ¥ may not have been stripped: raw=%q", i, tx.RawText)
	}
}

// TestParseAcceptsPlainAmount: some older or hand-edited exports omit the ¥
// prefix. Parsing must still succeed for "38.00".
func TestParseAcceptsPlainAmount(t *testing.T) {
	w := wechat.WeChat{}
	minimal := "微信支付账单明细\n" +
		"----------------------微信支付账单明细列表--------------------\n" +
		"交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注\n" +
		"2024-01-02 08:15:00,商户消费,星巴克,星巴克拿铁,支出,38.00,零钱,支付成功,4200001000000000001,SBX001,/\n"
	txns, err := w.Parse([]byte(minimal))
	assert.NoError(t, err)
	if assert.Len(t, txns, 1) {
		assert.True(t, txns[0].Amount.Equal(decimal.NewFromFloat(38.00)))
	}
}

// TestParseDropsFailedRows: any row whose 当前状态 contains "失败" is discarded.
// We construct a focused fixture so the assertion does not depend on the
// shared sample.csv.
func TestParseDropsFailedRows(t *testing.T) {
	w := wechat.WeChat{}
	minimal := "微信支付账单明细\n" +
		"----------------------微信支付账单明细列表--------------------\n" +
		"交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注\n" +
		"2024-01-02 08:15:00,商户消费,京东,日用品,支出,¥120.00,零钱通,支付失败,4200001000000000005,JD005,/\n" +
		"2024-01-03 08:15:00,商户消费,星巴克,拿铁,支出,¥38.00,零钱,支付成功,4200001000000000006,SBX,/\n"
	txns, err := w.Parse([]byte(minimal))
	assert.NoError(t, err)
	if assert.Len(t, txns, 1) {
		assert.Equal(t, "星巴克", txns[0].Payee)
	}
}

// TestParseSkipsNeutralTransfers: rows with 收/支 = "/" or empty represent
// pocket-to-pocket movements (零钱提现 etc.) and should be skipped by the
// importer — the user wants those tracked as transfers in the ledger, not
// as expenses.
func TestParseSkipsNeutralTransfers(t *testing.T) {
	w := wechat.WeChat{}
	minimal := "微信支付账单明细\n" +
		"----------------------微信支付账单明细列表--------------------\n" +
		"交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注\n" +
		"2024-01-20 16:30:00,零钱提现,/,零钱提现,/,¥50.00,零钱,提现已到账,4200001000000000008,WD,/\n"
	txns, err := w.Parse([]byte(minimal))
	assert.NoError(t, err)
	assert.Empty(t, txns)
}

// TestParseEmpty: an empty body (no separator, no header) must produce a
// helpful error rather than panicking.
func TestParseEmpty(t *testing.T) {
	w := wechat.WeChat{}
	_, err := w.Parse([]byte(""))
	assert.Error(t, err)
}

// TestRegisteredOnInit: importing the wechat package via init() must register
// the importer. Without this, the HTTP routes would never see the importer.
func TestRegisteredOnInit(t *testing.T) {
	found := importer.ByCode("wechat")
	if assert.NotNil(t, found, "wechat importer must auto-register via init()") {
		assert.Equal(t, "wechat", found.Code())
	}
}
