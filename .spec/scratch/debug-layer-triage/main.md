# Debug Agent Layer Triage — Research Notes

*Created: 2026-04-17*

---

## Source

**Video:** "You're Learning AI Wrong. Here's The Cheat Sheet." — AI Founders (2026-04-14)
**Transcript:** `yt/ai-founders/Zd8dA7bijzo/`
**Key quote (0:00-0:05):** "The most expensive mistake that you can make in AI isn't using the wrong tool. It's not knowing which layer the problem is on."

## The Concept

Speaker presents 8 groups of AI concepts as a "Periodic Table of AI Elements." The diagnostic value isn't the table itself (it's a teaching taxonomy, not a scientific framework — no predictive power like Mendeleev's actual table). The value is in the **misdiagnosis pattern**:

> People treat surface-level symptoms with surface-level fixes. They change the prompt when the problem is missing data. They switch models when the problem is missing guardrails. They rebuild the tool when the problem is infrastructure.

The fix: **identify which layer the problem is on BEFORE attempting fixes.** This is compatible with Agans' Rule 1 (Understand the System) and Rule 4 (Divide and Conquer), but adds a *dimension* they don't explicitly cover — the system isn't just horizontal (pipeline stage) but vertical (abstraction layer).

## Current Debug Agent Analysis

The debug agent is strong. It applies Agans' 9 rules with scripture parallels, works for both code and intellectual debugging. Key strengths:

1. **Characterize → Reproduce → Isolate → Fix → Verify** is a clean workflow
2. The Intellectual Debugging table makes it genuinely dual-purpose
3. Rule anchoring prevents random guessing
4. Abraham 4:18 ("watched until they obeyed") as Rule 9 — elegant

**What it doesn't do explicitly:** help the debugger quickly classify WHERE in the stack the problem lives. The 5-phase workflow starts with "State the problem clearly" but doesn't provide a triage framework for WHAT KIND of problem it is. You go straight from characterizing to reproducing.

Agans' Rules 1 and 4 implicitly cover this — "Understand the System" means know the layers, and "Divide and Conquer" means narrow the search. But the debug agent doesn't give you a layer map to divide against.

## Risk Assessment

The debug agent is one of Michael's most-used modes. He explicitly said it works for both code and ideas. Changes must:
- NOT break the intellectual debugging capability
- NOT make the agent heavier or more bureaucratic
- NOT require memorizing a new taxonomy
- Add a quick triage step that ACCELERATES the existing workflow

## Layer Map Options

### Option A: The Video's 8 Groups (Don't Do This)
Too AI-specific. Doesn't apply to intellectual debugging. Groups like "No-Code Builder Tools" and "Business Layer" aren't debugging layers — they're product categories.

### Option B: Generic 4-Layer Stack
Works for both code and ideas:

| Layer | Code | Ideas |
|-------|------|-------|
| Data | Wrong/missing input, stale cache, bad source | Wrong/missing source text, misread passage |
| Logic | Bad algorithm, wrong model, prompt flaw | Bad inference, logical fallacy, scope error |
| Integration | API failure, config mismatch, auth error | Framework mismatch, context collision |
| Output | Formatting, display, delivery | Writing, citation, presentation |

### Option C: 3-Zone Model (Simplest)
Even simpler — where does the problem ENTER the system?

| Zone | Question |
|------|----------|
| Input | Is the right information getting in? |
| Processing | Is the system doing the right thing with it? |
| Output | Is the result being delivered correctly? |

## Recommendation

Option B as a lightweight triage table inserted between Phase 1 (Characterize) and Phase 2 (Reproduce). Not a new phase — a 30-second classification step within Phase 1 that helps Rule 4 (Divide and Conquer) start faster.

**Proposal extracted to:** [.spec/proposals/debug-layer-triage.md](../../proposals/debug-layer-triage.md)
