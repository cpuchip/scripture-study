# Book of Mormon Walk — Per-Chapter Workflow

Adapted from the `study-workflow` skill into a per-chapter loop. **Reload this at each book boundary and after any compaction** to stay anchored over the long run.

## What this walk is (and isn't)

- **IS:** a start-to-finish digestion of the Book of Mormon, one chapter at a time — recording my own thoughts and connections, using the tools to surface links, building a knowledge graph we can pull from.
- **ISN'T:** 239 finished studies. Each chapter gets *genuine digestion* — read, follow the key footnotes, think it out, connect, note — not the full six-phase study apparatus. The unit of rigor is the chapter note, not a publishable study. A study goes deep on one binding question; this walk goes *wide and continuous*, and the depth lives in the connections and the honest thinking.

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
3. **THINK** — study it out. My own digestion: what strikes me, the tensions, the questions, how it connects to what came before in the walk. This is the heart of it — do it *before* moving to the next chapter.
4. **CONNECT** —
   - `gospel_search` (semantic or hybrid) on the chapter's binding idea → non-obvious scripture cross-refs.
   - `study_search` / `study_similar` → links into our existing 198 studies.
   - `webster_define` on load-bearing words where the 1828 sense differs from modern.
5. **VERIFY** — quotes verbatim against the text read; any count or claim traced to a tool call this session.
6. **WRITE** —
   - the chapter note → `study/bom-walk/{book}/{book}-{NN}.md` (format below). This note is *also* the scratch provenance.
   - extend `_graph.md` with new nodes/edges.
   - append to `_journal.md` **only if something genuinely stood out** (not forced every chapter).
7. **TRACK** — update `_progress.md` (mark done, advance the NEXT pointer). Commit per book (or every ~5 chapters) so any step walks back cleanly.
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

## Resume after compaction (ammon)

A 239-chapter run *will* cross context resets. On resume — do not restart:

1. Read `_progress.md` → the NEXT un-digested chapter.
2. Read this `_workflow.md` → re-anchor the loop.
3. Skim the last 1–2 chapter notes + the tail of `_graph.md` → recover the thread.
4. Continue from NEXT. "Remember all my commandments to execute them" (Alma 18:10).
