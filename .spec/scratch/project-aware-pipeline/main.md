# Scratch: Project-Aware Pipeline

**Binding Problem:** Pipeline agents operating on project-scoped entries don't know where the project lives, write scratch files to the wrong location, and never commit their work — making multi-project operation broken.

---

## Inventory

### Current State — What We Know

**Space Center project config (DB):**
- `workspace_type = "external"`, `workspace_path = "projects\space-center"`
- `context_file = "projects/space-center/.spec"` (truncated in DB display — likely longer)
- Directory exists at `projects/space-center/` with `.github/`, `.spec/proposals|scratch|memory`, `docs/`, `README.md`
- The project's `.spec/scratch/` is empty — research went to workspace root `.spec/scratch/` instead

**Entry: Build Physical Display Dashboard**
- `project_id = 4` (Space Center), `scratch_path = ".spec\scratch\build-physical-display-dashboard\main.md"`
- scratch_path is relative to workspace root, NOT to project dir
- The entry is at maturity "planned"

---

### Root Cause 1: `FormatProjectContext()` omits workspace path

**`BuildProjectContext()` (context.go:86)** fetches: `ProjectName`, `Description`, sibling entries, and `ContextDoc`.
It does NOT include: `WorkspaceType`, `WorkspacePath`, `GithubRepo`, `InitInstructions`.

**`FormatProjectContext()` (context.go:136)** renders: project name, description, sibling entries, context doc.
The agent prompt never mentions WHERE the project directory is.

**`resolveWorkDir()` (context.go:175)** correctly resolves to `projects/space-center` for external projects — so the agent's WorkingDir is set correctly. BUT the agent isn't told that the WorkingDir was changed or that `.spec/scratch/` inside it is where project files should go.

**Impact:** Every pipeline agent (research, plan, execute, review) calls `FormatProjectContext()` and gets the same incomplete info.

---

### Root Cause 2: Scratch paths are always workspace-root-relative

**In `runResearch()` (research.go:308)**:
```go
slug := slugify(entry.Title)
if entry.Category == "study" {
    scratchPath = filepath.Join("study", ".scratch", slug+".md")
} else {
    scratchPath = filepath.Join(".spec", "scratch", slug, "main.md")
}
absPath = filepath.Join(p.workspace, scratchPath)  // ← always workspace root
```

**In `runPlan()` (research.go:580)**: Same pattern — uses existing scratchPath or generates workspace-root-relative.

**The scratch path stored in DB is relative to p.workspace**, not relative to the project dir. Even though the agent's WorkingDir might be `projects/space-center/`, the prompt says "write findings to `C:\...\scripture-study\.spec\scratch\build-physical-display-dashboard\main.md`" — an absolute path in the workspace root.

**In `buildResearchPrompt()`**: Uses the absolute path directly: `fmt.Fprintf(&sb, "Write your findings to this file: %s\n", scratchPath)`.

**In `AllowedWritePaths`**:
- research: `{"study/.scratch", ".spec/scratch"}` — relative to WorkingDir
- plan: `{".spec/scratch", ".spec/proposals", "study/.scratch"}` — relative to WorkingDir
- execute: `{".", ".spec/scratch"}` — relative to WorkingDir

So if WorkingDir is `projects/space-center/`, then `.spec/scratch` in AllowedWritePaths means `projects/space-center/.spec/scratch` — but the prompt tells the agent to write to the workspace root's `.spec/scratch` via absolute path. The agent either (a) follows the prompt's absolute path and governance blocks it, or (b) the governance is lenient and writes to workspace root.

**Also:** `generateProposal()` (research.go:723) hard-codes workspace-root `.spec/proposals/`.

---

### Root Cause 3: No post-execution commit step

**`runExecute()` (execute.go:196)** after completion:
1. `SetAgentOutput(entry.ID, response, 0)` — stores response in DB
2. `AddSessionMessage(...)` — post session message
3. `UpdateRouteStatus(entry.ID, "your_turn")` — route to user
4. (notification) — that's it. No git operations.

**Existing git infrastructure:**
- `store/git.go`: `Git` struct with `CommitFile(relPath, message)`, `CommitAll(message)`
- Used only by `store.go:247` for archiving entries
- Has rate-limiting (`maxCommitsPerDay`), auto-commit toggle, commit prefix
- Operates on the data repo (`private-brain`), NOT on workspace repos

**Key issue:** The git infra is for the private-brain data repo. Execution creates files in workspace repos (scripture-study, or external project repos). Committing to those requires running git in different directories.

**The scaffold.go `gitCommitExternal()`** already has a pattern for this:
- Runs `git init`, `git add -A`, `git commit -m ...` in the project dir
- Optionally creates a GitHub repo with `gh repo create`
- But this only runs during project initialization, not after execution

---

### What Agents Are Affected

| Agent | WorkingDir | Scratch Path Source | Project Context | Git After |
|-------|-----------|-------------------|----------------|-----------|
| research | `resolveWorkDir()` ✓ | Hard-coded workspace root ✗ | Name+desc only ✗ | None ✗ |
| plan | `resolveWorkDir()` ✓ | Inherits or workspace root ✗ | Name+desc only ✗ | None ✗ |
| execute | `resolveWorkDir()` ✓ | Inherits from entry ✗ | Name+desc only ✗ | None ✗ |
| review | `resolveWorkDir()` ✓ | Reads from entry path ✗ | Name+desc only ✗ | N/A |
| scaffold | Uses project dir ✓ | N/A | N/A | External only ✓ |

---

## Critical Analysis

**Is this the RIGHT thing to build?**
Yes — this isn't speculative. Michael hit this bug in real usage. The Space Center pipeline test entry wrote to the wrong directory. Multi-project is a core design goal (the whole project/entry system exists for it) and it doesn't work yet.

**Does this solve the binding problem?**
Directly. Three surgical fixes to make project-scoped entries work correctly.

**What's the simplest version?**
Fix 1 (prompt context) and Fix 2 (scratch paths) are the minimum for correctness. Fix 3 (auto-commit) is a quality-of-life improvement — manual git commit is viable short-term. But Michael explicitly called it out ("I didn't see the haiku agent run to look at what was generated and generate commits"), so it's expected.

**What gets WORSE?**
- Fix 1 adds ~50-100 tokens to every project-scoped prompt. Trivial.
- Fix 2 changes scratch path conventions. Existing entries with workspace-root scratch paths need migration or compatibility handling.
- Fix 3 adds post-execution side effects (git operations). Could fail on uncommitted changes, merge conflicts, or git not being configured. Must be best-effort, not blocking.

**Does this duplicate something?**
No — `scaffold.go` has git commit logic for initialization only. We'd reuse the pattern but for a different trigger point.

**Is this the right time?**
Yes. Michael is actively testing the pipeline with Space Center. This is blocking real usage.

**Mosiah 4:27 check:**
These are targeted fixes to existing code, not new features. ~1 session of dev work. Low cognitive overhead, high payoff.

**Key design tension: scratch path migration**
The DB stores `scratch_path` as a string. Existing entries (like the dashboard entry) have workspace-root-relative paths. Options:
1. **Migrate existing paths** — update DB records for project-scoped entries to use project-relative paths. Move files on disk too.
2. **Dual-path resolution** — check project dir first, workspace root second. No migration.
3. **Only fix going forward** — new entries get project-scoped paths, existing ones keep working where they are.

Option 3 is cheapest but creates permanent inconsistency. Option 2 is robust but adds a subtle "where's my file?" debugging problem. Option 1 is cleanest but harder — need to know which entries are project-scoped and adjust.

**Recommendation:** Option 3 with a one-time manual move for the dashboard entry. The scope of project-scoped entries is small enough that migration isn't worth automating.

**Key design tension: git commit scope**
Post-execution commits could be:
1. **Mechanical commit** — `git add -A && git commit -m "brain: {entry title}"` in the project/workspace dir
2. **Haiku-generated commit message** — cheap agent reads the diff and writes a good message
3. **Per-workspace git instance** — different repos have different git configs

Michael said "I didn't see the haiku agent run" — he expected option 2. But option 1 is simpler and cheaper. The execution agent already produces a detailed session message about what it did. A mechanical commit with `"brain: {entry.Title}"` as the message is sufficient for a first pass. Haiku commit messages can be a Phase 2 enhancement.

**Recommendation:** Start with mechanical commits. The pattern already exists in `scaffold.go:gitCommitExternal()`.

---

## Decision Points for Michael

**D1: Scratch path strategy for existing entries**
- Go forward only (new entries get project-scoped paths, existing keep workspace-root paths)
- Or migrate existing dashboard entry manually after the fix

**D2: Commit message style**
- Mechanical: `"brain: Build Physical Display Dashboard"` — free, instant
- Haiku-generated from diff — 0.33 premium requests per commit, better messages

**D3: Commit scope for integrated projects**
- Integrated projects live in scripture-study workspace root. Committing there after every execution could be noisy. Should we only auto-commit for external/subfolder projects?
- Or commit for all but require the user to review with `git status` before push?

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | Pipeline agents can't operate correctly on project-scoped entries — they write to wrong dirs and never commit |
| Covenant | Rules of engagement? | Existing patterns: `resolveWorkDir()`, `FormatProjectContext()`, `gitCommitExternal()`. Extend, don't reinvent |
| Stewardship | Who owns what? | `context.go` owns project context, `research.go` owns scratch paths, `execute.go` owns post-execution flow |
| Spiritual Creation | Is the spec precise enough? | Yes — three discrete changes with code sketches, all in existing files |
| Line upon Line | Phasing? | Single phase — all three fixes are tightly coupled and small (~100 LOC total) |
| Physical Creation | Who executes? | dev agent |
| Review | How do we know it's right? | go vet, vue-tsc, go test, then manually run a research pass on the dashboard entry |
| Atonement | What if it goes wrong? | Scratch path changes are forward-only. Git commits are best-effort. No destructive changes |
| Sabbath | When do we stop and reflect? | After manual verification with Space Center entry |
| Consecration | Who benefits? | Michael — unblocks multi-project pipeline usage |
| Zion | How does this serve the whole? | Closes the gap between project setup (Phase 8) and project operation |
