-- =====================================================================
-- Batch J.12 — Error classification + brainstorm cap pre-flight
-- =====================================================================
-- Makes budget/cap failures easy to tell apart from generic failures.
-- Two surfaces:
--   1. classify_error(text) — read-time category for any stored error
--      string (work_items.error, work_queue.error). No rebuild: the raw
--      error is already captured (advance trigger -> work_item_fail);
--      this just labels it. Surfaced via the work_items API + UI.
--   2. start_brainstorm pre-flight — if a requested lens routes to a
--      provider whose enforced spend cap is already reached, RAISE a
--      clear message BEFORE spawning. The dispatch-level gate (J.11)
--      still protects every path, but its RAISE inside spawn_children is
--      swallowed by the trigger's EXCEPTION handler (0 children, logged).
--      The pre-flight surfaces it to the MCP tool / direct caller.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. classify_error(error_text) -> category
-- ---------------------------------------------------------------------
-- Categories (most-specific first):
--   spend_cap_reached     — our J.11 gate refused the dispatch
--   provider_budget       — provider rejected for quota/billing/balance
--                           (Gemini 429 RESOURCE_EXHAUSTED, billing off,
--                            out of credit). For a prepaid key this means
--                            "top up the balance".
--   rate_limited          — transient rate limit (retryable), not budget
--   auth                  — bad/missing key, permission denied
--   timeout               — request timed out / deadline exceeded
--   other                 — anything else
--   none                  — no error
CREATE OR REPLACE FUNCTION stewards.classify_error(p_error text)
RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT CASE
        WHEN p_error IS NULL OR btrim(p_error) = '' THEN 'none'
        WHEN p_error ILIKE '%spend cap reached%'
          OR p_error ILIKE '%provider_cap%'
          OR p_error ILIKE '%provider_cap_refill%'                 THEN 'spend_cap_reached'
        WHEN p_error ILIKE '%RESOURCE_EXHAUSTED%'
          OR p_error ILIKE '%exceeded your current quota%'
          OR p_error ILIKE '%billing%'
          OR p_error ILIKE '%out of credit%'
          OR p_error ILIKE '%insufficient%balance%'
          OR p_error ILIKE '%insufficient%credit%'
          OR p_error ILIKE '%quota%exceeded%'
          OR p_error ILIKE '%FAILED_PRECONDITION%'                 THEN 'provider_budget'
        WHEN p_error ILIKE '%rate limit%'
          OR p_error ILIKE '%rate_limit%'
          OR p_error ILIKE '%too many requests%'
          OR p_error ILIKE '%HTTP 429%'                            THEN 'rate_limited'
        WHEN p_error ILIKE '%HTTP 401%'
          OR p_error ILIKE '%HTTP 403%'
          OR p_error ILIKE '%PERMISSION_DENIED%'
          OR p_error ILIKE '%UNAUTHENTICATED%'
          OR p_error ILIKE '%API key%'
          OR p_error ILIKE '%invalid%key%'                         THEN 'auth'
        WHEN p_error ILIKE '%timeout%'
          OR p_error ILIKE '%timed out%'
          OR p_error ILIKE '%deadline%'                            THEN 'timeout'
        ELSE 'other'
    END
$$;

COMMENT ON FUNCTION stewards.classify_error(text) IS
'J.12: classify a stored error string into a category (spend_cap_reached | provider_budget | rate_limited | auth | timeout | other | none). Read-time labeling for the work_items API + UI. Note: Gemini returns HTTP 429 for BOTH rate limits and quota exhaustion; the quota/RESOURCE_EXHAUSTED wording is checked first so true budget exhaustion classifies as provider_budget, not rate_limited.';

-- ---------------------------------------------------------------------
-- 2. work_item_failures — convenience view for quick triage
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.work_item_failures AS
SELECT wi.id,
       wi.slug,
       wi.pipeline_family,
       wi.status,
       stewards.classify_error(wi.error) AS error_category,
       wi.error,
       wi.updated_at
  FROM stewards.work_items wi
 WHERE wi.status = 'failed'
   AND wi.error IS NOT NULL
 ORDER BY wi.updated_at DESC;

COMMENT ON VIEW stewards.work_item_failures IS
'J.12: failed work_items with a classified error_category. Quick triage: SELECT * FROM stewards.work_item_failures WHERE error_category = ''provider_budget'';';

-- ---------------------------------------------------------------------
-- 3. start_brainstorm pre-flight cap check (carry J.9.c forward).
-- ---------------------------------------------------------------------
DROP FUNCTION IF EXISTS stewards.start_brainstorm(text, text, text, text, text, bigint, jsonb, text[]);

CREATE OR REPLACE FUNCTION stewards.start_brainstorm(
    p_binding_question        text,
    p_destination             text,
    p_project_association     text     DEFAULT NULL,
    p_actor                   text     DEFAULT 'human',
    p_slug                    text     DEFAULT NULL,
    p_cost_cap_per_lens_micro bigint   DEFAULT 200000,
    p_models                  jsonb    DEFAULT NULL,
    p_lenses                  text[]   DEFAULT ARRAY['scamper', 'six-hats', 'crazy8s', 'reverse']
)
RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug             text;
    v_parent_id        uuid;
    v_manifest         jsonb;
    v_lens             text;
    v_lens_family      text;
    v_lens_slug        text;
    v_models_entry     jsonb;
    v_model_override   text;
    v_provider_override text;
    v_child            jsonb;
    v_children_arr     jsonb := '[]'::jsonb;
    v_unknown_lenses   text[];
    v_lens_provider    text;
    v_capped           text[] := ARRAY[]::text[];
BEGIN
    IF p_lenses IS NULL OR cardinality(p_lenses) = 0 THEN
        RAISE EXCEPTION 'start_brainstorm: p_lenses must contain at least one lens name';
    END IF;

    -- Validate every requested lens corresponds to an existing pipeline.
    SELECT array_agg(lens_name)
      INTO v_unknown_lenses
      FROM (SELECT unnest(p_lenses) AS lens_name) requested
     WHERE NOT EXISTS (
         SELECT 1 FROM stewards.pipelines
          WHERE family = 'brainstorm-' || requested.lens_name
     );
    IF v_unknown_lenses IS NOT NULL THEN
        RAISE EXCEPTION 'start_brainstorm: unknown lens name(s): %. Available lenses: %. (Introspect with SELECT regexp_replace(family, ''^brainstorm-'', '''') FROM stewards.pipelines WHERE family LIKE ''brainstorm-%%'')',
            v_unknown_lenses,
            (SELECT array_agg(regexp_replace(family, '^brainstorm-', ''))
               FROM stewards.pipelines WHERE family LIKE 'brainstorm-%');
    END IF;

    -- J.12 PRE-FLIGHT: refuse early (with a clear message) if any lens
    -- routes to a provider whose enforced spend cap is already reached.
    -- Resolves each lens's effective provider the same way dispatch does
    -- (p_models override -> pipeline default -> catalog default), so a
    -- capped Gemini lens surfaces here instead of being silently dropped
    -- by spawn_children's swallowed dispatch RAISE.
    FOREACH v_lens IN ARRAY p_lenses LOOP
        v_lens_provider := NULL;
        IF p_models IS NOT NULL AND (p_models ? v_lens)
           AND jsonb_typeof(p_models -> v_lens) = 'object' THEN
            v_lens_provider := (p_models -> v_lens) ->> 'provider';
        END IF;
        IF v_lens_provider IS NULL THEN
            v_lens_provider := COALESCE(
                (SELECT metadata->>'default_provider' FROM stewards.pipelines
                  WHERE family = 'brainstorm-' || v_lens),
                stewards.catalog_default_provider()
            );
        END IF;
        IF v_lens_provider IS NOT NULL
           AND stewards.provider_cap_exceeded(v_lens_provider)
           AND NOT (v_lens_provider = ANY(v_capped)) THEN
            v_capped := v_capped || v_lens_provider;
        END IF;
    END LOOP;

    IF cardinality(v_capped) > 0 THEN
        RAISE EXCEPTION 'start_brainstorm: refused — provider(s) % at spend cap. Top up + reset: SELECT stewards.provider_cap_refill(''<provider>''); (or drop the lens(es) routed to them).',
            v_capped;
    END IF;

    v_slug := COALESCE(p_slug, 'brainstorm-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS'));

    FOREACH v_lens IN ARRAY p_lenses LOOP
        v_lens_family    := 'brainstorm-' || v_lens;
        v_lens_slug      := v_slug || '-' || v_lens;
        v_model_override := NULL;
        v_provider_override := NULL;

        IF p_models IS NOT NULL AND (p_models ? v_lens) THEN
            v_models_entry := p_models -> v_lens;
            IF jsonb_typeof(v_models_entry) = 'string' THEN
                v_model_override := v_models_entry #>> '{}';
            ELSIF jsonb_typeof(v_models_entry) = 'object' THEN
                v_model_override    := v_models_entry ->> 'model';
                v_provider_override := v_models_entry ->> 'provider';
            END IF;
        END IF;

        v_child := jsonb_build_object(
            'slug',             v_lens_slug,
            'pipeline_family',  v_lens_family,
            'binding_question', p_binding_question,
            'cost_cap_micro',   p_cost_cap_per_lens_micro
        );
        IF v_model_override IS NOT NULL THEN
            v_child := v_child || jsonb_build_object('model_override', v_model_override);
        END IF;
        IF v_provider_override IS NOT NULL THEN
            v_child := v_child || jsonb_build_object('provider_override', v_provider_override);
        END IF;

        v_children_arr := v_children_arr || v_child;
    END LOOP;

    v_manifest := jsonb_build_object(
        'rationale', format('Brainstorm: %s lens(es) — %s. Synthesis aggregator combines.',
                            cardinality(p_lenses), array_to_string(p_lenses, ', ')),
        'children', v_children_arr,
        'aggregate', jsonb_build_object('destination', p_destination, 'synthesis', true)
    );

    INSERT INTO stewards.work_items (
        pipeline_family, current_stage, slug, input, intent_id, actor,
        project_association, stage_results, maturity, status
    ) VALUES (
        'decompose-fanout', 'decompose', v_slug,
        jsonb_build_object('binding_question', p_binding_question, 'lenses', to_jsonb(p_lenses)),
        (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
        p_actor, p_project_association,
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id INTO v_parent_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_parent_id;

    RAISE NOTICE 'start_brainstorm: parent=% slug=% lenses=% p_models=%',
        v_parent_id, v_slug, p_lenses, COALESCE(p_models::text, 'NULL');
    RETURN v_parent_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.start_brainstorm(text, text, text, text, text, bigint, jsonb, text[]) IS
'J.12: adds a pre-flight enforced-cap check (refuses with a clear message before spawning if any lens routes to an over-cap provider) on top of the J.9.c lens-subset + per-lens-model signature. The J.11 dispatch gate remains the universal enforcement; this just surfaces the cap cleanly on the brainstorm path.';

-- =====================================================================
-- Acceptance:
--   1. classify_error('... spend cap reached ...') = 'spend_cap_reached'
--   2. classify_error('chat HTTP 429: {... RESOURCE_EXHAUSTED ... quota ...}') = 'provider_budget'
--   3. classify_error('chat HTTP 401: invalid API key') = 'auth'
--   4. classify_error('') = 'none'
--   5. With google_gemini over cap, start_brainstorm(..., p_models with a
--      gemini lens) RAISEs 'refused — provider(s) {google_gemini} at spend
--      cap' BEFORE inserting the parent.
-- =====================================================================
