# Phase 2.1 importer — read every markdown file under study/ and call
# stewards.import_study() for each. Writes per-file SQL to a temp dir
# and runs via psql -f to avoid PowerShell heredoc / encoding pitfalls.

param(
    [int]$Limit = 0,
    [string]$StudyDir = "..\..\..\study"
)

$ErrorActionPreference = "Continue"
$container = "pg-ai-stewards-dev"
$resolvedDir = (Resolve-Path $StudyDir).Path
Write-Host "Importing from: $resolvedDir"

$tmpDir = Join-Path $env:TEMP "stewards-import"
if (Test-Path $tmpDir) { Remove-Item $tmpDir -Recurse -Force }
New-Item -ItemType Directory -Path $tmpDir | Out-Null

$files = Get-ChildItem -Path $resolvedDir -Filter *.md -File
if ($Limit -gt 0) { $files = $files | Select-Object -First $Limit }

$total = $files.Count
$ok = 0
$fail = 0
$failures = @()
$i = 0

foreach ($f in $files) {
    $i++
    $slug = [System.IO.Path]::GetFileNameWithoutExtension($f.Name)
    $relPath = "study/" + $f.Name
    # PS5 Get-Content -Raw defaults to system codepage; force UTF-8 so
    # em-dashes and other non-ASCII characters survive the round trip
    # (otherwise they end up as 'â€"' style mojibake in the database).
    $body = [System.IO.File]::ReadAllText($f.FullName, [System.Text.UTF8Encoding]::new($false))

    $titleMatch = [regex]::Match($body, '(?m)^#\s+(.+)$')
    $title = if ($titleMatch.Success) { $titleMatch.Groups[1].Value.Trim() } else { $slug }

    $frontmatter = @{}
    foreach ($line in ($body -split "`n" | Select-Object -First 20)) {
        if ($line -match '^\*([A-Za-z][A-Za-z ]+):\s*(.+?)\*\s*$') {
            $key = $matches[1].Trim().ToLower() -replace '\s+', '_'
            $frontmatter[$key] = $matches[2].Trim()
        }
    }
    $fmJson = ($frontmatter | ConvertTo-Json -Compress)
    if ([string]::IsNullOrEmpty($fmJson) -or $fmJson -eq '{}') { $fmJson = '{}' }

    $tag = "stewstudy" + ($i.ToString())
    $open = "`$" + $tag + "`$"
    $close = $open

    $sb = New-Object System.Text.StringBuilder
    [void]$sb.AppendLine("\set ON_ERROR_STOP 1")
    [void]$sb.AppendLine("SELECT stewards.import_study(")
    [void]$sb.Append("    "); [void]$sb.Append($open); [void]$sb.Append($slug);    [void]$sb.AppendLine("$close,")
    [void]$sb.Append("    "); [void]$sb.Append($open); [void]$sb.Append($relPath); [void]$sb.AppendLine("$close,")
    [void]$sb.Append("    "); [void]$sb.Append($open); [void]$sb.Append($title);   [void]$sb.AppendLine("$close,")
    [void]$sb.Append("    "); [void]$sb.Append($open); [void]$sb.Append($body);    [void]$sb.AppendLine("$close,")
    [void]$sb.Append("    "); [void]$sb.Append($open); [void]$sb.Append($fmJson);  [void]$sb.AppendLine("$close::jsonb")
    [void]$sb.AppendLine(") AS id;")

    $sqlPath = Join-Path $tmpDir "$slug.sql"
    [System.IO.File]::WriteAllText($sqlPath, $sb.ToString(), (New-Object System.Text.UTF8Encoding $false))

    docker cp $sqlPath ${container}:/tmp/import.sql 2>&1 | Out-Null
    $result = docker exec $container psql -U stewards -d stewards -f /tmp/import.sql 2>&1

    if ($LASTEXITCODE -eq 0) {
        $ok++
        if ($i % 10 -eq 0) {
            Write-Host "  [$i/$total] $slug -> OK"
        }
    } else {
        $fail++
        $failures += $slug
        Write-Host "  [$i/$total] $slug -> FAIL"
        Write-Host ($result | Out-String)
    }
}

Write-Host ""
Write-Host "=== Done. ok=$ok fail=$fail ==="
if ($failures.Count -gt 0) {
    Write-Host "Failed slugs:"
    $failures | ForEach-Object { Write-Host "  - $_" }
}

Write-Host ""
Write-Host "=== Row count + edge count ==="
"SELECT count(*) AS studies FROM stewards.studies; LOAD 'age'; SET search_path=ag_catalog,public; SELECT * FROM cypher('stewards_graph', `$`$ MATCH ()-[r:CITES]->() RETURN count(r) `$`$) AS (cites_edges agtype);" | docker exec -i $container psql -U stewards -d stewards 2>&1

Write-Host ""
Write-Host "=== Top 10 most-cited scripture chapters across all studies ==="
"LOAD 'age'; SET search_path = ag_catalog, public; SELECT * FROM cypher('stewards_graph', `$`$ MATCH (s:Study)-[:CITES]->(t:Scripture) RETURN t.uri AS uri, count(s) AS citing_studies ORDER BY citing_studies DESC LIMIT 10 `$`$) AS (uri agtype, citing_studies agtype);" | docker exec -i $container psql -U stewards -d stewards 2>&1
