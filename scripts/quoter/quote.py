#!/usr/bin/env python3
"""quote — the constructor. Insert verbatim source text + a canonical, correctly
re-based link into a target file. The write-side dual of the linter (see
.spec/proposals/study-tooling.md). No retype: the words are the source's own.

  quote scripture <ref> [--phrase "..."] [--block] [--into FILE] [--dry-run]
  quote webster   <word> [--def N] [--phrase "..."] [--into FILE] [--link]
  quote promote   <ref>  --from SCRATCH [--into STUDY] [--block]

Feature ladder (grounded in how the studies actually quote — see the survey in
the spec):
  v1  block        whole verse / def as a set-piece blockquote
  v2  inline phrase a verbatim sub-phrase woven into prose: "phrase" ([Ref](link))
  v3  free-flow     marked edits in the phrase — [brackets] and ... ellipses —
                    verified span-by-span; refuses to emit anything non-verbatim
  promote          carry a verified quote scratch -> study, RE-BASING the link and
                   re-verifying against the canonical source (the pass-through)

The constructor will not emit a quote it cannot verify. A FAIL is a refusal, not a
warning — that is the whole point.
"""
import sys, os, re, argparse

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import resolver, sources, grammar

IBECO = "https://1828.ibeco.me/word/"

# ---- emit helpers ----------------------------------------------------------

def _detect_nl(path):
    if path and os.path.exists(path):
        with open(path, "rb") as f:
            return "\r\n" if b"\r\n" in f.read(8192) else "\n"
    return "\n"

def _append(path, block, nl):
    block = block.replace("\r\n", "\n").replace("\n", nl)
    existing = b""
    if os.path.exists(path):
        with open(path, "rb") as f:
            existing = f.read()
    nlb = nl.encode()
    if not existing:
        sep = ""
    elif existing.endswith(nlb + nlb):
        sep = ""            # already separated by a blank line
    elif existing.endswith(nlb):
        sep = nl            # one trailing newline -> add one more for a blank line
    else:
        sep = nl + nl       # no trailing newline -> add a blank line
    with open(path, "ab") as f:
        f.write((sep + block + nl).encode("utf-8"))

def emit(block, into, nl, dry):
    if into and not dry:
        _append(into, block, nl)
        print(f"  + appended to {into}")
    else:
        print(block)

def fmt_inline(text, link_md):
    return f'"{text}" ({link_md})'

def fmt_block(text, link_md):
    return f'> "{text}"\n>\n> — {link_md}'

# ---- scripture -------------------------------------------------------------

def cmd_scripture(a):
    p = resolver.parse_ref(a.ref)
    fp = resolver.file_path(p)
    if not os.path.exists(fp):
        sys.exit(f"FAIL: source file missing for {p['label']} ({fp})")
    label, verse = sources.verses(p)
    target = a.into or a.rel or "study/x.md"
    link_md = resolver.md_link(p, target)

    if a.phrase:
        ok, spans = grammar.verify(a.phrase, verse)
        if not ok:
            bad = grammar.first_bad(spans)
            sys.exit(f"FAIL: not verbatim in {label} — this span has no contiguous "
                     f"match (unmarked drop/swap/paraphrase?):\n  {bad.text!r}\n"
                     f"source: {verse}")
        text = a.phrase.strip()
    else:
        text = verse  # whole verse/range

    if a.inline:
        block = fmt_inline(text, link_md)
    elif a.block or not a.phrase:   # whole verse/range defaults to a set-piece block
        block = fmt_block(text, link_md)
    else:                           # a phrase defaults to inline (the dominant form)
        block = fmt_inline(text, link_md)
    emit(block, a.into, _detect_nl(a.into), a.dry_run)

# ---- webster ---------------------------------------------------------------

def cmd_webster(a):
    defs = sources.webster_defs(a.word, "1828")
    if not defs:
        sys.exit(f"FAIL: {a.word!r} is not an 1828 headword (ABSENT). "
                 f"Don't attribute a quote to Webster 1828 for a word it doesn't define.")
    word = a.word.strip().lower()
    wlabel = f"[{word}]({IBECO}{word})" if a.link else f"*{word}*"

    if a.phrase:
        # verify the phrase against ALL defs (any one may carry it)
        joined = " ".join(defs)
        ok, spans = grammar.verify(a.phrase, joined)
        if not ok:
            bad = grammar.first_bad(spans)
            sys.exit(f"FAIL: not verbatim in Webster 1828 {word!r} — span has no "
                     f"contiguous match:\n  {bad.text!r}")
        block = f'{wlabel}, Webster 1828: "{a.phrase.strip()}"'
    else:
        d = defs[a.defn - 1] if a.defn and a.defn <= len(defs) else defs[0]
        block = f'{wlabel}, Webster 1828: "{d.strip()}"'
    emit(block, a.into, _detect_nl(a.into), a.dry_run)

# ---- promote (scratch -> study, re-based + re-verified) --------------------

_QUOTED = re.compile(r'"([^"]{4,})"')

def cmd_promote(a):
    p = resolver.parse_ref(a.ref)
    chap_file = os.path.basename(resolver.file_path(p))   # e.g. 5.md
    if not os.path.exists(a.from_):
        sys.exit(f"FAIL: scratch file not found: {a.from_}")
    folder = f"{p['folder']}/{p['chap']}.md"
    found = None
    with open(a.from_, encoding="utf-8") as f:
        for line in f:
            if folder in line:                      # a link to this chapter
                m = _QUOTED.search(line)
                if m:
                    found = m.group(1).strip()
                    break
    if a.phrase:
        found = a.phrase.strip()
    if not found:
        sys.exit(f"FAIL: no quoted text for {p['label']} found in {a.from_} "
                 f"(looked for a line linking {folder} with a \"quote\").")

    # RE-VERIFY against the canonical source — promote never carries a corrupt quote
    _, verse = sources.verses(p)
    ok, spans = grammar.verify(found, verse)
    if not ok:
        bad = grammar.first_bad(spans)
        sys.exit(f"FAIL: the scratch quote for {p['label']} does not verify against "
                 f"the source (span {bad.text!r}). Fix the scratch first.")
    target = a.into or "study/x.md"
    link_md = resolver.md_link(p, target)
    block = fmt_block(found, link_md) if a.block else fmt_inline(found, link_md)
    print(f"  promoted {p['label']} (re-verified, link re-based for {target}):")
    emit(block, a.into, _detect_nl(a.into), a.dry_run)

# ---- cli -------------------------------------------------------------------

def main(argv):
    ap = argparse.ArgumentParser(prog="quote", description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)

    s = sub.add_parser("scripture", help="quote a verse/range")
    s.add_argument("ref"); s.add_argument("--phrase")
    s.add_argument("--block", action="store_true"); s.add_argument("--inline", action="store_true")
    s.add_argument("--into"); s.add_argument("--rel"); s.add_argument("--dry-run", action="store_true")
    s.set_defaults(func=cmd_scripture)

    w = sub.add_parser("webster", help="quote a genuine 1828 definition")
    w.add_argument("word"); w.add_argument("--def", dest="defn", type=int, default=1)
    w.add_argument("--phrase"); w.add_argument("--link", action="store_true")
    w.add_argument("--into"); w.add_argument("--dry-run", action="store_true")
    w.set_defaults(func=cmd_webster)

    pr = sub.add_parser("promote", help="carry a verified quote scratch -> study")
    pr.add_argument("ref"); pr.add_argument("--from", dest="from_", required=True)
    pr.add_argument("--into"); pr.add_argument("--phrase")
    pr.add_argument("--block", action="store_true"); pr.add_argument("--dry-run", action="store_true")
    pr.set_defaults(func=cmd_promote)

    a = ap.parse_args(argv[1:])
    a.func(a)

if __name__ == "__main__":
    main(sys.argv)
