---
batch: SW
title: Scheduled workflows — cron-style periodic research + autonomous discovery
status: idea (stub — needs council)
proposed_by: michael
proposed_on: 2026-05-17
preceded_by:
  - substrate-pipelines-expansion.md
links:
  - "../../extension/src/bgworker.rs"
  - "proposals/substrate-pipelines-expansion.md"
---

# Scheduled Workflows — cron-style periodic work

> **Status: idea stub.** Captured 2026-05-17 from Michael's framing. Not yet
> councilled — the decisions below (D-SW1–D-SW7) need to be walked before any
> build. This file exists so the idea has a home in the proposal queue.

## Why this exists

The substrate runs work when a work_item is created — by Michael, by an agent
write-back, or by the watchman's dirty-queue pressure. There is no surface for
**"do this kind of work on a recurring schedule."** Michael wants three
concrete recurring jobs, and a general mechanism underneath them:

1. **Periodic exhibit research from physics news.** On a schedule, scan
   physics/science news, identify candidate exhibits (cf. the
   `everyday-science-to-exhibits` work item that drove Batch J), and kick off
   `research-write` — or fan-out — pipelines for the strong candidates.

2. **Autonomous YouTube AI-video review.** On a schedule, the substrate finds
   recent AI YouTube videos on its own (search + dedup against what it has
   already evaluated) and runs them through the `yt` evaluation pipeline.

3. **Public-playlist ingestion.** Michael maintains a public YouTube playlist
   as a curated inbox. The substrate polls it, detects entries new since the
   last poll, and runs each through the `yt` pipeline.

(1) and (2) are **autonomous discovery** — the substrate decides what to work
on. (3) is a **curated inbox** — Michael decides, the substrate picks up.

## Relationship to `substrate-pipelines-expansion.md`

That proposal already names "scheduled-pipeline machinery" inside its scope
(D-PE family) and defines the `research` and `yt` pipelines these jobs would
invoke. This proposal is the **scheduling + discovery layer** that sits on top:
the schedule definitions, the discovery sources, the new-since-last-poll
tracking, and the cost gate. The two should be councilled together, or
pipelines-expansion first (it provides the pipelines this depends on). A
council decision: does the scheduled-pipeline machinery live here or there, and
does this proposal absorb it. → **D-SW7.**

## The autonomous-spend concern (read this first)

The ES emergency-stop arc — closed two days before this idea — was entirely
about a runaway that spent ~230M tokens unattended. A scheduler that
**auto-spawns pipelines on a timer** is, by construction, autonomous spend. The
ES guardrails (work_item cost cap, kind circuit breaker, cancel cascade) all
still apply, but a schedule adds a *new* failure mode: a misconfigured cadence
or a discovery source that returns hundreds of candidates could enqueue far
more work than intended before anyone looks. The human-in-loop gate (D-SW5) and
per-schedule budget (D-SW6) are not optional polish — they are the load-bearing
decisions of this proposal.

## Decisions to walk (council)

- **D-SW1 — Schedule definition surface.** Where do schedules live? A new
  `stewards.scheduled_jobs` table (cron expression + pipeline + source config +
  enabled flag)? An extension of `watchman_config`? Authored via stewards-ui,
  via SQL, or via a YAML file like intent/covenant?
- **D-SW2 — Trigger mechanism.** Reuse the watchman's existing 60s scheduler
  tick to also evaluate due scheduled_jobs, or stand up a separate scheduler
  loop? (Reuse is cheaper and already soak-tested.)
- **D-SW3 — Discovery sources.** How does "physics news" get fetched — an RSS
  feed list, a recurring `web_search_exa` query, a curated source list? How
  does autonomous AI-video discovery query (yt_search terms + recency window +
  dedup key)?
- **D-SW4 — Public-playlist watcher.** How is the playlist registered, how is
  "new since last poll" tracked (last-seen video id / publish timestamp), and
  how are already-ingested videos deduped?
- **D-SW5 — Human-in-loop gate.** Does a scheduled discovery **auto-spawn** a
  full pipeline (spends immediately), or land candidates in a review queue for
  Michael to approve before any spend? Likely answer: discovery is autonomous,
  *spend* is gated — but this is the core ratification.
- **D-SW6 — Cost guardrails.** Per-schedule budget cap; behavior when a tick
  would exceed it (skip, defer, surface). How scheduled spend rolls into the
  existing 4-bucket `cost_buckets` tracking.
- **D-SW7 — Ownership of the scheduled-pipeline machinery.** Build here or in
  pipelines-expansion; does this proposal absorb that part of D-PE.

## Not in scope (for the stub)

- The `research` and `yt` pipeline shapes themselves — those are
  pipelines-expansion.
- Anything resembling self-modification of the schedules by an agent — a
  scheduled job changing its own cadence is a separate, later question.

## Next step

Council D-SW1–D-SW7 (paired with or after `substrate-pipelines-expansion.md`),
then this stub becomes a phased build proposal in the C–F cadence.
