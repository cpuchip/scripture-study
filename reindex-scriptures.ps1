<#
.SYNOPSIS
    Reindex all standard works in the gospel-vec vector database.

.DESCRIPTION
    This script builds gospel-vec and reindexes all scripture volumes:
    - Book of Mormon (bofm)
    - Doctrine & Covenants (dc-testament/dc)
    - Pearl of Great Price (pgp)
    - New Testament (nt)
    - Old Testament (ot)

    Includes verse, paragraph, summary, and theme layers.

.PARAMETER NoSummary
    Skip LLM summary/theme generation (faster, cache-only)

.PARAMETER Volumes
    Comma-separated list of volumes to index. Default: all
    Options: bofm, dc-testament/dc, pgp, nt, ot

.EXAMPLE
    .\reindex-scriptures.ps1
    # Full reindex with summaries

.EXAMPLE
    .\reindex-scriptures.ps1 -NoSummary
    # Quick reindex without generating new summaries

.EXAMPLE
    .\reindex-scriptures.ps1 -Volumes "nt,ot"
    # Only index New/Old Testament
#>

param(
    [switch]$NoSummary,
    [string]$Volumes = "bofm,dc-testament/dc,pgp,nt,ot"
)

$ErrorActionPreference = "Stop"

# Navigate to gospel-vec directory
$scriptRoot = $PSScriptRoot
$gospelVecDir = Join-Path $scriptRoot "scripts\gospel-vec"

Write-Host "📍 Working directory: $gospelVecDir" -ForegroundColor Cyan
Push-Location $gospelVecDir

try {
    # Build gospel-vec
    Write-Host "`n🔨 Building gospel-vec..." -ForegroundColor Yellow
    go build -o gospel-vec.exe .
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }
    Write-Host "✅ Build successful" -ForegroundColor Green

    # Test LM Studio connection
    Write-Host "`n🔌 Testing LM Studio connection..." -ForegroundColor Yellow
    .\gospel-vec.exe test
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠️  LM Studio connection test had issues, continuing anyway..." -ForegroundColor Yellow
    }

    # Build index command
    $layers = "verse,paragraph"
    if (-not $NoSummary) {
        $layers = "verse,paragraph,summary,theme"
    }

    Write-Host "`n📚 Starting indexing..." -ForegroundColor Yellow
    Write-Host "   Volumes: $Volumes" -ForegroundColor Cyan
    Write-Host "   Layers: $layers" -ForegroundColor Cyan

    $startTime = Get-Date

    # Run indexing
    if ($NoSummary) {
        .\gospel-vec.exe index -volumes $Volumes -layers $layers
    } else {
        .\gospel-vec.exe index -volumes $Volumes -layers $layers -summary
    }

    if ($LASTEXITCODE -ne 0) {
        throw "Indexing failed"
    }

    $elapsed = (Get-Date) - $startTime
    Write-Host "`n✅ Indexing complete!" -ForegroundColor Green
    Write-Host "   Time elapsed: $($elapsed.ToString('hh\:mm\:ss'))" -ForegroundColor Cyan

    # Show database stats
    Write-Host "`n📊 Database statistics:" -ForegroundColor Yellow
    .\gospel-vec.exe stats

} finally {
    Pop-Location
}

Write-Host "`n🎉 Done! The gospel-vec database is ready for searching." -ForegroundColor Green
