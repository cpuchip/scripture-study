-- =====================================================================
-- Phase 2.7b.4 — Watchman soak prep
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: NINTH file).
--
-- Builds on:
--   - 2.7a (dirty_queue view, verdicts, findings)
--   - 2.7b.1 (watchman_passes, completion trigger)
--   - 2.7b.2 (watchman_should_fire, scheduler tick)
--   - 2.7b.3 (token budget, estimate_chat_tokens)
--   - 2.6a/b/c (workstreams, todos, context_for, AGE graph)
--
-- This file adds:
--   1. Updated stewards.dirty_queue view — excludes docs whose
--      frontmatter has `watchman: skip` (or `exempt`). Implements
--      option 3 from the 2026-05-06 design conversation; closes
--      todo `watchman-frontmatter-exempt`.
--   2. stewards.regenerate_active_md() — generates a markdown
--      status report from current substrate state. Renders:
--        - In Flight (workstreams + their declared proposals)
--        - Open Findings (unacknowledged drift)
--        - Open Todos (status in {open,in_progress})
--        - Recent Watchman Activity (last 5 passes)
--        - Corpus Stats
--      Returns text. Does not write to disk; CLI is the surface.
--
-- The 7-day soak (the third 2.7b.4 deliverable) is runtime
-- observation, not code — start it by setting schedule_enabled=true
-- after this migration applies.
-- =====================================================================

-- ---------------------------------------------------------------------
-- dirty_queue: add watchman:skip frontmatter exemption.
--
-- New gate: docs whose frontmatter has `watchman: skip` (or `exempt`)
-- are excluded. The frontmatter jsonb column already exists with a
-- GIN index (Phase 2.1); zero schema change. Users add `watchman: skip`
-- to YAML frontmatter and re-import.
--
-- Existing gates preserved:
--   - dirty-bit (touched since last consolidated)
--   - open drift finding suppression
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.dirty_queue AS
SELECT s.id,
       s.slug,
       s.kind,
       s.title,
       s.updated_at,
       s.last_consolidated_at,
       (s.updated_at - coalesce(s.last_consolidated_at,
                                 'epoch'::timestamptz)) AS dirty_for
  FROM stewards.studies s
 WHERE (s.last_consolidated_at IS NULL
        OR s.updated_at > s.last_consolidated_at)
   AND coalesce(lower(s.frontmatter->>'watchman'), '')
       NOT IN ('skip', 'exempt')
   AND NOT EXISTS (
       SELECT 1 FROM stewards.findings f
        WHERE f.study_id = s.id
          AND f.kind = 'drift'
          AND f.acknowledged_at IS NULL
   )
 ORDER BY coalesce(s.last_consolidated_at, 'epoch'::timestamptz),
          s.updated_at;

COMMENT ON VIEW stewards.dirty_queue IS
'Phase 2.7b.4: docs that need (re-)consolidation. Three gates: dirty-bit (touched since last consolidated), no open drift finding (surface-once-stop), and frontmatter `watchman` is not "skip"/"exempt". Add `watchman: skip` to YAML to opt a doc out (e.g., journal entries that are point-in-time snapshots).';

-- ---------------------------------------------------------------------
-- regenerate_active_md() — markdown status report.
--
-- Generates the dynamic sections of an `active.md`-style document
-- from current substrate state. Does NOT cover human-curated parts
-- (Priorities list, Key Facts) — those stay in the hand-written file.
--
-- Sections:
--   ## In Flight        — workstreams + their declared proposals
--   ## Open Findings    — unacknowledged drift, severity-sorted
--   ## Open Todos       — open + in_progress, parent-grouped
--   ## Recent Watchman  — last 5 passes with verdict counts
--   ## Corpus Stats     — kind counts + dirty queue size
--
-- Returns text (markdown). Caller pipes to file if desired. Future
-- automation may write at end of each Watchman pass; for now it's
-- on-demand via CLI.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.regenerate_active_md()
RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_md          text := '';
    v_now         text := to_char(now() AT TIME ZONE 'UTC',
                                  'YYYY-MM-DD HH24:MI:SS"Z"');
    v_section     text;
    v_dirty_count int;
BEGIN
    -- Header
    v_md := v_md || format(
        E'# Active Context (generated)\n\n_Generated %s by stewards.regenerate_active_md()_\n\n',
        v_now);
    v_md := v_md || E'> This file is regenerated from substrate state. Human-curated\n'
                 || E'> sections (Priorities, Key Facts) live in `.mind/active.md` and are\n'
                 || E'> not produced here.\n\n';

    -- ----- In Flight -----
    v_md := v_md || E'## In Flight\n\n';
    SELECT string_agg(block, E'\n')
      INTO v_section
      FROM (
          SELECT format(
                     E'### %s — %s\n\n%s\n',
                     w.id,
                     coalesce(w.name, '(unnamed)'),
                     coalesce(
                         (SELECT string_agg(
                                     format('- %s **%s** — %s',
                                            CASE WHEN s.kind = 'proposal' THEN '📝'
                                                 WHEN s.kind = 'phase-doc' THEN '🔨'
                                                 ELSE '📄' END,
                                            coalesce(s.title, s.slug),
                                            s.slug),
                                     E'\n'
                                     ORDER BY s.title)
                            FROM stewards.studies s
                           WHERE s.frontmatter->>'workstream' = w.id),
                         '_(no declared proposals)_'
                     )
                 ) AS block
            FROM stewards.workstreams w
           WHERE coalesce(w.status, 'active') = 'active'
           ORDER BY w.id
      ) sub;
    v_md := v_md || coalesce(v_section, '_No active workstreams._') || E'\n\n';

    -- ----- Open Findings -----
    v_md := v_md || E'## Open Findings\n\n';
    SELECT string_agg(line, E'\n')
      INTO v_section
      FROM (
          SELECT format(
                     E'- **%s** [%s/%s] (`%s`)\n  %s%s',
                     coalesce(s.title, s.slug),
                     f.kind,
                     f.severity,
                     s.slug,
                     replace(coalesce(f.message, '(no message)'),
                             E'\n', E'\n  '),
                     CASE
                         WHEN f.suggested_action IS NOT NULL
                         THEN E'\n  → ' || replace(f.suggested_action,
                                                    E'\n', E'\n    ')
                         ELSE ''
                     END
                 ) AS line
            FROM stewards.findings f
            JOIN stewards.studies s ON s.id = f.study_id
           WHERE f.acknowledged_at IS NULL
           ORDER BY array_position(ARRAY['high','medium','low'], f.severity),
                    f.created_at DESC
      ) sub;
    v_md := v_md || coalesce(v_section, '_No open findings._') || E'\n\n';

    -- ----- Open Todos -----
    v_md := v_md || E'## Open Todos\n\n';
    SELECT string_agg(line, E'\n')
      INTO v_section
      FROM (
          SELECT format(
                     '- [%s] **%s** — %s (under `%s/%s`)',
                     CASE t.status WHEN 'in_progress' THEN '▶' ELSE ' ' END,
                     coalesce(t.slug, substring(t.id::text FROM 1 FOR 8)),
                     t.title,
                     t.parent_kind,
                     t.parent_slug
                 ) AS line
            FROM stewards.todos t
           WHERE t.status IN ('open', 'in_progress')
           ORDER BY t.parent_kind, t.parent_slug, t.created_at
      ) sub;
    v_md := v_md || coalesce(v_section, '_No open todos._') || E'\n\n';

    -- ----- Recent Watchman Activity -----
    v_md := v_md || E'## Recent Watchman Activity\n\n';
    SELECT string_agg(line, E'\n')
      INTO v_section
      FROM (
          SELECT format(
                     '- `%s` — %s, %s docs, %s verdicts',
                     pass_id,
                     to_char(started_at AT TIME ZONE 'UTC',
                             'YYYY-MM-DD HH24:MI"Z"'),
                     doc_count_done,
                     coalesce(verdict_counts::text, '{}')
                 ) AS line
            FROM stewards.watchman_passes
           ORDER BY started_at DESC
           LIMIT 5
      ) sub;
    v_md := v_md || coalesce(v_section, '_No passes recorded yet._') || E'\n\n';

    -- ----- Corpus Stats -----
    v_md := v_md || E'## Corpus Stats\n\n';
    v_md := v_md || E'| Kind | Total | Embedded | In dirty_queue |\n';
    v_md := v_md || E'|------|------:|---------:|---------------:|\n';
    SELECT string_agg(line, E'\n')
      INTO v_section
      FROM (
          SELECT format(
                     '| %s | %s | %s | %s |',
                     s.kind,
                     count(*),
                     count(s.embedding),
                     count(*) FILTER (
                         WHERE (s.last_consolidated_at IS NULL
                                OR s.updated_at > s.last_consolidated_at)
                           AND coalesce(lower(s.frontmatter->>'watchman'), '')
                               NOT IN ('skip', 'exempt')
                           AND NOT EXISTS (
                               SELECT 1 FROM stewards.findings f
                                WHERE f.study_id = s.id
                                  AND f.kind = 'drift'
                                  AND f.acknowledged_at IS NULL)
                     )
                 ) AS line
            FROM stewards.studies s
           GROUP BY s.kind
           ORDER BY s.kind
      ) sub;
    v_md := v_md || coalesce(v_section, '| _no docs_ | 0 | 0 | 0 |') || E'\n\n';

    -- Total dirty (cross-reference for sanity)
    SELECT count(*) INTO v_dirty_count FROM stewards.dirty_queue;
    v_md := v_md || format(
        E'_Total dirty queue: %s_\n', v_dirty_count);

    RETURN v_md;
END;
$func$;

COMMENT ON FUNCTION stewards.regenerate_active_md() IS
'Phase 2.7b.4: generate a markdown status report from current substrate state. Sections: In Flight, Open Findings, Open Todos, Recent Watchman Activity, Corpus Stats. Returns text — caller decides what to do with it (CLI prints; future automation may write to .mind/active.md at end of Watchman pass).';
