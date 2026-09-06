package usecases

import (
	"cashflow/internal/ports"
)

// TransactionService coordinates domain entities, storage repository, and VCS synchronization.
// It implements ports.TransactionUseCases.
type TransactionService struct {
	repo            ports.TransactionRepository
	sync            ports.SyncService
	defaultCurrency string
}

// NewTransactionService constructs a new TransactionService.
func NewTransactionService(
	repo ports.TransactionRepository,
	sync ports.SyncService,
	defaultCurrency string,
) *TransactionService {
	if defaultCurrency == "" {
		defaultCurrency = "USD"
	}
	return &TransactionService{
		repo:            repo,
		sync:            sync,
		defaultCurrency: defaultCurrency,
	}
}
