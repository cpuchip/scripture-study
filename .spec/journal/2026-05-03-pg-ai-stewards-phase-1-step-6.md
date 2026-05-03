# 2026-05-03 (Session E) — pg-ai-stewards Phase 1 step 6: real LM Studio embeddings

Fifth session in this stretch. The schema from yesterday now has actual vectors in it.

## What we did

Step 6 brief: at least one real provider call through the bgworker, ending with the embedding column filling on insert. Trust the plan, do step 6.

The plan said Ollama. Michael doesn't run Ollama locally; he runs LM Studio. Same OpenAI-compatible `/v1/embeddings` endpoint, same `nomic-embed-text-v1.5` model, same 768 dimensions. Swap the trigger's `provider='ollama'` for `provider='lm_studio'` and the model name from `nomic-embed-text:v1.5` (Ollama's tag style) to `nomic-embed-text-v1.5` (the canonical name gospel-engine-v2 already uses). One-line schema change. Documented in phases.md as a small substitution, not a deviation.

The actual work was in the bgworker: `dispatch(kind, provider, payload)`, an `embed` arm that calls the OpenAI shape, format the f64 array as pgvector's text literal, write back with `$2::vector(768)` so dimension mismatches surface as Postgres errors instead of silent corruption. `reqwest::blocking` with rustls-tls — the worker is already a sync per-tick loop, no tokio runtime needed; rustls keeps the runtime image free of libssl-dev. 120s HTTP timeout because LM Studio's first request after a cold start is 2–3s and we do not want to retry under it.

The thing I'm proudest of is the **three-phase dispatch**. Before step 6 the worker did claim + dispatch + write + notify all inside one transaction, which was fine for the echo stub because echo is microseconds. With real HTTP that pattern would hold a row lock across a 2-second model load and any other queue activity would block. So: Tx A claims and commits, HTTP runs in no transaction, Tx B writes the result and commits. The `process_one_pending` function got slightly longer; the locking behavior got considerably saner. This is the kind of thing that's a one-line bug if you skip it and a one-day debugging session if you discover it under load.

## What I deliberately did not skip

The version-snapshot trigger from step 3 was wrong. It fired on **every** UPDATE — including the bgworker's own UPDATE that just writes the embedding column. Five embed writes would create five junk `brain_versions` rows that say nothing changed except the embedding. I noticed it while reviewing my own code from yesterday, fixed it (gate on title/category/body/props change), bundled the fix into step 6 instead of leaving it for "later." Verified: 5 embed writes → 0 brain_versions rows. This is the stewardship-over-surfacing rule from the dev agent doc working in real time. Same bug, same file, same fix, no behavior change from the user's perspective. Fix it, name it.

The other thing I didn't skip was the inverse hypothesis. The rule is "build passed is not verification." So:
1. Rewrite the trigger to enqueue with `provider='no_such_provider'`, insert an entry. Result: `work_queue.status='error'`, error message says exactly `unknown provider: no_such_provider`, brain row's `embedding_error` is stamped, `embedding IS NULL`. The error path actually carries information.
2. Restore the trigger, UPDATE the entry. Result: embed succeeds, `embedded_model` populated, `embedding_error` cleared. The success path explicitly sets `embedding_error = NULL` — important because if it didn't, a row that succeeded after a previous failure would still appear failed in app queries.

Both halves of the loop work. Now I trust the system more than I would have if I'd just done the happy path twice.

## What surprised me

**Vector ranking just worked.** I expected to spend a session debugging cosine distances or pgvector behavior. Insert "Charity is the pure love of Christ" + "Faith hope and charity" + a "QUERY" entry whose body is "pure love of Christ moroni", let LM Studio embed all three, then `brain_entries.embedding <=> query.embedding`:

```
 Charity is the pure love of Christ | 0.1948
 Faith hope and charity             | 0.3631
 QUERY (self)                       | 0.0000
```

The Charity entry beats the Faith/Hope entry by a wide margin against a "pure love of Christ" query. Self-similarity is exactly zero. This is the system working *because* the embedding model is well-trained on this kind of language and *because* gospel-engine-v2 already chose the same model+dimension for the same reason. Step 6 inherited a year of decisions without paying for them.

**LM Studio cold load is 2-3s, warm calls are sub-second.** First embed: 2.96s. Next four: 280–680ms each, **610ms average**. That's plenty for an interactive feel after the first warm-up. The 120s timeout was paranoid; in practice 5s would probably work for warm and 10s for cold. Leaving the long timeout because the cost of a too-short timeout is a false failure and the cost of a too-long timeout is one slow tick.

## Relational note: the harness question

Michael paused after step 3 yesterday and asked the right question at the right time. **"In brain we relied on github-copilot-sdk to do the agentic work in the file system, and it handled the MCP servers, the skills, tool calls, agent modes, and copilot-instructions. Here that's on us right?"**

The honest answer is: yes, mostly. The schema sketch in proposal.md does include `stewards.tool_calls`, `stewards.skills`, `stewards.instructions` — but they're Phase 3, and we don't build any of them in Phase 1. Step 6 (today) and step 7 (OpenCode Go chat) are both deterministic single-shot calls, no agent loop, no tools[] in the chat completions payload, no skill retrieval in the system prompt. They prove provider dispatch works, which is necessary but not sufficient for the agentic story.

I gave him three options: (1) stay on plan, eat the gap when it appears after step 7, (2) insert a "Phase 1.5: harness sketch" — toy `stewards.run_agent(session_id, user_message)` that proves prompt-assembly + tools[] round-trip without the full loop, (3) do step 6 only, then think before committing to anything chat-shaped. He chose to trust the plan and do step 6 first. Right call — step 6 is the smallest provider integration and it makes the harness question more concrete by giving us one real round-trip already shipped.

The carry-forward is: **before step 7, decide whether to detour into the harness sketch.** Step 7 is "send messages, get completion back" — useful for a CLI but not agentic. The harness question is "what tools[] does the agent see, where does the system prompt come from, who composes the messages[] from sessions+skills+instructions+retrieved memory, where do tool_call results land before the next iteration." None of that is built. None of it has to be built before step 7, but writing the toy version of it would be cheaper than building step 7 first and then realizing step 7 needs to be rewritten to plug into a harness.

My read: option 2. But Michael owns intent; that's a question for next session start.

## Carry forward

- **Decide step 7 vs Phase 1.5 harness sketch** at next session start. Strong lean toward harness first — see relational note above.
- **First-boot FATAL still in logs** ("database 'stewards' does not exist" then `set_restart_time(5s)`). Acceptable but ugly. When step 7 lands and the worker actually does HTTP, this single ugly line becomes more visible. Worth a clean retry path before that ships.
- **The 120s embed timeout is paranoid.** In practice warm calls are sub-second and cold is 3s. Could tighten to ~30s without risk; would surface real provider failures faster. Defer until we see a real timeout in the wild.
- **`brain_search_vec` works** but the helper signature accepts a `vector(768)` directly — callers have to compute the embedding themselves. Phase 1.5 (or step 7) should add `brain_search_text_with_vec(query_text, k)` that enqueues a synchronous embed and queries against the result. The current shape is fine for SQL clients; an MCP wrapper would want the convenience version.
- **PowerShell heredoc + `\gset` is broken.** Wasted ~3 minutes on it today. Workaround: write SQL to a real file. Recorded in repo memory.
- **No tests yet.** Still verifying via psql probes after each rebuild. Will become uncomfortable around step 5 (brain CLI driver) when there's actual Go code that can be tested.

Five sessions, a working embedding pipeline, and a real semantic search end-to-end. The schema from yesterday is no longer notional. Friday energy held into Saturday.
