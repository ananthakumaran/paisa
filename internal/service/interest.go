package service

import (
	"strings"
	"sync"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type interestCache struct {
	sync.Once
	postings map[int64][]posting.Posting
}

var icache interestCache

func loadInterestCache(db *gorm.DB) {
	postings := query.Init(db).Like("Income:Interest:%").All()
	icache.postings = lo.GroupBy(postings, func(p posting.Posting) int64 { return p.Date.Unix() })
}

type interestRepaymentCache struct {
	sync.Once
	postings map[int64][]posting.Posting
}

var irepaymentCache interestRepaymentCache

func loadInterestRepaymentCache(db *gorm.DB) {
	postings := query.Init(db).Like("Expenses:Interest:%").All()
	irepaymentCache.postings = lo.GroupBy(postings, func(p posting.Posting) int64 { return p.Date.Unix() })
}

func ClearInterestCache() {
	icache = interestCache{}
	irepaymentCache = interestRepaymentCache{}
}

func CapitalGainsSourceAccount(account string) string {
	parts := strings.Split(account, ":")
	return "Assets:" + strings.Join(parts[2:], ":")
}

func CapitalGainsAccount(account string) string {
	parts := strings.Split(account, ":")
	return "Income:CapitalGains:" + strings.Join(parts[1:], ":")
}

func IsCapitalGains(p posting.Posting) bool {
	if utils.IsParent(p.Account, "Income:CapitalGains") {
		return true
	}

	return false
}

func IsRefund(p posting.Posting) bool {
	if utils.IsParent(p.Account, "Income:Refund") {
		return true
	}

	return false
}

func IsStockSplit(db *gorm.DB, p posting.Posting) bool {
	if utils.IsCurrency(p.Commodity) {
		return false
	}

	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return false
	}

	for _, tp := range t.Postings {
		if utils.IsCurrency(tp.Commodity) || tp.Account != p.Account {
			return false
		}
	}
	return true
}

// IsBonus checks whether the posting credits units of a commodity to an
// asset account against an Income account, like a bonus issue
//
//	Assets:Equity:ABC    10 ABC @ 150 INR
//	Income:Bonus:ABC
//
// Units credited against Income:CapitalGains, Income:Dividend or
// Income:Interest are not considered, as they are reinvested payouts
// which carry a real cost.
func IsBonus(db *gorm.DB, p posting.Posting) bool {
	if utils.IsCurrency(p.Commodity) || !p.Quantity.IsPositive() || !utils.IsParent(p.Account, "Assets") {
		return false
	}

	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return false
	}

	income := false
	for _, tp := range t.Postings {
		if tp.Account == p.Account {
			continue
		}

		if !utils.IsParent(tp.Account, "Income") ||
			IsCapitalGains(tp) ||
			utils.IsSameOrParent(tp.Account, "Income:Dividend") ||
			utils.IsSameOrParent(tp.Account, "Income:Interest") {
			return false
		}
		income = true
	}
	return income
}

func IsSellWithCapitalGains(db *gorm.DB, p posting.Posting) bool {
	if utils.IsCurrency(p.Commodity) {
		return false
	}

	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return false
	}

	for _, tp := range t.Postings {
		if IsCapitalGains(tp) {
			return true
		}
	}
	return false
}

// SellAmount returns the sale proceeds of the sell posting. When the
// gain is booked against Income:CapitalGains, the asset posting is
// valued at cost and the gain has to be added back.
//
//	Assets:Checking                     3900 INR
//	Income:CapitalGains:Equity:ABC     -1900 INR
//	Assets:Equity:ABC                    -30 ABC {66.67 INR} @ 130 INR
//
// The gain is taken from Income:CapitalGains:{name} for the account
// Assets:{name}, or from all the capital gains postings if only one
// account is sold in the transaction. It is distributed among the sell
// postings of the account in proportion to the units.
func SellAmount(db *gorm.DB, p posting.Posting) decimal.Decimal {
	if utils.IsCurrency(p.Commodity) || !p.Quantity.IsNegative() {
		return p.Amount
	}

	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return p.Amount
	}

	isSell := func(tp posting.Posting) bool {
		return !utils.IsCurrency(tp.Commodity) && tp.Quantity.IsNegative() && utils.IsParent(tp.Account, "Assets")
	}

	gains := lo.Filter(t.Postings, func(tp posting.Posting, _ int) bool {
		return tp.Account == CapitalGainsAccount(p.Account)
	})
	if len(gains) == 0 {
		sold := lo.Uniq(lo.Map(lo.Filter(t.Postings, func(tp posting.Posting, _ int) bool { return isSell(tp) }), func(tp posting.Posting, _ int) string { return tp.Account }))
		if len(sold) != 1 {
			return p.Amount
		}
		gains = lo.Filter(t.Postings, func(tp posting.Posting, _ int) bool { return IsCapitalGains(tp) })
		if len(gains) == 0 {
			return p.Amount
		}
	}

	sells := lo.Filter(t.Postings, func(tp posting.Posting, _ int) bool {
		return isSell(tp) && tp.Account == p.Account
	})
	quantity := utils.SumBy(sells, func(tp posting.Posting) decimal.Decimal { return tp.Quantity })
	if quantity.IsZero() {
		return p.Amount
	}

	proceeds := utils.SumBy(sells, func(tp posting.Posting) decimal.Decimal { return tp.Amount }).
		Add(utils.SumBy(gains, func(tp posting.Posting) decimal.Decimal { return tp.Amount }))
	return proceeds.Mul(p.Quantity).Div(quantity)
}

func IsContraPostingRefund(db *gorm.DB, p posting.Posting) bool {
	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return false
	}

	for _, tp := range t.Postings {
		if IsRefund(tp) {
			return true
		}
	}
	return false
}

func IsInterestRepayment(db *gorm.DB, p posting.Posting) bool {
	irepaymentCache.Do(func() { loadInterestRepaymentCache(db) })

	if !utils.IsCurrency(p.Commodity) {
		return false
	}

	if strings.HasPrefix(p.Account, "Expenses:Interest:") {
		return true
	}

	for _, ip := range irepaymentCache.postings[p.Date.Unix()] {

		if ip.Date.Equal(p.Date) &&
			ip.Amount.Neg().Equal(p.Amount) &&
			ip.Payee == p.Payee {
			return true
		}
	}

	return false
}

func IsInterest(db *gorm.DB, p posting.Posting) bool {
	icache.Do(func() { loadInterestCache(db) })

	if !utils.IsCurrency(p.Commodity) {
		return false
	}

	for _, ip := range icache.postings[p.Date.Unix()] {

		if ip.Date.Equal(p.Date) &&
			ip.Amount.Neg().Equal(p.Amount) &&
			ip.Payee == p.Payee {
			return true
		}
	}

	return false
}
