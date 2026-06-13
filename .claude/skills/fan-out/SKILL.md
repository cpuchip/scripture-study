---
name: fan-out
description: Parallelize a large task across many subagents — when the work is independent per-unit (verify / research / generate, the same operation across N files or items) it is a fan-out shape, and parallel fresh-eyes-per-unit beats one tiring serial operator. Use to triage whether a long task should be fanned out, and to run the fan-out safely under the presiding watch. Born from the 2026-06-13 scratch-audit (62 files, 6 Opus agents) that caught what a serial walk had missed.
---

# Fan-Out

## The triage (do this BEFORE starting any task with >~20 independent units)

Ask one question: **what is the SHAPE of this work?**

- **Fan-out shape** — independent, per-unit, the same operation across many units
  (verify N files against sources, research N videos, generate N stubs). No unit
  needs another unit's result. → **Fan out.** Parallel speedup *and* a higher
  quality ceiling: fresh eyes per unit beat one operator who accumulates blind
  spots ("I already know the pattern, I'll stop re-checking").
- **Single-pass shape** — centralizable (one known pattern, deterministic to
  locate — e.g. "remove all the X notes"), or sequentially dependent (each step
  needs the last). → **One careful pass.** Fan-out here just adds coordination,
  consistency risk, and review cost without improving the output.

The classic miss: a 469-file verification *walk* run serially for days because
"just do the next file" momentum never triggered the triage. It was a textbook
fan-out shape. **Surface the triage result before starting** — a quick "this is
~N independent units; I'd fan it out / I'd do it single-pass because ___."

## Why fan-out has a higher ceiling on verification (the refined principle)

"The full-context shepherd is the ceiling" is only half true:
- **Shepherd-for-integration** — one full-context mind is the ceiling for
  *cross-cutting* issues (a race spanning files, a thesis contradicting another
  study). Fan-out can't see those.
- **Fan-out-for-independent-verification** — for *per-unit* checks, parallel
  fresh eyes BEAT one accumulating operator. They are complementary, not rivals.

**Serial-probe, then parallel-scale.** A short serial pass finds the pattern
(e.g. "the dominant defect is Webster 1913-as-1828; also check the citations
*inside* entries"). Then front-load that knowledge into the fan-out spec so the
parallel agents inherit the shepherd's accumulated context without the blind
spot.

## The recipe

1. **Enumerate + batch.** List the units; split into N batches (aim ~10-14
   units/agent; balance by weight, not just count).
2. **Write ONE shared spec.** Self-contained — the agents start cold:
   - the *why* (one paragraph of context),
   - the exact method + tools (concrete commands / MCP tool names),
   - the rules (read-before-quoting; flag-don't-resolve for sensitive/bin-4
     items; conservative — correct only on a clear source mismatch, else FLAG),
   - the front-loaded pattern from the serial probe,
   - the **output format** (a compact structured per-unit report: corrections
     `<ref>: "stale" → "genuine" (verified via X)`, flags, clean items, tally).
   Give each agent the shared spec + its batch slice. Tell them: **edit + report,
   no git/commits** (the presider commits after review).
3. **Stage it — wave 1 to validate, then scale.** Spawn a SMALL first wave
   (~2 agents) on varied batches. Review their work before spawning the rest. If
   the spec produced clean, correct output, spawn the remaining waves; if not,
   fix the spec cheaply (git reverts the small wave). This is the D&C 121
   "reprove, then show increase of trust" applied to your own delegation.
4. **Spawn in parallel.** Multiple `Agent` calls in one message run concurrently
   (or `run_in_background: true`). Model: `opus` for careful verification.

## The presiding watch (non-negotiable)

A fan-out is a delegation. The presiding covenant governs it (`.spec/covenant.yaml`
→ `presiding.watch_what_you_order`, Abraham 4:18 — watch until it obeys the
INTENT, not until the job returns):

- **Review every agent's report.**
- **Spot-verify the high-stakes catches against the source yourself** — don't
  accept a confabulation-fix or a citation-correction on the agent's word; re-grep
  / re-MCP it. (In the scratch audit, 6 spot-checks across all agents — image→Matt
  22:20, D&C 8:11, Mosiah 18:29, D&C 107:87, transgression, ordain — confirmed
  reliability before accepting ~75 edits.)
- **Check the diff is bounded** (proportionate to the reported corrections; no
  wholesale rewrites).
- **Consolidate**: merge reports into the audit record (e.g. `findings.md`); fix
  any cross-cutting issues the units surfaced (a fan-out can reveal an error that
  reached a *parent* artifact — chase those down).

**Fan-out without the watch is offloading; fan-out with the watch is
multiplication.** The watch is the real bottleneck, which is why the output
format must be reviewable.

## Cost

Spawning is the expensive path — each agent re-derives context cold. Justified
when (parallel speedup + fresh-eyes-per-unit value) > (coordination + review +
cold-start cost). True for independent verification across many units; false for
centralizable single-pattern cleanup.

## Worked example (2026-06-13 scratch audit)

62 scratch/provenance files, ~9-18 min parallel. 6 Opus subagents, staged
2-then-4. Shared spec front-loaded the serial walk's pattern (Webster
1913-as-1828 + "check citations inside entries"). ~75 corrections; I reviewed all
reports + spot-verified 6 high-stakes catches + the diff before committing. It
caught a class the serial walk under-checked (fabricated citations inside Webster
entries) and one error that had reached a study file (alma5 `image`→Matt 22:20)
the serial pass marked CLEAN — the strongest evidence for the higher ceiling.

Principle + this example live in the memory note
[[reference_fanout_vs_shepherd]].
