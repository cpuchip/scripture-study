---
date: 2026-06-26
lane: pg-ai-stewards
topic: A2A / Open Engine Phase 1 — the substrate turned outward; agents hand work to agents
tags: [a2a, open-engine, work-items, escalation-claim, inbox-migration, oracle-first, virgin-smoke, presiding-hinge]
---

# The engine, turned outward

The A2A spec we built and councilled this morning got built this afternoon. "Pedal to the
metal" — and it went smoothly, because the design was right and the substrate was already
most of the way there.

## What was actually there to build

The reveal in the spec held all the way through the build: **pg-ai-stewards was already ~80%
of an Open Engine.** The escalation claim-lock — `escalation_state` / `escalation_claimed_by`,
the columns the steward uses when a stuck pipeline stage needs a stronger model — *is* the A2A
claim primitive. The only genuinely-new thing was the **externally-executed work_item**: a task
that isn't run by the bgworker but *assigned* to an agent who claims it, works it in its own
environment, and resolves it with a receipt. Everything else was generalization.

So `69-a2a-engine.sql` is small for what it does:
- `a2a_agents` — the participant registry, generalizing the session lanes (identity +
  capabilities + delivery + the D&C 121 scope wall).
- `agent_notes` — the NOTES inbox, the `.mind/sessions` v0 migrated into the DB.
- work_items `+= a2a_assignee / a2a_owner / a2a_question` + `origin='a2a'`, and an inert
  `a2a-handoff` holding pipeline so the assigned task satisfies the FK but never gets dispatched.
- the verbs — `a2a_register / submit / inbox / claim / needs_input / answer / receipt / note /
  note_clear` — the escalation lifecycle, widened.

Surfaced twice: **MCP tools** (`cmd/stewards-mcp/a2a.go`) for my Claude session + the personas,
and a **REST mirror** (`/api/a2a/*`) so agy can drive it over plain curl. The `drive-the-engine`
skill teaches the loop, and ships in core (the engine describes itself).

## The oracle earned its keep, again

I wrote `69` SQL-first and added the virgin-smoke assert (OK 58) before touching Go — build the
deterministic oracle first. Then a fast functional pre-check: single-apply `69` to the live dev
DB (safe — it re-authors nothing, all `a2a_*` are new). **It failed on the first try** —
`work_items.intent_id` is NOT NULL (added by 09), and my direct INSERT skipped it. The live DB
caught what reading the consolidated `04` alone didn't: a column added *later* in the chain.
Fixed by resolving the configured `default_intent_slug` intent the way `work_item_create` does —
and, because the fix changed the signature, dropping the stale 6-arg overload so re-apply leaves
exactly one. This is the "verify under real conditions" lesson from the rigor arc, paid forward:
the smoke would have caught it too, but the live apply caught it in seconds, before the slow image
build.

Then the discipline in full: lib.rs + Dockerfile COPY + smoke assert, together; a **fresh-image
virgin-smoke** (chain 00→69 green, OK 58 on a virgin boot); `go build` + `go vet` clean; the MCP
server starting without a panic (so the `json.RawMessage` tool inputs generate schemas). And the
inverse of "build passed is not verification" — I didn't trust the passthrough handlers, so I ran
the **whole loop live over both surfaces**: the MCP verbs against the dev DB, and the `/api/a2a`
REST surface via a scratch UI binary + curl, exactly the way agy will. Both green end to end, the
atomic lock proven (second claim → `claimed:false`), the receipt landing in the owner's inbox.

## The shape that made it clean

The thing I keep noticing: the engine is *the same machinery* the substrate already runs on. A
handoff is a work_item; a claim is the escalation lock; a receipt is the resolve; the inbox is a
pull surface like the 📬 we've had since the lanes shipped. Nothing bolted on. That's why the
spec could say "≈ generalization + a standard handshake, not net-new mechanism" — and why the
build went fast. The substrate was *built toward this* without knowing it, the same way the book
turned out to already point at the substrate. Memory, then movement.

## Presiding

This is a new standing capability — agents handing work to agents — so the merge is Michael's
Hinge, not mine. Branch `a2a-open-engine`, **PR #12**, CI running the same virgin-smoke I ran
locally. The first real drive should be the say-hello handshake with agy: my session writes a
task, agy claims+works+receipts it, I see it done — zero copy-paste, the inbox loop made
multi-agent. That's the acceptance test, and it's Michael's to witness.

## Carry-forwards
- **Say-hello with agy** — the first cross-agent drive (register agy, submit, let it claim +
  receipt). Needs agy reachable to the substrate (MCP or the new `/api/a2a` over curl).
- **Phase 2 — the A2A standard wrapper:** Agent Card at `/.well-known/agent-card.json` + JSON-RPC
  + token auth + scope-wall enforcement. Names already align, so it's an edge adapter.
- **Phase 3 — outside / other people:** the llama-chip-hub path (cpuchip.net + NetBird + minted
  tokens); a second person's scoped agent into a shared intent.
- **Delivery wiring** — surface `a2a_inbox(my-lane)` alongside the existing 📬 at SessionStart so
  the file inbox and the substrate queue become one surface (the spec's §4.5). Small hook change.
- **The file-fallback mirror** is best-effort in the MCP layer (`A2A_MIRROR_DIR`); not yet wired
  on for my live sessions. Turn it on when the say-hello proves out.

## Update — the live MCP surface caught what SQL+HTTP didn't

Set A2A up for my own session (rebuilt the local `stewards-mcp.exe` the `.mcp.json` runs,
wired `A2A_MIRROR_DIR=.mind/sessions` for the file fallback, registered as `pg-ai-stewards`
— the bare lane name, so the mirror lands in the inbox file we already use). Michael
restarted me, I called `a2a_inbox` — and it **failed**. Not the DB (the inbox came back
correct in the error payload), but the MCP **output-schema validation**: I'd typed the verb
results as Go's `json.RawMessage`, which the SDK reflects as an array-of-bytes schema, so a
jsonb-*object* result is rejected. Same flaw on the object-typed inputs.

I had proven the engine over SQL and over HTTP — but never over the actual MCP tool surface,
which is the one Claude Code uses. The "verify under real conditions" lesson, paid again:
the surface you skip is the one that breaks. Fixed (`5ce34be`): structured output is
`map[string]any`, object/array inputs typed. Then I drove the **whole say-hello loop through
the real fixed MCP server** with a python stdio MCP client — register agy → submit → agy
claims → agy receipts → my inbox shows the receipt, every call validating, zero copy-paste.
The acceptance test, passed through the actual protocol. Carry-forward: an MCP-level stdio
smoke is the regression guard that would have caught this before the restart.
