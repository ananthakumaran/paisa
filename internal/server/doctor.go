package server

import (
	"errors"
	"fmt"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
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
				Description: "Accounts like Expenses or Liabilities should never have debit entry."},
			Predicate: ruleNonDebitAccount}}
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
			if balance.Quantity < -0.01 {
				errs = append(errs, errors.New(fmt.Sprintf("<b>%s</b> account went negative on %s", account, balance.Date.Format(DATE_FORMAT))))
				break
			}
		}
	}
	return errs
}

func ruleNonCreditAccount(db *gorm.DB) []error {
	errs := make([]error, 0)
	incomes := query.Init(db).Like("Income:%").All()
	for _, p := range incomes {
		if p.Amount > 0.01 {
			errs = append(errs, errors.New(fmt.Sprintf("<b>%.4f</b> got credited to <b>%s</b> on %s", p.Amount, p.Account, p.Date.Format(DATE_FORMAT))))
		}
	}
	return errs
}

func ruleNonDebitAccount(db *gorm.DB) []error {
	errs := make([]error, 0)
	incomes := query.Init(db).Like("Expenses:%").OrLike("Liabilities:%").All()
	for _, p := range incomes {
		if p.Amount < -0.01 {
			errs = append(errs, errors.New(fmt.Sprintf("<b>%.4f</b> got debited from <b>%s</b> on %s", p.Amount, p.Account, p.Date.Format(DATE_FORMAT))))
		}
	}
	return errs
}
