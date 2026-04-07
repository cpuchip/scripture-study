# WS4 Phase 7 — Pre-Build Review

**Date:** April 6–7, 2026

## Binding Problem

The pipeline needs to support projects that live in different workspace configurations — some are subfolders of scripture-study (same git repo), some are separate git repos (nested or external). The current spec assumes all external projects need their own git repo, which misses the subfolder pattern.

## Research Findings (April 6)

### Governance Doc Pickup

The pipeline has **two layers of governance** injection:

**Layer 1: Global covenants** (always applied)
- `docs/governance/research-covenant.md` → injected into research agent system message
- `docs/governance/plan-covenant.md` → injected into plan agent system message
- `docs/governance/execute-covenant.md` → injected into execution agent system message
- `review-covenant.md` exists but is NOT yet wired into review.go nudge prompts
- These load from `p.codeDir` (scripts/brain/docs/governance/) — they are brain's internal governance
- They apply to ALL entries regardless of project

**Layer 2: Project context** (per-project, optional)
- `Project.ContextFile` → loaded from `p.workspace` (scripture-study root)
- Injected into ALL agents working on that project's entries
- Capped at 3000 chars

**Gap found:** `WorkspaceConfig.BaseInstructions` loads scripture-study's `.github/copilot-instructions.md` but this is NOT injected into pipeline agent prompts. Pipeline agents don't know about voice, covenant, or core principles from the workspace level.

### Workspace Patterns

Three patterns exist in practice:
- `scripts/becoming/` (ibeco.me) — subfolder within scripture-study's `.git`
- `scripts/brain/` — separate `.git` repo nested inside scripture-study workspace
- Future external projects — their own directory and repo

### VS Code Agent Visibility

If you open a separate repo in its own VS Code window, Copilot only sees that repo's `.github/`. Scripture-study's governance is invisible. For subfolders, VS Code agents see scripture-study's governance because you're in that workspace.

---

## Decisions (April 7)

### Decision 1: Inject workspace base instructions into pipeline agents

**YES.** Scripture-study's `.github/copilot-instructions.md` will be injected as Layer 0 into all pipeline agent system messages. This gives agents the personality, voice, covenant, and principles that make them work.

Prompt assembly order:
1. Workspace base instructions (scripture-study copilot-instructions — trimmed/summarized)
2. Phase-specific covenant (research/plan/execute/review)
3. Project context (project-specific architecture, conventions)
4. Task instructions (what to do with this specific entry)

### Decision 2: Option C — Thin project instructions on disk, fat at runtime

Each project's `copilot-instructions.md` is thin and project-specific — architecture, conventions, what makes this project unique. The pipeline injects the scripture-study base layer at runtime. Projects typically live as folders within the scripture-study workspace, so VS Code Copilot also sees the base governance.

For the rare case of opening a project in its own VS Code window, the project's instructions include a reference back to scripture-study for the human, but don't duplicate the full base instructions.

### Decision 3: Explicit workspace_type

The user explicitly chooses the workspace type when creating/editing a project. No inference from path — some subfolders have `.git` and some don't, and at project creation time the directory might not exist yet.

Three types:
- `integrated` — work happens in scripture-study root (default, current behavior)
- `subfolder` — relative path within scripture-study, same git
- `external` — own git repo in `./projects/{name}/`, full provenance structure

### Decision 4: ./projects/ folder for external projects

New external projects live at `./projects/{name}/` within scripture-study's directory tree. Each gets:
- Its own `.git` repo
- `.github/copilot-instructions.md` (thin, project-specific)
- `.spec/` for proposals, scratch, memory
- Own provenance — research, plans, work products all live there
- GitHub remote on cpuchip

This means they're accessible from the scripture-study workspace (pipeline and VS Code can see them) but are independent git repos that can be pushed, cloned, and worked on separately.

### Decision 5: Context file cap needs raising

The 3000-char cap on `Project.ContextFile` may be too tight for project-level instructions that serve as the primary governance layer. Will evaluate during implementation — may raise to 8000 or remove for workspace-type projects that have their own instructions file.

---

## Pre-Build Cleanup Items

1. **Wire review-covenant.md into review.go** — exists but isn't loaded in nudge prompts
2. **Raise or remove context_file char cap** — evaluate during implementation
