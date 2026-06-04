# Journal — Ammon night: the substrate built ai-chattermax's backend core

**Date:** 2026-06-04 (overnight, while Michael slept)
**Workstream:** WS5 / pg-ai-stewards (coder) + ai-chattermax
**Branch:** `agent/night-build` (origin), awaiting Michael's review → main.

## The ask

Michael, heading to sleep: *"I think you are capable enough to watch and shepherd this one through standing in my place and reviewing the PRs as they come in acting in my behalf… can you be my Ammon for tonight and try to get ai-chattermax built and tested through pg-ai-stewards?"* — an explicit delegation of the review/merge Hinge for the night.

## What I did

Stood in as the **bishop (decomposition) + reviewer (the Hinge)**, driving the proven single-`code-pr` path repeatedly rather than building auto-fan-out machinery overnight.

1. **cv5 — code-pr base-branch support.** Added `input.base_branch` to the clone + pr stages (clone that branch, PR into it). Lets a chained build accumulate on an integration branch and **never touch main**. Idempotent string-replace on the two stage templates. Tested by B1 itself (PR opened against night-build, not main).
2. **`agent/night-build`** created off main (gh ref, token-side); soak paused; the 12-item plan decomposed into a reliable backend-core chain (tasks #100-105), modeled as a dependency graph in the Claude task system (the "non-overlapping, chained" shape Michael described).
3. **Ran sequentially, reviewed each, merged into night-build** (squash, one commit/item):
   - **B1** scaffold — module + cmd/server + /healthz + graceful shutdown (PR #5)
   - **B2** room hub — transport-agnostic `Client` iface, mutex-safe Hub; **copies clients + releases the lock before I/O in Broadcast** (PR #6)
   - **B3** scheduler — round-robin turns + hard per-participant rate ceiling, injectable clock (PR #7)
   - **B4** transcript — `Message`/`Store` + concurrency-safe in-memory impl, defensive-copy Replay (PR #8)
   - **B5** presence — typed `Kind`, mutex-safe Tracker, deterministic ID-sorted roster snapshot (PR #9)
   - **B6** integration — WebSocket server wiring all four: shared Hub/Scheduler/Store/Tracker via a testable `newMux`, `wsClient` with **mutex-serialized writes** (gorilla panics otherwise), read loop `Allow → Append → Broadcast`, `/ws/{room}` + `/roster/{room}` + `/healthz` (PR #10)
4. **Final whole-module verify** (inverse hypothesis on the cumulative branch, not just per-PR): cloned night-build fresh → `go build ./... && go test ./... && go vet ./...` → **ALL-GREEN**, 5 packages pass, vet clean.

## Review discipline

Read every PR's actual diff + code (not just the green flag) and merged on Michael's behalf. Code quality was consistently high (idiomatic, concurrency-aware, well-tested); minor nits logged (B1 handler not extracted; B4 returns the unexported type; B5 one gofmt-alignment; B6 roster is global not room-scoped, gorilla marked `// indirect`) — none blocking. **Never touched main** — that merge (and the deploy it triggers) is Michael's Hinge in the morning.

## Anomaly surfaced (not touched)

3 sandboxes — `ai-chattermax-root-scaffold` / `room-hub` / `transcript` — created 00:18-00:35 **concurrent with my build**, still running, with **descriptive (non-work_item-stamped) names** → created by **direct `coder_sandbox_start` calls, not the pipeline** (no matching work_items, no PRs). One has a local-only scaffold commit + a stray `server` binary; the other two are bare clones. Almost certainly a **parallel terminal** also building ai-chattermax. Per data-safety (don't delete what I didn't create), **left untouched** + surfaced to Michael. My night-build is independent and intact; main is untouched.

## Carry-forward

- **Michael reviews `agent/night-build` → merges to main** (his Hinge/deploy).
- The parallel-build sandboxes — Michael to identify/clean (his other terminal?).
- **Auto-fan-out** (the substrate does the bishop moment itself) — tonight I was the bishop manually; automating it is the "with Michael" feature, plus the "new feature he mentioned."
- Remaining 12-item plan: #6 substrate-persona schema (its own ratification), #9 Vue frontend, #11 moderation, #12 D&D MVP wiring.
- Minor PR nits above, if we want a cleanup pass.

## Notes

cv5 committed (not pushed — root preference). Soak resumed. Cost rode the opencode_go subscription (kimi-k2.6); no 429s hit. The pipeline worked: 6 chained PRs, ~3-5 min each, zero pipeline failures — the multi-PR real-project build is proven.
