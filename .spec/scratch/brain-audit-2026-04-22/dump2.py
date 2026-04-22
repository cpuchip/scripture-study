import sqlite3
c = sqlite3.connect('private-brain/brain.db')
cur = c.cursor()
rows = cur.execute("""
SELECT e.id, COALESCE(p.name,'<inbox>'), e.maturity, e.status, e.action_done, e.title
FROM entries e LEFT JOIN projects p ON p.id=e.project_id
WHERE e.action_done=1 ORDER BY p.name, e.title
""").fetchall()
print(f'== action_done=1 ({len(rows)} rows) ==')
for r in rows:
    print(f'[{r[0][:8]}] {r[1]:<22} {r[2]:<10} st={str(r[3]):<9} | {r[5][:60]}')

print('\n\n== Workspace improvements raw (16) ==')
for r in cur.execute("""
SELECT e.id, e.maturity, e.status, e.action_done, e.title, COALESCE(e.one_liner, SUBSTR(e.body,1,140))
FROM entries e JOIN projects p ON p.id=e.project_id
WHERE p.name='Workspace improvements' AND e.maturity='raw'
ORDER BY e.updated_at DESC
"""):
    desc = (r[5] or '').replace('\n',' ').strip()[:100]
    print(f'[{r[0][:8]}] {r[1]:<10} st={str(r[2]):<9} done={r[3]} | {r[4][:60]}')
    if desc and desc.lower() != r[4].lower()[:100]:
        print(f'           {desc}')
