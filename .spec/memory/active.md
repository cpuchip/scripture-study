# Active Context

*Last updated: 2026-03-22*
*Archive: [archive/active-2026-03-22.md](archive/active-2026-03-22.md) — detailed records through Mar 22*

---

## Current State

**Last Sabbath:** March 22, 2026. Cycle "Infrastructure and Foundation" (Mar 18–22) declared good. Full record at [.spec/sabbath/2026-03-22-sabbath.md](../sabbath/2026-03-22-sabbath.md).

### Priorities
1. **Study** — Deep scripture study. "It keeps me in the spirit." Stewardship Pattern study COMPLETE + reflections written. Further calling-specific studies may follow.
2. **Model experiments** — Run same prompts through Haiku/Sonnet/Opus, evaluate quality. D&C 107 ratios for model-tier stewardship. Claude subscription likely needed next month.
3. **Debugging book** — DONE. Agans' "Debugging: The 9 Indispensable Rules" extracted to `books/debugging/9-indispensable-rules/` (17 chapter markdown files). Debug agent created at `.github/agents/debug.agent.md`. Connections mapped: Moroni 10:4 inverse hypothesis = falsification, scientific method = the 9 rules, Abraham 4:18 = Rule 9 (verify the fix), council moment = Rule 8 (get a fresh view). Analysis at `.spec/scratch/debugging-agent/main.md`. 2006 expanded edition (192pp, ISBN 9780814474570) available used ~$19 on AbeBooks.
4. **WS1 multi-agent framework** — Continue building. Next: Phase 3c (auto-routing + review queue).
5. **Desktop swap** — Decommission old desktop (migrate Plex, finalize). Do NOT repurpose.
6. **Server deployment** — App container on NOCIX. Domain rotated (Mar 22, confirmed working).

### Key Decisions (this cycle)
All settled decisions are in [decisions.md](decisions.md). New this cycle:
- **Covenant created.** `.spec/covenant.yaml`. Bilateral commitments. Added to session-start (Step 2).
- **Council moment added.** General principle for all agents. Abraham 4:26.
- **NOCIX server live.** Database migrated. App container not yet deployed.
- **R630 set down.** Existing Proxmox box works.
- **Old desktop: decommission only.** No repurposing.
- **Calling brain-app features: set down.** Paper, pencil, and existing ibeco.me practices for now.
- **Sabbath agent built.** Updates needed: scratch file support, model tiering.
- **WS-R and WS-P added.** Research and Planning as tracked workstreams. Organizing human and agent roles.

### Shipped (Mar 18–22)
- Data safety sprint ✅
- Server migration (NOCIX, 25 tables) ✅
- Disk crisis resolved ✅
- WS1 Phases 1–3b all shipped ✅
- Classifier hotfix ✅
- brain.exe WDAC fallback ✅
- Only Begotten study v2 ✅
- Sabbath agent ✅
- Stewardship Pattern study + reflections ✅
- Covenant + council moment ✅

*Full detail for all completed items: [archive/active-2026-03-22.md](archive/active-2026-03-22.md)*

---

## In Flight

### WS1 Phase 3c: Auto-Routing + Review Queue
- Phases 3a (agent pool + routing) and 3b (governance + token budgets) shipped Mar 21
- Next: auto-routing with human review of output
- Proposal: [.spec/proposals/brain-multi-agent/main.md](../proposals/brain-multi-agent/main.md)

### Squad Adoption Items (remaining)
- A2: Agent routing table — partially done (Phase 3a)
- A4: Reviewer lockout with model escalation
- A5: Response tier / model selection
- A6: Cost tracking
- Proposal: [.spec/proposals/squad-learnings.md](../proposals/squad-learnings.md)

### Progressive Trust Tracking (from stewardship reflections)
- Model capability experiments needed before assigning trust levels
- D&C 107 ratios framework: Haiku 1:12, Sonnet 1:48, Opus 1:96
- Details: [study/stewardship-pattern-reflections.md](../../study/stewardship-pattern-reflections.md)

### Pending Cleanup
- Delete `scripts/brain/internal/ai/tools.go` and `scripts/brain/test-spec.md`

---

## Plans Status

| Plan | Status | Notes |
|------|--------|-------|
| 15: Brain App Polish | Phases 1-2 DONE | |
| 16: Today Screen | Phases 1-3 DONE | Phase 4 absorbed into Plan 18 |
| 17: Proactive Surfacing | NOT STARTED | WS2 Phase 3 |
| 18: Widget Overhaul | Phase 1-2 DONE | Phase 3-4 PAUSED |
| 19: Brain App Ideas | Captured | Not started |
| Notifications | Phase 1 DONE | Phases 2-4 remaining |
| Data Safety | ALL DONE | |
| Overview | DECISIONS RECORDED | All guidance Qs answered |

---

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study or becoming entry? (Mar 4)
- What's the simplest version of the debugging book digestion that proves the concept? (Mar 22)
- Side quest: small classifier service on fermion/lepton for others? (Mar 19)
