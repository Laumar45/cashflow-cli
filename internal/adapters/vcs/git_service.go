package vcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"cashflow/internal/domain"
)

// GitService implements ports.SyncService via native Git CLI invocation.
type GitService struct {
	repoDir    string
	remote     string
	branch     string
	stepLogger io.Writer
}

// NewGitService constructs a GitService adapter.
func NewGitService(repoDir string, remote string, branch string) *GitService {
	if remote == "" {
		remote = "origin"
	}
	return &GitService{
		repoDir:    repoDir,
		remote:     remote,
		branch:     branch,
		stepLogger: os.Stdout,
	}
}

// SetStepLogger overrides the logger used for progress steps.
func (s *GitService) SetStepLogger(w io.Writer) {
	s.stepLogger = w
}

// run executes a git command in the repository directory and returns stdout and stderr.
func (s *GitService) run(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func (s *GitService) classifyGitError(stderr string, err error) error {
	lower := strings.ToLower(stderr)
	if strings.Contains(lower, "could not resolve host") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "network is unreachable") ||
		strings.Contains(lower, "operation timed out") ||
		strings.Contains(lower, "unable to access") {
		return fmt.Errorf("%w: %s", domain.ErrGitNetworkFailure, stderr)
	}

	if strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "fatal: repository") && strings.Contains(lower, "not found") {
		return fmt.Errorf("%w: %s", domain.ErrGitAuthFailure, stderr)
	}

	if strings.Contains(lower, "conflict") || strings.Contains(lower, "could not apply") {
		return fmt.Errorf("%w: %s", domain.ErrGitConflict, stderr)
	}

	return fmt.Errorf("git command error (%v): %s", err, stderr)
}

// Sync performs the 3-step synchronization cycle:
// [1/3] git add entries/ -> git commit -m "sync"
// [2/3] git pull --rebase
// [3/3] git push
func (s *GitService) resolveBranch(ctx context.Context) string {
	if s.branch != "" {
		return s.branch
	}

	// 1. Try symbolic-ref (works on unborn branches with 0 commits)
	currentBranch, _, err := s.run(ctx, "symbolic-ref", "--short", "HEAD")
	if err == nil && currentBranch != "" {
		return currentBranch
	}

	// 2. Try rev-parse --abbrev-ref HEAD
	currentBranch, _, err = s.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil && currentBranch != "" && currentBranch != "HEAD" {
		return currentBranch
	}

	return "main"
}

// Sync performs the 3-step synchronization cycle:
// [1/3] git add entries/ -> git commit -m "sync"
// [2/3] git pull --rebase
// [3/3] git push
func (s *GitService) Sync(ctx context.Context) error {
	// Verify git repo exists
	if _, _, err := s.run(ctx, "rev-parse", "--is-inside-work-tree"); err != nil {
		return domain.ErrStorageUninitialized
	}

	// Verify remote is configured
	if _, _, err := s.run(ctx, "remote", "get-url", s.remote); err != nil {
		return domain.ErrNoRemoteConfigured
	}

	// Resolve current branch
	branch := s.resolveBranch(ctx)

	// Step 1: Stage and commit local entries
	if s.stepLogger != nil {
		fmt.Fprintln(s.stepLogger, "[1/3] Guardando transacciones locales...")
	}

	_, stderr, err := s.run(ctx, "add", "entries")
	if err != nil {
		return fmt.Errorf("failed to stage entries: %s", stderr)
	}

	// Check if there are staged changes to commit
	status, _, _ := s.run(ctx, "status", "--porcelain", "entries")
	if status != "" {
		_, stderr, err := s.run(ctx, "commit", "-m", "sync: auto-commit financial entries")
		if err != nil {
			return fmt.Errorf("failed to commit entries: %s", stderr)
		}
	}

	// Step 2: Pull with rebase
	if s.stepLogger != nil {
		fmt.Fprintln(s.stepLogger, "[2/3] Descargando cambios remotos (pull rebase)...")
	}

	_, stderr, err = s.run(ctx, "pull", "--rebase", s.remote, branch)
	if err != nil {
		lowerErr := strings.ToLower(stderr)
		if strings.Contains(lowerErr, "couldn't find remote ref") {
			// If the branch wasn't found on remote, attempt alternate standard branch (master <-> main)
			altBranch := "main"
			if branch == "main" {
				altBranch = "master"
			}
			_, altStderr, altErr := s.run(ctx, "pull", "--rebase", s.remote, altBranch)
			if altErr == nil {
				branch = altBranch
			} else if !strings.Contains(strings.ToLower(altStderr), "couldn't find remote ref") {
				return s.classifyGitError(stderr, err)
			}
		} else {
			return s.classifyGitError(stderr, err)
		}
	}

	// Verify whether we have any commits in the repository to push
	if _, _, err := s.run(ctx, "rev-parse", "HEAD"); err != nil {
		// Both local and remote have 0 commits; nothing to push yet.
		return nil
	}

	// Step 3: Push to remote
	if s.stepLogger != nil {
		fmt.Fprintln(s.stepLogger, "[3/3] Enviando datos al repositorio remoto (push)...")
	}

	_, stderr, err = s.run(ctx, "push", "-u", s.remote, branch)
	if err != nil {
		return s.classifyGitError(stderr, err)
	}

	return nil
}
