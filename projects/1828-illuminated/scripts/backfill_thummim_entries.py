"""
Backfill stewards.thummim_entries from research/dictionary/thummim-*.md

The thummim-define pipeline's on_maturity_verified hook renders the
materialized markdown file at research/dictionary/<slug>.md but does NOT
populate the stewards.thummim_entries table. That's the D-THM-7 carry-forward
from a prior session — never built.

This is a one-shot backfill: parse the JSON block from each thummim-*.md,
INSERT (ON CONFLICT word DO UPDATE) into thummim_entries, linking to the
source work_item by slug.

The markdown files come in two shapes — bare JSON at the top, OR a ```json
fence. Both are handled.

Usage:
  python3 projects/1828-illuminated/scripts/backfill_thummim_entries.py
"""
from __future__ import annotations
import json
import re
import subprocess
import sys
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
DICT_DIR = REPO / "research" / "dictionary"
PG_CONTAINER = "pg-ai-stewards-dev"

# Match either:
#   ```json
#   { ... }
#   ```
# OR a bare JSON object at the top of the file.
# Capture everything between ```json and ``` — handles both multi-line pretty
# JSON and single-line compact JSON. json.loads validates the body.
JSON_FENCE = re.compile(r"```json\s*\n(.*?)\n```", re.DOTALL)


def extract_entry(md_path: Path) -> dict | None:
    text = md_path.read_text(encoding="utf-8")

    # Try ```json fence first
    m = JSON_FENCE.search(text)
    if m:
        try:
            return json.loads(m.group(1))
        except json.JSONDecodeError as e:
            print(f"  [FAIL]{md_path.name}: invalid fenced json — {e}", file=sys.stderr)
            return None

    # Fall back to bare JSON at top — find first '{' and walk braces to match
    stripped = text.lstrip()
    if not stripped.startswith("{"):
        print(f"  [FAIL]{md_path.name}: no json fence and doesn't start with {{", file=sys.stderr)
        return None
    # Walk to find matching closing brace
    depth = 0
    end = -1
    in_str = False
    esc = False
    for i, ch in enumerate(stripped):
        if esc:
            esc = False
            continue
        if ch == "\\" and in_str:
            esc = True
            continue
        if ch == '"':
            in_str = not in_str
            continue
        if in_str:
            continue
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                end = i + 1
                break
    if end < 0:
        print(f"  [FAIL]{md_path.name}: unbalanced braces", file=sys.stderr)
        return None
    try:
        return json.loads(stripped[:end])
    except json.JSONDecodeError as e:
        print(f"  [FAIL]{md_path.name}: invalid bare json — {e}", file=sys.stderr)
        return None


def dollar_quote(s: str) -> str:
    """Wrap s in a Postgres dollar-quoted string, finding an unused tag."""
    # Pick a tag not present in the body
    for tag in ("", "j", "json", "thummim", "payload_tag"):
        marker = f"${tag}$"
        if marker not in s:
            return f"{marker}{s}{marker}"
    # Last resort — should never hit this
    raise ValueError("could not find a safe dollar-quote tag")


def upsert(entry: dict, slug: str) -> bool:
    word = entry.get("word")
    if not word:
        print(f"  [FAIL]{slug}: no 'word' field in entry", file=sys.stderr)
        return False

    levels = entry.get("levels") or {}
    webster_compare = entry.get("webster_1828_compare")
    substrate_study = entry.get("substrate_study")

    payload_json = json.dumps({
        "word": word,
        "levels": levels,
        "webster_1828_compare": webster_compare,
        "substrate_study": substrate_study,
        "slug": slug,
    })
    payload_lit = dollar_quote(payload_json)

    sql = f"""
        WITH p AS (
          SELECT {payload_lit}::jsonb AS j
        ), wi AS (
          SELECT w.id FROM stewards.work_items w, p
          WHERE w.slug = p.j->>'slug'
          LIMIT 1
        )
        INSERT INTO stewards.thummim_entries
          (word, work_item_id, levels, webster_1828_compare, substrate_study, generated_at, updated_at)
        SELECT
          p.j->>'word',
          (SELECT id FROM wi),
          COALESCE(p.j->'levels', '{{}}'::jsonb),
          p.j->>'webster_1828_compare',
          p.j->>'substrate_study',
          now(),
          now()
        FROM p
        ON CONFLICT (word) DO UPDATE SET
          levels               = EXCLUDED.levels,
          webster_1828_compare = EXCLUDED.webster_1828_compare,
          substrate_study      = EXCLUDED.substrate_study,
          work_item_id         = COALESCE(EXCLUDED.work_item_id, stewards.thummim_entries.work_item_id),
          updated_at           = now()
        RETURNING word;
    """

    proc = subprocess.run(
        ["docker", "exec", "-i", PG_CONTAINER, "psql", "-U", "stewards", "-d", "stewards", "-t", "-A", "-c", sql],
        capture_output=True, text=True,
    )
    if proc.returncode != 0:
        print(f"  [FAIL]{word}: psql error — {proc.stderr.strip().splitlines()[-1]}", file=sys.stderr)
        return False
    if not proc.stdout.strip():
        print(f"  [WARN]{word}: returned no row", file=sys.stderr)
        return False
    print(f"  [OK] {word}")
    return True


def main() -> int:
    if not DICT_DIR.exists():
        print(f"ERR: {DICT_DIR} does not exist", file=sys.stderr)
        return 1
    files = sorted(DICT_DIR.glob("thummim-*.md"))
    if not files:
        print(f"ERR: no thummim-*.md files in {DICT_DIR}", file=sys.stderr)
        return 1

    print(f"Backfilling stewards.thummim_entries from {len(files)} markdown file(s):")
    ok = 0
    fail = 0
    for f in files:
        slug = f.stem
        entry = extract_entry(f)
        if entry is None:
            fail += 1
            continue
        if upsert(entry, slug):
            ok += 1
        else:
            fail += 1

    print(f"\nDone. {ok} upserted, {fail} failed.")
    return 0 if fail == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
