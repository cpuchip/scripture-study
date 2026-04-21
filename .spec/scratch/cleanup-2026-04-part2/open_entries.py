import sqlite3, json
db = sqlite3.connect('private-brain/brain.db')
db.row_factory = sqlite3.Row
cur = db.cursor()

# Open entries (not archived, not done) grouped by project
cur.execute("""
    SELECT
        COALESCE(p.name, '(no project)') AS proj,
        e.id, e.title, e.category, e.status, e.maturity, e.route_status,
        e.created_at, e.updated_at, e.next_action
    FROM entries e
    LEFT JOIN projects p ON e.project_id = p.id
    WHERE COALESCE(e.status, '') NOT IN ('archived', 'done')
    ORDER BY proj, e.updated_at DESC
""")
rows = cur.fetchall()

current = None
for r in rows:
    if r['proj'] != current:
        current = r['proj']
        print(f"\n## {current}")
    title = (r['title'] or '')[:80]
    mat = r['maturity'] or '-'
    st = r['status'] or '-'
    rs = r['route_status'] or '-'
    cat = r['category'] or '-'
    upd = (r['updated_at'] or '')[:10]
    print(f"  [{cat}/{mat}/{st}/{rs}] {title}  ({upd})")
    if r['next_action']:
        na = r['next_action'][:100]
        print(f"      -> {na}")

print(f"\n\nTotal open entries: {len(rows)}")
