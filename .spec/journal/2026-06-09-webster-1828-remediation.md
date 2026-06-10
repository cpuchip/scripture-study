# 2026-06-09 — Webster "1828" Data Integrity: Forensics + Remediation

**Session type:** dev/debug (dedicated session Michael queued after the v4 walk discovery)
**Binding question:** who introduced the "1828" label on 1913 data, and how do we fix every surface honestly?

## What happened

Michael opened with "let's dig into that 1828 issue." The spec brief
(`.spec/proposals/webster-1828-data-integrity.md`) carried the whole handoff — evidence,
blast radius, remediation plan. Worked it end to end.

### Forensics (steps 1–2) — the verdict

1. **The mislabel was ours.** ssvivian/WebstersDictionary's README says "Webster's
   Unabridged English Dictionary, provided by the Gutenberg Project" — no year, file
   named `dictionary.json`. Our commit `46c0092` (2026-02-04) renamed it
   `webster1828.json.gz`. The sharpest finding: the same-day plan doc
   (`scripts/plans/04_tool-improvements.md`) listed webstersdictionary1828.com as
   "Online reference for verification" — **the verification was planned and never
   executed.** Michael felt lied to by the data source; the data source never lied.
   We assumed the edition because the goal was 1828.
2. **Uniformly 1913, 10/10 words.** Anachronism probes: the tool defined *telephone*
   (1876), *phonograph* (1877), *bacterium* (1838); its *suffer* quotes Tennyson
   (1855). Text comparisons (spirit, intelligence, charity, glory, virtue, suffer):
   all 1913, zero 1828 matches.
3. **Replacement found + verified before adoption:** kayson-argyle/websters_1828
   (EGW Estate full-text preservation of the actual 1828; built for Standard-Works
   study). Same standard applied: 5/5 anachronisms absent, 6/6 words match the
   authority. Rejected: DataWar (claims its 1828 "came from Gutenberg" — same failure
   shape as ours), kljensen (provenance undocumented).

### Ratified (AskUserQuestion)

Re-parse their raw text (not their unlicensed SQLite); **keep the 1913 under a truthful
name** — Michael explicitly wants both editions side-by-side for the audit and for the
80-years-of-drift view. Scope: webster-mcp + 1828-illuminated this session; the
published-works audit (step 5) is a separate walk together.

### Built

- **`scripts/webster-mcp/tools/convert_1828.py`** — imports the upstream parsing
  pipeline from a local clone (not vendored; their scripts carry no license, the
  dictionary text is public domain), adds our cleaning: 248 junk headwords rejected,
  **414 scripture-ref "7→1" OCR fixes** ("7 Corinthians 8:1"), U+FFFD strip, the
  charity-duplication fix. Output: 63,280 words / 73,124 entries / 113,264 senses.
- **webster-mcp v2.0.0** — `webster_define` → genuine 1828; new `webster1913_define`;
  `define` → 1828 + 1913 + modern (three points in time); `webster_search`/`
  webster_search_definitions` take `edition`. Data renamed honestly
  (`webster1913.json.gz`); 1913 auto-discovered as sibling. Verified over real MCP
  stdio. README rewritten with provenance + history note; webster-analysis skill
  updated in both trees (its own D&C 93 example had quoted 1913-flavored text — fixed
  with verified genuine senses; D&C 93:29 read from the local file before quoting).
- **1828-illuminated** — backend seed swapped to genuine corpus; build_data.py re-run
  (832/853 tier words covered). **Adjacent-surface catch: the seeder's
  skip-if-populated guard would have kept 1913 rows in prod Postgres forever.** Added
  `seed_fingerprints` (migration 006) + sha256 compare + TRUNCATE-and-re-ingest.
  Verified via a scratch postgres:17 container: clean boot/migrate/seed,
  `/api/dict/1828/spirit` serves genuine senses, prod-state simulation (rows present,
  no fingerprint) triggers re-ingest, fingerprint match skips.

## Surprises

- The OCR 1→7 scripture-ref error was *pervasive* (414 instances), not a one-off.
- Genuine 1828 *glory* (12 senses, "Brightness; luster; splendor") differs far more
  from the 1913 than expected — the three-glories study audit will be substantive.
- ~27 tier words have no genuine-1828 entry; NAUGHTY, PESTILENCE, ALLEGE, HATH, HOSEN,
  BEFORETIME, SORCERIES, ZINC are real upstream OCR dropouts (neighbors present).
  Hand-add from the authority during the audit walk.

## Carry-forward

1. **Michael pushes root** → i1828 auto-redeploys → verify live re-ingest +
   `/word/spirit` + `/word/intelligence` (the book's QR targets). Task #6.
2. **Published-works audit walk together** (cpuchip.net studies, workspace `study/`,
   book F-19/20/21 via the v4 chat walk).
3. MCP reconnect to pick up webster-mcp v2 in the harness.
4. Optional: upstream the OCR fixes / missing headwords to kayson-argyle; ask about a
   license for their repo.

## The durable lesson

Verify the **edition** of a source, not just the quote. Anachronism probes are the
cheap decisive test for any historical dataset. And a verification tool can itself be
the wrong source — the 2026-05-29 book fact-check "verified" Webster quotes against
the mislabeled tool. Wrong-path verification recurring; see
`feedback_verify_via_real_path`.
