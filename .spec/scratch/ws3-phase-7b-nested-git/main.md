# WS3 Phase 7b — Nested Git Repo Awareness

## Binding Problem

The workspace contains 13 nested git repos. The current git status/diff endpoints only query the workspace root, so changes inside subrepos (scripts/brain/, teaching/, etc.) are invisible. Users can't see what changed in nested repos, and the file tree gives no indication which directories are independent git repos.

## Inventory of Nested Repos

From `Get-ChildItem -Recurse -Directory -Filter ".git" -Force -Depth 3`:

1. `.` (workspace root — scripture-study)
2. `external_context/autoresearch/`
3. `external_context/autoresearch-win-rtx/`
4. `external_context/coder/`
5. `external_context/modern-lcars/`
6. `external_context/squad/`
7. `external_context/superpowers/`
8. `external_context/tpg/`
9. `private/`
10. `private-brain/`
11. `scripts/brain/`
12. `scripts/brain-app/`
13. `scripts/chip-voice/`
14. `teaching/`

Note: `private-brain/` is already excluded from the file tree (in `skipDirs`). So effectively 12 visible nested repos.

## Current State

### Backend (server.go)
- `handleGitStatus`: runs `git status --porcelain` in workspace root only
- `handleGitDiff`: runs `git diff HEAD` / `git diff --no-index` in workspace root only
- `handleFileTree`: builds recursive tree, skips `.git` dirs. TreeNode has: Name, Path, IsDir, Children. No git repo indicator.

### Frontend
- `FileTreeNode` interface: name, path, is_dir, children. No is_git_repo field.
- `TreeNode.vue`: receives gitStatus Map, shows colored dots. No repo indicator.
- `LibraryView.vue`: calls gitStatus once on mount, builds single Map for whole workspace.

## Key Design Decision

**Option A: Backend discovers nested repos and runs git in each.**
- `handleGitStatus` walks workspace looking for `.git` dirs, runs `git status --porcelain` in each, prefixes paths with the repo's relative path.
- `handleGitDiff` detects which repo a file belongs to and runs `git diff` there.
- `handleFileTree` adds `is_git_repo: true` to TreeNode for directories containing `.git`.

**Option B: New endpoint for repo discovery, existing endpoints parameterized.**
- `GET /api/git/repos` returns list of nested repo paths.
- Frontend calls `gitStatus` once per repo path (or backend aggregates).

**Recommendation: Option A.** Simpler frontend — the backend does the aggregation. The frontend just sees a unified status map with an extra field on tree nodes.

## Implementation Notes

### handleGitStatus changes
- Walk workspace root looking for `.git` dirs (up to depth ~3)
- For each repo found, run `git status --porcelain` in that directory
- Prefix each file path with the repo's relative path from workspace root
- Return unified list across all repos
- Add `repo` field to GitFileStatus so frontend knows which repo a file belongs to

### handleGitDiff changes  
- Given a file path, determine which git repo it belongs to by checking parent dirs for `.git`
- Run `git diff` in that repo's root, with the path relative to that repo

### handleFileTree changes
- When building tree, check if directory contains `.git` child
- Add `is_git_repo` bool to TreeNode struct

### Frontend changes
- FileTreeNode gains `is_git_repo?: boolean`
- TreeNode.vue shows a repo indicator (icon/badge) for git repo directories
- gitStatus/gitDiff "just work" since backend aggregates

## Critical Analysis

1. **Right thing?** Yes — this is the natural completion of the git integration story. Without this, the brain subrepo (where most development happens) is invisible.
2. **Simplest version?** Backend aggregates, frontend mostly unchanged. One new field on TreeNode.
3. **What gets worse?** Slightly slower git status (runs git N times instead of 1). But N is small (13) and `git status --porcelain` is fast (~50ms).
4. **Right time?** Directly follows 7a. Same code, same mental model.
