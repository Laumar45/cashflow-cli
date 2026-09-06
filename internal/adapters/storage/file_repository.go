package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

// FileRepository implements ports.TransactionRepository storing each transaction
// as an immutable JSON file in ~/.cashflow/entries/<ulid>.json.
type FileRepository struct {
	baseDir    string
	entriesDir string
	warnWriter io.Writer
}

// NewFileRepository constructs a new FileRepository.
// If baseDir is empty, defaults to ~/.cashflow.
func NewFileRepository(baseDir string) (*FileRepository, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("%w: unable to determine user home directory: %v", domain.ErrStorageUnavailable, err)
		}
		baseDir = filepath.Join(home, ".cashflow")
	}

	cleanBaseDir := filepath.Clean(baseDir)
	entriesDir := filepath.Join(cleanBaseDir, "entries")

	return &FileRepository{
		baseDir:    cleanBaseDir,
		entriesDir: entriesDir,
		warnWriter: os.Stderr,
	}, nil
}

// SetWarnWriter overrides the writer used for non-fatal corruption warnings (useful in tests).
func (r *FileRepository) SetWarnWriter(w io.Writer) {
	r.warnWriter = w
}

// EntriesDir returns the directory path where entry files are stored.
func (r *FileRepository) EntriesDir() string {
	return r.entriesDir
}

// BaseDir returns the root cashflow directory.
func (r *FileRepository) BaseDir() string {
	return r.baseDir
}

// Init creates the base and entries directories if they do not already exist.
func (r *FileRepository) Init() error {
	if err := os.MkdirAll(r.entriesDir, 0755); err != nil {
		return fmt.Errorf("%w: failed to create storage directory: %v", domain.ErrStorageUnavailable, err)
	}
	return nil
}

// Save writes an immutable entry file entries/<ID>.json.
// Precondition: tx.ID is a valid ULID, tx.Amount.Cents > 0.
// Postcondition: Atomic file write. Returns ErrDuplicateTransaction if file already exists.
func (r *FileRepository) Save(ctx context.Context, tx domain.Transaction) error {
	if strings.TrimSpace(tx.ID) == "" {
		return fmt.Errorf("%w: transaction ID cannot be empty", domain.ErrStorageUnavailable)
	}
	if tx.Amount.Cents <= 0 {
		return domain.ErrInvalidAmount
	}

	// Verify storage is initialized
	if _, err := os.Stat(r.entriesDir); os.IsNotExist(err) {
		return domain.ErrStorageUninitialized
	}

	filePath := filepath.Join(r.entriesDir, fmt.Sprintf("%s.json", tx.ID))

	data, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: failed to serialize transaction: %v", domain.ErrStorageUnavailable, err)
	}

	// Use O_CREATE | O_EXCL to guarantee atomic creation without overwrite
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return domain.ErrDuplicateTransaction
		}
		return fmt.Errorf("%w: %v", domain.ErrStorageUnavailable, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("%w: failed to write entry file: %v", domain.ErrStorageUnavailable, err)
	}

	return nil
}

// FindAll reads and parses all entry files, filtering and sorting by Timestamp DESC.
// Corrupt files emit a warning and are skipped without aborting.
func (r *FileRepository) FindAll(ctx context.Context, filter ports.TransactionFilter) ([]domain.Transaction, error) {
	if _, err := os.Stat(r.entriesDir); os.IsNotExist(err) {
		return nil, domain.ErrStorageUninitialized
	}

	dirEntries, err := os.ReadDir(r.entriesDir)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read entries directory: %v", domain.ErrStorageUnavailable, err)
	}

	var transactions []domain.Transaction

	for _, entry := range dirEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(r.entriesDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			if r.warnWriter != nil {
				fmt.Fprintf(r.warnWriter, "Advertencia: Error al leer archivo '%s': %v\n", entry.Name(), err)
			}
			continue
		}

		var tx domain.Transaction
		if err := json.Unmarshal(data, &tx); err != nil {
			if r.warnWriter != nil {
				fmt.Fprintf(r.warnWriter, "Advertencia: Se omitió archivo corrupto '%s': JSON inválido\n", entry.Name())
			}
			continue
		}

		// Apply filter criteria
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

		transactions = append(transactions, tx)
	}

	// Sort descending by timestamp
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.After(transactions[j].Timestamp)
	})

	if filter.Limit > 0 && len(transactions) > filter.Limit {
		transactions = transactions[:filter.Limit]
	}

	return transactions, nil
}

// ListCategories extracts all unique category slugs sorted alphabetically.
func (r *FileRepository) ListCategories(ctx context.Context) ([]string, error) {
	txs, err := r.FindAll(ctx, ports.TransactionFilter{})
	if err != nil {
		if errors.Is(err, domain.ErrStorageUninitialized) {
			return []string{}, nil
		}
		return nil, err
	}

	categorySet := make(map[string]struct{})
	for _, tx := range txs {
		categorySet[tx.Category] = struct{}{}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	return categories, nil
}
