package domain

import "errors"

var (
	ErrInvalidAmount         = errors.New("amount must be greater than zero")
	ErrInvalidCategory       = errors.New("invalid category format: must be alphanumeric with dashes or underscores")
	ErrStorageUnavailable    = errors.New("storage is currently unavailable")
	ErrStorageUninitialized  = errors.New("storage has not been initialized")
	ErrDuplicateTransaction  = errors.New("duplicate transaction ID")
	ErrNoRemoteConfigured    = errors.New("no remote Git repository configured")
	ErrGitConflict           = errors.New("git conflict encountered")
	ErrGitNetworkFailure     = errors.New("network failure while connecting to remote git repository")
	ErrGitAuthFailure        = errors.New("git authentication failed")
	ErrCorruptFile           = errors.New("corrupt entry file")
)
