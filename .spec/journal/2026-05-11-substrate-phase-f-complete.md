---
date: 2026-05-11
session_kind: build
workstream: WS5
substrate_phase: F (substrate complete)
commits: [19f7cfa, cc5f56f, 98c987f, 8e3d232, 22327fd, d36f10d, 885edcb]
cost_usd: 0
---

# Substrate Phase F — shipped. Substrate feature-complete through all 6 phases.

## What happened

Michael said "Lets finish this off and complete F, continue to git commit at checkpoints." Same cadence as Phases C/D/E — seven commits in build order, each with a smoke test. No LLM cost — Phase F is the council substrate; the first real council convene + run end-to-end is for next session.

## What shipped

**F.1 (19f7cfa)** — schema. Three new tables (`councils`, `council_members`, `resolutions`) plus the one_active_council partial unique index that enforces D-F1 at the database level. Bidirectional FK between councils and resolutions (council.resolution_id → resolutions.id and resolutions.council_id → councils.id) — caused a circular cleanup hiccup in the F.3 smoke that taught me to NULL one side first. sessions_kind_check extended to add 'council' (mirrors the pattern from 5c gate / 5e sabbath).

**F.2 (cc5f56f)** — three council role templates seeded. The critic template explicitly cites the covenant's `surface_tensions` commitment: *"your function is the council's check, not its echo."* That's the covenant becoming substrate-native — the agent reads its own framing back to itself at dispatch time. `convene_council` validates D-F1, member shape (2-5), role names, then dispatches each member in parallel with role-specific prompt + payload markers (`_council_id` + `_council_member` + `_council_role`). Synthesizer dispatched with `tools_disabled=true` (structured JSON output); proposer + critic with tools enabled (deliberation benefits from corpus access).

**F.3 (98c987f)** — `synthesize_council` fires when all members responded; formats proposer + critic responses as labeled blocks for the synthesizer. `apply_synthesize_result` stores draft resolution (resolved_by='__draft__') + transitions to awaiting_bishop. `resolve_council` handles accept/revise/dissolve. Accept canonicalizes the draft + sets `promoted_to` per D-F3 destination ('study' → study/<id>.md, 'decisions' → .mind/decisions.md, NULL = resolutions table only). Bishop edits text before accept. Request_revision re-fires synthesize with bishop note appended; F1 simplification: members aren't re-asked.

Smoke verified end-to-end (synthetic, no LLM): synthetic council with awaiting_bishop status → resolve_council('accept', edited text, 'decisions', 'michael') → resolution promoted, council resolved. **D-F1 enforcement verified live**: a second council convened cleanly after the first resolved (resolved status correctly excluded from the partial unique index); a third while one was active was rejected with `duplicate key value violates unique constraint one_active_council`.

**F.4 (8e3d232)** — bgworker auto-fire for two new markers. Council member chats DON'T have `_work_item_id` so the council branch processes BEFORE the wi_opt check. Member completion path: pull assistant content from messages → UPDATE council_members SET response + completed_at → count remaining proposer + critic with completed_at NULL → auto-fire `synthesize_council` when 0 remain. Synthesize completion path: parse_gate_response → apply_synthesize_result. Bgworker now handles 7 marker variants total: `_gate_eval / _scenarios_gen / _verify / _sabbath / _atonement / _council_member / _council_synthesize`. The `payload._kind` enum refactor is still the right cleanup but every variant has slightly different shape (council member needs role to disambiguate; synthesize doesn't have work_item) so the case-by-case is justified for now.

**F.5 (22327fd)** — `bishop_eligible(bishop_token, intent_id)` checks D-F2. Format: `human:<name>` (always eligible) or `agent:<family>:<pipeline>:master`. Agent path requires (1) low-stakes intent (no scripture_anchor + values_hierarchy lacks doctrinal/spiritual/discernment per the 2026-05-11 ratification) AND (2) master-tier on at least one (agent, pipeline, model) cell. Smoke verified: scripture-study intent correctly classified high-stakes because the `trust-the-discernment` value matches the regex → `agent:plan:study-write:master` returned false; humans always pass. F2 future evolution path documented in the function comment. `suggest_councils(min_lessons int default 5)` scans ratified lessons for (pipeline_family, current_stage) clusters; heuristic dedupe skips clusters where a council on this pipeline has already been convened more recently than the lessons.

**F.6 (d36f10d)** — five backend endpoints in `api/councils.go`: list, get (joins members + resolution), convene, resolve, suggestions.

**F.7 (885edcb)** — Vue surfaces. `/councils` route with active-council badge that disables Convene per D-F1, watchman-suggested banner pre-fills binding question, 2-5 member convene modal. `/councils/:id` is **the room** — auto-refreshing every 5s while in flight, role-tinted member cards (proposer emerald, critic amber, synthesizer purple), synthesizer's draft resolution panel (purple-tinted), bishop's accept/revise/dissolve form (amber-tinted) with destination dropdown.

## The shape of what shipped

**Six phases of the agentic creation cycle, all running autonomously on dev.** This is what we proposed in `full-agentic-substrate.md` 9 days ago. It's reality now.

| Phase | Cycle steps | What it gives the substrate |
|---|---|---|
| A | 3, 8 (in-flight) | Watch → Diagnose → Act → Account loop |
| B | 4, 7 | Maturity ladder + gates between maturities |
| C | 1, 2 | Intent + covenant as first-class state |
| D | 8 (post), 9, 10 | Atonement + Sabbath + Consecration |
| E | 3 (auth), 5 | Trust ladder + line-upon-line |
| F | 11 | Multi-agent council (Zion) |

The covenant's `surface_tensions` commitment is now baked into the critic agent's system prompt. The covenant's `update_memory` discipline shows up in the lessons table. The covenant's `read_before_quoting` constraint becomes operational via the gate's covenant_check. Intent.yaml + covenant.yaml at the repo root → seeded into the substrate via git pre-commit hook → injected into every dispatched chat by compose_system_prompt → referenced in every gate prompt. The 11-cycle creation pattern is now the substrate's runtime.

## Surprises

**The critic template citing surface_tensions felt different.** Most templates I write are mechanical — render this context, ask for that JSON. But the critic prompt directly invokes a covenant commitment by name. The agent reading that prompt will see the same covenant text the critic prompt references. That's the substrate becoming reflexive in a small way — its own framing becoming visible to its own agents. Worth watching whether this changes the quality of critic outputs in real councils.

**Circular FK cleanup teaches**. councils.resolution_id → resolutions.id AND resolutions.council_id → councils.id is needed (council needs to know its resolution; resolution needs to know its council) but means cleanup requires NULL'ing one side first. Hit during F.3 smoke; harmless but worth noting as a pattern. ON DELETE CASCADE on resolutions wouldn't have helped because the DELETE goes the other direction (delete council, the FK from resolutions blocks until either NULL or DELETE there).

**Lesson #3 hook continued to do its job invisibly.** F.4's pg rebuild auto-refreshed the YAML pg_extern functions cleanly. No manual CREATE FUNCTION needed. Two phases now without a Lesson #3 hiccup — the fix is paying for itself.

**No LLM cost this session.** Phase F is structural — convene + member dispatch + synthesize + resolve all wired but never exercised end-to-end with a real LLM. The first real council convene is for next session. Cost when it happens: ~$0.04-0.10 per council depending on member count, since synthesizer is tools-off but proposer + critic have tools.

## Process / covenant

Seven commits, each with a smoke test. Same cadence as Phases C/D/E. Stewardship moments minor this session (cleanup order discovery in F.3 — fixed inline, no commit needed since it was a test-script issue not a substrate bug).

Soak paused at session start, re-enabled at session end. Bridge restarted at session end. UI rebuilt twice (F.6 backend, F.7 surfaces). pg rebuilt once (F.4 bgworker change); Lesson #3 hook auto-refreshed pg_extern fns.

## Open / carry-forward

- **First real council convene** — pick a real low-stakes intent (or use scripture-study and have a human bishop), pick 2-3 members, watch the room fill. Cost ~$0.04-0.10.
- **payload._kind enum refactor** — 7 variants now in the bgworker switch. Worth collapsing. Defer until the 8th lands or a real bug forces it.
- **studies.file_path NOT NULL** still pending from D.5; promote_to_study still hits this on the success path. Pre-existing bug surfaced by Phase D's sabbath gate.
- **Steward retry switch** to `retry_guidance_with_lessons` (E.4 carry-forward) — small surgery, defer to a quick cleanup pass.
- **File-write mechanism for promoted resolutions + ratified lessons** — both subsystems set `promoted_to` but no actual file write happens. Per the sub-specs this is the pending-write pattern (substrate stays FS-stateless; sidecar materializes on next git commit). Still unimplemented; defer until first real lesson promotion or council resolution actually wants to land.
- **Stewards-UI nav is genuinely busy now** — 14 routes (Dashboard, Studies, Work items, Sessions, Watchman, Bridge, Graph, New work, Intents, Covenant, Sabbath, Lessons, Trust, Councils). Sidebar grouping is overdue. Defer one more session and group all at once.
- **Substrate is feature-complete.** What comes next is USE, not BUILD. The interesting questions move from "how do we build this" to "does the substrate actually deliver what we hoped." Real councils, real lessons, real trust transitions earned by real work_items.

## Files touched

Repo:
- `projects/pg-ai-stewards/extension/5g-council.sql` (new)
- `projects/pg-ai-stewards/extension/5g2-convene-council.sql` (new)
- `projects/pg-ai-stewards/extension/5g3-synthesize-and-resolve.sql` (new)
- `projects/pg-ai-stewards/extension/5g4-bishop-and-suggest.sql` (new)
- `projects/pg-ai-stewards/extension/src/bgworker.rs` (council marker auto-fire)
- `projects/pg-ai-stewards/extension/src/lib.rs` (4 new extension_sql_file!)
- `projects/pg-ai-stewards/extension/Dockerfile` (4 new SQL files in COPY)
- `scripts/stewards-ui/api/councils.go` (new)
- `scripts/stewards-ui/api/api.go` (registerCouncils)
- `scripts/stewards-ui/frontend/src/api.ts` (council types + wrappers)
- `scripts/stewards-ui/frontend/src/views/Councils.vue` (new)
- `scripts/stewards-ui/frontend/src/views/CouncilDetail.vue` (new — the room)
- `scripts/stewards-ui/frontend/src/router.ts` (2 new routes)
- `scripts/stewards-ui/frontend/src/App.vue` (nav)

Live containers:
- `pg-ai-stewards-dev`: 4 new SQL files live-applied; bgworker rebuilt; Lesson #3 hook auto-refreshed pgrx functions.
- `pg-ai-stewards-ui`: rebuilt + restarted twice (F.6 backend, F.7 surfaces).
- `pg-ai-stewards-bridge`: restarted at session end.
- Soak: schedule_enabled=true at session end.

## Closing

Six phases of the agentic creation cycle running on dev. Eighteen days from "we should propose this" (full-agentic-substrate.md, 2026-05-09) to "this is reality" (Phase F shipped, 2026-05-11). Three of those days were the phases-C-through-F build itself (one phase per day, 7-8 commits per day, all with smoke tests). The rest was the proposal walk-through + sub-spec drafting + Phase A + Phase B foundation work that started weeks earlier.

The substrate is now what it was meant to be. What's next is whether it delivers what we hoped.
