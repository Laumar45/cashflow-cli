#!/usr/bin/env bash
# update.sh — Linux / macOS / Termux Automated Updater for CashFlow CLI
# Run with: ./update.sh

set -e

CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  CashFlow CLI — Updater (Unix/Termux)${NC}"
echo -e "${CYAN}========================================${NC}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}[1/2] Obteniendo últimos cambios del repositorio (git pull)...${NC}"
git pull --rebase --autostash

echo -e "${YELLOW}[2/2] Recompilando e instalando nueva versión...${NC}"
"$SCRIPT_DIR/install.sh"
