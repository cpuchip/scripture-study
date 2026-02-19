# Agent Skills & Customization Gap Analysis

*Created: 2026-02-17*
*Triggered by: Jerry Nixon's Azure SQL + DAB demo using agent skills — "Instructions set the rules. Skills do the work."*

---

## The Landscape

VS Code Copilot offers seven customization categories. We use three of them.

| Category | Status | Our Implementation |
|----------|--------|--------------------|
| **Always-on instructions** | ✅ Using | `.github/copilot-instructions.md` — ~70 lines, warmth-first, project structure, core principles |
| **Custom agents** | ✅ Using | 8 agents in `.github/agents/` — study, dev, eval, lesson, talk, review, journal, podcast |
| **MCP servers** | ✅ Using | 7 servers — gospel-mcp, gospel-vec, webster-mcp, becoming-mcp, yt-mcp, search-mcp, playwright |
| **Agent skills** | ❌ Not using | No `.github/skills/` directory |
| **Prompt files** | ❌ Not using | No `.github/prompts/` directory (but `docs/work-with-ai/expound-prompt.md` is a manual copy/paste prompt) |
| **File-based instructions** | ❌ Not using | No `*.instructions.md` files in the workspace |
| **Hooks** | ❌ Not using | No `.github/hooks/` directory |

---

## What We Have (Detailed Audit)

### Always-on Instructions
`.github/copilot-instructions.md` — lean and warm. Covers:
- Project identity (collaboration, not transaction)
- Project structure table
- Core principles (search → read → quote, follow footnotes, link everything)
- Agent mode table (pointers to agents, not duplicate instructions)
- Living documents (tool observations)

**Assessment:** Well-structured. Not bloated. This is the right size for always-on — it applies to every request regardless of mode.

### Custom Agents (8)

| Agent | Lines | Key Content |
|-------|-------|-------------|
| `study` | 77 | Two-phase workflow (discovery → deep reading), pre-publish checklist, link format |
| `dev` | 95 | Go conventions, MCP architecture, design principles, becoming app startup |
| `eval` | 53 | Video evaluation 7-step workflow, cite count rule, link format |
| `lesson` | 60 | Teaching framework, question design, lesson prep steps, link format |
| `talk` | 56 | Talk structure, scripture density goals, practical notes, link format |
| `review` | 52 | 5-dimension analysis framework, rhetoric patterns, link format |
| `journal` | 59 | Becoming layer, memorization, tone guidance, transformation focus |
| `podcast` | 108 | Episode structure (hook/setup/discovery/landing), example, link format |

**Handoffs configured:** study→journal/lesson, eval→study/journal, lesson→study/journal, talk→study/journal, review→talk/lesson, journal→study, podcast→study/journal

**Assessment:** Rich and purposeful. Each agent has a clear identity. But there's significant content duplication — see below.

### MCP Servers (7)

| Server | Purpose | Usage Frequency |
|--------|---------|-----------------|
| gospel-mcp | FTS5 keyword search | Every study session |
| gospel-vec | Semantic vector search | Most study sessions |
| webster-mcp | Webster 1828 + modern dictionary | When historical word meaning matters |
| becoming-mcp | Practice tracking, tasks, notes | Journal mode |
| yt-mcp | YouTube transcript download/search | Eval mode |
| search-mcp | DuckDuckGo web search | Occasional |
| playwright | Browser automation | Testing, rare |

### Templates (Manual, Not Automated)
- `docs/study_template.md` — study session patterns, pre-publish checklist
- `docs/lesson_template.md` — lesson outline with question design
- `docs/talk_template.md` — talk structure with multiple opening/body options
- `docs/yt_evaluation_template.md` — 4-step evaluation framework

**Assessment:** These are *referenced* by agents but never *automated*. You have to manually copy or remember them.

### Manual Prompt
- `docs/work-with-ai/expound-prompt.md` — a fully-formed prompt meant to be copy/pasted into chat sessions to mine examples for the lesson series

---

## Duplication Analysis

The strongest signal for what should become a skill is content that's duplicated across multiple agents.

### Scripture Link Format — 7 copies

Appears in: `copilot-instructions.md`, `study`, `eval`, `lesson`, `talk`, `review`, `podcast`

Every agent repeats the same link format examples:
```
[Moses 3:5](../gospel-library/eng/scriptures/pgp/moses/3.md)
[D&C 93:36](../gospel-library/eng/scriptures/dc-testament/dc/93.md)
```

This is ~5-10 lines duplicated across 7 files = 35-70 lines of wasted context every session.

### "Search Results Are Pointers" / Cite Count Rule — 5 copies

Appears in: `copilot-instructions.md`, `study`, `dev`, `eval`, plus `yt_evaluation_template.md`

The cite count rule (`for N citations, perform at least N read_file calls`) appears in study and eval explicitly, and is implied in lesson and talk.

### Pre-Publish Checklist — 3 copies

Appears in: `study.agent.md`, `study_template.md`, `yt_evaluation_template.md`

---

## Gap Analysis: What We Should Build

### Priority 1: Agent Skills (High Impact)

Skills load on-demand when relevant, reducing context bloat. Our duplicated content is the prime candidate.

#### 1. `scripture-linking` skill
**What:** Path conventions, link format examples, file verification pattern
**Why:** Currently duplicated in 7 files. Every agent carries this knowledge whether or not it's creating links.
**Description (frontmatter):** `"Format scripture and conference talk links using workspace-relative paths and verify files exist"`
**Properties:** `user-invokable: false` (background knowledge, auto-loaded when the agent is creating links)
**Impact:** Remove ~5-10 lines from each of 7 files. One source of truth.

**Content would include:**
- Path conventions (lowercase with hyphens, no leading zeros, `dc-testament/dc/`)
- Link format examples (scriptures, talks, manuals)
- The "verify files exist with file_search or list_dir" rule
- The "never link to a directory — always the specific file" rule

#### 2. `source-verification` skill
**What:** Two-phase workflow, cite count rule, pre-publish checklist, search-excerpt warnings
**Why:** Currently duplicated across 5 files. This is the single most important quality control in the project.
**Description (frontmatter):** `"Verify scripture and talk quotes against source files — read before quoting, never cite search excerpts"`
**Properties:** `user-invokable: false` (auto-loaded when writing studies, evaluations, or lessons)
**Impact:** Remove the rules/checklist sections from study, eval, and lesson agents. One authoritative checklist.

**Content would include:**
- The two-phase pattern: Discovery (search tools) → Deep Reading (read_file)
- The cite count rule
- Vector search summaries labeled `[AI SUMMARY]` are NOT direct quotes
- Pre-publish checklist
- The discovery-reading-writing quality rhythm

#### 3. `publish-and-commit` skill
**What:** Steps to run the publish pipeline, verify output, stage, commit, and push
**Why:** Done manually every session, same steps every time. Often requested as "please publish and commit."
**Description (frontmatter):** `"Run the publish script to convert study documents to public HTML, then git commit and push"`
**Properties:** `user-invokable: true` (invokable via `/publish-and-commit` slash command)

**Content would include:**
- Run `cd scripts/publish && go run ./cmd/main.go`
- Verify output (check file count, link conversion count)
- `git add -A && git commit -m "..."  && git push`
- Common gotchas (forgot to save, publish script path)

**Resources:**
- Could include a `publish.ps1` helper script in the skill directory

#### 4. `webster-analysis` skill
**What:** When and how to use Webster 1828, what to look for, how to integrate findings
**Why:** Webster 1828 is described as "the model tool" but the guidance is scattered across agents and reflections
**Description (frontmatter):** `"Look up words in Webster 1828 dictionary and analyze how historical meanings illuminate scripture"`
**Properties:** `user-invokable: true` (invokable via `/webster-analysis`)

**Content would include:**
- When to use it: historical meaning differs from modern, Restoration-era vocabulary, words Joseph Smith used
- What to look for: meanings that are narrower/broader/different from modern usage
- How to integrate: quote the definition, then connect to the scriptural context
- Examples from past studies where this was powerful

### Priority 2: Prompt Files (Medium Impact)

Prompt files are manual `/` commands — perfect for one-shot tasks we repeat.

#### 1. `/expound` prompt file
**What:** Convert the existing `docs/work-with-ai/expound-prompt.md` to a proper `.github/prompts/expound.prompt.md`
**Why:** Currently requires copy/paste. As a prompt file, it's one `/expound` command.
**Effort:** Near-zero — the prompt is already written.

**Frontmatter:**
```yaml
---
name: expound
description: Mine the current chat session for teaching examples for the Working with AI lesson series
agent: agent
---
```

#### 2. `/new-study` prompt file
**What:** Scaffold a new study document from template
**Why:** Currently you have to remember the template exists and manually adapt it
**Argument-hint:** `[topic]`

**Would:**
- Read `docs/study_template.md` for structure
- Ask about the topic and starting scriptures
- Create the file at `study/{topic-slug}.md` with metadata pre-filled

#### 3. `/new-lesson` prompt file
**What:** Scaffold a new lesson from template
**Argument-hint:** `[topic] [class] [date]`

#### 4. `/new-eval` prompt file
**What:** Scaffold a YouTube video evaluation
**Argument-hint:** `[youtube-url]`

**Would:**
- Download the transcript via yt-mcp
- Read `docs/yt_evaluation_template.md` for structure
- Create the file at `study/yt/{video-id}-{slug}.md` with transcript reference pre-filled

#### 5. `/reflect` prompt file
**What:** Start a periodic review — how are our interactions going?
**Why:** The "seventh day" practice from our lesson series, but applied to *our own* process

**Would:**
- Read recent entries from `/journal/` and `/study/`
- Read `docs/01_reflections.md` and `docs/04_observations.md` for past context
- Prompt: "How are our interactions going? What's working? What's not? What should we change?"

### Priority 3: File-Based Instructions (Low-Medium Impact)

File-based instructions auto-apply when working with certain file types.

#### 1. Go conventions — `scripts/**/*.go`
**What:** Go workspace patterns, MCP server structure, vet/build/test commands
**Why:** Currently embedded in `dev.agent.md` but should apply even when not in dev mode (e.g., quick bug fix in study mode)
**File:** `.github/instructions/go.instructions.md` or any `*.instructions.md` with `applyTo` frontmatter

**Content:**
- `go.work` multi-module management
- MCP server entry point pattern: `cmd/server/main.go` + `mcp.go`
- Run `go vet ./...` and `go build ./...` before committing
- Test with `go test ./...`

#### 2. Study document conventions — `study/**/*.md`
**What:** Link formatting, footnote expectations, cross-referencing standards
**Why:** Should apply regardless of which agent is active
**Note:** Overlaps with the `scripture-linking` and `source-verification` skills. May not be needed if those skills work well.

### Priority 4: Hooks (Exploratory)

Hooks are powerful but require care. Start with low-risk, high-value use cases.

#### 1. `SessionStart` — Inject project context
**What:** On session start, provide today's date, current git branch, recent study topics
**Why:** Saves the "what are we working on?" dance at the start of sessions
**Risk:** Low — read-only, additive context

**Example `.github/hooks/session-start.json`:**
```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "windows": "powershell -File .github/hooks/scripts/session-context.ps1",
        "timeout": 10
      }
    ]
  }
}
```

The script would output:
```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Date: 2026-02-17 | Branch: main | Recent studies: know-god, charity | Recent journal: 2026-02-08"
  }
}
```

#### 2. `Stop` — Publish reminder
**What:** When the agent stops, check if any study/lesson files were modified and remind to publish
**Why:** We often forget the publish step until explicitly asked
**Risk:** Medium — uses a premium request if it blocks. Better to just add `additionalContext` without blocking.
**Implementation:** Check `git status --short` for changes in `study/`, `lessons/`, `docs/work-with-ai/`. If changes found, add a system message: "There are uncommitted study/lesson changes. Consider running /publish-and-commit."
**Note:** Don't use `decision: "block"` — just remind via `additionalContext` or `systemMessage`.

#### 3. `PreCompact` — Preserve study context
**What:** Before the context window compacts, save key study context (current topic, sources read, citations made) to a summary file
**Why:** Long study sessions hit context limits. Losing track of what we've already read causes repetition.
**Risk:** Medium — needs careful design. Could save to a temp file that gets injected back.

---

## What NOT To Do

### Don't move agent core identity into skills
The study agent's two-phase workflow is its *identity*, not just a capability. Moving it to a skill would mean it might not load. Keep core agent workflows in the agent file and extract only the *shared* components (linking, verification) into skills.

### Don't over-hook
Hooks execute shell commands with your user permissions. Start with read-only hooks (SessionStart context injection) before adding any that modify files or block operations.

### Don't duplicate between skills and file-based instructions
If `scripture-linking` is a skill, you don't also need a `"study/**/*.md"` file-based instruction that says the same thing. Pick one mechanism per piece of knowledge.

### Don't create skills for rare workflows
Skills should be for things that happen regularly. The podcast agent's episode structure isn't worth a skill — it only applies in podcast mode, which already has its own agent. Skills are for *cross-cutting* capabilities used by multiple agents.

---

## Implementation Plan

### Phase 1: Quick Wins (1 session)
1. **Create `/expound` prompt file** — literally move existing content to `.github/prompts/expound.prompt.md`
2. **Create `scripture-linking` skill** — extract from 7 agent files, replace duplicated sections with "See the `scripture-linking` skill"
3. **Create `source-verification` skill** — extract from 5 files

### Phase 2: Prompt Files (1 session)
4. **Create `/new-study` prompt file** — scaffold from template
5. **Create `/new-eval` prompt file** — scaffold with yt-mcp integration
6. **Create `/new-lesson` prompt file** — scaffold from template

### Phase 3: Operational Skills (1 session)
7. **Create `publish-and-commit` skill** — publish pipeline as a slash command
8. **Create `webster-analysis` skill** — historical word study guidance

### Phase 4: Hooks (experimental, 1 session)
9. **Create `SessionStart` hook** — inject date, branch, recent work
10. **Create `Stop` hook** — publish reminder (non-blocking)

### Phase 5: Cleanup
11. **Slim down agent files** — remove duplicated content that's now in skills
12. **Update `copilot-instructions.md`** — reference skills instead of inline rules
13. **File-based Go instructions** — extract from dev agent if needed

---

## Measuring Success

After implementation, watch for:

| Signal | What It Means |
|--------|---------------|
| Agent files get shorter | Duplication removed, skills carrying the load |
| Context window lasts longer | Less always-on content = more room for actual work |
| `/publish-and-commit` used naturally | Operational workflow automated |
| Citation quality stays high | Source verification skill loading correctly |
| Fewer "forgot to publish" moments | Stop hook working |
| Session start is smoother | SessionStart hook providing context |

---

## References

- [VS Code Agent Skills docs](https://code.visualstudio.com/docs/copilot/customization/agent-skills)
- [VS Code Prompt Files docs](https://code.visualstudio.com/docs/copilot/customization/prompt-files)
- [VS Code Hooks docs](https://code.visualstudio.com/docs/copilot/customization/hooks)
- [VS Code Customization Overview](https://code.visualstudio.com/docs/copilot/copilot-customization)
- [Agent Skills Standard](https://agentskills.io/)
- [Jerry Nixon's example](https://github.com/JerryNixworkemail/copilot-agent-skills-azure-sql) — "Instructions set the rules. Skills do the work."
- [Awesome Copilot - community skills, agents, prompts](https://github.com/github/awesome-copilot)
- [Anthropic reference skills](https://github.com/anthropics/skills)

---

*This is a living document. Update as we implement and learn what works.*
