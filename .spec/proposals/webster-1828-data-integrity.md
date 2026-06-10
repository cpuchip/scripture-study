# Webster "1828" Data Integrity — Investigation & Remediation Brief

**Status: FORENSICS COMPLETE (steps 1–2 verified 2026-06-09, dedicated session) —
remediation steps 3–6 awaiting ratification.** Created 2026-06-09 by Claude Fable 5
during the scripture-book v4 honesty walk. Michael: *"I kinda feel lied to with the
source of the dictionary I got from git. It's something I want to resolve, and I'll
probably go through all of my published works (study / cpuchip.net / book) to fix
issues there."* This file is the self-sufficient handoff for that session.

---

## FORENSICS VERDICT (2026-06-09 dedicated session)

**1. The "1828" label was OURS, not upstream's.** ssvivian/WebstersDictionary's README
says only "Webster's Unabridged English Dictionary, provided by the Gutenberg Project"
— no year, file named `dictionary.json`. Our commit `46c0092` (2026-02-04) renamed it
`webster1828.json.gz` and wrote "1828" into the README and tool descriptions. The
original plan doc (`scripts/plans/04_tool-improvements.md` at that commit) noted the
source as "MIT License, Project Gutenberg source" and even listed
webstersdictionary1828.com as "Online reference for verification" — the verification
was planned and never executed. Nobody lied upstream; we assumed the edition because
the *goal* was 1828.

**2. The data is uniformly Webster 1913, confirmed on 10 words.** Anachronism probes:
the tool defines **telephone** (1876), **phonograph** (1877), **bacterium** (1838).
Text comparisons vs webstersdictionary1828.com: *spirit*, *intelligence* (from the
original audit), *charity*, *glory*, *virtue*, *suffer* — all 1913 text (the tool's
*suffer* quotes **Tennyson**, published 1855; *virtue* quotes De Quincey and Keble).
Zero 1828 matches. Note for the three-glories study audit: genuine 1828 *glory* has
12 senses beginning "Brightness; luster; splendor" — substantially different from the
1913 entry the study drew on.

**3. Replacement source FOUND and VERIFIED: `github.com/kayson-argyle/websters_1828`.**
SQLite DB (113,595 sense-rows) + raw text + Python parsing pipeline + 3,785 KJV
archaic-term lemma mappings. Provenance: the Ellen G. White Estate Archives' full-text
preservation of the actual 1828 (archive.org), built explicitly for KJV/LDS Standard
Works study. Passed the same standard: 5/5 anachronisms absent (telephone, phonograph,
bacterium, dinosaur, railroad), 6/6 sampled words match webstersdictionary1828.com
(charity, glory, virtue, suffer, spirit, intelligence). The two book-critical senses
are present verbatim: *spirit* 5 "intelligent, immaterial and immortal part of **human
beings**", 6 "An immaterial intelligent substance"; *intelligence* 4 "A spiritual
being; as a created intelligence."
  - **Caveats:** no declared license (underlying 1828 text is public domain; ask or
    re-parse from their raw text); OCR artifacts (junk headword rows like `- fall`,
    a doubled phrase in *charity* sense 1, `vur'�tu` encoding glitch, scripture refs
    with 1→7 OCR errors e.g. "7 Corinthians 8:1", "7 Peter 3:19"); headwords are
    lowercase; one row per sense (pos only on the first row of an entry).
  - **Rejected alternatives:** DataWar/1828-dictionary (README says its 1828 "came
    from Gutenberg" — Gutenberg has no 1828; same failure shape as ours);
    kljensen/websters (stardict conversion, provenance undocumented);
    CrossCrusaders/Websters1828API (no data provenance documented).
  - Eval clone at `%TEMP%\websters_1828_eval`.

---

## What we know (verified 2026-06-09)

**The workspace's "Webster 1828" tooling serves Webster *1913* (Revised Unabridged)
text under the 1828 label.** Caught during the book's v4 audit; verified three ways:

1. **Internal anachronisms.** The tool's *spirit* entry cites **"U. S. Disp."** (the
   United States Dispensatory, first published **1833**), quotes **N. P. Willis** (1830s
   travel writer), Keble, and defines "stannic chloride" — impossible content for an
   1828 text.
2. **Authoritative mismatch.** webstersdictionary1828.com (the standard faithful 1828
   transcription) lacks the quoted phrasing entirely. Genuine 1828 *spirit*: "5. The
   soul of man; the intelligent, immaterial and immortal part of **human beings**."
   "6. An immaterial intelligent substance." The tool's version — "an intelligence
   conceived of apart from any physical organization or embodiment; vital essence,
   force, or energy, as distinct from matter" / "…immortal part of **man**" — is
   **verbatim Webster 1913**.
3. **Second word confirmed.** Genuine 1828 *intelligence* = "1. Understanding; skill.
   2. Notice; information communicated… 3. Commerce of acquaintance… 4. **A spiritual
   being; as a created intelligence.**" The tool returns 1913's "the exercise of the
   understanding" / "the capacity to know or understand" instead.

**The likely chain (to be confirmed):** `scripts/webster-mcp/data/webster1828.json.gz`
← per `scripts/webster-mcp/README.md` line 83: **github.com/ssvivian/WebstersDictionary**
(MIT) — and README line 180 says the dictionary content is under the **Project
Gutenberg License**. Project Gutenberg's Webster etext is the **1913 Revised
Unabridged**, not 1828. So the data was probably honest *Webster-Unabridged-1913* that
got relabeled "1828" somewhere between that repo and our tooling. **Open question for
the investigation: did ssvivian's repo claim 1828, or did we assume it?** (Check their
README/commit history before assigning blame — we may have done the mislabeling
ourselves when we named the file `webster1828.json.gz`.)

**Why nobody caught it:** the 2026-05-29 book fact-check "verified" Webster quotes
*against this tool* — wrong-path verification (same failure shape as the glm-streaming
misdiagnosis, memory `feedback_verify_via_real_path`). A verification tool can itself be
the wrong source.

## Blast radius (what's downstream of the bad label)

| Surface | Exposure | Notes |
|---------|----------|-------|
| `scripts/webster-mcp` (`webster_define`, `define`, `webster_search`, `webster_search_definitions`) | **Root cause.** | 98k+ entries; "1828" in tool descriptions + README. |
| **1828.ibeco.me** / `projects/1828-illuminated/` | **High.** The site's whole identity is "1828"; 853-word tier list + pre-fetched defs built from this data. The book QR-links to `1828.ibeco.me/word/spirit` + `/word/intelligence`. | |
| ***Beyond the Prompt*** (book) | Identified: Ch 0 *spirit* quotes, Ch 1 *intelligence* quotes, Ch 9 "warily" gloss. | Already logged as **F-19 / F-20 / F-21** in `projects/scripture-book/.draft/20260609-v4-walk-findings.md`; fixes ride the book's v4 chat walk (genuine 1828 requotes drafted there — they *improve* both passages). |
| **cpuchip.net published studies** | **Unaudited.** `morm-8-three-glories-reading` leaned heavily on "1828" entries (fine / gain / glory / pollute / apparel; "telestial absent from Webster"); any other published study quoting webster_define as 1828. | |
| `study/` (unpublished workspace studies) | Unaudited; many used webster-analysis over months. | |
| `.github/skills/webster-analysis/` + `.claude/` twin | Skill teaches the tool as 1828. | |
| becoming app / ibeco.me | Check whether any surface quotes "1828" definitions. | |

## The remediation plan (steps 1–2 DONE — ratify 3–6 order with Michael)

1. ~~**Forensics first.**~~ **DONE 2026-06-09** — see Forensics Verdict above. We
   introduced the label; data is uniformly 1913 (10/10 words).
2. ~~**Get the genuine 1828.**~~ **DONE 2026-06-09** — kayson-argyle/websters_1828
   verified against webstersdictionary1828.com (see verdict §3 for caveats: license
   undeclared, OCR artifacts). Decision pending: use their SQLite as-is, re-parse
   their raw text with our own cleaning, or both.
3. ~~**Fix webster-mcp honestly.**~~ **DONE 2026-06-09 (ratified: re-parse + keep
   both editions).** New converter `scripts/webster-mcp/tools/convert_1828.py`
   (uses the kayson-argyle pipeline from a local clone — not vendored, their
   scripts are unlicensed; our cleaning on top: 248 junk headwords rejected,
   414 "7 Corinthians"-style 1→7 OCR ref fixes, U+FFFD strip, charity-dup fix)
   → `data/webster1828.json.gz` = genuine 1828 (63,280 words); old data renamed
   `data/webster1913.json.gz` (98,828 words). Server v2.0.0: `webster_define` →
   genuine 1828; NEW `webster1913_define`; `define` → 1828+1913+modern 3-way;
   search tools take `edition` param. Verified over real MCP stdio: spirit =
   genuine text, telephone absent from 1828 / present in 1913. README rewritten
   with full provenance + history note; webster-analysis skill updated both trees.
   **Known parse gaps (for the audit walk): 27 of 853 tier words lack entries —
   mostly coinages/proper nouns, but NAUGHTY, PESTILENCE, ALLEGE, HATH, HOSEN,
   BEFORETIME, SORCERIES, ZINC are real OCR dropouts (neighbors present);
   hand-add from webstersdictionary1828.com during the audit.**
   ⚠ This session's MCP connection still runs the old binary — reconnect to pick up v2.
4. **Rebuild 1828-illuminated** on the genuine data — **CODE DONE 2026-06-09,
   deploy rides Michael's root push** (i1828 Dokploy compose builds from
   cpuchip/scripture-study main, autoDeploy on). Backend seed data swapped to
   genuine 1828; **fixed the skip-if-populated seeder landmine** (new
   `seed_fingerprints` table, migration 006: SeedWebster1828 now sha256-compares
   the embedded corpus and TRUNCATE+re-ingests on change — without this, prod
   Postgres would have kept serving 1913 rows forever). build_data.py re-run
   (832/853 tier words have genuine defs). Verified against a scratch Postgres:
   boot→migrate→seed 63,280 words, /api/dict/1828/spirit serves genuine senses,
   prod-state simulation (rows + no fingerprint) triggers re-ingest.
5. **Audit the published works** (Michael wants to walk this together):
   - grep cpuchip.net `content/studies/*` + workspace `study/` for `1828|Webster`;
   - re-verify every quoted definition against genuine 1828;
   - fix + republish (cpuchip.net pushes auto-deploy);
   - the book's F-19/20/21 land via the v4 chat walk in the book repo.
6. **Close the loop:** docs/06 entry update (resolution), memory note, and a learnings
   entry — the durable lesson is *verify the edition of a source, not just the quote*.

## Verification standard for this whole effort

**webstersdictionary1828.com is the authority** for 1828 text until a better facsimile
source is ratified. Any "1828" claim that ships (tool output, site, study, book) must
trace to it — not to our own mirror, which is the thing under repair.

## Cross-references

- `docs/06_tool-use-observance.md` → "June 9, 2026 — webster-mcp serves Webster 1913
  text under the 1828 label" (the original incident log, same evidence).
- `projects/scripture-book/.draft/20260609-v4-walk-findings.md` → F-19, F-20, F-21, SQ-1.
- Memory: `feedback_verify_via_real_path` (the recurring lesson).
