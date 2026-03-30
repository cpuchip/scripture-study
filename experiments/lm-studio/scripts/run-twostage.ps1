<#
.SYNOPSIS
    Two-stage scoring: typological analysis first, then scoring with that analysis as context.

.DESCRIPTION
    Stage 1: Runs the typology prompt against the content to identify symbolic/thematic
    connections mapped to each scoring dimension. No scores produced.

    Stage 2: Creates a combined content file (original text + Stage 1 analysis) and runs
    the scoring prompt against it. The scorer has the typological map available when
    evaluating each dimension.

    Both stages use the same model, temperature, and other settings. Results from both
    stages are saved. The final scored result uses tag "{Tag}-s2" to distinguish it.

.PARAMETER Content
    Name of the content file (without extension) in the content/ directory.

.PARAMETER Model
    Short name of the model.

.PARAMETER MaxTokens
    Maximum tokens to generate per stage. Default: 32768.

.PARAMETER Temperature
    Sampling temperature. Default: 0.2.

.PARAMETER Tag
    Tag prefix for this run. Stage 1 gets "{Tag}-s1", Stage 2 gets "{Tag}-s2".

.PARAMETER Stage1Prompt
    Stage 1 prompt name. Default: titsw-stage1-typology.

.PARAMETER Stage2Prompt
    Stage 2 prompt name. Default: titsw-stage2-score.

.EXAMPLE
    .\run-twostage.ps1 -Content alma-32 -Model "ministral-3-14b-reasoning" -Tag "2stage"

.EXAMPLE
    .\run-twostage.ps1 -Content alma-32-with-refs -Model "ministral-3-14b-reasoning" -Tag "2stage-ctx" -Temperature 0.2
#>

param(
    [Parameter(Mandatory)][string]$Content,
    [string]$Model = "",
    [int]$MaxTokens = 32768,
    [double]$Temperature = 0.2,
    [string]$Tag = "2stage",
    [string]$Stage1Prompt = "titsw-stage1-typology",
    [string]$Stage2Prompt = "titsw-stage2-score",
    [int]$ContextLength = 0,
    [string]$BaseURL = "http://localhost:1234/v1"
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot

Write-Host "=== TWO-STAGE SCORING ===" -ForegroundColor Magenta
Write-Host "Content: $Content" -ForegroundColor Cyan
Write-Host "Model:   $Model" -ForegroundColor Cyan
Write-Host "Temp:    $Temperature" -ForegroundColor Cyan
Write-Host ""

# --- Stage 1: Typological Analysis ---

Write-Host "--- STAGE 1: Typological Analysis ---" -ForegroundColor Yellow

$s1Args = @{
    Prompt      = $Stage1Prompt
    Content     = $Content
    Model       = $Model
    MaxTokens   = $MaxTokens
    Temperature = $Temperature
    Tag         = "$Tag-s1"
    BaseURL     = $BaseURL
}
if ($ContextLength -gt 0) { $s1Args.ContextLength = $ContextLength }

& "$scriptDir\run-test.ps1" @s1Args

# --- Find the Stage 1 result file ---

$resultsDir = Join-Path $scriptDir "results"
$s1Files = Get-ChildItem $resultsDir -Filter "*-$Stage1Prompt-$Content.json" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1

if (-not $s1Files) {
    Write-Error "Stage 1 result not found. Cannot proceed to Stage 2."
    return
}

Write-Host "`nStage 1 result: $($s1Files.Name)" -ForegroundColor DarkGray

# Extract the response text from Stage 1
$s1Result = Get-Content $s1Files.FullName -Raw -Encoding UTF8 | ConvertFrom-Json
$typologyText = $s1Result.response

if (-not $typologyText -or $typologyText.Length -lt 50) {
    Write-Error "Stage 1 produced insufficient output ($($typologyText.Length) chars). Aborting."
    return
}

Write-Host "Stage 1 output: $($typologyText.Length) chars" -ForegroundColor Green

# --- Build combined content for Stage 2 ---

$contentFile = Join-Path $scriptDir "content" "$Content.md"
$originalContent = Get-Content $contentFile -Raw -Encoding UTF8

$combinedContent = @"
$originalContent

---

## TYPOLOGICAL ANALYSIS

The following analysis identifies symbolic, typological, and thematic connections in the text above. Use these connections when scoring each dimension — a symbol that maps to a dimension counts as evidence for that dimension.

$typologyText
"@

# Write temp content file
$tempContentName = "_twostage-temp-$Content"
$tempContentFile = Join-Path $scriptDir "content" "$tempContentName.md"
$combinedContent | Set-Content $tempContentFile -Encoding UTF8

Write-Host "`nCombined content: $($combinedContent.Length) chars" -ForegroundColor Cyan

# --- Stage 2: Scoring with Typological Context ---

Write-Host "`n--- STAGE 2: Scoring with Typological Context ---" -ForegroundColor Yellow

$s2Args = @{
    Prompt      = $Stage2Prompt
    Content     = $tempContentName
    Model       = $Model
    MaxTokens   = $MaxTokens
    Temperature = $Temperature
    Tag         = "$Tag-s2"
    BaseURL     = $BaseURL
}
if ($ContextLength -gt 0) { $s2Args.ContextLength = $ContextLength }

try {
    & "$scriptDir\run-test.ps1" @s2Args
} finally {
    # Clean up temp file
    if (Test-Path $tempContentFile) {
        Remove-Item $tempContentFile -Force
        Write-Host "`nCleaned up temp content file" -ForegroundColor DarkGray
    }
}

# --- Summary ---

Write-Host "`n=== TWO-STAGE COMPLETE ===" -ForegroundColor Magenta
Write-Host "Stage 1 (typology): tag=$Tag-s1" -ForegroundColor Green
Write-Host "Stage 2 (scoring):  tag=$Tag-s2" -ForegroundColor Green
