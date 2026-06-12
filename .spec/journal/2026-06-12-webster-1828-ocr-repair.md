# 2026-06-12 — Webster 1828 OCR Repair + the Study-Walk Plan

**Session type:** dev (webster-1828 lane)
**Binding question:** what OCR damage remains in the genuine 1828 data, and can
it be repaired honestly — without "fixing" toward modern expectations?

## What happened

Michael: "the new dictionary is paying off… track down the OCR issues and see
if we can repair it; maybe pull from another online 1828 if needed" + "a plan
to go through all the studies, in order, for correctness."

### Detection (scan_1828.py)

Validated all 6,031 scripture refs against the KJV canon (strongs-concordance
verse data — the MCP servers compounding). Found: 528 invalid refs, 359 lost
sentence junctions, 175 fragment senses, 122 junk fragments, 9 doubled
phrases, 24 broken See-refs.

### The three-digitization forensics (the session's discovery)

Evaluated DataWar/1828-dictionary (1828.mshaffer.com, MIT): passes all
anachronism probes — genuinely 1828 — **but carries the same charity
duplication and the same FINE "keep" as our EGW text.** The site
(webstersdictionary1828.com) is partially independent (cleaner charity, has
GAIN's noun) but shares other damage (CARRICK-BEND = CARRIER's senses in both).
**All available transcriptions descend from a common OCR ancestor at many
points.** Consequences:

1. Where all three agree on odd text ("Thin; keep; smoothly sharp"), we CANNOT
   distinguish shared OCR error from faithful transcription of a printer's
   error. Left alone, logged in `data/known-issues.md` — the facsimile
   (archive.org page scans) is the only higher authority.
2. **NAUGHTY, PESTILENCE, BEFORETIME are absent from all three** = genuine
   1828 gaps, NOT parse dropouts. This corrects my 2026-06-09 finding
   ("upstream OCR dropouts, neighbors present") — the neighbors-present
   heuristic misled me. ALLEGE/ZINC/SORCERIES/HOSEN = 1828 spellings and
   inflections (ALLEDGE, ZINK…), confirmed by the site's own "[See Alledge]"
   stub.

### Repair (repair_1828.py — every change ledgered)

- **426/543 scripture refs fixed**: fuzzy book names (Wark→Mark,
  Vatthew→Matthew, Solomon→Song of Solomon), 1↔7 digit swaps, and — the nice
  part — **ambiguous candidates disambiguated by KJV verse-text overlap**
  ("Corinthians 8:12" → whichever epistle's verse matches the quoted context;
  "Acis 17:79" → Acts 11:19 by the same test). 117 unresolved stay flagged
  (obviously-broken beats plausibly-wrong).
- 352 junctions, 166 junk strips, 9 doubled phrases, 53 stubs dropped,
  7 entries restored from the site (overlay, provenance-tagged). PISTACITE
  NOT added — the site's page is empty too; the overlay judge initially
  grabbed the next word's text (caught in spot-check, removed).
- **Variant layer**: `variants1828.json` + labeled fallback chain in
  `webster_define` (exact → 1828 spelling → archaic stem): "allege" shows
  ALLEDGE with an honest banner; "sleepeth" → SLEEP; "naughty" → NAUGHT.
- After-scan: junctions 0, doubled 0, refs 528→117-flagged. GAIN's noun was
  never lost (June-9 alarm was a truncated probe — checked this time).

### Shipped

Root `1336c8ce` pushed (1828-illuminated grant; sibling-lane commits rode
along, courtesy note left in pg-ai-stewards inbox). 1828-illuminated seed
updated → fingerprint reseed fires on deploy. webster-mcp rebuilt + stdio
smoke-tested.

### The study-walk plan (awaiting ratify)

`.spec/proposals/study-correctness-walk.md`: 469 study files inventoried,
**76 with quoted Webster definitions** (33 top-level / 13 yt / 12 bom-walk).
Tiered: T1 full requote (three-glories discipline: does the argument survive
genuine text), T2 mention-check, T3 link validation everywhere + quote
verification sampled. Chronological order, bom-walk progress-file mechanics,
ARGUMENT-class findings gated to Michael. Four ratify questions at the end.

## The durable lesson

Digitizations that look independent often share an OCR ancestor — agreement
between them is weaker evidence than it appears, and disagreement with
expectation is not license to "fix." The repair only changed text where a
mechanical inconsistency (invalid ref, lost junction) or a cleaner witness
proved the damage; everything else went to known-issues for the facsimile.
