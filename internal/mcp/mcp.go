package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ananthakumaran/paisa/internal/server"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"
)

// BuildMCPServer constructs a fully wired MCP server.
// Both the HTTP transport (mounted on Gin) and the stdio transport (paisa mcp)
// call this function with the same *gorm.DB — no network hop.
func BuildMCPServer(db *gorm.DB) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer(
		"Paisa Financial MCP",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithRecovery(),
	)

	// ── Transaction hygiene ──────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_unknown_transactions",
			mcpgo.WithDescription("List all transactions where the account is *:Unknown (e.g. Expenses:Unknown, Income:Unknown). Each item includes a TF-IDF-based suggested account."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetUnknownTransactions(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("classify_transactions",
			mcpgo.WithDescription("Rename the account (and optionally payee) for one or more transactions. Set dry_run=true to preview changes without writing. new_payee is optional."),
			mcpgo.WithBoolean("dry_run", mcpgo.Description("If true, return diffs without modifying files")),
			mcpgo.WithArray("classifications",
				mcpgo.Description("Array of {transaction_id, old_account, new_account, new_payee?}"),
				mcpgo.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"transaction_id": map[string]any{"type": "string"},
						"old_account":    map[string]any{"type": "string"},
						"new_account":    map[string]any{"type": "string"},
						"new_payee":      map[string]any{"type": "string"},
					},
					"required": []string{"transaction_id", "old_account", "new_account"},
				}),
			),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			var cr server.ClassifyRequest
			cr.DryRun = req.GetBool("dry_run", false)
			raw := req.GetArguments()["classifications"]
			if b, err := json.Marshal(raw); err == nil {
				_ = json.Unmarshal(b, &cr.Classifications)
			}
			return jsonResult(server.ClassifyTransactions(db, cr))
		},
	)

	s.AddTool(
		mcpgo.NewTool("add_transaction",
			mcpgo.WithDescription("Append a new ledger transaction to a file."),
			mcpgo.WithString("file", mcpgo.Description("Target ledger filename (relative), e.g. main.ledger")),
			mcpgo.WithString("date", mcpgo.Required(), mcpgo.Description("Date in YYYY/MM/DD format")),
			mcpgo.WithString("payee", mcpgo.Required(), mcpgo.Description("Transaction payee/description")),
			mcpgo.WithArray("postings",
				mcpgo.Required(),
				mcpgo.Description("Array of {account, amount?}"),
				mcpgo.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"account": map[string]any{"type": "string"},
						"amount":  map[string]any{"type": "string"},
					},
					"required": []string{"account"},
				}),
			),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			var ar server.AddTransactionRequest
			ar.File = req.GetString("file", "")
			ar.Date = req.GetString("date", "")
			ar.Payee = req.GetString("payee", "")
			raw := req.GetArguments()["postings"]
			if b, err := json.Marshal(raw); err == nil {
				_ = json.Unmarshal(b, &ar.Postings)
			}
			return jsonResult(server.AddTransaction(db, ar))
		},
	)

	s.AddTool(
		mcpgo.NewTool("suggest_account",
			mcpgo.WithDescription("Get top-5 account suggestions based on payee and amount using TF-IDF."),
			mcpgo.WithString("payee", mcpgo.Required(), mcpgo.Description("Transaction payee text")),
			mcpgo.WithString("amount", mcpgo.Description("Amount as string, e.g. '450'")),
			mcpgo.WithString("prefix", mcpgo.Description("Account prefix to filter by, e.g. 'Expenses'")),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			payee := req.GetString("payee", "")
			amount := req.GetString("amount", "")
			prefix := req.GetString("prefix", "")
			return jsonResult(server.SuggestAccounts(db, payee, amount, prefix))
		},
	)

	// ── Spending intelligence ────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_spending_trends",
			mcpgo.WithDescription("Month-over-month spending breakdown per Expenses:* account with rolling average and trend direction."),
			mcpgo.WithNumber("months", mcpgo.Description("Number of months to analyse (default 6)")),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			months := req.GetInt("months", 6)
			return jsonResult(server.GetSpendingTrends(db, months))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_spending_anomalies",
			mcpgo.WithDescription("Detect transactions where the amount for a payee deviates >2σ from historical median (potential fraud or billing errors)."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetSpendingAnomalies(db))
		},
	)

	// ── Budget management ────────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_budget",
			mcpgo.WithDescription("Current budget state: forecast vs actual spend per account."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetBudget(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("recommend_budget",
			mcpgo.WithDescription("Suggest budget amounts for each Expenses:* account based on 3-month averages (with 10% buffer)."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.RecommendBudget(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("set_budget",
			mcpgo.WithDescription("Write a new budget (periodic transaction) entry for an account."),
			mcpgo.WithString("account", mcpgo.Required(), mcpgo.Description("Expense account, e.g. Expenses:Dining")),
			mcpgo.WithString("amount", mcpgo.Required(), mcpgo.Description("Budget amount, e.g. 5000")),
			mcpgo.WithString("commodity", mcpgo.Description("Currency code, e.g. INR (defaults to config default)")),
			mcpgo.WithString("period", mcpgo.Description("monthly (default), yearly, or weekly")),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			var r server.BudgetSetRequest
			r.Account = req.GetString("account", "")
			r.Amount = req.GetString("amount", "")
			r.Commodity = req.GetString("commodity", "")
			r.Period = req.GetString("period", "")
			return jsonResult(server.SetBudget(db, r))
		},
	)

	// ── Cash flow & forecasting ──────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_upcoming_recurring",
			mcpgo.WithDescription("List upcoming scheduled transactions based on detected recurring patterns."),
			mcpgo.WithNumber("count", mcpgo.Description("Max items to return (default 10)")),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			count := req.GetInt("count", 10)
			return jsonResult(server.GetUpcomingRecurring(db, count))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_cashflow_forecast",
			mcpgo.WithDescription("Project checking account balance N months forward using recurring income/expense patterns."),
			mcpgo.WithNumber("months", mcpgo.Description("Months to forecast (default 3)")),
		),
		func(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			months := req.GetInt("months", 3)
			return jsonResult(server.GetCashFlowForecast(db, months))
		},
	)

	// ── Tax & compliance ─────────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_tax_summary",
			mcpgo.WithDescription("Composite tax view: tax paid YTD, capital gains (STCG/LTCG), tax-loss harvesting opportunities, and Schedule AL total."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetTaxSummary(db))
		},
	)

	// ── Investment & portfolio ───────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_portfolio_rebalance",
			mcpgo.WithDescription("Compare current allocation percentages against configured targets; returns per-category delta (how much to buy/sell)."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetPortfolioRebalance(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_gains",
			mcpgo.WithDescription("XIRR and gain breakdown per investment account."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetGain(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_capital_gains",
			mcpgo.WithDescription("Realized capital gains (STCG and LTCG) for the current financial year."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetCapitalGains(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_harvest_opportunities",
			mcpgo.WithDescription("Tax-loss harvesting opportunities — positions eligible for harvest with potential tax savings."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetHarvest(db))
		},
	)

	// ── Financial health ─────────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_health_score",
			mcpgo.WithDescription("Composite financial health score (0-100) covering savings rate, emergency fund, debt ratio, and budget adherence, with plain-language recommendations."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetHealthScore(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_goals",
			mcpgo.WithDescription("Progress towards configured financial goals (savings goals and retirement goals) with current vs target amounts."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetGoalProgress(db))
		},
	)

	// ── Broad context ────────────────────────────────────────────────────────

	s.AddTool(
		mcpgo.NewTool("get_financial_summary",
			mcpgo.WithDescription("Compact financial snapshot: net worth, 3-month income/expense totals, and top 10 expense accounts. Good as a first call to orient the agent."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetAISummary(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_networth",
			mcpgo.WithDescription("Full net worth timeline and overall XIRR."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetNetworth(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_transactions",
			mcpgo.WithDescription("All ledger transactions with their postings."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetTransactions(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_income_statement",
			mcpgo.WithDescription("Yearly income statement grouped by financial year."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetIncomeStatement(db))
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_diagnosis",
			mcpgo.WithDescription("Ledger diagnostic checks — unbalanced transactions, missing prices, etc."),
		),
		func(_ context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResult(server.GetDiagnosis(db))
		},
	)

	return s
}

// jsonResult serialises a gin.H map to a JSON MCP text result.
func jsonResult(data interface{}) (*mcpgo.CallToolResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("serialisation error: %v", err)), nil
	}
	return mcpgo.NewToolResultText(string(b)), nil
}
