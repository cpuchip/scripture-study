<#
.SYNOPSIS
    Run a model experiment for gospel-vec to evaluate embedding/summary quality.

.DESCRIPTION
    Indexes a small benchmark subset of scripture content into an isolated data
    directory, then runs benchmark queries and outputs scored results for
    evaluation. Each experiment gets its own data directory so production data
    is never touched.

    The script:
    1. Builds gospel-vec
    2. Indexes a small set of benchmark volumes (default: bofm only)
    3. Runs benchmark queries from experiments/benchmark-queries.json
    4. Outputs results to experiments/results/<experiment-name>.json
    5. Appends a summary to experiments/experiment-log.md

    Before running, load the desired models in LM Studio:
    - Embedding model at the /v1/embeddings endpoint
    - Chat model for summaries (optional, use -NoSummary to skip)

.PARAMETER Name
    Name for this experiment (used in output filenames).
    Default: auto-generated from timestamp.

.PARAMETER EmbeddingModel
    Override the embedding model name (must match what's loaded in LM Studio).
    Default: text-embedding-qwen3-embedding-4b

.PARAMETER ChatModel
    Override the chat model name for summaries.
    Default: auto-detected from LM Studio.

.PARAMETER Volumes
    Comma-separated scripture volumes to index for the benchmark.
    Default: bofm (Book of Mormon only, for speed)

.PARAMETER NoSummary
    Skip LLM summary/theme generation (test embeddings only).

.PARAMETER SearchLayers
    Comma-separated layers to search during evaluation.
    Default: verse,paragraph

.PARAMETER Limit
    Max results per query. Default: 10

.PARAMETER SkipIndex
    Skip the indexing step (reuse existing experiment data).
    Useful for re-running queries with different search parameters.

.EXAMPLE
    .\run-experiment.ps1 -Name "qwen3-4b-baseline"
    # Run with default embedding model, bofm only

.EXAMPLE
    .\run-experiment.ps1 -Name "nomic-embed" -EmbeddingModel "nomic-embed-text-v1.5"
    # Test a different embedding model

.EXAMPLE
    .\run-experiment.ps1 -Name "with-summaries" -Volumes "bofm,nt" -SearchLayers "verse,paragraph,summary"
    # Include NT and search summaries too

.EXAMPLE
    .\run-experiment.ps1 -Name "qwen3-4b-baseline" -SkipIndex -SearchLayers "verse,paragraph,summary"
    # Re-run queries on existing index with different search layers
#>

param(
    [string]$Name = "",
    [string]$EmbeddingModel = "",
    [string]$ChatModel = "",
    [string]$Volumes = "bofm",
    [switch]$NoSummary,
    [string]$SearchLayers = "verse,paragraph",
    [int]$Limit = 10,
    [switch]$SkipIndex
)

$ErrorActionPreference = "Stop"
$scriptRoot = $PSScriptRoot
$gospelVecDir = Join-Path $scriptRoot ".."  # scripts/gospel-vec/

# Auto-generate experiment name if not provided
if (-not $Name) {
    $Name = "exp-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
}

# Paths
$experimentsDir = Join-Path $gospelVecDir "experiments"
$resultsDir = Join-Path $experimentsDir "results"
$benchmarkFile = Join-Path $experimentsDir "benchmark-queries.json"
$resultFile = Join-Path $resultsDir "$Name.json"
$logFile = Join-Path $experimentsDir "experiment-log.md"

# Ensure directories exist
New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null

# Isolated data directory for this experiment
$expDataDir = Join-Path $experimentsDir "data-$Name"

Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  gospel-vec Experiment: $Name" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  Data dir:    $expDataDir" -ForegroundColor DarkGray
Write-Host "  Volumes:     $Volumes" -ForegroundColor DarkGray
Write-Host "  Layers:      $SearchLayers" -ForegroundColor DarkGray
Write-Host "  Summaries:   $(if ($NoSummary) { 'OFF' } else { 'ON' })" -ForegroundColor DarkGray

Push-Location $gospelVecDir
try {
    # ─────────────────────────────────────
    # Step 1: Build gospel-vec
    # ─────────────────────────────────────
    Write-Host "`n🔨 Building gospel-vec..." -ForegroundColor Yellow
    go build -o gospel-vec.exe .
    if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    Write-Host "✅ Build successful" -ForegroundColor Green

    # ─────────────────────────────────────
    # Step 2: Detect models from LM Studio
    # ─────────────────────────────────────
    Write-Host "`n🔌 Testing LM Studio connection..." -ForegroundColor Yellow
    .\gospel-vec.exe test
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠️  LM Studio test had issues, continuing..." -ForegroundColor Yellow
    }

    # Query loaded models via LM Studio API for logging
    $actualEmbeddingModel = if ($EmbeddingModel) { $EmbeddingModel } else { "text-embedding-qwen3-embedding-4b" }
    $actualChatModel = $ChatModel
    try {
        $modelsResponse = Invoke-RestMethod -Uri "http://localhost:1234/v1/models" -Method Get -ErrorAction SilentlyContinue
        $loadedModels = $modelsResponse.data | ForEach-Object { $_.id }
        Write-Host "   Loaded models: $($loadedModels -join ', ')" -ForegroundColor Cyan

        # Auto-detect chat model if not specified
        if (-not $actualChatModel -and -not $NoSummary) {
            $actualChatModel = ($loadedModels | Where-Object { $_ -notmatch "embed" } | Select-Object -First 1)
            if ($actualChatModel) {
                Write-Host "   Auto-detected chat model: $actualChatModel" -ForegroundColor Cyan
            }
        }
    } catch {
        Write-Host "   Could not query models API" -ForegroundColor Yellow
        $loadedModels = @()
    }

    # ─────────────────────────────────────
    # Step 3: Index benchmark content
    # ─────────────────────────────────────
    if (-not $SkipIndex) {
        Write-Host "`n📚 Indexing benchmark content..." -ForegroundColor Yellow
        Write-Host "   Volumes: $Volumes" -ForegroundColor Cyan
        Write-Host "   Data dir: $expDataDir" -ForegroundColor Cyan

        # Create experiment data directory
        New-Item -ItemType Directory -Path $expDataDir -Force | Out-Null

        $indexArgs = @(
            "index",
            "-volumes", $Volumes,
            "-v=true"
        )
        if ($NoSummary) {
            $indexArgs += @("-no-summary", "-no-theme")
        }
        if ($EmbeddingModel) {
            # We set the embedding model via environment variable approach
            # gospel-vec reads from config, so we'll use the config mechanism
        }
        if ($ChatModel) {
            $indexArgs += @("-chat-model", $ChatModel)
        }

        $indexStart = Get-Date

        # Set environment to use isolated data dir
        $env:GOSPEL_VEC_DATA_DIR = $expDataDir
        if ($EmbeddingModel) {
            $env:GOSPEL_VEC_EMBEDDING_MODEL = $EmbeddingModel
        }

        & .\gospel-vec.exe @indexArgs
        if ($LASTEXITCODE -ne 0) { throw "Indexing failed" }

        $indexElapsed = (Get-Date) - $indexStart
        Write-Host "`n✅ Indexing complete! ($($indexElapsed.ToString('mm\:ss')))" -ForegroundColor Green

        # Show stats
        .\gospel-vec.exe stats
    } else {
        Write-Host "`n⏭️  Skipping index (reusing existing data)" -ForegroundColor Yellow
        if (-not (Test-Path $expDataDir)) {
            throw "No existing experiment data at $expDataDir. Run without -SkipIndex first."
        }
    }

    # ─────────────────────────────────────
    # Step 4: Run benchmark queries
    # ─────────────────────────────────────
    Write-Host "`n🔍 Running benchmark queries..." -ForegroundColor Yellow

    # Load benchmark queries
    $benchmark = Get-Content $benchmarkFile -Raw | ConvertFrom-Json

    $queryResults = @()
    $totalQueries = $benchmark.queries.Count
    $queryIndex = 0

    foreach ($q in $benchmark.queries) {
        $queryIndex++
        Write-Host "   [$queryIndex/$totalQueries] $($q.id): $($q.query)" -ForegroundColor DarkGray

        # Run search
        $searchArgs = @(
            "search",
            "-layers", $SearchLayers,
            "-limit", $Limit,
            "-content=true",
            "-maxlen", 300,
            $q.query
        )

        $searchOutput = & .\gospel-vec.exe @searchArgs 2>&1 | Out-String

        # Parse results from output
        # Format: "1. [0.8234] Alma 32:21 (verse)"
        #         "   content text here..."
        $results = @()
        $lines = $searchOutput -split "`n"
        foreach ($line in $lines) {
            if ($line -match '^\d+\.\s+\[([\d.]+)\]\s+(.+?)\s+\((\w+)\)\s*$') {
                $results += @{
                    score     = [float]$Matches[1]
                    reference = $Matches[2].Trim()
                    layer     = $Matches[3]
                }
            }
        }

        # Check for expected matches
        $hits = 0
        $expectedCount = $q.expected_relevant.Count
        foreach ($expected in $q.expected_relevant) {
            # Extract the book-chapter from expected, e.g., "alma-32" from "alma-32 (v21: ...)"
            $expectedBook = ($expected -split '\s+\(')[0].Trim().ToLower()
            $found = $false
            foreach ($r in $results) {
                # Normalize reference: "Alma 32:21" → "alma-32"
                # Remove verse part (after colon), lowercase, replace spaces with hyphens
                $refBook = ($r.reference -replace ':.*$', '').ToLower() -replace '\s+', '-'
                if ($refBook -eq $expectedBook) {
                    $found = $true
                    break
                }
            }
            if ($found) { $hits++ }
        }

        $recall = if ($expectedCount -gt 0) { [math]::Round($hits / $expectedCount, 2) } else { 0 }

        $queryResults += @{
            id               = $q.id
            query            = $q.query
            category         = $q.category
            expected_count   = $expectedCount
            hits             = $hits
            recall           = $recall
            top_results      = $results | Select-Object -First 5
            raw_result_count = $results.Count
        }

        $recallPct = [math]::Round($recall * 100)
        $color = if ($recallPct -ge 67) { "Green" } elseif ($recallPct -ge 34) { "Yellow" } else { "Red" }
        Write-Host "     → Recall: $hits/$expectedCount ($recallPct%)" -ForegroundColor $color
    }

    # ─────────────────────────────────────
    # Step 5: Calculate aggregate metrics
    # ─────────────────────────────────────
    $avgRecall = [math]::Round(($queryResults | ForEach-Object { $_.recall } | Measure-Object -Average).Average, 3)
    $totalHits = ($queryResults | ForEach-Object { $_.hits } | Measure-Object -Sum).Sum
    $totalExpected = ($queryResults | ForEach-Object { $_.expected_count } | Measure-Object -Sum).Sum
    $perfectRecall = ($queryResults | Where-Object { $_.recall -eq 1.0 }).Count
    $zeroRecall = ($queryResults | Where-Object { $_.recall -eq 0 }).Count

    # ─────────────────────────────────────
    # Step 6: Save results
    # ─────────────────────────────────────
    $experimentResult = @{
        name             = $Name
        timestamp        = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
        config           = @{
            embedding_model = $actualEmbeddingModel
            chat_model      = $actualChatModel
            volumes         = $Volumes
            search_layers   = $SearchLayers
            no_summary      = $NoSummary.IsPresent
            limit           = $Limit
        }
        metrics          = @{
            avg_recall     = $avgRecall
            total_hits     = $totalHits
            total_expected = $totalExpected
            perfect_recall = $perfectRecall
            zero_recall    = $zeroRecall
            query_count    = $totalQueries
        }
        query_results    = $queryResults
    }

    $experimentResult | ConvertTo-Json -Depth 10 | Set-Content $resultFile -Encoding UTF8
    Write-Host "`n💾 Results saved to: $resultFile" -ForegroundColor Green

    # ─────────────────────────────────────
    # Step 7: Append to experiment log
    # ─────────────────────────────────────
    $logEntry = @"

## $Name

**Date:** $(Get-Date -Format "yyyy-MM-dd HH:mm")
**Embedding Model:** $actualEmbeddingModel
**Chat Model:** $(if ($actualChatModel) { $actualChatModel } else { 'none (no summary)' })
**Volumes:** $Volumes | **Layers Searched:** $SearchLayers
**Summaries:** $(if ($NoSummary) { 'OFF' } else { 'ON' })

### Results

| Metric | Value |
|--------|-------|
| Average Recall | $($avgRecall * 100)% |
| Total Hits | $totalHits / $totalExpected |
| Perfect Recall (100%) | $perfectRecall / $totalQueries queries |
| Zero Recall (0%) | $zeroRecall / $totalQueries queries |

### Per-Query Breakdown

| Query | Category | Recall | Hits |
|-------|----------|--------|------|
$($queryResults | ForEach-Object {
    "| $($_.id) | $($_.category) | $([math]::Round($_.recall * 100))% | $($_.hits)/$($_.expected_count) |"
} | Out-String)
**Observations:**
_TODO: Add observations about this experiment_

---
"@

    Add-Content -Path $logFile -Value $logEntry -Encoding UTF8
    Write-Host "📝 Log appended to: $logFile" -ForegroundColor Green

    # ─────────────────────────────────────
    # Summary
    # ─────────────────────────────────────
    Write-Host "`n══════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  Experiment Results: $Name" -ForegroundColor Cyan
    Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  Average Recall:   $($avgRecall * 100)%" -ForegroundColor $(if ($avgRecall -ge 0.67) { "Green" } elseif ($avgRecall -ge 0.34) { "Yellow" } else { "Red" })
    Write-Host "  Total Hits:       $totalHits / $totalExpected" -ForegroundColor White
    Write-Host "  Perfect Queries:  $perfectRecall / $totalQueries" -ForegroundColor White
    Write-Host "  Zero Recall:      $zeroRecall / $totalQueries" -ForegroundColor $(if ($zeroRecall -eq 0) { "Green" } else { "Red" })
    Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan

} finally {
    # Clean up environment variables
    Remove-Item Env:\GOSPEL_VEC_DATA_DIR -ErrorAction SilentlyContinue
    Remove-Item Env:\GOSPEL_VEC_EMBEDDING_MODEL -ErrorAction SilentlyContinue
    Pop-Location
}
