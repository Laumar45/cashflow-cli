package usecases_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
	"cashflow/internal/usecases"
)

// mockRepository implements ports.TransactionRepository in memory for deterministic unit testing.
type mockRepository struct {
	txs []domain.Transaction
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		txs: make([]domain.Transaction, 0),
	}
}

func (m *mockRepository) Save(ctx context.Context, tx domain.Transaction) error {
	m.txs = append(m.txs, tx)
	return nil
}

func (m *mockRepository) FindAll(ctx context.Context, filter ports.TransactionFilter) ([]domain.Transaction, error) {
	result := make([]domain.Transaction, 0)
	for _, tx := range m.txs {
		if filter.Month != nil {
			if tx.Timestamp.Year() != filter.Month.Year() || tx.Timestamp.Month() != filter.Month.Month() {
				continue
			}
		}
		if filter.Category != "" && tx.Category != filter.Category {
			continue
		}
		if filter.Type != nil && tx.Type != *filter.Type {
			continue
		}
		result = append(result, tx)
	}

	// Sort descending by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (m *mockRepository) ListCategories(ctx context.Context) ([]string, error) {
	catMap := make(map[string]struct{})
	for _, tx := range m.txs {
		catMap[tx.Category] = struct{}{}
	}

	cats := make([]string, 0, len(catMap))
	for cat := range catMap {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	return cats, nil
}

// mockSyncService implements ports.SyncService in memory.
type mockSyncService struct {
	synced bool
}

func (s *mockSyncService) Sync(ctx context.Context) error {
	s.synced = true
	return nil
}

func TestRecordIncomeAndExpense(t *testing.T) {
	repo := newMockRepository()
	sync := &mockSyncService{}
	svc := usecases.NewTransactionService(repo, sync, "USD")
	ctx := context.Background()

	// 1. Record Income
	income, err := svc.RecordIncome(ctx, "Sueldo", 3000.00, "Quincena", nil)
	if err != nil {
		t.Fatalf("unexpected error recording income: %v", err)
	}
	if income.Type != domain.TypeIncome {
		t.Errorf("expected type income, got %s", income.Type)
	}
	if income.Amount.Cents != 300000 {
		t.Errorf("expected 300000 cents, got %d", income.Amount.Cents)
	}
	if income.Category != "sueldo" {
		t.Errorf("expected category 'sueldo', got '%s'", income.Category)
	}

	// 2. Record Expense
	expense, err := svc.RecordExpense(ctx, "Comida Rápida", 15.50, "Cena", nil)
	if err != nil {
		t.Fatalf("unexpected error recording expense: %v", err)
	}
	if expense.Type != domain.TypeExpense {
		t.Errorf("expected type expense, got %s", expense.Type)
	}
	if expense.Amount.Cents != 1550 {
		t.Errorf("expected 1550 cents, got %d", expense.Amount.Cents)
	}
	if expense.Category != "comida-rapida" {
		t.Errorf("expected category 'comida-rapida', got '%s'", expense.Category)
	}

	// Verify repo state
	txs, err := svc.ListTransactions(ctx, ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions in repository, got %d", len(txs))
	}
}

func TestGetMonthlySummary(t *testing.T) {
	// Verifies Acceptance Criterion AC-02 from BRIEF.md:
	// "Para 2 ingresos de $1000 y 3 gastos de $200, cash summary reporta exactamente:
	// Ingresos: $2000.00, Gastos: $600.00, Balance: +$1400.00."
	repo := newMockRepository()
	sync := &mockSyncService{}
	svc := usecases.NewTransactionService(repo, sync, "USD")
	ctx := context.Background()

	targetDate := time.Date(2026, time.September, 15, 10, 0, 0, 0, time.UTC)

	// 2 incomes of $1000
	for i := 0; i < 2; i++ {
		_, err := svc.RecordIncome(ctx, "freelance", 1000.00, "Proyecto", &targetDate)
		if err != nil {
			t.Fatalf("failed recording income: %v", err)
		}
	}

	// 3 expenses of $200
	for i := 0; i < 3; i++ {
		_, err := svc.RecordExpense(ctx, "alquiler", 200.00, "Gasto oficina", &targetDate)
		if err != nil {
			t.Fatalf("failed recording expense: %v", err)
		}
	}

	summary, err := svc.GetMonthlySummary(ctx, 2026, time.September)
	if err != nil {
		t.Fatalf("failed getting monthly summary: %v", err)
	}

	if summary.TotalIncome.Cents != 200000 {
		t.Errorf("expected TotalIncome 200000 cents ($2000.00), got %d", summary.TotalIncome.Cents)
	}
	if summary.TotalExpense.Cents != 60000 {
		t.Errorf("expected TotalExpense 60000 cents ($600.00), got %d", summary.TotalExpense.Cents)
	}
	if summary.NetBalance.Cents != 140000 {
		t.Errorf("expected NetBalance 140000 cents (+$1400.00), got %d", summary.NetBalance.Cents)
	}
	if summary.TransactionCount != 5 {
		t.Errorf("expected TransactionCount 5, got %d", summary.TransactionCount)
	}
}

func TestListCategoriesAndSync(t *testing.T) {
	repo := newMockRepository()
	sync := &mockSyncService{}
	svc := usecases.NewTransactionService(repo, sync, "USD")
	ctx := context.Background()

	_, _ = svc.RecordIncome(ctx, "sueldo", 2000, "", nil)
	_, _ = svc.RecordExpense(ctx, "cafe", 3.5, "", nil)
	_, _ = svc.RecordExpense(ctx, "almuerzo", 12, "", nil)
	_, _ = svc.RecordExpense(ctx, "cafe", 4, "", nil)

	cats, err := svc.ListCategories(ctx)
	if err != nil {
		t.Fatalf("failed listing categories: %v", err)
	}

	expected := []string{"almuerzo", "cafe", "sueldo"}
	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d", len(expected), len(cats))
	}
	for i, c := range cats {
		if c != expected[i] {
			t.Errorf("expected category '%s' at %d, got '%s'", expected[i], i, c)
		}
	}

	// Test Sync delegation
	if err := svc.Synchronize(ctx); err != nil {
		t.Fatalf("failed synchronizing: %v", err)
	}
	if !sync.synced {
		t.Errorf("expected sync to be called, got false")
	}
}
