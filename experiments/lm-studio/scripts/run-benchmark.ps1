<#
.SYNOPSIS
    Benchmark multiple models on TITSW scoring against ground truth talks.

.DESCRIPTION
    Sequentially loads each model, runs the calibrated TITSW prompt on ground truth
    talks, and collects results for comparison. Handles unloading between models.

    For models with multiple quantization variants (e.g., ministral Q4_K_M + Q8_0),
    the script temporarily renames the unwanted variant so lms load picks the right one.

.PARAMETER Models
    Comma-separated list of model identifiers to benchmark.
    Default: the 3 new candidates.

.PARAMETER Prompt
    Prompt file to use. Default: titsw-calibrated

.PARAMETER Contents
    Comma-separated list of content files. Default: 3 ground truth talks.

.PARAMETER Tag
    Tag for this benchmark run. Default: "bench"

.PARAMETER ContextLength
    Context length to load models with. Default: 32768

.EXAMPLE
    .\run-benchmark.ps1
    # Run all 3 candidate models against ground truth talks

.EXAMPLE
    .\run-benchmark.ps1 -Models "mistralai/ministral-3-14b-reasoning"
    # Run a single model
#>

param(
    [string]$Models = "",
    [string]$Prompt = "titsw-calibrated",
    [string]$Contents = "kearon-receive-his-gift,bednar-their-own-judges,holland-and-now-i-see",
    [string]$Tag = "bench",
    [int]$MaxTokens = 16384,
    [double]$Temperature = 0.2,
    [int]$ContextLength = 32768,
    [switch]$NoThink,
    [string]$BaseURL = "http://localhost:1234/v1"
)

$ErrorActionPreference = "Stop"
$scriptDir = $PSScriptRoot

# --- Model definitions ---
# Each model has: key (for lms load), short name (for results), and optional
# quant-select info for multi-variant models.

$modelDefs = @(
    @{
        Key = "mistralai/ministral-3-14b-reasoning"
        Short = "ministral-14b-Q8"
        Quant = "Q8_0"
        # This model has Q4_K_M + Q8_0 variants. We want Q8_0.
        # Temporarily rename Q4_K_M during load.
        RenameFile = "Ministral-3-14B-Reasoning-2512-Q4_K_M.gguf"
        RenameDir = "lmstudio-community\Ministral-3-14B-Reasoning-2512-GGUF"
    },
    @{
        Key = "mistralai/magistral-small-2509"
        Short = "magistral-small-Q6K"
        Quant = "Q6_K"
    },
    @{
        Key = "qwen3.5-9b-abliterated-claude-4.6-opus-reasoning-distilled-v2"
        Short = "qwen35-9b-distill-Q8"
        Quant = "Q8_0"
    }
)

# Filter to requested models if specified
if ($Models) {
    $requested = $Models -split ',' | ForEach-Object { $_.Trim() }
    $modelDefs = $modelDefs | Where-Object {
        $def = $_
        $requested | Where-Object { $def.Key -match [regex]::Escape($_) -or $def.Short -match [regex]::Escape($_) }
    }
}

$contentList = $Contents -split ',' | ForEach-Object { $_.Trim() }
$totalTests = $modelDefs.Count * $contentList.Count

# --- LM Studio model directory ---
$lmModelsDir = "$env:USERPROFILE\.lmstudio\models"
if (-not (Test-Path $lmModelsDir)) {
    # Try reading from home pointer
    $pointer = "$env:USERPROFILE\.lmstudio-home-pointer"
    if (Test-Path $pointer) {
        $lmHome = (Get-Content $pointer -Raw).Trim()
        $lmModelsDir = Join-Path $lmHome "models"
    }
}

Write-Host "========================================" -ForegroundColor Magenta
Write-Host "  TITSW Model Benchmark" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "Models:    $($modelDefs.Count)" -ForegroundColor Cyan
Write-Host "Prompt:    $Prompt" -ForegroundColor Cyan
Write-Host "Content:   $($contentList -join ', ')" -ForegroundColor Cyan
Write-Host "Tag:       $Tag" -ForegroundColor Cyan
Write-Host "Context:   $ContextLength" -ForegroundColor Cyan
Write-Host "Temp:      $Temperature" -ForegroundColor Cyan
Write-Host "MaxTokens: $MaxTokens" -ForegroundColor Cyan
Write-Host "Total:     $totalTests tests" -ForegroundColor Cyan
Write-Host ""

# --- Ground truth for comparison ---
$groundTruth = @{
    "kearon-receive-his-gift" = @{ teach=8; help=8; love=4; spirit=3; doctrine=7; invite=8 }
    "bednar-their-own-judges" = @{ teach=5; help=5; love=2; spirit=3; doctrine=9; invite=6 }
    "holland-and-now-i-see"   = @{ teach=7; help=6; love=4; spirit=7; doctrine=6; invite=3 }
}

# --- Results collection ---
$allResults = @()

# --- Helper: extract scores from response text ---
function Extract-Scores {
    param([string]$Text)
    $scores = @{}
    $dims = @("TEACH_SCORE", "HELP_SCORE", "LOVE_SCORE", "SPIRIT_SCORE", "DOCTRINE_SCORE", "INVITE_SCORE")
    $shortNames = @("teach", "help", "love", "spirit", "doctrine", "invite")
    
    for ($i = 0; $i -lt $dims.Count; $i++) {
        if ($Text -match "$($dims[$i]):[*\s\[]*?(\d)") {
            $scores[$shortNames[$i]] = [int]$Matches[1]
        }
    }
    return $scores
}

# --- Helper: calculate MAE against ground truth ---
function Calculate-MAE {
    param([hashtable]$Scores, [hashtable]$Truth)
    $diffs = @()
    foreach ($dim in @("teach", "help", "love", "spirit", "doctrine", "invite")) {
        if ($Scores.ContainsKey($dim) -and $Truth.ContainsKey($dim)) {
            $diffs += [Math]::Abs($Scores[$dim] - $Truth[$dim])
        }
    }
    if ($diffs.Count -eq 0) { return -1 }
    return [Math]::Round(($diffs | Measure-Object -Sum).Sum / $diffs.Count, 2)
}

$current = 0

foreach ($modelDef in $modelDefs) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host "  Model: $($modelDef.Short) ($($modelDef.Quant))" -ForegroundColor Yellow
    Write-Host "========================================" -ForegroundColor Yellow

    # --- Unload any currently loaded model ---
    Write-Host "Unloading any loaded models..." -ForegroundColor DarkGray
    lms unload --all -y 2>&1 | Out-Null
    Start-Sleep -Seconds 3

    # --- Handle multi-variant quant selection ---
    $renamedFile = $null
    if ($modelDef.RenameFile -and $modelDef.RenameDir) {
        $origPath = Join-Path $lmModelsDir $modelDef.RenameDir $modelDef.RenameFile
        $hidePath = "$origPath.bak"
        if (Test-Path $origPath) {
            Write-Host "Temporarily hiding $($modelDef.RenameFile) to force $($modelDef.Quant) selection..." -ForegroundColor DarkYellow
            Rename-Item $origPath $hidePath
            $renamedFile = @{ Original = $origPath; Hidden = $hidePath }
            Start-Sleep -Seconds 1
        } else {
            Write-Host "Variant file not found at $origPath — may already be hidden or removed" -ForegroundColor DarkGray
        }
    }

    # --- Load the model ---
    Write-Host "Loading $($modelDef.Key) ($($modelDef.Quant))..." -ForegroundColor Cyan
    $loadOutput = lms load $modelDef.Key -c $ContextLength --gpu max -y 2>&1 | Out-String
    
    if ($loadOutput -match "loaded successfully|already loaded") {
        Write-Host "Model loaded successfully" -ForegroundColor Green
    } else {
        Write-Warning "Model load output: $loadOutput"
    }

    # Restore renamed file immediately after load
    if ($renamedFile) {
        Write-Host "Restoring $($modelDef.RenameFile)..." -ForegroundColor DarkGray
        if (Test-Path $renamedFile.Hidden) {
            Rename-Item $renamedFile.Hidden $renamedFile.Original
        }
    }

    # Brief pause for model to settle
    Start-Sleep -Seconds 2

    # Verify the model is loaded and check quantization
    $psOutput = lms ps 2>&1 | Out-String
    Write-Host $psOutput -ForegroundColor DarkGray

    # Also check via API
    try {
        $apiModels = Invoke-RestMethod -Uri "http://localhost:1234/api/v0/models" -TimeoutSec 5
        $loaded = $apiModels.data | Where-Object { $_.state -eq "loaded" }
        if ($loaded) {
            Write-Host "API confirms: $($loaded.id) @ $($loaded.quantization)" -ForegroundColor Green
        }
    } catch {
        Write-Host "Could not verify via API" -ForegroundColor DarkYellow
    }

    # --- Run tests for this model ---
    foreach ($content in $contentList) {
        $current++
        Write-Host ""
        Write-Host "[$current/$totalTests] $($modelDef.Short) × $content" -ForegroundColor Yellow

        $testArgs = @{
            Prompt = $Prompt
            Content = $content
            Model = $modelDef.Key
            MaxTokens = $MaxTokens
            Temperature = $Temperature
            Tag = "$Tag-$($modelDef.Short)"
            BaseURL = $BaseURL
        }
        if ($NoThink) { $testArgs['NoThink'] = $true }

        try {
            & (Join-Path $scriptDir "run-test.ps1") @testArgs

            # Parse the most recent result file for scores
            $resultsDir = Join-Path $scriptDir "results"
            $latest = Get-ChildItem $resultsDir -Filter "*.json" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
            if ($latest) {
                $result = Get-Content $latest.FullName -Raw | ConvertFrom-Json
                $scores = Extract-Scores $result.response
                
                $gt = $groundTruth[$content]
                $mae = if ($gt) { Calculate-MAE $scores $gt } else { -1 }

                $allResults += @{
                    Model = $modelDef.Short
                    Quant = $modelDef.Quant
                    Content = $content
                    Scores = $scores
                    MAE = $mae
                    TotalTime = $result.latency_ms
                    TokPerSec = $result.gen_tok_per_sec
                    TokensOut = $result.tokens_out
                }

                # Print inline comparison
                if ($gt -and $scores.Count -eq 6) {
                    $scoreStr = @("teach", "help", "love", "spirit", "doctrine", "invite") | ForEach-Object {
                        $s = $scores[$_]; $g = $gt[$_]; $diff = $s - $g
                        $color = if ($diff -eq 0) { "" } elseif ([Math]::Abs($diff) -le 1) { "" } else { "**" }
                        "${color}$_=$s($(if($diff -ge 0){"+$diff"}else{"$diff"}))${color}"
                    }
                    Write-Host "  Scores: $($scoreStr -join '  ')" -ForegroundColor $(if ($mae -le 1.5) { "Green" } elseif ($mae -le 2.0) { "Yellow" } else { "Red" })
                    Write-Host "  MAE: $mae  |  $([Math]::Round($result.total_time_ms/1000, 1))s  |  $($result.gen_tok_per_sec) tok/s" -ForegroundColor Cyan
                }
            }
        } catch {
            Write-Warning "FAILED: $($modelDef.Short) × $content — $_"
            $allResults += @{
                Model = $modelDef.Short
                Quant = $modelDef.Quant
                Content = $content
                Error = $_.ToString()
            }
        }
    }
}

# --- Summary ---
Write-Host ""
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "  BENCHMARK SUMMARY" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
Write-Host ""

Write-Host "Collected $($allResults.Count) results" -ForegroundColor DarkGray

# Group by model
$models = $allResults | Group-Object { $_.Model }

Write-Host ("{0,-30} {1,-10} {2,-8} {3,-8} {4,-8} {5,-8} {6,-8} {7,-8} {8,-8} {9,-10} {10,-8}" -f "Model × Content", "Quant", "teach", "help", "love", "spirit", "doct", "invite", "MAE", "Time(s)", "tok/s") -ForegroundColor White
Write-Host ("-" * 130) -ForegroundColor DarkGray

foreach ($group in $models) {
    $modelMae = @()
    $modelTokSec = @()
    
    foreach ($r in $group.Group) {
        if ($r.Error) {
            Write-Host ("{0,-30} {1,-10} ERROR: {2}" -f "$($r.Model) × $($r.Content.Substring(0, [Math]::Min(15, $r.Content.Length)))", $r.Quant, $r.Error.Substring(0, [Math]::Min(60, $r.Error.Length))) -ForegroundColor Red
            continue
        }
        
        $s = $r.Scores
        $timeS = [Math]::Round($r.TotalTime / 1000, 1)
        $line = "{0,-30} {1,-10} {2,-8} {3,-8} {4,-8} {5,-8} {6,-8} {7,-8} {8,-8} {9,-10} {10,-8}" -f `
            "$($r.Model.Substring(0, [Math]::Min(18, $r.Model.Length))) × $($r.Content.Substring(0, [Math]::Min(8, $r.Content.Length)))", `
            $r.Quant, `
            $(if($s.teach){"$($s.teach)"}else{"-"}), `
            $(if($s.help){"$($s.help)"}else{"-"}), `
            $(if($s.love){"$($s.love)"}else{"-"}), `
            $(if($s.spirit){"$($s.spirit)"}else{"-"}), `
            $(if($s.doctrine){"$($s.doctrine)"}else{"-"}), `
            $(if($s.invite){"$($s.invite)"}else{"-"}), `
            $r.MAE, $timeS, $r.TokPerSec
        
        $color = if ($r.MAE -le 1.5) { "Green" } elseif ($r.MAE -le 2.0) { "Yellow" } else { "Red" }
        Write-Host $line -ForegroundColor $color
        
        if ($r.MAE -ge 0) { $modelMae += $r.MAE }
        if ($r.TokPerSec) { $modelTokSec += $r.TokPerSec }
    }
    
    if ($modelMae.Count -gt 0) {
        $avgMae = [Math]::Round(($modelMae | Measure-Object -Average).Average, 2)
        $avgTok = [Math]::Round(($modelTokSec | Measure-Object -Average).Average, 1)
        Write-Host ("{0,-30} {1,-10} {2,-8} {3,-8} {4,-8} {5,-8} {6,-8} {7,-8} {8,-8} {9,-10} {10,-8}" -f `
            "  >> $($group.Name) AVG", "", "", "", "", "", "", "", $avgMae, "", $avgTok) -ForegroundColor Magenta
    }
    Write-Host ""
}

# Overall winner
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "  GROUND TRUTH REFERENCE" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
foreach ($talk in @("kearon-receive-his-gift", "bednar-their-own-judges", "holland-and-now-i-see")) {
    $gt = $groundTruth[$talk]
    Write-Host ("{0,-30} {1,-10} {2,-8} {3,-8} {4,-8} {5,-8} {6,-8} {7,-8}" -f $talk.Substring(0, [Math]::Min(28, $talk.Length)), "GT", $gt.teach, $gt.help, $gt.love, $gt.spirit, $gt.doctrine, $gt.invite) -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Previous benchmark (nemotron-3-nano T4): MAE = 1.83" -ForegroundColor DarkGray
Write-Host "Results saved to: $(Join-Path $scriptDir 'results')" -ForegroundColor DarkGray
