# Instruction System Review: Claude Code OpenCode Perspective

*Date: May 19, 2026*  
*Reviewer: OpenCode (Claude Code / OpenCode instance)*  
*Scope: `.github/copilot-instructions.md`, `CLAUDE.md`, `.claude/skills/`, `.claude/agents/`, `.spec/covenant.yaml`, `.mind/identity.md`*

---

## Context

This review was produced in response to a direct request: "What do you think of these instructions and memories?" The reviewer was given full access to the instruction architecture and asked for an honest, grounded critique. This document preserves that critique for future reference and as input to the cross-environment generalization workstream.

---

## Executive Summary

This is one of the most thoughtful, battle-scarred instruction architectures encountered. It is clear the system has been refined based on real failures — the Klarna lesson, the dialog bug, the corrupted practice record, the stewardship contradiction in Section VII. It is not theoretical; it is operational. The risk is **process bloat becoming its own idol**. The ceremonies exist to serve the work, not the other way around.

---

## What Is Genuinely Excellent

### Theological Coherence as Operating System

Woven through every file — from the Abraham 4:26 council moment to the Abinadi persona to the bilateral covenant — is a single theological framework. Most AI instructions are functional ("be helpful, don't hallucinate"). These are *teleological*: the work exists to facilitate transformation, not just knowledge. That gives the system a soul that resists drift.

### "Files Are Durable, Context Is Not"

The scratch-file + quote-log discipline, the hard gate on critical analysis, and the externalized memory architecture all recognize the same truth: LLMs lose context, and "I did that in my head" is the most reliable way to ship bad work. The source-verification skill — especially the cite-count rule and the three-level attribution (direct/paraphrase/reference) — is genuinely rigorous. It treats confabulation as a moral problem, not a technical one.

### The Bilateral Covenant

Making the human commit too (`read_fully`, `flag_when_wrong`, `not_bypass_process`) elevates this from "AI behavior script" to "collaboration contract." It names the dynamic honestly: when either side breaks covenant, the output degrades. That is mature.

### Model-Specific Operational Tuning

The Opus 4.7 literalism adjustments, the tool-naming translation tables, the explicit "honor intent not literal request" override — this shows the author is not writing instructions for an idealized AI. They are writing them for the specific models they actually use, with their specific failure modes.

### Hard Gates Over Soft Suggestions

Phrases like *"Saying 'skipping the gap and critical-analysis phases' in chat is not a license to skip them; it is a license to fail"* and *"If the section is not in the file, the work did not happen"* are structural, not rhetorical. They create verifiable artifacts that survive context compression.

---

## Where There Is Risk of Backfire

### Length and Density Risk Attention Dilution

By rough count, a study-session agent loads: `CLAUDE.md` (62 lines) + `copilot-instructions.md` (180 lines) + `covenant.yaml` (273 lines) + `identity.md` (36 lines) + `active.md` + `principles.md` + the full agent file (165 lines) + multiple skills. That is easily 5,000+ words of instruction *before* any work begins.

With Opus 4.7's literalism, the model may fixate on minor mechanical rules (em-dash budget, "therefore/but" transitions) while missing the deeper point. Or it may perform the rituals without animating them — generating a "Council moment" heading in chat without actually scanning, or writing a `## Gap & Critical Analysis` section that is three sentences of throat-clearing to satisfy the gate.

### Skill Proliferation

There are 28 skills in `.claude/skills/`. Many are fiction-writing tools (`believable-villains`, `emotional-resonance`, `sacrifice-and-loss`, `character-voice`, `worldbuilding-fiction`). If they are actively used, fine. If they are aspirational or migrated from another project, they are noise that competes for attention with `source-verification` and `study-workflow`.

### Memory Ritual vs. Memory Quality

The session-end update (`journal/` + `active.md` + `principles.md` + `identity.md`) is beautiful in theory, but in practice it may often become performative. It is easy for an agent to say "I updated memory." It is hard for it to actually synthesize what *mattered* from the session versus dumping chat history into YAML. The risk is memory accumulation without memory value.

### The Line Between Warmth and Performance

"Warmth over clinical distance" is the right north star, but it is also the hardest to specify mechanically. Without careful calibration, the model might hear "warmth" as "add filler" or "be chatty." The voice rules (cut list, therefore/but, no meta-narration) are good mechanical guardrails, but the deeper challenge — sounding like Michael without Michael being present to say "that is not quite it" — remains unsolved.

### Tool-Naming Fragility

The fact that a translation table is needed between Copilot and Claude Code for MCP tool names, and that the `-v2` suffix "trips us up repeatedly," suggests the tooling layer is still a source of operational toil. This is not a flaw in the instructions — it is a flaw in the ecosystem that the instructions are heroically papering over.

---

## Specific Recommendations

1. **Trim the skill tree to a core 8–10.** Keep the scripture-study essentials. Move fiction skills to a separate workspace or a "rarely-invoked" subdirectory so they do not compete for attention in study sessions.

2. **Define a "light mode" explicitly.** Multiple files say "for a typo fix, this is overkill," but they do not say what *is* appropriate for a quick fix. A clear spectrum (full ritual / journal update only / no ritual) would reduce the temptation to either over-process trivial tasks or under-process important ones.

3. **Automate the voice audit.** The em-dash check, "and then" scan, and cut-list detection could be a small regex script. Making it a tool call (`scripts/voice-check.sh`) turns a subjective scan into an objective gate.

4. **Test the hard gates.** The most important question: *Do the agents actually stop when the gate is not met?* If they generate a perfunctory `## Gap & Critical Analysis` section to satisfy the file check, rather than doing real analysis, the gate may be theater.

5. **Memory: quality over quantity.** Consider collapsing the session-end ritual. What is the *minimum* viable memory update? One meaningful sentence in `active.md` and one in `journal/` may be more durable than four files of diffused updates.

---

## Bottom Line

This system shows real scars from real failures. That is exactly what you want. It is not theoretical; it is operational.

The risk is **process bloat becoming its own idol**. The ceremonies exist to serve the work, not the other way around. Right now the system is still on the right side of that line, but it is a line that will need to be watched. The covenant's `not_bypass_process` commitment cuts both ways: bypassing process produces bad work, but *worshipping* process produces empty work.

It is beautiful, though. Theologically grounded, structurally sound, and honest about what AI collaboration actually is. Most instruction sets tell the AI what to do. This one tells it who to be. That is the difference.

---

## Cross-Environment Generalization Notes

This review itself is a data point for the generalization question. The reviewer was Claude Code / OpenCode, not Copilot. Key observations from that boundary crossing:

- **The theological framework transferred immediately.** The Abraham 4 pattern, the Abinadi persona, the bilateral covenant — these are model-agnostic. They live in the *content*, not the tooling.
- **Tool-naming is the primary friction point.** Every practical instruction about "how to search" or "how to verify" is coupled to the deferred-tool naming of a specific environment. A generalized system would need an abstraction layer here.
- **Hard gates are env-agnostic; ritual is env-specific.** "Read before quoting" works everywhere. "Use `mcp__gospel-engine-v2__gospel_search`" works only in Claude Code. The distinction between *principle* and *implementation* is the key to generalization.
- **Skills vs. agents vs. prompts:** The architecture is sound, but the packaging differs per platform. A generalized spec would describe the *role* of each layer, then map it to each platform's mechanism.

See the follow-on work for the full generalization proposal.

---

*Document created: May 19, 2026*  
*Related: [docs/09_post-skills-quality-review.md](../docs/09_post-skills-quality-review.md), [docs/09_post-skills-quality-review-followup.md](../docs/09_post-skills-quality-review-followup.md)*
