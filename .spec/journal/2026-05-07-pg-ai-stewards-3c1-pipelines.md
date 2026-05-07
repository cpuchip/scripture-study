# pg-ai-stewards Phase 3c.1 — pipelines + work_items orchestration

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

After 3a.1 (agent corpus imported) and the journal normalization
landed, the substrate had everything needed to actually orchestrate
agent work — agents in the registry, skills loadable, tool perms
encoded. What was missing was the *coordination* layer: how do you
chain agents across stages? How do you track work that takes
multiple model calls?

3c.1 ships that layer. Pipelines as immutable templates,
work_items as instances flowing through their stages.

The architectural decision came up first: should work_items be a
parallel dispatch primitive to `work_queue`, or a layer above it?
We chose option B (above). Same call we made for Watchman in 2.7b.1.
Michael's reasoning: "fundamentally it's the same flow right? at
this stage we're just working with LLMs for our AI and that's a
chat interface, and they'll have the same needs between both so
reusing makes sense to me." Right call. The bgworker stays generic;
work_items piggyback on the existing chat dispatch via payload
markers.

## What shipped

**Phase 3c.1 — pipelines + work_items orchestration above work_queue.**

### Files

- `extension/3c1-pipelines-work-items.sql` — new tables, transition
  functions, views, seed pipeline.
- `extension/src/lib.rs` — tenth `extension_sql_file!` reference.
- `extension/Dockerfile` — added `3c1-pipelines-work-items.sql` to
  the COPY directive.
- `cmd/stewards-cli/internal/show/pipelines.go` — new file with
  `PipelineList`, `PipelineShow`, `WorkItemCreate`, `WorkItemList`,
  `WorkItemShow`, `WorkItemDispatch`, `WorkItemAdvance`,
  `WorkItemCancel`, plus a shared `printWorkItemDetail` helper.
- `cmd/stewards-cli/main.go` — new top-level `pipeline` and
  `work-item` (alias `wi`) commands; `encoding/json` import added
  for the `--user-input` shorthand encoding.

### Schema

```sql
stewards.pipelines (
    family       text PRIMARY KEY,    -- ^[a-z0-9]+(-[a-z0-9]+)*$
    description  text,
    stages       jsonb NOT NULL,      -- [{name, agent_family, model, provider, next?, auto_advance?}]
    metadata     jsonb,
    created_at, updated_at
)

stewards.work_items (
    id              uuid PRIMARY KEY,
    slug            text UNIQUE,
    pipeline_family text REFERENCES pipelines(family),
    current_stage   text,
    status          text CHECK IN (pending, in_progress, awaiting_review,
                                   completed, failed, cancelled),
    input           jsonb,
    stage_results   jsonb,            -- {stage_name: {output, completed_at, tokens_in, tokens_out}}
    session_ids     text[],
    token_budget    int,
    tokens_in       int, tokens_out int,
    actor           text, error text,
    created_at, updated_at, completed_at
)
```

Plus two views: `work_items_active` (excludes terminal) and
`work_items_summary` (with stages_completed / stages_total).

### Transition functions

```
work_item_create(pipeline, input, slug?, actor?, budget?) → uuid
work_item_dispatch_stage(work_item_id, user_input?) → bigint (work_queue id)
work_item_advance(work_item_id, stage_output) → text (next stage or NULL)
work_item_fail(work_item_id, error) → void
work_item_cancel(work_item_id, reason?) → void
```

`work_item_dispatch_stage` mirrors the Watchman pattern from 2.7b.1
verbatim: builds the chat payload directly (not via `chat_enqueue`)
so it can inject `_work_item_id` / `_stage_name` / `_pipeline_family`
markers. The bgworker's existing chat dispatch loop is unchanged.

### Seed pipeline

```yaml
echo-test:
  - name: echo
    agent_family: stewards-explore
    model: kimi-k2.6
    provider: opencode_go
    next: null
    auto_advance: true
```

One stage. Smoke-test for the wiring. No real value beyond
verification.

### End-to-end smoke test (real model call)

```
1. work-item create --pipeline echo-test --slug smoke-test-2
                    --user-input "Reply with the single word 'ack'."
                    --actor verifier --budget 5000
2. work-item dispatch smoke-test-2
   → work_queue id=445, status=in_progress, payload contains
     _work_item_id, _stage_name='echo', _pipeline_family='echo-test'
3. bgworker drains the chat (kimi-k2.6 via opencode_go)
   → 2m52s elapsed (opencode was slow today as usual)
   → assistant "ack", 1935 tokens in / 30 out
4. work-item advance --output '{"output":"ack",...}' smoke-test-2
   → status=completed, stage_results.echo populated, completed_at set
```

The full cycle works. work_item carries through the pipeline; chat
dispatch is unchanged; the only Watchman-style addition is payload
markers.

## What was surprising

**Flag-stops-at-positional bit again.** Tried `work-item advance
smoke-test-2 --output '{...}'` first and got "id-or-slug required"
because Go's stdlib `flag` stops parsing at the first non-flag.
Fourth time this has bitten the project (per the 2026-05-04 journal
entry's carry-forward). cobra migration is the right answer when
bandwidth allows; for now, flags-before-positional is the workaround
documented in the help text.

**The `printWorkItemDetail` helper paid for itself immediately.**
WorkItemCreate, WorkItemAdvance, and WorkItemShow all use it; the
output is consistent across every command that mutates state. That
pattern (mutator → return updated detail view) makes CLI workflows
much more discoverable — the user sees the new state without having
to chase up with a separate `show` command.

**Token rollup is deferred to 3c.2.** The work_items table has
`tokens_in`/`tokens_out` columns but they stay 0 until 3c.2's
auto-advance trigger populates them by reading from `messages`
joined on `session_ids`. For 3c.1's manual advance, the user passes
the token counts inside the `--output` payload; that JSON lives in
`stage_results` for inspection. The aggregation column being unused
in 3c.1 is intentional — it lets 3c.2 own the trigger logic
end-to-end without 3c.1 setting up half a contract.

## What this unlocks

The substrate now has all three layers an agent platform needs:

1. **Capability registry** (3a.1) — agents, skills, tool perms.
   Answers: "what can this agent do?"
2. **Dispatch primitive** (Phase 1.5/1.6/2.7b.1) — chat work_kind,
   trigger-driven harvest. Answers: "execute one model call."
3. **Coordination layer** (3c.1) — pipelines + work_items.
   Answers: "execute a multi-step plan with state."

3c.2 closes the loop on (3) by making advance automatic — the chat
completion trigger advances the work_item rather than requiring a
manual `work-item advance` call. 3c.3 is the first real
demonstration: a study-write pipeline that orchestrates real
imported agents through real stages.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 3c.2** — `AFTER UPDATE` trigger on `work_queue` that auto-advances the work_item when its dispatched chat lands. Mirrors 2.7b.1's `handle_watchman_chat_completion`. Trigger reads `payload->'_work_item_id'`, finds the next stage from the pipeline definition, calls `work_item_advance` (or `work_item_dispatch_stage` for the next stage), updates token counters. |
| 2 | **Phase 3c.3** — first real pipeline. Probably `study-write`: outline (plan agent) → sources (study agent) → critical-analysis (critical-analysis skill) → draft (study agent) → review (study/talk agent). Real value, real agent corpus, real verification. |
| 3 | **Soak start** — flip `schedule_enabled=true` when ready. All gating items are clean. |
| 4 | **CLI flag-stops-at-positional papercut** — fifth occurrence noted today. Cobra migration when bandwidth allows. |
| 5 | **Image rebuild** — ten `extension_sql_file!` references now; never rebuilt since 2.7b.2. Live container has all SQL applied; rebuild before next deploy or container reset. |

## What's still solid

- The architecture pattern from 2.7b.1 (orchestration in SQL, payload
  markers, generic bgworker) generalized cleanly to 3c.1. We didn't
  have to invent new dispatch infrastructure; we extended payload
  shapes and added new transition functions. That's a sign the
  underlying decomposition is right.
- The CLI's `printWorkItemDetail` post-mutation pattern feels good
  enough to want to mirror in other CLI commands. Consider for
  future CLI work: every state-mutating command returns the updated
  detail view by default.
- `compose_system_prompt('stewards-explore', 'kimi-k2.6', 'wi--46906ee3--echo')`
  worked first try. The agent corpus imported in 3a.1 is being
  exercised by 3c.1's dispatch — first composition of the substrate's
  layers in production.
