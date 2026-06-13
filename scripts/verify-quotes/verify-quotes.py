#!/usr/bin/env python3
"""verify-quotes — a manual pre-publish quote checker (v1).

Born 2026-06-13 from the study-correctness walk + scratch-audit fan-out: the whole
effort was a deterministic check in disguise. This turns the highest-value,
most-tractable slice — verifying quoted **Webster 1828** definitions — into a
script, so the 1913-as-1828 contamination class becomes catchable in seconds.

KEY IDEA: 1913 and 1828 share phrases, so "the quote appears in 1828" is not
enough. We have BOTH editions (webster1828 + webster1913). For each quoted
Webster definition we measure how well it matches the genuine 1828 entry vs the
1913 entry for that word. If it matches **1913 better than 1828**, it is the
contamination class — FLAG it.

  OK      — the quote matches the genuine 1828 entry (and not 1913-better)
  FLAG    — the quote matches the 1913 entry better than 1828 (contaminated),
            or matches neither well (possible fabrication / paraphrase-in-quotes)
  ABSENT  — the word is not an 1828 headword at all (e.g. "telestial")

Quotes that match NEITHER edition for the word are treated as "not a Webster
definition quote" (e.g. a scripture quote sharing the paragraph) and skipped.

NOT a pre-commit hook yet — run it manually; promote it only once it keeps
earning its weight. v2+ roadmap (scripture verbatim, embedded-citation depth-2,
talk quotes) is in scripts/verify-quotes/README.md.

Usage:
  python scripts/verify-quotes/verify-quotes.py <file.md> ...
  git diff --name-only '*.md' | python scripts/verify-quotes/verify-quotes.py
Exit: 0 if no FLAG/ABSENT, else 1.
"""
import sys, os, re, gzip, json
from difflib import SequenceMatcher

HERE = os.path.dirname(os.path.abspath(__file__))
D828 = os.path.join(HERE, "..", "webster-mcp", "data", "webster1828.json.gz")
D1913 = os.path.join(HERE, "..", "webster-mcp", "data", "webster1913.json.gz")

def load(path):
    with gzip.open(path, "rt", encoding="utf-8") as f:
        entries = json.load(f)
    by = {}
    for e in entries:
        w = (e.get("word") or "").strip().upper()
        if not w:
            continue
        blob = " ".join(e.get("definitions") or [])
        by[w] = (by.get(w, "") + " " + blob).strip()
    return by

def norm(s):
    s = s.lower().replace("’", "'").replace("“", '"').replace("”", '"').replace("…", " ")
    s = re.sub(r"[^a-z0-9 ]+", " ", s)
    return re.sub(r"\s+", " ", s).strip()

def overlap(quote, entry):
    """Longest contiguous normalized match between quote and entry, as a
    fraction of the quote length (0..1). 1.0 = the quote sits verbatim in entry."""
    q, e = norm(quote), norm(entry)
    if not q or not e:
        return 0.0
    m = SequenceMatcher(None, q, e, autojunk=False).find_longest_match(0, len(q), 0, len(e))
    return m.size / len(q)

STEMS = ["", "S", "ES", "D", "ED", "ING", "LY", "ION", "OR", "ER"]
def lookup_word(word, by):
    W = word.upper()
    if W in by:
        return W
    for suf in ("S", "ES", "D", "ED", "ING", "LY", "OR", "ER", "ION", "MENT", "NESS"):
        if W.endswith(suf) and W[: -len(suf)] in by:
            return W[: -len(suf)]
        if W.endswith(suf) and (W[: -len(suf)] + "E") in by:
            return W[: -len(suf)] + "E"
    return None

WEBSTER_RE = re.compile(r"webster", re.I)
EIGHT28_RE = re.compile(r"1828")
ITALIC_WORD = re.compile(r"\*([A-Za-z][A-Za-z-]{2,})\*")
QUOTED = re.compile(r"[\"“]([^\"“”]{10,}?)[\"”]")
MATCH_FLOOR = 0.55   # a "Webster def quote" must match >=55% of one edition
MARGIN = 0.08        # 1913 must beat 1828 by this much to call it contaminated

def check_file(path, w828, w1913):
    text = open(path, encoding="utf-8").read()
    out = []
    for p in re.split(r"\r?\n\s*\r?\n", text):
        if not (WEBSTER_RE.search(p) and EIGHT28_RE.search(p)):
            continue
        quotes = QUOTED.findall(p)
        words = ITALIC_WORD.findall(p)
        if not quotes or not words:
            continue
        for w in dict.fromkeys(words):                 # de-dup, preserve order
            key = lookup_word(w, w828)
            if not key:
                # only report ABSENT if it looks like it's being defined here
                if any(overlap(q, w1913.get(w.upper(), "")) >= MATCH_FLOOR for q in quotes):
                    out.append(("ABSENT", w, "not an 1828 headword (but present in 1913)"))
                continue
            g828, g1913 = w828.get(key, ""), w1913.get(key, "")
            # a Webster-def quote is substantial (>=4 words) and matches an
            # edition; short fragments (scripture bits, study prose) are skipped.
            defquotes = [q for q in quotes if len(norm(q).split()) >= 4
                         and max(overlap(q, g828), overlap(q, g1913)) >= MATCH_FLOOR]
            if not defquotes:
                continue
            q = max(defquotes, key=lambda x: max(overlap(x, g828), overlap(x, g1913)))
            o828, o1913 = overlap(q, g828), overlap(q, g1913)
            # FLAG only on a STRONG 1913 match that clearly beats 1828 — the
            # contamination signature. (Weak/ambiguous cases are left for the
            # human; v1 optimizes for precision over recall.)
            if o1913 >= 0.82 and o1913 > o828 + MARGIN:
                out.append(("FLAG", key,
                            f"matches 1913 ({o1913:.2f}) > 1828 ({o828:.2f}) — "
                            f'quoted: "{q[:60]}…"'))
            else:
                out.append(("OK", key, f"1828 {o828:.2f}"))
    return out

def main(argv):
    files = argv[1:] or ([l.strip() for l in sys.stdin if l.strip()] if not sys.stdin.isatty() else [])
    files = [f for f in files if f.endswith(".md") and os.path.isfile(f)]
    if not files:
        print("usage: verify-quotes.py <file.md> ...", file=sys.stderr); return 2
    w828, w1913 = load(D828), load(D1913)
    bad = 0
    for path in files:
        res = check_file(path, w828, w1913)
        flags = [r for r in res if r[0] != "OK"]
        if not res:
            continue
        ok = sum(1 for r in res if r[0] == "OK")
        head = f"{path}  ({ok} Webster OK" + (f", {len(flags)} to review)" if flags else ")")
        print(("\n" + head) if flags else head)
        for level, word, msg in flags:
            print(f"  [{level}] {word}: {msg}"); bad += 1
    print(f"\n— {len(files)} file(s); {bad} Webster flag(s).")
    return 1 if bad else 0

if __name__ == "__main__":
    sys.exit(main(sys.argv))
