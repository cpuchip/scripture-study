#!/usr/bin/env python3
"""Scan the genuine-1828 webster JSON for OCR damage classes.

Detection classes:
  A  near-empty senses (no real text)
  B  encoding artifacts (U+FFFD remnants, stray accented chars in definitions)
  C  run-together words (lost line breaks: "henceAll")
  D  trailing/embedded junk (".W/", " = ", dangling "&", "(:)" fragments)
  E  invalid scripture references (validated against the KJV canon from
     strongs-concordance-mcp; suggests 1<->7 digit-confusion repairs)
  F  doubled phrases (repeated 4+-word runs inside one definition)
  G  missing headwords (See-references to absent entries; known tier dropouts)

Output: a machine-readable JSON report (input to the repair pipeline) and a
human summary on stdout.

Usage:
  python scan_1828.py --dict ../data/webster1828.json.gz \
      --kjv <strongs>/data/kjv-strongs.json.gz \
      [--missing-words <1828-illuminated>/frontend/src/data/definitions-1828.json] \
      --out scan-report.json
"""

import argparse
import gzip
import json
import re
from collections import Counter, defaultdict

# --- KJV canon ------------------------------------------------------

# Webster's citation style uses full book names; map them to the
# kjv-strongs abbreviations. Numbered books appear as "1 Corinthians" etc.
BOOK_NAME_TO_ABBREV = {
    "Genesis": "Gen", "Exodus": "Exo", "Leviticus": "Lev", "Numbers": "Num",
    "Deuteronomy": "Deu", "Joshua": "Jos", "Judges": "Jdg", "Ruth": "Rth",
    "1 Samuel": "1Sa", "2 Samuel": "2Sa", "1 Kings": "1Ki", "2 Kings": "2Ki",
    "1 Chronicles": "1Ch", "2 Chronicles": "2Ch", "Ezra": "Ezr",
    "Nehemiah": "Neh", "Esther": "Est", "Job": "Job", "Psalms": "Psa",
    "Psalm": "Psa", "Proverbs": "Pro", "Ecclesiastes": "Ecc",
    "Song of Solomon": "Sng", "Isaiah": "Isa", "Jeremiah": "Jer",
    "Lamentations": "Lam", "Ezekiel": "Eze", "Daniel": "Dan", "Hosea": "Hos",
    "Joel": "Joe", "Amos": "Amo", "Obadiah": "Oba", "Jonah": "Jon",
    "Micah": "Mic", "Nahum": "Nah", "Habakkuk": "Hab", "Zephaniah": "Zep",
    "Haggai": "Hag", "Zechariah": "Zec", "Malachi": "Mal",
    "Matthew": "Mat", "Mark": "Mar", "Luke": "Luk", "John": "Jhn",
    "Acts": "Act", "Romans": "Rom", "1 Corinthians": "1Co",
    "2 Corinthians": "2Co", "Galatians": "Gal", "Ephesians": "Eph",
    "Philippians": "Phl", "Colossians": "Col", "1 Thessalonians": "1Th",
    "2 Thessalonians": "2Th", "1 Timothy": "1Ti", "2 Timothy": "2Ti",
    "Titus": "Tit", "Philemon": "Phm", "Hebrews": "Heb", "James": "Jas",
    "1 Peter": "1Pe", "2 Peter": "2Pe", "1 John": "1Jo", "2 John": "2Jo",
    "3 John": "3Jo", "Jude": "Jde", "Revelation": "Rev",
}
BARE_BOOK_NAMES = {n.split(" ", 1)[-1] for n in BOOK_NAME_TO_ABBREV if n[0].isdigit()}


def load_canon(kjv_path):
    """canon[abbrev][chapter] = max verse"""
    with gzip.open(kjv_path, "rt", encoding="utf-8") as f:
        kjv = json.load(f)
    canon = {}
    for book, chapters in kjv.items():
        canon[book] = {int(c): max(int(v) for v in verses) for c, verses in chapters.items()}
    return canon


# Matches "1 Corinthians 8:1", "Ecclesiastes 12:7", and damaged variants
# like "7. Corinthians = 15:40" (stray period / equals).
REF_RE = re.compile(
    r"\b(?:([0-9])\s*\.?\s+)?"          # optional book number (possibly "7. ")
    r"([A-Z][a-z]{2,15})\s*=?\s+"        # book name, optional stray '='
    r"(\d{1,3}):(\d{1,3})\b"             # chapter:verse
)

RUNTOGETHER_RE = re.compile(r"[a-z]{2}[A-Z][a-z]{2}")
JUNK_RE = re.compile(r"\.W/|\s=\s|&\s*$|\(\s*:\s*\)|\s'\s|�")
WORD_RE = re.compile(r"[A-Za-z']+")


def digit_swaps(s):
    """Candidate 1<->7 (and 0-loss) corrections for a numeric string."""
    out = set()
    for i, ch in enumerate(s):
        if ch == "7":
            out.add(s[:i] + "1" + s[i + 1:])
        elif ch == "1":
            out.add(s[:i] + "7" + s[i + 1:])
    return out - {s}


def check_ref(canon, booknum, bookname, chapter, verse):
    """Returns (status, canonical_name) where status in valid|invalid_book|bad_chapter|bad_verse."""
    name = f"{booknum} {bookname}" if booknum else bookname
    abbrev = BOOK_NAME_TO_ABBREV.get(name)
    if abbrev is None:
        # A bare numbered-book name without its number ("Corinthians 8:1")
        if not booknum and bookname in BARE_BOOK_NAMES:
            return "missing_book_number", name
        return "unknown_book", name
    ch, v = int(chapter), int(verse)
    book = canon[abbrev]
    if ch not in book:
        return "bad_chapter", name
    if v > book[ch] or v < 1:
        return "bad_verse", name
    return "valid", name


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--dict", required=True)
    p.add_argument("--kjv", required=True)
    p.add_argument("--missing-words", default=None)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    canon = load_canon(args.kjv)
    with gzip.open(args.dict, "rt", encoding="utf-8") as f:
        entries = json.load(f)

    headwords = {e["word"].upper() for e in entries}
    report = defaultdict(list)
    stats = Counter()

    for e in entries:
        word = e["word"]
        for i, d in enumerate(e.get("definitions", [])):
            label = f"{word}#{i+1}"
            text = d.strip()

            # A: near-empty
            letters = re.sub(r"[^A-Za-z]", "", text)
            if len(letters) < 4:
                report["A_near_empty"].append({"w": label, "text": text})
                stats["A"] += 1

            # B: encoding artifacts in definition text
            odd = [ch for ch in text if ord(ch) > 127 and ch not in "—–‘’“”"]
            if odd:
                report["B_encoding"].append({"w": label, "chars": "".join(sorted(set(odd))), "text": text[:120]})
                stats["B"] += 1

            # C: run-together words
            for m in RUNTOGETHER_RE.finditer(text):
                ctx = text[max(0, m.start() - 30):m.end() + 30]
                report["C_runtogether"].append({"w": label, "ctx": ctx})
                stats["C"] += 1

            # D: junk fragments
            if JUNK_RE.search(text):
                report["D_junk"].append({"w": label, "text": text[-120:] if len(text) > 120 else text})
                stats["D"] += 1

            # E: scripture refs
            for m in REF_RE.finditer(text):
                booknum, bookname, ch, v = m.groups()
                status, name = check_ref(canon, booknum, bookname, ch, v)
                if status == "valid":
                    stats["E_valid"] += 1
                    continue
                suggestions = []
                if status in ("bad_chapter", "bad_verse", "unknown_book", "missing_book_number"):
                    # try digit swaps on book number, chapter, verse
                    for bn in ([booknum] if booknum else [None, "1", "2", "3"]):
                        bn_cands = {bn} | (digit_swaps(bn) if bn else set())
                        for bnc in bn_cands:
                            for chc in {ch} | digit_swaps(ch):
                                for vc in {v} | digit_swaps(v):
                                    if (bnc, chc, vc) == (booknum, ch, v):
                                        continue
                                    s2, n2 = check_ref(canon, bnc, bookname, chc, vc)
                                    if s2 == "valid":
                                        suggestions.append(f"{n2} {chc}:{vc}")
                report["E_bad_refs"].append({
                    "w": label, "ref": m.group(0), "status": status,
                    "suggest": sorted(set(suggestions))[:4],
                })
                stats["E_bad"] += 1

            # F: doubled phrases (adjacent repeated 4-grams)
            words = WORD_RE.findall(text.lower())
            for j in range(len(words) - 8):
                if words[j:j + 4] == words[j + 4:j + 8]:
                    report["F_doubled"].append({"w": label, "phrase": " ".join(words[j:j + 4])})
                    stats["F"] += 1
                    break

        # G: See-references to absent headwords
        for d in e.get("definitions", []):
            for m in re.finditer(r"\bSee\s+([A-Z][a-zA-Z\-]{2,})", d):
                target = m.group(1).upper().rstrip(".")
                if target not in headwords and target not in ("THE", "ALSO", "UNDER", "NOTE"):
                    report["G_missing_see_ref"].append({"w": e["word"], "target": target})
                    stats["G"] += 1

    # G2: known tier-word dropouts
    if args.missing_words:
        with open(args.missing_words, encoding="utf-8") as f:
            mw = json.load(f).get("missing_words", [])
        report["G_known_missing_tier_words"] = mw

    # de-duplicate G refs
    seen = set()
    dedup = []
    for r in report["G_missing_see_ref"]:
        k = (r["w"], r["target"])
        if k not in seen:
            seen.add(k)
            dedup.append(r)
    report["G_missing_see_ref"] = dedup

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump({"stats": dict(stats), "report": report}, f, ensure_ascii=False, indent=1)

    print(f"entries: {len(entries)}  unique headwords: {len(headwords)}")
    for k in sorted(stats):
        print(f"  {k}: {stats[k]}")
    print(f"G_missing_see_ref (deduped): {len(report['G_missing_see_ref'])}")
    print(f"report -> {args.out}")


if __name__ == "__main__":
    main()
