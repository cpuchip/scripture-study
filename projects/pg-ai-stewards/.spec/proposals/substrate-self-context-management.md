# Substrate proposal — Self-Context-Management (agent-governed context window)

**Status:** SPEC ONLY — ratified to *spec now, build after the ai-chattermax MVP* (Michael, 2026-06-04).
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

---

*Spec written 2026-06-04. Build deferred until the ai-chattermax MVP ships. Sibling track to the coder/critic harness.*
