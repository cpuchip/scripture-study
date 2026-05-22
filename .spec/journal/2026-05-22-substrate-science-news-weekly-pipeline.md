---
date: 2026-05-22
mode: substrate (PE-B seed)
workstream: WS5
project: pg-ai-stewards
title: "Substrate science-news-weekly scheduled pipeline seeded; first marsfield.org Research-side automation closes the loop"
status: shipped + verified. Row visible via /api/scheduled/list and direct SQL; next_due_at correctly populated for 2026-05-25 13:00 UTC by the BEFORE INSERT trigger. Migration file at `projects/pg-ai-stewards/extension/pe8-seed-science-news-weekly.sql`.
carry_forward:
  - "First natural fire on 2026-05-25 13:00 UTC (Monday 7am MT). Output will materialize to `research/science-news-weekly--2026-05-25-1300.md` per research-write's file_destination_template. After review, republish to marsfield.org /science-news following the 2026-05-22 workflow (frontmatter + h3→h2 promotion + honesty preamble)."
  - "**Cost uncertainty.** research-write is 3 stages with tools-on gather (kimi-k2.6) + synthesize (kimi-k2.6) + review (qwen3.6-plus). The 2026-05-11 manual physics-news dispatch cost is in `220cf35`'s context; for a recurring weekly cadence the monthly cost lands around 4×(that). Worth watching in cost_buckets and adjusting cadence if it surprises."
  - "**Manual test-fire optional.** To verify end-to-end before Monday, `UPDATE stewards.scheduled_pipelines SET next_due_at = now() WHERE slug = 'science-news-weekly';` will trigger on the next watchman tick (60s). Costs ~$0.30-1.00 in tokens. Not done by default — let the natural Monday fire validate."
  - "**Auto-republish bridge is a future thing.** Right now the loop is substrate → research/ → manual review → manual republish to marsfield.org. A future automation could watch research/science-news-weekly--*.md and PR/commit to marsfield.org/content/science-news/. Worth considering after a few natural fires shake out the prompt quality."
  - "**Aggregation work item open:** the substrate has two scheduled pipelines now (ai-news-7am daily, science-news-weekly Monday). When PE-C's /scheduled UI was built, it was sized for the single ai-news row. Should still render fine with two, but verify after first natural fire."
links:
  - "../../projects/pg-ai-stewards/extension/pe8-seed-science-news-weekly.sql"
  - "../../projects/pg-ai-stewards/extension/pe7-scheduled-pipelines-fire-and-tick.sql  (companion ai-news-7am seed)"
  - "../../projects/pg-ai-stewards/extension/h1-2-research-write-pipeline.sql  (the pipeline this dispatches)"
  - "../../research/physics-news-20260503-science-center-roundup.md  (format anchor)"
  - "../../projects/marsfield.org/content/science-news/2026-05-22-physics-roundup-may-2025.md  (the republish target)"
---

# 2026-05-22 — Substrate science-news-weekly seeded

Closing the loop the marsfield.org Science News category opened earlier
in the same session. The first post there was manually republished from
`research/physics-news-20260503-science-center-roundup.md` — itself a
one-off `research-write` dispatch from 2026-05-11 (commit `220cf35`). The
question was: can we automate the upstream so the public face gets fed
on a regular cadence?

Now there is a `science-news-weekly` row in `stewards.scheduled_pipelines`,
firing Mondays 13:00 UTC, using the `research-write` pipeline_family —
the same family that produced the format we love.

## Decisions surfaced + ratified

Three decisions via `AskUserQuestion` before any SQL hit disk (per
substrate CLAUDE.md §3 cadence — *decisions upfront*):

1. **Cadence**: Weekly Monday 7am MT (`0 13 * * 1`). Physics moves
   slower than AI; daily would force the substrate to pad. Daily costs
   would also be ~5× higher with the cadence offering little marginal
   value.
2. **Pipeline family**: `research-write`. The physics-news roundup used
   this (not `research-summary`, which ai-news uses). research-write is
   three stages with tools-on gather and a structured
   Headlines/Notable/Skeptical/Open-Questions synthesize prompt. Format
   matches because the engine matches.
3. **Scope**: Physics + general science (astronomy, chemistry, biology,
   materials, climate, neuroscience). Broader net than the existing
   "Physics News" first post. The science-center mission isn't
   discipline-bounded; the binding question shouldn't be either.

## What shipped

One new SQL file:
`projects/pg-ai-stewards/extension/pe8-seed-science-news-weekly.sql` —
105 lines, idempotent INSERT…ON CONFLICT DO UPDATE on
`stewards.scheduled_pipelines`. Live-applied via the substrate's
`docker cp + psql -f` pattern (no rebuild needed; not added to
`extension_sql_file!` chain since it's an operational seed, not
schema).

The seed row:

| field | value |
|---|---|
| slug | `science-news-weekly` |
| pipeline_family | `research-write` |
| intent_id | `general-research` (same as ai-news) |
| cron_pattern | `0 13 * * 1` (Monday 13:00 UTC = 7am MT) |
| enabled | true |
| missed_window_hours | 72 (catch-up window Mon→Thu, then skip + advance) |
| input_template.binding_question | physics + general-science roundup with science-center-translation requirement |
| next_due_at (auto-trigger) | **2026-05-25 13:00:00 UTC** |

The binding_question is the load-bearing prompt. It:
- States the binding question directly
- Names the seven sub-disciplines we want covered
- Requires "Science-center translation" for every headline finding
- Targets sub-$500 exhibits where possible
- Demands Headlines / Notable / Skeptical Takes / Open Questions / Synthesis structure
- Closes with tone guidance: "concrete, direct, unadorned … vague enthusiasm wastes their time"

## Verification

```
$ docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c \
    "SELECT slug, pipeline_family, cron_pattern, enabled, next_due_at AT TIME ZONE 'UTC' AS next_due_utc, missed_window_hours
     FROM stewards.scheduled_pipelines ORDER BY slug;"

        slug         | pipeline_family  | cron_pattern | enabled |    next_due_utc     | missed_window_hours
---------------------+------------------+--------------+---------+---------------------+---------------------
 ai-news-7am         | research-summary | 0 13 * * 1-5 | t       | 2026-05-25 13:00:00 |                  24
 science-news-weekly | research-write   | 0 13 * * 1   | t       | 2026-05-25 13:00:00 |                  72
```

Also verified via `/api/scheduled/list` — both rows returned with full
input_template visible.

## Path mangling note

`docker cp /c/...` followed by `:/tmp/pe8.sql` got mangled by Git Bash's
MSYS path translation into `C:\c:` on the destination side. The cleaner
fix was to drop into PowerShell which respects Windows paths natively.
Worth a one-line addition to substrate CLAUDE.md §4 — *on Windows/Git
Bash use PowerShell for `docker cp` to avoid MSYS path mangling* —
saved for the next substrate session that needs the same workaround.

## End-to-end automation arc (the bigger thing this session shipped)

Reading today's three journals in sequence shows the full arc:

1. `2026-05-22-marsfield-org-scaffold.md` — public face exists
2. `2026-05-22-marsfield-science-news-category.md` — first Science
   News post manually republished from substrate output; flagged the
   gap that there was no automated pipeline
3. *(this)* — substrate now auto-produces what we manually republished

Three sessions, one calendar day, ~16 commits across two repos, no
rollbacks. The substrate's *agentic creation cycle* claim — that the
substrate produces durable artifacts the public sees — has its
narrowest possible first instance closed: research substrate →
verified human review → marsfield.org. Now the wheel can turn weekly
without a human kicking it.

## Lessons

- **The format question routed me to the right pipeline.** First
  instinct was to mirror ai-news-7am exactly (research-summary +
  sources_spec). Tracing the format Michael loved back to its origin
  showed it was research-write, not research-summary. Two pipelines
  with similar names produce structurally different outputs. The
  substrate's `pipelines` table is the source of truth here; the seed
  is the easy part once you know which row to point at.
- **AskUserQuestion before SQL is the right substrate discipline.**
  Three decisions surfaced upfront cost ~30 seconds of human time and
  prevented a likely rewrite if I'd guessed wrong on any of cadence,
  pipeline family, or scope. The substrate's C-F build cadence
  (decisions → smoke → commit → memory) is load-bearing for exactly
  this reason.
- **`missed_window_hours` deserves attention per pipeline.** ai-news at
  24h is right because daily; weekly at 72h is right because the
  catch-up window doesn't overlap next week's run. Picking the right
  value requires actually thinking about the cadence interactions,
  not just copying ai-news.

## What's next

Natural Monday 2026-05-25 fire is the validation. If the output looks
like the May 3 anchor (Headlines / Notable / Skeptical / Open
Questions / Synthesis with exhibit translations), the prompt is
right. If it's thin or off-format, refine the binding_question and
re-ON-CONFLICT-DO-UPDATE the row.
