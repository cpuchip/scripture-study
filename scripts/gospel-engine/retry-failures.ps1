# retry-failures.ps1
# Re-enriches 1 Nephi (suspect garbage from 32K-context run) and the 27
# talks that failed (NULL titsw), then re-embeds all enrichments.
#
# Usage: powershell -ExecutionPolicy Bypass -File retry-failures.ps1

$ErrorActionPreference = "Stop"
$engineDir = $PSScriptRoot
$exe = Join-Path $engineDir "gospel-engine.exe"

$env:GOSPEL_ENGINE_CHAT_MODEL = "mistralai/ministral-3-14b-reasoning"
$env:GOSPEL_ENGINE_CHAT_URL = "http://localhost:1234/v1"

Set-Location $engineDir

# ── Step 1: Re-enrich 1 Nephi (all 22 chapters, --force) ─────────────────
Write-Host "`n=== Step 1: Re-enriching 1 Nephi (22 chapters, --force) ===" -ForegroundColor Cyan
& $exe enrich-scriptures --book=1-ne --force --verbose
if ($LASTEXITCODE -ne 0) { Write-Warning "1 Nephi enrichment exited with code $LASTEXITCODE" }

# ── Step 2: Enrich 27 remaining talks (NULL titsw data) ──────────────────
# These have no enrichment data at all — 4 from LLM 500 errors, 23 from
# parsing failures ("could not find teach score"). No --force needed since
# their titsw columns are NULL.
Write-Host "`n=== Step 2: Enriching remaining unenriched talks ===" -ForegroundColor Cyan
& $exe enrich --concurrency=3 --verbose
if ($LASTEXITCODE -ne 0) { Write-Warning "Talk enrichment exited with code $LASTEXITCODE" }

# ── Step 3: Re-embed all enrichments ─────────────────────────────────────
Write-Host "`n=== Step 3: Re-embedding all enrichments ===" -ForegroundColor Cyan
& $exe embed-enrichments --source=scriptures
& $exe embed-enrichments --source=conference

# ── Step 4: Final stats ──────────────────────────────────────────────────
Write-Host "`n=== Step 4: Final stats ===" -ForegroundColor Cyan
& $exe stats

Write-Host "`n✅ Retry complete!" -ForegroundColor Green
