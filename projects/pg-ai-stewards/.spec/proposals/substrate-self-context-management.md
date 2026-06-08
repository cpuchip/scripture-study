# Substrate proposal — Self-Context-Management (agent-governed context window)

**Status:** **CT2 CORE (§§1–6) COMPLETE 2026-06-08** — CT2.1 + CT2.2 + CT2.3 all shipped + verified (commits 1606675, d985b1c, 0fc6f32). The full self-management loop works: an agent calls `context_compress(handle)` → the dispatcher injects `_session_id` (the one Rust change, `exec_sql_fn_tool`) → the wrapper resolves handle→message → the lever sets the state → `compose_messages` honors it next turn (handles, lock-strip, pinned/compressed/muted, pressure line). Gated per-family by `agents.context_tools_enabled` (default off). Rebuild+restart clean. **Remaining: CT2.4** = the A/B (does it earn its keep on long runs — task #136). **§7 RATIFIED 2026-06-08** — see §7.6 for the as-ratified design: a **faceted-audience self-notes** model (selectors persona/kind/room/pipeline/global matched against per-dispatch facets — generalizes session/persona/global so it covers ALL substrate work, not just personas) + the §7.2 prompt split + §7.4 working tags = the **core build**; §7.3 self-editable base prompt is **greenlit but gated** for its own later pass. NOT built yet. — Original detail below:

**(superseded) CT2.1 + CT2.2 SHIPPED 2026-06-08** (commits 1606675, d985b1c). CT2.1 = SQL state model (`ct2-1-context-state-model.sql`). CT2.2 = the render honors it (`ct2-2-context-render.sql`) — handles, lock-strip (§4), pinned/compressed/muted, the §5 pressure line. **CT2.2 turned out to be PURE SQL, not Rust** — `compose_messages`/`compose_system_prompt` are plpgsql, so a `CREATE OR REPLACE` takes effect with no rebuild/restart; the "Rust rebuild" note in the plug-in table below was wrong about the RENDER layer. Built on the **l13** base (not k2 — a before/after md5 caught + reverted a regression where the k2 base would have dropped k6/k7/k8/k9/l1/l13 evolution); tools-off render is byte-identical, smoke-verified. **CT2.3 (expose the levers as agent-callable tools) IS the Rust+restart step** — `exec_sql_fn_tool` (tools.rs) calls `SELECT fn($1)` with only model args, no session, so the levers can't resolve `[ctx:handle]`→message_id; plan = inject `_session_id` into the sql_fn args (backward-compatible) + handle-resolution wrappers + tool_defs(sql_fn) + grants + rebuild/restart (Michael cleared restarts; ledger reconciled clean). Then CT2.4 (A/B). **§7 (durable self-notes + self-editable system prompt) is DESIGN-ONLY, UNRATIFIED, NOT built.**

## Decisions adopted (2026-06-05) — the 9 open questions resolved to the spec's proposed defaults
1. **N (lock cooldown)** = 3 turns.
2. **Handle** = short stable hash, `substr(md5(message_id::text),1,4)` → `[ctx:7a3f]` (stable across turns; same derivation in SQL + the Rust render).
3. **Mute** = tombstone (recoverable, never deleted) — the Atonement/forward-recovery shape.
4. **Pressure signal** = inform + light thresholds (the §5 example wording).
5. **Compress source** = reuse the existing engram (graduated L2); no new summarizer.
6. **Enablement** = per-pipeline/stage opt-in, off by default (like the critic). For chat personas, enable the `context-tools` group on `persona-turn` even though its other tools are disabled (context-tools is a distinct always-allowable group when enabled).
7. **Pin lifetime** = per-stage (each stage starts with a fresh context).
8. **Lock** applies to fold/mute/expand/clear; **pin/unpin are lock-exempt** (not thrash-prone).
9. **Tool-result clearing** = YES — add `context_clear_tool_result(handle)` as a distinct lever, and make stale-tool-result-clearing the automatic floor's default (cheapest, safest reclaim; keeps the call record). Anthropic ships this server-side (`clear_tool_uses_20250919`).

## Phased build plan (substrate C–F cadence)
- **CT2.1 — SQL state model** (live-apply, no restart, doesn't touch live personas): `stewards.messages` gains `context_state` + `locked_until_turn`; the `context_*` SQL functions (set-state + the lock); a `context_pressure(session_id)` helper; the handle derivation. Smoke: toggle a message's state in SQL, confirm the lock sets.
- **CT2.2 — Rust render path** (bgworker/dispatch rebuild → substrate restart; pause soak, expect a brief Starlet-cognition blip as persona-host reconnects): `compose_messages` honors the states, emits `[ctx:handle]` prefixes, **strips locked handles**, and appends the pressure line.
- **CT2.3 — context-tools MCP group** wired to the dispatch tool surface + `tool_defs`; enabled per-stage (incl. `persona-turn`). Lock enforcement on toggle.
- **CT2.4 — A/B verify** (don't assume): one long judgment-heavy dispatch with context-tools vs. automatic-only; measure tokens/cost/quality + thrash (toggle count, lock-hits) + whether mute ever dropped something needed.

**Note on pace (2026-06-05):** CT2's first *verifiable* increment is CT2.1+CT2.2 together (the SQL state model is inert without the Rust render that reads it), and CT2.2 restarts the live substrate that Starlet's cognition rides on. So it's a cohesive substrate batch best run with focus, not crammed at the tail of the session that already shipped the whole platform + multi-room. CT2.1 (SQL) is safe to start anytime.

**Original status:** SPEC ONLY — ratified to *spec now, build after the ai-chattermax MVP* (Michael, 2026-06-04).
**Author:** Claude Opus 4.8 (Claude Code), from Michael's driving-around idea + circuit-breaker refinement.
**Track:** substrate capability (sibling of the coder + critic harness), not on the ai-chattermax critical path.
**Builds on:** Batch K/L context engine (engram extraction, graduated rendering, `expand_message`, `re_extract_engrams`, `mark_engram_important`, `search_engrams`). Most of the *compression* already exists; this proposal adds *agent control*, a *pressure signal*, and an *anti-thrash circuit breaker*.

---

## Binding question

Can a dispatched agent govern its **own** context window — folding, muting, expanding, and pinning its messages on the fly — well enough to beat the substrate's existing *automatic* compaction on long, judgment-heavy runs, without thrashing its own memory?

## The problem

A dispatched agent's context = the system prompt (`compose_system_prompt`: covenant + intent + engrams) + the work-item message history. On long runs (multi-stage `code-pr`, deep studies, the coder's own iterations) the history outgrows the model window. Today the substrate compacts **automatically** (Batch K/L) — the context engine decides what to compress; the agent has **no say**. Automatic compaction is a good *floor* (it never overflows) but it's blind to what the *agent* knows is still live vs. truly done. The agent is the one entity that knows "I'm finished with that sub-thread" or "I'll need this spec for every remaining turn" — and right now it can't act on that knowledge.

## Design

### 1. Per-message context state

Each work-item message gains a context state (default `verbatim`):

| State | Rendered as | Set by |
|---|---|---|
| `verbatim` | the full message | default |
| `compressed` | its engram (graduated rendering L2) | agent `context_compress`, or automatic compaction |
| `muted` | a one-line tombstone `[ctx:7a3f — muted]` (recoverable, **not** deleted) | agent `context_mute` |
| `pinned` | full message, **exempt** from automatic compaction | agent `context_pin` |

`muted` renders a **tombstone, not nothing** — so the agent never loses track that the item exists and can pull it back. (Open Q3: tombstone vs. full omission.)

### 2. Addressable handles

Today the agent sees rendered *text*, not message IDs — so it can't say "fold *that*." The render path prefixes each foldable message with a short stable handle:

```
[ctx:7a3f] <message text…>
```

The agent calls context-tools with that handle. **Critical interaction with the circuit breaker (below): a locked message renders WITHOUT its handle** — so the agent literally cannot reference, and therefore cannot re-toggle, a message under cooldown.

### 3. The agent's levers (new MCP tools, exposed to dispatched agents)

- `context_compress(handle)` → `compressed` (render the engram, reclaim tokens).
- `context_expand(handle)` → `verbatim` (pull the full message back).
- `context_mute(handle)` → `muted` (tombstone; for resolved sub-threads).
- `context_pin(handle)` / `context_unpin(handle)` → protect a message from automatic compaction (the agent's "keep this no matter the pressure" — e.g. the spec/acceptance-criteria it needs every turn).

All operations take effect on the **next** turn's render (the current turn already built its context). Wire these as a `context-tools` group on the dispatch tool surface, gated like any tool group (off by default; enabled per-pipeline/stage — see Open Q6).

### 4. ★ The harness circuit breaker (Michael's addition — the anti-thrash lock)

The failure mode of agent-driven context management is **thrash**: compress → expand → compress flip-flop, burning tokens managing memory instead of doing the work. The breaker makes thrash *structurally impossible*, not merely discouraged:

- When the agent performs **any** toggle on a message, the harness sets `locked_until_turn = current_turn + N` on that message **and strips its handle from the render** for those N turns.
- With no handle, the agent **cannot reference** the message → cannot re-toggle it. The lock is enforced by *absence*, not by a refusal the model could argue with.
- After N turns the harness **restores the handle**; the agent may toggle again if still warranted.
- `N` is configurable (default proposal: **3 turns**).

This is the agent's own self-governance with a hardware-style rate limiter: a decision, once made, holds for a cooldown before it can be revisited. It is **involuntary and temporary** — distinct from `pin`, which is **voluntary and persistent**.

> Distinction worth keeping crisp: **`pin` = the agent chooses to protect.** **lock = the substrate enforces a cooldown the agent can't override** (it can't even see the handle). Pin is the agent's will; lock is the guardrail on the agent's will.

### 5. The pressure signal

`compose_system_prompt` adds a line each turn so the agent knows when to act and what it costs:

```
CONTEXT PRESSURE: 47,200 / 200,000 tokens (24%).
Foldable now: [ctx:7a3f] 3.1k · [ctx:9b2e] 5.8k · [ctx:c41d] 2.2k
(Below 60% you needn't fold. Above ~80%, compress or mute the least-relevant.)
```

The agent can't manage what it can't measure. The signal is the trigger; the levers are the response.

### 6. Hybrid — automatic floor + agent ceiling

Keep the existing automatic engram compaction as the **floor**: it still fires at a hard pressure threshold regardless of agent action, so the context never overflows and the agent is never *forced* to manage. The agent's levers are the **ceiling** — smarter-than-default folding the automatic pass can't know to do. `pinned` messages are exempt from the automatic pass. This is the safety property: if the agent does nothing, behavior is exactly today's.

## 7. Self-authored DURABLE memory + self-system-prompt (2026-06-05 expansion — Michael) — FOR RATIFICATION

Sections 1–6 govern the **current** window (fold/mute/pin existing messages). Michael
added two capabilities that go further: the agent should be able to (a) write
**durable** things into its own context that *survive* compaction and session
boundaries, and (b) shape its own **system prompt**. He flagged (b) as "dangerous but
could be super powerful." The key design move is that (a) and the *safe* part of (b)
are the **same mechanism**, and the dangerous part of (b) is split off behind a gate.

### 7.1 Durable self-notes — model writes AND removes (the Hermes loop)

A new persistent store `agent_self_notes` (keyed by scope): notes the agent authors
for its **future self**, rendered into `compose_system_prompt` every turn as a
"YOUR DURABLE NOTES" block — so they outlive both automatic compaction and the
session that wrote them.

- **Tools:** `remember(note, scope)` → add · `forget(handle)` → remove · notes
  surface in the prompt block with `[note:xxxx]` handles (same handle scheme as §2).
- **scope:** `session` (this run) · `persona`/`agent_family` (this character) ·
  `global` (the steward). Default `persona`.
- **Removable by the model — not append-only.** Michael's refinement: a model may
  park a fact to survive an imminent compaction and later *clear* it once it's been
  integrated elsewhere. So the agent has full add/remove over its OWN notes. This is
  the self-curating-memory loop behind self-learning agents like **Hermes** — the
  model maintains its own working knowledge, pruning as it goes.
- **Safety (it's the "safer area"):** these are the agent's own notes, never a
  guardrail. Bounded by a token budget + a count cap; every note is an inspectable,
  human-prunable DB row; the §4 circuit-breaker cooldown applies to add/remove so it
  can't thrash. Builds directly on the existing engram/`mark_engram_important`
  machinery (Batch K/L) rather than a new subsystem.

### 7.2 The system prompt, split in two

The agent prompt becomes **base + self-notes**:

- **Immutable base** — the guardrails (posture, the SILENCE gate, read-before-quoting,
  tool framing). The model can **never** write here.
- **Self-notes block** — §7.1, rendered as part of the prompt. So "the model edits its
  own system prompt" is *realized* by curating this appended block. ~80% of the upside
  (a self-tuning posture — "I tend to over-explain; be terser") at near-zero risk,
  because the base — and its safety rules — is untouchable.

### 7.3 Editing the BASE prompt — propose → ratify (the gated, dangerous path)

Direct self-edit of the base prompt is **not** allowed. The danger is specific:
drift/runaway (loosens itself each cycle), safety erosion (drops the SILENCE gate or
the quoting discipline), and a **sticky jailbreak** (a user message that rewrites the
prompt *persists*). Instead:

- The agent emits a **proposed diff** to its base prompt (a `propose_prompt_change`
  tool) — it never takes effect directly.
- A **critic agent + the human ratify** before it applies (D&C 88:122 council; the
  human is the **Hinge** — the [[full-context-shepherd-is-the-ceiling]] lesson).
- Every change is a **versioned, revertible** `agent_prompt_history` row; the base is
  always recoverable.
- Gated behind a per-agent **`allow_self_base_prompt` flag, OFF by default.**

### 7.4 Working tags — batch context by task (Michael, 2026-06-05)

Folding messages one handle at a time (§3) is fine for a stray item, but the real
unit of work is a **task**: an agent grinds on a todo across many turns + tool calls,
then moves on — and wants to sweep *all* of that out of context in one move. Tags
make a whole task's footprint a single addressable thing.

- **`context_set_tag(tag)`** — sets the agent's **working tag**. From then on, every
  new message + tool call the agent produces is stamped with `tag` (a
  `context_tags text[]` column on the message), until the agent sets a different tag
  or calls `context_clear_tag()` (untagged work resumes). One active working tag at a
  time; the current tag is echoed in the §5 pressure line.
- **Batch levers (operate on every message bearing the tag, in one call):**
  `context_fold_tag(tag)` / `context_mute_tag(tag)` / `context_expand_tag(tag)` /
  `context_pin_tag(tag)`.
- **The flow:** `set_tag("todo-3")` → do the work (everything auto-stamped) →
  finish → `mute_tag("todo-3")` reclaims the whole task's tokens at once →
  `set_tag("todo-4")` and start fresh. If todo-3 reopens, `expand_tag("todo-3")`
  brings it all back (mute is a recoverable tombstone, §1).
- **Interactions:** a tag batch op is **one** circuit-breaker event (§4) — it locks
  the tagged set together for the cooldown, not each message separately, so a deliberate
  task-sweep isn't penalized as thrash. `pin_tag` protects a whole task (e.g. the spec +
  acceptance criteria for the item in flight) from the automatic floor (§6). Tags are
  orthogonal to per-message state, so a single message can be both individually pinned
  and part of a folded tag (pin wins — it's the stronger, voluntary signal).
- **Auto-tagging is the key ergonomic:** the agent sets the tag *once* and keeps
  working; it does NOT have to tag each message. That's what makes "fold the last task
  away" a single cheap call instead of N.

### 7.5 Build split (Mosiah 4:27 — don't run faster than strength)

- **CT2 core (build first):** §§1–6 + §7.1 durable self-notes + §7.2 the self-notes
  block + §7.4 working tags. Safe, additive, high-value.
- **Separate, later, ratify-gated:** §7.3 base-prompt propose/ratify. Its own design +
  ratification pass; not bundled into CT2's first build.

### 7.6 ★ RATIFIED design (2026-06-08) — supersedes the §7.1 scope sketch

Walked with Michael 2026-06-08. Decisions:

**Build scope:** all of §7 (incl. §7.3, gated). **Notes cap:** generous — ~40 notes /
~4,000 tokens rendered per dispatch (count + token cap, hard-enforced). **Default
audience:** the authoring persona. **§7.3:** greenlit as a gated design (build later).

**The big change — a faceted audience model replaces session/persona/global.** The
original §7.1 scope enum doesn't generalize (Michael: it "gets messy" once you also
want per-room and per-work-type notes). So a note is targeted by **audience
selectors** matched against each **dispatch's facets** — one match rule instead of N
tiers, and it works for ALL pg-ai-stewards work, not just chat personas.

- **`agent_self_notes`** row: `note`, `handle` ([note:xxxx], §2 scheme), `audience`
  jsonb (selectors), `tags` text[] (free-form), `created_by`, timestamps.
- **audience selectors** (dimension→value, AND-combinable): `persona`, `kind`,
  `room`, `pipeline` (work-type), `global:true`, `session`. Default `{persona:<self>}`.
- **dispatch facets** — computed in `compose_system_prompt`: `session_id`,
  `agent_family`, `kind`, `pipeline` always; `persona`, `room` when a chat persona
  (threaded from persona-host). A new dimension later = a new facet key, **no schema
  change, no new tier.**
- **render rule:** a note renders into a dispatch's "YOUR DURABLE NOTES" block iff
  every one of its selectors matches that dispatch's facet (`global:true` always
  renders; a selector whose facet is absent on this dispatch → no match).
- **kind** is a new field on the agent/persona (`roleplay` / `code` / `librarian` /
  `general`, extensible) — drives the `kind` facet. A `{kind:code}` note is the
  shared per-kind pool Michael picked (any code persona writes it, all code personas
  read it; roleplay personas never see it).
- **tags** = labels for search/curation only; they do **not** gate delivery (a
  dispatch has no "purpose" facet), so `purpose:code-style` is organization, not routing.
- **Tools:** `remember(note, audience, tags)` / `forget(handle)`. Removable by the
  model (the Hermes self-curation loop). §4 cooldown applies to add/remove (anti-thrash).
  Human-prunable DB rows. Builds on the engram/`mark_engram_important` machinery.

**§7.2 prompt split** (build with core): immutable base + the rendered self-notes block.

**§7.4 working tags** (build with core): `context_set_tag` + batch `*_tag` levers, as in
§7.4 above. (Distinct from §7.6 `tags`: working tags address *current-window messages*
for batch fold/mute; self-note `tags` are durable-note labels.)

**§7.3 self-editable BASE prompt** — **greenlit, gated, separate later build pass:**
`propose_prompt_change` → critic + human ratify → versioned/revertible
`agent_prompt_history`; `allow_self_base_prompt` flag OFF by default.

**Build order (Mosiah 4:27):** ① core = §7.1 faceted self-notes + caps + the `kind`
field + dispatch-facet computation + §7.2 split + §7.4 working tags. ② §7.3, its own pass.

*(§7 RATIFIED 2026-06-08. The core is build-ready; §7.3 is design-ratified for a later pass.)*

## Prior art & validation (web search, 2026-06-04)

This design is not speculative — it sits squarely on a validated lineage. Searched the field; three findings sharpen the build.

**1. Anthropic already shipped the primitives — but mostly *automatic*, not *agent-driven*.** The Claude Developer Platform offers three first-party context tools ([context-management](https://www.anthropic.com/news/context-management), [cookbook](https://platform.claude.com/cookbook/tool-use-context-engineering-context-engineering-tools)):

| Primitive | Operation | Trigger | Our analogue |
|---|---|---|---|
| **Compaction** (`compact_20260112`) | whole-transcript summary, restart | server, token threshold | our automatic floor (Batch K/L) |
| **Tool-result clearing** (`clear_tool_uses_20250919`) | *sub-transcript* — drop old re-fetchable tool payloads, keep the call record + placeholder | server, token threshold | **we have no equivalent — see implication below** |
| **Memory tool** (`memory_20250818`) | model writes/reads files outside the window | **the model (tool call)** | engram store / brain |

The eval numbers justify the track: memory + context-editing **+39%** on agentic search; context-editing alone **+29%**; a 100-turn web-search run cut tokens **84%**. The gap our proposal fills is exactly the column that reads "server, token threshold": Anthropic's compaction/clearing fire *automatically* — the only *agent-decided* lever they ship is the memory tool. **Our `context_compress/mute/pin` with addressable handles is the agent-driven curation layer they don't expose.** That's the bet, and it's a real gap, not a reinvention.

**2. MemGPT/Letta is the academic root — and validates two of our specifics.** [MemGPT (2023)](https://arxiv.org/abs/2310.08560) framed "LLM as an OS" with virtual context paging, and its mechanism includes a **memory-pressure warning at 70% / 100%** that prompts the model to evict — i.e. our §5 pressure signal is a named, tested pattern, not a hunch. [Letta's memory blocks](https://www.letta.com/blog/memory-blocks) are addressable, size-capped, and either agent-editable **or developer-locked `read_only`** — and that read-only block *is* Michael's circuit breaker (§4): a slot the model structurally cannot toggle. The API's `exclude_tools` (never-clear list) is the same primitive at the platform layer. **Our lock + pin are both validated prior art.**

**3. Implication — split the cheap lever from the expensive one.** The search surfaced a distinction our spec collapses: **tool-result clearing (sub-transcript, surgical, safe — drop a re-fetchable file-read payload, keep the record) is a different and *cheaper* operation than message compaction (whole-transcript, lossy).** On a `code-pr` run the bloat is dominated by exactly this — large file reads and verify dumps the agent already processed. So add a fourth lever, `context_clear_tool_result(handle)`, that drops a stale tool *result* while keeping the `tool_use` record (the agent can always re-read the file). It's the safest, highest-yield reclaim and should arguably be the *default automatic* behavior of the floor, with the agent lever for the cases the threshold misses. (New Open Q9.)

## Where it plugs in

| Piece | Layer | Rebuild? |
|---|---|---|
| Per-message `context_state` + `locked_until_turn` columns; turn counter | SQL (message store) | live-apply |
| `context_compress/expand/mute/pin` SQL functions | SQL | live-apply |
| Render path honors states, emits handles, strips locked handles, appends the pressure line | Rust (dispatch/`compose_system_prompt`) | **yes** (bgworker/dispatch rebuild) |
| `context-tools` MCP group wired to the dispatch tool surface | Rust + tools registry | **yes** |
| Lock enforcement (set on toggle, release at turn N) | Rust (advance/render path) | **yes** |

So: SQL for the state model, a Rust rebuild for the render + tool surface + lock enforcement. Reuses the engram engine for the actual compression (no new summarizer).

## Verify it earns its keep (the lesson — don't assume)

The council A/B taught us to measure, not assume (m3-on-plan looked good, cost 2× for no gain; the plan-critic *did* earn it). Same discipline here:

**A/B:** one long, judgment-heavy dispatch (a full multi-stage `code-pr`, or a deep study) run **with** context-tools enabled vs. **automatic-only**. Measure: total tokens, cost, end-to-end quality, and — specifically — whether the agent *thrashed* (toggle count, lock-hits) and whether muting ever dropped something it later needed. Hypothesis: marginal on short runs, real on long ones. If it's marginal everywhere, it stays a per-stage opt-in, not a default.

## Creation-cycle framing (for the book audit)

This deepens three steps of the cycle and is worth a note in the blueprint audit:
- **Step 5 (Line upon Line).** The audit found a *granted-context* direction (the steward grants what the sandbox can't reach). This adds a third: **self-governed** context — the agent administering its own graduated rendering. Line-upon-line becomes something the agent does to itself, not only something done to it.
- **Step 10 (Consecration).** "Every token accountable to intent" moves from *enforced at dispatch* (spend caps) to *self-enforced in-flight* — the agent consecrating its own window.
- **Atonement (mute-not-delete).** Reversibility as a first-class property: nothing is destroyed, only set aside, and recoverable — the forward-recovery shape, applied to memory.

## Open ratification questions (resolve at build time)

1. **N (lock cooldown turns)?** Proposed default 3.
2. **Handle format** — short hash (`7a3f`) vs. sequential (`#12`)? Stability across turns matters (the agent may reference a handle it saw two turns ago).
3. **Mute = tombstone or full omission?** Proposed: tombstone (the agent shouldn't forget an item exists). 
4. **Pressure-signal tone** — prescriptive thresholds vs. inform-only?
5. **Compress source** — reuse the existing engram (graduated L2) or mint a fresh purpose-built summary?
6. **Enablement** — per-pipeline/stage opt-in (like the critic) or always-on above a context-length threshold?
7. **Pin lifetime** — does an agent's pin survive across pipeline stages, or reset per stage (each stage a fresh context)?
8. **Does the lock apply to `pin` too,** or only to the fold/mute/expand toggles? (Pinning is arguably not thrash-prone; leaning lock-exempt for pin.)
9. **Tool-result clearing as a distinct lever** (from Prior Art #3) — add `context_clear_tool_result(handle)` separate from `context_compress`, and make stale-tool-result clearing the *automatic floor's* default (cheapest, safest reclaim, keeps the call record)? Leaning yes — this is the highest-yield, lowest-risk piece and Anthropic ships it server-side as `clear_tool_uses_20250919`.

---

*Spec written 2026-06-04. Build deferred until the ai-chattermax MVP ships. Sibling track to the coder/critic harness.*
