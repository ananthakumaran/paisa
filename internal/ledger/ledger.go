package ledger

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/google/btree"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"encoding/json"

	"github.com/ananthakumaran/paisa/internal/binary"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
)

type LedgerFileError struct {
	LineFrom uint64 `json:"line_from"`
	LineTo   uint64 `json:"line_to"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

type Ledger interface {
	ValidateFile(journalPath string) ([]LedgerFileError, string, error)
	Parse(journalPath string, prices []price.Price) ([]*posting.Posting, error)
	Prices(jornalPath string) ([]price.Price, error)
}

type LedgerCLI struct{}
type HLedgerCLI struct{}

func Cli() Ledger {
	if config.GetConfig().LedgerCli == "hledger" {
		return HLedgerCLI{}
	}

	return LedgerCLI{}
}

func (LedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, string, error) {
	errors := []LedgerFileError{}

	ledgerPath, err := binary.LedgerBinaryPath()
	if err != nil {
		return errors, "", err
	}

	var output, error bytes.Buffer
	err = utils.Exec(ledgerPath, &output, &error, "--args-only", "-f", journalPath, "balance")
	if err == nil {
		return errors, output.String(), nil
	}

	re := regexp.MustCompile(`(?m)While parsing file "[^"]+", line ([0-9]+):\s*[\r\n]+(?:(?:While|>).*[\r\n]+)*((?:.*[\r\n]+)*?Error: .*[\r\n]+)`)

	matches := re.FindAllStringSubmatch(error.String(), -1)

	for _, match := range matches {
		line, _ := strconv.ParseUint(match[1], 10, 64)
		errors = append(errors, LedgerFileError{LineFrom: line, LineTo: line, Message: match[2]})
	}
	return errors, "", err
}

func (LedgerCLI) Parse(journalPath string, _prices []price.Price) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	postings, err := execLedgerCommand(journalPath, []string{})

	if err != nil {
		return nil, err
	}

	budgetPostings, err := execLedgerCommand(journalPath, []string{"--now", strconv.Itoa(utils.Now().Year() + 3), "--budget"})
	budgetPostings = lo.Filter(budgetPostings, func(p *posting.Posting, _ int) bool {
		return p.Payee == "Budget transaction"
	})

	if err != nil {
		return nil, err
	}

	return append(postings, budgetPostings...), nil
}

func (LedgerCLI) Prices(journalPath string) ([]price.Price, error) {
	var prices []price.Price

	ledgerPath, err := binary.LedgerBinaryPath()
	if err != nil {
		return prices, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(ledgerPath, &output, &error, "--args-only", "-f", journalPath, "pricesdb")
	if err != nil {
		return prices, err
	}

	return parseLedgerPrices(output.String(), config.DefaultCurrency())
}

func (HLedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, string, error) {
	errors := []LedgerFileError{}
	path, err := binary.LookPath("hledger")
	if err != nil {
		return errors, "", err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, "-f", journalPath, "--auto", "balance")
	if err == nil {
		return errors, output.String(), nil
	}

	re := regexp.MustCompile(`(?m)hledger: Error: [^:]*:([0-9:-]+)[\r\n]+((?:.*[\r\n]+)*)`)
	matches := re.FindAllStringSubmatch(error.String(), -1)

	for _, match := range matches {
		lineRange := match[1]
		var lineFrom uint64 = 1
		var lineTo uint64 = 1

		multiline := regexp.MustCompile(`^([0-9]+)-([0-9]+):?$`)
		if multiline.MatchString(lineRange) {
			lineMatch := multiline.FindStringSubmatch(lineRange)
			lineFrom, _ = strconv.ParseUint(lineMatch[1], 10, 64)
			lineTo, _ = strconv.ParseUint(lineMatch[2], 10, 64)
		} else {
			lineMatch := regexp.MustCompile(`^([0-9]+).*$`).FindStringSubmatch(lineRange)
			lineFrom, _ = strconv.ParseUint(lineMatch[1], 10, 64)
			lineTo = lineFrom
		}

		errors = append(errors, LedgerFileError{LineFrom: lineFrom, LineTo: lineTo, Message: match[2]})
	}

	return errors, "", err
}

func (HLedgerCLI) Parse(journalPath string, prices []price.Price) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	postings, err := execHLedgerCommand(journalPath, prices, []string{})
	if err != nil {
		return nil, err
	}

	timeRange := fmt.Sprintf("%d..%d", utils.Now().Year()-3, utils.Now().Year()+3)
	budgetPostings, err := execHLedgerCommand(journalPath, prices, []string{"--forecast=" + timeRange, "tag:_generated-transaction"})

	if err != nil {
		return nil, err
	}

	return append(postings, budgetPostings...), nil

}

func (HLedgerCLI) Prices(journalPath string) ([]price.Price, error) {
	var prices []price.Price

	path, err := binary.LookPath("hledger")
	if err != nil {
		log.Error(err)
		return prices, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, "-f", journalPath, "--auto", "--infer-market-prices", "--infer-costs", "prices")
	if err != nil {
		return prices, err
	}

	return parseHLedgerPrices(output.String(), config.DefaultCurrency())
}

func parseLedgerPrices(output string, defaultCurrency string) ([]price.Price, error) {
	var prices []price.Price
	re := regexp.MustCompile(`P (\d{4}\/\d{2}\/\d{2}) (?:\d{2}:\d{2}:\d{2}) ([^\s\d.-]+|"[^"]+") ([^\r\n]+)[\r\n]+`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		target, value, err := parseAmount(match[3])
		if err != nil {
			return nil, err
		}

		if target != defaultCurrency {
			continue
		}

		commodity := utils.UnQuote(match[2])

		date, err := time.ParseInLocation("2006/01/02", match[1], time.Local)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price.Price{Date: date, CommodityName: commodity, CommodityID: commodity, CommodityType: config.Unknown, Value: value})

	}
	return prices, nil
}

func parseHLedgerPrices(output string, defaultCurrency string) ([]price.Price, error) {
	var prices []price.Price
	re := regexp.MustCompile(`P (\d{4}-\d{2}-\d{2}) ([^\s\d.-]+|"[^"]+") ([^\r\n]+)[\r\n]+`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		target, value, err := parseAmount(match[3])
		if err != nil {
			return nil, err
		}

		commodity := utils.UnQuote(match[2])
		if target != defaultCurrency {
			if commodity == defaultCurrency && !value.Equal(decimal.Zero) {
				commodity = target
				target = defaultCurrency
				value = decimal.NewFromInt(1).Div(value)
			} else {
				continue
			}
		}

		date, err := time.ParseInLocation("2006-01-02", match[1], time.Local)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price.Price{Date: date, CommodityName: commodity, CommodityID: commodity, CommodityType: config.Unknown, Value: value})

	}
	return prices, nil
}

func parseAmount(amount string) (string, decimal.Decimal, error) {
	match := regexp.MustCompile(`^(-?[0-9.,]+)([^\d,.-]+|\s*"[^"]+")$|([^\d,.-]+|\s*"[^"]+"\s*)(-?[0-9.,]+)$`).FindStringSubmatch(amount)
	if len(match) == 0 {
		log.Errorf("Could not parse amount: <%s>", amount)
		return "", decimal.Zero, fmt.Errorf("Could not parse amount: <%s>", amount)
	}

	if match[1] != "" {
		value, err := decimal.NewFromString(strings.ReplaceAll(match[1], ",", ""))
		return utils.UnQuote(strings.Trim(match[2], " ")), value, err

	}
	value, err := decimal.NewFromString(strings.ReplaceAll(match[4], ",", ""))
	return utils.UnQuote(strings.Trim(match[3], " ")), value, err

}

func execLedgerCommand(journalPath string, flags []string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	const (
		Date = iota
		Payee
		Account
		Commodity
		Quantity
		Amount
		FileName
		SequenceID
		Status
		TransactionBeginLine
		TransactionEndLine
		LotPrice
		TagRecurring
		TagPeriod
	)
	args := append(append([]string{"--args-only", "-f", journalPath}, flags...), "csv", "--csv-format", "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(scrub(market(amount,date,'"+config.DefaultCurrency()+"') * 100000000))),%(quoted(xact.filename)),%(quoted(xact.id)),%(quoted(cleared ? \"*\" : (pending ? \"!\" : \"\"))),%(quoted(xact.beg_line)),%(quoted(xact.end_line)),%(quoted(lot_price(amount))),%(quoted(tag('Recurring'))),%(quoted(tag('Period')))\n")

	ledgerPath, err := binary.LedgerBinaryPath()
	if err != nil {
		return postings, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(ledgerPath, &output, &error, args...)
	if err != nil {
		log.Error(error)
		return nil, err
	}

	// https://github.com/ledger/ledger/issues/2007
	fixedOutput := bytes.ReplaceAll(output.Bytes(), []byte(`\"`), []byte(`""`))
	reader := csv.NewReader(bytes.NewBuffer(fixedOutput))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(config.GetJournalPath())

	for _, record := range records {
		date, err := time.ParseInLocation("2006/01/02", record[Date], time.Local)
		if err != nil {
			return nil, err
		}

		quantity, err := decimal.NewFromString(record[Quantity])
		if err != nil {
			return nil, err
		}

		var amount decimal.Decimal
		amountAvailable := false

		lotString := record[LotPrice]
		if lotString != "" {
			lotCurrency, lotAmount, err := parseAmount(record[LotPrice])
			if err != nil {
				return nil, err
			}

			if lotCurrency == config.DefaultCurrency() {
				amount = lotAmount.Mul(quantity)
				amountAvailable = true
			}
		}

		if !amountAvailable {
			_, amount, err = parseAmount(record[Amount])
			if err != nil {
				return nil, err
			}
			amount = amount.Div(decimal.NewFromInt(100000000))
		}

		if record[Payee] == "Budget transaction" {
			amount = amount.Neg()
		}

		namespace := uuid.Must(uuid.FromString("45964a1b-b24c-4a73-835a-9335a7aa7de5"))
		var transactionID string
		var fileName string
		var forecast bool

		if record[Payee] == "Budget transaction" || record[Payee] == "Forecast transaction" {
			transactionID = uuid.NewV5(namespace, record[Date]+":"+record[Payee]).String()
			forecast = true
		} else {
			fileName, err = filepath.Rel(dir, record[FileName])
			if err != nil {
				return nil, err
			}

			transactionID = uuid.NewV5(namespace, fileName+":"+record[SequenceID]).String()
			forecast = false
		}

		var status string
		if record[Status] == "*" {
			status = "cleared"
		} else if record[Status] == "!" {
			status = "pending"
		} else {
			status = "unmarked"
		}

		transactionBeginLine, err := strconv.ParseUint(record[TransactionBeginLine], 10, 64)
		if err != nil {
			return nil, err
		}

		transactionEndLine, err := strconv.ParseUint(record[TransactionEndLine], 10, 64)
		if err != nil {
			return nil, err
		}

		var tagRecurring string
		if record[TagRecurring] != "" {
			tagRecurring = record[TagRecurring]
		}

		var tagPeriod string
		if record[TagPeriod] != "" {
			tagPeriod = record[TagPeriod]
		}

		posting := posting.Posting{
			Date:                 date,
			Payee:                record[Payee],
			Account:              record[Account],
			Commodity:            utils.UnQuote(record[Commodity]),
			Quantity:             quantity,
			Amount:               amount,
			TransactionID:        transactionID,
			Status:               status,
			TagRecurring:         tagRecurring,
			TagPeriod:            tagPeriod,
			TransactionBeginLine: transactionBeginLine,
			TransactionEndLine:   transactionEndLine,
			Forecast:             forecast,
			FileName:             fileName}
		postings = append(postings, &posting)

	}

	return postings, nil
}

func execHLedgerCommand(journalPath string, prices []price.Price, flags []string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	path, err := binary.LookPath("hledger")
	if err != nil {
		return nil, err
	}

	args := append([]string{"-f", journalPath, "--auto", "print", "-Ojson"}, flags...)

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, args...)
	if err != nil {
		log.Error(error)
		return nil, err
	}

	type Transaction struct {
		Date        string     `json:"tdate"`
		Description string     `json:"tdescription"`
		ID          int64      `json:"tindex"`
		Status      string     `json:"tstatus"`
		Tags        [][]string `json:"ttags"`
		TSourcePos  []struct {
			SourceColumn uint64 `json:"sourceColumn"`
			SourceLine   uint64 `json:"sourceLine"`
			SourceName   string `json:"sourceName"`
		} `json:"tsourcepos"`
		Postings []struct {
			Account string     `json:"paccount"`
			Tags    [][]string `json:"ptags"`
			Amount  []struct {
				Commodity string `json:"acommodity"`
				Quantity  struct {
					Value float64 `json:"floatingPoint"`
				} `json:"aquantity"`
				Price struct {
					Contents struct {
						Commodity string `json:"acommodity"`
						Quantity  struct {
							Value float64 `json:"floatingPoint"`
						} `json:"aquantity"`
					} `json:"contents"`
					Tag string `json:"tag"`
				} `json:"aprice"`
			} `json:"pamount"`
		} `json:"tpostings"`
	}

	var transactions []Transaction
	err = json.Unmarshal(output.Bytes(), &transactions)
	if err != nil {
		return nil, err
	}

	pricesTree := make(map[string]*btree.BTree)
	for _, price := range prices {
		if pricesTree[price.CommodityName] == nil {
			pricesTree[price.CommodityName] = btree.New(2)
		}

		pricesTree[price.CommodityName].ReplaceOrInsert(price)
	}

	dir := filepath.Dir(config.GetJournalPath())

	for _, t := range transactions {
		forecast := false
		date, err := time.ParseInLocation("2006-01-02", t.Date, time.Local)
		if err != nil {
			return nil, err
		}

		for _, p := range t.Postings {
			amount := p.Amount[0]
			totalAmount := decimal.NewFromFloat(amount.Quantity.Value)
			totalAmountSet := false

			if amount.Commodity != config.DefaultCurrency() {
				if amount.Price.Contents.Quantity.Value != 0 {
					if amount.Price.Contents.Commodity != config.DefaultCurrency() {
						pt := pricesTree[amount.Commodity]
						if pt != nil {
							pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
							if !pc.Value.Equal(decimal.Zero) {
								totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(pc.Value)
								totalAmountSet = true
							}
						}

						if !totalAmountSet {
							pt = pricesTree[amount.Price.Contents.Commodity]
							if pt != nil {
								pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})

								if !pc.Value.Equal(decimal.Zero) {
									totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(decimal.NewFromFloat(amount.Price.Contents.Quantity.Value).Mul(pc.Value))
								}
							}
						}
					} else {
						if amount.Price.Tag == "TotalPrice" {
							totalAmount = decimal.NewFromFloat(amount.Price.Contents.Quantity.Value)
						} else {
							totalAmount = decimal.NewFromFloat(amount.Price.Contents.Quantity.Value).Mul(decimal.NewFromFloat(amount.Quantity.Value))
						}
					}
				} else {
					pt := pricesTree[amount.Commodity]
					if pt != nil {
						pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
						if !pc.Value.Equal(decimal.Zero) {
							totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(pc.Value)
						}
					}
				}
			}

			var tagRecurring, tagPeriod string

			for _, tag := range t.Tags {
				if len(tag) == 2 {
					if tag[0] == "Recurring" {
						tagRecurring = tag[1]
					}

					if tag[0] == "Period" {
						tagPeriod = tag[1]
					}

					if tag[0] == "_generated-transaction" {
						forecast = true
					}
				}
				break
			}

			for _, tag := range p.Tags {
				if len(tag) == 2 && tag[0] == "Recurring" {
					tagRecurring = tag[1]
				}

				if len(tag) == 2 && tag[0] == "Period" {
					tagPeriod = tag[1]
				}
				break
			}

			var fileName string
			if !forecast {
				fileName, err = filepath.Rel(dir, t.TSourcePos[0].SourceName)
				if err != nil {
					return nil, err
				}

			}

			posting := posting.Posting{
				Date:                 date,
				Payee:                t.Description,
				Account:              p.Account,
				Commodity:            amount.Commodity,
				Quantity:             decimal.NewFromFloat(amount.Quantity.Value),
				Amount:               totalAmount,
				TransactionID:        strconv.FormatInt(t.ID, 10),
				Status:               strings.ToLower(t.Status),
				TagRecurring:         tagRecurring,
				TagPeriod:            tagPeriod,
				TransactionBeginLine: t.TSourcePos[0].SourceLine,
				TransactionEndLine:   t.TSourcePos[1].SourceLine,
				Forecast:             forecast,
				FileName:             fileName}
			postings = append(postings, &posting)

		}

	}

	return postings, nil
}
