"""
Harness Phase 1 — schema migration.
Adds `workstream` and `proposal_path` columns to entries.
Backfills workstream from project_id default mapping.
Backfills proposal_path by scraping body for `.spec/proposals/...` references.

Idempotent: re-runs check column existence before ALTER.
Run with --commit to apply.
"""
import sqlite3, sys, re, datetime

DRY = '--commit' not in sys.argv
DB = 'private-brain/brain.db'

# Default project → workstream mapping.
# Mixed-workstream projects (3, 6) get a default; per-entry tuning happens in Phase C/D walks.
PROJECT_TO_WS_DEFAULT = {
    1: 'WS6',  # study → Studies
    2: 'WS8',  # Sunday School
    3: 'WS5',  # Workspace improvements (default; WS3 for gospel items, WS1 for brain core items)
    4: 'WS9',  # Space Center → Other apps
    5: 'WS4',  # ibeco.me → study.ibeco.me
    6: 'WS2',  # 2nd Brain (default; WS1 for pipeline/classifier items)
    7: 'WS7',  # YouTube / Content → Teaching
    8: 'WS9',  # Budget App → Other apps
    9: 'WS9',  # Notebook → Other apps
    10: 'WS9', # cpuchip.net → Other apps
    11: None,  # Personal → not workstream-tracked
}

# Per-entry workstream overrides where the default project mapping is wrong.
# Format: { entry_id_prefix: 'WSn' }
ENTRY_WS_OVERRIDES = {
    # Project 6 entries that are pipeline/classifier (WS1) not UX (WS2)
    'a86ebb14': 'WS1',  # TITSW Enrichment
    '537edd66': 'WS1',  # Multi-Agent Routing + Classification
    '989e9375': 'WS1',  # Data Safety
    '4b1df2ca': 'WS1',  # Classifier Qwen Fix
    '12fa3f78': 'WS1',  # Classification Quality Benchmark
    # Project 3 entries that are gospel-engine (WS3) not memory/process (WS5)
    '732b63de': 'WS3',  # Gospel Engine FTS+Vector
    'f1f8dc89': 'WS3',  # Gospel Graph
    '0cde48e2': 'WS3',  # Gospel Engine 1.5
    # Project 3 entries that are infra/auth for ibeco.me (WS4)
    '17733627': 'WS4',  # Implement Gateway Auth (this is in ibeco.me project actually, leave it)
}

PROPOSAL_PATH_RE = re.compile(r'\.spec[/\\]proposals[/\\]([a-zA-Z0-9._/-]+\.md)')


def has_column(cur, table, col):
    return any(r[1] == col for r in cur.execute(f"PRAGMA table_info({table})"))


def main():
    conn = sqlite3.connect(DB)
    cur = conn.cursor()
    cur.execute('BEGIN')

    # 1. Add columns if missing
    if not has_column(cur, 'entries', 'workstream'):
        cur.execute("ALTER TABLE entries ADD COLUMN workstream TEXT")
        print("ALTER: added entries.workstream")
    else:
        print("OK: entries.workstream already exists")

    if not has_column(cur, 'entries', 'proposal_path'):
        cur.execute("ALTER TABLE entries ADD COLUMN proposal_path TEXT")
        print("ALTER: added entries.proposal_path")
    else:
        print("OK: entries.proposal_path already exists")

    # 2. Backfill workstream from project_id (only where currently NULL)
    backfill_count = 0
    for pid, ws in PROJECT_TO_WS_DEFAULT.items():
        if ws is None:
            continue
        cur.execute("UPDATE entries SET workstream=? WHERE project_id=? AND workstream IS NULL",
                    (ws, pid))
        if cur.rowcount:
            print(f"  backfill project={pid} → ws={ws}: {cur.rowcount} rows")
            backfill_count += cur.rowcount

    # 3. Apply per-entry overrides
    override_count = 0
    for eid_prefix, ws in ENTRY_WS_OVERRIDES.items():
        cur.execute("UPDATE entries SET workstream=? WHERE id LIKE ?",
                    (ws, eid_prefix + '%'))
        if cur.rowcount:
            override_count += cur.rowcount
    print(f"  per-entry overrides applied: {override_count}")

    # 4. Backfill proposal_path by scraping body
    path_count = 0
    rows = cur.execute("SELECT id, body, one_liner FROM entries WHERE proposal_path IS NULL").fetchall()
    for eid, body, one_liner in rows:
        text = (body or '') + ' ' + (one_liner or '')
        m = PROPOSAL_PATH_RE.search(text)
        if m:
            path = m.group(0).replace('\\', '/')
            cur.execute("UPDATE entries SET proposal_path=? WHERE id=?", (path, eid))
            path_count += 1
    print(f"  proposal_path scraped: {path_count}")

    # 5. Summary
    print(f"\n=== Coverage ===")
    print(f"  entries with workstream:    {cur.execute('SELECT COUNT(*) FROM entries WHERE workstream IS NOT NULL').fetchone()[0]} / 115")
    print(f"  entries with proposal_path: {cur.execute('SELECT COUNT(*) FROM entries WHERE proposal_path IS NOT NULL').fetchone()[0]} / 115")
    print(f"  entries still NULL ws (inbox + Personal): {cur.execute('SELECT COUNT(*) FROM entries WHERE workstream IS NULL').fetchone()[0]}")

    print(f"\n=== Workstream distribution ===")
    for r in cur.execute("SELECT COALESCE(workstream,'<null>'), COUNT(*) FROM entries GROUP BY workstream ORDER BY 2 DESC"):
        print(f"  {r[0]:<8} {r[1]}")

    if DRY:
        conn.rollback()
        print('\n*** DRY RUN — rolled back. Run with --commit to apply. ***')
    else:
        conn.commit()
        print('\n*** COMMITTED ***')

    conn.close()


if __name__ == '__main__':
    main()
