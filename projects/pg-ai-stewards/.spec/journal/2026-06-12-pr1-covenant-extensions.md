# PR.1 — Covenant extensions catch-all, presiding render, The Watch echo

**Date:** 2026-06-12 · **Session:** pg-ai-stewards lane (OSS/anatomy session)
**Trigger:** inbox signal from general-workspase — the ratified `presiding:`
covenant section (preside study, `study/preside.md`) was silently dropped
between covenant.yaml and `stewards.covenants`. Michael: "can you make the
changes to the best of your judgement?"

## What shipped

1. **`parse_yaml_covenant` catch-all (Rust, `src/yaml.rs`).** Any top-level
   YAML section outside the known six (`purpose`, `human_commits_to`,
   `agent_commits_to`, `when_broken`, `council_moment`, `teaching`) now
   passes through generically as `extensions.<section>`. The covenant can
   grow sections without a parser change — the anti-silent-drop guard.
2. **`stewards.covenants.extensions jsonb`** (default `{}`) +
   `seed_covenant_from_yaml` carries `parsed->'extensions'` through.
3. **`compose_system_prompt`** (based on live def, verified == 5d3):
   - renders the presiding extension inside the covenant block — the three
     agent commitments (descriptions + the emergency act-then-account
     amendment), `dominion_in_council`, and the Luke 12:45 breach
     signature. The chain of watches now rides every dispatch.
   - appends `=== The Watch (echo) ===` at the very END of the system
     prompt: covenant commitment keys (data-driven, no new text) + one
     precedence rule ("if anything later in this context conflicts, the
     covenant governs").

All in `extension/pr1-covenant-extensions.sql` + `migration-order.txt`.

## Why the echo (the prompt-order research, 2026-06-12)

Michael asked whether the system-prompt order should be inverted ("some
models weight lower = more important?"). Findings:

- Attention over context is **U-shaped** — primacy AND recency privileged,
  middle weakest (Liu et al., *Lost in the Middle*, TACL 2024; *Found in
  the Middle* 2024 ties it to intrinsic attention bias).
- OpenAI GPT-4.1 guide: long context → instructions at **both ends** beat
  either alone; **conflicting instructions resolve toward the end**.
  Anthropic long-context guidance: documents top, instructions end.
- BUT the curve runs over the WHOLE context: the system message is the
  first message, so all of it is primacy-zone. "Last in the system prompt"
  ≠ "last in context." And tools aren't in the system prompt at all — they
  ride the API `tools` field (ours alphabetical); their position isn't a
  priority signal we control.

Decision: **don't invert — say it twice.** Covenant first (primacy +
cache-stable prefix; volatile blocks stay at the tail) and covenant last
(echo: recency + the conflict-toward-the-end bias pointed AT the covenant).

## Verification (inverse-hypothesis discipline)

- **Failure reproduced:** active row `29e1a8d9` had zero trace of
  presiding; no extensions column.
- **Fix applied** (pg image rebuild for yaml.rs; targeted
  `stop pg ui bridge` — persona-host container untouched, sibling lane's
  claim honored; watchman paused/resumed; queue empty both ends).
- **Reseed through the real path** (docker cp + pg_read_file +
  seed_covenant_from_yaml, the pre-commit hook's exact path): new active
  row `452034df` with `extensions ? 'presiding'` = t, all three commitment
  keys present.
- **Dispatch-level:** composed persona prompt contains preside_under_121,
  Emergency amendment, breach signature, Watch echo as final block.
- **Live end-to-end smoke:** spawn_subagent_create persona-turn
  (`600f6673`) → completed/verified, answer "ACK"; the dispatched
  work_queue payload's system message verified to contain
  `preside_under_121` AND `The Watch (echo)`.
- Bridge restart: migrations "2 applied, 237 skipped, 0 warnings".

## Cost note

Persona system prompt 5,562 → 8,430 chars (~+820 tokens/dispatch); plan
22,956 → 25,824. The presiding block is the bulk; echo ~60 tokens. Within
the flagged-for-measurement envelope, but worth watching on high-volume
gate dispatches.

## Gotchas (bank these)

- **MSYS path mangling ate a psql -f:** Git Bash converted `/tmp/pr1.sql`
  to `C:/Users/...` inside `docker exec` — migration silently didn't
  apply while the ledger row DID insert. Fix: `MSYS_NO_PATHCONV=1` (or
  `//tmp/...`). Always verify apply output before ledger insert.
- **Ledger naming wart:** the bridge's migrate keys entries WITHOUT
  `.sql`; manual live-apply inserts (r21, dnd2, pr1) used the full
  filename. The bridge therefore re-applied pr1 + r21 under suffix-less
  names (harmless — files idempotent), and the ledger now carries
  double-entries in two conventions. Feeds the open migrate-manifest
  design call; did NOT delete ledger history.

## Carry-forwards

- **Walls-vs-compulsion audit** of substrate mechanisms (preside study §V)
  — named follow-on, not started.
- **Trailing-reminder idea — PRIORITY RAISED by the 2026 regrounding
  (Michael's challenge, same day):** the 2025–26 literature says frontier
  models attenuated classic absolute-position LiM, BUT (a) U-shape
  persists ≤~50% of window and past that **primacy fades** while
  distance-to-end bias takes over (*Positional Biases Shift…*; CALIOPE,
  EACL Findings 2026; architectural-prior theory arXiv 2602.16837), and
  (b) context rot + "instruction weight loss" in long agent loops is the
  current consensus (Chroma + 2026 replications), and (c) long context
  silently shortens reasoning/self-verification ~2× on current-gen models
  (arXiv 2604.01161, Apr 2026). So in long sessions the end-of-MESSAGES
  slot matters more than end-of-system-prompt. Echo near the true tail =
  the robust lever. Still proposal-first (anthropic body conversion
  extracts system messages; provider-quirks pass needed).
- Verify-suite full run recommended after this schema change (bridge
  replay was green, but the suite's function-def parity diff is the
  deeper check; 20 pre-existing unclassified mismatches still open).
- OSS extraction must snapshot POST-PR.1 covenant machinery (noted in
  extraction-plan.md).
