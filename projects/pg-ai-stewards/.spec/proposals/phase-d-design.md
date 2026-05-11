---
title: Phase D — Atonement + Sabbath + Consecration
date: 2026-05-11
status: design sub-spec — ready for Phase D implementation
parent: full-agentic-substrate.md (D-D1..D3 ratification, plus 2026-05-11 re-validation)
purpose: >
  Add the post-completion phase to the substrate. Today work_items succeed,
  fail, or quarantine and that's the whole story. Phase D adds three
  rituals that the gospel framework treats as essential but the substrate
  currently lacks: Atonement (extracting lessons from failure), Sabbath
  (marking the ending of completed work), Consecration (gating promotion
  on Sabbath having run).
---

# Phase D — Atonement + Sabbath + Consecration

## I. Binding problem

The substrate runs hot. A work_item completes (success or quarantine) and the next dispatch starts immediately. Nothing extracts what was learned. Nothing marks endings.

Two consequences observed during Phases A and B build:
1. **The same failure repeats across work_items.** Phase B's gate-eval discovered a structural blind spot in a study outline (Hebrew/Greek word study as a category error for an 1843 English revelation). That insight lives in `gate_decisions.reasoning` for one work_item, then evaporates. The next outline written in the same pipeline has no awareness of it.
2. **Endings aren't recorded.** A study reaching `verified` triggers `work_item_promote_to_study` and the work just… ends. There's no marker that says "this cycle closed; here's what it produced; rest before starting more." The gospel framework treats this rest as load-bearing, not optional.

Phase D introduces three substrate primitives:
- **Atonement** — a post-quarantine LLM dispatch that extracts proposed lessons; human ratifies before they enter `.mind/principles.md`.
- **Sabbath** — a pre-promotion LLM dispatch that produces a structured reflection journaled to the substrate.
- **Consecration** — `work_item_promote_to_study` becomes gated on Sabbath having run for sabbath-enabled pipelines.

## II. Success criteria

1. **Sabbath fires at the right moment.** A work_item reaching `maturity='verified'` on a sabbath-enabled pipeline triggers a Sabbath dispatch within 30s. Reflection is journaled to `stewards.lessons` (kind='sabbath_reflection') and visible in the Sabbath Log UI within 60s.
2. **Sabbath blocks promotion.** `work_item_promote_to_study` raises if the work_item's pipeline is sabbath-enabled and no Sabbath reflection exists. The block is intentional — the discipline is recording the ending.
3. **Atonement fires on quarantine.** A quarantined work_item triggers an Atonement dispatch within 60s. Output is parsed to one or more rows in `stewards.lessons` (kind='principle' | 'decision' | 'lesson'), all with `ratified_at IS NULL`.
4. **Tools are off** for both Sabbath and Atonement dispatches (same fix Phase C applies to covenant-check). Both are structured-output prompts; no research loops.
5. **Lessons accumulate.** `stewards.lessons` grows over time; pipeline-keyed views surface patterns. Phase E will consume these in retry context.
6. **Stewards-UI surfaces ratification.** Unratified lessons appear in a dedicated review panel; a single click promotes a lesson to ratified, optionally with destination = `.mind/principles.md` or `.mind/decisions.md`.

## III. Constraints and boundaries

**In scope:**
- `stewards.lessons` table (audit-ledger shape mirroring `gate_decisions`)
- `stewards.sabbath_dispatch(work_item_id)` SQL function — enqueues a Sabbath chat with `_sabbath=true` marker
- `stewards.atonement_dispatch(work_item_id)` SQL function — enqueues an Atonement chat with `_atonement=true` marker
- `apply_sabbath_result` and `apply_atonement_result` SQL functions
- bgworker auto-fire extension (two more markers added to the existing 3-marker switch from Phase B)
- `pipeline_families` config column for sabbath_enabled (default per family — study/lesson/talk = true; debug/dev = false)
- `work_item_promote_to_study` revision — checks sabbath gate
- New `gate_prompts` rows for `sabbath` and `atonement` templates
- Stewards-UI Sabbath Log + Lessons Review surfaces

**Out of scope (explicitly):**
- Auto-curation (lessons → `.mind/principles.md` without human ratification — D-D3 ratified human curation)
- Lesson de-duplication (the same insight surfacing twice is itself a signal)
- Sabbath dispatch on intermediate maturities (verified-only — the proposal is firm)
- Atonement on every failure (only quarantine; in-flight failures are Phase A's territory)
- Cost-cap enforcement on Atonement specifically (uses the existing per-work_item cap)

## IV. Prior art

- **Phase 5a `gate_decisions`** — append-only audit ledger with `at`, `action`, `reasoning`, `feedback`, `work_id`, `revision_count`, `raw_response`. Stewards-UI already renders this in WorkItemDetail. `stewards.lessons` mirrors this shape (D-D3 reconfirmed 2026-05-11).
- **Phase 5b bgworker auto-fire** — three markers (`_gate_eval`, `_scenarios_gen`, `_verify`) trigger automatic apply functions after chat completion. Phase D adds two more (`_sabbath`, `_atonement`) using the same pattern.
- **`work_item_promote_to_study`** — exists, called explicitly. Phase D adds a precondition check; existing call sites need no change unless the work_item's pipeline is sabbath-enabled.
- **`stewards.gate_prompts`** — Phase 5a/5b seeded `evaluate`, `generate_scenarios`, `verify`. Phase C adds `covenant_check`. Phase D adds `sabbath` and `atonement`. Five total templates by end of D.
- **Phase B's tools-cost lesson (2026-05-11)** — gate-eval through the `plan` agent with tools enabled cost ~5× because the model researched before deciding. Sabbath and Atonement are JSON-output prompts; tools=off avoids the same blowout.

## V. Proposed approach

### V.1 Schema

```sql
-- pipeline_families config table (or column on existing pipelines table)
ALTER TABLE stewards.pipelines
    ADD COLUMN IF NOT EXISTS sabbath_enabled boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS atonement_enabled boolean NOT NULL DEFAULT false;

UPDATE stewards.pipelines SET sabbath_enabled = true
 WHERE pipeline_family IN ('study-write', 'study-write-qwen', 'lesson', 'talk');

-- atonement is opt-in; nothing flipped on by default in this migration.

-- Audit-ledger lessons table (mirrors gate_decisions shape per 2026-05-11 ratification)
CREATE TABLE stewards.lessons (
    id              bigserial PRIMARY KEY,
    work_item_id    uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    at              timestamptz NOT NULL DEFAULT now(),
    kind            text NOT NULL CHECK (kind IN
                        ('principle', 'decision', 'lesson', 'sabbath_reflection')),
    content         text NOT NULL,
    raw_response    jsonb,
    ratified_at     timestamptz,
    ratified_by     text,                 -- 'human' typically
    promoted_to     text,                 -- '.mind/principles.md' | '.mind/decisions.md' | NULL
    work_id         bigint                -- the work_queue id of the dispatch
);

CREATE INDEX lessons_at         ON stewards.lessons (at);
CREATE INDEX lessons_work_item  ON stewards.lessons (work_item_id);
CREATE INDEX lessons_unratified ON stewards.lessons (ratified_at) WHERE ratified_at IS NULL;
CREATE INDEX lessons_kind       ON stewards.lessons (kind);

-- Aggregation view for Phase E's retry context (last 3 ratified per pipeline+stage)
CREATE VIEW stewards.lessons_recent_ratified AS
SELECT l.*, wi.pipeline_family, wi.current_stage
  FROM stewards.lessons l
  JOIN stewards.work_items wi ON wi.id = l.work_item_id
 WHERE l.ratified_at IS NOT NULL
   AND l.kind IN ('lesson', 'principle')
 ORDER BY l.at DESC;
```

### V.2 Sabbath dispatch + apply

`stewards.sabbath_dispatch(work_item_id uuid) RETURNS bigint` — pattern mirrors Phase 5b `verify_work_item`:
1. Read work_item; verify maturity='verified' and pipeline.sabbath_enabled.
2. Create a session `wi--<short-id>--sabbath--<epoch>` with `kind='sabbath'` (extend `sessions_kind_check` to add 'sabbath' — same pattern as 5c added 'gate').
3. Render the `sabbath` template (see V.4) with the work_item's input + final stage_results.
4. INSERT into work_queue with payload markers: `_work_item_id`, `_sabbath=true`, `tools_disabled=true`.
5. Return work_queue id.

`stewards.apply_sabbath_result(work_item_id uuid, result_jsonb jsonb, work_id bigint) RETURNS bigint` — INSERT into `stewards.lessons` (kind='sabbath_reflection', content=result.reflection, raw_response=result, work_id). Returns lesson id. Marks the work_item with `sabbath_completed_at` (new column).

### V.3 Atonement dispatch + apply

`stewards.atonement_dispatch(work_item_id uuid) RETURNS bigint` — mirrors sabbath_dispatch but:
1. Verify quarantined_at IS NOT NULL (or status='quarantined').
2. Render `atonement` template with full stage_results history + failure reasons + steward_actions log.
3. Marker `_atonement=true`, tools_disabled=true.

`stewards.apply_atonement_result(work_item_id, result_jsonb, work_id) RETURNS int` — Atonement returns:
```json
{
  "principles_to_record": ["...", "..."],
  "decisions": ["..."],
  "lessons": ["..."]
}
```
Function INSERTs one `stewards.lessons` row per item across the three arrays, kind matching the array name. Returns count inserted. All rows land with `ratified_at IS NULL` (D-D3).

### V.4 Templates

`gate_prompts.sabbath`:
```
A work_item just reached verified maturity. Mark its ending with a structured
reflection. This is not more work — it is the recording of an ending.

Pipeline: {{pipeline_family}}
Intent: {{intent_purpose}}
Binding question: {{input_binding_question}}
Final output (truncated): {{stage_results_summary}}

Reflect on:
- What did this work produce that you did not expect at the start?
- What got harder than predicted? What got easier?
- What pattern would you carry forward to the next work in this pipeline?
- What is the one sentence the human should remember from this work?

Return JSON: {reflection: string, carry_forward: string, surprise: string}.
No tool calls. JSON only.
```

`gate_prompts.atonement`:
```
A work_item was quarantined after {{failure_count}} failures. Walk back through
what was tried, what failed, what was eventually completed (or not), and propose
lessons that should outlive this work_item.

Pipeline: {{pipeline_family}}
Intent: {{intent_purpose}}
Failure history:
{{steward_actions_summary}}
Final state: {{quarantine_reason}}

Distinguish three kinds of takeaways:
- principles: enduring insights about HOW the work should be done (candidate for .mind/principles.md)
- decisions: specific choices made about THIS pipeline/stage that should be recorded (candidate for .mind/decisions.md)
- lessons: ephemeral observations relevant only for similar future work (substrate-only)

Return JSON: {principles_to_record: [string], decisions: [string], lessons: [string]}.
Be sparse. Three lessons that survive scrutiny beat thirty that get pruned.
No tool calls. JSON only.
```

### V.5 bgworker auto-fire extension

`src/bgworker.rs` already inspects payload for `_gate_eval`, `_scenarios_gen`, `_verify`. Add `_sabbath` and `_atonement`. The dispatch shape:
```rust
if let Some(wi_str) = wi_opt {
    if is_gate_eval || is_scenarios_gen || is_verify || is_sabbath || is_atonement {
        let parsed = parse_gate_response(work_id);
        match (is_gate_eval, is_scenarios_gen, is_verify, is_sabbath, is_atonement) {
            (true, _, _, _, _) => apply_gate_decision(...),
            (_, true, _, _, _) => apply_scenarios_result(...),
            (_, _, true, _, _) => apply_verify_result(...),
            (_, _, _, true, _) => apply_sabbath_result(...),
            (_, _, _, _, true) => apply_atonement_result(...),
            _ => unreachable!(),
        }
    }
}
```
Cleaner alternative once the 5-way matrix lands: enum in payload (`_kind: 'gate' | 'scenarios' | 'verify' | 'sabbath' | 'atonement'`). Refactor when adding the 6th.

### V.6 Promotion gate

`work_item_promote_to_study(work_item_id)` revision:
```sql
-- Existing precondition check at the top
PERFORM 1 FROM stewards.work_items WHERE id = $1 AND maturity = 'verified';
IF NOT FOUND THEN RAISE EXCEPTION ...;

-- NEW: sabbath gate
IF EXISTS (
    SELECT 1 FROM stewards.pipelines p
      JOIN stewards.work_items wi ON wi.pipeline_family = p.pipeline_family
     WHERE wi.id = $1 AND p.sabbath_enabled = true
) AND NOT EXISTS (
    SELECT 1 FROM stewards.lessons l
     WHERE l.work_item_id = $1 AND l.kind = 'sabbath_reflection'
) THEN
    RAISE EXCEPTION 'sabbath required before promotion: dispatch via stewards.sabbath_dispatch(%) first', $1;
END IF;
```

This is the deliberately blocking version (D-D-Sabbath-blocker ratified 2026-05-11). The exception message points the human at the fix.

### V.7 Triggering

Sabbath fires when maturity transitions to 'verified'. Two paths:
1. `apply_gate_decision` for action='advance' that lands on verified — directly enqueue sabbath.
2. `apply_verify_result` (Phase 5b) when all_passed=true and the verify_results row was the verified-stage one — directly enqueue sabbath.

Atonement fires when a work_item transitions to quarantined. The steward (Phase 4a `steward_tick`) is the one that quarantines. Add a one-liner at the quarantine point:
```sql
PERFORM stewards.atonement_dispatch(v_work_item_id) WHERE pipeline.atonement_enabled;
```

### V.8 Stewards-UI

Two new surfaces:
- **Sabbath Log** (top-level route `/sabbath`): chronological list of recent sabbath_reflection rows. Each card shows pipeline, slug, reflection text, carry_forward sentence, surprise. Links to the work_item.
- **Lessons Review** (top-level route `/lessons` or panel on the dashboard): unratified lessons grouped by kind. Each row has Approve / Approve & promote to .mind/X.md / Reject buttons. Approve sets `ratified_at`, `ratified_by`. Promote also writes the lesson text to the .mind/ file (append with timestamp).

Backend additions to `api/`:
- `api/lessons.go` — list (with `?kind=`, `?ratified=`), get, ratify (POST), reject (POST).
- `api/sabbath.go` — list reflections (with optional `?pipeline=`).

## VI. Open questions / follow-ups

- **Promotion-to-.mind/ mechanism.** "Approve & promote" writes to `.mind/principles.md`. The substrate doesn't have file-write capability today. Two options: (a) write via host-mount + plpython3u, (b) emit a "pending file write" record that a sidecar (or next git commit) materializes. Recommend option (b) — keeps substrate stateless on the file system.
- **What's the right cap on Atonement length?** A quarantined work_item with 30 failures might overflow the prompt. Truncate steward_actions to last 20 entries; full history available on follow-up.
- **Sabbath on already-promoted work_items.** Existing verified-and-promoted work_items predate Phase D. Backfill skipped — Sabbath only fires going forward. Document.
- **Atonement cost guardrail.** Even tools-off, Atonement on a long failure history is one of the larger prompts in the substrate. Worth a separate cost cap? Or trust the per-work_item cap?
- **De-duplication signal.** If the same lesson surfaces 5 times across pipelines, that's a signal worth surfacing. Phase E may want to track this; Phase D just records.
- **Cycle interaction with Phase E.** E's retry composer pulls last 3 ratified lessons. If Atonement happens on quarantine and lessons aren't ratified for hours/days, retry context is empty. Acceptable — ratified means the human stood behind the lesson; un-ratified shouldn't influence future dispatches.

## VII. Estimated programming time

- V.1 schema + V.2/V.3 dispatch + apply functions + V.4 templates: 1 session
- V.5 bgworker auto-fire + V.6 promotion gate + V.7 triggering + smoke test: 1 session
- V.8 Stewards-UI surfaces (Sabbath Log + Lessons Review + APIs): 1 session

**Total: 3 sessions** (one over the proposal's 2-session estimate; the UI surface accounts for the extra session).

## VIII. Acceptance scenarios

- A study-write work_item completes Phase B's verify stage with all_passed=true. Sabbath dispatch fires within 30s; reflection appears in `/sabbath` UI within 60s.
- An attempt to call `work_item_promote_to_study` on that work_item before sabbath completes raises `sabbath required before promotion`.
- After sabbath completes, `work_item_promote_to_study` succeeds and the study appears in `study_search` MCP.
- A quarantined work_item with `atonement_enabled=true` triggers an Atonement dispatch; 3 unratified lessons appear in `/lessons` UI within 60s.
- Human clicks Approve on one lesson; `ratified_at` set; lesson now visible in `lessons_recent_ratified` view.
- Phase E's retry composer (when built) pulls that lesson into the next retry's context for the same pipeline+stage.
