# Design history & provenance

`pg-ai-stewards` was built in the open, with a journal entry and a decision
record for nearly every session. This directory (plus the files it points to)
is the **internal provenance** — kept on purpose. It's heavy and reflects the
author's scripture-study workspace where the substrate grew up; read it for
*why* things are shaped the way they are, not as user-facing docs.

For using the substrate, start at [../../README.md](../../README.md) +
[../../QUICKSTART.md](../../QUICKSTART.md). For the runtime map, see
[../architecture.md](../architecture.md).

## The arc

- **[2026-05-02-research-verdict.md](2026-05-02-research-verdict.md)** — the
  original "should we build this, and on what?" research. The verdict (build a
  pgrx extension; don't fork; pair with pgvector + AGE) and every source behind
  it. This was the README before the substrate existed.
- **[../../proposal.md](../../proposal.md)** — what to build and why.
- **[../../phases.md](../../phases.md)** — the Phase 1 → F delivery plan
  (1500+ lines; carry-forwards now live in `.spec/open-items.md`).
- **[../../scratch.md](../../scratch.md)** — full source provenance for the
  research phase.

## Live design records

- **[../../.spec/proposals/](../../.spec/proposals/)** — 28+ design proposals,
  one per batch (phases C–F, the G–L batches, the ES emergency-stop arc, Council
  ① pipelines-expansion, and the standalone-extraction spec).
- **[../../.spec/journal/](../../.spec/journal/)** — per-session memory:
  discoveries, surprises, what shipped, carry-forwards.
- **[../../.spec/open-items.md](../../.spec/open-items.md)** — the live
  "what's next" queue.

## Other technical notes in docs/

- [../AGE-QUIRKS.md](../AGE-QUIRKS.md) — Apache AGE gotchas learned the hard way.
- [../3e-mcp-findings.md](../3e-mcp-findings.md) — MCP integration findings.
- [../lib-rs-refactor-findings.md](../lib-rs-refactor-findings.md) — the
  module-split lessons.
