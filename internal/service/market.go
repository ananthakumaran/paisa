package service

import (
	"sync"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/google/btree"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type priceCache struct {
	mu         sync.Mutex
	loaded     bool
	pricesTree map[string]*btree.BTree
}

var pcache priceCache

func loadPriceCache(db *gorm.DB) {
	pcache.mu.Lock()
	defer pcache.mu.Unlock()

	if pcache.loaded {
		return
	}

	var prices []price.Price
	result := db.Find(&prices)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	pcache.pricesTree = make(map[string]*btree.BTree)

	for _, price := range prices {
		if pcache.pricesTree[price.CommodityName] == nil {
			pcache.pricesTree[price.CommodityName] = btree.New(2)
		}

		pcache.pricesTree[price.CommodityName].ReplaceOrInsert(price)
	}

	var postings []posting.Posting
	result = db.Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	for commodityName, postings := range lo.GroupBy(postings, func(p posting.Posting) string { return p.Commodity }) {
		if postings[0].Commodity != "INR" && pcache.pricesTree[commodityName] == nil {
			pcache.pricesTree[commodityName] = btree.New(2)
			for _, p := range postings {
				pcache.pricesTree[commodityName].ReplaceOrInsert(price.Price{Date: p.Date, CommodityID: p.Commodity, CommodityName: p.Commodity, Value: p.Amount / p.Quantity})
			}
		}
	}

	pcache.loaded = true
}

func GetMarketPrice(db *gorm.DB, p posting.Posting, date time.Time) float64 {
	loadPriceCache(db)
	if p.Commodity == "INR" {
		return p.Amount
	}

	pt := pcache.pricesTree[p.Commodity]
	if pt != nil {

		pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
		if pc.Value != 0 {
			return p.Quantity * pc.Value
		}
	} else {
		log.Info("Not found ", p)
	}

	return p.Amount
}
