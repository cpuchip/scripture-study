"""
Build the static data bundle the 1828-illuminated frontend ships.

Three outputs into frontend/src/data/:
  - tier-words.json    — tiered highlight list (A++, A+, B, C, D)
                          extracted from the P5 synthesis JSON
  - definitions-1828.json — for each tier word, the full 1828 entries
                            (extracted from scripts/webster-mcp/data/webster1828.json.gz)
  - studies.json       — for each tier word, the list of substrate study
                          files that lensed it (from P1 raw citations)

Modern definitions are added separately by fetch_modern_defs.py (rate-limited
to be friendly to the Free Dictionary API). This script can be re-run any
time; it overwrites the static data files.
"""
from __future__ import annotations
import gzip
import json
from collections import defaultdict
from pathlib import Path

REPO = Path("C:/Users/cpuch/Documents/code/stuffleberry/scripture-study")
PROJECT = REPO / "projects/1828-illuminated"
DATA = PROJECT / "frontend/src/data"

WEBSTER = REPO / "scripts/webster-mcp/data/webster1828.json.gz"
INTERSECT_JSON = REPO / "research/gospel/1828/04-canon-words-with-1828-entries.json"
RAW_CITES = REPO / "research/gospel/1828/.work/raw-citations.jsonl"


def main():
    DATA.mkdir(parents=True, exist_ok=True)

    # --- Load 1828 dictionary ---
    print("Loading Webster 1828…")
    with gzip.open(WEBSTER, "rt", encoding="utf-8") as f:
        webster_entries = json.load(f)
    webster_by_word: dict[str, list[dict]] = defaultdict(list)
    for e in webster_entries:
        w = (e.get("word") or "").lower()
        if w:
            webster_by_word[w].append({
                "pos": e.get("pos", ""),
                "definitions": e.get("definitions", []),
            })
    print(f"  {len(webster_by_word):,} unique 1828 headwords loaded")

    # --- Load P4 intersect data ---
    print("Loading P4 intersect…")
    p4 = json.loads(INTERSECT_JSON.read_text(encoding="utf-8"))
    p4_by_word = {r["word"]: r for r in p4["records"]}

    # --- Load P1 raw citations (study cross-references) ---
    print("Loading P1 study citations…")
    study_by_word: dict[str, list[dict]] = defaultdict(list)
    with RAW_CITES.open("r", encoding="utf-8") as f:
        for line in f:
            row = json.loads(line.strip())
            if row["word"]:
                study_by_word[row["word"]].append({
                    "study": row["study"],
                    "line": row["line"],
                    "tag": row["tag"],
                    "excerpt": row["definition_excerpt"][:300],
                })

    # --- Reconstruct tier mapping ---
    # Replicate P5 logic. Reject list (matches compose_p5_final.py).
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

    from collections import Counter
    word_records: list[dict] = []

    # Words known to studies
    by_word_studies: dict[str, list[dict]] = defaultdict(list)
    for word, rows in study_by_word.items():
        if word in REJECT:
            continue
        studies = sorted({r["study"] for r in rows})
        tag_count = Counter(r["tag"] for r in rows)
        has_differ = tag_count["differ"] > 0
        if len(studies) >= 3 and has_differ:
            study_t = "A"
        elif len(studies) >= 2:
            study_t = "B"
        else:
            study_t = "C"
        by_word_studies[word] = [{"study": s, "tag": tag_count} for s in studies]

        p4_r = p4_by_word.get(word)
        p4_high = p4_r is not None and p4_r["score"] >= 6
        p4_archaic = p4_r is not None and "archaic marker" in p4_r["reasons"]

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
        else:
            tier = "C"

        word_records.append({
            "word": word,
            "tier": tier,
            "study_tier": study_t,
            "studies": studies,
            "study_excerpts": [r["excerpt"] for r in rows[:3]],
            "p4_score": p4_r["score"] if p4_r else None,
            "p4_reasons": p4_r["reasons"] if p4_r else [],
        })

    # Words NOT in studies but P4-high — these are Tier C
    studied_set = {r["word"] for r in word_records}
    for word, p4_r in p4_by_word.items():
        if word in studied_set or word in REJECT:
            continue
        if p4_r["score"] >= 6:
            tier = "C"
            word_records.append({
                "word": word,
                "tier": tier,
                "study_tier": None,
                "studies": [],
                "study_excerpts": [],
                "p4_score": p4_r["score"],
                "p4_reasons": p4_r["reasons"],
            })
        elif "archaic marker" in p4_r["reasons"]:
            word_records.append({
                "word": word,
                "tier": "D",
                "study_tier": None,
                "studies": [],
                "study_excerpts": [],
                "p4_score": p4_r["score"],
                "p4_reasons": p4_r["reasons"],
            })

    tier_order = {"A++": 0, "A+": 1, "B": 2, "C": 3, "D": 4}
    word_records.sort(key=lambda r: (tier_order[r["tier"]], r["word"]))

    counts: Counter = Counter(r["tier"] for r in word_records)
    print(f"  Tier counts: {dict(counts)}")

    # --- Write tier-words.json (the master highlight list for the frontend) ---
    (DATA / "tier-words.json").write_text(json.dumps({
        "generated_at": "2026-05-20",
        "source": "research/gospel/1828/00-FINAL-highlight-candidates.md (regenerated from P1/P4 source data)",
        "tier_counts": dict(counts),
        "words": word_records,
    }, indent=2), encoding="utf-8")
    print(f"  Wrote tier-words.json ({len(word_records)} words)")

    # --- Write definitions-1828.json (full 1828 entries for every tier word) ---
    defs_1828 = {}
    missing = []
    for r in word_records:
        word = r["word"]
        if word in webster_by_word:
            defs_1828[word] = webster_by_word[word]
        else:
            missing.append(word)
    print(f"  1828 definitions: {len(defs_1828):,} present, {len(missing)} missing")

    (DATA / "definitions-1828.json").write_text(json.dumps({
        "generated_at": "2026-05-20",
        "source": "scripts/webster-mcp/data/webster1828.json.gz (direct read)",
        "missing_words": missing,
        "definitions": defs_1828,
    }, indent=2), encoding="utf-8")
    print(f"  Wrote definitions-1828.json")

    # --- Word list for the fetcher (just the words it needs to grab modern defs for) ---
    # Limit to non-D tiers (A++ + A+ + B + C) for the first fetch run.
    fetch_words = [r["word"] for r in word_records if r["tier"] != "D"]
    (PROJECT / "scripts" / "fetch-wordlist.txt").write_text(
        "\n".join(fetch_words) + "\n", encoding="utf-8",
    )
    print(f"  Wrote fetch-wordlist.txt ({len(fetch_words)} words to fetch modern defs for)")


if __name__ == "__main__":
    main()
