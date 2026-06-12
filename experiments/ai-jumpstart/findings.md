# ai-jumpstart — cold-model experiment findings

*2026-06-12. Method: stage the kit in a fresh folder, role-play the human using
Michael's voice profile (`docs/voice-profile-michael.md`), run a cold model headless,
grade against the seed's intent. Runs preserved in `run1-sonnet/`, `run2-gemini/`.*

## Run 1 — Claude Sonnet (`claude -p`, 2 turns)

**Turn 1** (vague vision: "simple chore tracker for my kids, nothing fancy"):
- ✅ Read AGENTS.md, did NOT scaffold, asked SIX numbered specific questions
  (ages/interaction, cycle, rewards, editability, hosting, tech), closed with "once
  the vision has edges I'll restate a plan and we'll agree before writing a line."
  Practice 1 + 2 working as designed.

**Turn 2** (Michael-voiced numbered answers):
- ✅ Created all four working files from templates; intent.md has CHECKABLE done-
  criteria; journal written UNPROMPTED with decisions-and-why including a calibrated
  "pending Michael's confirmation"; five-phase plan in active.md; no premature code;
  asked 4 final pre-build questions.
- ⚠ Finding 1: **assumed the user's name was "Michael"** — inferred from the book
  credit in the kit (the only name present). Seed fix: setup now says ask, don't
  assume. (v0.1.1)
- ⚠ Contamination note: running inside the workspace inherited the workspace's MCP
  config (`becoming-mcp.log` appeared). True cold tests should run outside the
  workspace. Doesn't invalidate the behavioral results.

## Run 2 — Gemini (`agy -p`, 1 turn)

**Turn 1** (same vague vision):
- ✅ Read AGENTS.md, asked FOUR numbered questions — including a parent-approval-queue
  question Sonnet missed. Volunteered explicit bounds unprompted. No files created
  prematurely.
- ⚠ Finding 2: **proposed a full tech stack + plan in the same message as its
  questions** — "sleek, highly interactive SPA… premium dark/light mode with rich
  micro-animations" — to a human who had said **"nothing fancy."** Pre-committing a
  design defeats the questions, and "premium polish" directly upgraded a stated
  constraint. Seed fix: "Ask, then stop" + "stated constraints are bounds, not modesty
  to upgrade." (v0.1.1)
- Harness notes: agy stdout-drop reproduced (recovered via the transcript jsonl, per
  the agy-cli skill); turn-2 continuation under headless agy not attempted yet.

## Verdict so far

The seed steers both models through the core gate — **ask before building, set up
memory, no premature code** — on the first try. The drift modes are model-flavored
(Sonnet: identity inference; Gemini: eager pre-design + polish inflation) and both are
addressable with one-line seed amendments, which is exactly the iteration loop Michael
wanted.

## Next experiments (queued)

- Gemini turn 2 (answers → file setup fidelity) once agy continuation is worked out.
- A run from OUTSIDE the workspace (no inherited MCP config) for a true cold room.
- GPT-class model when available (codex CLI or via API).
- The substrate path: hand the kit to chattercode in a Docker sandbox and watch it
  flush out a repo end-to-end (Michael's original suggestion — needs a substrate
  session).
- Longitudinal: does session 2 actually read memory first? (The real portability test.)
