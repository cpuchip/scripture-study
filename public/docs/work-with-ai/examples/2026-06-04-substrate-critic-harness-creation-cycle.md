# Case study: the substrate that learned to review itself

**Date:** 2026-06-04
**For:** the Working-with-AI book — fits Part 5 (The Complete Cycle), specifically **Step 7 (Review)** and **Step 8 (Atonement)**, with a coda for **Step 11 (Zion) / the Bishop-vs-Conductor** section.
**One-line:** a coding agent shipped correct-but-incomplete work; the fix was not a better model but the two creation-cycle steps the industry skips — review-against-intent and redemptive error-recovery — built directly into the pipeline.

> **Honesty note for the author (Ben Test):** I did **not** consciously walk the 11 steps as a checklist while building this. I applied the *principles* (covenant, line-upon-line, foresight, read-before-trusting) the way they've become habit. The step-by-step mapping below is **mostly seen in hindsight** — and that hindsight is itself the finding: the substrate maps to the cycle because it was *built* on it, so a gap in the substrate turned out to be a *missing step of the cycle*. Don't write this up as "we deliberately engineered each step." Write it as "the pattern was already there, and the failure showed us which rung was missing."

---

## The arc (what actually happened)

1. **The setup.** The substrate (pg-ai-stewards) had just gained the ability to work in a real repo and open PRs (coder-v2). Overnight, unattended, it built the backend core of a chat app (ai-chattermax) as six chained, reviewed, tested PRs. Every PR built green and passed tests. It looked like a clean win.

2. **The disappointment.** On review, the human (Michael) found the plan wasn't fully carried out, and the code that *was* built had quiet gaps: a presence tracker built **global** when the design implied **per-room**; a "room isolation" test that tested no isolation; a "concurrency" test with no goroutines. All of it compiled. All of it passed. None of it was caught.

3. **The diagnosis (the important part).** The gaps were **not** the model failing. They were two other things:
   - **The harness only reviewed for correctness.** The `verify` stage ran build + test — "does it work?" — and never asked "does it match the plan?" or "does it serve the intent?" So a *correct implementation of the wrong thing* sailed through.
   - **The agent was context-starved.** The reviewing human can see every repo in the workspace; the sandboxed coder sees only the one cloned repo. When the spec said "do what we did for the other site's auth," the coder literally could not look. The human fills that gap reflexively; the sandbox can't.

4. **The fix.** Not a stronger model. A **critic stage**: a *different* strong model reads the real diff against an explicit acceptance-criteria checklist and either passes it or bounces it back to the implementer with specific, file-cited feedback — capped, then escalated to a human. Plus a discipline that the spec must carry the whole plan and the context the sandbox can't see.

5. **The proof.** Re-ran the exact failure on purpose: a vague spec to the implementer, strict criteria to the critic. The critic caught all three planted gaps with line citations, bounced it, the implementer fixed every one, and the second review passed on the merits. The substrate now catches the class of gap it had shipped silently the night before.

6. **The measurement.** With the harness fixed, ran four coding models head-to-head on the same task. **All four passed the critic on the first try; all four were race-clean and correctly scoped.** The harness, not the model, had been the gap.

---

## The mapping to the 11-step cycle

The reason this is book material: the fix lands precisely on the two steps the guide says the industry is *missing* (Steps 8-11 "entirely uncharted").

### Step 7 — Review ("watched until they obeyed")
The guide's own table names three layers of review:

| Layer | Question | Tool |
|---|---|---|
| Correctness | Does it work? | tests, CI |
| Specification | Does it match the spec? | spec diff, acceptance criteria |
| Intent | Does it serve the purpose? | intent audit |

The substrate already had **layer 1** (the `verify` stage: build + test). The night-build failure was a textbook **"correct implementation of the wrong thing"** — exactly the failure the guide says correctness-only review misses. **The critic stage we built is layers 2 and 3 made executable:** it checks the diff against the acceptance criteria (specification) and against the binding question (intent). We didn't invent this from scratch; we implemented a table that was already written in Part 5.

### Step 8 — Atonement (redemptive error recovery, "all things work together for good")
The guide's Atonement pattern: *don't just revert → analyze why → capture the learning → forward-recover with the learning incorporated → adjust the covenant.* The critic's **revise loop is this pattern in code**: the critic names what's wrong (the learning), the feedback is injected into the implementer's next attempt (captured + carried forward), and the implementer **forward-recovers** — it fixes from the feedback rather than starting over or rolling back. It is not retry-the-same-way (the industry's "Atonement = missing" cell); it is move-forward-changed ("go thy way and sin no more").

And at the meta level: the night-build failure itself ran the Atonement pattern. The failure wasn't reverted; it was analyzed ("the harness was the gap"), the learning was captured, and we forward-recovered by building the missing step. The covenant got refined — the pipeline now *demands* the whole plan. The failure became the reason the system is better. "All things work together for good" is not a sentiment here; it's the literal shape of what happened.

### The headline insight
**The harness gap was a missing step of the creation cycle.** The substrate's pipeline already instantiated most of the cycle (spiritual creation = `plan`, physical creation = `implement`, correctness-review = `verify`). What it lacked was Review-beyond-correctness (Step 7 layers 2-3) and Atonement-as-forward-recovery (Step 8). The "bug" was theological before it was technical: we had skipped the same two steps the whole industry skips. Building them fixed it.

### Step 11 — Zion / the Bishop-vs-Conductor coda (the open thread)
After the bake-off, Michael proposed staffing each pipeline phase with the model best-suited to it — "a ward council for development": a documenter phase, a builder phase, a critical-review phase, a verifier. That is **Step 3 (stewardship by ability — "distributed according to ability") meeting Step 11 (Zion — unified purpose across agents)**, and it is *exactly* the guide's Bishop-vs-Conductor argument: not one model conducting everything, but a council of stewardships, each with its gift, aligned by a shared intent (the binding question + acceptance criteria flow to every stage). The bake-off was the council *interview* — it measured each model's gift so the roles can be assigned on evidence rather than reputation.

---

## The findings (all of them, for the record)

**The harness mechanism (what was built):**
- A `review` critic stage added to the code-pr pipeline (clone → plan → implement → verify → **review** → pr).
- The critic is a *different* model from the implementer (default qwen3.7-max), held constant so it never grades its own work — "fresh eyes," the council member who didn't write the code.
- It inspects the **real diff** (not the implementer's self-report) against an explicit acceptance-criteria checklist carried in the work item.
- On a deficiency it loops back to `implement` with the feedback injected, capped at 2 cycles, then escalates to a human — the human-merge stays the final gate ("the Hinge").

**The bake-off (four models, same task, same critic):**

| model | wall-clock | tokens | impl LOC | read |
|---|---|---|---|---|
| kimi-k2.6 | 7.3 min | 531k | 85 | leanest, fastest, clean (snapshot-before-send). The default. |
| deepseek-v4-pro | 10 min | 783k | 116 | clean channel-actor, sound. Strong #2. |
| glm-5.1 | 13 min | 953k | 161 | most code/tests, but **a latent data race** (actor + mutex mixed); verbosity ≠ quality. |
| minimax-m3 | 22 min | 1.3M | 118 | best docs + defensive copy, but lock-held-during-send and 3-4× the cost. |

- **All four passed the critic first-try; all four build/vet/`go test -race` clean; all four correctly room-scoped.** With a rich spec + a critic, every model produced correct, sound-enough, tested code. *The model barely mattered; the harness was the lever.*
- **The meta-finding worth the most:** glm's data race passed build, vet, the race detector, *and* the critic — every automated gate — because the tests never exercised that path. **The full-context shepherd caught it** — the orchestrating Opus holding the whole arc, not a human and not any single gate. (The human *did* catch the separate plan/scope gaps on review; the *race* was the shepherd's catch.) So: narrow, diff-scoped gates **raise the floor** (no silent dropped requirements / vacuous tests); a **reviewer that holds the whole picture raises the ceiling.** The axis is **full-context vigilance vs. narrow gate — not human-vs-AI.** The human stays the **Hinge** — the merge authority who owns the consequence — which is about authority, not about having the sharpest eyes. The lesson for the book (Ch 4): keep a full-context watcher in the loop *and* the human as the Hinge; don't read this as "humans unnecessary."

**The context-starvation finding (ties to Step 5, Line upon Line):**
- The sandbox sees one repo; the orchestrator sees the workspace. "Do what we did elsewhere" is uninstructable to a context-starved agent. The fix is the *steward's* job: the spec must carry the cross-repo pattern, not point at code the agent can't open. This is line-upon-line inverted — the human must *grant* the context the agent has demonstrated it needs, because the agent cannot reach for it.

---

## What I'd want a reader to take away
1. The breakthrough was not a better model. It was building the two cycle-steps the industry skips (intent-review, redemptive recovery). The "uncharted territory" the guide names is where the real gain was.
2. A failure that compiles is the dangerous one. Correctness-only review certifies "correct implementation of the wrong thing." The cure is review against the spec and the intent — and it can be made a pipeline stage.
3. Automating review raises the floor; it does not remove the watcher. Keep the human at the merge.
4. Measure the gifts before assigning the council. The ward-council model for AI is real, and the bake-off is how you staff it honestly.
