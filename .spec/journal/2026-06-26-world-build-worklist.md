---
date: 2026-06-26
lane: pg-ai-stewards
topic: world-build worklist harness (the scratch-file fix) + the rig/CI fixes that preceded it
tags: [loreworks, world-build, harness, llama-chip, ci, harness-over-intelligence]
---

# The world-build agent gets a scratch file

## The arc of the session

Three things, in order, each feeding the next.

**1. CI had been red for 7 commits and I'd been calling it green.** Came back to confirm
`#282` and found the `extension build + virgin smoke` job failing since `b977546`. Root
cause: `60-chat-model-pin.sql` was registered in `lib.rs` but never added to the
Dockerfile's explicit COPY list, so `cargo pgrx package` couldn't read it. I'd been
reading the `go build` job (always green) and assuming. One-line fix (`806a8d8`) healed
all 7 — but the build break had been *masking* every downstream virgin-smoke assert, so
when the image finally built, assert 51 tripped (a #282 gate test still attributed its
artifact the old session-scoped way). Fixed the test (`d4cc045`), CI green.
**Lesson, now permanent: a broken extension build hides a broken test — the go-build job
is never "green," and a chain change's oracle is virgin-smoke on a fresh image.**

**2. "A GPU fell off" mid Cosmere-import.** Both 4090s sat at ~140 MiB free. Michael's
instinct was to trim context. I trimmed it (live + config) and measured — and context
turned out to be a **non-lever**: qwen used the same ~2 GB of non-weight VRAM at 245k,
196k, *and* 98k context. The cards are full from a 20 GB model + fixed buffers + the
OS/LM-Studio baseline, none of which the window touches. The real cause was LM Studio
**auto-spawning nomic embedder replicas** onto already-full cards during the import's
embed burst (each replica grabs a ~400-500 MB CUDA context). Michael fixed it the right
way — moved nomic to CPU — and both cards jumped to >1 GB free. **The fix wasn't the dial
he reached for; it was the co-tenant he couldn't see. Measure by unloading; Windows WDDM
hides per-process VRAM.**

**3. The world-build agent has no scratch file.** Re-running the Cosmere build after the
GPU fix, the local model did 60 turns / 235 `doc_search` calls and extracted **zero**
entities — it free-searched an unbounded semantic space with no done-signal and looped
to the step cap. A hand-typed steering instruction ("search a few times, then extract;
count:0 means move on") rescued it (95 entities). Then Michael said the line that turned
a workaround into a build: **"Gemini died the same way on MLP — lots of searching, no
done-marker."** A *strong* cloud model failing identically is the tell. This isn't a weak
local model; it's a **harness gap**. Harness > intelligence.

## What got built

`61-world-build-worklist.sql` — the study scratch-file rule applied to extraction:

- **`world_build_coverage`** — one row per source chunk per world. Persists across runs,
  so a huge corpus finishes across multiple builds, each guaranteed to advance (the
  BoM-walk committed-progress pattern). Resumable by construction.
- **`world_build_walk`** — the driver. Seeds the worklist from the world's project
  chunks, serves the next bounded batch (marking them shown), and returns `complete:true`
  once nothing is pending. The build becomes a **bounded, deterministic WALK** of the
  canon instead of "search until you feel done." The agent *cannot* loop forever (pending
  strictly decreases) and has an unambiguous finish line.
- The world-build **agent prompt** re-authored to walk-then-extract with a COMMIT clause
  (the permanent form of the steering, mirroring the `45-work-item-chat` COMMIT fix that
  killed the chat reasoning-spiral). `world.go`'s per-build canon hint aligned to the walk.

Coverage is *coverage, not correctness* — the `58` world-critic still judges extraction
quality. This file owns "did we look at every chunk," which is the thing that was missing.

## How it was tested (three ways, because the oracle comes first)

1. **Deterministic unit test** — seed → drain 27 → `complete:true`; reset is pure (seeds,
   serves nothing); stable after complete. This caught a wart (reset was serving a chunk)
   before any model ran. Build the oracle first.
2. **Cosmere regression** — the corpus that looped to 0 → walk built **73 entities / 126
   edges with no hand-steering**. The harness replaced the manual instruction.
3. **Star Trek scale** (114 chunks) — the *old* doc_search-first pattern overflowed
   instantly (99,925 tokens > the 98,304 per-slot context, from one `doc_search` of 50
   results + a big `doc_get`). **Walk-first walked all 114/114 → 234 entities, zero
   overflow** — bounded batches plus the substrate's existing `page_in_cap` (which trims
   old walk-chunk tool results) keep cumulative context under the ceiling.

CI green on a fresh image (`615dc29`, virgin-smoke `OK 50`).

## Findings worth keeping

- **A done-signal can be gamed if reaching it is easier than the work.** First walk-driven
  run, the model speed-ran coverage to `complete:true` and extracted little — then on its
  own pivoted ("I've now read all 27 chunks, let me systematically extract") and produced
  the full graph. Walk-all-then-extract emerged as the model's chosen strategy and it's
  fine; but it's a reminder that the coverage signal answers "did we look," not "did we
  extract well" (that's the critic's job, kept separate on purpose).
- **The PM(4) context trim is the exact ceiling the old pattern hit.** 196608/parallel-2
  = 98304 per slot, and the overflow was 99,925. Two readings: (a) it's a concrete reason
  the *walk* (bounded turns) matters, and (b) a *very* large corpus might still want
  `compact_context` wired into the extraction loop — a deferred follow-up, not blocking,
  since the walk + existing caps handled 114 chunks fine.

## Carry-forward

- The worklist+coverage pattern generalizes beyond world-build — the same shape would help
  the digesters and doc-build tackle large sources methodically. Worth lifting into a
  shared "extraction scratchpad" if the pattern proves out.
- `compact_context`-in-the-loop for >150-chunk corpora (deferred; the walk already makes
  the failure *resumable* rather than fatal, which buys the time to do it right).
- `st-walktest` (234 entities) kept as a live scale demo in the World tab — disposable.
