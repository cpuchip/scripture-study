# Brain UX Audit — Space Center Entry Walkthrough

**Date:** 2026-04-08
**Scope:** Walk through creating and advancing a Space Center project entry (Star Trek theme + clock/calculator apps)
**Budget:** 50 premium requests
**Method:** Playwright browser automation, acting as user

## Problem Statement

Is the Brain agentic OS flow smooth? Where are the gaps, stumbling blocks, and automation opportunities when working on a real entry end-to-end?

## Pre-Audit Observations

### Bug Found: Plan Agent Premium Cost
- Plan agent was using `PremiumRequestCost: 1.0` with comment "Sonnet 4"
- But model was already changed to `config.PipelineBigModel` (Opus 4.6, costs 3.0)
- **Fixed:** Updated to `PremiumRequestCost: 3.0` with accurate comment

### Model Layout (current)
| Agent | Model | Cost |
|-------|-------|------|
| Research | Haiku 4.5 | 0.33 |
| Plan | Opus 4.6 | 3.0 |
| Execute | Sonnet 4.6 | 1.0 |
| Review/Nudge | Haiku 4.5 | 0.33 |
| Commit msg | Haiku 4.5 | 0.33 |

## Premium Budget Tracking (50 max)

| Step | Agent | Model | Cost | Running Total |
|------|-------|-------|------|---------------|
| Research | research | Haiku 4.5 | 0.33 | 0.33 |
| Plan | plan | Opus 4.6 | 3.0 | 3.33 |
| Auto-advance fail (x2) | — | — | 0 | 3.33 |
| **Project total (all entries)** | — | — | — | **4.33** |

## Walkthrough Log

### Step 1: Capture → Capture page
- Typed thought about LCARS Vue3 theme + clock + calculator
- **No project selector** on capture page — must go to Edit after saving
- **No title field** — title auto-derived from body text (truncated at ~60 chars)
- **No success toast** after save — textarea clears, entry appears in Recent list, but no explicit "Saved!" confirmation
- Entry classified as "inbox" with 0% confidence (no auto-classify on save)

### Step 2: Entry Detail → Edit → Project Assignment
- Entry detail shows truncated title: "Star Trek LCARS-style Vue3 theme for Space Center apps. Buil"
- Pipeline/Sabbath checkboxes visible but meaning unclear (see UX Bug below)
- Clicked **Edit** to access project assignment dropdown
- **Project dropdown only available in Edit mode** — not visible on detail view or board
- Set title to "LCARS Vue3 Theme — Clock & RPN Calculator", category to "projects", project to Space Center
- API auto-generated: tags, 2 subtasks, next_action, 95% confidence, "planned" maturity

### Step 3: Pipeline Toggle UX Bug
- **Pipeline checkbox label is misleading:** Shows "🔄 Pipeline" when unchecked
- Clicking it MOVED the entry to Notebook mode (opposite of expected!)
- Label changed to "📓 Notebook" — the checkbox describes the action taken, not the current state
- User expectation: "check Pipeline to enable pipeline" → Reality: "checking it exits the pipeline"
- This violates the principle of least surprise

### Step 4: Board View — Advance Failure (planned→specced)
- Entry at "planned" maturity (auto-set by classification, no actual research/plan done)
- Clicked ▶ Advance → **window.alert()** error: "advancing to specced requires scenarios"
- **3 UX failures:**
  1. Uses `window.alert()` instead of toast/inline error (violates UX agent rules)
  2. No guidance on what "scenarios" are or how to provide them
  3. No scenario input mechanism anywhere in the UI
- **Root cause:** Entry was auto-classified to "planned" maturity, skipping research+plan phases
  - The pipeline wasn't aware the entry was new — it just saw "planned" maturity
  - The next step from "planned" is "specced" which needs human scenarios

### Step 5: Research Phase (raw → researched)
- Reset maturity to "raw" via API to walk full pipeline
- Clicked ▶ Advance from board → buttons disabled (good: prevents double-click)
- **No progress indicator visible!** User sees:
  - Disabled buttons (that's it)
  - No spinner, no "Researching..." text, no elapsed time
  - No way to see what the agent is doing  
  - No way to cancel
- Server logs show: research agent fetching Wikipedia (LCARS, Stardate, RPN), StackOverflow, Vue docs, PrimeVue
- Research taking 2+ minutes — user has no idea
- TOKEN WARNING at 112K (threshold 100K) — agent is eating context

### Step 6: Auto-Continue → Plan Phase (researched → planned)
- Research completed at ~21:14:34 (4.5 minutes from start)
- `maybeAutoContinue` fired automatically (auto_continue enabled)
- Plan agent (Opus 4.6) started immediately, 3.0 premium cost
- Plan agent: read project docs, spawned `explore-space-center` subagent
- TOKEN WARNING for plan agent at 157,657 tokens (threshold 150K)
- Plan agent wrote comprehensive plan: 4 phases, scenarios, decisions, risk analysis, consecration/Zion checks
- Plan agent created 4 todos with dependencies via SQL tool
- Plan completed at 21:17:10 (~2.5 minutes)
- **Good:** Auto-continue from research → plan is seamless

### Step 7: Auto-Continue Bug — planned → specced (CRITICAL)
- After plan completed, `maybeAutoContinue` fired AGAIN for "planned"  
- Tried to advance planned → specced with empty AdvanceRequest (no scenarios)
- **Failed twice** — two "⚠️ planned pass failed" messages in conversation
- `window.alert()` dialog popped up on the project board
- **Root cause:** `maybeAutoContinue` checked `NewMaturity != "researched" && NewMaturity != "planned"` — firing auto-advance for "planned" entries even though specced requires human-provided scenarios
- **Fixed:** Changed condition to `NewMaturity != "researched"` only. Comment updated to explain: "planned → specced requires human-provided scenarios."
- Rebuilt binary. Need to restart server to apply.

### Step 8: Entry Detail Panel UX
- Entry detail accessible via click on board card (opens side panel) or "Open →" link
- **Good:**
  - "🔔 Your Turn" badge correctly signals human gate
  - Conversation section shows agent message timeline
  - Full body text visible
  - Advance/Revise/Defer buttons available
- **Gaps:**
  - No scenario input field anywhere (the next step)
  - No link to scratch file (user must know the path)
  - No tags display in side panel
  - No subtasks display in side panel or board card
  - No indication of what to do next — "Your Turn" but for what?
  - Research said "28 open questions" but no way to answer them in the UI
  - Plan said "Review before adding scenarios" but no mechanism to add scenarios

### Step 9: Research Output Quality Assessment
- Research file: `projects/space-center/.spec/scratch/lcars-vue3-theme-clock-rpn-calculator/main.md`
- ~300 lines, well-structured: What This Is About, What Already Exists, External Context, Open Questions, Raw Sources, Plan
- 28 open questions organized by topic — genuine questions, not padding
- 6 web sources fetched and synthesized
- Cross-referenced workspace docs (architecture.md, README, existing scratch)
- **Good:** Research connected LCARS Vue3 to existing physical display dashboard entry — "complementary not competing"
- **Good:** Identified CBS copyright considerations unprompted
- **Concern:** Plan section was written in research phase — plan agent then extended it. Some overlap between research output and plan agent output
- **Concern:** TOKEN WARNING at 112K in research, 157K in plan — context budgets need monitoring

## Bugs Found (2 fixed, 4 open)

### Fixed This Session
1. **Plan agent premium cost** — Was 1.0 (Sonnet), should be 3.0 (Opus). Fixed in research.go.
2. **Auto-continue from planned** — `maybeAutoContinue` tried planned→specced without scenarios, causing double failure message + alert dialog. Fixed: only auto-continue from "researched".

### Open UX Bugs
3. **Pipeline/Notebook checkbox semantics** — Checking "Pipeline" exits the pipeline. Reversed user expectation.
4. **window.alert() for pipeline errors** — Should be toast/inline. All pipeline errors use this.
5. **No progress indicator during agent operations** — Buttons disabled but no spinner, status, elapsed time, or cancel option. Multi-minute operations with no feedback.
6. **No scenario input in UI** — "planned→specced" transition has nowhere to input scenarios. Must use API.

## Missing UI Features

1. **Project selector on Capture page** — Must save first, then Edit to assign project
2. **Title field on Capture page** — Auto-derived from body, often truncated poorly
3. **Scenario input mechanism** — Textarea or checklist that passes scenarios to pipeline advance
4. **Scratch file link** — Entry detail should link to the scratch file for reading research/plan
5. **Tags display** — Not shown on detail panel or board card
6. **Subtasks display** — Not shown on detail panel or board card
7. **Next-step guidance** — "Your Turn" badge but no explanation of what the user should do
8. **Agent activity log** — During agent operations, show what the agent is reading/fetching/writing

## Recommendations (Priority Order)

### P0 — Blocking: Fix before next demo
1. **Replace window.alert() with toast notifications** — Every pipeline error uses alert(). Should be inline toast with clear message.
2. **Add scenario input** — When entry is at "planned" maturity, show a textarea/list for scenarios alongside the Advance button.
3. **(Done) Fix auto-continue from planned** — prevent auto-advance when next step needs human input.

### P1 — Important: Next sprint
4. **Add progress indicator** for agent operations — spinner, status text ("Researching..."), elapsed time, cancel button.
5. **Fix Pipeline/Notebook toggle** — Either change to two separate radio buttons ("Pipeline" / "Notebook") or relabel to match actual behavior.
6. **Add project selector to Capture** — Dropdown or tag-style selector on the capture form.
7. **Add title field to Capture** — Optional override, auto-fill from body if blank.

### P2 — Nice-to-have
8. **Show scratch file link** on entry detail — clickable path to open in editor.
9. **Show tags on board cards** — small badges below the category tag.
10. **Show subtasks on entry detail** — checklist-style display. *(Already exists! Collapsed by default, works well.)*
11. **Next-step guidance** — When "Your Turn" is active, show hint text: "Add scenarios to advance" or "Review research findings".

## Detailed Entry View Assessment

### What Works Well
- **"🔔 Your Turn" badge** correctly signals the human gate
- **Conversation timeline** with timestamps shows full agent message history
- **Sub-tasks** (expandable, editable, deletable, addable) — solid implementation
- **Agent Context disclosure** shows what agents see: project, description, related entries
- **Reclassify buttons** — one-click category change from detail view
- **Premium counter** (🎟️ 3.33) tracks per-entry cost
- **Failure counter** (🔴 1 failure) with tooltip showing last reason
- **Sabbath/Auto toggle** — clear labels: ⚡ Auto / 🕊️ Sabbath (though same UX pattern issue)
- **Tags display** — shown as badges (lcars, space-center, vue3)
- **Mark complete button** — top-level action
- **Reply textbox** with Ctrl+Enter — conversation mechanism exists

### What Needs Work
- **Checkbox labels show current state, not target state** — "🔄 Pipeline" (unchecked) suggests "check to enable pipeline" but actually exits pipeline. Standard: label the target state, or use radio buttons.
- **No Advance/Revise/Defer on detail page** — only on board cards. The detail page is where you'd read the plan and decide to advance, but the button is elsewhere.
- **Scratch file paths are plain text** — agent messages say "Findings at projects/space-center/.spec/scratch/..." but not clickable
- **No scenario input tied to Advance** — reply textbox sends messages, but scenarios need to be passed to `/pipeline/advance` endpoint separately
- **Conversation order** — oldest first (chronological), which reads naturally, but error messages from before research appear above the research results, potentially confusing

## Pipeline Timing Analysis

| Phase | Agent | Model | Duration | Tokens | Premium |
|-------|-------|-------|----------|--------|---------|
| Research | research | Haiku 4.5 | ~4.5 min | 112K (warning at 100K) | 0.33 |
| Plan | plan | Opus 4.6 | ~2.5 min | 157K (warning at 150K) | 3.00 |
| **Total** | | | **~7 min** | | **3.33** |

Observations:
- Research is slower than plan despite using a faster model — more tool calls (8 web fetches, 12+ file reads/edits)
- Plan is faster with Opus because it does fewer tool calls (mostly reading + writing)
- Both agents hit token warnings — context pressure is real
- 7 minutes total for research+plan with no user feedback — need progress indicators
- Auto-continue is seamless between research → plan but breaks at plan → specced

## Root Cause Analysis

The core issue isn't individual bugs — it's a **human gate design gap**. The pipeline has three types of transitions:

| Transition | Type | UI Support |
|-----------|------|------------|
| raw → researched | Agent runs | ✅ Advance button triggers agent |
| researched → planned | Agent runs | ✅ Auto-continue or Advance |
| planned → specced | **Human input** (scenarios) | ❌ No input mechanism |
| specced → executing | Agent runs | ✅ Advance button triggers agent |
| executing → verifying | Agent runs | ✅ Auto-advance |
| verifying → complete | Human accepts | ❓ Mark complete? |

The pipeline was built for agent-driven transitions but the human gates (planned→specced, verifying→complete) lack their own UI. The "Your Turn" badge *signals* the gate but provides no *mechanism* to pass through it.

## The Fix Chain

1. **(Done)** Fix auto-continue — don't auto-advance from planned
2. **Next:** Add scenario textarea to entry detail when maturity=planned
3. **Next:** Wire scenario textarea to Advance button on detail page
4. **Then:** Add progress indicators for agent phases
5. **Then:** Replace window.alert with toast/inline errors
6. **Then:** Fix checkbox toggle labeling

