package service

import (
	"strings"
	"sync"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
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
