-- Inverse hypothesis: a tool that ereports must NOT crash the bgworker.
-- The error must surface as a role='tool' message so the model can recover.

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target)
VALUES (
    'always_fails',
    'A tool that always errors, for testing the error path. The function it points at does not exist.',
    '{"type":"object","properties":{}}'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"nonexistent_function"}'::jsonb
) ON CONFLICT (name) DO NOTHING;

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES ('stewards-explore', 'always_fails', 'allow')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;

-- First a sanity success path
INSERT INTO stewards.sessions (id, label, kind)
VALUES ('loop-3', 'success path post-fix', 'chat')
ON CONFLICT (id) DO NOTHING;

SELECT stewards.chat_enqueue(
    'stewards-explore', 'kimi-k2.6', 'loop-3',
    'In one sentence, name two virtues from Moroni 7.', 'opencode_go'
) AS work_id_success;

-- Then the inverse: ask kimi to call the broken tool
INSERT INTO stewards.sessions (id, label, kind)
VALUES ('loop-err2', 'inverse hypothesis post-fix', 'chat')
ON CONFLICT (id) DO NOTHING;

SELECT stewards.chat_enqueue(
    'stewards-explore', 'kimi-k2.6', 'loop-err2',
    'Please call the always_fails tool with no arguments and report what happens.',
    'opencode_go'
) AS work_id_inverse;
