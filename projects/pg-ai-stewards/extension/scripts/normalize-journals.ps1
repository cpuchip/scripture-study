# normalize-journals.ps1
#
# Bulk-edit .spec/journal/*.yaml to:
#   1. Prepend `watchman: skip` (idempotent — skips files that already have it)
#   2. Apply top-level key renames for confirmed aliases:
#        session:               → session_id:
#        summary:               → intent:
#        relational_dynamics:   → relationship:
#        relational:            → relationship:
#        open_questions:        → questions:
#
# Confirmed before running: no file uses BOTH variants of any pair, so all
# renames are pure aliases. Nested fields (e.g., discoveries[].label) are
# deliberately NOT touched — at least one carry_forward block uses `label:`
# with different semantics, and the importer already aliases the variants
# transparently.
#
# Reads/writes UTF-8 without BOM. Reports per-file changes.

param(
    [string]$Root = ".spec/journal",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

# Resolve the journal directory relative to the workspace root.
$workspaceRoot = (Resolve-Path .).Path
$journalDir = Join-Path $workspaceRoot $Root
if (-not (Test-Path $journalDir)) {
    throw "journal directory not found: $journalDir"
}

$utf8NoBom = New-Object System.Text.UTF8Encoding $false

$files = Get-ChildItem -Path $journalDir -Filter *.yaml | Sort-Object Name
Write-Output "Found $($files.Count) journal file(s)"
Write-Output ""

$summary = @{
    total      = 0
    tagged     = 0
    renamed    = 0
    untouched  = 0
    perRename  = @{
        "session"             = 0
        "summary"             = 0
        "relational_dynamics" = 0
        "relational"          = 0
        "open_questions"      = 0
    }
}

foreach ($f in $files) {
    $summary.total++
    $orig = [System.IO.File]::ReadAllText($f.FullName, [System.Text.Encoding]::UTF8)

    $changes = @()
    $content = $orig

    # 1. Prepend watchman: skip if not already present (anywhere top-level).
    if ($content -notmatch '(?m)^watchman:\s*') {
        $content = "watchman: skip`n" + $content
        $changes += "tag"
        $summary.tagged++
    }

    # 2. Top-level renames. (?m) = multiline so ^ matches line starts.
    #    Each pattern requires an immediate `:` to avoid hitting `session_id:`
    #    when renaming `session:`, etc.
    $pairs = @(
        @{old = '^session:';             new = 'session_id:';   key = 'session'             },
        @{old = '^summary:';             new = 'intent:';       key = 'summary'             },
        @{old = '^relational_dynamics:'; new = 'relationship:'; key = 'relational_dynamics' },
        @{old = '^relational:';          new = 'relationship:'; key = 'relational'          },
        @{old = '^open_questions:';      new = 'questions:';    key = 'open_questions'      }
    )
    foreach ($p in $pairs) {
        $before = $content
        $content = $content -replace "(?m)$($p.old)", $p.new
        if ($content -ne $before) {
            $changes += $p.key
            $summary.perRename[$p.key]++
        }
    }

    if ($changes.Count -eq 0) {
        $summary.untouched++
        continue
    }

    if ($changes -contains "tag" -or ($changes | Where-Object { $_ -ne "tag" }).Count -gt 0) {
        if ($changes -notcontains "tag") {
            $summary.renamed++
        }
    }

    $relPath = $f.FullName.Substring($workspaceRoot.Length + 1)
    Write-Output ("{0,-50} {1}" -f $relPath, ($changes -join ", "))

    if (-not $DryRun) {
        [System.IO.File]::WriteAllText($f.FullName, $content, $utf8NoBom)
    }
}

Write-Output ""
Write-Output "=== summary ==="
Write-Output "  total files:    $($summary.total)"
Write-Output "  tagged:         $($summary.tagged)  (added watchman: skip)"
Write-Output "  rename-only:    $($summary.renamed) (already tagged, only renames applied)"
Write-Output "  untouched:      $($summary.untouched)"
Write-Output ""
Write-Output "  rename counts:"
foreach ($k in $summary.perRename.Keys | Sort-Object) {
    Write-Output ("    {0,-22} {1}" -f $k, $summary.perRename[$k])
}
if ($DryRun) {
    Write-Output ""
    Write-Output "  (DRY RUN — no files written)"
}
