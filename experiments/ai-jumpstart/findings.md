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


## ABLATION RUN — 2026-06-12 (cold room at ../ai-jumpstart, OUTSIDE the workspace)

4 arms, turn-1, identical task ("simple chore tracker... nothing fancy"), v0.1.1 seed.
Cold room verified clean (no MCP leakage, no scaffolding visible). DISCIPLINES.md =
the heavier pack (council moment, intent check, cite-or-hedge, reversibility, autonomy
bins, keep-the-watch-whole, when-you-delegate-you-preside — the last from study/
preside.md + the covenant's new presiding extension).

| Marker (turn-1) | A sonnet seed | B sonnet+pack | C flash seed | D flash+pack |
|---|---|---|---|---|
| Premature build | 0 | 0 | 0 | 0 |
| Name assumed | no — asked | no — asked | no — asked | no — asked |
| Pre-design w/ questions | none ("I'll hold off") | none ("Ask, then stop") | none | none |
| Council moment surfaced | — | ★ YES, unprompted ("your stated constraints are real bounds I'll honor") + gap-scan | — | — |
| File-placement comprehension | ok | ok | ✗ proposed working files INSIDE the kit folder | ✓ "in your project directory" |
| Sprawl / branding | normal | tight | 2,207 chars + "I'm Antigravity" intro | 1,192 chars, no branding |

**Findings:**
1. **The v0.1.1 seed fix, not the pack, is what tamed flash.** Run 2 (v0.1, warm room)
   had flash pre-designing "premium micro-animations"; arm C (v0.1.1, cold room) had
   ZERO pre-design. Michael's "flash realllly needs them" hypothesis sharpens to:
   flash needs the EXPLICIT rule in the seed; once written, it follows it. The seed
   iteration loop demonstrably works.
2. **The pack moves smaller models toward larger-model behavior.** Sonnet+pack
   surfaced forethought BEFORE questions (the council moment) — the very thing Opus
   does natively; flash+pack fixed a comprehension slip and halved the sprawl.
3. **Floor/ceiling holds across the matrix:** all arms clear the hard gates (floor =
   the seed); the deltas are forethought-shaped (ceiling = model + pack).

**Actions:** DISCIPLINES.md graduated into the kit (v0.2, pushed) as the optional
heavier pack. Ablation artifacts archived at `ablation-2026-06-12/` (sonnet outputs +
the pack; flash replies recovered from agy transcripts); cold room marked ARCHIVED,
kept for reproduction. Docker path still open (CLI auth inside containers unsolved —
the one-folder-up cold room proved sufficient isolation for now).


## NIGHT LAP 2 — 2026-06-12 (cold Opus + turn-2s + THE SESSION-2 TEST)

**Arm E — Opus, cold room, seed-only (the uncontaminated ceiling baseline):**
- Turn 1: made "nothing fancy" itself a clarifying question ("usable this week, or a
  fun project we polish?"), and articulated ask-then-stop's REASON unprompted ("I don't
  want to answer these *for* you by picking a stack first"). Ceiling confirmed clean.
- Turn 2: ★ NEW FINDING — **over-gating**. Asked to "get the working files setup," it
  created NOTHING, holding the granted reversible action behind four design forks
  (excellent forks, with recommendations — wrong sequencing). Sonnet+pack, same
  instruction, created the files AND asked its remaining questions alongside.
  → Seed v0.2.1: "counsel is not a toll-booth" (pushed). Also noted: headless -p
  degraded an interactive-question tool gracefully ("I'll skip the form").

**★★ THE SESSION-2 TEST — PASSED at the floor tier.** Fresh Sonnet conversation (no
--continue, zero shared context) in arm B's folder, prompt only "this is session 2,
lets pick up where we left off": it opened "Welcome back, Michael" (name read FROM the
memory, where session 1 put it), summarized the exact state (plan drafted), and resumed
at the three named open questions with options. The kit's central promise — the next
session, or a different AI entirely, arrives already knowing — demonstrated end-to-end
on the smallest model in the matrix.

**Matrix status:** A/B/C/D turn-1 ✓, B turn-2 ✓, B session-2 ✓, E turns 1-2 ✓ (cold).
Still open: flash turn-2 (agy headless continuation undocumented), cross-model
session-2 (e.g., OPUS reads SONNET's memory — the model-swap test), GPT-class arm.

## NIGHT LAP 3 — 2026-06-12 (haiku + the Gemini-variant attempt)

**Arm F — Haiku, cold, seed-only: ✅ PASSED turn 1.** Name question, five sharpeners,
no premature design, and a genuinely good probe: "when you say 'nothing fancy,' what
does done look like?" The seed's floor holds at the smallest Claude tier. (Fresh-folder
methodology confirmed: every arm gets an untouched copy of the kit; no model sees
another's leavings; the cold room is archived as a unit when the campaign closes.)

**Arms G/H/I — gemini-3.1-flash-lite / 3.1-pro-preview / 3-flash-preview: ✗ INVALID.**
agy silently ignored `--model` — all three replies state "I am running Gemini 3.5
Flash," and the flag leaked into the model's context (one reply muses about "a specific
command or context related to --model"). agy is pinned to its brain model; no
GEMINI/GOOGLE API key on this machine and no standalone gemini CLI. The three runs'
sloppy seed-following is CONFOUNDED — discarded, not scored. To test these models
properly: install Google's gemini CLI (and auth) or provide an API key for an
inline-AGENTS.md turn-1 harness.

## Tomorrow's queue (with Michael)

- **kimi-k2.6** — via opencode (his call: "would probably need opencode").
- **Local models via LM Studio's OpenAI-compatible endpoint** (inline-AGENTS.md turn-1
  harness): **gemma-4-31B**, **qwen3.6-27b** (memory: it ALWAYS reasons — give ≥2000
  max_tokens, answer arrives in `content`, thinking in `reasoning_content`),
  **nemotron** (non-thinking, per the Spin voice notes).
- **The Gemini 3.1/3 variants** — once a CLI or key path exists.
- **Cross-model session-2** (Opus reads Sonnet's memory — the model-swap proof) +
  flash turn-2.
## Next experiments (queued)

- Gemini turn 2 (answers → file setup fidelity) once agy continuation is worked out.
- A run from OUTSIDE the workspace (no inherited MCP config) for a true cold room.
- GPT-class model when available (codex CLI or via API).
- The substrate path: hand the kit to chattercode in a Docker sandbox and watch it
  flush out a repo end-to-end (Michael's original suggestion — needs a substrate
  session).
- Longitudinal: does session 2 actually read memory first? (The real portability test.)
