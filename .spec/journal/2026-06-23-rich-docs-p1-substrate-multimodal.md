# Rich-docs P1 — the substrate carries an image (multimodal carry-through)

**Date:** 2026-06-23 (late) / 2026-06-24
**Session:** pg-ai-stewards lane (post-compact continuation)
**Michael:** "lets roll! pick up P1+"

## What shipped

The load-bearing slice of the rich-docs-in-chat arc (spec ratified 06-23). One
new OSS file does the whole substrate carry-through.

**OSS `2d1b379` — `extension/47-multimodal.sql` (chain → 47), later-file-wins:**
1. `content_parts jsonb` on `messages` — an OpenAI-style content array
   (`[{type:text},{type:image_url}…]`). `content` (text) stays populated as a
   fallback (pressure estimate, history-text, text-only models).
2. `model_capability.supports_vision` bit + `model_supports_vision()` helper.
3. `page_in_cap` re-author — a leading guard: ARRAY content is never truncated
   (the cap does `->> 'content'` then `jsonb_set` a truncated *string* over the
   array → corruption; the guard makes it a pass-through).
4. `compose_messages` re-author — a FIRST CASE branch: a `content_parts` row
   renders as the message `content` VERBATIM (no `[ctx:]` handle prefix, no
   engram/state/injection rewrite). Carries the 15b FINAL body unchanged
   otherwise. **This file now owns compose_messages.**
5. `dispatch_chat_turn` re-author — optional 6th `p_content_parts jsonb`: media
   attached → auto-select the `vision` alias (falls back to the requested
   alias), insert the user turn as `[text part] + media parts`, enqueue via
   `chat_post_internal`. Text-only turns are byte-identical to 45. The 5-arg
   overload is **dropped** (a 5-arg call would be ambiguous with the defaulted
   6-arg).

**No bgworker change.** The OpenAI dispatch path already `body_orig.clone()`s
the messages array verbatim (the de-risk bright spot held), so an array
`content` reaches a vision model untouched. The `:8090` router passes it
through. (`anthropic_body_from_openai` flattens content to a string → Claude
vision is a later phase; P1 routes vision to the local openai-kind rig.)

**Workspace `87bcaf6` — overlays:** a `vision` alias →
`flexllama/gemma-4-26b-a4b` (local, free, no-train, so a `file_private` chat's
image stays local) in `role-aliases.sql`; `supports_vision=true` on gemma +
qwen3.6-35b-a3b in `flexllama-models.sql`.

## Proven both ways

- **Live e2e** (the inverse-hypothesis proof): a First Orbit screenshot injected
  via `dispatch_chat_turn(…, p_content_parts)` → bgworker → `gemma-4-26b-a4b`
  (the `vision` alias, $0/local) → the model **quoted the on-screen telemetry
  verbatim**: ALT 0m, Periapsis −600.0 km, Stage 1/2 4.5k fuel, TWR 0.00. Those
  numbers exist only in the image — the text fallback held only the question —
  so the image demonstrably reached the model *through the substrate*.
- **Virgin-smoke 00→47 green** (fresh `docker build` + `CREATE EXTENSION`): new
  **OK 36** asserts the column + the compose_messages array passthrough + the
  page_in_cap array guard. The whole chain installs clean from scratch.

## Surprises / lessons

- **pg18 `pg_get_function_identity_arguments` includes parameter NAMES** here
  (`p_session_id text, …`), not bare types. The OK 34 assertion's exact
  type-string match (`'text, text, …'`) was brittle and failed in the virgin
  build even though the function was correct. Hardened to `pronargs=6 AND
  proargtypes[5]='jsonb'::regtype` — robust across pg versions. (Worth checking
  whether other smoke assertions use the same brittle pattern.)
- **gemma-4-26b-a4b is a reasoning model:** a vision probe with max_tokens 400
  returned empty content + 2572 reasoning tokens (thinking ate the budget);
  ≥~2500 works. The chat dispatch path uses llama-server's default n_predict
  (fine), but a manual probe must budget for the thinking.
- Re-applying a single end-of-chain file to live is safe *when it carries the
  full FINAL bodies* and nothing supersedes it — 47 does, so the live hot-apply
  (CREATE OR REPLACE + ADD COLUMN IF NOT EXISTS + DROP IF EXISTS) was a clean,
  fast first oracle before the full virgin build.

## Carry-forward

- **P2 (the MVP, first real win):** `chat_attachments` table (session-scoped
  bytea + extracted text) + an upload API + the chat injecting attached image(s)
  as subject + the UI upload control **AND inline media rendering** (Michael's
  explicit add). Works in empty AND work-item chats. Then P3 (docs +
  corpus-as-lens) → P4 (spawn from the combination via start_task).
- **P1 is substrate-only — no UI surface yet** (that's P2). Vision works in
  Stewdio chat at the substrate level (the alias + functions are live).
- The Anthropic dispatch path still flattens arrays → Claude vision is deferred
  (the local rig + Vertex/Gemini openai-compat cover the need).
- Cost: every turn re-sends history's content_parts (the image) to the model —
  acceptable for MVP; a "cap old images" optimization is a later concern.

Autonomy stayed PAUSED throughout (Michael's innovation-week call). The
digester-reads-our-repos council item remains OPEN in the inbox (deferred until
sandboxed; `dominion_in_council`).
