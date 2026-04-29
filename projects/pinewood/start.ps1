# Pinewood Derby — build & run
# Usage: .\start.ps1 [-SkipBuild] [-Port 8765] [-Db derby.db] [-Log derby.log]
#
# Prerequisites:
#   - Go 1.23+
#   - Node.js 20+ (npm run build for the embedded SPA)
param(
    [switch]$SkipBuild,
    [int]$Port = 8080,
    [string]$Db = 'derby.db',
    [string]$Log = 'derby.log',
    [switch]$UseGoRun
)

$ErrorActionPreference = 'Stop'
$base = $PSScriptRoot

Write-Host "`n  Pinewood Derby" -ForegroundColor Cyan
Write-Host "  ==============`n"

# Stop any running pinewood process
Get-Process -Name pinewood -ErrorAction SilentlyContinue | Stop-Process -Force

if (-not $SkipBuild) {
    # 1. Build frontend (Vite -> ../cmd/pinewood/dist/, embedded via go:embed)
    Write-Host "  [1/2] Building frontend..." -ForegroundColor Yellow
    Push-Location "$base\frontend"
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "         npm install..." -ForegroundColor DarkGray
            npm install
            if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
        }
        npm run build
        if ($LASTEXITCODE -ne 0) { throw "Frontend build failed" }
    }
    finally { Pop-Location }

    # 2. Build Go binary
    Write-Host "  [2/2] Building server..." -ForegroundColor Yellow
    Push-Location $base
    try {
        go build -o pinewood.exe ./cmd/pinewood
        if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
    }
    finally { Pop-Location }
}

Write-Host "`n  Starting pinewood server..." -ForegroundColor Green
Write-Host "  Web UI: http://localhost:$Port" -ForegroundColor Cyan
Write-Host "  DB:     $Db" -ForegroundColor DarkGray
Write-Host "  Log:    $Log" -ForegroundColor DarkGray
Write-Host "  Press Ctrl+C to stop`n" -ForegroundColor DarkGray

Push-Location $base
try {
    if ($UseGoRun) {
        go run ./cmd/pinewood serve -addr ":$Port" -db $Db -log $Log
    } else {
        & "$base\pinewood.exe" serve -addr ":$Port" -db $Db -log $Log
    }
}
catch {
    $msg = $_.Exception.Message
    if ($msg -match "Application Control policy has blocked this file") {
        Write-Host "  pinewood.exe blocked by Application Control policy." -ForegroundColor Yellow
        Write-Host "  Falling back to: go run ./cmd/pinewood ..." -ForegroundColor Yellow
        go run ./cmd/pinewood serve -addr ":$Port" -db $Db -log $Log
    } else {
        throw
    }
}
finally {
    Pop-Location
}
