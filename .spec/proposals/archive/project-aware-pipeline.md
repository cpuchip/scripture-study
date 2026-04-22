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
- New entries going forward get project-scoped paths. Existing dashboard entry already moved manually (D1 done).
- Git commits are best-effort — failures log warnings but don't block the pipeline.
- Haiku-generated commit messages from the start (D2), not mechanical.
- Auto-commit for ALL project types including integrated (D3), but selective — only files the session changed.

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

## 9c — Post-Execution Git Commit (Selective, All Project Types)

**Problem:** After `runExecute()` completes, no git operations happen. Files created by the agent sit uncommitted.

**Fix:** Track which files the agent writes during the session, then selectively `git add` only those files and commit with a Haiku-generated message. Applies to ALL project types — external, subfolder, and integrated.

### Design Decisions (from Michael)

- **D2: Haiku commit messages from the start.** Not mechanical — the review agent (Haiku 3.5, 0.33 premium requests) generates the commit message. Format: `brain({slug}): {haiku-generated description}`. The slug comes from the entry title (e.g., `brain(build-physical-display-dashboard): scaffold dashboard layout with tile grid and API routes`).
- **D3: Selective commits for ALL project types.** Including integrated projects in the base workspace. But ONLY the files that the agent session actually touched — never `git add -A`. This prevents accidentally committing Michael's in-progress work.

### Part 1: File Tracking in Agent Sessions

The PostToolUse hook in `agent.go` already fires for every tool call and we already have `isWriteTool()` and `extractPathCandidates()` in `governance.go`. We reuse these to collect written file paths.

**`ai/agent.go` — Add file tracking:**
```go
// In Agent struct, add:
writtenFiles map[string]bool // set of absolute file paths written during session

// In createSession() PostToolUse hook, after existing AUDIT log:
if isWriteTool(input.ToolName) {
    for _, path := range extractPathCandidates(input.ToolArgs) {
        abs := path
        if !filepath.IsAbs(abs) {
            abs = filepath.Join(a.config.WorkingDir, abs)
        }
        abs = filepath.Clean(abs)
        a.writtenFiles[abs] = true
    }
}

// New method:
func (a *Agent) WrittenFiles() []string {
    files := make([]string, 0, len(a.writtenFiles))
    for f := range a.writtenFiles {
        files = append(files, f)
    }
    return files
}
```

**Key:** `extractPathCandidates()` already handles the common tool arg keys (path, filepath, dirpath, old_path, new_path, workspacefolder). For `run_in_terminal`, file tracking won't capture paths (terminal commands are opaque) — but that's acceptable since most agent writes use the file tools.

### Part 2: Selective Git Commit

After execution, commit only the tracked files. Determine the correct git repo for each file.

**New function in `execute.go`:**
```go
// commitAfterExecution selectively commits files written during the agent session.
// Groups files by git repo and commits each group separately.
// Best-effort — failures log but don't block the pipeline.
func (p *Pipeline) commitAfterExecution(entry *store.Entry, writtenFiles []string) {
    if len(writtenFiles) == 0 {
        return
    }

    // Group files by their git repo root
    repoFiles := map[string][]string{} // repoRoot -> []relativePaths
    for _, absPath := range writtenFiles {
        repoRoot := findGitRoot(absPath)
        if repoRoot == "" {
            continue // not in a git repo
        }
        rel, err := filepath.Rel(repoRoot, absPath)
        if err != nil {
            continue
        }
        repoFiles[repoRoot] = append(repoFiles[repoRoot], rel)
    }

    for repoRoot, files := range repoFiles {
        msg := p.generateCommitMessage(entry, files)
        if err := gitCommitSelective(repoRoot, files, msg); err != nil {
            log.Printf("post-execution commit failed in %s: %v", repoRoot, err)
            continue
        }
        log.Printf("post-execution commit: %s (%d files in %s)", msg, len(files), repoRoot)
    }
}

// findGitRoot walks up from the given path to find the nearest .git directory.
func findGitRoot(path string) string {
    dir := filepath.Dir(path)
    for {
        if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            return "" // reached filesystem root
        }
        dir = parent
    }
}

// gitCommitSelective stages specific files and commits.
func gitCommitSelective(repoRoot string, files []string, message string) error {
    // git add <file1> <file2> ...
    args := append([]string{"add", "--"}, files...)
    add := exec.Command("git", args...)
    add.Dir = repoRoot
    if out, err := add.CombinedOutput(); err != nil {
        return fmt.Errorf("git add: %w (%s)", err, string(out))
    }

    // git commit -m message
    commit := exec.Command("git", "commit", "-m", message)
    commit.Dir = repoRoot
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

### Part 3: Haiku Commit Message Generation

The commit message is generated by Haiku (0.33 premium requests) with the entry title prepended as context.

**New function in `execute.go`:**
```go
// generateCommitMessage uses Haiku to generate a short commit message.
// Falls back to mechanical message if Haiku fails.
func (p *Pipeline) generateCommitMessage(entry *store.Entry, files []string) string {
    slug := slugify(entry.Title)
    prefix := fmt.Sprintf("brain(%s)", slug)

    // Build a short prompt listing the files changed
    fileList := strings.Join(files, "\n  ")
    prompt := fmt.Sprintf(
        "Generate a concise git commit message (one line, max 72 chars after the prefix) "+
            "for these changes made while working on '%s':\n  %s\n"+
            "Reply with ONLY the message body, no prefix.",
        entry.Title, fileList,
    )

    // Quick Haiku call — single turn, no tools
    body, err := p.quickHaikuCall(prompt)
    if err != nil || body == "" {
        // Fallback: mechanical message
        return fmt.Sprintf("%s: pipeline execution", prefix)
    }

    // Trim and enforce length
    body = strings.TrimSpace(body)
    if len(body) > 72-len(prefix)-2 {
        body = body[:72-len(prefix)-2]
    }
    return fmt.Sprintf("%s: %s", prefix, body)
}
```

`quickHaikuCall()` creates a minimal single-turn Haiku session (no tools, no MCP) just for the commit message. Cost: 0.33 premium requests per commit.

### Verification
- [ ] After executing a Space Center entry, files in `projects/space-center/` are selectively committed
- [ ] After executing an integrated entry, only the files the agent wrote in the workspace root are committed
- [ ] Files written via `create_file`, `replace_string_in_file`, etc. are tracked
- [ ] The commit message starts with `brain({slug}):` and includes a Haiku-generated description
- [ ] If Haiku fails, the fallback message is `brain({slug}): pipeline execution`
- [ ] If nothing changed, no empty commit is created
- [ ] If git fails, execution still succeeds (best-effort)
- [ ] Files in different git repos get separate commits
- [ ] `run_in_terminal` writes are not tracked (known limitation, acceptable)

---

## Phased Delivery

**Single phase (this session):** All three fixes — 9a, 9b, 9c. They're tightly coupled and individually small. 9c is the largest due to file tracking + Haiku commit messages, but the pieces are well-defined.

Estimated changes:
- `context.go`: ~10 lines (struct fields + format output)
- `research.go`: ~15 lines (projectRelPath helper + 3 call sites)
- `execute.go`: ~80 lines (commitAfterExecution, findGitRoot, gitCommitSelective, generateCommitMessage)
- `agent.go`: ~20 lines (writtenFiles tracking in PostToolUse hook + WrittenFiles() method)

---

## Costs & Risks

| Item | Cost | Risk |
|------|------|------|
| 9a: Prompt context | ~50 tokens/prompt | None — additive |
| 9b: Scratch paths | Code change only | Existing entries keep old paths — minor inconsistency |
| 9c: File tracking | Memory — map in Agent struct | `run_in_terminal` writes not tracked (known limitation) |
| 9c: Haiku commit msg | 0.33 premium requests/commit | Could fail — falls back to mechanical message |
| 9c: Git commit | `exec.Command("git", ...)` | Could fail if git not configured; best-effort mitigates |
| 9c: Multi-repo | findGitRoot walk | 13 nested repos — must find correct one per file |
| Total dev time | ~1 session | Medium — file tracking is new infrastructure |
