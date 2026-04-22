import sqlite3, textwrap
c = sqlite3.connect('private-brain/brain.db')
cur = c.cursor()

print("=== Counts by project x maturity x status ===")
print(f"{'Proj':<24} {'Maturity':<12} {'Status':<14} {'Done':<5} {'Count'}")
for r in cur.execute("""
SELECT COALESCE(p.name, '<INBOX/none>') AS proj,
       COALESCE(e.maturity,'-') AS mat,
       COALESCE(e.status,'<null>') AS estatus,
       e.action_done AS done,
       COUNT(*) AS n
FROM entries e
LEFT JOIN projects p ON p.id = e.project_id
GROUP BY proj, mat, estatus, done
ORDER BY proj, mat, estatus, done
"""):
    print(f"{r[0][:23]:<24} {r[1]:<12} {r[2]:<14} {r[3]!s:<5} {r[4]}")

print("\n=== Totals ===")
for label, q in [
    ("total entries", "SELECT COUNT(*) FROM entries"),
    ("project_id IS NULL (inbox)", "SELECT COUNT(*) FROM entries WHERE project_id IS NULL"),
    ("status IS NULL", "SELECT COUNT(*) FROM entries WHERE status IS NULL OR status=''"),
    ("action_done=1", "SELECT COUNT(*) FROM entries WHERE action_done=1"),
    ("maturity=raw", "SELECT COUNT(*) FROM entries WHERE maturity='raw'"),
    ("maturity=researched", "SELECT COUNT(*) FROM entries WHERE maturity='researched'"),
    ("maturity=planned", "SELECT COUNT(*) FROM entries WHERE maturity='planned'"),
    ("maturity=verified", "SELECT COUNT(*) FROM entries WHERE maturity='verified'"),
    ("maturity=specced", "SELECT COUNT(*) FROM entries WHERE maturity='specced'"),
]:
    print(f"  {label}: {cur.execute(q).fetchone()[0]}")

print("\n=== Maturity values present ===")
for r in cur.execute("SELECT maturity, COUNT(*) FROM entries GROUP BY maturity ORDER BY 2 DESC"):
    print(f"  {r[0]!s:<14} {r[1]}")

print("\n=== Status values present ===")
for r in cur.execute("SELECT status, COUNT(*) FROM entries GROUP BY status ORDER BY 2 DESC"):
    print(f"  {r[0]!s:<14} {r[1]}")
