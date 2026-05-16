-- ES.3.s3 synthetic smoke — consult_subagent_dispatch. Rolled back.
-- Lives in extension/smoke/ (subdir — the migration ledger skips it).
\set ON_ERROR_STOP on
BEGIN;

-- A fake judged message + its preserved document + judge session.
INSERT INTO stewards.sessions (id, kind, label)
VALUES ('es8-host', 'chat', 'es8 smoke host');
INSERT INTO stewards.messages (id, session_id, role, content)
VALUES (9990001, 'es8-host', 'tool', '[JUDGE BRIEF]\nstate: done\n• [hot] x');
INSERT INTO stewards.messages_raw_overflow
    (message_id, parent_ordinal, content, byte_size, tool_name, binding_question, content_sha256)
VALUES (9990001, 0, repeat('the preserved document body. ', 400), 11600,
        'fetch_md', 'what does the doc say?', 'deadbeef');
INSERT INTO stewards.sessions (id, kind, label)
VALUES ('judge-9990001', 'tool', 'judge brief for message 9990001');
INSERT INTO stewards.messages (session_id, role, content)
VALUES ('judge-9990001', 'assistant',
        '{"engrams":[],"state":"done","discarded":"the prior brief"}');

\echo '--- consult #1 (under soft cap) ---'
SELECT stewards.consult_subagent_dispatch('judge-9990001', 'what about temperature?') AS chat_wq \gset
SELECT 'C1 chat enqueued: ' || (:chat_wq IS NOT NULL)::text;
SELECT 'C2 [CONSULT] user msg recorded: ' || count(*)::text
  FROM stewards.messages WHERE session_id='judge-9990001' AND role='user' AND content LIKE '[CONSULT]%';
SELECT 'C3 body has document: ' ||
       ((payload#>>'{body,messages}') LIKE '%preserved document body%')::text
  FROM stewards.work_queue WHERE id = :chat_wq;
SELECT 'C4 body has follow-up question: ' ||
       ((payload#>>'{body,messages}') LIKE '%what about temperature%')::text
  FROM stewards.work_queue WHERE id = :chat_wq;
SELECT 'C5 reask index: ' || (payload->>'_consult_reask_index')
  FROM stewards.work_queue WHERE id = :chat_wq;
SELECT 'C6 model deepseek-v4-flash, no max_tokens: ' ||
       ((payload#>>'{body,model}')='deepseek-v4-flash'
        AND NOT (payload#>'{body}' ? 'max_tokens'))::text
  FROM stewards.work_queue WHERE id = :chat_wq;

\echo '--- soft cap: simulate 5 prior consults, expect NOTICE on #6 ---'
INSERT INTO stewards.messages (session_id, role, content)
SELECT 'judge-9990001', 'user', '[CONSULT] prior ' || g
  FROM generate_series(1,4) g;
SELECT stewards.consult_subagent_dispatch('judge-9990001', 'sixth question') AS chat6 \gset
SELECT 'C7 soft-cap NOTICE injected on 6th: ' ||
       ((payload#>>'{body,messages}') LIKE '%STEWARD NOTICE — soft cap%')::text
  FROM stewards.work_queue WHERE id = :chat6;

ROLLBACK;
\echo '--- rolled back ---'
