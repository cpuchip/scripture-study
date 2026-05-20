# Cross-Environment Generalization Stress Test

*Date: May 19, 2026*  
*Reviewer: OpenCode (Claude Code instance)*  
*Scope: The CORE/ADAPTER/LOCAL architecture proposed in `cross-env-instruction-audit.md`*

---

## Executive Summary

The generalization proposal is **directionally correct but optimistically scoped**. The three-category architecture (CORE/ADAPTER/LOCAL) is sound in principle. In practice, the boundary between CORE and ADAPTER is **muddy, not clean** — and the audit's claim that "most skills are 80-90% CORE" understates the contamination by a significant margin.

The proposal is not flawed enough to abandon. But it is flawed enough to **shrink the first prototype dramatically**. The 2-day validation should not be `source-verification` or `study-workflow` — it should be `intent-check` + `critical-analysis` + a deliberately stripped `source-verification-core` that contains *only* the principle layer, not the operational layer. If that three-skill bundle transfers cleanly between Claude Code and Copilot, the architecture is viable. If it doesn't, we've learned that cheaply.

---

## 1. Is the CORE/ADAPTER/LOCAL Split Actually Workable?

### Short Answer: It's a gradient, not a boundary.

The audit presents the categories as discrete. They are not. After reading the actual skill files, the picture is more contaminated than the audit admits.

### What the audit got right

Some skills genuinely are pure CORE:

- **`intent-check`** — Four questions, no tool names, no paths, no model assumptions. It would transfer to ChatGPT, Gemini, or a typewriter.
- **`ben-test`** — Calibrated language, evidence levels, practice-vs-principle. Pure methodology.
- **`discernment-rubric`** — Six text-checkable properties. No tooling dependencies.

These validate the *concept* of a transferable core. But they are the minority.

### What the audit understated

**`source-verification` is not "mostly core with 3-5 tool name references."** It is a tool operation manual disguised as a principle document. Every phase description names specific tools:

- Phase 1: `mcp__gospel-engine-v2__gospel_search`, `mcp__webster__define`, `mcp__exa-search__web_search_exa`
- Phase 2: `Read` on `gospel-library/`, `mcp__byu-citations__byu_citations`, `mcp__gospel-engine-v2__gospel_get`
- YouTube section: `mcp__yt__yt_download`
- Pre-publish checklist: `Glob` or `Bash ls`
- The cite-count rule defines itself in terms of `Read` calls and specific MCP tool names

Removing all tool names from `source-verification` would leave a skeleton: "verify quotes against sources" and "count your citations." The *how* — which is 70% of the skill's value — lives in the tool layer. The audit's claim that "a generalized system could keep the skills as CORE documents and inject tool names via an environment config" works for `intent-check` but barely works for `source-verification`. You would need to inject not just names, but entire workflow steps.

**`council-moment` is similarly contaminated.** The Three Scans section says:

> - `Grep` on key terms in `study/`, `lessons/`, `docs/`, `becoming/`, `.spec/proposals/`
> - `mcp__gospel-engine-v2__gospel_search` on the binding question (semantic mode)
> - `Glob` for related filenames

That's three tool references in the first three bullets. The skill also ties explicitly to `agent_commits_to.check_existing_work` and the Abraham 4:26 council — both core *concepts*, but the operationalization assumes file-search tooling.

**`critical-analysis` was tagged "Pure CORE" but isn't.** Its "Tool check" section says:

> If you have not run `mcp__gospel-engine-v2__gospel_search` (semantic or hybrid mode) on the binding question this session, do it now. This is the single most common cause of missing voices on Opus 4.7...

One tool name + one model name (`Opus 4.7`) in a "pure CORE" skill. Small, but it proves the gradient: even the skills that *look* clean have adapter fingerprints.

**`becoming` contains Michael's personal north star.** The skill explicitly states: "The user's personal north star is: *'I seek to always point to the savior who made all things, and all things testify of him.'*" This is not a transferable principle. It is personal theology tied to a specific user. The audit tagged this as "Core principle, local paths/context" — but the north star isn't a path. It's identity. It should be LOCAL.

**`sabbath-close` assumes git.** "Use `git status` and `git diff --stat` to ground the inventory in reality." It also references specific file paths (`.spec/journal/`, `.mind/active.md`, `scripts/session-journal/journal.go`). This is mostly LOCAL, not CORE.

**`quote-log` references `Read` and `study/.scratch/{topic}.md`.** The path is local; the tool reference is adapter.

### The gradient, visualized

| Skill | Audit's Estimate | Actual Assessment |
|-------|------------------|-------------------|
| `intent-check` | Pure CORE | ✅ Pure CORE |
| `ben-test` | Pure CORE | ✅ Pure CORE |
| `critical-analysis` | Pure CORE | ⚠️ 95% CORE (one tool + model ref) |
| `council-moment` | "Minor tool refs" | ⚠️ 60% CORE (tool names are the scans) |
| `study-workflow` | "Mostly core" | ⚠️ 50% CORE (phases are tool-dependent) |
| `source-verification` | "Mostly core, 3-5 tool refs" | ❌ 30% CORE (it's a tool manual) |
| `becoming` | "Core principle, local paths" | ❌ 40% CORE (north star is personal, not principle) |
| `sabbath-close` | Pure CORE | ❌ 20% CORE (git, paths, memory architecture) |

### Verdict on workability

The split is **workable for principle-level skills** (`intent-check`, `ben-test`, `discernment-rubric`) and **workable in theory for operational skills** if we accept that the CORE version will be significantly thinner than the current skills. The current `source-verification` skill cannot become a CORE document without losing most of its procedural value. It would need to become a *principles* document, with a separate *procedures* document per environment.

---

## 2. What Would Break When Trying to Implement This?

### Tool names are woven into prose, not just frontmatter

The audit's architecture diagram shows:

```
core/
  source-verification.md  # "Discovery→Reading→Writing→Becoming, cite count"
env/
  claude-code/
    skills/             # "Claude Code copies with tool name adjustments"
```

This assumes you can take `source-verification.md`, do a find-and-replace on tool names, and get a valid Claude Code skill. You cannot. The skill's *logic* is tool-shaped:

- "Use `gospel_search` (semantic mode) for discovery" — this assumes a tool that supports semantic search.
- "Use `Read` to verify" — this assumes a file-reading tool with access to local markdown.
- "Download the transcript first (`mcp__yt__yt_download`)" — this assumes a YouTube MCP server.

If an environment lacks semantic search, or lacks local files, or lacks YouTube tools, the adapted skill becomes a broken instruction: "Do this thing you cannot do."

**The hard case:** What if an environment has *analogous* tools but not identical ones? Say a environment has a scripture search API but no semantic mode, only keyword. The CORE skill says "run a semantic search on the binding question." The adapter can only map to keyword search. Does the adapter silently degrade? Does it insert a warning? The architecture doesn't specify the contract between CORE and ADAPTER when capabilities don't map 1:1.

### The hard gate degrades poorly in chat-only environments

The study workflow's hard gate is structural, not rhetorical:

> "If that section is not in the file, the work did not happen — regardless of what was said in chat."

This assumes:
1. The agent can write files.
2. The agent can read files back.
3. The file persists across the session.

In a chat-only UI (ChatGPT web interface, Claude.ai web interface, most mobile AI apps), none of these are true. The entire workflow collapses because its verification mechanism is file-based.

**How would it degrade?** The audit mentions a "light mode" but doesn't specify what a chat-only study workflow looks like. Options:
- The agent writes the scratch file content into the chat thread (but chat threads compact too, and the "file" is not inspectable by the user in the same way).
- The agent holds the scratch in its context window (but this is exactly what the workflow was designed to avoid).
- The workflow is simply unavailable in chat-only environments.

None of these are graceful degradation. They're abandonment or compromise of the core principle.

**This is not a minor issue.** The hard gate is the architectural signature of this instruction system. If it can't survive environment variation, the generalization is cosmetic, not structural.

### The Abinadi persona and theological framework assume large context

Load the study agent in full and count what the model must hold simultaneously:

- `CLAUDE.md` / `copilot-instructions.md` (~180 lines)
- `covenant.yaml` (273 lines)
- `identity.md` (36 lines)
- `principles.md` (~hundreds of lines)
- The full study agent file (165-176 lines)
- Multiple skills loaded during the session
- The actual study content being researched and written

That's easily **4,000-6,000 tokens of instruction before any work begins.**

On an 8K-context model (older GPT-4, some local models), this consumes 50-75% of the context window before the first source is read. The entire architecture — externalized memory, scratch files, quote logs — was built *because* context compacts. But it was built assuming the model has *enough* context to hold the instructions and some working memory simultaneously.

**What breaks on 8K context?**
- The agent may not be able to load both the identity layer and the skill layer in the same session.
- The Abinadi persona (4 dense paragraphs of theological framing) may need to be compressed to a single sentence.
- The phased workflow may need to be simplified to 2-3 phases instead of 7.
- The "read the scratch file in full" step may fail because the scratch file + instructions exceed context.

The audit notes that the theological framework "transferred immediately" across the Claude Code/Copilot boundary. That's true — but both environments use Opus 4.7 or similar large-context models. The transfer test to a small-context environment has not been performed.

### Model-specific voice rules may not translate

The writing voice rules (em-dash budget, therefore/but audit, cut list) were tuned for Opus 4.7 literalism. The instructions explicitly say:

> "Per Anthropic's 4.7 migration guide, this model uses tools less by default — you have to explicitly reach for them."
> "Per Anthropic's 4.7 guidance, positive examples shape voice better than negative rules."

What happens when these rules load into GPT-4o? Or Gemini? Or a local Llama model?

- **GPT-4o** doesn't have Opus 4.7's literalism problem. It may interpret "therefore/but audit" as a rigid formula and produce mechanical, formulaic prose — the opposite of the intended effect.
- **Gemini** has different tool conventions and different default voice characteristics. The cut list ("let that land," "sit with that") may be completely irrelevant because Gemini doesn't default to that vocabulary.
- **Local models** may lack the instruction-following fidelity to enforce mechanical rules at all, making the voice audit a waste of context.

The audit tags writing voice as CORE. But it's actually **model-specific corrective guidance** — closer to ADAPTER than CORE.

---

## 3. Are There Hidden Dependencies?

### `check_existing_work` requires tools some environments lack

The covenant's `agent_commits_to.check_existing_work` says:

> "Before writing new claims, check what we've already written. Search the `study/` folder, the `docs/` folder, the guide series."

This is tagged as a behavioral principle (CORE). But it is inseparable from file-search capability. If an environment has no `Grep`, `Glob`, `file_search`, or `list_dir` equivalent, the agent literally cannot fulfill this commitment.

**This is a hidden capability dependency.** The audit treats it as "search the corpus" — a semantic action. But the implementation is "run `Grep` on key terms." An environment with only a web-search tool cannot do this.

### Writing voice rules were tuned for Opus 4.7

As noted above, the voice rules are corrective lenses for a specific model's myopia. Transferring them uncritically to other models may:
- Be unnecessary (if the target model doesn't have the problem)
- Be counterproductive (if the target model interprets the rules differently)
- Waste context (if the target model ignores mechanical rules)

The audit's claim that "positive examples shape voice better than negative rules" is itself a model-specific observation. It may not hold for models with different training architectures.

### The "Becoming" north star is personal context, not transferable principle

The `becoming` skill states:

> "The user's personal north star is: *'I seek to always point to the savior who made all things, and all things testify of him.'*"

This is Michael's personal statement. It is not a general principle of scripture study. It should not be in a generalized CORE skill. It should be in a LOCAL user-config file that the skill references.

**The deeper issue:** The `becoming` skill mixes three things:
1. The *principle* that studies should land personally (CORE)
2. The *format* for becoming sections (CORE)
3. Michael's *specific north star* (LOCAL)
4. The `becoming/` directory path (LOCAL)

The audit only separated #4 as local. #3 should be local too.

### `covenant.yaml` and `identity.md` reference specific files

These are tagged as "Pure CORE" but they contain LOCAL anchors:

- `covenant.yaml` references `study/stewardship-pattern.md`, `docs/work-with-ai/guide/05_complete-cycle.md`, `.spec/proposals/teaching-workstream.md`, and dates like "Mar 18 / Mar 19 / Apr 23."
- `identity.md` references `.spec/journal/2026-01-21--project-genesis.yaml`, `docs/04_observations.md`, `docs/biases.md`, and the specific date March 4, 2026.

These references make the files feel *grounded* — and they are, in this project. But in a generalized context, they read like inside jokes. A user in another environment loading this "CORE" identity will wonder: "What is the 'March 22, 2026 reflection'? What is 'the bias wall'?"

The files need two versions: a **canonical CORE** version with generic language, and a **localized** version with the project's specific history.

### Agents reference each other by name

The `.github/agents/study.agent.md` has a `handoffs:` block referencing `journal` and `lesson` agents. This is an architectural dependency on the agent system itself. In an environment without agent handoffs (ChatGPT, standard Claude web), these references are meaningless.

---

## 4. What's the Simplest Version That Proves the Concept?

### Don't prototype with `source-verification`

The audit's action item #5 says: "Prototype one generated skill. Pick `source-verification` or `study-workflow`." This is the hardest possible test. It's like proving a bridge can hold weight by driving a tank across it. If it fails, you won't know if the concept is bad or if the load was too heavy.

### The 2-day version

**Day 1: Extract three genuinely core skills**

1. **`intent-check`** — Already pure. Extract to `instructions/core/intent-check.md` unchanged.
2. **`ben-test`** — Already pure. Extract to `instructions/core/ben-test.md` unchanged.
3. **`critical-analysis`** — Remove the one tool reference (`mcp__gospel-engine-v2__gospel_search`) and the model name (`Opus 4.7`). Replace with: "If your environment has a semantic search capability, run it on the binding question now. Semantic search surfaces non-obvious cross-references that recall does not." Extract to `instructions/core/critical-analysis.md`.

These three prove the *easy* case: principle-only skills can transfer cleanly.

**Day 2: Extract one semi-core skill and test the split**

4. **Create `instructions/core/source-verification-principles.md`** — This is NOT the full skill. It contains ONLY:
   - The Core Rule ("Search results are pointers, not sources")
   - The cite-count rule (abstracted: "For N citations, perform at least N verified source reads")
   - The three levels of attribution (Direct/Paraphrase/Reference)
   - The confabulation warning
   - The pre-publish checklist (without tool-specific items like `Glob` or `Bash ls`)

   Then create:
   - `instructions/env/claude-code/source-verification-procedures.md` — The tool-specific layer: "Use `mcp__gospel-engine-v2__gospel_search` for discovery..."
   - `instructions/env/copilot/source-verification-procedures.md` — The Copilot tool names.

5. **Test the split manually.** Load the CORE + Claude Code adapter in Claude Code. Load the CORE + Copilot adapter in Copilot. Run a tiny study (one binding question, 2-3 sources) in each. Ask:
   - Does the agent follow the cite-count rule?
   - Does it distinguish direct quote from paraphrase?
   - Does it try to verify against source files?

If the split works, the agent should behave similarly in both environments despite different tool names. If it fails, the failure mode tells us whether the problem is in the CORE layer (too vague) or the ADAPTER layer (wrong mapping).

**What NOT to do in the 2-day version:**
- Don't build a generator. Hand-write both adapters. The generator is a Month-2 problem.
- Don't touch `study-workflow` yet. It has the hard-gate dependency on file writing.
- Don't move the fiction or tech skills. They're irrelevant to the generalization question.
- Don't try to abstract the agent files yet. The agent split is harder than the skill split.

### If the prototype fails, the most likely failure modes are:

1. **CORE too vague:** The agent reads "verify quotes against sources" and doesn't know what that means without tool names.
2. **ADAPTER incomplete:** The adapter writer forgets to mention a tool, and the agent skips a verification step.
3. **Context collision:** Loading CORE + ADAPTER + the actual study content exceeds the model's ability to track the workflow.

All three are valuable findings. Any of them tells us the architecture needs adjustment before we commit to full extraction.

---

## 5. What Would Michael Wish We Had Thought Of Before Starting?

### The Adjacent Surface Audit

The audit itself needs an adjacent surface audit. Here's what it misses:

#### Scope: What exactly are we generalizing?

The current `.claude/skills/` directory has 28 skills. The audit mentions:
- 14 "scripture study skills" (the "essential set")
- 5 fiction skills
- 5 tech/dev skills
- Plus agents, memory files, base instructions

Are we generalizing ALL of these? Just the scripture study subset? What about the covenant, identity, and principles files — are they part of the generalized package, or are they Michael-specific?

**The honest answer:** Most users don't need the fiction skills. Many users won't have MCP servers. Some users won't even have a local file system. A "generalized scripture study framework" that includes `pgrx-rust` and `believable-villains` is not generalized — it's this project's instruction set with the names changed.

**Michael would wish we'd scoped "generalized" to mean:** "The minimum set of instructions that would allow a different user, in a different environment, to do disciplined scripture study with AI assistance." That's probably 6-8 skills, not 28.

#### Discoverability: How does the agent know what to load?

The audit's architecture diagram shows a `core/` directory and an `env/` directory. But it doesn't specify:

- How does the agent know it's in a "Claude Code" environment vs a "Copilot" environment?
- Does the user manually configure this?
- Does the environment auto-detect?
- What happens if the environment is something new (OpenCode, local Ollama, a future platform)?

The current system handles this implicitly: Copilot loads `.github/copilot-instructions.md`; Claude Code loads `CLAUDE.md`. The generalization proposal replaces this implicit mechanism with... nothing specified. The user would need to wire up the correct adapter themselves.

**Michael would wish we'd specified the loading mechanism.** Without it, the architecture is a filing system, not a runtime system.

#### Contracts: What does CORE assume about ADAPTER capabilities?

The CORE skills assume certain capabilities exist:
- File reading (`Read` / `read_file`)
- File writing (`write` / `edit` / `create`)
- Semantic search
- Scripture text access (local files or API)
- Dictionary lookup (Webster 1828)

What if an environment has keyword search but not semantic search? What if it has scripture access but no conference talks? What if it has file writing but no persistent storage across sessions?

The architecture needs a **capability manifest** — an explicit list of what CORE requires, so an environment can say "I can provide A, B, and C but not D" and the CORE skills can degrade gracefully.

**Michael would wish we'd defined the CORE-to-ADAPTER contract.** Right now it's implicit and optimistic.

#### Spec gaps: What the audit left unspecified

1. **How do file paths get injected?** The audit says "A generalized `scripture-linking` skill says 'link to the specific chapter file.' It doesn't say `gospel-library/eng/scriptures/bofm/alma/32.md`." But it doesn't specify the injection mechanism. Template variables? Frontmatter? A local config file?

2. **How do we handle the "light mode"?** Multiple files say "for a typo fix, this is overkill," but the audit doesn't specify what a light-mode study workflow looks like. If the full workflow requires file writing, what's the chat-only equivalent?

3. **What about markdown links in CORE documents?** The `council-moment` skill contains a link to `gospel-library/eng/scriptures/pgp/abr/4.md`. If this is a CORE document, should it contain links to local files? Or should links also be injected?

4. **How do we keep the two env copies in sync?** The audit says "per-environment skills are generated, not hand-maintained." But there's no generator specified. Hand-maintaining two copies of 14 skills is how drift happens — and the audit already notes drift between `.github/agents/dev.agent.md` and `.claude/agents/dev.md`.

5. **What about model-specific guidance that isn't tool-related?** The Opus 4.7 literalism notes are model-specific but not tool-specific. Are they ADAPTER or LOCAL? The audit puts them in `local/model-guidance.md`, but that means every environment using Opus 4.7 (whether Copilot or Claude Code) would duplicate them.

---

## Bottom Line

The CORE/ADAPTER/LOCAL architecture is **sound in theory and premature in scope.** The audit correctly identified the *direction* (principle vs. mechanism vs. context) but underestimated how deeply the mechanism is woven into the principle.

**The honest assessment:**

- **Start with 3-4 truly core skills.** `intent-check`, `ben-test`, `critical-analysis` (with one edit), and a deliberately thin `source-verification-principles`. Prove these transfer.
- **Accept that `study-workflow` and `source-verification` cannot be cleanly split** without losing their operational value. They may need to stay as per-environment documents that *reference* core principles, rather than being generated from a core template.
- **Define the loading mechanism and capability contract** before building more of the architecture. Without these, the `instructions/` directory is a taxonomy, not a system.
- **Scope "generalized" narrowly.** A transferable scripture study framework is 6-8 skills, not 28. The fiction skills, tech skills, and project-specific memory files stay here.

The system is beautiful. It is also, as the original review noted, at risk of "process bloat becoming its own idol." Generalization is another process. We should build it with the same skepticism we apply to our own studies: check the strongest claims, find the weakest links, and ask whether we're reading to discover or to confirm.

---

*In honor of Ben: perhaps too complimentary? Perhaps. But better honest than enthusiastic.*
