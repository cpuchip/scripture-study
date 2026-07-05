# War-gaming: turning a smart model's simulation into a cheap model's executable

**Source:** Mark Kashef, "Do THIS Before You Lose Access to Fable 5" ([nuwlyQXrADg](https://www.youtube.com/watch?v=nuwlyQXrADg), 2026-07-05, 13:58)
**Read for:** does it change how we think about pg-ai-stewards? Yes — it names the 20% we haven't built.

## The claim, stripped of the "spend your last tokens" wrapper

Anthropic's own guidance is: don't waste your smartest model on *plans* — a plan is a diagram of the surgery, not the surgery. But the video's move is sharper than "use it to execute": use it to **war-game**. A plan assumes a blue-sky line — phase 1, phase 2, high success. A war-game assumes reality will humble every move.

The structure it prescribes, move by move:
- every move states its **expected observation** — exactly what you'd see if it worked, and what you'd see if it didn't;
- every move carries its **most-likely failure**, that failure's **signal**, and the **countermove**;
- forks get **triggers** — *if you observe X, take route A, else B*;
- **unresolved assumptions** (what recon couldn't settle) are flagged, not guessed;
- it ends with **abort conditions** — the errors at which the plan stops rather than thrashing;
- and it traces **2nd/3rd/4th-order consequences** — the failures a few layers down from the first move.

He names the loop out loud: **action → reaction (reality throws an error) → counteraction.** "That's the modern agentic loop." The output is a durable war-game artifact per mission; a *cheaper* model then executes it confidently, because the realities are pre-simulated. He wraps it in a folder — `tasks/`, `wargames/`, `success.md` (what counts as done), `ledger.md` (blockers, with `(variable)` placeholders for what needs the human) — and runs `/goal` over the missions, fanning out parallel agents that recon then write.

## Why this matters to us: it's a manual, static version of what we already are

The uncomfortable-in-a-good-way part: his folder-of-markdown is a hand-rolled version of pg-ai-stewards' live control flow.

| His war-game rig | pg-ai-stewards |
|---|---|
| `tasks/` fanned to parallel agents | `decompose-fanout` pipeline |
| move → "failure signal → countermove" | steward `diagnose_failure → retry / failover` |
| fork triggers (`if X → route`) | `route_on` edges |
| abort conditions | spend caps, breaker, quarantine, step budgets |
| `ledger.md` `(unresolved-variable)` | `needs_attention` view + `ask_up` ladder |
| `success.md` | maturity gates + the critique checklist |
| smart-writes / cheap-executes | loom-sonnet judges, cheap models execute |

A mass-market creator arrived independently at our founding thesis: the smart model's value is planning and simulation, cheap models execute, and failure-handling should be **explicit and durable**, not improvised per run. That's validation. It also hands us a one-line frame: **war-gaming as infrastructure** — his markdown is the artisanal version; the substrate is the industrialized one, where the war-game's forks and aborts become live database edges the executor actually obeys, with budget and governance enforced structurally rather than hoped for.

## The 20% it names that we don't have

We handle failure **reactively** — we react *when* a stage breaks (retry the blip, walk to the next alias member, escalate to the human). We do not **pre-simulate**. Our failover is generic: any transient → retry. The video's real insight is that a strong model, *asked to war-game*, surfaces the **unknown-unknowns** — the task-specific failure modes you didn't know to guard against — *before* execution. We have no task-specific pre-mortem.

Two builds follow (design: `.spec/proposals/war-game-pipeline.md` in the OSS repo):

1. **A `war-game` pipeline** — decompose-fanout's cousin. A strong loom stage takes a mission brief and emits the war-game artifact (moves, observations, failure signals, countermoves, fork triggers, abort conditions, unresolved-assumption ledger, higher-order consequences); pool it as a doc. Cheap, and immediately demoable.
2. **War-game-informed execution** — the categorical differentiator. When a work item carries a war-game, inject it into the executor's context **and materialize its abort conditions as real gate checks, its fork triggers as `route_on` edges.** *This* is what separates us from "feed a model a markdown file": the guardrails are enforced, not suggested.

The honest synthesis is belt-and-suspenders, not replacement. A pre-simulation can be wrong — it war-games failures that never come and misses ones that do. So the war-game is a *prior*, and the reactive failover we hardened this week stays the backstop. Prospective plus reactive is stronger than either.

## Honest caveats

The video's war-games are **static** — written once, fed to a model, no feedback. That's precisely the artisanal ceiling our live control flow transcends, so we take the idea, not the mechanism. Its "tailor the war-game to a specific executor model's system card" is real but less useful to us — we abstract over models via aliases, so the routing layer owns model choice and the war-game should stay model-agnostic. And the whole thing is a gimmick wrapper (use your expiring Fable tokens) around a genuinely good pattern; the pattern is what survives the wrapper.

## The timely angle

The best use of a smart model before the Fable window closes is his own pitch, turned on our own roadmap: **war-game the substrate's hardest open items** — D2A (pack-as-extension), D3C (policy layer), multi-tenancy — producing war-game artifacts that cheaper models, or a later session, execute. Durable intelligence that outlives the window. It folds straight into the Lab's Fable-hinge A/B, and it makes a real demo beat: launch a war-game work item live, watch the move/countermove/abort artifact appear, then watch a cheap executor obey its forks.
