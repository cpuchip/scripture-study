INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target)
VALUES (
    'always_fails',
    'A tool that always errors, for testing the error path.',
    '{"type":"object","properties":{}}'::jsonb,
    '{"kind":"sql_fn","schema":"stewards","name":"nonexistent_function"}'::jsonb
) ON CONFLICT (name) DO NOTHING;

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES ('stewards-explore', 'always_fails', 'allow')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;

INSERT INTO stewards.sessions (id, label, kind)
VALUES ('loop-err', 'inverse hypothesis', 'chat')
ON CONFLICT (id) DO NOTHING;

SELECT stewards.chat_enqueue(
    'stewards-explore',
    'kimi-k2.6',
    'loop-err',
    'Please call the always_fails tool with no arguments.',
    'opencode_go'
) AS work_id;
