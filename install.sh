#!/usr/bin/env bash
# install.sh — Linux / macOS / Termux Automated Installer for CashFlow CLI
# Run with: ./install.sh

set -e

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  CashFlow CLI — Installer (Unix/Termux)${NC}"
echo -e "${CYAN}========================================${NC}"

# 1. Verify Go
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}Error: Go no está instalado o no se encuentra en tu PATH.${NC}"
    echo "Instálalo desde https://go.dev/dl/ o mediante tu gestor de paquetes (ej: pkg install golang en Termux)."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}[1/3] Compilando binario de producción (cash)...${NC}"
go build -ldflags="-s -w" -o cash ./cmd/cash

# 2. Determine installation directory
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
    # Termux on Android
    TARGET_DIR="$PREFIX/bin"
elif [ "$(id -u)" -eq 0 ]; then
    # Running as root
    TARGET_DIR="/usr/local/bin"
else
    # Non-root user: use ~/.local/bin
    TARGET_DIR="$HOME/.local/bin"
    mkdir -p "$TARGET_DIR"
fi

echo -e "${YELLOW}[2/3] Instalando en: $TARGET_DIR...${NC}"
install -m 755 cash "$TARGET_DIR/cash"
rm -f cash

# 3. Check PATH
echo -e "${YELLOW}[3/3] Verificando PATH...${NC}"
if [[ ":$PATH:" != *":$TARGET_DIR:"* ]]; then
    echo -e "${YELLOW}Aviso: '$TARGET_DIR' no se encuentra en tu PATH.${NC}"
    echo "Añade la siguiente línea a tu ~/.bashrc o ~/.zshrc:"
    echo -e "  ${CYAN}export PATH=\"\$PATH:$TARGET_DIR\"${NC}"
else
    echo -e "${GREEN}✔ El directorio ya está en tu PATH.${NC}"
fi

# 4. Verify which binary takes precedence
ACTIVE_BIN="$(command -v cash 2>/dev/null || true)"
if [ -n "$ACTIVE_BIN" ] && [ "$ACTIVE_BIN" != "$TARGET_DIR/cash" ]; then
    echo ""
    echo -e "${RED}⚠️  ADVERTENCIA: Existe otro binario 'cash' con mayor prioridad en tu PATH:${NC}"
    echo -e "${YELLOW}   Ruta activa: $ACTIVE_BIN${NC}"
    echo -e "${YELLOW}   Ruta recién instalada: $TARGET_DIR/cash${NC}"
    echo -e "${RED}   Elimina el binario antiguo para evitar ejecutar una versión obsoleta.${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ✔ Instalación completada con éxito!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Ahora puedes usar 'cash' en tu terminal:"
echo -e "  ${CYAN}cash --help${NC}     (ver lista de comandos)"
echo -e "  ${CYAN}cash in ...${NC}     (registrar un ingreso)"
echo -e "  ${CYAN}cash out ...${NC}    (registrar un gasto)"
echo -e "  ${CYAN}cash tui${NC}        (abrir dashboard interactivo)"
echo ""
