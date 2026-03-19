# Data Safety: Dev Agent Hardening + Audit Log

## Binding Problem

On March 18, 2026, a single bell icon toggle on ibeco.me corrupted a practice record — wiping its name, type, status, and active flag. The practice vanished from all views. Three bugs worked together: a frontend partial PUT, a backend blind full-column UPDATE, and an MCP boolean inversion. None were caught before deployment.

**The problem is not that the bug happened.** Bugs are inevitable. The problem is that **nothing in our process detected it before production**, and **nothing in the database prevented or recorded the corruption**.

Two binding problems, one root:

1. **Prevention:** The dev agent has no testing requirements, no destructive-operation checklist, and no guardrails for data-mutating code. There are zero Go unit tests in the entire becoming backend. The database has zero CHECK constraints.

2. **Recovery:** When data is corrupted or lost, there is no row-level recovery mechanism. The only recovery path today is hand-written admin endpoints built after the fact. An audit log would preserve previous row states automatically.

---

## Scope

### A. Dev Agent Hardening (Prevention)

Process and convention changes that make it harder for the dev agent (or a human) to ship data-corrupting code:

1. **Pre-commit checklist** added to `.github/agents/dev.agent.md`
2. **DB constraint migration** — NOT NULL / CHECK constraints on critical columns
3. **Handler remediation** — convert 6 blind-overwrite PUT handlers to read-modify-write
4. **Go test infrastructure** — test helpers + seed tests for CRUD partial-update behavior

### B. Audit Log (Recovery)

Automatic row-level history that captures the previous state before any UPDATE or DELETE:

1. **Audit log table + trigger** — PostgreSQL and SQLite
2. **Admin query endpoint** — view audit history for a given table/row
3. **MCP tool** — query audit log from the MCP server

### Not in Scope

- **S3/backup infrastructure** — Michael handles this separately via Dokploy
- **Local DB mirror** — rejected as unnecessary complexity
- **Full ORM or migration framework change** — we keep goose + Go-code SQLite migrations

---

## Success Criteria

- [ ] Dev agent includes a data-safety checklist that would have caught the March 18 bug
- [ ] `practices` table has CHECK constraints on `status` and `type` columns
- [ ] `practices.active` has NOT NULL constraint
- [ ] All 6 blind-overwrite handlers converted to read-modify-write
- [ ] At least one Go test file exists testing partial-update behavior
- [ ] Audit log captures previous row state for UPDATE/DELETE on `practices` table
- [ ] Audit log is queryable via admin endpoint
- [ ] Both SQLite and PostgreSQL implementations work

---

## Prior Art & Related Work

- **The March 18 fix itself** is the reference implementation: `updatePractice` in `internal/api/router.go` now does read-modify-write with `json.RawMessage` field detection (lines 278-360). This is the pattern all handlers should follow.
- **Admin endpoints** already exist: `GET /api/admin/corrupted-practices`, `POST /api/admin/recover-practice/:id`, protected by `AdminRequired` middleware.
- **No existing audit log, Go tests, or CHECK constraints** anywhere in the codebase. This is entirely new infrastructure.
- **Playwright e2e tests** exist for the frontend. The Go backend has zero test files.

---

## Proposed Approach

### Phase 1: Dev Agent Checklist (immediate, zero code)

Add a "Data Safety Checklist" section to `.github/agents/dev.agent.md` under "When Making Changes." The checklist triggers when any change touches a PUT/PATCH handler, a DB UPDATE/DELETE query, or a migration.

**Checklist content:**

```markdown
### Data Safety Checklist

When a change touches any PUT/PATCH handler, UPDATE/DELETE query, or database migration:

1. **Partial update safe?** Does the handler use read-modify-write (fetch existing → overlay sent fields → save)? A blind `UPDATE ... SET col1=?, col2=?, ...` from decoded request body will zero-value any field the client didn't send. See `updatePractice` in `internal/api/router.go` for the correct pattern.

2. **DB constraints enforced?** Are critical columns protected by NOT NULL and CHECK constraints? If adding a new column that has a finite set of valid values, add a CHECK constraint in the migration.

3. **Both databases migrated?** Does the change require a schema change? If yes:
   - PostgreSQL: goose migration in `internal/db/migrations/postgres/`
   - SQLite: migration in `internal/db/` Go code
   Forgetting one breaks either dev or production.

4. **Test coverage?** Is there a Go test that sends a partial update (missing fields) and verifies the existing values are preserved? If not, write one.

5. **Frontend sends full object?** If the frontend calls PUT on a resource, does it send the complete current state, not just the changed field? Check the API call payload.

6. **Destructive operation review?** DELETE endpoints, status changes, archive operations — verify the operation is reversible or has confirmation UI.
```

**Why this is Phase 1:** Zero code changes. Purely agent instructions. Could have prevented March 18 entirely. Can be done right now.

### Phase 2: DB Constraints Migration

Add migration `015_data_safety_constraints.sql` (PostgreSQL) and corresponding SQLite migration.

**PostgreSQL migration:**

```sql
-- +goose Up

-- practices: enforce valid status and type values
ALTER TABLE practices ALTER COLUMN active SET NOT NULL;
ALTER TABLE practices ADD CONSTRAINT practices_status_check 
    CHECK (status IN ('active', 'paused', 'completed', 'archived'));
ALTER TABLE practices ADD CONSTRAINT practices_type_check 
    CHECK (type IN ('memorize', 'exercise', 'habit', 'task'));

-- tasks: enforce valid status and type values  
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check 
    CHECK (status IN ('active', 'completed', 'paused', 'archived'));
ALTER TABLE tasks ADD CONSTRAINT tasks_type_check 
    CHECK (type IN ('once', 'daily', 'weekly', 'ongoing'));

-- practice_logs: enforce quality range
ALTER TABLE practice_logs ADD CONSTRAINT practice_logs_quality_check 
    CHECK (quality IS NULL OR (quality >= 0 AND quality <= 5));

-- +goose Down
ALTER TABLE practices ALTER COLUMN active DROP NOT NULL;
ALTER TABLE practices DROP CONSTRAINT IF EXISTS practices_status_check;
ALTER TABLE practices DROP CONSTRAINT IF EXISTS practices_type_check;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_type_check;
ALTER TABLE practice_logs DROP CONSTRAINT IF EXISTS practice_logs_quality_check;
```

**SQLite migration** (in Go code — `runSQLiteMigrations`):

SQLite doesn't support `ALTER TABLE ADD CONSTRAINT`. Two options:
- **Option A (recommended):** Add CHECK constraints to `schema.sql` for new databases. Existing dev databases accept any value but new inserts are validated. Accept the gap — dev databases are ephemeral.
- **Option B:** Recreate tables with constraints (complex, risky for dev). Not worth it.

**Pre-flight check:** Before deploying, verify no existing rows violate the new constraints:
```sql
SELECT id, name, status, type FROM practices WHERE status NOT IN ('active','paused','completed','archived');
SELECT id, name, status, type FROM practices WHERE type NOT IN ('memorize','exercise','habit','task');
SELECT id FROM practices WHERE active IS NULL;
```

### Phase 3: Handler Remediation

Convert these 6 handlers from blind overwrite to read-modify-write, following the `updatePractice` pattern:

| Handler | Table | Risk Level |
|---------|-------|------------|
| `updateTask` | tasks | HIGH — same field structure as practices |
| `updateNote` | notes | MEDIUM — fewer critical fields |
| `updatePrompt` | prompts | LOW — simple structure |
| `updatePillar` | pillars | LOW — simple structure |
| `updateSource` | sources | MEDIUM — document references |
| `handleUpdateSettings` | user_settings | LOW — but still wrong pattern |

**Pattern to apply (from the fixed `updatePractice`):**
1. `io.ReadAll(r.Body)` → get raw bytes
2. `json.Unmarshal` into `map[string]json.RawMessage` → know which fields were sent
3. Fetch existing record from DB
4. For each field: if present in map, overlay from decoded struct
5. Save the merged result

Each handler is independent — can be done one at a time, tested, and committed separately.

### Phase 4: Go Test Infrastructure

Create the first Go test file: `internal/api/router_test.go` (or `internal/api/practice_test.go`).

**Infrastructure needed:**
- Test database setup (SQLite in-memory or temp file)
- Schema initialization helper
- Test user creation helper
- HTTP test helpers (httptest.NewServer or httptest.ResponseRecorder)

**Seed test cases for `updatePractice`:**

```go
func TestUpdatePractice_PartialUpdate_PreservesExistingFields(t *testing.T) {
    // Create practice with name="Test", type="habit", category="fitness"
    // Send PUT with only {config: "..."}
    // Assert name, type, category are unchanged
}

func TestUpdatePractice_FullUpdate_UpdatesAllFields(t *testing.T) {
    // Create practice with all fields
    // Send PUT with all fields changed
    // Assert all fields updated
}

func TestUpdatePractice_EmptyStringName_Rejected(t *testing.T) {
    // Create practice
    // Send PUT with {name: ""}
    // Assert 400 or name unchanged (depends on desired behavior)
}
```

After infrastructure is in place, add similar tests for each handler fixed in Phase 3.

### Phase 5: Audit Log

**PostgreSQL migration** (`016_audit_log.sql`):

```sql
-- +goose Up
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    row_id INTEGER NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('UPDATE', 'DELETE')),
    old_data JSONB NOT NULL,
    changed_by INTEGER,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_table_row ON audit_log(table_name, row_id);
CREATE INDEX idx_audit_log_changed_at ON audit_log(changed_at);

-- Audit function: captures OLD row as JSON before UPDATE/DELETE
CREATE OR REPLACE FUNCTION audit_trigger_func() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (table_name, row_id, operation, old_data, changed_by)
    VALUES (
        TG_TABLE_NAME, 
        OLD.id, 
        TG_OP, 
        row_to_json(OLD)::jsonb,
        NULLIF(current_setting('app.current_user_id', true), '')::integer
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Attach to practices and tasks tables
CREATE TRIGGER practices_audit
    BEFORE UPDATE OR DELETE ON practices
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER tasks_audit
    BEFORE UPDATE OR DELETE ON tasks
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

-- +goose Down
DROP TRIGGER IF EXISTS practices_audit ON practices;
DROP TRIGGER IF EXISTS tasks_audit ON tasks;
DROP FUNCTION IF EXISTS audit_trigger_func();
DROP TABLE IF EXISTS audit_log;
```

**Setting user context per request:**

Add middleware or per-request setup that calls:
```sql
SET LOCAL app.current_user_id = ?;
```
before each request's DB operations. This allows the trigger to capture who made the change. `SET LOCAL` scopes to the transaction and automatically resets.

**SQLite equivalent:**

SQLite triggers work but require explicit column enumeration:
```sql
CREATE TRIGGER IF NOT EXISTS practices_audit_update
BEFORE UPDATE ON practices
FOR EACH ROW
BEGIN
    INSERT INTO audit_log (table_name, row_id, operation, old_data, changed_at)
    VALUES ('practices', OLD.id, 'UPDATE', 
        json_object('id', OLD.id, 'name', OLD.name, 'type', OLD.type, 
                     'status', OLD.status, 'active', OLD.active, 
                     'category', OLD.category, 'config', OLD.config,
                     'description', OLD.description),
        datetime('now'));
END;
```
No `changed_by` in SQLite (no session variables). Acceptable for dev mode.

**Admin query endpoint:**

```
GET /api/admin/audit-log?table=practices&row_id=17&limit=50
```

Returns audit entries sorted by `changed_at DESC`. Protected by `AdminRequired` middleware.

**MCP tool (optional, Phase 5b):**

`get_audit_history` tool — table name + row ID → list of previous states. Lower priority than the admin endpoint.

---

## Phased Delivery

| Phase | Deliverable | Scope | Dependencies |
|-------|-------------|-------|--------------|
| 1 | Dev agent checklist | `.github/agents/dev.agent.md` edit | None |
| 2 | DB constraints migration | Migration 015 + schema.sql update | Pre-flight data check |
| 3 | Handler remediation | 6 handlers in `router.go` | Phase 1 (follow checklist) |
| 4 | Go test infrastructure | Test helpers + seed tests | Phase 3 (test the fixes) |
| 5 | Audit log | Migration 016 + admin endpoint | Phase 2 (after constraints) |

Each phase delivers value independently. Phase 1 can be done right now with no code changes. Phases 2-5 can be deferred, reordered, or dropped without invalidating earlier phases.

---

## Verification Criteria

### Phase 1
- [ ] The dev agent, given a task to "add a notification toggle that calls PUT with `{notify: true}`", follows the checklist and identifies the partial-update hazard before writing code

### Phase 2
- [ ] `INSERT INTO practices (user_id, name, type, status) VALUES (1, '', 'habit', 'active')` fails (empty name)
- [ ] `INSERT INTO practices (user_id, name, type, status) VALUES (1, 'Test', 'invalid', 'active')` fails (bad type)
- [ ] `UPDATE practices SET status = 'bogus' WHERE id = 1` fails (bad status)
- [ ] Existing data passes pre-flight check before migration

### Phase 3
- [ ] For each fixed handler: send PUT with only one field changed → verify all other fields preserved
- [ ] `go vet ./...` passes
- [ ] `npx vue-tsc --noEmit` passes

### Phase 4
- [ ] `go test ./internal/api/...` runs and passes
- [ ] Partial-update test sends `{config: "{}"}` and confirms name/type/status unchanged

### Phase 5
- [ ] Update a practice → `SELECT * FROM audit_log WHERE table_name='practices'` shows the old values
- [ ] Delete a practice → audit log captures the deleted row
- [ ] `GET /api/admin/audit-log?table=practices&row_id=17` returns history
- [ ] SQLite triggers work in dev mode

---

## Costs and Risks

### Costs
- **Phase 1:** ~10 minutes of writing. No code risk.
- **Phase 2:** One migration. Risk: existing data violates constraints (mitigated by pre-flight check).
- **Phase 3:** 6 handler changes. Moderate risk — each could introduce a new bug (mitigated by Phase 4 tests). Touches the most sensitive code in the app.
- **Phase 4:** Test infrastructure from scratch. One-time investment, pays off every future PR.
- **Phase 5:** Database trigger + admin endpoint. Low risk, well-understood pattern.

### Risks
- **Handler remediation is the riskiest phase.** Changing 6 update endpoints at once is exactly the kind of batch change that created the original bug. Mitigation: do them one at a time, test each, commit each.
- **Constraint migration could fail** if existing data is dirty. Mitigation: run pre-flight SQL before deploying.
- **SQLite parity is imperfect.** SQLite can't do `ALTER TABLE ADD CONSTRAINT` or session-scoped variables. Accept the gap — dev mode doesn't need 100% parity, just functional equivalence.

---

## Creation Cycle Review

| Step | Question | This Proposal |
|------|----------|---------------|
| **1. Intent** | Why are we doing this? | A real data loss incident happened. This prevents the next one and enables recovery if prevention fails. |
| **2. Covenant** | Rules of engagement? | Dev agent checklist is a covenant — the agent commits to checking before shipping data-touching code. Human commits to not bypassing the checklist for "quick fixes." |
| **3. Stewardship** | Who owns what? | Dev agent owns the checklist. Developer (human or agent) owns the migrations and handler fixes. AdminRequired middleware owns access to audit data. |
| **4. Spiritual Creation** | Is the spec precise enough? | Yes. Each phase has concrete deliverables, file locations, SQL, and Go code patterns. An executing agent can build against this without ambiguity. |
| **5. Line upon Line** | What's the phasing? | 5 phases, each independent. Phase 1 (checklist) stands alone and delivers value in minutes. |
| **6. Physical Creation** | Who executes? | Dev agent for Phases 2-5. Michael can do Phase 1 manually or hand it to dev agent. |
| **7. Review** | How do we know it's right? | Verification criteria for each phase. Phase 2 has pre-flight SQL. Phase 3-4 have Go tests. Phase 5 has audit log query verification. |
| **8. Atonement** | What if it goes wrong? | This entire proposal IS the Atonement step — born from the March 18 failure. If a migration fails, goose rollback. If a handler fix introduces a new bug, the new Go tests should catch it. |
| **9. Sabbath** | When do we stop and reflect? | After each phase. Phase 1 is a natural pause — deploy the checklist, use it for a week, then decide if Phases 2-5 are still needed. |
| **10. Consecration** | Who benefits? | Michael's daughter (the app user). Michael (the developer). Future dev agents (the checklist). Any contributor. |
| **11. Zion** | How does this serve the whole? | Integrates with existing admin endpoints, existing middleware, existing migration infrastructure. Doesn't create parallel systems. The checklist improves all future dev agent work, not just this fix. |

---

## Status: READY FOR REVIEW
