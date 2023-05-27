package ledger

import (
	"bytes"
	"encoding/csv"
	"regexp"

	"os/exec"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/spf13/viper"
)

type LedgerFileError struct {
	LineFrom uint64 `json:"line_from"`
	LineTo   uint64 `json:"line_to"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

type Ledger interface {
	ValidateFile(journalPath string) ([]LedgerFileError, error)
	Parse(journalPath string) ([]*posting.Posting, error)
}

type LedgerCLI struct{}
type HLedgerCLI struct{}

func Cli() Ledger {
	if viper.GetString("ledger_cli") == "hledger" {
		return HLedgerCLI{}
	}

	return LedgerCLI{}
}

func (LedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, error) {
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
		return errors, nil
	}

	re := regexp.MustCompile(`(?m)While parsing file "[^"]+", line ([0-9]+):\s*\n(?:(?:While|>).*\n)*((?:.*\n)*?Error: .*\n)`)

	matches := re.FindAllStringSubmatch(error.String(), -1)

	for _, match := range matches {
		line, _ := strconv.ParseUint(match[1], 10, 64)
		errors = append(errors, LedgerFileError{LineFrom: line, LineTo: line, Message: match[2]})
	}
	return errors, err
}

func (LedgerCLI) Parse(journalPath string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	_, err := exec.LookPath("ledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("ledger", "-f", journalPath, "csv", "--csv-format", "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(to_int(scrub(market(amount,date) * 100000)))),%(quoted(xact.id))\n")
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
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		date, err := time.Parse("2006/01/02", record[0])
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

		posting := posting.Posting{Date: date, Payee: record[1], Account: record[2], Commodity: record[3], Quantity: quantity, Amount: amount, TransactionID: record[6]}
		postings = append(postings, &posting)

	}

	return postings, nil
}

func (HLedgerCLI) ValidateFile(journalPath string) ([]LedgerFileError, error) {
	errors := []LedgerFileError{}
	_, err := exec.LookPath("hledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("hledger", "-f", journalPath, "--auto", "check")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err == nil {
		return errors, nil
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
	return errors, err
}

func (HLedgerCLI) Parse(journalPath string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	_, err := exec.LookPath("hledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("hledger", "-f", journalPath, "--auto", "print", "-Ojson")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err != nil {
		log.Fatal(error.String())
		return nil, err
	}

	type Transaction struct {
		Date        string `json:"tdate"`
		Description string `json:"tdescription"`
		ID          int64  `json:"tindex"`
		Postings    []struct {
			Account string `json:"paccount"`
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

	for _, t := range transactions {
		date, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return nil, err
		}

		for _, p := range t.Postings {
			amount := p.Amount[0]
			totalAmount := amount.Quantity.Value
			if amount.Price.Contents.Quantity.Value != 0 {
				totalAmount = amount.Price.Contents.Quantity.Value * amount.Quantity.Value
			}
			posting := posting.Posting{Date: date, Payee: t.Description, Account: p.Account, Commodity: amount.Commodity, Quantity: amount.Quantity.Value, Amount: totalAmount, TransactionID: strconv.FormatInt(t.ID, 10)}
			postings = append(postings, &posting)

		}

	}

	return postings, nil
}
