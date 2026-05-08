---
name: pgrx-rust
description: "Authoritative reference for pgrx (Postgres extension in Rust) project structure, attribute placement, and module organization. Load when working in `projects/pg-ai-stewards/extension/` or any pgrx Rust extension. Captures patterns harvested from real production extensions (pg_vectorize, ParadeDB pg_search) so you don't have to guess what `#[pg_guard]` or `extension_sql!` need at the crate level."
user-invokable: false
---

# pgrx-rust — Module organization + attribute placement

This skill answers the questions that block multi-file pgrx refactors. Every claim here was verified against real production extensions in 2026-05.

## When to load

- Editing or refactoring code in `projects/pg-ai-stewards/extension/src/`
- Adding a new `#[pg_extern]` function and unsure where it can live
- Splitting a monolithic `lib.rs` into modules
- Hitting a build error that mentions `_PG_init`, `pg_module_magic`, or `pgrx_embed`
- Designing the file structure for a brand-new pgrx extension

## The four questions everyone asks (with answers)

### 1. Does `_PG_init` need to be at the crate root?

**No.** `_PG_init` can live in any submodule. Postgres finds it at `dlopen` time via C linkage — `#[pg_guard]` applied to a `pub extern "C-unwind" fn _PG_init()` makes the symbol externally visible regardless of Rust module path. No `pub use` re-export needed in `lib.rs`. Just declare the module: `mod bgworker;` (or `pub mod`, doesn't matter).

**Reference:** [pg_vectorize puts `_PG_init` two levels deep](https://github.com/ChuckHend/pg_vectorize/blob/main/extension/src/workers/pg_bgw.rs) at `extension/src/workers/pg_bgw.rs`. Their `lib.rs` is ~30 lines and contains only module declarations + `pg_module_magic!()`.

**Practical rule:** Keep `_PG_init` next to the `BackgroundWorkerBuilder::new(...).load()` calls it owns. That's where it's most readable.

### 2. Do `extension_sql!` and `extension_sql_file!` work in submodules?

**Yes, both work anywhere.** pgrx's SQL emitter builds the binary with a `pgrx_embed.rs` driver, runs it, and harvests metadata symbols across the whole crate. Module location doesn't affect discovery. `name=` and `requires=[…]` dependency declarations resolve via the dependency graph, not lexical position.

**One critical gotcha:** `extension_sql_file!("../sql/foo.sql")` paths are **relative to the file containing the macro**, not the crate root. If you move a block from `lib.rs` into `src/schema.rs` (one level deeper), a path like `"../foo.sql"` may need to become `"../foo.sql"` still (same depth) or `"../../foo.sql"` if you nest into `src/schema/mod.rs`. Always re-check the path after a move.

**Reference:** [pgrx custom_sql example](https://github.com/pgcentralfoundation/pgrx/blob/develop/pgrx-examples/custom_sql/src/lib.rs) lines 21-36 explicitly documents that `creates`, `requires`, `name` work via the dependency graph (graph-based, not lexical-position based).

### 3. Does `#[pg_extern]` need to be at the crate root?

**No, also any submodule.** From the [pgrx schemas example](https://github.com/pgcentralfoundation/pgrx/blob/develop/pgrx-examples/schemas/src/lib.rs):

> *"All top-level pgrx objects, **regardless** of the .rs file they're defined in, are created in the schema determined by `CREATE EXTENSION`."*

If you wrap a module in `#[pg_schema] mod foo { ... }`, the contained `#[pg_extern]`s land in Postgres schema `foo`. A plain `mod foo;` puts them in the extension's default schema (the one declared in `.control`).

**Practical rule:** Put `#[pg_extern]` definitions next to the Rust functions they wrap. Don't manufacture a `pg_externs.rs` "facade" module — that's anti-pattern. pg_vectorize's `api.rs` is a good model.

### 4. When do I need `pub use` for re-exports?

**Almost never, for pgrx items.** pgrx attributes emit metadata symbols at compile time; they don't rely on Rust path resolution. Plain `mod foo;` is enough. `pub use submodule::thing;` is only needed when:

- Rust code in another submodule imports `thing` by short name and you want it crate-internal-but-not-private
- You're publishing a library crate (which pgrx extensions usually aren't)

For internal cross-module access, prefer `pub(crate)` on the item itself over `pub use` re-exports.

## Patterns to copy

These are harvested from real production pgrx extensions (pg_vectorize, ParadeDB pg_search):

1. **Slim `lib.rs`.** Module declarations, `pgrx::pg_module_magic!()`, optionally a couple of crate-wide constants, the `#[cfg(test)] pub mod pg_test` block. Target ≤200 lines.

2. **One module per concern.** Reasonable splits:
   - `schema.rs` — `extension_sql!` blocks for DDL
   - `bgworker.rs` — `_PG_init`, BackgroundWorker, tick loop
   - `api.rs` — `#[pg_extern]` functions
   - `types.rs` — shared structs/enums used across modules
   - `providers.rs`, `tools.rs`, `auth.rs`, etc. — domain modules

3. **Plain `mod foo;` is enough** for pgrx-decorated items. No `pub use` parade.

4. **`pg_module_magic!()` is invoked exactly once, at crate root.** Don't try to put it in a submodule.

5. **`pub mod pg_test { ... }` block stays at crate root.** The pgrx test framework looks for it there. Same for `#[cfg(any(test, feature = "pg_test"))] mod tests` if used.

6. **Use `pub(crate)` for cross-module visibility on plain Rust items** (structs, enums, statics, helpers). Avoids `pub` (which implies external API) and avoids `pub use` parades.

## Cross-module visibility quick reference

| Item | When | Annotation |
|------|------|------------|
| pgrx-decorated function (`#[pg_extern]`, `#[pg_guard]`) | always | `pub` (the macro requires it) |
| Struct/enum used across submodules | item not exposed to Postgres | `pub(crate)` |
| Static (`OnceLock`, `LazyLock`) used across submodules | internal | `pub(crate) static` |
| Helper function used across submodules | internal | `pub(crate) fn` |
| Item exposed to external crates (rare for extensions) | publishing | `pub` |

## Build / test workflow

For our project specifically (`projects/pg-ai-stewards/extension/`):

```bash
# Full build (Dockerfile-driven, ~60s warm cache, ~5min cold):
cd projects/pg-ai-stewards/extension && docker compose build pg

# Smoke test on ephemeral container:
docker run --rm -d --name pg-smoke \
  -e POSTGRES_USER=stewards -e POSTGRES_PASSWORD=stewards -e POSTGRES_DB=stewards \
  pg-ai-stewards-dev:pg18 \
  postgres -c shared_preload_libraries=pg_ai_stewards
sleep 6
docker exec pg-smoke psql -U stewards -d stewards \
  -c "CREATE EXTENSION IF NOT EXISTS pg_ai_stewards CASCADE;"
docker stop pg-smoke

# Live container restart (after smoke clean):
cd projects/pg-ai-stewards/extension && docker compose down && docker compose up -d
```

**Pause the soak first** if you're going to do multiple restarts:
```sql
UPDATE stewards.watchman_config SET schedule_enabled = false WHERE id = 1;
-- (do refactor work, restart container as needed)
UPDATE stewards.watchman_config SET schedule_enabled = true  WHERE id = 1;
```

## Refactor recipe (multi-file split of an existing lib.rs)

This is the recipe that worked for our 2026-05-08 split:

1. **Map sections.** Use Grep with `^// ==|^// ---|^pub fn |^fn |^struct |^extension_sql_file!|^#\[pg_extern\]` to get a structural outline of the file with line numbers.

2. **Pick the leaf module first.** A module with no pgrx macros and no cross-module dependencies. (For us: `providers.rs` — pure data types + parsing.)

3. **Copy-extract.** Create the new file with the moved code. Mark items `pub(crate)` as needed for cross-module access.

4. **Add the module declaration.** In `lib.rs`: `mod foo;` (with `use foo::{Type, fn, STATIC};` if you want short names in lib.rs).

5. **Delete the moved block from `lib.rs`.** Leave a one-line breadcrumb comment ("Provider types moved to providers.rs (Phase 3c.3.6 module split)").

6. **Build to verify.** `docker compose build pg` from `extension/`. Compiler errors will surface visibility issues. Fix with `pub(crate)` on whatever items the compiler flags.

7. **Smoke test.** Ephemeral container + `CREATE EXTENSION CASCADE`. Counts of agents/tool_defs/pipelines should match pre-split.

8. **Commit + restart live.** Don't restart live until smoke is clean.

## Gotchas (real ones, not theoretical)

- **`pg_module_magic!()` at crate root only.** Calling it twice or in a submodule produces obscure linker errors.
- **`extension_sql_file!` path is file-relative.** Re-check on every move that involves nested directories.
- **The `#[cfg(test)] mod tests` and `pub mod pg_test` blocks must be visible from crate root.** Either keep them in `lib.rs` or `mod`-declare them from `lib.rs`.
- **`#[pg_schema] mod foo { ... }` puts contents in Postgres schema `foo`.** Don't accidentally use it when you wanted the default extension schema.
- **`shared_preload_libraries` discovery is C-side, not Rust-side.** Postgres looks for the `.so` library named in `.control` and the `_PG_init` C symbol. Module location is invisible to it.
- **GitHub issue #2202** (linker errors with `_PG_init` + bindings) is unrelated to module placement — it's about `build.rs` external library linkage. Don't conflate symptoms.
- **`OnceLock` statics in submodules** need `pub(crate) static FOO: OnceLock<T>` to be readable from `lib.rs`. Default visibility hides them.

## Reference projects

When in doubt, look at how production extensions structure things:

- **[pg_vectorize](https://github.com/ChuckHend/pg_vectorize/tree/main/extension/src)** — clean multi-module layout. lib.rs ~30 lines. Modules: `api.rs`, `chat.rs`, `executor.rs`, `guc.rs`, `init.rs`, `transformers/`, `workers/pg_bgw.rs`. Best canonical example.
- **[ParadeDB pg_search](https://github.com/paradedb/paradedb/tree/main/pg_search/src)** — heavier but real. Modules: `api/`, `bootstrap/`, `gucs.rs`, `index/`, `postgres/customscan/`, `schema/`. Shows what large pgrx projects look like.
- **[pgrx official examples](https://github.com/pgcentralfoundation/pgrx/tree/develop/pgrx-examples)** — `bgworker/`, `custom_sql/`, `schemas/`, `triggers/`. Single-file each, but the doc-comments in their lib.rs files are authoritative on attribute semantics.

## Project-specific

For `projects/pg-ai-stewards/extension/` specifically: the multi-module split is in flight as Phase 3c.3.6. Findings doc at `projects/pg-ai-stewards/docs/lib-rs-refactor-findings.md` records what's moved, what's left, and surprises encountered. Keep that file updated as moves land.
