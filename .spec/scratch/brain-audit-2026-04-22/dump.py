import sqlite3
c = sqlite3.connect('private-brain/brain.db')
cur = c.cursor()

def dump(title, where):
    print(f"\n\n========== {title} ==========")
    rows = cur.execute(f"""
SELECT e.id, COALESCE(p.name,'<inbox>') AS proj, e.maturity, e.status, e.action_done,
       e.title, COALESCE(e.one_liner, SUBSTR(e.body,1,140)) AS desc,
       e.created_at, e.updated_at
FROM entries e
LEFT JOIN projects p ON p.id=e.project_id
WHERE {where}
ORDER BY proj, e.maturity DESC, e.updated_at DESC
""").fetchall()
    for r in rows:
        eid, proj, mat, st, done, ttl, desc, ca, ua = r
        desc = (desc or '').replace('\n',' ').strip()[:120]
        print(f"[{eid[:8]}] {proj:<22} {mat:<10} st={str(st):<9} done={done} | {ttl[:60]}")
        if desc and desc.lower() != ttl.lower()[:120]:
            print(f"           {desc}")
    print(f"--- {len(rows)} rows ---")

# 1. INBOX (no project assigned)
dump("INBOX (project_id IS NULL)", "e.project_id IS NULL")

# 2. Researched/Planned/Specced/Verified/Complete in any project
dump("Mature entries (planned/specced/verified/complete) — need workspace records",
     "e.maturity IN ('planned','specced','verified','complete')")

# 3. action_done=1 but status not closed
dump("Done by action but status open — needs closure",
     "e.action_done=1 AND (e.status IS NULL OR e.status NOT IN ('done','archived'))")
