---
title: Memory & Context Research Bundle
status: proposed
workstream: WS5 Memory & Process
created: 2026-04-22
source_brain_entries: ["a4eae47c", "17755019", "17756547"]
see_also_brain_entries: ["17750225", "17765208"]
binding_problem: We have multiple parallel memory/context research threads (mempalace, MetaClaw, continual learning, agentic AI memory frameworks, postgres-style memory). They all point at the same underlying problem — current agent memory systems don't fit our needs. Investigating each in isolation produces fragmented findings. A single bundled research pass produces a coherent direction.
---

# Memory & Context Research Bundle

## Binding Problem

Michael's stated direction: a postgres-style memory and context management system for brain + the VS Code harness. Multiple research threads have surfaced over the last several months, all pointing roughly the same direction:

- **mempalace** (https://github.com/milla-jovovich/mempalace) — memory framework
- **MetaClaw** (https://github.com/aiming-lab/MetaClaw) — memory issue solutions
- **Continual Learning for AI agents** (https://blog.langchain.com/continual-learning-for-ai-...) — LangChain piece
- **Agentic AI memory** (https://machinelearningmastery.com/7-steps-to-mastering-memory-in-agentic-ai-systems/) — overview
- **Practical guide to memory for autonomous LLM agents** (https://towardsdatascience.com/a-practical-guide-to-memory-for-autonomous-llm-agents/)
- (Possibly) proxy-pointer RAG and LightRAG — adjacent retrieval-architecture work tracked separately.

Doing 5 separate research spikes wastes effort and produces 5 disconnected memos. One pass that surveys all of them together, identifies the shared abstractions, and produces a single direction memo gets us to a designable v1 faster.

## Success Criteria

A single research document at `.spec/scratch/memory-research-bundle/findings.md` that:

1. Summarizes each source faithfully.
2. Identifies what each one solves and what it leaves on the table.
3. Maps onto our pain points (context window pressure, .mind/ memory architecture limits, brain entry recall).
4. Recommends a direction for our own memory layer — which patterns to adopt, which to invent, what postgres-backed shape this takes.
5. Output: either a "design v1" proposal OR an explicit "current architecture is good, defer this" verdict.

## Constraints

- Don't build during this phase. Research only.
- Single pass — resist scope creep into adjacent topics (pure RAG, pure embeddings).
- Output must connect back to brain.exe + .mind/ + the VS Code harness specifically.

## Related

- Adjacent to `lightrag-investigation` and `gospel-engine-v3-proxy-pointer` (RAG cousins, but distinct).
- The output may feed into `brain-vscode-bridge` Phase 3+.

## Phase 1

Research agent (Sonnet/Opus 4.7) reads all 5 sources + scans .mind/ + scans brain schema. Produces the findings doc. Estimate: one focused research session.
