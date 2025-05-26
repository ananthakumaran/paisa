package prediction

import (
	"regexp"
	"strings"

	"github.com/ananthakumaran/paisa/internal/model/rule"
	"github.com/ananthakumaran/paisa/internal/query"
	"gorm.io/gorm"
)

// TransactionDetails contains the details of a transaction for prediction
type TransactionDetails struct {
	Payee       string `json:"payee"`
	Description string `json:"description"`
	Account     string `json:"account"`
}

// PredictWithRules predicts the account for a transaction using rules
func PredictWithRules(db *gorm.DB, details TransactionDetails) (string, bool) {
	rules, err := rule.FindAll(db)
	if err != nil {
		return "", false
	}

	for _, r := range rules {
		if matchesRule(r.Pattern, details) {
			return r.Account, true
		}
	}

	return "", false
}

// matchesRule checks if the transaction details match the rule pattern
func matchesRule(pattern string, details TransactionDetails) bool {
	// Handle complex expressions with AND, OR operators
	if strings.Contains(pattern, " AND ") {
		parts := strings.Split(pattern, " AND ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !evaluateExpression(part, details) {
				return false
			}
		}
		return true
	} else if strings.Contains(pattern, " OR ") {
		parts := strings.Split(pattern, " OR ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if evaluateExpression(part, details) {
				return true
			}
		}
		return false
	}

	// Simple expression
	return evaluateExpression(pattern, details)
}

// evaluateExpression evaluates a single expression against transaction details
func evaluateExpression(expr string, details TransactionDetails) bool {
	// Handle grouped expressions
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return matchesRule(expr[1:len(expr)-1], details)
	}

	// Handle field-specific expressions
	if strings.HasPrefix(expr, "payee=") {
		value := strings.TrimPrefix(expr, "payee=")
		return matchFieldValue(value, details.Payee)
	} else if strings.HasPrefix(expr, "description=") {
		value := strings.TrimPrefix(expr, "description=")
		return matchFieldValue(value, details.Description)
	} else if strings.HasPrefix(expr, "account=") {
		value := strings.TrimPrefix(expr, "account=")
		return matchFieldValue(value, details.Account)
	}

	// Default to matching against all fields
	return matchFieldValue(expr, details.Payee) ||
		matchFieldValue(expr, details.Description) ||
		matchFieldValue(expr, details.Account)
}

// matchFieldValue checks if a field value matches a pattern
func matchFieldValue(pattern, value string) bool {
	// Handle regex patterns
	if strings.HasPrefix(pattern, "~/") && strings.HasSuffix(pattern, "/i") {
		// Case-insensitive regex
		regexStr := pattern[2 : len(pattern)-2]
		regex, err := regexp.Compile("(?i)" + regexStr)
		if err != nil {
			return false
		}
		return regex.MatchString(value)
	} else if strings.HasPrefix(pattern, "~/") && strings.HasSuffix(pattern, "/") {
		// Case-sensitive regex
		regexStr := pattern[2 : len(pattern)-1]
		regex, err := regexp.Compile(regexStr)
		if err != nil {
			return false
		}
		return regex.MatchString(value)
	}

	// Exact match
	return pattern == value
}