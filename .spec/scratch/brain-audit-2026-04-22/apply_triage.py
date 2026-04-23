"""
Apply 2026-04-22 triage decisions to brain.db.

Each tuple: (id_prefix, action, optional new_project_id, optional proposal_path)
- action in: done, archived, someday, keep-active, move-to-notebook, promote
- For 'promote': proposal_path links the entry to the proposal we created.

Run with --commit to apply. Default is dry-run preview.
"""
import sqlite3, sys

PROMOTES = {
    # entry_id_prefix -> proposal_path (workspace-relative)
    '17735267': '.spec/proposals/vscode-agent-hooks-integration.md',
    '17742746': '.spec/proposals/study-nate-jones-delegation.md',
    '17767284': '.spec/proposals/gospel-engine-v3-proxy-pointer.md',
    '17762740': '.spec/proposals/johari-window-agent-mode.md',
    '17752818': '.spec/proposals/motivation-coach-agent-mode.md',
    'c168be5c': '.spec/proposals/ai-presentation-site-tool.md',
    '17751138': '.spec/proposals/launch-youtube-channel.md',
    '17751939': '.spec/proposals/lightrag-investigation.md',
    '17751941': '.spec/proposals/lightrag-investigation.md',  # merged into same proposal
    'a4eae47c': '.spec/proposals/memory-research-bundle.md',
    '17755019': '.spec/proposals/memory-research-bundle.md',
    '17756547': '.spec/proposals/memory-research-bundle.md',
}

# (id_prefix, action) — action precedence: filled-in Action overrides recommendation
ACTIONS = [
    # ============ verified (23) ============
    ('17752826', 'someday'),
    ('17748767', 'archived'),
    ('a249be5d', 'archived'),
    ('17753205', 'archived'),
    ('17752229', 'archived'),
    ('17735051', 'someday'),
    ('5d7d781a', 'done'),
    ('17753587', 'someday'),
    ('17752218', 'archived'),
    ('17753137', 'done'),
    ('17733627', 'someday'),
    ('17743563', 'someday'),
    ('17746750', 'done'),
    ('17744822', 'someday'),
    ('17743715', 'archived'),
    ('17735267', 'promote'),
    ('17739741', 'done'),
    ('17742746', 'promote'),
    ('17750525', 'someday'),
    ('17740194', 'done'),
    ('17739252', 'someday'),
    ('22b8d8b2', 'keep-active'),
    ('20ce036a', 'keep-active'),
    # ============ planned (1) ============
    ('17729989', 'keep-active'),
    # ============ raw (45) ============
    ('17767284', 'promote'),
    ('17763075', 'someday'),
    ('17763948', 'done'),
    ('17762740', 'promote'),
    ('17749663', 'someday'),
    ('17766038', 'someday'),
    ('17765208', 'someday'),
    ('17729953', 'archived'),
    ('82e58844', 'someday'),
    ('1348874b', 'keep-active'),
    ('17755019', 'promote'),
    ('e8bdac42', 'keep-active'),
    ('17756547', 'promote'),
    ('17752818', 'promote'),
    ('486827f9', 'archived'),
    ('17764487', 'someday'),
    ('de10491b', 'someday'),
    ('ac28114d', 'keep-active'),
    ('17729928', 'archived'),
    ('17752798', 'someday'),
    ('17752858', 'someday'),
    ('a4eae47c', 'promote'),
    ('17754148', 'someday'),
    ('17746749', 'someday'),
    ('17751941', 'promote'),
    ('10839464', 'archived'),
    ('17750573', 'archived'),
    ('17751939', 'promote'),
    ('17750225', 'someday'),
    ('17761796', 'someday'),
    ('9262c7be', 'archived'),  # user typed "archive" - same intent
    ('17740452', 'someday'),
    ('17743564', 'archived'),
    ('17752224', 'archived'),
    ('17743562', 'archived'),
    ('17744148', 'someday'),
    ('17730988', 'someday'),
    ('17739561', 'done'),
    ('c168be5c', 'promote'),
    ('17751138', 'promote'),
    ('17735945', 'done'),
    ('17730291', 'someday'),
    ('17752820', 'keep-active'),
    ('17730308', 'move-to-notebook'),
    ('17749618', 'someday'),
]


def expand_id(cur, prefix):
    rows = cur.execute("SELECT id FROM entries WHERE id LIKE ?", (prefix + '%',)).fetchall()
    if len(rows) != 1:
        raise RuntimeError(f"Prefix {prefix} matched {len(rows)} entries")
    return rows[0][0]


def main(commit):
    c = sqlite3.connect('private-brain/brain.db')
    cur = c.cursor()
    counts = {'done': 0, 'archived': 0, 'someday': 0, 'keep-active': 0,
              'move-to-notebook': 0, 'promote': 0}
    misses = []

    for prefix, action in ACTIONS:
        try:
            full_id = expand_id(cur, prefix)
        except RuntimeError as e:
            misses.append(str(e))
            continue

        if action == 'done':
            cur.execute("UPDATE entries SET status='done', action_done=1 WHERE id=?", (full_id,))
        elif action == 'archived':
            cur.execute("UPDATE entries SET status='archived' WHERE id=?", (full_id,))
        elif action == 'someday':
            cur.execute("UPDATE entries SET status='someday' WHERE id=?", (full_id,))
        elif action == 'keep-active':
            cur.execute("UPDATE entries SET status='active' WHERE id=? AND (status IS NULL OR status='')", (full_id,))
        elif action == 'move-to-notebook':
            cur.execute("UPDATE entries SET project_id=9, status='active' WHERE id=?", (full_id,))
        elif action == 'promote':
            ppath = PROMOTES.get(prefix)
            if not ppath:
                misses.append(f"promote {prefix}: no proposal_path mapping")
                continue
            cur.execute("UPDATE entries SET status='active', proposal_path=? WHERE id=?",
                        (ppath, full_id))
        counts[action] = counts.get(action, 0) + 1

    print(f"--- Triage application {'COMMIT' if commit else 'DRY-RUN'} ---")
    for k, v in sorted(counts.items()):
        print(f"  {k}: {v}")
    print(f"  TOTAL: {sum(counts.values())} (expected 69)")
    if misses:
        print("\nWARNINGS:")
        for m in misses:
            print(f"  - {m}")

    if commit:
        c.commit()
        print("\nCOMMITTED")
    else:
        c.rollback()
        print("\nROLLED BACK (use --commit to apply)")
    c.close()


if __name__ == '__main__':
    main('--commit' in sys.argv)
