package ports

import "context"

// SyncService defines the outbound secondary port for distributed Git synchronization.
type SyncService interface {
	// Sync executes the distributed synchronization cycle.
	// Precondition: Storage directory is an initialized Git repository with configured remote.
	// Postcondition: Performs git add -> git commit -> git pull --rebase -> git push.
	// Error: ErrNoRemoteConfigured, ErrGitConflict, ErrGitNetworkFailure, ErrGitAuthFailure.
	Sync(ctx context.Context) error
}
