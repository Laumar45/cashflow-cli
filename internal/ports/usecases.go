package ports

import (
	"context"
	"time"

	"cashflow/internal/domain"
)

// TransactionUseCases defines the primary inbound port for CLI orchestration.
type TransactionUseCases interface {
	RecordIncome(ctx context.Context, category string, amount float64, desc string, date *time.Time) (*domain.Transaction, error)
	RecordExpense(ctx context.Context, category string, amount float64, desc string, date *time.Time) (*domain.Transaction, error)
	GetMonthlySummary(ctx context.Context, year int, month time.Month) (*domain.MonthlySummary, error)
	ListTransactions(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error)
	ListCategories(ctx context.Context) ([]string, error)
	Synchronize(ctx context.Context) error
}
