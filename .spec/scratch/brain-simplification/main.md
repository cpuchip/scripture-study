# Brain Simplification Reflection — Apr 5, 2026

*Scratch file: research provenance for brain dashboard simplification decisions*
*Triggered by: KISS brain entry and honest reflection on over-engineering*

---

## The KISS Tension

The brain pipeline has 6 maturity stages (raw → researched → planned → specced → executing → verified), a review nudge bot firing 4x/day, execution gates, verification steps, context injection, and project-based orchestration. This is sophisticated infrastructure.

But most entries are **simple captures** — a thought while walking, a study idea, a task to do. These don't need a 6-stage maturity ladder. They need tags, search, and done.

The pipeline was built for **delegation** — the use case where Michael says "research this and bring me back a plan." That's real and valuable. But it's maybe 10% of entries. The other 90% are being forced through infrastructure designed for the 10%.

## How We Actually Work

When Michael and I build something together, the real workflow is:

1. Michael has an idea (verbal capture or typed)
2. We discuss it together (council moment)
3. I write a spec in `.spec/proposals/`
4. Michael says "build it"
5. I build it in one session
6. Michael tests it
7. We iterate next session

That's 7 conversational steps. The formal pipeline maps these to automated stages, but the *live collaboration* is what actually works. The pipeline is for when Michael isn't in the room.

## Four Types of Work

| Type | What it needs | What it doesn't need |
|------|--------------|---------------------|
| **Capture** | Tags, search, done | Classification, pipeline, agent routing |
| **Task** | Tracking, due date, completion | 6-stage maturity ladder |
| **Project idea** | Council → spec → build cycle | Automated pipeline (done live in conversation) |
| **Delegation** | Full pipeline | Human presence during execution |

Only Delegation needs the full pipeline. The other three types are being over-served by infrastructure they don't need.

## Board Simplification: 3 Columns

Current board has horizontal scroll with too many columns (raw, researched, planned, specced, executing, verified). Recommendation:

| Column | Contains | Sub-stage shown as badge |
|--------|----------|--------------------------|
| **Inbox** | raw entries | — |
| **Working** | researched, planned, specced, executing | Badge shows which stage |
| **Done** | verified, closed, dismissed | Badge shows outcome |

This gives the Kanban view its visual power back (see everything, scan status) without needing horizontal scroll. The badge tells you *where* in the pipeline something is, but the column tells you the *meaningful* state.

## Auto-Continuation

For the fully automated pipeline (delegation use case), an auto-continuation checkbox:

- When checked: after each pipeline stage completes, automatically advance to the next stage without waiting for human review
- "Run until it hits a question" — the agent keeps going until it needs human input
- Like submitting a batch job: come back when it's done or stuck
- Pairs with the review nudge bot: auto-continuation for the happy path, nudge bot for when it gets stuck

## Notebook Mode

For entries that are just captures and don't need pipeline processing:

- A way to mark an entry as "notebook" — it stays as-is, searchable, taggable, but never enters the pipeline
- Think of it as the journal/notes that don't need to become projects
- Could be a category, or a flag on any entry

## 11-Step Creation Cycle Gap Analysis

Mapping the brain pipeline against the 11-step creation cycle from the guide:

| Step | Name | Represented in Pipeline? |
|------|------|-------------------------|
| 1 | Intent | ✅ Entry creation, binding problem |
| 2 | Covenant | ❌ No rules of engagement per-entry |
| 3 | Stewardship | ✅ Project assignment, agent routing |
| 4 | Spiritual Creation | ✅ Spec phase — "build it first in mind" |
| 5 | Line Upon Line | ✅ Phased maturity ladder |
| 6 | Physical Creation | ✅ Execution phase |
| 7 | Review | ✅ Verification gate + nudge bot |
| 8 | Atonement | ❌ No error recovery, no "what went wrong" |
| 9 | Sabbath | ❌ No reflection pause — it just keeps going |
| 10 | Consecration | ❌ No "who benefits" check |
| 11 | Zion | ❌ No integration check |

Steps 2, 8, 9, 10, 11 have no representation in pipeline tooling. Step 9 (Sabbath) is especially notable — the pipeline has no natural stopping point for reflection.

## Review Nudge Bot Concerns

Current behavior (from `pipeline/review.go`):
- Fires at hours [7, 11, 15, 19] local time — 4 times per day
- Uses Haiku (ResearchModel) — 0.33 premium requests per nudge
- Scans up to 10 stale entries per wake (via `ListStaleEntries`)
- Creates a NEW `ai.NewAgent` per nudge — each creates a Copilot SDK session visible in VS Code sidebar
- After nudge: sets `route_status="your_turn"`, `agent_route="review"`
- Stale thresholds: raw after 24h, researched after 48h, complete after 24h

Problems identified:
1. **Invisible** — doesn't appear in Scheduled Tasks tab. It's a hardcoded goroutine started at `server.go:51`
2. **No pause control** — can't be paused from UI; requires code change or restart
3. **Clutters VS Code** — each nudge creates a new Copilot SDK session in the sidebar
4. **Fires when user isn't around** — 4x/day regardless of whether Michael is at the desk
5. **Unclear value** — "not sure if it actually moves things along"
6. **Session zombies** — `ListStaleEntries` excludes `route_status IN ('your_turn', 'running', 'pending')`, so it won't re-nudge entries it already nudged. But those entries sit in "your_turn" indefinitely if the user doesn't respond.

## Key Decisions (Apr 5 Session)

1. **Both paths wanted.** Michael wants BOTH simplified workflow (notebook mode, 3 columns) AND fully automated pipeline (auto-continuation). Not one or the other.
2. **KISS for captures, power for delegation.** Most entries need simple UX. Delegation entries get the full pipeline.
3. **Nudge bot needs controls.** Must be visible in Scheduled Tasks, pausable, and transparent about what it's doing.
4. **Space Center as test bed.** Practice the fully automated pipeline on Space Center project (project_id=4) as a low-stakes test.
5. **"By small and simple things."** Alma 37:6 as design principle — don't build infrastructure ahead of use case.

## Recommended Action Sequence

1. **Build inline panel spec** (reply + close) — small, solves real friction (ALREADY SPECCED)
2. **Add nudge bot to Scheduled Tasks** — make it visible and controllable

---

## Graduated to Proposal (Apr 6)

This scratch file's findings — the 11-step creation cycle gap analysis, notebook mode, auto-continuation, 3-column board, and nudge bot concerns — have been graduated into a formal proposal:

**[Brain Pipeline Evolution — Creation Cycle Completion](../../proposals/brain-pipeline-evolution.md)**

Research provenance for the proposal: [.spec/scratch/brain-pipeline-evolution/main.md](../brain-pipeline-evolution/main.md)
3. **Simplify board** — 3 columns (Inbox/Working/Done), badges for sub-stages
4. **Add notebook mode** — captures that are just captures
5. **Auto-continuation for delegation** — let it run until it hits a question
6. **Stop adding infrastructure** — let it settle, use it, see what's actually needed

Don't do all 6 at once. Do #1, use it. Then #2, use it. Mosiah 4:27.
