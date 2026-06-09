# Lunch-block build: verify-suite (4 DR bugs) · P1.5 + A/B · r11 · CT2 §7.3 · arch refresh

**Date:** 2026-06-09 · **Mode:** dev (autonomous; Michael at lunch; first Fable 5 substrate session)
**Ratified upfront (AskUserQuestion):** verify-suite harness · #135 handler+A/B ·
voice-bridge spec-only · architecture.md refresh · "any ct2.x left to implement" (→ §7.3).
Soak paused at start, resumed at close. Root commits unpushed (Michael pushes root).

## 1. Verify-suite harness (`extension/scripts/run-verify-suite.ps1`, commit fad61a8)

First-ever test of "can the substrate be rebuilt from its repo files?" Answer: **no — 4 DR bugs**, found before the suite even fully ran:

1. **CREATE EXTENSION fails on any virgin DB** — `5d4-backfill-intent.sql` (in the
   bundled chain) RAISEs when the scripture-study intent doesn't exist, which on a
   fresh DB it never can. FIXED: NOTICE+RETURN guard; live ledger sha re-recorded
   (the established drift-fix pattern, NOT re-applied); takes effect in the bundle at
   the next pg image build. The harness carries an in-container sed shim until then.
2. **`stewards-cli migrate` applies lexically; history was chronological.** Same-object
   redefinitions break: ct2-2/ct2-7a2 (June) sort before k*/l13 (May) → a fresh
   bootstrap ends with l13's compose_messages, CT2 silently gone. The harness
   generates `extension/migration-order.txt` from git first-add dates and replays in
   that order. (Migrate-side fix = an order manifest; Michael's design call.)
3. **migrate would EXECUTE verify-*.sql as migrations** on a fresh DB (they're
   ledger-recorded but only because backfill recorded everything).
4. **Bundle-vs-replay conflicts:** CREATE EXTENSION pre-installs the *newest* bundled
   definitions, so 6 historical files fail on replay (DROP on extension-owned
   functions: 4c/4d/es11; old constraint vs new-shape seed rows: 5d5/5e2; old view
   shape: 2-7b1). Inherent "bundle + replay" tension — design question, not a today-fix.

**Suite results:** 197/203 replay ok in 68s; parity vs live = columns/views/triggers
**100% identical** (763/11/36); functions 311 vs 313 — 20 def mismatches (the live↔repo
drift inventory, unclassified; rerun with `-KeepContainer` and diff
pg_get_functiondef to classify), `ct2_echo_tool` live-only (CT2.3 smoke leftover),
record_cost_event overload tangle (explained by the es11 replay failure + a stale
9-arg live overload). 21/54 verify files pass hermetically (rest need data/bgworker —
triage table in `.spec/scratch/verify-suite-report-2026-06-09-1341.md`).

Also: migrate's header comment says failures let "subsequent files still attempt";
the code `os.Exit(1)`s on first failure. And lib.rs's lone `#[pg_test]` asserts
version 0.1.0 against a 0.2.0 crate. Small, real, recorded.

## 2. #135 P1.5 — research_codebase MCP handler + model A/B (commit 3911644)

- Handler in `cmd/stewards-mcp/heavyweight_tools.go` (the l6 wrapper pattern);
  `normalizeRepoURL` owns the allow-list contract (bare name → full clone URL).
  Unit-tested; e2e-proven through the REAL MCP stdio path (initialize → tools/call).
- **★ The overnight "flash fumbles the sandbox loop" verdict is OVERTURNED:** the
  fumble was the bare-name allow-list rejection, not model weakness. With the URL
  normalized, **deepseek-v4-flash ran the full loop clean** — sandbox_start → grep →
  targeted reads → sandbox_stop → correct, cited answer — for **$0.00** (free tier).
  kimi-k2.6 on the same question: deeper + line-precise, $0.84, 40s. **Flash stays
  the pipeline default.** (Environment-before-model-upgrade: a data point that
  rhymes with [[project_council_review_beats_gift_matching]]'s "measure, don't
  gift-match.")
- **Observation:** the kimi run blew through its $0.50 cost cap and completed at
  $0.84 — cap semantics are steward-quarantine-on-failure, not a hard ceiling.
- **r11 (adjacent fix, the stewardship rule):** the one-shot auto-verify trigger
  covered persona-%/redline%/brainstorm-% but NOT subagent-% (an r7 rebuild
  narrowing) → run 1 hung at completed|raw until manual verify; EVERY L.6 wrapper
  pipeline had the same 20-min-timeout exposure. `r11-subagent-auto-verify.sql`
  (based on the LIVE def, l13 lesson) live-applied + ledger-recorded; run 2
  self-verified = acceptance.
- `bin/stewards-mcp.exe` rebuilt — research_codebase reaches Claude Code on next
  MCP connect. `tool_defs.active` stays **false**: substrate-internal activation
  needs the bridge image rebuilt (it bakes stewards-mcp) → deferred to **P2** with
  the chattermax code-persona wiring, paired with Michael's next restart window.

## 3. CT2 §7.3 — self-editable base prompt, gated (ct2-7e, commit 3911644)

The full ratified shape, pure SQL, live-applied + ledger-recorded:
`propose_prompt_change` (double-gated: context_tools_enabled AND new
`allow_self_base_prompt`, OFF everywhere) → proposal row + `prompt-critic` one-shot
(qwen3.7-max, deny-*, json verdict) → SQL completion trigger stamps the verdict →
**human-only** `prompt_proposal_apply/_reject` + `prompt_revert` over a versioned
`agent_prompt_history` (`prompt_set` gives human edits the same ledger). Apply/reject
are deliberately NOT tool_defs — no agent path to ratification exists.

Smoked end-to-end with a real proposal on the researcher family: gate-off refusal →
flags on → tool visible → real propose → critic dispatched ($0.01) → **endorse**
verdict stamped by trigger → 3-pending cap refused a 4th → rejected the test
proposal → flags restored → **compose_tools hash = pre-7e baseline for all 50
families** (the §6 byte-identical proof; the only diff across the migration is the
new prompt-critic family itself). Apply/revert roundtrip proven on a scratch family.

## 4. Voice-bridge spec + architecture refresh

- `projects/spin/.spec/proposals/voice-room-bridge.md` (spin 6ee2d68, PUSHED):
  Pipecat client joins a chattermax room as Michael's voice — STT→room post,
  room→serial TTS queue, voice-per-persona, NO LLM in the pipeline; text stays the
  bus, voice is the human's codec. 4–12s/reply = D&D cadence. V0/V1/V2 phases;
  4 open questions for build-time ratify.
- `docs/architecture.md` (03a1c93): drift fixed (23-vs-65 tables), now 2026-06-09
  (70 tables/322 fns/43 pipelines/51 families), neighborhoods 8–12 added
  (orchestration, cost/governance, context engine, sidecars, ledger), deeper-reading
  table points at migration-order.txt + the verify suite.
- **#136 designed, not run** (spend unscoped): concrete A/B in the CT2 spec §CT2.4 —
  bacteriopolis-class workload, family-clone arms, queryable metrics, blind quality
  read, ~$1.40–2.20. Needs Michael's nod.

## Carry-forward

- **Michael:** push root (fad61a8, 3911644, 03a1c93 + spec/memory commits); nod or
  amend the #136 spend; ratify the voice-bridge build; decide the migrate-order
  manifest question (DR bug 2/3); next pg image rebuild picks up the 5d4 fix
  (then the harness shim no-ops and fresh-mode is the true DR test).
- **P2 (#135):** bridge image rebuild + `tool_defs.active=true` + refresh-tools +
  chattermax code persona — kick off together.
- **Open:** classify the 20 function mismatches (rerun suite `-KeepContainer`,
  diff defs); stale 9-arg record_cost_event + ct2_echo_tool cleanup candidates;
  cost-cap hard-ceiling question; prompt-critic-1 work item left as smoke record.
