# Phase 1.6.1 mode-4 verification: bgworker crashes mid-dispatch.
$ErrorActionPreference = "Continue"
$container = "pg-ai-stewards-dev"

Write-Host "=== Setup: orphaned in_progress tool_dispatch row ==="
Get-Content verify-1-6-1-reaper-setup.sql | docker exec -i $container psql -U stewards -d stewards 2>&1

Write-Host ""
Write-Host "=== Restarting container (bgworker reaper runs at startup) ==="
docker compose restart pg 2>&1 | Select-Object -Last 1
Start-Sleep -Seconds 6

Write-Host ""
Write-Host "=== Post-restart inspection (BEFORE continuation runs) ==="
Get-Content verify-1-6-1-reaper-check.sql | docker exec -i $container psql -U stewards -d stewards 2>&1

Write-Host ""
Write-Host "=== Bgworker reaper logs ==="
docker compose logs pg 2>&1 | Select-String -Pattern 'reaper|synthesize_tool_failure' | Select-Object -Last 10

Write-Host ""
Write-Host "=== Wait for continuation to complete ==="
$attempts = 0
$done = $false
while (-not $done -and $attempts -lt 30) {
    Start-Sleep -Seconds 4
    $attempts++
    $state = "SELECT 'pending=' || (SELECT count(*) FROM stewards.work_queue WHERE payload->>'session_id'='mode-4-reaper' AND status NOT IN ('done','error')) || ' last=' || coalesce((SELECT finish_reason FROM stewards.messages WHERE session_id='mode-4-reaper' AND role='assistant' ORDER BY id DESC LIMIT 1), 'none');" | docker exec -i $container psql -U stewards -d stewards -tA 2>&1
    Write-Host "  [$attempts] $state"
    if ($state -match 'pending=0') { $done = $true }
}

Write-Host ""
Write-Host "=== Final state (AFTER recovery) ==="
Get-Content verify-1-6-1-reaper-check.sql | docker exec -i $container psql -U stewards -d stewards 2>&1

Write-Host ""
Write-Host "=== Continuation chat detail ==="
"SELECT id, status, left(error, 300) AS err, jsonb_pretty(payload->'body'->'messages') AS msgs FROM stewards.work_queue WHERE kind='chat' AND payload->>'session_id' = 'mode-4-reaper' ORDER BY id DESC LIMIT 1;" | docker exec -i $container psql -U stewards -d stewards 2>&1

Write-Host ""
Write-Host "=== Skipping cleanup so state is inspectable ==="

Write-Host "Done."
