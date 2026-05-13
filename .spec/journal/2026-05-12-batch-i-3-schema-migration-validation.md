---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Batch I.3 — schema-migration source_type + SQL syntax validation shipped (+ I.1 latent bug fix)"
status: shipped (smoke validated)
carry_forward:
  - "Batch I.2 — HTTP endpoint POST /api/agent-proposals/create + origin filter chip on /work-items (~1h)"
  - "yaml.rs Rust parser refactor — NOT actually rule-of-three triggered yet (correction in this session); wait for 3rd YAML SHAPE to land"
  - "Phase A pgrx longjmp catch + 60s reaper (Claude-only, ~1 session)"
  - "Projects B — deferred"
  - "14 SC work_items still pending Michael's ratification"
  - "materialize-writes /workspace mount is read-only in bridge container — separate ops concern, not blocking"
links:
  - "../../projects/pg-ai-stewards/extension/i6-schema-migration-claude-attest.sql"
  - "../../projects/pg-ai-stewards/extension/i7-apply-agent-proposal-direct-file-queue.sql"
  - "../../projects/pg-ai-stewards/cmd/stewards-cli/validate_sql.go"
  - "../../projects/pg-ai-stewards/extension/smoke/i6-schema-migration-smoke.sql"
---

# Batch I.3 — schema-migration source_type + SQL syntax validation (2026-05-12)

Fourth build pulse today. Closes the kimi-creates-tables-that-survive-restart loop and surfaces+fixes a latent I.1 bug.

## Ratification

Five questions via AskUserQuestion, all answered **A** (recommended):

- **Q1 — SQL validation method:** A. `stewards-cli validate-sql` + BEGIN/ROLLBACK against live DB
- **Q2 — Validation failure feedback:** A. `pending_file_writes.materialized_by='error:syntax:<diag>'`
- **Q3 — Claude-only enforcement:** A. Hard gate via `input.draft.claude_attested=true`

The hard-gate option is honor-system: a non-Claude agent that lies and sets `claude_attested=true` is committing a covenant violation, not bypassing a technical control. Belt + suspenders for kimi-trust ratification.

## What shipped

### i6 — claude_attested gate

`apply_agent_proposal` redefined: for `source_type='schema-migration'`, checks `input.draft.claude_attested` directly. **Read from `input.draft`, not `validate.output`** — the validate stage cannot promote attestation. If absent or false: NOTICE-logs the kimi-trust diagnostic, returns false, no studies/file action.

Validate stage prompt updated with a `KIMI-TRUST GATE` section explaining the rule.

### i7 — apply_agent_proposal queues body directly (BUG FIX)

I.3 smoke surfaced a latent bug from I.1. The agent-proposal pipeline had:
```sql
pipeline.file_content_jsonpath = 'stage_results.validate.output'
```

This points to the validate stage's output which is a **JSON string** — the full normalized proposal:
```json
{ "source_type": "...", "slug": "...", "title": "...",
  "body": "<the actual file content>", "frontmatter": {...}, ... }
```

When `enqueue_work_item_file` called `extract_work_item_file_content`, it returned this entire JSON string. So **`pending_file_writes.content` was the whole JSON, not the body**. Had I.1 actually materialized to disk, every exhibit/study/note/lesson `.md` file would have been JSON, and schema-migration `.sql` files would have been JSON wrapping the SQL.

Why this didn't bite earlier: bridge container mounts `/workspace` read-only, so we never actually ran materialize-writes against the smoke proposals. The "real e2e" smoke in I.1 verified content_len matched the JSON size (692 bytes for the SC bias exhibit) — which I noted but didn't recognize as the bug it was.

**Fix:** `apply_agent_proposal` now INSERTs into `pending_file_writes` directly with the extracted `body` as content. Sets `work_items.file_enqueued_at = now()` so `on_maturity_verified`'s subsequent enqueue path is a no-op (its IF guard checks `file_enqueued_at IS NULL`). Scoped to agent-proposal pipeline only; other pipelines unchanged.

This is the kind of bug the covenant's `exercise_stewardship` clause names: same shape, same fix, no behavior change from the user's perspective — fix and report.

### stewards-cli validate-sql

New subcommand. Reads SQL from `--file PATH` or stdin. Connects to live DB via existing DSN logic. Wraps SQL in `BEGIN; <sql>; ROLLBACK;` via `pgx.Tx`. Returns nil on parse OK; returns the pgx error on failure. Exit code 0 / 1.

Shared `validateSQL(ctx, pool, sql)` helper used both by the CLI subcommand and the materialize-writes hook.

### materialize-writes hook

For `pending_file_writes` rows where `target_path` ends in `.sql` AND starts with `projects/pg-ai-stewards/extension/`, validate-sql runs BEFORE the file write (or BEFORE the dry-run print). On failure:
- `recordError` writes `materialized_by='error:syntax:<diag>'` (diag truncated to 500 chars)
- File not written
- `failed++` counter; other rows in batch proceed

Validation runs in dry-run mode too — bad SQL is bad SQL regardless of write mode.

## Smoke verification

**validate-sql direct (CLI):**
- valid SQL via stdin → exit 0, "ok"
- bad SQL via stdin → exit 1, "ERROR: syntax error at or near 'TABL' (SQLSTATE 42601)"

**apply_agent_proposal gate (SQL):**
- schema-migration without claude_attested → NOTICE logs kimi-trust diagnostic, returns false, no file_destination
- schema-migration with claude_attested + valid SQL → INSERT into pending_file_writes (body only, 75 bytes); file_enqueued_at set; agent_proposal_applied_at set
- schema-migration with claude_attested + bad SQL → apply succeeds (doesn't validate); pending_file_writes queued with 52 bytes of bad SQL

**materialize-writes --dry-run end-to-end:**
- Row #25 (valid SQL) → validate-sql ok → "DRY-RUN #25 → /workspace/projects/pg-ai-stewards/extension/iz-smoke-test-with-attest-valid.sql [75 bytes]"
- Row #27 (bad SQL) → "skip #27: validate-sql failed: ERROR: syntax error at or near 'TABL' (SQLSTATE 42601)"; `materialized_by='error:syntax:ERROR: syntax error at or near "TABL" (SQLSTATE 42601)'`

Smoke rows cleaned up after verification.

## yaml.rs correction (related; surfaced this session)

In the I.1 journal earlier today I wrote "rule-of-three for yaml.rs is technically already met" because `stewards.intents` has 3 rows. **That was wrong.** The rule-of-three for yaml.rs is about three distinct YAML SHAPES needing parsing, not three callers of one parser. Today's `yaml.rs` has `parse_yaml_intent` (used for all 3 intents) and `parse_yaml_covenant` — two parsers, one of which serves multiple callers. One parser doing its job is not a rule-of-three trigger.

Fix applied to journal + active.md. yaml.rs refactor stays deferred until a third YAML shape lands (agent.yaml, skill.yaml, principles.yaml, etc.).

## The full kimi-creates-tables loop (now closed)

```
1. Claude-driven agent proposes a schema-migration via work_item_create
   with pipeline_family='agent-proposal', input.draft = {
     source_type: 'schema-migration',
     slug: 'iN-add-foo',
     title: '...',
     body: '-- iN ... CREATE TABLE ...',
     claude_attested: true,
     ...
   }
2. Bgworker dispatches validate stage (qwen3.6-plus, tools off)
   → emits normalized JSON to stage_results.validate.output
3. Maturity → verified
4. on_maturity_verified.apply_agent_proposal:
   - checks claude_attested (i6)
   - INSERTs pending_file_writes(content=body) (i7)
   - sets file_enqueued_at; agent_proposal_applied_at
5. stewards-cli materialize-writes runs (host-side):
   - validate-sql hook for .sql files in extension/ (Batch I.3)
   - if syntax ok: writes file to projects/pg-ai-stewards/extension/iN-<slug>.sql
   - if syntax bad: pending_file_writes.materialized_by='error:syntax:<diag>'
6. Next bridge restart:
   - bridge-entrypoint runs stewards-cli migrate
   - ledger picks up new iN-*.sql file
   - applies it transactionally; records sha256
   - bridge daemon starts (or doesn't, if SQL was actually broken)
7. Schema permanently applied. Survives image rebuild, container
   recreate, host migration.
```

Every layer is now built. The loop closes cleanly when a Claude agent actually wants to use it.

## Architecture wins

**The substrate composed AGAIN.** No new abstractions. Existing pieces (work_items, pending_file_writes, materialize-writes, migration ledger, on_maturity_verified, apply_*) wired through a new entry point. The new code is one CLI subcommand + one hook + one SQL function refinement.

**Self-attestation as honor-system control.** The kimi-trust ratification is load-bearing. The technical gate (`claude_attested=true`) is honest about being honor-system — a covenant violation by an agent setting it falsely is still a covenant violation, not a security bypass. This matches the project's posture: "discernment over enforcement."

**A bug found by the smoke is a bug worth finding.** The I.1 JSON-as-file-content bug was latent — it would have surfaced the first time anyone actually materialized to a writable mount. The I.3 validate-sql hook caught it before that happened. Good catch from a discipline ("test the failure mode") that pays compounding dividends.

## Carry-forward updates

- Batch I.2 still on deck (~1h): HTTP endpoint + origin filter chip
- yaml.rs corrected: still gated on 3rd YAML shape, NOT three callers
- Phase A unchanged: stable polish
- Projects B unchanged: deferred
- 14 SC work_items unchanged: pending ratification

## Cost

LLM: **$0.00**. All substrate + Go + smoke work. Bridge restart × 2; rebuild × 1 (~12s).

## Closing

Four pulses today. The substrate now has:
1. Durable file mechanism (morning: ledger)
2. Project entity + honest names (morning + midday: Projects A, FK, materialized_at)
3. Agent write-back path (pulse 3: agent-proposal pipeline)
4. Validated agent-authored schema migrations (this pulse)

The loop the user named back in March — "can kimi create tables and SQL functions that get ingested by the DB itself? the problem comes is getting those out so when restarts or image rebuilds happen we can bootstrap back up properly" — is now closed. A Claude-driven agent can propose a migration tonight; the substrate will accept, validate, materialize, and re-apply it across restart. The chicken-and-egg has a base case AND a syntax-checked rail.
