# Project-Aware Pipeline

**Status:** Draft
**Binding Problem:** Pipeline agents operating on project-scoped entries don't know where the project lives, write scratch files to the wrong location, and never commit their work — making multi-project operation broken.

**Prior Art:** Phase 8 (Agent-Driven Project Initialization) in `.spec/proposals/brain-pipeline-evolution.md` — solved directory creation and governance files but didn't address ongoing agent operations within the project.

---

## Problem Statement

When an entry belongs to an external project (like Space Center at `projects/space-center/`), three things go wrong:

1. **Agent prompts don't mention the project directory.** `FormatProjectContext()` includes the project name and description but omits `workspace_type`, `workspace_path`, and any indication of where the project's files live. The agent's CWD is correctly set via `resolveWorkDir()`, but the agent isn't told this or told what the project's directory structure looks like.

2. **Scratch files are created in the workspace root.** `runResearch()` hard-codes scratch paths as `.spec/scratch/{slug}/main.md` relative to the workspace root — even when the project has its own `.spec/scratch/` directory. The absolute path is passed to the agent in the prompt, so the agent dutifully writes there.

3. **No git commit happens after execution.** The pipeline marks the entry as "your_turn" and posts a session message, but never commits the files the agent created. The existing git infrastructure only handles private-brain data archiving, not workspace repo operations.

---

## Success Criteria

- An agent working on a Space Center entry knows it's operating in `projects/space-center/` and that its `.spec/scratch/`, `.spec/proposals/`, etc. are there.
- Research creates scratch files at `projects/space-center/.spec/scratch/{slug}/main.md`, not at workspace root.
- After execution completes, changed files are committed in the appropriate git repo.
- Integrated project entries continue working as before (no regression).
- Study entries continue using `study/.scratch/` (no change).

---

## Constraints

- No database schema changes. scratch_path is just a string — the fix is in how it's computed and resolved.
- New entries going forward get project-scoped paths. Existing entries may keep workspace-root paths (option 3 from analysis).
- Git commits are best-effort — failures log warnings but don't block the pipeline.
- No Haiku commit agent in Phase 1 — mechanical commits with `"brain: {entry.Title}"` messages. Can upgrade later.

---

## 9a — Project Context in Agent Prompts

**Problem:** `FormatProjectContext()` renders project name, description, sibling entries, and context doc — but never mentions workspace_type, workspace_path, or the project directory structure.

**Fix:** Add workspace location fields to `ProjectContext` struct and `FormatProjectContext()` output.

### Changes

**`context.go` — `ProjectContext` struct:**
Add fields:
```go
WorkspaceType string // "integrated", "subfolder", "external"
WorkspacePath string // relative path from workspace root (e.g. "projects/space-center")
```

**`context.go` — `BuildProjectContext()`:**
Populate new fields from project:
```go
ctx.WorkspaceType = project.WorkspaceType
ctx.WorkspacePath = project.WorkspacePath
```

**`context.go` — `FormatProjectContext()`:**
Add workspace location to rendered output:
```go
if ctx.WorkspacePath != "" {
    sb.WriteString(fmt.Sprintf("**Project directory:** %s\n", filepath.ToSlash(ctx.WorkspacePath)))
    sb.WriteString("All project files (scratch, proposals, docs) should be created within this directory.\n")
}
if ctx.WorkspaceType == "external" {
    sb.WriteString("This is an external project with its own git repository.\n")
}
```

### Verification
- [ ] `FormatProjectContext()` output includes workspace path for project-scoped entries
- [ ] For integrated projects (no workspace_path), output is unchanged
- [ ] All agent prompts (research, plan, execute, review) inherit the fix via existing `FormatProjectContext()` calls

---

## 9b — Project-Scoped Scratch Paths

**Problem:** `runResearch()` and `runPlan()` compute scratch paths relative to workspace root. `generateProposal()` also hard-codes workspace root for proposals.

**Fix:** When an entry belongs to a project with a `workspace_path`, prefix the scratch/proposal paths with that workspace path.

### Changes

**New helper — `context.go` or `research.go`:**
```go
// projectRelPath returns a path prefixed with the project's workspace path
// if the entry belongs to a project with one. Otherwise returns the path unchanged.
func (p *Pipeline) projectRelPath(entry *store.Entry, relPath string) string {
    if entry.ProjectID == nil {
        return relPath
    }
    project, err := p.store.DB().GetProject(*entry.ProjectID)
    if err != nil || project == nil || project.WorkspacePath == "" {
        return relPath
    }
    return filepath.Join(project.WorkspacePath, relPath)
}
```

**`research.go` — `runResearch()`:**
Change scratch path computation:
```go
slug := slugify(entry.Title)
var scratchPath string
if entry.Category == "study" {
    scratchPath = filepath.Join("study", ".scratch", slug+".md")
} else {
    scratchPath = filepath.Join(".spec", "scratch", slug, "main.md")
}
scratchPath = p.projectRelPath(entry, scratchPath)  // NEW
```

**`research.go` — `runPlan()`:**
Same pattern — when generating a new scratch path (not reusing existing):
```go
scratchPath = p.projectRelPath(entry, scratchPath)
```

**`research.go` — `generateProposal()`:**
```go
proposalPath := filepath.Join(".spec", "proposals", slug+".md")
proposalPath = p.projectRelPath(entry, proposalPath)  // NEW
```

**`research.go` — `AllowedWritePaths` for research/plan agents:**
Currently hard-coded to `{".spec/scratch", ".spec/proposals"}`. These are relative to WorkingDir. For external projects, WorkingDir is already the project dir, so `.spec/scratch` resolves correctly. No change needed here — the fix is in the scratch path computation + prompt path.

**Key insight:** For external projects, the agent's WorkingDir is already `projects/space-center/` (via `resolveWorkDir()`). The scratch path just needs to match — `projects/space-center/.spec/scratch/{slug}/main.md` relative to workspace root. The absolute path in the prompt will then point into the project dir, and `AllowedWritePaths` relative to WorkingDir will permit it.

For integrated projects (WorkspaceType="" or "integrated", no WorkspacePath), nothing changes.

### Verification
- [ ] New research on a Space Center entry creates scratch at `projects/space-center/.spec/scratch/{slug}/main.md`
- [ ] New research on an integrated entry still creates scratch at `.spec/scratch/{slug}/main.md`
- [ ] Plan pass finds and appends to the project-scoped scratch file
- [ ] Study entries continue using `study/.scratch/` unchanged
- [ ] Proposal generation for project entries goes to `projects/space-center/.spec/proposals/`

---

## 9c — Post-Execution Git Commit

**Problem:** After `runExecute()` completes, no git operations happen. Files created by the agent sit uncommitted.

**Fix:** Add a `commitAfterExecution()` step that runs `git add -A && git commit` in the appropriate directory after successful execution.

### Changes

**`execute.go` — after execution success (after `SetAgentOutput`, before route_status update):**
```go
// Commit changes if applicable
p.commitAfterExecution(entry)
```

**New function in `execute.go` or `context.go`:**
```go
// commitAfterExecution runs git add + commit in the project directory after
// successful execution. Best-effort — failures log but don't block.
func (p *Pipeline) commitAfterExecution(entry *store.Entry) {
    dir := p.resolveWorkDir(entry)

    // Only auto-commit for external/subfolder projects with their own directory.
    // Integrated projects share the workspace root — too noisy to auto-commit.
    if dir == p.workspace {
        return
    }

    // Check if this is actually a git repo
    gitDir := filepath.Join(dir, ".git")
    if _, err := os.Stat(gitDir); os.IsNotExist(err) {
        return
    }

    msg := fmt.Sprintf("brain: %s", entry.Title)
    if err := gitCommitAll(dir, msg); err != nil {
        log.Printf("post-execution commit failed for %s: %v", entry.ID, err)
        return
    }
    log.Printf("post-execution commit: %s in %s", msg, dir)
}

// gitCommitAll stages all changes and commits in the given directory.
func gitCommitAll(dir, message string) error {
    // git add -A
    add := exec.Command("git", "add", "-A")
    add.Dir = dir
    if out, err := add.CombinedOutput(); err != nil {
        return fmt.Errorf("git add: %w (%s)", err, string(out))
    }

    // git commit -m message (--allow-empty is not needed; skip if nothing staged)
    commit := exec.Command("git", "commit", "-m", message)
    commit.Dir = dir
    out, err := commit.CombinedOutput()
    if err != nil {
        if strings.Contains(string(out), "nothing to commit") {
            return nil
        }
        return fmt.Errorf("git commit: %w (%s)", err, string(out))
    }
    return nil
}
```

**Scope:** Only external/subfolder projects get auto-commit. Integrated projects share the workspace root where Michael's own work is in progress — auto-committing there would be disruptive. This matches Michael's expectation: project repos should be self-managing.

### Verification
- [ ] After executing a Space Center entry, `projects/space-center/` gets a git commit
- [ ] The commit message includes the entry title
- [ ] If nothing changed, no empty commit is created
- [ ] If git fails, execution still succeeds (best-effort)
- [ ] Integrated entries do NOT trigger auto-commit in workspace root
- [ ] The commit appears in `git log` for the project directory

---

## Phased Delivery

**Phase 1 (this session):** All three fixes — 9a, 9b, 9c. They're tightly coupled and individually small. Total: ~100 lines of Go changes across context.go, research.go, execute.go.

**Phase 2 (future, optional):** Haiku commit message generation. After Phase 1 proves the pattern, upgrade from mechanical to AI-generated commit messages that describe what changed. Cost: 0.33 premium requests per commit.

---

## Costs & Risks

| Item | Cost | Risk |
|------|------|------|
| 9a: Prompt context | ~50 tokens/prompt | None — additive |
| 9b: Scratch paths | Code change only | Existing entries keep old paths — minor inconsistency |
| 9c: Git commit | `exec.Command("git", ...)` | Could fail if git not configured; best-effort mitigates |
| Total dev time | ~1 session | Low — pattern exists in scaffold.go |
