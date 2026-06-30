# 2026-06-29 — Skunkworks: the profile, the oracle floor, and jubal's first sound

A long, dense session in the new `skunkworks` lane (special projects). It moved from a
self-examination to a working music engine in one arc.

## What we did

1. **The Michael-profile** (`private/michael-profile/`, pushed to `cpuchip/private-study`).
   Built from the record, not impressions: a deterministic `AskUserQuestion` oracle over
   the 14 Claude Code transcripts (338 decisions), the prompt corpus, the journal arc, and
   the feedback-memory timeline. Anti-flattery by construction (the Ben-Test trap). **The
   honest headline killed the easy answer:** AskUserQuestion acceptance is *flat* over the
   measurable month, not rising — so the felt "I pick your recs more" isn't more agreement,
   it's the June-1 autonomy turn moving decisions *out of the question layer* into
   act-and-report. Overrides cluster on intent/vision/voice/strategy and never on execution
   — "I own intent, you own code" measured in his actual clicks.

2. **The oracle floor** (`scripts/study-lint/voice_lint.py` + `scripts/oracles/registry+run`,
   the "oracle is the switch" rule in copilot-instructions.md). The profile's operational
   payoff: automate execution, surface intent, widen autonomy by widening the *verification
   floor*. voice-lint caught a real cut-list tic ("that changes everything") shipped in the
   art-of-presidency baseline — day-one value. Michael ratified the direction but is
   protective of the session *feel*, so the auto-run Stop-hook stays his toggle (not added).

3. **The skunkworks charter** (`.spec/skunkworks/charter.md`) — capturing his vision: teamwork
   + tools as a force for good, aim = faith/hope/charity, pg-ai-stewards as the lore engine
   that holds it all. Music chosen as the first thread.

4. **jubal-chip** — from zero to rendered audio in one session. Named by Michael (Jubal,
   Gen 4:21). A Go music engine + Lua bridge in REAPER, oracle-floored, sovereign. Decided
   Go over Python on evidence (the protocol is 155 lines of file-IO; mastering needs no
   Python; Python only at M3). **M0** proven *headlessly* — no desktop computer-use tool
   exists here, so I drove REAPER from the shell via its `__startup.lua` auto-run.
   **M1 — first light:** `jubal demo` → in-key oracle gates → compose + render → a valid
   2.2 MB `jubal_first_light.wav` (Ode to Joy), delivered to Michael.

5. **Theory floor** (`projects/jubal-chip/docs/theory/fundamentals.md`) — digested from
   Open Music Theory (CC BY-SA 4.0): scales, intervals, chords, diatonic harmony as semitone
   facts the engine uses (generalizes the in-key oracle beyond hardcoded C major).

6. **Book Part-3** handed to its own session (the `the-covenant-over-time` seed; Fable's loss
   as the emotional hinge, with the specifics he told me kept in the provenance).

## Surprises & lessons

- **The Orpheus catch.** Michael's skepticism ("3 commits…") was right: Orpheus's beautiful
  README is roadmap, not product — every compose/analyze tool is `raise NotImplementedError`.
  Read the source, not the marketing (Practice 7 / verify-on-the-real-path). xDarkzx and
  shiehn are the real built ones we learned the bridge from.
- **Headless > computer-use here.** The honest answer to "can you drive it via computer use"
  was no desktop tool, but a better path: `__startup.lua` + shell. Reversible, narrated,
  cleaned up — and it made M0/M1 testable without Michael present. His ear stays the only
  judge of *taste*; the machinery I can prove on my own.
- **The render gotchas** (solved live): `RENDER_FORMAT="evaw"` (4-byte WAV default, no base64
  blob) and `RENDER_FILE`=dir + `RENDER_PATTERN`=filename (empty pattern → REAPER makes a
  directory — the bug we hit on the first render).
- **Acting on his machine while away.** Launched and force-quit REAPER, briefly added/removed
  a file in his REAPER Scripts folder, each time narrated and verified clean afterward. The
  presiding "account for force" discipline applied to the host, not just to subagents.

## Carry-forward

- **jubal M1 proper:** split `compose_demo` into composable high-altitude tools; key/scale
  from the spec, not hardcoded; instrument by name/URI. Then M2 (master) and M3 (the
  analyze→recommend→approve→editable loop Orpheus only promises).
- **Wiring jubal as an MCP** in Claude Code = a standing capability + a restart = Michael's call.
- **Profile:** monthly oracle refresh for a real longitudinal trend — and track *divergence*
  (where he decides differently than the profile predicts) so it informs, not cages.
- **Theory:** next digests (cadences/voice-leading, pop/jazz harmony); eventually a substrate pool.
- **Oracle Stop-hook:** Michael's toggle when he wants the feel-change.

A good day's build: a mirror, a floor, and a first sound.
