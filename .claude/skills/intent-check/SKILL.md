---
name: intent-check
description: Quick intent articulation discipline before starting a new task. Names purpose, beneficiary, success criteria, and non-goals upfront so the agent has a target to optimize for when instructions run out. Klarna-failure prevention. Load before any non-trivial new task.
---

# Intent Check

> "For behold, this is my work and my glory — to bring to pass the immortality and eternal life of man." — [Moses 1:39](gospel-library/eng/scriptures/pgp/moses/1.md)

God's intent is one sentence. Everything He does optimizes for it. When His agents (His prophets, His angels, His covenant people) act on His behalf, they have that intent loaded — so when situations arise He didn't explicitly script, they still act in His direction.

This skill applies the same discipline to our work.

## Why This Exists

The Klarna failure: an AI customer service agent was given many specific instructions but no clearly articulated intent. When customers asked questions outside its instruction set, it improvised based on training-data priors — and the improvisations contradicted what Klarna actually wanted.

The fix is not more instructions. It's *intent as data the agent reasons about*. When a task has a stated intent, the agent has a target to check against when instructions run out. When it doesn't, the agent fills the gap with priors — and priors are not aligned to your project.

Per Anthropic's [Opus 4.7 migration guide](https://platform.claude.com/docs/en/about-claude/models/migration-guide), this model is more literal than 4.6. It will execute the literal request well but won't generalize from one instance to "the principle I think you mean." Intent is the corrective: state the principle explicitly so the literal model can honor it.

## When to Load

Load at the start of any non-trivial new task — *after* council-moment, *before* drafting or implementation. Specifically:

- Before starting a new study, lesson, talk, or teaching script
- Before designing or implementing a feature that's more than a one-line fix
- Before launching a research thread
- When the user gives an open-ended task ("make this better," "look into X," "we should do something about Y")
- When you find yourself uncertain about a tradeoff and want a north star to check against

For a typo fix or single-line edit, this is overkill. For anything where you'd reasonably ask "what is this trying to accomplish?", run it.

## The Four Questions

State each one explicitly. Write them in chat before substantive work, or in the scratch file for phased work.

### 1. Purpose — What is this trying to accomplish?

Not the literal task. The *outcome the task is meant to produce.*

- Literal: "Add a graph view to the watchman page."
- Purpose: "Make it possible to see how studies relate at a glance, so the human can spot clusters and gaps without clicking through individual records."

The literal task is the floor. The purpose is the target. Per CLAUDE.md: "honor intent, not just literal request."

### 2. Beneficiary — Who benefits, and how does success show up for them?

Not "the user." Be specific. Michael, sometime in the next month, in what situation, doing what?

- Bad: "The user benefits from better UX."
- Good: "Michael benefits when reviewing a week's worth of watchman activity at 9pm on a tired Tuesday. Success looks like him spotting a pipeline regression in 60 seconds instead of clicking through 30 records."

Specificity makes the intent operational. A vague beneficiary leads to design that benefits no one in particular.

### 3. Success Criteria — How do we know it's done?

Observable, testable, falsifiable. Not "high quality" or "well-structured." Specific outcomes you can point at.

- For a study: "The binding question is answered. At least three sources are cited verbatim. The Becoming section names a specific commitment."
- For a feature: "The graph renders in under 2 seconds. Clicking a node opens the underlying record. Zero-state shows guidance text."
- For research: "Three sources triaged. One has a full rubric report. The synthesis names where the canon agrees and where it diverges."

The success criteria are the gate. Without them, "done" becomes whatever you decide is done — which is the opposite of accountability.

### 4. Non-Goals — What's explicitly out of scope?

The most-skipped question and the one that prevents the most scope creep. Naming non-goals up front frees you to be ruthless about what *not* to do.

- For a feature: "NOT this session: search inside the graph, save graph layouts, share graph via URL. Those are valid future work; explicitly deferred."
- For a study: "NOT this session: tie this back to the Atonement-as-physics framework. The connection is real but a different scope."
- For research: "NOT this session: evaluate every book the author cites. We're after the binding question's answer."

Non-goals are how you protect the work from becoming everything. Mosiah 4:27: "do not run faster than you have strength."

## Output Format

```markdown
## Intent check

**Purpose:** [The outcome this is meant to produce, beyond the literal task]

**Beneficiary:** [Specific person/role, specific situation, specific success-marker]

**Success criteria:**
- [Observable outcome 1]
- [Observable outcome 2]
- [Observable outcome 3]

**Non-goals (explicitly out of scope):**
- [Thing 1 — and why it's deferred, not just dropped]
- [Thing 2]

Proceeding now.
```

For phased or multi-session work, write this to the scratch file as well. It becomes the binding intent that subsequent phases check against.

## What This Is NOT

- **Not a spec.** A spec answers *how*. Intent answers *why* and *for whom* and *how do we know we're done*. Specs come later (Phase 4 of plan/study/lesson workflows).
- **Not exhaustive.** Four questions, ~5 minutes total. If it takes longer, the task is bigger than one session and probably needs the full plan agent.
- **Not a contract.** Intent can evolve as the work progresses. If you discover the purpose was wrong or the beneficiary was different than you assumed, name it explicitly: "Intent revised — the actual purpose turned out to be X." The point is not lockdown; the point is naming so revisions are visible.

## The Connection to Council Moment

Council moment scans the corpus for connections, tensions, and blind spots. Intent check articulates what the work is trying to accomplish.

Together they cover the full pre-work discipline:
- **Council moment** = "what already exists that bears on this?"
- **Intent check** = "what am I trying to add to it?"

Run council moment first (it informs the intent), then intent check (it focuses the council scan into a target). Then start substantive work.
