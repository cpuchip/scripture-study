-- =====================================================================
-- Phase 3c.2.5 — Study tool registration (sql_fn)
--
-- Live-DB migration. Folds into extension/src/lib.rs at next
-- intentional rebuild (foldback debt: 12th file).
--
-- Builds on:
--   - Phase 1.5 (tool_defs schema, brain_search_text wrapper pattern)
--   - Phase 2.1 (studies table + body_tsv FTS index)
--   - Phase 2.3 (refresh_study_similarity edges)
--   - Phase 2.6c (context_for graph walk)
--   - Phase 3a.1 (imported study agent + its tool_pattern allow rules)
--   - Phase 3c.1/3c.2 (work_items + auto-advance trigger that uses
--     the existing tool_dispatch loop)
--
-- This file adds:
--   1. Two budget-hook columns on stewards.tool_defs (left NULL).
--   2. stewards.study_search_text(query, kinds[], limit) — FTS over
--      studies.body_tsv with multi-kind filter.
--   3. stewards.study_get(slug, include_body, line_offset, line_count,
--      max_chars) — line-paginated read returning jsonb.
--   4. Five _tool wrapper functions (jsonb → jsonb) following the
--      Phase 1.5 brain_search_text_tool pattern.
--   5. Five rows in stewards.tool_defs registering the tools with
--      JSON Schemas.
--
-- After this lands, compose_tools('study') (and any imported agent
-- whose tool_pattern allow-list matches `study_*`) emits these tools
-- in the chat body. The existing tool_dispatch loop (Phase 1.6)
-- handles execution.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Budget hook columns (Phase 3c.2.5: designed, left empty)
-- ---------------------------------------------------------------------
ALTER TABLE stewards.tool_defs
    ADD COLUMN IF NOT EXISTS expected_result_tokens int,
    ADD COLUMN IF NOT EXISTS expected_invocation_tokens int;

COMMENT ON COLUMN stewards.tool_defs.expected_result_tokens IS
'Phase 3c.2.5: typical token weight of this tool''s result. NULL = unknown. Future: estimate_chat_tokens uses this to refine per-stage cost prediction. Populate once we have observation data — leaving empty per the gated-autonomy principle (don''t pretend to know what we haven''t measured).';

COMMENT ON COLUMN stewards.tool_defs.expected_invocation_tokens IS
'Phase 3c.2.5: typical token weight of one tool invocation (args + dispatch overhead). NULL = unknown.';

-- ---------------------------------------------------------------------
-- Underlying SQL function: study_search_text
--
-- FTS over studies.body_tsv (GIN-indexed since Phase 2.1). Multi-kind
-- filter via array. Empty array = all kinds. websearch_to_tsquery
-- handles natural-language queries better than to_tsquery.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.study_search_text(
    p_query text,
    p_kinds text[] DEFAULT ARRAY[]::text[],
    p_limit int DEFAULT 10
) RETURNS TABLE (
    slug    text,
    kind    text,
    title   text,
    snippet text,
    rank    real
)
LANGUAGE sql STABLE AS $func$
    SELECT s.slug,
           s.kind,
           s.title,
           ts_headline('english', coalesce(s.body, ''), q,
                       'MaxWords=20, MinWords=10, ShortWord=3') AS snippet,
           ts_rank(s.body_tsv, q) AS rank
      FROM stewards.studies s,
           websearch_to_tsquery('english', p_query) q
     WHERE s.body_tsv @@ q
       AND (cardinality(p_kinds) = 0 OR s.kind = ANY(p_kinds))
     ORDER BY rank DESC
     LIMIT greatest(p_limit, 1);
$func$;

COMMENT ON FUNCTION stewards.study_search_text(text, text[], int) IS
'Phase 3c.2.5: FTS over stewards.studies.body_tsv. Multi-kind filter via array (empty = all). Ordered by ts_rank.';

-- ---------------------------------------------------------------------
-- Underlying SQL function: study_get
--
-- Returns the doc + frontmatter + citation count + (optional) body
-- with line-based pagination. Line slicing avoids mid-word splits;
-- max_body_chars is the safety cap that wins if a slice is dense.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.study_get(
    p_slug          text,
    p_include_body  boolean DEFAULT true,
    p_line_offset   int     DEFAULT 0,
    p_line_count    int     DEFAULT 200,
    p_max_chars     int     DEFAULT 20000
) RETURNS jsonb
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_study           stewards.studies%ROWTYPE;
    v_lines           text[];
    v_total_lines     int;
    v_actual_count    int;
    v_body_slice      text;
    v_truncated       bool := false;
    v_citation_count  int;
    v_result          jsonb;
BEGIN
    SELECT * INTO v_study FROM stewards.studies WHERE slug = p_slug;
    IF v_study.id IS NULL THEN
        RETURN jsonb_build_object(
            'error', format('study not found: %s', p_slug));
    END IF;

    SELECT count(*)::int INTO v_citation_count
      FROM stewards.study_citations(p_slug);

    v_result := jsonb_build_object(
        'slug',           v_study.slug,
        'kind',           v_study.kind,
        'title',          v_study.title,
        'frontmatter',    coalesce(v_study.frontmatter, '{}'::jsonb),
        'citation_count', v_citation_count
    );

    IF p_include_body THEN
        v_lines := string_to_array(coalesce(v_study.body, ''), E'\n');
        v_total_lines := cardinality(v_lines);

        IF p_line_offset < 0 THEN p_line_offset := 0; END IF;
        IF p_line_count < 1  THEN p_line_count  := 200; END IF;

        v_actual_count := least(
            p_line_count,
            greatest(0, v_total_lines - p_line_offset)
        );

        IF v_actual_count > 0 THEN
            v_body_slice := array_to_string(
                v_lines[p_line_offset + 1 : p_line_offset + v_actual_count],
                E'\n'
            );
        ELSE
            v_body_slice := '';
        END IF;

        IF p_max_chars > 0 AND length(v_body_slice) > p_max_chars THEN
            v_body_slice := substring(v_body_slice FROM 1 FOR p_max_chars);
            v_truncated  := true;
        END IF;

        v_result := v_result
            || jsonb_build_object(
                'body',                    v_body_slice,
                'body_line_offset',        p_line_offset,
                'body_lines_returned',     v_actual_count,
                'body_total_lines',        v_total_lines,
                'body_truncated_by_chars', v_truncated
            );
    ELSE
        -- Surface the line count even when body is omitted, so the
        -- agent can decide whether to fetch and at what offset.
        v_lines := string_to_array(coalesce(v_study.body, ''), E'\n');
        v_result := v_result
            || jsonb_build_object(
                'body_total_lines', cardinality(v_lines)
            );
    END IF;

    RETURN v_result;
END;
$func$;

COMMENT ON FUNCTION stewards.study_get(text, boolean, int, int, int) IS
'Phase 3c.2.5: read a doc + frontmatter + citation count + (optional) body with line-based pagination. Mirrors the Read tool''s offset/limit semantics. Returns jsonb.';

-- ---------------------------------------------------------------------
-- Tool wrappers (jsonb → jsonb)
--
-- All decode args from the model''s tool_call.arguments jsonb, apply
-- defaults for omitted fields, and call the underlying typed function.
-- ---------------------------------------------------------------------

-- 1. study_search_text_tool
CREATE OR REPLACE FUNCTION stewards.study_search_text_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql STABLE AS $func$
    SELECT coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)
    FROM stewards.study_search_text(
        p_args->>'query',
        coalesce(
            (SELECT array_agg(value::text)
               FROM jsonb_array_elements_text(coalesce(p_args->'kinds', '[]'::jsonb)) AS value),
            ARRAY[]::text[]
        ),
        coalesce((p_args->>'limit')::int, 10)
    ) t;
$func$;

-- 2. study_get_tool
CREATE OR REPLACE FUNCTION stewards.study_get_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql STABLE AS $func$
    SELECT stewards.study_get(
        p_args->>'slug',
        coalesce((p_args->>'include_body')::boolean, true),
        coalesce((p_args->>'body_line_offset')::int, 0),
        coalesce((p_args->>'body_line_count')::int, 200),
        coalesce((p_args->>'max_body_chars')::int, 20000)
    );
$func$;

-- 3. study_similar_tool
CREATE OR REPLACE FUNCTION stewards.study_similar_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql STABLE AS $func$
    SELECT coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)
    FROM stewards.study_similar(
        p_args->>'slug',
        coalesce((p_args->>'limit')::int, 5)
    ) t
    WHERE coalesce((p_args->>'min_score')::float, 0.0) <= t.score;
$func$;

-- 4. study_citations_tool
CREATE OR REPLACE FUNCTION stewards.study_citations_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql STABLE AS $func$
    SELECT coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)
    FROM stewards.study_citations(p_args->>'slug') t;
$func$;

-- 5. study_context_for_tool
CREATE OR REPLACE FUNCTION stewards.study_context_for_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql STABLE AS $func$
    SELECT coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)
    FROM stewards.context_for(
        p_args->>'slug',
        coalesce((p_args->>'depth')::int, 2)
    ) t;
$func$;

-- ---------------------------------------------------------------------
-- tool_defs registrations
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target)
VALUES
(
    'study_search_text',
    'Full-text search over the substrate''s document corpus (studies, journals, proposals, docs, phase-doc). Returns ranked matches with slug, kind, title, snippet, and ts_rank score. Use this to find docs by topic before reading them with study_get. Filter to specific kinds via the `kinds` array (empty = all). Backed by Postgres FTS over body_tsv.',
    '{
        "type": "object",
        "required": ["query"],
        "properties": {
            "query":  {"type": "string", "minLength": 1, "maxLength": 200,
                       "description": "Natural-language search terms. Phrases in quotes are matched verbatim."},
            "kinds":  {"type": "array",
                       "items": {"type": "string",
                                 "enum": ["study","doc","proposal","journal","phase-doc"]},
                       "description": "Filter to one or more kinds. Empty/omitted = search all kinds."},
            "limit":  {"type": "integer", "minimum": 1, "maximum": 20,
                       "description": "Max results (default 10)."}
        }
    }'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"study_search_text_tool"}'::jsonb
),
(
    'study_get',
    'Read a doc by slug. Returns title, frontmatter, citation count, and body with line-based pagination. The body slice is bounded by `body_line_count` (line-aligned, no mid-word splits) AND `max_body_chars` (hard cap that wins if the slice is dense). For long docs, paginate via `body_line_offset = previous_offset + body_lines_returned` until `body_total_lines` is reached. Set `include_body=false` to fetch only metadata + total line count.',
    '{
        "type": "object",
        "required": ["slug"],
        "properties": {
            "slug":             {"type": "string", "description": "Doc slug (e.g. \"charity\", \"proposal-token-efficiency\")."},
            "include_body":     {"type": "boolean", "description": "Default true. Set false for metadata only."},
            "body_line_offset": {"type": "integer", "minimum": 0, "description": "Lines to skip before the slice (default 0)."},
            "body_line_count":  {"type": "integer", "minimum": 1, "maximum": 1000, "description": "Max lines per call (default 200)."},
            "max_body_chars":   {"type": "integer", "minimum": 100, "maximum": 50000, "description": "Hard char cap on the returned slice (default 20000)."}
        }
    }'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"study_get_tool"}'::jsonb
),
(
    'study_similar',
    'Return docs semantically similar to the given slug, using precomputed pgvector cosine similarity edges (Phase 2.3). No on-the-fly embedding; cheap. Each result has a score (0..1, higher = more similar) and direction (outgoing | incoming | mutual). Use after study_search_text to expand a topic''s neighborhood.',
    '{
        "type": "object",
        "required": ["slug"],
        "properties": {
            "slug":      {"type": "string"},
            "limit":     {"type": "integer", "minimum": 1, "maximum": 10, "description": "Max neighbors (default 5)."},
            "min_score": {"type": "number",  "minimum": 0,  "maximum": 1, "description": "Filter results below this score."}
        }
    }'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"study_similar_tool"}'::jsonb
),
(
    'study_citations',
    'Return the canonical sources (scriptures, talks, manuals) cited by a doc. Backed by AGE :CITES edges parsed from markdown links during import. Returns cited_uri (workspace path), cited_kind (scripture | talk | manual | reference), anchor_text (the link text the doc used), and citation_count (how many times that uri appears).',
    '{
        "type": "object",
        "required": ["slug"],
        "properties": {
            "slug": {"type": "string"}
        }
    }'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"study_citations_tool"}'::jsonb
),
(
    'study_context_for',
    'Walk the AGE graph outward from a doc, returning typed-edge neighbors up to `depth` hops. Surfaces structural connections (Workstream, Proposal, Phase, Todo) and semantic ones (CITES, FEEDS, REFINES, IMPLEMENTS, SIMILAR_TO). Use this when "what''s connected to X?" is the question; use study_similar when only semantic similarity is needed.',
    '{
        "type": "object",
        "required": ["slug"],
        "properties": {
            "slug":  {"type": "string"},
            "depth": {"type": "integer", "minimum": 1, "maximum": 4, "description": "Hops to walk (default 2). Capped at 4."}
        }
    }'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"study_context_for_tool"}'::jsonb
)
ON CONFLICT (name) DO UPDATE
SET description    = EXCLUDED.description,
    args_schema    = EXCLUDED.args_schema,
    execute_target = EXCLUDED.execute_target;

-- ---------------------------------------------------------------------
-- Grant study_* tool perms to applicable imported agents.
--
-- Background: 3a.1's agent import preserved the Copilot frontmatter
-- `tools:` list verbatim as agent_tool_perms rows (e.g.,
-- `gospel-engine-v2/*: allow`). Those patterns are aspirational —
-- they match Copilot/MCP tool names that don't exist in the substrate
-- yet. Substrate-internal tools like study_search_text don't match
-- those patterns, so the deny-* fallback wins and compose_tools
-- emits nothing useful.
--
-- We blanket-allow `study_*` across all non-watchman agents. The
-- tools are read-only over substrate state; there's no destructive
-- risk in granting broad access. Watchman's deny-everything pattern
-- is preserved (it ships with its own tools=none design).
--
-- glob_match: `study_*` is length 7, beats `*: deny` (length 1)
-- via the longest-match-wins resolver.
-- ---------------------------------------------------------------------
-- Tagged source='broadcast' (per 3c3-3) so the importer's reimport-DELETE
-- (filtered by source='frontmatter') doesn't wipe it. ON CONFLICT updates
-- only `action` to avoid downgrading a row that the agent's frontmatter
-- has since declared explicitly.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
SELECT DISTINCT a.family, 'study_*', 'allow', 'broadcast'
  FROM stewards.agents a
 WHERE a.family NOT LIKE 'watchman%'
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
SET action = EXCLUDED.action;
