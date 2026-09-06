package ports

import (
	"context"
	"time"

	"cashflow/internal/domain"
)

// TransactionFilter specifies query parameters for retrieving transactions.
type TransactionFilter struct {
	Month    *time.Time
	Category string
	Type     *domain.TransactionType
	Limit    int
}

// TransactionRepository defines the outbound secondary port for transaction persistence.
type TransactionRepository interface {
	// Save stores a new transaction as an immutable event file entries/<ID>.json.
	// Precondition: tx.ID is a valid non-empty ULID, tx.Amount.Cents > 0.
	// Postcondition: Writes the entry to disk atomically. Never overwrites an existing ID.
	// Error: ErrStorageUnavailable, ErrDuplicateTransaction.
	Save(ctx context.Context, tx domain.Transaction) error

	// FindAll retrieves all transactions matching the filter.
	// Precondition: filter contains valid parameters or zero values.
	// Postcondition: Returns slice sorted by Timestamp DESC. Returns empty slice if no matches.
	// Error: ErrStorageUnavailable.
	FindAll(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error)

	// ListCategories retrieves all unique categories used.
	// Postcondition: Returns slice sorted alphabetically without duplicates.
	ListCategories(ctx context.Context) ([]string, error)
}
