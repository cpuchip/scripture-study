# pg-ai-stewards — Claude Code project context

> Per-project context auto-loaded when working in `projects/pg-ai-stewards/`. Repo-root [CLAUDE.md](../../CLAUDE.md) still applies; this file adds substrate-specific guidance.

## 1. Substrate state (2026-05-11)

Feature-complete through **Phase F**. Six phases of the agentic creation cycle running on dev:

| Phase | What |
|---|---|
| A | Watch → Diagnose → Act → Account loop |
| B | Maturity ladder + gates |
| C | Intent + covenant as first-class state |
| D | Atonement + Sabbath + Consecration |
| E | Trust ladder + line-upon-line |
| F | Multi-agent council (Zion) |

Next moves are about **USE, not BUILD** — see `.spec/open-items.md` Section 0 for the active proposal queue.

## 2. Where things live

| Need | Path |
|---|---|
| **What's open / what to work on next** | [`.spec/open-items.md`](.spec/open-items.md) — navigation hub |
| **Active proposals (build-ready or needing ratification)** | `.spec/proposals/substrate-*.md` |
| **Deferred catalog** (don't start sessions from here) | [`.spec/proposals/substrate-deferred-items.md`](.spec/proposals/substrate-deferred-items.md) |
| **Phase journals** (per-session memory) | repo-root `.spec/journal/2026-05-*-substrate-*.md` |
| **Historical phase tracking** | [`phases.md`](phases.md) — 1500+ lines; carry-forwards now in open-items |
| **Foundational substrate proposal** | [`.spec/proposals/full-agentic-substrate.md`](.spec/proposals/full-agentic-substrate.md) — §VI has all ratified decisions + 2026-05-11 amendments |
| **Extension code** | `extension/src/*.rs` + `extension/N*.sql` |
| **Stewards-UI code** | `../../scripts/stewards-ui/` (api/ + frontend/) |
| **Stewards-MCP code** | `cmd/stewards-mcp/` (the MCP server Claude Code uses to read substrate) |

## 3. Build cadence (the C–F pattern that worked)

1. **Decisions upfront.** Walk ratification questions via `AskUserQuestion` batches BEFORE writing code. Phases C/D/E/F each opened with 3-7 decision points.
2. **Gated phased commits.** Each commit is one named sub-step (C.1, C.2, …). Smoke test BEFORE committing.
3. **Memory update at session end.** Always: journal entry in `.spec/journal/` + active.md update. Covenant `update_memory` is load-bearing.

Don't batch multiple sub-steps into one commit. Don't skip the smoke. The C–F cadence shipped 30+ commits in 4 sessions with zero rollbacks because of this discipline.

## 4. Key conventions

**SQL file naming.** `N{a-z}-name.sql` per phase. `extension_sql_file!` chain in `src/lib.rs` defines dependency order. Same file added to Dockerfile COPY list. Idempotent statements only (`CREATE TABLE IF NOT EXISTS`, `CREATE OR REPLACE FUNCTION`, `ON CONFLICT DO UPDATE`).

**Lesson #3 fix.** After `docker compose build pg`, the PostToolUse hook auto-runs `scripts/bump-extension.sh` which extracts pgrx `CREATE FUNCTION` blocks from the bundled SQL and re-registers them in `pg_proc`. **Don't manually `CREATE FUNCTION ... AS '$libdir/pg_ai_stewards', '<name>_wrapper'`** — that's the workaround the bump script replaced. See `.claude/skills/pgrx-extension-bump/SKILL.md`.

**Live-migration for SQL.** New SQL files apply via `docker cp + psql -f`. Pattern:
```powershell
docker cp "C:/.../extension/NX-name.sql" pg-ai-stewards-dev:/tmp/X.sql
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -f /tmp/X.sql
```

## 5. Live container topology

| Container | Role |
|---|---|
| `pg-ai-stewards-dev` | Postgres + extension + bgworker (4 parallel workers, 500ms poll) |
| `pg-ai-stewards-ui` | Stewards-UI Go binary at http://127.0.0.1:8080 |
| `pg-ai-stewards-bridge` | MCP bridge daemon (outbound tool dispatch) |

**Soak pause/resume protocol** (do this for build sessions):
```sql
-- Pause at session start
UPDATE stewards.watchman_config SET schedule_enabled = false WHERE id = 1;
-- Resume at session end
UPDATE stewards.watchman_config SET schedule_enabled = true  WHERE id = 1;
```

**Restart sequence for code changes:**
- SQL-only: live-apply via `docker cp + psql -f`, no restart
- bgworker.rs / yaml.rs / tools.rs: `docker compose build pg && docker compose down && docker compose up -d pg ui` (bridge stays down during pg rebuild, restart after)
- UI: `docker compose build ui && docker compose up -d ui` (no pg downtime)
- **`docker compose down -v` wipes the data volume.** Never run unless intentionally resetting dev.

## 6. Bgworker auto-fire markers

7 markers currently switched on in `bgworker.rs` chat-completion path (refactor to `payload._kind` enum tracked as carry-forward when the 8th lands):

| Marker | Fires | Apply function |
|---|---|---|
| `_gate_eval` | evaluate_gate completes | `apply_gate_decision` |
| `_scenarios_gen` | generate_scenarios completes | `apply_scenarios_result` |
| `_verify` | verify_work_item completes | `apply_verify_result` |
| `_sabbath` | sabbath_dispatch completes | `apply_sabbath_result` |
| `_atonement` | atonement_dispatch completes | `apply_atonement_result` |
| `_council_member` | proposer/critic responds | UPDATE council_members + auto-fire `synthesize_council` when all done |
| `_council_synthesize` | synthesizer completes | `apply_synthesize_result` |

Errors logged via `pgrx::log!`, never propagated. Failed auto-apply leaves the work_item un-transitioned for human re-trigger.

## 7. Cost discipline

- **Tools-disabled gate prompts.** Phase B lesson 2026-05-11: gate-eval through `plan` agent with tools enabled cost 11 chats (~$0.04); tools-off cost 1 chat ($0.005) — 7× reduction. Every JSON-output gate (`_gate_eval`, `_scenarios_gen`, `_verify`, `_sabbath`, `_atonement`, `_council_synthesize`) sets `tools_disabled=true` on the payload.
- **Per-work-item cost cap** enforced by the steward at quarantine time. Cap configured in `model_pricing` rates × token usage.
- **Bucket caps:** $12/day + $60/month soft caps (OpenCode Go subscription); 4-bucket tracking (5h/daily/weekly/monthly) in `stewards.cost_buckets`.
- **`compose_system_prompt` adds ~600 tokens/dispatch** (covenant + intent blocks). Acceptable; flagged for measurement on first real workload — see `open-items.md` § X.10 in deferred items.

## 8. Don't do

- **Don't** manually `CREATE FUNCTION ... AS '$libdir/pg_ai_stewards', '<name>_wrapper'` — use the bump script (§4).
- **Don't** `docker compose down -v` — wipes substrate data.
- **Don't** skip the smoke between commits in a phased build — the C–F discipline depends on it.
- **Don't** edit Cargo.toml / .control versions manually — the bump script handles version bookkeeping when needed.
- **Don't** start a build session from `open-items.md` — it's the index. Start from an active proposal (§2).
- **Don't** ship destructive SQL (`DROP TABLE`, `TRUNCATE`, schema renames) without an explicit ratification in the same session — substrate data is durable on purpose.
- **Don't** add new pg_extern functions without confirming the bump-extension hook fired post-rebuild (function will be silently missing from pg_proc otherwise).
- **Don't** forget the `update_memory` covenant — journal + active.md at session end is non-optional for substantive sessions.
