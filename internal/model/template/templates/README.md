# Builtin import templates

This directory was used by upstream Paisa to embed Handlebars-based
import templates for India-specific banks/brokers (Paytm, Zerodha,
ICICI, SBI, HDFC, NPS, Kuvera, 26AS, Value Research, IDFC, Mint).
They are removed for the China localization.

The directory must continue to exist so `go:embed all:templates` in
`internal/model/template/template.go` resolves. China importers will
be added via the `internal/importer/` plugin framework (issue #19),
not Handlebars templates.
