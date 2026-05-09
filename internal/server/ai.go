package server

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/server/goal"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ── Types ────────────────────────────────────────────────────────────────────

type UnknownTransaction struct {
	TransactionID  string          `json:"transaction_id"`
	Date           string          `json:"date"`
	Payee          string          `json:"payee"`
	Amount         decimal.Decimal `json:"amount"`
	Commodity      string          `json:"commodity"`
	CurrentAccount string          `json:"current_account"`
	AccountPrefix  string          `json:"account_prefix"`
	SuggestedAccount string        `json:"suggested_account"`
	FileName       string          `json:"file_name"`
	BeginLine      uint64          `json:"begin_line"`
	EndLine        uint64          `json:"end_line"`
}

type ClassificationItem struct {
	TransactionID string `json:"transaction_id"`
	OldAccount    string `json:"old_account"`
	NewAccount    string `json:"new_account"`
	NewPayee      string `json:"new_payee,omitempty"`
}

type ClassifyRequest struct {
	DryRun          bool                 `json:"dry_run"`
	Classifications []ClassificationItem `json:"classifications"`
}

type DryRunDiff struct {
	FileName  string `json:"file_name"`
	OldLine   string `json:"old_line"`
	NewLine   string `json:"new_line"`
	LineNumber uint64 `json:"line_number"`
}

type AddTransactionPosting struct {
	Account string `json:"account"`
	Amount  string `json:"amount,omitempty"`
}

type AddTransactionRequest struct {
	File     string                  `json:"file"`
	Date     string                  `json:"date"`
	Payee    string                  `json:"payee"`
	Postings []AddTransactionPosting `json:"postings"`
}

type AccountSuggestion struct {
	Account    string  `json:"account"`
	Confidence float64 `json:"confidence"`
}

type SpendingTrend struct {
	Account        string          `json:"account"`
	Monthly        map[string]decimal.Decimal `json:"monthly"`
	RollingAverage decimal.Decimal `json:"rolling_average_3m"`
	Trend          string          `json:"trend"`
}

type SpendingAnomaly struct {
	TransactionID string          `json:"transaction_id"`
	Date          string          `json:"date"`
	Payee         string          `json:"payee"`
	Account       string          `json:"account"`
	Amount        decimal.Decimal `json:"amount"`
	Median        decimal.Decimal `json:"median"`
	ZScore        float64         `json:"z_score"`
}

type BudgetSetRequest struct {
	Account   string `json:"account"`
	Amount    string `json:"amount"`
	Commodity string `json:"commodity"`
	Period    string `json:"period"`
}

type HealthScore struct {
	Score             float64 `json:"score"`
	SavingsRate       float64 `json:"savings_rate"`
	BudgetAdherence   float64 `json:"budget_adherence"`
	EmergencyFundMonths float64 `json:"emergency_fund_months"`
	DebtRatio         float64 `json:"debt_ratio"`
	Recommendations   []string `json:"recommendations"`
}

type GoalProgress struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Current    decimal.Decimal `json:"current"`
	Target     decimal.Decimal `json:"target"`
	ProgressPct float64        `json:"progress_pct"`
}

type PortfolioRebalanceItem struct {
	Name    string          `json:"name"`
	Target  decimal.Decimal `json:"target_pct"`
	Current decimal.Decimal `json:"current_pct"`
	Delta   decimal.Decimal `json:"delta_amount"`
}

// ── GET /api/ai/transactions/unknown ─────────────────────────────────────────

func GetUnknownTransactions(db *gorm.DB) gin.H {
	postings := query.Init(db).All()
	tfIdfResult := prediction.GetTfIdf(db)
	vectors, _ := tfIdfResult["tf_idf"].(map[string]map[string]float64)
	idx, _ := tfIdfResult["index"].(map[string]interface{})
	_ = idx

	var unknowns []UnknownTransaction
	seen := map[string]bool{}

	for _, p := range postings {
		if !strings.HasSuffix(p.Account, ":Unknown") {
			continue
		}
		if seen[p.TransactionID] {
			continue
		}
		seen[p.TransactionID] = true

		parts := strings.SplitN(p.Account, ":", 2)
		prefix := parts[0]

		suggested := suggestAccount(p, prefix, vectors)

		unknowns = append(unknowns, UnknownTransaction{
			TransactionID:    p.TransactionID,
			Date:             p.Date.Format("2006/01/02"),
			Payee:            p.Payee,
			Amount:           p.Amount,
			Commodity:        p.Commodity,
			CurrentAccount:   p.Account,
			AccountPrefix:    prefix,
			SuggestedAccount: suggested,
			FileName:         p.FileName,
			BeginLine:        p.TransactionBeginLine,
			EndLine:          p.TransactionEndLine,
		})
	}

	return gin.H{"unknown_transactions": unknowns}
}

// ── POST /api/ai/transactions/classify ───────────────────────────────────────

func ClassifyTransactions(db *gorm.DB, req ClassifyRequest) gin.H {
	if config.GetConfig().Readonly {
		return gin.H{"error": "Readonly mode", "updated": 0}
	}

	// Group classifications by transaction ID for quick lookup
	byTxn := map[string]ClassificationItem{}
	for _, c := range req.Classifications {
		byTxn[c.TransactionID] = c
	}

	// Fetch all relevant postings to find file names and line ranges
	allPostings := query.Init(db).All()

	// Group postings by file
	type filePatch struct {
		fileName  string
		beginLine uint64
		endLine   uint64
		item      ClassificationItem
	}
	var patches []filePatch

	seen := map[string]bool{}
	for _, p := range allPostings {
		item, ok := byTxn[p.TransactionID]
		if !ok || seen[p.TransactionID] {
			continue
		}
		seen[p.TransactionID] = true
		patches = append(patches, filePatch{
			fileName:  p.FileName,
			beginLine: p.TransactionBeginLine,
			endLine:   p.TransactionEndLine,
			item:      item,
		})
	}

	// Group patches by file
	byFile := map[string][]filePatch{}
	for _, patch := range patches {
		byFile[patch.fileName] = append(byFile[patch.fileName], patch)
	}

	journalDir := filepath.Dir(config.GetJournalPath())

	var diffs []DryRunDiff
	updated := 0
	skipped := 0
	var errors []string

	for fileName, filePatches := range byFile {
		absPath := fileName
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(journalDir, fileName)
		}

		contentBytes, err := os.ReadFile(absPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot read %s: %v", fileName, err))
			skipped += len(filePatches)
			continue
		}

		lines := strings.Split(string(contentBytes), "\n")

		// Sort patches by begin line descending so line numbers stay valid
		sort.Slice(filePatches, func(i, j int) bool {
			return filePatches[i].beginLine > filePatches[j].beginLine
		})

		for _, patch := range filePatches {
			begin := int(patch.beginLine) - 1 // 0-indexed
			end := int(patch.endLine)          // exclusive

			if begin < 0 || end > len(lines) {
				errors = append(errors, fmt.Sprintf("line range out of bounds for txn %s", patch.item.TransactionID))
				skipped++
				continue
			}

			txnLines := make([]string, end-begin)
			copy(txnLines, lines[begin:end])

			newLines, changed := applyPatch(txnLines, patch.item, &diffs, fileName, uint64(begin))
			if changed {
				lines = append(lines[:begin], append(newLines, lines[end:]...)...)
				updated++
			} else {
				skipped++
			}
		}

		if !req.DryRun {
			result := SaveFile(db, LedgerFile{
				Name:      filepath.Base(fileName),
				Content:   strings.Join(lines, "\n"),
				Operation: "update",
			})
			if saved, ok := result["saved"].(bool); !ok || !saved {
				msg := "unknown error"
				if m, ok := result["message"].(string); ok {
					msg = m
				}
				errors = append(errors, fmt.Sprintf("save failed for %s: %s", fileName, msg))
			}
		}
	}

	return gin.H{
		"updated":       updated,
		"skipped":       skipped,
		"dry_run":       req.DryRun,
		"dry_run_diffs": diffs,
		"errors":        errors,
	}
}

func applyPatch(lines []string, item ClassificationItem, diffs *[]DryRunDiff, fileName string, baseLineOffset uint64) ([]string, bool) {
	changed := false
	accountRegex := regexp.MustCompile(
		`^((?:\t|\s{2})\s*)` + regexp.QuoteMeta(item.OldAccount) + `((?:\t|\s{2}).*|\s*)$`,
	)
	payeeRegex := regexp.MustCompile(`^(\d{4}[/-]\d{2}[/-]\d{2})\s+(.+)$`)

	result := make([]string, len(lines))
	copy(result, lines)

	for i, line := range result {
		if item.OldAccount != "" {
			if m := accountRegex.FindStringSubmatchIndex(line); m != nil {
				newLine := line[:m[3]] + item.NewAccount + line[m[4]:]
				*diffs = append(*diffs, DryRunDiff{
					FileName:   fileName,
					OldLine:    line,
					NewLine:    newLine,
					LineNumber: baseLineOffset + uint64(i) + 1,
				})
				result[i] = newLine
				changed = true
				continue
			}
		}
		if item.NewPayee != "" && i == 0 {
			if m := payeeRegex.FindStringSubmatch(line); m != nil {
				newLine := m[1] + " " + item.NewPayee
				*diffs = append(*diffs, DryRunDiff{
					FileName:   fileName,
					OldLine:    line,
					NewLine:    newLine,
					LineNumber: baseLineOffset + uint64(i) + 1,
				})
				result[i] = newLine
				changed = true
			}
		}
	}
	return result, changed
}

// ── POST /api/ai/transactions/add ────────────────────────────────────────────

func AddTransaction(db *gorm.DB, req AddTransactionRequest) gin.H {
	if config.GetConfig().Readonly {
		return gin.H{"error": "Readonly mode", "saved": false}
	}

	if req.Date == "" || req.Payee == "" || len(req.Postings) == 0 {
		return gin.H{"error": "date, payee, and postings are required", "saved": false}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s\n", req.Date, req.Payee))
	for _, p := range req.Postings {
		if p.Amount != "" {
			sb.WriteString(fmt.Sprintf("    %-40s  %s\n", p.Account, p.Amount))
		} else {
			sb.WriteString(fmt.Sprintf("    %s\n", p.Account))
		}
	}
	newTxn := sb.String()

	journalDir := filepath.Dir(config.GetJournalPath())
	targetFile := req.File
	if targetFile == "" {
		targetFile = filepath.Base(config.GetJournalPath())
	}

	absPath := filepath.Join(journalDir, targetFile)
	existing, err := os.ReadFile(absPath)
	if err != nil {
		return gin.H{"error": fmt.Sprintf("cannot read %s: %v", targetFile, err), "saved": false}
	}

	content := strings.TrimRight(string(existing), "\n") + "\n\n" + newTxn

	result := SaveFile(db, LedgerFile{
		Name:      targetFile,
		Content:   content,
		Operation: "update",
	})
	return result
}

// ── GET /api/ai/accounts/suggest ─────────────────────────────────────────────

func SuggestAccounts(db *gorm.DB, payee, amountStr, prefix string) gin.H {
	tfIdfResult := prediction.GetTfIdf(db)
	vectors, _ := tfIdfResult["tf_idf"].(map[string]map[string]float64)

	key := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%s %s", amountStr, payee), " "), "")
	tokens := tokenizeText(key)

	type scored struct {
		account string
		score   float64
	}
	var scores []scored

	for account, vec := range vectors {
		if prefix != "" && !strings.HasPrefix(strings.ToLower(account), strings.ToLower(prefix)) {
			continue
		}
		s := cosineSimilarity(tokens, vec)
		if s > 0 {
			scores = append(scores, scored{account, s})
		}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if len(scores) > 5 {
		scores = scores[:5]
	}

	suggestions := lo.Map(scores, func(s scored, _ int) AccountSuggestion {
		return AccountSuggestion{Account: s.account, Confidence: s.score}
	})
	return gin.H{"suggestions": suggestions}
}

// ── GET /api/ai/spending/trends ───────────────────────────────────────────────

func GetSpendingTrends(db *gorm.DB, months int) gin.H {
	if months <= 0 {
		months = 6
	}
	postings := query.Init(db).Like("Expenses:%").LastNMonths(months).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })

	var trends []SpendingTrend
	for _, account := range utils.SortedKeys(byAccount) {
		ps := byAccount[account]
		monthly := map[string]decimal.Decimal{}
		for _, p := range ps {
			key := p.Date.Format("2006-01")
			monthly[key] = monthly[key].Add(p.Amount)
		}

		vals := lo.Values(monthly)
		avg := decimal.Zero
		if len(vals) > 0 {
			sum := decimal.Zero
			for _, v := range vals {
				sum = sum.Add(v)
			}
			avg = sum.Div(decimal.NewFromInt(int64(len(vals))))
		}

		trend := "stable"
		keys := lo.Keys(monthly)
		sort.Strings(keys)
		if len(keys) >= 2 {
			last := monthly[keys[len(keys)-1]]
			prev := monthly[keys[len(keys)-2]]
			if last.GreaterThan(prev.Mul(decimal.NewFromFloat(1.1))) {
				trend = "up"
			} else if last.LessThan(prev.Mul(decimal.NewFromFloat(0.9))) {
				trend = "down"
			}
		}

		trends = append(trends, SpendingTrend{
			Account:        account,
			Monthly:        monthly,
			RollingAverage: avg,
			Trend:          trend,
		})
	}
	return gin.H{"trends": trends, "months": months}
}

// ── GET /api/ai/spending/anomalies ────────────────────────────────────────────

func GetSpendingAnomalies(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Expenses:%").All()

	byPayee := lo.GroupBy(postings, func(p posting.Posting) string { return p.Payee })

	var anomalies []SpendingAnomaly
	for _, ps := range byPayee {
		if len(ps) < 3 {
			continue
		}
		amounts := lo.Map(ps, func(p posting.Posting, _ int) float64 { return p.Amount.InexactFloat64() })
		median := medianFloat(amounts)
		stddev := stddevFloat(amounts, median)
		if stddev == 0 {
			continue
		}
		for _, p := range ps {
			z := math.Abs((p.Amount.InexactFloat64() - median) / stddev)
			if z > 2.0 {
				anomalies = append(anomalies, SpendingAnomaly{
					TransactionID: p.TransactionID,
					Date:          p.Date.Format("2006/01/02"),
					Payee:         p.Payee,
					Account:       p.Account,
					Amount:        p.Amount,
					Median:        decimal.NewFromFloat(median),
					ZScore:        z,
				})
			}
		}
	}
	sort.Slice(anomalies, func(i, j int) bool { return anomalies[i].ZScore > anomalies[j].ZScore })
	return gin.H{"anomalies": anomalies}
}

// ── GET /api/ai/budget/recommend ─────────────────────────────────────────────

func RecommendBudget(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Expenses:%").LastNMonths(3).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })

	type Recommendation struct {
		Account   string          `json:"account"`
		Average3M decimal.Decimal `json:"average_3m"`
		Suggested decimal.Decimal `json:"suggested"`
	}
	var recs []Recommendation
	for _, account := range utils.SortedKeys(byAccount) {
		ps := byAccount[account]
		total := accounting.CostSum(ps)
		avg := total.Div(decimal.NewFromInt(3))
		// Suggest 10% buffer over 3-month average
		suggested := avg.Mul(decimal.NewFromFloat(1.10)).Round(0)
		recs = append(recs, Recommendation{Account: account, Average3M: avg, Suggested: suggested})
	}
	return gin.H{"recommendations": recs}
}

// ── POST /api/ai/budget/set ───────────────────────────────────────────────────

func SetBudget(db *gorm.DB, req BudgetSetRequest) gin.H {
	if config.GetConfig().Readonly {
		return gin.H{"error": "Readonly mode", "saved": false}
	}
	if req.Account == "" || req.Amount == "" {
		return gin.H{"error": "account and amount are required", "saved": false}
	}
	period := req.Period
	if period == "" {
		period = "monthly"
	}
	commodity := req.Commodity
	if commodity == "" {
		commodity = config.GetConfig().DefaultCurrency
	}

	// Derive the periodic prefix (~ Monthly / ~ Yearly / ~ Weekly)
	var periodicLine string
	switch strings.ToLower(period) {
	case "yearly":
		periodicLine = "~ Yearly"
	case "weekly":
		periodicLine = "~ Weekly"
	default:
		periodicLine = "~ Monthly"
	}

	newBlock := fmt.Sprintf("\n%s\n    %-40s  %s %s\n    Assets:Checking\n",
		periodicLine, req.Account, req.Amount, commodity)

	journalDir := filepath.Dir(config.GetJournalPath())
	budgetFile := "budget.ledger"
	absPath := filepath.Join(journalDir, budgetFile)

	var existing string
	if b, err := os.ReadFile(absPath); err == nil {
		existing = string(b)
	}

	content := strings.TrimRight(existing, "\n") + newBlock

	result := SaveFile(db, LedgerFile{
		Name:      budgetFile,
		Content:   content,
		Operation: "overwrite",
	})
	return result
}

// ── GET /api/ai/recurring/upcoming ───────────────────────────────────────────

type UpcomingTransaction struct {
	Key      string          `json:"key"`
	Period   string          `json:"period"`
	Interval int             `json:"interval_days"`
	NextDate string          `json:"next_date"`
	Amount   decimal.Decimal `json:"amount"`
	Payee    string          `json:"payee"`
}

func GetUpcomingRecurring(db *gorm.DB, count int) gin.H {
	if count <= 0 {
		count = 10
	}
	postings := query.Init(db).All()
	sequences := ComputeRecurringTransactions(postings)

	now := utils.EndOfToday()
	var upcoming []UpcomingTransaction
	for _, seq := range sequences {
		if len(seq.Transactions) == 0 || seq.Interval == 0 {
			continue
		}
		last := seq.Transactions[0]
		nextDate := last.Date.AddDate(0, 0, seq.Interval)

		// compute typical amount from last transaction
		var amount decimal.Decimal
		for _, p := range last.Postings {
			if p.Amount.IsPositive() {
				amount = amount.Add(p.Amount)
			}
		}

		if nextDate.After(now) {
			upcoming = append(upcoming, UpcomingTransaction{
				Key:      seq.Key,
				Period:   seq.Period,
				Interval: seq.Interval,
				NextDate: nextDate.Format("2006/01/02"),
				Amount:   amount,
				Payee:    last.Payee,
			})
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].NextDate < upcoming[j].NextDate
	})
	if len(upcoming) > count {
		upcoming = upcoming[:count]
	}
	return gin.H{"upcoming": upcoming}
}

// ── GET /api/ai/cashflow/forecast ─────────────────────────────────────────────

func GetCashFlowForecast(db *gorm.DB, months int) gin.H {
	if months <= 0 {
		months = 3
	}
	checkingBalance := accounting.CostSum(query.Init(db).AccountPrefix("Assets:Checking").All())

	postings := query.Init(db).All()
	sequences := ComputeRecurringTransactions(postings)

	type MonthForecast struct {
		Month          string          `json:"month"`
		StartBalance   decimal.Decimal `json:"start_balance"`
		RecurringIn    decimal.Decimal `json:"recurring_income"`
		RecurringOut   decimal.Decimal `json:"recurring_expenses"`
		ProjectedBalance decimal.Decimal `json:"projected_balance"`
	}

	var forecasts []MonthForecast
	balance := checkingBalance
	now := utils.EndOfToday()

	for m := 1; m <= months; m++ {
		futureDate := now.AddDate(0, m, 0)
		monthKey := futureDate.Format("2006-01")

		var recurringIn, recurringOut decimal.Decimal
		for _, seq := range sequences {
			if seq.Interval == 0 || len(seq.Transactions) == 0 {
				continue
			}
			last := seq.Transactions[0]
			for _, p := range last.Postings {
				if strings.HasPrefix(p.Account, "Income:") {
					recurringIn = recurringIn.Add(p.Amount.Abs())
				} else if strings.HasPrefix(p.Account, "Expenses:") {
					recurringOut = recurringOut.Add(p.Amount.Abs())
				}
			}
		}

		start := balance
		balance = balance.Add(recurringIn).Sub(recurringOut)
		forecasts = append(forecasts, MonthForecast{
			Month:            monthKey,
			StartBalance:     start,
			RecurringIn:      recurringIn,
			RecurringOut:     recurringOut,
			ProjectedBalance: balance,
		})
	}

	return gin.H{"forecast": forecasts, "current_balance": checkingBalance}
}

// ── GET /api/ai/tax/summary ───────────────────────────────────────────────────

func GetTaxSummary(db *gorm.DB) gin.H {
	capitalGains := GetCapitalGains(db)
	harvest := GetHarvest(db)
	scheduleAL := GetScheduleAL(db)
	taxPaid := accounting.CostSum(query.Init(db).AccountPrefix("Expenses:Tax").All()).Abs()

	return gin.H{
		"tax_paid_ytd":    taxPaid,
		"capital_gains":   capitalGains,
		"harvest":         harvest,
		"schedule_al":     scheduleAL,
	}
}

// ── GET /api/ai/portfolio/rebalance ──────────────────────────────────────────

func GetPortfolioRebalance(db *gorm.DB) gin.H {
	allocationResult := GetAllocation(db)
	targets, _ := allocationResult["allocation_targets"].([]AllocationTarget)

	now := utils.EndOfToday()
	postings := query.Init(db).Like("Assets:%").All()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, now)
		return p
	})
	totalPortfolio := decimal.Zero
	for _, p := range postings {
		totalPortfolio = totalPortfolio.Add(p.MarketAmount)
	}

	var items []PortfolioRebalanceItem
	for _, t := range targets {
		currentPct := decimal.Zero
		if !totalPortfolio.IsZero() {
			currentPct = t.Current.Div(totalPortfolio).Mul(decimal.NewFromInt(100))
		}
		delta := t.Target.Sub(currentPct).Div(decimal.NewFromInt(100)).Mul(totalPortfolio)
		items = append(items, PortfolioRebalanceItem{
			Name:    t.Name,
			Target:  t.Target,
			Current: currentPct,
			Delta:   delta,
		})
	}
	return gin.H{"rebalance": items, "total_portfolio": totalPortfolio}
}

// ── GET /api/ai/health ────────────────────────────────────────────────────────

func GetHealthScore(db *gorm.DB) gin.H {
	cashFlows := GetCurrentCashFlow(db)

	var totalIncome, totalExpenses decimal.Decimal
	for _, cf := range cashFlows {
		totalIncome = totalIncome.Add(cf.Income)
		totalExpenses = totalExpenses.Add(cf.Expenses)
	}

	savingsRate := 0.0
	if !totalIncome.IsZero() {
		savingsRate = totalIncome.Sub(totalExpenses).Div(totalIncome).InexactFloat64() * 100
	}

	// Emergency fund: checking balance / avg monthly expenses
	checkingBalance := accounting.CostSum(query.Init(db).AccountPrefix("Assets:Checking").All())
	avgMonthlyExpenses := decimal.Zero
	if len(cashFlows) > 0 {
		sum := decimal.Zero
		for _, cf := range cashFlows {
			sum = sum.Add(cf.Expenses)
		}
		avgMonthlyExpenses = sum.Div(decimal.NewFromInt(int64(len(cashFlows))))
	}
	emergencyMonths := 0.0
	if !avgMonthlyExpenses.IsZero() {
		emergencyMonths = checkingBalance.Div(avgMonthlyExpenses).InexactFloat64()
	}

	// Debt ratio
	assets := accounting.CostSum(query.Init(db).Like("Assets:%").All())
	liabilities := accounting.CostSum(query.Init(db).Like("Liabilities:%").All()).Abs()
	debtRatio := 0.0
	if !assets.IsZero() {
		debtRatio = liabilities.Div(assets).InexactFloat64() * 100
	}

	// Budget adherence
	budgetResult := GetCurrentBudget(db)
	budgetsRaw, _ := budgetResult["budgets"]
	adherence := 100.0 // default if no budgets configured

	// score each dimension 0-100
	savingsScore := math.Min(100, math.Max(0, savingsRate/30.0*100))
	emergencyScore := math.Min(100, (emergencyMonths/6.0)*100)
	debtScore := math.Max(0, 100-debtRatio)
	_ = budgetsRaw

	score := (savingsScore + emergencyScore + debtScore + adherence) / 4.0

	var recs []string
	if savingsRate < 20 {
		recs = append(recs, fmt.Sprintf("Savings rate is %.1f%% — aim for 20%%+", savingsRate))
	}
	if emergencyMonths < 6 {
		recs = append(recs, fmt.Sprintf("Emergency fund covers %.1f months — aim for 6+", emergencyMonths))
	}
	if debtRatio > 40 {
		recs = append(recs, fmt.Sprintf("Debt-to-assets ratio is %.1f%% — consider paying down liabilities", debtRatio))
	}

	return gin.H{
		"health": HealthScore{
			Score:               score,
			SavingsRate:         savingsRate,
			BudgetAdherence:     adherence,
			EmergencyFundMonths: emergencyMonths,
			DebtRatio:           debtRatio,
			Recommendations:     recs,
		},
	}
}

// ── GET /api/ai/goals ─────────────────────────────────────────────────────────

func GetGoalProgress(db *gorm.DB) gin.H {
	summaries := goal.GetGoalSummaries(db)
	var progress []GoalProgress
	for _, s := range summaries {
		pct := 0.0
		if !s.Target.IsZero() {
			pct = s.Current.Div(s.Target).InexactFloat64() * 100
		}
		progress = append(progress, GoalProgress{
			Type:        s.Type,
			Name:        s.Name,
			Current:     s.Current,
			Target:      s.Target,
			ProgressPct: math.Min(100, pct),
		})
	}
	return gin.H{"goals": progress}
}

// ── GET /api/ai/summary ───────────────────────────────────────────────────────

func GetAISummary(db *gorm.DB) gin.H {
	networth := GetCurrentNetworth(db)
	cashFlows := GetCurrentCashFlow(db)

	var totalIncome, totalExpenses decimal.Decimal
	for _, cf := range cashFlows {
		totalIncome = totalIncome.Add(cf.Income)
		totalExpenses = totalExpenses.Add(cf.Expenses)
	}

	topExpenses := query.Init(db).Like("Expenses:%").LastNMonths(3).All()
	byAccount := lo.GroupBy(topExpenses, func(p posting.Posting) string { return p.Account })
	type accountSum struct {
		Account string
		Total   decimal.Decimal
	}
	var sums []accountSum
	for acc, ps := range byAccount {
		sums = append(sums, accountSum{acc, accounting.CostSum(ps)})
	}
	sort.Slice(sums, func(i, j int) bool { return sums[i].Total.GreaterThan(sums[j].Total) })
	if len(sums) > 10 {
		sums = sums[:10]
	}

	return gin.H{
		"networth":             networth,
		"income_3m":            totalIncome,
		"expenses_3m":          totalExpenses,
		"top_expense_accounts": sums,
	}
}

// ── Gin HTTP wrappers ─────────────────────────────────────────────────────────

func RegisterAIRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/api/ai/transactions/unknown", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetUnknownTransactions(db))
	})

	router.POST("/api/ai/transactions/classify", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(http.StatusOK, gin.H{"error": "Readonly mode"})
			return
		}
		var req ClassifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ClassifyTransactions(db, req))
	})

	router.POST("/api/ai/transactions/add", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(http.StatusOK, gin.H{"error": "Readonly mode"})
			return
		}
		var req AddTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, AddTransaction(db, req))
	})

	router.GET("/api/ai/accounts/suggest", func(c *gin.Context) {
		c.JSON(http.StatusOK, SuggestAccounts(db, c.Query("payee"), c.Query("amount"), c.Query("prefix")))
	})

	router.GET("/api/ai/spending/trends", func(c *gin.Context) {
		months := 0
		fmt.Sscanf(c.DefaultQuery("months", "6"), "%d", &months)
		c.JSON(http.StatusOK, GetSpendingTrends(db, months))
	})

	router.GET("/api/ai/spending/anomalies", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetSpendingAnomalies(db))
	})

	router.GET("/api/ai/budget/recommend", func(c *gin.Context) {
		c.JSON(http.StatusOK, RecommendBudget(db))
	})

	router.POST("/api/ai/budget/set", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(http.StatusOK, gin.H{"error": "Readonly mode"})
			return
		}
		var req BudgetSetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, SetBudget(db, req))
	})

	router.GET("/api/ai/recurring/upcoming", func(c *gin.Context) {
		count := 10
		fmt.Sscanf(c.DefaultQuery("count", "10"), "%d", &count)
		c.JSON(http.StatusOK, GetUpcomingRecurring(db, count))
	})

	router.GET("/api/ai/cashflow/forecast", func(c *gin.Context) {
		months := 3
		fmt.Sscanf(c.DefaultQuery("months", "3"), "%d", &months)
		c.JSON(http.StatusOK, GetCashFlowForecast(db, months))
	})

	router.GET("/api/ai/tax/summary", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetTaxSummary(db))
	})

	router.GET("/api/ai/portfolio/rebalance", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetPortfolioRebalance(db))
	})

	router.GET("/api/ai/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetHealthScore(db))
	})

	router.GET("/api/ai/goals", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetGoalProgress(db))
	})

	router.GET("/api/ai/networth", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetNetworth(db))
	})

	router.GET("/api/ai/summary", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetAISummary(db))
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func suggestAccount(p posting.Posting, prefix string, vectors map[string]map[string]float64) string {
	if vectors == nil {
		return ""
	}
	key := fmt.Sprintf("%s %s", p.Amount.String(), p.Payee)
	tokens := tokenizeText(key)

	best := ""
	bestScore := -1.0
	for account, vec := range vectors {
		if !strings.HasPrefix(strings.ToLower(account), strings.ToLower(prefix)) {
			continue
		}
		if strings.HasSuffix(account, ":Unknown") {
			continue
		}
		s := cosineSimilarity(tokens, vec)
		if s > bestScore {
			bestScore = s
			best = account
		}
	}
	return best
}

func tokenizeText(s string) map[string]float64 {
	re := regexp.MustCompile(`[ .()/:]+`)
	parts := re.Split(strings.ToLower(s), -1)
	counts := map[string]float64{}
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			counts[p]++
		}
	}
	return counts
}

func cosineSimilarity(query map[string]float64, vec map[string]float64) float64 {
	dot := 0.0
	for token, qw := range query {
		if vw, ok := vec[token]; ok {
			dot += qw * vw
		}
	}
	return dot
}

func medianFloat(vals []float64) float64 {
	s := make([]float64, len(vals))
	copy(s, vals)
	sort.Float64s(s)
	n := len(s)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (s[n/2-1] + s[n/2]) / 2
	}
	return s[n/2]
}

func stddevFloat(vals []float64, mean float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}
