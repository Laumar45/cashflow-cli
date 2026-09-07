package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

// Model encapsulates the deterministic Bubble Tea Elm model for the dashboard.
type Model struct {
	svc          ports.TransactionUseCases
	currentMonth time.Time
	summary      *domain.MonthlySummary
	transactions []domain.Transaction
	cursorIndex  int
	width        int
	height       int
	isCompact    bool
	syncInfo     SyncStatusInfo
	logoContent  string
	err          error
}

// NewModel constructs and initializes a new TUI Model.
func NewModel(
	svc ports.TransactionUseCases,
	initialMonth time.Time,
	syncInfo SyncStatusInfo,
	logoContent string,
) Model {
	if initialMonth.IsZero() {
		initialMonth = time.Now().UTC()
	}

	m := Model{
		svc:          svc,
		currentMonth: initialMonth,
		syncInfo:     syncInfo,
		logoContent:  logoContent,
		width:        100, // Default width until WindowSizeMsg arrives
		height:       30,
		isCompact:    false,
	}

	m.loadMonthData()
	return m
}

// loadMonthData queries usecases for monthly summary and filtered transactions.
func (m *Model) loadMonthData() {
	ctx := context.Background()
	year := m.currentMonth.Year()
	month := m.currentMonth.Month()

	summary, err := m.svc.GetMonthlySummary(ctx, year, month)
	if err != nil {
		m.err = err
		return
	}
	m.summary = summary

	filterMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	txs, err := m.svc.ListTransactions(ctx, ports.TransactionFilter{
		Month: &filterMonth,
	})
	if err != nil {
		m.err = err
		return
	}
	m.transactions = txs
	m.cursorIndex = 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.isCompact = m.width < CompactBreakpoint

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "left", "h":
			m.currentMonth = m.currentMonth.AddDate(0, -1, 0)
			m.loadMonthData()

		case "right", "l":
			m.currentMonth = m.currentMonth.AddDate(0, 1, 0)
			m.loadMonthData()

		case "t":
			m.currentMonth = time.Now().UTC()
			m.loadMonthData()

		case "up", "k":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}

		case "down", "j":
			if m.cursorIndex < len(m.transactions)-1 {
				m.cursorIndex++
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width > 0 && m.width < MinTerminalWidth {
		return "Terminal demasiado pequeña para renderizar CashFlow TUI.\n"
	}

	header := renderHeader(m.logoContent, m.currentMonth, m.isCompact)
	cards := renderCards(m.summary, m.isCompact)

	// Available table rows estimation
	tableRows := 10
	if m.height > 25 {
		tableRows = m.height - 18
	}
	table := renderTable(m.transactions, m.cursorIndex, m.isCompact, tableRows)

	// Status Line (Overflow + Local Sync status)
	totalTxs := len(m.transactions)
	currRow := 0
	if totalTxs > 0 {
		currRow = m.cursorIndex + 1
	}

	var statusLine string
	if m.isCompact {
		statusLine = fmt.Sprintf("Fila %d/%d  •  ↻ %s (sin sync: %d)",
			currRow, totalTxs, m.syncInfo.LastSyncRelative, m.syncInfo.UnsyncedCount)
	} else {
		statusLine = fmt.Sprintf("Fila %d/%d  •  Última sync: %s  •  Cambios sin sync: %d",
			currRow, totalTxs, m.syncInfo.LastSyncRelative, m.syncInfo.UnsyncedCount)
	}
	styledStatus := StyleFooter.Render(statusLine)

	// Footer Keybindings
	var helpBar string
	if m.isCompact {
		helpBar = "[h/l] Mes  •  [j/k] Nav  •  [t] Hoy  •  [q] Salir"
	} else {
		helpBar = "[←/→ h/l] Cambiar Mes  •  [↑/↓ j/k] Navegar Filas  •  [t] Hoy  •  [q] Salir"
	}
	styledHelp := StyleFooter.Render(helpBar)

	divider := StyleFooter.Render(lipgloss.NewStyle().Width(m.width).Render("--------------------------------------------------------------------------------"))

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		divider,
		cards,
		divider,
		table,
		divider,
		styledStatus,
		styledHelp,
	)
}
