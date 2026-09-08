package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"

	"cashflow/internal/domain"
)

// renderCards generates the 4 metric cards adhering to symbolic accessibility (DEC-TUI-05).
func renderCards(summary *domain.MonthlySummary, isCompact bool, width int) string {
	incomeVal := "+USD 0.00"
	expenseVal := "-USD 0.00"
	balanceVal := "+USD 0.00"
	topCat := "—"
	isPositiveBalance := true

	if summary != nil {
		incomeVal = "+" + summary.TotalIncome.Format()
		expenseVal = "-" + summary.TotalExpense.Format()
		balanceVal = summary.NetBalance.Format()
		if summary.NetBalance.Cents >= 0 {
			balanceVal = "+" + balanceVal
			isPositiveBalance = true
		} else {
			isPositiveBalance = false
		}

		// Calculate Top Category from expense aggregates (DEC-TUI-08)
		if len(summary.ExpenseByCategory) > 0 {
			type catItem struct {
				name  string
				cents int64
			}
			var items []catItem
			for name, cat := range summary.ExpenseByCategory {
				items = append(items, catItem{name: name, cents: cat.Amount.Cents})
			}
			sort.Slice(items, func(i, j int) bool {
				return items[i].cents > items[j].cents
			})
			if len(items) > 0 {
				topCat = items[0].name
			}
		}
	}

	// 1. Income Card
	incomeTitle := lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("INGRESOS  ▲")
	incomeContent := StyleIncomeText.Render(incomeVal)
	cardIncome := StyleIncomeCard.Render(fmt.Sprintf("%s\n%s", incomeTitle, incomeContent))

	// 2. Expense Card
	expenseTitle := lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("GASTOS    ▼")
	expenseContent := StyleExpenseText.Render(expenseVal)
	cardExpense := StyleExpenseCard.Render(fmt.Sprintf("%s\n%s", expenseTitle, expenseContent))

	// 3. Balance Card
	var netTitle, netContent string
	var cardNet string
	if isPositiveBalance {
		netTitle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("BALANCE   ▲")
		netContent = StyleIncomeText.Render(balanceVal)
		cardNet = StyleNetCardPositive.Render(fmt.Sprintf("%s\n%s", netTitle, netContent))
	} else {
		netTitle = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("BALANCE   ▼")
		netContent = StyleExpenseText.Render(balanceVal)
		cardNet = StyleNetCardNegative.Render(fmt.Sprintf("%s\n%s", netTitle, netContent))
	}

	// 4. Top Category Card
	topCatTitle := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true).Render("TOP CAT.  ★")
	topCatContent := lipgloss.NewStyle().Foreground(ColorHighlight).Render(topCat)
	cardTopCat := StyleTopCatCard.Render(fmt.Sprintf("%s\n%s", topCatTitle, topCatContent))

	if isCompact {
		// Use two columns when the terminal can fit two cards without wrapping.
		if width >= 48 {
			rowOne := lipgloss.JoinHorizontal(lipgloss.Top, cardIncome, cardExpense)
			rowTwo := lipgloss.JoinHorizontal(lipgloss.Top, cardNet, cardTopCat)
			return lipgloss.JoinVertical(lipgloss.Left, rowOne, rowTwo)
		}

		return lipgloss.JoinVertical(lipgloss.Left,
			cardIncome,
			cardExpense,
			cardNet,
			cardTopCat,
		)
	}

	// Standard horizontal row
	return lipgloss.JoinHorizontal(lipgloss.Top,
		cardIncome,
		cardExpense,
		cardNet,
		cardTopCat,
	)
}
