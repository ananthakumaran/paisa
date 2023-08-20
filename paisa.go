package main

import (
	"github.com/ananthakumaran/paisa/cmd"
	"github.com/shopspring/decimal"
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true
	cmd.Execute()
}
