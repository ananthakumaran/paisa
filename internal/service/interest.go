package service

import (
	"sync"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type interestCache struct {
	sync.Once
	postings map[time.Time][]posting.Posting
}

var icache interestCache

func loadInterestCache(db *gorm.DB) {
	var postings []posting.Posting
	result := db.Where("account like ?", "Income:Interest:%").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	icache.postings = lo.GroupBy(postings, func(p posting.Posting) time.Time { return p.Date })
}

func IsInterest(db *gorm.DB, p posting.Posting) bool {
	icache.Do(func() { loadInterestCache(db) })

	if p.Commodity != "INR" {
		return false
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
