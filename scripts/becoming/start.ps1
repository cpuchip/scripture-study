# Becoming App — build & run
# Usage: .\start.ps1 [-Dev] [-SkipBuild]
param(
    [switch]$Dev,       # Run in dev mode (CORS enabled, use Vite on :5173 for hot reload)
    [switch]$SkipBuild  # Skip frontend/backend build, just start the server
)

$ErrorActionPreference = 'Stop'
$base = $PSScriptRoot

Write-Host "`n  Becoming App" -ForegroundColor Cyan
Write-Host "  ============`n"

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

# Run
$scriptures = Resolve-Path "$base\..\..\gospel-library\eng\scriptures"
$args = @("-db", "$base\becoming.db", "-scriptures", $scriptures)
if ($Dev) { $args += "-dev" }

Write-Host "`n  Starting server..." -ForegroundColor Green
if ($Dev) {
    Write-Host "  Dev mode: run 'cd frontend && npm run dev' in another terminal"
    Write-Host "  Frontend: http://localhost:5173"
}
Write-Host "  Server:   http://localhost:8080`n" -ForegroundColor Cyan

& "$base\server.exe" @args
