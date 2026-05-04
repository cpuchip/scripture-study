# stewards.ps1 — thin CLI wrapper for the pg_ai_stewards extension.
#
# Usage:
#   ./stewards.ps1 study show <slug> [-Container <name>] [-DB <name>]
#       [-User <name>] [-Sim <int>] [-Cites <int>] [-VerseChars <int>]
#
# This is intentionally minimal — every command is a thin wrapper over
# a SQL function in the extension. As more commands are added (e.g.
# `study list`, `study refresh`), each is one more case in the switch.
# When the surface grows past ~10 commands, port to a Go binary that
# uses libpq directly (no codepage friction).

param(
    [Parameter(Mandatory=$true, Position=0)] [string] $Noun,
    [Parameter(Mandatory=$true, Position=1)] [string] $Verb,
    [Parameter(Mandatory=$false, Position=2)] [string] $Slug,
    [string] $Container = "pg-ai-stewards-dev",
    [string] $Database = "stewards",
    [string] $DbUser = "stewards",
    [int]    $Sim = 5,
    [int]    $Cites = 20,
    [int]    $VerseChars = 140
)

# Force UTF-8 for the current console so em-dashes and other non-ASCII
# round-trip cleanly out of psql -t -A. Without this Windows defaults to
# cp1252 and renders ΓÇö instead of —.
$prevOutEnc = [Console]::OutputEncoding
$prevInEnc  = [Console]::InputEncoding
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::InputEncoding  = [System.Text.UTF8Encoding]::new($false)
$env:PGCLIENTENCODING = "UTF8"

try {
    switch ("$Noun $Verb") {
        "study show" {
            if (-not $Slug) {
                Write-Error "study show requires a <slug>"
                exit 1
            }
            $sqlSlug = $Slug.Replace("'", "''")
            $sql = "SELECT stewards.study_show('$sqlSlug', $Sim, $Cites, $VerseChars);"
            $sql | docker exec -i $Container psql -U $DbUser -d $Database -t -A
        }
        "study list" {
            $sql = @"
SELECT slug, title,
       coalesce(to_char(embedded_at, 'YYYY-MM-DD'), '(unembedded)') AS embedded
  FROM stewards.studies
 ORDER BY title ASC;
"@
            $sql | docker exec -i $Container psql -U $DbUser -d $Database
        }
        "study refresh" {
            if ($Slug) {
                $sqlSlug = $Slug.Replace("'", "''")
                $sql = @"
SELECT stewards.refresh_study_refs('$sqlSlug')       AS resolves_enqueued,
       stewards.refresh_study_similarity('$sqlSlug') AS similarity_edges;
"@
            } else {
                $sql = @"
SELECT stewards.refresh_all_study_refs()       AS resolves_enqueued,
       stewards.refresh_all_study_similarity() AS similarity_edges;
"@
            }
            $sql | docker exec -i $Container psql -U $DbUser -d $Database
        }
        default {
            Write-Error "unknown command: $Noun $Verb. Try: study show <slug> | study list | study refresh [<slug>]"
            exit 1
        }
    }
}
finally {
    [Console]::OutputEncoding = $prevOutEnc
    [Console]::InputEncoding  = $prevInEnc
}
