-- R19 — persona turn budgets → 16k (Michael, 2026-06-10, right after r18).
-- "Keep responses small in chat, but comfortable thinking room — and headroom
-- for summarize work — is probably good." Replies stay short by prompt
-- ("2-4 sentences"); the budget stops being the thing that mutes a persona
-- mid-thought. Applies to ALL persona-turn pipelines, including the code ones
-- r13 had at 3000. Idempotent.
UPDATE stewards.pipelines
SET stages = jsonb_set(stages, '{0,max_tokens}', '16000'::jsonb),
    updated_at = now()
WHERE family LIKE 'persona-turn%'
  AND (stages->0->>'max_tokens')::int < 16000;
