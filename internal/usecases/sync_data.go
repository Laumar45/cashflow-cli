package usecases

import (
	"context"

	"cashflow/internal/domain"
)

// Synchronize triggers the distributed Git synchronization cycle.
func (s *TransactionService) Synchronize(ctx context.Context) error {
	if s.sync == nil {
		return domain.ErrNoRemoteConfigured
	}
	return s.sync.Sync(ctx)
}
