#!/usr/bin/env python3
"""scripture-verbatim — the detector half of the study-tooling suite.

For each scripture link in a study, verify the quoted text next to it is verbatim
against the actual gospel-library verse. The #1 thing the 469-file walk did by hand;
this turns it into a script. Reuses the quoter's spine — `resolver` (ref->file),
`sources` (verse text), `grammar` (the verbatim-spans + marked-edits engine). One
engine, both ends: the quoter PRODUCES well-formed quotes, this ENFORCES them.

Two hard parts, both handled here:
  1. ASSOCIATION — pair each scripture link with the quote it belongs to. A quote
     immediately precedes its link, inline ("phrase" ([Ref](link))) or in a
     blockquote (> "verse" / > — [Ref](link)). Streamed, nearest-preceding,
     within a proximity window.
  2. PRECISION — speak the quote grammar (partial phrases, [brackets], ... ellipses
     are legitimate; only UNMARKED deviations are errors), and don't flag a quote
     that isn't even from this verse. A failing quote is FLAGGED only if it clearly
     IS this verse (high token coverage) but deviates; a low-coverage failure is a
     different source near the link -> skipped, not flagged.

NOT a pre-commit hook yet — run it manually; promote once it keeps earning weight
(same posture as verify-quotes). Exit 0 if clean, 1 if any FLAG.

Usage:
  python scripts/study-lint/scripture_verbatim.py <file.md> ...
  find study -name '*.md' | grep -v .audit | xargs python scripts/study-lint/scripture_verbatim.py
"""
import sys, os, re
from difflib import SequenceMatcher

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "quoter"))
import resolver, sources, grammar

ACCEPTED = os.path.join(os.path.dirname(os.path.abspath(__file__)), "accepted.tsv")

def load_accepted(path=ACCEPTED):
    """Reviewed-and-accepted flags (an unmarked trim that's fine, or a deferred
    carry-forward) so they stop surfacing. TSV: relpath <TAB> ref <TAB> quote-prefix.
    The *why* is recorded in study/.audit/scripture-verbatim-carryforward.md."""
    acc = []
    if os.path.exists(path):
        for line in open(path, encoding="utf-8"):
            if not line.strip() or line.lstrip().startswith("#"):
                continue
            parts = line.rstrip("\n").split("\t")
            if len(parts) >= 3:
                acc.append((parts[0].replace("\\", "/").strip(), parts[1].strip(), parts[2].strip()))
    return acc

def is_accepted(relpath, ref, quote, acc):
    rp = relpath.replace("\\", "/")
    return any(a_path == rp and a_ref == ref and quote.startswith(a_pre)
               for a_path, a_ref, a_pre in acc)

LINK = re.compile(r'\[([^\]]+)\]\(([^)]+?\.md)\)')
QUOTE = re.compile(r'"([^"]{4,}?)"|“([^”]{4,}?)”')
WINDOW = 320          # max chars between a quote's end and its link
COV_FLAG = 0.90       # gate floor: almost all words present in order
ANCHOR_FLAG = 0.85    # a long single contiguous run = one boundary edit
MIN_TOK = 4           # ignore quotes shorter than this many tokens
_MARKS = re.compile(r"\[[^\]]*\]|\.\.\.|…")

def ref_from_path(path):
    """Fallback when a link label doesn't parse: derive a whole-chapter ref from
    the gospel-library path itself."""
    m = re.search(r"scriptures/([^/]+)/([^/]+)/(\d+)\.md", path.replace("\\", "/"))
    if not m:
        return None
    vol, folder, chap = m.group(1), m.group(2), int(m.group(3))
    label = resolver._LABEL.get(folder, folder)
    return {"label": f"{label} {chap}", "vol": vol, "folder": folder,
            "chap": chap, "v1": None, "v2": None}

def parse_link(label, path):
    """(parsed_ref, is_scripture). Prefer the label (gives verses); fall back to
    the path (whole chapter)."""
    try:
        return resolver.parse_ref(label), True
    except resolver.RefError:
        if "gospel-library/eng/scriptures/" in path.replace("\\", "/"):
            p = ref_from_path(path)
            return p, p is not None
        return None, False

_VNUM = re.compile(r"(?<!\d)\b\d{1,3}\.(?=\s)")   # "5." verse markers inside a quote

def pre(q):
    """Strip markdown emphasis and embedded verse-number markers ("5." in a block
    quote of a whole verse — the source has them stripped, so they read as spurious
    deviations). KEEPS marked edits ([..], ...) so grammar.verify can parse them."""
    return _VNUM.sub(" ", q.replace("*", "").replace("_", ""))

def bare(q):
    """pre() plus the marked edits removed — the quote's verbatim-claimed tokens,
    for scoring."""
    return _MARKS.sub(" ", pre(q))

def scores(quote, verse_text):
    """Contiguous-run signals on the quote's verbatim-claimed tokens vs the verse.
    A real near-miss has FEW deviations, so it survives as a few LONG contiguous
    runs; a label / paraphrase / quote-of-another-verse only scatter-matches common
    words (many short runs). Returns (coverage, sorted_run_lengths, n_tokens)."""
    qt = grammar.tokens(bare(quote))
    vt = grammar.tokens(verse_text)
    if not qt:
        return 0.0, [], 0
    sm = SequenceMatcher(None, qt, vt, autojunk=False)
    blocks = sorted((b.size for b in sm.get_matching_blocks() if b.size), reverse=True)
    return sum(blocks) / len(qt), blocks, len(qt)

def is_near_miss(cov, blocks, ntok):
    """Flag iff the quote is clearly this verse but deviates — and precisely so:
      (a) a long single anchor (>=85%): one edit at a boundary, the body intact; or
      (b) two genuinely LONG runs (each >=4 tokens) covering >=90% at cov >=0.95:
          one deviation in the MIDDLE, between two substantial verbatim chunks.
    Short runs (a 2+2 label) and fragmented condensations fail both -> not flagged."""
    if ntok < MIN_TOK or cov < COV_FLAG or not blocks:
        return False
    anchor = blocks[0] / ntok
    if anchor >= ANCHOR_FLAG:
        return True
    # one MIDDLE deviation: two LONG runs (>=4 tokens each) covering >=90%. A drop
    # keeps cov ~1.0; a swap adds one wrong token (cov ~0.93) — both qualify here, so
    # this catches mid-quote swaps the single anchor misses. Three+ short runs (a
    # condensation) or a 2+2 label fail the run-length / top2 bar.
    if (len(blocks) >= 2 and blocks[0] >= 4 and blocks[1] >= 4
            and (blocks[0] + blocks[1]) / ntok >= 0.90):
        return True
    return False

def events(text):
    ev = []
    for m in QUOTE.finditer(text):
        ev.append((m.start(), m.end(), "quote", (m.group(1) or m.group(2))))
    for m in LINK.finditer(text):
        ev.append((m.start(), m.end(), "link", (m.group(1), m.group(2))))
    ev.sort()
    return ev

def check_file(path, accepted=()):
    text = open(path, encoding="utf-8").read()
    results = []      # (level, ref_label, msg)
    n_ok = 0
    pending = None    # (end_pos, quote_text)
    for start, end, kind, payload in events(text):
        if kind == "quote":
            pending = (end, payload)
            continue
        label, lpath = payload
        p, is_scr = parse_link(label, lpath)
        if not is_scr:
            pending = None
            continue
        if not pending or (start - pending[0]) > WINDOW:
            pending = None
            continue                     # a bare reference, no quote to verify
        quote = pending[1]
        pending = None
        try:
            _, verse = sources.verses(p)
        except (FileNotFoundError, KeyError) as e:
            results.append(("FLAG", p["label"], f"cannot verify — broken link / missing verse: {e}"))
            continue
        ok, spans = grammar.verify(pre(quote), verse)
        if ok:
            n_ok += 1
            continue
        cov, blocks, ntok = scores(quote, verse)
        if is_near_miss(cov, blocks, ntok) and not is_accepted(path, p["label"], quote, accepted):
            anc = blocks[0] / ntok
            results.append(("FLAG", p["label"],
                            f'near-verbatim (anchor {anc:.2f}, cov {cov:.2f}) — '
                            f'unmarked drop/swap/insert? review: "{quote[:66]}"'))
        # else: no long contiguous anchor -> a label, a paraphrase, or a quote of a
        # different verse near this link; not a verbatim-claim of it -> skip.
    return n_ok, results

def main(argv):
    files = argv[1:] or ([l.strip() for l in sys.stdin if l.strip()] if not sys.stdin.isatty() else [])
    files = [f for f in files if f.endswith(".md") and os.path.isfile(f)]
    if not files:
        print("usage: scripture_verbatim.py <file.md> ...", file=sys.stderr)
        return 2
    accepted = load_accepted()
    bad = 0
    for path in files:
        try:
            n_ok, results = check_file(path, accepted)
        except Exception as e:
            print(f"{path}: ERROR {e}", file=sys.stderr)
            continue
        if not n_ok and not results:
            continue
        head = f"{path}  ({n_ok} scripture OK" + (f", {len(results)} to review)" if results else ")")
        print(("\n" + head) if results else head)
        for level, ref, msg in results:
            print(f"  [{level}] {ref}: {msg}")
            bad += 1
    print(f"\n— {len(files)} file(s); {bad} scripture flag(s).")
    return 1 if bad else 0

if __name__ == "__main__":
    sys.exit(main(sys.argv))
