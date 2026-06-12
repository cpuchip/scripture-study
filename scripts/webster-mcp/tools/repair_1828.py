#!/usr/bin/env python3
"""Repair OCR damage in the genuine-1828 webster JSON.

Mechanical passes (every change goes to the ledger):
  1  scripture refs   — fuzzy book-name repair (Wark->Mark), 1<->7 digit
                        swaps, KJV-verse-text disambiguation for ambiguous
                        candidates (e.g. "Corinthians 8:12")
  2  run-together     — lost sentence breaks ("adversityEcclesiastes") get
                        ". " inserted at the junction
  3  junk strip       — " = ", "' . ", U+FFFD/euro/copyright/yen/guillemet,
                        ".W/" fragments, dangling "&", double spaces
  4  fragments        — standalone "Obs." merges into the previous sense;
                        pronunciation stubs ("ake.") drop when the entry
                        has real senses; sole-sense fragments are kept and
                        listed for authority repair
  5  doubled phrases  — adjacent repeated 4+-word runs collapse to one

Overlay pass (from fetch_site_entries.py output):
  6  entries whose content was destroyed are replaced from the
     webstersdictionary1828.com transcription (provenance-tagged).

Usage:
  python repair_1828.py --dict ../data/webster1828.json.gz \
      --kjv <strongs>/data/kjv-strongs.json.gz \
      [--overlay site-entries.json] [--dry-run] \
      --out ../data/webster1828.json.gz --ledger repair-ledger.json
"""

import argparse
import gzip
import json
import re
from collections import Counter

from scan_1828 import BOOK_NAME_TO_ABBREV, BARE_BOOK_NAMES, REF_RE, digit_swaps

STOPWORDS = {
    "the", "and", "of", "to", "a", "in", "that", "for", "is", "be", "with",
    "his", "he", "they", "them", "shall", "not", "it", "i", "my", "thy",
    "ye", "you", "their", "but", "as", "are", "was", "were", "have", "hath",
}


def load_kjv(path):
    with gzip.open(path, "rt", encoding="utf-8") as f:
        return json.load(f)


def verse_text(kjv, abbrev, ch, v):
    try:
        return kjv[abbrev][str(ch)][str(v)]["text"]
    except KeyError:
        return None


def content_words(text):
    return {w for w in re.findall(r"[a-z']+", text.lower()) if w not in STOPWORDS and len(w) > 2}


def _edit1(a, b):
    """True if edit distance(a, b) == 1 (substitution / insert / delete)."""
    if len(a) == len(b):
        return sum(1 for x, y in zip(a, b) if x != y) == 1
    longer, shorter = (a, b) if len(a) > len(b) else (b, a)
    if len(longer) - len(shorter) != 1:
        return False
    return any(longer[:i] + longer[i + 1:] == shorter for i in range(len(longer)))


def _edit_le2(a, b):
    """True if edit distance(a, b) <= 2 (full DP; strings are short)."""
    if abs(len(a) - len(b)) > 2:
        return False
    prev = list(range(len(b) + 1))
    for i, ca in enumerate(a, 1):
        cur = [i]
        for j, cb in enumerate(b, 1):
            cur.append(min(prev[j] + 1, cur[j - 1] + 1, prev[j - 1] + (ca != cb)))
        prev = cur
    return prev[-1] <= 2


# "Solomon 5:6" — Webster cites Song of Solomon by its last word.
NAME_ALIASES = {"Solomon": "Song of Solomon", "Songs": "Song of Solomon",
                "Canticles": "Song of Solomon"}


def book_candidates(bookname):
    """Exact, aliased, or fuzzy matches to known bare book names."""
    if bookname in NAME_ALIASES:
        return [NAME_ALIASES[bookname]]
    names = {n.split(" ", 1)[-1] for n in BOOK_NAME_TO_ABBREV} | set(BOOK_NAME_TO_ABBREV)
    bare = {n for n in names if " " not in n}
    if bookname in bare:
        return [bookname]
    out = [n for n in bare if abs(len(n) - len(bookname)) <= 1 and _edit1(n, bookname)]
    if not out and len(bookname) >= 6:
        # longer names tolerate two errors when the match is unique
        out2 = [n for n in bare if len(n) >= 6 and _edit_le2(n, bookname)]
        if len(out2) == 1:
            out = out2
    return out


def valid_ref(kjv, booknum, bookname, ch, v):
    name = f"{booknum} {bookname}" if booknum else bookname
    abbrev = BOOK_NAME_TO_ABBREV.get(name)
    if abbrev is None:
        return None
    t = verse_text(kjv, abbrev, int(ch), int(v))
    return (name, abbrev) if t is not None else None


def repair_ref(kjv, context, booknum, bookname, ch, v):
    """Returns (fixed_text, how) or None. context = definition text for
    quote-overlap disambiguation."""
    candidates = []
    for bn in book_candidates(bookname):
        nums = [booknum] if booknum else ([None] +
               (["1", "2", "3"] if bn in BARE_BOOK_NAMES else []))
        for num in nums:
            num_cands = {num} | (digit_swaps(num) if num else set())
            for nc in num_cands:
                for chc in {ch} | digit_swaps(ch):
                    for vc in {v} | digit_swaps(v):
                        ok = valid_ref(kjv, nc, bn, chc, vc)
                        if ok:
                            candidates.append((ok[0], ok[1], chc, vc))
    # drop duplicates, drop the original (it was invalid or we wouldn't be here)
    seen = set()
    uniq = []
    for c in candidates:
        if c not in seen:
            seen.add(c)
            uniq.append(c)
    if not uniq:
        return None
    if len(uniq) == 1:
        name, _, chc, vc = uniq[0]
        return f"{name} {chc}:{vc}", "single-candidate"
    # disambiguate by quote overlap with the verse text
    ctx_words = content_words(context)
    scored = []
    for name, abbrev, chc, vc in uniq:
        vt = verse_text(kjv, abbrev, int(chc), int(vc)) or ""
        overlap = len(ctx_words & content_words(vt))
        scored.append((overlap, name, chc, vc))
    scored.sort(reverse=True)
    if scored[0][0] >= 3 and (len(scored) == 1 or scored[0][0] > scored[1][0]):
        _, name, chc, vc = scored[0]
        return f"{name} {chc}:{vc}", f"text-overlap({scored[0][0]})"
    return None


RUNTOGETHER_RE = re.compile(r"([a-z]{2})([A-Z][a-z]{2})")
PRON_STUB_RE = re.compile(r"^[a-zA-Z�'é .,-]{1,10}$")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--dict", required=True)
    p.add_argument("--kjv", required=True)
    p.add_argument("--overlay", default=None)
    p.add_argument("--out", required=True)
    p.add_argument("--ledger", required=True)
    p.add_argument("--dry-run", action="store_true")
    args = p.parse_args()

    kjv = load_kjv(args.kjv)
    with gzip.open(args.dict, "rt", encoding="utf-8") as f:
        entries = json.load(f)

    ledger = {"refs": [], "runtogether": [], "junk": [], "fragments": [],
              "doubled": [], "overlay": [], "unresolved_refs": [],
              "sole_sense_fragments": []}
    stats = Counter()

    overlay = {}
    if args.overlay:
        with open(args.overlay, encoding="utf-8") as f:
            overlay = {k.upper(): v for k, v in json.load(f).items()}

    out_entries = []
    overlaid_words = set()

    for e in entries:
        word = e["word"]

        # Pass 6: full-entry overlay from the site transcription
        if word.upper() in overlay and word.upper() not in overlaid_words:
            for oe in overlay[word.upper()]["entries"]:
                out_entries.append({"word": word.upper(), **oe})
            ledger["overlay"].append({"w": word, "source": overlay[word.upper()].get("source", "webstersdictionary1828.com")})
            stats["overlay"] += 1
            overlaid_words.add(word.upper())
            continue
        if word.upper() in overlaid_words:
            continue  # additional pos-blocks of an overlaid word

        new_defs = []
        for d in e.get("definitions", []):
            orig = d

            # Pass 3: junk strip
            d2 = d.replace("�", "").replace("€", "").replace("©", "")
            d2 = d2.replace("¥", "").replace("«", "")
            d2 = d2.replace(" = ", " ").replace("' . ", "").replace(".W/ .", ".").replace(".W/", ".")
            d2 = re.sub(r"\s+'\s*\.\s*$", ".", d2)   # trailing "… ' ."
            d2 = re.sub(r"\s+'\s+", " ", d2)          # stray mid-text " ' "
            d2 = re.sub(r"\s*&\s*$", "", d2)
            d2 = re.sub(r"[ \t]{2,}", " ", d2).strip()
            if d2 != orig:
                ledger["junk"].append({"w": word, "before": orig[:90], "after": d2[:90]})
                stats["junk"] += 1

            # Pass 2: run-together
            def rt_fix(m):
                return m.group(1) + ". " + m.group(2)
            d3 = RUNTOGETHER_RE.sub(rt_fix, d2)
            if d3 != d2:
                ledger["runtogether"].append({"w": word, "ctx": d2[:120], "after": d3[:120]})
                stats["runtogether"] += 1

            # Pass 1: scripture refs
            def ref_fix(m):
                booknum, bookname, ch, v = m.groups()
                if valid_ref(kjv, booknum, bookname, ch, v):
                    return m.group(0)
                fixed = repair_ref(kjv, d3, booknum, bookname, ch, v)
                if fixed:
                    ledger["refs"].append({"w": word, "before": m.group(0), "after": fixed[0], "how": fixed[1]})
                    stats["refs"] += 1
                    return fixed[0]
                ledger["unresolved_refs"].append({"w": word, "ref": m.group(0)})
                stats["unresolved_refs"] += 1
                return m.group(0)
            d4 = REF_RE.sub(ref_fix, d3)

            # Pass 5: doubled phrases (adjacent repeated 4-gram, collapse once)
            words_l = re.findall(r"\S+", d4)
            for j in range(len(words_l) - 8):
                a = [re.sub(r"\W", "", w).lower() for w in words_l[j:j + 4]]
                b = [re.sub(r"\W", "", w).lower() for w in words_l[j + 4:j + 8]]
                if a == b and any(a):
                    collapsed = " ".join(words_l[:j + 4] + words_l[j + 8:])
                    ledger["doubled"].append({"w": word, "phrase": " ".join(words_l[j:j + 4])})
                    stats["doubled"] += 1
                    d4 = collapsed
                    break

            new_defs.append(d4)

        # Pass 4: fragment senses
        cleaned = []
        for d in new_defs:
            stripped = d.strip()
            letters = re.sub(r"[^A-Za-z]", "", stripped)
            if len(letters) >= 4:
                cleaned.append(stripped)
                continue
            if stripped.rstrip(".").lower() == "obs" and cleaned:
                cleaned[-1] = cleaned[-1].rstrip() + " Obs."
                ledger["fragments"].append({"w": word, "action": "merged-obs"})
                stats["frag_obs"] += 1
            elif len(new_defs) > 1 and PRON_STUB_RE.match(stripped):
                ledger["fragments"].append({"w": word, "action": "dropped-stub", "text": stripped})
                stats["frag_drop"] += 1
            else:
                cleaned.append(stripped)
                ledger["sole_sense_fragments"].append({"w": word, "text": stripped})
                stats["frag_kept"] += 1

        if cleaned:
            out_entries.append({**e, "definitions": cleaned})
        else:
            ledger["fragments"].append({"w": word, "action": "entry-emptied-kept-original"})
            out_entries.append(e)

    # overlay words that are NEW (not in the original data)
    existing = {e["word"].upper() for e in entries}
    for w, ov in overlay.items():
        if w not in existing:
            for oe in ov["entries"]:
                out_entries.append({"word": w, **oe})
            ledger["overlay"].append({"w": w, "source": ov.get("source", "site"), "new": True})
            stats["overlay_new"] += 1

    print("repair stats:")
    for k in sorted(stats):
        print(f"  {k}: {stats[k]}")

    with open(args.ledger, "w", encoding="utf-8") as f:
        json.dump(ledger, f, ensure_ascii=False, indent=1)
    print(f"ledger -> {args.ledger}")

    if not args.dry_run:
        with gzip.open(args.out, "wt", encoding="utf-8") as f:
            json.dump(out_entries, f, ensure_ascii=False)
        print(f"wrote {args.out} ({len(out_entries)} entries)")
    else:
        print("dry run — no output written")


if __name__ == "__main__":
    main()
