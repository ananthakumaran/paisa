package accounting

import (
	"sync"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type accountCache struct {
	sync.Once
	accounts []string
}

var acache accountCache

func loadAccountCache(db *gorm.DB) {
	db.Model(&posting.Posting{}).Distinct().Pluck("Account", &acache.accounts)
}

func all(db *gorm.DB) []string {
	acache.Do(func() { loadAccountCache(db) })
	return acache.accounts
}

func IsLeafAccount(db *gorm.DB, account string) bool {
	return slices.Contains(all(db), account)
}

func ClearCache() {
	acache = accountCache{}
}
