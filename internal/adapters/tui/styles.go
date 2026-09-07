package tui

import "github.com/charmbracelet/lipgloss"

const (
	CompactBreakpoint = 80
	MinTerminalWidth  = 30
	MinTerminalHeight = 10
)

var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorDanger    = lipgloss.Color("#FF4444")
	ColorMuted     = lipgloss.Color("#626262")
	ColorNeutral   = lipgloss.Color("#DDDDDD")
	ColorHighlight = lipgloss.Color("#E5C07B")
	ColorBgSelect  = lipgloss.Color("#2E3440")

	StyleTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	StyleSubtitle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	StyleMonthSelector = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight)

	StyleCardBase = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginRight(1)

	StyleIncomeCard = StyleCardBase.Copy().
		BorderForeground(ColorSuccess)

	StyleExpenseCard = StyleCardBase.Copy().
		BorderForeground(ColorDanger)

	StyleNetCardPositive = StyleCardBase.Copy().
		BorderForeground(ColorSuccess)

	StyleNetCardNegative = StyleCardBase.Copy().
		BorderForeground(ColorDanger)

	StyleTopCatCard = StyleCardBase.Copy().
		BorderForeground(ColorHighlight)

	StyleTableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	StyleSelectedRow = lipgloss.NewStyle().
		Bold(true).
		Background(ColorBgSelect).
		Foreground(lipgloss.Color("#ECEFF4"))

	StyleNormalRow = lipgloss.NewStyle().
		Foreground(ColorNeutral)

	StyleIncomeText = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	StyleExpenseText = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	StyleFooter = lipgloss.NewStyle().
		Foreground(ColorMuted)
)
