"""
Fetch modern definitions for the tier-words list from the Free Dictionary API
(same source webster-mcp uses for `modern_define`).

Rate-limited to 1 word/second to be polite to the API. Resumable: on each
run, reads existing definitions-modern.json and skips already-fetched words.

Output: frontend/src/data/definitions-modern.json — keyed by word, value is
list of pos+definitions entries (or null when the word isn't in the modern
dictionary, which is a real signal — it means the word is sufficiently
archaic that mainstream modern dictionaries don't cover it).
"""
from __future__ import annotations
import json
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
PROJECT = REPO / "projects/1828-illuminated"
DATA = PROJECT / "frontend/src/data"
WORDLIST = PROJECT / "scripts/fetch-wordlist.txt"
OUTFILE = DATA / "definitions-modern.json"

API_URL = "https://api.dictionaryapi.dev/api/v2/entries/en/"
DELAY_SECONDS = 1.0  # Be polite — Free Dictionary API is no-auth + community-supported


def fetch_one(word: str) -> dict | None:
    """Returns a normalized definitions structure, or None on 404 (not in dictionary)."""
    url = API_URL + urllib.request.quote(word)
    req = urllib.request.Request(url, headers={"User-Agent": "1828-illuminated-build/0.1"})
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return None
        raise
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as e:
        print(f"  WARN: {word} → {e}", file=sys.stderr)
        return {"error": str(e)}

    # Normalize: [{pos, definitions: [text...]}]
    out: list[dict] = []
    if isinstance(data, list):
        for entry in data:
            for meaning in entry.get("meanings", []):
                defs = [d.get("definition", "") for d in meaning.get("definitions", [])]
                if defs:
                    out.append({
                        "pos": meaning.get("partOfSpeech", ""),
                        "definitions": defs,
                    })
    return {"entries": out} if out else None


def main() -> int:
    if not WORDLIST.exists():
        print(f"ERR: wordlist {WORDLIST} not found — run build_data.py first", file=sys.stderr)
        return 1
    words = [w.strip() for w in WORDLIST.read_text(encoding="utf-8").splitlines() if w.strip()]

    # Resume from existing data
    if OUTFILE.exists():
        existing = json.loads(OUTFILE.read_text(encoding="utf-8"))
        results = existing.get("definitions", {})
    else:
        results = {}
    todo = [w for w in words if w not in results]
    skipped = len(words) - len(todo)
    print(f"Wordlist: {len(words)} total · {skipped} already fetched · {len(todo)} to fetch", file=sys.stderr)
    print(f"Estimated time: {len(todo) * DELAY_SECONDS:.0f}s ({len(todo) * DELAY_SECONDS / 60:.1f} min)", file=sys.stderr)

    n_found = sum(1 for v in results.values() if v and "entries" in v)
    n_404 = sum(1 for v in results.values() if v is None)
    n_err = sum(1 for v in results.values() if v and "error" in v)

    try:
        for i, word in enumerate(todo, 1):
            res = fetch_one(word)
            results[word] = res
            if res is None:
                n_404 += 1
                status = "404"
            elif "error" in (res or {}):
                n_err += 1
                status = "ERR"
            else:
                n_found += 1
                status = f"OK ({len(res['entries'])} sense-groups)"
            print(f"  [{i:4d}/{len(todo):4d}] {word:28s} {status}", flush=True)

            # Persist every 10 words so a crash doesn't lose progress
            if i % 10 == 0:
                OUTFILE.write_text(json.dumps({
                    "generated_at": "2026-05-20",
                    "source": "https://api.dictionaryapi.dev/api/v2/entries/en (Free Dictionary API)",
                    "stats": {"found": n_found, "not_found_404": n_404, "errors": n_err},
                    "definitions": results,
                }, indent=2), encoding="utf-8")

            time.sleep(DELAY_SECONDS)
    except KeyboardInterrupt:
        print("\nInterrupted — saving progress…", file=sys.stderr)

    # Final write
    OUTFILE.parent.mkdir(parents=True, exist_ok=True)
    OUTFILE.write_text(json.dumps({
        "generated_at": "2026-05-20",
        "source": "https://api.dictionaryapi.dev/api/v2/entries/en (Free Dictionary API)",
        "stats": {"found": n_found, "not_found_404": n_404, "errors": n_err},
        "definitions": results,
    }, indent=2), encoding="utf-8")
    print(f"\nDone. found={n_found} not_found={n_404} errors={n_err}", file=sys.stderr)
    print(f"Wrote {OUTFILE.relative_to(REPO)}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
