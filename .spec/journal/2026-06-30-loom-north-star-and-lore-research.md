# 2026-06-30 — loom finds its north star (the substrate's agent fabric) + the lore research

**Lane:** general-workspace. **Arc:** lore research handed to pg-ai-stewards; loom marched v0.1→v0.3 and reviewed *itself*; Michael convened a council moment on what loom actually is.

## What was done

**Lore research → pg-ai-stewards OSS.** Adapted David Khourshid's (XState/Stately) *"Goodbye slop; welcome determinism"* + Stately Sketch (cloned `external_context/stately-sketch`, MIT) for Loreworks world-graphing. The thesis maps 1:1: the lore engine is a **static knowledge graph** (entities/edges + 3d-force-graph + read-only LOREMASTER) with the agentic shell but **no deterministic core** — no entity state, events, transitions, or world-advance. Proposal `deterministic-core-lore.md` committed to the OSS repo (`7627e14`, beside `world-graph-spec.md`): model lore entities as **statecharts** (JSONB, XState-serializable, Postgres-native) + a `world_events` log (= the missing deterministic core + temporal dimension); LLM proposes→engine validates; borrow `@statelyai/graph` + the Sketch simulator + Mermaid I/O; view stack keeps 3d-force-graph + adds **Vue Flow** (edit) + **Cytoscape** (analyze). The pg-ai-stewards session **registered it** (adjacent to `#291 world-graph`).

**loom v0.2/v0.3 — and it reviewed itself.** Shipped: the **`local` backend** (free models via llama-chip `:8090`; cloud+local `panel` council proven), **structured event streaming** (`--events`, observable — claude's tool calls/thinking surface live), and **`loom review`** (diff/files → fan a review across agents). The standout: **dogfooding `loom review` ON ITS OWN CODE found + fixed real bugs** — claude caught history-poisoning (the user message was appended to history before the request, never rolled back on error → two consecutive `role:"user"` messages corrupt the session), a `SessionID` data race, and (reviewing my own CoT fix) the **orphan-`</think>`** case (qwen/vLLM seed the opening tag in the prompt → reasoning→`</think>`→answer; my `stripThink` only handled matched pairs). The harness improves itself, human at the commit gate. Commits `a87f2ed`→`c092951`.

**The north star (Michael convened a council moment — "I'm not sure I know what we're building").** loom had gone muddy because it's **two things in one coat**: a *model provider* (the `local` backend — redundant with the substrate's own `:8090` dispatch) and an *agent runner* (claude/agy — a full autonomous agent the Postgres substrate **can't run itself**). Resolution: **loom is the substrate's hands — a uniform, *walled* interface for summoning intelligence, whose soul is running real agentic harnesses (Claude Code, agy) the substrate can't run, safely.** Two axes: **agency** (raw model ↔ agentic harness) and **trust/place** (local-direct ↔ sandboxed ↔ remote). The claude-runs-in-the-spawned-dir CWD is the **asset** (reach — digest a corpus into pg-ai-stewards from *outside*) AND the **risk** (a full-FS agent commanded by the substrate = host access). **Docker isolation = the lawful wall** (D&C 121 — *the walls are lawful*) that makes the north star tenable; **remote claude sessions** (Michael's new plan) = the same star at distance.

## Lessons / surprises

- **The recursive dogfood is real and valuable.** loom reviewing loom found a genuine bug in loom's *own* fix (the orphan-`</think>`). A self-hardening harness — the verify-real-path discipline turned into a tool.
- **Naming the two axes dissolved the identity confusion.** "A harness around harnesses that *also* runs raw models" was muddy; "agent fabric — provide *intelligence* (agents, not just completions), walled" is clear. The raw-model backend is a useful side, not the soul.
- **Isolation IS the presiding covenant made literal.** Delegating to a powerful agent needs a wall (D&C 121). Docker isolation isn't a feature — it's what makes loom-as-agent-fabric *safe*. [[feedback_presiding_covenant]]

## Carry-forward

- **Next build (direction ratified in the council moment): docker isolation** for the claude/agy backends — run the agent IN a container (workdir mounted, host walled off). **claude-in-docker first** (auth via mounted `~/.claude` or `ANTHROPIC_API_KEY`); **agy-in-docker deferred** (its Antigravity auth is gnarlier). The wall the north star rests on.
- **Michael's plan: pg-ai-stewards managing REMOTE claude sessions** via loom (the trust axis at distance) — keep the local-claude work; it's the reach.
- loom backlog: **panel role-routing** (doer→critic — the council-beats-gift-matching payoff), `--agent`/`--agents` flag consistency, `--events` through panel, session resume, a condenser for long reviews.

## Commits

`cpuchip/loom`: `a87f2ed` local backend · `b1a6b81` events · `ef6f3f1` history-poison/data-race fix · `c092951` review command + CoT fix. `pg-ai-stewards-oss`: `7627e14` deterministic-core-lore. Memory: `project_loom` (→ v0.3 + north star). Prior arc: `2026-06-29-loom-and-the-agentic-harness-arc.md`.
