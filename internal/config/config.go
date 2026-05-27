package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "embed"

	log "github.com/sirupsen/logrus"

	"dario.cat/mergo"
	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/ananthakumaran/paisa/internal/model/account"
	"gopkg.in/yaml.v3"
)

type CommodityType string

const (
	MutualFund CommodityType = "mutualfund"
	Stock      CommodityType = "stock"
	Unknown    CommodityType = "unknown"
)

type BoolType string

const (
	Yes BoolType = "yes"
	No  BoolType = "no"
)

type ImportTemplate struct {
	Name    string `json:"name" yaml:"name"`
	Content string `json:"content" yaml:"content"`
}

type Price struct {
	Provider string `json:"provider" yaml:"provider"`
	Code     string `json:"code" yaml:"code"`
}

type Commodity struct {
	Name  string        `json:"name" yaml:"name"`
	Type  CommodityType `json:"type" yaml:"type"`
	Price Price         `json:"price" yaml:"price"`
}

type Account struct {
	Name string              `json:"name" yaml:"name"`
	Icon string              `json:"icon" yaml:"icon"`
	Kind account.AccountKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type UserAccount struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type Goals struct {
	Retirement []RetirementGoal `json:"retirement" yaml:"retirement"`
	Savings    []SavingsGoal    `json:"savings" yaml:"savings"`
}

type RetirementGoal struct {
	Name           string   `json:"name" yaml:"name"`
	Icon           string   `json:"icon" yaml:"icon"`
	SWR            float64  `json:"swr" yaml:"swr"`
	Expenses       []string `json:"expenses" yaml:"expenses"`
	Savings        []string `json:"savings" yaml:"savings"`
	YearlyExpenses float64  `json:"yearly_expenses" yaml:"yearly_expenses"`
	Priority       int      `json:"priority" yaml:"priority"`
}

type SavingsGoal struct {
	Name             string   `json:"name" yaml:"name"`
	Icon             string   `json:"icon" yaml:"icon"`
	Target           float64  `json:"target" yaml:"target"`
	TargetDate       string   `json:"target_date" yaml:"target_date"`
	Rate             float64  `json:"rate" yaml:"rate"`
	PaymentPerPeriod float64  `json:"payment_per_period" yaml:"payment_per_period"`
	Accounts         []string `json:"accounts" yaml:"accounts"`
	Priority         int      `json:"priority" yaml:"priority"`
}

type Budget struct {
	Rollover BoolType `json:"rollover" yaml:"rollover"`
}

type AllocationTarget struct {
	Name     string   `json:"name" yaml:"name"`
	Target   float64  `json:"target" yaml:"target"`
	Accounts []string `json:"accounts" yaml:"accounts"`
}

type CreditCard struct {
	Account         string `json:"account" yaml:"account"`
	CreditLimit     int    `json:"credit_limit" yaml:"credit_limit"`
	StatementEndDay int    `json:"statement_end_day" yaml:"statement_end_day"`
	DueDay          int    `json:"due_day" yaml:"due_day"`
	Network         string `json:"network" yaml:"network"`
	Number          string `json:"number" yaml:"number"`
	ExpirationDate  string `json:"expiration_date" yaml:"expiration_date"`
}

type LiabilityKind string

const (
	AmortizingLoan LiabilityKind = "amortizing_loan"
)

type LiabilitySchedule string

const (
	LiabilityEqualPayment   LiabilitySchedule = "equal_payment"
	LiabilityEqualPrincipal LiabilitySchedule = "equal_principal"
)

type Liability struct {
	Name       string            `json:"name" yaml:"name"`
	Kind       LiabilityKind     `json:"kind,omitempty" yaml:"kind,omitempty"`
	Principal  float64           `json:"principal,omitempty" yaml:"principal,omitempty"`
	Rate       float64           `json:"rate,omitempty" yaml:"rate,omitempty"`
	StartDate  string            `json:"start_date,omitempty" yaml:"start_date,omitempty"`
	TermMonths int               `json:"term_months,omitempty" yaml:"term_months,omitempty"`
	Schedule   LiabilitySchedule `json:"schedule,omitempty" yaml:"schedule,omitempty"`
}

type Config struct {
	JournalPath           string       `json:"journal_path" yaml:"journal_path"`
	DBPath                string       `json:"db_path" yaml:"db_path"`
	SheetsDirectory       string       `json:"sheets_directory" yaml:"sheets_directory"`
	Readonly              bool         `json:"readonly" yaml:"readonly"`
	LedgerCli             string       `json:"ledger_cli" yaml:"ledger_cli"`
	DefaultCurrency       string       `json:"default_currency" yaml:"default_currency"`
	DisplayPrecision      int          `json:"display_precision" yaml:"display_precision"`
	AmountAlignmentColumn int          `json:"amount_alignment_column" yaml:"amount_alignment_column"`
	Locale                string       `json:"locale" yaml:"locale"`
	TimeZone              string       `json:"time_zone" yaml:"time_zone"`
	WeekStartingDay       time.Weekday `json:"week_starting_day" yaml:"week_starting_day"`
	Strict                BoolType     `json:"strict" yaml:"strict"`

	Budget Budget `json:"budget" yaml:"budget"`

	AllocationTargets []AllocationTarget `json:"allocation_targets" yaml:"allocation_targets"`

	Commodities []Commodity `json:"commodities" yaml:"commodities"`

	ImportTemplates []ImportTemplate `json:"import_templates" yaml:"import_templates"`

	Accounts []Account `json:"accounts" yaml:"accounts"`

	Goals Goals `json:"goals" yaml:"goals"`

	UserAccounts []UserAccount `json:"user_accounts" yaml:"user_accounts"`

	CreditCards []CreditCard `json:"credit_cards" yaml:"credit_cards"`

	Liabilities []Liability `json:"liabilities" yaml:"liabilities"`

	TransferAccounts []string `json:"transfer_accounts" yaml:"transfer_accounts"`
}

var config Config
var configPath string
var location *time.Location

var defaultConfig = Config{
	Readonly:              false,
	LedgerCli:             "ledger",
	DefaultCurrency:       "INR",
	DisplayPrecision:      0,
	AmountAlignmentColumn: 52,
	Locale:                "en-IN",
	TimeZone:              "",
	Budget:                Budget{Rollover: Yes},
	Strict:                No,
	WeekStartingDay:       0,
	AllocationTargets:     []AllocationTarget{},
	Commodities:           []Commodity{},
	ImportTemplates:       []ImportTemplate{},
	Accounts:              []Account{},
	Goals:                 Goals{Retirement: []RetirementGoal{}, Savings: []SavingsGoal{}},
	UserAccounts:          []UserAccount{},
	CreditCards:           []CreditCard{},
	Liabilities:           []Liability{},
	TransferAccounts:      []string{},
}

var itemsUniquePropertiesMeta = jsonschema.MustCompileString("itemsUniqueProperties.json", `{
  "properties": {
    "itemsUniqueProperties": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "minItems": 1
    }
  }
}`)

type itemsUniquePropertiesSchema []string
type itemsUniquePropertiessCompiler struct{}

func (itemsUniquePropertiessCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {

	if items, ok := m["itemsUniqueProperties"]; ok {
		itemsInterface := items.([]interface{})
		itemsString := make([]string, len(itemsInterface))
		for i, v := range itemsInterface {
			itemsString[i] = v.(string)
		}
		return itemsUniquePropertiesSchema(itemsString), nil
	}

	return nil, nil
}

func (s itemsUniquePropertiesSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	for _, uniqueProperty := range s {
		items := v.([]interface{})
		seen := make(map[string]bool)
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if _, ok := itemMap[uniqueProperty]; ok {
				value := itemMap[uniqueProperty].(string)
				if seen[value] {
					return ctx.Error("itemsUniqueProperty", "duplicate %s %s", uniqueProperty, value)
				}
				seen[value] = true
			}
		}
	}
	return nil
}

//go:embed schema.json
var SchemaJson string
var schema *jsonschema.Schema

func init() {
	c := jsonschema.NewCompiler()
	c.AssertFormat = true
	c.Draft = jsonschema.Draft2020
	c.RegisterExtension("itemsUniqueProperties", itemsUniquePropertiesMeta, itemsUniquePropertiessCompiler{})
	err := c.AddResource("schema.json", strings.NewReader(SchemaJson))
	if err != nil {
		log.Fatal(err)
	}

	schema = c.MustCompile("schema.json")
}

func SaveConfigObject(config Config) error {
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return SaveConfig(content)
}

func SaveConfig(content []byte) error {
	err := LoadConfig(content, "")
	if err != nil {
		return err
	}

	yamlContent, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, yamlContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

// stripDeprecatedKeys mutates the parsed-yaml tree to drop config
// keys that were removed in M2-A (India-specific tax / harvest /
// Schedule AL / CII / NPS / metal / in-mfapi). For every key (or
// whole commodity entry) that gets stripped, an INFO log line is
// emitted so users notice their config has stale fields.
//
// The schema has `additionalProperties: false` at root and inside
// `commodities[*]`, plus tight enums on `commodities[*].type` and
// `commodities[*].price.provider`, so without this scrubbing a legacy
// yaml would fail validation and crash the binary on first upgrade.
func stripDeprecatedKeys(parsed interface{}) {
	root, ok := parsed.(map[string]interface{})
	if !ok {
		return
	}

	for _, key := range []string{"schedule_al", "cii"} {
		if _, present := root[key]; present {
			log.Infof("Ignored deprecated config key '%s' (removed in M2-A)", key)
			delete(root, key)
		}
	}

	commodities, ok := root["commodities"].([]interface{})
	if !ok {
		return
	}

	// Two passes: (1) per-item key strip (harvest, tax_category);
	// (2) drop whole commodity entries whose `type` or `price.provider`
	// is no longer supported (nps, metal, in-mfapi, com-purifiedbytes-*).
	deprecatedTypes := map[string]bool{"nps": true, "metal": true}
	deprecatedProviders := map[string]bool{
		"in-mfapi":                true,
		"com-purifiedbytes-nps":   true,
		"com-purifiedbytes-metal": true,
	}

	kept := make([]interface{}, 0, len(commodities))
	for _, c := range commodities {
		item, ok := c.(map[string]interface{})
		if !ok {
			kept = append(kept, c)
			continue
		}
		name, _ := item["name"].(string)
		for _, key := range []string{"harvest", "tax_category"} {
			if _, present := item[key]; present {
				if name != "" {
					log.Infof("Ignored deprecated config key 'commodities[%s].%s' (removed in M2-A)", name, key)
				} else {
					log.Infof("Ignored deprecated config key 'commodities[].%s' (removed in M2-A)", key)
				}
				delete(item, key)
			}
		}

		commodityType, _ := item["type"].(string)
		var providerCode string
		if price, ok := item["price"].(map[string]interface{}); ok {
			providerCode, _ = price["provider"].(string)
		}

		if deprecatedTypes[commodityType] || deprecatedProviders[providerCode] {
			label := name
			if label == "" {
				label = "<unnamed>"
			}
			log.Infof("Dropping deprecated commodity '%s' (type=%q provider=%q removed in M2-A)", label, commodityType, providerCode)
			continue
		}

		kept = append(kept, item)
	}
	root["commodities"] = kept
}

func LoadConfigFile(path string) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		log.Warn("Failed to read config file: ", path)
		log.Fatal(err)
	}

	err = LoadConfig(content, path)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Using config file: ", path)
}

func LoadConfig(content []byte, cp string) error {
	var configJson interface{}
	err := yaml.Unmarshal(content, &configJson)
	if err != nil {
		return err
	}

	// M2-A removed the India-specific tax / harvest / Schedule AL stack
	// (and the in-mfapi / NPS / metal commodity surface). The schema has
	// `additionalProperties: false` at root and inside commodities[*],
	// so a legacy yaml that still carries any of those keys would fail
	// validation and crash `paisa update` / `paisa serve`.
	//
	// To keep the upgrade path painless, strip the deprecated keys from
	// the parsed structure BEFORE schema validation and emit an INFO log
	// so the user knows their config still loaded but the field has no
	// effect. We do NOT mutate the on-disk file — `paisa config save`
	// would, but plain reads stay non-destructive.
	stripDeprecatedKeys(configJson)

	err = schema.Validate(configJson)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid configuration\n%#v", err))
	}

	// Re-marshal the scrubbed tree so the typed unmarshal below sees the
	// post-strip view too. Otherwise nps / metal commodities (or the now-
	// dead in-mfapi / com-purifiedbytes-* providers) would still land in
	// config.Commodities and trip `Unknown price provider: ...` at sync
	// time.
	scrubbed, err := yaml.Marshal(configJson)
	if err != nil {
		return err
	}

	config = Config{}
	err = yaml.Unmarshal(scrubbed, &config)
	if err != nil {
		return err
	}

	err = mergo.Merge(&config, defaultConfig, mergo.WithOverrideEmptySlice)

	if err != nil {
		return err
	}

	if cp != "" && configPath == "" {
		configPath = cp
	}

	if config.TimeZone == "" {
		location = time.Local
	} else {
		location, err = time.LoadLocation(config.TimeZone)
		if err != nil {
			location = time.Local
			return errors.New(fmt.Sprintf("Invalid time zone: %s\n%#v", config.TimeZone, err))
		}
	}

	return nil
}

func GetConfig() Config {
	return config
}

func GetJournalPath() string {
	if !filepath.IsAbs(config.JournalPath) {
		return filepath.Join(GetConfigDir(), config.JournalPath)
	}

	return config.JournalPath
}

func GetSheetDir() string {
	if config.SheetsDirectory == "" {
		return filepath.Dir(GetJournalPath())
	}

	dir := config.SheetsDirectory
	if !filepath.IsAbs(config.SheetsDirectory) {
		dir = filepath.Join(GetConfigDir(), config.SheetsDirectory)
	}

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		log.Fatal("Failed to create sheets directory", err)
	}

	return dir
}

func GetDBPath() string {
	if !filepath.IsAbs(config.DBPath) {
		return filepath.Join(GetConfigDir(), config.DBPath)
	}

	return config.DBPath
}

func GetConfigDir() string {
	return filepath.Dir(configPath)
}

func GetConfigPath() string {
	return configPath
}

func GetSchema() any {
	var schemaObject any
	err := json.Unmarshal([]byte(SchemaJson), &schemaObject)
	if err != nil {
		log.Fatal(err)
	}
	return schemaObject
}

func EnsureLogFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(cacheDir, "paisa", "paisa.log")

	err = os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	return path, err
}

func DefaultCurrency() string {
	return config.DefaultCurrency
}

func TimeZone() *time.Location {
	if location != nil {
		return location
	}

	return time.Local
}
