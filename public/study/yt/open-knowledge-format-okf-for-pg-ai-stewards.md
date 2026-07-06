# Open Knowledge Format (OKF) — research, and what's of use for pg-ai-stewards

**Trigger:** [youtu.be/k4sMSsMzX2g](https://youtu.be/k4sMSsMzX2g) — AI LABS, "Google's New Release Just
Fixed AI Systems" (11:53). **Researched:** 2026-06-28 against the primary source, not the video.
**Primary sources:** [Google Cloud blog (2026-06-12)](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
· [the v0.1 SPEC](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
· [the repo](https://github.com/GoogleCloudPlatform/knowledge-catalog/tree/main/okf).
**Question:** what, if anything, is of use to pg-ai-stewards?

## What OKF actually is

OKF v0.1 is an open, vendor-neutral spec (Google Cloud's Data Cloud team, 12 June 2026) for representing
**knowledge** — the metadata, context, and curated insight around data and systems — as **a directory of
markdown files with YAML frontmatter**. Its own one-line pitch: *"If you can `cat` a file, you can read
OKF; if you can `git clone` a repo, you can ship it."* It formalizes Karpathy's "LLM-wiki" pattern into a
portable interchange format. The whole v0.1 spec fits on a page.

- **Bundle** = a directory of markdown files (the unit of distribution). **Concept** = one file (one unit
  of knowledge). A concept's identity is its file path minus `.md` (`tables/users.md` → `tables/users`).
- **Frontmatter:** exactly **one required field — `type`** (a short, producer-chosen string for routing/
  filtering). Recommended (priority order): `title`, `description`, `resource` (a URI for the underlying
  asset), `tags`, `timestamp` (ISO 8601). Producers may add any keys; **consumers must preserve unknown
  keys and must never reject a bundle for unknown values.** The spec defines the *interoperability
  surface, not the content model.*
- **Cross-links** are ordinary inline markdown links (bundle-relative `/paths` survive file moves). The
  directory becomes a **graph** richer than the folder tree — there is no `links:` field.
- **Reserved files:** `index.md` (a curated directory listing for **progressive disclosure** — read the
  root index, pick a subdir, read its index, open one concept, all without loading the whole bundle) and
  `log.md` (chronological change history, newest first). Bundle-root `index.md` declares `okf_version`.
- **Ships with three reference implementations** (the format is the contribution, "not the tooling"): a
  **BigQuery enrichment agent** (ingest a source → emit an OKF bundle), a **static HTML visualizer**
  (`visualize` → one self-contained interactive force-directed graph, no backend, "Cited by" backlinks +
  search + type filter), and **three sample bundles**. The BigQuery ingest path is Google-platform-
  specific; OKF itself is platform-independent (composes with Obsidian/Notion/MkDocs/git/any agent).

The progressive-disclosure point is the load-bearing one: *"For a 10,000-concept bundle, the pattern is
the difference between a usable corpus and an unusable one."*

## The key framing: complementary, not competing

This is the same shape as the Chase AI "Agentic OS" review — a **file-system, agent-navigable knowledge
layer** set against pg-ai-stewards' **database-native semantic engine**. But OKF is not a rival to the
substrate. They sit on opposite sides of one boundary:

- **pg-ai-stewards is knowledge *in motion*** — live, queryable, compounding: the docs corpus with FTS +
  vector + RRF hybrid retrieval, engrams, the doc-construction digesters, the per-intent Zion pools. For
  the live query path, DB-native semantic retrieval *beats* OKF's keyword + index-tree navigation.
- **OKF is knowledge *at rest*** — portable, git-shippable, human- and agent-readable, consumer-
  independent. It is an *interchange format*, not an engine.

So the use isn't to rebuild the substrate around OKF. It's to have **the substrate speak OKF at its
edges** — the same way it already speaks MCP. OKF is being positioned exactly like MCP and skills were:
Google standardizing an emerging pattern it expects every agent to adopt. Speaking it keeps the
substrate interoperable as the ecosystem standardizes knowledge interchange.

## What's of use — concrete

**1. An OKF *export* adapter for intent/Zion pools — the sharpest win.** The substrate accumulates
knowledge per intent (work, Marsfield, …) in Postgres. An `okf_export(intent)` could emit a bundle: one
concept file per doc (frontmatter `type`/`title`/`description`/`tags`/`timestamp` straight from doc
metadata; the doc body as markdown), `index.md` files auto-generated from each doc's one-line
description, `log.md` from the doc/engram history, and markdown cross-links from the doc relationship
graph. The payoff is large: the substrate's private knowledge becomes **portable, git-versionable,
forkable, and consumable by any agent** — a teammate's Claude Code, another framework, a static viewer.
That *is* the reflect-steward's "Zion knowledge pool" made shareable, and the realization of OKF's
"separate knowledge from its consumer" principle — the substrate's knowledge escapes the DB silo when
sharing is wanted, without giving up the DB-native core.

**2. An OKF *import* path into the doc corpus.** The doc-construction digesters could ingest an OKF
bundle (a partner's catalog, a shared team wiki, eventually websites shipping OKF beside `llms.txt`):
parse each concept + frontmatter → a row in the docs table (frontmatter → metadata, body → content +
embedding, markdown links → graph edges). The world is about to *produce* OKF; the substrate being able
to *consume* it is ecosystem interop for almost no cost (it's just markdown + YAML).

**3. Align doc metadata to OKF's frontmatter — and inherit its discipline.** Shaping the digesters'
output toward `{type, title, description, resource, tags, timestamp}` makes export trivial *and* imposes
good chunking hygiene: **one concept = one thing** (minimalism), a `type` for routing, and a forced
**one-sentence `description`** per doc. That description is what powers progressive-disclosure indexes and
search snippets — and better-separated concepts retrieve better. (Echoes the BINEVAL note: decompose, but
don't over-decompose.)

**4. It reinforces the "cheap index-map tier" idea — now from two directions.** The Chase AI review
already flagged a deterministic `index.md`-style map as a cheap complement to embedding retrieval. OKF
formalizes exactly that (the `description`-first index for progressive disclosure). Two independent
sources pointing at the same lever is a signal: a generated concept→description index over the corpus
could front-run the expensive RRF call for plain navigation. Worth a spike.

## Honest caveats

- **v0.1 is three weeks old** (June 2026) and adoption is unproven. The video's own verdict applies:
  *"until it becomes an open standard that agents support out of the box, this is more of an optimization
  than something you really need."* Even Claude didn't recognize OKF until a `CLAUDE.md` section taught it.
- **Don't touch the engine.** For the live query path the substrate's semantic retrieval is stronger than
  OKF navigation. OKF earns its place as a *boundary adapter* (export/import), not a core change.
- **Build it when there's a concrete need** — the reflect-steward wanting to publish a work pool to the
  team, or ingesting a partner's OKF catalog. Until then it's a clean, low-cost capability to keep on the
  shelf, not an urgent build.

## Verdict

OKF is a small, sane, vendor-neutral standard for knowledge *at rest*, and it slots cleanly against
pg-ai-stewards as a boundary format rather than a competitor. The one worth building, when a sharing or
ingest need is real, is the **export/import adapter**: it turns the substrate's compounding private
knowledge into a portable, standard-compliant artifact the wider agent ecosystem can read, and lets the
substrate ingest the OKF bundles that ecosystem is about to start producing — all for the price of
markdown plus YAML, with the DB-native engine untouched.
