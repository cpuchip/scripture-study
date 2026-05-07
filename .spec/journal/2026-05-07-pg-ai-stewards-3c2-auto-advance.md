# pg-ai-stewards Phase 3c.2 — auto-advance trigger

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

Two pieces in one session: (1) resolved the three open design
questions on the 3c.2.5 proposal (study tool registration); (2)
shipped 3c.2 — the auto-advance trigger that closes the loop on
3c.1's pipeline orchestration.

After 3c.2, work_items move through their stages without manual
intervention. The bgworker dispatches each stage's chat; the trigger
fires when the chat completes; the work_item advances. Same pattern
as Watchman's harvest trigger from 2.7b.1, retargeted from
`_watchman_pass_id` payload markers to `_work_item_id`.

## What shipped

### 3c.2.5 design decisions resolved

Per Michael's notes:

1. **`study_get` body inclusion:** YES default `include_body=true`
   (it's the point of the command). **Plus line-based pagination**
   — the agent reads docs in slices the way the Read tool does,
   with `body_line_offset` + `body_line_count` + `max_body_chars`
   safety cap. Defaults: 200 lines / 20K chars; "play with the
   limits once we have data."
2. **`study_search_text` multi-kind filter:** YES, `kinds` array
   (renamed from singular `kind`). Empty array means all kinds.
3. **Naming consistency:** YES, agent-tool name is `study_context_for`
   (anticipates `brain_context_for`, `todo_context_for`). The
   underlying SQL function stays `stewards.context_for` because the
   CLI + watchman_input already call it; renaming the SQL function
   would break both.

Proposal updated, committed `d978667`.

### 3c.2 — auto-advance trigger

**Files:**

- `extension/3c2-work-item-advance-trigger.sql` — the trigger
  function and the trigger itself.
- `extension/src/lib.rs` — eleventh `extension_sql_file!` reference.
- `extension/Dockerfile` — added the new SQL to the COPY directive.
- `extension/verify-3c2-inverse.sql` — 4-trial synthetic inverse
  hypothesis (zero model tokens spent during verification).

**The trigger logic** (in `handle_work_item_chat_completion`):

```
WHEN: NEW.kind='chat' AND NEW.payload ? '_work_item_id' AND
      NEW.status IN ('done','error') AND OLD.status DISTINCT FROM NEW

ON ERROR:
  → work_item_fail(id, formatted error)

ON DONE:
  → read latest assistant message for the session
  → roll up tokens_in/out into work_items (always — even on
    intermediate tool-loop iterations)
  → detect intermediate vs final:
      final = (no tool_calls AND finish_reason in stop|length|content_filter)
              OR (loop_stop_reason in steps_exhausted|truncated_tool_calls)
  → if intermediate: return (let the existing tool_dispatch loop run)
  → if final:
      build stage_output jsonb (output, model, tokens, finish_reason,
                                loop_stop_reason if present)
      call work_item_advance(id, stage_output)
        → if terminal: status=completed, return
        → else: status=pending (or awaiting_review if !auto_advance)
      check token_budget gate:
        → if (tokens_in+tokens_out) >= token_budget:
            status=awaiting_review with explanatory error,
            do NOT auto-dispatch
      auto-dispatch next stage:
        work_item_dispatch_stage(id)
        → on dispatch failure: status=awaiting_review with error,
          don't fail the whole work_item (prior stage's results valid)
```

Every record_*/advance/dispatch call wrapped in BEGIN/EXCEPTION
WHEN OTHERS → RAISE WARNING, so a bug in the harvester never breaks
the bgworker's underlying status flip.

### Verification — three layers

**Layer 1 — live end-to-end smoke test (real chat).** Created
`3c2-trigger-smoke` work_item on echo-test pipeline, dispatched.
Dispatched at 09:38:22, completed_at 09:38:27 — **5 seconds** end-
to-end (opencode_go's kimi-k2.6 was warm). Status=completed,
tokens 1935/75 rolled up automatically, stage_results.echo
populated with the assistant's "auto" response. Zero manual
advance calls.

**Layer 2 — synthetic inverse hypothesis (4 trials, zero tokens).**
A 2-stage `inverse-test-2stage` pipeline + a `pg_temp.fake_stage_completion`
helper that synthesizes user+assistant messages and a work_queue
row, then UPDATEs status='done' to fire the trigger.

| # | Setup | Result |
|---|-------|--------|
| 1 | trigger present, fake stage 1 completion | status=in_progress, current_stage=second, tokens 100/50 ✓ |
|   | + fake stage 2 completion | status=completed, tokens 180/90 ✓ |
| 2 | trigger DROPPED, fake completion | status=in_progress, current_stage=first, tokens 0/0 ✓ proves trigger is load-bearing |
| 3 | trigger restored, fake completion | advances again ✓ |
| 4 | budget=100, stage 1 spends 150 cumulative | status=awaiting_review, current_stage=second (advance fired), error="token budget exhausted at stage first (150/100); next stage second not auto-dispatched" ✓ |

**Layer 3 — token budget gate proven** by trial 4 specifically.
The trigger will not auto-dispatch when the cumulative spend has
crossed the per-work-item budget. Status=awaiting_review with a
human-readable error means the pipeline halts in a recoverable
state, not a failure.

## What was surprising

**The 5-second end-to-end was unexpected.** Recent opencode_go
chats had been taking 2-7 minutes (2.7b.1 5-doc verification: avg
~93s/doc; 3c.1 smoke test: 2m52s). This one came back in 5
seconds. Probably opencode_go's cache being warm + the response
being tiny ("auto" = 1 token). Worth noting: Watchman's per-pass
elapsed time is dominated by provider latency, which is
structurally variable. The bgworker's bookkeeping (status flips,
trigger firing, work_item updates) is sub-second.

**Token rollup includes reasoning tokens.** The trigger pulls
`tokens_in + tokens_out + reasoning_tokens` from the assistant
message and rolls them all into work_items.tokens_out. Kimi-k2.6
sometimes bills reasoning separately; the work_item's budget
counts everything as billable output. This matches the cost model
already established in Phase 1 step 7 (`messages.reasoning_tokens`)
and in Watchman.

**The intermediate-vs-final detection was the design's load-bearing
piece.** Without it, every tool-loop iteration would have advanced
the work_item, wreaking havoc. With it, work_items only advance
when the agent is *actually* done with the stage (clean stop, or
loop budget exhausted). The chat handler's `loop_stop_reason`
result field — added in Phase 1.6 — was already there waiting for
us to use; we just had to read it.

## What's now unblocked

Phase 3c.3 (first real multi-stage pipeline) is now blocked **only**
on 3c.2.5 (study tools). Once the agent has a real tool surface
(study_search_text, study_get, study_similar, study_citations,
study_context_for), an end-to-end multi-stage pipeline like
`study-write` becomes a real demonstration: dispatch → tools →
intermediate chats → tool replies → continuation → next-stage
auto-dispatch → terminal stage → completed work_item with all
stages' outputs in stage_results.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 3c.2.5** — register the 5 study tools (proposal already spec'd in detail; ~30 min in a fresh session per the estimate). |
| 2 | **Phase 3c.3** — first real multi-stage pipeline using the imported agent corpus. |
| 3 | **Soak start** — schedule_enabled=true. Independent of the 3c stack. |
| 4 | **Image rebuild** — eleventh `extension_sql_file!` now; container has been live-applied through all of them but never docker-rebuilt since 2.7b.2. |
| 5 | **Tool-call cost observation** for 3c.2.5's empty budget hooks — need a real pipeline run to know what numbers to populate. |

## What's still solid

- The 2.7b.1 → 3c.2 trigger pattern is becoming the canonical
  shape for "harvest something from work_queue completions." Both
  use AFTER UPDATE OF status with a WHEN-clause prefilter; both
  do their work in the same tx as the bgworker's status flip; both
  defensively wrap every side-effect call in BEGIN/EXCEPTION. The
  shape is reproducible — when 2.8 (LLM-inferred edges) ships it
  will likely use the same shape with `_edge_proposal_id` markers.
- The token rollup-on-every-chat (not just final) is the right
  call. It means a work_item's tokens_in/out is the live spend at
  any moment, regardless of where in the loop it is. The budget
  gate has accurate data to work with.
- Synthetic verification scaled cleanly. The
  `pg_temp.fake_stage_completion` helper is ~30 lines and gave us
  4 trials at zero token cost. That's the pattern for any
  trigger-driven mechanism going forward.
