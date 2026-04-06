# Brain Pipeline Evolution — Research & Findings

**Binding Problem:** The pipeline handles the mechanical stages (research → plan → spec → execute → verify) but skips 5 of the 11 creation cycle steps. It has no per-entry covenant, no error recovery beyond silent rollback, no reflection pause, no "who benefits" check, and no integration verification. It also forces all entries through the same pipeline regardless of type, and the nudge bot operates invisibly.

**Session:** 2026-04-06
**Source material:** [brain-simplification scratch](../brain-simplification/main.md) (creation cycle gap analysis, Apr 5)

---

## Current Pipeline Architecture

### Maturity Stages & Transitions

```
raw → researched → planned → specced → executing → verified
                                 ↑                    |
                                 └── (verify fails) ──┘
```

**No "failed" maturity state.** Any string accepted by `SetMaturity()` — no schema constraint.

### Error Handling Today

| Failure Point | Current Behavior | State After |
|---------------|-----------------|-------------|
| Research agent fails | Sync error returned | Still "raw" |
| Plan agent fails | Sync error returned | Still "researched" |
| Execute agent fails | Async — logs error, rolls back | Back to "specced" + session message |
| Verify scenarios fail | Sync — rolls back with feedback | Back to "planned" |
| Nudge agent fails | Logged, loop continues | No change |

Key observations:
- Execute is the only stage that posts a human-readable failure message
- Other failures just return errors to the HTTP endpoint — user sees "pipeline advance failed"
- No dead-letter queue — failed entries sit in their pre-transition state silently
- No retry mechanism — user must manually re-advance

### Governance Documents

**Not created.** Phase 4 proposal specified governance docs per pipeline layer (`research-covenant.md`, `plan-covenant.md`, `execution-covenant.md`). The code reads them if present but falls back to hardcoded prompts. Warnings logged but pipeline runs without them.

### Existing Extension Points

1. **`notify()` method** — Called after stage transitions. Could inject pre/post hooks.
2. **`Advance()` entry point** — All transitions flow through here. Natural middleware point.
3. **Session messages** — Agent posts reflections/questions. Human reads and responds. This IS the reflection loop.
4. **Route status "your_turn"** — Pause mechanism already exists. Set after execute, after nudge. Could set after ANY stage.
5. **Actions: advance/revise/reject/defer** — Already 4 human choices. Could add more (e.g., "reflect", "pivot").

### Four Types of Work (from simplification analysis)

| Type | % of Entries | What They Need | Pipeline Fit |
|------|-------------|----------------|-------------|
| Capture | ~60% | Tags, search, done | Ovserved |
| Task | ~20% | Tracking, due date, completion | Overserved |
| Project idea | ~10% | Council → spec → build (live in conversation) | Overserved |
| Delegation | ~10% | Full automated pipeline | Good fit |

### Nudge Bot Issues

- Hardcoded goroutine, not visible in Scheduled Tasks
- No pause/resume from UI
- Creates Copilot SDK sessions per nudge (clutters VS Code)
- Fires at [7,11,15,19] regardless of user presence
- Excludes entries already at "your_turn" — zombie entries stay there
- 0.33 premium requests per nudge × up to 10 entries × 4 times/day

---

## Creation Cycle Gap Analysis

From the 11-step creation cycle ([docs/work-with-ai/guide/05_complete-cycle.md](../../docs/work-with-ai/guide/05_complete-cycle.md)):

| Step | Name | Pipeline Status | Gap |
|------|------|----------------|-----|
| 1 | Intent | ✅ Entry creation, binding problem | — |
| 2 | Covenant | ❌ | No per-entry rules of engagement. Governance docs not created. |
| 3 | Stewardship | ✅ Project assignment, agent routing | — |
| 4 | Spiritual Creation | ✅ Spec phase | — |
| 5 | Line Upon Line | ✅ Phased maturity ladder | — |
| 6 | Physical Creation | ✅ Execution phase | — |
| 7 | Review | ✅ Verification gate + nudge bot | Nudge bot invisible/uncontrollable |
| 8 | Atonement | ❌ | No error recovery, no "what went wrong" audit, silent rollback |
| 9 | Sabbath | ❌ | No reflection pause. Pipeline advances immediately. No "stop and see" moments. |
| 10 | Consecration | ❌ | No "who benefits" check. Work has no checked purpose beyond task completion. |
| 11 | Zion | ❌ | No integration check. New work lands in isolation. |

### Analysis of Each Gap

**Covenant (Step 2):** The Phase 4 proposal specified governance documents per pipeline agent, loaded into system messages. These were never written. The code reads them if present (`research-covenant.md`, `plan-covenant.md`, `execution-covenant.md`) but falls back to hardcoded prompts. Writing these documents IS the covenant implementation — no code changes needed, just content creation.

**Atonement (Step 8):** When things go wrong, the pipeline silently rolls back (execute → specced, verify → planned) or returns HTTP errors. There's no:
- Human-readable "what happened and why" for non-execution failures
- Pattern tracking: "this entry has failed 3 times — something structural is wrong"
- Recovery prompts: "Would you like to revise the plan, change scope, or abandon?"
- Dead-letter visibility: entries stuck in pre-transition states are invisible

**Sabbath (Step 9):** The pipeline has no reflection pause. After research completes → plan can start immediately. After plan → spec is immediate. The "your_turn" route_status exists but is only set after execution. Earlier stages auto-advance without the human ever seeing the research output or plan output before it continues. There's no built-in "stop, look at what was made, and declare it good" moment.

**Consecration (Step 10):** No pipeline step asks "who is this for?" or "what purpose does this serve beyond task completion?" This is less about code and more about prompting — the plan agent could include a "beneficiaries" section in its output.

**Zion (Step 11):** New work lands in isolation. There's no "how does this connect to existing work?" check. The research agent could cross-reference existing studies/entries/proposals, but currently only searches for direct context, not integration points.

---

## Critical Analysis

**Is this the right thing to build?** Mixed. The gaps are real, but not all of them need code.

- **Covenant**: Write the governance docs. No code changes. Immediate value. Highest priority.
- **Atonement**: Failure visibility + retry UX are genuine engineering work. Moderate priority.
- **Sabbath**: The route_status="your_turn" mechanism already exists. Extending it to earlier stages is small. But auto-continuation (which Michael also wants) goes the opposite direction — less pausing, not more. Tension to resolve.
- **Consecration + Zion**: These are prompt engineering in the plan agent, not pipeline mechanics. Low code, high value if the prompts are good.

**Sabbath vs. Auto-Continuation tension:** Michael wants BOTH reflection pauses AND auto-continuation. The resolution: auto-continuation is the delegate-and-forget mode; reflection pauses are the engaged-and-present mode. The entry's type/flag determines which path. Delegation entries auto-continue. Everything else pauses.

**Scope creep risk:** This proposal touches pipeline internals, agent prompts, UI, and nudge bot. That's wide. Phasing is critical — each phase must ship standalone value.
