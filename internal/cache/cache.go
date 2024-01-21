package cache

import (
	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/service"
)

func Clear() {
	service.ClearInterestCache()
	service.ClearPriceCache()
	accounting.ClearCache()
	prediction.ClearCache()
	transaction.ClearCache()
}
