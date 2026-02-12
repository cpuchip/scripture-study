# Becoming App — build & run with TLS (for Google OAuth local dev)
# Usage: .\start-ssl.ps1 [-Dev] [-SkipBuild] [-Port 8443]
#
# Prerequisites: mkcert (install via: winget install FiloSottile.mkcert)
# First run will generate trusted local certificates in .certs/
param(
    [switch]$Dev,
    [switch]$SkipBuild,
    [int]$Port = 8443
)

$ErrorActionPreference = 'Stop'
$base = $PSScriptRoot

Write-Host "`n  Becoming App (TLS)" -ForegroundColor Cyan
Write-Host "  ==================`n"

# --- TLS certificate setup ---
$certDir = "$base\.certs"
$cert = "$certDir\localhost.pem"
$key = "$certDir\localhost-key.pem"

if (-not (Test-Path $cert)) {
    if (-not (Get-Command mkcert -ErrorAction SilentlyContinue)) {
        Write-Host "  mkcert not found!" -ForegroundColor Red
        Write-Host "  Install: winget install FiloSottile.mkcert" -ForegroundColor Yellow
        Write-Host "  Then run: mkcert -install`n" -ForegroundColor Yellow
        exit 1
    }

    Write-Host "  Generating TLS certificates..." -ForegroundColor Yellow
    if (-not (Test-Path $certDir)) { New-Item -ItemType Directory -Path $certDir | Out-Null }

    mkcert -install 2>$null
    Push-Location $certDir
    mkcert localhost 127.0.0.1 ::1
    # mkcert names files "localhost+2.pem" / "localhost+2-key.pem"
    Get-ChildItem "*+*-key.pem" | Rename-Item -NewName "localhost-key.pem"
    Get-ChildItem "*+*.pem" | Where-Object { $_.Name -notmatch "key" } | Rename-Item -NewName "localhost.pem"
    Pop-Location

    Write-Host "  Certificates created in .certs/`n" -ForegroundColor Green
}

# Set Google OAuth redirect URL for local HTTPS
$env:GOOGLE_REDIRECT_URL = "https://localhost:${Port}/auth/google/callback"

# Stop any running server
Get-Process -Name server -ErrorAction SilentlyContinue | Stop-Process -Force

if (-not $SkipBuild) {
    # 1. Build frontend
    Write-Host "  [1/3] Building frontend..." -ForegroundColor Yellow
    Push-Location "$base\frontend"
    npm run build
    if ($LASTEXITCODE -ne 0) { Pop-Location; throw "Frontend build failed" }
    Pop-Location

    # 2. Copy dist into Go embed directory
    Write-Host "  [2/3] Copying dist..." -ForegroundColor Yellow
    $dist = "$base\cmd\server\dist"
    if (Test-Path $dist) { Remove-Item -Recurse -Force $dist }
    Copy-Item -Recurse "$base\frontend\dist" $dist

    # 3. Build Go binary
    Write-Host "  [3/3] Building server..." -ForegroundColor Yellow
    Push-Location $base
    go build -o server.exe ./cmd/server/
    if ($LASTEXITCODE -ne 0) { Pop-Location; throw "Go build failed" }
    Pop-Location
}

# Run with TLS
$scriptures = Resolve-Path "$base\..\..\gospel-library\eng\scriptures"
$serverArgs = @(
    "-db", "$base\becoming.db",
    "-scriptures", $scriptures,
    "-addr", ":$Port",
    "-tls-cert", $cert,
    "-tls-key", $key
)
if ($Dev) { $serverArgs += "-dev" }

Write-Host "`n  Starting server with TLS..." -ForegroundColor Green
if ($Dev) {
    Write-Host "  Dev mode: run 'cd frontend && npm run dev' in another terminal"
    Write-Host "  Frontend: http://localhost:5173"
}
Write-Host "  Server:   https://localhost:$Port" -ForegroundColor Cyan
Write-Host "  OAuth:    $env:GOOGLE_REDIRECT_URL`n" -ForegroundColor DarkGray

& "$base\server.exe" @serverArgs
