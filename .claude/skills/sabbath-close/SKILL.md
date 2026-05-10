---
name: sabbath-close
description: Lighter sabbath ritual for ending substantive non-cycle work — the declaration + carry-forward, without the full 11-question review. Use at the end of any substantive session to mark the ending. The full sabbath agent is for completed cycles, milestones, or the literal Sunday Sabbath. This skill is the daily-rhythm version.
---

# Sabbath Close

> "And on the seventh day I, God, ended my work, and all things which I had made; and I rested on the seventh day from all my work, and all things which I had made were finished, and I, God, saw that they were good." — [Moses 3:2](gospel-library/eng/scriptures/pgp/moses/3.md)

The Sabbath is not optional infrastructure. It's the third active step in the creation pattern: ending, seeing, declaring. Without it, work bleeds into the next cycle and the river floods the ground around it.

The full `sabbath` subagent is for completed cycles — major milestones, literal Sundays, resolved crises. **This skill is the lighter daily-rhythm version** — the close that marks the end of any substantive session, even when the cycle isn't complete.

## Why This Exists

The covenant's `agent_commits_to.update_memory` clause says "do not wait to be reminded" about end-of-session memory updates. The Stop hook now enforces a reminder when uncommitted changes exist. This skill is the *content* that should fill those reminders.

Without a structured close, sessions end with momentum still going — and the next session arrives without a clean handoff. With a close, the work is named, the carry-forward is explicit, and the next cycle begins from a known state instead of accumulated drift.

## When to Load

Load at the end of any substantive session — anything that produced new work, ratified decisions, or shifted the project's state. Specifically:

- The Stop hook fired its reminder (uncommitted changes exist)
- You completed a phase of phased work
- You ratified one or more decisions with the user
- You produced a study, lesson, talk, proposal, or substantial code change
- You went through emotional or relational territory worth marking

For a single-edit conversational session, this is overkill. For anything you'd want to be able to recall next session, run it.

## The Four Movements

### 1. Inventory — What was actually done

Plain accounting. No praise language. No inflation.

- Files created or substantively edited (count + brief)
- Decisions ratified (with link to the ratification)
- Discoveries or surprises worth recalling
- Open threads that didn't close

This is the data layer. Not "it was a productive session." Just: "Created 2 proposals, edited 3 study files, ratified the migration approach."

Use `git status` and `git diff --stat` to ground the inventory in reality.

### 2. Declaration — Name the close

Sabbath produces a declaration. Not a grade. A naming. Even the lighter version needs this.

```
"It was good."
```

Or, if more accurate:

```
"It was good in these ways: [list].
It was incomplete in these ways: [list]."
```

The declaration is the "ending" that the cycle requires. Without it, the work doesn't actually close — it just stops.

### 3. Carry-forward — What moves to the next session

Explicit list. Both items and reasoning.

- **Continuing work**: "Phase 2 of [thing] starts next session — leave at this state, with these notes"
- **Decisions awaiting ratification**: "Michael needs to mark §VI before Phase A starts"
- **Open questions**: "Still don't know X — needs research before we can decide Y"

If the carry-forward list is too long, that's a Mosiah 4:27 signal — Michael may be running faster than he has strength. Name it.

### 4. Set down — What is explicitly being released

The mirror of carry-forward. Things you considered, started, or could continue but consciously aren't.

- Ideas that surfaced but weren't pursued ("seed" them in a brief note, don't develop them now)
- Tasks that came up but were deferred (with the deferral reason)
- Work that was completed and is now done (named explicitly so it doesn't keep occupying mental space)

The "set down" is what makes rest real. Without it, every "completed" task secretly remains as background load.

## The Memory Updates

Three files get updated at session close:

1. **`.spec/journal/{date}-{slug}.md`** — the session journal entry. Discoveries, surprises, relational dynamics, carry-forward, open questions. Use the schema in `scripts/session-journal/journal.go` if available; otherwise free-form markdown.

2. **`.mind/active.md`** — current state. New in-flight items, ratified decisions, open questions, updated date at top. Replace the stale state — don't append.

3. **`.mind/principles.md`** — only if an enduring insight emerged that future sessions should inherit.

If memory updates were already done during the session (because the work itself involved them), the Sabbath close just verifies they're current.

## Output Format

In chat, before yielding:

```markdown
## Sabbath close

**Inventory:** [what was actually done, in plain accounting]

**Declaration:** "[the declaration — "it was good" or fuller form]"

**Carry-forward:**
- [thing 1]
- [thing 2]

**Set down:**
- [thing 1]
- [thing 2]

**Memory updates:**
- ✓ .spec/journal/{date}-{slug}.md written
- ✓ .mind/active.md updated
- [Optional: .mind/principles.md updated because X]
```

Then actually do the memory updates if they aren't already done. Saying you'll do them is not the same as doing them — that was the canonical update_memory failure.

## When the Full Sabbath Agent Is Needed Instead

This skill is the daily-rhythm close. Use the full `sabbath` subagent when:

- A complete cycle finished (study series shipped, major refactor done, crisis resolved)
- The literal Sunday Sabbath
- A long phase concluded (multi-session work that completed)
- A milestone with stakes (publication, ship, deployment, public commit)

The full agent runs the 11-question review and writes a Sabbath Record to `.spec/sabbath/`. This skill writes lighter outputs to the journal and active.md.

## What This Is NOT

- **Not a project status report.** This is internal/personal — for the next-session-self, not for stakeholders.
- **Not optional theatre.** The decoration is empty. The substance — inventory, declaration, carry-forward, set-down — is the cycle's actual close. If you skip it, you skip the cycle's third step.
- **Not a delay.** Five to ten minutes total. The cost is low. The cost of *not* doing it is the next session arriving as a stranger to its own context.

## The Connection to Council Moment + Intent Check

Council moment opens the cycle (what already exists?).
Intent check sets the target (what are we trying to accomplish?).
Sabbath close ends the cycle (what was done; what carries; what we set down).

Together they form the cycle's three observation points: before, during, and after. Skipping the close means the cycle never actually completes — and an uncompleted cycle drifts into the next one, accumulating debt. Mosiah 4:27 lives here: don't start more than you can faithfully close.
