---
title: Gospel Engine v3 — Proxy-Pointer RAG Investigation
status: proposed
workstream: WS3 Gospel Engine
created: 2026-04-22
source_brain_entry: 17767284
binding_problem: A new RAG approach (proxy-pointer) claims 100% accuracy with smarter retrieval. We're already doing some of this implicitly in gospel-engine-v2. Question is whether v3 should move toward this pattern explicitly, or if we already have enough of it.
---

# Gospel Engine v3 — Proxy-Pointer RAG Investigation

## Binding Problem

Article: https://towardsdatascience.com/proxy-pointer-rag-structure-meets-scale-100-accuracy-with-smarter-retrieval/

Glancing over it, our current FTS + semantic combined search shares some DNA with proxy-pointer. Before designing a v3, we want to know: are we already doing this, or is there a structural change worth making?

## Success Criteria

- A research-spike document in scripts/gospel-engine-v2/research/ that:
  1. Summarizes proxy-pointer RAG faithfully.
  2. Maps it against gospel-engine-v2's current architecture.
  3. Names what (if anything) v3 should adopt.
- Decision: continue iterating on v2, OR draft a v3 architecture proposal.

## Approach

Use Sonnet/Opus 4.7 research agent on the article. Cross-reference with our existing gospel-engine-v2 search code.

## Related

- Pairs with `lightrag-investigation` — both are RAG architecture questions for the same engine.
- May fold into a single "gospel-engine-v3-architecture" proposal if both directions converge.

## Phase 1

Research agent reads the article + scans our FTS/semantic code paths in scripts/gospel-engine-v2. Produces evaluation memo.
