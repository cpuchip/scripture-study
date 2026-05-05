---
workstream: WS5
status: 2.5 shipped; 2.6 spec'd; 2.7 sketched
created: 2026-05-04
shipped: 2026-05-04
feeds:
  - proposal-sabbath-agent
  - proposal-token-efficiency
  - proposal-brain-vscode-bridge-main
---

# pg-ai-stewards Phase 2.5 + 2.6 + 2.7 — Generic Substrate, Typed Edges, Watchman

> *Scratch context: this proposal exists because Phase 2.4 surfaced
> that `stewards.studies` was already generic enough in schema to hold
> any document, and that the implicit graph of Michael's work spans
> studies, planning docs, proposals, journals, and ai-collaboration
> writing. Folder names were load-bearing for nothing. This proposal
> closes the gap in two scoped phases.*

## The vision

**Three layers, each its own phase:**

| Layer | What | Phase |
|-------|------|-------|
| **L1** | Beyond just studies — `docs/work-with-ai/`, lessons, etc. flow into the same embedding/similarity graph | **2.5** |
| **L2** | Typed documents as first-class shapes — `proposal`, `phase`, `journal`, `scratch`, `memory` get a `kind` column + per-kind ingest | **2.5** |
| **L3** | Typed edges + hierarchy — `Workstream →[:HAS_PROPOSAL]→ Proposal →[:HAS_PHASE]→ Phase →[:HAS_TODO]→ Todo`, plus cross-edges `:FEEDS`, `:REFINES`, `:IMPLEMENTS`. Progressive context disclosure via graph walks. | **2.6** |

L1 and L2 collapse into one schema change (one `kind` column) and one
parser-shape pass. L3 is its own design — it needs its own ontology
sweep and shouldn't be guessed during implementation.

## Phase 2.5 outcome (shipped 2026-05-04)

**Done.** All five "done when" gates met. Notable findings beyond the spec:

1. **The old PowerShell importer was silently losing 12% of the corpus.**
   188 study .md files exist, but the old slug-from-basename strategy
   collapsed 30 of them via collisions (`art-of-delegation.md` exists
   in both `study/` and `study/talks/`; the second always overwrote
   the first). Phase 2.5's slug builder — bare basename for root-level
   studies, kind+subdir prefix elsewhere — recovered all of them. Total
   import after fix: 359 docs (188 study + 73 proposal + 65 journal +
   32 doc + 1 phase-doc).
2. **Journal YAML schema drifted across versions.** Early entries used
   `[]string` discoveries and `relational_dynamics`; later entries used
   `[{title, detail}]` and `relationship`. The first parser pass failed
   on 14 entries. Loosening to `map[string]any` + `firstString(keys...)`
   + per-shape dispatch fixed 13 of 14. The 15th was a genuine YAML
   syntax error in the source file; we now fall back to indexing the
   raw text rather than rejecting the entry.
3. **Cross-kind bridge formed strongly.** The full verification is in
   the [stewards-cli README](../../projects/pg-ai-stewards/cmd/stewards-cli/README.md) but
   the headline: `proposal-pg-ai-stewards-phase-2-5-generic-substrate`'s
   top-3 mutual neighbors are the journal entries from Phase 2.2, 2.3,
   and 2.4 — the substrate rendered the timeline of its own creation.
4. **CLI shipped:** `projects/pg-ai-stewards/cmd/stewards-cli/` with `import`,
   `study show|list|refresh`. Cross-compiles cleanly to linux/amd64
   (13.8 MB binary).

## Phase 2.5 — multi-corpus ingest + Go CLI

### Done when

1. `stewards-cli` is a Go binary, cross-compiled to linux/amd64 and
   windows/amd64. Subcommands: `import`, `study show|list|refresh`.
   Replaces both `stewards.ps1` and `import-studies.ps1`.
2. `stewards.studies` has a `kind text NOT NULL DEFAULT 'study'`
   column with a btree index. Existing rows backfilled to `'study'`.
3. The full corpus is imported (~167 documents):
   - 69 × `kind='study'` from `study/*.md` (existing)
   - 9 × `kind='doc'` from `docs/work-with-ai/*.md`
   - 23 × `kind='proposal'` from `.spec/proposals/*.md`
   - 1 × `kind='phase-doc'` from `projects/pg-ai-stewards/phases.md`
     (kept as one document; splitting per-phase deferred to 2.6)
   - 65 × `kind='journal'` from `.spec/journal/*.yaml`
4. `stewards study show <slug>` prints the `kind` line in the header
   and works on all kinds (not just kind='study'). Similarity edges
   freely cross kinds.
5. The bridge prediction is verified: `study_similar('charity')` and
   `study_similar('creation')` should now include `01_planning-then-create-gospel`
   (or `intelligence-cleaveth-gospel`) as neighbors. If it doesn't,
   that's a finding worth understanding.

### Why Go CLI now

- **Cross-platform deploy.** Linux hosting server needs the same
  tooling without a PowerShell port.
- **No more codepage fights.** PowerShell + psql + Windows = em-dash
  mojibake by default. `pgx` over libpq is unicode-clean.
- **No more reserved-name traps.** `-DB` aliasing `-Debug`, `-User`
  aliasing `-UserName` — both bit us in 2.4.
- **One importer, multiple sources.** PowerShell importer was
  hardcoded to `study/`. A Go subcommand `import --source <kind>:<dir>`
  takes the same shape across kinds.

### Schema migration (additive, non-breaking)

```sql
ALTER TABLE stewards.studies
    ADD COLUMN kind text NOT NULL DEFAULT 'study';

CREATE INDEX studies_kind_idx ON stewards.studies (kind);

-- Optional CHECK once we know the canonical list:
-- ALTER TABLE stewards.studies
--     ADD CONSTRAINT studies_kind_check
--     CHECK (kind IN ('study','doc','proposal','phase-doc','journal','scratch','memory'));
-- Defer the CHECK to 2.6 once the kind taxonomy stabilizes.

-- import_study gains an optional p_kind parameter (default 'study')
-- so existing callers don't break.
```

### Shape considerations per kind

This is where the parser pain Michael wants to surface lives. Each
kind has different "natural fields":

| Kind | Title source | Body source | Frontmatter | Quirks |
|------|--------------|-------------|-------------|--------|
| `study` | First H1 | full markdown | `*key: value*` italic lines (legacy) | em-dashes, gospel links |
| `doc` | First H1 | full markdown | YAML frontmatter optional | mostly clean |
| `proposal` | First H1 after frontmatter | full markdown | YAML: `workstream`, `status`, `created` | always has frontmatter |
| `phase-doc` | First H1 | full markdown (very large) | none | one big file with many `## Phase X.Y` sections |
| `journal` | derived from `intent` field (truncated) or `session_id` | synthesized: `intent` + flattened `discoveries[].title + .detail` + `surprises` + `relationship[].quality + .detail` | full YAML structure | NOT markdown — YAML throughout |

**The journal is the parser challenge.** It's structured YAML, not
markdown. The "body" we embed is a synthesized text composition. The
synthesis function lives in the Go importer, not in PL/pgSQL — way
easier to test there.

**The phases.md is the structural challenge** — we import it as one
document for now; 2.6 will split per-phase. Naming the split work
explicitly here so we don't forget.

### The Go CLI shape

```
projects/pg-ai-stewards/cmd/stewards-cli/   # moved here 2026-05-04
├── go.mod
├── main.go              # cobra/flag, subcommand routing
├── internal/
│   ├── db/conn.go       # pgx pool, env config
│   ├── importer/
│   │   ├── importer.go      # interface: Parse(path) → DocFields
│   │   ├── markdown.go      # study, doc, proposal, phase-doc
│   │   └── journal.go       # YAML synthesis
│   └── show/show.go     # calls SQL functions, prints
└── README.md
```

Subcommands:
- `stewards-cli import --source study:study --source doc:docs/work-with-ai --source proposal:.spec/proposals --source journal:.spec/journal --source phase-doc:projects/pg-ai-stewards/phases.md`
- `stewards-cli study show <slug>`
- `stewards-cli study list [--kind <kind>]`
- `stewards-cli study refresh [<slug>]`

Connection config: `STEWARDS_DSN` env var (default
`postgres://stewards:stewards@localhost:5432/stewards?sslmode=disable`).
For local dev against the docker container, the script writer can
either expose 5432 or `docker exec` (a `--via-docker` flag fallback
for users who don't want the host port published).

### Verification gates

1. **Import idempotency.** Running import twice doesn't duplicate or
   change row counts. Embedded count == total count after embed
   bgworker drains.
2. **Cross-kind similarity bridge.** Pick the prediction:
   `study_similar('creation')` returns at least one `kind='doc'` row.
   If not, the finding is recorded — it might mean the embedder is
   over-weighting domain vocabulary, or it might mean the gospel ↔
   AI bridge isn't as strong as expected. Either is information.
3. **Show works for all kinds.** `stewards study show <some-proposal-slug>`
   and `stewards study show <some-journal-slug>` both render. Citations
   for kinds that don't cite scriptures (proposals, journals) gracefully
   show "no citations" instead of erroring.
4. **Linux build smoke test.** `GOOS=linux GOARCH=amd64 go build` produces
   a binary. Don't deploy it yet — just verify it compiles and connects
   when targeted at the local docker container's exposed port.

### Risks

- **Journal synthesis is opinionated.** What text do we embed for a
  journal entry? If we get the synthesis wrong, journals cluster
  with each other but don't bridge to studies. Mitigation: log the
  synthesized body for the first few during import; tune if the
  similarity output looks weird.
- **Proposal body includes the proposal of THIS proposal.** Recursion
  isn't the issue (it's a flat document); the issue is the proposal
  about \"generic substrate\" will cluster with itself in a navel-gazing
  way. Acceptable; it's just one row.
- **Phases.md is 700+ lines.** Embedding such a long document is a
  blunt instrument; the embed will be a coarse "what is this project
  about" vector. That's fine for 2.5; 2.6 splits per-phase for
  fine-grained edges.

## Phase 2.6 — typed edges + hierarchy

### The vision (Michael's words, paraphrased)

> Workstream is the line of work. Then proposals link to it. Then
> phases link to them with specific details, and todos/info/scratch
> link to phases — and you can link and update cross-ways and allow
> progressive context disclosure through queries and what you're
> working on. By connecting graph/semantic edges it can surface
> automatically how things are connected.

### What this requires

1. **Workstream as a first-class vertex** (`:Workstream {id: 'WS5', title: 'pg-ai-stewards', status: 'active'}`).
   Currently lives as a frontmatter `workstream:` field in proposals.
   Lift it.
2. **Typed structural edges:**
   - `:Workstream -[:HAS_PROPOSAL]-> :Proposal`
   - `:Proposal -[:HAS_PHASE]-> :Phase` (requires phase-doc splitting)
   - `:Phase -[:HAS_TODO]-> :Todo` (requires a Todo concept)
   - `:Phase -[:PRODUCED]-> :Journal` (link journal entries by `tags:` + `session_id` matching)
3. **Typed semantic edges (alongside existing `:SIMILAR_TO`):**
   - `:FEEDS` — this doc's ideas show up later in another doc
   - `:REFINES` — this doc supersedes / clarifies an earlier one
   - `:IMPLEMENTS` — code or proposal implements an earlier idea
4. **Phase splitting parser** — turn `projects/*/phases.md` into N
   `:Phase` vertices, one per `## Phase X.Y` section.
5. **Progressive context disclosure.** A new SQL function
   `stewards.context_for(slug, depth)` that walks the graph from the
   given vertex outward, returning a ranked list of related vertices
   for the agent to optionally pull in. Replaces "load this whole
   proposal" with "load the workstream, the parent proposal, sibling
   phases, and the most-similar prior journal entries."

### Resolved ontology decisions (2026-05-04 — answers Michael landed on after Phase 2.5 shipped)

These were the four open questions; they now have answers, and 2.6
implements against them rather than re-deriving.

**1. How are typed semantic edges authored? Three doors, one schema.**

All three authorship paths land in the same edge structure with a
`provenance` discriminator:

```cypher
(:Study|Doc|Proposal|...) -[:FEEDS {
    provenance: 'declared' | 'linked' | 'inferred',
    confidence: 0.0..1.0,
    source: '<frontmatter-key | markdown-link | model-id>',
    created_at: timestamptz,
    confirmed_by: 'human' | 'agent-session-id' | null
}]-> (:Study|Doc|Proposal|...)
```

- **`'declared'`** — frontmatter (`feeds: [other-slug]`,
  `supersedes: [other-slug]`, `implements: [other-slug]`).
  Authoritative. Confidence 1.0. No confirmation needed.
- **`'linked'`** — markdown links (`[X](path/to/Y.md)`) parsed during
  import become `:REFERENCES` edges by default; certain link contexts
  ("see also", "supersedes", "based on") promote to typed edges via
  regex over the surrounding sentence. Confidence 0.7-0.9. Auto-merged.
- **`'inferred'`** — LLM pass (kimi-k2.6, cheap and good) runs as a
  bgworker job over similarity neighbors above a threshold. Writes to
  `stewards.edge_proposals` rather than directly to the graph. Never
  auto-merges into the active graph. Confidence per the model's output.

Default queries filter to `provenance IN ('declared', 'linked')` for
trust. The inferred layer is a *suggestion surface* the agent or human
walks separately. This is the **Restoration discernment standard
applied to graph edges**: the substrate can propose, only the steward
can confirm.

**2. What's a "todo" as a vertex?** A persistent connector vertex with
state and lifecycle. Implemented in 2.6. Key properties:

- Scope: any document kind (workstream | proposal | phase | etc.)
- Lifecycle: `open | in_progress | done | dropped` — todos persist
  after completion as a permanent record of what was done. Marking
  done is a state change, not deletion.
- Roll-up as audit: when a parent (phase, proposal) is marked done
  with open child todos, that's a *correctness signal* — we shipped
  with known unfinished work. Surface as a query, never auto-resolve.
- Connector shape: a todo with both `:ON` (parent scope) and
  `:LINKS_TO` (other docs the work touches) edges is itself an
  edge-with-state — a hyperedge that can carry assignee, blocker, and
  agent-session provenance.

**3. Workstream backfill — solved by making it agent work, not
migration work.** Untagged proposals don't block 2.6. The system
proposes its own taxonomy:

- Frontmatter `workstream:` declarations are authoritative when present
- A bgworker job scans untagged docs, examines frontmatter date +
  similarity neighborhood + active.md context, proposes a workstream
  assignment, writes to `stewards.tag_proposals`
- The next agent session reviews and confirms (or corrects) those
  proposals — same mechanism as inferred edges
- This is the test case for whether the agentic-pipeline framing works
  on the system's own state. If it can't tag its own corpus correctly,
  it's not going to manage anyone else's work either.

**4. "What am I working on" is a status query, not a separate table.**
The truth lives in the graph. The query:

```cypher
MATCH (ws:Workstream {status: 'active'})
      -[:HAS_PROPOSAL]-> (p:Proposal {status: 'in-progress'})
      -[:HAS_PHASE]-> (ph:Phase {status: 'in-progress'})
      -[:HAS_TODO]-> (t:Todo)
WHERE t.status IN ('open', 'in_progress')
RETURN ws, p, ph, t
```

`stewards.context_for(slug, depth)` walks outward from a vertex,
returning ranked related vertices. `.mind/active.md` becomes a
*generated artifact* that Watchman (Phase 2.7) renders out of this
query during weekly consolidation — git-tracked, human-readable, but
never the source of truth. **Files become a renderable view of
substrate state.**

Full background and the conversation that produced these answers:
[2026-05-04 journal](../journal/2026-05-04--pg-ai-stewards-2-5.yaml)
and the agent-side reasoning in the Copilot session that landed 2.5.

---

## Phase 2.6 — Workstream + Todo + 3-door edges

**Status:** spec'd 2026-05-04. Implementation not yet started.
**Builds on:** 2.5 (kind column, generic substrate, importer).
**Defers to:** 2.7 (Watchman / consolidation), 2.8 (LLM-inferred edges).

### Binding question

Make the graph carry the *structure of how work happens*, not just the
*similarity of what's written*. After 2.6, the canonical query "what
am I working on, and what's adjacent to it" should run against typed
structural edges, with similarity edges available as a fallback layer
beneath them.

### Done when

1. `:Workstream` vertices exist for WS1-WS6 (and any others in
   `.mind/active.md`'s workstream table).
2. `:Todo` vertices exist with the lifecycle states above. The CLI
   has `stewards-cli todo create | done | drop | list` subcommands.
   Roll-up audit query exists and runs clean (no false-positive done
   parents with open children).
3. `phases.md` is split per-`## Phase X.Y`, producing N `:Phase`
   vertices per project. Slug pattern: `phase-{project-dir}-{X-Y}`.
4. Three-door edge ingestion works: declared edges from frontmatter,
   linked edges from markdown link parsing, edge proposals table for
   inferred (no LLM pass yet — that's 2.8).
5. `stewards.context_for(slug, depth)` returns the right neighborhood
   for at least three test cases:
   - `proposal-pg-ai-stewards-phase-2-6` returns the workstream, sibling
     phases, related journals, declared `:FEEDS` from 2.5
   - `study-charity` returns its kindred studies + any docs that
     declare `:FEEDS` from it
   - `journal-2026-05-04--pg-ai-stewards-2-5` returns its phase, sibling
     journal entries from the same workstream
6. `.mind/active.md` can be regenerated from the graph and matches
   what we'd write by hand (modulo formatting). Doesn't have to ship
   automated — proves the data is sufficient.

### What this requires

**1. Schema additions to `stewards.studies` family:**

- New `stewards.workstreams` table:
  `(id text PK, title text, status text, created_at, frontmatter jsonb)`
- New `stewards.todos` table:
  ```sql
  (id uuid PK, slug text UNIQUE, title text, body text,
   status text CHECK IN ('open','in_progress','done','dropped'),
   created_at, updated_at, completed_at,
   parent_kind text, parent_slug text,    -- denormalized for fast roll-up
   created_by_session text,
   completed_by_session text,
   frontmatter jsonb)
  ```
  Todos live in their own table, not in `studies`, because their
  lifecycle is different (rapid mutation vs. write-once).
- New `stewards.edge_proposals` table:
  ```sql
  (id uuid PK, from_slug text, to_slug text, edge_type text,
   provenance text, confidence float, source text,
   created_at, status text CHECK IN ('pending','confirmed','rejected'),
   reviewed_by text, reviewed_at)
  ```

**2. Importer changes:**

- Parse `workstream:`, `feeds:`, `supersedes:`, `implements:` from
  frontmatter → `:DECLARED` edges to graph.
- Parse markdown link bodies during import. Default: `:REFERENCES`
  edge with `provenance='linked'`. Promotion regex for "supersedes",
  "see also", "based on", "feeds into" → typed edge.
- Phase splitter: parse `projects/*/phases.md` by `## Phase X.Y`
  headers. Each section becomes a `:Phase` vertex. Section body is
  the embedded text.

**3. CLI additions:**

```
stewards-cli workstream list | show <id>
stewards-cli todo create --parent <kind>:<slug> --title "..." [--body "..."]
stewards-cli todo done <id-or-slug>
stewards-cli todo list [--scope <slug>] [--status <state>]
stewards-cli todo audit          # roll-up correctness check
stewards-cli context <slug> [--depth N]
stewards-cli edges proposed [--accept <id>] [--reject <id>]
```

**4. SQL functions:**

- `stewards.context_for(slug text, depth int DEFAULT 2)` — graph walk
- `stewards.todo_rollup_audit()` — returns rows where parent is done
  with open children, or all-children-done with parent still open
- `stewards.regenerate_active_md()` — produces the active.md content
  from the graph; doesn't write the file (CLI does that, optionally)

### Phasing within 2.6

Three sub-phases that each deliver value independently:

- **2.6a — Workstream + frontmatter edges.** Workstreams as vertices,
  frontmatter declarations parsed into `:DECLARED` edges. No todos
  yet. Done when `proposal-pg-ai-stewards-phase-2-6`'s graph
  neighborhood includes WS5 and the 2.5 proposal as `:DECLARED` edges.
- **2.6b — Todos.** Schema, CLI subcommands, roll-up audit. Backfill
  by importing existing `manage_todo_list` snapshots from journal
  entries. Done when an active session can create a todo, mark it
  done, see it in roll-up.
- **2.6c — Phase splitting + context_for.** Phase splitter for
  phases.md. `context_for()` query. `regenerate_active_md()`. Done
  when the regenerated active.md matches the hand-written one.

Each sub-phase ships independently. 2.6a unblocks 2.6b unblocks 2.6c,
but each is its own session's work.

### Risks and explicit non-goals

- **Not in scope: the LLM inference pass.** That's 2.8. The
  `edge_proposals` table exists; no producer fills it yet. Manual
  inserts for testing are fine.
- **Not in scope: Watchman / consolidation.** That's 2.7. 2.6 builds
  the substrate Watchman runs on but doesn't run any scheduled work.
- **Risk: todo schema drift.** The `parent_kind`/`parent_slug`
  denormalization could fall out of sync with the graph. Mitigation:
  a single function (`stewards.create_todo`) writes both the row AND
  the `:HAS_TODO` edge in one transaction; never write directly.
- **Risk: link parsing produces edge spam.** Every `[X](path/Y.md)` is
  candidate `:REFERENCES`. The corpus has thousands of these. Mitigation:
  cap edges per source-doc (e.g., 50), and exclude link targets that
  resolve to the same kind+slug already linked structurally.

---

## Phase 2.7 — Watchman (consolidation, freshness, anti-loop discipline)

**Status:** spec'd 2026-05-04 (sketch level — full design lives in
[study/yt/matt-pocock-ai-workflow-research.md](../../study/yt/matt-pocock-ai-workflow-research.md#point-1--sabbath-as-rem-sleep-cycle-for-spec-freshness)).
**Builds on:** 2.6 (workstream + todos + typed edges).
**Hard constraint:** must not loop on the same items burning tokens —
this is the brain v1 nudge bot's failure mode and is the one thing
2.7 absolutely cannot inherit.

### Binding question

Make the substrate self-maintaining without burning tokens on
already-evaluated work. The scheduler runs forever; the work it does
must shrink as items reach a terminal state.

### Anti-loop discipline (the load-bearing constraint)

Brain v1's nudge bot ran every 4 hours and re-scanned the same stale
entries on every pass. Token waste and no progress. Watchman 2.7 must
make this *structurally impossible*, not merely *unlikely*.

The discipline:

1. **Every doc has a `last_consolidated_at` and a `last_touched_at`
   timestamp.** Touched is set by any agent edit; consolidated is set
   by Watchman. Watchman skips docs where
   `last_consolidated_at >= last_touched_at` AND prior verdict was
   terminal (`done`, `superseded`, `archived`).
2. **Verdicts are terminal-or-not.** A consolidation pass produces one
   of: `clean` (still matches code/spec, no action), `drift` (action
   needed, surfaces to human), `done` (acceptance criteria met → archive
   recommendation), `superseded` (newer doc replaces it → archive
   recommendation), `stale` (untouched and not in active.md). Only
   `done`, `superseded`, and `archived` are terminal — items in those
   states never re-enter the queue unless explicitly touched.
3. **Drift surfaces to the human and exits the queue until acted on.**
   Watchman doesn't keep nagging. It writes ONE recommendation row to
   `stewards.consolidation_findings`, marks the doc with a
   `pending_review_since` timestamp, and stops looking at it. The
   human (or a session-end agent) clears the pending-review state by
   either resolving the drift or marking the finding as
   acknowledged/wontfix. *Until that happens, Watchman doesn't touch
   the doc again.*
4. **Token budget per pass is bounded.** Default: 50K tokens per pass,
   configurable. When the budget is hit, the pass stops cleanly and
   resumes from the same cursor next run. No infinite loops; every
   run terminates.
5. **A doc can be re-evaluated only when its `last_touched_at` advances
   past `last_consolidated_at`.** This is the dirty-bit. Without it,
   the same doc can never be consolidated twice, period. With it, the
   doc gets one consolidation pass per touch — bounded by how often
   the doc actually changes.

The combination: terminal verdicts + dirty-bit + per-pass token cap +
"surface once and stop" produces a system whose worst case is
"Watchman runs forever doing zero work" rather than "Watchman runs
forever burning tokens."

### Three phases per pass (mirroring SCM paper biology)

1. **NREM (consolidation).** For each dirty doc, evaluate against
   current code/spec/active state. Write verdict + finding row.
   Cheap model (haiku, qwen-3, or kimi-k2.6 in cheap-mode). This
   is most of the work.
2. **REM (synthesis).** Pick 3-5 docs from the `clean` set that are
   thematically clustered, ask the model "do these connect in a way
   the docs don't note?" Output to `stewards.learnings` as candidate
   insights. Skip on most passes — budget controls when this fires.
   Expensive model (full Opus or Kimi). Optional per pass.
3. **Forgetting.** For terminal verdicts (`done`, `superseded`),
   write archive recommendations to a queue. Never auto-archive —
   moves stay human-confirmed. Removes from active queries.

### Triggers

- **Time-based:** weekly cron (e.g., Sunday 03:00). The Sabbath
  schedule is doctrinally honest, not just cute.
- **Pressure-based:** when `active.md` exceeds a threshold (e.g., 15K
  tokens), a smaller pass runs to find archive candidates.
- **Idle-based:** when no human-in-loop session has run for >48h,
  a small synthesis-only pass runs.

### Done when (2.7)

1. Watchman runs as a bgworker job in pg-ai-stewards.
2. A 7-day soak test shows: total tokens-per-day decreasing as the
   corpus stabilizes (proves the anti-loop discipline works). Tighter
   than the original 30-day target because we don't have 30 days —
   the *trend* is what matters, and 7 days of touched-but-now-stable
   docs is enough signal. If the trend is wrong at day 7, fix and
   re-soak; if it's right, ship and let production extend the proof.
3. At least 3 drift findings surfaced to the human and resolved
   through the recommendation→action→acknowledgement loop.
4. At least 1 synthesis (REM) finding produced an insight the human
   judges genuinely new (not "you already wrote this").
5. `regenerate_active_md()` (from 2.6c) is automated as a Watchman
   output: a fresh `.mind/active.md` is written at the end of each
   weekly pass.

### Risks and explicit non-goals

- **Risk: confirmation bias in the dirty-bit.** If an agent edits a
  doc in a way that doesn't actually change its alignment, Watchman
  re-evaluates anyway and burns the budget. Mitigation: agents that
  edit docs should set `last_touched_at` only when they make
  semantically meaningful changes, not formatting-only edits. Add a
  `--touched` flag to the CLI's edit operations.
- **Risk: REM synthesis hallucinates connections.** Mitigation: REM
  output is always to `learnings` as candidate, never directly to the
  graph. Same discernment frame as inferred edges in 2.6.
- **Not in scope: Watchman editing its own instructions.** Phase 2.9 or
  later, if at all. The vision Michael named — "agent rewrites its own
  prompts/skills/agent-modes" — needs more covenant scaffolding than
  exists yet. Watchman 2.7 is read-mostly with one write surface
  (recommendations); it doesn't modify agents, instructions, or skills.

---

## Future phases (sketches, not specs)

**Phase 2.8 — LLM-inferred edges (kimi-k2.6).** The producer for
`edge_proposals`. Walks similarity neighbors above a confidence
threshold, asks the model "is there a typed semantic relationship
here, and if so which?" Writes proposals. Same anti-loop discipline as
Watchman: dirty-bit per pair, terminal verdicts, bounded budget.

**Phase 2.9 — Agent self-modification surface.** The hard one. Lets
agents propose changes to their own instructions/skills/agent-modes,
gated through a `stewards.self_modifications` review queue. Only
implementable after Watchman has 7+ days of clean operation, because
this is the surface where "the system goes off the rails" becomes
plausible. Covenant scaffolding TBD.

**Phase 3 — External arms.** Docker-sandboxed git work, MCP wiring,
multi-model dispatch (Google Gemini Pro/Flash, Veo, TTS for the space
center work; Anthropic Opus/Sonnet via API; Kimi via opencode go/zen; local
models via lm studio). The pg-ai-stewards bgworker pattern is the spine;
each model becomes a tool sidecar. Token cost per task becomes a
queryable metric, not an estimate. Tokenomics (per Michael's coining)
becomes first-class telemetry.

**Phase 4 — Project work.** Marsfield Science Center exhibits, D&D
campaign generation, Empty Epsilon prop/mission/voice content. The
substrate proves itself by producing tangible outputs in domains
outside its own development.

---

## Plan persistence

This file IS the plan. Updates as 2.5 progresses go directly into the
relevant section above. When 2.5 ships, summarize in
[extension/README.md](../../projects/pg-ai-stewards/extension/README.md)
and write the journal entry. When 2.6 starts, fork its open-questions
section into its own proposal file.
