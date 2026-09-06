package domain

import (
	"time"
)

// CategorySummary aggregates total amounts per category.
type CategorySummary struct {
	Category string `json:"category"`
	Amount   Money  `json:"amount"`
	Count    int    `json:"count"`
}

// MonthlySummary aggregates metrics for a specific month.
type MonthlySummary struct {
	Year             int                        `json:"year"`
	Month            time.Month                 `json:"month"`
	TotalIncome      Money                      `json:"total_income"`
	TotalExpense     Money                      `json:"total_expense"`
	NetBalance       Money                      `json:"net_balance"`
	TransactionCount int                        `json:"transaction_count"`
	IncomeByCategory map[string]CategorySummary `json:"income_by_category"`
	ExpenseByCategory map[string]CategorySummary `json:"expense_by_category"`
}

// CalculateMonthlySummary computes financial aggregates from a slice of transactions for a specific month.
func CalculateMonthlySummary(year int, month time.Month, txs []Transaction, currency string) *MonthlySummary {
	if currency == "" {
		currency = "USD"
	}

	summary := &MonthlySummary{
		Year:              year,
		Month:             month,
		TotalIncome:       Money{Cents: 0, Currency: currency},
		TotalExpense:      Money{Cents: 0, Currency: currency},
		NetBalance:        Money{Cents: 0, Currency: currency},
		IncomeByCategory:  make(map[string]CategorySummary),
		ExpenseByCategory: make(map[string]CategorySummary),
	}

	for _, tx := range txs {
		if tx.Timestamp.Year() != year || tx.Timestamp.Month() != month {
			continue
		}

		summary.TransactionCount++

		if tx.Type == TypeIncome {
			summary.TotalIncome.Cents += tx.Amount.Cents

			catSum := summary.IncomeByCategory[tx.Category]
			catSum.Category = tx.Category
			catSum.Count++
			catSum.Amount.Cents += tx.Amount.Cents
			catSum.Amount.Currency = currency
			summary.IncomeByCategory[tx.Category] = catSum
		} else if tx.Type == TypeExpense {
			summary.TotalExpense.Cents += tx.Amount.Cents

			catSum := summary.ExpenseByCategory[tx.Category]
			catSum.Category = tx.Category
			catSum.Count++
			catSum.Amount.Cents += tx.Amount.Cents
			catSum.Amount.Currency = currency
			summary.ExpenseByCategory[tx.Category] = catSum
		}
	}

	summary.NetBalance.Cents = summary.TotalIncome.Cents - summary.TotalExpense.Cents
	return summary
}
