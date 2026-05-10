---
name: council-moment
description: Three-minute connection scan + tension surface before substantive work. Codifies the CLAUDE.md mandate to "actively scan for connections to previous studies, tensions with existing work, and things the human might not be looking for" at session start. Load at the start of any substantive session, after grounding files but before diving into the task.
---

# Council Moment

> "And there stood one among them that was like unto God, and he said unto those who were with him: We will go down, for there is space there, and we will take of these materials, and we will make an earth whereon these may dwell." — [Abraham 3:24](gospel-library/eng/scriptures/pgp/abr/3.md)

> "Howbeit there were many among them who began to be exceedingly humble before God." — Helaman 4:25

The Gods took counsel before they acted. So should we.

## Why This Exists

Most failures in this project don't come from lack of capability. They come from acting before scanning. Section VII of the stewardship study contradicted Step 2 of the work-with-ai guide because the agent didn't check existing work. The `agent_commits_to.check_existing_work` covenant clause exists because of that incident.

The council moment is the inverse: a deliberate three minutes *before* substantive work where you scan for connections, tensions, and blind spots. Not exhaustive search — just enough to surface what would have been missed.

This is what the [Abraham 4:26 council](gospel-library/eng/scriptures/pgp/abr/4.md) looks like in practice. The Gods did not "go down" until they had counseled. We don't begin substantive work until we've counseled too — even if the counsel is just three minutes alone with the corpus.

## When to Load

Load at session start, after grounding files (intent.yaml, covenant.yaml, identity.md, active.md, principles.md) and before any substantive work begins. Specifically:

- Before drafting a study, lesson, talk, or teaching script
- Before designing or implementing a non-trivial feature
- Before evaluating a source, video, or proposal
- Before a multi-session phased task
- When the user gives you an open-ended task and you're tempted to start by writing

For a quick fix to a single typo, this is overkill. For anything that would touch 3+ files or take 30+ minutes, run it.

## The Three Scans

### Scan 1 — Connections (3 minutes)

What in our existing corpus connects to this task?

- `Grep` on key terms in `study/`, `lessons/`, `docs/`, `becoming/`, `.spec/proposals/`
- `mcp__gospel-engine-v2__gospel_search` on the binding question (semantic mode) — surfaces non-obvious cross-references in the canon
- `Glob` for related filenames in the project

**The bar:** name 2-3 specific files (or scripture references) that bear on this task. If you can't, you haven't scanned widely enough.

**What you're looking for:** the existing study that already answered half the question. The proposal that already debated this. The principle in `.mind/principles.md` that constrains the answer space.

### Scan 2 — Tensions (3 minutes)

What in our existing corpus *complicates* this task?

- A previous study that reaches an opposite conclusion
- A `.mind/decisions.md` entry that already ruled out an approach
- A constraint in the covenant or active.md that this task pushes against
- A failure mode logged in `.spec/learnings/` that this task could re-trigger

**The bar:** name at least one tension or surface "no significant tension found, here's why" in writing. Silence is not the same as no-tension.

**What you're looking for:** the place where the task's stated framing collides with something we already believe. Naming the collision *before* drafting prevents the contradiction from making it into the artifact.

### Scan 3 — Blind Spots (2 minutes)

What is the user *not* asking that they probably should be?

- An adjacent surface the task implies but doesn't name (the Adjacent Surface Audit at task-start, not just task-end)
- A spec gap — what does the user assume I'll cover that wasn't written?
- A discoverability question — will the user find what I built tomorrow?
- A contracts question — does the API/file/format actually carry what the work assumes?

**The bar:** name at least one thing the user didn't ask about that would change the work. Surface it as a question or as a noted assumption — don't silently expand scope.

## Output

Council moment outputs go in chat at the start of the session, before substantive work. Format:

```markdown
## Council moment

**Connections:** I found 2-3 specific files/refs that bear on this:
- [file:line or scripture ref] — [why it matters]
- ...

**Tensions:** [Either: "I found no significant tensions, here's why I checked" OR a list of specific tensions]

**Blind spots:** [The thing the user didn't ask but probably should — surfaced as a question or noted assumption]

Proceeding with [the actual task] now.
```

For longer or more substantive tasks (a phased study, a major proposal), write the council moment to a scratch file rather than chat. The file becomes part of the work's audit trail.

## What This Is NOT

- **Not exhaustive search.** Three minutes per scan, not thirty. The point is to surface what would have been missed, not to read everything.
- **Not theatrical.** If the scans return clean, say so plainly and proceed. Don't manufacture tensions to look thorough.
- **Not a delay tactic.** This is six minutes total. If it surfaces something real, those six minutes save hours. If it surfaces nothing, those six minutes earned the right to proceed confidently.

## The Stewardship Connection

Council moment is the entry point for stewardship. The covenant's `agent_commits_to.exercise_stewardship` clause says "owning the code includes keeping it sound." You can't steward what you haven't scanned. Council moment IS the scan.

The work that comes after is shaped by what the council surfaced. A study with no council moment may be brilliant in isolation but incoherent with the corpus. A study built on a council moment is a contribution to the body — "they were all one church" (Mosiah 25:22).
