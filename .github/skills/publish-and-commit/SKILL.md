---
name: publish-and-commit
description: "Run the publish script to convert study documents to public HTML, then git add, commit, and push. Use when the user says publish, commit, push, or asks to share their work."
user-invokable: true
argument-hint: "[commit message]"
---

# Publish & Commit

## What the Publish Script Does

The Go script at `scripts/publish/cmd/main.go` copies documents from source directories (`study/`, `lessons/`, `docs/work-with-ai/`, `callings/`, `journal/`) into `public/`, converting internal gospel-library links to churchofjesuschrist.org URLs.

## Steps

### 1. Run the Publish Script
```powershell
go run .\scripts\publish\cmd\main.go
```

Run from the **workspace root** (`scripture-study/`).

**Expected output:**
- File count (e.g., "Published 103 files")
- Link conversion count (e.g., "Converted 3551 links")
- Any warnings about broken links or missing files

**Verify:** Check that the file count and link count are reasonable. If either is 0, something went wrong.

### 2. Check What Changed
```powershell
git status --short
```

Review the changes. New files in `public/` should correspond to new or updated source files.

### 3. Stage, Commit, and Push
```powershell
git add -A
git commit -m "publish: {description of changes}"
git push
```

Use a descriptive commit message. If the user provided one, use it. Otherwise, summarize what was published (e.g., "publish: add know-god study" or "publish: update lesson series with periodic review").

## Common Issues

| Problem | Solution |
|---------|----------|
| Publish shows 0 files | Check you're in the workspace root, not a subdirectory |
| Link conversion count drops | A source file may have broken relative links — check warnings |
| `git push` fails | Check if you need to `git pull --rebase` first |
| Files in `public/` but not in source | The publish script copies, not syncs — delete stale `public/` files manually if needed |

## When to Publish

- After completing or updating a study document
- After creating or updating a lesson
- After modifying the `docs/work-with-ai/` lesson series
- When the user explicitly asks ("please publish", "publish and commit", "push it up")
