---
title: LightRAG — Investigation for Gospel Engine and Beyond
status: proposed
workstream: WS3 Gospel Engine
created: 2026-04-22
source_brain_entries: ["17751939", "17751941"]
binding_problem: LightRAG is a well-regarded graph-augmented RAG framework. We want to evaluate whether it could replace or augment gospel-engine-v2's current FTS+semantic retrieval — and whether it has carryover value for non-gospel work too.
---

# LightRAG — Investigation

## Binding Problem

Sources:
- https://github.com/HKUDS/LightRAG
- https://lightrag.github.io/
- https://youtu.be/QHlB-RJfx8w

Gospel-engine-v2 currently uses FTS5 + chromem-go semantic search combined. LightRAG offers a graph-augmented approach that could improve cross-reference discovery — exactly the thing scripture study most needs.

Open questions: does it actually help on our corpus shape (highly cross-referenced, hierarchical scripture text)? Is it operationally cheap enough to run? Does it have carryover for non-gospel work like memory-research-bundle?

## Success Criteria

- A research-spike doc in scripts/gospel-engine-v2/research/ that:
  1. Summarizes LightRAG faithfully.
  2. Tests it on a slice of our corpus (e.g., one volume of scripture).
  3. Compares retrieval quality against current gospel-engine-v2 on a fixed query set.
  4. Names the decision: adopt / hybridize / pass.

## Related

- Pairs with `gospel-engine-v3-proxy-pointer` — both are RAG-architecture investigations for the same engine. May converge into one v3 proposal.
- Pairs with `memory-research-bundle` — graph approaches to retrieval and memory are kissing cousins.

## Phase 1

Stand it up locally on a slice. Run the same queries against gospel-engine-v2 and against LightRAG. Compare side-by-side.
