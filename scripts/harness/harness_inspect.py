"""
Harness Phase 1 — sync inspector (read-only).
Lists workstream alignment between brain entries and proposal files.

Usage:
  python harness_inspect.py              # all workstreams
  python harness_inspect.py WS2          # single workstream
  python harness_inspect.py --gaps       # entries without proposal_path that look mature
  python harness_inspect.py --orphans    # proposals without any brain entry
"""
import sqlite3, sys, os, re
from pathlib import Path

DB = 'private-brain/brain.db'
PROPOSALS_DIR = Path('.spec/proposals')

WORKSTREAMS = ['WS1', 'WS2', 'WS3', 'WS4', 'WS5', 'WS6', 'WS7', 'WS8', 'WS9']
WS_NAMES = {
    'WS1': 'Brain Core', 'WS2': 'Brain UX', 'WS3': 'Gospel Engine',
    'WS4': 'study.ibeco.me', 'WS5': 'Memory & Process', 'WS6': 'Studies',
    'WS7': 'Teaching', 'WS8': 'Sunday School', 'WS9': 'Other apps',
}


def load_proposal_frontmatter():
    """Returns list of (path, frontmatter_dict) for every proposal with frontmatter."""
    results = []
    for md in list(PROPOSALS_DIR.glob('*.md')) + list(PROPOSALS_DIR.glob('*/*.md')):
        if 'archive' in md.parts:
            continue
        try:
            text = md.read_text(encoding='utf-8')
        except Exception:
            continue
        if not text.startswith('---'):
            continue
        end = text.find('---', 3)
        if end == -1:
            continue
        fm = {}
        for line in text[3:end].strip().splitlines():
            if ':' in line:
                k, v = line.split(':', 1)
                fm[k.strip()] = v.strip().strip('"').strip("'")
        rel = str(md.relative_to(Path('.'))).replace('\\', '/')
        fm['_path'] = rel
        results.append(fm)
    return results


def main():
    args = sys.argv[1:]
    only_ws = next((a for a in args if a.startswith('WS')), None)
    gaps_mode = '--gaps' in args
    orphans_mode = '--orphans' in args

    conn = sqlite3.connect(DB)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    proposals = load_proposal_frontmatter()
    proposals_by_ws = {}
    for p in proposals:
        ws = p.get('workstream', '?')
        proposals_by_ws.setdefault(ws, []).append(p)

    # GAPS mode: mature entries without proposal_path
    if gaps_mode:
        print("=== GAPS: mature brain entries without proposal_path ===")
        rows = cur.execute("""
            SELECT e.id, e.workstream, e.maturity, e.status, e.title, COALESCE(p.name,'<inbox>') as proj
            FROM entries e LEFT JOIN projects p ON p.id=e.project_id
            WHERE e.proposal_path IS NULL
              AND e.maturity IN ('verified','specced','planned','complete')
              AND (e.status IS NULL OR e.status NOT IN ('done','archived'))
            ORDER BY e.workstream, e.maturity DESC
        """).fetchall()
        for r in rows:
            print(f"  [{r['id'][:8]}] {str(r['workstream']):<6} {r['maturity']:<10} {r['proj']:<22} {r['title'][:60]}")
        print(f"\nTotal gaps: {len(rows)}")
        return

    # ORPHANS mode: proposals without any brain entry pointing to them
    if orphans_mode:
        print("=== ORPHANS: proposals with no linked brain entry ===")
        linked = {r[0] for r in cur.execute("SELECT DISTINCT proposal_path FROM entries WHERE proposal_path IS NOT NULL")}
        for p in proposals:
            path = p.get('_path', '')
            if not any(linked_p and linked_p in path for linked_p in linked):
                print(f"  {p.get('workstream','?'):<6} {p.get('status','?'):<10} {path}")
        return

    # ALIGNMENT mode: per-workstream side-by-side
    targets = [only_ws] if only_ws else WORKSTREAMS
    for ws in targets:
        ws_proposals = proposals_by_ws.get(ws, [])
        ws_entries = cur.execute("""
            SELECT e.id, e.maturity, e.status, e.action_done, e.title, e.proposal_path,
                   COALESCE(p.name,'<inbox>') as proj
            FROM entries e LEFT JOIN projects p ON p.id=e.project_id
            WHERE e.workstream=? ORDER BY e.maturity DESC, e.updated_at DESC
        """, (ws,)).fetchall()

        if not ws_proposals and not ws_entries:
            continue

        print(f"\n{'='*88}")
        print(f"  {ws} — {WS_NAMES[ws]}")
        print(f"{'='*88}")

        print(f"\n  PROPOSALS ({len(ws_proposals)}):")
        for p in sorted(ws_proposals, key=lambda x: (x.get('status',''), x.get('_path',''))):
            print(f"    [{p.get('status','?'):<10}] {p['_path']}")

        # Entries grouped by status
        print(f"\n  BRAIN ENTRIES ({len(ws_entries)}):")
        active = [e for e in ws_entries if e['status'] == 'active']
        done   = [e for e in ws_entries if e['status'] == 'done']
        other  = [e for e in ws_entries if e['status'] not in ('active','done')]

        for label, group in [('active', active), ('other/null', other), ('done', done)]:
            if not group:
                continue
            print(f"    -- {label} ({len(group)}) --")
            for e in group:
                link = '🔗' if e['proposal_path'] else '  '
                print(f"    {link} [{e['id'][:8]}] {e['maturity']:<10} st={str(e['status']):<8} {e['title'][:55]}")

    conn.close()


if __name__ == '__main__':
    main()
