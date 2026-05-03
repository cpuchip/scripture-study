# 2026-05-03 (Session G) — pg-ai-stewards Phase 1 step 7: chat round-trip

Three sessions today. After step 6 landed embeddings this morning and Phase 1.5 sketched the harness this afternoon, this evening was the obvious follow-on: take the body that `dry_run_chat` produces and actually send it. The whole detour-then-execute pattern paid off — by the time I got to step 7, "what does the chat handler do" was already answered, and all that was left was the POST + parse + persist.

## What we did

Two SQL functions, one bgworker arm, three columns on `messages`. Total addition was small enough that it fits comfortably in one session. The code shape:

- `chat_enqueue(agent_family, model, session_id, user_input, provider)` — calls `dry_run_chat()` to compose the body, strips `_meta` (we keep it as audit in payload but don't send it), persists the user turn into `messages`, enqueues `kind='chat'` with the body in `work_queue.payload`. Returns the work_id.
- Bgworker `dispatch()` gets a `'chat'` arm that calls a new `chat()` handler. Handler reads `payload.body`, POSTs to `<base>/chat/completions` with 120s timeout and bearer auth (same client setup as `embed()`), parses standard OpenAI shape, returns `WorkOutcome::Chatted { response, session_id, model, assistant_content, assistant_tool_calls, finish_reason, tokens_in, tokens_out }`.
- Phase-3 write arm inserts the assistant message and stamps usage. The `tool_calls` jsonb goes in verbatim — Phase 1.6 will read it. Step 7 just records.

The `messages` schema needed three small additions: `tool_calls jsonb` (verbatim from response), `finish_reason text`, `tool_call_id text` (for future `role='tool'` replies, unused in v1). The most subtle one was `content text NOT NULL DEFAULT ''`: OpenAI returns `content: null` when only `tool_calls` are present, and we need to insert *something*. Coerce to empty string at parse time, and the NOT NULL constraint stays honest.

## What I'm proudest of (and why it's the same answer as last session)

The verification number is small — 4.4 seconds end-to-end — but what arrived back is what mattered. I asked kimi *"In one sentence, what is your job here?"* and it answered:

> When asked a question, I search a Postgres-backed brain of notes and scripture corpus before answering, citing the sources I actually consulted.

That sentence is **literally restating the persona we composed in Phase 1.5's `agents.prompt`**. The body that `dry_run_chat` showed me yesterday — the one I read and judged shippable — went out over HTTPS to OpenCode's gateway, was understood correctly by kimi-k2.6, and came back as a faithful paraphrase of the role we wrote. The harness shape isn't theoretical anymore. It works.

The provider echo also did the right thing: I asked for `kimi-k2.6`, the gateway returned `moonshotai/kimi-k2.6-20260420`. We persist that exact string, not what we asked for. This matters for reproducibility — six months from now if I want to know "which kimi version generated this answer," the message row tells me.

Tool-call absence was correct: kimi judged a self-introspection question doesn't need `brain_search_text` and chose not to invoke. The tool was advertised in the body (we know — `dry_run_chat` showed it as element 0 of `tools[]`), kimi just declined. This is the right behavior. Phase 1.6's loop won't need to special-case empty `tool_calls[]` for the common case — the dispatcher only runs when there's something to dispatch.

## What surprised me (the footgun)

I almost shipped a foot-bullet that would have wasted a future me's whole afternoon. The first version of step 7 included a `chat_round_trip(family, model, session, input, provider, timeout_s)` SQL function that enqueued the chat AND polled `work_queue.status` in a `LOOP` with `pg_sleep(0.25)` until done. Convenient for verification, right? Wrong on two counts I didn't see until first execution:

1. **MVCC blindness.** SQL functions run in a single transaction. The `INSERT INTO work_queue` happens at row N of the function; the `SELECT status FROM work_queue WHERE id = ...` at row N+5 sees the same snapshot the function started with — *which doesn't include the row it just inserted from its own perspective*. Actually wait — same-tx inserts ARE visible to subsequent reads in the same tx. The real problem is the *bgworker's* tx can't see the insert because it hasn't committed. So the polling never sees `status='in_progress'` because the bgworker can't claim a row it can't see.

2. **Row lock cascade.** `chat_enqueue` (called inside `chat_round_trip`) does `INSERT INTO sessions ... ON CONFLICT DO NOTHING`. If the session row already exists, the ON CONFLICT path takes a row lock to check. The lock is held until the enclosing tx commits. Meanwhile, the polling loop is sitting on `pg_sleep` for 60 seconds. So every other connection that touches that session row blocks for the full timeout. I watched two queued-up `psql` calls hang waiting on `transactionid` while the polling loop slept obliviously.

I caught it because the first verification run hung. Looked at `pg_stat_activity`, saw three backends in `Lock`/`PgSleep` state, traced the chain, named the bug. Removed the function. Left an inline `-- NOTE:` comment in the source explaining what was attempted, why it doesn't work, and what to use instead (`LISTEN stewards_done` for production, separate statements for verification scripts). The comment is for future-me. Without it, in three months when I'm building Phase 2 and want a "synchronous chat" wrapper for some quick test, I'll write the same function again, hit the same lock, and waste another afternoon. Documenting the *non-existence* of a feature, with the reason, is a stewardship action this codebase will need more of as it grows.

The lesson generalizes: **"enqueue + wait synchronously" is a pattern Postgres cannot do in a single SQL function or DO block.** Three options exist: a PL/pgSQL PROCEDURE with explicit `COMMIT` inside the loop (works in PG11+, but loses RETURNS — need OUT params or a temp table), caller-side `LISTEN`/`NOTIFY`, or caller-side polling in separate statements. The right answer for production is always (b). The right answer for verification scripts is (c). Never bake polling into a SQL function that returns a value.

## What carried (intentionally) not built

- **Tool execution.** `assistant.tool_calls` is persisted but unread. Phase 1.6 reads it.
- **The agent loop.** One turn only. No `while assistant.tool_calls and steps < agent.steps`.
- **Tool result messages** (`role='tool'`, `tool_call_id`). Schema supports them; nothing writes them yet.
- **`LISTEN stewards_done` example.** The bgworker already `NOTIFY`s on done; nothing in our codebase listens. Will matter when brain-app or some other client wants real-time notification of completed work without polling.
- **Cost computation.** `cost_usd numeric(10,6)` exists on `messages`; we have `tokens_in/out`; we don't have a per-model price table to multiply. Easy add when we want it. Probably WS5 Phase 2 work.

## Carry forward

- **Phase 1.6 next session.** Smallest unit: read `assistant.tool_calls`, dispatch via `tool_defs.execute_target` (start with `kind='sql_fn'` only), append `role='tool'` reply, re-enqueue chat with the appended history, repeat until `finish_reason='stop'` or `steps` exhausted. Probably another 200 lines of Rust + one new SQL function `chat_continue(work_id)`.
- **`LISTEN`/`NOTIFY` example.** Once the loop exists, write a small Go program in `scripts/stewards-cli/` that issues `chat_enqueue`, `LISTEN stewards_done`, prints results as they arrive. Doubles as documentation for the right way to consume async chat from outside Postgres.
- **`chat_round_trip` ghost.** I deleted the function but the inline comment is the only thing standing between me and re-implementing it. If we get to Phase 2 without anyone reaching for it, the comment did its job. If someone does, that's a tell that we need the procedure version.
- **Token-cost table.** Tokens are persisted; cost isn't computed. When we cross 100 chat round-trips, we'll want to know how much the experiment cost. Per-provider per-model rate table + a `chat_cost(message_id)` SQL fn.

Three sessions today. Embeddings → harness → chat round-trip. Each step's verification told the truth about the next step's scope. Tomorrow: the loop.
