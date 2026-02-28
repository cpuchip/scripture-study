# Proposal: Next Steps After the Guide

*February 28, 2026*
*Status: Proposal — ideas to plan out, not a development plan*
*Context: After writing the 7-part Working with AI guide, verifying all sources, discovering the confabulation problem, refining our skills, and building the dev-plan for Tier 1.*

---

## Why This Document Exists

The dev-plan ([dev-plan.md](guide/dev-plan.md)) covers Tier 1 in detail (verify-quotes, learnings pipeline, intent YAML). But the session also surfaced several bigger ideas that will evaporate if they stay in a context window. This doc captures them as proposals — not planned, not prioritized, just *recorded* so they can be picked up later.

---

## Proposal 1: Creative / Synthesis Agent Mode

### The Problem

We used the `study` agent to write the 7-part guide series. But the guide isn't a study document — it's a *teaching document* written for a general audience. Study mode optimizes for personal depth, footnote-following, and Becoming sections. The guide needed:

- Clear explanations for readers who don't share our context
- Analogies that bridge gospel concepts to engineering concepts
- Structured progression across 7 documents
- Source verification at a different level (external sources, YouTube transcripts, not just scriptures)
- No "Becoming" section (the guide IS the becoming — it teaches others)

### The Idea

A `creative` or `synthesis` agent mode designed for producing outward-facing content:

| Aspect | Study Agent | Creative Agent |
|--------|-------------|----------------|
| **Audience** | Self (+ future self) | Others — readers, learners, practitioners |
| **Tone** | Personal exploration, warmth | Clear teaching, accessible, still warm |
| **Structure** | Follow the text where it leads | Planned structure across multiple documents |
| **Sources** | Scriptures, talks, Webster | Scriptures + external (videos, articles, frameworks, tools) |
| **Verification** | cite-count rule, read-before-quoting | Same, PLUS external source verification (transcripts, URLs) |
| **Output** | Single study document | Multi-document series, guides, tutorials |
| **Ending** | Becoming (personal) | Call to action (reader-facing) |

**Potential scope:**
- Guide writing (like the Working with AI series)
- Tutorial creation (how to set up MCP servers, how to use intent YAML)
- Blog post drafting from existing studies
- Podcast script generation (currently in `podcast` agent, but that's narrower)

**Key question:** Is this a new agent, or an extension of `podcast`? The podcast agent already has the "outward-facing" orientation. Maybe it's `publish` or `teach` rather than `creative`.

### Handoffs

- `study` → `creative`: "I've done the research, now help me teach it"
- `creative` → `study`: "I need deeper research on this subtopic before I can explain it"
- `creative` → `dev`: "This section describes a tool — let's build it"

---

## Proposal 2: Repo Organization

### The Problem

The repo has grown organically over 2+ months of daily use. It's rich and alive — but it's also a maze. Some symptoms:

- `docs/` has 15+ numbered files (reflections, observations, skill-gaps, intent-development) with no clear taxonomy
- `docs/work-with-ai/` has loose files (01-04), plus `guide/`, `intent/`, `prompt/`, `examples/`
- `scripts/plans/` has architecture docs that are really specs, not scripts
- Study documents have no lifecycle tracking (some are drafts, some are definitive, some are superseded)
- The relationship between files is implicit (you have to read them to know they connect)

This was identified in [10_intent-development.md](../10_intent-development.md) as the "Markdown Flood" problem. The ideas there (document registry, taxonomy) are still valid.

### Constraints

**The publish script creates a public-facing URL structure.** People have shared links like:
- `<site>/study/enoch.html`
- `<site>/study/charity.html`
- `<site>/lessons/cfm/...`
- `<site>/docs/work-with-ai/guide/...`

Any reorganization must either:
1. Keep the public paths stable (move source files but adjust publish script to output to the same paths), OR
2. Generate redirects from old paths to new paths

Option 1 is simpler. The publish script already decouples source paths from output paths — we'd just need to update mappings.

### Ideas (Not Decisions)

**A. Flat docs → categorized docs**
```
docs/
  meta/           ← process reflections, methodology, biases
  templates/      ← study, lesson, talk, yt_evaluation templates
  marketing/      ← (already exists)
  work-with-ai/   ← (already exists, keep)
```

**B. Plans → .spec/**
Move `scripts/plans/` content to `.spec/` alongside the new `learnings/` directory. Plans *are* specs — they just weren't called that.

**C. Document frontmatter**
Add YAML frontmatter to key documents:
```yaml
---
title: Enoch and the City of Zion
type: study
status: active        # draft | active | superseded | archived
created: 2026-01-15
updated: 2026-02-10
connects_to:
  - study/charity.md
  - becoming/charity.md
intent: "Understand how Enoch's pattern of walking with God applies to personal discipleship"
---
```

This is the lightest-weight version of doc 10's "document registry" idea — metadata lives in the file, not a separate index.

**D. Auto-generated index**
A script that scans frontmatter across all documents and generates a browsable index (for the public site and for agent consumption). This is the "document registry" from doc 10 but automated.

**E. Study lifecycle**
Some studies are explorations (first pass). Some are definitive (thoroughly researched, verified, published). Some are superseded (a later study covers the same ground better). Making this explicit helps both humans and agents know which to trust.

### Relationship to OpenSpec

The interest in OpenSpec or similar frameworks is about formalizing what we're already doing informally. Our "spec" process today:
1. Write a markdown doc in `docs/` or `scripts/plans/`
2. Discuss it in chat
3. Start building
4. Sometimes update the doc, sometimes not

A formal spec approach would add: version tracking, explicit status, decision records, and machine-readable structure. The `intent.yaml` we just created is a step in this direction. The `.spec/learnings/` pipeline is another step. But the full framework is a bigger conversation.

---

## Proposal 3: Tier 2 — ibecome Improvements

Captured from dev-plan, expanded here for future planning:

### Tasks Overhaul
The Tasks view is the thinnest part of the app (138 lines). It's the bridge between study insights and daily action, and it needs:
- Edit capability (currently create/toggle/delete only)
- Due dates and recurrence
- Pillar linking (tasks connect to growth areas)
- Notes integration (attach insights to tasks)
- Priority/sorting
- Source document linking (which study generated this task?)

### Study Mode in Nav
StudyView (873 lines, adaptive memorization exercises) is fully built but not in the navigation. Adding it to the nav bar is a one-line change but needs UX consideration — where in the nav order? What's the entry flow?

### PWA Support
For a daily practice tool, mobile access matters. Progressive Web App with:
- Install prompt (add to home screen)
- Service worker for offline access to today's practices and due cards
- Future: push notifications for due memorization cards
- Future: background sync for offline logs

### Data Export
Privacy policy promises data portability. Add export button in Settings → download all user data as JSON.

### Dark Mode Cleanup
115 lines of `!important` CSS overrides in App.vue. Migrate to Tailwind v4's proper dark mode support.

---

## Proposal 4: Tier 3 — Intent Architecture

The big ideas from the guide that nobody else is building:

### Covenant Blocks
Agent configs that include mutual commitments — not just "follow these rules" but "I commit to X, you commit to Y." The existing copilot-instructions already gesture at this ("a collaboration between a human who brings faith, agency, and the Spirit, and an AI that brings processing capacity..."). Making it structural:

```yaml
# In agent config or intent.yaml
covenant:
  human_commits:
    - "Bring genuine curiosity, not just task execution"
    - "Trust the discernment — if it doesn't feel right, say so"
    - "Verify what matters — don't blindly accept output"
  agent_commits:
    - "Read before quoting — never write from memory"
    - "Follow the text where it leads, not just the template"
    - "Stay present — warmth over clinical distance"
  mutual:
    - "Honest accounting of what went wrong (biases.md, learnings/)"
    - "The system improves through use, not despite it"
```

### Progressive Trust Engine
Agents earn expanded scope through demonstrated faithfulness. Start simple:
- Track success/failure per task type per agent session
- After N successful completions in a domain, expand the agent's default context for that domain
- After a failure, narrow scope and require more explicit verification

This is the Parable of the Talents as an algorithm.

### Spec Directory Workflow
Formalize `.spec/` as the spiritual creation layer:
```
.spec/
  intent.yaml          ← root intent (already created)
  learnings/           ← error→growth pipeline (already started)
  decisions/           ← architectural decision records
  reviews/             ← intent-alignment reviews of completed work
  proposals/           ← this document would live here
```

### Agent Config Generation from Intent
Given an intent YAML with values, constraints, and success criteria, generate the agent instructions that implement it. Currently the agent markdown files are hand-crafted. If intent were structured, agent behavior could be derived — and verified against intent.

---

## Proposal 5: webecome (Future)

Multi-user version of ibecome. The same study→becoming→practice→reflection cycle, but with:
- Shared accountability groups (study groups, quorum groups)
- Shared document sources (a ward shares their Come Follow Me studies)
- Aggregated anonymized insights ("42 people are studying Alma 32 this week")
- Mentor/mentee relationships (a parent tracks with a child, a bishop supports a missionary)

This is Zion-level tooling — "one heart and one mind" applied to personal development. It's the biggest idea and the furthest out. But it's worth recording because it connects the personal tool (ibecome) to the community vision (webecome) to the organizational pattern (the guide's enterprise architecture).

The hostname-dependent branding already exists in the frontend (`I Become` vs `We Become`). The multi-user database schema is already in place (`user_id` on every table). The infrastructure is closer than it looks.

---

## The Meta-Observation

This repo is messy because it's *alive*. Study documents, tool code, personal journals, teaching guides, app code, process reflections, and meta-methodology all live side by side because they all emerged from the same practice: sitting down to study, discovering something, building a tool to help, reflecting on what happened, and doing it again.

The mess is the evidence that the cycle works. The organization challenge is real — but the goal isn't to eliminate the organic growth. It's to give it enough structure that we (human + AI) can navigate it, and that new things can find their place without someone having to reorganize everything.

Document frontmatter + auto-generated index is probably the right first move. It adds discoverability without forcing a reorganization. The reorganization can happen incrementally as documents naturally get touched.

---

*This is a proposal document. Nothing here is committed. Pick it up when the energy is right.*
