# install.ps1 — Windows Automated Installer for CashFlow CLI
# Run with: .\install.ps1

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  CashFlow CLI — Instalador Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 1. Verify Go toolchain
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Error: Go no esta instalado o no se encuentra en tu PATH. Descargalo desde https://go.dev/dl/"
    exit 1
}

Write-Host "[1/3] Compilando binario de produccion (cash.exe)..." -ForegroundColor Yellow
$rootDir = $PSScriptRoot
if (-not $rootDir) { $rootDir = Get-Location }

Push-Location $rootDir
try {
    go build -ldflags="-s -w" -o cash.exe ./cmd/cash
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Error durante la compilacion con Go."
        exit 1
    }
} finally {
    Pop-Location
}

# 2. Define user installation folder
$installDir = Join-Path $env:LOCALAPPDATA "cashflow\bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$targetExe = Join-Path $installDir "cash.exe"
$sourceExe = Join-Path $rootDir "cash.exe"

Write-Host "[2/3] Instalando en: $installDir..." -ForegroundColor Yellow
Copy-Item -Path $sourceExe -Destination $targetExe -Force

# 3. Ensure installDir is in user PATH
Write-Host "[3/3] Configurando variable de entorno PATH..." -ForegroundColor Yellow
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$pathParts = $userPath -split ";"

if ($pathParts -notcontains $installDir) {
    $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "✔ Directorio anadido permanentemente a tu PATH de usuario." -ForegroundColor Green
} else {
    Write-Host "✔ El directorio ya se encuentra en tu PATH." -ForegroundColor Green
}

# Update PATH in current session
if ($env:PATH -notmatch [regex]::Escape($installDir)) {
    $env:PATH = "$env:PATH;$installDir"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  ✔ Instalacion completada con exito!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "Ahora puedes usar 'cash' en cualquier ventana de PowerShell o Terminal:" -ForegroundColor White
Write-Host "  cash --help     (ver lista de comandos)" -ForegroundColor Cyan
Write-Host "  cash in ...     (registrar un ingreso)" -ForegroundColor Cyan
Write-Host "  cash out ...    (registrar un gasto)" -ForegroundColor Cyan
Write-Host "  cash tui        (abrir el dashboard interactivo)" -ForegroundColor Cyan
Write-Host ""
