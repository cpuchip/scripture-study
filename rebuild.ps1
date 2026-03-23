#!/usr/bin/env pwsh
# Rebuild all MCP servers and tools in the scripture-study workspace.
# Usage: .\rebuild.ps1            — rebuild everything
#        .\rebuild.ps1 gospel-vec  — rebuild just one server

param(
    [string]$Target
)

$ErrorActionPreference = "Stop"
$root = $PSScriptRoot

# Each entry: Name, Directory (relative to scripts/), build commands
# Build commands are [ordered] pairs of output-name → build-path
$servers = @(
    @{ Name = "gospel-vec";      Dir = "gospel-vec";      Builds = @{ "gospel-vec.exe" = "." } }
    @{ Name = "gospel-mcp";      Dir = "gospel-mcp";      Builds = @{ "gospel-mcp.exe" = "./cmd/gospel-mcp" } }
    @{ Name = "search-mcp";      Dir = "search-mcp";      Builds = @{ "search-mcp.exe" = "./cmd/search-mcp" } }
    @{ Name = "webster-mcp";     Dir = "webster-mcp";     Builds = @{ "webster-mcp.exe" = "./cmd/webster-mcp" } }
    @{ Name = "yt-mcp";          Dir = "yt-mcp";          Builds = @{ "yt-mcp.exe" = "." } }
    @{ Name = "byu-citations";   Dir = "byu-citations";   Builds = @{ "byu-citations.exe" = "./cmd/byu-citations" } }
    @{ Name = "becoming";        Dir = "becoming";        Builds = @{ "mcp.exe" = "./cmd/mcp"; "server.exe" = "./cmd/server" } }
    @{ Name = "brain";           Dir = "brain";           Builds = @{ "brain.exe" = "./cmd/brain"; "brain-mcp.exe" = "./cmd/brain-mcp"; "brain-cli.exe" = "./cmd/brain-cli" } }
    @{ Name = "session-journal"; Dir = "session-journal"; Builds = @{ "session-journal.exe" = "./cmd/session-journal" } }
    @{ Name = "publish";         Dir = "publish";         Builds = @{ "publish.exe" = "./cmd" } }
)

if ($Target) {
    $servers = $servers | Where-Object { $_.Name -eq $Target }
    if (-not $servers) {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Write-Host "Available: $( ($servers | ForEach-Object { $_.Name }) -join ', ' )"
        exit 1
    }
}

$failed = @()
$succeeded = @()

foreach ($srv in $servers) {
    $dir = Join-Path $root "scripts" $srv.Dir
    if (-not (Test-Path $dir)) {
        Write-Host "  SKIP $($srv.Name) — directory not found" -ForegroundColor Yellow
        continue
    }

    Push-Location $dir
    try {
        foreach ($exe in $srv.Builds.Keys) {
            $pkg = $srv.Builds[$exe]
            Write-Host "  Building $($srv.Name)/$exe ... " -NoNewline
            go build -o $exe $pkg 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "FAILED" -ForegroundColor Red
                $failed += "$($srv.Name)/$exe"
            } else {
                Write-Host "OK" -ForegroundColor Green
                $succeeded += "$($srv.Name)/$exe"
            }
        }
    } finally {
        Pop-Location
    }
}

Write-Host ""
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host "  Succeeded: $($succeeded.Count)" -ForegroundColor Green
if ($failed.Count -gt 0) {
    Write-Host "  Failed:    $($failed.Count)" -ForegroundColor Red
    $failed | ForEach-Object { Write-Host "    - $_" -ForegroundColor Red }
    exit 1
} else {
    Write-Host "  All builds passed." -ForegroundColor Green
}
