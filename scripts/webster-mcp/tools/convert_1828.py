#!/usr/bin/env python3
"""Convert the kayson-argyle/websters_1828 parse output into webster-mcp's JSON format.

Provenance chain (see scripts/webster-mcp/README.md):
  Noah Webster, American Dictionary of the English Language (1828, public domain)
  -> Ellen G. White Estate Archives full-text preservation (archive.org)
  -> github.com/kayson-argyle/websters_1828 (raw text + parsing pipeline)
  -> this converter (grouping, extra cleaning, webster-mcp JSON shape)

This script imports kayson-argyle's extraction modules directly from a local clone
(it does NOT vendor their code -- clone their repo and point --src at it). The
dictionary *text* is public domain; their parsing scripts remain theirs.

Usage (run with PYTHONUTF8=1 on Windows -- upstream scripts open files
without an explicit encoding):
  python convert_1828.py --src <clone-dir> --out ../data/webster1828.json.gz [--report report.txt]

Output: gzipped JSON array of WebsterEntry objects:
  {"word": "CHARITY", "pos": "noun", "definitions": ["...", ...],
   "etymology": "...", "notes": ["[Not used.]"]}

Our cleaning pass on top of theirs:
  1. Reject junk headwords (OCR artifacts like "- fall", "(:);").
  2. Strip U+FFFD replacement characters and collapse whitespace.
  3. Fix invalid numbered-book scripture refs (OCR reads "1" as "7":
     "7 Corinthians 8:1" -> "1 Corinthians 8:1"). Only applied when the
     number exceeds the canonical maximum for that book AND is 7; every
     fix is logged to the report.
  4. Targeted phrase fixes for known OCR duplications (TEXT_FIXES below).
All rejects and fixes are written to the conversion report.
"""

import argparse
import gzip
import json
import os
import re
import sys

# Headwords must look like words: letters, then letters/apostrophe/hyphen/space.
HEADWORD_RE = re.compile(r"^[A-Za-z][A-Za-z'\- ]*$")

# Canonical maximum book numbers for numbered Bible books (KJV).
BOOK_MAX = {
    "Corinthians": 2,
    "Peter": 2,
    "Timothy": 2,
    "Thessalonians": 2,
    "John": 3,
    "Samuel": 2,
    "Kings": 2,
    "Chronicles": 2,
}
REF_RE = re.compile(r"\b([0-9])\s+(" + "|".join(BOOK_MAX) + r")\b")

# Known OCR duplications / glitches verified against webstersdictionary1828.com.
# Format: (wrong, right). Applied as plain string replacement on definitions.
TEXT_FIXES = [
    (
        "inclines men to think favorably of their fellow men to think favorably "
        "of their fellow men, and to do them good",
        "inclines men to think favorably of their fellow men, and to do them good",
    ),
]

WS_RE = re.compile(r"[ \t]+")


def clean_definition(text, ref_fixes, fix_counts):
    text = text.replace("�", "")
    for wrong, right in TEXT_FIXES:
        if wrong in text:
            fix_counts[wrong] = fix_counts.get(wrong, 0) + 1
            text = text.replace(wrong, right)

    def repl(m):
        num, book = m.group(1), m.group(2)
        if int(num) > BOOK_MAX[book] and num == "7":
            fixed = "1 " + book
            ref_fixes.append(f"{m.group(0)} -> {fixed}")
            return fixed
        return m.group(0)

    text = REF_RE.sub(repl, text)
    text = WS_RE.sub(" ", text).replace(" \n", "\n").strip()
    return text


def convert(src, out_path, report_path):
    src = os.path.abspath(src)
    sys.path.insert(0, src)
    os.chdir(src)  # upstream scripts use relative paths

    from clean_ocr import clean_text  # noqa: E402
    from full_parser import parse_full_dictionary  # noqa: E402

    clean_text()
    dictionary = parse_full_dictionary()

    entries_out = []
    rejected = []
    ref_fixes = []
    fix_counts = {}
    sense_count = 0

    for key in sorted(dictionary.keys()):
        data = dictionary[key]
        headword = (data["headword"] or "").strip()
        if not HEADWORD_RE.match(headword):
            rejected.append(headword)
            continue

        # Group consecutive senses that came from the same raw entry block:
        # one block = one headword+POS entry in the original text.
        blocks = []
        last_raw = None
        for e in data["entries"]:
            if not blocks or e.get("raw_text") != last_raw:
                blocks.append([])
                last_raw = e.get("raw_text")
            blocks[-1].append(e)

        for block in blocks:
            definitions = []
            for e in block:
                d = clean_definition(e.get("definition") or "", ref_fixes, fix_counts)
                if d:
                    definitions.append(d)
                    sense_count += 1
            if not definitions:
                continue
            entry = {
                "word": headword.upper(),
                "pos": (block[0].get("part_of_speech") or "").strip(),
                "definitions": definitions,
            }
            etym = (block[0].get("etymology") or "").strip()
            if etym:
                entry["etymology"] = clean_definition(etym, ref_fixes, fix_counts)
            notes = [n.strip() for n in (block[0].get("notes") or []) if n.strip()]
            if notes:
                entry["notes"] = notes
            entries_out.append(entry)

    out_path = os.path.abspath(out_path)
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    with gzip.open(out_path, "wt", encoding="utf-8") as f:
        json.dump(entries_out, f, ensure_ascii=False)

    unique_words = len({e["word"] for e in entries_out})
    lines = [
        "Webster 1828 conversion report",
        f"source: {src}",
        f"output: {out_path}",
        f"unique words: {unique_words}",
        f"entries (word+pos blocks): {len(entries_out)}",
        f"senses: {sense_count}",
        f"rejected junk headwords: {len(rejected)}",
        f"scripture-ref fixes (7 -> 1): {len(ref_fixes)}",
        f"targeted text fixes applied: {sum(fix_counts.values())}",
        "",
        "--- rejected headwords ---",
        *rejected,
        "",
        "--- scripture-ref fixes ---",
        *ref_fixes,
        "",
        "--- targeted fixes ---",
        *(f"{count}x: {wrong[:80]}" for wrong, count in fix_counts.items()),
    ]
    if report_path:
        with open(report_path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
    print("\n".join(lines[:10]))


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--src", required=True, help="path to a kayson-argyle/websters_1828 clone")
    p.add_argument("--out", required=True, help="output .json.gz path")
    p.add_argument("--report", default=None, help="conversion report path")
    args = p.parse_args()
    convert(args.src, os.path.abspath(args.out), os.path.abspath(args.report) if args.report else None)


if __name__ == "__main__":
    main()
