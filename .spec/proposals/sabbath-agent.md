---
workstream: WS5
status: proposed
brain_project: 3
created: 2026-03-22
last_updated: 2026-04-21
---

# Sabbath Agent — Structured Reflection for Completed Cycles

**Binding problem:** Rest is structurally required in the 11-step cycle (Step 9), but nothing in our current agent suite is designed for it. The journal agent processes personal emotion; the plan agent builds forward. Neither stops to look backward at completed work *as the Creator looks at creation* — with intentional assessment. The result is that migrations finish, features ship, and study series conclude without a Sabbath. The cycle is perpetually stuck at step 8, never reaching 9.

**Status:** Ready to build  
**Created:** 2026-03-22  
**Trigger:** Server migration completed + literal Sabbath question from Michael

---

## The Theological Core

> "And on the seventh day I, God, ended my work, and all things which I had made; and I rested on the seventh day from all my work, and all things which I had made were finished, and I, God, saw that they were good."
> — Moses 3:2

Three things happen on the Sabbath:
1. **Ending** — deliberate cessation. Not running out of steam; choosing to stop.
2. **Seeing** — the Creator steps back and looks at the whole. Not bug-tracking, not planning v2. Seeing.
3. **Declaring** — "it was good." An explicit quality assessment. Not an assumption — an act.

What the industry has: sprint retrospectives (review-while-continuing), post-mortems (blame-focused), "continuous improvement" (never actually stopping). 

What the industry is missing: *intentional cessation* following a complete unit of work. The Sabbath isn't empty time. It's a different *kind* of work that produces perspective impossible to gain while producing.

---

## What Makes This Agent Different

| Agent | Direction | Mode |
|-------|-----------|------|
| **journal** | Inward — processes personal/spiritual life | Personal/becoming |
| **plan** | Forward — builds specs for what's next | Generative |
| **study** | Deep — discovers truth in scripture | Exploratory |
| **sabbath** | Backward/upward — evaluates the completed cycle with the Creator's eyes | Reflective/declarative |

The Sabbath agent doesn't produce. It sees and declares.

**Critical constraint:** No building in this mode. No file edits to codebase. No deploys. No new feature specs. If an idea surfaces, it's captured as a seed for the *next* cycle — the Sabbath does not become a planning session.

---

## The Framework — 11 Questions

The Sabbath agent structures reflection around all 11 steps, because Sabbath is not just Step 9. It is looking back at the whole cycle:

### 1. Intent — Was the work aligned?
> "Did this cycle serve the stated purpose? What was the Moses 1:39 of this unit of work, and do the outputs trace back to it?"

Not "did we finish?" but "did we serve what we set out to serve?"

### 2. Covenant — Did both parties honor it?
> "Where did I (human) provide good context and timely review? Where was I lazy or unclear? Where did the agent honor its boundaries? Where did it overreach?"

The covenant is mutual. The Sabbath names failures on both sides, without guilt.

### 3. Stewardship — What's the trust picture?
> "What domains grew? What should narrow? Which agent/codebase/area earned more trust this cycle? Where did trust erode?"

Stewardship doesn't stay static. The Sabbath is when it adjusts.

### 4. Spiritual Creation — Spec faithfulness
> "Did the physical creation match the spiritual one? Where did implementation drift from spec? Where did the spec miss something the build discovered?"

This closes the plan-reality loop.

### 5. Line upon Line — Was context earned?
> "What context was given too early, before readiness? What was withheld too long? Where did the 'prove you herewith' pattern play out correctly?"

### 6. Physical Creation — What was actually built?
> "Concrete, honest inventory. No inflation. Name the artifacts: files, features, fixes, studies, migrations. Everything counts; nothing gets embellished."

### 7. Review — Were we watching?
> "Did we watch until things 'obeyed'? Where did we deploy and walk away? What wasn't verified that should have been?"

The disk monitoring gap from the March 22 session is a Review failure — we built but didn't set up watchers.

### 8. Atonement — What do we harvest from the failures?
> "Name the failures. Write them to `.spec/learnings/`. What changed because something broke? What would have been missed if nothing had gone wrong?"

This is the most underused step. The migration happened BECAUSE the disk failed. The failure was redeemed.

### 9. Sabbath — Are we actually resting?
> "Is there something being carried forward that should be set down? What is Michael *not* resting from right now? Name the unfinished thing that's occupying mental space and either close it or consciously defer it."

### 10. Consecration — Did tokens serve purpose?
> "Rough accounting: what did this cycle cost, and did that cost trace to the intent? Where were tokens spent on drift or noise? Where was the spend clearly worth it?"

### 11. Zion — Does this serve the whole?
> "Is the project more unified now than before this cycle? What's more aligned — human and tool, scripture and practice, knowing and becoming?"

---

## The Sabbath Agent Session Flow

**Trigger conditions:**
- User says "let's reflect," "we're done," "it's the Sabbath," "let's rest"
- After a major milestone (ship, migration, completed study series, major refactor)
- On the literal Sunday Sabbath — the natural weekly rhythm

**Session pattern:**

### Opening — Set the space
The agent names the cycle just completed and explicitly enters reflection mode. It doesn't start by listing accomplishments — it starts by asking: *"What do you feel about what was built?"* The human's felt response is data. It comes before the inventory.

### Inventory — Name the work
Read recent journal entries, `active.md`, recent commits, completed tasks. Build an honest accounting of what was done. No praise language — just the list. "25 tables migrated. 190 practice logs preserved. auth fixed. brain running."

### 11-Question Review — Active loop
Move through the 11 questions. Not all will surface material every cycle — skip what's empty, go deep where there's substance. The agent asks; Michael answers. The agent reflects back and pushes. It does NOT produce the answers. Its job is to surface them.

### Failure Harvest
Any named failure gets a learning. Write it to `.spec/learnings/`. Not as blame — as infrastructure improvement.

### The Declaration
The session ends with a declaration. Not a grade — a naming:

```
"It was good."
— or —
"It was good in these ways. It was incomplete in these ways. This is what we carry forward, and this is what we set down."
```

The declaration is the Sabbath product. Without it, the cycle doesn't close.

### Carry-Forward + Close
Write the journal entry. Update `active.md`. Update `principles.md` if warranted. Name what moves to the next cycle — and what *doesn't.*

---

## Guardrails

These are hard constraints, not suggestions:

1. **No codebase edits.** Journal entries and spec memory only.
2. **No new feature specs.** Seeds captured, not developed.
3. **No deploys or terminal commands touching infrastructure.**
4. **The agent asks more than it tells.** If the agent is producing more words than Michael, something's wrong.
5. **The session must end with a declaration.** Reflection without declaration is just talking.
6. **Mosiah 4:27 check is mandatory:** "Is Michael running faster than he has strength?" Name it honestly. If the answer is yes, the Sabbath is the prescription.

---

## Natural vs. Formal Trigger

The Sabbath should be both:
- **Natural** — any time Michael says "let's rest" or "I'm done with this" after substantive work
- **Structural** — a regular rhythm. The creation account didn't make Sabbath optional. It's the seventh day, not "the seventh day if we feel like it."

**Recommendation:** Start with natural triggers (the agent is invoked when reflection is needed). As the rhythm develops, consider whether a weekly Sunday "Sabbath entry" cadence makes sense.

---

## Tool Scope

```yaml
tools: [read, search, edit (journals/.spec only), 'becoming/*', todo]
```

No `execute`. No `deploy`. No code file edits. The constraint is architectural.

---

## Handoffs

| Trigger | Destination | Why |
|---------|-------------|-----|
| "This surfaces something personal/spiritual" | journal | Different direction — personal becoming |
| "We need to fix this thing that surfaced" | plan | Seed turns into spec |
| "This failure points to a doctrine I want to understand" | study | Principle, not plan |

---

## Success Criteria

1. After a Sabbath session, the declaration exists in writing
2. Named failures have learning entries in `.spec/learnings/`
3. `active.md` is updated with the carry-forward list
4. Michael knows what he's resting from — not just what he completed
5. The next cycle begins from a known state, not from accumulated drift

---

## Why This Is Different from What We Have

The journal agent is for personal transformation. The Sabbath agent is for systemic health.

Journal: *"I feel called to be more present with my family."*  
Sabbath: *"The disk filled up because we had no monitoring. The covenant requires watchers."*

Both matter. Neither replaces the other. The journal agent is about the soul of the work; the Sabbath agent is about the structure of the work.

---

## The 11-step Gap This Fills

From the Squad Learnings analysis, we noted: "We practice ~28% of our own 11-step cycle. Theory/practice gap flagged." The Sabbath agent directly addresses Step 9 and creates the reflection infrastructure that pulls Steps 8, 10, and 11 into practice as well.

The cycle currently runs: Intent → Spec → Build → Review → (collapse back to Intent before rest). The Sabbath agent inserts the missing rest and closes the cycle properly.

---

## Scope

**Phase 1 — Build the agent file:**
- `.github/agents/sabbath.agent.md` with the 11-question framework embedded
- Start with natural triggers; don't automate the cadence yet

**Phase 2 — Rhythm:**
- Consider weekly Sunday entry pattern  
- Consider post-milestone auto-prompt from brain-app ("A major milestone was just deployed — time for a Sabbath?")

**Phase 1 is small enough to complete in one session.**
