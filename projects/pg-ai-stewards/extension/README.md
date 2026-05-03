# pg_ai_stewards extension — dev stack

The actual Postgres extension. Phase 1 of [the project](../).

## Status

**Phase 1, steps 1+2+3+6 done (2026-05-02 / 2026-05-03):**
- pgrx 0.18 extension builds, loads on PG18 alongside pgvector + AGE.
- Bgworker registered via `shared_preload_libraries`, three-phase
  dispatch (claim/HTTP/write) so cold model loads don't hold row locks.
- Provider registry parsed from `STEWARDS_PROVIDER_*` env in
  `_PG_init`; visible without secrets via `stewards.providers_loaded()`.
- Brain schema: `brain_entries` + companions, `vector(768)` + HNSW
  cosine, generated tsvector + GIN, version-snapshot trigger gated
  on content change, embed-enqueue trigger.
- **Real LM Studio embeddings landing in the vector column.**
  Average **610ms warm**, ~3s cold. `brain_search_vec` ranks
  correctly against real query vectors.

**Phase 1.5 done (2026-05-03) — harness sketch (detour before step 7):**
- `stewards.agents`, `stewards.skills`, `stewards.instructions`,
  `stewards.tool_defs`, `stewards.agent_tool_perms`,
  `stewards.agent_skill_perms`, `stewards.tool_calls` schema.
- **Variant-by-glob:** agents/skills/instructions can ship multiple
  rows per logical family, differentiated by `model_match` glob
  (`'kimi-*'`, with `'*'` as the catch-all default). Resolver picks
  the longest matching pattern. Same workflow rules, model-tuned
  personas.
- `glob_match(pattern, value)` — sanitized `LIKE` translation,
  reused by `tool_permission`/`skill_permission` (3-state
  `allow`/`ask`/`deny`, last-matching wins, default-allow).
- `compose_system_prompt` / `compose_messages` / `compose_tools` —
  pure read-only assembly, all `STABLE`.
- `dry_run_chat(family, model, session, input)` returns the exact
  JSON body that would POST to `/v1/chat/completions`. The
  verification target: read the bytes, judge the shape, then build
  step 7 against a frozen contract.
- Skill advertising follows the OpenCode pattern (`<available_skills>`
  XML block inside the `skill` tool description, NOT in the system
  prompt body) — token-efficient, agent loads on demand.
- Verified: kimi system prompt is exactly 86 chars longer than
  gpt-5 system prompt for the same agent family, because the
  `kimi-*` agent variant adds a "be terse" clause and nothing else
  varies. Inverse hypothesis: unknown agent family raises cleanly.

Everything else from the [phase plan](../phases.md#phase-1--foundation-extension-scaffold--bgworker--brain-port)
(brain CLI driver in step 5, OpenCode Go chat in step 7, Go
migrator in step 4) is still ahead. Step 7 is now smaller because
the composition shape is frozen.

**Phase 1, step 7 done (2026-05-03) — chat round-trip via OpenCode Go:**
- `stewards.chat_enqueue(agent_family, model, session_id, user_input,
  provider)` composes body via `dry_run_chat`, persists user turn,
  enqueues `kind='chat'`. Returns work_queue id.
- Bgworker `dispatch()` `chat` arm POSTs to `<base>/chat/completions`,
  parses OpenAI shape (`choices[0].message`, `usage`, `model`),
  phase 3 inserts assistant message into `stewards.messages` with
  `tool_calls` jsonb (verbatim, for Phase 1.6), `finish_reason`,
  `tokens_in/out`.
- Verified: **4.4s round-trip** to kimi-k2.6 via OpenCode Go
  (`https://opencode.ai/zen/go/v1`). Kimi accurately restated the
  persona we composed in Phase 1.5 — proving the harness shape
  arrives intact at the model.
- Provider echo persisted (asked `kimi-k2.6`, got
  `moonshotai/kimi-k2.6-20260420`). We record what the provider
  actually used.
- Inverse hypothesis: bad provider → `unknown provider:
  does_not_exist` in `work_queue.error`, no row leaks.
- **Stewardship action surfaced:** a draft `chat_round_trip()` SQL
  fn was caught on first run — it polled inside its own tx, hiding
  its own enqueued row from the bgworker (MVCC). Removed with an
  inline `-- NOTE:` comment for future-me. SQL functions cannot
  COMMIT mid-loop; real callers should `LISTEN stewards_done`.
- Tool dispatch + agent loop NOT here — that's Phase 1.6.
  `assistant.tool_calls` is persisted but unread.

## Layout

```
extension/
├── Cargo.toml                  # pgrx 0.18.0, default-features = ["pg18"]
├── pg_ai_stewards.control      # PG control file, schema = stewards
├── src/
│   └── lib.rs                  # one-function scaffold (version, pgrx_version)
├── Dockerfile                  # multi-stage: rust builder + runtime w/ pgvector+AGE
├── docker-compose.yaml         # dev stack on host port 55433
└── init/
    └── 00-extensions.sql       # CREATE EXTENSION x3 on first boot
```

## Build & run

```pwsh
cd projects\pg-ai-stewards\extension
copy .env.example .env       # then fill in OPENCODE_GO_API_KEY (others have defaults)
docker compose build         # ~2 min cold; ~30s warm thanks to layer cache
docker compose up -d
```

`.env` is optional — the compose file falls back to inline defaults
if it's missing, so `docker compose up -d` works without it. Real
provider keys (OpenCode Go etc.) only matter once Phase 1 step 6/7
wires the bgworker; for now `.env` is just the committed shape.
See [proposal § Provider abstraction and secrets](../proposal.md#provider-abstraction-and-secrets)
for the full design.

### Secrets — what stays local

**`.env` never enters the Docker image.** The [Dockerfile](Dockerfile)
only copies `Cargo.toml`, `pg_ai_stewards.control`, and `src/` into
the builder stage. There is no `COPY .env` and no `COPY . .`.
[`.dockerignore`](.dockerignore) is belt-and-suspenders: even if the
Dockerfile is later refactored to `COPY . .`, `.env` and `.env.*`
are excluded from the build context (only `.env.example` passes through).

`docker compose` reads `.env` at *runtime* via `env_file:` and sets
the values as environment variables on the running **container**.
Those values are:

- visible to processes inside the container (the bgworker reads them
  on startup)
- visible via `docker inspect <running-container>` on your local machine
- **NOT** in the image filesystem
- **NOT** in any layer (`docker history` is clean)
- **NOT** included if you `docker push` the image or `docker save` it

You can verify this for yourself:

```pwsh
# Layer history — should print nothing
docker history pg-ai-stewards-dev:pg18 --no-trunc --format "{{.CreatedBy}}" `
  | Select-String -Pattern 'STEWARDS_PROVIDER|API_KEY' -SimpleMatch

# Image-level Env — should only show stock Postgres vars (PG_MAJOR, LANG, etc.)
docker image inspect pg-ai-stewards-dev:pg18 --format "{{json .Config.Env}}"

# Filesystem grep — should print nothing
docker run --rm --entrypoint sh pg-ai-stewards-dev:pg18 `
  -c "grep -rI 'STEWARDS_PROVIDER_OPENCODE' / 2>/dev/null | head -5"
```

**For a future standalone public repo** (when this project graduates
out of `scripture-study`), the same model works: ship `.env.example`
and `.dockerignore`, never ship `.env`. For shared dev environments
or production, swap `.env` for [Docker secrets](https://docs.docker.com/engine/swarm/secrets/)
or a real secret manager (Vault, 1Password, AWS Secrets Manager) —
the bgworker reads env vars regardless of how they got there, so the
bootstrap surface doesn't change.

Then verify:

```pwsh
docker exec -it pg-ai-stewards-dev psql -U stewards -d stewards `
  -c "SELECT extname, extversion FROM pg_extension WHERE extname IN ('vector','age','pg_ai_stewards') ORDER BY extname;" `
  -c "SELECT stewards.version();"
```

Expected:

```
    extname     | extversion
----------------+------------
 age            | 1.7.0
 pg_ai_stewards | 0.1.0
 vector         | 0.8.2

 version
---------
 0.1.0
```

The dev stack runs on **port 55433** so it doesn't collide with the
probe stack on 55432. Both can run simultaneously.

## Tear down

```pwsh
docker compose down -v      # -v drops the volume so init runs again
```

## Dev loop

This is a deliberately **slow** dev loop for now: every code change
requires a full image rebuild. That's fine for the scaffold step
because changes are infrequent. When iteration starts to bite, swap in
a mounted-source dev container (Rust + cargo-pgrx with the source
directory bind-mounted) that builds in place and re-installs into the
running Postgres without rebuilding the image. Track that as a Phase 1
quality-of-life upgrade.

## Notes for next session

- pgrx `pg_module_magic!` in 0.18 wants `CStr` arguments if you pass
  named ones; the no-arg form is simpler and pulls metadata from
  `Cargo.toml`. Already applied here.
- `cargo pgrx package --out-dir /out` produces a tree rooted at `/`
  (e.g. `/out/usr/lib/postgresql/18/lib/pg_ai_stewards.so`), NOT a
  named subdirectory. The `COPY --from=builder /out/ /` line in the
  Dockerfile depends on this.
- The runtime image is `pgvector/pgvector:pg18` + Apache AGE built
  from source, exactly matching the [probe](../probe/Dockerfile).
  When AGE or pgvector versions change, change them in both places.

## Next steps (per phases.md Phase 1)

1. **bgworker scaffold** — `cargo pgrx new --bgworker` template, then
   register a worker that listens on `LISTEN stewards_dispatch` and
   reads from `stewards.work_queue`. Reference: [pg_vectorize](https://github.com/ChuckHend/pg_vectorize).
2. **Schema for brain replacement** — `stewards.brain_entries`,
   `stewards.messages`, HNSW index, JSONB props.
3. **Migrator** — Go binary reading `scripts/brain/`'s SQLite +
   chromem-go vector store, writing into Postgres.
4. **Brain CLI driver** — Postgres backend behind the existing brain
   API surface; SQLite stays as read-only fallback for ~30 days.
5. **Real provider call through bgworker** — Ollama embedding for new
   brain entries, end-to-end.
