---
workstream: WS5
status: shipped (2.5 done; 2.6 design pending)
created: 2026-05-04
shipped: 2026-05-04
---

# pg-ai-stewards Phase 2.5 + 2.6 — Generic Document Substrate

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
   the [stewards-cli README](../../scripts/stewards-cli/README.md) but
   the headline: `proposal-pg-ai-stewards-phase-2-5-generic-substrate`'s
   top-3 mutual neighbors are the journal entries from Phase 2.2, 2.3,
   and 2.4 — the substrate rendered the timeline of its own creation.
4. **CLI shipped:** `scripts/stewards-cli/` with `import`,
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
scripts/stewards-cli/
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

### Open ontology questions (NOT to be answered while implementing 2.5)

- **How are typed semantic edges (`:FEEDS`, `:REFINES`) authored?**
  Three options: (a) declared in frontmatter (`feeds: [other-slug]`);
  (b) inferred from markdown link text ("see also X", "supersedes Y");
  (c) computed by an LLM pass over the corpus. Probably some
  combination — frontmatter for the user-authored cases, LLM pass for
  retroactive discovery on existing documents.
- **What's a "todo" as a vertex?** Right now todos live in
  `manage_todo_list` calls (ephemeral) and in `.mind/active.md`
  bullets (semi-persistent). Neither is a file. Does Phase 2.6
  introduce a `stewards.todos` table, or extract todos from existing
  documents as virtual vertices?
- **How does workstream backfill happen?** Some workstreams have
  multiple proposals; some proposals have no workstream tag. What's
  the migration story?
- **What edges does the agent walk for "what am I working on"?** This
  is the canonical query. Proposed: `MATCH path = (active_phase)-[:HAS_PHASE|HAS_PROPOSAL|HAS_TODO|FEEDS|SIMILAR_TO*1..2]-(other) WHERE active_phase.status = 'in-progress' RETURN other ORDER BY [some scoring]`. The scoring needs design.

### Phase 2.6 explicitly defers to its own writeup

**Do not start 2.6 work until the open questions above have a written
answer.** The temptation will be to "just add an edge while we're in
there" during 2.5. Resist. The point of separating 2.6 is so the edge
ontology is designed coherently, not accreted.

## Plan persistence

This file IS the plan. Updates as 2.5 progresses go directly into the
relevant section above. When 2.5 ships, summarize in
[extension/README.md](../../projects/pg-ai-stewards/extension/README.md)
and write the journal entry. When 2.6 starts, fork its open-questions
section into its own proposal file.
