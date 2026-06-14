# Doctrine & Covenants Walk — Per-Section Workflow

Walk #3 of the canon-walk series. Adapted from the [PoGP walk's `_workflow.md`](../pogp-walk/_workflow.md) (itself from the BoM walk) — **the honesty / cost / resilience / resume frames carry over unchanged.** **Reload this at each "decade" boundary (every ~10 sections) and after any compaction** to stay anchored over the long run (140 units — this is a long walk, like the BoM's 239).

## What this walk is (depth scales to the section — see the calibration note)

- **IS:** a start-to-finish walk where *each section gets a digestion note* — the historical setting engaged, footnote chains followed across all volumes, real cross-referencing (`gospel_search` + our corpus), load-bearing word work (`webster_define`; `strongs_*` where a section quotes/echoes the KJV), the critical-analysis discipline (strongest claims, counter-readings, named tensions), and an analytical note.
- **★ Depth scales to the section (D&C-specific).** The D&C is wildly uneven in length and weight: one-verse sections (D&C 13) and situational calls to named individuals (instructions to Thomas Marsh, the Whitmers, etc.) sit beside the great revelations (76, 88, 93, 107, 121-122, 132, 138). **Calibrate:** the doctrinal landmarks get full mini-study depth; short / situational / administrative sections get tight, faithful notes (setting + the verse(s) that matter + connections). Don't pad a 3-verse section to look like section 88. *(Final depth/coverage dial = Michael's call at kickoff; see _progress.md header.)*
- **The one trim vs. a full study:** no per-section voice-audit polish, no forced Becoming section, no finished-study ratification. The note is a *digestion + provenance*, not a publishable study. Still **bin-1/2** (gather / draft-for-Michael).
- **One-go intent:** ammon + the progress file carry it to completion across many compactions. 140 units; `_progress.md` is the thread that survives the resets.

## Honesty frame (non-negotiable)
- Exploratory digestion — **clearly mine, never doctrine.** Speculation labeled. Calibrated confidence.
- **Read before quoting.** Every quoted phrase verbatim from the section text read this session (`gospel_get`). If unverified, paraphrase — no quotation marks. (Verify counts/dates/biographical claims the same way — the D&C is full of names, dates, and places.)
- bin-1/2 under `stuffy-in-the-loop`, authorized unsupervised; never published doctrine without ratification.

## Cost discipline (the $18 zen budget)
- pg-ai-stewards = retrieval index only (`study_search`, `study_similar`, `study_get`). No opus, no cost.
- **Never** spawn per-section LLM generation on the substrate (`panel_redline`, `start_brainstorm`, `spawn_subagent`, `consult_subagent`, `deep_research`, `audit_*`). The thinking is mine.
- `byu_citations` IS used here (D&C citation density is genuinely informative) — local query, not opus. `brain_*` is dead.

## The per-section loop
1. **READ** — `gospel_get "D&C {N}"`. Read the full section, its **historical heading** (who/when/where/why — load-bearing for the D&C), and the footnotes.
2. **SETTING** — name the historical context from the heading: date, place (Harmony / Fayette / Kirtland / Missouri / Liberty Jail / Nauvoo), the person(s) addressed, the occasion. The D&C is revelation *embedded in history*; the setting often is the key.
3. **FOLLOW** — trace footnote chains that *add, clarify, or reframe* (into OT/NT/BoM/PoGP, not just nearby verses).
4. **THINK (study it out).** Work the section: what it says, what strikes me, tensions, counter-readings, questions. Critical-analysis lens. Name tensions rather than resolving them.
5. **CONNECT (cross-volume + cross-walk).**
   - `gospel_search` on the section's binding idea → non-obvious cross-refs across all five standard works.
   - `study_search` / `study_similar` → our existing studies (name the slug + why). **HARD link-don't-duplicate** — we have deep D&C-theology studies (see below).
   - `webster_define` on load-bearing words (1828/1830s register); `strongs_*` where a section quotes the KJV.
   - `byu_citations` where how the Brethren have used a section changes the reading, or to mark heavily-cited vs. underexplored sections.
6. **VERIFY** — quotes verbatim; any date/name/count traced to a tool call this session.
7. **WRITE** — the section note → `study/dc-walk/{NNN}.md` (zero-padded: `001.md`…`138.md`, plus `od-1.md`, `od-2.md`). Extend `_graph.md`. Append to `_journal.md` only if something genuinely stood out.
8. **TRACK** — update `_progress.md` (mark done, advance NEXT). **Commit per section** (or per 2-3 short ones), **no push** (Michael pushes when he reads). [Confirm cadence with Michael — default to per-section like PoGP.]
9. **RELOAD** — at each decade boundary (10/20/30…), re-read this file + `_progress.md`.

## The section note format
```markdown
# D&C {N} — {short title}

**Setting:** date · place · who/why (from the heading)
**Binding idea:** the one thread this section turns on
**Read:** what the section says — brief, my framing
**Thoughts:** my digestion — what strikes me, tensions, questions
**Connections:**
- scripture: cross-refs from footnotes + gospel_search
- our studies: study_search / study_similar hits (slug — why)
- words: webster_define / strongs where it earns it
- citations: byu_citations where it changes the reading
**Entities:** people · places · doctrines · covenants · ordinances · offices
**Edges:** {this} —[cross-ref|fulfillment|parallel|covenant-thread|links-to-study|cross-walk]→ {that}
**Notable / flag:** anything that "pops" worth Michael's eye
**Verified:** quotes checked against the section ✓
```

## D&C-specific notes

### Link, don't duplicate — our D&C-theology corpus is deep
The D&C overlaps heavily with studies we already hold. **D&C 93 = `truth.md` ("the unified field theory")**; the matter-spectrum (`truth`, `truth-atonement`, `intelligence`, `mechanics-of-refinement`) leans on **88, 93, 130, 131**; **`priesthood-oath-and-covenant` = D&C 84**; **`consumption-decreed` = D&C 87**; `divine-love`, `nevertheless`, `plan-of-salvation`, `enoch`/`zion-blueprint` (Zion sections 42/57/105) all have D&C anchors. **Run `study_search` early on each section so I know what we hold before writing.** The walk's contribution is to digest the section on its own terms, name its distinct contribution, and link — not re-derive the studies.

### ★ Project-source-text moments — flag them as they land (like Abr 4:18 in the PoGP walk)
The D&C is the densest book for *this project's own foundations*. Mark these as significant when reached:
- **D&C 82:10** ("I, the Lord, am bound when ye do what I say") — the epigraph of `covenant.yaml`.
- **D&C 88:119** ("organize yourselves; prepare every needful thing…") — the "Organize Before Building" principle in `.mind/principles.md`.
- **D&C 121:34-46** ("no power or influence… only by persuasion… without compulsory means") — the **source of the presiding covenant extension** (`covenant.yaml presiding:`).
- **D&C 130:18-19** ("whatever principle of intelligence we attain unto in this life, it will rise with us") — the project's thesis thread (intent.yaml, identity).
- **D&C 107:99** ("let every man learn his duty, and… act in the office") — `honor_scope`.
- **D&C 9:7-9** ("study it out in your mind; then ask me… you shall feel that it is right") — the verification epistemology the whole project rests on.
- **D&C 50, 121:26 / 8:2-3** — revelation by the Spirit / light (pairs with the matter-spectrum + the "period-language + Spirit-distillation" principle).

### ★ Cross-walk closures — the threads from walks #1 and #2 land here
- **OD 2 (1978)** resolves the **priesthood-lineage bin-4 thread** flagged across the BoM (2 Ne 5:21, Alma 3) AND the PoGP (Moses 7:8/22, Abr 1:21-27). The thread that spanned both prior walks lands at OD 2 — treat it with the full bin-4 care + the 2013 "Race and the Priesthood" essay context. **Counterweight already logged: Abr 2:10** (adoption-by-gospel).
- **D&C 84 / 107** (priesthood) ↔ Moses 6:7 / Abr 1:3 (the unbroken line) — the priesthood thread continues from the PoGP.
- **D&C 76 / 88 / 93 / 131** ↔ the matter-spectrum studies (the BoM/PoGP "is-Christ-God" + creation threads mature here into exaltation/glory).
- **D&C 87** (the consumption decreed) ↔ the BoM walk's secret-combinations + consumption threads.

### Recurring-thread ledger (track in `_progress.md`)
- **Matter-spectrum / light & truth / intelligence** (88, 93, 130, 131…)
- **Priesthood — restoration, oath & covenant, offices, the presiding terms** (13, 20, 84, 107, 121)
- **Degrees of glory / exaltation / the nature of God & man** (76, 88, 131, 132, 137, 138)
- **Consecration / Zion / the law** (42, 51, 57, 82, 104, 105)
- **Covenant & sealing / eternal marriage / the keys** (110, 124, 131, 132)
- **Gathering / last days / the Second Coming** (29, 45, 87, 88, 133)
- **★ Project-source-texts** (82, 88:119, 121, 130, 107:99, 9:8 — see above)

## Spin-off avenue (lean toward linking)
Some sections deserve a full study (76, 88, 93, 132, 138 are candidates). But **first check whether an existing study already holds it** — given our corpus, most deep D&C theology is covered, so usually *link*, not spin off. Spin off ONLY genuinely new ground: `Agent(subagent_type="study", model="opus"|"sonnet")` → `study/dc-walk/studies/{sec}_{subject}.md`; then read it, record the takeaway + link in the section note, move on.

## Unattended-run resilience (carried)
529s strike background subagent spawns first. (1) Degrade gracefully — fall back to fuller inline treatment, keep advancing. (2) Re-probe on a cadence (every ~10 sections). (3) Make degradation LOUD at the top of `_progress.md`.

## Resume after compaction (ammon)
A 140-unit run crosses context resets. On resume — do not restart: (1) read `_progress.md` → NEXT section; (2) read this `_workflow.md`; (3) skim the last 1-2 notes + the tail of `_graph.md`; (4) continue from NEXT. "Remember all my commandments to execute them" (Alma 18:10).
