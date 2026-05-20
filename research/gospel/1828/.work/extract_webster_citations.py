"""
Extract every Webster 1828 citation from study/**/*.md (excluding .scratch),
including the word(s) being defined.

Strategy: for each line that mentions "Webster 1828" or "Webster's 1828",
collect a 3-line window (the citation line + 2 lines after, since
definitions typically follow). From that window, extract every word
that appears in explicit markup — bold, italic, or double quotes —
and treat each as a defined-word candidate. Reject stop words and
words that appear in the surrounding *Restoration* scripture quote
context rather than as the dictionary lookup target.

A single citation line can yield multiple defined words (e.g. "Webster
1828 draws a distinction between *burden* and *load*"). Output one
row per (study, line, word).

Provenance: 2026-05-20 autonomous overnight task per Michael's directive.
No MCP calls; pure file read.
"""
from __future__ import annotations
import json
import re
import sys
from collections import Counter
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
STUDY_DIR = REPO / "study"
OUT_JSONL = REPO / "research/gospel/1828/.work/raw-citations.jsonl"

CITATION_PATTERN = re.compile(r"Webster(?:'s)? 1828", re.IGNORECASE)

# Word-extractors. All require explicit markup so we don't grab function words.
WORD_QUOTED = re.compile(r"[\"“]([A-Za-z][A-Za-z\-]{1,40})[\"”]")
WORD_ITALIC = re.compile(r"(?<!\*)\*([A-Za-z][A-Za-z\-]{2,40})\*(?!\*)")
WORD_BOLD = re.compile(r"\*\*([A-Za-z][A-Za-z\-]{2,40})\*\*")
WORD_HEADING_COLON = re.compile(
    r"^#{1,6}\s+Webster(?:'s)? 1828:?\s*[\"“*_]?([A-Za-z][A-Za-z\-]{2,40})[\"”*_]?\s*$",
    re.IGNORECASE,
)

STOP_WORDS = {
    "the", "and", "is", "a", "an", "of", "to", "in", "on", "for", "as", "by",
    "with", "or", "but", "if", "then", "than", "that", "this", "these", "those",
    "it", "its", "was", "were", "be", "been", "being", "are", "have", "has", "had",
    "i", "we", "you", "he", "she", "they", "them", "us", "me", "my", "our", "your",
    "his", "her", "their", "our", "ours",
    "defines", "definition", "definitions", "defining", "means", "meaning",
    "says", "saying", "gives", "giving", "captures", "shows", "showing",
    "draws", "drawing", "catches", "tracks", "includes", "noted", "notes",
    "two", "three", "four", "five", "sense", "senses", "first", "second", "third",
    "dictionary", "dictionaries",
    "webster", "websters", "noah",
    "lord", "god", "christ", "jesus", "spirit", "holy",  # too generic for highlight
    "v", "vs",
    # Quoted scripture-text fragments that aren't dictionary lookups
    "and now i say unto you", "behold", "yea", "amen",
}

# Heuristic tags
DIFFER_HINTS = re.compile(
    r"\b(differ|differs|diverge|contrast|whereas|opposite|surprising|not what|"
    r"isn't|isn.t|isnt|doesn't mean|doesn.t mean|different from|broader than|"
    r"narrower than|sharper than|reverses|illuminat|expand|recover|miss|"
    r"flatten|forgotten|archaic|drift|shift|recover|hidden)",
    re.IGNORECASE,
)
REINFORCE_HINTS = re.compile(
    r"\b(consistent|aligns|matches|same|reinforce|confirms|echoes|tracks|"
    r"clarif|sharpen|precise|exact|fits|directly|already)",
    re.IGNORECASE,
)


def _clean(word: str) -> str | None:
    w = word.strip().lower().rstrip(",.;:!?")
    if not w or w in STOP_WORDS or len(w) < 3:
        return None
    # Reject words that are obviously not single dictionary entries
    if " " in w or w.isdigit():
        return None
    return w


def extract_words(window_lines: list[str]) -> list[str]:
    """Pull every marked-up word from a 3-line window. Dedupe but preserve order."""
    found: list[str] = []
    seen: set[str] = set()

    def push(w: str | None):
        if w and w not in seen:
            seen.add(w)
            found.append(w)

    for ln in window_lines:
        # Heading style with colon
        mh = WORD_HEADING_COLON.match(ln)
        if mh:
            push(_clean(mh.group(1)))
        # Bold caps/title
        for m in WORD_BOLD.finditer(ln):
            push(_clean(m.group(1)))
        # Quoted
        for m in WORD_QUOTED.finditer(ln):
            push(_clean(m.group(1)))
        # Italic (single *word*)
        for m in WORD_ITALIC.finditer(ln):
            push(_clean(m.group(1)))
    return found


def classify(context: str) -> str:
    if DIFFER_HINTS.search(context):
        return "differ"
    if REINFORCE_HINTS.search(context):
        return "reinforce"
    return "neutral"


def relative(p: Path) -> str:
    return str(p.relative_to(REPO)).replace("\\", "/")


def scan_file(path: Path) -> list[dict]:
    if ".scratch" in path.parts:
        return []
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []
    lines = text.splitlines()
    out: list[dict] = []
    for i, line in enumerate(lines):
        if not CITATION_PATTERN.search(line):
            continue
        # 3-line window: citation line + 2 after
        window = lines[i : min(len(lines), i + 3)]
        words = extract_words(window)
        # Context for classification: ±3 lines around
        start = max(0, i - 3)
        end = min(len(lines), i + 6)
        ctx = "\n".join(lines[start:end])
        tag = classify(ctx)
        # Truncate excerpt
        excerpt = " ".join(line for line in window if line.strip())[:400]

        if not words:
            out.append({
                "study": relative(path),
                "line": i + 1,
                "word": None,
                "citation_line": line.strip()[:300],
                "definition_excerpt": excerpt,
                "tag": tag,
            })
        else:
            for w in words:
                out.append({
                    "study": relative(path),
                    "line": i + 1,
                    "word": w,
                    "citation_line": line.strip()[:300],
                    "definition_excerpt": excerpt,
                    "tag": tag,
                })
    return out


def main() -> int:
    OUT_JSONL.parent.mkdir(parents=True, exist_ok=True)
    files = sorted(STUDY_DIR.rglob("*.md"))
    all_rows: list[dict] = []
    for fp in files:
        all_rows.extend(scan_file(fp))
    OUT_JSONL.write_text(
        "\n".join(json.dumps(r, ensure_ascii=False) for r in all_rows) + "\n",
        encoding="utf-8",
    )
    print(f"Wrote {len(all_rows)} rows to {OUT_JSONL.relative_to(REPO)}", file=sys.stderr)

    attributed = [r for r in all_rows if r["word"]]
    unattributed = [r for r in all_rows if not r["word"]]
    print(f"\nAttributed: {len(attributed)}", file=sys.stderr)
    print(f"Unattributed: {len(unattributed)}", file=sys.stderr)

    words = Counter(r["word"] for r in attributed)
    print(f"\nUnique words: {len(words)}", file=sys.stderr)
    print("Top 50 by citation count:")
    for w, n in words.most_common(50):
        print(f"  {n:3d}  {w}")

    by_tag = Counter(r["tag"] for r in attributed)
    print(f"\nTag distribution (attributed): {dict(by_tag)}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
