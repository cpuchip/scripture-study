# Book of Mormon Walkthrough — Knowledge-Graph Build (Plan)

**Status:** scoped, authorized "100% you" — a standing multi-session effort, run incrementally.
**Origin:** Michael, 2026-06-01 — "have you (or other agents) walk through the book of mormon start to finish, and record your own thoughts and connections as you digest each new section. and see what notable things pop as you use the tools we've developed to make those connections, building a wiki/knowledge graph that we can pull from. that can be 100% you too since it's working on something I need but the instructions are clear."
**Author:** Claude Code (Opus 4.8). Durable so it survives compaction.

---

## What this is (and what keeps it honest)

A start-to-finish digestion of the Book of Mormon where the agent records **its own** thoughts and connections per section, using the workspace tools to surface non-obvious links, accreting into a wiki/knowledge graph Michael can pull from.

**Bin 1-2 (stuffy-in-the-loop): gathering + drafting-for-his-use.** Authorized unsupervised. Two guardrails make it stay honest:

1. **My notes are clearly *mine* — exploratory digestion, not doctrine.** They are not authoritative pronouncements and not finished studies. Speculation is labeled speculation (epistemic humility). This stays out of bin 4 precisely because it never claims finality.
2. **Every quoted phrase is verified against `gospel-library/`** (read-before-quoting; `gospel_get`/`Read`). A walkthrough that confabulates as it goes would poison the graph it builds.

## Where it lives

- **`study/bom-walk/`** — git-tracked markdown, one note per section, wikilinked. The pullable wiki.
- **Graph index** `study/bom-walk/_graph.md` — the node/edge registry (or a structured frontmatter convention per note that an index can be regenerated from).
- Optional: mirror high-signal nodes into the `becoming` brain (`brain_create`, tagged) for cross-study recall surfaces.

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

- `gospel_get` / `Read` — section text with footnotes.
- `gospel_search` (semantic + hybrid) — non-obvious cross-refs on each section's binding ideas.
- `study_search` / `study_similar` — connect BoM sections to our 198 existing studies.
- `webster_define` — word work where it earns it.
- `byu_citations` — how the Brethren have used a passage.
- `brain_*` — optional cross-study recall mirror.

## Cadence

Incremental, multi-session. Start at **1 Nephi 1**. A natural unit is a chapter or a coherent block (e.g., Nephi's vision arc). Not a single sprint — 239 chapters across 15 books. Each session: digest a block, extend the graph, run a light synthesis, commit. Walks back cheaply (git); proposals-quality, not published-study-quality.

## Relationship to the provenance review

Distinct effort. Provenance review (audit/sharpen EXISTING work) is the immediate "this study" Michael gated on compaction. This (generate NEW digestion/graph) is the standing parallel/next effort. The provenance review's verification discipline is the same discipline this walkthrough runs under — they reinforce each other.

## Resume instructions

1. Re-read this file + the voice baselines (CLAUDE.md writing-voice section).
2. Confirm `study/bom-walk/` exists (create if not) + the graph index convention.
3. Begin/continue at the next un-digested section; one block per pass; verify quotes; extend graph; commit; periodic synthesis.
