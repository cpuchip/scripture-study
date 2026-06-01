# Study Provenance Review — Plan

**Status:** scoped, awaiting execution (gated on Michael's compaction)
**Origin:** Michael, 2026-06-01 — "a full review of our studies and study provenances. there are a lot of early studies that don't have provenance .scratch files. it shows our history a little, and how we've come, I think it could sharpen some of our earlier work too." Asked which vehicle: pg-ai-stewards / subagents / agy -p. "we have a fair bit of resources to play with for fun takes like this."
**Author:** Claude Code (Opus 4.8). Written pre-compaction so the plan survives the context reset and execution resumes cleanly (ammon).

---

## The data (measured 2026-06-01)

- **198** study docs under `study/` (incl. subdirs: `yt/`, `talks/`, `cfm/`, `eq/`, `plan-of-salvation/`, `podcast/`, `teaching-in-the-saviors-way/`, etc.)
- **54** provenance files under `study/.scratch/`
- **The fault line is datable.** Every top-level `study/*.md` modified before **~2026-03-27** has no `.scratch`. The convention begins at `art-of-presidency.md` (2026-03-27) and holds from there forward. The genesis cohort — `creation`, `intelligence`, `word`, `heavenly_mother` (2026-01-25) through the February doctrinal run — is the gap Michael means.

The early cohort (no provenance), top-level, ~40 files:
agency, ai-responsible-use(+reflections), atoning-love-andersen, charity, consumption-decreed(+modern-warning), covenants, creation, doctrines-principles-programs, end-times, enjoy-the-words-of-eternal-life(+reflection), enoch(+charity), faith-01(+a), gadianton-robbers, gifts, heavenly_mother, helaman-ten-virgins, helaman-why, intelligence(+01), know-god, language-of-adam, mazzaroth(+01+02), miracles-references, moses-6-gospel-to-adam, order-of-god(+modern-lens), priestcraft(+beguile), priesthood-and-gifts, priesthood-oath-and-covenant, priesthood-obtaining-exploration, receive, serpent-and-dragon, ten-virgins-parable, translated-beings, truth(+atonement+enjoy+modern-prophets), way-truth-life, word, zion-blueprint. Plus early subdir work to map separately.

(Note: `.scratch` ↔ study matching is NOT 1:1 by basename. Several `.scratch` files have no same-name study — `divine-love`, `how-is-it-done`, `nevertheless`, `only-begotten`↔`only-begotten-deeper`. Phase 1 builds the real map.)

---

## The hazard (load-bearing — read before executing)

Memory `project-scripture-book-provenance-redemption` (2026-05-26): backfilling provenance after a doc exists produces **"a documentation pass that confirms what was already produced,"** not an audit. That is exactly how a fabricated D&C 104:11-12 quote entered the scripture-book audit trail. **Provenance only works as a gate, not as a postscript.**

Therefore this review does **NOT** fabricate the working notes that "would have" preceded a January study. That would re-commit the drift we just redeemed. Instead:

## The reframe — three honest things, not one dishonest one

1. **Verify (the real value, bin 1 — ground truth).** The early studies predate the read-before-quoting discipline. They are the likeliest home of confabulated/close-enough quotes. Read each, extract every quoted scripture/talk/source, verify against `gospel-library/`. Emit a findings list: `study → quote → {match | mismatch + correct text | unverifiable}`. This is the same work the scripture-book redemption did, and it is the highest-value, ground-truth-checkable, unsupervised-safe core.

2. **Reconstruct provenance honestly, dated (bin 2).** Where we add a `.scratch` for an early study, it is explicitly a **2026-06 verification log / re-derivation**, marked as backfilled — NOT a fake "here's how we discovered this in January." It records: the verification findings; what re-running `gospel_search` on the binding question surfaces *now*; any drift caught. This is the scripture-book pattern (name the reconstruction; don't perform a false history).

3. **Propose sharpening (bin 4 — Michael's call).** "Sharpen earlier work" = rewriting a study = his judgment + voice + the Spirit. The agent produces redline-style PROPOSALS, surfaced for review and ratification. Never republishes a study autonomously.

---

## Vehicle — right tool per phase (the answer to Michael's question)

The phases have different judgment profiles, so they want different tools. The spine: **verification by Claude reading actual sources** (a cheap model fanned out claiming "verified" reintroduces the exact confabulation we're auditing for — feedback_verify_via_real_path); **generation/proposals by the substrate panel** (human-reviewed anyway, and its sweet spot); **fresh-eyes diversity optionally by agy -p**.

| Phase | Work | Bin | Vehicle | Why |
|-------|------|-----|---------|-----|
| **1. Map** | Inventory + match `.scratch`↔study + classify cohort | 1 | **Claude Code direct** | Fast, exact, filesystem ground truth. Mostly done above. |
| **2. Verify** | Quote-check early studies against `gospel-library/` | 1 | **Claude Code subagents (Agent fan-out)** | Needs our source-verification discipline; the value IS the trustworthiness of the check, so the checker must read the real file. NOT a fanned-out cheap model (would re-confabulate). NOT agy/Gemini (harness gap — doesn't reliably carry the discipline). |
| **3. Reconstruct** | Honest dated provenance `.scratch` + re-derivation `gospel_search` | 2 | **Claude Code** (+ optional substrate pre-gather) | Honesty frame must hold; this is where the scripture-book trap lives. |
| **4. Sharpen** | Redline edit proposals on early studies | 4→his | **pg-ai-stewards `panel_redline`** (now with real Claude via zen) + optional **agy -p** fresh eyes | The substrate's purpose-built strength; produces proposals he ratifies. "Plays with the resources" where it genuinely fits. |

**One-line vehicle answer:** verification = Claude (me/subagents) on real sources; sharpening-proposals = the substrate redline panel; agy -p = optional second take. Not the substrate for the verification gate — that's precisely where a fanned-out model would re-introduce the confabulation we're hunting.

---

## Autonomous vs. his (stuffy-in-the-loop)

- **Unsupervised-safe (bins 1-2):** Phase 1 map, Phase 2 verification findings, Phase 3 honest-dated reconstruction. These gather/verify/draft and emit reviewable artifacts. Can run without him.
- **His (bins 3-4):** any actual rewrite of study content (Phase 4 is proposals only); any change that alters a published study's claims; ratifying the redline panel's output. Surface, don't execute.

## Output artifacts

- `study/.scratch/_provenance-review-2026-06.md` — the master findings log (per-study verification table, drift caught, reconstruction status).
- Per-study `study/.scratch/{name}.md` honest dated verification/re-derivation notes for the early cohort.
- A sharpening-proposals queue (redline panel output) for council review — NOT applied.

## Resume instructions (post-compaction)

1. Re-read this file + memory `project-scripture-book-provenance-redemption` + `feedback_verify_via_real_path`.
2. Phase 1 map is essentially done (see data above) — formalize it into `_provenance-review-2026-06.md`.
3. Start Phase 2 verification on the early cohort, fanning out to subagents in batches of ~8 studies; each returns a quote-by-quote findings table verified via `gospel_get`/`Read` against `gospel-library/`.
4. Phase 3/4 follow. Sharpening stays proposals-only until Michael ratifies.
