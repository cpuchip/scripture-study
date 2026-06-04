# Journal — critic harness + 4-model bake-off

**Date:** 2026-06-04
**Workstream:** WS5 / pg-ai-stewards (coder harness) + ai-chattermax (test target)

## The turn the disappointment took

The night-build review showed the plan wasn't fully carried out — and honestly that was mostly *me* (conservative decomposition) + the *harness* (verify only checks compile/test, never plan-conformance; the sandbox is context-starved vs. my whole-workspace view), not kimi failing. Michael's call: **fix the harness so it asks for the whole plan + hands over context, then bake-off the models on the fixed harness.**

## Built (SQL-only, no rebuild)

- **cv6 — critic stage + revise loop.** code-pr: clone→plan→implement→verify→**review**→pr. The `review` critic (a different strong model, default qwen3.7-max) inspects the *real* diff against an explicit `acceptance_criteria` checklist and returns `REVIEW: passes`/`revise`. Loop-back is one branch in `work_item_advance` (gated to code-pr review): revise → implement with feedback injected (cap 2) → past cap, awaiting_review. The implement stage gained a REVISION-REQUESTED section fed by `input.review_feedback`.
- **cv7 — critic immune to model_override.** So a bake-off dev-model override sets every dev stage but the critic stays the constant fresh judge (no self-review). Verified live: m3's review ran on qwen3.7-max.

Mechanism notes (H1): the auto-advance trigger `handle_work_item_chat_completion` (3c2) calls `work_item_advance` (SQL) then dispatches the next stage — so the loop-back was a contained SQL edit, no bgworker/Rust rebuild. The existing l28 gate only *halts* a bad review; revise-proposal is a separate human-triggered flow; neither loops — so the loop was net-new.

## H4 — the critic proven on the night-build's exact gap

Deliberately vague spec → kimi built a global (non-room-scoped) presence Tracker with a "room isolation" test that tested no isolation and a "concurrency" test with no goroutines. The critic caught **all three** with file/line citations, bounced (rev 0→1), the implementer fixed them (`map[RoomID]map[ID]Participant` + a real isolation test), and the second review passed. The harness now catches the class of gap it shipped silently the night before — and catches vacuous tests, the harder win.

## H5/H6 — the bake-off

Same room-scoped `chatcore` module (room hub + presence + ratelimit, strict acceptance criteria), built by all four dev models on their own branches (PRs #12-15), qwen3.7-max critic constant.

**Every model passed the critic first-try (r0, zero bounces); all build/vet/`go test -race` clean; all room-scoped.** With rich specs + the critic, the harness lifted *every* model to correct, sound-enough, tested output. That's the headline: the harness, not the model, was the gap.

Differentiation (human read):
- **kimi-k2.6** — 7.3 min / 531k tok / 85 LOC. RWMutex, snapshot-clients-then-send (lock not held during I/O). Leanest + fastest + clean. The default. (Minor: no defensive message copy.)
- **deepseek-v4-pro** — 10 min / 783k. Clean channel-actor, auto-started, sound. Strong #2.
- **m3 (minimax-m3)** — 22 min / 1.3M tok. Best docs + defensive copy + careful contracts, BUT holds the lock through Send (a slow client stalls the room) and 3-4× the cost. The care/1M-context pick, not the default.
- **glm-5.1** — 13 min / 953k / 1532 LOC (most). Over-engineered (channel-actor **+** a mutex) with a **latent data race**: `Run()` writes `h.rooms` without the mutex while `Rooms()` reads under `mu.RLock()`. The tests don't exercise that path, so build + vet + `-race` + the critic all passed it. Verbosity ≠ quality.

**Meta-finding worth keeping:** glm's race cleared every automated gate; the **full-context shepherd (the orchestrating Opus holding the whole arc) caught it** — not a human, not any single gate. So narrow gates **raise the floor**; a **full-context reviewer raises the ceiling** — the axis is full-context-vigilance vs. narrow-gate, *not* human-vs-AI. The human stays the **Hinge** (merge authority / owns the consequence), not the sole defect-catcher. The critic is a backstop, not a guarantee. (Michael's correction 2026-06-04 — see memory `feedback_full_context_shepherd_is_the_ceiling`.)

## Carry-forward

- night-build still awaits Michael's review → main (his Hinge/deploy).
- A possible 2nd-pass concurrency-soundness critic, or use glm (thorough test-writer) as a test author / second critic.
- The deferred 12-item items: substrate persona schema (own ratification), Vue frontend, moderation, D&D wiring, auth (needs the becoming/1828 cookie pattern *injected into the spec* — the sandbox can't see those repos).
- ⚠ Parallel terminal still active + shadowing my slugs (foreign sandboxes/worktrees) — left untouched.
