<#
.SYNOPSIS
    Test LM Studio parallel request throughput with GPU monitoring.

.DESCRIPTION
    Fires N concurrent requests to LM Studio and measures wall-clock time,
    individual request times, and GPU utilization via nvidia-smi.
    Uses the same prompt/content/context system as run-test.ps1.

.PARAMETER Concurrency
    Number of simultaneous requests. Default: 1. LM Studio supports up to
    the "Parallel" setting shown in the UI (typically 4).

.PARAMETER Prompt
    Name of the prompt file (without extension) in the prompts/ directory.

.PARAMETER Contents
    Comma-separated list of content file names (without extension).
    If fewer contents than concurrency, they cycle. If more, only the
    first N are used.

.PARAMETER Model
    Short name of the model for display/logging.

.PARAMETER MaxTokens
    Maximum tokens to generate per request. Default: 32768.

.PARAMETER Temperature
    Sampling temperature. Default: 0.2.

.PARAMETER Context
    Optional context directory name (loads all *.md files as additional system context).

.PARAMETER NoThink
    Prepend /no_think to user message.

.PARAMETER Tag
    Tag for result logging.

.PARAMETER BaseURL
    LM Studio API base URL. Default: http://localhost:1234/v1

.PARAMETER GpuPollMs
    How often to poll nvidia-smi in milliseconds. Default: 2000.

.EXAMPLE
    .\run-parallel-test.ps1 -Concurrency 2 -Prompt titsw-enriched-talk -Contents "kearon-receive-his-gift,bednar-their-own-judges" -Model nemotron-3-nano -NoThink -Context context-t4 -Tag parallel-2x
#>

param(
    [int]$Concurrency = 1,
    [Parameter(Mandatory)][string]$Prompt,
    [Parameter(Mandatory)][string]$Contents,
    [string]$Model = "",
    [int]$MaxTokens = 32768,
    [double]$Temperature = 0.2,
    [string]$Context = "",
    [switch]$NoThink,
    [string]$Tag = "",
    [string]$BaseURL = "http://localhost:1234/v1",
    [int]$GpuPollMs = 2000
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot

# --- Parse content list ---
$contentNames = $Contents -split ','
if ($contentNames.Count -lt $Concurrency) {
    # Cycle through contents if we have fewer than concurrency
    $expanded = @()
    for ($i = 0; $i -lt $Concurrency; $i++) {
        $expanded += $contentNames[$i % $contentNames.Count]
    }
    $contentNames = $expanded
} elseif ($contentNames.Count -gt $Concurrency) {
    $contentNames = $contentNames[0..($Concurrency - 1)]
}

# --- Resolve files ---
$contextFile = Join-Path $scriptDir "context.md"
$promptFile = Join-Path $scriptDir "prompts" "$Prompt.md"
$resultsDir = Join-Path $scriptDir "results"

if (-not (Test-Path $contextFile)) { Write-Error "Context file not found: $contextFile"; return }
if (-not (Test-Path $promptFile)) { Write-Error "Prompt file not found: $promptFile"; return }

foreach ($cn in $contentNames) {
    $cf = Join-Path $scriptDir "content" "$cn.md"
    if (-not (Test-Path $cf)) { Write-Error "Content file not found: $cf"; return }
}

# --- Load system message ---
$systemMessage = Get-Content $contextFile -Raw -Encoding UTF8

if ($Context) {
    $contextDir = Join-Path $scriptDir $Context
    if (-not (Test-Path $contextDir)) { Write-Error "Context directory not found: $contextDir"; return }
    $contextFiles = Get-ChildItem $contextDir -Filter "*.md" | Sort-Object Name
    foreach ($cf in $contextFiles) {
        $systemMessage += "`n`n" + (Get-Content $cf.FullName -Raw -Encoding UTF8)
    }
}

$promptTemplate = Get-Content $promptFile -Raw -Encoding UTF8

# --- Detect model ---
try {
    $modelsResponse = Invoke-RestMethod -Uri "$BaseURL/models" -Method Get -TimeoutSec 5
} catch {
    Write-Error "Cannot reach LM Studio at $BaseURL. Is the server running?"
    return
}

$loadedModels = $modelsResponse.data | Where-Object { $_.id -notmatch 'embedding' }
if ($loadedModels.Count -eq 0) { Write-Error "No inference models loaded."; return }

$modelId = ""
if ($Model) {
    $match = $loadedModels | Where-Object { $_.id -match [regex]::Escape($Model) }
    if ($match) { $modelId = ($match | Select-Object -First 1).id }
    else { $modelId = $loadedModels[0].id }
} else {
    $modelId = $loadedModels[0].id
    $Model = ($modelId -split '/')[-1]
}

# --- Build request bodies ---
$requestBodies = @()
for ($i = 0; $i -lt $Concurrency; $i++) {
    $cn = $contentNames[$i]
    $contentText = Get-Content (Join-Path $scriptDir "content" "$cn.md") -Raw -Encoding UTF8
    $userMessage = $promptTemplate -replace '\{\{CONTENT\}\}', $contentText
    if ($NoThink) { $userMessage = "/no_think`n$userMessage" }

    $body = @{
        model = $modelId
        messages = @(
            @{ role = "system"; content = $systemMessage }
            @{ role = "user"; content = $userMessage }
        )
        max_tokens = $MaxTokens
        temperature = $Temperature
        stream = $false
        cache_prompt = $true
    } | ConvertTo-Json -Depth 10

    $requestBodies += @{ Body = $body; Content = $cn; Index = $i }
}

# --- Print test info ---
Write-Host "`n=== PARALLEL THROUGHPUT TEST ===" -ForegroundColor Yellow
Write-Host "Model:       $modelId" -ForegroundColor Cyan
Write-Host "Concurrency: $Concurrency" -ForegroundColor Cyan
Write-Host "Prompt:      $Prompt" -ForegroundColor Cyan
Write-Host "Contents:    $($contentNames -join ', ')" -ForegroundColor Cyan
if ($Context) { Write-Host "Context:     $Context" -ForegroundColor Cyan }
Write-Host "MaxTokens:   $MaxTokens" -ForegroundColor Cyan
Write-Host "NoThink:     $NoThink" -ForegroundColor Cyan
Write-Host ""

# --- GPU baseline ---
Write-Host "--- GPU Baseline ---" -ForegroundColor Green
$gpuBaseline = & nvidia-smi --query-gpu=utilization.gpu,utilization.memory,power.draw,temperature.gpu,memory.used,memory.total --format=csv,noheader,nounits 2>$null
if ($gpuBaseline) {
    $gpus = $gpuBaseline | ForEach-Object {
        $parts = $_ -split ',\s*'
        @{ GpuUtil = $parts[0]; MemUtil = $parts[1]; Power = $parts[2]; Temp = $parts[3]; MemUsed = $parts[4]; MemTotal = $parts[5] }
    }
    for ($g = 0; $g -lt $gpus.Count; $g++) {
        Write-Host "  GPU $g`: $($gpus[$g].GpuUtil)% util, $($gpus[$g].Power)W, $($gpus[$g].Temp)C, $($gpus[$g].MemUsed)/$($gpus[$g].MemTotal) MiB" -ForegroundColor DarkGray
    }
} else {
    Write-Host "  nvidia-smi not available" -ForegroundColor DarkGray
}
Write-Host ""

# --- GPU monitoring (shared-state runspace) ---
$gpuSamples = [System.Collections.Concurrent.ConcurrentBag[PSObject]]::new()
$gpuStop = [System.Threading.ManualResetEventSlim]::new($false)

$monitorRunspace = [RunspaceFactory]::CreateRunspace()
$monitorRunspace.Open()
$monitorRunspace.SessionStateProxy.SetVariable('gpuSamples', $gpuSamples)
$monitorRunspace.SessionStateProxy.SetVariable('gpuStop', $gpuStop)
$monitorRunspace.SessionStateProxy.SetVariable('pollMs', $GpuPollMs)

$monitorPS = [PowerShell]::Create()
$monitorPS.Runspace = $monitorRunspace
[void]$monitorPS.AddScript({
    while (-not $gpuStop.IsSet) {
        $raw = & nvidia-smi --query-gpu=index,utilization.gpu,utilization.memory,power.draw,temperature.gpu --format=csv,noheader,nounits 2>$null
        if ($raw) {
            $ts = Get-Date -Format "HH:mm:ss.fff"
            foreach ($line in $raw) {
                $parts = $line -split ',\s*'
                $gpuSamples.Add([PSCustomObject]@{
                    Time = $ts
                    GPU = $parts[0]
                    GpuUtil = [int]$parts[1]
                    MemUtil = [int]$parts[2]
                    Power = [double]$parts[3]
                    Temp = [int]$parts[4]
                })
            }
        }
        Start-Sleep -Milliseconds $pollMs
    }
})
$monitorHandle = $monitorPS.BeginInvoke()

# --- Fire concurrent requests using runspaces ---
Write-Host "Firing $Concurrency request(s)..." -ForegroundColor Yellow

$wallStopwatch = [System.Diagnostics.Stopwatch]::StartNew()

$runspacePool = [RunspaceFactory]::CreateRunspacePool(1, $Concurrency)
$runspacePool.Open()

$handles = @()
foreach ($req in $requestBodies) {
    $ps = [PowerShell]::Create()
    $ps.RunspacePool = $runspacePool

    [void]$ps.AddScript({
        param($body, $url, $contentName, $idx)
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($body)
            $httpRequest = [System.Net.HttpWebRequest]::Create("$url/chat/completions")
            $httpRequest.Method = "POST"
            $httpRequest.ContentType = "application/json; charset=utf-8"
            $httpRequest.Timeout = 600000
            $reqStream = $httpRequest.GetRequestStream()
            $reqStream.Write($bodyBytes, 0, $bodyBytes.Length)
            $reqStream.Close()

            $httpResponse = $httpRequest.GetResponse()
            $reader = New-Object System.IO.StreamReader($httpResponse.GetResponseStream())
            $responseText = $reader.ReadToEnd()
            $reader.Close()
            $httpResponse.Close()

            $sw.Stop()
            $parsed = $responseText | ConvertFrom-Json
            $content = $parsed.choices[0].message.content
            $tokensIn = $parsed.usage.prompt_tokens
            $tokensOut = $parsed.usage.completion_tokens

            return @{
                Index = $idx
                Content = $contentName
                Response = $content
                TokensIn = $tokensIn
                TokensOut = $tokensOut
                ElapsedMs = $sw.ElapsedMilliseconds
                Error = $null
            }
        } catch {
            $sw.Stop()
            return @{
                Index = $idx
                Content = $contentName
                Response = $null
                TokensIn = 0
                TokensOut = 0
                ElapsedMs = $sw.ElapsedMilliseconds
                Error = $_.Exception.Message
            }
        }
    })
    [void]$ps.AddArgument($req.Body)
    [void]$ps.AddArgument($BaseURL)
    [void]$ps.AddArgument($req.Content)
    [void]$ps.AddArgument($req.Index)

    $handle = $ps.BeginInvoke()
    $handles += @{ PS = $ps; Handle = $handle }
}

# --- Wait for all to complete ---
$results = @()
foreach ($h in $handles) {
    $result = $h.PS.EndInvoke($h.Handle)
    $results += $result
    $h.PS.Dispose()
}

$wallStopwatch.Stop()
$wallMs = $wallStopwatch.ElapsedMilliseconds

$runspacePool.Close()
$runspacePool.Dispose()

# --- Stop GPU monitor ---
$gpuStop.Set()
$monitorPS.EndInvoke($monitorHandle)
$monitorPS.Dispose()
$monitorRunspace.Close()
$monitorRunspace.Dispose()
$gpuData = @($gpuSamples.ToArray())

# --- Display results ---
Write-Host "`n=== RESULTS ($Concurrency concurrent) ===" -ForegroundColor Green
Write-Host "Wall time: $($wallMs)ms ($([math]::Round($wallMs/1000, 1))s)" -ForegroundColor Cyan
Write-Host ""

$totalTokensOut = 0
$totalTokensIn = 0
foreach ($r in ($results | Sort-Object { $_.Index })) {
    $status = if ($r.Error) { "FAILED: $($r.Error)" } else { "OK" }
    $tokPerSec = if ($r.ElapsedMs -gt 0 -and $r.TokensOut -gt 0) { [math]::Round($r.TokensOut / ($r.ElapsedMs / 1000), 1) } else { 0 }
    Write-Host "  [$($r.Index)] $($r.Content): $status" -ForegroundColor $(if ($r.Error) { "Red" } else { "White" })
    if (-not $r.Error) {
        Write-Host "      Tokens: $($r.TokensIn) in / $($r.TokensOut) out | Time: $($r.ElapsedMs)ms | Tok/s: $tokPerSec"
        # Show first 100 chars of response
        $preview = ($r.Response -replace '\r?\n', ' ').Substring(0, [Math]::Min(120, $r.Response.Length))
        Write-Host "      Preview: $preview..." -ForegroundColor DarkGray
        $totalTokensOut += $r.TokensOut
        $totalTokensIn += $r.TokensIn
    }
}

# --- Throughput summary ---
$wallSec = $wallMs / 1000
$aggTokPerSec = if ($wallSec -gt 0) { [math]::Round($totalTokensOut / $wallSec, 1) } else { 0 }
$avgRequestMs = ($results | Where-Object { -not $_.Error } | Measure-Object -Property ElapsedMs -Average).Average
$maxRequestMs = ($results | Where-Object { -not $_.Error } | Measure-Object -Property ElapsedMs -Maximum).Maximum

Write-Host "`n--- Throughput ---" -ForegroundColor Green
Write-Host "Total tokens out:    $totalTokensOut"
Write-Host "Wall time:           $([math]::Round($wallSec, 1))s"
Write-Host "Aggregate tok/s:     $aggTokPerSec (total output / wall time)" -ForegroundColor Cyan
Write-Host "Avg request time:    $([math]::Round($avgRequestMs/1000, 1))s"
Write-Host "Max request time:    $([math]::Round($maxRequestMs/1000, 1))s"
Write-Host "Speedup vs serial:   ~$([math]::Round($avgRequestMs * $Concurrency / $maxRequestMs, 2))x (ideal: ${Concurrency}x)" -ForegroundColor Cyan

# --- GPU summary ---
if ($gpuData -and $gpuData.Count -gt 0) {
    Write-Host "`n--- GPU Usage ---" -ForegroundColor Green
    $gpuIds = $gpuData | Select-Object -ExpandProperty GPU -Unique | Sort-Object
    foreach ($gid in $gpuIds) {
        $samples = $gpuData | Where-Object { $_.GPU -eq $gid }
        $avgUtil = [math]::Round(($samples | Measure-Object -Property GpuUtil -Average).Average, 1)
        $maxUtil = ($samples | Measure-Object -Property GpuUtil -Maximum).Maximum
        $avgPower = [math]::Round(($samples | Measure-Object -Property Power -Average).Average, 1)
        $maxPower = [math]::Round(($samples | Measure-Object -Property Power -Maximum).Maximum, 1)
        $maxTemp = ($samples | Measure-Object -Property Temp -Maximum).Maximum
        Write-Host "  GPU $gid`: Util avg=$avgUtil% max=$maxUtil% | Power avg=${avgPower}W max=${maxPower}W | Max temp=${maxTemp}C" -ForegroundColor DarkGray
    }
    Write-Host "  Samples: $($gpuData.Count) over $([math]::Round($wallSec, 1))s" -ForegroundColor DarkGray
} else {
    Write-Host "`n--- GPU: No samples collected (run may have been too short for polling) ---" -ForegroundColor DarkGray
}

# --- Save result summary ---
New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$safeModel = $Model -replace '[/\\:]', '-'
$summaryFile = Join-Path $resultsDir "$timestamp-parallel-${Concurrency}x-$safeModel.json"

$summary = @{
    timestamp = (Get-Date -Format "o")
    model = $modelId
    concurrency = $Concurrency
    prompt = $Prompt
    contents = $contentNames
    context = $Context
    max_tokens = $MaxTokens
    temperature = $Temperature
    no_think = [bool]$NoThink
    tag = $Tag
    wall_ms = $wallMs
    total_tokens_in = $totalTokensIn
    total_tokens_out = $totalTokensOut
    aggregate_tok_per_sec = $aggTokPerSec
    requests = @($results | Sort-Object { $_.Index } | ForEach-Object {
        @{
            index = $_.Index
            content = $_.Content
            tokens_in = $_.TokensIn
            tokens_out = $_.TokensOut
            elapsed_ms = $_.ElapsedMs
            error = $_.Error
            response = $_.Response
        }
    })
    gpu_summary = if ($gpuData -and $gpuData.Count -gt 0) {
        $gpuIds = $gpuData | Select-Object -ExpandProperty GPU -Unique | Sort-Object
        @($gpuIds | ForEach-Object {
            $samples = $gpuData | Where-Object { $_.GPU -eq $_ }
            @{
                gpu = $_
                avg_util = [math]::Round(($samples | Measure-Object -Property GpuUtil -Average).Average, 1)
                max_util = ($samples | Measure-Object -Property GpuUtil -Maximum).Maximum
                avg_power = [math]::Round(($samples | Measure-Object -Property Power -Average).Average, 1)
                max_power = [math]::Round(($samples | Measure-Object -Property Power -Maximum).Maximum, 1)
                max_temp = ($samples | Measure-Object -Property Temp -Maximum).Maximum
                sample_count = $samples.Count
            }
        })
    } else { @() }
} | ConvertTo-Json -Depth 10

$summary | Set-Content $summaryFile -Encoding UTF8
Write-Host "`nSaved: $summaryFile" -ForegroundColor DarkGray
