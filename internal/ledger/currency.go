package ledger

// symbolToISO maps common currency symbols to their ISO 4217 codes.
// This is intentionally non-exhaustive — only the symbols paisa users encounter
// in their journals. Ambiguous mappings choose the most common (¥ → CNY since
// paisa's locale stack targets China; $ → USD).
//
// If you need different disambiguation (e.g. ¥ → JPY), add a `currency_aliases`
// field to paisa.yaml. For now this is hardcoded.
var symbolToISO = map[string]string{
	"¥": "CNY",
	"$": "USD",
	"€": "EUR",
	"£": "GBP",
	"₹": "INR",
}

// NormalizeCurrency returns the ISO 4217 code for a known currency symbol,
// or the input unchanged if it's already ISO (or unknown).
func NormalizeCurrency(s string) string {
	if iso, ok := symbolToISO[s]; ok {
		return iso
	}
	return s
}
