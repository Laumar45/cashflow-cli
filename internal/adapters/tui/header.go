package tui

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

//go:embed assets/dollar-sign.txt
var embeddedDollarLogo string

var spanishMonths = map[time.Month]string{
	time.January:   "ENERO",
	time.February:  "FEBRERO",
	time.March:     "MARZO",
	time.April:     "ABRIL",
	time.May:       "MAYO",
	time.June:      "JUNIO",
	time.July:      "JULIO",
	time.August:    "AGOSTO",
	time.September: "SEPTIEMBRE",
	time.October:   "OCTUBRE",
	time.November:  "NOVIEMBRE",
	time.December:  "DICIEMBRE",
}

// renderHeader renders the ASCII logo alongside the dynamic month navigation selector.
func renderHeader(logo string, currentMonth time.Time, isCompact bool) string {
	monthName := currentMonth.Format("January")
	if name, ok := spanishMonths[currentMonth.Month()]; ok {
		monthName = name
	}
	year := currentMonth.Year()

	monthSelector := fmt.Sprintf("<  %s %d  >", monthName, year)
	monthBox := StyleMonthSelector.Render(monthSelector)

	cleanLogo := strings.TrimSpace(logo)
	if cleanLogo == "" {
		cleanLogo = strings.TrimSpace(embeddedDollarLogo)
	}
	if cleanLogo == "" {
		cleanLogo = "  $$$   C A S H F L O W\n $   $  ---------------\n  $$$   T U I"
	}

	styledLogo := lipgloss.NewStyle().Foreground(ColorHighlight).Render(cleanLogo)

	if isCompact {
		return lipgloss.JoinVertical(lipgloss.Left,
			styledLogo,
			"",
			monthBox,
		)
	}

	titleBlock := lipgloss.JoinVertical(lipgloss.Left,
		StyleTitle.Render("C A S H F L O W   T U I"),
		StyleSubtitle.Render("Terminal Cashflow & Financial Dashboard"),
		"",
		monthBox,
	)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		styledLogo,
		"    ",
		titleBlock,
	)
}
