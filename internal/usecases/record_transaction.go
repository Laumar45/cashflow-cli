package usecases

import (
	"context"
	"time"

	"cashflow/internal/domain"
)

// RecordIncome validates and persists a new income transaction.
func (s *TransactionService) RecordIncome(
	ctx context.Context,
	category string,
	amount float64,
	desc string,
	date *time.Time,
) (*domain.Transaction, error) {
	return s.record(ctx, domain.TypeIncome, category, amount, desc, date)
}

// RecordExpense validates and persists a new expense transaction.
func (s *TransactionService) RecordExpense(
	ctx context.Context,
	category string,
	amount float64,
	desc string,
	date *time.Time,
) (*domain.Transaction, error) {
	return s.record(ctx, domain.TypeExpense, category, amount, desc, date)
}

func (s *TransactionService) record(
	ctx context.Context,
	txType domain.TransactionType,
	category string,
	amount float64,
	desc string,
	date *time.Time,
) (*domain.Transaction, error) {
	money, err := domain.NewMoney(amount, s.defaultCurrency)
	if err != nil {
		return nil, err
	}

	var ts time.Time
	if date != nil && !date.IsZero() {
		ts = *date
	} else {
		ts = time.Now().UTC()
	}

	tx, err := domain.NewTransaction(txType, category, money, desc, ts)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		return nil, err
	}

	return &tx, nil
}
