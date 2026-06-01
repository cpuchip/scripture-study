# Book of Mormon Walk — Per-Chapter Workflow

Adapted from the `study-workflow` skill into a per-chapter loop. **Reload this at each book boundary and after any compaction** to stay anchored over the long run.

## What this walk is (calibrated 2026-06-01: heavier / mini-study depth)

Michael chose **mini-study depth per chapter** — richer over faster. So:

- **IS:** a start-to-finish walk where *each chapter gets a small study* — footnote chains followed across all volumes, real cross-referencing (`gospel_search` + our corpus), load-bearing word work, the critical-analysis discipline (strongest claims, counter-readings, named tensions), and a substantial analytical note. Plus the connection graph and the standout journal.
- **The one trim vs. a full study:** I don't run the full publish apparatus every chapter — no per-chapter voice-audit polish, no forced Becoming section, no finished-study ratification. The chapter note is a *mini-study + provenance*, not a publishable polished study. **Depth of engagement = study-grade; level of finish = note-grade.**
- Still **bin-1/2** (gather / draft-for-Michael) — exploratory, mine, quotes verified — never published doctrine without his ratification.
- **One-go intent:** the goal tool + ammon carry it to completion across many compactions. It is a long run by design; `_progress.md` is the thread that survives the resets.

## Honesty frame (non-negotiable)

- My notes are **exploratory digestion — clearly mine, never doctrine.** Speculation is labeled as speculation. Calibrated confidence: strong where the text is plain, tentative where I'm inferring.
- **Read before quoting.** Every quoted phrase is verbatim from the chapter text I read this session via `gospel_get`. If I haven't verified it, I paraphrase — no quotation marks.
- This is **bin-1/2 work** (gather / draft-for-Michael's-use) under `stuffy-in-the-loop`, authorized unsupervised. It never becomes a published study without his ratification.

## Cost discipline (the $18 zen budget)

- pg-ai-stewards is used **only as a retrieval / connection index**: `study_search`, `study_similar`, `study_get`. These are DB queries against the indexed corpus — **no opus, no cost.**
- **Never** spawn per-chapter LLM generation on the substrate (`panel_redline`, `start_brainstorm`, `spawn_subagent`, `consult_subagent`, `deep_research`, `audit_*`). The thinking is mine — Claude Code tokens, which are plentiful. The $18 is reserved for real study workflows.
- `brain_*` is **dead** — do not use. The substrate study tools replace it.

## The per-chapter loop

1. **READ** — `gospel_get "{Book} {N}"`. Read the full chapter, its heading, and the footnotes.
2. **FOLLOW** — trace the footnote chains that *add, clarify, or reframe* (deep-reading). Note the ones that matter; don't chase all of them.
3. **THINK (study it out — mini-study depth).** Genuinely work the chapter: what it says, what strikes me, the tensions, the counter-readings, the questions. Run the critical-analysis lens — strongest claims against the text, weakest links, missing voices, speculation vs. doctrine. Name tensions rather than resolving them. This is the heart; do it fully *before* moving on.
4. **CONNECT (cross-volume).**
   - `gospel_search` (semantic or hybrid) on the chapter's binding idea → non-obvious cross-refs across all five standard works.
   - Follow the footnote chains that *reframe* — into OT/NT/D&C/PGP, not just the nearby verses.
   - `study_search` / `study_similar` → links into our existing 198 studies (name the slug + why it connects).
   - `webster_define` on load-bearing words where the 1828 sense differs from modern.
   - `byu_citations` sparingly, where how the Brethren have used a verse changes the reading.
5. **VERIFY** — quotes verbatim against the text read; any count or claim traced to a tool call this session.
6. **WRITE** —
   - the chapter note → `study/bom-walk/{book}/{book}-{NN}.md` (format below). This note is *also* the scratch provenance.
   - extend `_graph.md` with new nodes/edges.
   - append to `_journal.md` **only if something genuinely stood out** (not forced every chapter).
7. **TRACK** — update `_progress.md` (mark done, advance the NEXT pointer). Commit per chapter or every 2–3 (the notes are heavier now) so any step walks back cleanly.
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
- words: webster_define where it earns it
**Entities:** people · places · doctrines · types/symbols · prophecies · covenants
**Edges:** {this} —[cross-ref|fulfillment|parallel|type→antitype|covenant-thread|links-to-study]→ {that}
**Notable / flag:** anything that "pops" worth Michael's eye
**Verified:** quotes checked against the chapter ✓
```

## Prophetic-quotation protocol (Isaiah blocks & long OT quotations)

For 2 Ne 12–24 (Isa 2–14), 1 Ne 20–21 (Isa 48–49), 3 Ne 22 (Isa 54), Jacob's Isaiah in 2 Ne 6–8, Abinadi on Isa 53 (Mosiah 14), and similar: **full exegesis AND Nephite framing.** Two layers, every such chapter:

1. **The text itself** — full prophetic-text depth: what Isaiah means, dual / latter-day fulfillment, Hebrew sense where it earns it, and the **Book-of-Mormon-vs-KJV textual variants** (the BoM Isaiah differs from the KJV in instructive places — note them).
2. **Why it's here — Michael's four questions, answered explicitly:**
   - Why did *this* writer (Nephi / Jacob / Abinadi / Christ) include *this* passage *here*?
   - Was it for us — the last-dispensation readers Nephi says he wrote for?
   - How does it apply to us today?
   - How do we read it *better* through the prophet's own framing?

   Anchor to the prophet's stated method — Nephi's likening (1 Ne 19:23; the interpretive keys in 2 Ne 25), Christ's command to search Isaiah (3 Ne 23:1). Verify and quote these when reached; here they are pointers, not quotations.

## Resume after compaction (ammon)

A 239-chapter run *will* cross context resets. On resume — do not restart:

1. Read `_progress.md` → the NEXT un-digested chapter.
2. Read this `_workflow.md` → re-anchor the loop.
3. Skim the last 1–2 chapter notes + the tail of `_graph.md` → recover the thread.
4. Continue from NEXT. "Remember all my commandments to execute them" (Alma 18:10).
