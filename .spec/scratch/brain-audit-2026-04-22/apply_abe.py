"""
Phase A + B + E executor for brain-audit-2026-04-22.
Idempotent. Wrapped in single transaction. Prints diff before commit.
Run with --dry to preview, --commit to apply.
"""
import sqlite3, sys, datetime

DRY = '--commit' not in sys.argv
DB = 'private-brain/brain.db'

# ============ PHASE A: mechanical closure ============
# action_done=1 → status=done (or archived for personal items already in inbox)
PHASE_A_DONE = [
    # 2nd Brain bug reports / features
    '190f9ad5', '4ed38e7b', '4ce75e3d', '17729990',
    '0d756be3', '17730350', 'c0796e7c',
    # Personal/notebook
    '17731960',  # birthday present for mom
]

# Personal items already action_done=1 → confirm status=archived (no-op if already)
PHASE_A_ARCHIVE_PERSONAL = [
    'b22c77c9',  # Keda sync with ben — work item, done
]

# Verified inbox entries that say "Shipped" — set project + status=done
# (proposal_path stays in body; we'll add proposal_path column later in harness phase)
PHASE_A_SHIPPED_INBOX = [
    # (entry_id, project_id, project_name)
    ('22e0426d', 3, 'Workspace improvements'),  # Session Journal Tool
    ('88cf9fb9', 5, 'ibeco.me'),                 # Becoming App Phases 1-3
    ('1038828b', 5, 'ibeco.me'),                 # Desktop Notifications
    ('a86ebb14', 6, '2nd Brain'),                # TITSW Enrichment
    ('732b63de', 3, 'Workspace improvements'),  # Gospel Engine FTS+Vector
    ('537edd66', 6, '2nd Brain'),                # Multi-Agent Routing
    ('01fd99fd', 6, '2nd Brain'),                # Brain Relay via ibeco.me
    ('a632a729', 6, '2nd Brain'),                # Brain Project-Kanban
]

# Space Center LCARS theme — already status=done, just verify
PHASE_A_VERIFY_DONE = ['03549801']

# ============ PHASE B: inbox → project + status=active ============
# (entry_id, project_id, status, note)
PHASE_B_RELOCATE = [
    # Specced inbox → active proposals
    ('d6c00fb2', 3, 'active',   'Claude Code Integration → claude-code-integration.md'),
    ('6988583c', 6, 'active',   'Brain Windows Service → brain-windows-service.md'),
    ('f1f8dc89', 3, 'active',   'Gospel Graph → gospel-graph/main.md'),
    ('989e9375', 6, 'active',   'Data Safety → data-safety/main.md'),
    ('4b1df2ca', 6, 'active',   'Classifier Qwen Fix → classifier-qwen-fix.md'),
    ('c68b4a64', 3, 'active',   'Sabbath Agent → sabbath-agent.md'),
    # Specced inbox → proposals we archived in cleanup-2026-04-part2
    ('ed09792d', 6, 'archived', 'brain-ibecome-layer2 archived in cleanup-2026-04-part2'),
    ('cd6ccf77', 3, 'archived', 'embedding-comparison archived in cleanup-2026-04-part2'),
    ('36b54a69', 3, 'archived', 'brain-workspace-aware archived in cleanup-2026-04-part2'),
    ('2318a15b', 3, 'archived', 'memory-architecture superseded by .mind/ adoption'),
    # Planned inbox → active
    ('0cde48e2', 3, 'active',   'Gospel Engine 1.5 → gospel-engine/phase1.5-ergonomics.md'),
    ('b0257582', 7, 'active',   'Teaching Workstream → teaching-workstream.md'),
    ('ff51406e', 1, 'active',   'Study Workstream → study-workstream.md'),
    ('12fa3f78', 6, 'active',   'Classification Quality Benchmark → classify-bench.md'),
]

# ============ PHASE E: Personal project + move ============
PHASE_E_CREATE_PROJECT = ('Personal', '🏠', 'active',
                          'Personal life: family, health, errands, home')
PHASE_E_MOVE_TO_PERSONAL = [
    'acef973d',  # Custom Desk Project
    '17753153',  # Reflect back what wife says
    '17753148',  # KISS principle reminder
    '17729329',  # Grocery shopping list (active)
    '1597e29f',  # Grocery (milk + hotdogs)
    '17729138',  # Schedule Temple Visit with Kurt
    '0ede8a82',  # Build PC hardware assembly  (will check actual id)
    '0ede0bf2',  # placeholder
]

# Better: query the inbox raw entries that look personal
PERSONAL_TITLES = [
    'Custom Desk Project',
    'Reflect back what wife says',
    'KISS principle reminder from comment',
    'Grocery shopping list',
    'Grocery Shopping List',
    'Schedule Temple Visit with Kurt',
    'Build PC hardware assembly',
    'Uncertainty about AI',  # this one's mixed — leave in inbox
]


def main():
    conn = sqlite3.connect(DB)
    cur = conn.cursor()
    cur.execute('BEGIN')
    now = datetime.datetime.utcnow().isoformat() + 'Z'

    changes = []

    # PHASE A: mark done
    for eid in PHASE_A_DONE:
        cur.execute("UPDATE entries SET status='done', updated_at=? WHERE id LIKE ? AND (status IS NULL OR status='active')",
                    (now, eid + '%'))
        changes.append(('A.done', eid, cur.rowcount))

    for eid in PHASE_A_ARCHIVE_PERSONAL:
        cur.execute("UPDATE entries SET status='archived', updated_at=? WHERE id LIKE ? AND status != 'archived'",
                    (now, eid + '%'))
        changes.append(('A.archive_personal', eid, cur.rowcount))

    for eid, pid, pname in PHASE_A_SHIPPED_INBOX:
        cur.execute("UPDATE entries SET project_id=?, status='done', updated_at=? WHERE id LIKE ?",
                    (pid, now, eid + '%'))
        changes.append((f'A.shipped→{pname}', eid, cur.rowcount))

    # PHASE B: relocate + status
    for eid, pid, status, note in PHASE_B_RELOCATE:
        cur.execute("UPDATE entries SET project_id=?, status=?, updated_at=? WHERE id LIKE ?",
                    (pid, status, now, eid + '%'))
        changes.append((f'B.{status}→pid{pid}', eid, cur.rowcount))

    # PHASE E: create Personal project
    cur.execute("SELECT id FROM projects WHERE name='Personal'")
    row = cur.fetchone()
    if row:
        personal_id = row[0]
        changes.append(('E.project_exists', f'Personal=id{personal_id}', 0))
    else:
        cur.execute("""INSERT INTO projects (name, emoji, status, description, created_at, updated_at)
                       VALUES (?, ?, ?, ?, ?, ?)""",
                    (PHASE_E_CREATE_PROJECT[0], PHASE_E_CREATE_PROJECT[1],
                     PHASE_E_CREATE_PROJECT[2], PHASE_E_CREATE_PROJECT[3], now, now))
        personal_id = cur.lastrowid
        changes.append(('E.project_created', f'Personal=id{personal_id}', 1))

    # PHASE E: move personal entries by title match (safer than guessed IDs)
    for title in PERSONAL_TITLES:
        if title.startswith('Uncertainty'):
            continue  # explicitly leave in inbox
        cur.execute("UPDATE entries SET project_id=?, updated_at=? WHERE project_id IS NULL AND title=?",
                    (personal_id, now, title))
        changes.append(('E.move', title, cur.rowcount))

    # Print diff
    print(f"{'Action':<32} {'Target':<50} {'Rows'}")
    print('-' * 90)
    total = 0
    for action, target, rows in changes:
        print(f"{action:<32} {str(target)[:49]:<50} {rows}")
        total += rows
    print('-' * 90)
    print(f'TOTAL row changes: {total}')

    if DRY:
        conn.rollback()
        print('\n*** DRY RUN — rolled back. Run with --commit to apply. ***')
    else:
        conn.commit()
        print('\n*** COMMITTED ***')

    conn.close()

if __name__ == '__main__':
    main()
