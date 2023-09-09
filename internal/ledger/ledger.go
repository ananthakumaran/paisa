package ledger

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"os/exec"
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

	command := exec.Command(binary.LedgerBinaryPath(), "-f", journalPath, "balance")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err := command.Run()
	if err == nil {
		return errors, output.String(), nil
	}

	re := regexp.MustCompile(`(?m)While parsing file "[^"]+", line ([0-9]+):\s*\n(?:(?:While|>).*\n)*((?:.*\n)*?Error: .*\n)`)

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

	budgetPostings, err := execLedgerCommand(journalPath, []string{"--now", strconv.Itoa(time.Now().Year() + 3), "--budget"})
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

	command := exec.Command(binary.LedgerBinaryPath(), "-f", journalPath, "pricesdb")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err := command.Run()
	if err != nil {
		return prices, err
	}

	return parseLedgerPrices(output.String(), config.DefaultCurrency())
}

func (HLedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, string, error) {
	errors := []LedgerFileError{}
	_, err := exec.LookPath("hledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("hledger", "-f", journalPath, "--auto", "balance")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err == nil {
		return errors, output.String(), nil
	}

	re := regexp.MustCompile(`(?m)hledger: Error: [^:]*:([0-9:-]+)\n((?:.*\n)*)`)
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

	_, err := exec.LookPath("hledger")
	if err != nil {
		log.Fatal(err)
	}

	postings, err = execHLedgerCommand(journalPath, prices, []string{})
	if err != nil {
		return nil, err
	}

	timeRange := fmt.Sprintf("%d..%d", time.Now().Year()-3, time.Now().Year()+3)
	budgetPostings, err := execHLedgerCommand(journalPath, prices, []string{"--forecast=" + timeRange, "tag:_generated-transaction"})

	if err != nil {
		return nil, err
	}

	return append(postings, budgetPostings...), nil

}

func (HLedgerCLI) Prices(journalPath string) ([]price.Price, error) {
	var prices []price.Price

	_, err := exec.LookPath("hledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("hledger", "-f", journalPath, "--auto", "--infer-market-prices", "prices")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err != nil {
		return prices, err
	}

	return parseHLedgerPrices(output.String(), config.DefaultCurrency())
}

func parseLedgerPrices(output string, defaultCurrency string) ([]price.Price, error) {
	var prices []price.Price
	re := regexp.MustCompile(`P (\d{4}\/\d{2}\/\d{2}) (?:\d{2}:\d{2}:\d{2}) ([^\s\d.-]+|"[^"]+") (.+)\n`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		target, value, err := parseAmount(match[3])
		if err != nil {
			return nil, err
		}

		if target != defaultCurrency {
			continue
		}

		commodity := match[2]

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
	re := regexp.MustCompile(`P (\d{4}-\d{2}-\d{2}) ([^\s\d.-]+) (.+)\n`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		target, value, err := parseAmount(match[3])
		if err != nil {
			return nil, err
		}

		if target != defaultCurrency {
			continue
		}

		commodity := match[2]

		date, err := time.ParseInLocation("2006-01-02", match[1], time.Local)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price.Price{Date: date, CommodityName: commodity, CommodityID: commodity, CommodityType: config.Unknown, Value: value})

	}
	return prices, nil
}

func parseAmount(amount string) (string, decimal.Decimal, error) {
	match := regexp.MustCompile(`^(-?[0-9.,]+)([^\d,.-]+)|([^\d,.-]+)(-?[0-9.,]+)$`).FindStringSubmatch(amount)
	if match[1] == "" {
		value, err := decimal.NewFromString(strings.ReplaceAll(match[4], ",", ""))
		return strings.Trim(match[3], " "), value, err
	}

	value, err := decimal.NewFromString(strings.ReplaceAll(match[1], ",", ""))
	return strings.Trim(match[2], " "), value, err
}

func execLedgerCommand(journalPath string, flags []string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	args := append(append([]string{"-f", journalPath}, flags...), "csv", "--csv-format", "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(scrub(market(amount,date,'"+config.DefaultCurrency()+"') * 100000000))),%(quoted(xact.filename)),%(quoted(xact.id)),%(quoted(cleared ? \"*\" : (pending ? \"!\" : \"\"))),%(quoted(tag('Recurring'))),%(quoted(xact.beg_line)),%(quoted(xact.end_line))\n")

	command := exec.Command(binary.LedgerBinaryPath(), args...)
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err := command.Run()
	if err != nil {
		log.Fatal(error.String())
		return nil, err
	}

	// https://github.com/ledger/ledger/issues/2007
	fixedOutput := bytes.ReplaceAll(output.Bytes(), []byte(`\"`), []byte(`""`))
	reader := csv.NewReader(bytes.NewBuffer(fixedOutput))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(config.GetConfig().JournalPath)

	for _, record := range records {
		date, err := time.ParseInLocation("2006/01/02", record[0], time.Local)
		if err != nil {
			return nil, err
		}

		quantity, err := decimal.NewFromString(record[4])
		if err != nil {
			return nil, err
		}

		_, amount, err := parseAmount(record[5])
		if err != nil {
			return nil, err
		}
		amount = amount.Div(decimal.NewFromInt(100000000))
		if record[1] == "Budget transaction" {
			amount = amount.Neg()
		}

		namespace := uuid.Must(uuid.FromString("45964a1b-b24c-4a73-835a-9335a7aa7de5"))
		var transactionID string
		var fileName string
		var forecast bool

		if record[1] == "Budget transaction" || record[1] == "Forecast transaction" {
			transactionID = uuid.NewV5(namespace, record[0]+":"+record[1]).String()
			forecast = true
		} else {
			fileName, err = filepath.Rel(dir, record[6])
			if err != nil {
				return nil, err
			}

			transactionID = uuid.NewV5(namespace, fileName+":"+record[7]).String()
			forecast = false
		}

		var status string
		if record[8] == "*" {
			status = "cleared"
		} else if record[8] == "!" {
			status = "pending"
		} else {
			status = "unmarked"
		}

		var tagRecurring string
		if record[9] != "" {
			tagRecurring = record[9]
		}

		transactionBeginLine, err := strconv.ParseUint(record[10], 10, 64)
		if err != nil {
			return nil, err
		}

		transactionEndLine, err := strconv.ParseUint(record[11], 10, 64)
		if err != nil {
			return nil, err
		}

		posting := posting.Posting{
			Date:                 date,
			Payee:                record[1],
			Account:              record[2],
			Commodity:            record[3],
			Quantity:             quantity,
			Amount:               amount,
			TransactionID:        transactionID,
			Status:               status,
			TagRecurring:         tagRecurring,
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

	args := append([]string{"-f", journalPath, "--auto", "print", "-Ojson"}, flags...)

	command := exec.Command("hledger", args...)
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err := command.Run()
	if err != nil {
		log.Fatal(error.String())
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
						Quantity struct {
							Value float64 `json:"floatingPoint"`
						} `json:"aquantity"`
					} `json:"contents"`
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

	dir := filepath.Dir(config.GetConfig().JournalPath)

	for _, t := range transactions {
		forecast := false
		date, err := time.ParseInLocation("2006-01-02", t.Date, time.Local)
		if err != nil {
			return nil, err
		}

		for _, p := range t.Postings {
			amount := p.Amount[0]
			totalAmount := decimal.NewFromFloat(amount.Quantity.Value)

			if amount.Commodity != config.DefaultCurrency() {
				if amount.Price.Contents.Quantity.Value != 0 {
					totalAmount = decimal.NewFromFloat(amount.Price.Contents.Quantity.Value).Mul(decimal.NewFromFloat(amount.Quantity.Value))
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

			var tagRecurring string

			for _, tag := range t.Tags {
				if len(tag) == 2 {
					if tag[0] == "Recurring" {
						tagRecurring = tag[1]
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
				TransactionBeginLine: t.TSourcePos[0].SourceLine,
				TransactionEndLine:   t.TSourcePos[1].SourceLine,
				Forecast:             forecast,
				FileName:             fileName}
			postings = append(postings, &posting)

		}

	}

	return postings, nil
}
