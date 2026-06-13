# Per-file workflow — study correctness walk

Ratified scope: `.spec/proposals/study-correctness-walk.md`. Worked example of
every finding class: the three-glories correction (2026-06-09, journal
`projects/cpuchip.net/.spec/journal/2026-06-09-three-glories-edition-correction.md`).

## For each file (top to bottom of `_progress.md`)

1. **Read the file fully.** No skimming — ARGUMENT-class findings hide in how
   a quote is *used*, not just its wording.
2. **Extract claims by class:**
   - Webster quotes (T1) — every quoted definition, including paraphrased
     "1828 says…" claims (T2)
   - direct quotes (FULL: all; SAMPLE: 3 spread across the file) — scripture,
     talks, transcripts
   - stats/dates/counts ("cited six times", "first published 1944")
   - links (all tiers)
3. **Verify:**
   - Webster → `webster_define` (genuine 1828; v2 with variant fallback);
     `webster1913_define` to identify where a wrong quote came from;
     webstersdictionary1828.com for disputes (data/known-issues.md lists the
     facsimile-only cases — do not chase those)
   - quotes → `Read` the actual gospel-library / source file (cite-count rule)
   - links → md-link-validate (md-mcp) or Read-check relative paths
4. **Classify + act:**
   - CLEAN — note in findings.md (one line)
   - REQUOTE — wrong-edition or drifted wording, argument unaffected → fix
     in place now, log before/after (act-and-report)
   - ARGUMENT — the claim leaned on wrong text (apparel-bridge class) →
     log under "needs Michael" in findings.md; do NOT rewrite solo
   - LINK/TYPO — fix now, log
5. **Published surfaces:** if the study is republished (cpuchip.net), apply
   the same fixes there + a visible correction note (three-glories precedent);
   cpuchip.net push = deploy (granted).
6. **Mark `[x]` in `_progress.md` with the finding count, commit both files**
   (+ the fixed study). The commit is the resume point.

## SAMPLE escalation rule (learned 2026-06-12, file #18)

If ANY sampled quote fails, escalate the file to FULL on the spot. The eq/
atoning-love file failed its first sample (invented closing quote) and FULL
verification then found a confabulated biography of a living man.

**Risk marker (evidence from files #17-18):** quotes from talks the model
"knows" only thinly — especially Apr/Oct 2025 conference — are the
confabulation hot zone. The split is consistent: talks READ from disk get
quoted exactly; talks RECALLED get invented wholesale (titles right, links
right, words fabricated). Any file quoting 2024-2025 talks: verify every
talk quote regardless of tier. Scripture blocks stay near-perfect
everywhere; talk-quote density × recency is the risk product.

## Unattended laps (gated autonomy, ratified)

- Bins 1–2 only: verify, requote, links. ARGUMENT class always queues.
- 529/overload: degrade loud, log it in findings.md, re-probe every ~30 min
  (the BoM-walk lesson — never stay silently degraded).
- End every lap: commit progress + a one-line lap summary in findings.md.
- Spending judgment: if a file's findings feel like they need taste (voice,
  doctrine, public correction wording), stop at the queue — that is the line.

## Post-boundary rule (learned 2026-06-12, file #86 — the Feb-19 verdict)

The Feb-19 tool boundary held for SCRIPTURE: the first post-boundary file
had every scripture and talk quote verbatim. It did NOT hold for Webster:
all six 1828 definitions in that same file were the 1913 edition's,
because the webster-mcp dictionary itself served 1913-as-1828 until the
2026-06-09 repair. Study date is no shield for Webster quotes.

Amended tiering for post-Feb-19 files:
- **Webster quotes: verify EVERY one, every tier.** They are now the
  highest-risk class (16 of the first 22 corrupt words were found via
  full Webster checks).
- **Scripture: sample normally** per tier; escalate on any failure as
  usual. The discipline is holding.
- **2024-25 talk quotes: still verify regardless of tier** (hot-zone rule
  unchanged).
