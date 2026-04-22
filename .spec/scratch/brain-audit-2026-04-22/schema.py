import sqlite3, sys
c = sqlite3.connect('private-brain/brain.db')
cur = c.cursor()

print("=== TABLES ===")
for r in cur.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"):
    print(r[0])

print("\n=== entries schema ===")
for r in cur.execute("PRAGMA table_info(entries)"):
    print(r)

print("\n=== projects schema ===")
for r in cur.execute("PRAGMA table_info(projects)"):
    print(r)

print("\n=== projects ===")
for r in cur.execute("SELECT id, name FROM projects ORDER BY id"):
    print(r)
