package ledger

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLedgerErrors(t *testing.T) {
	unbalanced := `In file included from "/journal/main.ledger", line 1:
While parsing file "/journal/included file.ledger", line 3:
While balancing transaction from "/journal/included file.ledger", lines 1-3:
> 2026-10-18 Unbalanced example
>     Expenses:Food    100 INR
>     Assets:Bank      -90 INR
Unbalanced remainder is:
              10 INR
Amount to balance against:
             100 INR
Error: Transaction does not balance
`
	parseError := `While parsing file "/journal/main.ledger", line 12:
Error: Invalid date
`
	for _, tc := range []struct {
		name   string
		output string
	}{
		{name: "LF", output: unbalanced + parseError},
		{name: "CRLF", output: strings.ReplaceAll(unbalanced+parseError, "\n", "\r\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errors := parseLedgerErrors(utils.Dos2Unix(tc.output), "/journal/main.ledger")
			require.Len(t, errors, 2)
			assert.Equal(t, "included file.ledger", errors[0].FileName)
			assert.Equal(t, uint64(1), errors[0].LineFrom)
			assert.Equal(t, uint64(3), errors[0].LineTo)
			assert.Equal(t, "Unbalanced remainder is:\n              10 INR\nAmount to balance against:\n             100 INR\nError: Transaction does not balance\n", errors[0].Message)
			assert.Equal(t, LedgerFileError{
				FileName: "main.ledger",
				LineFrom: 12,
				LineTo:   12,
				Message:  "Error: Invalid date\n",
			}, errors[1])
		})
	}
}

func TestParseLedgerErrorRelativePaths(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal", "main.ledger")
	for _, tc := range []struct {
		name     string
		fileName string
		want     string
	}{
		{name: "main journal", fileName: journalPath, want: "main.ledger"},
		{name: "nested include", fileName: filepath.Join(dir, "journal", "transactions", "investment.ledger"), want: filepath.Join("transactions", "investment.ledger")},
		{name: "outside journal directory", fileName: filepath.Join(dir, "shared.ledger"), want: filepath.Join("..", "shared.ledger")},
		{name: "already relative", fileName: filepath.Join("transactions", "investment.ledger"), want: filepath.Join("transactions", "investment.ledger")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := "While parsing file \"" + tc.fileName + "\", line 3:\nError: Invalid date\n"
			errors := parseLedgerErrors(output, journalPath)
			require.Len(t, errors, 1)
			assert.Equal(t, tc.want, errors[0].FileName)
			assert.Equal(t, uint64(3), errors[0].LineFrom)
			assert.Equal(t, uint64(3), errors[0].LineTo)
		})
	}
}

func TestLedgerFileErrorString(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  LedgerFileError
		want string
	}{
		{
			name: "transaction range",
			err:  LedgerFileError{FileName: "investment.ledger", LineFrom: 10, LineTo: 13, Message: "Error: Transaction does not balance"},
			want: "investment.ledger:10-13\nError: Transaction does not balance",
		},
		{
			name: "single line",
			err:  LedgerFileError{FileName: "main.ledger", LineFrom: 8, LineTo: 8, Message: "Error: Invalid date"},
			want: "main.ledger:8\nError: Invalid date",
		},
		{
			name: "existing errors without filename",
			err:  LedgerFileError{LineFrom: 8, LineTo: 8, Message: "Original error"},
			want: "Original error",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.err.String())
		})
	}
}

func TestLedgerValidateUnbalancedIncludedFile(t *testing.T) {
	if _, err := exec.LookPath("ledger"); err != nil {
		t.Skip("ledger executable is not installed")
	}
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.ledger")
	includedPath := filepath.Join(dir, "unbalanced.ledger")
	require.NoError(t, os.WriteFile(mainPath, []byte("include unbalanced.ledger\n"), 0600))
	require.NoError(t, os.WriteFile(includedPath, []byte("2026-10-18 Unbalanced example\n    Expenses:Food    100 INR\n    Assets:Bank      -90 INR\n"), 0600))

	errors, _, err := (LedgerCLI{}).ValidateFile(mainPath)
	require.Error(t, err)
	require.Len(t, errors, 1)
	assert.Equal(t, "unbalanced.ledger", errors[0].FileName)
	assert.Equal(t, uint64(1), errors[0].LineFrom)
	assert.Equal(t, uint64(3), errors[0].LineTo)
	assert.Contains(t, errors[0].String(), "unbalanced.ledger:1-3\n")
	assert.Contains(t, errors[0].Message, "Error: Transaction does not balance")
}
