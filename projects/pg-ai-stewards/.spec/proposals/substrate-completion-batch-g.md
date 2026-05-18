---
title: Substrate Completion (Batch G) — make the substrate land in real files
date: 2026-05-11
status: SHIPPED 2026-05-11 (8 commits — see .spec/journal/2026-05-11-substrate-batch-g-shipped.md)
parent: open-items.md (Sections I, II, III.1)
purpose: >
  Close the gap between "Phases A–F shipped" and "first real Phase D + E + F
  end-to-end run is possible." Four items, all small-to-medium effort.
  After Batch G ships, the substrate's built primitives connect into a
  working loop with real outputs.
ratifications:
  - 2026-05-11 architecture nuance from Michael: file outputs are an
    opt-in deliverable per work_item, not a default. DB-default
    everywhere; explicit human gesture to materialize. Generalized
    mechanism that any pipeline can use.
  - D-G1 — Pipeline file_destination_template is UI suggestion only,
    NOT enforced default. Human chooses per work_item.
  - D-G2 — Materialization triggered by explicit "Materialize now"
    button on WorkItemDetail. Auto-fire deferred to v2 if the
    explicit-gesture pattern proves trustworthy.
  - D-G3 — Lessons + resolutions keep their separate promoted_to
    columns + per-item buttons. The new mechanism handles work_item-
    level file outputs (a different unit of decision).
---

# Batch G — make the substrate land in real files

## I. Binding problem

Phases A–F shipped six layers of the agentic creation cycle (Watch→Diagnose→Act→Account, maturity ladder, intent/covenant, post-completion ritual, trust ladder, multi-agent council). All six work in isolation; all six have been smoke-verified. But the loop between them isn't complete:

- Studies that reach `verified` can't promote to `study/` files (`file_path NOT NULL` blocks)
- Steward retries don't pull ratified lessons into prompt context (built but uncalled)
- Quarantined work_items don't fire Atonement (built but uncalled)
- Lessons + resolutions that get "Approve & promote → .mind/X.md" set the DB column but no actual file write happens

Each item is small. Together they're the difference between a substrate that demos its components and a substrate that produces things.

## II. Success criteria

1. **First real Phase D + E + F end-to-end run is possible** without manual intervention. A study-write work_item runs through all stages → reaches verified → fires sabbath → reflection lands → promotes to `study/<slug>.md` on the filesystem.
2. **Steward retries see prior lessons.** When the steward re-dispatches a failed work_item, the retry prompt includes "Recent lessons from this pipeline + stage" — the line-upon-line discipline actually fires.
3. **Quarantine fires Atonement.** When the steward quarantines a work_item on an `atonement_enabled` pipeline, the atonement dispatch fires automatically + ratified lessons land in `stewards.lessons`.
4. **File-write mechanism exists.** Lessons promoted to `.mind/principles.md` or `.mind/decisions.md` actually appear there on disk after the next git commit (or manual materializer run).

## III. Constraints and boundaries

**In scope:**
- Pre-existing `studies.file_path NOT NULL` fix
- Two two-line edits in `4c-steward-dispatch.sql` (retry composer switch + maybe_enqueue_atonement call)
- New `stewards.pending_file_writes` table + `stewards-cli materialize-writes` CLI command
- Smoke + first real e2e run through Phase D's Sabbath + Phase E's trust increment

**Out of scope:**
- New pipelines (covered by `substrate-pipelines-expansion.md`)
- UI for intent/covenant authoring (covered by `stewards-ui-evolution.md`)
- Soak observation week (Batch K — followup to this batch)
- File writes via plpython3u or host-mount sidecar — pending-write pattern is explicit choice per phase-d-design.md V.6 and phase-f-design.md V.4

## IV. Prior art

- **`stewards.lessons.promoted_to`** column (Phase D) — already populated by UI ratify-and-promote actions. Today it's a write-only audit field.
- **`stewards.resolutions.promoted_to`** column (Phase F) — same shape, same gap.
- **substrate's existing live-migration pattern** — docker cp + psql -f for SQL changes — works fine for the steward_dispatch edit.
- **Phase D's `maybe_enqueue_atonement(work_item_id)`** helper — already exists; nothing calls it.
- **Phase E's `retry_guidance_with_lessons(diagnosis, attempt, pipeline, stage)`** — already exists; steward still calls `retry_guidance` instead.

## V. Proposed approach

### V.1 G.1 — Fix `studies.file_path NOT NULL`

Inspect the existing studies table:
```sql
\d stewards.studies
```

Two options, pick whichever matches the substrate's design intent:
- **Option A (recommended):** make `file_path` nullable. Substrate-promoted studies don't have a canonical file path until the file-write mechanism (G.4) materializes them. NULL means "exists in DB, no on-disk file yet."
- **Option B:** populate `file_path` from slug at insert time: `'substrate--' || slug || '.md'`. Mirrors current `work_item_promote_to_study` slug pattern.

**Recommendation:** A. Once G.4 lands, the materializer sets `file_path` to the actual on-disk path it wrote to. Until then, NULL is honest.

```sql
ALTER TABLE stewards.studies ALTER COLUMN file_path DROP NOT NULL;
```

Add a comment explaining the semantics.

### V.2 G.2 — Steward retry pulls lessons

Edit `extension/4c-steward-dispatch.sql` (or the latest steward dispatch file):

```sql
-- Before:
v_retry_guidance := stewards.retry_guidance(v_diagnosis, v_attempt);

-- After:
v_retry_guidance := stewards.retry_guidance_with_lessons(
    v_diagnosis, v_attempt,
    v_wi.pipeline_family, v_wi.current_stage);
```

Live-apply via docker cp + psql -f. No restart needed.

**Smoke:** synthetic failed work_item on study-write/outline with ≥1 ratified lesson in `lessons_recent_ratified` for that pipeline+stage. Trigger steward_tick. Verify the dispatched chat's user message includes "Recent lessons from this pipeline + stage:" block.

### V.3 G.3 — Quarantine fires Atonement

Edit the same `4c-steward-dispatch.sql` (the quarantine path):

```sql
-- At the quarantine point, after UPDATE work_items SET quarantined_at = now():
PERFORM stewards.maybe_enqueue_atonement(v_work_item_id);
```

`maybe_enqueue_atonement` is a no-op if `pipelines.atonement_enabled = false`, so this is safe to add unconditionally.

**Smoke:** synthetic work_item with `failure_count = 3` on a pipeline with `atonement_enabled = true`. Trigger steward_tick. Verify (a) work_item gets quarantined, (b) atonement_dispatch chat enqueued, (c) on completion stewards.lessons populates with kind in (principle | decision | lesson).

### V.4 G.4 — File-write mechanism

**Schema:**

```sql
CREATE TABLE stewards.pending_file_writes (
    id              bigserial PRIMARY KEY,
    requested_at    timestamptz NOT NULL DEFAULT now(),
    requested_by    text NOT NULL,        -- 'lesson_promote' | 'council_resolve' | 'manual'
    target_path     text NOT NULL,        -- '.mind/principles.md' | 'study/<slug>.md' | etc.
    write_mode      text NOT NULL CHECK (write_mode IN ('append', 'create')),
    content         text NOT NULL,
    source_id       text,                 -- lesson_id | resolution_id | NULL
    source_kind     text,                 -- 'lesson' | 'resolution' | NULL
    materialized_at timestamptz,
    materialized_by text                  -- 'cli' | 'pre-commit-hook' | etc.
);

CREATE INDEX pending_file_writes_unmaterialized
    ON stewards.pending_file_writes (requested_at)
    WHERE materialized_at IS NULL;
```

**Producer hooks** (existing functions get small additions):

- `apply_lesson_ratify` (when promoted_to is set): INSERT into pending_file_writes with target_path = promoted_to, write_mode='append', content = "[YYYY-MM-DD] " + lesson.content
- `resolve_council` (when destination = 'study' or 'decisions'): INSERT with appropriate target_path + content

**Consumer CLI:**

```bash
stewards-cli materialize-writes [--dry-run] [--limit N]
```

Lives in `projects/pg-ai-stewards/cmd/stewards-cli/`. Reads unmaterialized rows, performs the file write (append or create), UPDATEs `materialized_at` + `materialized_by='cli'`. Idempotent — already-materialized rows are skipped.

**Optional pre-commit hook integration:** the existing `scripts/git-hooks/pre-commit` (from Phase C.3) gains a section that calls `stewards-cli materialize-writes` if the substrate container is running. Skip gracefully if down (same pattern as the YAML seed).

**Why pending-write + CLI rather than direct file write from pg:**
- substrate stays FS-stateless — no plpython3u dep, no host-mount required
- materialization is reviewable — `stewards-cli materialize-writes --dry-run` shows what would change
- pre-commit hook integration means file writes land naturally before each commit; manual run is escape hatch

## VI. Open questions / follow-ups

- **Conflict resolution for append mode.** If two lessons get promoted to `.mind/principles.md` between commits, both append in order of `requested_at`. Should there be a section header per entry? Recommend: each entry gets a "## YYYY-MM-DD: <slug>" header so the file stays browsable.
- **`promoted_to` semantics.** Today the DB column says "where it WILL be written." After G.4 ships, should the column mean "where it WAS written" (set by materializer) or stay as the request? Recommend: keep as request; add `pending_file_writes.id` FK so the audit trail is complete.
- **Failure mode.** What if materialize-writes fails (file permissions, disk full, malformed path)? Recommend: log to materialized_by='error:<reason>', leave materialized_at NULL, surface in next run.
- **Study promotion via this path?** Phase D's `work_item_promote_to_study` writes to `stewards.studies` directly (in-DB). The on-disk study/<slug>.md file is separate. Should the promote flow also queue a pending file write to actually create the .md file? Recommend: yes — that's where G.1's "file_path NULL until materialized" pattern pays off.

## VII. Estimated programming time

- G.1 file_path fix: 10 min + smoke
- G.2 retry pulls lessons: 10 min + smoke
- G.3 quarantine fires atonement: 10 min + smoke
- G.4 file-write mechanism: 1 session (schema + producer hooks + Go CLI + pre-commit hook integration)
- First real e2e run through Sabbath + trust increment: 30 min (cost ~$0.05)

**Total: ~1 session** for the full batch.

## VIII. Acceptance scenarios

1. Synthetic study-write work_item runs outline → draft → review → verify with all_passed=true. Sabbath fires (Phase D). Reflection lands in stewards.lessons. work_items.sabbath_completed_at timestamped. work_item_promote_to_study succeeds (G.1 + sabbath gate passed). New row in stewards.studies. New pending_file_write row queued for `study/<slug>.md`. `stewards-cli materialize-writes` writes the file. Trust counter for (agent, pipeline, model) increments.

2. Synthetic failed work_item triggers steward retry. Retry prompt includes "Recent lessons from this pipeline + stage:" with 1-3 ratified lessons (G.2).

3. Synthetic 3-failure quarantine on atonement_enabled pipeline. Quarantine fires atonement_dispatch (G.3). Apply produces 3 unratified lessons.

4. Human clicks "Approve & promote → .mind/principles.md" on a lesson. Pending file write queued. Pre-commit hook runs on next commit. `.mind/principles.md` appends the lesson content. lesson row's pending_file_write FK populated.

## IX. Why this is a batch, not a phase

Phases A–F were architectural builds (new substrate primitives). Batch G is *plumbing* — connecting primitives already built. No new ratification required; all decisions in the sub-specs were made.

Naming convention: phase letters were exhausted with F. Batch G uses the next letter to mark the BUILD→USE transition without claiming a new architectural layer.
