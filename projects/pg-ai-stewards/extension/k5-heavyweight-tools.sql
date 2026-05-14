-- =====================================================================
-- Batch K.5 — Heavyweight wrapper tools (minimum-viable: deep_research)
-- =====================================================================
-- Of the 7 ratified wrapper tools (deep_research, summarize_url,
-- audit_files, investigate_session, summarize_study, investigate_study,
-- audit_studies), this migration ships deep_research as the proof-of-
-- pattern. It uses the existing research-write pipeline (no new
-- pipeline needed), so it's immediately useful for J.3-shaped retry
-- workloads.
--
-- The other 6 wrappers each follow the SAME shape:
--   1. Go file: <wrapper>.go with WrapperInput struct + makeWrapper
--      handler that constructs binding_question + picks pipeline_family
--      + calls makeSpawnSubagent (already exists).
--   2. tool_defs row mapping the wrapper name to kind=mcp_proxy,
--      server=pg-ai-stewards.
--   3. Optional: new pipeline if the wrapper needs a tightly-scoped
--      single-stage pipeline (e.g. summarize_url, audit_files).
--
-- See journal 2026-05-14 for the full build pattern. Adding each of
-- the remaining 6 is ~15-30 lines Go + 1-3 SQL rows.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. deep_research tool definition.
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'deep_research',
    'Delegate broad multi-source research to a sub-agent running the research-write pipeline. ' ||
    'Returns a sourced prose digest with verbatim URLs / dates / quotes preserved (covenant cite chain). ' ||
    'Use for: topics requiring 3+ web sources, comparison across vendors / studies, historical lineage. ' ||
    'DO NOT use for: a single URL fetch, or work you can answer with one web_search call.',
    $JSON$
    {
      "type": "object",
      "required": ["topic"],
      "additionalProperties": false,
      "properties": {
        "topic": {
          "type": "string",
          "description": "The subject to research (5-20 words; the binding question will be built around this)."
        },
        "focus": {
          "type": "string",
          "description": "Optional narrowing focus (e.g. 'safety considerations only' or 'pre-1960 history only')."
        },
        "cost_cap_micro": {
          "type": "integer",
          "default": 1500000,
          "description": "Max micro-dollars (default $1.50)."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'deep_research'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of k5-heavyweight-tools.sql
-- =====================================================================
