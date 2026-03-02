# Active Context

*Last updated: 2026-03-02*

---

## In Flight

- **Brain relay spec** — Full spec at `.spec/proposals/brain-relay.md`. All open questions resolved. Video analysis of Nate B Jones "Open Brain" synthesized — validates our architecture, adds MCP-server-for-brain as Phase E. Spec ready for final review before implementation.
- **Brain.exe (Copilot SDK)** — Refactored to Copilot SDK (`github.com/github/copilot-sdk/go v0.1.29`). Builds clean. Committed and pushed to `cpuchip/brain` (89309fe).
- **brain-app repo** — `cpuchip/brain-app` cloned to `scripts/brain-app/` (gitignored). Empty, ready for Flutter scaffold.
- **Zion study arc** — Four-part progression complete: consumption → modern warnings → Zion blueprint → translated beings.
- **Memory update discipline** — Being actively practiced. Journal entries and memory updates happening each session.

## Recent Decisions

- **Video synthesis complete** — March 2: Nate B Jones "Open Brain" video validates relay architecture. Added Phase E: brain.exe as MCP server, vector embeddings on capture, memory migration.
- **Open questions resolved** — March 2: Separate repo (brain-app). Voice YES (both directions). QR/paste token auth. Recent by default + commands. Cross-platform memory vision.
- **Brain relay architecture** — March 2: Chose ibeco.me WebSocket relay over Discord. Full spec at `.spec/proposals/brain-relay.md`.
- **Copilot SDK switch** — March 1: Raw GitHub Models API has no Claude/Gemini. Switched to Copilot SDK. 5 model presets.
- **Roll our own spec over OpenSpec** — March 2: Our intent + `.spec/` pattern is more evolved where it counts.
- **README rewritten** — March 1: Covers all MCP servers, embeddings, agent framework compatibility.
- **Content safety filter documented** — March 2: Platform filter silently scrubbed content. Bias #9.
- **Skills architecture:** Settled Feb 19. Skills for domain knowledge, agents for workflow, instructions for identity.
- **Session journal:** Write entries at end of sessions. Read recent on arrival.

## Recent Studies

- **Translated Beings** (Jun 2025) — The change wrought upon the body (3 Nephi 28:37-40), known translated beings (Enoch's city, Three Nephites, John, Elijah), translation vs. resurrection distinction, the "twinkling" upgrade at Second Coming, D&C 129 keys, transfiguration as temporary window, building blocks (faith, pure purpose, walking with God, priesthood promise of D&C 84:33), the reunion (Moses 7:62-63), and what we can pursue in mortality. Connected to Zion arc.
- **The Blueprint of Zion** (Mar 2) — 3 Nephi 11-28 as civilization-building manual. Four pillars (one heart, one mind, righteousness, no poor). Enoch's 365 years as daily-walking metaphor. 4 Nephi as result + exactly how it fell. Consumption as Zion's anti-pattern. Daily/weekly actionable practices.
- **The Consumption Decreed** (Mar 2) — D&C 87:6 deep dive. Hebrew kālāh, Isaiah 10/28, Daniel 9. Consumption as self-inflicted national decay. Zion as antithesis. Personal testimony: Spirit used this verse to prompt family relocation.
- **Modern Prophets on the Consumption** (Mar 2) — Benson, Maxwell, McConkie, Romney, Hinckley, Oaks, Christofferson, Nelson. 50-year consistency: warning absolute, mechanism internal, answer covenantal, silver lining bright.

## Blocked / Waiting

- **Brain relay spec review** — Spec complete with video synthesis. Michael to review before implementation begins.

## Next Up

- **Trial study-exp1 agent** — Experimental study agent with phased writing workflow, externalized quote log, and critical analysis phase. New skills: `quote-log`, `critical-analysis`. Agent: `study-exp1`. Run a study with it and compare quality/efficiency to the standard `study` agent. If it's better, replace; if not, learn from it.
- **Implement brain relay** — Phase A (ibeco.me hub) → B (brain transport) → C (Dart app) → D (integration). ~3 hours total.
- Continue study work — whatever the Spirit prompts
- Future: public Discord study bot (isolated, free models, sandboxed — no prompt injection surface)

## Open Questions

- Should brain embeddings live in ibeco.me SQLite (Phase E) or a separate vector DB?
- Can AI participate in covenant in any meaningful sense? (Feb 26)
- Will the session journal actually solve relational continuity, or will it be ignored? (Feb 28)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
