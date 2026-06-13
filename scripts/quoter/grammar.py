#!/usr/bin/env python3
"""grammar — the shared quote grammar (study-tooling spec, §"quote grammar").

A quote = verbatim spans + MARKED edits. Brackets [..] (insertion/substitution)
and ellipses ... / … are legitimate; the spans between them must be verbatim.
Any UNMARKED deviation (a dropped word, a swapped word, a paraphrase-in-quotes) is
an error. This is the exact judgment the 469-file walk made by hand — "unmarked
elision -> mark it", the Alma 22:18 dropped "and".

The quoter PRODUCES well-formed marked quotes (verify before emit). The linter will
ENFORCE the same grammar on hand-written quotes. One engine, both ends.

verify(quote, source_text) -> (ok, spans) where each span records whether its
verbatim words appear contiguously in the source, in order.
"""
import re

_EDIT = re.compile(r"(\[[^\]]*\]|\.\.\.|…)")

def parse_quote(q):
    """Split a quote into segments: ('span'|'bracket'|'ellipsis', text)."""
    segs, last = [], 0
    for m in _EDIT.finditer(q):
        if m.start() > last:
            segs.append(("span", q[last:m.start()]))
        tok = m.group(1)
        segs.append(("bracket" if tok.startswith("[") else "ellipsis", tok))
        last = m.end()
    if last < len(q):
        segs.append(("span", q[last:]))
    return segs

def tokens(s):
    """Normalized word tokens — lowercase, curly punctuation folded, dashes are
    separators, apostrophes kept inside words. Punctuation/case differences are
    intentionally invisible (free quotes re-case and re-punctuate at splices); a
    dropped or swapped WORD is not."""
    s = s.lower().replace("’", "'").replace("‘", "'")
    s = s.replace("“", '"').replace("”", '"')
    s = re.sub(r"[—–-]", " ", s)
    return re.findall(r"[a-z0-9]+(?:'[a-z0-9]+)*", s)

def _find(src, sub, start):
    if not sub:
        return start
    n, m = len(src), len(sub)
    for i in range(start, n - m + 1):
        if src[i:i + m] == sub:
            return i
    return -1

class Span:
    def __init__(self, text, ok, at):
        self.text, self.ok, self.at = text, ok, at
    def __repr__(self):
        return f"Span({self.text!r}, ok={self.ok})"

def verify(quote, source_text):
    """True iff every verbatim span of `quote` appears contiguously in
    `source_text`, in left-to-right order (gaps allowed only across a marked
    edit). Returns (ok, [Span...])."""
    src = tokens(source_text)
    pos, ok, spans = 0, True, []
    for kind, text in parse_quote(quote):
        if kind != "span":
            continue
        wq = tokens(text)
        if not wq:
            continue
        idx = _find(src, wq, pos)
        if idx < 0:
            ok = False
            spans.append(Span(text.strip(), False, -1))
        else:
            spans.append(Span(text.strip(), True, idx))
            pos = idx + len(wq)
    return ok, spans

def first_bad(spans):
    for s in spans:
        if not s.ok:
            return s
    return None

if __name__ == "__main__":
    import sys
    src = sys.argv[1]
    for q in sys.argv[2:]:
        ok, spans = verify(q, src)
        print(("OK  " if ok else "FAIL"), repr(q))
        for s in spans:
            print("      ", "ok " if s.ok else "BAD", repr(s.text))
