# Book of Mormon Walkthrough — Knowledge-Graph Build (Plan)

**Status:** rigor built (`study/bom-walk/` scaffold + `_workflow.md`), awaiting Michael's go + goal-tool. Intended as a **one-go continuous run** (goal tool + ammon carry it to completion across compactions), one chapter at a time.
**Origin:** Michael, 2026-06-01 — "have you (or other agents) walk through the book of mormon start to finish, and record your own thoughts and connections as you digest each new section. and see what notable things pop as you use the tools we've developed to make those connections, building a wiki/knowledge graph that we can pull from. that can be 100% you too since it's working on something I need but the instructions are clear."
**Author:** Claude Code (Opus 4.8). Durable so it survives compaction.

---

## What this is (and what keeps it honest)

A start-to-finish digestion of the Book of Mormon where the agent records **its own** thoughts and connections per section, using the workspace tools to surface non-obvious links, accreting into a wiki/knowledge graph Michael can pull from.

**Bin 1-2 (stuffy-in-the-loop): gathering + drafting-for-his-use.** Authorized unsupervised. Two guardrails make it stay honest:

1. **My notes are clearly *mine* — exploratory digestion, not doctrine.** They are not authoritative pronouncements and not finished studies. Speculation is labeled speculation (epistemic humility). This stays out of bin 4 precisely because it never claims finality.
2. **Every quoted phrase is verified against `gospel-library/`** (read-before-quoting; `gospel_get`/`Read`). A walkthrough that confabulates as it goes would poison the graph it builds.

## Where it lives

- **`study/bom-walk/`** — git-tracked markdown, one note per chapter, in book subfolders. The pullable wiki. Scaffold built 2026-06-01: `README.md`, `_workflow.md` (per-chapter rigor), `_progress.md` (tracker + NEXT pointer = the ammon resume anchor), `_journal.md` (Opus's standout reflections), `_graph.md` (node/edge index).
- **Graph** `study/bom-walk/_graph.md` — markdown node/edge registry, grown per chapter. (Not persisted into the substrate corpus — markdown is git-tracked, reviewable, and pullable; pg-ai-stewards is used to *discover* connections, not to *store* the graph.)
- **`brain_*` is dead — not used.** The substrate study tools replace it.

## Per-section note format

```
# {Book} {Chapter(s)} — {short title}
- **Read:** what the section says (brief, my framing)
- **Thoughts:** my own digestion — what strikes me, tensions, questions
- **Connections (tool-surfaced):** cross-refs found via gospel_search (semantic/hybrid),
  links to our existing studies via study_search/study_similar, word work via webster
- **Entities:** people / places / doctrines / types-&-symbols / prophecies / covenants touched
- **Edges:** {this} —[cross-ref|fulfillment|parallel|type→antitype|covenant-thread]→ {that}
- **Notable / flag:** anything that "pops" worth Michael's eye
- **Verified:** quotes checked against gospel-library ✓
```

## The graph schema

- **Node types:** person, place, doctrine, type/symbol, prophecy, covenant, event, study-link.
- **Edge types:** cross-reference, fulfillment (prophecy→event), parallel (type→antitype), covenant-thread, doctrinal-development, links-to-our-study.
- Built incrementally; each section adds nodes/edges; periodic "what's emerging" synthesis passes surface the notable patterns Michael wants.

## Tools (the point — let them surface what recall wouldn't)

- `gospel_get` — chapter text *with footnotes* (verified live: returns the full footnote cross-ref chains).
- `gospel_search` (semantic + hybrid) — non-obvious scripture cross-refs on each chapter's binding idea.
- `study_search` / `study_similar` (pg-ai-stewards) — connect BoM chapters to our 198 existing studies. **Retrieval only — DB queries, no opus, no cost** (verified live).
- `webster_define` — word work where it earns it.
- `byu_citations` — how the Brethren have used a passage (use sparingly).

**Cost rule:** the substrate is a *connection index*, never a per-chapter generator. No `panel_redline` / `start_brainstorm` / `spawn_subagent` / `deep_research` in the loop — that would burn the $18 zen budget (≈ one study workflow). The thinking is Claude Code's, where tokens are plentiful.

## Cadence

Incremental, multi-session. Start at **1 Nephi 1**. A natural unit is a chapter or a coherent block (e.g., Nephi's vision arc). Not a single sprint — 239 chapters across 15 books. Each session: digest a block, extend the graph, run a light synthesis, commit. Walks back cheaply (git); proposals-quality, not published-study-quality.

## Relationship to the provenance review

Distinct effort. Provenance review (audit/sharpen EXISTING work) is the immediate "this study" Michael gated on compaction. This (generate NEW digestion/graph) is the standing parallel/next effort. The provenance review's verification discipline is the same discipline this walkthrough runs under — they reinforce each other.

## Resume instructions

1. Re-read this file + the voice baselines (CLAUDE.md writing-voice section).
2. Confirm `study/bom-walk/` exists (create if not) + the graph index convention.
3. Begin/continue at the next un-digested section; one block per pass; verify quotes; extend graph; commit; periodic synthesis.
