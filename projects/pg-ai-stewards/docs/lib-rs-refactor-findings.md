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

## Risks identified before starting

1. **`pgrx` macro visibility.** `#[pg_extern]` and `extension_sql_file!` may need to be at crate root to register correctly with pgrx's build system. Splitting them across modules could break extension generation.
2. **`extension_sql_file!` `requires=` chain.** Macros declare named dependencies between SQL blocks. If we move them to a submodule, the macro names still need to be unique across the crate. Probably fine since they're already namespaced.
3. **`#[pg_guard]` on `_PG_init`.** Bgworker entry point. May need to stay in `lib.rs` or be `pub use`'d from `bgworker.rs`. Will verify.
4. **Cross-module visibility.** Functions currently default-private (no `pub`). Moving them across modules requires `pub` annotations or `pub(crate)` for crate-internal sharing.
5. **`reqwest` and tokio runtime.** The bgworker uses a blocking reqwest client. Moving the HTTP dispatch to a separate module shouldn't change runtime behavior, but type signatures may need adjusting.
6. **Compile-time impact.** The whole point. Want to measure before/after.

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

*(filled in as moves land)*
