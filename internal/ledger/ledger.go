package ledger

import (
	"bytes"
	"encoding/csv"

	"os/exec"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
)

func Parse(journalPath string) ([]*posting.Posting, error) {
	var postings []*posting.Posting

	_, err := exec.LookPath("ledger")
	if err != nil {
		log.Fatal(err)
	}

	command := exec.Command("ledger", "-f", journalPath, "csv", "--csv-format", "%(quoted(date)),%(quoted(payee)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(to_int(scrub(market(amount,date) * 100000))))\n")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err = command.Run()
	if err != nil {
		log.Fatal(error.String())
		return nil, err
	}

	reader := csv.NewReader(&output)
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

		posting := posting.Posting{Date: date, Payee: record[1], Account: record[2], Commodity: record[3], Quantity: quantity, Amount: amount}
		postings = append(postings, &posting)

	}

	return postings, nil
}
