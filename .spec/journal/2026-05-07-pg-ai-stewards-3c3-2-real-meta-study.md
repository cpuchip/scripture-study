# pg-ai-stewards Phase 3c.3.1 + 3c.3.2 — the substrate writes its first real study

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

3c.3 v1 (the prior session) shipped the pipeline, ran end-to-end, and
surfaced three substrate bugs. This session: fix the bugs, re-run on
the same binding question, see what happens.

What happened: the substrate produced a real 6-section meta-study
with self-review revision notes, in 17 minutes 14 seconds, for
~$0.30 with prompt-cache pricing. The system did what it was built
to do.

## Bug fixes — Phase 3c.3.1

### Fix 1: `chat_post_internal` propagates `_*` markers

```sql
SELECT jsonb_object_agg(je.key, je.value) INTO v_inherited_markers
  FROM stewards.work_queue wq
  CROSS JOIN LATERAL jsonb_each(wq.payload) je
 WHERE wq.payload->>'session_id' = p_session_id
   AND wq.kind = 'chat'
   AND wq.id = (SELECT max(id) FROM stewards.work_queue
                 WHERE payload->>'session_id' = p_session_id AND kind = 'chat')
   AND je.key LIKE '\_%' ESCAPE '\';

v_payload := jsonb_build_object(...) || coalesce(v_inherited_markers, '{}');
```

Continuation chats now inherit any marker key starting with
underscore. Generic — works for `_watchman_pass_id`,
`_work_item_id`, `_stage_name`, `_pipeline_family`,
`_watchman_estimate`, and any future system that puts markers on
its first chat in a session.

### Fix 2: `v_is_final` coalesce

```sql
v_is_final := coalesce(
    (NOT v_has_tool_calls AND v_finish_reason IS NOT NULL
     AND v_finish_reason IN ('stop', 'length', 'content_filter'))
    OR (v_loop_stop IS NOT NULL
        AND v_loop_stop IN ('steps_exhausted', 'truncated_tool_calls')),
    false
);
```

Every clause guarded against NULL. If `v_loop_stop IS NULL`, the
whole expression collapses to a clean boolean. The 3c.2 trigger no
longer advances on intermediate chats.

### Fix 3: `agents.steps` 8 → 50

```sql
UPDATE stewards.agents SET steps = 50
 WHERE family NOT LIKE 'watchman%' AND steps < 50;
-- 21 rows updated; watchman variants (steps=1) preserved.
```

Watchman stays at steps=1 (single-shot, no tools by design).
Everyone else gets 50 to allow real tool-using research.

### Synthetic verification (zero tokens)

Marker inheritance proven:
- Inserted a fake chat row with `_test_marker_id`, `_other_field`,
  `normal_key`
- Called `chat_post_internal` for the same session
- New row inherited `_test_marker_id` and `_other_field`; did NOT
  inherit `normal_key`. Underscore filter works.

## Re-run — Phase 3c.3.2

Created `work_item ftc-wtl-meta-v2` with the same binding question:

> How do the triplets *Faith, Hope, and Charity* and *The Way, The
> Truth, The Life* interrelate? Are they the same concepts viewed
> from different angles, or genuinely different? Is one
> human-centered and the other Christ-centered? Are they the same
> point seen from different vantage points?

Budget: 2,000,000 tokens. Pipeline: study-write (outline → draft →
review). Provider: opencode_go. Model: kimi-k2.6.

### What happened, stage by stage

**Outline (plan agent, 9 chats)** — Started with broad searches:
"faith hope charity", "way truth life", "charity love of christ".
Got 8 hits across kinds. Read the most-relevant docs in full
(`way-truth-life`, `hope-and-the-grammar-of-pairs`, `enoch-charity`,
`faith-01`, `discernment-and-the-comprehending-eye`,
`tree-of-life-and-the-chain`, `plan-of-salvation`,
`charity`, `truth`, `best-books-and-the-spirit-of-discernment`).
Pulled `study_citations` to discover canonical scripture coverage
(John 14, Hebrews 6, 10, 11, Moroni 7, 10, D&C 88, 93, Romans 5,
Alma 32, 2 Nephi 31). Then synthesized a 5-section outline naming
which slugs each section would draw on and which canonical sources
those docs reference.

**Draft (study agent, 8 chats)** — Started with a self-imposed
critical-analysis pass before drafting (the agent did this on its
own initiative — the input_template didn't ask for it):
> *Critical Analysis Notes:*
> - The strongest structural link is Hebrews 10:20 — the veil as Christ's flesh, connecting "the way" directly to temple architecture. This is explicit text, not inference.
> - The "fruit is Christ Himself" reading from `tree-of-life-and-the-chain` openly admits: "The synthesis here... is not stated by any single verse... The mapping is built by reading the two systems against each other..." I need to carry that honesty forward.
> - Missing voice check: Do I have all five standard works? Yes...
> - Posture check: Am I claiming these triplets are identical? No...

Then produced "The Two Triplets: One Ascent, Two Descriptions" —
18,336 chars, 5 sections (Two Triplets as Ordered Progressions /
The Veil as Shared Hinge / Terminal Point / Perceiver and
Perceived / Synthesis).

**Review (study agent, ~5 chats including revision)** — Loaded its
own draft, ran voice + source-verification + binding-question
coverage checks. Returned `REVIEW: revised` with substantial
revision notes:

> *Voice / em-dashes:* Restructured two paragraphs that contained paired em-dashes (one per paragraph max). Removed meta-narrative frames ("Here is where the binding question finds its sharpest answer"). Removed the three-sentence closing refrain at the end.
>
> *Source verification:* Fixed four non-verbatim quotes:
> (1) changed "one assimilation, viewed at three stages" to the exact "three names for one assimilation, viewed at three stages";
> (2-3) removed quote marks from paraphrased passages;
> (4) removed "which is" from Romans 5:5 quote to match the retrieved source.
> Fixed "Mormon 7:48" typo to "Moroni 7:48."
>
> *Binding question:* Coverage remains complete — the study answers whether the triplets are synonymous (no), genuinely different (yes but complementary), us-centric vs. Christ-centered (neither; both are Christ-centered), and the same point from different vantage points (one ascent, two descriptions).

The agent voluntarily applied Michael's voice discipline rules
without the input_template prescribing them. The 17K-char `study`
agent prompt carried the discipline; the substrate just gave the
agent room to apply it.

### The synthesis the agent produced

> *These triplets are not the same point seen from different vantage points, yet they are not separate programs either. They are the instrument and the music, the eye and the light, the traveler and the road.*
>
> *The work this week is simple and severe: check the eye. Is it single? Is it aimed at Christ as prototype, or at something else — a political outcome, a family ideal, a personal project, even a theological system? Whatever the eye stays on, the body fills with. And what the body fills with is what the body will eventually comprehend. The two triplets agree on this completely. Faith, hope, and charity tune the instrument. The way, the truth, and the life are what the tuned instrument finally hears. The ascent is one, the descriptions are two, and the Person at the threshold is Christ.*

That is a real, defensible answer to the binding question. Saved to
[study/.scratch/ftc-wtl-meta-v2-review-output.md](../../study/.scratch/ftc-wtl-meta-v2-review-output.md).

### Numbers

| | |
|---|---|
| Elapsed | 17m 14s |
| Total chats | 18 (outline 9 + draft 8 + review ~5, with overlapping rows) |
| Tool calls | ~50 across all stages |
| tokens_in | 626,404 |
| tokens_out | 64,373 |
| Estimated $ | ~$0.30 with kimi-k2.6 cached pricing |
| Budget headroom | 65% (used 691K of 2M) |

Token rollup verified exact: `work_items.tokens_in/out` matches
`SUM(messages.tokens_in/out + reasoning_tokens)` across all three
session_ids. The bug 2/3 fixes are clean.

### Auto-advance fired correctly between stages

- 16:07:53: outline → draft (9 outline chats then advance)
- 16:11:58: draft → review
- 16:21:52ish: review terminal → status=completed

Each advance fired the trigger on the FINAL chat of the prior stage
(detected via `loop_stop_reason` or `finish_reason='stop'` + no tool
calls). The next stage's first chat dispatched immediately,
inheriting the work_item markers via the new
`chat_post_internal` propagation.

## What was surprising

**The agent self-imposed Michael's writing discipline.** The
`input_template` for the draft stage said "Quote text VERBATIM only
from study_get results." It said nothing about em-dash budgets or
the "and-then vs therefore/but" rule. The agent did the
critical-analysis pass on its own, did the voice check on its own,
applied the cite-count rule on its own. The 17K-char study agent
prompt carried all that discipline; the substrate just gave it
enough steps to apply it.

This is the architecture working as intended: the agent's
*persona* carries the standards; the substrate carries the *tools
and dispatch*. They compose without the substrate having to know
anything about scripture-study writing voice.

**Cached pricing is the difference between viable and not.** 626K
of input tokens at full price would be ~$1.25; at kimi-k2.6 cached
pricing, ~$0.05-0.10 of that 626K is the actual bill. Total run
~$0.30. Without caching, this work would cost meaningfully more.

**The agent corrected its own quotes.** The review stage caught
four non-verbatim quotes and fixed them by either re-retrieving the
exact text or converting to paraphrase. This is the
source-verification discipline actually being executed by the
substrate's review mechanism — not just claimed in the prompt.
Worth flagging because it's the most direct evidence that the
substrate's tool surface (`study_get` for re-retrieval) makes
honest writing easier than dishonest writing.

**The agent acknowledged its own epistemic limits.** From the
draft:
> "An honest caveat from `tree-of-life-and-the-chain`: 'No prophet writes the seed is faith, the tree is hope, the fruit is charity. The mapping is built by reading the two systems against each other and noticing they share vocabulary, motion, and end. That makes this a reading, not a doctrine.' The same holds for the larger synthesis between the two triplets."

The agent imported a calibrated-confidence frame from one of
Michael's own studies and applied it to its own claim. That kind of
intertextual humility is what the corpus was supposed to cultivate.

## What this means for the project

The substrate produced a real meta-study end-to-end without human
in the dispatch loop. The chain that works:

```
work_item created
  → input_template renders binding question
  → first chat enqueued with markers
  → agent runs tool loop (study_search_text, study_get, study_citations, etc.)
  → loop terminates with synthesized output
  → trigger advances stage, dispatches next stage's first chat
  → markers propagate through chat_post_internal
  → next stage's tool loop runs
  → ...
  → terminal stage completes
  → work_item status='completed'
  → human reads stage_results.review.output and judges
```

That's not theoretical anymore. We have an artifact.

The substrate is now USEFUL for the agentic real work Michael's
been heading toward since January. The pieces — agent corpus,
study tools, pipelines + work_items, auto-advance + token rollup +
budget gates — all compose. The next step depends on what we want
to do with this capability.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Read the produced study** — Michael's call on whether the synthesis holds up under his discernment. The substrate produced it; only he can say if it's good. |
| 2 | **Watchman soak start** — flip `schedule_enabled = true`. Independent of the 3c stack; we've been holding off because it costs tokens. Now we know the bgworker handles real load. |
| 3 | **3c.3.3** (if needed) — auto-promotion of completed work_items to `stewards.studies` rows. Currently the output lives in `stage_results`; auto-INSERT-as-study would close the loop on "the substrate writes its own studies." Speculative; only build if Michael wants pipeline-produced studies to participate in the corpus. |
| 4 | **3c.4** — HTTP tool registration for gospel-engine-v2 (Path A from earlier). Would let the agent quote scripture VERBATIM from canonical sources, not just from substrate study slugs. |
| 5 | **Image rebuild** — 14 SQL files folded; container has been live-applied through all. Rebuild before any container reset. The IDE flagged the `rust:1-bookworm` base image at 19 known vulns; bump to a current security tag during the rebuild cycle. |

## What's still solid

- The substrate's foundational decomposition (work_queue + dispatch
  + payload markers + triggers) handled a 3-stage / 18-chat /
  690K-token run cleanly. No bgworker errors. No stalled rows. No
  manual interventions.
- The agent corpus (3a.1) + study tools (3c.2.5) + pipelines (3c.1)
  + auto-advance (3c.2 with 3c.3.1 fixes) compose without further
  glue. Each layer does its job.
- The "files as projections" architecture means the produced study
  ALSO lands in `study/.scratch/` as a markdown file, ready for
  Michael to read and (maybe) promote. Substrate writes; human
  judges.
- Most importantly: the work the substrate produced honors Michael's
  voice and discipline. It made internal slug links; it acknowledged
  its synthesis was a "reading, not a doctrine"; it caught its own
  non-verbatim quotes; it ended with a Christ-centered framing. The
  Abraham 4-5 pattern Michael named in March — council, plan, watch,
  redemptive correction, rest and reflect — happened inside this
  pipeline. Not as decoration. As the actual mechanism.
