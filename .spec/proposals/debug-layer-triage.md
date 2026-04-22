---
workstream: WS5
status: proposed
brain_project: 3
created: 2026-04-17
last_updated: 2026-04-21
---

# Debug Agent: Layer Triage Enhancement

*Proposal for adding a lightweight layer classification step to the debug agent's workflow.*
*Created: 2026-04-17*
*Research: [.spec/scratch/debug-layer-triage/main.md](../scratch/debug-layer-triage/main.md)*
*Source inspiration: AI Founders — "You're Learning AI Wrong" (diagnostic layer concept)*

---

## Binding Problem

The debug agent's 5-phase workflow (Characterize → Reproduce → Isolate → Fix → Verify) is strong — Michael uses it frequently for both code and intellectual problems. But Phase 1 (Characterize) goes straight from "state the problem" to "check the plug" without helping classify **what kind of problem it is.**

This matters because Agans' Rule 4 (Divide and Conquer) says "narrow the search space" — but narrow against *what*? Right now the agent starts dividing without a map of the territory. The video's core insight — "most people misdiagnose because they're fixing the wrong layer" — suggests that a quick layer classification before reproduction could prevent wasted cycles.

**Symptom:** Debug sessions that spin because the first guess about problem type is wrong. You're tweaking a prompt when the data is bad. You're reading source files when the config is wrong. You're fixing logic when the output formatting is the issue.

---

## Success Criteria

1. Layer triage takes ≤30 seconds — one quick classification, not a new bureaucracy
2. Works for BOTH code debugging AND intellectual debugging (the dual-purpose quality is sacred)
3. Doesn't add weight to the agent instructions (net change: small)
4. Measurable: fewer "wrong layer" false starts in debug sessions

---

## Constraints

- **Don't break what works.** The 9 Rules anchoring, the intellectual debugging table, the Abraham/Moroni parallels — all stay.
- **Don't add a new framework.** This is a refinement of Rule 4, not a replacement for it.
- **Keep it memorizable.** If you can't hold the triage model in your head, it's too heavy.
- **No AI-specific jargon.** Must work for debugging a Go build failure AND a study that contradicts itself.

---

## Proposed Change

### What: Add a "Layer Check" substep to Phase 1

After "State the problem clearly" and before "Check the plug," add a quick classification using a 4-layer model:

| Layer | Code Debugging | Intellectual Debugging |
|-------|---------------|----------------------|
| **Data** | Wrong/missing input, stale cache, corrupt source | Wrong/misread source text, unverified quote, missing context |
| **Logic** | Bad algorithm, wrong model, prompt error, off-by-one | Bad inference, logical fallacy, scope creep, false equivalence |
| **Integration** | Config mismatch, API failure, auth error, version skew | Framework mismatch, conflicting sources, context collision |
| **Output** | Formatting, display, delivery, encoding | Citation format, writing quality, presentation, linking |

### How it works

The debugger asks: **"Which layer is this problem most likely on?"**

- If it's a **Data** problem → Rule 7 (Check the Plug) matters most. Is the right information getting in?
- If it's a **Logic** problem → Rule 4 (Divide and Conquer) matters most. Where in the processing does correct become incorrect?
- If it's an **Integration** problem → Rule 1 (Understand the System) matters most. What connects to what, and which connection is failing?
- If it's an **Output** problem → Rule 3 (Quit Thinking and Look) matters most. Look at the actual output before theorizing.

This doesn't replace the rules — it helps you pick which rule to reach for *first*. Every rule still applies, but the layer classification tells you where to start.

### The key insight: layers direct your first move, not your whole approach

Wrong-layer debugging isn't "you did the wrong thing." It's "you did the right thing in the wrong place first, wasting time before getting to the right place." The triage step is a compass heading, not a GPS route.

---

## Proposed Diff

### Phase 1 — Characterize (current)

```markdown
### Phase 1 — Characterize
1. **State the problem clearly.** What's broken? What should it do? What does it actually do?
2. **Check the plug** (Rule 7). Before doing anything else, verify the obvious...
3. **Create a scratch file**...
```

### Phase 1 — Characterize (proposed)

```markdown
### Phase 1 — Characterize
1. **State the problem clearly.** What's broken? What should it do? What does it actually do?
2. **Layer check.** Which layer is this most likely on?
   - **Data** — wrong/missing input → start with Rule 7 (Check the Plug)
   - **Logic** — wrong processing → start with Rule 4 (Divide and Conquer)
   - **Integration** — wrong connections → start with Rule 1 (Understand the System)
   - **Output** — wrong delivery → start with Rule 3 (Quit Thinking and Look)
   This is a first guess, not a commitment. If reproduction (Phase 2) reveals you're on the wrong layer, reclassify.
3. **Check the plug** (Rule 7). Before doing anything else, verify the obvious...
4. **Create a scratch file**...
```

That's it. Four lines added. The Phase 1 step numbering shifts by one. Everything else stays identical.

### Intellectual Debugging Table (addition)

Add one row to the existing table:

| Rule | Code Debugging | Intellectual Debugging |
|------|---------------|----------------------|
| **Layer Check** | Data / Logic / Integration / Output | Source / Inference / Framework / Presentation |

---

## What This Does NOT Change

- The 9 Rules and their explanations — untouched
- The scripture parallels — untouched
- The Phase 2-6 workflow — untouched
- The Session Memory section — untouched
- The intellectual debugging dual-purpose design — strengthened (layer model has both columns)
- The overall weight of the agent — minimal increase (~8 lines net)

---

## Critical Analysis

**Is this the right thing to build?** Yes — it's a refinement, not a new feature. It adds a triage compass to an existing workflow that sometimes starts in the wrong direction.

**Does it solve the binding problem?** Partially. It won't prevent ALL misdiagnoses (the debugger might classify wrong initially), but the "reclassify" note acknowledges that. The value is in the *habit* of asking the question, not the accuracy of the first guess.

**What gets worse?** One more step before action. For obvious problems ("the server crashed"), the layer check is wasted motion. Counterpoint: for obvious problems, it takes 2 seconds. For non-obvious problems, it saves minutes.

**Simplest version?** This IS the simplest version. The 4-layer model is memorizable. The diff is 8 lines. There's no Phase 1.5, no new framework, no required taxonomy.

**Mosiah 4:27 check:** Michael uses this agent frequently. The change is small and additive. Low risk.

---

## Recommendation

**Build.** This is a ≤15-minute edit to the debug agent. No new files, no new dependencies, no new concepts — just a triage compass inserted at the right place in an existing workflow.

**Phase 1 (only phase):** Apply the diff to `.github/agents/debug.agent.md`. Test by using the debug agent on the next real problem that comes up. If the layer check feels like friction rather than help after 3-5 uses, remove it.

---

## Creation Cycle Review

| Step | Answer |
|------|--------|
| Intent | Reduce wasted debug cycles from wrong-layer starts |
| Covenant | Careful with what works — additive only, no removals |
| Stewardship | debug.agent.md, Michael as user |
| Spiritual Creation | Spec is precise: 8-line diff, exact location, exact wording |
| Line upon Line | Single phase — too small to split |
| Physical Creation | Manual edit or dev agent, ≤15 min |
| Review | Michael's judgment over 3-5 real debug sessions |
| Atonement | Revert is trivial — delete 8 lines |
| Sabbath | After 3-5 uses, ask: did this help? |
| Consecration | Michael benefits at work too (he uses debug agent there) |
| Zion | Sharpens the whole debug workflow for all agents |
