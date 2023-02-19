package ledger

import (
	"bytes"
	"encoding/csv"
	"regexp"

	"os/exec"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
)

type LedgerFileError struct {
	LineFrom uint64 `json:"line_from"`
	LineTo   uint64 `json:"line_to"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

func ValidateFile(journalPath string) ([]LedgerFileError, error) {
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

func Parse(journalPath string) ([]*posting.Posting, error) {
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
