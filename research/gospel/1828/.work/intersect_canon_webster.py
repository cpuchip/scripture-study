"""
Phase 4 — intersect the canon word list with the local 1828 dictionary.

For every unique canon word (from P3), look up all 1828 entries (same word
may have multiple entries for different parts of speech). Score each word
by heuristic signals that suggest meaning-shift or theological depth:

- multi-sense: 3+ distinct entries / senses across POS
- archaic markers: "obsolete", "formerly", "in old English", "anciently"
- theological depth: scripture / divine / soul / spirit / heaven / etc.
- legal-classical depth: jurisprudence / philosophy markers
- length: longer aggregate definition = more nuance

Output:
- 04-canon-words-with-1828-entries.json — full intersect data
- 04-canon-words-with-1828-entries.md — full intersect summary
- 04-flagged-candidates.md — top-scored highlight candidates (shortlist)

No MCP / LM Studio load: direct file read of webster1828.json.gz.

Provenance: 2026-05-20 autonomous overnight task.
"""
from __future__ import annotations
import gzip
import json
import re
import sys
from collections import defaultdict
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
WEBSTER = REPO / "scripts/webster-mcp/data/webster1828.json.gz"
CANON_JSON = REPO / "research/gospel/1828/03-canonical-unique-words.json"

OUT_JSON = REPO / "research/gospel/1828/04-canon-words-with-1828-entries.json"
OUT_FULL_MD = REPO / "research/gospel/1828/04-canon-words-with-1828-entries.md"
OUT_FLAG_MD = REPO / "research/gospel/1828/04-flagged-candidates.md"

# Signal patterns
ARCHAIC_RE = re.compile(
    r"\b(obsolete|formerly|anciently|in old English|in ancient|archaic|"
    r"out of use|deprecated|seldom used)\b",
    re.IGNORECASE,
)
THEOLOGICAL_RE = re.compile(
    r"\b(scripture|scriptures|scriptural|divine|deity|deities|heaven|heavenly|"
    r"spirit|spiritual|soul|holy|sacred|sanctif|gospel|christ|christian|"
    r"god|godly|godhead|providence|sin|sins|repent|covenant|salvation|"
    r"redeem|redemption|grace|sanctuary|temple|priest|prophet|apostle|"
    r"pharisee|sabbath|tabernacle|atonement|virtue|wisdom|grace|charity|"
    r"piety|reverence|truth)\b",
    re.IGNORECASE,
)
CLASSICAL_RE = re.compile(
    r"\b(jurispr|philosoph|logic|grammar|rhetoric|metaphysics|alchemy|"
    r"astrology|astronomy|natural philosophy|ethic|moral|virtue)\b",
    re.IGNORECASE,
)


def load_webster() -> dict[str, list[dict]]:
    with gzip.open(WEBSTER, "rt", encoding="utf-8") as f:
        entries = json.load(f)
    by_word: dict[str, list[dict]] = defaultdict(list)
    for e in entries:
        w = e.get("word", "")
        if w:
            by_word[w.lower()].append(e)
    return dict(by_word)


def score_word(word: str, entries: list[dict]) -> dict:
    """Score a word by 1828 entry signals. Returns dict with score + reasons."""
    n_entries = len(entries)
    n_senses = sum(len(e.get("definitions", [])) for e in entries)
    all_def_text = " ".join(d for e in entries for d in e.get("definitions", []))
    pos_set = sorted({e.get("pos", "") for e in entries if e.get("pos")})

    reasons = []
    score = 0

    if n_senses >= 5:
        reasons.append(f"{n_senses} senses")
        score += 2
    elif n_senses >= 3:
        reasons.append(f"{n_senses} senses")
        score += 1

    if n_entries >= 2:
        reasons.append(f"{n_entries} entries across POS ({','.join(pos_set)})")
        score += 1

    if ARCHAIC_RE.search(all_def_text):
        reasons.append("archaic marker")
        score += 3

    theo_hits = len(THEOLOGICAL_RE.findall(all_def_text))
    if theo_hits >= 3:
        reasons.append(f"{theo_hits} theological terms")
        score += 2
    elif theo_hits >= 1:
        reasons.append(f"{theo_hits} theological term")
        score += 1

    if CLASSICAL_RE.search(all_def_text):
        reasons.append("classical/philosophical")
        score += 1

    if len(all_def_text) >= 800:
        reasons.append(f"{len(all_def_text)} chars of definition")
        score += 1

    return {
        "n_entries": n_entries,
        "n_senses": n_senses,
        "pos": pos_set,
        "def_chars": len(all_def_text),
        "score": score,
        "reasons": reasons,
        "first_def": (entries[0].get("definitions", [""])[0] or "")[:300] if entries else "",
    }


def main() -> int:
    print("Loading Webster 1828…", file=sys.stderr)
    webster = load_webster()
    print(f"  {len(webster):,} unique words in 1828", file=sys.stderr)

    print("Loading canon word list…", file=sys.stderr)
    canon = json.loads(CANON_JSON.read_text(encoding="utf-8"))
    canon_words = set(canon["words"].keys())
    print(f"  {len(canon_words):,} unique words in canon", file=sys.stderr)

    intersect = sorted(canon_words & set(webster.keys()))
    print(f"  intersect: {len(intersect):,}", file=sys.stderr)

    records = []
    for w in intersect:
        s = score_word(w, webster[w])
        s["word"] = w
        s["canon_total"] = canon["words"][w]["total"]
        s["canon_by_volume"] = canon["words"][w]["by_volume"]
        records.append(s)

    # Sort by score desc, then by canon frequency desc
    records.sort(key=lambda r: (-r["score"], -r["canon_total"], r["word"]))

    # Full JSON
    OUT_JSON.write_text(json.dumps({
        "provenance": {
            "date": "2026-05-20",
            "canon_source": "research/gospel/1828/03-canonical-unique-words.json",
            "webster_source": "scripts/webster-mcp/data/webster1828.json.gz (direct read, no MCP)",
            "scoring": "multi-sense + archaic + theological + classical + length signals",
        },
        "n_canon": len(canon_words),
        "n_webster": len(webster),
        "n_intersect": len(intersect),
        "records": records,
    }, indent=2), encoding="utf-8")

    # Full markdown
    lines = []
    lines.append("# Canon ∩ Webster 1828 — full intersect")
    lines.append("")
    lines.append("**Source:**")
    lines.append(f"- Canon words: [`03-canonical-unique-words.json`](03-canonical-unique-words.json) — {len(canon_words):,} unique tokens from BoM/D&C/PGP/NT/OT")
    lines.append(f"- 1828 dictionary: `scripts/webster-mcp/data/webster1828.json.gz` ({len(webster):,} unique headwords; direct local read, no MCP)")
    lines.append(f"- **Intersect:** **{len(intersect):,} words** appear in both")
    lines.append("**Date:** 2026-05-20")
    lines.append("**Provenance:** generated by `research/gospel/1828/.work/intersect_canon_webster.py`")
    lines.append("")
    lines.append("## Scoring rubric")
    lines.append("")
    lines.append("Each canon-∩-1828 word scored by signals that suggest meaning-shift or depth:")
    lines.append("- **+2** if 5+ senses across all POS entries; **+1** if 3-4 senses")
    lines.append("- **+1** if 2+ entries (e.g. noun-and-verb forms)")
    lines.append("- **+3** if 1828 definition contains an archaic marker (\"obsolete\", \"formerly\", \"anciently\")")
    lines.append("- **+2** if 3+ theological terms in definitions (3+ scripture/divine/soul/etc.); **+1** if 1-2")
    lines.append("- **+1** if classical/philosophical vocabulary (jurisprudence, philosophy, ethics)")
    lines.append("- **+1** if total definition length ≥ 800 chars (substantial nuance)")
    lines.append("")
    lines.append(f"## All intersect words (top 200 by score)")
    lines.append("")
    lines.append("| Word | Score | Senses | POS | Canon freq | First sense (truncated) |")
    lines.append("|------|-------|--------|-----|------------|------------------------|")
    for r in records[:200]:
        pos = ",".join(r["pos"][:3])
        first = r["first_def"].replace("\n", " ").replace("|", "\\|")[:120]
        lines.append(f"| **{r['word']}** | {r['score']} | {r['n_senses']} | {pos} | {r['canon_total']} | {first} |")
    lines.append("")
    lines.append(f"*(Full {len(intersect):,}-entry intersect available in `04-canon-words-with-1828-entries.json`. Top 200 shown here for readability.)*")
    OUT_FULL_MD.write_text("\n".join(lines), encoding="utf-8")

    # Flagged candidates — highest score buckets
    high = [r for r in records if r["score"] >= 6]
    mid = [r for r in records if 4 <= r["score"] < 6]
    low_with_archaic = [
        r for r in records
        if r["score"] < 4 and "archaic marker" in r["reasons"]
    ]

    lines = []
    lines.append("# 1828 highlight candidates — heuristic shortlist")
    lines.append("")
    lines.append("**Source:** [`04-canon-words-with-1828-entries.json`](04-canon-words-with-1828-entries.json) — the canon ∩ 1828 intersect with per-word scoring.")
    lines.append("**Date:** 2026-05-20")
    lines.append("**Provenance:** generated by `research/gospel/1828/.work/intersect_canon_webster.py`")
    lines.append("")
    lines.append("Three signal buckets, ordered by highlight-worthiness:")
    lines.append("")
    lines.append("- **High signal (score ≥ 6)** — multiple meaning-shift indicators converge. Worth highlighting in the MVP without further review.")
    lines.append("- **Mid signal (score 4-5)** — at least one strong signal. Manual review before highlighting.")
    lines.append("- **Archaic-marker-only** — lower aggregate score but the 1828 entry explicitly marks the word's modern sense as different from its 1828 sense. These are pure semantic-drift words.")
    lines.append("")

    for label, recs in [
        (f"## High signal — score ≥ 6 ({len(high)} words)", high),
        (f"## Mid signal — score 4-5 ({len(mid)} words)", mid),
        (f"## Archaic-marker — even at low overall score ({len(low_with_archaic)} words)", low_with_archaic),
    ]:
        lines.append(label)
        lines.append("")
        if not recs:
            lines.append("(none)")
            lines.append("")
            continue
        lines.append("| Word | Score | Reasons | Canon freq | First sense |")
        lines.append("|------|-------|---------|------------|-------------|")
        for r in recs[:150]:
            reasons = "; ".join(r["reasons"])[:120]
            first = r["first_def"].replace("\n", " ").replace("|", "\\|")[:140]
            lines.append(f"| **{r['word']}** | {r['score']} | {reasons} | {r['canon_total']} | {first} |")
        if len(recs) > 150:
            lines.append(f"| *(+{len(recs)-150} more in JSON)* | | | | |")
        lines.append("")

    OUT_FLAG_MD.write_text("\n".join(lines), encoding="utf-8")

    print(f"\nWrote {OUT_JSON.relative_to(REPO)}", file=sys.stderr)
    print(f"Wrote {OUT_FULL_MD.relative_to(REPO)}", file=sys.stderr)
    print(f"Wrote {OUT_FLAG_MD.relative_to(REPO)}", file=sys.stderr)
    print(f"\nHigh signal: {len(high)}")
    print(f"Mid signal: {len(mid)}")
    print(f"Archaic-marker-only: {len(low_with_archaic)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
