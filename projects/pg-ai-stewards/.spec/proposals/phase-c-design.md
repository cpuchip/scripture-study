---
title: Phase C — Intent + Covenant as first-class substrate state
date: 2026-05-11
status: design sub-spec — ready for Phase C implementation
parent: full-agentic-substrate.md (D-C1..C4 ratification, plus 2026-05-11 re-validation)
purpose: >
  Make intent and covenant operational at the substrate layer. Today they live
  in YAML files (intent.yaml, .spec/covenant.yaml) — read by humans, recited by
  CLAUDE.md, but invisible to dispatched LLM calls. Phase C surfaces both into
  the system prompt of every dispatch and gives the gate something concrete to
  evaluate against ("does this honor the covenant?").
---

# Phase C — Intent + Covenant as first-class substrate state

## I. Binding problem

The substrate today dispatches work_items with no explicit awareness of:
- **Why** the work_item exists (its intent)
- **What we've committed to** in how the work gets done (the covenant)

Both exist in the repo (`intent.yaml`, `.spec/covenant.yaml`) and shape the agent's behavior at the harness level (CLAUDE.md preamble). But the substrate's own dispatched chats see neither directly. The model that drafts a study outline doesn't know whether the binding question came from a "deep transformation study" intent or a "quick reference lookup" intent. The gate that ratifies the outline can't ask "does this honor the covenant?" because the covenant isn't loaded.

This is the Klarna-failure shape: a perfectly disciplined agent succeeding at the wrong objective because the objective was implicit. Phase A added discipline within a work_item; Phase B added gates between maturities; neither addresses what the work_item is *for*.

## II. Success criteria

1. **Intent is required at work_item creation.** NewWork enforces. No NULL intent_id allowed.
2. **Every dispatched stage's system prompt** carries the active covenant's commitments and the work_item's intent purpose.
3. **The gate prompt** (Phase B's `evaluate` template) references intent: "does this output advance the stated intent? does it honor the covenant?"
4. **YAML stays canonical.** `intent.yaml` and `.spec/covenant.yaml` remain the source of truth. The substrate reflects them; humans edit YAML, not the database.
5. **A ratified change to a YAML file makes its way into the substrate within one git commit.** Pre-commit hook calls the seed SQL functions.
6. **A new top-level Stewards-UI route** lists active intents and the active covenant. Read-mostly; create/edit happens in YAML.

## III. Constraints and boundaries

**In scope:**
- `stewards.intents` and `stewards.covenants` tables
- `seed_intents_from_yaml(text)` and `seed_covenant_from_yaml(text)` SQL functions
- Git pre-commit hook that calls the SQL fns when YAML changes
- `compose_system_prompt` extension to inject active covenant + work_item intent
- Phase B `gate_prompts.evaluate` template revision to reference intent
- A new free-form covenant gate prompt (separate from `evaluate`) with **tools=off**
- `work_items.intent_id` FK column + NewWork form integration
- `/intents` and `/covenants` Stewards-UI routes

**Out of scope:**
- Multi-version covenants (versioning lives in git; the substrate carries the active one)
- Intent inheritance / nesting (each work_item has exactly one intent for now)
- Auto-derivation of intent from the binding question (that's a Phase F-ish feature)
- Per-stage intent overrides (intent is work_item-level)

## IV. Prior art

- **`intent.yaml`** at repo root: 68 lines. Carries `purpose`, `values` (5 keyed entries with description + source), `constraints` (4 keyed entries with severity + source + enforcement), `success_criteria` (5 strings), and the 11-step cycle commentary.
- **`.spec/covenant.yaml`** at repo root: 274 lines. Bilateral commitments — `human_commits_to` (5 entries), `agent_commits_to` (6 entries), `when_broken`, `recovery`, `council_moment`, plus a `teaching` extension with its own bilateral commitments.
- **`compose_system_prompt(agent_family, model, session_id)`** at `src/schema.rs:686` — already concatenates agent.prompt + matching instructions + skill XML block. Phase C extends this function, not replaces it.
- **`stewards.gate_prompts`** table (Phase 5a) — already holds three templates (`evaluate`, `generate_scenarios`, `verify`). Phase C adds a fourth (`covenant_check`) and revises `evaluate` to reference intent.
- **Phase B's gate-eval cost surprise (2026-05-11):** `evaluate_gate` ran through the `plan` agent which has tools enabled, leading to 5 tool-calling rounds before returning JSON (~5× cost). Phase C's covenant gate must avoid the same pattern by dispatching with tools disabled.

## V. Proposed approach

### V.1 Schema

```sql
-- Intents — reusable across work_items and pipeline_families.
CREATE TABLE stewards.intents (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text UNIQUE NOT NULL,
    purpose             text NOT NULL,
    beneficiary         text,
    values_hierarchy    jsonb NOT NULL DEFAULT '[]'::jsonb,  -- ordered list of trade-off priorities
    non_goals           text[] DEFAULT ARRAY[]::text[],
    scripture_anchor    text,                                 -- e.g. "D&C 88:118"
    source_file         text,                                 -- e.g. "intent.yaml"
    source_yaml_sha     text,                                 -- sha256 of YAML at last seed
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Covenants — typically one active row scoped 'global'; pipeline/work_item
-- scopes available for finer enforcement.
CREATE TABLE stewards.covenants (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope               text NOT NULL,        -- 'global' | 'pipeline:<family>' | 'work_item:<id>'
    human_commits_to    jsonb NOT NULL,       -- array of {key, description, why}
    agent_commits_to    jsonb NOT NULL,
    when_broken         text,
    recovery            text,
    council_moment      text,
    teaching_extension  jsonb,                -- optional Section 7 covenant
    activated_at        timestamptz NOT NULL DEFAULT now(),
    deactivated_at      timestamptz,          -- NULL = active
    ratified_by         text NOT NULL,        -- 'human' | 'agent' | 'both'
    source_file         text,
    source_yaml_sha     text,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX covenants_active ON stewards.covenants (scope) WHERE deactivated_at IS NULL;

-- work_items gains the FK
ALTER TABLE stewards.work_items
    ADD COLUMN intent_id uuid REFERENCES stewards.intents(id);
-- NOT NULL added in a follow-up migration AFTER existing rows are backfilled
-- with a "scripture-study" default intent (see V.5).
```

### V.2 Seed SQL functions

```sql
-- seed_intents_from_yaml: parses the YAML text, upserts rows by slug.
-- Intent.yaml today carries purpose + values map + constraints map + success_criteria.
-- We model the root as intent slug='scripture-study' with values_hierarchy =
-- the values map serialized; non_goals = success_criteria phrased as positive (not goals,
-- but the success-vs-failure boundary). Future: support multiple intents per file.
CREATE OR REPLACE FUNCTION stewards.seed_intents_from_yaml(p_yaml text)
RETURNS uuid
LANGUAGE plpgsql AS $$
DECLARE
    v_intent_id uuid;
    v_sha       text := encode(sha256(convert_to(p_yaml, 'utf8')), 'hex');
BEGIN
    -- Implementation: parse YAML via plpython3u OR ship a tiny Rust helper as
    -- a #[pg_extern]. plpython3u is simpler but adds an extension dep; the
    -- Rust path keeps the substrate self-contained. Recommend Rust helper.
    -- Function signature here is the contract; body uses parse_yaml_intent helper.
    --
    -- Upsert keyed on slug. Returns the intent_id.
    -- Idempotent: if source_yaml_sha matches, no-op early exit.
    ...
END;
$$;

-- Same shape for covenant. Source file = .spec/covenant.yaml. Always scope='global'
-- unless YAML specifies otherwise. Deactivates the prior 'global' row before
-- inserting new (atomic — single transaction).
CREATE OR REPLACE FUNCTION stewards.seed_covenant_from_yaml(p_yaml text) RETURNS uuid ...;
```

YAML parsing: write a small `parse_yaml_intent` / `parse_yaml_covenant` helper in `src/schema.rs` (or new `src/yaml.rs`) using `serde_yaml` (already pulled in for tool definitions). Expose as `stewards.parse_yaml_intent(text) RETURNS jsonb`. The seed functions pass the YAML through the parser then upsert from the resulting jsonb. Keeps the substrate self-contained — no plpython3u dependency.

### V.3 Pre-commit hook

`.git/hooks/pre-commit` (also tracked in `scripts/git-hooks/pre-commit` and symlinked):

```bash
#!/bin/bash
# Re-seed substrate when intent.yaml or .spec/covenant.yaml changes.
set -e

CHANGED=$(git diff --cached --name-only)

if echo "$CHANGED" | grep -q '^intent\.yaml$'; then
    echo "intent.yaml changed — reseeding stewards.intents"
    YAML=$(cat intent.yaml)
    docker exec pg-ai-stewards-dev psql -U stewards -d stewards \
        -c "SELECT stewards.seed_intents_from_yaml(\$\$${YAML}\$\$);"
fi

if echo "$CHANGED" | grep -q '^\.spec/covenant\.yaml$'; then
    echo ".spec/covenant.yaml changed — reseeding stewards.covenants"
    YAML=$(cat .spec/covenant.yaml)
    docker exec pg-ai-stewards-dev psql -U stewards -d stewards \
        -c "SELECT stewards.seed_covenant_from_yaml(\$\$${YAML}\$\$);"
fi
```

Single-quoted dollar-quoting protects against PowerShell's quoting issues (Lesson #4 from prior session). If the dev container isn't running, the hook prints a warning and continues — the seed can be re-run manually from the CLI.

### V.4 compose_system_prompt extension

Today `compose_system_prompt(agent_family, model, session_id)` returns:
```
<agent.prompt>
<matching instructions>
<available_skills XML>
```

After Phase C, when the session is tied to a work_item with an intent:
```
=== Active Covenant ===
<covenant.human_commits_to formatted as bullets>
<covenant.agent_commits_to formatted as bullets>

=== Intent ===
Purpose: <intent.purpose>
Beneficiary: <intent.beneficiary>
Values (in order):
  <intent.values_hierarchy formatted>
Non-goals:
  <intent.non_goals formatted>
Scripture anchor: <intent.scripture_anchor>

=== Agent ===
<agent.prompt>
<matching instructions>
<available_skills XML>
```

The intent block is appended only when `session_id` resolves to a work_item with `intent_id IS NOT NULL`. Sessions without work_items (ad-hoc chats, watchman) get only the global covenant + agent block. Token cost: ~300-600 added input tokens per dispatch. Acceptable; the cost discipline already in place will surface it if it spikes.

### V.5 Migration: backfilling existing work_items

Existing work_items have NULL intent_id. Phase C migration:
1. INSERT a default `scripture-study` intent (sourced from intent.yaml).
2. UPDATE all existing work_items SET intent_id = (default row).
3. ALTER TABLE work_items ALTER COLUMN intent_id SET NOT NULL.

This means existing soak work_items keep working. New work_items must pick (D-C3 ratified — friction is the discipline).

### V.6 NewWork form

Adds `intent_id` dropdown above the existing pipeline picker. Options:
- Active intents from `stewards.intents` (slug + truncated purpose)
- "Create new intent inline…" — opens a modal that posts to `/api/intents/create` (slug + purpose required, the rest optional). The created intent is selected automatically.

`new_work.go` extends `workItemCreateReq` with `IntentID string`. UPDATEs `work_items.intent_id` after `work_item_create()` (same pattern used for `destination_maturity` in Phase B push 3).

### V.7 Free-form covenant gate

Phase B's `evaluate` template revision:
```
You are evaluating the output of stage '{{current_stage}}' of pipeline
'{{pipeline_family}}', currently at maturity '{{maturity}}'.

The intent of this work is:
  Purpose: {{intent_purpose}}
  Values: {{intent_values}}
  Non-goals: {{intent_non_goals}}

The active covenant commits the agent to:
  {{covenant_agent_commits}}

The stage produced this output:
  {{stage_output}}

Decide one of: advance | revise | surface.
- advance: output meets the bar AND advances the stated intent AND honors the covenant
- revise: output meets the technical bar but misses something fixable; explain
- surface: output drifts from intent OR violates covenant — human review needed

Return JSON: {action, reasoning, feedback}. No tool calls; respond with JSON only.
```

The new `gate_prompts.covenant_check` template (separate from `evaluate`):
```
A work_item is about to advance to maturity '{{target_maturity}}'.

Active covenant commitments (agent):
{{covenant_agent_commits}}

Active covenant commitments (human):
{{covenant_human_commits}}

The work_item's stated intent:
{{intent_purpose}}

The work produced:
{{stage_output}}

Question: does this output honor the covenant? Specifically, does it
respect surface_tensions, read_before_quoting, check_existing_work,
honor_scope, exercise_stewardship?

Return JSON: {honors_covenant: bool, concerns: [string], recommendation: 'pass'|'flag'}.
No tool calls; respond with JSON only.
```

Both prompts dispatched with **tools=off** at the chat-payload level. This requires extending the chat dispatch to accept a `tools_disabled: true` payload flag, which `compose_system_prompt` honors by omitting the `<available_skills>` block.

### V.8 Stewards-UI

Two new top-level routes:
- `/intents` — list view of all intents with usage counts (how many work_items reference each); detail view shows the intent + linked work_items. Read-only; "edit" links to the YAML file path with instructions to commit + push.
- `/covenants` — single page showing the active global covenant rendered nicely (human + agent commitments side-by-side, when_broken / recovery / council_moment as expandable cards). Same edit-via-YAML pattern.

`api/intents.go` + `api/covenants.go` — list, get, create-inline (for NewWork's "create new intent" flow). No update endpoints (YAML is the editor).

## VI. Open questions / follow-ups

- **YAML parsing**: serde_yaml in pgrx — is it already in Cargo.toml or do we need to add it? Check before committing to the parsing approach.
- **Intent schema vs. intent.yaml shape**: today's intent.yaml has `values:` as a *map* (named keys with description + source). Proposal §V.1 says `values_hierarchy jsonb` (ordered list). Compromise: store the map in `values_hierarchy` as `[{key, description, source}, ...]` preserving order via array semantics.
- **Pre-commit hook discoverability**: a hook that requires the dev container to be running can fail silently if the container is down. Print loudly.
- **Session-without-work_item dispatches**: ad-hoc chats and watchman don't have an intent. Should they get the global covenant only, or should they also get a default "ad-hoc" intent? Recommend: covenant only; ad-hoc chats are by definition not goal-directed work.
- **How does compose_system_prompt know which is the active covenant?** Query: `SELECT * FROM stewards.covenants WHERE scope='global' AND deactivated_at IS NULL ORDER BY activated_at DESC LIMIT 1`. Cache for session lifetime to avoid the per-dispatch query.
- **Token cost of injection**: 300-600 tokens × every dispatch is non-trivial. Cost panel will show it. May want a `compose_system_prompt(skip_covenant=true)` option for stage chats that don't benefit from re-stating commitments mid-loop.

## VII. Estimated programming time

- V.1 schema + seed functions + YAML parser helper: 1 session
- V.3 pre-commit hook + V.4 compose_system_prompt extension + V.5 migration: 1 session
- V.6 NewWork form + V.7 gate prompt revisions + V.8 UI surfaces: 1 session

**Total: 3 sessions.** Matches the proposal's 2-3 estimate.

## VIII. Acceptance scenarios

- A new work_item cannot be created without `intent_id`. NewWork form requires it; `work_item_create()` raises if NULL; existing work_items have a default backfilled intent.
- A dispatched stage's system prompt (verifiable via `dry_run_chat`) starts with `=== Active Covenant ===` followed by `=== Intent ===` followed by the agent block.
- Editing `intent.yaml` and `git commit`-ing fires the pre-commit hook; subsequent `dry_run_chat` shows the updated intent values in the system prompt.
- A gate evaluation that returns `surface` with reasoning "this output meets technical criteria but doesn't advance the stated intent" appears in the gate decisions audit panel with the intent quoted.
- `/intents` Stewards-UI route lists current intents; clicking shows the YAML source path.
- A covenant `covenant_check` dispatch costs roughly the same as Phase B's gate-eval at parity (no 5× tool-loop blowout). Cost panel verifies.
