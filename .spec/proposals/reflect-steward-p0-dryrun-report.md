# Reflect-Steward P0 — Dry-Run Report

**For:** Michael's morning review · **Run:** 2026-06-15 (overnight) · **Stack:** OSS scratch (`stewards-oss-*`)
**Status of the autonomous loops:** NOT flipped on. This was one **manual** cycle. The scheduled steward + the Claude-watchman wake await your explicit "go live."

---

## TL;DR

I pointed the **real** planning engine at a new **Vivint intent** and ran one
reflect cycle by hand. The headline isn't the proposals — it's that **the
engine we need is ~90% already built**: the substrate's `planning` pipeline
*is* the reflect-steward loop (gather → explore → synthesize → propose → review),
and `enqueue_proposed_work_items` already lands proposals as parked,
human-reviewable work-items. The dry-run's main value was **surfacing real bugs**
(which is exactly what you said it was for), and they're caught + handled below.

## What I did

1. Confirmed the engine exists: the `planning` pipeline has the exact loop shape
   (`context_gather → explore → synthesize → propose_work → review_plan`), all
   stages auto-advance, and `enqueue_proposed_work_items` reads the `propose_work`
   output and inserts each proposal as an `origin='agent_planning'` work-item —
   parked (`__proposal_only`), visible, non-running. **That is the "PROPOSED"
   state from the spec, already implemented.** So the reflect-steward is mostly
   *configuration + scheduling + the review verbs*, not a new pipeline.
2. Created the **Vivint intent** (`slug=vivint`) with the public-only scope baked
   into its values hierarchy: *public-info-only · faithful-to-the-source ·
   actionable-improvement · strongest-signal-first*, anchor = "understand the
   real complaint before proposing the fix."
3. Ran one manual cycle (a `planning` work-item against the Vivint intent) with
   the binding question: research what people publicly say about Vivint
   (reviews, products, customer service, billing, pain points) and propose 5
   concrete improvements.

## Bugs the dry-run surfaced (the real payoff)

**1. The `research` agent was referenced everywhere but never seeded — PROPERLY FIXED.**
The `planning`, `research-write`, `research-summary` pipelines (`13`) AND the
`echo-test` smoke pipeline (`04`) all run on `agent_family='research'`, but
**core seeded no such agent**. Worse, the two example digesters had drifted onto
a *second* name, `stewards-explore`, that nothing shippable seeds either — this
stack only worked because dev/testing left a `stewards-explore` row behind,
masking the bug. A truly fresh `CREATE EXTENSION` would fail at dispatch with
`no agent variant resolved: family=research`.

The fix (per your "properly fix so fresh installs get the proper landscape") —
**unify on one core-seeded generic agent, `research`:**
- `13` now **seeds the `research` agent** (`model_match='*'`, a grounded
  research prompt) and grants it `web_search_exa` + `fetch_url` so it can
  actually research the web out of the box (exa-search is a core default server).
- Both example digesters + `echo-test` repointed `stewards-explore → research`;
  zero shippable `stewards-explore` remains.
- A **virgin-smoke assertion** (OK 3b) now guards it: the `research` agent must
  be seeded + web-capable, so this can never silently regress.
- **VERIFIED:** rebuilt the extension image, ran virgin-smoke on a fresh
  container — **all 6 OK blocks pass, incl. OK 3b**, and the fresh install shows
  `research | * | active`. (It failed before the fix, passes now — inverse
  hypothesis satisfied.) Pushed to the OSS repo as `b116ab9`.

**2. The `yt` MCP server was broken in the running bridge — FIXED.** The bridge
was running *without* `WITH_YT=1`, so `/usr/local/bin/yt-mcp` was missing and
`mcp_bridge_state` showed `yt` erroring `fork/exec ... no such file or directory`
— the playlist loop couldn't fetch new transcripts. **Rebuilt the bridge with the
yt overlay** (`docker compose -f docker-compose.yaml -f docker-compose.override.yaml
-f docker-compose.yt.yaml up -d --build bridge` — all three files so the coder
overlay, which lives in the auto-loaded `override`, isn't dropped). After
`refresh-tools`: **7/7 servers OK, `yt` now serving 5 tools** (yt_playlist,
yt_download, yt_get, yt_list, yt_search), `last_error` clear. The playlist loop
can fetch again. (Note: the stack must be brought up with the yt overlay to keep
it — yt is opt-in by design.)

**3. (Minor, already fixed earlier today) the `(unknown)` playlist digest** —
a failed listing published a placeholder; fixed at the write boundary
(`playlist_publish` now rejects non-YouTube ids), pushed to the OSS repo.

## Dry-run cycle output — it works, and the output is good

The cycle ran **end to end** (context_gather → explore → synthesize →
propose_work → review_plan → completed) on `kimi-k2.6`/`qwen3.6-plus`
(opencode_go), ~124K tokens, well under the $1 cap.

**What it produced (the real test — are these worth acting on? yes):** given a
*cold start* (empty Vivint pool), it correctly proposed a **research plan** — it
recognized it must gather before it can recommend, and decomposed the work into
five grounded next-steps:

1. `vivint-sentiment-bbb-complaints` — BBB billing/cancellation/contract patterns
2. `vivint-sentiment-trustpilot-reviews` — installation + service quality
3. `vivint-sentiment-reddit-threads` — unfiltered long-term owner narratives
4. `vivint-sentiment-app-store-reviews` — app-specific failures (connectivity, notifications, UI)
5. `vivint-sentiment-synthesis-proposals` — fuse the four into the final ranked improvement proposals

Each is a valid, parked **PROPOSED** work-item (`origin=agent_planning`,
`status=pending`, non-running) — exactly the review queue from the spec. The
**explore** stage gathered real findings *and surfaced its own assumptions +
asked clarifying questions back* ("competitor benchmarks in scope? policy
changes or strictly engineering?"). The **review_plan** stage returned a
structured `verdict: pass` (JSON valid, assumptions surfaced, risks concrete,
work-items appropriately sized). This is a careful analyst, not a slop generator.

**Three engine findings the run surfaced (beyond the agent/yt bugs):**

- **A — `input.today` hard-fail.** The resolver *raises* on a missing template
  field (`resolve_template_path: path input.today resolved to NULL`) instead of
  defaulting it, so it parked once mid-run. The reflect-steward's launcher must
  inject `today` (as `enqueue_proposed_work_items` already does), or the resolver
  should default it. Easy, but it *will* bite the autonomous launcher.
- **B — proposals did not auto-enqueue (the one that matters for autonomy).** The
  run completed at `maturity=planned`, not `verified`, so the
  `on_maturity_verified` planning branch that calls `enqueue_proposed_work_items`
  never fired — I had to call it by hand. **For the loop to run unattended, the
  planning pipeline must reach `verified` and auto-enqueue.** This is the #1 thing
  to resolve before flipping the schedule on. (Flagged, not yet fixed — it's a
  maturity-ladder mechanism question, not one of tonight's three asks.)
- **C — the `research`/`stewards-explore` agent gap** (below; fixed properly).

## What's done — UPDATE 2026-06-15 (you said "proceed with fixing and finishing")

Everything that makes the **engine work** is now done, verified, and pushed:

- ✅ **`research` agent** — core-seeded, unified off `stewards-explore`,
  web-capable, virgin-smoke-guarded, fresh-install-verified. `b116ab9`.
- ✅ **`yt` pipeline** — bridge rebuilt `WITH_YT`, 7/7 servers OK.
- ✅ **Finding B (the autonomy blocker) — FIXED + PROVEN.** Root cause: the
  review-verify gate only promotes a review stage to `verified` when its output
  starts with `REVIEW: passes|revised`, but planning's `review_plan` emitted a
  (dead, never-parsed) JSON verdict. Switched it to the `REVIEW:` convention. A
  fresh Vivint cycle then reached `verified` and **auto-enqueued 3 proposals with
  no manual call** (vs 0 before). `dd34715`.
- ✅ **Context + memory + pool-read grants** — `research` now has `doc_search/
  get/similar` + `compact_context` + the durable-memory & context tools
  (remember/forget/expand/summarize/context_*), minus the gated base-prompt
  editor. `dd34715`.
- ✅ **Finding A** — `work_item_create` now injects `input.today`, so no launch
  (manual or scheduled) hard-fails on it. Verified on live. `f8ea1bb`.

The loop now closes end-to-end on its own: launch → 5 stages → `verified` →
auto-enqueue. The dry-run mandate is fully met.

## The go-live bundle — BUILT + PROVEN (`c285ff5`, `22-reflect-steward.sql`)

Built to your two decisions (kill switch = global + per-intent; approve = queue,
not auto-dispatch, capacity-gated so work never floods):

1. **Kill switch** — `config.autonomy_paused` gates `scheduled_pipelines_fire`
   AND the drain (global stop); `reflect_pause_intent('<intent>')` disables that
   intent's schedules + makes the drain skip it (decommission one runaway intent
   while the rest run). *Note: the global switch now also governs the existing
   book/playlist digesters — one autonomy switch for everything.*
2. **Approval queue + capacity-gated drain** — `reflect_approve` QUEUES a
   proposal (does not fire it); `reflect_drain_approved` (every tick) launches
   approved proposals oldest-first only while in-flight < `reflect_max_concurrent`
   (default 2). *Proven on live: pause→drain=0; approve→in_flight stays 0;
   cap=0→0, cap=2→1.*
3. **Check-in verbs** — `reflect_status / reflect_proposals / reflect_approve /
   reflect_decline / reflect_steer / reflect_pause(_intent) / reflect_resume(_intent)`,
   driven via the `reflect-checkin` skill.
4. **Vivint schedule** — `scheduled_pipelines` row `vivint-reflect` (planning,
   cron `0 */3 * * *`, the Vivint intent), seeded **`enabled = false`**.

Verified on a fresh container (virgin-smoke OK 7, chain 00→22).

## The only thing left — the flip (your Hinge)

```sql
-- 1. turn the schedule on (it then fires every 3h):
UPDATE stewards.scheduled_pipelines SET enabled = true WHERE slug = 'vivint-reflect';
```
2. I start the **Claude-watchman wake** (`ScheduleWakeup`, every 2–5h) to run
   `reflect_status` and `reflect_pause` if anything drifts.

Say "flip it on" and I'll run both. To stop at any time:
`SELECT stewards.reflect_pause('<why>');`. Until you flip it, nothing runs on its
own (the schedule is disabled and `autonomy_paused` is off but there's nothing
enabled for it to fire).

## Decisions for you

1. **The 5 (well, 8) proposals — worth acting on?** The go/no-go for the whole
   idea. (My read: yes — grounded, actionable, ending at ticketable epics.)
2. **Flip it on?** Say the word and the Vivint steward starts running every 3h
   with the watch in place. Or approve a couple of the existing proposals first
   (`reflect_approve <slug>`) to watch one real research cycle run end-to-end
   before committing to the schedule.
