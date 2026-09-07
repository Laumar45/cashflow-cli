# 💸 CashFlow CLI (`cash`)

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS%20%7C%20Termux-blue)](https://github.com/)

> **Fast, offline-first personal cash flow tracker for the terminal.**  
> Powered by an immutable event log, multi-device Git sync, and an interactive Bubble Tea dashboard.

---

## ⚡ Features

- **Blazing Fast (<15ms):** Instant command execution for logging expenses and income on the fly.
- **Append-Only Event Log:** Each transaction is saved as an individual ULID-named JSON file (`~/.cashflow/entries/<ulid>.json`). Zero database locks, zero merge conflicts.
- **Multi-Device Git Sync:** Native Git adapter automates staging, committing, and syncing your financial log across machines, laptops, and mobile devices (Termux).
- **Interactive TUI Dashboard (`cash tui`):** Built with Charm's [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss). Includes responsive layout, embedded Braille dollar logo, transaction ledger, and accessibility markers (`▲` / `▼`).
- **Clean Architecture:** Strict boundary separation (Domain, Ports, Use Cases, Adapters).
- **Self-Contained Binary:** All UI assets and Braille art are embedded at compile-time via `//go:embed`.

---

## 🚀 Installation

### Option 1: Automated Windows Installer (PowerShell)

Clones or compiles the project, moves `cash.exe` to `%LOCALAPPDATA%\cashflow\bin`, and registers it permanently in your User `PATH`:

```powershell
# From the repository root:
.\install.ps1
```

### Option 2: Automated Unix / macOS / Termux Installer (Bash)

Compiles `cash` and installs it to `~/.local/bin` (or `$PREFIX/bin` under Android Termux):

```bash
# From the repository root:
chmod +x install.sh
./install.sh
```

### Option 3: Go Developer Install

If you have Go installed and your `$GOPATH/bin` in your `PATH`:

```bash
go install ./cmd/cash
```

---

## 🔄 Updating to Latest Version

When pulling changes made from another machine or updating your local checkout, run the automated updater from the repository root:

**Windows (PowerShell):**
```powershell
.\update.ps1
```

**Linux / macOS / Termux (Bash):**
```bash
chmod +x update.sh
./update.sh
```

This runs `git pull --rebase` and automatically recompiles and replaces the active binary.

---

## 📖 Quickstart & Usage

### 1. Initialize CashFlow

Set up the storage directory structure in `~/.cashflow/`:

```bash
# Offline-only local setup:
cash init

# Automatic multi-device sync setup (detects or creates a private 'cashflow-data' repo on GitHub via gh):
cash init --auto

# Or manually link an existing remote Git repository:
cash init https://github.com/username/cashflow-data.git
```

### 2. Record Transactions

```bash
# Record income
cash in 1500.00 "Freelance website design" -c income

# Record an expense
cash out 42.50 "Supermarket groceries" -c food
cash out 12.00 "Cinema ticket" -c entertainment
```

### 3. Review Summary & List

```bash
# View aggregated totals and breakdown by category
cash summary

# List recent transactions
cash list

# List all available categories
cash categories
```

### 4. Launch the Interactive Dashboard (TUI)

```bash
cash tui
```

* **Keyboard Controls:**
  * `↑` / `k` — Navigate up in the transaction ledger
  * `↓` / `j` — Navigate down in the transaction ledger
  * `q` / `Ctrl+C` — Exit dashboard

### 5. Multi-Device Git Sync

Sync your immutable transaction files with your private Git repository:

```bash
# Link your private storage repository automatically:
cash init --auto

# Synchronize local ledger with remote Git repository:
cash sync
```

---

## 🏗 Architecture & Storage Model

```text
cmd/
  └── cash/                     # Application entrypoint
internal/
  ├── domain/                   # Pure entities (Transaction, Money, Category)
  ├── ports/                    # Repository & VCS Interfaces
  ├── usecases/                 # Core business flows (Record, Summarize, Sync)
  └── adapters/
      ├── storage/              # Immutable file-system event log (ULID JSON)
      ├── vcs/                  # Git automation service
      ├── cli/                  # Cobra CLI command suite
      └── tui/                  # Bubble Tea + Lip Gloss terminal dashboard
```

### Why Append-Only JSON Files?

Traditional single-file ledgers (like SQLite or a single `transactions.json`) suffer from write contention and inevitable Git merge conflicts when used across multiple devices.

CashFlow generates a unique **ULID** (Universally Unique Lexicographically Sortable Identifier) for every single entry:
```text
~/.cashflow/
  ├── entries/
  │   ├── 01ARZ3NDEKTSV4RRFFQ69G5FAV.json
  │   └── 01ARZ3NDEKTSV4RRFFQ69G5FAW.json
  └── config.json
```
Since new transactions are always distinct files, **Git merges are always 100% conflict-free**.

---

## 🧪 Testing

Run the test suite across all architectural layers:

```bash
go test -v ./...
```

---

## 📄 License

MIT License.
