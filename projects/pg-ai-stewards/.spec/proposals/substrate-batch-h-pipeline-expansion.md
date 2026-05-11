---
title: Substrate Batch H — pipeline expansion (the BUILD→USE transition for many domains)
date: 2026-05-11
status: RATIFIED 2026-05-11 (D-H1 through D-H7); H.1 build-ready
parent: full-agentic-substrate.md (§VI ratifications); substrate-pipelines-expansion.md (light prior proposal — see §IV)
supersedes: substrate-pipelines-expansion.md (this proposal absorbs and replaces it; older proposal can be retired or kept as a historical-only reference)
purpose: >
  Six phases (A–F) built the substrate. Batch G plumbed it into the
  filesystem. Batch H is the arc where the substrate proves it
  generalizes off the single domain it was designed against
  (scripture-study). Five new pipeline families, sequenced. H.1
  (research-write) is spec'd in build-ready depth; H.2–H.5 are
  scoped at moderate depth with the load-bearing design
  decisions surfaced.
---

# Batch H — pipeline expansion

> **Status (2026-05-11):** §VI RATIFIED. Seven decisions recorded inline
> as **RATIFIED:** lines under each D-H. H.1 build-ready; scope expands
> per D-H2 / D-H3 / D-H5 ratifications (see §VI.summary below).
>
> **§VI summary of ratifications (2026-05-11):**
> - **D-H1 → A (lightweight + rule of three).** Recommended path.
> - **D-H2 → B (per-family ladder column NOW).** Heavier than recommended.
>   H.1 must add `pipelines.maturity_ladder jsonb` + seed study-write's
>   current rungs + research-write's rungs + refactor any code that
>   hardcodes the ladder to read from the column.
> - **D-H3 → B (three intents: general-research + professional-awareness
>   + creative-fidelity).** Heavier than recommended. H.1 creates
>   general-research; H.3 creates professional-awareness;
>   H.4 creates creative-fidelity. Storage: one file per intent in
>   `.spec/intents/<slug>.yaml`.
> - **D-H4 → A (tools_disabled=true payload flag).** Recommended path.
> - **D-H5 → B (per-work_item override).** Heavier than recommended.
>   H.1 must add `work_items.sabbath_enabled boolean NULL` +
>   `work_items.atonement_enabled boolean NULL` (NULL = inherit from
>   pipeline). Trigger / dispatch logic resolves work_item override first,
>   falls back to pipeline default.
> - **D-H6 → A (defer H.5 Bridge Sim NPCs to substrate-aware-chat).**
>   Recommended path. Document deferral; revisit when substrate-aware-chat
>   is designed. Fallback: H.4 fiction-scene can produce scripts if science-
>   center deadline forces it.
> - **D-H7 → A (ship origin column in H.3).** Recommended path. Add
>   `work_items.origin text` (human|scheduled|watchman|steward|council).
>
> **Net effect on H.1 scope:** original H.1 spec + four new schema/refactor
> items (maturity_ladder column + ladder code refactor + work_item sabbath/
> atonement overrides + general-research intent). Estimated ~+30-45 min of
> programming over the original H.1 estimate. The heavier choices (D-H2,
> D-H3, D-H5) all build substrate primitives that pay dividends across
> H.2–H.4 instead of needing retrofitting.

## Council moment — what this document is and isn't

This is not "more substrate." Phases A–F finished that work; Batch G
proved the substrate can land outputs as files. What's still untested
is whether the substrate's six rituals — intent, gates, scenarios,
verify, sabbath, atonement — are *gospel-shaped* (they only fit
scripture work) or *creation-shaped* (they fit any disciplined work
that has a binding question and an output).

The honest framing: today's substrate runs one shape of work
(study-write) and a placeholder (echo-test). Every primitive
generalizes *in theory*. Batch H is the experiment that finds out
which generalize *in practice*.

Two failure modes to design against:

1. **Lightweight generalization that hides coupling.** Copy
   study-write, swap tools, ship. Discover at pipeline 3 or 4 that
   some shared template or function quietly assumes scripture work.
   Refactor under duress.
2. **Heavyweight generalization that never ships.** Audit every
   shared template, function, and prompt for gospel-specific
   assumptions before adding any new pipeline. Spend three sessions
   in theory before producing one new pipeline.

The recommendation in §VI D-H1 is the middle path: lightweight with
*eyes open* — apply the rule of three. Build H.1 noting every fork
of substrate primitives. After H.2, if three primitives have forked,
pause and audit. If fewer, keep going.

---

## I. Binding problem

The substrate's primitives are built and proven on one pipeline
shape. Five categories of work want the same machinery but have no
pipeline yet:

1. **Non-gospel internet research** (e.g. "What shipped in AI today?")
2. **YouTube transcript analysis** (gospel + secular variants)
3. **Scheduled / cron-fired pipelines** (e.g. AI-news at 7am MT weekdays)
4. **Fiction / D&D / Bridge Sim creative writing** (premise → scene → draft)
5. **Bridge Sim live NPC dialogue** (Michael's science-center project)

These are not five copies of the same problem. They differ on:
- whether maturity-as-ladder fits the work shape (research yes; fiction no)
- whether tools belong on (research yes; fiction mostly no)
- whether sabbath / atonement are meaningful (research yes; AI-news daily digest probably no; fiction unclear)
- whether the work fits a `work_item` shape at all (Bridge Sim live dialogue probably doesn't)

The binding problem is: **port the substrate's discipline into each
domain in the lightest possible way that doesn't break the
discipline.** Not "make every primitive fire on every pipeline."

---

## II. Success criteria

1. **The substrate generalizes off scripture work.** At least three
   non-scripture pipeline families run end-to-end (research +
   yt-secular + scheduled-news minimally) using the substrate's
   intent/covenant/gate/sabbath machinery.
2. **The gospel-specific forks are named.** Every place a prompt
   template, function, or gate had to fork for a non-scripture
   domain is documented in §VII of this batch's retrospective so the
   next pipeline family knows where the seams are.
3. **A scheduled pipeline fires unattended.** AI-news-7am dispatches
   without manual trigger, runs to verified, materializes a daily
   digest, all before Michael wakes up the next morning. Cost stays
   under $0.50 per run.
4. **Fiction either fits or honestly doesn't.** H.4 builds the
   minimum fiction pipeline; if the maturity ladder doesn't fit, the
   batch's retrospective names that explicitly and proposes the
   per-pipeline-family-ladder solution as a future batch — rather
   than papering over with a forced fit.
5. **H.5 is correctly placed.** Either H.5 ships as a substrate
   pipeline OR it's explicitly deferred to a future "substrate-aware
   chat" workstream with the deferral reasoning recorded.

---

## III. Constraints and boundaries

**In scope:**
- New pipeline definitions: `research-write` (H.1), `yt-gospel-evaluate` + `yt-secular-digest` (H.2), `ai-news-summary` as canonical scheduled example + `stewards.scheduled_pipelines` table + bgworker scheduler hook (H.3), `fiction-scene` (H.4, minimal variant)
- New intent(s) for non-scripture work — at least `general-research`; possibly `professional-awareness` and `creative-fidelity`
- Per-pipeline-family forks of gate prompts where required (named, not implicit)
- A documented decision on whether sabbath / atonement enable per-family or per-pipeline (see D-H6)
- UI: NewWork.vue dynamic pipeline dropdown already exists from Batch G; per-pipeline input form variants needed
- A new `/scheduled` UI surface for H.3

**Out of scope:**
- New MCP tool development beyond what already exists (fetch-md-mcp, yt-mcp, exa-search, byu-citations all built). If H.1/H.2 surface a gap (e.g. needing a structured news-feed reader), that's a follow-up proposal
- Building new agent personas from scratch — reuse existing `research`, `yt`, `yt-gospel`, `fiction` agents
- H.5 implementation as a substrate pipeline (almost certainly belongs in a future "substrate-aware chat" workstream — see §V.5)
- Multi-tenancy on schedules (single-user)
- Catch-up / backfill on missed scheduled runs (D-PE4 from older proposal — keep that ratification: no backfill)
- Cron parser beyond `MM HH * * DAY` subset for v1

---

## IV. Prior art

- **`substrate-pipelines-expansion.md`** (2026-05-11) — lighter prior
  proposal covering research + YT + scheduled (no fiction, no live
  NPCs). Six decisions D-PE1..D-PE7 ratified. **This proposal
  supersedes it.** Three of its ratifications carry forward
  unchanged (D-PE3 frequency floor, D-PE4 no-backfill, D-PE6 cron
  subset); the rest are subsumed into D-H1..D-H6 below.
- **`full-agentic-substrate.md` §VI** — all ratified decisions A1–F4
  plus the 2026-05-11 amendments. Several decisions implicitly
  assumed scripture work; Batch H is the moment to make those
  assumptions explicit.
- **`substrate-completion-batch-g.md`** — file-write mechanism
  (pending_file_writes + materialize-writes CLI + pre-commit hook)
  lets new pipelines declare their own file_destination_templates
  without each one having to reinvent the materialization path.
- **`phase-c-design.md` §V.4** — `compose_system_prompt` injects
  active covenant + work_item intent. New pipelines automatically
  inherit this. Means: a new pipeline with its own intent gets the
  intent in its prompts for free.
- **Live database (read 2026-05-11):**
  - 3 pipelines today (`study-write`, `study-write-qwen`, `echo-test`).
    Stages for study-write: `outline → draft → review`. Maturity
    mappings: outline→planned, draft→executing, review→verified.
  - 9 gate_prompt templates (`evaluate`, `generate_scenarios`,
    `verify`, `covenant_check`, `sabbath`, `atonement`,
    `council_proposer`, `council_critic`, `council_synthesizer`).
    Each loaded by ID, rendered with `{{...}}` vars at dispatch time.
  - 21 stage_models rows. Schema is **(pipeline_family, stage_name,
    default_model)** — no agent_family or provider columns; those
    are read from `pipelines.stages[i]`.
  - One intent today: `scripture-study`. values_hierarchy carries
    9 entries (5 values + 4 constraints). scripture_anchor is empty
    on this intent (worth noting — even the canonical intent
    doesn't bind itself to one verse).
  - One active covenant, scope=`global`. ratified_by=`both`.
- **Agent tool grants are agent-level, not pipeline-level.** Live
  data: `research` has 27 allow grants, `yt-gospel` has 23, `yt` has
  14. The substrate enforces these at agent-dispatch time. **There
  is no per-pipeline tool restriction layer** — see D-H4.
- **Watchman scheduler (extension/2-7b2)** is *not* a generic cron.
  It's a singleton decision function for one workflow (when to fire
  watchman passes). H.3's scheduled-pipelines machinery is a NEW
  design, not a reuse.

---

## V. Proposed approach

Five sub-phases, in build order. Rationale for ordering:

| Sub-phase | What | Why this position |
|---|---|---|
| H.1 | `research-write` pipeline | First real non-scripture pipeline. Tests whether the maturity ladder + gates generalize cleanly when the inputs aren't doctrinal. Build-ready spec. |
| H.2 | `yt-gospel-evaluate` + `yt-secular-digest` | Two pipelines sharing an ingest stage but diverging on rubric. Tests per-family gate-prompt forking pattern surfaced in H.1. |
| H.3 | Scheduled pipelines infrastructure + ai-news-summary | Cron-fired dispatch. Tests "the work_item doesn't have a human originator" — every gate, sabbath, and atonement assumption that implicitly depended on a human convener is exposed here. |
| H.4 | `fiction-scene` (minimal) | Tests whether the maturity ladder fits creative work at all. Likely answer: not cleanly. Batch H's job is to find out and propose the future-batch fix. |
| H.5 | Bridge Sim NPC live dialogue | **Recommend defer.** Almost certainly not a `work_item`-shaped problem. See §V.5 for the recommend-defer rationale. |

### V.1 H.1 — `research-write` pipeline (build-ready spec)

#### V.1.1 Pipeline family

`research-write` — chosen to match the `study-write` shape (verb +
domain). The companion `study-write-qwen` model variant pattern
(qwen3.6-27b for the entire pipeline) can be added later as
`research-write-qwen` if cost or voice tuning calls for it.

#### V.1.2 Stages

Mirrors study-write's 3-stage shape with renames:

```jsonb
[
  {"name": "gather",     "agent_family": "research", "model": "kimi-k2.6", "provider": "opencode_go"},
  {"name": "synthesize", "agent_family": "research", "model": "kimi-k2.6", "provider": "opencode_go"},
  {"name": "review",     "agent_family": "research", "model": "kimi-k2.6", "provider": "opencode_go"}
]
```

Maturity mapping (matches study-write rungs intentionally — the
maturity ladder should be the SAME for both, only the stage names
differ):

| Stage | produces_maturity |
|---|---|
| gather | researched |
| synthesize | planned |
| review | verified |

Note: study-write's outline produces `planned` but draft produces
`executing`. Research has no draft-vs-synthesize distinction —
synthesize IS the draft. Therefore research-write skips the
`executing` rung. This is a meaningful difference and the first
gospel-shape-vs-creation-shape fork: a pipeline can produce a
non-contiguous maturity path. Document this in V.1.7.

#### V.1.3 Stage models

Insert into `stewards.stage_models`:

```sql
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model) VALUES
  ('research-write', 'gather',     'kimi-k2.6'),
  ('research-write', 'synthesize', 'kimi-k2.6'),
  ('research-write', 'review',     'qwen3.6-plus');
```

Rationale:
- `gather` and `synthesize` mirror study-write at kimi-k2.6 — known
  good for research voice on the OpenCode Go bucket.
- `review` runs cheaper (qwen3.6-plus) — same logic as the `_gate`
  stages. A verification pass doesn't need the heavier model.

#### V.1.4 Tool grants (matrix vs §V.1.4 design tension)

**The substrate enforces tools at agent-family level** (D-H4
ratifies what to do about that). Today's `research` agent grants
include: exa-search/*, fetch_url, fetch_urls, fetch_url_raw,
extract_links, web_search_exa, byu-citations/* (verify byu fits
research's brief — see open question), gospel-engine-v2/*, yt/*,
webster/*, brain_*, study_*.

That grant set is broad. Two paths forward, scored in D-H4:

- **Path A (lightweight, recommended):** use the existing `research`
  agent for all H.1 stages. Tools are uniformly available. The
  `gather` stage uses them heavily; `synthesize` could use them
  lightly; `review` should run tools-disabled (it's structured-output
  verification, same pattern as gates).
- **Path B (heavyweight):** add `pipeline_stage_tool_perms` table
  that overrides agent perms per-stage. Lets us deny tools on
  `review` even though `research` agent has them. Real
  infrastructure work.

Recommend Path A for H.1 with a tools_disabled flag on the dispatch
payload for `review` (mechanism already exists from Phase C/D).

#### V.1.5 Gate prompts — transfer matrix

Walked each of the 9 templates against the research domain.
Verdicts:

| Template | Verdict for research-write | Note |
|---|---|---|
| `evaluate` | **transfers cleanly** | Already domain-neutral ("intent and covenant for this work"). Will run against a `general-research` intent rather than `scripture-study` — that's what intent indirection buys us. |
| `generate_scenarios` | **transfers cleanly** | "Generate 3-7 testable acceptance criteria" works for "did the gather collect 5+ credible sources?" as well as for "did the draft hit the binding question?" |
| `verify` | **transfers cleanly** | Per-criterion pass/fail is domain-neutral. |
| `covenant_check` | **needs LIGHT REWRITE** | Today references `read_before_quoting`, `check_existing_work`, `surface_tensions`, `honor_scope`, `exercise_stewardship`. The first three apply to research work directly; `check_existing_work` and `exercise_stewardship` are universal. **No template rewrite needed if the active covenant carries those commitments.** Since the global covenant DOES carry them (see live data), the template works as-is. *No fork.* |
| `sabbath` | **transfers cleanly** | "Mark its ending with a structured reflection" is domain-neutral. The reflection on "what got harder than predicted" is just as meaningful for a research piece as for a study. |
| `atonement` | **transfers cleanly** | "Walk back through what was tried, what failed, what was eventually completed" — domain-neutral. The principles/decisions/lessons trichotomy is universal. |
| `council_proposer` / `_critic` / `_synthesizer` | **not invoked by research-write directly** | Councils convene on intents, not pipelines. If a research-domain council is ever convened, the same templates work. *No fork.* |

**Important finding:** the gate templates are already mostly
domain-neutral. The gospel-specificity lives in the **intent** (via
`compose_system_prompt` injection), NOT in the templates themselves.
This is good design from Phase C that pays off here. Confirms the
lightweight-with-eyes-open approach is the right call (D-H1).

The one place a template fork might be needed: if H.4 (fiction)
proves that "verify against scenarios" doesn't fit creative work,
THAT's where the fork pressure shows up — not at H.1.

#### V.1.6 Intent for research-write

Create a new intent `general-research` (slug). Initial values
suggested below; final wording to live in `intent.yaml` (D-H3
ratifies whether to grow `intent.yaml` or create a new
`research-intent.yaml`):

```yaml
slug: general-research
purpose: >
  Cast a wider net than scripture-study — gather, summarize, and
  reason about non-doctrinal sources to inform Michael's
  understanding of fields he's actively working in (AI, engineering,
  product, education).
beneficiary: Michael (primarily); secondary: anyone reading the digest
scripture_anchor: ~  # explicitly null — see D-H5
values_hierarchy:
  - key: credibility-over-volume
    description: One credible source beats five rumors. Refuse to summarize what can't be sourced.
  - key: skepticism-as-default
    description: Treat each claim as needing evidence. Note where the source is opinion vs reporting vs primary.
  - key: recency-matters
    description: A 2024 take on AI tooling is obsolete. Weight recency where the domain moves fast.
  - key: honest-uncertainty
    description: "I couldn't find a credible source on X" is a valid output. Better than fabrication.
  - key: cross-reference
    description: Where claims appear in multiple independent sources, say so. Where they appear in only one, say so.
non_goals:
  - Doctrinal claims (those go through scripture-study)
  - Personal recommendations (Michael draws his own conclusions; we summarize)
  - Source-laundering (rephrasing without attribution)
```

**Critical:** `scripture_anchor` is intentionally NULL on this
intent. Per the 2026-05-11 §VI amendment for D-F2, intents with
`scripture_anchor IS NULL AND values_hierarchy lacks
doctrinal/spiritual/discernment` count as **low-stakes**, which
means master-tier agents can bishop councils convened on this
intent. This is desirable for research work and was already designed
for via the Phase F amendment.

#### V.1.7 File destination template

```sql
UPDATE stewards.pipelines
   SET file_destination_template = 'research/<slug>.md'
 WHERE family = 'research-write';
```

The `research/` directory doesn't exist in the repo yet — it'll be
created on first materialization (G.4.2's CLI handles parent
directory creation). Per the §IV agent grants listing, `research`
agent has `yt/*` and `gospel-engine-v2/*` — note this for the
gather stage; the agent may pull from local corpus where helpful.

`file_content_jsonpath` — to revisit when synthesize's output shape
is settled. v1 use the full stage output (NULL jsonpath means whole
content); v2 narrow if the agent embeds metadata.

#### V.1.8 Sabbath + atonement defaults

| Pipeline | sabbath_enabled | atonement_enabled | Rationale |
|---|---|---|---|
| `research-write` (deep-research output_kind) | **true** | **true (opt-in)** | A deep research piece is a creative artifact. The Sabbath reflection captures what surprised the researcher; atonement on quarantine captures what made the research hard. |
| `ai-news-summary` (daily-digest output_kind, scheduled) | **false** | **false** | A daily digest is an ephemeral artifact. Sabbath every weekday morning becomes ritual-as-noise. |

This bifurcation IS the design tension surfaced in D-H6. Two paths:
(a) split into two pipelines with different sabbath flags (clean but
duplicates); (b) keep one pipeline, add `sabbath_enabled` per
work_item (not per pipeline) so the dispatch decides at work_item
creation time. Recommend (a) for H.1 because it's simpler; revisit
if duplication grows past 3 pipelines.

#### V.1.9 Acceptance scenarios

Five binding questions Michael can run through `research-write` as
the first-real-use validation:

1. **"What shipped in AI tooling this week that I should know about?"**
   gather pulls 8-12 sources from Anthropic / OpenAI / Google /
   Microsoft / vendor blogs; synthesize produces a Headlines + Notable
   + Skeptical-takes structure; review verifies each claim cites a
   gather source. End-to-end cost target: < $0.20.

2. **"What are the strongest critiques of the 'AI-replaces-engineers'
   thesis in serious publications since Jan 2026?"**
   gather identifies dissenting voices; synthesize summarizes their
   arguments steelmanned; review checks no straw-man rewriting.

3. **"What's the state of Postgres extension distribution in 2026?
   How do other extensions (pgvector, paradedb, citus) ship and
   version?"**
   gather pulls READMEs, docs, GitHub releases; synthesize produces
   a comparison matrix; review checks each cell has a citation.

4. **"What are the LDS Church's recent (2024-2026) communications
   about AI?"**
   Test of the *general-research* intent on a topic that ALSO falls
   under scripture-study's domain. Verifies the intent-shaped
   prompt context steers the work toward research-shape, not
   study-shape. (Specifically: does the agent cite official Church
   sources rather than launching a scripture exegesis?)

5. **"Compare Tauri vs Electron for shipping a desktop app in 2026."**
   A purely technical research question. Confirms research-write
   handles topics with zero gospel surface.

#### V.1.10 Programming time

- V.1.1–V.1.4 pipeline definition + stage_models + tools_disabled flag for review stage: ~30 min
- V.1.5 (gate prompts) — no code, the analysis is the work: done
- V.1.6 intent seed + YAML extension: ~30 min (assuming D-H3 path A)
- V.1.7 file_destination_template: 5 min
- V.1.8 sabbath/atonement flags: 5 min
- First real e2e run on scenario #1 (AI tooling weekly): ~30 min + monitoring
- Tuning based on first-real-use: ~1 session

**Total H.1: 1-2 sessions.** Most of the time goes to e2e tuning,
not infrastructure.

---

### V.2 H.2 — YouTube pipelines (gospel + secular)

Two pipelines sharing an ingest stage but diverging on rubric. The
shared-ingest pattern is the design test: does the substrate
gracefully handle two pipelines that share a stage definition?
Answer today is "they share the agent_family and tool grants, but
not a row" — each pipeline has its own stages array. Two rows. Fine.

**Pipeline families:**
- `yt-gospel-evaluate` (uses `yt-gospel` agent)
- `yt-secular-digest` (uses `yt` agent)

**Stages (both):** `ingest → analyze → review`

The `ingest` stage calls `yt_download(video_url)` to fetch a
transcript. **Tension surface here (D-H4 territory):** today's `yt`
agent has 14 tool grants including `playwright/*` and `vscode`
— way more than `ingest` needs. The `gather` stage of H.1 has the
same shape problem. If D-H4 ratifies the heavyweight path
(per-stage tool perms), H.2 ingest is where it pays off most.

**Gate prompts:** all 6 transfer cleanly (same conclusion as H.1's
§V.1.5).

**Intents:**
- yt-gospel-evaluate → reuses `scripture-study` intent. Gospel
  video evaluation is scripture-study-shaped work (Restoration
  discernment rubric applies).
- yt-secular-digest → reuses `general-research` intent (created in
  H.1). Secular video digest is research-shaped work.

**File destinations:**
- `yt-gospel-evaluate` → `study/yt/<slug>.md` (matches existing
  manual convention)
- `yt-secular-digest` → `study/yt/secular/<slug>.md` OR
  `research/yt/<slug>.md` — D-H3 has a sub-decision on directory
  scheme here. Recommend `study/yt/<slug>.md` for both initially
  (existing convention), revisit if secular volume grows.

**Sabbath/atonement:**
- yt-gospel-evaluate → sabbath=true (study-shaped), atonement=true (gospel video evaluations have failure modes worth recording)
- yt-secular-digest → sabbath=true, atonement=false (default off)

**Acceptance scenarios:**
- yt-gospel-evaluate on a 2024 GC talk (transcript known good).
  Output evaluates against Restoration framework.
- yt-secular-digest on a Nate B. Jones AI video. Output captures key
  claims + skeptical questions.

**Programming time: 1 session.** Pipeline rows + ingest stage call
+ file destination + first run each.

---

### V.3 H.3 — Scheduled pipelines + ai-news-summary

This is the sub-phase with the most genuinely new infrastructure.
The watchman_scheduler (extension/2-7b2) is **not** a generic cron;
it's a decision function tied to one workflow. H.3 needs a separate
machinery.

#### V.3.1 Schema

```sql
CREATE TABLE stewards.scheduled_pipelines (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text UNIQUE NOT NULL,
    pipeline_family     text NOT NULL REFERENCES stewards.pipelines(family),
    intent_id           uuid NOT NULL REFERENCES stewards.intents(id),
    cron_pattern        text NOT NULL,           -- 'MM HH * * DAY' subset v1
    input_template      jsonb NOT NULL,          -- merged into work_item.input
    enabled             boolean NOT NULL DEFAULT true,
    last_dispatched_at  timestamptz,
    next_due_at         timestamptz,             -- recomputed by tick after each dispatch
    min_interval_secs   int NOT NULL DEFAULT 3600,  -- D-PE3 hard floor (1h)
    file_destination_override text,              -- optional per-schedule destination
    notes               text,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX scheduled_pipelines_due_idx
    ON stewards.scheduled_pipelines (next_due_at)
    WHERE enabled = true;
```

#### V.3.2 Bgworker hook

New file `extension/Nx-scheduled-pipelines.sql` with a
`scheduled_pipelines_tick()` SQL function. Bgworker calls it every
60s on the leader (mirrors the watchman pattern).

```sql
CREATE OR REPLACE FUNCTION stewards.scheduled_pipelines_tick()
RETURNS jsonb
LANGUAGE plpgsql AS $$
DECLARE
    v_row RECORD;
    v_work_id uuid;
    v_dispatched int := 0;
BEGIN
    FOR v_row IN
        SELECT * FROM stewards.scheduled_pipelines
         WHERE enabled = true
           AND (next_due_at IS NULL OR next_due_at <= now())
           AND (last_dispatched_at IS NULL
                OR now() - last_dispatched_at >= make_interval(secs => min_interval_secs))
         ORDER BY next_due_at NULLS FIRST
         LIMIT 5  -- per-tick cap
    LOOP
        BEGIN
            v_work_id := stewards.work_item_create(
                p_pipeline_family => v_row.pipeline_family,
                p_input           => v_row.input_template,
                p_intent_id       => v_row.intent_id,
                p_slug            => v_row.slug || '-' || to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD')
            );
            PERFORM stewards.work_item_dispatch_stage(v_work_id);
            UPDATE stewards.scheduled_pipelines
               SET last_dispatched_at = now(),
                   next_due_at = stewards.compute_next_due(cron_pattern, now())
             WHERE id = v_row.id;
            v_dispatched := v_dispatched + 1;
        EXCEPTION WHEN OTHERS THEN
            -- log to a new stewards.scheduled_pipelines_errors table;
            -- don't propagate (same isolation pattern as steward_tick).
            INSERT INTO stewards.scheduled_pipelines_errors (sched_id, error_at, error_msg)
                 VALUES (v_row.id, now(), SQLERRM);
        END;
    END LOOP;
    RETURN jsonb_build_object('dispatched', v_dispatched);
END;
$$;
```

The cron parser `compute_next_due(pattern, base_ts)` lives in
extension/src/cron.rs as a small Rust helper (pgrx-exported).
Supports only `MM HH * * DAY` v1.

#### V.3.3 Seed: ai-news-summary

```sql
INSERT INTO stewards.scheduled_pipelines
    (slug, pipeline_family, intent_id, cron_pattern, input_template, notes)
VALUES (
    'ai-news-7am-mt-weekdays',
    'research-write',  -- reuse H.1's pipeline
    (SELECT id FROM stewards.intents WHERE slug='general-research'),
    '0 13 * * 1-5',    -- 1300 UTC = 0700 MT (during MST; MDT is 0600); confirm
    '{"binding_question":"What shipped in AI tooling, models, or research overnight that I should know about?",
      "sources_spec":{"queries":["AI news today","claude release","openai release","anthropic announcement","google deepmind"],
                      "since":"24h"},
      "output_kind":"daily-digest"}'::jsonb,
    'Weekdays 7am MT. Daily AI digest. Lives in research/ai-news/. Does NOT trigger sabbath (overridden per V.1.8).'
);
```

Note: this scheduled run uses `research-write` as the pipeline. The
`output_kind: "daily-digest"` flag in input_template would signal
the pipeline to skip sabbath/atonement (D-H6 ratifies how).

#### V.3.4 Tension surface — scheduled dispatches and human-presence assumptions

This is the deepest gospel-shape vs creation-shape fork in Batch H.
Several substrate assumptions implicitly bake in a human convener:

1. **"surface" gate action** — when a gate decides `surface`, the
   work_item routes to `your_turn` so a human steers. With
   scheduled dispatches, "your_turn" piles up overnight and the
   human sees a backlog in the morning. Acceptable, but new.
2. **Cost cap quarantine** — a scheduled run that hits cost cap at
   3am has no human to talk to until 7am. Atonement fires
   automatically (post-Batch G) so the lesson gets recorded; that's
   actually a clean answer. Good design.
3. **Sabbath on every daily run** — already addressed in V.1.8
   (sabbath disabled for daily-digest output_kind).
4. **Council convening** — councils require a bishop. Scheduled
   work that produces a watchman-suggested council can't auto-
   convene. Fine; this stays manual.

The deepest assumption: **the substrate has no concept of "this
work_item was machine-originated" vs "this work_item was human-
originated."** Today the only signal is `work_items.created_by`
which is text. Recommend: add `work_items.origin` column with values
`human` | `scheduled` | `watchman` | `steward` | `council`. Let gate
prompts and atonement prompts vary based on origin. Doesn't have
to ship in H.3 — but **propose** in §VI D-H7.

#### V.3.5 UI

New `/scheduled` route:
- List view: slug, pipeline_family, cron_pattern, next_due_at,
  enabled toggle, last 3 dispatches with links.
- Modal: edit cron / disable / re-test (computes next_due_at).
- Dashboard card: "Next scheduled in N hours" + "Last 7 scheduled
  runs" status summary.

#### V.3.6 Programming time

- Schema + cron parser (Rust) + tick function: 1 session
- Bgworker hook (mirrors watchman pattern): ~30 min
- UI surfaces: 1 session
- First scheduled run overnight + tune: 30 min next morning

**Total H.3: 2-3 sessions.**

---

### V.4 H.4 — Fiction (`fiction-scene` minimal variant)

This sub-phase is **the maturity ladder fit test.** Expect to learn
that the ladder doesn't fit and propose the per-family-ladder fix as
a follow-up batch.

#### V.4.1 Why fiction is shaped differently

Creative work doesn't move through raw → researched → planned →
specced → executing → verified. It moves through something more like
premise → scene → draft → revision → polish. The motion is
recursive, not linear. A scene draft might force a premise rewrite.
Verify-against-scenarios doesn't fit ("the scene should make the
reader cry" is not a checkable acceptance criterion the way "the
study cites 3+ primary sources" is).

#### V.4.2 v1 design: collapsed ladder

Build the minimum pipeline:

**Pipeline family:** `fiction-scene`

**Stages:** `premise → draft → polish`

**Stage models:** all kimi-k2.6 (the model handles creative voice
reasonably well; alternative: glm-5.1 for synthesis if voice quality
matters more than cost). Tune later.

**Maturity mapping (the experiment):**
- premise → researched (force-fit; "the idea is sketched")
- draft → executing (force-fit; "the scene exists")
- polish → verified (force-fit; "the scene is presentable")

The force-fits ARE the experiment. Each one is somewhat dishonest
about what the rung means in creative work. Document this in
retrospective.

#### V.4.3 Tool grants

`fiction` agent already exists in the workspace. Live data: 14
allow grants. Most are tools fiction shouldn't be calling during
scene-writing (vscode, edit, web, search). The pipeline should
dispatch with tools_disabled=true for ALL three stages. Creative
writing doesn't research mid-scene; if it needs research, the human
runs a separate research-write pipeline and feeds results in via
input_template.

#### V.4.4 Gate prompts — fork pressure

This is where the templates strain:

| Template | Verdict for fiction-scene | Note |
|---|---|---|
| `evaluate` | **needs FORK** | "Does the output advance the stated intent" works; "covenant's surface_tensions" doesn't (a scene shouldn't surface counterarguments — it should immerse). Need a `evaluate_creative` variant. |
| `generate_scenarios` | **needs FORK or skip** | "3-7 testable acceptance criteria" doesn't fit creative work. Either skip scenarios entirely for fiction or rewrite as "narrative beats" (does the scene hit the emotional turn? does the dialogue stay in character? does the pacing land?). |
| `verify` | **needs FORK or skip** | Per-beat verification might work IF generate_scenarios produces beats rather than criteria. |
| `covenant_check` | **transfers loosely** | The agent commitments around honest paraphrase + scope still apply to fiction. The gospel-specific items (read_before_quoting, gospel sources) don't bite for original fiction. |
| `sabbath` | **transfers cleanly** | "What surprised you, what got harder than predicted" applies beautifully to creative work. |
| `atonement` | **transfers cleanly** | Failed creative attempts produce real lessons (this premise didn't work because X). |

**Recommendation for H.4 v1:** ship with the existing templates
unchanged but flag the forks for the retrospective. Don't pre-build
the fiction-specific templates until we see how the existing ones
fail. The retrospective IS H.4's deliverable as much as the
pipeline itself.

#### V.4.5 Intent

Create `creative-fidelity` intent:
- purpose: fiction work that stays true to the world, the
  characters, and the emotional truth being explored
- scripture_anchor: NULL (low-stakes per F2 amendment; bishops can
  be agent for fiction councils)
- values_hierarchy: character-consistency, world-coherence,
  emotional-truth, voice-fidelity
- non_goals: doctrinal teaching disguised as fiction; lecture in
  dialogue form

#### V.4.6 File destination

`fiction/<world>/<slug>.md` — note the `<world>` parameterization.
Bridge Sim work goes under `fiction/bridge-sim/`; D&D under
`fiction/dnd/`; standalone short stories under `fiction/standalone/`.
The substrate's `<slug>` template substitution supports this if
input_template carries `world`.

#### V.4.7 Sabbath / atonement

Both **true**. Creative work benefits from ending-records as much as
study work, maybe more — fiction sessions don't have natural
verification gates.

#### V.4.8 Acceptance scenario (single)

Run one scene through the pipeline. A Bridge Sim NPC scene where
the captain receives bad news from a junior officer. Premise stage
sketches stakes + stakes; draft writes the scene; polish tightens
dialogue + cuts. Verify by reading. Note every place the maturity
ladder felt forced.

#### V.4.9 Programming time

- Pipeline definition + stage_models + intent: ~30 min
- Tools_disabled flag application across all 3 stages: ~10 min
- First run + retrospective notes: 1 session
- Retrospective document for follow-up batch (the per-family-ladder
  proposal): 1 session

**Total H.4: 1-2 sessions** including the retrospective.

---

### V.5 H.5 — Bridge Sim NPC live dialogue (**RECOMMEND DEFER**)

Michael flagged the tension in the request: "this might NOT want a
pipeline at all — it wants live interactive dialogue."

Argument for keeping H.5 in Batch H:
- Bridge Sim NPCs use the same world + characters + voice as fiction
  pipelines (H.4). Keeping them in one batch surfaces shared
  primitives early.
- The substrate's intent + covenant injection (Phase C) is exactly
  the right shape for "this NPC has THIS personality and knows
  THESE things about the world."

Argument for deferring H.5:
- A `work_item` is **task-shaped**: there is an input, stages
  transform it, an output is produced, the work item closes. Live
  NPC dialogue is **session-shaped**: a conversation has many turns,
  no single completion, no maturity ladder, no acceptance scenarios.
- Forcing live dialogue into the work_item shape means creating a
  work_item per turn, or per conversation, or per session. Each
  choice has the wrong granularity. Per-turn pollutes the work
  queue. Per-conversation has no natural close (when does a
  Bridge Sim session "end"?). Per-session conflates many
  conversations.
- The substrate's six rituals — gates, scenarios, verify, sabbath,
  atonement, councils — don't fire usefully on dialogue. Sabbath
  after every NPC turn is absurd. Atonement on a bad NPC line is
  over-engineered.
- Michael's note explicitly named "substrate-aware chat" as a
  future direction. THAT is the right shape for live NPC dialogue:
  a chat session that reads substrate state (the world, the
  characters, the recent events) but doesn't *itself* go through
  the work_item lifecycle.

**Recommendation:** **Defer H.5 to a future "substrate-aware chat
for fiction" workstream.** That workstream should answer:

- Does the substrate-aware chat (already a separately-proposed
  feature; see `stewards-ui-evolution.md`) gain a "character"
  context-mode where an intent + a character persona + a world
  history get loaded into the system prompt?
- Do Bridge Sim NPC interactions get archived as `sessions` rows
  (which the substrate already supports) rather than work_items?
- What's the analog to sabbath for an NPC session? (Possibly: after
  N turns or M minutes of inactivity, write a session-reflection
  lesson — that's a much lighter ritual than a work_item sabbath.)

**Caveat for the deferral.** If the science-center deadline forces
*some* substrate involvement before substrate-aware-chat ships,
ship a minimal **H.5-stub:** a `fiction-scene` pipeline run that
generates a *script* (not live dialogue) the human reads to the
center visitors. Same as H.4. That's a fallback, not the real
answer.

**§VI D-H5 ratifies this defer/include decision.**

---

## VI. Decisions Michael needs to ratify (D-H1 through D-H6)

Following the §VI convention from `full-agentic-substrate.md`. Each
decision lists options + a recommendation; Michael ratifies before
H.1 coding starts.

---

**D-H1: Lightweight vs heavyweight generalization.**

Two paths for adding new pipeline families:
- **Path A (lightweight, recommended):** Copy study-write's shape,
  swap stages + tools + intent, ship. Note every place a primitive
  forks. Apply rule of three: after three forks, pause and audit
  for shared template extractions.
- **Path B (heavyweight):** Before any new pipeline, audit every
  shared template / function for gospel-specific assumptions.
  Extract a domain-agnostic substrate-core; pipelines become thin
  configuration over it.

**Recommendation: Path A.** §V.1.5's gate-prompt transfer matrix
already shows the templates are mostly domain-neutral; the
gospel-specificity lives in intents. The pre-audit cost of Path B
exceeds its expected benefit. Rule of three serves as the
escape hatch if forking explodes.

**RATIFIED 2026-05-11: A (lightweight + rule of three).** Per recommendation.

---

**D-H2: Maturity ladder fit.**

Study-write fits raw→researched→planned→specced→executing→verified.
Research probably fits (with a non-contiguous skip from `planned` to
`verified` for review-stage pipelines — see V.1.2). Fiction
probably doesn't fit. Options:
- **A:** Force-fit fiction (and any other non-fit) onto the existing
  ladder; document the strain; revisit in a follow-up batch if
  strain compounds.
- **B:** Per-pipeline-family ladder column. `pipelines.maturity_ladder`
  is a jsonb array of rung names. study-write declares
  `[researched, planned, executing, verified]`; fiction-scene
  declares `[premise, draft, polish]`. Each pipeline brings its own
  ladder; gates run against the pipeline-declared ladder.
- **C:** Drop the maturity concept for pipelines that don't fit;
  fiction work flows stage-to-stage without intermediate maturity
  rungs.

**Recommendation: A for Batch H, propose B in H.4 retrospective for
the next batch.** Don't try to design the per-family-ladder system
without first feeling where the strain is. Force-fit H.4 in v1;
let the retrospective produce the real proposal.

**RATIFIED 2026-05-11: B (per-family ladder column NOW).** Heavier than
recommended. H.1 adds `pipelines.maturity_ladder jsonb NOT NULL DEFAULT
'["raw","researched","planned","specced","executing","verified"]'::jsonb`;
seeds study-write with the existing six-rung ladder; seeds research-write
with its appropriate ladder (likely the same six, possibly compressed —
H.1 decides); refactors any code that hardcodes the ladder (e.g., the
maturity transition gates) to read from the column. Rationale: the
retrofit cost is real once Phase B/D/E logic is built against a hardcoded
ladder; building the column-driven version once is cheaper than refactoring
later.

---

**D-H3: Intent granularity.**

Today there's one intent (scripture-study). New pipelines surface
the granularity question:
- **A (minimal):** Two new intents total: `general-research` (covers
  H.1, H.2 yt-secular, H.3 ai-news, H.5 stub) and
  `creative-fidelity` (covers H.4, future fiction). yt-gospel
  reuses scripture-study.
- **B (medium):** Add a third intent `professional-awareness`
  specifically for AI-news / industry-tracking work; keep
  `general-research` for one-off "really dig into X" research.
- **C (fine-grained):** Per-pipeline-family intent (research,
  yt-secular-digest, ai-news-summary, fiction-scene each own one).

Storage: where does the intent YAML live?
- **A:** Grow `intent.yaml` to a multi-document YAML file (root +
  multiple intent docs).
- **B:** New file `.spec/intents/<slug>.yaml` per intent. Update the
  pre-commit hook to glob.

**Recommendation: intent count = A (minimal), storage = B (one file
per intent in `.spec/intents/`).** Minimal intents avoid premature
fragmentation; one-file-per-intent prevents the root YAML from
becoming unwieldy. Pre-commit hook grows the loop.

**RATIFIED 2026-05-11: intent count = B (three new intents:
general-research + professional-awareness + creative-fidelity),
storage = B (one file per intent in `.spec/intents/<slug>.yaml`).**
Heavier on count than recommended. H.1 creates `general-research`
(covers H.1 research-write + H.2 yt-secular); H.3 creates
`professional-awareness` (covers H.3 ai-news-summary + similar industry
tracking); H.4 creates `creative-fidelity` (covers H.4 fiction-scene +
future creative work). yt-gospel reuses `scripture-study`. Pre-commit
hook grows to glob `.spec/intents/*.yaml`.

---

**D-H4: Per-pipeline-per-stage tool grants.**

Today tools are agent-level. H.1's `review` stage and H.2's
`ingest` stages both want tighter restrictions than the agent grant
allows. Options:
- **A (minimum):** Use `tools_disabled=true` payload flag for
  stages that should run without tools. Existing mechanism.
- **B (medium):** New `stewards.pipeline_stage_tool_perms` table
  that overrides agent grants per (pipeline_family, stage_name,
  tool_pattern). Real infrastructure.
- **C (heavyweight):** Move ALL tool grant authority to
  pipeline_stage level; deprecate agent_tool_perms. Big refactor.

**Recommendation: A for H.1/H.2/H.3, propose B in H.2 retrospective
if `tools_disabled=true` proves too coarse.** Most stages either
want all the agent's tools or none. The binary suffices until a
real case demands per-tool grants.

**RATIFIED 2026-05-11: A (tools_disabled=true payload flag).** Per
recommendation. Revisit in H.2 retrospective if binary proves too coarse.

---

**D-H5: Sabbath + atonement applicability per family.**

Today sabbath_enabled / atonement_enabled are boolean columns on
`pipelines`. Three cases proposed in Batch H:
- research-write (deep): sabbath true, atonement true (opt-in)
- ai-news-summary (scheduled daily-digest): sabbath false, atonement false
- yt-gospel-evaluate: sabbath true, atonement true
- yt-secular-digest: sabbath true, atonement false
- fiction-scene: sabbath true, atonement true

Options:
- **A:** Encode at pipeline level (today's design). One pipeline per
  sabbath/atonement combination — duplicate pipelines if needed.
- **B:** Add `sabbath_enabled` per work_item override (column on
  `work_items`). Default from pipeline; per-work_item override
  available. Lets one `research-write` pipeline serve both
  deep-research (sabbath on) and daily-digest (sabbath off) by
  varying the work_item flag.
- **C:** Encode at intent level. `intents.sabbath_enabled`
  cascades to work_items via the intent_id FK.

**Recommendation: A for H.1, propose B in retrospective.** If the
ai-news-summary pipeline ends up as a separate pipeline family from
research-write (V.3.3 wires it that way), Option A is sufficient.
Cross-pipeline duplication only becomes painful at 3+ pipelines, by
which point we have evidence for the right Option B design.

**RATIFIED 2026-05-11: B (per-work_item override).** Heavier than
recommended. H.1 adds `work_items.sabbath_enabled boolean NULL` +
`work_items.atonement_enabled boolean NULL` (NULL = inherit from
pipeline default). Sabbath dispatch + atonement enqueue logic resolves
work_item override first, falls back to pipeline. Lets a single
research-write pipeline serve both deep-research (sabbath on) and
daily-digest (sabbath off) without pipeline duplication. Rationale:
ai-news being a separate pipeline solves the daily-digest case but
not the general "this specific work_item doesn't warrant sabbath"
case — having the per-item knob upfront is cheaper than the column
addition + trigger refactor later.

---

**D-H6: H.5 — include in Batch H or defer to substrate-aware-chat?**

Detailed analysis in §V.5. Three positions:
- **A (recommended): Defer.** Bridge Sim NPCs are session-shaped,
  not work_item-shaped. They belong in the substrate-aware-chat
  workstream where session-shaped interactions live naturally.
  Document the deferral; revisit when substrate-aware-chat is
  designed.
- **B: Include as fiction-scene stub.** Ship H.4's fiction-scene
  pipeline. Bridge Sim NPCs use it to generate pre-written scripts.
  Not live dialogue; a fallback that lets the science-center
  deadline land without waiting on substrate-aware-chat.
- **C: Force-include as live dialogue.** Add a `live-dialogue`
  pipeline shape that's fundamentally different (per-turn dispatch,
  no maturity ladder, ephemeral). Major substrate work; almost
  certainly the wrong place to put it.

**Recommendation: A.** B is a reasonable fallback if science-center
timing demands something now. C should be rejected outright.

**RATIFIED 2026-05-11: A (defer to substrate-aware-chat).** Per
recommendation. Bridge Sim NPCs are session-shaped, not work_item-shaped.
Revisit when substrate-aware-chat workstream is designed. Fallback if
science-center deadline forces something: H.4 fiction-scene pipeline
can produce pre-written scripts the human reads to visitors.

---

**D-H7 (NEW, surfaced in V.3.4): Origin tracking on work_items.**

Add `work_items.origin` column with enum values `human` | `scheduled`
| `watchman` | `steward` | `council`? Lets gates and atonement vary
behavior based on whether a human convened the work or a schedule
fired it.

**Recommendation: Yes, ship in H.3.** Cost is one column +
backfill. Value is real: scheduled-run atonement might want to
mention "this fired at 3am, no human was available to steer" in
the lesson context. Light enough to ship now, valuable enough
that retrofitting would be annoying.

**RATIFIED 2026-05-11: A (ship origin column in H.3).** Per
recommendation. `work_items.origin text` enum (human | scheduled |
watchman | steward | council). Backfill existing rows to 'human'
(the only origin that's existed). NOT NULL after backfill.

---

## VII. Carry-forward — what Batch H unlocks for the three larger proposals

Per the request: how does Batch H's completion compose with the
three Michael-flagged workstreams?

### UI authoring for intents + covenants (`stewards-ui-evolution.md`)
- Batch H **multiplies** the intent count from 1 to 3-4 (D-H3 A).
  Single-intent UI was acceptable; multi-intent UI is mandatory.
- The `.spec/intents/<slug>.yaml` per-file pattern (D-H3 storage B)
  is the right shape for a UI authoring tool that creates one file
  per intent.
- The intent-edit UI needs to know which pipelines reference each
  intent. Both `stewards.scheduled_pipelines.intent_id` and
  `stewards.work_items.intent_id` FKs make this queryable.

### Substrate-aware chat (`stewards-ui-evolution.md` chat surface)
- Batch H **defers** H.5 (Bridge Sim NPCs) into this workstream.
  That's a significant feature requirement for substrate-aware chat:
  it needs to support character-personality context-modes, not just
  read-only substrate inspection.
- H.3's `work_items.origin` column (D-H7) creates the precedent for
  marking non-work-item interactions. Sessions originating from
  substrate-aware chat will need similar typing.
- The retrospective from H.4 about maturity-ladder strain will
  shape what substrate-aware chat does NOT need to enforce.

### Multi-pipeline ecosystem (`substrate-pipelines-expansion.md`)
- This proposal **supersedes** the older one. After Batch H ships,
  retire `substrate-pipelines-expansion.md` (move to an `archive/`
  folder under `.spec/proposals/`) so the canonical pipeline
  expansion proposal is this one.
- Future pipelines (weekly steward review, monthly studies audit,
  whatever) become 1-session adds rather than 1-week proposals
  because the patterns are now in place.

---

## VIII. Estimated programming time (rolled up)

| Sub-phase | Sessions |
|---|---|
| H.1 research-write | 1-2 |
| H.2 yt-gospel + yt-secular | 1 |
| H.3 scheduled pipelines + ai-news-summary | 2-3 |
| H.4 fiction-scene + retrospective | 1-2 |
| H.5 (deferred per recommendation) | 0 (defer) or 0 (B fallback ships with H.4) |
| **Batch H total** | **5-8 sessions** |

Comparable to Phase D (2 sessions) + Phase E (2 sessions) +
Phase F (3-4 sessions). Larger than any single A-F phase because
Batch H is breadth (5 new pipelines) rather than depth (one new
architectural primitive).

---

## IX. Risks + blind spots

**1. Cost discipline on scheduled runs.** ai-news-summary fires
every weekday. At $0.20 per run × 5 days × 4 weeks = $4/month for
that one schedule. Acceptable. But: the substrate has no per-
schedule cost dashboard yet. Need to either reuse the per-work-item
cost panel or add a per-schedule cost view. Surface this as
follow-up if H.3 ships before the cost panel is extended.

**2. Tools eagerness on research pieces.** A research piece
*legitimately* needs many tool calls. Today's cost-cap mechanism +
the steward's retry logic might fire too eagerly on research work
that's just doing its job (calling exa-search 5 times, fetch_url 10
times, byu-citations 3 times). Worth measuring on H.1 acceptance
scenario #1 and tuning cost cap defaults per pipeline if needed.

**3. Council fit on research.** Phase F's council is designed for
intent-level deliberation. If a research-write run hits cost cap
and atonement extracts "the question was malformed; we need to
deliberate before re-running," does a council make sense for a
research question? Probably yes, but the council templates assume
some doctrinal-stake framing. Worth a test convening on a research
question during H.1 e2e.

**4. ai-news-summary trustworthiness.** The substrate has no LLM
hallucination check for news claims. The `review` stage helps
catch "this fact doesn't appear in the gather sources," but
doesn't catch "this gather source itself is rumor." H.1 needs
explicit guidance in the agent prompt + the synthesize template to
distinguish primary reporting from secondary reporting from rumor.
Otherwise the daily digest becomes high-confidence noise.

**5. Fiction agent isn't well-tuned yet.** The `fiction` agent
exists but the voice work for creative writing is less mature than
for studies. H.4's force-fit experiment may actually surface "the
agent isn't ready" before it surfaces "the maturity ladder doesn't
fit." Worth distinguishing in the retrospective.

**6. Phase F council over-engineering for research.** A research
piece that legitimately produces 5 sources, synthesizes them,
reviews them — has no doctrinal stake, no spirit-discernment load.
Phase F councils are designed for doctrinal stake. Most research
work should never convene a council. Verify this stays the default
behavior; document if research patterns suggest auto-convening.

---

## X. Closing

The substrate has spent six phases learning to walk the gospel
cycle one way. Batch H asks whether it can walk other paths.

The honest expectation: H.1 (research) and H.2 (YouTube) will fit
the existing primitives with minor forks. H.3 (scheduling) will
require genuine new infrastructure but the substrate's primitives
will continue to serve. H.4 (fiction) will expose the maturity
ladder as gospel-shaped, not creation-shaped, and produce the
proposal for the next batch's work. H.5 (live NPCs) doesn't belong
in Batch H — it belongs in substrate-aware-chat.

What matters most: each fork is *named*, not hidden. The
retrospective from each sub-phase is as important as the pipeline
itself. The substrate's value isn't that it works on five
pipelines — it's that we know exactly where its shape ends and
where new shapes need to begin.

Michael: ratify §VI D-H1 through D-H7 (or surface alternatives)
before H.1 coding starts. Recommend doing this in one
AskUserQuestion batch the same way Phases C–F were ratified.
