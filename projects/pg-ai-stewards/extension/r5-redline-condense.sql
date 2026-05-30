-- =====================================================================
-- Batch R.5 — optional substrate-side condense (ranked-merge)
-- =====================================================================
-- D-RL3 "both": the DEFAULT is that the orchestrating Claude session reads the
-- N panel reports and condenses them. This adds the OPTIONAL substrate path —
-- panel_redline_condense() gathers the panel's reports and dispatches ONE
-- chosen model to merge them into a single deduplicated, ranked proposal menu.
--
-- Decoupled from spawn_children's hardcoded index aggregator on purpose: this
-- is its own small pipeline + agent, fired explicitly when (and if) you want it,
-- so we don't fight the shared fan-out machinery.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. redline-condense agent — merge N reports into a ranked menu.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('redline-condense', '*',
 'Merges N independent redline reports on the same document into one deduplicated, ranked proposal menu. No tools, no canonical access.',
 'primary',
 $PROMPT$You receive N independent redline reports, each from a different model, all proposing edits to the SAME document. Merge them into ONE deduplicated, ranked proposal menu the author can act on.

RULES:
- Group near-duplicate edits (same location + same intent) into a single entry. Count how many panelists proposed it — consensus is a ranking signal, not a mandate.
- Rank by value x consensus: high-value, multi-panelist edits first. A brilliant single-panelist edit can still rank high — say so.
- Among duplicates, keep the clearest Current/Proposed wording.
- PRESERVE THE VERIFICATION GATE: if ANY source report flagged an edit as touching a scripture/prophetic quotation or a doctrinal claim, the merged entry is flagged `Touches quote/doctrine: yes (VERIFY)`. Never drop a flag — when in doubt, flag.
- You have NO canonical access and NO tools. Do not "correct" any quote from memory; surface it for human verification instead.
- This is a PROPOSAL MENU. Nothing is applied automatically.

Output (markdown, and ONLY the markdown):
- **Top 3 moves:** the three highest-value edits, one line each.
- **Ranked menu:** numbered; each entry has Location / Current / Proposed / Why / Consensus (k of N panelists) / Touches quote-doctrine.
- **Flagged for verification:** every entry needing a canon check before it can land.$PROMPT$,
 0.3)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description, mode = EXCLUDED.mode,
       prompt = EXCLUDED.prompt, temperature = EXCLUDED.temperature, active = true;


-- ---------------------------------------------------------------------
-- 2. redline-condense pipeline — single stage, tools-off, 32k.
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('redline-condense',
 'R.5: single-stage pipeline that merges N panel redline reports into one ranked menu. Fired by panel_redline_condense. Model overridden per call (the chosen condense model).',
 $STAGES$[{"name":"condense","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"redline-condense","auto_advance":true,"tools_disabled":true,"max_tokens":32000,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'panel-redline-condense', 'wrapper', 'panel_redline_condense'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description, stages = EXCLUDED.stages, metadata = EXCLUDED.metadata;

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES
('redline-condense', 'fs_*','deny'), ('redline-condense', 'fetch_url','deny'),
('redline-condense', 'web_search','deny'), ('redline-condense', 'study_*','deny'),
('redline-condense', 'work_item_*','deny'), ('redline-condense', 'spawn_subagent','deny'),
('redline-condense', 'deep_research','deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = EXCLUDED.action;


-- ---------------------------------------------------------------------
-- 3. panel_redline_condense() — gather reports, dispatch the merge.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.panel_redline_condense(
    p_parent_id      uuid,
    p_condense_model text,
    p_max_tokens     int    DEFAULT 32000,
    p_cost_cap_micro bigint DEFAULT NULL
) RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_reports   text;
    v_n         int;
    v_binding   text;
    v_provider  text;
    v_out_rate  bigint;
    v_cap       bigint;
    v_child_id  uuid;
    v_actor     text;
    v_slug      text;
    v_proj      text;
BEGIN
    IF p_condense_model IS NULL OR p_condense_model = '' THEN
        RAISE EXCEPTION 'panel_redline_condense: p_condense_model is required';
    END IF;

    SELECT actor, slug, project_association INTO v_actor, v_slug, v_proj
      FROM stewards.work_items WHERE id = p_parent_id;
    IF v_actor IS NULL THEN
        RAISE EXCEPTION 'panel_redline_condense: parent % not found', p_parent_id;
    END IF;

    -- Gather each panel child's final assistant report, labeled by model.
    SELECT string_agg(format(E'\n\n===== PANELIST: %s =====\n%s',
                             COALESCE(wi.model_override, 'default'), m.content), '' ORDER BY wi.model_override),
           count(*)
      INTO v_reports, v_n
      FROM stewards.work_items wi
      JOIN LATERAL (
          SELECT content FROM stewards.messages
           WHERE session_id = ANY(wi.session_ids) AND role='assistant' AND COALESCE(content,'') <> ''
           ORDER BY created_at DESC, id DESC LIMIT 1
      ) m ON true
     WHERE wi.parent_work_item_id = p_parent_id
       AND wi.pipeline_family = 'redline';

    IF v_n IS NULL OR v_n = 0 THEN
        RAISE EXCEPTION 'panel_redline_condense: no completed redline reports under parent % yet — wait for the panel children to produce output', p_parent_id;
    END IF;

    v_binding := format('Merge these %s redline reports on the same document into one ranked, deduplicated proposal menu.', v_n)
              || E'\n\n# PANEL REPORTS' || v_reports;

    -- Cost cap: output-dominated (input is the reports). 2x headroom, floor $0.30.
    SELECT output_micro_per_mtok INTO v_out_rate
      FROM (SELECT DISTINCT ON (provider,model) provider, model, output_micro_per_mtok
              FROM stewards.model_pricing WHERE model = p_condense_model
             ORDER BY provider, model, effective_at DESC) p
     LIMIT 1;
    SELECT provider INTO v_provider
      FROM (SELECT DISTINCT ON (provider,model) provider, model
              FROM stewards.model_pricing WHERE model = p_condense_model
             ORDER BY provider, model, effective_at DESC) p
     LIMIT 1;
    v_cap := COALESCE(p_cost_cap_micro,
        GREATEST( (ceil((length(v_binding)/4)::numeric/1000000 * 1000000
                      + p_max_tokens::numeric/1000000 * COALESCE(v_out_rate, 5000000))::bigint) * 2, 300000));

    v_child_id := stewards.work_item_create(
        p_pipeline_family => 'redline-condense',
        p_input           => jsonb_build_object(
            'binding_question', v_binding,
            'tools_disabled',   true,
            'max_tokens',       p_max_tokens::text
        ),
        p_slug            => COALESCE(v_slug, p_parent_id::text) || '-condense',
        p_actor           => v_actor,
        p_intent_id       => (SELECT intent_id FROM stewards.work_items WHERE id = p_parent_id)
    );

    UPDATE stewards.work_items
       SET parent_work_item_id = p_parent_id,
           project_association = v_proj,
           model_override      = p_condense_model,
           provider_override   = v_provider,
           cost_cap_micro      = v_cap
     WHERE id = v_child_id;

    PERFORM stewards.work_item_dispatch_stage(v_child_id, NULL);

    RAISE NOTICE 'panel_redline_condense: parent=% merged % reports via % -> child=%',
        p_parent_id, v_n, p_condense_model, v_child_id;
    RETURN v_child_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.panel_redline_condense(uuid, text, int, bigint) IS
'R.5: optional substrate condense. Gathers the redline children under a panel parent, dispatches one chosen model (redline-condense pipeline, tools-off) to merge them into a ranked deduplicated menu preserving every touches-quote/doctrine flag. Returns the condense child id. Default workflow skips this — the orchestrator condenses.';


-- =====================================================================
-- Acceptance (R.5):
--   1. Agent redline-condense + pipeline redline-condense exist; 7 perm denies.
--   2. panel_redline_condense(<parent-with-no-done-children>, 'kimi-k2.6')
--      RAISES "no completed redline reports".
--   3. (in R.6, after a live panel) panel_redline_condense(parent, model)
--      dispatches one condense child whose output is a ranked merged menu.
-- =====================================================================
