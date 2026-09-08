package tui_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"cashflow/internal/adapters/tui"
	"cashflow/internal/domain"
	"cashflow/internal/ports"
	"cashflow/internal/usecases"
)

// mockRepo provides deterministic in-memory storage for TUI model testing.
type mockRepo struct {
	txs []domain.Transaction
}

func (m *mockRepo) Save(ctx context.Context, tx domain.Transaction) error {
	m.txs = append(m.txs, tx)
	return nil
}

func (m *mockRepo) FindAll(ctx context.Context, filter ports.TransactionFilter) ([]domain.Transaction, error) {
	var res []domain.Transaction
	for _, tx := range m.txs {
		if filter.Month != nil {
			if tx.Timestamp.Year() != filter.Month.Year() || tx.Timestamp.Month() != filter.Month.Month() {
				continue
			}
		}
		res = append(res, tx)
	}
	return res, nil
}

func (m *mockRepo) ListCategories(ctx context.Context) ([]string, error) {
	return []string{"sueldo", "almuerzo"}, nil
}

func setupTestService() ports.TransactionUseCases {
	repo := &mockRepo{}
	ctx := context.Background()

	// Month 1: September 2026
	t1 := time.Date(2026, time.September, 15, 10, 0, 0, 0, time.UTC)
	m1, _ := domain.NewMoney(3000.00, "USD")
	tx1, _ := domain.NewTransaction(domain.TypeIncome, "sueldo", m1, "Quincena", t1)
	_ = repo.Save(ctx, tx1)

	m2, _ := domain.NewMoney(20.00, "USD")
	tx2, _ := domain.NewTransaction(domain.TypeExpense, "almuerzo", m2, "Menu ejecutivo", t1)
	_ = repo.Save(ctx, tx2)

	// Month 2: August 2026
	t2 := time.Date(2026, time.August, 10, 10, 0, 0, 0, time.UTC)
	m3, _ := domain.NewMoney(1500.00, "USD")
	tx3, _ := domain.NewTransaction(domain.TypeIncome, "freelance", m3, "Web project", t2)
	_ = repo.Save(ctx, tx3)

	return usecases.NewTransactionService(repo, nil, "USD")
}

func TestTUIModelRenderingAndAccessibility(t *testing.T) {
	// Verifies AC-05 (Renderizado de Dashboard) & AC-06 (Accesibilidad Simbólica)
	svc := setupTestService()
	sept := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	syncInfo := tui.SyncStatusInfo{
		LastSyncRelative: "hace 2h",
		UnsyncedCount:    1,
	}

	model := tui.NewModel(svc, sept, syncInfo)

	viewOutput := model.View()

	// Check Header & Month
	if !strings.Contains(viewOutput, "SEPTIEMBRE 2026") {
		t.Errorf("expected view to contain 'SEPTIEMBRE 2026', got: %s", viewOutput)
	}

	// Check AC-06: Symbolic Accessibility
	if !strings.Contains(viewOutput, "INGRESOS  ▲") {
		t.Errorf("expected income card with symbol ▲, got: %s", viewOutput)
	}
	if !strings.Contains(viewOutput, "GASTOS    ▼") {
		t.Errorf("expected expense card with symbol ▼, got: %s", viewOutput)
	}
	if !strings.Contains(viewOutput, "BALANCE   ▲") {
		t.Errorf("expected positive balance card with symbol ▲, got: %s", viewOutput)
	}
	if !strings.Contains(viewOutput, "TOP CAT.") {
		t.Errorf("expected top category card, got: %s", viewOutput)
	}

	// Check Table Content
	if !strings.Contains(viewOutput, "sueldo") || !strings.Contains(viewOutput, "almuerzo") {
		t.Errorf("expected table to display transactions, got: %s", viewOutput)
	}

	// Check Sync Info in footer
	if !strings.Contains(viewOutput, "hace 2h") || !strings.Contains(viewOutput, "Cambios sin sync: 1") {
		t.Errorf("expected sync status line in footer, got: %s", viewOutput)
	}
}

func TestTUIMonthNavigation(t *testing.T) {
	// Verifies AC-08 (Navegación Temporal Fluida <15ms)
	svc := setupTestService()
	sept := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	model := tui.NewModel(svc, sept, tui.SyncStatusInfo{})

	// Press left arrow to navigate to August
	start := time.Now()
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	elapsed := time.Since(start)

	if cmd != nil {
		t.Errorf("expected no cmd on navigation, got %v", cmd)
	}

	viewAug := updatedModel.View()
	if !strings.Contains(viewAug, "AGOSTO 2026") {
		t.Errorf("expected updated view to show 'AGOSTO 2026', got: %s", viewAug)
	}
	if !strings.Contains(viewAug, "freelance") {
		t.Errorf("expected August view to show 'freelance' transaction, got: %s", viewAug)
	}

	t.Logf("month navigation completed in %v", elapsed)
	if elapsed > 15*time.Millisecond {
		t.Errorf("month navigation took %v, expected <15ms", elapsed)
	}
}

func TestTUIResponsivenessUnder80Columns(t *testing.T) {
	// Verifies AC-07 (Responsividad <80 Columnas para Termux)
	svc := setupTestService()
	sept := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	model := tui.NewModel(svc, sept, tui.SyncStatusInfo{LastSyncRelative: "hace 10m", UnsyncedCount: 0})

	// Resize to 80 columns: common mobile terminal width, still compact mode.
	updatedModel, _ := model.Update(tea.WindowSizeMsg{
		Width:  80,
		Height: 25,
	})

	viewCompact := updatedModel.View()

	// In compact mode, help bar must be compact
	if !strings.Contains(viewCompact, "[h/l] Mes") {
		t.Errorf("expected compact help bar '[h/l] Mes', got: %s", viewCompact)
	}

	// Compact table shouldn't show DESCRIPCIÓN header
	if strings.Contains(viewCompact, "DESCRIPCIÓN") {
		t.Errorf("expected compact table to hide DESCRIPCIÓN column, but it was present")
	}
	cardRowFound := false
	for _, line := range strings.Split(viewCompact, "\n") {
		if strings.Contains(line, "INGRESOS") && strings.Contains(line, "GASTOS") {
			cardRowFound = true
			break
		}
	}
	if !cardRowFound {
		t.Errorf("expected compact cards to use two columns at width 80")
	}

	for _, width := range []int{30, 60, 80} {
		resizedModel, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: 25})
		for _, line := range strings.Split(resizedModel.View(), "\n") {
			if lipgloss.Width(line) > width {
				t.Errorf("expected compact view line to fit width %d, got %d: %q", width, lipgloss.Width(line), line)
			}
		}
	}
}

func TestTUIQuitCommands(t *testing.T) {
	// Verifies Done-when condition 4: q/esc emits tea.Quit
	svc := setupTestService()
	model := tui.NewModel(svc, time.Now(), tui.SyncStatusInfo{})

	for _, key := range []string{"q", "esc", "ctrl+c"} {
		var keyMsg tea.KeyMsg
		if key == "q" {
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		} else if key == "esc" {
			keyMsg = tea.KeyMsg{Type: tea.KeyEscape}
		} else {
			keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
		}

		_, cmd := model.Update(keyMsg)
		if cmd == nil {
			t.Errorf("expected tea.Quit command on key %s, got nil", key)
		}
	}
}
