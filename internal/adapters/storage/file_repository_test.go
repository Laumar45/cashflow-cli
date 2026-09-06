package storage_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cashflow/internal/adapters/storage"
	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

func TestFileRepositorySaveAndFindAll(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Prior to Init(), Save should fail with ErrStorageUninitialized
	money, _ := domain.NewMoney(10.00, "USD")
	sampleTx, _ := domain.NewTransaction(domain.TypeIncome, "sueldo", money, "Init test", time.Now())
	if err := repo.Save(ctx, sampleTx); !errors.Is(err, domain.ErrStorageUninitialized) {
		t.Fatalf("expected ErrStorageUninitialized, got %v", err)
	}

	// Initialize storage directory
	if err := repo.Init(); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Done-when Condition 1: Cada llamada a Save() crea un archivo único ~/.cashflow/entries/<ulid>.json
	t1 := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, time.September, 2, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, time.September, 3, 14, 0, 0, 0, time.UTC)

	tx1, _ := domain.NewTransaction(domain.TypeIncome, "sueldo", money, "Pago 1", t1)
	m2, _ := domain.NewMoney(15.50, "USD")
	tx2, _ := domain.NewTransaction(domain.TypeExpense, "almuerzo", m2, "Comida", t2)
	m3, _ := domain.NewMoney(5.00, "USD")
	tx3, _ := domain.NewTransaction(domain.TypeExpense, "transporte", m3, "Bus", t3)

	for _, tx := range []domain.Transaction{tx1, tx2, tx3} {
		if err := repo.Save(ctx, tx); err != nil {
			t.Fatalf("failed to save transaction %s: %v", tx.ID, err)
		}
		expectedPath := filepath.Join(repo.EntriesDir(), tx.ID+".json")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Fatalf("expected file to exist at %s, but not found", expectedPath)
		}
	}

	// Test duplicate save rejection
	if err := repo.Save(ctx, tx1); !errors.Is(err, domain.ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction on saving duplicate ID, got %v", err)
	}

	// Done-when Condition 2: FindAll() lee el directorio, parsea los archivos JSON y retorna los registros ordenados por timestamp descendente
	allTxs, err := repo.FindAll(ctx, ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("failed to retrieve all transactions: %v", err)
	}

	if len(allTxs) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(allTxs))
	}

	// Verify descending sort order: tx3 (Sept 3) -> tx2 (Sept 2) -> tx1 (Sept 1)
	if allTxs[0].ID != tx3.ID {
		t.Errorf("expected first tx to be tx3 (%s), got %s", tx3.ID, allTxs[0].ID)
	}
	if allTxs[1].ID != tx2.ID {
		t.Errorf("expected second tx to be tx2 (%s), got %s", tx2.ID, allTxs[1].ID)
	}
	if allTxs[2].ID != tx1.ID {
		t.Errorf("expected third tx to be tx1 (%s), got %s", tx1.ID, allTxs[2].ID)
	}
}

func TestFileRepositoryCorruptFileHandling(t *testing.T) {
	// Done-when Condition 3: La presencia de un archivo con JSON inválido emite una advertencia pero no aborta el parseo de los demás archivos
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	var warnBuf bytes.Buffer
	repo.SetWarnWriter(&warnBuf)

	ctx := context.Background()

	// 1. Save valid transaction
	m, _ := domain.NewMoney(20.00, "USD")
	validTx, _ := domain.NewTransaction(domain.TypeIncome, "freelance", m, "Valid project", time.Now())
	if err := repo.Save(ctx, validTx); err != nil {
		t.Fatalf("failed to save valid transaction: %v", err)
	}

	// 2. Introduce corrupt JSON file directly into entries directory
	corruptPath := filepath.Join(repo.EntriesDir(), "corrupt-entry.json")
	if err := os.WriteFile(corruptPath, []byte("NOT_A_VALID_JSON{{{"), 0644); err != nil {
		t.Fatalf("failed to create corrupt file: %v", err)
	}

	// 3. FindAll must succeed, returning the valid transaction and emitting warning
	txs, err := repo.FindAll(ctx, ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("FindAll should not return error when corrupt file is present, got: %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("expected 1 valid transaction, got %d", len(txs))
	}
	if txs[0].ID != validTx.ID {
		t.Errorf("expected valid tx ID %s, got %s", validTx.ID, txs[0].ID)
	}

	// Verify warning was emitted
	warnOutput := warnBuf.String()
	if !strings.Contains(warnOutput, "corrupt-entry.json") || !strings.Contains(warnOutput, "Advertencia") {
		t.Errorf("expected warning about corrupt-entry.json, got: '%s'", warnOutput)
	}
}

func TestFileRepositoryListCategories(t *testing.T) {
	tempDir := t.TempDir()
	repo, _ := storage.NewFileRepository(tempDir)
	_ = repo.Init()
	ctx := context.Background()

	m, _ := domain.NewMoney(5.00, "USD")
	tx1, _ := domain.NewTransaction(domain.TypeExpense, "transporte", m, "", time.Now())
	tx2, _ := domain.NewTransaction(domain.TypeExpense, "comida", m, "", time.Now())
	tx3, _ := domain.NewTransaction(domain.TypeIncome, "sueldo", m, "", time.Now())
	tx4, _ := domain.NewTransaction(domain.TypeExpense, "comida", m, "", time.Now())

	for _, tx := range []domain.Transaction{tx1, tx2, tx3, tx4} {
		_ = repo.Save(ctx, tx)
	}

	cats, err := repo.ListCategories(ctx)
	if err != nil {
		t.Fatalf("failed to list categories: %v", err)
	}

	expected := []string{"comida", "sueldo", "transporte"}
	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d", len(expected), len(cats))
	}
	for i, c := range cats {
		if c != expected[i] {
			t.Errorf("expected '%s' at index %d, got '%s'", expected[i], i, c)
		}
	}
}
