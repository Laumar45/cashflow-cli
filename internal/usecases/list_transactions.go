package usecases

import (
	"context"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

// ListTransactions returns transactions matching the specified filter criteria.
func (s *TransactionService) ListTransactions(
	ctx context.Context,
	filter ports.TransactionFilter,
) ([]domain.Transaction, error) {
	return s.repo.FindAll(ctx, filter)
}

// ListCategories returns all unique categories recorded in storage.
func (s *TransactionService) ListCategories(ctx context.Context) ([]string, error) {
	return s.repo.ListCategories(ctx)
}
