# Contributing to pg-ai-stewards

Thanks for looking. This is a single-maintainer research-grade project that
grew into a real substrate; the conventions below are what kept ~200 commits
across a dozen build batches at zero rollbacks.

## The build cadence (the discipline that worked)

1. **Decisions upfront.** Walk the design/ratification questions *before*
   writing code. Each phase opened with 3–7 explicit decision points.
2. **Gated, phased commits.** One named sub-step per commit. Smoke-test
   *before* committing.
3. **Memory at session end.** A journal entry + state update every substantive
   session. Provenance is load-bearing here, not busywork — see
   [docs/history/](docs/history/).

Don't batch multiple sub-steps into one commit. Don't skip the smoke.

## Where things live

| Need | Path |
|---|---|
| Runtime map ("what do I query?") | [docs/architecture.md](docs/architecture.md) |
| Extension code (Rust/pgrx + SQL) | `extension/src/*.rs`, `extension/*.sql` |
| Go sidecars (MCP, CLI, bridge) | `cmd/stewards-mcp/`, `cmd/stewards-cli/`, `cmd/fs-read-mcp/` |
| Web console | `../../scripts/stewards-ui/` (pre-extraction; moves in standalone) |
| Design history / provenance | [docs/history/](docs/history/), `.spec/proposals/`, `.spec/journal/` |
| What's open / next | `.spec/open-items.md` |
| AI-agent contributor context | [CLAUDE.md](CLAUDE.md) |

## SQL conventions

- **File naming:** `N{a-z}-name.sql` per phase/batch. The `extension_sql_file!`
  chain in `src/lib.rs` defines load order for the embedded extension; the same
  file is added to the Dockerfile COPY list. Live migrations apply via
  `stewards-cli migrate` (a sha-tracked ledger).
- **Idempotent statements only:** `CREATE TABLE IF NOT EXISTS`,
  `CREATE OR REPLACE FUNCTION`, `ON CONFLICT DO UPDATE`. Migrations re-run safely.
- **Live-apply pattern** (no rebuild for SQL-only changes):
  ```
  docker cp extension/NX-name.sql pg-ai-stewards-dev:/tmp/x.sql
  docker exec pg-ai-stewards-dev psql -U stewards -d stewards -f /tmp/x.sql
  ```
- **Never** ship destructive SQL (`DROP TABLE`, `TRUNCATE`, renames) without an
  explicit decision in the same change — substrate data is durable on purpose.

## The load-bearing architectural rule

**Foreground SQL functions never call an LLM provider.** They write rows and
`NOTIFY`. The Rust bgworker (its own tokio runtime) dispatches model + tool
calls and writes results back. This isolation is the whole point — keep it.

## Rebuild vs. live-apply

- **SQL-only:** live-apply, no restart.
- **bgworker / Rust change:** `docker compose build pg && docker compose down &&
  docker compose up -d pg ui`, then bring `bridge` back. (Pause the watchman
  soak first if one is running.)
- **UI change:** `docker compose build ui && docker compose up -d ui` — no pg
  downtime.
- **Never** `docker compose down -v` — it wipes the data volume.

## Cost discipline

LLM calls cost money. Gate-style dispatches set `tools_disabled=true` (≈7×
cheaper). Work items carry per-item cost caps; providers can carry enforced
prepaid spend caps (`stewards.provider_spend_caps`) that refuse dispatch
before spending. Don't remove these gates.

## Status

Feature-complete and running, but **not yet independently buildable** as a
standalone checkout — see
[.spec/proposals/standalone-extraction.md](.spec/proposals/standalone-extraction.md)
for the decoupling work. Until then, build from the parent workspace.
