# P2 restart bundle: research_codebase live + codewright persona + CT2.4 RUN 1

**Date:** 2026-06-09 (cont., same Fable-5 day as the lunch block)
**Mode:** dev — ratified "tackle together" menu; Michael picked all of A/B/C/D.
**Discipline call:** built A + B; HELD C (voice-bridge V0) + D (Spin offload) to a
fresh session (both large; C ends in Michael's mic test; stacking on the day's volume
= the Mosiah 4:27 strength limit I'd flagged in my own review). Soak paused→resumed.

## A — restart-window bundle (#135 P2), all verified

The one thing that structurally needed the restart Hinge. Sequence:
1. **pg image rebuilt** — bakes the 5d4 fresh-DB guard into the bundle, so the
   verify-suite's DR mode is now shim-free (the harness sed-shim is legacy).
2. **r12** — flipped `research_codebase` `tool_defs.active=true`. After refresh-tools
   the pg-ai-stewards self-catalog lists **30** tools incl. research_codebase.
3. **r13 — codewright**, a tool-using CODE chat persona (engineering twin of the R9
   librarian): `deny *` + allow research_codebase/read_corpus_parents/expand_message;
   `persona-turn-code` pipeline (kimi-k2.6, auto-verifies via the persona-% trigger).
   Read-only by construction. Room-join (mint key + grant + persona-host env) is the
   human/ops step, deliberately left to Michael.

**★ Two real bugs found + fixed (both known-gotcha families):**
- **grant ≠ catalog, one layer deeper:** research_codebase active in tool_defs but
  absent from `mcp_tool_cache` for the pg-ai-stewards self-server → substrate dispatch
  got "unknown tool (session invalidated)". Fixed by refresh-tools once the binary was
  real.
- **★ bridge.Dockerfile go.work bug (the root cause of 2 failed rebuilds):** `go.work`
  began listing `../persona-host` when that sidecar landed, but bridge.Dockerfile's
  COPY list never included it. So EVERY clean (`--no-cache`) bridge build failed the Go
  compile (`cannot load module ../persona-host`) and `docker compose up --force-recreate`
  silently re-ran the **stale 2026-06-04 image**. This is why my first two "rebuilds"
  didn't pick up the P1.5 binary (strings=0). A Dokploy-stale-build cousin: the build
  "succeeds" via cache while shipping old code. Added the COPY; clean build → binary
  has research_codebase (strings=3). **Lesson: when go.work gains a module, every
  Dockerfile that copies go.work must copy that module too, or GOWORK-mode builds break.**

**★ Codewright cognition proven TWICE, the second time the better proof:**
- Run 1 (tool down): codewright called research_codebase, got an error, and **refused
  to fabricate** — "Without a search I won't guess at the implementation," then offered
  clearly-labeled general patterns. The read-before-quoting discipline held in a
  brand-new 0-history persona. That refusal is worth more than a happy-path pass.
- Run 2 (tool live): full nested chain — codewright → research_codebase → flash subagent
  cloned+greped ai-chattermax (110s, $0.00 free tier) → codewright answered alice's
  bot-ping-pong question in room voice with **real file:line citations**
  (`internal/gateway/handler.go:180–182`, `hub.go:75–76`, `examples/echo-persona/main.go:98`).
  $0.0067 total. The engineering bot works.

## B — CT2.4 RUN 1 (#136): the informative null

One control/treatment pair on the bacteriopolis binding. Treatment = `research-ct2`
(research + context_tools_enabled + explicit context-lever allows) on a cloned
`research-write-ct2` pipeline.

| arm | cost | context-lever calls |
|---|---|---|
| control | $0.371 | n/a |
| treatment | $0.360 | **0** (fetch_url ×12, fs_read ×6, … but 0 context_*) |

**`context_tools_enabled` alone is INERT.** A long tool-heavy run, never once a
compress/mute/pin/remember. Cost delta = noise (0 levers). Two causes: the `research`
prompt never mentions managing context, and the multi-stage pipeline resets context
per stage so it never crossed the §5 pressure threshold. **Verdict is about
activation, not value** — stays opt-in/off-default (as built). RUN 2 (~$1) needs
prompt scaffolding OR a single-accumulating-session workload (research_codebase on a
big repo / a long code-pr implement) OR verifying the §5 pressure line actually fires.
Recorded in the CT2 spec §CT2.4. Reusable arms left in place (`research-ct2`,
`research-write-ct2`) for RUN 2.

## Commits (root, UNPUSHED — Michael pushes)
`93d593d` (r12+r13+Dockerfile fix) · `2f48c13` (CT2.4 RUN 1 finding) — plus the lunch
block's fad61a8/3911644/03a1c93/2fd881c still unpushed.

## Carry-forward
- **Michael:** push root; the codewright room-join (mint a codewright persona key in
  ai-chattermax → grant a room → add to persona-host CHATTERMAX_PERSONAS → restart
  persona-host) — then you can @-ask an engineering bot in chat; CT2.4 RUN 2 nod (~$1).
- **NEXT SESSION (held, large):** voice-bridge V0 (C; contract in hand from the
  research_codebase digest) · Spin offload Tier 3 (D).
- Today's opencode spend $1.93 / $12.
- Note: the becoming-mcp / fetch-md `fetch_url tool-error` lines in the bridge log
  during the A/B gather stage — transient web fetch failures, didn't block the run;
  worth a glance if they recur.
