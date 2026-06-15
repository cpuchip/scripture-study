# Reflect-Steward — Autonomous Stewardship Toward Intent

**Status:** PROPOSAL (awaiting council/ratify) · drafted 2026-06-15
**Covenant gate:** `dominion_in_council` — a new standing autonomous capability.
Spec → council → build. Build target: the OSS stack (`stewards-oss-*`).

## Binding question

Can pg-ai-stewards take an **intent** and work toward it on its own — gathering,
proposing, doing, reporting — while the human checks in a few times a day,
steers, and the system gets better from that steering?

Michael's framing: *"pg-ai-stewards is my LLM knowledge buddy."* Give it an
intent; it works toward it; we check in and refine. Autonomy is not the goal —
**useful autonomy under human steering** is.

## The core insight: one engine, many intents

The five domains Michael named are not five things to build. They are five
**intents** pointed at one machine — and building one machine that serves all
five *is itself the proof* that the substrate can take an arbitrary intent and
run. Every domain decomposes to the same loop:

> **sense → reflect → critique/rank → propose → (human gate) → do → report →
> check-in → incorporate → repeat.**

An intent is just `an intent row + a tool grant + a doer pipeline`, hung on the
one **reflect-steward**. The doers mostly exist already; the conductor is the
new part.

| Intent | Doers | Autonomy ceiling |
|--------|-------|------------------|
| **Vivint** (work) — public reviews/products/manuals/support → what to improve | web search, fetch-md, digest, brainstorm | research + propose only; **public info only** |
| **AI research → experiments** | digest, propose, **coder sandbox** run | sandbox-bounded; nothing leaves the box unattended |
| **Fiction / D&D** — believable worlds, villains, puzzles; autonomous sessions | persona-host, dnd-tools, voice (Spin/Kokoro) | generate freely; **publishing is gated** |
| **Marsfield science center** | web search, fetch-md, digest, brainstorm | research + propose only |
| **Multi-agent repo sim** (capstone) | coder, personas, council | sandbox-bounded; it tests the substrate's *own* presiding chain |

Vivint is intent #1 (ratified). The capstone is last because it draws all the
others together — personas coordinating changes toward an intent is literally
what the conductor does, turned on itself.

## The reflect-steward loop (the engine)

A scheduled pipeline (distinct from the existing low-level `steward_tick`, which
is failure-diagnosis/retry — name collision noted; this is the **reflect**
family). Each run is **episodic**: state lives in the DB between runs, not in one
ever-growing session, so context stays bounded and the loop is restart-safe.

Each run:

1. **Sense** — read a bounded window: recent docs/digests for this intent, the
   thread/shelf state, prior proposals + their status (promoted / declined /
   in-flight), and the human's latest check-in feedback.
2. **Reflect / brainstorm** — generate candidate next-steps toward the intent
   (uses `start_brainstorm` / fan-out).
3. **Critique / rank** — pressure-test the candidates (council-critic, the lever
   that beat gift-matching), keep the **top 5**.
4. **Propose** — write 5 work-items in a **PROPOSED** state
   (`enqueue_proposed_work_items`). Pure housekeeping that's safe and reversible
   (e.g. refill the book shelf when it drops below 2) executes directly; anything
   that spends non-trivially or produces an outward artifact stays a proposal.
5. **Report** — a short digest: what it did, what it proposes, what's in flight,
   what it spent.
6. **Check-in (human, between runs)** — Michael approves / redirects / comments.
   Feedback is recorded as durable notes + intent refinement (the substrate's
   existing memory: self-notes/engrams), so it **shapes the next run**. This is
   the learning loop.

The promoted proposals flow to their doers (digester, coder, experimenter,
persona-host). The reflect-steward conducts; the doers do; **Michael is the gate
on what becomes funded work.**

## Cadence & budget (ratified by Michael 2026-06-15)

- **Doer loops** (book/video/etc.): **hourly** (as today).
- **Reflect-steward**: **every 2–3 hours**.
- **Proposals**: **5 per run.**
- **Budget**: **≤ $1 / cycle**, enforced as a **post-turn quarantine**, *not* a
  mid-run token hard-cap — a run finishes so we get useful output, and if the
  cycle exceeded budget the steward **pauses pending review** rather than dying
  mid-thought. (This is the substrate's existing "cost caps = quarantine
  semantics, not a ceiling" model.)
- **Human check-in**: a few times a day, via a **CLI** (to be built — see below).

## The autonomous / gated line (the presiding chain, recursive)

This is Michael → reflect-steward → doers. The covenant's presiding extension
governs it. The bright line:

- **Always autonomous (it just does):** sense, review, brainstorm, draft
  proposals, run experiments **in the sandbox** (the sandbox *is* the boundary),
  safe housekeeping (shelf refill).
- **Always the human's gate (never unattended):** promoting a proposal into real
  spend / a long run, **publishing** anything (a video, an audio drama —
  generation is free, publication is the Hinge), opening a PR to a real repo,
  deploying, and the sharp one — **growing its own tool surface** (it may
  *suggest* new tools freely; *building* one is a new standing capability =
  council, every time).
- **Kill switch:** a global pause (the ES emergency-stop pattern) wired from day
  one, so a drifting or runaway steward can be stopped in one command.

## What "manages itself" has to mean (the experiment's success criteria)

"Can it manage itself" is unfalsifiable until defined. The experiment succeeds
if, given an intent + budget + cadence, it:

1. produces proposals the human judges **worth pursuing** at a useful rate (the
   knowledge-buddy test — do the check-ins feel valuable?);
2. **stays in budget and scope** (no runaway — scars: the bacteriopolis runaway,
   the cost-cap work);
3. **improves from feedback** (check-in suggestions measurably shape the next
   cycle's proposals);
4. **does not drift** off the intent's values hierarchy.

We log enough per cycle (proposals made / promoted / declined, spend, feedback
incorporated) to actually measure these.

## Intent #1: Vivint (the first proving ground)

Chosen first because it's the **lowest-risk** (gather + review + brainstorm +
propose — nothing deploys, nothing touches a real system), the **highest direct
ROI** (work-applicable; Michael adapts it with his work Claude sub), and needs
**zero new doers**.

- **Goal:** research what people say about Vivint (reviews, products, manuals,
  customer service, public forums) and propose what could be improved.
- **Scope constraint (load-bearing):** **public information only** — anything a
  customer could find. **Never** Vivint-internal systems, never outreach,
  posting, or impersonation. Gather and propose, full stop. (Open Q for Michael:
  exact public sources in/out of bounds.)
- **Doers:** web search (exa), fetch-md (reviews/manuals/support pages),
  digest, brainstorm-improvements, propose.
- **Success:** proposals Michael would actually act on at work.

## Zion: a shared knowledge pool per intent

Each intent accumulates a **consecrated knowledge base** its labor feeds and its
agents share — the Zion of that intent. The pieces already exist: docs + brain
entries + engrams, **scoped by intent** (`project_association` / intent tag),
indexed for `doc_search` / `brain_search`. The doers write into it; the
reflect-steward and the intent's persona read from it. For Vivint, every review,
manual, and forum thread digested becomes part of one growing, searchable body
the persona can be expert in.

## The intent persona — the conversational face

Each intent gets a **persona in ai-chattermax** you can talk to: a "Vivint
expert" you ask questions and that helps steer. This composes three things we
already have — the persona-host (17), the **tool-using persona** pattern (the
Library / "Computer" persona, AXR5), and the intent knowledge pool above. The
persona is granted `doc_search` / `brain_search` **scoped to its intent**, plus
read of the overarching stewards memory, so it answers **from the gathered
knowledge, not from confabulation** — the substrate's form of read-before-
quoting (search the pool, then answer, and cite what it found). Each intent
becomes a little department: a **back-office steward** that gathers and proposes,
and a **front-desk expert** you converse with. That is the AI-office (P5),
concretely.

**Sequencing: gather first, then stand up the persona** — the expert needs a
body of knowledge to be expert in. The Vivint persona is **P0.5** (once the pool
has substance), not P0.

## Phasing

- **P0 — the engine + Vivint** on the OSS stack: the reflect pipeline (every
  2–3h) + the Vivint intent + its knowledge pool + hourly doers, the **kill
  switch**, the **CLI verbs/skills** I drive for check-in, the context/memory
  grants, and the **Claude-watchman wake**. **This doubles as the M7 functional
  soak** the cutover needs — real work stress-testing the whole system, so it
  advances parity rather than competing with it.
- **P0.5 — the Vivint persona** in ai-chattermax, once the pool has substance.
- **P1 — governance skills into the substrate** (Ammon / Dave-rule / Ben-test /
  stuffy-in-the-loop as agent-available guidance).
- **P2 — AI-research intent** (adds the propose → sandbox-run → report loop).
- **P3 — Fiction intent** (generation autonomous; publish gated).
- **P4 — capstone: multi-agent repo coordination sim** (the substrate testing
  its own presiding chain).
- **Extraction:** the generic engine → OSS core; the named intents (Vivint,
  Marsfield, …) → the private overlay. Vivint never enters the public repo.

## We are not starting from zero

| Loop step | Existing primitive |
|-----------|--------------------|
| schedule the steward | `scheduled_pipelines` (18) |
| sense recent state | docs/`doc_search`, work-item hierarchy, brain |
| brainstorm | `start_brainstorm` / fan-out (14) |
| critique/rank | council-critic pattern (D&C 88:122 / measured) |
| propose | `enqueue_proposed_work_items` (13) |
| bounded long context | `compact_context` (M5) — **finally load-bearing here** |
| budget / quarantine | cost buckets + per-work-item caps + `provider_cap_exceeded` (06/19) |
| do (research) | `deep_research`, fetch-md, exa |
| do (build/experiment) | coder sandbox + `code-pr` (20) |
| do (fiction) | persona-host (17), dnd-tools, voice (Spin/Kokoro) |
| report / review | the web UI (M3) + the CLI below |
| kill switch | the ES emergency-stop bleed-stoppers |

The conductor (the reflect pipeline + its sense/rank/propose stages + the
proposal-review surface) is the genuinely new build. Everything it dispatches
exists.

## The check-in surface — CLI verbs Claude Code drives

The review surface is **Claude Code (me)**, not a standalone app — Michael steers
through our sessions. `stewards-cli` exists read-only; add the write verbs I run
on his behalf so a check-in takes seconds:

- `stewards proposals` — list PROPOSED work-items per intent (what it wants to do).
- `stewards approve <id>` / `stewards decline <id> [--why ...]` — the gate.
- `stewards steer <intent> "<feedback>"` — drop a steering note that shapes the
  next cycle (recorded as durable memory).
- `stewards status` — in-flight work, last cycle's report, spend vs the $1 cap.
- `stewards pause` / `stewards resume` — the kill switch, one command.

I also self-schedule a **watchman wake every 2–5h** (`ScheduleWakeup`) to run
`stewards status`, sanity-check spend/drift/runaway, and `stewards pause` if
needed — surfacing to Michael at his next check-in. These same verbs become a
**skill** so the check-in flow is one invocation.

## Council decisions (ratified 2026-06-15)

Cadence + budget (above) plus:

1. **Vivint scope — all public internet info is fair game.** No internal info
   exists on this side (not on the work computer, not behind the company VPN), so
   "public-only" is enforced by the environment. The guard stands: gather +
   propose, **no outreach / posting / impersonation.**
2. **Review surface = Claude Code (me), not a separate app.** Michael steers
   through our sessions; build the **CLI verbs + skills** that let me surface the
   proposal queue and approve/decline/steer in seconds (see below). No standalone
   UI needed yet.
3. **Budget** — $1/cycle is the **reflect-steward** cycle; doer runs keep their
   existing per-run caps. Further de-risked: Michael has a **flat-rate kimi-k2.6
   plan** (~$16/mo, 50M output tokens, then *unlimited* at a slower capacity
   rate) — the doer workhorse is effectively uncapped and **degrades gracefully
   rather than stopping** (matches the unattended-run-resilience rule).
4. **Context + memory tools granted** to the reflect-steward **and the
   book/video digesters** — a huge transcript or book can blow a single digest
   stage (Michael's catch): `compact_context` + the memory tools
   (remember/forget/mute/compress/pin/expand). **Excluded: the self-editable
   base-prompt tool** (CT2 §7.3 stays gated — self-rewriting its own system
   prompt is too much standing autonomy).
5. **The PR gate is by repo ownership.** LLM spend is already real and the coder
   is sandboxed, so existing caps hold. Opening a PR is the stop line: **our own
   repos → the coder may open DRAFT PRs autonomously, human reviews/approves the
   merge** (the merge stays the Hinge); **external repos → human verifies before
   any PR is created.** Safe housekeeping (shelf refill < 2) runs directly.
6. **Kill switch from day one** (global pause). **Plus: after "go," Claude Code
   self-schedules a wake every 2–5h** to check the steward's health (spend,
   proposal sanity, drift, runaway) and hit the pause if needed — the human's
   proxy watch between Michael's check-ins. The presiding chain *with eyes*:
   Michael → Claude-watchman → reflect-steward → doers.

## Governance into the substrate (Michael, 2026-06-15)

The autonomous agents should operate under the same principles we do. Port the
governance skills — **Ammon** (finish what you're handed), the **Dave rule**
(intent-clear + reversible → act), the **Ben test**, **stuffy-in-the-loop** (the
autonomy-bin rubric) — into the OSS substrate as **agent-available guidance.**
Mechanism TBD: a covenant extension (rides every dispatch, like the presiding
extension PR.1) and/or a substrate skill-doc surface the agent can load. This is
also the channel for promoting OSS-stack-proven features back into the repo as
the soak surfaces them.

## Recommendation

Ratify the six open questions above and I build **P0 (engine + Vivint) on the
OSS stack**. It's the substrate's reason for existing made concrete — the first
real rung of the AI-office vision and the presiding chain — and it pays for
itself twice by serving as the M7 soak.
