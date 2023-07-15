package config

import (
	"time"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

type TaxCategoryType string

const (
	Debt           TaxCategoryType = "debt"
	Equity         TaxCategoryType = "equity"
	Equity65       TaxCategoryType = "equity65"
	Equity35       TaxCategoryType = "equity35"
	UnlistedEquity TaxCategoryType = "unlisted_equity"
)

type CommodityType string

const (
	MutualFund CommodityType = "mutualfund"
	NPS        CommodityType = "nps"
	Stock      CommodityType = "stock"
	Unknown    CommodityType = "unknown"
)

type Commodity struct {
	Name        string          `json:"name" yaml:"name"`
	Type        CommodityType   `json:"type" yaml:"type"`
	Code        string          `json:"code" yaml:"code"`
	Harvest     int             `json:"harvest" yaml:"harvest"`
	TaxCategory TaxCategoryType `json:"tax_category" yaml:"tax_category"`
}

type Retirement struct {
	SWR            float64  `json:"swr" yaml:"swr"`
	Expenses       []string `json:"expenses" yaml:"expenses"`
	Savings        []string `json:"savings" yaml:"savings"`
	YearlyExpenses float64  `json:"yearly_expenses" yaml:"yearly_expenses"`
}

type ScheduleAL struct {
	Code     string   `json:"code" yaml:"code"`
	Accounts []string `json:"accounts" yaml:"accounts"`
}

type AllocationTarget struct {
	Name     string   `json:"name" yaml:"name"`
	Target   float64  `json:"target" yaml:"target"`
	Accounts []string `json:"accounts" yaml:"accounts"`
}

type Config struct {
	JournalPath                string     `json:"journal_path" yaml:"journal_path"`
	DBPath                     string     `json:"db_path" yaml:"db_path"`
	LedgerCli                  string     `json:"ledger_cli" yaml:"ledger_cli"`
	DefaultCurrency            string     `json:"default_currency" yaml:"default_currency"`
	Locale                     string     `json:"locale" yaml:"locale"`
	FinancialYearStartingMonth time.Month `json:"financial_year_starting_month" yaml:"financial_year_starting_month"`

	Retirement Retirement `json:"retirement" yaml:"retirement"`

	ScheduleALs []ScheduleAL `json:"schedule_al" yaml:"schedule_al"`

	AllocationTargets []AllocationTarget `json:"allocation_targets" yaml:"allocation_targets"`

	Commodities []Commodity `json:"commodities" yaml:"commodities"`
}

var config Config

var defaultConfig = Config{
	LedgerCli:                  "ledger",
	DefaultCurrency:            "INR",
	Locale:                     "en-IN",
	Retirement:                 Retirement{SWR: 4, Savings: []string{"Assets:*"}, Expenses: []string{"Expenses:*"}, YearlyExpenses: 0},
	FinancialYearStartingMonth: 4,
}

func LoadConfig(content []byte) error {
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return err
	}

	err = mergo.Merge(&config, defaultConfig)

	if err != nil {
		return err
	}

	return nil
}

func GetConfig() Config {
	return config
}

func SetConfig(cfg Config) {
	config = cfg
}

func JournalPath() string {
	return config.JournalPath
}

func DefaultCurrency() string {
	return config.DefaultCurrency
}
