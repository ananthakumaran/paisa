package liabilities

import (
	"strconv"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/loan"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Amortization is the per-liability schedule returned to the frontend.
// Only liabilities with kind == amortizing_loan get one.
type Amortization struct {
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Schedule       string          `json:"schedule"`
	Principal      decimal.Decimal `json:"principal"`
	APR            decimal.Decimal `json:"apr"`
	TermMonths     int             `json:"term_months"`
	StartDate      string          `json:"start_date,omitempty"`
	MonthlyRate    decimal.Decimal `json:"monthly_rate"`
	MonthlyPayment decimal.Decimal `json:"monthly_payment"`
	TotalPayment   decimal.Decimal `json:"total_payment"`
	TotalPrincipal decimal.Decimal `json:"total_principal"`
	TotalInterest  decimal.Decimal `json:"total_interest"`
	Months         []loan.Month    `json:"months"`
}

func GetRepayment(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Liabilities:%").Credit().All()
	postings = service.PopulateMarketPrice(db, postings)
	expenses := query.Init(db).Like("Expenses:Interest:%").All()
	postings = append(postings, expenses...)

	amortizations := buildAmortizations(config.GetConfig().Liabilities)
	return gin.H{"repayments": postings, "amortizations": amortizations}
}

// decimalFromYAMLFloat converts a YAML-decoded float64 to decimal.Decimal
// without round-tripping the IEEE-754 bits. We format the float with the
// shortest representation that exactly reproduces it (strconv 'g' / -1 prec),
// then parse it as decimal text. For typical YAML inputs like `1100000` or
// `4.9` this yields exactly `1100000` and `4.9` — no `1.0999999...` noise.
// Per CLAUDE.md "Don't introduce native floats for money", money never
// actually flows through float arithmetic; the float64 here is just the
// shape of the YAML AST.
func decimalFromYAMLFloat(v float64) decimal.Decimal {
	d, err := decimal.NewFromString(strconv.FormatFloat(v, 'g', -1, 64))
	if err != nil {
		// strconv 'g' / -1 round-trips every finite float64, so this should
		// never trigger; fall back to NewFromFloat to stay safe.
		return decimal.NewFromFloat(v)
	}
	return d
}

func buildAmortizations(liabs []config.Liability) []Amortization {
	result := make([]Amortization, 0)
	for _, l := range liabs {
		if l.Kind != config.AmortizingLoan {
			continue
		}
		schedule, err := loan.Amortize(
			decimalFromYAMLFloat(l.Principal),
			decimalFromYAMLFloat(l.Rate),
			l.TermMonths,
			loan.ScheduleKind(l.Schedule),
		)
		if err != nil {
			// Skip malformed liability rows; schema validation should have caught them.
			continue
		}
		result = append(result, Amortization{
			Name:           l.Name,
			Kind:           string(l.Kind),
			Schedule:       string(l.Schedule),
			Principal:      schedule.Principal,
			APR:            schedule.APR,
			TermMonths:     schedule.TermMonths,
			StartDate:      l.StartDate,
			MonthlyRate:    schedule.MonthlyRate,
			MonthlyPayment: schedule.MonthlyPayment,
			TotalPayment:   schedule.TotalPayment,
			TotalPrincipal: schedule.TotalPrincipal,
			TotalInterest:  schedule.TotalInterest,
			Months:         schedule.Months,
		})
	}
	return result
}
