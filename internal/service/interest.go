package service

import (
	"strings"
	"sync"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type interestCache struct {
	sync.Once
	postings map[time.Time][]posting.Posting
}

var icache interestCache

func loadInterestCache(db *gorm.DB) {
	postings := query.Init(db).Like("Income:Interest:%").All()
	icache.postings = lo.GroupBy(postings, func(p posting.Posting) time.Time { return p.Date })
}

func ClearInterestCache() {
	icache = interestCache{}
}

func IsInterest(db *gorm.DB, p posting.Posting) bool {
	icache.Do(func() { loadInterestCache(db) })

	if p.Commodity != "INR" {
		return false
	}

	if strings.HasPrefix(p.Account, "Expenses:Interest:") {
		return true
	}

	for _, ip := range icache.postings[p.Date] {
		if ip.Date.Equal(p.Date) &&
			-ip.Amount == p.Amount &&
			ip.Payee == p.Payee {
			return true
		}
	}

	return false
}
