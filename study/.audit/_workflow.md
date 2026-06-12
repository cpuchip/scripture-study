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

## Unattended laps (gated autonomy, ratified)

- Bins 1–2 only: verify, requote, links. ARGUMENT class always queues.
- 529/overload: degrade loud, log it in findings.md, re-probe every ~30 min
  (the BoM-walk lesson — never stay silently degraded).
- End every lap: commit progress + a one-line lap summary in findings.md.
- Spending judgment: if a file's findings feel like they need taste (voice,
  doctrine, public correction wording), stop at the queue — that is the line.
