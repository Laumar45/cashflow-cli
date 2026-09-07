# update.ps1 — Windows Automated Updater for CashFlow CLI
# Run with: .\update.ps1

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  CashFlow CLI — Actualizador Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$rootDir = $PSScriptRoot
if (-not $rootDir) { $rootDir = Get-Location }

Push-Location $rootDir
try {
    # 1. Pull latest changes
    Write-Host "[1/2] Obteniendo ultimos cambios del repositorio (git pull)..." -ForegroundColor Yellow
    git pull --rebase --autostash
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Error al actualizar con git pull. Verifica el estado de tu repositorio."
        exit 1
    }

    # 2. Run install.ps1 to build and install
    Write-Host "[2/2] Recompilando e instalando nueva version..." -ForegroundColor Yellow
    & "$rootDir\install.ps1"
} finally {
    Pop-Location
}
