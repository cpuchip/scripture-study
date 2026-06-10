-- R18 — persona turn token budget: 1200 → 3000.
--
-- The 2026-06-10 Holodeck-3 failure: Starlet's turn "completed" with
-- finish_reason=length, content EMPTY, reasoning_content 4817 chars — kimi
-- spent the whole 1200-token budget REASONING about the turn and was cut off
-- before writing a single reply character. The host surfaced it as "session
-- produced no assistant reply". r7's "chat turns are short" was right about
-- the reply and wrong about the thinking: reasoning models bill their thought
-- against max_tokens (the same gotcha as qwen3.6 on LM Studio — give ≥2000).
-- codewright pipelines already ran at 3000 (r13); this brings the
-- conversational personas up to match. Idempotent.
UPDATE stewards.pipelines
SET stages = jsonb_set(stages, '{0,max_tokens}', '3000'::jsonb),
    updated_at = now()
WHERE family IN ('persona-turn', 'persona-turn-lmstudio', 'persona-turn-gemini')
  AND (stages->0->>'max_tokens')::int < 3000;
