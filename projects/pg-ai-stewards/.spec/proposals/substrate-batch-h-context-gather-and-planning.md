---
title: Substrate Batch H — context-gather + planning (H.1.7, H.2, H.3)
date: 2026-05-11
status: H.1.7 build-ready (ratified this session); H.2 build-ready; H.3 design-with-open-questions
parent: substrate-batch-h-pipeline-expansion.md
purpose: >
  Batch H.1 shipped research-write as the first non-scripture
  pipeline. This proposal extends Batch H with three sub-batches
  that turn the substrate from "agent reads external sources" into
  "agent reads + reflects on what we've already built." H.1.7
  grants the research agent read access to the substrate's own
  state (DB + journal/proposal/mind/docs trees). H.2 adds a
  first-class context-gather stage to research-write so every
  research run begins with situational awareness. H.3 introduces a
  planning pipeline family — the first pipeline whose output isn't
  a research artifact but a plan + proposed follow-up work_items.
---

# Batch H.1.7 / H.2 / H.3 — context-gather and planning

> Pinky-and-the-Brain reference logged for the record: this batch
> is the substrate becoming the partner that helps Michael
> accomplish the things he's "always wanted to do but never had
> time for." Science center first. World domination Tuesdays only.

## 0. Why this batch exists

The substrate compounds. Every completed work_item leaves residue
— sabbath lessons, study artifacts, tool_dispatch trails, journal
entries naming what was built and what was learned. Today that
residue is *passive*. The vision this batch operationalizes:
**agents read it before they act, write to it when they learn,
next agent starts smarter than the last.**

Three of the substrate's existing phases already point this way:

- **Phase D sabbath lessons** — already a per-pipeline knowledge
  accretion. Lesson #17 from the physics-news run ("pair every
  abstract discovery with a buildable exhibit") is exactly the
  shape: substrate-observed insight, retrievable, applicable next
  time.
- **Phase E trust ladder** — already gates what an agent can do by
  track record. Extending from "retry-with-lessons" to
  "write-back-to-context-store" is the natural next rung.
- **Phase F council** — multi-agent council was always the
  endgame. A `context-gather` agent briefing the main agent before
  it acts IS a council member providing situational awareness.
  Phase F in concrete clothes.

So this batch is not new substrate; it's the substrate's
existing primitives reaching their intended use.

---

## I. Ratifications (this session — 2026-05-11)

### D-H1.7-A — Validation case for H.1.7 (RATIFIED)

**Substrate-reflective AI-tooling question.** Modified from the
"follow-up to AI tooling roundup" recommendation:

> *"Looking over this last week's roundup, what other improvements
> to our tooling can we make that industry is moving towards, that
> would be interesting for us to adopt? Self-reflect on the DB +
> agent substrate we've made and how what's going down in AI now
> could help us build that better. Consult our prior journals,
> proposals, mind files, and the 11-cycle guide."*

**Implication on H.1.7 scope:** original H.1.7 was DB-read only
(grant `pg-ai-stewards` MCP to research-write gather stage).
Validating *this* binding question requires reading
non-DB context: journals (`.spec/journal/*`), proposals
(`.spec/proposals/*`), mind files (`.mind/*`), and docs (`docs/*`).
The 11-cycle guide is at `.spec/proposals/pg-ai-stewards-11-cycle-review.md`
— already inside the scope above.

**Stewardship decision:** pull a minimal `fs-read` MCP forward
into H.1.7 with that exact scope. Cost goes from ~2hr to ~3-4hr.
Same machinery reused for per-pipeline-scoped fs-read in H.2/H.3.

### D-H3-A — Planning pipeline output (RATIFIED)

**Both: plan file + proposed work_items.** When an H.3 planning
run finishes, it produces:
1. A plan document materialized via existing
   `pending_file_writes` mechanism. File destination per pipeline
   convention.
2. Proposed follow-up work_items inserted at `maturity='raw'`
   into `stewards.work_items` with an `origin` of `'agent_planning'`
   (extending the `D-H7` `origin` column ratification — see
   `substrate-batch-h-pipeline-expansion.md` §VI). User reviews
   and ratifies (`maturity` advance) before they actually run.

### D-H3-B — Studies table generalization timing (RATIFIED)

**H.3 — alongside the planning family.** When the planning
pipeline lands, extend the `studies` schema with `tags text[]`,
`source_type text` (was: gospel-only implicit), and
`project_association text` so non-scripture studies are first-class
and searchable cross-domain. Context-gather stages query it.

### D-H3-C — fs-read scope (RATIFIED)

**Per-pipeline scoped.** Each pipeline family declares its fs-read
allowlist in `pipelines.fs_read_paths jsonb[]`. The fs-read MCP
server enforces the scope at request time using the pipeline_family
on the dispatching work_item.

For H.1.7 specifically (research-write): `.spec/journal/*`,
`.spec/proposals/*`, `.mind/*`, `docs/*`.

For H.3 planning, scopes are per-pipeline-instance (e.g., a
space-center planning pipeline gets `/projects/space-center/*`
added).

### Stewardship decisions (within ratified scope)

- **H.1.7 tool grants:** read-only subset of `pg-ai-stewards`
  MCP: `study_search`, `study_get`, `study_similar`,
  `study_citations`, `work_item_list`, `work_item_show`,
  `watchman_pass_show`, `watchman_passes_list`. Escalation
  write tools excluded.
- **Add `context-gather` to research-write too**, not just the
  new planning family. Marginal cost; every future research run
  gets smarter.
- **`context-gather` is not a new stage type.** It's a stage with
  a name + tool grants. Existing `stage_results` JSONB +
  `input_template` substitution handles output flow. No new
  substrate machinery.
- **fs-read MCP lives at `projects/pg-ai-stewards/cmd/fs-read-mcp/`**
  per user direction to co-locate MCPs related to the substrate.

---

## II. H.1.7 — research agent reads substrate state

### II.1 Scope

Grant the research-write `gather` stage tools that let it read:
1. Prior studies, work_items, and watchman passes (via existing
   `pg-ai-stewards` MCP — read tools only)
2. Project context files — journals, proposals, mind, docs (via
   new `fs-read` MCP)

### II.2 fs-read MCP server

**Location:** `projects/pg-ai-stewards/cmd/fs-read-mcp/`

**Language:** Go (matches stewards-mcp, mcp-server-go skill).

**Tools:**

| Tool | Purpose |
|---|---|
| `fs_list` | List files matching a glob within the configured scope. Inputs: `glob` (string), optional `limit` (int). Output: array of paths. |
| `fs_read` | Read a file's contents, max ~50KB to prevent runaway. Inputs: `path` (string, must be within scope). Output: text content. |
| `fs_search` | grep-style search for text within scoped files. Inputs: `pattern` (regex), optional `path_glob`. Output: matches with file + line. |

**Path-scope enforcement:** server takes `--allowed-paths` flag at
startup (comma-separated glob list). Every request validates the
target path resolves within the allowed set; reject otherwise.

**Repo-root resolution:** server takes `--repo-root` flag pointing
to the project root. All paths in tool calls are relative to repo
root. (Same convention as stewards-mcp's reading of `.spec/`.)

**Bridge wiring:** registered in `stewards.mcp_servers` with
`transport='stdio'`. Bridge spawns the binary at request time
(or as a long-lived process — TBD during build, follow whatever
stewards-mcp does).

### II.3 Substrate wiring

1. Insert/update row in `stewards.mcp_servers` for `pg-ai-stewards`
   if not already present. Mark `enabled=true`.
2. Insert row for `fs-read` with `enabled=true` and the H.1.7 path
   scope.
3. Update `pipelines.tool_grants` for research-write's `gather`
   stage to include both servers' read tool names.
4. Tighten the gather `input_template` to instruct the agent to
   **first** consult prior work via these new tools before
   external search. Pattern:
   ```
   Before external search, check what we already know:
   - study_search for related prior work
   - fs_search for journal entries / proposals on this topic
   - work_item_list to see what's been queued or completed
   Then external search for what's NEW since our last touch.
   ```

### II.4 Validation

Run the D-H1.7-A binding question. Verify:
- Agent issues at least 2 substrate-read tool calls before any
  external search
- Final draft references at least one specific prior journal or
  proposal by name
- Cost stays under $0.40 (the H.1.6.5 baseline was $0.19)
- File materializes via existing auto-mat path

If validation passes: commit. If not: surface what broke and
discuss before committing.

---

## III. H.2 — context-gather as research-write's first stage

### III.1 Scope

Today's research-write has three stages: `gather → synthesize →
review`. H.2 adds a new first stage `context-gather` with these
properties:

- Tool grants: same as H.1.7 (`pg-ai-stewards` read tools +
  `fs-read`)
- Tools_disabled = false (it needs the tools — this is its
  whole point)
- Input template: receives the binding question; produces a
  briefing of "here's what we already know about this topic"
- Output flows to the `gather` stage via
  `{{stage_results.context_gather.output}}` prepended into
  gather's input_template

### III.2 Why this is different from H.1.7

H.1.7 gave the gather stage the *tools* to consult prior work
mixed with its external research. The agent might or might not
use them well. H.2 makes the consultation a *separate stage* with
its own model invocation and its own output — so:
- Context-gathering and external research can use different
  models (context-gather can be cheaper / smaller; gather stays
  on the strong model)
- The briefing is inspectable in `stage_results` — we can debug
  whether context-gather found the right things even if the
  main run goes sideways
- Future pipelines (H.3 planning, future yt pipelines) can
  reuse the exact same context-gather stage shape

### III.3 Implementation

1. SQL migration: update `pipelines.stages` for research-write to
   insert the new stage at position 0. Existing stages renumber.
2. Stage definition (jsonb):
   ```json
   {
     "name": "context_gather",
     "next": "gather",
     "agent": "research",
     "model_preference": "qwen3.6-plus",
     "tool_grants": ["fs_read", "fs_search", "fs_list",
                     "study_search", "study_get", "study_similar",
                     "work_item_list", "work_item_show"],
     "input_template": "<see below>",
     "produces_maturity": null
   }
   ```
3. Stage maturity mapping: context_gather does NOT produce a new
   maturity rung. The first real maturity advance happens at gather
   (raw → researched).
4. Update research-write's gather input_template to consume
   `{{stage_results.context_gather.output}}` as a "Prior context"
   section before the original gather instructions.

### III.4 Validation

Run a different real research question. Inspect:
- `stage_results.context_gather.output` — is the briefing
  substantive?
- Gather stage chat input — does the briefing actually appear
  prepended?
- Final draft — does it reflect awareness of prior work?

Commit on pass.

---

## IV. H.3 — planning pipeline family (design with open questions)

H.3 is design-ready but has real open questions that benefit
from a ratification pass before code. This section is the design;
§IV.6 lists the open questions.

### IV.1 The shape

`planning` pipeline family. Stages:

1. **context_gather** (reused from H.2) — what do we already know
   about this project/topic?
2. **explore** — open exploration; surface assumptions, identify
   risks, ask back questions. Maturity: → researched. Tools:
   pg-ai-stewards reads + fs-read (per-pipeline scope) + external
   search.
3. **synthesize** — pull the exploration into a structured plan
   document. Maturity: → planned. Tools off (just shape the text).
4. **propose_work** — emit JSON listing proposed follow-up
   work_items. Maturity: → planned. Tools off.
5. **review** — verify the plan and the proposed work_items are
   coherent. Maturity: → verified.

### IV.2 Output: dual artifact

When the pipeline reaches verified, the trigger from H.1.6.2
fires:
- `enqueue_work_item_file` materializes the plan document
- New: `enqueue_proposed_work_items` reads
  `stage_results.propose_work.output` (JSON array of
  `{slug, binding_question, pipeline_family_hint}` items) and
  inserts each as `maturity='raw'`, `origin='agent_planning'`,
  `parent_work_item_id` pointing back at the planning run.

User reviews the proposed work_items in the UI and ratifies
(advances maturity) before they fire.

### IV.3 Intent

New intent file: `.spec/intents/planning-partner.yaml`. Values
that distinguish it from general-research:
- "Surface assumptions before recommending."
- "Ask back questions when the binding is underspecified — don't
  invent the answer."
- "Propose follow-up work small enough to actually finish."
- "Prefer one strong plan that ships over five branches that
  don't."

### IV.4 File destination

Pipeline-level template: `projects/<project>/plans/<slug>.md`
where `<project>` comes from the work_item's `project_association`
column (new — see D-H3-B studies generalization; same column
extended to work_items).

For non-project planning (e.g., a planning run that's not tied to
a project), fallback: `plans/<slug>.md`.

### IV.5 Studies generalization (D-H3-B)

Migration to make studies cross-domain:
```sql
ALTER TABLE stewards.studies
    ADD COLUMN tags text[] DEFAULT '{}',
    ADD COLUMN source_type text,
    ADD COLUMN project_association text;
CREATE INDEX studies_tags_gin ON stewards.studies USING gin(tags);
CREATE INDEX studies_project_association_idx ON stewards.studies(project_association);
```

`source_type` values seeded: `scripture-study` (existing rows),
`research`, `plan`, `yt-evaluation`, etc.

Existing `studies.file_path NOT NULL` blocker (open-items §1.1)
must be resolved as part of this migration — make it nullable or
compute at insert.

### IV.6 Open questions (need ratification before H.3 build)

These are the questions I'm NOT deciding within stewardship —
each is a load-bearing design choice that affects future work.

**Q-H3.1 — How does the `propose_work` stage emit work_items?**
- (a) Strict JSON output, substrate validates schema, rejects
  malformed → revise loop
- (b) Free-form output, post-stage parser extracts proposed items
  by pattern
- (c) Tool call: give the stage a `propose_work_item` tool that
  inserts directly into a `proposed_work_items` staging table

**Q-H3.2 — When proposed work_items land, do they auto-show up in
the UI's "Inbox" or "Proposals" surface, or live in a separate
review queue?**
- Affects stewards-ui scope. May be a follow-up UI batch.

**Q-H3.3 — Cost cap for planning pipelines.** Research-write caps
at ~$0.40. Planning has 5 stages vs 3, plus uses tools more
heavily. Likely cap: $0.75-$1.00. Need ratification.

**Q-H3.4 — Does planning need its own review template?** The
existing review template asks "does this verify the binding
question?" — planning's binding question is exploratory, so
"verifies" is a different shape. Probably needs a
`review_plan` gate prompt.

**Q-H3.5 — `project_association` enum or freeform?** Freeform is
flexible; enum prevents typos and enables project-page UI
features later. Probably freeform with a "known projects" view.

### IV.7 H.3 build plan once ratified

Six commits, sized like the C/D/E/F cadence:
1. studies schema migration + work_items.project_association
2. planning-partner intent
3. planning pipeline definition + stages
4. propose_work_items mechanism (Q-H3.1 answer)
5. review_plan gate prompt
6. UI surface for proposed work_items (if Q-H3.2 says inline)

---

## V. Sequencing this session

| Sub-batch | Status |
|---|---|
| Write this proposal | doing now |
| H.1.7 — fs-read MCP + grant + validate | this session, full ship |
| H.2 — context-gather stage on research-write | this session if H.1.7 validates |
| H.3 — design proposal + open Qs | this session (the §IV.6 questions to user) |
| H.3 build | NEXT session, after §IV.6 ratification |

---

## VI. Carry-forward & risks

- **fs-read sandbox correctness is load-bearing.** The MCP runs
  inside the bridge container with whatever filesystem access the
  container has. The `--allowed-paths` flag must be enforced
  *before* path expansion — and symlink traversal must resolve
  + re-check. Sandbox bypass = the agent can read anything the
  bridge can.
- **`context-gather` model selection matters.** A weak model
  here produces a weak briefing that pollutes downstream stages.
  Start on qwen3.6-plus (cheaper), upgrade if briefings are thin.
- **Briefing length budget.** A 5KB context-gather output
  consumes the gather stage's prompt budget. Cap context-gather
  output at ~3KB; instruct it to summarize, not transcribe.
- **Existing studies rows have `file_path NOT NULL`.** The pre-
  existing bug (open-items §1.1) gets resolved in H.3's studies
  migration.

---

## VII. Update points

- Add H.1.7/H.2/H.3 entries to
  `substrate-batch-h-pipeline-expansion.md` §V table.
- Update `.spec/open-items.md` § 0 active proposal queue.
- Update `.mind/active.md` at session end.
- Journal entry to `.spec/journal/` capturing what shipped + the
  open §IV.6 questions for next session.
