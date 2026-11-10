package service

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StockSplitUnits returns the number of units removed and added by
// the stock split transaction p is part of. A transaction which only
// adds or only removes units doesn't tell the split ratio and is not
// considered as a split.
func StockSplitUnits(db *gorm.DB, p posting.Posting) (removed decimal.Decimal, added decimal.Decimal, found bool) {
	if !IsStockSplit(db, p) {
		return decimal.Zero, decimal.Zero, false
	}

	t, _ := transaction.GetById(db, p.TransactionID)
	for _, tp := range t.Postings {
		if tp.Commodity != p.Commodity {
			return decimal.Zero, decimal.Zero, false
		}

		if tp.Quantity.IsPositive() {
			added = added.Add(tp.Quantity)
		} else {
			removed = removed.Add(tp.Quantity.Neg())
		}
	}

	return removed, added, removed.IsPositive() && added.IsPositive()
}

// IsBonus returns true if the units are received without giving up
// anything other than the same units from an income account. When the
// price is specified as zero, ledger and beancount omit the income
// posting altogether, while hledger includes it with zero quantity.
func IsBonus(db *gorm.DB, p posting.Posting) bool {
	if utils.IsCurrency(p.Commodity) || !p.Quantity.IsPositive() {
		return false
	}

	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return false
	}

	for _, tp := range t.Postings {
		if tp.Quantity.IsZero() {
			continue
		}

		if tp.Commodity != p.Commodity {
			return false
		}

		if tp.Quantity.IsNegative() && (!tp.HasBehaviour(posting.INCOME) || isTaxedOnReceipt(tp)) {
			return false
		}
	}
	return true
}

// isTaxedOnReceipt returns true for income which is taxed at the
// market value when it is received. The market value becomes the cost
// of the units.
func isTaxedOnReceipt(p posting.Posting) bool {
	return p.HasBehaviour(posting.INCOME_DIVIDEND) ||
		p.HasBehaviour(posting.INCOME_INTEREST) ||
		utils.IsSameOrParent(p.Account, "Income:Salary")
}

// SellAmount returns the amount for which the units are sold. If the
// lot is specified, the posting is valued at cost and the gain would
// be available as a separate capital gains posting.
func SellAmount(db *gorm.DB, p posting.Posting) decimal.Decimal {
	t, found := transaction.GetById(db, p.TransactionID)
	if !found {
		return p.Amount
	}

	balance := decimal.Zero
	gain := decimal.Zero
	amount := decimal.Zero
	quantity := decimal.Zero
	for _, tp := range t.Postings {
		balance = balance.Add(tp.Amount)
		if IsCapitalGains(tp) && CapitalGainsSourceAccount(tp.Account) == p.Account {
			gain = gain.Add(tp.Amount)
		} else if tp.Account == p.Account && tp.Commodity == p.Commodity && tp.Quantity.IsNegative() {
			amount = amount.Add(tp.Amount)
			quantity = quantity.Add(tp.Quantity)
		}
	}

	// if the transaction doesn't balance, then the posting is already
	// valued at the sell price.
	if gain.IsZero() || quantity.IsZero() || balance.Abs().GreaterThan(gain.Abs().Div(decimal.NewFromInt(2))) {
		return p.Amount
	}

	return amount.Add(gain).Mul(p.Quantity).Div(quantity)
}
