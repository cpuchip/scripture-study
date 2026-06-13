#!/usr/bin/env python3
"""sources — pull verbatim text from the two deterministic sources.

  scripture: verse text from a gospel-library chapter file (the `**N.** text`
             format, footnote <sup> markers stripped).
  webster:   genuine 1828 definition(s) for a word (from webster1828.json.gz —
             the same data verify-quotes diffs against).

No retyping anywhere: the quoter pulls from here, so what lands in the document is
the source's own bytes.
"""
import os, re, gzip, json
from functools import lru_cache
import resolver

HERE = os.path.dirname(os.path.abspath(__file__))
D828 = os.path.join(HERE, "..", "webster-mcp", "data", "webster1828.json.gz")
D1913 = os.path.join(HERE, "..", "webster-mcp", "data", "webster1913.json.gz")

_SUP = re.compile(r"<sup[^>]*>.*?</sup>", re.S)
_VERSE = re.compile(r"^\*\*(\d+)\.\*\*\s*(.*)$")

def _clean(text):
    text = _SUP.sub("", text)
    return re.sub(r"\s+", " ", text).strip()

@lru_cache(maxsize=64)
def _chapter_verses(path):
    """{verse_int: clean_text} for a chapter file."""
    out = {}
    with open(path, encoding="utf-8") as f:
        for line in f:
            m = _VERSE.match(line.rstrip("\n"))
            if m:
                out[int(m.group(1))] = _clean(m.group(2))
    return out

def verses(ref):
    """Verbatim verse text for a ref. Single verse -> that verse; range -> the
    verses joined with a space (matching how the studies render inline ranges);
    chapter-only -> all verses joined. Returns (label, text)."""
    p = ref if isinstance(ref, dict) else resolver.parse_ref(ref)
    path = resolver.file_path(p)
    if not os.path.exists(path):
        raise FileNotFoundError(path)
    chap = _chapter_verses(path)
    if p["v1"] is None:
        nums = sorted(chap)
    else:
        v2 = p["v2"] or p["v1"]
        nums = list(range(p["v1"], v2 + 1))
    missing = [n for n in nums if n not in chap]
    if missing:
        raise KeyError(f"{p['label']}: verse(s) {missing} not found in {os.path.basename(path)}")
    text = " ".join(chap[n] for n in nums)
    return p["label"], text

@lru_cache(maxsize=2)
def _webster(path):
    with gzip.open(path, "rt", encoding="utf-8") as f:
        entries = json.load(f)
    by = {}
    for e in entries:
        w = (e.get("word") or "").strip().upper()
        if w:
            by.setdefault(w, []).extend(e.get("definitions") or [])
    return by

def webster_defs(word, edition="1828"):
    """List of genuine definition strings for a word (1828 by default)."""
    by = _webster(D828 if edition == "1828" else D1913)
    return by.get(word.strip().upper(), [])

if __name__ == "__main__":
    import sys
    if sys.argv[1] == "webster":
        for i, d in enumerate(webster_defs(sys.argv[2]), 1):
            print(f"[{i}] {d}")
    else:
        label, text = verses(sys.argv[1])
        print(f"{label}: {text}")
