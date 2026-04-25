---
name: study-plan
description: "Plan a multi-phase scripture study — like sprint planning for deep, interconnected topics that need multiple sessions"
agent: study
argument-hint: "[broad topic or multi-part question]"
tools: [read, edit, search, "gospel-engine-v2/*", "webster/*", "search/*"]
---

Plan a multi-phase scripture study on the given topic.

## Philosophy

This is the spec-driven approach from our [Working with AI series](../../docs/work-with-ai/01_planning-then-create.md) applied to scripture study. Some topics are too big for one session: the Plan of Salvation, the Abrahamic Covenant, the nature of God across dispensations. These need a **plan** — a blueprint that breaks the study into focused phases, each building on the last.

**This is NOT a replacement for one-shot studies.** One-shot (`/new-study`) works great for focused topics. Use this when:
- The topic spans multiple dispensations or books
- You want to telescope in (deep reading) AND widen out (semantic search) deliberately
- Previous one-shot studies have surfaced threads you want to weave together
- The topic is too big to hold in one context window

## Step 1: Envision

Before creating any files, explore the topic with the user:
- What prompted this study? What's the driving question?
- What do we already know? (Check `study/*.md` for existing related studies)
- What are the sub-questions or facets?
- What would "done" look like — a single comprehensive document? Multiple linked documents? A synthesis that connects existing studies?

## Step 2: Create the Plan

Create a directory and plan file at `study/${input:slug:topic-slug}/00_plan.md`:

```markdown
# Study Plan: ${input:topic:Study Topic}

*Created: ${input:date:YYYY-MM-DD}*

## Driving Question
<!-- The one question this whole study is trying to answer -->

## What We Already Know
<!-- Links to existing studies that touch this topic -->

## Phases

### Phase 1: [Deep Reading — Core Text(s)]
**Skill:** `deep-reading`
**Focus:** [Specific passages to telescope into]
**Output:** Intermediate findings in `notes-01.md`
**Questions to answer:**
- [Specific questions for this phase]

### Phase 2: [Deep Reading or Wide Search — Second Focus]
**Skill:** `deep-reading` or `wide-search`
**Focus:** [What to explore]
**Output:** Intermediate findings in `notes-02.md`
**Questions to answer:**
- [Specific questions for this phase]

### Phase 3: [Wide Search — Cross-Library Connections]
**Skill:** `wide-search`
**Focus:** Threads from Phases 1-2 searched broadly
**Output:** Connection findings in `notes-03.md`

### Phase 4: Synthesis
**Output:** Final study document(s) — either one comprehensive doc or multiple linked docs
**Structure:** [Outline of the final document]

### Phase 5: Becoming
**Skill:** `becoming`
**Output:** Personal application section in the final doc + becoming/ companion if warranted

## Status

| Phase | Status | Date | Notes |
|-------|--------|------|-------|
| Phase 1 | ❌ Not started | | |
| Phase 2 | ❌ Not started | | |
| Phase 3 | ❌ Not started | | |
| Phase 4 | ❌ Not started | | |
| Phase 5 | ❌ Not started | | |
```

## Step 3: Begin Phase 1

After the plan is reviewed and approved by the user, start Phase 1 using the appropriate skill (`deep-reading` or `wide-search`). Create the intermediate notes file and begin working.

## Principles

- **Each phase is one session.** Don't try to do everything at once. Commit findings, update the plan status, and hand off clearly.
- **Intermediate notes are the workbench.** They don't need to be polished — they need to be useful for the next phase.
- **The plan is alive.** As phases complete, new questions will emerge. Update the plan with new phases or reorder existing ones.
- **Existing studies are assets.** A phased study might connect 3-4 existing one-shot studies into something larger. Don't redo work — link to it and build on it.
- **Becoming is not optional.** Phase 5 always exists. It might be brief, but it's always there.
