# Project Archaeology — the 2025 archive

*Compiled overnight 2026-06-11→12 by Claude Fable 5, from the old-machine archives
Michael copied over: `external_context/old-code/` (repos), `external_context/
Code-old-session/` (VS Code chat sessions, Sept 2025–Jan 2026), `external_context/
.copilot-old/` (Copilot CLI sessions). "There is all of our old docs too to sift
through. have fun for me." — I did.*

## The timeline (attested by git + session dates)

| When (2025) | Project | What it was | Evidence |
|---|---|---|---|
| Sept 5–8 | **journal-mcp** | The first MCP server — task-based work journaling with AI integration. Ancestor of today's public MCP suite. | git: 21 commits |
| Sept 18 – Oct 2 | **forkirk** | Video-transcript quoter: search transcripts, embedded player, quotes with timestamp links, quote groups. Go + Vue + Mongo, dokploy deploys. Set down unfinished. | sessions + recovered repo (`old-code/astrotreks/forkirk`) |
| Oct 5 | — | The roadmap moment: "I'm feeling a bit overwhelmed by the amount of work left to do" — written to Copilot with the quote-groups roadmap attached, then at exactly **eight milestones**. The reply: "treating it as a checklist for launch will overwhelm you… treat it as your long-term vision." | `.copilot-old/history-session-state/session_aa43e58e…` |
| Oct 7–10 | **astrotreks** | The big integrated bridge simulator — one backend, many clients, each client a control surface of the ship (Helm throttle/heading sessions). | sessions + repo (no .git in copy) |
| Oct 22 → Nov 7 | **storygames** | Interactive story game with his 9-year-old's original characters; Dad reads, daughter chooses, AI weaves. TTS audio pipeline (`tts-builder/`), checkpoints. Alive again in 2026 (`projects/storygames`). | git: 14 commits |
| Oct 25 | **mobile-games** | The founding chat: "Ask me a few clarifying questions and lets get some documentation down and a plan together before we start coding." → ten questions → **Shields Down!** (byte-sized bridge sim). `SHIELDS_DOWN_PLAN.md` + `SHIELDS_DOWN_ROGUE_ONE.md` + the Nov 13 multi-agent stations experiment (`PARALLEL_WORK_PLAN.md` — Agent 2 Tactical, Agent 3 Engineering). | session `be90ec81…` + git: 60 commits (first 10-26) |
| Nov 18 → Dec 27 | **simple-games** | mobile-games' continuation (handoff day Nov 18 — last commit of one = first of the other). Flutter/Dart, mDNS host/join, WebSockets. Final wired roster: **eight networked games** (tic-tac-toe, connect 4, dots-and-boxes, desert dash, clock stoppers, hot potato, snake, wizard TD) + a ninth in the wings (wizard spellbook). Play-store packaging docs. | git: 90 commits |

**Adjacent life projects in the same archive:** **trek-experiments/
hubble-hometown-space-experience** — the *physical* bridge: `BRIDGE_MASTER_PLAN.md`,
console furniture designs, ESP32 console integration, KiCad boards, EmptyEpsilon setup,
grant research (the Mars-field Science Center workshop thread). **homeschool-keeper** —
offline-first homeschool tracking/journaling. **farm-store** — Farm Pickup Coordinator
("the communication layer" for small farms). **fambam** — family notes.
**storybook-web** — the website arm of storygames.

## What the archive verified for *Beyond the Prompt* (v4 audit closures)

1. **P1's founding conversation, verbatim** — the real prompt and the model's **ten**
   numbered questions (book said fourteen; corrected, and the four example questions
   in the book are now the real ones).
2. **P2's scar, fully** — the roadmap was exactly eight milestones at the
   staring-at-it moment, and "I wrote in the log that I felt overwhelmed" is his
   literal sentence from Oct 5. Bonus: the AI's reply that day *is* P2's lesson
   ("a plan is a map, not a debt"), received eight months before he wrote it.
3. **P6's roster context** — final wired count was eight networked games (book says
   seven — true during the arc it describes; Michael's call whether to update).
4. **The dream timeline** — pg-ai-stewards founded 2026-05-02; P1's "following
   spring" framing confirmed against storygames' Oct 22 start.

## Threads picked up later (the arcs continue)

- journal-mcp (Sept) → the public MCP suite (2026).
- astrotreks + trek-experiments (Oct) → the space-center workshop + ai-chattermax
  D&D holodeck's bridge ambitions.
- storygames (Oct) → running again with his daughter (2026, `projects/storygames`).
- Shields Down's multi-agent stations experiment (Nov 13: "You are Agent 2 — Tactical
  Station") → an early shadow of the substrate's multi-agent council patterns.
- forkirk's quote-with-source discipline → the gospel-engine/read-before-quoting
  ethos, rebuilt on firmer ground.

## Housekeeping notes

- astrotreks, storybook-web, and trek-experiments copied over **without `.git`**
  (git walks up to the workspace repo from those dirs — don't trust `git -C` there).
- forkirk's transcripts/assets intentionally not copied (Michael's call; the code,
  notes, and roadmap are intact).
- Raw voice corpora extracted to `.spec/scratch/voice-corpus-*.txt` (left
  **untracked** pending Michael's call on committing personal chat text — see
  `docs/voice-profile-michael.md` for the distilled profile, which is committed).
