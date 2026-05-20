"""
Export stewards.thummim_entries → frontend/src/data/thummim-entries.json

After the thummim-define pipeline (substrate-yt-transcripts §VI carry-forward
+ thummim-restoration-dictionary.md proposal) has populated the substrate
table, this script dumps the entries into the static JSON shape the
frontend Dictionary.vue view expects.

Replaces the hand-curated `thummim-seed.json` once real entries exist.
Frontend reads `thummim-entries.json` first; falls back to `thummim-seed.json`
if the export hasn't been run yet.

Usage:
  python3 projects/1828-illuminated/scripts/export_thummim.py

  # Then rebuild the frontend bundle to pick up the new data:
  cd projects/1828-illuminated && docker build -t 1828-illuminated .

The script connects via psycopg2 if available; otherwise falls back to
shelling out to `docker exec pg-ai-stewards-dev psql ...` which always
works from the host.
"""
from __future__ import annotations
import json
import subprocess
import sys
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
OUT = REPO / "projects/1828-illuminated/frontend/src/data/thummim-entries.json"

QUERY = """
SELECT json_build_object(
    'word',                 word,
    'levels',               levels,
    'webster_1828_compare', webster_1828_compare,
    'substrate_study',      substrate_study,
    'generated_at',         to_char(generated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
)::text
FROM stewards.thummim_entries
ORDER BY word;
"""


def fetch_via_docker() -> list[dict]:
    """Query through `docker exec` since we know the container is up."""
    cmd = [
        "docker", "exec", "pg-ai-stewards-dev",
        "psql", "-U", "stewards", "-d", "stewards",
        "-tA", "-c", QUERY,
    ]
    out = subprocess.run(cmd, capture_output=True, text=True, check=True).stdout
    entries = []
    for line in out.splitlines():
        line = line.strip()
        if not line:
            continue
        entries.append(json.loads(line))
    return entries


def main() -> int:
    try:
        entries = fetch_via_docker()
    except subprocess.CalledProcessError as e:
        print(f"ERR: docker exec failed: {e.stderr}", file=sys.stderr)
        return 1
    except FileNotFoundError:
        print("ERR: docker CLI not found in PATH", file=sys.stderr)
        return 1

    # Re-shape to match the seed format ({word: entry}) for drop-in compatibility
    by_word = {e["word"]: e for e in entries}

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps({
        "_provenance": "Generated from stewards.thummim_entries by export_thummim.py",
        "_count": len(entries),
        "entries": by_word,
    }, indent=2), encoding="utf-8")

    print(f"Wrote {OUT.relative_to(REPO)} ({len(entries)} entries)", file=sys.stderr)
    if not entries:
        print(
            "(No entries yet. Dispatch a thummim-define work_item to start. "
            "Example: see scripts/example-dispatch-thummim.sql.)",
            file=sys.stderr,
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
