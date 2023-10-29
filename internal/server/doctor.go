package server

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Level string

const (
	WARN  Level = "warning"
	ERROR Level = "danger"
)

type Issue struct {
	Level       Level  `json:"level"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Details     string `json:"details"`
}

type Rule struct {
	Issue     Issue
	Predicate func(db *gorm.DB) []error
}

const DATE_FORMAT string = "02 Jan 2006"

var rules []Rule

func init() {
	rules = []Rule{
		{
			Issue: Issue{
				Level:       ERROR,
				Summary:     "Negative Balance",
				Description: "The running balance of an <b>asset</b> account is not supposed to go negative at any time. This issue typically happens due to incorrect transaction entries."},
			Predicate: ruleAssetRegisterNonNegative},
		{
			Issue: Issue{
				Level:       ERROR,
				Summary:     "Credit Entry",
				Description: "Income account should never have credit entry."},
			Predicate: ruleNonCreditAccount},
		{
			Issue: Issue{
				Level:       ERROR,
				Summary:     "Debit Entry",
				Description: "Expense Account should never have debit entry."},
			Predicate: ruleNonDebitAccount},
		{
			Issue: Issue{
				Level:       ERROR,
				Summary:     "Exchange Price Missing",
				Description: "Exchange price is missing for the commodity."},
			Predicate: ruleExchangePriceMissing},
		{
			Issue: Issue{
				Level:       WARN,
				Summary:     "Unit Price Mismatch",
				Description: "Unit price used in the journal doesn't match the price fetched from external system."},
			Predicate: ruleJournalPriceMismatch}}
}

func GetDiagnosis(db *gorm.DB) gin.H {
	issues := make([]Issue, 0)
	for _, rule := range rules {
		for _, error := range rule.Predicate(db) {
			issue := rule.Issue
			issue.Details = error.Error()
			issues = append(issues, issue)
		}
	}
	return gin.H{"issues": issues}
}

func ruleAssetRegisterNonNegative(db *gorm.DB) []error {
	errs := make([]error, 0)
	assets := query.Init(db).Like("Assets:%").All()
	for account, ps := range lo.GroupBy(assets, func(posting posting.Posting) string { return posting.Account }) {
		for _, balance := range accounting.Register(ps) {
			if balance.Quantity.LessThan(decimal.NewFromFloat(0.01).Neg()) {
				errs = append(errs, errors.New(fmt.Sprintf("<b>%s</b> account went negative (%.2f) on %s", account, balance.Quantity.InexactFloat64(), balance.Date.Format(DATE_FORMAT))))
				break
			}
		}
	}
	return errs
}

func ruleNonCreditAccount(db *gorm.DB) []error {
	errs := make([]error, 0)
	incomes := query.Init(db).Like("Income:%").NotLike("Income:CapitalGains:%").All()
	for _, p := range incomes {
		if p.Amount.GreaterThan(decimal.NewFromFloat(0.01)) {
			errs = append(errs, errors.New(fmt.Sprintf("<b>%.4f</b> got credited to <b>%s</b> on %s", p.Amount.InexactFloat64(), p.Account, p.Date.Format(DATE_FORMAT))))
		}
	}
	return errs
}

func ruleNonDebitAccount(db *gorm.DB) []error {
	errs := make([]error, 0)
	incomes := query.Init(db).Like("Expenses:%").All()
	for _, p := range incomes {
		if p.Amount.LessThan(decimal.NewFromFloat(0.01).Neg()) {
			errs = append(errs, errors.New(fmt.Sprintf("<b>%.4f</b> got debited from <b>%s</b> on %s", p.Amount.InexactFloat64(), p.Account, p.Date.Format(DATE_FORMAT))))
		}
	}
	return errs
}

func ruleExchangePriceMissing(db *gorm.DB) []error {
	errs := make([]error, 0)
	postings := query.Init(db).Desc().All()

	for _, p := range postings {
		if !utils.IsCurrency(p.Commodity) {
			externalPrice := service.GetUnitPrice(db, p.Commodity, p.Date)
			if externalPrice.CommodityName != p.Commodity {
				errs = append(errs, errors.New(fmt.Sprintf("Exchange price from <b>%s</b> to your default currency <b>%s</b> is not specified for posting %s", p.Commodity, config.DefaultCurrency(), formatPosting(p))))
			}
		}
	}
	return errs
}

func ruleJournalPriceMismatch(db *gorm.DB) []error {
	errs := make([]error, 0)
	postings := query.Init(db).Desc().All()
	for _, p := range postings {
		if !utils.IsCurrency(p.Commodity) {
			externalPrice := service.GetUnitPrice(db, p.Commodity, p.Date)
			diff := externalPrice.Value.Sub(p.Price()).Abs()
			if externalPrice.CommodityName == p.Commodity &&
				externalPrice.CommodityType != config.Unknown &&
				diff.GreaterThanOrEqual(decimal.NewFromFloat(0.0001)) {
				errs = append(errs, errors.New(fmt.Sprintf("The price specified in your posting %s doesn't match the price <b>%.4f</b> (%s) fetched from external system", formatPosting(p), externalPrice.Value.InexactFloat64(), externalPrice.Date.Format(DATE_FORMAT))))
			}
		}
	}
	return errs
}

func formatPosting(p posting.Posting) string {
	var price string
	if p.Quantity.Equal(p.Amount) {
		price = fmt.Sprintf("%.4f %s", p.Quantity.InexactFloat64(), p.Commodity)
	} else {
		price = fmt.Sprintf("%.4f %s @ %.4f %s", p.Quantity.InexactFloat64(), p.Commodity, p.Price().InexactFloat64(), config.DefaultCurrency())
	}

	postingUrl := fmt.Sprintf("/ledger/editor/%s#%d", url.PathEscape(p.FileName), p.TransactionBeginLine)
	return fmt.Sprintf("<a href=\"%s\"> %s\t%s\t%s</a>", postingUrl, p.Date.Format(DATE_FORMAT), p.Account, price)
}
