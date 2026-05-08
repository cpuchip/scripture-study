# lib.rs refactor findings — Phase 3c.3.6

*Started 2026-05-08, soak paused for the duration.*

This is a working journal of the lib.rs module split. Updated incrementally as moves land, including surprises, gotchas, and "what would I tell the next person who tries this." Entries kept terse — pgrx + extension_sql_file! has subtle constraints and we want them captured while fresh.

## Pre-state baseline

```
$ wc -l extension/src/lib.rs
4246 lib.rs

$ ls extension/src/
lib.rs
```

Single 4246-line file, no module structure. Mixes:
- `pg_module_magic!()` declaration
- ~20+ `extension_sql!` blocks (DDL embedded as Rust raw strings)
- 16 `extension_sql_file!` macros (foldback chain pointing at SQL files)
- ~30 `#[pg_extern]` functions exposed to SQL
- Provider registry + parsing from env (`Provider`, `ProviderRegistry`)
- `_PG_init` postmaster entry point + bgworker registration
- Bgworker tick loop + `process_one_pending` + dispatch helpers
- Provider HTTP dispatch (`embed`, `chat` via reqwest)
- Tool dispatch (`tool_dispatch`, `exec_one_tool`, `exec_sql_fn_tool`, `exec_http_tool`)
- Reference resolution helpers (`resolve_ref`, `url_encode_query_value`)

## Section map (line ranges, approximate)

| Lines | Section | Target module |
|-------|---------|---------------|
| 1-23 | Crate docs, imports, `pg_module_magic!()` | `lib.rs` |
| 26-422 | work_queue + brain schema DDL | `schema.rs` |
| 424-464 | Tool wrappers (brain_search_text_tool, load_skill_tool) | `schema.rs` |
| 466-1306 | Phase 1.5 harness (agents, skills, instructions, tool_defs, chat helpers) | `schema.rs` |
| 1308-1695 | Phase 2.1 studies + AGE citations | `schema.rs` |
| 1696-2287 | Watchman, dirty_queue, more schema | `schema.rs` |
| 2287-2438 | More schema/functions | `schema.rs` |
| 2440-2466 | Foldback comment header | `lib.rs` |
| 2468-2562 | 16 `extension_sql_file!` macros | `lib.rs` |
| 2564-2610 | `#[pg_extern]` wrappers (version, enqueue, providers_loaded) | `pg_externs.rs` (or stay in `lib.rs`) |
| 2620-2737 | Provider types (`Provider`, `ProviderRegistry`, `GospelEngineConfig`) | `providers.rs` |
| 2739-2944 | `_PG_init` + bgworker registration | `bgworker.rs` |
| 2950-3517 | `check_watchman_schedule`, `process_one_pending` | `bgworker.rs` |
| 3518-3815 | `dispatch`, `embed`, `chat` (HTTP via reqwest) | `bgworker.rs` (or split to `chat.rs`) |
| 3816-3910 | `resolve_ref` + URL helpers | `tools.rs` |
| 3911-4218 | `tool_dispatch`, `exec_one_tool`, `exec_sql_fn_tool`, `exec_http_tool` | `tools.rs` |
| 4220-4246 | (final tail — likely diagnostic helpers) | TBD |

## Risks identified before starting (and how research resolved them)

**Update 2026-05-08 after research:** all four pgrx-macro-visibility concerns turned out to be unfounded. Detailed answers in the new `.github/skills/pgrx-rust/SKILL.md` skill. Summary:

1. **`pgrx` macro visibility — RESOLVED.** `#[pg_extern]`, `extension_sql!`, `extension_sql_file!`, `#[pg_guard]` all work in submodules. pgrx's SQL emitter walks the entire crate via `pgrx_embed.rs` and harvests metadata symbols regardless of source-file location. **Reference:** pg_vectorize puts `_PG_init` two levels deep at `extension/src/workers/pg_bgw.rs` with no `pub use` re-export.
2. **`extension_sql_file!` `requires=` chain — RESOLVED.** Names resolve via dependency graph, not lexical position. **One real gotcha discovered:** `extension_sql_file!("../sql/foo.sql")` paths are FILE-relative, not crate-relative. Adjust paths if blocks move to a nested directory. Our moves stay at `src/` depth so this doesn't bite us, but worth remembering.
3. **`#[pg_guard]` on `_PG_init` — RESOLVED.** Works in any submodule. Postgres finds it at `dlopen` time via C linkage; module path doesn't matter. Plain `mod bgworker;` is enough — no `pub use` needed.
4. **Cross-module visibility — minor.** `pub(crate)` on items used across modules. Default visibility hides items from siblings. Compiler errors are clear and easy to fix.
5. **`reqwest` and tokio runtime — non-issue.** Just plain Rust function calls; module split doesn't affect runtime behavior.
6. **Compile-time impact** still TBD. First move (providers.rs) didn't show measurable improvement because Cargo recompiles the whole crate on any source change. The win is *navigation* and future-edit-blast-radius, not compile time.

## Strategy

Conservative, incremental moves. After each module split:
1. `cargo check` to verify compile
2. `cargo build` (or `cargo pgrx package`) to verify the full pipeline
3. Smoke test if anything functional changed (most moves should be code-shape-only)
4. Document findings here before the next move

Order of moves (least risky → most risky):
1. **`providers.rs`** — pure data types + parsing. No pgrx macros. Lowest risk.
2. **`tools.rs`** — `tool_dispatch` and helpers. Pure Rust dispatcher functions called from bgworker. Should move cleanly.
3. **`bgworker.rs`** — `_PG_init`, tick loop, HTTP dispatch. Bigger but cohesive. The `#[pg_guard]` attributes need careful handling.
4. **`schema.rs`** — all the `extension_sql!` blocks. Largest chunk. Risk: pgrx might require all `extension_sql!` calls to be visible at the crate level. Will test with a tiny move first.
5. **`pg_externs.rs`** (maybe) — pgrx-exposed functions. Risk: same as schema.rs — may need crate-level visibility.

## Findings log

### Move 1 — `providers.rs` (2026-05-08)

**Scope:** Lines 2620-2737 of original lib.rs. ~120 lines extracted to `extension/src/providers.rs`. Items moved: `Provider`, `ProviderSummary`, `ProviderRegistry` + `impl`, `split_provider_key`, `PROVIDER_REGISTRY` static, `GospelEngineConfig`, `GOSPEL_ENGINE_CONFIG` static.

**Mechanics:**
- Created `extension/src/providers.rs` with the moved code, wholesale-copied
- Marked all moved items `pub(crate)` so the bgworker + dispatch code in lib.rs can reach them
- Added `mod providers;` + `use providers::{...};` to lib.rs (after the `use` block, before `pg_module_magic!()`)
- Removed the original block from lib.rs, leaving a 4-line breadcrumb comment
- Removed the now-unused `use std::sync::OnceLock;` from lib.rs

**Result:**
- lib.rs: 4246 → 4138 lines (-108)
- providers.rs: 127 lines (new)
- Total: 4265 lines (+19 from comments/headers)
- `cargo pgrx package` clean. 30 SQL entities discovered (was 30, no change). Image built end-to-end.

**Surprises / non-surprises:**
- **No pgrx macros in this section, so no visibility issues.** The `OnceLock` statics needed `pub(crate)` because they're accessed from bgworker code in lib.rs, but otherwise the move was mechanical.
- **`pgrx::log!` inside `from_env`** continued to work without any feature-gating — pgrx exports it from the prelude that was used by ProviderRegistry's parent module, but since I imported `pgrx::log!` is a macro and macros need a separate `use` path... actually it just worked because it's `pgrx::log!()` fully-qualified in the source. No fix needed.
- **`pub(crate)` on every field** (not just types) was necessary because the bgworker code in lib.rs reads field values directly (e.g., `provider.api_key`, `provider.base_url`). Could have used getters/methods instead but mechanical-extraction-first is the right discipline.

**What would I tell the next person:** Start with the leaf module that has no pgrx macros and no `extension_sql!` calls. The compiler error messages around pub(crate) visibility are immediate and clear — fix one, build, fix the next. Don't preemptively over-engineer with getters/setters; just expose the fields as `pub(crate)`.

**Build time observation:** First build after the split took 36s for the cargo step (Stage 1 of Dockerfile). Need a baseline run to compare — currently no before-numbers. Noted as a TODO for the next module move.

**Smoke test on fresh DB:** ephemeral container + `CREATE EXTENSION CASCADE` clean — 4 agents, 7 tool_defs, 2 pipelines (matches pre-split baseline). Live container restarted onto the new image: 6 substrate studies preserved, 193 dirty queue, soak re-enabled, bgworker poll loop active with 4 providers.

### Decision after Move 1: stop here, document the rest as future work

After completing providers.rs cleanly, weighed pushing forward to `tools.rs` (the planned next move). Re-reading the section map showed `tool_dispatch` returns `Result<WorkOutcome, String>`, and `WorkOutcome` is an enum still living in lib.rs (line 3347). Cleanly extracting tools.rs would require either:

1. Also extracting `WorkOutcome` to a shared `types.rs` module (chains the move into a third file), or
2. Splitting only the leaf execute helpers (`exec_sql_fn_tool`, `exec_http_tool`) and leaving the orchestrator (`tool_dispatch`) in lib.rs (cuts the savings in half — only ~150 lines moved instead of ~400)

Neither felt like a clean ship. The diminishing return on a second move during a single autonomous session — versus the proven pattern from move 1 — pushed me to commit the providers.rs split as a standalone phase 3c.3.6 v1, document the rest as future work, and stop.

The rhythm "Things in order and wisdom, not faster than we have strength" applies. One clean move with full validation > two half-finished moves with build pressure.

## Future moves (planned, not done)

In order of risk/effort:

### Move 2 — `tools.rs` (planned)

**Scope:** Lines 3722-4109 of current lib.rs (~390 lines). `resolve_ref`, `url_encode_query_value`, `tool_dispatch`, `exec_one_tool`, `exec_sql_fn_tool`, `exec_http_tool`.

**Prerequisite:** Decide on `WorkOutcome`'s home. Two options:
- a) Create `types.rs` first, move `WorkOutcome` + any other shared types there, then move tools.rs cleanly
- b) Move only the leaf helpers (`exec_sql_fn_tool`, `exec_http_tool`) and leave `tool_dispatch` + the WorkOutcome dependency in lib.rs

Recommendation: **option (a)**. `WorkOutcome` is referenced by `dispatch`, `embed`, `chat`, `resolve_ref`, `tool_dispatch` — at least 5 callers in lib.rs. Once a future bgworker.rs split lands, all of those move out of lib.rs anyway. Centralizing `WorkOutcome` early in `types.rs` avoids re-doing visibility work twice.

### Move 3 — `bgworker.rs` (planned)

**Scope:** Lines ~2739-3815 of current lib.rs (~1075 lines). `_PG_init` + bgworker registration, `check_watchman_schedule`, `process_one_pending`, `dispatch`, `embed`, `chat`.

**Risks:**
- `#[pg_guard]` on `_PG_init` — must verify pgrx accepts this attribute on a function in a non-root module. The pgrx examples I've seen all keep `_PG_init` in lib.rs. May require `pub use` re-export from lib.rs, or `_PG_init` may need to stay in lib.rs as a thin shim that calls into bgworker.rs.
- The `#[pg_guard] pub extern "C-unwind" fn _PG_init()` signature is generated by pgrx attribute macros that may not be portable across modules. Test with a minimal stub before doing the full move.

### Move 4 — `schema.rs` (planned, biggest)

**Scope:** Lines ~26-2438 of current lib.rs (~2400 lines). All `extension_sql!` blocks for the in-Rust-source schema definitions. Does NOT include the `extension_sql_file!` macros (those stay in lib.rs).

**Risks:**
- pgrx might require all `extension_sql!` calls to be visible at the crate root for the SQL emitter to find them. Need to test with a tiny `extension_sql!` move first before bulk-moving 2400 lines.
- The `requires=` chain spans all the moved blocks plus the file-based ones. Names need to remain unique across the split files. Should be OK since names are global.
- This is the bulk of the file's complexity. Defer until moves 2-3 establish confidence in the pattern.

### Move 5 — Lower priority

- `pg_externs.rs` — the few `#[pg_extern]`-marked Rust wrappers (version, enqueue, providers_loaded). Probably fine to keep in lib.rs since they're small and the pgrx attribute may need crate-root visibility.

## Final state after move 1

```
extension/src/
├── lib.rs        4138 lines  (was 4246, -108)
└── providers.rs   127 lines  (new)
```

Total LOC: 4265 (+19 from comments/headers). Image rebuilds cleanly. Live container restarted, soak re-enabled. Phase 3c.3.6 ships as v1; v2-v4 are documented above.

### Move 2a — `types.rs` (2026-05-08)

**Scope:** `WorkOutcome` enum (~58 lines). Pure type, prerequisite for tools.rs and bgworker.rs splits.

**Result:**
- lib.rs: 4137 → 4081 lines (-56)
- types.rs: 64 lines (new)
- Build clean

**Surprises:** none. Single enum, marked `pub(crate)`, plain `mod types;` + `use types::WorkOutcome;` in lib.rs. Same pattern as providers.rs.

### Move 2b — `tools.rs` (2026-05-08)

**Scope:** `resolve_ref` + `tool_dispatch` + `exec_*` helpers (~390 lines).

**Result:**
- lib.rs: 4081 → 3685 lines (-396)
- tools.rs: 418 lines (new, includes 23-line header)
- Build clean

**Mechanics:** Used `sed -i '3658,4045d' lib.rs` for the line-range delete since the block was too large for a clean Edit-tool match. Added `mod tools;` + `use tools::{resolve_ref, tool_dispatch};` to lib.rs.

**Surprises:** none. The functions touched `pgrx::Spi`, `BackgroundWorker::transaction`, `reqwest`, `pgrx::PgTryBuilder` — all available via `use pgrx::prelude::*;` + `use pgrx::bgworkers::*;` in tools.rs.

### Move 3 — `bgworker.rs` (2026-05-08)

**Scope:** `_PG_init` + bgworker tick loop + `dispatch`/`embed`/`chat` (~1018 lines). The biggest single move.

**Result:**
- lib.rs: 3685 → 2658 lines (-1027)
- bgworker.rs: 1041 lines (new)
- Build clean

**Mechanics:**
- Used `sed -n '2634,3651p' lib.rs > /tmp/bgworker-body.rs` to extract
- `cat header > bgworker.rs && cat body >> bgworker.rs` to assemble
- `sed -i '2634,3651d' lib.rs` to remove
- Added `mod bgworker;` to lib.rs (no `use` needed — nothing in lib.rs references bgworker functions; Postgres calls `_PG_init` via C linkage)
- Cleaned up `use pgrx::bgworkers::*;` and `use std::time::{Duration, Instant};` from lib.rs (only bgworker uses them now)

**Surprises:** none. The pgrx-rust skill research was 100% correct: `_PG_init` works in any submodule. No `pub use` needed. No `#[pg_guard]` placement issue. Plain `mod bgworker;` is enough.

### Move 4 — `schema.rs` (2026-05-08)

**Scope:** All `extension_sql!` blocks (~2412 lines). The largest extraction.

**Result:**
- lib.rs: 2658 → 246 lines (-2412)
- schema.rs: 2432 lines (new)
- Build clean
- Smoke test on fresh DB: 4 agents, 7 tool_defs, 2 pipelines (matches pre-split baseline)
- 30+ SQL entities discovered correctly

**Mechanics:** Same `sed` extract-pattern as move 3. Added `mod schema;` to lib.rs declarations.

**Surprises:** none. The `extension_sql!` blocks emitted correctly from the submodule. Per the pgrx-rust skill, pgrx walks the entire crate via `pgrx_embed.rs` for SQL discovery — module location doesn't matter. The `extension_sql_file!` macros stayed in lib.rs because their relative paths reference the file location of the macro call.

## Final state after all moves

```
extension/src/
├── lib.rs         246 lines  (was 4246, -94%)
├── schema.rs     2432 lines  (extension_sql! DDL blocks)
├── bgworker.rs   1041 lines  (_PG_init, tick loop, dispatch/embed/chat)
├── tools.rs       418 lines  (resolve_ref, tool_dispatch, exec_* helpers)
├── providers.rs   127 lines  (Provider, ProviderRegistry, Gospel config)
└── types.rs        64 lines  (WorkOutcome enum)
```

**Total: 4328 lines (+82 from headers/breadcrumbs).**

lib.rs now contains:
- Module declarations (`mod bgworker; mod providers; mod schema; mod tools; mod types;`)
- `use pgrx::prelude::*;` and `use providers::{...};`
- `pgrx::pg_module_magic!()`
- 16 `extension_sql_file!` macros (foldback chain — kept in lib.rs because relative paths point at `../foo.sql`)
- 4 `#[pg_extern]` SQL-exposed wrappers: `version`, `pgrx_version`, `enqueue`, `providers_loaded`
- The `#[cfg(test)] mod tests` and `pub mod pg_test` blocks (test framework requires crate-root visibility)

## Verification

- **Build:** all 5 commits' images built cleanly with `cargo pgrx package` reporting "Discovered 30 SQL entities" (no change from pre-split baseline)
- **Smoke test on fresh DB after schema.rs split:** `CREATE EXTENSION CASCADE` clean, all expected substrate state present (4 agents, 7 tool_defs, 2 pipelines, work_item_promote_trg trigger, agent_tool_perms_source_check constraint, all major tables)
- **Live container restart on the final image:** state preserved (6 substrate studies, 188 dirty queue, 19 broadcast / 275 frontmatter perm provenance), soak schedule_enabled re-flipped on, bgworker poll loop active with 4 providers

## What this enables

1. **Easier next-feature work.** When 3e (MCP) lands, the new module(s) will declare `mod mcp;` + `mod http_dispatch;` etc. without touching lib.rs's bulk.
2. **Clearer ownership.** A bug in chat dispatch lives in bgworker.rs. A bug in DDL lives in schema.rs. Grep ranges are bounded.
3. **Faster compile times for partial edits** (in theory — Cargo's incremental compilation works at file granularity for some cases, though `cargo pgrx package` does a full rebuild). Not measured.
4. **Lower context-window cost** when reading parts of the codebase via the Read tool — no more 4000-line files.

## Key skill: `pgrx-rust`

Authored as `.github/skills/pgrx-rust/SKILL.md` based on research into pg_vectorize, ParadeDB pg_search, and pgrx examples. Captures the four critical questions (`_PG_init` placement, `extension_sql!` placement, `#[pg_extern]` placement, `pub use` re-exports) so future sessions don't have to re-derive them.

## Time and effort

- Research phase (delegated to general-purpose subagent): ~3 minutes
- Skill authoring: ~10 minutes
- Move 2a (types.rs): ~5 minutes including build
- Move 2b (tools.rs): ~10 minutes including build
- Move 3 (bgworker.rs): ~15 minutes including build (biggest move)
- Move 4 (schema.rs): ~10 minutes including build
- Final restart + verification: ~5 minutes

**Total: ~1 hour** for what was estimated at 3-5 hours. The research front-loaded all the uncertainty; the moves themselves were mechanical once the pgrx-rust skill answered the visibility questions.

The `sed` line-range pattern (`sed -n 'A,Bp' lib.rs > /tmp/body.rs && sed -i 'A,Bd' lib.rs`) was the right tool for >100-line moves where Edit's exact-string-match would be cumbersome.

