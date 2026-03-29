<#
.SYNOPSIS
    Run a single LM Studio model test: prompt + content → model → recorded response.

.DESCRIPTION
    Sends a prompt with content to an LM Studio model via the OpenAI-compatible API.
    Records the response, timing, and token counts. Saves raw response as JSON and
    appends a summary line to results.tsv.

    The system message is always loaded from context.md (covenant + intent extract).
    The user message is constructed from the prompt template + content file.

.PARAMETER Prompt
    Name of the prompt file (without extension) in the prompts/ directory.

.PARAMETER Content
    Name of the content file (without extension) in the content/ directory.

.PARAMETER Model
    Short name of the model. Used for display and result logging.
    The actual model ID is read from the LM Studio /v1/models endpoint.

.PARAMETER MaxTokens
    Maximum tokens to generate. Default: 2048.

.PARAMETER Temperature
    Sampling temperature. Default: 0.7.

.PARAMETER Tag
    Optional tag for this run (e.g., "pass2", "tailored"). Appears in results.tsv.

.PARAMETER NoSave
    Don't save the raw response JSON (useful for quick tests).

.PARAMETER BaseURL
    LM Studio API base URL. Default: http://localhost:1234/v1

.EXAMPLE
    .\run-test.ps1 -Prompt summarize -Content kearon-receive-his-gift -Model nemotron-3-nano

.EXAMPLE
    .\run-test.ps1 -Prompt cross-reference -Content alma-32 -Model qwen3.5-35b -Tag pass2
#>

param(
    [Parameter(Mandatory)][string]$Prompt,
    [Parameter(Mandatory)][string]$Content,
    [string]$Model = "",
    [int]$MaxTokens = 16384,
    [double]$Temperature = 0.7,
    [string]$Tag = "",
    [string]$Context = "",
    [switch]$NoSave,
    [switch]$NoThink,
    [switch]$DisableThinking,
    [string]$BaseURL = "http://localhost:1234/v1"
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot

# --- Resolve files ---

$contextFile = Join-Path $scriptDir "context.md"
$promptFile = Join-Path $scriptDir "prompts" "$Prompt.md"
$contentFile = Join-Path $scriptDir "content" "$Content.md"
$resultsDir = Join-Path $scriptDir "results"
$resultsTsv = Join-Path $scriptDir "results.tsv"

if (-not (Test-Path $contextFile)) {
    Write-Error "Context file not found: $contextFile"
    return
}
if (-not (Test-Path $promptFile)) {
    Write-Error "Prompt file not found: $promptFile. Available prompts:"
    Get-ChildItem (Join-Path $scriptDir "prompts") -Filter "*.md" | ForEach-Object { Write-Host "  $($_.BaseName)" }
    return
}
if (-not (Test-Path $contentFile)) {
    Write-Error "Content file not found: $contentFile. Available content:"
    Get-ChildItem (Join-Path $scriptDir "content") -Filter "*.md" | ForEach-Object { Write-Host "  $($_.BaseName)" }
    return
}

# --- Load files ---

$systemMessage = Get-Content $contextFile -Raw -Encoding UTF8

if ($Context) {
    $contextDir = Join-Path $scriptDir $Context
    if (-not (Test-Path $contextDir)) {
        Write-Error "Context directory not found: $contextDir"
        return
    }
    $contextFiles = Get-ChildItem $contextDir -Filter "*.md" | Sort-Object Name
    foreach ($cf in $contextFiles) {
        $systemMessage += "`n`n" + (Get-Content $cf.FullName -Raw -Encoding UTF8)
    }
    Write-Host "Context: $($contextFiles.Count) files from $Context" -ForegroundColor Cyan
}

$promptTemplate = Get-Content $promptFile -Raw -Encoding UTF8
$contentText = Get-Content $contentFile -Raw -Encoding UTF8

# Build user message: prompt template with content inserted
$userMessage = $promptTemplate -replace '\{\{CONTENT\}\}', $contentText

# For thinking models (e.g., Qwen3.5), prepend /no_think to prevent
# the model from spending all tokens on internal reasoning
if ($NoThink) {
    $userMessage = "/no_think`n$userMessage"
}

# --- Detect model from LM Studio ---

if (-not $Model) {
    Write-Host "No model specified, detecting from LM Studio..."
}

try {
    $modelsResponse = Invoke-RestMethod -Uri "$BaseURL/models" -Method Get -TimeoutSec 5
} catch {
    Write-Error "Cannot reach LM Studio at $BaseURL. Is the server running?"
    return
}

$loadedModels = $modelsResponse.data | Where-Object { $_.id -notmatch 'embedding' }
if ($loadedModels.Count -eq 0) {
    Write-Error "No inference models loaded in LM Studio."
    return
}

# Pick the model: match by short name if provided, otherwise use first loaded
$modelId = ""
if ($Model) {
    $match = $loadedModels | Where-Object { $_.id -match [regex]::Escape($Model) }
    if ($match) {
        $modelId = ($match | Select-Object -First 1).id
    } else {
        Write-Warning "Model '$Model' not found in loaded models. Using first available."
        $modelId = $loadedModels[0].id
    }
} else {
    $modelId = $loadedModels[0].id
    $Model = ($modelId -split '/')[-1]
}

Write-Host "Model: $modelId" -ForegroundColor Cyan
Write-Host "Prompt: $Prompt" -ForegroundColor Cyan
Write-Host "Content: $Content ($($contentText.Length) chars)" -ForegroundColor Cyan
if ($Context) { Write-Host "Context: $Context" -ForegroundColor Cyan }

# --- Build request ---

$body = @{
    model = $modelId
    messages = @(
        @{ role = "system"; content = $systemMessage }
        @{ role = "user"; content = $userMessage }
    )
    max_tokens = $MaxTokens
    temperature = $Temperature
    stream = $true
    stream_options = @{ include_usage = $true }
    cache_prompt = $true
}

if ($DisableThinking) {
    $body.chat_template_kwargs = @{ enable_thinking = $false }
    Write-Host "API: chat_template_kwargs.enable_thinking = false" -ForegroundColor Magenta
}

$requestBody = $body | ConvertTo-Json -Depth 10

# --- Send streaming request and time it ---

Write-Host "`nSending request..." -ForegroundColor Yellow
$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

try {
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($requestBody)
    $httpRequest = [System.Net.HttpWebRequest]::Create("$BaseURL/chat/completions")
    $httpRequest.Method = "POST"
    $httpRequest.ContentType = "application/json; charset=utf-8"
    $httpRequest.Timeout = 600000
    $reqStream = $httpRequest.GetRequestStream()
    $reqStream.Write($bodyBytes, 0, $bodyBytes.Length)
    $reqStream.Close()

    $httpResponse = $httpRequest.GetResponse()
    $reader = New-Object System.IO.StreamReader($httpResponse.GetResponseStream())

    $ttftMs = $null
    $responseText = ""
    $tokensIn = 0
    $tokensOut = 0

    while (-not $reader.EndOfStream) {
        $line = $reader.ReadLine()
        if (-not $line -or -not $line.StartsWith("data: ")) { continue }
        $data = $line.Substring(6)
        if ($data -eq "[DONE]") { break }
        try {
            $chunk = $data | ConvertFrom-Json
            # Check for content delta
            $delta = $chunk.choices[0].delta.content
            if ($delta) {
                if (-not $ttftMs) { $ttftMs = $stopwatch.ElapsedMilliseconds }
                $responseText += $delta
            }
            # Check for usage in final chunk
            if ($chunk.usage) {
                $tokensIn = $chunk.usage.prompt_tokens
                $tokensOut = $chunk.usage.completion_tokens
            }
        } catch { }
    }
    $reader.Close()
    $httpResponse.Close()
} catch {
    $stopwatch.Stop()
    Write-Error "API call failed after $($stopwatch.ElapsedMilliseconds)ms: $_"
    return
}

$stopwatch.Stop()
$latencyMs = $stopwatch.ElapsedMilliseconds
if (-not $ttftMs) { $ttftMs = $latencyMs }
$genTimeMs = $latencyMs - $ttftMs
$totalTokens = $tokensIn + $tokensOut

# Calculate tok/s: overall (output/wall) and generation-only (output/gen_time)
$tokPerSec = if ($latencyMs -gt 0) { [math]::Round($tokensOut / ($latencyMs / 1000), 1) } else { 0 }
$genTokPerSec = if ($genTimeMs -gt 0) { [math]::Round($tokensOut / ($genTimeMs / 1000), 1) } else { 0 }

# --- Display results ---

Write-Host "`n--- Response ---" -ForegroundColor Green
Write-Host $responseText
Write-Host "`n--- Stats ---" -ForegroundColor Green
Write-Host "Tokens in:  $tokensIn"
Write-Host "Tokens out: $tokensOut"
Write-Host "Total:      $totalTokens"
Write-Host "TTFT:       $($ttftMs)ms"
Write-Host "Gen time:   $($genTimeMs)ms"
Write-Host "Total time: $($latencyMs)ms"
Write-Host "Gen tok/s:  $genTokPerSec" -ForegroundColor Cyan
Write-Host "Overall tok/s: $tokPerSec"

# --- Save raw response ---

if (-not $NoSave) {
    New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $safeModel = $Model -replace '[/\\:]', '-'
    $resultFile = Join-Path $resultsDir "$timestamp-$safeModel-$Prompt-$Content.json"

    $resultObj = @{
        timestamp = (Get-Date -Format "o")
        model = $modelId
        model_short = $Model
        prompt = $Prompt
        content = $Content
        tag = $Tag
        temperature = $Temperature
        max_tokens = $MaxTokens
        tokens_in = $tokensIn
        tokens_out = $tokensOut
        tokens_total = $totalTokens
        ttft_ms = $ttftMs
        gen_time_ms = $genTimeMs
        latency_ms = $latencyMs
        gen_tok_per_sec = $genTokPerSec
        tok_per_sec = $tokPerSec
        system_message = $systemMessage
        user_message = $userMessage
        response = $responseText
    }

    $resultObj | ConvertTo-Json -Depth 10 | Set-Content $resultFile -Encoding UTF8
    Write-Host "`nSaved: $resultFile" -ForegroundColor DarkGray
}

# --- Append to results.tsv ---

# Create header if file doesn't exist
if (-not (Test-Path $resultsTsv)) {
    "timestamp`tmodel`tprompt`tcontent`ttag`ttokens_in`ttokens_out`tgen_tok_per_sec`ttok_per_sec`tttft_ms`tgen_time_ms`tlatency_ms`tscore`tnotes" |
        Set-Content $resultsTsv -Encoding UTF8
}

$tsvLine = @(
    (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
    $Model
    $Prompt
    $Content
    $Tag
    $tokensIn
    $tokensOut
    $genTokPerSec
    $tokPerSec
    $ttftMs
    $genTimeMs
    $latencyMs
    ""  # score — filled in by Michael
    ""  # notes — filled in by Michael
) -join "`t"

Add-Content $resultsTsv $tsvLine -Encoding UTF8
Write-Host "Logged to results.tsv" -ForegroundColor DarkGray
