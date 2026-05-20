---
title: Substrate pipelines expansion — beyond study-write
date: 2026-05-11
status: design proposal — needs ratification before build
parent: open-items.md (Michael's 2026-05-11 ask — three new pipeline categories)
purpose: >
  The substrate's six-phase agentic creation cycle is built; today only
  study-write + study-write-qwen + echo-test pipelines exercise it.
  Three new pipeline categories the substrate is ready for: research
  (e.g. "AI news at 7am"), YouTube analysis (gospel + secular), and
  scheduled-pipeline infrastructure. This proposal scopes each.
---

# Substrate pipelines expansion

## I. Binding problem

The substrate has powerful infrastructure — intent + covenant + gates + sabbath + atonement + trust + councils — but only **one shape of work** uses it today: writing studies from the corpus. Three categories of work are natural fits for the same infrastructure but have no pipeline definitions yet:

1. **Research pipelines.** Cast a wider net than scripture-study — gather + summarize external sources on a topic. Example concrete use: "Summarize today's AI news at 7am UTC so I can see what shipped overnight."
2. **YouTube analysis pipelines.** Two paths sharing transcript+ingest but diverging on rubric:
   - **YT gospel** — evaluate a video against the Restoration discernment standard (current `yt-gospel` agent's lane)
   - **YT secular** — extract patterns from non-doctrinal content (current `yt` agent's lane — AI / engineering / relationships / skills)
3. **Scheduled pipelines.** Today every work_item is human-triggered or watchman-pressure-triggered. Some pipelines benefit from cron-style scheduling: "every weekday at 7am, run AI-news-summary."

All three are within reach because the substrate primitives are built. What's missing: pipeline definitions, supporting MCP tools where they don't exist yet, and a scheduling table.

## II. Success criteria

1. **A research pipeline exists** that takes a topic + sources spec, gathers content via web search + page fetch, synthesizes a structured summary, and produces a substrate-promoted study (or a different output kind — see V.1).
2. **YouTube gospel and YouTube secular pipelines exist** that share an ingest-transcript stage but diverge on the analysis rubric. Each produces a substrate-promoted evaluation document.
3. **Scheduled pipelines work.** A `stewards.scheduled_pipelines` config row says "run pipeline_family=ai-news-summary every weekday at 13:00 UTC with this input template." Bgworker tick dispatches when due.
4. **The first scheduled run actually fires** end-to-end without manual intervention. Output appears in the substrate by 7am local time the next morning.

## III. Constraints and boundaries

**In scope:**
- 3–4 new pipeline definitions in `stewards.pipelines` (research, yt-gospel-evaluate, yt-secular-digest, ai-news-summary as the canonical scheduled example)
- `stewards.scheduled_pipelines` table + bgworker scheduler tick extension
- Optional: new MCP tools where gaps exist (research probably wants a more capable `fetch_url` or paginated search; YouTube needs the existing yt-mcp tools wired to pipeline agents)
- Substrate-UI surface: scheduled-pipelines view (list + edit cron), "Last 7 scheduled runs" badge on dashboard

**Out of scope:**
- Building NEW agents for these pipelines from scratch — reuse existing `research`, `yt-gospel`, `yt` agents from the workspace where possible, then tune via the deferred `phase-3h` per-model-prompt-tuning effort
- Multi-tenancy on schedules (single-user only for now)
- Email / push notifications on scheduled completion (it'll just appear in Studies + Sabbath log when finished)
- Web-app scrape engines beyond fetch-md-mcp (Readability + chromedp js mode) and exa-search

## IV. Prior art

- **fetch-md-mcp** (Phase 3e.4) — Mozilla Readability + html-to-markdown + chromedp js mode. 4 tools: `fetch_url`, `fetch_urls`, `extract_links`, `fetch_url_raw`. Already granted to `research` agent.
- **exa-search** — remote MCP at mcp.exa.ai. Already used by research agent for web search.
- **yt-mcp** — Go MCP at `scripts/yt-mcp/`. 4 tools: `yt_download`, `yt_get`, `yt_list`, `yt_search`. Downloads transcripts via yt-dlp.
- **byu-citations** — for the gospel YT path (scripture citation density checking).
- **Phase 2.7b.2 scheduler tick** — bgworker already has a 60s scheduler tick that fires watchman passes on cron OR pressure triggers. Same mechanism extends to scheduled pipelines.
- **work_item_create + work_item_dispatch_stage** — the canonical entry points. Scheduled pipelines just call these on cron.

## V. Proposed approach

### V.1 Research pipeline

**Pipeline family:** `research-summary` (placeholder name; ratify in §VI)

**Stages:**
1. **gather** — `research` agent uses `web_search_exa` + `fetch_url` to collect 5–15 sources on the topic. Outputs a list of {url, title, retrieved_at, excerpt}.
2. **synthesize** — same agent (or kimi-k2.6 for synthesis) takes the gather output + intent + binding question, writes a structured summary. Output: markdown with sections (Headlines, Notable, Skeptical-takes, Carry-forward).
3. **review** — quick voice + verification pass (does each claim cite a source? are there obvious holes?).

**Input shape:**
```jsonb
{
  "binding_question": "What shipped in AI today that I should know about?",
  "sources_spec": {
    "queries": ["AI news today", "claude code update", "openai release"],
    "max_per_query": 10,
    "since": "24h"
  },
  "output_kind": "daily-digest"   -- vs 'deep-research'
}
```

**Output kind:**
- `daily-digest` — short summary, no substrate-study promotion (lives in stewards.studies with kind='daily-digest', browsable in UI but doesn't pollute Studies list with one-off news)
- `deep-research` — full study promotion, eligible for sabbath + atonement like any study-write run

### V.2 YouTube gospel pipeline

**Pipeline family:** `yt-gospel-evaluate`

**Stages:**
1. **ingest** — `yt` agent (or a thin agent variant) calls `yt_download(video_url)` to capture transcript. Outputs transcript text + metadata.
2. **evaluate** — `yt-gospel` agent applies the Restoration discernment rubric. Checks scriptural citation density via byu-citations where claims are made. Output: evaluation document with sections (binding question, evidence, alignment with canon, witness questions, becoming).
3. **review** — voice + source verification.

**Input shape:**
```jsonb
{
  "binding_question": "Does X's argument about Y align with the Restoration framework?",
  "video_url": "https://www.youtube.com/watch?v=...",
  "evaluator_focus": ["doctrinal" | "rhetorical" | "fruit-bearing"]   -- optional
}
```

### V.3 YouTube secular pipeline

**Pipeline family:** `yt-secular-digest`

**Stages:**
1. **ingest** — same as V.2.1
2. **digest** — `yt` agent extracts what's worth keeping (insights, contradictions to existing notes, what to skeptically question). Output: digest document with sections (1-sentence summary, key claims, contradictions, application).
3. **review** — voice pass.

**Input shape:**
```jsonb
{
  "binding_question": "What about Nate B Jones's talk is worth holding on to?",
  "video_url": "...",
  "context_tags": ["ai", "engineering", "agents"]   -- aids cross-reference
}
```

### V.4 Scheduled pipelines

**Schema:**

```sql
CREATE TABLE stewards.scheduled_pipelines (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            text UNIQUE NOT NULL,            -- 'ai-news-7am'
    pipeline_family text NOT NULL REFERENCES stewards.pipelines(family),
    intent_id       uuid NOT NULL REFERENCES stewards.intents(id),
    cron_pattern    text NOT NULL,                   -- '0 13 * * 1-5' (UTC; 7am MT weekdays)
    input_template  jsonb NOT NULL,                  -- merged into work_item.input on dispatch
    enabled         boolean NOT NULL DEFAULT true,
    last_dispatched_at timestamptz,
    next_due_at     timestamptz,                     -- materialized by scheduler tick
    created_at      timestamptz NOT NULL DEFAULT now(),
    notes           text
);

CREATE INDEX scheduled_pipelines_due
    ON stewards.scheduled_pipelines (next_due_at)
    WHERE enabled = true;
```

**Bgworker scheduler tick:**
- Extends the existing 60s watchman scheduler tick (`extension/2-7b2-watchman-scheduler.sql`).
- Each tick: scans `scheduled_pipelines WHERE enabled = true AND next_due_at <= now()`.
- For each due row: `work_item_create(pipeline_family, input_template, slug=auto-from-cron-pattern-+-now)`, then `work_item_dispatch_stage(new_id)`, then UPDATE `last_dispatched_at = now()` and recompute `next_due_at` via a small cron-parser.
- Cron parsing: a minimal subset is fine for v1 (`MM HH * * D`). Skip ranges, lists, ?, L, etc. Use a Go function in stewards-cli or a small Rust crate (`cron-parser` if licensing fits).

**Seed:**
```sql
INSERT INTO stewards.scheduled_pipelines (slug, pipeline_family, intent_id, cron_pattern, input_template, notes) VALUES
    ('ai-news-7am',
     'research-summary',
     (SELECT id FROM stewards.intents WHERE slug='scripture-study'),   -- or a new 'professional-awareness' intent
     '0 13 * * 1-5',                                                   -- 7am MT weekdays in UTC
     '{"binding_question":"What shipped in AI today that I should know about?","sources_spec":{"queries":["AI news today","claude release","openai update","anthropic announcement"],"max_per_query":10,"since":"24h"},"output_kind":"daily-digest"}'::jsonb,
     'Daily AI news digest. Weekdays 7am MT. Output is daily-digest kind (lives in studies, doesn''t bloat the Studies list).');
```

### V.5 Stewards-UI surfaces

- New `/scheduled` route showing scheduled_pipelines with enable/disable toggles + edit-cron modal + "next due" countdown.
- Dashboard card: "Last 7 scheduled runs" — quick list with status badges + click-through to the work_item.
- NewWork.vue: pipeline dropdown grows from {study-write, echo-test} to include {research-summary, yt-gospel-evaluate, yt-secular-digest, study-write, study-write-qwen, echo-test}. Per-pipeline input form schema (research-summary needs source queries; yt-* needs video URL; study-* needs binding question).

### V.6 Pipeline-stage maturity mapping

Per phase-b D-B4 ratification: maturity-to-stage mapping in `stewards.pipeline_stage_maturity`. New pipelines need rows:

```sql
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('research-summary',   'gather',    'researched', 'sources collected'),
    ('research-summary',   'synthesize','planned',    'structure proposed'),
    ('research-summary',   'review',    'verified',   'voice + verification'),
    ('yt-gospel-evaluate', 'ingest',    'researched', 'transcript captured'),
    ('yt-gospel-evaluate', 'evaluate',  'planned',    'rubric applied'),
    ('yt-gospel-evaluate', 'review',    'verified',   'voice + verification'),
    ('yt-secular-digest',  'ingest',    'researched', 'transcript captured'),
    ('yt-secular-digest',  'digest',    'planned',    'patterns extracted'),
    ('yt-secular-digest',  'review',    'verified',   'voice + verification');
```

Sabbath enabled per D-D1: all three default ON (they're study/lesson/talk-like work).

## VI. Decision points for Michael

These need ratification before build.

- **D-PE1: Research pipeline output kind.** Two options offered (daily-digest vs deep-research). Default `daily-digest` for scheduled runs (don't bloat Studies); default `deep-research` for human-triggered "really dig into X." Or: always deep-research, just tag scheduled runs differently. **Recommendation: keep both; let the input shape carry it.**
- **D-PE2: Should research / YT pipelines share intent with scripture-study, or get their own?** scripture-study's values include `faith-as-framework` and `trust-the-discernment` which fit YT-gospel but feel weird for AI news. Recommend: **separate intent for professional-awareness work** (`professional-awareness` slug; values: stay-current / honest-skepticism / breadth-then-depth). YT-gospel can stay under scripture-study.
- **D-PE3: Scheduled pipeline frequency cap.** Should the scheduler refuse cron patterns that fire more than once per hour to prevent runaway? Recommend: yes, hard floor 1h between runs of the same scheduled_pipelines row.
- **D-PE4: Catch-up on missed runs.** If the substrate is down at 7am, when it comes back up should it run the missed schedule? Recommend: no — fire once per next_due_at being reached, no backfill. Otherwise a 3-day outage produces a flood.
- **D-PE5: YT pipeline scope.** v1 ingests transcripts only (no video frames, no audio analysis beyond what yt-dlp's transcript provides). Confirm.
- **D-PE6: Cron-pattern v1 subset.** `MM HH * * DAY` is enough for "every weekday at 7am" + "every Sunday at noon." Ranges + lists + step values can wait. Confirm.
- **D-PE7: AGE Cypher integration for research output.** Should research-summary outputs participate in the AGE citation graph? Recommend: yes for deep-research, no for daily-digest. Daily digests pollute the graph; deep-research benefits from it.

## VII. Estimated programming time

- V.1 Research pipeline definition + seed: 1 session
- V.2 + V.3 YouTube pipelines (gospel + secular share most): 1 session
- V.4 Scheduled pipelines schema + bgworker tick + cron parser: 1–2 sessions
- V.5 UI surfaces (new /scheduled route + dashboard card + per-pipeline form): 1 session
- First real e2e scheduled run + tuning: ~30 min (overnight; observe morning)

**Total: 4–5 sessions.**

Should be sequenced AFTER Batch G (file-write mechanism lets these new pipelines actually produce on-disk outputs). Could ship the schema + pipelines first (PE.1 + PE.2/3), then scheduler (PE.4 + PE.5), as two separate batches.

## VIII. Acceptance scenarios

1. `INSERT INTO stewards.scheduled_pipelines (ai-news-7am, ...)`. Bgworker scheduler tick at next 13:00 UTC dispatches a research-summary work_item. By 13:05 UTC the gather stage chat has fired; by 13:10 the synthesize is done; daily digest appears in `/studies?kind=daily-digest`.
2. Manual `work_item_create('yt-gospel-evaluate', {binding_question, video_url})` via NewWork form. Stage 1 calls `yt_download`, stage 2 evaluates, stage 3 reviews. Output study appears in Studies list with kind='gospel-evaluation'.
3. Cron pattern '0 * * * *' rejected with `frequency_floor_violation` (D-PE3 hard floor).
4. Substrate down for 6 hours covering a scheduled run; on recovery, scheduler advances `next_due_at` to the NEXT scheduled time, doesn't backfire the missed one (D-PE4 no-backfill).

## IX. Why now

The substrate's primitives are mature; one pipeline is a poor showcase of what they enable. The proposed pipelines also have direct utility: a daily AI news digest at 7am is something Michael will actually read. YouTube digestion replaces an ad-hoc manual workflow.

The scheduled-pipelines machinery is independently useful — once it exists, future pipelines (e.g. weekly "stewardship review", monthly "studies-not-yet-cited audit") become 5-minute additions, not 5-day proposals.

---

## X. Ratification + PE-A build log (2026-05-19)

Council ① of the post-ES-arc queue. Decisions ratified, two new ones surfaced during build, PE-A shipped in one session.

### X.1 Decisions ratified

| ID | Choice | Note |
|----|--------|------|
| D-PE1 | Input field carries it | Both kinds, output_kind on the input |
| D-PE2 | New 'professional-awareness' intent | Initial decision — superseded by D-PE2' during build |
| D-PE3 | No hard floor | Trust the operator; cost-cap + quarantine + bucket caps are the safety net |
| D-PE4 | Fire one missed run on recovery | Scheduler needs a missed-window threshold (probably 24h) |
| D-PE5 | Transcript + metadata enrichment | yt ingest also pulls chapters + full description + top comments |
| D-PE6 | Standard 5-field cron with ranges + lists | Pulls in a Rust cron crate; PE-B will need a soak pause + pg rebuild |
| D-PE7 | All research output in the graph | Both research-write and research-summary; daily-digest included |

### X.2 Two decisions surfaced during build

**D-PE1' — pipeline selection itself is a judgment.** When `research-write` was discovered already in production (15 work_items, latest 2026-05-17), Option B was ratified: keep research-write for deep-research, add research-summary for daily-digest, and let the agent (or the human in NewWork) judge which pipeline fits the task. Same shape as `study-write` vs. `study-write-qwen` splitting on model choice. The "one family with output_kind switching" framing in §V.1 is superseded by per-family routing.

**D-PE2' — reuse `general-research` over creating `professional-awareness`.** When the existing `general-research` intent was discovered (with concrete values + non-goals already source-backed via `.spec/intents/general-research.yaml`), reuse was chosen with two YT-aware values appended (`separate-claim-from-charisma`, `surface-the-rhetoric`). The "rule of three" Rust parser refactor stays deferred since no third intent is being introduced.

**D-PE7' — non-study-write promotion path was missing.** During PE.5 build, found that `work_item_promote_to_study` is hardcoded to `study-write*` and `on_maturity_verified` only enqueues the file write — never inserts into `stewards.studies`. So D-PE7 ("all research output in the graph") required first building the studies-promotion path for non-study-write pipelines. Scope B chosen: narrow promotion path + backfill of the 15 existing research-write rows.

### X.3 PE-A shipped (2026-05-19)

Five sub-steps, five commits, zero rollbacks:

| Sub-step | What | File |
|----------|------|------|
| PE.1 | Two YT-aware values appended to `general-research` intent | `extension/h1-1-general-research-intent.sql` (edit) + `.spec/intents/general-research.yaml` (doc sync) |
| PE.2 | `research-summary` pipeline (daily-digest; sabbath/atonement OFF; auto-materialize ON) | `extension/pe2-research-summary-pipeline.sql` |
| PE.3 | `yt-gospel-evaluate` pipeline (sabbath/atonement ON; agent yt-gospel; byu_citations/gospel_*/yt/* perms already granted) | `extension/pe3-yt-gospel-evaluate-pipeline.sql` |
| PE.4 | `yt-secular-digest` pipeline (sabbath/atonement ON; agent yt; cross-reference via study_search_text/brain_search) | `extension/pe4-yt-secular-digest-pipeline.sql` |
| PE.5 | `promote_to_study()` for the four non-study-write families + `on_maturity_verified` wiring + backfill of 14/15 research-write runs (1 skipped, un-sabbathed) | `extension/pe5-promote-to-study-non-study-write.sql` |

Smoke after PE.5:
- `stewards.studies` kinds now: study/195, proposal/73, journal/70, doc/32, **research/14**, phase-doc/1
- AGE :Study nodes kind=research: **14** (was 0)
- CITES edges from research nodes: 0 (expected — secular research outputs have no gospel-library citations)

### X.4 Still pending — PE-B and PE-C

PE-A intentionally scoped to SQL-only changes; no pg rebuild needed; soak stayed running throughout. Remaining work:

- **PE-B (scheduled machinery)** — `stewards.scheduled_pipelines` schema + bgworker scheduler-tick extension + Rust cron crate + fire-one-missed logic. Requires soak pause + pg rebuild (Cargo.toml change).
- **PE-C (UI surfaces)** — `/scheduled` route + dashboard "Last 7 scheduled runs" card + NewWork.vue per-pipeline forms. UI-only; no soak pause needed.

### X.5 Carry-forward for future sessions

- **1 un-sabbathed research-write row** (id `2c7a501d-eb6e-4cbe-ad0d-44ebf482353e`) skipped by the sabbath gate during PE.5.C backfill. Call `stewards.sabbath_dispatch(<id>)` then `stewards.promote_to_study(<id>)` to bring it into the graph.
- **The proposal-was-stale pattern.** The proposal didn't anticipate `research-write`, `general-research`, registered `yt`/`yt-gospel` agents, or the missing non-study-write promotion path. Three duplications caught in a single session. Future councils should `\d` the table + `SELECT ... FROM stewards.pipelines / intents / agents` before assuming clean-slate. The pattern is recorded in `2026-05-19-substrate-pe-a-shipped.md`.

---

## XI. PE-B + PE-C ship log (2026-05-19, same session as PE-A)

After PE-A closeout, Michael chose to push through PE-B + PE-C in the same session ("pausing only for ratifications/clarifications"), then test all three pipelines end-to-end, then sabbath-close council ①. Four PE-C sub-steps shipped on top of PE-B; one cron-crate-vs-plpgsql implementation reframe; one end-to-end smoke of all three pipelines.

### XI.1 Implementation reframe — PE-B as plpgsql, not Rust crate

The pre-PE-B survey counted 86 post-G SQL files needing replay after a pg rebuild (the rebuild-discipline pattern). Pulling in a Rust `cron` crate (the original PE-B plan implied by D-PE6) would have forced that rebuild + replay event in this session. Reframe: implement the cron parser in plpgsql instead, keeping the pg rebuild a future dedicated session with its own attention.

- D-PE6 (standard 5-field with ranges + lists + step values) is honored entirely in `stewards.cron_field_values` + `stewards.cron_next_after`.
- `cron_next_after` runs in plpgsql (~50ms worst case per call) and is called once per dispatch — performance fine for scheduled workloads.
- Soak stayed paused through PE-B as a defensive measure, then resumed unchanged after; no work was lost.

### XI.2 PE-B shipped — scheduled machinery

| Sub-step | File | What |
|----------|------|------|
| PE-B.1 | `extension/pe6-scheduled-pipelines-schema-and-cron.sql` | `stewards.scheduled_pipelines` table (uuid pk, slug + FKs to pipelines + intents, cron_pattern, input_template, enabled, missed_window_hours default 24) + `cron_field_values()` + `cron_next_after()` (plpgsql, OR-semantics between day-of-month and day-of-week) + BEFORE INSERT/UPDATE trigger that materializes `next_due_at`. Smoke: five canonical cron patterns all return expected `next_due_at`. |
| PE-B.2 | `extension/pe7-scheduled-pipelines-fire-and-tick.sql` | `stewards.scheduled_pipelines_fire()` (FOR UPDATE SKIP LOCKED scan, dispatch via `work_item_create` + `work_item_dispatch_stage`, D-PE4 fire-one-missed with `missed_window_hours` cutoff, child-slug pattern `<schedule.slug>--YYYY-MM-DD-HHMM` in UTC) + extends `watchman_scheduler_fire` to call the new function at the top (so scheduled pipelines fire even when watchman soak is paused) + seeded `ai-news-7am` row pointing at research-summary with cron `0 13 * * 1-5`. Smoke: forced-due test dispatched + completed cleanly via echo-test; 25h-stale test was correctly skipped + advanced. |

No bgworker.rs change. Both schedulers ride the existing 60s leader tick.

### XI.3 PE-C shipped — UI surfaces

| Sub-step | Files | What |
|----------|-------|------|
| PE-C.1 | `scripts/stewards-ui/api/scheduled.go` + `api/api.go` (Register hookup) | `GET /api/scheduled/list`, `GET /get?id=\|slug=`, `POST /create`, `PUT /update?id=`, `POST /toggle?id=`, `DELETE /delete?id=`, `GET /recent-runs?limit=N`. `recent-runs` identifies scheduler-spawned work_items via `actor='scheduler'` and exposes both the schedule slug (split on `--`) and the work_item slug for click-through. |
| PE-C.2 | `frontend/src/views/Scheduled.vue` + `api.ts` types + `router.ts` + `App.vue` nav | Full-CRUD `/scheduled` route: list with enabled checkbox (live toggle), next-due countdown that ticks every second, edit modal (cron + JSON input_template + missed_window + notes), create modal with pipeline-family dropdown (all current + new pipelines) and intent-slug dropdown. Sort: enabled-first, then by next_due_at. |
| PE-C.3 | `frontend/src/views/Dashboard.vue` | "Last 7 scheduled runs" table card added after Recent Errors. Loads on the same 5s refresh cycle; isolated error state so a /api/scheduled error does not flag the overall dashboard. Status badges + click-through to WorkItemDetail. Empty state explicitly points to the `ai-news-7am` seed. |
| PE-C.4 | `frontend/src/views/NewWork.vue` | Per-pipeline input fields composed dynamically into `inputJson` based on selected pipeline: research-* gets a sources_spec block (queries textarea, max_per_query, since); yt-* gets a video URL field; yt-gospel-evaluate gets an evaluator_focus dropdown; yt-secular-digest gets a context_tags input. The pipeline dropdown already populated from `/api/pipelines/list` so the three new pipelines appeared automatically once PE-A landed. |

ui container rebuilt + restarted once at the end of PE-C (no pg restart since PE-B was pure SQL). Deployment smoke: `/healthz` ok; `/api/scheduled/list` returns the ai-news-7am seed with `next_due_at: 2026-05-20T13:00:00Z` and full input_template; `/api/scheduled/recent-runs` returns the PE-B.2 smoke echo work_item.

### XI.4 PE-final — end-to-end smoke of all three pipelines

Three real work_items dispatched, one per new pipeline:

| Pipeline | work_item slug | Input |
|----------|----------------|-------|
| `research-summary` | `pe-final-research-summary-smoke` | "What shipped in AI today that I should know about?" + sources_spec with 3 queries (AI news, claude release, anthropic announcement), max_per_query=5, since=24h |
| `yt-secular-digest` | `pe-final-yt-secular-pinecone-knowledge-layer` | Nate B Jones, "Pinecone Just Demoted Vector Search. Here's the Knowledge Layer." (May 13). Context tags: ai, vector-search, knowledge-graph, substrate. Binding question ties it to the substrate's vector+AGE-graph approach. |
| `yt-gospel-evaluate` | `pe-final-yt-gospel-morgan-philpot` | Morgan Philpot (`https://youtu.be/9UTrPgjLW7g`), from Michael's hometown. Binding question on alignment with the Restoration framework. |

Status at dispatch + 3 minutes:
- research-summary: synthesize stage in progress, 157k tokens, $0.17
- yt-secular-digest: digest stage in progress, 44k tokens, $0.05
- yt-gospel-evaluate: ingest stage just started, 6k tokens, $0.006

End-to-end status (final): see XI.5 below.

### XI.5 Council ① closed

**Total PE-A + PE-B + PE-C: 13 commits across one calendar day (2026-05-19), zero rollbacks.** SQL-only changes throughout for the substrate side. Go + Vue + container rebuild for UI. The `/scheduled` route surfaces a new substrate primitive (scheduled cron-style pipelines) that ② `substrate-scheduled-workflows` will use directly when it gets walked. PE-C is the small preview slice of what ③ `stewards-ui-evolution` builds out — write-action UIs, per-row CRUD, sidebar grouping.

Council ② and ③ remain. Next session can start ② whenever Michael is ready; PE-B's `scheduled_pipelines` schema and `cron_next_after` carry forward as ready-to-use machinery.

### XI.6 Carry-forward (in addition to X.5)

- **Final work_item end-states for the three pipelines** captured in `2026-05-19-substrate-council-1-closed.md` once they all reach terminal status.
- **The pg rebuild stays deferred.** 86 post-G SQL files would need replay. Next time a pg rebuild becomes necessary (new pgrx function, new Rust crate dep, breaking schema change), the rebuild-discipline ritual runs: rebuild → replay all post-G files in dependency order → smoke. Until then, all new SQL applies via live-apply.
- **The proposal-was-stale pattern stays a council ritual.** Already named once in X.5; held across PE-B + PE-C with no further duplications surfaced (Scheduled.vue, Dashboard.vue, NewWork.vue all read from current schema before extending).
