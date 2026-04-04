# Active Context

*Last updated: 2026-04-04*
*Previous cycle archived: [archive/active-2026-04-04.md](archive/active-2026-04-04.md)*
*Hardware: Dual 4090s desktop (Mar 27). NOCIX server live.*

---

## Current State

**Last Sabbath:** March 22, 2026. Cycle "Infrastructure and Foundation" (Mar 18–22) declared good.

### Priorities
1. **Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" (third sabbath seed). Workstream: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md).
2. **Teaching** — 11-episode experiential arc (Option C). Teaching agent + repo scaffolded. Content not yet started. Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md).
3. **Brain pipeline** — Phase 4c shipped. Phase 4d next (REST API + execution + emergency stop). Billing model overhauled to premium requests.
4. **Claude Code integration** — NEW (Apr 4). Plan: add as alternative agent backend alongside Copilot SDK. Both cost spreading and gaining experience. Proposal forthcoming.

### Key Facts
- **GitHub Copilot billing:** 1500 premium requests/month ($40/mo Pro+). Multipliers: Haiku=0.33, Sonnet=1.0, Opus=3.0, GPT-4.1/4o/5-mini=0 (free). Tool calls within agentic sessions are free. 19% utilization on Apr 4.
- **Claude Code:** Pro $20/mo (includes Claude Code + Sonnet + Opus). Usage-based with 5-hour session window + weekly limits. 200K context. Project caching. Different billing model from Copilot.
- **Pipeline costs:** Research pass=0.33 + Plan pass=1.0 = 1.33 premium requests per entry.
- **Space Center:** Dream business — planetarium, space/science center, starship bridge simulator. Related repos on GitHub (cpuchip). Haiku-class models need this in prompt context (not in training data).

---

## In Flight

### WS1 Phase 4d: Pipeline REST API + Execution
- **Next up.** REST endpoints for pipeline operations, execution integration, emergency stop.
- **Phase 4a-c all shipped.** Maturity ladder, research agent (Haiku), plan agent (Sonnet), scenario enforcement, proposal generation. 24 tests green.
- Proposal: [.spec/proposals/brain-phase4-pipeline.md](../proposals/brain-phase4-pipeline.md)

### Claude Code Integration
- Alternative agent backend to Copilot SDK via CLI subprocess.
- Motivated by: cost spreading, gaining experience, different capabilities (200K context, project caching).
- Status: Research done, proposal written. [Proposal](../proposals/claude-code-integration.md).

### Brain Windows Service (Systray)
- brain.exe should auto-start on login, show systray icon.
- Proposal: [.spec/proposals/brain-windows-service.md](../proposals/brain-windows-service.md)

### ibeco.me Security Audit + Showcase
- OWASP Top 10, relay WebSocket, auth flows, adversarial testing.
- Becoming coach agent: customer-facing AI with trust boundaries.
- AI Skills Track 2 (Security Engineering).

### Gospel Engine Phase 1.5 (Ergonomics)
- Verse-level get, cross-ref retrieval, includeIgnoredFiles fix.
- Proposal: [.spec/proposals/gospel-engine/phase1.5-ergonomics.md](../proposals/gospel-engine/phase1.5-ergonomics.md)

### Gospel Graph Visualization
- study.ibeco.me — standalone site. Sequenced AFTER gospel-engine stabilizes.
- Proposal: [.spec/proposals/gospel-graph/main.md](../proposals/gospel-graph/main.md)

### Teaching Workstream
- 11-episode arc. Agent + scaffold built. Content not yet started.
- Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md)

### Study Workstream
- Next: "Zion in a Presidency" → "The Weight of Watching" → "Commission and Council"
- Proposal: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md)

---

## Deferred / Paused

| Item | Status | Revisit When |
|------|--------|------|
| Plan 17: Proactive Surfacing | Not started | WS2 Phase 3 |
| Plan 18: Widget Overhaul (Ph 3-4) | Paused | Agent infra proves out |
| Plan 19: Brain App Ideas | Captured | Natural pause |
| Notifications (Ph 2-4) | Remaining | After systray |
| Progressive Trust Tracking | Noted | D&C 107 ratios, model capability experiments |
| Squad A4/A5/A9 | Not started | After pipeline stabilizes |
| Review 4-Step & 11-Step Guides | Low urgency | Natural pause |
| Brain as Agent OS Platform | Researched | Ready for plan pass when prioritized |

---

## Plans Status

| Plan | Status |
|------|--------|
| 15–16: Brain App | DONE |
| 17: Proactive Surfacing | NOT STARTED |
| 18: Widget Overhaul | PAUSED (Ph 1-2 done) |
| 21: Gospel Engine | Phase 5 DONE (all enrichment) |
| Brain Pipeline (WS1) | Phase 4c DONE, 4d next |

---

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study? (Mar 4)
- What model should Claude Code use for brain pipeline? Sonnet for cost, Opus for quality? (Apr 4)
- How does Claude Code's usage-based billing interact with agentic tool-heavy workflows? (Apr 4)
