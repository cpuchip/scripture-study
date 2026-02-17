<#
.SYNOPSIS
    Build and reindex both gospel-mcp (FTS) and gospel-vec (vector) databases.

.DESCRIPTION
    This script builds both search servers and runs a full reindex:

    1. gospel-mcp  — SQLite FTS5 full-text search (fast, seconds)
    2. gospel-vec  — Vector embeddings + LLM summaries (slow, hours)

    Content indexed: scriptures, conference talks, manuals, books, music.

.PARAMETER NoSummary
    Skip LLM summary/theme generation for gospel-vec (faster, cache-only)

.PARAMETER VecOnly
    Skip gospel-mcp build and reindex (only do gospel-vec)

.PARAMETER FtsOnly
    Skip gospel-vec build and reindex (only do gospel-mcp)

.PARAMETER Volumes
    Comma-separated scripture volumes for gospel-vec index command.
    Only applies when running gospel-vec in scripture-only mode (not index-all).
    Options: bofm, dc-testament/dc, pgp, nt, ot

.PARAMETER Source
    gospel-mcp: Index only a specific source type.
    Options: scriptures, conference, manual, magazine, books, music

.EXAMPLE
    .\reindex-scriptures.ps1
    # Full reindex of both databases (all content)

.EXAMPLE
    .\reindex-scriptures.ps1 -NoSummary
    # Full reindex without generating new LLM summaries

.EXAMPLE
    .\reindex-scriptures.ps1 -FtsOnly
    # Rebuild only the FTS database (gospel-mcp), fast

.EXAMPLE
    .\reindex-scriptures.ps1 -VecOnly -NoSummary
    # Rebuild only the vector database, skip summaries

.EXAMPLE
    .\reindex-scriptures.ps1 -Source scriptures
    # gospel-mcp: only reindex scriptures; gospel-vec: index-all as usual
#>

param(
    [switch]$NoSummary,
    [switch]$VecOnly,
    [switch]$FtsOnly,
    [string]$Volumes = "",
    [string]$Source = ""
)

$ErrorActionPreference = "Stop"
$scriptRoot = $PSScriptRoot
$startTime = Get-Date

# ─────────────────────────────────────────────
# Phase 1: gospel-mcp (FTS5 full-text search)
# ─────────────────────────────────────────────
if (-not $VecOnly) {
    $gospelMcpDir = Join-Path $scriptRoot "scripts\gospel-mcp"
    Write-Host "`n══════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  Phase 1: gospel-mcp (FTS5 full-text search)" -ForegroundColor Cyan
    Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan

    Push-Location $gospelMcpDir
    try {
        # Build
        Write-Host "`nBuilding gospel-mcp..." -ForegroundColor Yellow
        go build -tags "fts5" -o gospel-mcp.exe ./cmd/gospel-mcp/
        if ($LASTEXITCODE -ne 0) { throw "gospel-mcp build failed" }
        Write-Host "Build successful" -ForegroundColor Green

        # Index
        Write-Host "`nIndexing gospel-mcp (full reindex)..." -ForegroundColor Yellow
        $ftsArgs = @("index", "--force", "--root", $scriptRoot)
        if ($Source) {
            $ftsArgs += @("--source", $Source)
            Write-Host "   Source filter: $Source" -ForegroundColor Cyan
        }

        $ftsStart = Get-Date
        & .\gospel-mcp.exe @ftsArgs
        if ($LASTEXITCODE -ne 0) { throw "gospel-mcp indexing failed" }

        $ftsElapsed = (Get-Date) - $ftsStart
        Write-Host "`ngospel-mcp indexing complete ($($ftsElapsed.ToString('mm\:ss')))" -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

# ─────────────────────────────────────────────
# Phase 2: gospel-vec (vector embeddings)
# ─────────────────────────────────────────────
if (-not $FtsOnly) {
    $gospelVecDir = Join-Path $scriptRoot "scripts\gospel-vec"
    Write-Host "`n══════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  Phase 2: gospel-vec (vector embeddings)" -ForegroundColor Cyan
    Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan

    Push-Location $gospelVecDir
    try {
        # Build
        Write-Host "`nBuilding gospel-vec..." -ForegroundColor Yellow
        go build -o gospel-vec.exe .
        if ($LASTEXITCODE -ne 0) { throw "gospel-vec build failed" }
        Write-Host "Build successful" -ForegroundColor Green

        # Test LM Studio connection
        Write-Host "`nTesting LM Studio connection..." -ForegroundColor Yellow
        .\gospel-vec.exe test
        if ($LASTEXITCODE -ne 0) {
            Write-Host "LM Studio connection test had issues, continuing anyway..." -ForegroundColor Yellow
        }

        # Run index-all (scriptures + talks + manuals + music)
        $vecStart = Get-Date
        $vecArgs = @("index-all")
        if ($NoSummary) {
            $vecArgs += @("-no-summary", "-no-theme")
        }

        Write-Host "`nStarting full vector indexing..." -ForegroundColor Yellow
        Write-Host "   This covers scriptures, talks, manuals, and music." -ForegroundColor Cyan
        if ($NoSummary) {
            Write-Host "   Summaries/themes: SKIPPED (cache-only)" -ForegroundColor Cyan
        } else {
            Write-Host "   Summaries/themes: generating (this will take hours)" -ForegroundColor Cyan
        }

        & .\gospel-vec.exe @vecArgs
        if ($LASTEXITCODE -ne 0) { throw "gospel-vec indexing failed" }

        $vecElapsed = (Get-Date) - $vecStart
        Write-Host "`ngospel-vec indexing complete ($($vecElapsed.ToString('hh\:mm\:ss')))" -ForegroundColor Green

        # Show database stats
        Write-Host "`nDatabase statistics:" -ForegroundColor Yellow
        .\gospel-vec.exe stats
    } finally {
        Pop-Location
    }
}

# ─────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────
$totalElapsed = (Get-Date) - $startTime
Write-Host "`n══════════════════════════════════════════" -ForegroundColor Green
Write-Host "  Reindex complete!" -ForegroundColor Green
Write-Host "  Total time: $($totalElapsed.ToString('hh\:mm\:ss'))" -ForegroundColor Green
Write-Host "══════════════════════════════════════════" -ForegroundColor Green
