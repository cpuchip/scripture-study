-- =====================================================================
-- Phase 3c.3 — Stage input templating + study-write pipeline
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: 13th file).
--
-- Builds on:
--   - Phase 3c.1 (work_items, work_item_dispatch_stage)
--   - Phase 3c.2 (auto-advance trigger that uses dispatch on next stage)
--   - Phase 3c.2.5 (study tools — what the agent calls during stages)
--
-- This file adds:
--   1. stewards.resolve_template_path(input, stage_results, path) —
--      walks a {{root.a.b.c}} path against work_item.input or
--      work_item.stage_results. Errors loudly on missing paths.
--   2. stewards.render_stage_input(work_item_id) — looks up the
--      current stage's `input_template` and renders {{...}}
--      placeholders against work_item state.
--   3. Updated stewards.work_item_dispatch_stage that uses the
--      template when the stage definition has one (else falls back
--      to the existing user_input/stringified-input behavior).
--   4. The `study-write` pipeline definition — 3 stages
--      (outline → draft → review).
-- =====================================================================

-- ---------------------------------------------------------------------
-- resolve_template_path
--
-- Path syntax: {{root.a.b.c}} where root is "input" or "stage_results"
-- and a/b/c are nested jsonb keys.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.resolve_template_path(
    p_input         jsonb,
    p_stage_results jsonb,
    p_path          text
) RETURNS text
LANGUAGE plpgsql IMMUTABLE AS $func$
DECLARE
    v_parts text[];
    v_root  text;
    v_value jsonb;
    i       int;
BEGIN
    v_parts := string_to_array(trim(p_path), '.');
    IF cardinality(v_parts) < 1 OR v_parts[1] IS NULL OR v_parts[1] = '' THEN
        RAISE EXCEPTION
            'resolve_template_path: empty path';
    END IF;

    v_root := v_parts[1];
    IF v_root = 'input' THEN
        v_value := p_input;
    ELSIF v_root = 'stage_results' THEN
        v_value := p_stage_results;
    ELSE
        RAISE EXCEPTION
            'resolve_template_path: unknown root % in path %; expected "input" or "stage_results"',
            v_root, p_path;
    END IF;

    -- Walk the rest of the path through nested jsonb objects.
    FOR i IN 2..cardinality(v_parts) LOOP
        IF v_value IS NULL OR jsonb_typeof(v_value) <> 'object' THEN
            RAISE EXCEPTION
                'resolve_template_path: path % not resolvable; stopped at %',
                p_path, v_parts[i-1];
        END IF;
        v_value := v_value -> v_parts[i];
    END LOOP;

    IF v_value IS NULL THEN
        RAISE EXCEPTION
            'resolve_template_path: path % resolved to NULL', p_path;
    END IF;

    -- Strings unwrap (no quotes); other types stringify.
    IF jsonb_typeof(v_value) = 'string' THEN
        RETURN v_value #>> '{}';
    ELSE
        RETURN v_value::text;
    END IF;
END;
$func$;

COMMENT ON FUNCTION stewards.resolve_template_path(jsonb, jsonb, text) IS
'Phase 3c.3: walk a {{root.a.b.c}} template path against work_item.input or work_item.stage_results. Errors loudly on missing paths so template bugs surface at dispatch, not in agent output.';

-- ---------------------------------------------------------------------
-- render_stage_input
--
-- For the work_item's CURRENT stage, looks up its input_template from
-- the pipeline definition and renders {{...}} placeholders. If the
-- stage has no input_template, returns NULL (caller falls back).
--
-- Multiple references to the same path are all substituted (regex /g).
-- Whitespace inside {{ ... }} is tolerated.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.render_stage_input(p_work_item_id uuid)
RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_wi       stewards.work_items%ROWTYPE;
    v_stage    jsonb;
    v_template text;
    v_rendered text;
    v_match    text[];
    v_path     text;
    v_value    text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'render_stage_input: work_item % not found', p_work_item_id;
    END IF;

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION
            'render_stage_input: stage % not found in pipeline %',
            v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_template := v_stage->>'input_template';
    IF v_template IS NULL THEN
        RETURN NULL;  -- caller falls back
    END IF;

    v_rendered := v_template;
    -- Walk every distinct {{...}} match.
    FOR v_match IN
        SELECT regexp_matches(v_template, '\{\{\s*([^}]+?)\s*\}\}', 'g')
    LOOP
        v_path := v_match[1];
        v_value := stewards.resolve_template_path(
            v_wi.input, v_wi.stage_results, v_path);
        -- Replace every literal {{<path>}} occurrence (with surrounding
        -- whitespace tolerance via a regex_replace).
        v_rendered := regexp_replace(
            v_rendered,
            '\{\{\s*' || regexp_replace(v_path, '([\\.()|*+?\[\]{}^$])', '\\\1', 'g') || '\s*\}\}',
            v_value,
            'g'
        );
    END LOOP;

    RETURN v_rendered;
END;
$func$;

COMMENT ON FUNCTION stewards.render_stage_input(uuid) IS
'Phase 3c.3: render the current stage''s input_template against work_item state. Returns NULL if the stage has no template (caller falls back to legacy behavior).';

-- ---------------------------------------------------------------------
-- Update work_item_dispatch_stage to use templating when present.
--
-- Diff from 3c.1:
--   v_user_input := coalesce(
--       p_user_input,
--       v_wi.input->>'user_input',
--       v_wi.input::text
--   );
--
-- becomes:
--   IF p_user_input IS NOT NULL THEN
--       v_user_input := p_user_input;
--   ELSE
--       v_user_input := stewards.render_stage_input(p_work_item_id);
--       IF v_user_input IS NULL THEN
--           v_user_input := coalesce(v_wi.input->>'user_input', v_wi.input::text);
--       END IF;
--   END IF;
--
-- Everything else identical.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id uuid,
    p_user_input   text DEFAULT NULL
) RETURNS bigint
LANGUAGE plpgsql AS $func$
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
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;
    IF v_wi.status NOT IN ('pending', 'awaiting_review') THEN
        RAISE EXCEPTION 'work_item %: cannot dispatch from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_agent    := v_stage->>'agent_family';
    v_model    := v_stage->>'model';
    v_provider := v_stage->>'provider';
    IF v_agent IS NULL OR v_model IS NULL OR v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family/model/provider',
            p_work_item_id, v_wi.current_stage;
    END IF;

    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    -- Phase 3c.3: input resolution priority.
    --   1. Explicit p_user_input override (CLI dispatch w/ --user-input).
    --   2. Stage's input_template rendered against work_item state.
    --   3. work_item.input.user_input field (legacy fallback).
    --   4. Stringified work_item.input (last-resort fallback).
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

-- ---------------------------------------------------------------------
-- The `study-write` pipeline definition.
--
-- Three stages: outline (plan agent) → draft (study agent) → review
-- (study agent + ben-test skill). Smaller than the 5-stage spec to
-- limit v1 cost and isolate variables.
--
-- The system prompt for each stage comes from the imported agent
-- (e.g., the 17K-char study agent prompt). The user message is
-- composed by render_stage_input from the input_template below.
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages)
VALUES (
    'study-write',
    'First real multi-stage pipeline (Phase 3c.3 v1). Produces a meta-study from the substrate''s document corpus only — no external scripture lookup. Three stages: outline (plan agent surveys what''s in the corpus relevant to the binding question, proposes structure + source plan), draft (study agent writes the study using study_get to re-read sources), review (study agent self-checks against voice + source-verification, optionally loading the ben-test skill).',
    jsonb_build_array(
        -- Stage 1: outline
        jsonb_build_object(
            'name',          'outline',
            'agent_family',  'plan',
            'model',         'kimi-k2.6',
            'provider',      'opencode_go',
            'next',          'draft',
            'auto_advance',  true,
            'input_template',
$tmpl$Binding question: {{input.binding_question}}

Outline a meta-study addressing this question using the SUBSTRATE'S existing document corpus. The substrate has 364 documents across 5 kinds (study, doc, proposal, journal, phase-doc) — no external scripture lookup is available in this pipeline.

You have access to substrate-internal tools:
- study_search_text(query, kinds, limit) — FTS search across the corpus
- study_get(slug, include_body, body_line_offset, body_line_count) — read a doc with line-based pagination
- study_similar(slug, limit, min_score) — find similar docs via precomputed embedding edges
- study_citations(slug) — see what canonical sources a doc references (URIs only; no text)
- study_context_for(slug, depth) — typed graph walk (CITES, FEEDS, SIMILAR_TO, etc.)
- skill(name) — load a skill body on demand

Use these tools to discover what's already in the corpus that bears on the binding question. Don't try to call gospel_search or read_file — those tools don't exist in this environment.

Produce an outline with:
- Section headings (3-6 sections that together answer the binding question)
- For each section: which doc slugs (`study_search_text` results) you intend to draw on
- A note on what canonical sources (scriptures/talks via study_citations) those docs reference

Don't write the study itself. The next stage drafts it.$tmpl$
        ),
        -- Stage 2: draft
        jsonb_build_object(
            'name',          'draft',
            'agent_family',  'study',
            'model',         'kimi-k2.6',
            'provider',      'opencode_go',
            'next',          'review',
            'auto_advance',  true,
            'input_template',
$tmpl$Binding question: {{input.binding_question}}

The outline + source plan from the previous stage:

{{stage_results.outline.output}}

Now write the study. Draw on the slugs identified in the outline. Use study_get (with body_line_offset to paginate long docs) to re-read any source you need to cite directly.

Quote text VERBATIM only from study_get results in this session. If you don't have the source text in front of you, paraphrase ("Maxwell teaches that...") rather than quote.

The substrate's tools are study_search_text, study_get, study_similar, study_citations, study_context_for, skill. Don't try gospel_search / read_file / web_search.

Produce a complete study in markdown: title, sections matching the outline, paragraph prose, citations as `[doc-slug](#)` links to other studies in the corpus. Aim for 2000-5000 words.$tmpl$
        ),
        -- Stage 3: review
        jsonb_build_object(
            'name',          'review',
            'agent_family',  'study',
            'model',         'kimi-k2.6',
            'provider',      'opencode_go',
            'next',          null,
            'auto_advance',  true,
            'input_template',
$tmpl$Binding question: {{input.binding_question}}

The draft from the previous stage:

{{stage_results.draft.output}}

Review the draft against three criteria:

1. **Voice discipline.** Em-dashes — at most one per paragraph; bibliographic citation dashes don't count. Therefore/but transitions between sections, not "and then." No meta-narration ("This matters because...", "Read that again," "Sit with that"). No closing refrain.

2. **Source verification.** Any text in quotation marks must be a verbatim excerpt from a source you actually retrieved via study_get this session. If a quote can't be verified, either remove it or convert to paraphrase.

3. **Binding question coverage.** Does the study answer what was asked? If not, name what's missing.

Optionally load `skill('ben-test')` for a structured rubric on whether the study practices what it preaches.

Return ONE of:
(a) The same draft, verbatim and unchanged, if it passes all three criteria. Prefix with a single line: "REVIEW: passes" then a blank line then the draft.
(b) A revised draft. Prefix with "REVIEW: revised" then a blank line, the revised draft, and at the end a brief notes section listing what changed and why.$tmpl$
        )
    )
)
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages      = EXCLUDED.stages,
       updated_at  = now();
