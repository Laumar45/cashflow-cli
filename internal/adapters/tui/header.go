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
func renderHeader(currentMonth time.Time, isCompact bool, width int) string {
	monthName := currentMonth.Format("January")
	if name, ok := spanishMonths[currentMonth.Month()]; ok {
		monthName = name
	}
	year := currentMonth.Year()

	monthSelector := fmt.Sprintf("<  %s %d  >", monthName, year)
	monthBox := StyleMonthSelector.Render(monthSelector)

	cleanLogo := strings.TrimSpace(embeddedDollarLogo)
	logoLines := strings.Split(cleanLogo, "\n")
	for i, line := range logoLines {
		logoLines[i] = strings.TrimRight(line, " \t")
	}
	cleanLogo = strings.Join(logoLines, "\n")
	if isCompact {
		cleanLogo = "  $$$\n $   $\n $    \n $   $\n  $$$"
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

	if width < MediumBreakpoint {
		return lipgloss.JoinVertical(lipgloss.Left,
			styledLogo,
			"",
			titleBlock,
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		styledLogo,
		"    ",
		titleBlock,
	)
}
