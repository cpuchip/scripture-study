# Pearl of Great Price Walk — Per-Chapter Workflow

Walk #2 of the canon-walk series. Adapted from the [BoM walk's `_workflow.md`](../bom-walk/_workflow.md) — **the honesty / cost / resilience / resume frames carry over unchanged.** **Reload this at each book boundary and after any compaction** to stay anchored.

## What this walk is (same calibration as the BoM walk: mini-study depth)

- **IS:** a start-to-finish walk where *each chapter gets a small study* — footnote chains followed across all volumes, real cross-referencing (`gospel_search` + our corpus), load-bearing word work (`webster_define`; `strongs_*` where a KJV-shared phrase has Hebrew/Greek underneath), the critical-analysis discipline (strongest claims, counter-readings, named tensions), and a substantial analytical note. Plus the connection graph and the standout journal.
- **The one trim vs. a full study:** no per-chapter voice-audit polish, no forced Becoming section, no finished-study ratification. The chapter note is a *mini-study + provenance*, not a publishable polished study. **Depth of engagement = study-grade; level of finish = note-grade.**
- Still **bin-1/2** (gather / draft-for-Michael) — exploratory, mine, quotes verified — never published doctrine without his ratification.
- **One-go intent:** ammon + the progress file carry it to completion across compactions. Small (16 chapters) but dense; `_progress.md` is the thread that survives the resets.

## Honesty frame (non-negotiable)

- My notes are **exploratory digestion — clearly mine, never doctrine.** Speculation is labeled as speculation. Calibrated confidence: strong where the text is plain, tentative where I'm inferring.
- **Read before quoting.** Every quoted phrase is verbatim from the chapter text I read this session via `gospel_get`. If I haven't verified it, I paraphrase — no quotation marks.
- This is **bin-1/2 work** under `stuffy-in-the-loop`, authorized unsupervised. It never becomes a published study without his ratification.

## Cost discipline (the $18 zen budget)

- pg-ai-stewards is used **only as a retrieval / connection index**: `study_search`, `study_similar`, `study_get`. DB queries — **no opus, no cost.**
- **Never** spawn per-chapter LLM generation on the substrate (`panel_redline`, `start_brainstorm`, `spawn_subagent`, `consult_subagent`, `deep_research`, `audit_*`). The thinking is mine — Claude Code tokens. The $18 is reserved for real study workflows.
- `brain_*` is **dead** — do not use.

## The per-chapter loop

1. **READ** — `gospel_get "{Book} {N}"`. Read the full chapter, its heading, and the footnotes.
2. **FOLLOW** — trace the footnote chains that *add, clarify, or reframe* (deep-reading). Note the ones that matter; don't chase all of them.
3. **THINK (study it out — mini-study depth).** Genuinely work the chapter: what it says, what strikes me, the tensions, the counter-readings, the questions. Run the critical-analysis lens — strongest claims against the text, weakest links, missing voices, speculation vs. doctrine. Name tensions rather than resolving them. This is the heart; do it fully *before* moving on.
4. **CONNECT (cross-volume).**
   - `gospel_search` (semantic or hybrid) on the chapter's binding idea → non-obvious cross-refs across all five standard works.
   - Follow the footnote chains that *reframe* — into OT/NT/D&C/BoM, not just nearby verses.
   - `study_search` / `study_similar` → links into our existing studies (name the slug + why). **PoGP-specific: lean HARD on this — see "Link, don't duplicate" below.**
   - `webster_define` on load-bearing words where the 1828 sense differs from modern; `strongs_*` where a KJV phrase (esp. Moses paralleling Genesis, JS–M paralleling Matt 24) has Hebrew/Greek worth surfacing.
   - `byu_citations` sparingly, where how the Brethren have used a verse changes the reading.
5. **VERIFY** — quotes verbatim against the text read; any count or claim traced to a tool call this session.
6. **WRITE** —
   - the chapter note → `study/pogp-walk/{book}/{book}-{NN}.md` (format below). This note is *also* the scratch provenance.
   - extend `_graph.md` with new nodes/edges.
   - append to `_journal.md` **only if something genuinely stood out** (not forced every chapter).
7. **TRACK** — update `_progress.md` (mark done, advance the NEXT pointer). **Commit per chapter** (Michael's directive 2026-06-14 — every chapter, not every 2-3) so any step walks back cleanly. **Do NOT push** (his directive — he pushes when he reads).
8. **RELOAD** — at each book boundary, re-read this file + `_progress.md` to re-anchor the loop and resist drift.

## The chapter note format

```markdown
# {Book} {N} — {short title}

**Binding idea:** the one thread this chapter turns on
**Read:** what the chapter says — brief, my framing
**Thoughts:** my own digestion — what strikes me, tensions, questions
**Connections:**
- scripture: cross-refs from footnotes + gospel_search
- our studies: study_search / study_similar hits (slug — why)
- words: webster_define / strongs where it earns it
**Entities:** people · places · doctrines · types/symbols · prophecies · covenants
**Edges:** {this} —[cross-ref|fulfillment|parallel|type→antitype|covenant-thread|links-to-study]→ {that}
**Notable / flag:** anything that "pops" worth Michael's eye
**Verified:** quotes checked against the chapter ✓
```

## PoGP-specific notes

### Link, don't duplicate — our creation/council corpus is already deep
The PoGP overlaps heavily with studies we've already done. Moses 1–8 and Abraham 3–5 retread ground covered in `creation.md`, `moses-6-gospel-to-adam.md`, `enoch.md`, `enoch-charity.md`, `only-begotten.md` / `only-begotten-deeper.md`, `language-of-adam.md`, the `plan-of-salvation/` folder, and `is-jesus-christ-god.md`. **The walk's contribution is NOT to re-derive what those studies already hold — it's to (a) digest the chapter on its own terms, (b) name the chapter's distinct contribution, and (c) link back to the existing study.** Run `study_search` / `study_similar` early on each chapter so I know what we already have before I write.

### Facsimiles (default: fold into the Abraham chapters)
The local corpus has only Abraham 1–5 — the three facsimile **explanations are not in the gospel-library files.** Default treatment: fold facsimile references into the relevant Abraham chapter notes (Fac. 1 = the altar/attempted sacrifice → Abr 1; Fac. 2 = the hypocephalus, Kolob & the governing-stars cosmology → Abr 3; Fac. 3 = Abraham in Pharaoh's court → Abr 1/3). If Michael wants dedicated facsimile notes, pull the explanation text via `gospel_get` first (read-before-quoting still applies) and give Fac. 2 its own note.

### Joseph Smith–History is narrative, not exposition
JS–H is the founding-events account (First Vision, Moroni, the plates, Restoration of priesthood). The note format still works, but the treatment shifts toward *what happened, the sources Joseph engages (James 1:5, Joel, Malachi 3–4, Acts 3), and the doctrinal weight of the events* rather than verse-by-verse exegesis. Same for the Articles of Faith — 13 distilled doctrinal statements; the note traces each article to its scriptural roots.

### Recurring threads to track (the PoGP's own datapoint lines)
The BoM walk tracked *is-Jesus-Christ-God* and *2 Ne 5:21 (curse/lineage)* threads. The PoGP has its own. Track these across chapters as a running ledger in `_progress.md`:
- **Christ as Jehovah/Creator/Only Begotten** — Moses 1:6, 32–33; 2:1; Abr 3 (the Lord among the noble and great); the Only Begotten language throughout Moses. Continues the BoM's is-Christ-God thread.
- **The premortal council & the plan** — Moses 4:1–4 (Satan's rebellion, agency); Abr 3:22–28 (the noble and great, the two estates, "we will prove them"). Cross-link to `plan-of-salvation/`.
- **The Abrahamic covenant** — Abr 1:18–19; 2:8–11 (seed, land, priesthood, the gospel to all nations). Cross-link to BoM covenant threads + D&C 132/Abr.
- **The pattern of apostasy & restoration** — Moses' lost-then-restored text frame; JS–H. The "plain and precious things" thread (1 Ne 13).
- **Enoch & Zion** — Moses 7. Cross-link to `enoch.md`, `enoch-charity.md`, `zion-blueprint.md`.

## Spin-off avenue (lean toward linking, not spinning off)
Some chapters surface something deserving a full study. Michael's directive (carried from the BoM walk): **don't chew my own cycles on it mid-walk.** But for PoGP, **first check whether an existing study already holds it** — given our corpus, most "deep" PoGP material is already covered, so the right move is usually *link*, not spin off. Spin off ONLY genuinely new ground:
- **Worth a full study (new ground):** `Agent(subagent_type="study", model="opus"|"sonnet")` with instructions to run the full study-workflow and write to `study/pogp-walk/studies/{ch-tag}_{subject}.md` (e.g. `moses1_endless-vision.md`, `abr3_governing-stars.md`). Sonnet for moderate; opus for the hard ones.
- After it returns, **read it, record the takeaway + link in the chapter note and `_graph.md`, then move on.**
- Still bin-1/2 drafts-for-Michael until ratified.

## Unattended-run resilience — subagent overload (carried from the BoM walk)
A long run *will* hit transient API failures — most often **HTTP 529 "Overloaded."** Three rules:
1. **Degrade gracefully — never stall the main goal.** Subagent spawn fails → fall back to a fuller **inline** treatment and keep advancing. Finishing the walk is load-bearing; spin-offs are an enhancement.
2. **Re-probe the degraded capability on a cadence.** 529s are transient. After deferring a spin-off, retry at the next spin-off-worthy chapter, or every ~5 chapters (this walk is short). Don't stay degraded for the whole run.
3. **Make degradation LOUD.** Log it at the top of `_progress.md` (e.g. `DEGRADED HH:MM — subagent spawns 529'ing, carrying inline, will re-probe`), not only in a commit message.

## Resume after compaction (ammon)
On resume — do not restart:
1. Read `_progress.md` → the NEXT un-digested chapter.
2. Read this `_workflow.md` → re-anchor the loop.
3. Skim the last 1–2 chapter notes + the tail of `_graph.md` → recover the thread.
4. Continue from NEXT. "Remember all my commandments to execute them" (Alma 18:10).
