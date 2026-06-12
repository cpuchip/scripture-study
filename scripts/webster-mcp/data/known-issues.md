# Webster 1828 data — known issues (post-repair, 2026-06-12)

State after `tools/repair_1828.py` (full ledger:
`.spec/scratch/webster-1828-repair-ledger.json`). What was fixed: 426
scripture references (fuzzy book names, 1↔7 digit swaps, KJV-verse-text
disambiguation), 352 lost sentence junctions, 166 junk fragments, 9 doubled
phrases, 53 pronunciation-stub senses dropped, 7 entries restored from
webstersdictionary1828.com, 1 entry added isn't — PISTACITE stayed out (see
below). Variant map `variants1828.json` covers 1828 spellings (ALLEGE→ALLEDGE,
ZINC→ZINK…) via labeled tool fallback.

## Left as-is, deliberately

1. **117 unresolved scripture refs** — damaged beyond confident automated
   repair (e.g. "Acis 26:173", verse numbers with extra digits). Listed under
   `unresolved_refs` in the ledger. They read as obviously-broken rather than
   plausibly-wrong, which is the safer failure mode. Candidates for the
   facsimile pass.
2. **115 sole-sense fragment entries** — entries whose only content is a
   pronunciation stub or "Obs." marker (ABLE's adjective block has real
   content; its stub sibling entry remains). webstersdictionary1828.com is
   equally thin for these. Facsimile pass.
3. **"Thin; keep; smoothly sharp" (FINE #3)** and similar suspected
   character-level errors: identical in all three digitizations (EGW text,
   webstersdictionary1828.com, DataWar/mshaffer) — they share an OCR ancestor.
   Without the 1828 facsimile we cannot distinguish "shared OCR error" from
   "faithful transcription of a printer's error." DO NOT correct toward modern
   expectation.
4. **CARRICK-BEND carries CARRIER's senses** — the same misalignment exists on
   webstersdictionary1828.com. Shared-ancestor damage; facsimile pass.
5. **PISTACITE has no definition text in any available transcription** (the
   site's page is empty too); the entry is omitted rather than fabricated.
6. **Genuine 1828 gaps (not data bugs):** NAUGHTY, PESTILENCE, BEFORETIME are
   absent from all three digitizations — Webster didn't enter them. The
   variant map routes NAUGHTY→NAUGHT; PESTILENT/PESTILENTIAL exist.

## The facsimile pass (future)

The only authority above all transcriptions is the 1828 printing itself
(archive.org page scans). Items 1–5 carry a bounded word list; a session with
the facsimile (reading page images) could settle every one. Until then the
three-digitization agreement is the best evidence available, and disagreement
with it is flagged, not silently fixed.
