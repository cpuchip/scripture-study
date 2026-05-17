-- =====================================================================
-- ES.5.s3 — Grant consult_subagent to all pipeline agents
-- =====================================================================
-- consult_subagent (ES.3.s3) shipped but inert — agent_tool_perms is
-- deny-by-default and no agent had the grant. The ES.5 council
-- ratified granting it to ALL pipeline agents: any pipeline-stage
-- agent may re-engage a persistent sub-agent (a judge that compiled a
-- brief, or a spawned child) with a new question.
--
-- The grant is derived straight from the pipelines table — every
-- distinct agent_family referenced by any pipeline stage. A specific
-- 'allow' row overrides the per-agent 'deny *' baseline (the resolver
-- resolves by specificity; debug already runs this way with '* deny'
-- plus specific allows).
--
-- New pipelines added later need their own grant — granting is a
-- deliberate act under the deny-by-default model, not automatic.
-- =====================================================================

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
SELECT DISTINCT
       stage ->> 'agent_family',
       'consult_subagent',
       'allow',
       'manual'
  FROM stewards.pipelines,
       jsonb_array_elements(stages) AS stage
 WHERE stage ->> 'agent_family' IS NOT NULL
ON CONFLICT (agent_family, tool_pattern)
   DO UPDATE SET action = 'allow', source = 'manual';

-- =====================================================================
-- End of es10-grant-consult-subagent.sql
-- =====================================================================
