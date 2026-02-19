---
name: expound
description: "Mine the current chat session for real-world teaching examples for the Working with AI lesson series"
agent: agent
tools: [read, edit, search]
---

I'm building a 3-part lesson series on working with AI effectively. The series exists in two versions — a secular version for software engineers and a gospel-centered version for YouTube.

**Read these docs to understand the framework:**

- [docs/work-with-ai/01_planning-then-create.md](../../docs/work-with-ai/01_planning-then-create.md) — Part 1: Spec-driven development
- [docs/work-with-ai/02_the-feedback-loop.md](../../docs/work-with-ai/02_the-feedback-loop.md) — Part 2: Review, diagnose, correct, verify
- [docs/work-with-ai/03_live-build.md](../../docs/work-with-ai/03_live-build.md) — Part 3: Applying the full pattern live
- [docs/work-with-ai/01_planning-then-create-gospel.md](../../docs/work-with-ai/01_planning-then-create-gospel.md) — Part 1 (gospel): The creation pattern from Abraham 4-5
- [docs/work-with-ai/02_watching-until-they-obey-gospel.md](../../docs/work-with-ai/02_watching-until-they-obey-gospel.md) — Part 2 (gospel): The feedback loop as divine pattern
- [docs/work-with-ai/03_intelligence-cleaveth-gospel.md](../../docs/work-with-ai/03_intelligence-cleaveth-gospel.md) — Part 3 (gospel): How what you bring shapes what emerges

**Now review this chat session's history** and extract:

### 1. Feedback Loop Examples
Identify moments where:
- I gave a correction and it worked (or didn't)
- The AI drifted from a spec or intent and I steered it back
- We went in circles and I had to change strategy
- A fix revealed a deeper issue

For each, capture:
- **What happened** (1-2 sentences)
- **The diagnosis** (spec gap, missing context, wrong approach, small bug)
- **The correction** (what I said or should have said)
- **The outcome** (did it stick? did it reveal more?)
- **Which lesson it fits** (Part 2 secular, Part 2 gospel, or both)

### 2. Planning Patterns
Identify moments where:
- Having a plan/spec prevented problems
- Lack of a plan caused problems
- We created a spec mid-session and it improved things
- The "spiritual creation before temporal" pattern was visible

For each, capture the same format as above, noting which Part 1 it fits.

### 3. Quality-of-Engagement Observations
Identify moments where:
- The quality of my question visibly affected the quality of the output
- Genuine curiosity led to unexpected depth
- Impatience or vagueness led to shallow results
- "Intelligence cleaveth unto intelligence" was observable in the interaction

For each, note what happened and how it illustrates D&C 88:40 (for the gospel version) or the "what you bring shapes what emerges" principle (for the secular version).

### 4. Trust Gradient Observations
Identify moments where:
- I started reviewing closely and gradually trusted more
- I trusted too quickly and it bit me
- I appropriately let the AI run in an area it had proven reliable
- The Abraham 4 progression (saw they obeyed → watched until → shall be very obedient) was visible

### 5. Novel Insights
Anything in this session that the existing lesson docs don't cover — a pattern, a failure mode, a success mode, a better analogy, a clearer way to explain something.

---

**Output format:**

Create a file at `docs/work-with-ai/examples/${input:date:YYYY-MM-DD}-${input:context:short-description}.md` with this structure:

```markdown
# Session Examples: {context}
**Applicable lessons:** {list of which lesson docs these examples strengthen}
**Date:** {date}
**Session type:** coding / study / mixed
**Tools used:** {list}

## Feedback Loop Examples
...

## Planning Patterns
...

## Quality-of-Engagement Observations
...

## Trust Gradient Observations
...

## Novel Insights
...

## Suggested Additions
For each example above, note specifically where it could be inserted into the existing lesson docs — which file, which section, and how it strengthens the teaching point.
```
