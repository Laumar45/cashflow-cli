package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"cashflow/internal/domain"
)

// renderTable formats the transactions list with row cursor and responsive columns.
func renderTable(transactions []domain.Transaction, cursorIndex int, isCompact bool, maxRows int, width int) string {
	if len(transactions) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(2, 4).
			Render("No hay transacciones registradas para este mes.")
		return emptyMsg
	}

	if maxRows <= 0 {
		maxRows = 8
	}

	// Calculate visible window for scroll
	totalRows := len(transactions)
	startIdx := 0
	if cursorIndex >= maxRows {
		startIdx = cursorIndex - maxRows + 1
	}
	endIdx := startIdx + maxRows
	if endIdx > totalRows {
		endIdx = totalRows
	}

	var b strings.Builder

	if isCompact {
		// Compact columns (<80 cols): FECHA | TIPO | MONTO
		dateWidth, typeWidth, separator := 6, 10, "  "
		if width < 48 {
			dateWidth, typeWidth, separator = 5, 9, " "
		}
		amountWidth := width - dateWidth - typeWidth - (len(separator) * 2)
		if amountWidth < 12 {
			amountWidth = 12
		}
		header := fmt.Sprintf("%-*s%s%-*s%s%*s", dateWidth, "FECHA", separator, typeWidth, "TIPO", separator, amountWidth, "MONTO")
		b.WriteString(StyleTableHeader.Render(header))
		b.WriteString("\n")

		for i := startIdx; i < endIdx; i++ {
			tx := transactions[i]
			dateStr := tx.Timestamp.Format("01-02")

			var typeStr, amountStr string
			if tx.Type == domain.TypeIncome {
				typeStr = "▲ INCOME"
				amountStr = "+" + tx.Amount.Format()
			} else {
				typeStr = "▼ EXPENSE"
				amountStr = "-" + tx.Amount.Format()
			}

			rowText := fmt.Sprintf("%-*s%s%-*s%s%*s", dateWidth, dateStr, separator, typeWidth, typeStr, separator, amountWidth, amountStr)

			if i == cursorIndex {
				b.WriteString(StyleSelectedRow.Render(rowText))
			} else {
				if tx.Type == domain.TypeIncome {
					b.WriteString(StyleIncomeText.Render(rowText))
				} else {
					b.WriteString(StyleExpenseText.Render(rowText))
				}
			}
			b.WriteString("\n")
		}
	} else {
		// Standard columns (≥80 cols): FECHA | TIPO | CATEGORÍA | MONTO | DESCRIPCIÓN
		header := fmt.Sprintf("%-12s  %-11s  %-16s  %14s  %s", "FECHA", "TIPO", "CATEGORÍA", "MONTO", "DESCRIPCIÓN")
		b.WriteString(StyleTableHeader.Render(header))
		b.WriteString("\n")

		for i := startIdx; i < endIdx; i++ {
			tx := transactions[i]
			dateStr := tx.Timestamp.Format("2006-01-02")

			var typeStr, amountStr string
			if tx.Type == domain.TypeIncome {
				typeStr = "▲ INCOME"
				amountStr = "+" + tx.Amount.Format()
			} else {
				typeStr = "▼ EXPENSE"
				amountStr = "-" + tx.Amount.Format()
			}

			catStr := tx.Category
			if len(catStr) > 16 {
				catStr = catStr[:14] + ".."
			}

			descStr := tx.Description
			if len(descStr) > 25 {
				descStr = descStr[:23] + ".."
			}

			rowText := fmt.Sprintf("%-12s  %-11s  %-16s  %14s  %s",
				dateStr,
				typeStr,
				catStr,
				amountStr,
				descStr)

			if i == cursorIndex {
				b.WriteString(StyleSelectedRow.Render(rowText))
			} else {
				if tx.Type == domain.TypeIncome {
					b.WriteString(StyleIncomeText.Render(rowText))
				} else {
					b.WriteString(StyleExpenseText.Render(rowText))
				}
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
