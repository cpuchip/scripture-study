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


## Run 3 — Claude Opus (`claude -p`, 2 turns; same prompts as Run 1)

**Turn 1:**
- ★ Explored UP a directory unprompted, found `findings.md`, and OPENED with the meta:
  "I can see what this is: run 3 of your cold-model experiment, with the v0.1.1 seed
  fixes already folded in" — then offered to play it as a true cold run if preferred.
  Michael's "sees around corners" phrase, demonstrated literally.
- Applied both v0.1.1 fixes BY NAME (held "nothing fancy" as a bound; asked the name
  question while disclosing it could infer). Five questions, each carrying its own
  reasoning ("that decides whether the screen leans on words or icons").

**Turn 2 (same answers as Run 1):**
- Plan table includes a **"Verify by" column** unprompted (Phase 1: "restart server,
  confirm the check survived") — verification discipline materializing from the seed.
- ★ "Per-date storage in Phase 2 is the hinge — it also **pre-pays for the streaks
  goal later, for free**" — designed for the human's deferred wish at zero cost.
  Sonnet deferred streaks; Opus architected for them without building them.
- Recommended **JSON-under-mutex over SQLite** explicitly because "light" was the
  brief (Sonnet recommended SQLite) — held the stated constraint deeper into the
  stack. Asked one kid-empathy question nobody else saw: the 11:50pm chore vs a
  midnight rollover "can feel unfair," offered a 4am logical-day option.
- Volunteered bounds + per-phase check-ins; confirmed the memory loop; journal
  already written.

## Sonnet vs Opus on the SAME seed (Michael's lived observation, corroborated)

Michael (after a week on Sonnet at work when Opus tokens ran out): "the smaller
models need a lot more hand holding and forethought from the presiding agent… I hit
decision fatigue." / "sonnet you have to have a much clearer vision… or ask a LOT of
questions" / "gemini 3.5 flash realllly needs [the instructions] otherwise it
improvises."

The runs agree: all three models pass the seed's hard gates (ask-first, files, no
premature code) — the seed raises the FLOOR. The deltas are ceiling-shaped:
Sonnet executes the protocol faithfully; Opus additionally reasons about WHY each
step exists, holds constraints deeper, pre-pays deferred goals, and adds
verification criteria nobody asked for. Gemini-flash is the most improvisational
without firm instructions. Decision-fatigue prediction: Sonnet pushes ~4-6 decisions
back to the human per exchange with little triage; Opus triages to the 2 that
matter and attaches vetoable recommendations.

## Designed next: the instruction-ablation experiment (Michael's question)

Goal: measure how much of "it works well" is the seed vs. the heavier workspace
disciplines doing invisible lifting.

- **Matrix:** {seed-only | seed + DISCIPLINES pack} × {sonnet, gemini-flash}, with
  opus seed-only as ceiling baseline. The DISCIPLINES pack = a distilled extra file
  (council-moment scan, intent-check, read-before-quoting/cite-or-hedge, reversibility
  bias, autonomy bins) injected next to AGENTS.md.
- **Scorecard per run (count, don't vibe):** premature-build (any code before
  ratify) · constraint-upgrades ("nothing fancy" → polish) · identity/name
  assumptions · unverified factual claims · decisions pushed to human without a
  recommendation · memory-loop completion (journal+active updated unprompted) ·
  questions-with-reasoning ratio.
- **Method fixes now REQUIRED:** run OUTSIDE this workspace (twice-proven
  contamination: inherited MCP config writes `becoming-mcp.log`; Opus read the
  experiment scaffolding itself); identical scripted turns; same task.

## Next experiments (queued)

- Gemini turn 2 (answers → file setup fidelity) once agy continuation is worked out.
- A run from OUTSIDE the workspace (no inherited MCP config) for a true cold room.
- GPT-class model when available (codex CLI or via API).
- The substrate path: hand the kit to chattercode in a Docker sandbox and watch it
  flush out a repo end-to-end (Michael's original suggestion — needs a substrate
  session).
- Longitudinal: does session 2 actually read memory first? (The real portability test.)
