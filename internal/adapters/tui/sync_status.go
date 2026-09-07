package tui

import (
	"bytes"
	"os/exec"
	"strings"
)

// SyncStatusInfo contains local read-only Git status for the TUI footer.
type SyncStatusInfo struct {
	LastSyncRelative string
	UnsyncedCount    int
}

// GetLocalSyncStatus queries local Git commits and porcelain status without network I/O.
func GetLocalSyncStatus(repoDir string) SyncStatusInfo {
	info := SyncStatusInfo{
		LastSyncRelative: "sin sincronizar",
		UnsyncedCount:    0,
	}

	if repoDir == "" {
		return info
	}

	// 1. Get relative timestamp of last commit
	cmdLog := exec.Command("git", "log", "-1", "--format=%cr")
	cmdLog.Dir = repoDir
	var outLog bytes.Buffer
	cmdLog.Stdout = &outLog
	if err := cmdLog.Run(); err == nil {
		rel := strings.TrimSpace(outLog.String())
		if rel != "" {
			info.LastSyncRelative = rel
		}
	}

	// 2. Count modified/untracked files in entries directory
	cmdStatus := exec.Command("git", "status", "--porcelain", "entries")
	cmdStatus.Dir = repoDir
	var outStatus bytes.Buffer
	cmdStatus.Stdout = &outStatus
	if err := cmdStatus.Run(); err == nil {
		raw := strings.TrimSpace(outStatus.String())
		if raw != "" {
			lines := strings.Split(raw, "\n")
			info.UnsyncedCount = len(lines)
		}
	}

	return info
}
