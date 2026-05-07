# pg-ai-stewards Phase 3c.3 v1 — first real pipeline run; three bugs found

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

3c.3 v1 — the first time the substrate actually ran a multi-stage
agent pipeline against a real binding question. Michael picked the
question himself: *"How do the triplets Faith/Hope/Charity and
The Way/Truth/Life interrelate? Are they the same concepts viewed
from different angles, or genuinely different? Is one human-centered
and the other Christ-centered?"*

What v1 was supposed to deliver: the templating + pipeline + a real
end-to-end run that we'd judge for quality.

What v1 actually delivered: the templating + pipeline + an
**instructive failure** that surfaced three substrate bugs in the
3c.2 auto-advance trigger. The agent worked. The substrate's
foundation is sound. The trigger needs fixes before 3c.3.2 can
produce a real study.

## What worked (genuinely encouraging)

**The agent used substrate tools correctly despite the imported
prompt.** The 17K-char `study` agent prompt was authored for
Copilot — it explicitly references `gospel_search`, `read_file`,
etc. Those tools don't exist in the substrate. The substrate has
`study_search_text`, `study_get`, `study_citations`, etc.

The agent figured it out from the `tools[]` JSON Schema
descriptions alone. Across three stages, it called:

- `study_search_text` with multi-kind queries
- `study_get` to pull bodies of relevant docs (charity,
  way-truth-life, enoch-charity, faith-and-the-grammar-of-pairs,
  discernment-and-the-comprehending-eye, tree-of-life-and-the-chain,
  plan-of-salvation, faith-1 lectures-on-faith, etc.)
- `study_citations` on multiple docs to discover canonical scripture
  references (John 14, Moses 6, Moroni 7, D&C 88, 1 Nephi 11,
  Hebrews 11, D&C 138)

The agent's research pattern was sound: broad search → read key
docs → trace citations. This is meaningful evidence the imported
agent corpus + substrate tool surface ARE compatible without
prompt rewriting.

**Stage input templating worked.** `render_stage_input` resolved
`{{input.binding_question}}` and `{{stage_results.outline.output}}`
correctly. The substitution layer is the right size; future
pipelines can use the same pattern.

**The pipeline + work_item dispatch infrastructure worked at the
mechanical level.** Stages enqueued, the bgworker dispatched, tool
calls round-tripped, continuation chats fired. All 33 work_queue
rows ran cleanly. The substrate's load-bearing parts are sound.

## What broke — three substrate bugs

### Bug 1: `v_is_final` evaluates to NULL when `loop_stop_reason IS NULL`

In `handle_work_item_chat_completion` (3c.2 trigger):

```sql
v_is_final := (NOT v_has_tool_calls
               AND v_finish_reason IN ('stop', 'length', 'content_filter'))
            OR v_loop_stop IN ('steps_exhausted', 'truncated_tool_calls');
```

When `v_loop_stop` is NULL (the normal case for an intermediate
chat with `finish_reason='tool_calls'`), `NULL IN (...)` returns
NULL, not FALSE. Then `FALSE OR NULL` is NULL. So `v_is_final`
becomes NULL.

Then:

```sql
IF NOT v_is_final THEN
    -- Intermediate. Wait for the next chat.
    RETURN NEW;
END IF;
```

`NOT NULL` is NULL. PL/pgSQL `IF` treats NULL as false → branch
not taken → function falls through and advances the work_item
incorrectly.

**Fix:** wrap the second clause in `coalesce(... , false)` or guard
with `IS NOT NULL` so NULL collapses to FALSE.

### Bug 2: continuation chats don't carry `_work_item_id` markers

The 3c.1 `work_item_dispatch_stage` injects `_work_item_id`,
`_stage_name`, `_pipeline_family` into the FIRST chat's
`work_queue.payload`. But after a tool_dispatch + continuation,
the bgworker re-enqueues a chat via `chat_post_internal`. That
function builds payload via:

```sql
v_payload := jsonb_build_object(
    'session_id', ..., 'agent_family', ..., 'requested_model', ...,
    'meta', ..., 'body', ...);
```

No marker propagation. So continuation chats have no
`_work_item_id` in payload.

Result: the 3c.2 trigger's WHEN clause `payload ? '_work_item_id'`
filters them out. The trigger only fires on the FIRST chat per
stage. The actual *final* chat in the loop (which has
`finish_reason='stop'` or `loop_stop_reason='steps_exhausted'`)
never reaches the trigger.

This is what made bug 1 visible. With marker propagation, the
trigger would fire on every chat in the loop and the NULL bug
would still exist — but the LAST fire would correctly identify
the final chat.

**Fix options for next session:**

A. Pass markers through `chat_post_internal` (add `p_meta jsonb`
   parameter; tool_dispatch continuation copies from parent).
B. `chat_post_internal` looks up the LATEST chat work_queue row
   for the session and copies any `_*_id` markers it finds.
C. Trigger does session-level lookup: when fired on any chat,
   queries the session's FIRST chat to get the markers.

Option A is most explicit. Option C is the smallest diff to the
existing code. I'd lean toward **A** — explicit beats implicit
for this kind of cross-cutting metadata.

### Bug 3: token rollup undercounts massively

The 3c.2 trigger reads the LATEST assistant message tokens at the
moment of trigger firing. With bug 2, the trigger only fires on
the FIRST chat per stage, so it only sees the FIRST iteration's
assistant message. Subsequent iterations grow the input context
substantially (5K → 50K+ tokens of accumulated context per
stage), but those tokens never get rolled into `work_items.tokens_in`.

In this run:

| Session | Assistant msgs | Sum input tok | Sum output tok |
|---------|----:|--------:|--------:|
| outline | 8 | 256,789 | 3,620 |
| draft | 8 | 184,814 | 1,389 |
| review | 2 | 12,545 | 1,095 |
| **total** | 18 | **454,148** | **6,104** |

The work_item showed `tokens_in=16,639, tokens_out=1,636` — under by
**>27×** on input. Roughly **460K actual tokens** vs the configured
200K budget. The budget gate never tripped because the rollup
that feeds it was wrong.

**Fix:** when bug 2 is fixed, the trigger fires per chat-iteration
and tokens roll up correctly per iteration. Budget gate then has
real data.

## Secondary finding: agent step budget is too tight

Even with bugs 1-3 fixed, the imported agents have `steps=8`
(default from 3a.1's import). With kimi calling 3-5 tools per
iteration on a real research task, 8 iterations isn't enough to
reach synthesis. The agent step-exhausted on every stage with
NO substantive content produced — every assistant message had
`finish_reason='tool_calls'` and the loop ran out before the
agent wrote a synthesis answer.

For the next run: bump `steps` for `study` and `plan` agents to
something like 20-30. Or set per-stage step budgets in the pipeline
definition (overriding the agent default).

This isn't a substrate bug — it's a parameter that needs tuning
based on the kind of work being asked. Note for 3c.3.1.

## Pipeline state right now

`work_item ftc-wtl-meta` shows `status='completed'` because the
buggy trigger advanced it. Its `stage_results` contains short
"I'll start by searching..." stubs — not real outputs. The session
message logs show the real work the agent did via tools. We're
keeping the work_item as-is for forensic value; 3c.3.2 will create
a fresh one after bugs land.

## What 3c.3 v1 actually delivers

- `extension/3c3-stage-templating-and-study-write.sql` — the
  templating infrastructure + study-write pipeline. Both work as
  designed.
- `lib.rs` + Dockerfile foldback (13th `extension_sql_file!`).
- The honest learnings catalogued here.
- A test fixture (`work_item ftc-wtl-meta`) we can re-use for
  regression testing once 3c.3.1 lands.

The substrate's templating layer + pipeline definition + agent
tool dispatch all work. The only thing between 3c.3 v1 and a real
study output is the trigger fix list above.

## Carry-forward (Phase 3c.3.1)

| Priority | Item |
|----------|------|
| 1 | **Fix bug 2 first** (markers propagate through continuations). Without it, bug 1 is half-visible and bug 3 stays. |
| 2 | **Fix bug 1** (NULL → FALSE in `v_is_final`). One-line change once bug 2 is in. |
| 3 | **Bump `steps` for plan/study agents** to ~20-30. Maybe via per-stage override in pipeline def rather than mutating agent rows. |
| 4 | **Re-run 3c.3.2** on the same FtC/WtL question with a fresh work_item. Now we should get a real outline → draft → review flow with synthesis output. |
| 5 | **Verify token rollup correct** post-fix: work_item.tokens_in/out should match `SUM(messages.tokens_in/out)` across all session_ids. |

## What's still solid

- The substrate's tool dispatch loop (Phase 1.6) handled 33
  work_queue rows + 44 tool replies across 3 sessions without a
  hiccup. The bgworker is rock-solid; this run was the most
  intensive load it's seen and it didn't blink.
- The agent corpus import (3a.1) plus tool registration (3c.2.5)
  plus tool perm broadcast produced an agent that uses substrate
  tools without prompt modification. That's a real architectural
  win — the imported prompts may reference tools that don't exist,
  but the model picks the right ones from the JSON Schema.
- The instructive-failure framing is honest. v1 of anything
  exists to surface what we don't know yet. We learned three
  concrete bugs and the agent-tool compatibility insight in a
  single ~$0.25 run.
