package vcs_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"cashflow/internal/adapters/storage"
	"cashflow/internal/adapters/vcs"
	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed in %s: %v (stderr: %s)", args, dir, err, stderr.String())
	}
}

func TestGitServiceNoRemoteAndUninitialized(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()

	// 1. Non-git directory -> ErrStorageUninitialized
	svc := vcs.NewGitService(tempDir, "origin", "main")
	if err := svc.Sync(ctx); !errors.Is(err, domain.ErrStorageUninitialized) {
		t.Fatalf("expected ErrStorageUninitialized, got: %v", err)
	}

	// 2. Git repo without remote -> ErrNoRemoteConfigured
	runGit(t, tempDir, "init", "-b", "main")
	runGit(t, tempDir, "config", "user.name", "Test User")
	runGit(t, tempDir, "config", "user.email", "test@example.com")

	if err := svc.Sync(ctx); !errors.Is(err, domain.ErrNoRemoteConfigured) {
		t.Fatalf("expected ErrNoRemoteConfigured, got: %v", err)
	}
}

func TestMultiDeviceConflictFreeSync(t *testing.T) {
	// Verifies Acceptance Criterion AC-03 from BRIEF.md:
	// "Si el Dispositivo A crea una transacción desconectado y el Dispositivo B crea otra desconectado,
	// al ejecutar cash sync en ambos, ambos repositorios contienen ambas transacciones sin intervención manual ni conflicto de Git."

	root := t.TempDir()
	bareRemoteDir := filepath.Join(root, "remote.git")
	deviceADir := filepath.Join(root, "deviceA")
	deviceBDir := filepath.Join(root, "deviceB")

	ctx := context.Background()

	// 1. Setup bare remote repo
	runGit(t, root, "init", "--bare", "--initial-branch=main", bareRemoteDir)

	// 2. Setup Device A
	repoA, _ := storage.NewFileRepository(deviceADir)
	_ = repoA.Init()
	runGit(t, deviceADir, "init", "-b", "main")
	runGit(t, deviceADir, "config", "user.name", "Device A")
	runGit(t, deviceADir, "config", "user.email", "deviceA@example.com")
	runGit(t, deviceADir, "remote", "add", "origin", bareRemoteDir)

	// Initial commit to establish main branch on remote
	readmeFile := filepath.Join(deviceADir, "README.md")
	_ = os.WriteFile(readmeFile, []byte("# CashFlow Storage"), 0644)
	runGit(t, deviceADir, "add", "README.md")
	runGit(t, deviceADir, "commit", "-m", "init repo")
	runGit(t, deviceADir, "push", "-u", "origin", "main")

	// 3. Setup Device B by cloning bare remote
	runGit(t, root, "clone", bareRemoteDir, deviceBDir)
	repoB, _ := storage.NewFileRepository(deviceBDir)
	_ = repoB.Init()
	runGit(t, deviceBDir, "config", "user.name", "Device B")
	runGit(t, deviceBDir, "config", "user.email", "deviceB@example.com")

	// 4. Offline write on Device A
	mA, _ := domain.NewMoney(100.00, "USD")
	txA, _ := domain.NewTransaction(domain.TypeIncome, "freelance", mA, "Pago Cliente A", time.Now().Add(-1*time.Hour))
	if err := repoA.Save(ctx, txA); err != nil {
		t.Fatalf("failed saving tx on Device A: %v", err)
	}

	// 5. Offline write on Device B (concurrent, before pulling Device A)
	mB, _ := domain.NewMoney(25.00, "USD")
	txB, _ := domain.NewTransaction(domain.TypeExpense, "almuerzo", mB, "Almuerzo B", time.Now())
	if err := repoB.Save(ctx, txB); err != nil {
		t.Fatalf("failed saving tx on Device B: %v", err)
	}

	// 6. Device A syncs first
	svcA := vcs.NewGitService(deviceADir, "origin", "main")
	svcA.SetStepLogger(&bytes.Buffer{})
	if err := svcA.Sync(ctx); err != nil {
		t.Fatalf("Device A sync failed: %v", err)
	}

	// 7. Device B syncs (must auto-merge independent entry files without conflict!)
	svcB := vcs.NewGitService(deviceBDir, "origin", "main")
	svcB.SetStepLogger(&bytes.Buffer{})
	if err := svcB.Sync(ctx); err != nil {
		t.Fatalf("Device B sync failed with conflict: %v", err)
	}

	// 8. Device A syncs to pull Device B's changes
	if err := svcA.Sync(ctx); err != nil {
		t.Fatalf("Device A second sync failed: %v", err)
	}

	// 9. Verify BOTH devices have BOTH transactions!
	txsOnA, err := repoA.FindAll(ctx, ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("failed to list txs on Device A: %v", err)
	}
	if len(txsOnA) != 2 {
		t.Errorf("Device A expected 2 transactions, got %d", len(txsOnA))
	}

	txsOnB, err := repoB.FindAll(ctx, ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("failed to list txs on Device B: %v", err)
	}
	if len(txsOnB) != 2 {
		t.Errorf("Device B expected 2 transactions, got %d", len(txsOnB))
	}
}
