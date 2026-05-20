"""
Phase 5 — synthesis. Combine P2 (study-confirmed) + P4 (heuristic-scored)
into a tiered final highlight-candidate list.

Tiers:
- A++ definitive: study-confirmed Tier A AND P4 high signal (≥6).
- A+ very strong: study-confirmed Tier A OR (Tier B AND P4 high signal).
- B  strong: study-confirmed Tier B alone, OR (Tier C AND P4 high signal).
- C  worth review: study-confirmed Tier C alone (lensed once), OR P4 high
                   signal alone (heuristically interesting, never lensed).
- D  archaic-marker pool: P4 archaic-marker words not otherwise tiered.

Provenance: 2026-05-20 autonomous overnight task per Michael's directive.
"""
from __future__ import annotations
import json
import re
from collections import defaultdict, Counter
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
RAW_JSONL = REPO / "research/gospel/1828/.work/raw-citations.jsonl"
INTERSECT_JSON = REPO / "research/gospel/1828/04-canon-words-with-1828-entries.json"
OUT = REPO / "research/gospel/1828/00-FINAL-highlight-candidates.md"

# Same REJECT list as P2 to keep the synthesis consistent
REJECT = {
    "why", "how", "what", "when", "where", "who", "which",
    "both", "all", "any", "every", "some", "one", "two",
    "from", "into", "with", "through", "without", "across",
    "before", "after", "during", "until", "while", "since",
    "above", "below", "between", "among",
    "stops", "starts", "ends", "becomes", "become", "stays",
    "saith", "thereof", "whereof",
    "use", "take", "make", "give", "set", "put", "see", "know",
    "find", "feel", "look", "say", "want", "need", "tell", "ask",
    "man", "men", "woman", "women", "people", "person",
    "day", "days", "night", "year", "years", "time",
    "way", "ways", "thing", "things",
    "very", "just", "only", "still", "again", "even", "also",
    "more", "less", "most", "least", "much", "many", "few",
}


def main():
    # Load P2 data (recompute tiers from raw)
    rows = [json.loads(l) for l in RAW_JSONL.read_text(encoding="utf-8").splitlines() if l.strip()]
    by_word: dict[str, list[dict]] = defaultdict(list)
    for r in rows:
        if r["word"] and r["word"] not in REJECT:
            by_word[r["word"]].append(r)

    study_tier: dict[str, str] = {}
    study_data: dict[str, dict] = {}
    for word, cites in by_word.items():
        studies = sorted({c["study"] for c in cites})
        tag_count = Counter(c["tag"] for c in cites)
        has_differ = tag_count["differ"] > 0
        if len(studies) >= 3 and has_differ:
            tier = "A"
        elif len(studies) >= 2:
            tier = "B"
        else:
            tier = "C"
        study_tier[word] = tier
        study_data[word] = {
            "tier": tier,
            "n_studies": len(studies),
            "n_citations": len(cites),
            "studies": studies,
            "has_differ": has_differ,
        }

    # Load P4 data
    intersect = json.loads(INTERSECT_JSON.read_text(encoding="utf-8"))
    p4 = {r["word"]: r for r in intersect["records"]}

    # Synthesize: for every word in study_data OR in p4-high-score, compute final tier
    final_words: set[str] = set(study_data.keys()) | {
        w for w, r in p4.items() if r["score"] >= 6 or "archaic marker" in r["reasons"]
    }
    # Also keep all P4-high (score >= 6) even without study confirmation
    final_words |= {w for w, r in p4.items() if r["score"] >= 6}

    records = []
    for w in final_words:
        sd = study_data.get(w)
        pd = p4.get(w)
        study_t = sd["tier"] if sd else None
        p4_high = pd is not None and pd["score"] >= 6
        p4_mid = pd is not None and 4 <= pd["score"] < 6
        p4_archaic = pd is not None and "archaic marker" in pd["reasons"]
        in_canon = pd is not None

        # Compute final tier
        if study_t == "A" and p4_high:
            tier = "A++"
        elif study_t == "A":
            tier = "A+"
        elif study_t == "B" and p4_high:
            tier = "A+"
        elif study_t == "B":
            tier = "B"
        elif study_t == "C" and p4_high:
            tier = "B"
        elif study_t == "C":
            tier = "C"
        elif p4_high:
            tier = "C"
        elif p4_archaic:
            tier = "D"
        else:
            tier = "D"

        records.append({
            "word": w,
            "tier": tier,
            "study_tier": study_t,
            "study_studies": sd["studies"] if sd else [],
            "study_n_citations": sd["n_citations"] if sd else 0,
            "study_has_differ": sd["has_differ"] if sd else False,
            "p4_score": pd["score"] if pd else None,
            "p4_reasons": pd["reasons"] if pd else [],
            "p4_n_senses": pd["n_senses"] if pd else 0,
            "p4_first_def": pd["first_def"] if pd else "",
            "in_canon": in_canon,
            "canon_total": pd["canon_total"] if pd else 0,
        })

    tier_order = {"A++": 0, "A+": 1, "B": 2, "C": 3, "D": 4}
    records.sort(key=lambda r: (
        tier_order[r["tier"]],
        -(r["study_n_citations"] or 0),
        -(r["p4_score"] or 0),
        r["word"],
    ))

    counts = Counter(r["tier"] for r in records)

    lines = []
    lines.append("# 1828 highlight candidates — final synthesis")
    lines.append("")
    lines.append("**Output of:** the four-phase word-list groundwork for the [1828-illuminated scriptures proposal](../../../.spec/proposals/1828-illuminated-scriptures.md).")
    lines.append("")
    lines.append("**Date:** 2026-05-20 (autonomous overnight, per Michael's directive)")
    lines.append("**Provenance:** synthesizes [`02-confirmed-from-studies.md`](02-confirmed-from-studies.md) (P2) and [`04-flagged-candidates.md`](04-flagged-candidates.md) (P4). Generated by `research/gospel/1828/.work/compose_p5_final.py`.")
    lines.append("")
    lines.append("## How the tiers combine")
    lines.append("")
    lines.append("Two independent signals converge on each word:")
    lines.append("")
    lines.append("- **Study signal** (P2) — has our own substrate work *already lensed* this word via Webster 1828? Tier A=3+ studies + differ; B=2+ studies; C=1 study.")
    lines.append("- **Dictionary signal** (P4) — does the 1828 entry itself carry meaning-shift markers? high=score ≥6 (multi-sense + archaic + theological depth + length); mid=score 4-5; archaic-only=explicit \"obsolete/formerly\" marker.")
    lines.append("")
    lines.append("**Combined tier:**")
    lines.append("")
    lines.append("| Final | Definition |")
    lines.append("|-------|------------|")
    lines.append("| **A++** | Study Tier A AND P4 high. Both signals converge — these are the **definitive MVP highlight words**. |")
    lines.append("| **A+**  | Study Tier A alone, OR Study Tier B + P4 high. Strong; ship in MVP. |")
    lines.append("| **B**   | Study Tier B alone, OR Study Tier C + P4 high. Strong candidates pending one review pass. |")
    lines.append("| **C**   | Study Tier C alone, OR P4 high alone (never lensed in our studies). Worth the highlight tool's broader pool. |")
    lines.append("| **D**   | P4 archaic-marker words not otherwise tiered. Pure semantic-drift candidates; review individually. |")
    lines.append("")
    lines.append(f"**Counts:** A++={counts['A++']} · A+={counts['A+']} · B={counts['B']} · C={counts['C']} · D={counts['D']} · total={len(records)}")
    lines.append("")
    lines.append("---")
    lines.append("")

    for tier in ["A++", "A+", "B", "C", "D"]:
        recs = [r for r in records if r["tier"] == tier]
        lines.append(f"## Tier {tier} ({len(recs)} words)")
        lines.append("")
        if not recs:
            lines.append("(none)")
            lines.append("")
            continue

        lines.append("| Word | Study tier | Studies | Diff? | P4 score | P4 reasons | 1828 first sense |")
        lines.append("|------|------------|---------|-------|----------|------------|------------------|")
        for r in recs:
            st = r["study_tier"] or "—"
            n_st = r["study_n_citations"]
            differ = "✓" if r["study_has_differ"] else ""
            p4s = r["p4_score"] if r["p4_score"] is not None else "—"
            reasons = "; ".join(r["p4_reasons"])[:100]
            first = (r["p4_first_def"] or "").replace("\n", " ").replace("|", "\\|")[:120]
            lines.append(f"| **{r['word']}** | {st} | {n_st} | {differ} | {p4s} | {reasons} | {first} |")
        lines.append("")

    # Provenance + how-to-use footer
    lines.append("---")
    lines.append("")
    lines.append("## How to use this list (for the MVP frontend)")
    lines.append("")
    lines.append("Recommended highlight strategy for the first 1828.ibeco.me cut:")
    lines.append("")
    lines.append("1. **Highlight Tier A++ and A+ words inline by default.** These have converging signals (substrate work AND dictionary depth). Hover or click reveals the 1828 definition.")
    lines.append("2. **Tier B opt-in: \"deep mode\".** A toggle that adds these as additional highlights for users who want richer coverage.")
    lines.append("3. **Tier C + D available on user-initiated lookup.** Don't highlight inline — too much visual noise — but any word the user clicks should fall back to 1828 lookup if a entry exists.")
    lines.append("4. **Cross-link to study work for A++/A+.** Words lensed in our existing studies should link from the hover-card to those studies. Highest-leverage feature: \"see how we used this word\".")
    lines.append("5. **Honest cautions** (from proposal §IV) apply throughout: 1828 isn't always deeper, don't encode mysteries, good-faith reads still differ.")
    lines.append("")
    lines.append("## Carry-forward")
    lines.append("")
    lines.append("- Tier C and D need manual review before any are added to the inline-highlight set.")
    lines.append("- The Webster 1828 entries themselves are not exposed in this list (only first-sense excerpts). The hover-card UI should pull from `scripts/webster-mcp/data/webster1828.json.gz` directly or via webster-mcp's `webster_define` tool.")
    lines.append("- Modern-definition diffing (D-1828-4 in proposal) was deferred — would need webster-mcp's `modern_define` on a few thousand words. Carry forward as a hardening pass once the MVP is live and we know which words actually need the modern comparison.")
    lines.append("- The full study citation table is in [`01-webster-references-in-studies.md`](01-webster-references-in-studies.md); the heuristic flagged shortlist is in [`04-flagged-candidates.md`](04-flagged-candidates.md). Both feed this synthesis.")

    OUT.write_text("\n".join(lines), encoding="utf-8")
    print(f"Wrote {OUT.relative_to(REPO)}")
    for tier in ["A++", "A+", "B", "C", "D"]:
        print(f"  Tier {tier}: {counts[tier]}")


if __name__ == "__main__":
    main()
