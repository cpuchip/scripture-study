#!/usr/bin/env python3
"""voice_lint — the AI-tic detector.

The mechanical half of the writing-voice rules (CLAUDE.md § Writing Voice,
.mind/principles.md § Voice Discipline) made into a deterministic check, so the
agent's drift gets caught before Michael has to police "a common opus phrasing for
us" by hand. Precision over recall, same posture as the rest of study-lint: a flag
should be a real violation, not a borderline judgment call.

Catches (v1, high-precision):
  - EM-DASH BUDGET   : >1 prose em-dash in one paragraph (citation/attribution
                       dashes excluded). Rule: "one per paragraph max."
  - CUT-LIST         : the presenter tics on the explicit cut list.
  - META-NARRATION   : narrating the document's own structure.

Known recall gaps (v1, by design — the human + the voice-audit cover these):
  - therefore/but vs "and then" transitions (semantic; noisy to detect)
  - the "isn't just X — it's Y" formula at section cadence (per-doc judgment)
  - closing-refrain restatement (semantic)

Usage:
  python scripts/study-lint/voice_lint.py study/foo.md ...
  find study -name '*.md' | grep -vE '.audit|.scratch' | xargs python scripts/study-lint/voice_lint.py

Exit 0 clean, 1 on any flag. Accepted-flags ignore list: voice_accepted.tsv
(relpath <TAB> rule <TAB> first ~40 chars), '#' comments allowed.
"""
import os
import re
import sys

# ── tunable thresholds (single lines, like the rest of the suite) ──────────────
EMDASH_PARA_MAX = 1            # prose em-dashes allowed per paragraph

CUT_LIST = [
    "let that land", "sit with that", "here's the thing", "here is the thing",
    "this matters because", "read that again", "that's not nothing",
    "that is not nothing", "that changes everything", "stops me cold",
]

META_NARRATION = [
    r"what i notice:",
    r"\bi want to name\b",
    r"there is a specific point",
    r"section [\dIVXLC]+ is (the )?(answer|key|crux|heart|point)",
    r"\bthe (answer|point) (here )?is (this|the following)\b",
    r"^what (this|the) (section|study|paragraph) (does|is)",
]

EMDASH = "—"  # —

# An em-dash is a CITATION dash (excluded from the budget) when it opens an
# attribution: "— [Source](...)", "— Author (1990)", or a line that is just "— X".
CITATION_DASH = re.compile(
    r"—\s*(\[[^\]]+\]\(|[A-Z][^.\n]{0,60}\(\d{4}\)|[\"“])"
)
ATTRIB_LINE = re.compile(r"^\s*>?\s*—\s")  # "> — Name" / "— Name" attribution line
TABLE_ROW = re.compile(r"^\s*\|.*\|")       # markdown table row — not prose
LIST_ITEM = re.compile(r"^\s*([-*+]|\d+\.)\s")  # bullet/numbered list item


def load_accepted(root):
    path = os.path.join(root, "voice_accepted.tsv")
    acc = set()
    if os.path.exists(path):
        for line in open(path, encoding="utf-8"):
            line = line.rstrip("\n")
            if not line or line.startswith("#"):
                continue
            parts = line.split("\t")
            if len(parts) >= 3:
                acc.add((parts[0].strip(), parts[1].strip(), parts[2].strip()))
    return acc


def strip_noncontent(text):
    """Drop YAML frontmatter and fenced code blocks (not prose)."""
    lines = text.split("\n")
    out = []
    in_code = False
    in_front = False
    if lines and lines[0].strip() == "---":
        in_front = True
        lines = lines[1:]
        # consume to closing ---
        cut = 0
        for i, ln in enumerate(lines):
            if ln.strip() == "---":
                cut = i + 1
                break
        lines = lines[cut:]
    for ln in lines:
        if ln.lstrip().startswith("```"):
            in_code = not in_code
            out.append("")  # paragraph break, keep line numbers loose
            continue
        out.append("" if in_code else ln)
    return out


def _line_emdashes(line):
    """Em-dashes on a line that are NOT citation/attribution dashes."""
    if ATTRIB_LINE.match(line):
        return 0
    return CITATION_DASH.sub("", line).count(EMDASH)


def emdash_units(lines):
    """Yield (lineno, count, snippet) for each PROSE unit's em-dash total.

    The budget is per prose paragraph. Markdown table rows are skipped (not
    prose). Each list item is its own unit (a list of definition dashes is not
    one over-dashed paragraph). A prose unit is a run of consecutive flowing
    (non-list, non-table, non-blank) lines."""
    buf, start = [], None

    def flush():
        nonlocal buf, start
        if buf:
            n = sum(_line_emdashes(l) for l in buf)
            snip = re.sub(r"\s+", " ", " ".join(buf))[:40]
            res = (start, n, snip)
            buf, start = [], None
            return res
        buf, start = [], None
        return None

    for i, ln in enumerate(lines, 1):
        if ln.strip() == "" or TABLE_ROW.match(ln):
            r = flush()
            if r:
                yield r
            continue
        if LIST_ITEM.match(ln):
            r = flush()
            if r:
                yield r
            n = _line_emdashes(ln)
            yield (i, n, re.sub(r"\s+", " ", ln)[:40])
            continue
        if start is None:
            start = i
        buf.append(ln)
    r = flush()
    if r:
        yield r


def lint_file(path, accepted):
    try:
        text = open(path, encoding="utf-8").read()
    except Exception as e:
        return [(0, "READ", str(e))]
    rel = path.replace("\\", "/")
    hard, advisory = [], []   # hard → exit 1; advisory → surfaced, exit 0
    lines = strip_noncontent(text)

    # ADVISORY: em-dash budget. A judgment call by nature — even approved
    # baseline studies run ~2/unit — so it is surfaced, not gated. Tighten the
    # prose or amend the rule; the detector does not decide that for you.
    for start, n, snip in emdash_units(lines):
        if n > EMDASH_PARA_MAX:
            if (rel, "EMDASH", snip) not in accepted:
                advisory.append((start, "EMDASH", f"{n} prose em-dashes in one unit (rule: max {EMDASH_PARA_MAX}): \"{snip}…\""))

    # HARD: cut-list + meta-narration. The exact phrases are unambiguous tics.
    low_lines = [ln.lower() for ln in lines]
    for i, low in enumerate(low_lines, 1):
        for phrase in CUT_LIST:
            if phrase in low:
                snip = re.sub(r"\s+", " ", lines[i - 1])[:40]
                if (rel, "CUTLIST", snip) not in accepted:
                    hard.append((i, "CUTLIST", f"cut-list tic: \"{phrase}\""))
        for pat in META_NARRATION:
            if re.search(pat, low):
                snip = re.sub(r"\s+", " ", lines[i - 1])[:40]
                if (rel, "META", snip) not in accepted:
                    hard.append((i, "META", f"meta-narration: \"{snip}…\""))
    return hard, advisory


def main(argv):
    root = os.path.dirname(os.path.abspath(__file__))
    accepted = load_accepted(root)
    files = [a for a in argv if a.endswith(".md")]
    hard_total, adv_total = 0, 0
    for path in files:
        hard, advisory = lint_file(path, accepted)
        if hard or advisory:
            print(f"\n{path}")
            for lineno, rule, msg in sorted(hard):
                print(f"  {lineno:>5}  [{rule}] {msg}")
            for lineno, rule, msg in sorted(advisory):
                print(f"  {lineno:>5}  [{rule}~adv] {msg}")
            hard_total += len(hard)
            adv_total += len(advisory)
    print(f"\nvoice_lint: {hard_total} hard flag(s), {adv_total} advisory, across {len(files)} file(s).")
    return 1 if hard_total else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
