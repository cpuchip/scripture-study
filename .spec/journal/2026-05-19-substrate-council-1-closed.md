---
title: substrate council ① CLOSED — pipelines-expansion fully shipped
date: 2026-05-19
workstream: WS5
status: shipped
priority: high
---

# Substrate council ① closed — pipelines-expansion fully shipped

In one calendar day, PE-A + PE-B + PE-C + the end-to-end smoke of all three new pipelines all shipped. 13+ commits, zero rollbacks. Council ① closed at sabbath-close.

## The arc

Started this morning with a soak-error review (9 errors, all from the ES.4 closeout window, all recovered). Walked the seven D-PE decisions via two AskUserQuestion batches. Shipped PE-A as a SQL-only slice. Closed it out with proposal §X + journal + memory. Michael said let's keep going — push through PE-B + PE-C, test all three together, sabbath-close at the end.

Three more reframes surfaced over the next few hours, each caught before code:

- **D-PE1' Option B** (PE-A) — pipeline selection is itself a judgment, not output_kind branching inside one pipeline. `research-write` already in production with 15 work_items; keep both research-write and research-summary, let the agent (or human in NewWork) judge which fits.
- **D-PE2' reuse general-research** (PE-A) — existing intent already covered the ground the proposal called "professional-awareness"; reuse + append two YT-aware values rather than create a duplicate.
- **D-PE7' build the missing promotion path** (PE-A) — auto-materialize wrote to disk but never inserted into stewards.studies for non-study-write pipelines; PE.5 grew from "wire AGE" to "build the missing promotion path + backfill 14 of 15 research-write rows."
- **PE-B as plpgsql, not Rust crate** — the survey before PE-B counted 86 post-G SQL files. Pulling in a Rust cron crate would have forced a pg rebuild + replay event with its own attention. Reframe: implement cron_next_after in plpgsql (~150 lines, ~50ms per call), keep the rebuild deferred to a dedicated future session.

Two of these were "the proposal didn't know about existing state." Two were "what I was about to build conflicts with what already exists." All four were caught by `check_existing_work` — query the table, read the existing function, count the files — before writing the new code.

## What shipped

**PE-A** (5 sub-steps, SQL-only):
- general-research intent extended with two YT-aware values
- research-summary pipeline (daily-digest)
- yt-gospel-evaluate pipeline (sabbath+atonement on, agent yt-gospel)
- yt-secular-digest pipeline (sabbath+atonement on, agent yt)
- promote_to_study + on_maturity_verified wiring + 14-of-15 research-write backfill

**PE-B** (2 sub-steps, SQL-only, plpgsql cron):
- stewards.scheduled_pipelines schema + cron_field_values + cron_next_after (standard 5-field with ranges, lists, step values per D-PE6) + BEFORE INSERT/UPDATE trigger
- scheduled_pipelines_fire dispatcher (FOR UPDATE SKIP LOCKED, fire-one-missed per D-PE4, child-slug pattern with UTC date suffix) + wired into watchman_scheduler_fire so scheduled pipelines fire even when watchman soak is paused + ai-news-7am seed (cron 0 13 * * 1-5)

**PE-C** (4 sub-steps, Go + Vue, one container rebuild):
- /api/scheduled/* endpoints (list, get, create, update, toggle, delete, recent-runs)
- /scheduled Vue route — full CRUD with cron + JSON template editor + next-due countdown + enable/disable toggle + sort by enabled-then-next-due
- Dashboard "Last 7 scheduled runs" card (loads on the same 5s refresh, isolated error state)
- NewWork.vue per-pipeline input fields composed dynamically into inputJson based on selected pipeline (research-* gets sources_spec; yt-* gets video URL + per-family optional fields)

**PE-final** (3 real pipeline dispatches):
- research-summary: "What shipped in AI today" with 3-query sources_spec → 3383-char daily-digest, $0.21
- yt-secular-digest: Nate B Jones on Pinecone + vector search → 2796-char yt-digest, $0.08
- yt-gospel-evaluate: Morgan Philpot (from Michael's hometown) → 3370-char gospel-evaluation, $0.07
- All three completed, sabbathed where required, in studies + AGE. Total $0.36.

## What surprised me

**The proposal-was-stale pattern, now confirmed across two batches.** Caught three times in PE-A (research-write, general-research, yt-agents) plus the missing promotion path. Held cleanly through PE-B + PE-C — no further duplications surfaced because I was now actively checking before extending.

**The 86-file count gave me pause.** Counted before PE-B because I assumed pg rebuild was inevitable. 86 was big enough that "go with plpgsql" became obviously right. That single count changed the shape of PE-B from "Cargo.toml + Rust crate + rebuild + replay" to "150 lines of plpgsql." The rebuild-discipline memory was right: rebuilds are events, and they want their own attention.

**One bug surfaced + fixed mid-smoke.** yt-* pipelines were missing `auto_materialize_on_verified=true` in their seeds. When yt-secular-digest reached verified + sabbathed, on_maturity_verified's auto-materialize block never fired, so promote_to_study (wired inside that block in PE.5) also didn't fire. Caught when I queried stewards.studies and saw only research-summary. Fixed by (a) UPDATE pipelines SET auto_materialize_on_verified=true, (b) manually calling promote_to_study on the completed yt-secular-digest row, (c) updating the pe3/pe4 source SQL so the next rebuild gets it right. The yt-gospel-evaluate that was still running when I flipped the flag also needed the manual promote since it had already transitioned. This is the stewardship pattern: find a bug adjacent to your work, fix it, report it. Boundary test: "would Michael say yes obviously" — yes.

**Council ① closed via real-world fire of all three pipelines, not just deploy smoke.** Michael chose "dispatch all three now" over deferring to natural use. $0.36 was the price of real coverage. The yt-secular-digest particularly: the Nate B Jones video on Pinecone demoting vector search lands directly on the substrate's vector+AGE-graph approach. Smoke became a real study with carry-value.

## What carries forward

- **Council ② substrate-scheduled-workflows is the next thing.** PE-B's `stewards.scheduled_pipelines` table + `cron_next_after` + dispatcher are ready-to-use machinery. D-SW1–D-SW7 ratification questions are unwalked.
- **Pg rebuild stays deferred.** 86 post-G files still want replay when a rebuild becomes necessary. Until then, all new SQL applies via live-apply.
- **One un-sabbathed research-write row** (id `2c7a501d-...`) still skipped from the PE.5.C backfill. Run `sabbath_dispatch` then `promote_to_study` when it's worth bringing forward.
- **AGE n_cites=0 on the three smoke outputs.** parse_gospel_links didn't find gospel-library citations in any of the three. For yt-gospel-evaluate Morgan Philpot specifically, that's unexpected — gospel evaluations should cite scripture. Worth checking next session whether (a) the output legitimately had no scripture citations in linkable form, or (b) parse_gospel_links has a pattern gap for the link format the agent chose. Carry-forward investigation.
- **The four new substrate kinds** (`research`, `daily-digest`, `gospel-evaluation`, `yt-digest`) now coexist with the original five (`study`, `proposal`, `journal`, `doc`, `phase-doc`) in stewards.studies. The UI's Studies page filter dropdown should grow when ③ stewards-ui-evolution gets walked.

## What the work taught

**Reframes save more than they cost.** Each of the four mid-build reframes was 3–5 minutes of AskUserQuestion or self-survey. Each saved between "one duplicate to delete later" and "an 86-file replay event." Surface tensions before deciding is the highest-leverage covenant commitment when the codebase has grown faster than the proposal that ratified the work.

**The C–F cadence scales beyond SQL.** One sub-step per commit; smoke before commit (for SQL, live-apply + query; for Go, build; for Vue, npm run build; for end-to-end, real dispatch); journal + memory at the close. Held for 13 commits across SQL + Go + Vue + container rebuild + real LLM dispatches. The cadence is the moat, not the model.

**Stewardship over surfacing was a real call today.** The plpgsql-vs-Rust-crate decision could have been a question to Michael. Two factors made it a self-decision: (a) it honored a ratified spec (D-PE6 standard 5-field), and (b) Michael's "pause only for ratifications/clarifications" was explicit standing authorization for implementation choices. The yt-* auto_materialize fix similarly: would Michael say yes obviously? Yes. Fix and report. Don't surface-without-acting.

**Council ① is the first complete cycle through the new substrate.** Pipelines defined → dispatched → ran → completed → sabbathed (where required) → promoted to studies → indexed into AGE. The full Abraham-4 cycle (council → spiritual creation → physical creation → watch until obeyed → corrective action) on a per-work-item scale. Today the substrate watched itself; the yt-* bug surfaced because the watching was real, not performative.

Going into sabbath-close with all three pipelines visible as real evidence in studies + AGE.
