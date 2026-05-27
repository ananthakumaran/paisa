package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIssueSeverityField verifies the Issue struct exposes a Severity field
// distinct from Level. Severity uses semantic values (error/warning/info)
// while Level remains the bulma class hint (danger/warning/info).
func TestIssueSeverityField(t *testing.T) {
	issue := Issue{
		Level:    ERROR,
		Severity: SeverityError,
		Summary:  "test",
	}
	assert.Equal(t, SeverityError, issue.Severity)
	assert.Equal(t, Severity("error"), issue.Severity)
}

// TestRulesAssignSeverity walks every registered rule and confirms it sets a
// non-empty Severity. This is the structural guarantee for the frontend that
// every Issue can be grouped by severity.
func TestRulesAssignSeverity(t *testing.T) {
	assert.NotEmpty(t, rules, "expected at least one rule registered")
	for _, rule := range rules {
		assert.NotEmpty(t, rule.Issue.Severity,
			"rule %q must declare a Severity", rule.Issue.Summary)
		switch rule.Issue.Severity {
		case SeverityError, SeverityWarning, SeverityInfo:
		default:
			t.Errorf("rule %q has unexpected Severity %q",
				rule.Issue.Summary, rule.Issue.Severity)
		}
	}
}

// TestDebitEntryIsInfo locks in the M0-C behavior: refund-style debit entries
// in Expenses accounts (Expenses:* -<amount>) are a normal way to record
// 红冲/refunds in ledger format and must not be reported as errors.
func TestDebitEntryIsInfo(t *testing.T) {
	for _, rule := range rules {
		if rule.Issue.Summary == "Debit Entry" {
			assert.Equal(t, SeverityInfo, rule.Issue.Severity,
				"Debit Entry must be info severity (refunds are legitimate)")
			return
		}
	}
	t.Fatal("no Debit Entry rule found")
}

// TestNegativeBalanceIsError guards against accidentally demoting real
// problems. A negative asset balance is a structural integrity issue that
// users must fix.
func TestNegativeBalanceIsError(t *testing.T) {
	for _, rule := range rules {
		if rule.Issue.Summary == "Negative Balance" {
			assert.Equal(t, SeverityError, rule.Issue.Severity,
				"Negative Balance must remain error severity")
			return
		}
	}
	t.Fatal("no Negative Balance rule found")
}

// TestCreditEntryIsError confirms Income credit entries stay at error.
func TestCreditEntryIsError(t *testing.T) {
	for _, rule := range rules {
		if rule.Issue.Summary == "Credit Entry" {
			assert.Equal(t, SeverityError, rule.Issue.Severity)
			return
		}
	}
	t.Fatal("no Credit Entry rule found")
}
