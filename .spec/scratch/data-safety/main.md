# Data Safety — Scratch File

## Binding Problem (copied from proposal)

On March 18, 2026, a single bell icon toggle corrupted practice id 17 ("Pants"). Three bugs: frontend partial PUT, backend blind full-column UPDATE, MCP boolean inversion. Nothing caught it before production. Nothing in the DB prevented it. No audit trail existed.

## Research Log

### 2026-03-18: Research Complete

---

### 1. Dev Agent Gaps (`dev.agent.md`)

**What exists:**
- Go conventions section (go vet, go build, go test commands listed)
- Multi-codebase development rules (6 rules for brain ecosystem)
- "When Making Changes" section (check docs, test against real workflows, update descriptions)
- Build verification commands for all three codebases
- Playwright test instructions

**What's MISSING (and would have caught the March 18 bug):**
- No pre-commit checklist for data-mutating code
- No guidance on partial-update hazards (blind PUT vs. read-modify-write)
- No requirement to write Go unit tests for new handlers
- No destructive-operation warnings or review triggers
- No mention of DB constraints or schema validation
- No testing requirements beyond "go test ./..." (which finds zero tests)
- No plan-exp1 routing suggestion for features that touch data

---

### 2. PUT/PATCH Handler Audit (all handlers in `internal/api/router.go`)

**6 blind overwrite handlers (VULNERABLE — same class of bug as March 18):**
1. `updateTask` — blind full-column UPDATE
2. `updateNote` — blind full-column UPDATE
3. `updatePrompt` — blind full-column UPDATE
4. `updatePillar` — blind full-column UPDATE
5. `updateSource` — blind full-column UPDATE
6. `handleUpdateSettings` — blind full-column UPDATE

**3 now safe (read-modify-write):**
1. `updatePractice` (FIXED March 18 — uses json.RawMessage field detection)
2. `HandleBrainEntryUpdate` — brain entries
3. `HandleBrainSubTaskUpdate` — brain subtasks

**4 specialized/narrow scope (lower risk):**
1. `updateSourceTreeCache` — cache only
2. `updateBookmarkNote` — single field
3. `UpdateMe` — user profile
4. `ChangePassword` — single operation

---

### 3. Database Constraint Inventory

**Current constraints on `practices` table (from schema.sql):**
- `name TEXT NOT NULL` ✓
- `type TEXT NOT NULL` ✓
- `status TEXT NOT NULL DEFAULT 'active'` ✓
- `active BOOLEAN DEFAULT 1` (no NOT NULL!)
- No CHECK constraints on ANY column in the entire database

**Missing constraints that should exist:**
| Table | Column | Missing |
|-------|--------|---------|
| practices | status | CHECK (status IN ('active','paused','completed','archived')) |
| practices | type | CHECK (type IN ('memorize','exercise','habit','task')) |
| practices | active | NOT NULL DEFAULT 1 |
| tasks | status | CHECK (status IN ('active','completed','paused','archived')) |
| tasks | type | CHECK (type IN ('once','daily','weekly','ongoing')) |
| brain_messages | status | CHECK constraint on valid statuses |
| practice_logs | quality | CHECK (quality BETWEEN 0 AND 5) |

**Migration infrastructure:**
- PostgreSQL: goose migrations at `internal/db/migrations/postgres/` (14 migrations, next = 015)
- SQLite: Go-code migrations in `internal/db/` (runSQLiteMigrations, EnsureBrainEntriesTable, etc.)

---

### 4. Test Infrastructure

- **Go unit tests:** ZERO. `file_search("scripts/becoming/**/*_test.go")` returns nothing.
- **Playwright e2e tests:** Exist in `scripts/becoming/frontend/`. Run against dev server.
- **No test database setup:** No test helpers, no test fixtures, no mock DB.
- **Go test command is documented** (`go test ./...`) but has nothing to run.

---

### 5. Audit Log Design Research

**PostgreSQL trigger pattern (standard):**
```sql
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    row_id INTEGER NOT NULL,
    operation TEXT NOT NULL, -- UPDATE | DELETE
    old_data JSONB NOT NULL,
    changed_by INTEGER,     -- user_id if available
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION audit_trigger_func() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (table_name, row_id, operation, old_data, changed_by)
    VALUES (TG_TABLE_NAME, OLD.id, TG_OP, row_to_json(OLD)::jsonb, 
            current_setting('app.current_user_id', true)::integer);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER practices_audit
    BEFORE UPDATE OR DELETE ON practices
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
```

**SQLite equivalent:**
- SQLite supports triggers with same semantics
- No `row_to_json` — must enumerate columns explicitly or use json_object()
- `json_object('id', OLD.id, 'name', OLD.name, ...)` — verbose but works
- SQLite has `json_group_object` in newer versions

**Storage estimation:**
- ~21 active practices, updated ~5-10 times/day → ~50-100 rows/day → ~2-3 KB/day
- Tasks: ~similar volume
- At this scale, retention is effectively unlimited — 1 year = ~1 MB
- Could add retention policy (DELETE WHERE changed_at < NOW() - INTERVAL '1 year') but not urgent

**Key design decisions:**
- Which tables to audit? Start with `practices` and `tasks` (highest risk). Add others later.
- Store full `OLD` row as JSON vs. just changed fields? Full row is simpler, slightly more storage, much simpler to restore from.
- User context in triggers? PostgreSQL `SET LOCAL app.current_user_id = ?` per request. SQLite: pass user_id in application code.

---

### 6. Prior Art in This Project

- No existing audit log pattern anywhere in the codebase
- No existing Go test patterns to follow
- The `updatePractice` fix (read-modify-write with json.RawMessage) is the pattern to replicate
- Admin endpoints (`/api/admin/corrupted-practices`, `/api/admin/recover-practice/:id`) already exist for manual recovery
- `AdminRequired` middleware already built — audit query endpoints would use it

---

## Phase 3a — Critical Analysis

### Is this the RIGHT thing to build, or just the EXCITING thing?

This isn't exciting at all. It's infrastructure born from an actual data loss incident. That's the strongest possible signal — real pain, real cost, real user impact. This is definitively the right thing.

### Does this solve the binding problem, or a different one?

Yes, directly. The binding problem is "nothing detected the bug before production, and nothing in the DB prevented or recorded the corruption." Dev agent hardening addresses detection. DB constraints address prevention. Audit log addresses recording.

### What's the simplest version that would be useful?

**Dev agent hardening:** Just the checklist addition to `dev.agent.md`. Zero code changes. Could be done in 10 minutes and would have prevented March 18. This is Phase 1.

**DB constraints:** One migration file adding NOT NULL / CHECK constraints. Small, safe, high-value. Phase 2.

**Handler fixes:** Apply the read-modify-write pattern to the 6 vulnerable handlers. Moderate scope but each is independent. Phase 3.

**Go tests:** One test file for `updatePractice` partial-update behavior. Proves the pattern, creates the test infrastructure. Phase 4.

**Audit log:** One migration + one trigger + one query endpoint. Phase 5.

### What gets WORSE if we build this?

- **Dev agent checklist:** Adds friction to every data-touching change. Could slow down simple features. But the March 18 incident cost more time than a thousand checklist pauses.
- **DB constraints:** Could break existing code if there are rows violating constraints. Need to audit existing data first. Migration must be tested.
- **Handler fixes:** Risk introducing NEW bugs while fixing old ones. Each fix should be tested.
- **Audit log:** Adds DB write overhead (tiny at this scale). Adds storage (negligible). Adds complexity to migrations. Worth it.

### Does this duplicate something we already have?

No. Zero tests, zero constraints, zero audit logging exist. This is greenfield infrastructure.

### Is this the right time?

The March 18 incident just happened. The pain is fresh, the details are clear, the motivation is real. This is exactly the right time. Deferring would mean the next incident catches us unprepared again — and next time it might be worse than one practice record.

### Mosiah 4:27 check

Michael has a lot in flight (see `.spec/memory/active.md` — overview proposal, brain ecosystem, chip-voice, etc.). But this is defensive infrastructure, not a new feature. It protects existing work. The phasing means he can do Phase 1 (checklist) in minutes and defer the rest if capacity is tight.

### Creation Cycle alignment

This is born from Phase 8 (Atonement — recovery from failure) cycling back to Phase 4 (Spiritual Creation — the spec) before going to Phase 6 (Physical Creation — building it). That's the right order. We learned from the failure, now we're designing the prevention before rushing to code.

### Recommendation: PROCEED — phased delivery, Phase 1 is tiny and high-value
