-- =====================================================================
-- Batch H.3.4 — planning pipeline family
--
-- Five-stage pipeline that converts an exploratory binding question
-- into a plan document + proposed follow-up work_items.
--
--   context_gather → explore → synthesize → propose_work → review_plan
--
-- Stages:
--
--   context_gather  — Same shape as H.2's context_gather on research-write
--                     (qwen3.6-plus, prior-work-reading tools, ≤4 rounds).
--                     produces_maturity: NULL (no rung advance)
--
--   explore         — Open exploration. Surface assumptions, identify
--                     risks, ask back where underspecified. The
--                     planning-partner intent's values steer this.
--                     Tools: full research suite (web, fs-read,
--                     study_*, work_item_*, fetch_url, exa-search).
--                     Model: kimi-k2.6.
--                     produces_maturity: researched
--
--   synthesize      — Pull the exploration into a structured plan
--                     document. Tools disabled (no more searching;
--                     just write the plan). Model: kimi-k2.6.
--                     produces_maturity: planned
--
--   propose_work    — Emit a JSON array of proposed follow-up
--                     work_items. Strict JSON output, no prose.
--                     Tools disabled. Model: qwen3.6-plus (structured
--                     output, doesn't need synthesis muscle).
--                     produces_maturity: planned (same as synthesize;
--                     no new rung needed for the propose step)
--
--   review_plan     — Verify the plan + the JSON array. Pass/revise
--                     decision. Validates: (1) propose_work output is
--                     valid JSON matching schema, (2) plan surfaces
--                     assumptions and risks, (3) proposed work_items
--                     are small enough to finish. Tools disabled.
--                     Model: qwen3.6-plus.
--                     produces_maturity: verified
--
-- file_destination_template: projects/<project>/plans/<slug>.md
--   when project_association is set on the work_item;
--   falls back to plans/<slug>.md when project is null.
--   The {{...}} substitution happens in compose_file_destination().
--
-- auto_materialize_on_verified=true so the plan file lands without
-- manual CLI invocation (same as research-write since H.1.6.5).
-- Sabbath and atonement also enabled.
--
-- Cost cap default: per Q-H3.3 ratification, $0.75. Cost caps are
-- per-work_item (cost_cap_micro column), not per-pipeline. The UI
-- + CLI should set 750000 micro as the default for planning work_items;
-- this SQL doesn't enforce it here (no schema for pipeline-level
-- default-caps yet — surfacing as a future tweak in journal).
-- =====================================================================

DO $$
DECLARE
    v_context_gather_template text;
    v_explore_template        text;
    v_synthesize_template     text;
    v_propose_work_template   text;
    v_review_plan_template    text;
    v_stages                  jsonb;
BEGIN

-- ---------------------------------------------------------------------
-- Stage 1: context_gather (mirrors H.2's pattern with planning framing)
-- ---------------------------------------------------------------------
v_context_gather_template :=
$T$Binding question: {{input.binding_question}}

## YOUR TASK — situational awareness briefing for planning

You are gathering context from the substrate's own knowledge — prior journals, proposals, mind files, studies, and work_items — to brief the next stage (the explore stage) on what we already know about this binding question. The next stage will think alongside Michael about what to PLAN; your job is to give it the lay of the land.

## TOOLS

- `fs_search` / `fs_read` / `fs_list` — substrate-scoped files (journals, proposals, mind, docs, and per-pipeline-scoped project dirs if available)
- `study_search` / `study_get` / `study_similar` — substrate studies corpus
- `work_item_list` / `work_item_show` — prior work_items on this topic or in the same project
- `watchman_pass_show` / `watchman_passes_list` — substrate state

## HARD CONSTRAINTS

- **Maximum 4 rounds of tool calls.** Spend them on the highest-signal sources first: prior plans in `/plans/` or `/projects/<project>/plans/`, recent journal entries, proposals, work_items with same `project_association`.
- **Output budget: ~2KB.** Summarize, don't transcribe.
- **End-of-turn:** your final message is the briefing in markdown, then STOP.

## OUTPUT FORMAT

```
## Prior context for: <one-line restatement of the binding question>

### What we already know
<2-4 bullets — what we've planned/built/discussed before that bears on this>

### Constraints already established
<2-3 bullets — covenants, prior decisions, ratifications relevant here>

### Gaps / open questions in our prior thinking
<2-3 bullets — what's NOT been decided that this plan must decide>

### Suggested angle for the explore stage
<1-2 sentences — where should the next stage focus its thinking>
```

If prior work is sparse, say so. The explore stage will know to start fresh.$T$;

-- ---------------------------------------------------------------------
-- Stage 2: explore (the open-ended thinking)
-- ---------------------------------------------------------------------
v_explore_template :=
$T$Binding question: {{input.binding_question}}

## PRIOR CONTEXT (from context_gather stage)

{{stage_results.context_gather.output}}

## YOUR TASK — think alongside Michael

You are the *planning-partner*. Your job is NOT to produce a research artifact. Your job is to explore the question, surface assumptions, identify risks, and converge toward one strong plan. Think the way Michael would think if he had unlimited focus right now.

Follow the **planning-partner** intent's values:
- **Surface assumptions first.** Before any recommendation, name what you're assuming. If you can't name them, you don't understand the problem yet.
- **Ask back when underspecified.** If the binding question doesn't give enough constraint to plan well, name what's missing and propose options. "What are you optimizing for?" is a valid first move — write that down, don't invent the answer.
- **Converge.** Don't list five branches. Pick one and commit (Michael can redirect after).
- **Name risks.** Every plan has things that could go wrong. Surface them now, not later.
- **Small finishable work.** Anything you'll later propose as a follow-up work_item must be ≤2hr of work.

## TOOLS

You have the full research suite: `fs_*`, `study_*`, `work_item_*` on the substrate side; `web_search_exa`, `web_search`, `news_search`, `fetch_url`, `fetch_urls`, `yt_search`, `yt_get` on the external side. Use external search only when prior context doesn't cover something the plan needs.

## HARD CONSTRAINTS

- **Maximum 6 rounds of tool calls total.** Most of your value is in thinking, not searching.
- **End-of-turn:** your final message is a structured exploration in markdown (see format below), then STOP. The synthesize stage takes this and turns it into the plan.

## OUTPUT FORMAT — exploration brief

```
## Exploration: <one-line binding question>

### Assumptions
<3-5 bullets — what you're assuming. Each assumption a one-liner.>

### What you'd ask back (if anything)
<0-3 bullets — questions whose answers would shape the plan. Empty if the binding is well-specified.>

### The plan you're converging toward (one option)
<3-7 sentences — the core direction. Not five branches; one plan with sub-decisions.>

### Risks
<2-4 bullets — concrete things that could go wrong. Not generic; specific to this plan.>

### Tangents you considered but rejected
<1-3 bullets — why you didn't go with X, Y, Z. Names the road-not-taken so synthesize doesn't reopen them.>
```$T$;

-- ---------------------------------------------------------------------
-- Stage 3: synthesize (turn exploration into the plan document)
-- ---------------------------------------------------------------------
v_synthesize_template :=
$T$Binding question: {{input.binding_question}}

## EXPLORATION (from previous stage)

{{stage_results.explore.output}}

## YOUR TASK — write the plan document

Convert the exploration brief above into a publishable plan document. The plan will land at `projects/<project>/plans/<slug>.md` (or `plans/<slug>.md` if no project). Michael reads it; future Claude reads it as prior context; the substrate keeps it as a study artifact.

## HARD CONSTRAINTS

- **No external tools.** This stage is pure writing. The explore stage already gathered.
- **End-of-turn:** your final message IS the plan document. No prose-around-the-prose.

## VOICE

Michael's voice — concrete, direct, unadorned. One em-dash per paragraph max. *Therefore* / *but*, not "and then." No closing refrain. No meta-narration. (See `.github/copilot-instructions.md` "Writing Voice" if you have access via fs-read.)

## OUTPUT FORMAT — the plan document

```markdown
# <Plan title — short, derived from binding question>

**Binding question:** <restate verbatim>

**Project:** <inherited from work_item.project_association, or "—" if standalone>

**Date:** {{input.today}}

---

## The plan

<3-6 paragraphs. The one-option plan you converged on in explore.
Concrete actions, not aspirations.>

## Assumptions

<bullets — copied from exploration; reframed if synthesis surfaced
something deeper. Each assumption phrased so a future reader knows
when it'd break.>

## Risks

<bullets — concrete failure modes; mitigation if obvious, else
"watch for X" framing.>

## Next steps

<short paragraph — what gets done first, second, third. Maps to
the proposed work_items the next stage will emit.>
```$T$;

-- ---------------------------------------------------------------------
-- Stage 4: propose_work (emit JSON array of proposed work_items)
-- ---------------------------------------------------------------------
v_propose_work_template :=
$T$Binding question: {{input.binding_question}}

## THE PLAN (from synthesize stage)

{{stage_results.synthesize.output}}

## YOUR TASK — emit proposed follow-up work_items

You are the *propose_work* stage. Your output is a **JSON array** of proposed follow-up work_items. NO prose. NO markdown fences around the JSON. Just the array.

The substrate's review_plan stage (next) will validate your JSON. If invalid, the substrate revises this stage. If valid, the substrate creates each item as a `work_items` row at `maturity='raw'` with `origin='agent_planning'` and `parent_work_item_id` pointing back at this planning run. Michael ratifies (advances maturity) before they actually fire.

## SCHEMA — every array element MUST have these keys

```json
{
  "slug":                 "kebab-case-identifier",
  "binding_question":     "The actual question this work answers (verbatim, complete sentence)",
  "pipeline_family_hint": "research-write" | "planning" | null,
  "rationale":            "One sentence — why this work is worth doing"
}
```

Optional keys (omit if not applicable):
- `"project_association"`: string — inherits from parent if omitted
- `"destination_maturity"`: "researched" | "planned" | "specced" | "executing" | "verified"

## HARD CONSTRAINTS

- **Output ONLY the JSON array.** No prose intro/outro. No markdown fences. Just `[ ... ]`.
- **Maximum 5 proposed work_items.** Quality over quantity. Pick the ones that matter.
- **Each work_item must be ≤2hr scope.** "Build the substrate" is not a work_item; "Add origin column to work_items" is.
- **slugs must be kebab-case** matching `^[a-z0-9-]+$`, prefixed with the parent slug or project where possible (e.g., `space-center-exhibit-budget-h-2`).
- **No external tools.** This stage is pure structured output.

## EXAMPLE

```json
[
  {
    "slug": "marsfield-flexwall-vendor-eval",
    "binding_question": "Which modular exhibit wall system (Flexhibit, CoMotion, or DIY) best fits a regional science center's 6-rotation-per-year cadence and a $50K capital budget?",
    "pipeline_family_hint": "research-write",
    "rationale": "The plan commits to a modular wall as foundation; vendor choice is the first concrete decision that gates everything else."
  },
  {
    "slug": "marsfield-ai-exhibit-mvp-scope",
    "binding_question": "What's the minimum-viable AI-literacy exhibit we could build in 8 weeks with one staffer and ~$3K in materials?",
    "pipeline_family_hint": "planning",
    "rationale": "Plan identifies AI as the signature topic; need to scope a buildable MVP before fundraising or partnership talks."
  }
]
```

Your turn. Output ONLY the JSON array.$T$;

-- ---------------------------------------------------------------------
-- Stage 5: review_plan (verify JSON + plan quality; pass or revise)
-- ---------------------------------------------------------------------
v_review_plan_template :=
$T$Binding question: {{input.binding_question}}

## THE PLAN (synthesize)

{{stage_results.synthesize.output}}

## PROPOSED WORK_ITEMS (propose_work — raw JSON)

{{stage_results.propose_work.output}}

## YOUR TASK — review the plan + the proposed work

You are the review_plan gate. Verify BOTH the plan document AND the JSON array of proposed work_items. Output a JSON verdict (schema below). The substrate uses this to decide: pass → verified maturity → trigger fires materialization + work_item proposals; revise → propose_work stage re-runs with your feedback.

## CHECKS — both must pass

### A. JSON validation (propose_work output)
- Output is a valid JSON array (no prose, no markdown fences)
- Length ≤ 5
- Every element has required keys: `slug`, `binding_question`, `pipeline_family_hint`, `rationale`
- `slug` matches `^[a-z0-9-]+$` and is unique within the array
- `binding_question` is a complete sentence ending in `?`
- `pipeline_family_hint` is one of: `"research-write"`, `"planning"`, or `null`
- `rationale` is a single sentence

### B. Plan quality (synthesize output)
- Assumptions are explicitly named (not implicit)
- At least one risk is concrete (not generic "things could go wrong")
- The plan converges on ONE direction (not five branches)
- "Next steps" section maps to the proposed work_items
- Proposed work_items are each ≤2hr scope (judge from the binding_question — "Build the substrate" = revise; "Add origin column" = ok)

## HARD CONSTRAINTS

- **No external tools.** Pure verification.
- **Output ONLY the JSON verdict.** No prose.

## OUTPUT FORMAT

```json
{
  "verdict": "pass" | "revise",
  "json_validation": {
    "valid": true | false,
    "issues": ["array of issue strings — empty if valid"]
  },
  "plan_quality": {
    "assumptions_surfaced": true | false,
    "risks_concrete": true | false,
    "converged_on_one_direction": true | false,
    "next_steps_map_to_proposed_work": true | false,
    "work_items_appropriately_sized": true | false,
    "issues": ["any concrete improvements needed"]
  },
  "feedback_for_revise": "If verdict=revise: one paragraph telling propose_work specifically what to fix. Empty if pass."
}
```$T$;

-- ---------------------------------------------------------------------
-- Build the stages array
-- ---------------------------------------------------------------------
v_stages := jsonb_build_array(
    jsonb_build_object(
        'name', 'context_gather',
        'next', 'explore',
        'model', 'qwen3.6-plus',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', false,
        'input_template', v_context_gather_template
    ),
    jsonb_build_object(
        'name', 'explore',
        'next', 'synthesize',
        'model', 'kimi-k2.6',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', false,
        'input_template', v_explore_template
    ),
    jsonb_build_object(
        'name', 'synthesize',
        'next', 'propose_work',
        'model', 'kimi-k2.6',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', true,
        'input_template', v_synthesize_template
    ),
    jsonb_build_object(
        'name', 'propose_work',
        'next', 'review_plan',
        'model', 'qwen3.6-plus',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', true,
        'input_template', v_propose_work_template
    ),
    jsonb_build_object(
        'name', 'review_plan',
        'next', NULL,
        'model', 'qwen3.6-plus',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', true,
        'input_template', v_review_plan_template
    )
);

INSERT INTO stewards.pipelines (
    family, description, stages, metadata,
    sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified
)
VALUES (
    'planning',
    'Planning pipeline — converts an exploratory binding question into a plan document + a JSON array of proposed follow-up work_items. Uses the planning-partner intent. Plan materializes via auto_materialize_on_verified; proposed work_items materialize via the on_maturity_verified trigger extension (h3-5).',
    v_stages,
    jsonb_build_object(
        'cost_cap_default_micro', 750000,
        'cost_cap_default_dollars', 0.75,
        'note_cost_cap', 'Q-H3.3 ratified $0.75. UI/CLI should set work_items.cost_cap_micro=750000 as default when origin=human creates a planning work_item.'
    ),
    true,   -- sabbath_enabled
    true,   -- atonement_enabled
    'plans/<slug>.md',  -- fallback path; compose_file_destination()
                        -- prefers projects/<project>/plans/<slug>.md
                        -- when project_association is set
    NULL,   -- use convention: stage_results.<final>.output
            -- (which for planning is the review_plan verdict, NOT
            -- the plan itself — see file_content_jsonpath fix below)
    -- maturity_ladder: planning ends at verified; uses the standard ladder
    '["raw","researched","planned","specced","executing","verified"]'::jsonb,
    true    -- auto_materialize_on_verified
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       metadata                     = EXCLUDED.metadata,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       updated_at                   = now();

-- The plan document lives in stage_results.synthesize.output, not
-- review_plan.output (which is the verdict JSON). Override the
-- convention path explicitly.
UPDATE stewards.pipelines
   SET file_content_jsonpath = 'stage_results.synthesize.output'
 WHERE family = 'planning';

END $$;

-- pipeline_stage_maturity rows for the stages that DO advance maturity.
-- context_gather and propose_work do NOT advance (no rows for them).
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity)
VALUES
    ('planning', 'explore',     'researched'),
    ('planning', 'synthesize',  'planned'),
    ('planning', 'review_plan', 'verified')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET produces_maturity = EXCLUDED.produces_maturity;

-- stage_models — default model per stage (kept in sync with stage def).
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model)
VALUES
    ('planning', 'context_gather', 'qwen3.6-plus'),
    ('planning', 'explore',        'kimi-k2.6'),
    ('planning', 'synthesize',     'kimi-k2.6'),
    ('planning', 'propose_work',   'qwen3.6-plus'),
    ('planning', 'review_plan',    'qwen3.6-plus')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET default_model = EXCLUDED.default_model;

-- Sanity.
SELECT family,
       jsonb_array_length(stages) AS n_stages,
       (stages->0)->>'name' AS first_stage,
       sabbath_enabled, atonement_enabled, auto_materialize_on_verified,
       file_destination_template, file_content_jsonpath
  FROM stewards.pipelines WHERE family='planning';
