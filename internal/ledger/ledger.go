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
type Beancount struct{}

func Cli() Ledger {
	if config.GetConfig().LedgerCli == "hledger" {
		return HLedgerCLI{}
	}

	if config.GetConfig().LedgerCli == "beancount" {
		return Beancount{}
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
	args := []string{"--args-only"}
	if config.GetConfig().Strict == config.Yes {
		args = append(args, "--pedantic")
	}
	args = append(args, "-f", journalPath, "balance")
	err = utils.Exec(ledgerPath, &output, &error, args...)
	if err == nil {
		return errors, utils.Dos2Unix(output.String()), nil
	}

	re := regexp.MustCompile(`(?m)While parsing file "[^"]+", line ([0-9]+):\s*\n(?:(?:While|>).*\n)*((?:.*\n)*?Error: .*\n)`)

	matches := re.FindAllStringSubmatch(utils.Dos2Unix(error.String()), -1)

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
	err = utils.Exec(ledgerPath, &output, &error, "--args-only", "-f", journalPath, "pricesdb", "--pricedb-format", "P %(datetime) %(display_account) %(quantity(scrub(display_amount))) %(commodity(scrub(display_amount)))\n")
	if err != nil {
		log.Error(error.String())
		return prices, err
	}

	return parseLedgerPrices(utils.Dos2Unix(output.String()), config.DefaultCurrency())
}

func (HLedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, string, error) {
	errors := []LedgerFileError{}
	path, err := binary.LookPath("hledger")
	if err != nil {
		return errors, "", err
	}

	var output, error bytes.Buffer
	args := []string{"-f", journalPath, "--auto"}
	if config.GetConfig().Strict == config.Yes {
		args = append(args, "--strict")
	}
	args = append(args, "balance")
	err = utils.Exec(path, &output, &error, args...)
	if err == nil {
		return errors, utils.Dos2Unix(output.String()), nil
	}

	re := regexp.MustCompile(`(?m)hledger: Error: [^:]*:([0-9:-]+)\n((?:.*\n)*)`)
	matches := re.FindAllStringSubmatch(utils.Dos2Unix(error.String()), -1)

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

	commodities, err := parseHLedgerCommodities(journalPath)
	if err != nil {
		log.Error(err)
		return prices, err
	}

	commoditiesStyles := lo.Map(commodities, func(c string, _ int) string {
		return fmt.Sprintf(`--commodity-style="%s" 1,000.00`, c)
	})

	args := append([]string{"-f", journalPath, "--infer-market-prices", "--infer-costs", "prices"}, commoditiesStyles...)

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, args...)
	if err != nil {
		log.Error(error.String())
		return prices, err
	}

	return parseHLedgerPrices(utils.Dos2Unix(output.String()), config.DefaultCurrency())
}

func (Beancount) ValidateFile(journalPath string) ([]LedgerFileError, string, error) {
	errors := []LedgerFileError{}

	path, err := binary.LookPath("bean-check")
	if err != nil {
		return errors, "", err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, journalPath)
	if err == nil {

		path, err = binary.LookPath("bean-report")
		if err != nil {
			return errors, "", err
		}

		err = utils.Exec(path, &output, &error, journalPath, "bal")
		if err != nil {
			log.Error(error)
			return nil, "", err
		}
		return errors, utils.Dos2Unix(output.String()), nil
	}

	re := regexp.MustCompile(`(?:.*):([0-9]+):\s+(.+)`)

	lines := strings.Split(utils.Dos2Unix(error.String()), "\n")
	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			lastError := errors[len(errors)-1]
			lastError.Message = lastError.Message + "\n" + line
			errors[len(errors)-1] = lastError
		} else {
			lineno, _ := strconv.ParseUint(match[1], 10, 64)
			errors = append(errors, LedgerFileError{LineFrom: lineno, LineTo: lineno, Message: match[2]})
		}
	}
	return errors, "", err
}

func (Beancount) Parse(journalPath string, prices []price.Price) ([]*posting.Posting, error) {
	pricesTree := buildPricesTree(prices)

	type Range struct {
		Begin uint64
		End   uint64
	}

	transactionRanges := make(map[string]Range)

	var postings []*posting.Posting
	const (
		Date = iota
		Payee
		Narration
		Account
		Commodity
		Quantity
		Amount
		FileName
		Location
		TransactionID
		Status
		TagRecurring
		TagPeriod
	)
	args := []string{"-f", "csv", journalPath, "select date,payee,narration,account,currency,units(position),cost(position),filename,location,id,flag,ANY_META('recurring'),ANY_META('period')"}

	path, err := binary.LookPath("bean-query")
	if err != nil {
		return postings, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, args...)
	if err != nil {
		log.Error(error)
		return nil, err
	}

	reader := csv.NewReader(bytes.NewBuffer(output.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(config.GetJournalPath())

	locationRegex := regexp.MustCompile(`.*:(\d+):`)

	for _, record := range records[1:] {
		date, err := time.ParseInLocation("2006-01-02", record[Date], time.Local)
		if err != nil {
			return nil, err
		}

		_, quantity, err := parseAmount(strings.TrimSpace(record[Quantity]))
		if err != nil {
			return nil, err
		}

		costCurrency, amount, err := parseAmount(strings.TrimSpace(record[Amount]))
		if err != nil {
			return nil, err
		}

		if costCurrency != config.DefaultCurrency() {
			pr := lookupPrice(pricesTree, costCurrency, date)
			if !pr.Equal(decimal.Zero) {
				amount = amount.Mul(pr)
			}
		}

		fileName, err := filepath.Rel(dir, record[FileName])

		payee := strings.TrimSpace(record[Payee])
		narration := strings.TrimSpace(record[Narration])
		if narration != "" {
			if payee != "" {
				payee += " | "
			}
			payee += narration
		}

		var status string
		if record[Status] == "*" {
			status = "cleared"
		} else if record[Status] == "!" {
			status = "pending"
		} else {
			status = "unmarked"
		}

		match := locationRegex.FindStringSubmatch(strings.TrimSpace(record[Location]))
		if len(match) == 0 {
			return nil, fmt.Errorf("Could not parse location: %s", record[Location])
		}

		lineNumber, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return nil, err
		}

		r := transactionRanges[record[TransactionID]]
		if r.Begin == 0 || r.Begin > lineNumber {
			r.Begin = lineNumber
		}

		if r.End == 0 || r.End < lineNumber {
			r.End = lineNumber
		}
		transactionRanges[record[TransactionID]] = r

		posting := posting.Posting{
			Date:          date,
			Payee:         payee,
			Account:       strings.TrimSpace(record[Account]),
			Commodity:     utils.UnQuote(strings.TrimSpace(record[Commodity])),
			Quantity:      quantity,
			Amount:        amount,
			TransactionID: record[TransactionID],
			Status:        status,
			TagRecurring:  strings.TrimSpace(record[TagRecurring]),
			TagPeriod:     strings.TrimSpace(record[TagPeriod]),
			Forecast:      false,
			FileName:      fileName}
		postings = append(postings, &posting)

	}

	postings = lo.Map(postings, func(p *posting.Posting, _ int) *posting.Posting {
		r := transactionRanges[p.TransactionID]
		p.TransactionBeginLine = r.Begin - 1
		p.TransactionEndLine = r.End + 1
		return p
	})

	return postings, nil
}

func (Beancount) Prices(journalPath string) ([]price.Price, error) {
	var prices []price.Price
	path, err := binary.LookPath("bean-report")
	if err != nil {
		log.Error(err)
		return prices, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, journalPath, "pricesdb")
	if err != nil {
		log.Error(error.String())
		return prices, err
	}

	return parseBeancountPrices(utils.Dos2Unix(output.String()), config.DefaultCurrency())
}

func parseHLedgerCommodities(journalPath string) ([]string, error) {
	var commodities []string

	path, err := binary.LookPath("hledger")
	if err != nil {
		return commodities, err
	}

	var output, error bytes.Buffer
	err = utils.Exec(path, &output, &error, "-f", journalPath, "commodities")
	if err != nil {
		log.Error(error.String())
		return commodities, err
	}

	lines := strings.Split(utils.Dos2Unix(output.String()), "\n")

	for _, line := range lines {
		commodities = append(commodities, utils.UnQuote(strings.TrimSpace(line)))
	}

	return commodities, nil
}

func parseLedgerPrices(output string, defaultCurrency string) ([]price.Price, error) {
	var prices []price.Price
	re := regexp.MustCompile(`P (\d{4}\/\d{2}\/\d{2}) (?:\d{2}:\d{2}:\d{2}) ([^\s\d.-]+|"[^"]+") ([^\n]+)\n`)
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
	re := regexp.MustCompile(`P (\d{4}-\d{2}-\d{2}) ([^\s\d.-]+|"[^"]+") ([^\n]+)\n`)
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

func parseBeancountPrices(output string, defaultCurrency string) ([]price.Price, error) {
	var prices []price.Price
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}) price ([^ ]+)\s*([^\n]+)\n`)
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
		LotCommodity
		TagRecurring
		TagPeriod
		Note
		TransactionNote
	)
	args := append(append([]string{"--args-only", "-f", journalPath}, flags...), "csv", "--csv-format", "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(quantity(scrub(market(amount,date,'"+config.DefaultCurrency()+"') * 100000000)))),%(quoted(xact.filename)),%(quoted(xact.id)),%(quoted(cleared ? \"*\" : (pending ? \"!\" : \"\"))),%(quoted(xact.beg_line)),%(quoted(xact.end_line)),%(quoted(quantity(lot_price(amount)))),%(quoted(commodity(lot_price(amount)))),%(quoted(tag('Recurring'))),%(quoted(tag('Period'))),%(quoted(note)),%(quoted(xact.note))\n")

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
		if lotString != "" && lotString != "0" {
			lotAmount, err := decimal.NewFromString(record[LotPrice])
			if err != nil {
				return nil, err
			}

			lotCurrency := utils.UnQuote(record[LotCommodity])
			if lotCurrency == config.DefaultCurrency() {
				amount = lotAmount.Mul(quantity)
				amountAvailable = true
			}
		}

		if !amountAvailable {
			amount, err = decimal.NewFromString(record[Amount])
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

		note := record[Note]
		transactionNote := record[TransactionNote]
		if transactionNote != "" {
			note = utils.ReplaceLast(note, transactionNote, "")
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
			FileName:             fileName,
			Note:                 note,
			TransactionNote:      transactionNote}
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
		Comment     string     `json:"tcomment"`
		Tags        [][]string `json:"ttags"`
		TSourcePos  []struct {
			SourceColumn uint64 `json:"sourceColumn"`
			SourceLine   uint64 `json:"sourceLine"`
			SourceName   string `json:"sourceName"`
		} `json:"tsourcepos"`
		Postings []struct {
			Account string     `json:"paccount"`
			Comment string     `json:"pcomment"`
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

	pricesTree := buildPricesTree(prices)
	dir := filepath.Dir(config.GetJournalPath())

	for _, t := range transactions {
		forecast := false
		date, err := time.ParseInLocation("2006-01-02", t.Date, time.Local)
		if err != nil {
			return nil, err
		}

		for _, p := range t.Postings {
			// ignore balance assertions
			if len(p.Amount) == 0 {
				continue
			}
			amount := p.Amount[0]
			totalAmount := decimal.NewFromFloat(amount.Quantity.Value)
			totalAmountSet := false

			if amount.Commodity != config.DefaultCurrency() {
				if amount.Price.Contents.Quantity.Value != 0 {
					if amount.Price.Contents.Commodity != config.DefaultCurrency() {
						pr := lookupPrice(pricesTree, amount.Commodity, date)
						if !pr.Equal(decimal.Zero) {
							totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(pr)
							totalAmountSet = true
						}
						if !totalAmountSet {
							pr = lookupPrice(pricesTree, amount.Price.Contents.Commodity, date)
							if !pr.Equal(decimal.Zero) {
								totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(decimal.NewFromFloat(amount.Price.Contents.Quantity.Value).Mul(pr))
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
					pr := lookupPrice(pricesTree, amount.Commodity, date)
					if !pr.Equal(decimal.Zero) {
						totalAmount = decimal.NewFromFloat(amount.Quantity.Value).Mul(pr)
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
				FileName:             fileName,
				Note:                 p.Comment,
				TransactionNote:      t.Comment}
			postings = append(postings, &posting)

		}

	}

	return postings, nil
}

func buildPricesTree(prices []price.Price) map[string]*btree.BTree {
	pricesTree := make(map[string]*btree.BTree)
	for _, price := range prices {
		if pricesTree[price.CommodityName] == nil {
			pricesTree[price.CommodityName] = btree.New(2)
		}

		pricesTree[price.CommodityName].ReplaceOrInsert(price)
	}

	return pricesTree
}

func lookupPrice(pricesTree map[string]*btree.BTree, commodity string, date time.Time) decimal.Decimal {
	pt := pricesTree[commodity]
	if pt != nil {
		pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
		if !pc.Value.Equal(decimal.Zero) {
			return pc.Value
		}
	}

	return decimal.Zero
}
