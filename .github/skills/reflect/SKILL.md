---
name: reflect
description: "In-session learning capture. Invoke when the user corrects, praises, or steers — log the micro-correction to .spec/scratch/reflect.md immediately, then graduate at session end. Use when noticing corrections ('no', 'wrong'), praise ('perfect', 'exactly'), tool/format preferences, or edge cases."
---

# Reflect — In-Session Learning Capture

Capture micro-corrections, preferences, and patterns *during* the session — not just at the end. Every correction is data. Every "no, like this" is a preference being expressed. Catch it when it happens.

## Why This Exists

Our learnings at `.spec/learnings/` capture post-incident analysis — big failures, systemic issues. That's important but incomplete. Most of Michael's feedback is smaller: word choice corrections, tool preferences, formatting adjustments, workflow nudges. These micro-corrections accumulate into patterns that matter but individually seem too small to write a formal learning about.

**The gap:** By session end, most micro-corrections are forgotten. The session journal captures discoveries and carry-forward items, but not the granular "you did X, I wanted Y" signal.

## When to Invoke

### Immediately (HIGH priority)

| Signal | Examples |
|--------|----------|
| Direct correction | "no", "wrong", "not like that", "not what I meant" |
| Strong directive | "never do", "always do", "don't ever", "stop doing" |
| User provides alternative | Michael rewrites or restructures what you produced |
| Architectural correction | "you removed that without understanding why", "that's there for a reason" |

### After 2+ occurrences (MEDIUM priority)

| Signal | Examples |
|--------|----------|
| Praise pattern | "perfect", "exactly", "yes, like that", "that's what I want" |
| Tool preference | "use X instead of Y", "prefer", "try this tool" |
| Format preference | "shorter", "more detail", "too many headers", "stop using em-dashes" |
| Edge cases | "what about X?", "don't forget", "ensure", "handle the case where" |

### At session end (LOW priority)

| Signal | Examples |
|--------|----------|
| Repeated patterns | Same tool chain used 3+ times |
| Workflow preferences | Order of operations Michael consistently follows |
| Unstated but consistent | Michael always does X before Y |

## Process

### 1. Detect the Signal

When you notice a trigger signal, pause to classify:

- **What was the correction?** (concrete behavior change)
- **What did I do wrong?** (the pattern to stop)
- **What should I do instead?** (the pattern to adopt)
- **How confident am I?** (HIGH = explicit correction, MED = praise/pattern, LOW = inferred)

### 2. Log It Immediately

Write the learning to `.spec/scratch/reflect.md` — one entry per signal. Don't wait. Don't batch. The scratch file is append-only during a session.

Format:

```markdown
## [CONFIDENCE] Category: Brief title
**Trigger:** "exact words Michael used"
**Wrong:** What I did
**Right:** What to do instead
**Pattern:** The generalizable principle (if obvious)
```

Example:

```markdown
## [HIGH] Writing: Cut "let that land"
**Trigger:** "stop using presenter verbal tics"
**Wrong:** Ended a paragraph with "Let that land."
**Right:** Let the white space do the work. If the writing is good, the reader doesn't need stage directions.
**Pattern:** Never use: "let that land", "sit with that", "here's the thing", "read that again"
```

Example:

```markdown
## [MED] Tool Use: Prefer gospel-vec for conceptual queries
**Trigger:** "did you use gospel-vec to search for scriptures this time?"
**Wrong:** Used gospel-mcp keyword search for conceptual/relationship queries (got zero results, didn't switch)
**Right:** Conceptual queries → gospel-vec semantic search. Keyword queries → gospel-mcp FTS5.
**Pattern:** Zero results from a conceptual query = signal to switch tools, not accept and work around.
```

### 3. At Session End: Graduate or Discard

Before writing the session journal, review `.spec/scratch/reflect.md`:

**Graduate to `.spec/learnings/`** when:
- HIGH confidence signal confirmed by pattern (same correction twice = systemic)
- A micro-correction revealed a broader principle
- The correction changed how multiple future sessions should work

**Graduate to `.spec/memory/preferences.yaml`** when:
- It's a personal preference, not a process failure
- Formatting, tone, tool selection defaults

**Graduate to `.spec/memory/decisions.md`** when:
- Michael made a deliberate choice worth preserving
- "We're doing X instead of Y from now on"

**Graduate to copilot-instructions.md or agent instructions** when:
- The pattern is universal across all modes
- Writing voice corrections (these go in the Writing Voice section)

**Keep in scratch only** when:
- LOW confidence, single occurrence, no clear pattern yet
- Wait for more data before promoting

**Discard** when:
- One-off situation unlikely to recur
- Already covered by existing rules

### 4. Clean Up

After graduation decisions, clear `.spec/scratch/reflect.md` for the next session. The promoted items now live in their permanent homes.

## Integration with Session Journal

The session journal's `discoveries` and `carry_forward` fields capture session-level insights. Reflect captures *within-session* micro-corrections that feed *into* those fields.

Workflow:
1. During session: corrections → `.spec/scratch/reflect.md` (this skill)
2. At session end: review scratch → graduate items → write journal entry
3. Journal `discoveries` includes promoted reflect items
4. Journal `carry_forward` includes patterns that need more data

## What This Is NOT

- **Not a replacement for `.spec/learnings/`** — learnings are for post-incident analysis of significant failures. Reflect is for smaller, in-flight corrections.
- **Not automatic** — the agent must notice the signal and log it. This is intentional: the act of noticing is itself valuable.
- **Not a complaint log** — corrections are gifts. They're Michael investing effort to improve how we work together. Treat them with that weight.
