-- SCRATCH (not a migration — CT2.4 RUN 2 experiment apparatus, 2026-06-09).
-- codewright-ct2 = codewright + context tools + scaffolding (treatment).
-- Control = live codewright (unchanged). Drive both on a long research-heavy
-- room session and compare: context-size curve, cost, lever usage, answer quality.

INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, kind, context_tools_enabled)
SELECT 'codewright-ct2', model_match, description || ' [CT2 RUN2 treatment]', mode,
       prompt || E'\n\nMANAGING YOUR OWN CONTEXT: this is one long-lived session — the whole room conversation plus every research result piles up. Keep it lean so you stay fast and focused:\n- After you have delivered a cited answer from a research_codebase result, mute that bulky tool result with context_mute(handle) — you already captured what mattered in your reply.\n- pin (context_pin) a fact you will reuse across the conversation (e.g. a repo''s auth flow).\n- remember() a durable cross-session fact (e.g. where a repo keeps X) so a future session has it too; forget() it once integrated.\nHandles like [ctx:ab12] appear next to messages — pass them to these tools. Do this housekeeping briefly between turns; never let it crowd out actually answering the question.',
       temperature, kind, true
  FROM stewards.agents WHERE family = 'codewright'
ON CONFLICT (family, model_match) DO UPDATE
   SET context_tools_enabled = true, prompt = EXCLUDED.prompt, kind = EXCLUDED.kind, active = true;

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
SELECT 'codewright-ct2', tool_pattern, action, source
  FROM stewards.agent_tool_perms WHERE agent_family = 'codewright'
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = EXCLUDED.action;

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
('codewright-ct2','context_compress','allow','manual'),
('codewright-ct2','context_mute','allow','manual'),
('codewright-ct2','context_expand','allow','manual'),
('codewright-ct2','context_pin','allow','manual'),
('codewright-ct2','context_unpin','allow','manual'),
('codewright-ct2','remember','allow','manual'),
('codewright-ct2','forget','allow','manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = EXCLUDED.action;

INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled, file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
SELECT 'persona-turn-code-ct2', description || ' [CT2 RUN2]',
       (SELECT jsonb_agg(jsonb_set(s, '{agent_family}', '"codewright-ct2"')) FROM jsonb_array_elements(stages) s),
       sabbath_enabled, atonement_enabled, file_destination_template, file_content_jsonpath, maturity_ladder, false,
       COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('ct2_run2', 'treatment')
  FROM stewards.pipelines WHERE family = 'persona-turn-code'
ON CONFLICT (family) DO UPDATE SET stages = EXCLUDED.stages, metadata = EXCLUDED.metadata;

SELECT 'codewright-ct2 ctx_on=' || stewards.context_tools_on('codewright-ct2')::text AS check1;
SELECT 'context levers: ' || string_agg(t->'function'->>'name', ', ' ORDER BY 1) AS check2
  FROM jsonb_array_elements(stewards.compose_tools('codewright-ct2')) t
 WHERE t->'function'->>'name' LIKE 'context%' OR t->'function'->>'name' IN ('remember','forget');
SELECT 'pipeline agent=' || (stages->0->>'agent_family') AS check3 FROM stewards.pipelines WHERE family='persona-turn-code-ct2';
