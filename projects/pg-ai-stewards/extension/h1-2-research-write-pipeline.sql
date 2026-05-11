-- =====================================================================
-- Batch H.1.2 + H.1.3 — research-write pipeline + tools_disabled wiring
--
-- H.1.2: seed the `research-write` pipeline. Three stages:
--   - gather (kimi-k2.6, research agent, tools ENABLED) — external sources
--   - synthesize (kimi-k2.6, research agent, tools enabled lightly)
--   - review (qwen3.6-plus, research agent, tools DISABLED — verification)
--
-- Per D-H1 (lightweight): mirror study-write's stages JSONB shape so the
-- existing dispatch + bgworker machinery handles research-write without
-- code changes. The one substrate-side change rides in this same file:
--
-- H.1.3: extend work_item_dispatch_stage to forward stage.tools_disabled
--   into the payload. Today the dispatch builds payload with standard
--   fields; it doesn't propagate a per-stage tools_disabled flag.
--   Adding the COALESCE into the payload jsonb is the minimum change
--   that makes the review stage actually run tools-off.
-- =====================================================================

-- ---------------------------------------------------------------------
-- H.1.2: pipeline definition
-- ---------------------------------------------------------------------

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder
)
VALUES (
    'research-write',
    jsonb_build_array(
        jsonb_build_object(
            'name',            'gather',
            'next',            'synthesize',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'You are gathering sources for a research piece that will answer the binding question above.' || E'\n\n' ||
                'Use the tools available to you (web_search_exa, web_search, fetch_url, fetch_urls, yt_*, etc.) to find 6-12 credible sources that bear on the binding question. The general-research intent applies — your active system prompt carries its values; reread them if you forget.' || E'\n\n' ||
                'For each source you keep, capture:' || E'\n' ||
                '  - Title + URL + publication date' || E'\n' ||
                '  - One-sentence summary of what it adds to the binding question' || E'\n' ||
                '  - A short verbatim quote (1-3 sentences) you might draw on in the synthesis' || E'\n' ||
                '  - Source type: primary documentation, news reporting, opinion/analysis, vendor blog, academic, social-media-thread, etc.' || E'\n\n' ||
                'Recency: where the domain moves fast (AI tooling, ML research, frontier-lab announcements), strongly prefer 2025-2026 sources. Where the domain is slower (epistemics, foundational engineering principles), older sources are fine. Flag any source where the publication date is older than half the topic''s relevance horizon.' || E'\n\n' ||
                'Credibility: prefer primary documentation (the vendor''s own docs, the paper itself, the official announcement) over secondary reporting. Mark each source you keep with a credibility note. Refuse to summarize what you can''t source.' || E'\n\n' ||
                'Cross-reference: if a claim appears in multiple independent sources, note it. If it appears in only one, note that too.' || E'\n\n' ||
                'Produce a sources brief — a structured list of every source kept, with the four fields above. The next stage drafts the synthesis from this brief; quality of sources here governs quality of output downstream. Do NOT write the synthesis yet.'
        ),
        jsonb_build_object(
            'name',            'synthesize',
            'next',            'review',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Sources brief from the gather stage:' || E'\n\n' ||
                '{{stage_results.gather.output}}' || E'\n\n' ||
                'Now write the research piece. Draw on the sources collected in the gather stage. You MAY re-fetch any source via fetch_url if you need to re-read it; you SHOULD NOT introduce new sources here — that''s a sign the gather stage was incomplete and would be better fixed by re-running gather.' || E'\n\n' ||
                'Quote text VERBATIM only when you have the source text in front of you in this session. Paraphrase otherwise — "Vendor X says that..." is honest; an unverified direct quote is not.' || E'\n\n' ||
                'Attribution: every non-trivial claim cites the source it came from. Use inline markdown links: [Source Title](https://url). Where a claim is your synthesis across multiple sources, say so explicitly.' || E'\n\n' ||
                'Structure suggestion (adapt to what the binding question actually needs):' || E'\n' ||
                '  - **Headlines** — the 3-5 most important findings that answer the binding question' || E'\n' ||
                '  - **Notable** — second-tier findings worth knowing' || E'\n' ||
                '  - **Skeptical takes** — credible dissenting voices, if any' || E'\n' ||
                '  - **Open questions** — what the sources don''t answer' || E'\n\n' ||
                'Length: aim for 800-2500 words depending on topic depth. Resist the urge to pad. Honest uncertainty ("I couldn''t find a credible source on X") is preferred over fabrication.' || E'\n\n' ||
                'Produce the complete research piece in markdown. The next stage reviews it.'
        ),
        jsonb_build_object(
            'name',            'review',
            'next',            NULL,
            'model',           'qwen3.6-plus',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  true,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'The draft from the previous stage:' || E'\n\n' ||
                '{{stage_results.synthesize.output}}' || E'\n\n' ||
                'Review the draft against four criteria:' || E'\n\n' ||
                '1. **Source credibility.** Every claim of fact has a citation. Citations point to credible sources (primary docs or established reporting, not random blog posts presented as fact). Where a claim is uncited or cited weakly, flag it.' || E'\n\n' ||
                '2. **Recency.** Where the domain moves fast, sources are 2025-2026. Older sources are explicitly flagged or appropriate to a slow-moving domain.' || E'\n\n' ||
                '3. **Binding question coverage.** Does the draft answer what was asked? If not, name what''s missing.' || E'\n\n' ||
                '4. **Honest uncertainty.** Where the sources don''t support a strong claim, the draft says so. No fabricated certainty.' || E'\n\n' ||
                'Tools are DISABLED for this stage. You CANNOT fetch URLs or re-search — your review must rest on the draft itself plus the sources it cites in-line. If a claim looks unverifiable from the draft alone, flag it as unverifiable rather than try to verify externally.' || E'\n\n' ||
                'Return ONE of:' || E'\n' ||
                '(a) The same draft, verbatim and unchanged, if it passes all four criteria. Prefix with a single line: "REVIEW: passes" then a blank line then the draft.' || E'\n' ||
                '(b) A revised draft. Prefix with "REVIEW: revised" then a blank line, the revised draft, and at the end a brief notes section listing what changed and why.'
        )
    ),
    true,   -- sabbath_enabled (research is creative; sabbath reflection on a deep piece is valuable)
    true,   -- atonement_enabled (research that hits cost cap is worth atoning over)
    'research/<slug>.md',
    NULL,   -- file_content_jsonpath: v1 uses whole stage output
    '["raw","researched","planned","specced","executing","verified"]'::jsonb
)
ON CONFLICT (family) DO UPDATE SET
    stages                    = EXCLUDED.stages,
    sabbath_enabled           = EXCLUDED.sabbath_enabled,
    atonement_enabled         = EXCLUDED.atonement_enabled,
    file_destination_template = EXCLUDED.file_destination_template,
    file_content_jsonpath     = EXCLUDED.file_content_jsonpath,
    maturity_ladder           = EXCLUDED.maturity_ladder;

-- Stage models for the three cells
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('research-write', 'gather',     'kimi-k2.6',    'External-source gather; tools enabled (exa, web_search, fetch_url, yt_*).'),
    ('research-write', 'synthesize', 'kimi-k2.6',    'Draft synthesis from gather brief; tools enabled lightly (re-fetch only).'),
    ('research-write', 'review',     'qwen3.6-plus', 'Tools-disabled verification pass; cheaper model is sufficient for structured-output review.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- Maturity rung mapping. Research-write skips "executing" intentionally:
-- synthesize IS the draft (no separate draft + execute stages). This is
-- the first gospel-shape-vs-creation-shape fork — documented in
-- substrate-batch-h-pipeline-expansion.md §V.1.2.
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('research-write', 'gather',     'researched', 'Sources collected + summarized; ready for synthesis.'),
    ('research-write', 'synthesize', 'planned',    'Draft is the plan. No separate executing rung — research has no draft-vs-execute distinction.'),
    ('research-write', 'review',     'verified',   'Review pass complete; piece is verified.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;

-- ---------------------------------------------------------------------
-- H.1.3: extend work_item_dispatch_stage to forward stage.tools_disabled.
-- Verbatim copy of the live body with one additional jsonb key in the
-- payload. The bgworker already honors a tools_disabled payload field
-- (Phase C wiring); this just makes per-stage declarations effective.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id uuid,
    p_user_input text DEFAULT NULL,
    p_allow_failed_status boolean DEFAULT false
)
RETURNS bigint
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_stage       jsonb;
    v_agent       text;
    v_model       text;
    v_provider    text;
    v_session_id  text;
    v_user_input  text;
    v_body        jsonb;
    v_payload     jsonb;
    v_work_id     bigint;
    v_was_failed  boolean := false;
    v_tools_off   boolean;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    IF v_wi.status NOT IN ('pending', 'awaiting_review')
       AND NOT (p_allow_failed_status AND v_wi.status = 'failed')
    THEN
        RAISE EXCEPTION 'work_item %: cannot dispatch from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_was_failed := (v_wi.status = 'failed');

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_agent    := v_stage->>'agent_family';
    v_model    := COALESCE(v_wi.model_override,    v_stage->>'model');
    v_provider := COALESCE(v_wi.provider_override, v_stage->>'provider');

    IF v_agent IS NULL OR v_model IS NULL OR v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family/model/provider',
            p_work_item_id, v_wi.current_stage;
    END IF;

    -- H.1.3: per-stage tools_disabled flag honored at dispatch
    v_tools_off := COALESCE((v_stage->>'tools_disabled')::boolean, false);

    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    IF p_user_input IS NOT NULL THEN
        v_user_input := p_user_input;
    ELSE
        v_user_input := stewards.render_stage_input(p_work_item_id);
        IF v_user_input IS NULL THEN
            v_user_input := coalesce(
                v_wi.input->>'user_input',
                v_wi.input::text
            );
        END IF;
    END IF;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_user_input, v_model);

    v_body := stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               v_body->'_meta',
        'body',               (v_body - '_meta')
                              || jsonb_build_object('user', v_session_id),
        'tools_disabled',     v_tools_off,
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    UPDATE stewards.work_items
       SET status      = 'in_progress',
           session_ids = session_ids || v_session_id,
           updated_at  = now()
     WHERE id = p_work_item_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_dispatch_stage(uuid, text, boolean) IS
'H.1.3 (Batch H): payload now forwards stage.tools_disabled (default false) so per-stage tools-off declarations (e.g., research-write review) take effect at the bgworker. Otherwise verbatim from Phase 4b dispatch.';
