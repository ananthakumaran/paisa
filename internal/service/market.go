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
	sync.Once
	pricesTree map[string]*btree.BTree
}

var pcache priceCache

func loadPriceCache(db *gorm.DB) {
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
		if !utils.IsCurrency(postings[0].Commodity) && pcache.pricesTree[commodityName] == nil {
			pcache.pricesTree[commodityName] = btree.New(2)
			for _, p := range postings {
				pcache.pricesTree[commodityName].ReplaceOrInsert(price.Price{Date: p.Date, CommodityID: p.Commodity, CommodityName: p.Commodity, Value: p.Amount / p.Quantity})
			}
		}
	}
}

func ClearPriceCache() {
	pcache = priceCache{}
}

func GetUnitPrice(db *gorm.DB, commodity string, date time.Time) price.Price {
	pcache.Do(func() { loadPriceCache(db) })

	pt := pcache.pricesTree[commodity]
	if pt == nil {
		log.Fatal("Price not found ", commodity)
	}

	return utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
}

func GetAllPrices(db *gorm.DB, commodity string) []price.Price {
	pcache.Do(func() { loadPriceCache(db) })

	pt := pcache.pricesTree[commodity]
	if pt == nil {
		log.Fatal("Price not found ", commodity)
	}
	return utils.BTreeToSlice[price.Price](pt)
}

func GetMarketPrice(db *gorm.DB, p posting.Posting, date time.Time) float64 {
	pcache.Do(func() { loadPriceCache(db) })

	if utils.IsCurrency(p.Commodity) {
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

func PopulateMarketPrice(db *gorm.DB, ps []posting.Posting) []posting.Posting {
	date := time.Now()
	return lo.Map(ps, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = GetMarketPrice(db, p, date)
		return p
	})
}
