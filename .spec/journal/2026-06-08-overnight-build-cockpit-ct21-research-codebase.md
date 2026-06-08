---
date: 2026-06-08
title: Overnight build — cockpit P1, CT2.1, research_codebase (Ammon night)
workstream: pg-ai-stewards
mode: dev (autonomous, unattended)
tags: [cockpit, ct2, code-persona, agentic-tools, ammon, inverse-hypothesis, mosiah-4-27]
---

# Overnight build — three deliverables while Michael slept

Michael handed me three builds before bed (~midnight): the cockpit CLI, CT2, and
the repo persona — "report on your progress when I wake up," plus "set a durable
timer in 5 hours just in case we run out of the 5 hour session." An Ammon night
(finish what you're handed, full stewardship). I ordered by safety × value and
kept every Hinge (merge/deploy/live-restart) his.

## What shipped (all committed to root — NOT pushed; Michael pushes root)

**① Cockpit P1 — read-only `stewards` Go CLI** (commit 2b2715b). New module
`cmd/stewards/`, pgxpool to the substrate (port 55433, like persona-host), zero
writes. Verbs: `project` (list + switch the sticky active project, ~/.stewards.json),
`board` (work-items by project, `--all` spans), `watch` (one item: stage/status/
cost/tokens/escalation/error + recent cost_events; slug or id-prefix), `cost`
(`--by project|model|day`, totals). go build/vet/test green (unit tests:
money/token formatting, rel-time, count summary, config round-trip). Live
read-only smoke: project counts match the work_items distribution; cost
surfaced deepseek-v4-flash at 363M tokens for $0 (the cheap tier the code
persona uses); `watch` resolved a real failed item + its event tail.

**② CT2.1 — self-context SQL state model** (commit 1606675). The SAFE, inert-by-
itself layer (`extension/ct2-1-context-state-model.sql`): messages.context_state
+ locked_until_turn; context_handle (substr md5 → [ctx:7a3f]); session_turn
(monotonic message count = the lock's time unit); the 5 levers — compress/mute/
expand (lockable toggles, lock = turn+3) and pin/unpin (lock-exempt, decision #8);
context_pressure (chars/4 proxy + foldable list). Live-applied (single tx) +
smoke-verified including the INVERSE: compress sets lock=turn+3; a re-toggle
within cooldown is REFUSED with the exact lock exception (the anti-thrash breaker);
pin bypasses the lock; pressure returned sane numbers; test message reset clean.
Nothing reads the state yet (compose_messages unchanged) so it cannot alter
behavior.

**③ Code persona P1 — research_codebase agentic tool** (commit 8417e7f). The
flagship: the first agentic tool for CODE, the exact l6 heavyweight-wrapper
pattern (agent + single-stage pipeline + read-only tool-grant denies + tool_defs
+ return contract). `extension/r10-research-codebase.sql`. Cognition PROVEN GREEN.

## The research_codebase smoke — three runs, inverse-hypothesis each time

This is the night's best story. Each failed run revealed the real next blocker
(not blind retry — a diagnosis each time):

1. **deepseek-v4-flash** (the spec's default "flash" tier): the WIRING works —
   tools dispatch, the coder sandbox starts, the allow-list correctly rejects a
   bad URL. But the model FUMBLED the orchestration: guessed a wrong-org URL,
   started a no-repo sandbox, then tried a raw `git clone`. → the cheap-tier
   question (spec Open Q2) is real.
2. **kimi-k2.6 + bare name "ai-chattermax"**: rejected by the allow-list, and
   kimi correctly STOPPED and returned low-confidence (the prompt's "say so and
   don't git-clone" rule working as designed). → revealed the allow-list wants
   the full clone URL, not a slug. (`repoAllowed` does a `Contains()` match on
   CODER_REPO_ALLOWLIST, default `github.com/cpuchip/ai-chattermax`.)
3. **kimi-k2.6 + https://github.com/cpuchip/ai-chattermax**: GREEN. Cloned the
   real repo, read `internal/gateway/*.go` + `internal/auth/*.go`, returned an
   excellent cited report — 8 findings, **12 file:line citations**, confidence
   high + 3 honest caveats — for ~$0.19. The return contract is exactly right,
   and the citations point at real code.

The pattern works: a model-as-tool researches a repo read-only and hands back
curated, verifiable, cited findings; the orchestrator never reads the whole repo.

## Deferred / handed off (Michael's call)

- **CT2.2 (Rust render)** — the next CT2 step (compose_messages honors
  context_state, strips locked handles, appends the §5 pressure line). Written-
  handoff, NOT started: it requires `docker compose build pg` + a live restart,
  and Starlet/Computer (chat.ibeco.me) ride on the dev substrate. That's a
  bin-3 op (could leave his live personas down), so it's his to run in a focused
  session — not an unattended call. CT2 §7 (durable self-notes / self-prompt /
  tags) remains UNRATIFIED, not built.
- **Code persona P1.5/P2** (task #135): the Go handler in cmd/stewards-mcp to
  make research_codebase first-class callable (maps short name → clone URL;
  flip tool_defs active=true + refresh-tools); the model-tier A/B (escalate the
  default off deepseek-flash, or scaffold); wire a read-only code persona in
  ai-chattermax.

## Findings that need Michael

- **Pre-existing migration-ledger drift** — a read-only `migrate --dry-run`
  surfaced 4 files changed-after-record (4a-cost-tracking, h1-1-general-research-
  intent, j10/j11 provider files) + 4 other live-applied-but-unrecorded files
  (cv12, r7/r8/r9) from earlier sessions. A bridge-startup migrate would exit-2
  on the drift. NOT touched — resolving drift is judgment (did each file change
  meaningfully?). ct2-1 + r10 left as clean PENDING migrations (apply+record from
  canonical bytes on the next migrate) rather than hand-recorded (which could
  create a line-ending drift landmine).
- **CODER_REPO_ALLOWLIST is empty in the live bridge** — falls back to the
  cpuchip/ai-chattermax default (fine for the smoke); set it explicitly before
  researching more repos.
- **The "durable" timer isn't truly disk-durable** — this environment returns
  "session-only" for every cron regardless of durable:true. The 4:52am one-shot
  fires only if the process is alive then (the 5-hour cap usually pauses-and-
  resumes the same process, so it should fire). Flagged, not hidden.

## Reflections

- The night was the stewardship tree made concrete: the cockpit (Michael drives),
  CT2.1 (the agent governs its own context), research_codebase (the orchestrator
  delegates the grind down to a cheap model). Three layers, one evening.
- Inverse-hypothesis paid off three times on one smoke — each failure was a
  signal, not a wall. The discipline ("reproduce → diagnose → fix the named
  thing → re-run") is what turned a fumbled deepseek run into a green kimi report.
- Mosiah 4:27 held the line: stopped at three verified deliverables rather than
  starting the CT2.2 Rust + live-restart deep in the night. The restart is the
  one thing genuinely his — guarding his live personas mattered more than a 4th
  commit.
- The soak was left running (additive SQL is safe; no pause needed) — nothing to
  resume.
