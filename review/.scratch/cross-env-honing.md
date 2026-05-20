# Cross-Environment Instruction Honing — Scratch File

## Intent Check

**Purpose:** Create a clear separation between environment-agnostic instruction *principles* and environment-specific *implementations*, so the core system can be used across Copilot, Claude Code, OpenCode, and future AI environments without duplicating or diluting the theological/collaborative framework.

**Beneficiary:** Michael, when setting up a new AI environment, when switching between Copilot and Claude Code on different machines, or when sharing the instruction pattern with others. Success looks like: opening a new project in a new AI tool and having the *character* of the collaboration (covenant, council, Abinadi) transfer immediately, while the *mechanics* (tool names, agent invocation, file paths) adapt cleanly.

**Success criteria:**
- A review of the instruction architecture exists in `review/instruction-system-review.md` ✅
- A proposal document exists that maps each instruction layer to its environment dependencies
- At least one agent was engaged to stress-test the generalization
- The existing `.claude/` and `.github/` divergence pattern is honored, not erased

**Non-goals (explicitly out of scope):**
- NOT this session: rewriting all instructions into a single shared format. The divergence between `.claude/` and `.github/` is deliberate and useful.
- NOT this session: solving tool-naming once and for all. That's an ecosystem problem that each platform will handle differently.
- NOT this session: moving fiction skills out of the repo. That may be needed but it's a different task.
- NOT this session: building a converter script or automation. This is about *clarity* (what is core vs env-specific), not about *automation*.

## Council Moment

**Connections:** I found 3 specific files/refs that bear on this:
- `.spec/proposals/claude-code-integration.md` — already proposed a `AgentBackend` abstraction interface for brain.exe. The same abstraction principle applies to instructions: define an interface, implement per-backend.
- `.spec/proposals/archive/memory-architecture.md` line 27, 42 — explicitly states "Build portable, not platform-dependent" and rates YAML-in-git as "Excellent" for portability. This is prior art we should build on.
- `CLAUDE.md` itself is an env-adapter layer over `.github/copilot-instructions.md`. It's the working example of what we're trying to generalize.
- `docs/09_post-skills-quality-review-followup.md` identified "copilot-instructions.md still has overlap" with skills, adding token cost to every request. Generalization is a chance to trim that overlap.

**Tensions:**
- The system is heavily tuned for Opus 4.7 literalism. Generalizing to "any model" would lose that precision. However, the core principles (read before quoting, binding question, critical analysis) are model-agnostic.
- `.claude/skills/` and `.github/skills/` are explicitly independent copies allowed to drift. A generalized framework must preserve that drift allowance — convergence is not the goal.
- The "hard gate" pattern works because the environment supports `Read` and file-writing. Some environments (Copilot chat, web UIs) don't have the same tool access. The *principle* (externalize memory) transfers; the *mechanism* (scratch file in markdown) may need adaptation.
- Tool-naming fragility is the #1 operational toil. Any generalized framework that doesn't address this will inherit the pain.

**Blind spots:**
- What's the actual goal? Is this about making a template for new projects? Making the existing project portable across Michael's machines? Or defining a standard that could be adopted by other AI users (OpenCode, Roo, Cline, Continue.dev)? The answer changes the shape of the output.
- Who else would use this? If it's just Michael, portability means "works on my other machine." If it's meant to be shared, portability means "works for anyone with any AI tool."
- The fiction skills (`believable-villains`, `emotional-resonance`, etc.) — are they part of the "scripture study" system or a separate creative-writing system that shares the same repo? If the latter, they shouldn't be in the generalized framework at all.
- Do we need an "instruction manifest" — a single file that declares what skills/agents exist, what envs they work in, and how to load them? That would be a new artifact, not a reorganization.

## Proposed Approach

1. Audit the full instruction stack and tag each component as **CORE** (env-agnostic), **ADAPTER** (env-specific translation), or **LOCAL** (this-project-only).
2. Produce a proposal doc in `.spec/proposals/` or `review/` mapping the stack.
3. Engage the `plan` agent (or its principles) to structure the generalization proposal.
4. Update the review doc with the generalization findings.

## Next Step

Present intent-check and council-moment to user, then proceed with the audit/engagement.
