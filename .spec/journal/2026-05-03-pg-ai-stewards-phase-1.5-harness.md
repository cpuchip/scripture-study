# 2026-05-03 (Session F) — pg-ai-stewards Phase 1.5: harness sketch (detour)

Two sessions in one day. After step 6 landed real LM Studio embeddings this morning, Michael's question from yesterday — *"in brain we relied on github-copilot-sdk to do the agentic work; here that's on us right?"* — was the next thing to answer. I gave him three options after step 6: stay on plan and eat the gap, insert a Phase 1.5 harness sketch, or pause and think. He chose harness sketch. The right call.

## What we did

Read OpenCode's docs first. He pointed me at anomalyco/opencode — open-source coding agent, well-documented, has the exact same shape problems we're about to have. Three things worth borrowing landed in the schema:

1. **Skills are not injected into the system prompt.** They're advertised in an `<available_skills>` XML block inside the `skill` tool's description, and the agent calls `skill({name})` to load a body when it needs one. Token-efficient. The brain.exe approach was to stuff everything in; OpenCode learned to be more surgical, and we should too.
2. **The agent IS its config row.** `(name, mode, prompt, temperature, top_p, steps, permissions)`. Subagent invocation is just another tool call. There's no separate "agent runtime" object — the loop reads the row and runs.
3. **Tool name = `<prefix>_<name>` is universal.** Permissions glob on prefix. `brain_*: allow`. This is the right convention from day one because it makes per-server permission management trivial.

What I deliberately didn't borrow: filesystem-coupled tool context (we're a DB), real MCP client (v1 "MCP equivalent" is `execute_target: {kind:'sql_fn', name:...}` — a row that says how to dispatch), and `steps` enforcement (column exists, loop doesn't yet).

Then Michael added the contribution that mattered most for the long-term shape: **different models reason about the same instructions differently.** Kimi over-explains; GPT-5 ignores temperature; Qwen wants different defaults. We need a way to ship the same agent against different models without duplicating workflow rules. We considered a separate variants table, a JSON column, an inheritance/diff system. The cleanest answer was a single column added to three tables: `model_match` glob (`'kimi-*'`, with `'*'` as catch-all). Resolver picks longest match. **Tools deliberately don't get variants** — a tool's description is structural ("what does this do"), not stylistic.

The implementation almost shipped with a real bug. I had `model_match text` nullable and `PRIMARY KEY (family, model_match)`. PG requires NOT NULL on PK columns — `CREATE TABLE` would have failed at boot. I caught it during code review (not at build) by reasoning about what `NULL` actually means in three places: PK constraint, ON CONFLICT clause, and resolver SQL. Switching from NULL-as-fallback to `'*'`-as-sentinel fixed all three at once because `'*'` glob-matches everything through the existing `glob_match` function, so the resolver loses a special case AND the PK works AND ON CONFLICT is honest. Pattern worth remembering: when NULL causes friction in three places, reach for a sentinel that satisfies the type system AND unifies the function logic.

The deliverable is `stewards.dry_run_chat(family, model, session, input)` returning the exact JSON body that would POST to `/v1/chat/completions`. No HTTP. No agent loop. Just composition. The point is to *look* at what we'd send before we send it.

## What I'm proudest of

The verification numbers are the kind I trust because they didn't have to come out this clean:

```
Kimi system prompt:  1049 chars  agent_variant_match: 'kimi-*'
GPT-5 system prompt:  963 chars  agent_variant_match: '*'
                       86 char delta = the "be terse" paragraph
```

Same instructions block, same `<available_skills>`, same tools[], same temperature. Persona is the *only* delta. This is the right layering — agent.prompt is the persona ("Kimi-specific: be terse"), instructions are the workflow rules ("read before quoting"). The model variant changes one and not the other.

The tools[] array came out canonical without me having to massage it: `{type: function, function: {name, description, parameters}}` with JSON Schema fully intact (enum values, min/max, required fields). The `<available_skills>` block is alphabetized name + description, no body bloat — exactly what OpenCode does, exactly what Kimi or GPT-5 will know how to read.

Inverse hypothesis: `dry_run_chat('does-not-exist', ...)` raises `no agent variant resolved: family=does-not-exist model=gpt-5.1`. Clean failure. Message points to the fix. This is what the bgworker step 6 also did and it's becoming a habit — the error path carries information, the success path explicitly clears prior errors. Two failures in a row that both helped me trust the system more.

## What surprised me

How small the schema turned out. I'd been imagining "agent harness" as a major surface — runtime, dispatcher, message router, tool registry, permission engine. The actual schema is six tables and seven functions, and four of those functions are 5-line SQL helpers. The compose functions are pure data shaping. There's no "engine" — there's just *queries that return the bytes you'd send*. Step 7 becomes "POST the bytes, parse the response, write the rows."

This makes me think the chat loop will also be small. The dispatcher for `tool_defs.execute_target` is a `match kind { 'sql_fn' => ..., 'http' => ... }` block. The "agentic loop" is `while not finished and steps < limit { send, parse, dispatch tool_calls, append, repeat }`. Maybe 200 lines of Rust in the bgworker. We'll see.

The other surprise was how much OpenCode's "skills advertised in tool description, loaded on demand" pattern helped clarify our `.github/skills/` directory's purpose. We have ~25 skill files at this point, many of which are 200+ lines. If those all went into every system prompt, we'd be torching 30k tokens per call before the user said anything. The advertise-and-load pattern means a session that doesn't need `byu-citations` never pays for `byu-citations`. Worth back-porting this insight even into the brain.exe / copilot-sdk world.

## What carried (intentionally) not built

- **No agent loop.** Single-turn dry run only. Step 7 (or "Phase 1.6") will close this.
- **No real tool execution.** `execute_target` is data; nothing reads it yet.
- **No real MCP client.** "MCP equivalent" in v1 is the `execute_target` shape. Real MCP transport comes when we want to consume gospel-engine's MCP from inside stewards.
- **No `steps` enforcement.** Column exists; loop that respects it doesn't.
- **No session-scoped instructions.** Schema supports it (`scope='session:<id>'`); nothing writes them yet. Will matter when we want a per-session "stay focused on this PR" instruction without polluting the agent definition.

## Carry forward

- **Step 7 next session.** Now that the body is shippable, wire one round-trip: pick a session, call `dry_run_chat`, POST to OpenCode Go's `/chat/completions` (kimi-k2.6), parse the response, append the assistant message to `stewards.messages`. No tool execution yet — just verify the chat round-trip lands and we can read what came back.
- **Then "Phase 1.6" agent loop.** Once chat round-trips, add the dispatcher for `execute_target.kind='sql_fn'` (call the named function, capture result, append as `{role:'tool', content:result}`), and the loop guard (`while assistant.tool_calls and steps < agent.steps`).
- **Migration story still owed.** `docker compose down` (no -v) preserves the volume, so schema changes need either drop-volume-rebuild (dev) or `ALTER EXTENSION ... UPDATE` with versioned migration files (production). I keep deferring this. By step 7 it'll start mattering because we'll have real chat history we don't want to throw away.
- **Skills back-port.** OpenCode's "advertise in tool description, load on demand" pattern is so much better than our current `.github/skills/SKILL.md` injection that I want to reconsider how copilot-sdk handles the skill catalog. Not urgent. Note for the next session that touches that.
- **Verification SQL files cleaned up.** Removed `verify-phase-1.5-a.sql` and `-b.sql` from `extension/`. Future verification sessions should write to a temp dir, not the source tree.

Two sessions in one day, both clean, both with inverse-hypothesis verification. The harness has a body. Tomorrow we send it.
