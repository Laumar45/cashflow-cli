package usecases

import (
	"context"
	"time"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

// GetMonthlySummary retrieves all transactions for a given month and computes financial metrics.
func (s *TransactionService) GetMonthlySummary(
	ctx context.Context,
	year int,
	month time.Month,
) (*domain.MonthlySummary, error) {
	filterMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	filter := ports.TransactionFilter{
		Month: &filterMonth,
	}

	txs, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	return domain.CalculateMonthlySummary(year, month, txs, s.defaultCurrency), nil
}
