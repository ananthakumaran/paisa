package ledger

import (
	"bytes"
	"encoding/csv"

	"io"
	"log"
	"os/exec"
	"strconv"
	"time"
)

type Posting struct {
	date      time.Time
	account   string
	commodity string
	quantity  float64
}

func Parse(journalPath string) ([]Posting, error) {
	var postings []Posting

	command := exec.Command("ledger", "-f", journalPath, "csv", "--csv-format", "%(quoted(date)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount))))\n")
	var output, error bytes.Buffer
	command.Stdout = &output
	command.Stderr = &error
	err := command.Run()
	if err != nil {
		log.Fatal(error.String())
		return postings, err
	}

	reader := csv.NewReader(&output)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return postings, err
		}

		date, err := time.Parse("2006/01/02", record[0])
		if err != nil {
			return postings, err
		}

		quantity, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return postings, err
		}

		posting := Posting{date: date, account: record[1], commodity: record[2], quantity: quantity}
		postings = append(postings, posting)
	}

	return postings, nil
}
