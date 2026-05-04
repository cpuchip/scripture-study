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

**Phase 1.6 done (2026-05-03) — agent loop closes:**
- Schema: `messages.parent_work_id` chains iterations,
  `messages.reasoning_content` + `messages.reasoning_details`
  capture and replay thinking-model state (Moonshot returns 400
  without it).
- New `kind='tool_dispatch'` work item: reads parent assistant's
  `tool_calls`, executes each (sql_fn or http), inserts
  `role='tool'` messages with proper `tool_call_id` echo, enqueues
  continuation chat. Work-item-per-iteration architecture — every
  step durable, observable, cancellable, no starvation.
- Phase-3 of `chat`: when response has `tool_calls` AND iteration
  < `agent.steps`, enqueues `tool_dispatch` instead of stopping.
- Two seeded sql_fn tools: `brain_search_text_tool`, `load_skill_tool`.
- Bgworker resilience: stale-claim reaper at startup (zero window),
  `pg_proc` pre-flight before sql_fn dispatch (workaround for
  pgrx 0.18 quirk where PgTryBuilder doesn't catch ereports
  through `BackgroundWorker::transaction`).
- **Verified end-to-end** (`verify-loop.sql`):
  - Success: "name two virtues from Moroni 7" → kimi calls
    `brain_search_text` + `skill` → reads replies → answers →
    `finish_reason='stop'`. 18s, ~$0.0005.
  - Inverse (Agans Rule 9): broken tool → error JSON lands as
    `role='tool'` content → kimi recovers and finishes cleanly.
    **Bgworker did not crash.**
  - Reasoning replay verified: 266 + 2982 chars round-tripped.
- Spec gap named for Phase 1.6.1: if `tool_dispatch` *itself*
  errors (vs. tool returning an error string), no `role='tool'`
  reply gets written and the loop stalls. Acceptable for
  developer-driven use; addressed in [phases.md § Phase 1.6.1](../phases.md#phase-161--tool_dispatch-error-recovery-planned-not-started).

**Phase 1.6.1 done (2026-05-03) — tool_dispatch error recovery:**
- New SQL fn `synthesize_tool_failure(parent_work_id, agent_family,
  model, session_id, provider, error)` writes synthetic
  `role='tool'` replies (one per `tool_call_id` on the parent
  assistant message, idempotent via dedup) AND enqueues the
  continuation chat. Synthetic content is JSON with
  `_synthetic: true` marker so callers can distinguish.
- Dispatcher's `Err(msg)` arm wired through the helper for
  `kind='tool_dispatch'` rows. Loop never stalls on dispatcher
  failure — model receives the error and recovers.
- Stale-claim reaper enhanced: for every reaped `tool_dispatch`
  row, calls `synthesize_tool_failure` before marking errored.
  Bgworker crashes are now self-healing on next start.
- New `stewards.session_status` view: one row per session with
  `last_finish_reason`, `last_loop_stop_reason`, `pending_work`,
  `errored_work`, `total_tokens_in`, `total_billable_out`. Single
  SELECT answers "did this loop finish or stall?".
- Verified end-to-end via `verify-1-6-1-reaper.ps1`: insert
  orphaned `tool_dispatch` row → `docker compose restart pg` →
  reaper synthesizes failure reply → continuation chat enqueued →
  kimi reads the failure → retries with real `brain_search_text`
  call → finishes with `finish_reason='stop'`. Zero stalled rows.
- Pure-SQL unit tests in `verify-1-6-1.sql`: synthetic-reply
  insertion, idempotency, session_status output across multiple
  prior sessions.
- Per-tool retry policy deliberately deferred (YAGNI): the model
  already decides what to do with errors. Add a tool-level retry
  layer only if we observe the model looping on broken tools.

**Phase 2.1 done (2026-05-04) — studies + AGE citations:**
- New `stewards.studies` table mirroring brain_entries shape (slug PK,
  title, body, frontmatter jsonb, vector(768) embedding via the
  existing `embed` work_kind, FTS, version snapshots).
- AGE graph `stewards_graph` with `Study` and `Scripture`/`Talk`/
  `Manual` vertices and `CITES` edges. Bootstrap fn
  `stewards.ensure_studies_graph()` is idempotent and called both
  from init and defensively from import.
- `stewards.parse_gospel_links(body)` extracts `[text](.../gospel-library/eng/...)`
  links, normalizes to canonical URIs (`eng/scriptures/bofm/mosiah/18.md`,
  `eng/.../moro/7.md#47`, `eng/general-conference/2024/04/<slug>.md`).
- `stewards.import_study(slug, file_path, title, body, frontmatter)`
  upserts the row + syncs the graph (deletes existing CITES edges
  then re-creates from current body).
- `stewards.study_citations(slug)` reads the graph back to relational rows.
- PowerShell importer `import-studies.ps1` bulk-loads all 69 studies
  via `docker cp + psql -f` (avoids heredoc encoding pitfalls).
- Verified end-to-end: 69 studies, 432 unique scripture vertices,
  1256 CITES edges, all 69 embeddings populated. `verify-2-1.sql`
  runs seven inverse-hypothesis tests including the apostrophe/em-dash
  survival case that broke 13 of 69 imports on the first attempt.

**Critical AGE pattern (recorded everywhere AGE writes happen):**
Cypher does NOT honor PG's `''` single-quote escape inside string
literals. `format()`-built Cypher with `replace(x, '''', '''''')` is
a latent bug. Always use `cypher()`'s 3-argument form to bind via
`$param` placeholders: `cypher('graph', $$ ... $name ... $$, $1)`
where `$1` is `(jsonb_build_object(...)::text)::ag_catalog.agtype`.

**Phase 2.2 done (2026-05-04) — gospel-engine resolver:**
- `stewards.resolved_refs` cache table keyed by single-verse
  reference string ("Mosiah 18:8", "D&C 88:67"). Verse ranges in
  citation anchor_text fan out to one row per verse — a 5-citation
  range is one round-trip per verse but each verse is reusable.
- `stewards.parse_reference(text)` — parses anchor_text into the
  canonical reference shapes, normalizing en-dashes and chapter
  numbering. Returns empty for chapter-only refs ("D&C 76") and
  for non-scripture anchors ("Maxwell 1991") — those gracefully
  show empty `resolved_verses` arrays rather than enqueuing waste.
- `stewards.normalize_book(book)` — maps LDS-standard abbreviations
  ("Rom.", "3 Ne.", "Heb.", "Jas.", "Psalm") to the full forms
  gospel-engine v2 stores ("Romans", "3 Nephi", etc.). Also fixes
  the singular/plural slip "Psalm" → "Psalms".
- New `resolve_ref` work_kind. Bgworker hits
  `{GOSPEL_ENGINE_URL}/api/get?ref=<ref>` with the bearer token from
  `GOSPEL_ENGINE_TOKEN`. Both env vars read once in `_PG_init`.
- 404 from gospel-engine is a soft-error: the response body is
  cached in `resolved_refs.error` rather than retried. Errors are
  sticky — `enqueue_resolve` skips refs with ANY cached row. Use
  `stewards.invalidate_ref(ref)` to force a re-resolve after fixing
  the parser or backfilling gospel-engine.
- `stewards.refresh_study_refs(slug)` and
  `stewards.refresh_all_study_refs()` enqueue all unresolved refs
  for a study or the whole corpus. Both are idempotent.
- `stewards.study_citations_resolved(slug)` joins citations to
  cached verse text in one query. UI gets a jsonb array of
  `{ref, content, error}` objects per CITES edge, ready to render.
- Verified end-to-end on the full corpus: **1363 verses cached,
  87.2% success rate, 0 retries on cached rows.** The 12.8% misses
  are real corpus gaps in gospel-engine v2 (Hebrews, James,
  1 Corinthians, Jeremiah, Ezekiel, Job, Proverbs, several minor
  prophets, most general epistles) — confirmed by direct curl;
  those books simply aren't indexed there yet.

**Resolver finding to feed back to gospel-engine v2:** `/api/get?ref=`
returns 404 for many books that ARE present in the source markdown
under `gospel-library/eng/scriptures/nt/heb/*.md` etc. Either the
ingest skipped those books or the canonical-reference index missed
them. 175 verses across 21 studies are blocked on this. The list
of missing books and counts is in `stewards.resolved_refs WHERE
error IS NOT NULL`.

**Phase 2.3 done (2026-05-04) — similarity bridge:**
- `stewards.refresh_study_similarity(slug, top_k=5, min_score=0.5)`
  ports the probe's bridge pattern into production. For one source
  study, drops outgoing `:SIMILAR_TO {method:'pgvector_cosine'}`
  edges and writes top-K neighbors above min_score using pgvector's
  `<=>` (cosine distance) operator and the HNSW index from Phase 1.
- `stewards.refresh_all_study_similarity()` loops over every embedded
  study; corpus refresh (69 × 5 = 345 edges) runs in <1s.
- `stewards.study_similar(slug, limit=10)` reads edges back in BOTH
  directions and labels each result `outgoing` / `incoming` / `mutual`.
  The asymmetry is meaningful: mutual = both studies picked each other
  in their top-K; outgoing-only = A picked B but B didn't reciprocate.
- Score distribution across the 69-study corpus: min=0.62, p50=0.74,
  p95=0.81, max=0.94. Default 0.5 floor is permissive enough that
  every embedded study gets its top-K; UI / agents can tighten via
  the `min_score` parameter.
- Real clusters surface immediately. `art-of-delegation` →
  `stewardship-pattern` (0.888), `art-of-presidency`,
  `stewardship-pattern-reflections`. `charity` → `enoch-charity`
  (0.843), `tree-of-life-and-the-chain`, `miracles-references`.
  `give-away-all-my-sins` → `atoning-love-andersen`,
  `only-begotten`, `broken-heart-and-contrite-spirit`.
- Bug caught by inverse hypothesis: original `refresh_study_similarity`
  returned early when `embedding IS NULL` BEFORE deleting outgoing
  edges, so a freshly-nulled embedding kept stale edges. Fix: delete
  always, skip writes only when no embedding. Verify Test 6 now
  round-trips cleanly: baseline → null+refresh → 0 → restore+refresh
  → original neighbors. Without that ordering the test reveals stale
  cache that survives “refresh.”

**Phase 2.4 done (2026-05-04) — `stewards study show` view:**
- `stewards.study_show(slug, sim_limit, cite_limit, verse_chars)`
  pulls together everything Phase 2 built into one formatted text
  blob: study row + frontmatter + embedded_at, resolved citations
  with verse text inlined, similar studies ranked by score with
  outgoing/incoming/mutual labels, and a footer count.
- Thin PowerShell wrapper [stewards.ps1](stewards.ps1): `study show`,
  `study list`, `study refresh [slug]`. Forces UTF-8 console encoding
  so em-dashes survive psql's text output (Windows defaults to cp1252
  and renders — as mojibake).
- **Phase 2 done criteria met:** running
  `.\stewards.ps1 study show give-away-all-my-sins` returns the
  study, 14 resolved scripture verses across 6 citations, and 5
  similar studies (`atoning-love-andersen`, `only-begotten`,
  `moses-6-gospel-to-adam`, `know-god`, `receive`).
- Talk citations and chapter-only refs gracefully render as
  "_(no resolvable verses for this anchor)_" — Phase 2.2's
  deferred-by-design path working end-to-end.

**Generalization question raised by 2.4 output (→ Phase 2.5 candidate):**
The importer is hardcoded to `study/`, but the schema is generic.
`docs/work-with-ai/*-gospel.md` files ("The Creation Pattern: Working
with AI as God Works with Intelligence", "Watching Until They Obey:
The Feedback Loop as Divine Pattern") belong in the same graph as the
scripture studies. Recommended Phase 2.5: parameterize the importer
over a `-Sources` list, add a `corpus text` column to
`stewards.studies` (study/doc/lesson/journal), keep the `:Study` AGE
label as the slight misnomer it becomes — cost of corpus-wide rename
outweighs cosmetic gain.

**Phase 1 deliverables 4 + 5 (brain migrator + brain CLI port) deferred** —
substrate work (1.5/1.6) turned out to matter more than the brain
port, which becomes a "do it when SQLite hurts" item rather than
a Phase 2 blocker. See [phases.md](../phases.md#phase-1--foundation-extension-scaffold--bgworker--brain-port).

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
