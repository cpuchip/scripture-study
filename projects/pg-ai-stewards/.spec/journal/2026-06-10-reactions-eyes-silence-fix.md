# Reactions + Eyes shipped; the Codewright/Chattercode SILENCE day

**Date:** 2026-06-10 (morning, Fable 5) · **Mode:** dev
**Trigger:** Michael read chattercode's extension brainstorm in Engineering
(@mentions/reactions/roster-DM), added his own (hover-react, persona 👀 on the
message it's reading), asked for thoughts → "lets spec and ratify."

## Ratified (4 decisions, all as recommended)
`reactions-eyes-mentions.md` (ai-chattermax .spec): durable reactions + history
backfill · fixed-six palette · mentions = alerts AND respond_policy routing ·
full arc R→E→M + roster touch under stewardship.

## REM-1 Reactions — SHIPPED + PROVEN ON PROD (`dcf1d1f`)
0003 migration (one reactor XOR, unique per message/emoji/reactor) · store
add/remove/MessageInChannel guard/batched backfill on room+DM history · gateway
`reaction` frame both directions (broadcast includes sender; idempotent patch) ·
frontend hover ☺+ → palette → chips with counts/tooltips · roster one-click DM
buttons (persona dmEnabled-gated, humans minus self). Store test harness now
applies ALL migrations lexically; full flow proven against live Postgres.
Live e2e: add echoed in 100ms with resolved reactor; REST showed the reaction
while it existed; remove cleared it.

## REM-2 Eyes — SHIPPED + PROVEN LIVE
message.id parsed from frames; 👀 add at turn start, remove on kindDone
(answer, SILENCE, or error), hops on coalesced follow-up. New `rawFn` test seam;
race-clean. Live: **👀 at 0.1s → "🔍 let me check the boot path" at 9s → cited
answer (main.go:45, verified correct) at 48s → 👀 off.** Eyes-off on a real
Fireworks 503 also observed — error path correct.

## The SILENCE day (found during E's live verify — the real story)
Chattercode answered SILENCE to four straight direct questions, fresh session
included. Chased through: session-pattern theory (partial), addressing theory
(necessary, not sufficient — added "(You were directly addressed.)" and it STILL
silenced), until **kimi's reasoning_content gave it up verbatim: "But I am
Codewright, not Chattercode. The message is directed at…"** The host/prompt
character name ("Codewright") differs from the platform display name
("Chattercode") — the persona genuinely believed every question was for someone
else. The model out-reasoned the framing hint.

**Fixes (persona-host `9972e8a`):** isAddressed matches plain slug + platform
name (captured from the gateway `ready` frame); turn-zero framing emits an
identity bridge when names differ ("messages addressed to 'Chattercode' are
addressed to YOU; lines from 'Chattercode' below are your own earlier
messages"). Regression tests for both. Post-fix live run = the happy path above.
Also fixed the "🔍 🔍" mood double-prefix.

**Lessons:**
- *The model's reasoning_content is primary evidence.* Three layers of plumbing
  theories (session patterns, transcript bias, addressing flags) fell to one
  read of what the model actually said to itself. Check it FIRST next time.
- *Identity must be bridged everywhere it's split.* Slug, host display name,
  character prompt name, platform display name — four names, and the one humans
  type matched none of the ones the model knew.
- Queue result shape gotcha: it's `result->response->choices`, NOT
  `result->choices` — my first reads showed "empty content" and nearly sent me
  down a wrong path.

## Carry-forward
- **REM-3 Mentions** (alerts + respond_policy routing + human mood UI) — next PR,
  ratified, not started.
- Residual: SILENCE rows accumulate in long sessions and can bias consults
  (secondary effect, observed only while the identity bug primed it). If quiet
  returns: scrub SILENCE rows or rotate the session.
- Naming hygiene (Michael's call): align persona_host display_name
  ("Codewright") with platform name ("Chattercode") — removes the split at the
  source. The bridge handles it either way.
- Coalesce-duplicate watch item from yesterday: noted in spec; predates the
  async-loop fix; revisit only if it recurs.

Commits: ai-chattermax `38ccd12` (journal correction) `dcf1d1f` (REM-1, pushed =
deployed) `34c2ceb` (spec, pushed); root `9972e8a` (persona-host, UNPUSHED per
preference). Test account flow + Engineering room id in
`.spec/scratch/test-credentials.env`.
