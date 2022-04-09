package service

import (
	"sync"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/google/btree"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type priceCache struct {
	mu         sync.Mutex
	pricesTree map[string]*btree.BTree
}

var cache priceCache

func loadCache(db *gorm.DB) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.pricesTree != nil {
		return
	}

	var prices []price.Price
	result := db.Find(&prices)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	cache.pricesTree = make(map[string]*btree.BTree)

	for _, price := range prices {
		if cache.pricesTree[price.CommodityName] == nil {
			cache.pricesTree[price.CommodityName] = btree.New(2)
		}

		cache.pricesTree[price.CommodityName].ReplaceOrInsert(price)
	}
}

func GetMarketPrice(db *gorm.DB, p posting.Posting, date time.Time) float64 {
	loadCache(db)
	if p.Commodity == "INR" {
		return p.Amount
	}

	pt := cache.pricesTree[p.Commodity]
	if pt != nil {
		var pc price.Price
		var found bool

		pt.DescendLessOrEqual(price.Price{Date: date}, func(item btree.Item) bool {
			pc = item.(price.Price)
			found = true
			return false
		})

		if found {
			return p.Quantity * pc.Value
		}
	}

	return p.Amount
}
