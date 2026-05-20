"""
Phase 3 — extract unique words across the standard works.

Source: gospel-library/eng/scriptures/{bofm,dc-testament,pgp,nt,ot}/*.md
(local copies, gitignored). Tokenize all .md files in each volume.

Tokenization rules:
- Word = [a-zA-Z][a-zA-Z'\-]+ (letters + apostrophes + hyphens)
- Lowercased
- Strip leading/trailing apostrophes and hyphens
- Minimum length 2

Output:
- 03-canonical-unique-words.json — full data: per-word per-volume frequency
- 03-canonical-unique-words.txt — sorted plain word list (one per line)

Provenance: 2026-05-20 autonomous overnight task.
"""
from __future__ import annotations
import json
import re
import sys
from collections import Counter, defaultdict
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
SCRIPTURE_ROOT = REPO / "gospel-library/eng/scriptures"
VOLUMES = ["bofm", "dc-testament", "pgp", "nt", "ot"]

OUT_JSON = REPO / "research/gospel/1828/03-canonical-unique-words.json"
OUT_TXT = REPO / "research/gospel/1828/03-canonical-unique-words.txt"

# Markdown noise patterns to strip before tokenization
MARKDOWN_LINKS = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")
MARKDOWN_FOOTNOTE = re.compile(r"\[\^[^\]]+\]")
HTML_TAGS = re.compile(r"<[^>]+>")
CODE_FENCE = re.compile(r"```[^`]*```", re.DOTALL)
INLINE_CODE = re.compile(r"`[^`]+`")

# Word tokenizer
WORD_RE = re.compile(r"[a-zA-Z][a-zA-Z'\-]+")


def clean_text(text: str) -> str:
    """Strip markdown markers and code blocks; keep link anchor-text only."""
    # Replace links with their anchor text (we keep the visible text, drop the URL)
    text = MARKDOWN_LINKS.sub(r"\1", text)
    text = MARKDOWN_FOOTNOTE.sub("", text)
    text = CODE_FENCE.sub("", text)
    text = INLINE_CODE.sub("", text)
    text = HTML_TAGS.sub("", text)
    return text


def tokenize(text: str) -> list[str]:
    out = []
    for m in WORD_RE.finditer(text):
        w = m.group(0).strip("'-").lower()
        if len(w) >= 2:
            out.append(w)
    return out


def process_volume(volume: str) -> tuple[Counter, int, int]:
    """Returns (word_counter, n_files, n_tokens) for the volume."""
    vol_dir = SCRIPTURE_ROOT / volume
    if not vol_dir.exists():
        print(f"WARN: {vol_dir} does not exist", file=sys.stderr)
        return Counter(), 0, 0
    counter: Counter = Counter()
    n_files = 0
    n_tokens = 0
    for fp in vol_dir.rglob("*.md"):
        try:
            text = fp.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        cleaned = clean_text(text)
        tokens = tokenize(cleaned)
        counter.update(tokens)
        n_files += 1
        n_tokens += len(tokens)
    return counter, n_files, n_tokens


def main() -> int:
    OUT_JSON.parent.mkdir(parents=True, exist_ok=True)
    all_counts: dict[str, Counter] = {}
    stats = {}
    for vol in VOLUMES:
        print(f"Processing {vol}…", file=sys.stderr)
        c, nf, nt = process_volume(vol)
        all_counts[vol] = c
        stats[vol] = {"files": nf, "tokens": nt, "unique_words": len(c)}
        print(f"  files={nf}, tokens={nt:,}, unique={len(c):,}", file=sys.stderr)

    # Merge into per-word-per-volume structure
    all_words: set[str] = set()
    for c in all_counts.values():
        all_words.update(c.keys())

    word_records: dict[str, dict] = {}
    for w in sorted(all_words):
        rec = {
            "word": w,
            "total": sum(c.get(w, 0) for c in all_counts.values()),
            "by_volume": {vol: c.get(w, 0) for vol, c in all_counts.items() if c.get(w, 0) > 0},
        }
        word_records[w] = rec

    OUT_JSON.write_text(json.dumps({
        "provenance": {
            "date": "2026-05-20",
            "source": "gospel-library/eng/scriptures/{bofm,dc-testament,pgp,nt,ot}/**/*.md",
            "tokenizer": "Python regex [a-zA-Z][a-zA-Z'\\-]+ lowercased, strip leading/trailing apostrophes-hyphens, len>=2",
            "stats": stats,
        },
        "total_unique_words": len(word_records),
        "words": word_records,
    }, indent=2), encoding="utf-8")

    OUT_TXT.write_text("\n".join(sorted(all_words)) + "\n", encoding="utf-8")

    print(f"\nWrote {OUT_JSON.relative_to(REPO)}", file=sys.stderr)
    print(f"Wrote {OUT_TXT.relative_to(REPO)}", file=sys.stderr)
    print(f"Total unique words across canon: {len(word_records):,}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
