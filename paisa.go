package main

import (
	"fmt"
	"os"

	"github.com/ananthakumaran/paisa/internal/ledger"
)

func main() {
	postings, _ := ledger.Parse(os.Args[1])
	fmt.Println(postings)
}
