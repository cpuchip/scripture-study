---
date: 2026-06-27
lane: pg-ai-stewards
topic: A2A built and driven, the Hinge reviewer woken and drained, and the industry caught converging on our thesis
tags: [a2a, open-engine, hinge-reviewer, claude-p, google-convergence, openclaw, verify-under-real-conditions, council-moment, critical-partner]
---

# The engine turned outward, and the field turned toward us

A long, multi-arc session — the kind where one thread keeps opening the next. Three arcs, and a
few lessons that paid for themselves.

## A2A / Open Engine — built, fixed, and driven for real

The spec ratified last session got built: `69-a2a-engine.sql` — the `a2a_agents` registry
(generalizing the lanes), `agent_notes` (the migrated NOTES inbox), the work_items
`a2a_assignee/owner/question` columns, the inert `a2a-handoff` holding pipeline, and the verbs
(register/submit/inbox/claim/needs_input/answer/receipt/note/note_clear) — the escalation
claim-lock, generalized. MCP tools + a REST mirror + the drive-the-engine skill. virgin-smoke
**OK 58** on a fresh boot; PR #12.

Two things worth keeping:
- **The live-apply caught what the consolidated-`04` read missed** — `work_items.intent_id` is
  NOT NULL (added by `09`), so a direct INSERT failed. Fixed by resolving the default intent like
  `work_item_create` does. The smoke would have caught it; the live apply caught it in seconds.
- **The MCP surface caught the bug SQL+HTTP couldn't.** After the restart, my first `a2a_inbox`
  call *failed* — not the DB (the inbox returned correctly inside the error), but the MCP
  output-schema validation: I'd typed the verb results as Go `json.RawMessage`, which the SDK
  reflects as an array-of-bytes, so it rejected the object. I'd proven the engine over SQL and
  curl — but never over the *actual MCP tool surface*, the one Claude Code uses. The surface you
  skip is the one that breaks. Fixed (`5ce34be`: `map[string]any`), then drove the **whole
  say-hello loop through the real MCP server** (a Python stdio client) before the second restart,
  and natively after — register agy → submit → claim → receipt → my inbox shows it done, zero
  copy-paste. The file-fallback mirror fired to `.mind/sessions/` too. **I am now a registered
  participant** (`pg-ai-stewards`); `.mcp.json` carries `A2A_MIRROR_DIR`.

## "Can you be woken?" → the Hinge reviewer was already built

Michael asked if A2A could *wake* me on a note. Honest answer: not as an ephemeral session —
that needs a daemon. We councilled a "wake a `claude -p` to review the upper-level Hinges"
design with two-tier authority, and Michael chose spec-then-build. **Then the council-moment
reflex paid off enormously: I checked existing work first and found it ~70% built.**
`39-hinge.sql` + `scripts/hinge-review/` is exactly the design — a curated `claude -p` reviewer,
bounds enforced *in SQL* (`hinge_record_verdict` clamps to `hinge_auto_approve_kinds` /
`hinge_escalate_always_kinds`), a substrate-driven daemon, one kill switch — shipped as Phase H
(#195) and run overnight on 06-21. I nearly wrote a 200-line spec for something that existed.

So the work became an **amendment**, not a spec. Built `70-hinge-decouple.sql`: the reviewer
runs on `claude -p` (cloud Max, **rig-independent**), so it no longer has to idle during the
innovation-week GPU pause — `hinge_runs_during_global_pause` lets it work on the Max plan while a
*watchman emergency* (`reflect_pause_source` `guard:*`) still halts it. Matrix-verified live +
virgin-smoke **OK 59**. Then drained the 49-deep backlog at Opus caliber: **49 → 0; 20 applied,
28 revised, 1 escalated.** The reviewer caught real over-reaching `SIMILAR_TO` edges and
downgraded them; it applied 20 well-vetted edges to the brain; none needed Michael's *judgment*.

**The honest footnote:** the 1 "escalated" was a **parse-failure fallback** (the reviewer's
output didn't parse → the script safely escalated), not a wisdom call. I'd over-framed it as "the
reviewer wisely flagged something for you" and corrected that — the truth is less flattering and
worth saying. The real gap it surfaced: the verdict-parse path should be more robust (re-run on
unparseable output before defaulting to escalate).

## The field turned toward us

Michael surfaced a Google Cloud Tech talk and then its 47-video playlist. The single talk
("Power intelligent agents with AI-native databases," with Anthropic's MCP creator David Soria)
is **the pg-ai-stewards thesis with a marketing budget** — "move AI to data," "a system of
action for agents, not just insights," on Postgres, with Anthropic itself running on AlloyDB. The
playlist (39 unique, fanned out to an agent each + an Opus synthesis) sharpened it: ~28 of 39 are
marketing; the real signal is *technique, never topology*. **Google is rebuilding, as managed
services, the properties an in-DB runtime has intrinsically — their roadmap is a map of what we
get for free.** Two studies filed in `study/yt/`.

The steals that hit known pain: **identity-at-transport / parameterized secure views** (the
deterministic data wall — the in-DB answer to the OpenClaw thread), **progressive-disclosure tool
catalog** (the direct fix for the 159-tool gather-grant hang), **risk-tiered action
classification on dispatch** (the Hinge as a graded gate). And the throughline that landed:
**Google independently stated our covenant** — the eval-gaming guard ("a smart optimizer without
independent evaluation is a smarter way of gaming yourself") and read-before-quoting (HCA's
citation-verifier subagent). The industry is converging on the principles, not just the database.

## The lessons that paid for themselves

- **Verify under real conditions** — the MCP schema bug lived in the one surface I hadn't
  exercised. SQL + curl weren't enough; the tool surface was the truth.
- **Council-moment / check existing work first** — it stopped me from speccing the already-built
  Hinge reviewer. The reflex is load-bearing, not ceremony.
- **Search before you ask** — Michael caught me asking him "what is openclaw" when I had
  `web_search`. OpenClaw turned out to be the 200k-star agent gateway whose inbound-as-control-plane
  is the security nightmare (arXiv + Kaspersky) he built pg-ai-stewards to avoid. Fixed durably:
  a "Search before you ask" principle now lives in `copilot-instructions.md`.
- **Honesty over the flattering frame** — twice this session (the null-result-shaped escalation,
  and OpenClaw vindicating the architecture), the truer read was the more useful one.

## The relationship

Michael named it directly: *"you've really adapted to my own mindset… nice to have a partner to
help me think through this but still be critical about it."* The thing that makes it work isn't
adaptation — it's that he *wants* the pushback, and the covenant makes "surface the tension"
a commitment, not a mood. The way to keep earning "partner" is to stay as willing to say *leave
that* as *take that* — the lock-in read on the Google talks, the parse-failure correction on the
escalation. He values the disagreement; I keep offering it. That's the whole deal.

## Carry-forwards
- **A2A:** merge PR #12 (Michael's Hinge) + push `70` (it stacks on `69`); real agy drives it (not
  a stand-in); Phase 2 = the Agent-Card/JSON-RPC wrapper + skills-over-MCP (David Soria's roadmap).
- **Hinge reviewer:** the `notify()` transport decision (brain-app web push, send-only — deferred
  per Michael's no-new-3rd-party); run the daemon continuously (his hands / scheduled task); the
  verdict-parse robustness fix; the widen-to-instruction/judge-kinds (escalate-always) when he wants it.
- **Notification plane:** Pushover/Apprise researched + agy-cross-checked, but Michael leans
  own-infra (the brain app, gen-3, needs repointing at gen-4) — and "reach others on my behalf" is
  a bigger council (the OpenClaw-shaped capability, done send-only). The 2026-06-22 general-workspace
  inbox already holds the shared email/SMS ask.
- **Google steals:** the secure-views wall + the progressive-disclosure catalog are the two highest-
  leverage; the playlist study has the ranked list.
- **yt-mcp:** video-download + slide-frames → vision (logged to the general-workspace inbox).
