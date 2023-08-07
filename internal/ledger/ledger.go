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
	log "github.com/sirupsen/logrus"

	"encoding/json"

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
	_, err := exec.LookPath("ledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("ledger", "-f", journalPath, "balance")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
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

func ledgerCmd(args ...string) (records [][]string, err error) {
	command := exec.Command("ledger", args...)
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err != nil {
		log.Fatal(error.String())
		return nil, err
	}

	// https://github.com/ledger/ledger/issues/2007
	fixedOutput := bytes.ReplaceAll(output.Bytes(), []byte(`\"`), []byte(`""`))
	reader := csv.NewReader(bytes.NewBuffer(fixedOutput))
	records, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return
}

func (LedgerCLI) Parse(journalPath string, _prices []price.Price) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	_, err := exec.LookPath("ledger")
	if err != nil {
		log.Fatal(err)
	}

	csvFormat := "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(to_int(scrub(market(amount,date,'" + config.DefaultCurrency() + "') * 100000)))),%(quoted(xact.filename)),%(quoted(xact.id)),%(quoted(cleared ? \"*\" : (pending ? \"!\" : \"\"))),%(quoted(tag('Recurring'))),%(quoted(xact.beg_line)),%(quoted(xact.end_line))\n"
	records, err := ledgerCmd("-f", journalPath, "--forecast", "d>[1900]", "csv", "--csv-format", csvFormat)
	if err != nil {
		return nil, err
	}
	budgetRecords, err := ledgerCmd("-f", journalPath, "--budget", "csv", "@", "^Budget transaction$", "--csv-format", csvFormat)
	if err != nil {
		return nil, err
	}
	records = append(records, budgetRecords...)

	dir := filepath.Dir(config.GetConfig().JournalPath)

	for _, record := range records {
		date, err := time.ParseInLocation("2006/01/02", record[0], time.Local)
		if err != nil {
			return nil, err
		}

		quantity, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return nil, err
		}
		amount = amount / 100000

		fileName, err := filepath.Rel(dir, record[6])
		if err != nil {
			return nil, err
		}

		namespace := uuid.Must(uuid.FromString("45964a1b-b24c-4a73-835a-9335a7aa7de5"))
		transactionID := uuid.NewV5(namespace, fileName+":"+record[7]).String()

		var status string
		if record[8] == "*" {
			status = "cleared"
		} else if record[8] == "!" {
			status = "pending"
		} else {
			status = "unmarked"
		}

		var tagRecurring, tagBudgeting string
		if record[9] != "" {
			tagRecurring = record[9]
		}

		if record[1] == "Forecast transaction" {
			tagBudgeting = record[1]
		} else if record[1] == "Budget transaction" {
			tagBudgeting = record[1]
			quantity = -quantity
			amount = -amount
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
			TagBudgeting:         tagBudgeting,
			TransactionBeginLine: transactionBeginLine,
			TransactionEndLine:   transactionEndLine,
			FileName:             fileName}
		postings = append(postings, &posting)

	}

	return postings, nil
}

func (LedgerCLI) Prices(journalPath string) ([]price.Price, error) {
	var prices []price.Price

	_, err := exec.LookPath("ledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("ledger", "-f", journalPath, "pricesdb")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
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
	currentYear, _, _ := time.Now().Date()
	forecastYearEnd := currentYear + 5 // currentYear + config.GetConfig().Budgeting.EndYearDelta
	command := exec.Command("hledger", "-f", journalPath, "--auto", fmt.Sprintf("--forecast=..%d", forecastYearEnd), "print", "-Ojson")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
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
		date, err := time.ParseInLocation("2006-01-02", t.Date, time.Local)
		if err != nil {
			return nil, err
		}

		for _, p := range t.Postings {
			amount := p.Amount[0]
			totalAmount := amount.Quantity.Value

			if amount.Commodity != config.DefaultCurrency() {
				if amount.Price.Contents.Quantity.Value != 0 {
					totalAmount = amount.Price.Contents.Quantity.Value * amount.Quantity.Value
				} else {
					pt := pricesTree[amount.Commodity]
					if pt != nil {
						pc := utils.BTreeDescendFirstLessOrEqual(pt, price.Price{Date: date})
						if pc.Value != 0 {
							totalAmount = amount.Quantity.Value * pc.Value
						}
					}
				}
			}

			var tagRecurring, tagBudgeting string

			//? Does order matter here?
			for _, tag := range append(t.Tags, p.Tags...) {
				if len(tag) == 2 && tag[0] == "Recurring" {
					tagRecurring = tag[1]
					break
				} else if len(tag) == 2 && tag[0] == "_generated-transaction" {
					tagBudgeting = tag[1]
					break
				}
			}

			fileName, err := filepath.Rel(dir, t.TSourcePos[0].SourceName)
			if err != nil {
				return nil, err
			}

			posting := posting.Posting{
				Date:                 date,
				Payee:                t.Description,
				Account:              p.Account,
				Commodity:            amount.Commodity,
				Quantity:             amount.Quantity.Value,
				Amount:               totalAmount,
				TransactionID:        strconv.FormatInt(t.ID, 10),
				Status:               strings.ToLower(t.Status),
				TagRecurring:         tagRecurring,
				TagBudgeting:         tagBudgeting,
				TransactionBeginLine: t.TSourcePos[0].SourceLine,
				TransactionEndLine:   t.TSourcePos[1].SourceLine,
				FileName:             fileName}
			postings = append(postings, &posting)

		}

	}

	return postings, nil
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
	re := regexp.MustCompile(`P (\d{4}\/\d{2}\/\d{2}) (?:\d{2}:\d{2}:\d{2}) ([^\s\d.-]+) (.+)\n`)
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

func parseAmount(amount string) (string, float64, error) {
	match := regexp.MustCompile(`^(-?[0-9.,]+)([^\d,.-]+)|([^\d,.-]+)(-?[0-9.,]+)$`).FindStringSubmatch(amount)
	if match[1] == "" {
		value, err := strconv.ParseFloat(strings.ReplaceAll(match[4], ",", ""), 64)
		return strings.Trim(match[3], " "), value, err
	}

	value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
	return strings.Trim(match[2], " "), value, err
}
