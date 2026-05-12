-- =====================================================================
-- Batch H.2 — context_gather stage as research-write's new first stage
--
-- H.1.7 gave the gather stage tools to consult prior work mixed with
-- external research. H.2 splits that into a distinct stage so:
--   (a) context-gathering and external search can use different models
--       (context_gather on cheaper qwen3.6-plus; gather stays on kimi-k2.6)
--   (b) the briefing is inspectable via stage_results.context_gather.output
--   (c) future pipelines (H.3 planning, future yt pipelines) can reuse
--       the exact same context_gather stage shape
--
-- Stage definition:
--   - name: context_gather
--   - next: gather
--   - agent_family: research (reuses tool grants from H.1.7)
--   - model: qwen3.6-plus (cheaper; context-gathering is structured, not creative)
--   - provider: opencode_go
--   - auto_advance: true
--   - tools_disabled: false (needs fs-read + study_search etc.)
--
-- Gather stage's input_template is rewritten to:
--   (a) drop the "CONSULT PRIOR WORK FIRST" section (context_gather owns
--       that job now)
--   (b) prepend the context_gather output as a "## PRIOR CONTEXT" section
--   (c) lower the round budget back from 8 to 5 (gather now only does
--       external search since prior-work reading happened in context_gather)
--
-- pipeline_first_stage_name(family) reads stages->0->>name, so prepending
-- context_gather makes it the new first stage automatically. Existing
-- work_items in flight retain their current_stage value; only NEW work_items
-- start at context_gather.
-- =====================================================================

DO $$
DECLARE
    v_pipeline    stewards.pipelines%ROWTYPE;
    v_new_stages  jsonb;
    v_new_gather_template text;
    v_context_gather_template text;
    v_context_gather_stage jsonb;
BEGIN
    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family='research-write';
    IF v_pipeline.family IS NULL THEN
        RAISE EXCEPTION 'H.2: research-write pipeline not found';
    END IF;

    v_context_gather_template :=
$T$Binding question: {{input.binding_question}}

## YOUR TASK — situational awareness briefing

You are gathering context from the substrate's own knowledge — prior journals, proposals, mind files, studies, and work_items — to brief the next stage (the external-research gather stage) on what we already know about this binding question. Your output is NOT the final research piece. It is a *briefing* the next stage reads before doing external search.

## TOOLS

You have:
- `fs_search` (regex search across `.spec/journal/*`, `.spec/proposals/*`, `.mind/*`, `docs/**`)
- `fs_read` (read a file in full)
- `fs_list` (list files matching a glob)
- `study_search` (substrate's studies corpus — gospel + research + planning)
- `study_get` (read a study by slug)
- `study_similar` (related studies via embedding edges)
- `work_item_list` / `work_item_show` (prior work_items on this binding)

## HARD CONSTRAINTS

- **Maximum 4 rounds of tool calls.** Spend them on the most likely prior-work sources first (journals named for the topic; proposals; mind files like `.mind/active.md`, `.mind/principles.md`).
- **Output budget: ~2KB.** Summarize, don't transcribe. The gather stage reads your briefing in addition to its own template; keep it tight.
- **End-of-turn:** your final message is the briefing in markdown, then STOP.

## OUTPUT FORMAT — the briefing

```
## Prior context for: <one-line restatement of the binding question>

### What we already know
<2-4 bullets: the most relevant prior journals/proposals/studies and what they say>

### Gaps in our prior work
<2-3 bullets: what the prior work does NOT cover that the binding question needs>

### Suggested external-search angle for the next stage
<1-2 sentences: where the gather stage should focus its external search to fill the gaps>
```

If prior work is sparse or absent (e.g., this is a brand-new topic for us), say so explicitly — "We have no prior journals or proposals on X" — and the gather stage will know to start fresh externally.$T$;

    v_new_gather_template :=
$T$Binding question: {{input.binding_question}}

## PRIOR CONTEXT (from context_gather stage)

{{stage_results.context_gather.output}}

## YOUR TASK

Given the prior context above, find external sources to fill the gaps and answer what prior work doesn't cover. Then **STOP**, produce the sources brief, and end your turn.

## HARD CONSTRAINTS

- **Maximum 8 strong sources** in the final brief. The prior context above counts as 0 of those — your job is the EXTERNAL sources.
- **Maximum 5 rounds of tool calls.** Cast wide early, narrow with `fetch_url` on high-value hits.
- **End-of-turn:** your final message is the sources brief in markdown. No further tool calls.

## TOOL GUIDANCE

You have `web_search_exa` (Exa neural search), `web_search` (DuckDuckGo), `news_search`, `fetch_url`, `fetch_urls`, `yt_search`, `yt_get`, and others. Use 1-2 search calls per round to cast wide; use `fetch_url` to read a specific high-value source. Parallel tool calls in one round = ONE round.

You can also still use `fs_*` and `study_*` if the prior context surfaces a substrate document you want to read directly — but skip another full sweep; context_gather already did that.

## FOR EACH SOURCE YOU KEEP

- **Title** + **URL** + **publication date**
- **One-sentence summary** of what it adds (especially what prior context didn't already cover)
- **Short verbatim quote** (1-3 sentences) you might draw on in synthesis
- **Source type:** primary documentation / news reporting / opinion / vendor blog / academic / etc.
- **Credibility note:** primary source for this claim? secondary? recency vs domain half-life?

## OUTPUT FORMAT

Produce a markdown sources brief: a numbered list of up to 8 sources, each with the five fields above. **No prose intro. No prose outro.** Just the structured list. The synthesize stage drafts the actual research piece from your brief + the prior context.$T$;

    v_context_gather_stage := jsonb_build_object(
        'name',           'context_gather',
        'next',           'gather',
        'model',          'qwen3.6-plus',
        'provider',       'opencode_go',
        'agent_family',   'research',
        'auto_advance',   true,
        'tools_disabled', false,
        'input_template', v_context_gather_template
    );

    -- Build the new stages array: [context_gather] + existing stages
    -- with gather's input_template replaced.
    SELECT jsonb_build_array(v_context_gather_stage)
        || jsonb_agg(
            CASE
                WHEN s->>'name' = 'gather'
                    THEN jsonb_set(s, '{input_template}', to_jsonb(v_new_gather_template))
                ELSE s
            END
            ORDER BY ord
        )
    INTO v_new_stages
    FROM jsonb_array_elements(v_pipeline.stages) WITH ORDINALITY AS arr(s, ord);

    UPDATE stewards.pipelines
       SET stages     = v_new_stages,
           updated_at = now()
     WHERE family = 'research-write';

    RAISE NOTICE 'H.2: research-write now has % stages (was %)',
        jsonb_array_length(v_new_stages),
        jsonb_array_length(v_pipeline.stages);
END
$$;

-- pipeline_stage_maturity: context_gather DOES NOT advance maturity, so
-- we deliberately do NOT insert a row. work_item_advance reads this
-- table with a single-row lookup; when no row matches the completing
-- stage, v_new_maturity stays NULL and the COALESCE in the UPDATE
-- leaves work_items.maturity unchanged. That's the "no-op advance"
-- semantics we want for context_gather → gather.
--
-- Confirmed by reading work_item_advance source: the function has
-- explicit "IF v_new_maturity IS NULL THEN v_new_maturity := NULL"
-- guard and uses COALESCE(v_new_maturity, maturity) on UPDATE.

-- stage_models: context_gather defaults to qwen3.6-plus (set inline on the
-- stage def, but the stage_models table is consulted by some retry paths,
-- so keep them in sync).
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model)
VALUES ('research-write', 'context_gather', 'qwen3.6-plus')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET default_model = EXCLUDED.default_model;

-- Sanity check.
SELECT 'stages:' AS check_name,
       jsonb_array_length(stages) AS n_stages,
       (stages->0)->>'name' AS first_stage,
       (stages->1)->>'name' AS second_stage
  FROM stewards.pipelines WHERE family='research-write';
