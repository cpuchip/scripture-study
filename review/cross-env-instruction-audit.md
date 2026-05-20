# Cross-Environment Instruction Audit

*Date: May 19, 2026*  
*Scope: Full instruction stack — `.github/`, `.claude/`, `.mind/`, `.spec/`*  
*Produced by: OpenCode (Claude Code instance) with user direction*

---

## The Three Categories

| Tag | Meaning | Transferable? |
|-----|---------|---------------|
| **CORE** | Theological, methodological, or relational principles. Model-agnostic, tool-agnostic, environment-agnostic. | Yes — anywhere |
| **ADAPTER** | Environment-specific mechanics: tool names, invocation syntax, file paths, model-specific guidance. | Yes — but must be re-implemented per env |
| **LOCAL** | Project-specific or workstream-specific content. Tied to this repo, this codebase, this user's context. | No — stays here |

---

## Base Instructions

### `.github/copilot-instructions.md` — MIXED (Core + Adapter)

**CORE layers** (transferable):
- "Who We Are Together" — theological framework, warmth, depth, faith-as-framework
- "Covenant" — bilateral commitment pattern
- "Core Principles" — read before quoting, verify numbers/dates, paraphrase when unverified, link everything, prefer local copies
- "Writing Voice" — therefore/but, em-dash budget, cut list, no meta-narration, no closing refrain
- "Adjacent Surface Audit" — scope, discoverability, contracts, spec gaps
- "Inverse hypothesis" (Agans Rule 9)

**ADAPTER layers** (needs translation per env):
- Tool naming tables: `mcp_gospel-engine_gospel_search`, `read_file`, `grep_search`, `file_search` / `list_dir`
- `includeIgnoredFiles: true` flag for Copilot's `grep_search`
- "This project uses custom agents (`.github/agents/`)" — agent invocation mechanism
- Agent mode table (16 agents) — these are real but the *selection mechanism* is Copilot-specific
- "Model context (2026-04 onward): GitHub Copilot now runs on Claude Opus 4.7" — model-specific guidance

**LOCAL layers** (this repo only):
- Project structure table — paths like `/gospel-library/eng/scriptures/`, `/study/`, `/becoming/`
- "7 MCP servers configured in `.vscode/mcp.json`"
- Specific file references: `biases.md`, `work-with-ai/guide/05_complete-cycle.md`

**Verdict:** This file should be *split*. The core layers go into a shared base. The adapter layers become per-environment overlays. The local layers stay in a project-specific file.

---

### `CLAUDE.md` — ADAPTER (by design)

This file exists precisely because the base instructions are written for Copilot and need translation. It is the working model of what an adapter layer looks like.

**Contents:**
- Tool naming translation table (`mcp_gospel-engine_gospel_search` → `mcp__gospel-engine-v2__gospel_search`)
- Tool name mapping (`read_file` → `Read`, `grep_search` → `Grep`)
- `.gitignore` handling difference (`includeIgnoredFiles: true` → use `Bash`/`ls`)
- Subagent system explanation (`.claude/agents/*.md` vs `.github/agents/*.agent.md`)
- Skills divergence policy (`.claude/skills/` and `.github/skills/` are independent copies)
- Slash command translation (`${input:foo}` → `$ARGUMENTS`)

**Verdict:** This is the *correct* pattern. Every environment needs its own `CLAUDE.md`-equivalent. The question is whether we can *generate* these from a canonical mapping rather than hand-maintaining them.

---

## Memory / Identity

### `.spec/covenant.yaml` — CORE

Bilateral covenant between human and AI. Model-agnostic. Environment-agnostic. The commitments (`read_fully`, `provide_binding_question`, `flag_when_wrong`, `not_bypass_process`, `read_before_quoting`, `check_existing_work`, `exercise_stewardship`) are behavioral principles, not tool instructions.

**Verdict:** Pure CORE. Transfer as-is to any environment.

---

### `.mind/identity.md` — CORE

Relational identity of the collaboration. Theological framework (Abraham 4-5 pattern). Posture (warmth, honest exploration, depth, trust). Known bias patterns (safety-posture coldness, instruction-compliance flatness, tool-familiarity defaults, finding-without-reading).

**Verdict:** Pure CORE. This is the *character* layer. Transfer as-is.

---

### `.mind/principles.md` — MOSTLY CORE, some LOCAL

**CORE:**
- Theological framework (Matter Spectrum, Atonement as Refinement, Prophetic Theologizing, Epistemic Humility, Becoming Is the Point, Gospel Pattern Is Eternal, Consumption vs Consecration, Love vs Blessings, Zion Is Built Daily)
- Study methodology (Two-Phase Workflow, Finding vs Reading, Webster 1828 as Model Tool, Follow the Footnotes, Binding Questions and Ring Checks)
- Collaboration principles (Intent Over Instruction, Instruction Minimalism, Compression Is Curation, Stability After Improvement, Organize Before Building, Abraham 4-5 Pattern, Judges Not Executors)
- Writing craft (Therefore/But, Monson Principle, Omission Earns Weight, Ma in the Writing, Voice Discipline)
- Agentic architecture (Harness > Intelligence, Convergent Pattern, Reduce Before Adding, Ben Test, Verification Is the Surviving Skill, Gated Autonomy, Cost Unit)

**LOCAL:**
- Tool Selection Heuristics table — references `gospel-mcp`, `gospel-vec`, `webster-mcp`, `byu-citations`, `yt-mcp`, `becoming-mcp`. The *heuristic* (when to use keyword vs semantic) is core. The *tool names* are local.
- "1500 premium requests/month" — specific to GitHub Copilot billing.

**Verdict:** Split the tool selection heuristics into a separate LOCAL file. The rest is CORE.

---

### `.mind/active.md`, `.mind/preferences.yaml`, `.mind/decisions.md` — LOCAL

Current state, personal context, project decisions. Tied to this user, this moment, this codebase.

**Verdict:** LOCAL. Not transferable.

---

## Agents

### `.github/agents/study.agent.md` vs `.claude/agents/study.md`

These are the same agent implemented for two environments. Comparing them reveals the core/adapter boundary.

**CORE (shared identity & workflow):**
- Abinadi persona and four principles (read to Christ, answer binding question, deliver whole message, write it down)
- "Who We Are Together" section (faith, warmth, depth, trust)
- Phased workflow: Outline → Source Gathering → Gap Analysis → Critical Analysis → First Draft → Review → Becoming → Clean Up
- Binding question discipline
- Ring check / posture check
- Hard gate on critical analysis
- Voice audit rules
- Progress updates between phases

**ADAPTER (differs per env):**
- `.github`: `tools: [vscode, execute, read, agent, edit, search, web, browser, ...]` — Copilot tool registry
- `.github`: `handoffs:` block with `label`, `agent`, `prompt`, `send` — Copilot-specific agent handoff
- `.github`: `gospel_search` (no server prefix) — Copilot deferred-tool naming
- `.github`: `read_file` — Copilot naming
- `.claude`: `model: opus` — Claude Code model selection
- `.claude`: No tool registry (Claude Code doesn't use frontmatter tool lists)
- `.claude`: `mcp__gospel-engine-v2__gospel_search` — full MCP naming
- `.claude`: `Read` — Claude Code tool name

**LOCAL:**
- Both reference `study/{topic}.md`, `study/.scratch/{topic}.md`, `quote-log` skill — project-specific paths
- Reference to `study/what-abides.md` — this project's study history

**Verdict:** The core agent (identity, workflow, hard gates) should be written once and *referenced*, not duplicated. The adapter layer (tool names, invocation syntax) should be per-environment. The local layer (file paths) should be in a project config.

---

### `.github/agents/dev.agent.md` vs `.claude/agents/dev.md`

Same pattern. The core is stewardship, foresight, adjacent surface audit. The adapter is tool names and project paths. The local is brain.exe / ibeco.me / becoming specific architecture.

**Notable:** The `.claude/agents/dev.md` is significantly more detailed (289 lines vs likely shorter `.github` version). This is the "drift" that `CLAUDE.md` explicitly allows.

---

### Fiction agents (`fiction`, `story`, `storytime`) — LOCAL

Creative writing agents. Part of a different workstream. Not core to scripture study.

**Verdict:** LOCAL. Keep them, but don't include in a generalized scripture-study framework.

---

## Skills

### Scripture Study Skills (the essential set)

| Skill | Core | Adapter | Local | Notes |
|-------|------|---------|-------|-------|
| `study-workflow` | ✅ Phases, binding question, hard gate | ⚠️ `Read`, `gospel_search` naming | ✅ File paths (`study/.scratch/`) | Mostly core |
| `source-verification` | ✅ Discovery→Reading→Writing→Becoming, cite count, quote hygiene | ⚠️ `Read`, `gospel_search`, `gospel_get`, `webster_define`, `byu_citations` | ✅ File paths (`gospel-library/`) | Mostly core |
| `quote-log` | ✅ Format, externalize as you read | — | ✅ `study/.scratch/` path | Core principle, local path |
| `scripture-linking` | ✅ Link conventions, always specific file | — | ✅ Path patterns (`gospel-library/eng/...`) | Core principle, local path pattern |
| `becoming` | ✅ Bridge to life, specific commitments, north star | — | ✅ `becoming/` directory, user's north star | Core principle, local paths/context |
| `council-moment` | ✅ Three scans, connections/tensions/blind spots | ⚠️ `Grep`, `gospel_search`, `Glob` | — | Core principle, minor tool refs |
| `intent-check` | ✅ Four questions, purpose/beneficiary/criteria/non-goals | — | — | Pure CORE |
| `critical-analysis` | ✅ Stress-test claims, weakest links, missing voices | — | — | Pure CORE |
| `deep-reading` | ✅ Read full chapters, follow footnotes, keep intermediate findings | ⚠️ `read_file` | — | Core principle, minor tool ref |
| `wide-search` | ✅ Broad net, semantic + keyword, cross-volume | ⚠️ `gospel_search` | — | Core principle, minor tool ref |
| `webster-analysis` | ✅ Historical meaning illuminates scripture | ⚠️ `webster_define` | — | Core principle, minor tool ref |
| `ben-test` | ✅ Calibrated language, practice vs principle | — | — | Pure CORE |
| `reflect` | ✅ Log corrections immediately, graduate at session end | — | ✅ `.spec/scratch/reflect.md` path | Core principle, local path |
| `sabbath-close` | ✅ Declaration + carry-forward | — | — | Pure CORE |
| `publish-and-commit` | ✅ Publish script, git add/commit/push | ⚠️ `publish` command mechanism | ✅ `scripts/publish/` path | Mixed |
| `discernment-rubric` | ✅ Six text-checkable properties | — | — | Pure CORE |

**Key insight:** Most skills are 80-90% CORE. The adapter contamination is usually just 3-5 tool name references. A generalized system could keep the skills as CORE documents and inject tool names via an environment config.

---

### Fiction Skills — LOCAL

| Skill | Verdict |
|-------|---------|
| `believable-villains` | LOCAL — creative writing |
| `character-voice` | LOCAL — creative writing |
| `emotional-resonance` | LOCAL — creative writing |
| `sacrifice-and-loss` | LOCAL — creative writing |
| `worldbuilding-fiction` | LOCAL — creative writing |

**Recommendation:** Move to `.claude/skills/fiction/` or a separate workspace. They compete for attention with study skills.

---

### Tech/Dev Skills — LOCAL

| Skill | Verdict |
|-------|---------|
| `pgrx-rust` | LOCAL — specific to `projects/pg-ai-stewards/` |
| `pgrx-extension-bump` | LOCAL — specific to pgrx projects |
| `mcp-server-go` | LOCAL — specific to Go MCP server in this repo |
| `playwright-cli` | LOCAL — specific to frontend testing in this repo |
| `dokploy` | LOCAL — specific to deployment of this repo |

**Recommendation:** Keep in `.claude/skills/` but mark as project-specific. Or move to `projects/pg-ai-stewards/.claude/skills/` if Claude Code supports subdirectory skills.

---

## Proposals and Meta-Docs

### `.spec/proposals/claude-code-integration.md` — LOCAL (but relevant)

Proposes `AgentBackend` interface for brain.exe. Relevant because the same abstraction principle applies to instructions.

### `.spec/proposals/archive/memory-architecture.md` — CORE (principle)

"Build portable, not platform-dependent." "YAML files in git. Fully portable, model-agnostic, vendor-independent." This is the architectural principle behind the generalization effort.

### `.spec/proposals/deferred/second-brain-architecture.md` — CORE (principle)

"Architecture is portable, tools are not." "Markdown/YAML as source of truth."

---

## The Generalization Architecture

Based on this audit, a cross-environment instruction system would look like this:

```
instructions/
  core/
    identity.md           # Who We Are — theological, relational
    covenant.yaml         # Bilateral commitments
    principles.md         # Methodological, hermeneutical, agentic
    voice.md              # Writing voice rules (therefore/but, cut list, etc.)
    study-workflow.md     # Phased workflow, binding question, hard gates
    source-verification.md # Discovery→Reading→Writing→Becoming, cite count
    quote-log.md          # Scratch file format
    scripture-linking.md  # Link conventions (generic, no paths)
    becoming.md           # Bridge to life, specific commitments
    council-moment.md     # Three scans
    intent-check.md       # Four questions
    critical-analysis.md  # Stress-test checklist
    deep-reading.md       # Read full chapters, follow footnotes
    wide-search.md        # Broad net methodology
    webster-analysis.md   # Historical meaning methodology
    ben-test.md           # Calibrated self-assessment
    reflect.md            # In-session learning capture
    sabbath-close.md      # Ending ritual
    discernment-rubric.md # Non-canonical source evaluation
  env/
    copilot/
      AGENTS.md           # Copilot-specific: tool names, agent frontmatter
      skills/             # Copilot copies with tool name adjustments
    claude-code/
      CLAUDE.md           # Claude Code adapter (what exists today)
      skills/             # Claude Code copies with tool name adjustments
    opencode/
      AGENTS.md           # OpenCode adapter
      skills/             # OpenCode copies
  local/
    project-structure.md  # Paths, directories, repo layout
    tool-registry.md      # Which MCP servers exist, how they're configured
    tool-heuristics.md    # When to use which tool (keyword vs semantic)
    agent-index.md        # Which agents exist, what they do
    model-guidance.md     # Opus 4.7 literalism, etc. — model-specific, env-specific
```

**Key design decisions:**

1. **CORE skills contain no tool names.** They describe *what* to do, not *which function* to call. The environment layer injects tool names.
2. **File paths live in LOCAL, not CORE.** A generalized `scripture-linking` skill says "link to the specific chapter file." It doesn't say `gospel-library/eng/scriptures/bofm/alma/32.md`.
3. **The covenant, identity, and principles are loaded everywhere.** These are the "why" that makes the "how" meaningful.
4. **Per-environment skills are generated, not hand-maintained.** If a core skill says "run a semantic search on the binding question," the Copilot version says "`gospel_search` (semantic mode)" and the Claude Code version says "`mcp__gospel-engine-v2__gospel_search` (mode: 'semantic')". The delta is small enough to automate.
5. **Agents are split into CORE (persona + workflow) and ADAPTER (tool registry + invocation).** The Abinadi persona and phased workflow live once in CORE. The frontmatter tool lists and handoff syntax live in the env layer.

---

## Immediate Action Items

1. **Create `review/core/` directory.** Extract the core layers from existing files into canonical core documents.
2. **Audit `copilot-instructions.md` for overlap.** The follow-up review identified it still duplicates source-verification and scripture-linking. Trimming this reduces token cost AND makes generalization easier.
3. **Tag every skill with its category.** Add a `category: core | local` field to skill frontmatter.
4. **Move fiction skills to a subdirectory.** Reduce noise in the main skill tree.
5. **Prototype one generated skill.** Pick `source-verification` or `study-workflow`. Write the CORE version without tool names. Then write a Copilot adapter and a Claude Code adapter. See if the pattern holds.

---

*Related: [instruction-system-review.md](./instruction-system-review.md) · [docs/09_post-skills-quality-review-followup.md](../docs/09_post-skills-quality-review-followup.md) · [.spec/proposals/claude-code-integration.md](../.spec/proposals/claude-code-integration.md)*
