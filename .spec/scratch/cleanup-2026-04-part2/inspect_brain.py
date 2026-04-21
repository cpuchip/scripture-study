import sqlite3, sys, json

db = sqlite3.connect('private-brain/brain.db')
db.row_factory = sqlite3.Row
cur = db.cursor()

print("=== TABLES ===")
cur.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
tables = [r[0] for r in cur.fetchall()]
for t in tables:
    cur.execute(f"SELECT COUNT(*) FROM {t}")
    n = cur.fetchone()[0]
    print(f"  {t}: {n} rows")

# Try projects
if 'projects' in tables:
    print("\n=== PROJECTS (all) ===")
    cur.execute("PRAGMA table_info(projects)")
    cols = [r[1] for r in cur.fetchall()]
    print(f"cols: {cols}")
    cur.execute("SELECT * FROM projects ORDER BY 1")
    for r in cur.fetchall():
        d = dict(r)
        print(json.dumps(d, default=str, indent=2))
        print("---")

# entries: just counts by category and recent open
if 'entries' in tables:
    print("\n=== ENTRIES schema ===")
    cur.execute("PRAGMA table_info(entries)")
    for r in cur.fetchall():
        print(f"  {r[1]} {r[2]}")
    print("\n=== ENTRIES by category ===")
    cur.execute("SELECT category, COUNT(*) FROM entries GROUP BY category")
    for r in cur.fetchall():
        print(f"  {r[0]}: {r[1]}")
    print("\n=== ENTRIES by status ===")
    try:
        cur.execute("SELECT status, COUNT(*) FROM entries GROUP BY status")
        for r in cur.fetchall():
            print(f"  {r[0]}: {r[1]}")
    except Exception as e:
        print(f"  (no status col: {e})")
    print("\n=== ENTRIES by route_status ===")
    try:
        cur.execute("SELECT route_status, COUNT(*) FROM entries GROUP BY route_status")
        for r in cur.fetchall():
            print(f"  {r[0]}: {r[1]}")
    except Exception as e:
        print(f"  (no route_status: {e})")
