# Study Correctness Walk — every study, in order, against the fixed dictionary

**Status: RATIFIED 2026-06-12 (AskUserQuestion).** Michael's decisions:
**T3 = FULL direct-quote verification on all 90 top-level studies** (not just
the 33 Webster ones; sample + links elsewhere) · **chronological order** ·
**gated autonomy** (mechanical fixes act-and-report; ARGUMENT-class findings
queue for sessions with Michael; overnight laps allowed under
unattended-resilience rules) · **bom-walk = the 12 Webster files** full
treatment, the other 227 link-validation only.

Working dir: `study/.audit/` (progress + findings + workflow). Drafted
2026-06-12 (webster-1828 lane). Companion to
[`webster-1828-data-integrity.md`](webster-1828-data-integrity.md); leg 1
(three-glories) already walked 2026-06-09 and is the template.

**Binding question:** does each study say what its sources actually say — now
that we know four months of "Webster 1828" quotes were really Webster 1913?

## The numbers (counted 2026-06-12)

| Surface | Files | Webster exposure |
|---|---|---|
| `study/**` (excl. scratch) | 469 | 108 mention Webster/1828 |
| — with **quoted** definitions | **76** | 33 top-level · 13 `yt/` · 12 `bom-walk/` · 10 `plan-of-salvation/` · 8 other |
| cpuchip.net published studies | ~8 | three-glories ✅ done; abinadi + voice-from-the-dust + others to check |
| *Beyond the Prompt* (book) | — | F-19/20/21 ride the book's own v4 chat walk (not this walk) |

## Tiers (the dial Michael sets)

- **T1 — full Webster requote (76 files).** Every quoted definition re-verified
  against the genuine 1828 (`webster_define`, now actually 1828) with the
  three-glories discipline: not just text-swap — check whether the *argument*
  survives the genuine text, and let the genuine entry strengthen it where it
  does (it did, three times, in three-glories). Use `define`'s 1828→1913→modern
  view when drift itself is the point. Wrong-edition quotes that carried an
  argument get the apparel-bridge treatment: surface to Michael before rewriting.
- **T2 — Webster-mention check (32 files).** Mentions without quotes: confirm
  no paraphrased "1828 says…" claims hide there; verify the few that do.
- **T3 — general correctness (all 469, sampled).** While in each file:
  (a) **link validation** — mechanical, every file (md-mcp `md-link-validate`);
  (b) **direct-quote verification** — full for the 33 top-level T1 studies
  (read-before-quoting applied retroactively, cite-count rule); a 3-quote
  sample per file elsewhere; (c) **stat/date/count claims** — flag and verify
  the "Maxwell cited it six times" class. T3 depth is the main cost dial:
  full-everything ≈ months; the sampling above ≈ tractable.

## Order

Chronological by file creation (`git log --diff-filter=A`), oldest first —
"in order" per Michael, and the oldest studies predate the strongest
verification discipline, so findings-per-file should be highest there.
Published studies get fixed+republished as they come up in sequence (not
batched at the end) — a wrong published page outranks an unpublished one.

## Mechanics (bom-walk pattern)

- Working dir `study/.audit/`: `_progress.md` (file list in order, one line
  per file: pending → done + finding count; the resume mechanism — commit
  after every file), `findings.md` (per-study sections, append-only),
  `_workflow.md` (the per-file checklist below, including the
  unattended-resilience rules: degrade loud, re-probe on a cadence).
- **Per-file workflow:** read fully → extract claims (Webster quotes, direct
  quotes, stats, links) → verify each (webster_define / Read the source file /
  md-link-validate) → classify: CLEAN · REQUOTE (mechanical — fix now, log) ·
  ARGUMENT (the claim leaned on wrong text — queue for Michael, do not rewrite
  solo) · LINK/TYPO (fix now) → commit file + progress.
- **Autonomy split (stuffy-in-the-loop):** mechanical requotes and link fixes
  = act + report (bins 1–2; Dave rule). ARGUMENT-class findings = surface-first:
  batched in `findings.md` under "needs Michael," walked together like the
  apparel bridge. Published-page corrections ship with a visible correction
  note (three-glories precedent) — that note policy is itself act+report.
- **Sessions:** T1 top-level (33 files) ≈ 2–4 focused sessions or a guarded
  autonomous run with the ARGUMENT-class gate; `yt/` + subdirs autonomous-able;
  bom-walk's 12 = one session. Each session ends with progress committed, so
  any session can resume the walk cold.

## Dependencies / preconditions

1. **The repaired dictionary ships first** (OCR repair underway in this lane:
   ~400 scripture-ref fixes, ~350 junction fixes, destroyed-sense restoration
   from webstersdictionary1828.com, variant layer for 1828 spellings).
   Walking studies against a dictionary we're about to change would double work.
2. webster MCP v2 connected (this session has `webster1913_define` — yes).
3. Three-glories correction (done) stands as the worked example of every
   classification: REQUOTE (glory ladder), ARGUMENT (apparel bridge), and the
   strengthen-don't-just-fix posture.

## Ratify with Michael

1. **T3 depth** — sampling as proposed, or full direct-quote verification on
   more than the 33 top-level studies?
2. **Order confirmation** — strict chronological, or published-first then
   chronological?
3. **Autonomy** — may the walk run unattended (overnight laps) with the
   ARGUMENT-class gate, or stay interactive?
4. **bom-walk scope** — its 12 Webster-quoting chapter notes only, or spot-walk
   all 239 notes under T3 sampling?
