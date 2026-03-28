<#
.SYNOPSIS
    Run all prompts against a model (or all models), logging results.

.DESCRIPTION
    Iterates through all prompt files in prompts/ and all content files in content/,
    running each combination through run-test.ps1. This is the "run the suite" command.

    After the suite completes, review results.tsv and fill in scores manually.

.PARAMETER Model
    Short name of the model to test. Required (load one model at a time in LM Studio).

.PARAMETER Prompts
    Comma-separated list of prompt names to run. Default: all prompts in prompts/.

.PARAMETER Contents
    Comma-separated list of content names to run. Default: all content in content/.

.PARAMETER Tag
    Tag for this suite run (e.g., "pass1", "tailored"). Default: "pass1".

.PARAMETER MaxTokens
    Maximum tokens to generate per test. Default: 2048.

.PARAMETER Temperature
    Sampling temperature. Default: 0.7.

.PARAMETER BaseURL
    LM Studio API base URL. Default: http://localhost:1234/v1

.EXAMPLE
    .\run-suite.ps1 -Model nemotron-3-nano
    # Run all prompts × all content for nemotron

.EXAMPLE
    .\run-suite.ps1 -Model qwen3.5-35b -Prompts "summarize,cross-reference" -Tag pass2
    # Run specific prompts only
#>

param(
    [Parameter(Mandatory)][string]$Model,
    [string]$Prompts = "",
    [string]$Contents = "",
    [string]$Tag = "pass1",
    [int]$MaxTokens = 2048,
    [double]$Temperature = 0.7,
    [string]$BaseURL = "http://localhost:1234/v1"
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot

# --- Resolve prompt and content lists ---

$promptDir = Join-Path $scriptDir "prompts"
$contentDir = Join-Path $scriptDir "content"

if ($Prompts) {
    $promptList = $Prompts -split ',' | ForEach-Object { $_.Trim() }
} else {
    $promptList = Get-ChildItem $promptDir -Filter "*.md" | ForEach-Object { $_.BaseName }
}

if ($Contents) {
    $contentList = $Contents -split ',' | ForEach-Object { $_.Trim() }
} else {
    $contentList = Get-ChildItem $contentDir -Filter "*.md" | ForEach-Object { $_.BaseName }
}

if ($promptList.Count -eq 0) {
    Write-Error "No prompts found in $promptDir"
    return
}
if ($contentList.Count -eq 0) {
    Write-Error "No content found in $contentDir"
    return
}

$totalTests = $promptList.Count * $contentList.Count
Write-Host "=== Suite: $Model ===" -ForegroundColor Magenta
Write-Host "Prompts:  $($promptList -join ', ')" -ForegroundColor Cyan
Write-Host "Content:  $($contentList -join ', ')" -ForegroundColor Cyan
Write-Host "Tag:      $Tag" -ForegroundColor Cyan
Write-Host "Tests:    $totalTests" -ForegroundColor Cyan
Write-Host ""

# --- Run each combination ---

$current = 0
$failed = 0

foreach ($prompt in $promptList) {
    foreach ($content in $contentList) {
        $current++
        Write-Host "[$current/$totalTests] $prompt × $content" -ForegroundColor Yellow

        try {
            & (Join-Path $scriptDir "run-test.ps1") `
                -Prompt $prompt `
                -Content $content `
                -Model $Model `
                -MaxTokens $MaxTokens `
                -Temperature $Temperature `
                -Tag $Tag `
                -BaseURL $BaseURL
        } catch {
            Write-Warning "FAILED: $prompt × $content — $_"
            $failed++
        }

        Write-Host ""
    }
}

# --- Summary ---

Write-Host "=== Suite Complete ===" -ForegroundColor Magenta
Write-Host "Passed: $($totalTests - $failed) / $totalTests"
if ($failed -gt 0) {
    Write-Host "Failed: $failed" -ForegroundColor Red
}
Write-Host "`nNext: Review results.tsv and fill in scores (0-5)."
Write-Host "Then load the next model in LM Studio and run again."
