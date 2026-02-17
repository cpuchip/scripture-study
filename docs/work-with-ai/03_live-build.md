# Live Build: Applying the Full Pattern

**Series:** Working with AI — Part 3 of 3
**Duration:** 30 minutes
**Audience:** Software engineers adopting AI-assisted development (VS Code + GitHub Copilot / Cursor)
**Date:** February 2026

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| 1 | [Planning Then Creating](01_planning-then-create.md) | Why specification-driven development matters more than ever with AI |
| 2 | [The Feedback Loop](02_the-feedback-loop.md) | How to review, steer, and iterate — you as the architect, AI as the builder |
| **3** | **[Live Build](03_live-build.md)** | **Build something real together, start to finish, applying the patterns** |

### Glossary

| Term | Definition |
|------|------------|
| **Session** | One prompt-and-response cycle. You say something, the AI processes and responds with text, tool calls, file edits, etc. |
| **Chat session** | The full conversation containing multiple sessions. Your ongoing back-and-forth in one chat window. |
| **Spec / Blueprint** | The planning document created collaboratively before implementation begins. |
| **Feedback loop** | Review → diagnose → correct → verify → repeat. |

---

## Part 3: Live Build (30 min)

### Opening (3 min)

**Recap:**
- **Part 1:** Plan before you build. The spec is the product.
- **Part 2:** Review, diagnose, correct, verify. The feedback loop is the skill.

Today we do it live. We're going to pick a project, write a spec, build it with AI, and handle whatever goes wrong — all in 30 minutes.

**Why live?** Because reading about the feedback loop is not the same as watching it happen. You need to see the messiness — the drift, the corrections, the moments where you have to decide: steer or start fresh? You need to see that this isn't a clean, linear process. It's iterative, sometimes frustrating, and it *still* produces better results faster than writing everything by hand.

**The goal isn't a polished product.** The goal is to demonstrate the *pattern* in action. You'll see me:
1. Write a quick spec
2. Feed it to the AI
3. Review the output
4. Correct when something's off
5. Ship something that works

---

### Choose the Project (2 min)

If someone brought a planning doc from the Part 2 homework — use that. Real projects are better than contrived examples.

If no one has a doc, here are fallback options (pick the one that matches the audience):

| Project | Complexity | Good For Demonstrating |
|---------|-----------|----------------------|
| CLI tool that converts CSV to JSON | Low | Basic spec → build → verify cycle |
| REST API with 2-3 endpoints | Medium | Architecture decisions, error handling patterns, the feedback loop at scale |
| VS Code extension that highlights TODOs | Medium | Working with an unfamiliar API, the AI filling knowledge gaps |
| Refactoring an existing messy function | Low-Medium | How specs work for *changing* code, not just creating it |

The sweet spot is something with 3-5 components that can be specced in 5 minutes and built in 15. Complex enough that the AI will make at least one meaningful mistake. Simple enough that we can ship in the session.

---

### Write the Spec Together (5 min)

**Do this live, with the audience watching.**

Describe the project to the AI in the chat — what it does, what you're thinking for structure. Then ask the AI to create `docs/plan.md`. Ask the audience to contribute:

- What does this thing do? (1-2 sentences)
- What are the inputs and outputs?
- What are the components? (no more than 3-4)
- What's the data model? (if applicable)
- What edge cases should we handle?

Feed their answers into the conversation. Let the AI update the spec. Go back and forth once or twice — the audience sees the collaborative refinement in real time. This demonstrates Phase 2 from Part 1 — the blueprint.

**Example spec (if using the CSV→JSON CLI):**

```markdown
# csv2json — Planning Doc

## What It Does
CLI tool that reads a CSV file and outputs formatted JSON.

## Usage
csv2json input.csv [--output file.json] [--pretty]

## Components
1. CLI argument parser (flag package)
2. CSV reader (encoding/csv)
3. JSON writer (encoding/json)
4. Main orchestrator

## Data Flow
stdin/file → csv.Reader → []map[string]string → json.Marshal → stdout/file

## Edge Cases
- Empty file → empty JSON array
- Missing output flag → write to stdout
- Malformed CSV → clear error message with line number
- Headers with spaces → preserve as-is in JSON keys
```

**Point out:** This took 3 minutes to write. It's not a 50-page design doc. It's just enough to give the AI — and us — a shared understanding of what we're building. Spec-driven doesn't mean heavy process. It means *thinking before typing*.

---

### Build with AI (12 min)

Now the live part. Open VS Code. Start a Copilot / Cursor chat. Feed it the spec.

**What the audience should watch for:**

#### 1. The First Draft (3-4 min)
Paste the spec. Ask the AI to implement it. The first output will probably be 80-90% right.

**Narrate what you're doing:**
- "I'm reading the output before running it."
- "I'm checking: does the argument parsing match our spec?"
- "The CSV reader looks fine — standard library, nothing unusual."
- "Wait — it's not handling the empty file case. That was in our spec."

#### 2. The First Correction (2-3 min)
Point out the gap. Show the audience what a good correction looks like:

- Bad: "Handle the edge case."
- Good: "The empty file case is missing. When `csv.ReadAll()` returns an empty slice, write `[]` to the output instead of proceeding to the JSON marshal step. Add this check right after the ReadAll call."

**Call out the diagnosis:** "This is a 'missing edge case' problem, not a 'wrong approach' problem. The architecture is fine. I'm adding a requirement the AI missed, not redesigning the solution."

#### 3. The Organic Problem (5-6 min)
Something will go wrong that you didn't plan for. Maybe the AI uses a deprecated API. Maybe it structures the code in a way that doesn't match your style. Maybe the error handling is inconsistent.

**This is the gold.** This is the feedback loop in its natural habitat. Don't script this. Let it happen. Then narrate your thinking:

- "I notice it's logging errors instead of returning them. In our codebase we return errors to the caller. Let me show it the pattern."
- "It generated the whole thing as one function. The spec said four components. Let me ask it to break this into the pieces we planned."
- "It's going in a direction I don't like. The output format isn't what I described. I'm going to give a very specific correction with an example of what I want."

#### 4. Verification (2 min)
Run the tool. Test the happy path. Test one edge case. Show the output.

"Does this match our spec? Let me check:
- ✅ Reads CSV
- ✅ Outputs JSON
- ✅ Pretty print flag
- ❌ Empty file — let me test that..."

If the empty file case still isn't handled, *that's another teaching moment*. Fix it live.

---

### Debrief: What Just Happened (5 min)

Pull the audience back from the code. Connect what they saw to the framework:

**Map it to the pattern:**
1. **Envision** — "What does this thing do?" discussion (2 min)
2. **Specify** — The planning doc (3 min)
3. **Build** — AI implementation against the spec (first draft)
4. **Watch and steer** — Reading before running. Finding the missing edge case. Correcting the error handling. Restructuring the code.

**Observations to draw out:**

- **How much time was spent writing code?** Almost none. Most of the time was spent *reading*, *thinking*, and *communicating*.
- **Did the spec prevent problems?** Yes — we caught the empty file case because it was in the spec. Without the spec, we might not have thought to test it until production.
- **Was the first draft perfect?** No. It never is. The first draft is a starting point. The feedback loop is what turns it into something good.
- **How long would this have taken without AI?** For a simple tool like this, maybe 30 minutes of typing vs. 15 minutes of directing. But for a complex system? The ratio gets much more dramatic — hours of typing vs. minutes of specifying and steering.

**The key takeaway:** You just watched the entire workflow in 20 minutes. You saw a spec turn into working code. You saw the AI make mistakes and get corrected. You saw the trust gradient develop in real time — from careful review of every line to quick verification of familiar patterns. This is the skill. Not prompting. Not typing. *This*.

---

### Wrap-Up (3 min)

**The three-part framework:**

| Part | Core Skill | Mantra |
|------|-----------|--------|
| 1. Plan | Thinking clearly | Spec before code |
| 2. Review | Reading critically | Verify against spec |
| 3. Build | Applying both live | The pattern is the product |

**What to do next:**
1. **Pick a real project.** Something you've been putting off because of the implementation effort.
2. **Write a 1-page spec.** Components, data model, edge cases.
3. **Build it with AI.** Apply the full pattern: spec → build → review → correct → verify.
4. **Notice your feedback loop.** Where do you steer well? Where do you waste time? The meta-awareness is the skill that compounds.

**The honest truth:** This is how I build everything now. Personal projects, work projects, tools, docs, all of it. The spec-driven approach with a tight feedback loop is not a nice-to-have. It's the most efficient way I've found to turn ideas into working software. The engineers who adopt this pattern are going to be dramatically more productive. The engineers who resist are going to wonder why they're falling behind.

This isn't a fad. It's a fundamental shift in how software gets built. And now you know the pattern.

---

## Facilitator Notes

### Pre-Session Prep
- **Have a fallback project ready.** Don't count on someone bringing a homework doc. Have your own spec drafted (but don't *show* it drafted — write it "live" from your notes).
- **Test the AI tool in the room's environment.** Network restrictions, proxy issues, and authentication problems are show-stoppers. Verify everything works 30 minutes before the session.
- **Have backup screenshots.** If live coding fails (network goes down, tool crashes), have screenshots of a real session you ran earlier. The teaching points are the same even without the live element.

### Running the Live Build
- **Narrate constantly.** The audience can't see your thoughts. Say out loud: "I'm reading the function signature... checking the error handling... this doesn't match our spec because..."
- **Don't fake perfection.** If the AI does something unexpected, don't pretend you planned for it. Say "Huh, I didn't expect that. Let me figure out what happened." That's the most valuable thing the audience can see — a skilled practitioner encountering the unknown and reasoning through it.
- **Keep time ruthlessly.** The spec should take no more than 5 minutes. If the audience is debating features, cut it off: "Good enough. We can always revise. Let's build."
- **If you spiral, name it.** If you hit the "going in circles" pattern from Part 2, stop and say: "This is the spiral we talked about last time. Watch what I do next." Then demonstrate the recovery strategy.

### Common Objections

| Objection | Response |
|-----------|----------|
| "You cherry-picked an easy project" | Fair. The pattern scales but the demo needs to fit 30 min. Try it on your own complex project — the feedback loop is the same, just longer. |
| "This wouldn't work for my codebase — it's too complex" | The more complex the codebase, the more the spec matters. AI without a spec in a complex codebase is a disaster. AI *with* a spec and codebase context is incredibly powerful. |
| "What about testing?" | Great question. The spec is your test plan. Each bullet point in the spec is a test case. You can even ask the AI to generate tests from the spec. |
| "My company won't let us use AI for production code" | Learn the pattern anyway. The spec-driven approach and the feedback loop make you better at directing *any* implementer — human or AI. And company policies are changing fast. |

### Adapting for Different Audiences

| Audience | Adjust |
|----------|--------|
| **Junior engineers** | Emphasize that the spec is training wheels — the AI compensates for syntax gaps while you focus on design. Show that not knowing every API is fine. |
| **Senior engineers** | Emphasize leverage. "You already think in systems. Now you can implement at the speed of thought instead of the speed of typing." |
| **Managers / leads** | Emphasize the review pattern. "Your code review skills are more valuable than ever. Every review you do teaches the session context." |
| **Skeptics** | Don't argue. Demonstrate. Let the live build speak. Then say: "Try it for a week. If it doesn't work for you, no harm done." |

### Full Series Summary

| Session | You Learned | You Practiced |
|---------|-------------|---------------|
| Part 1 | Spec-driven development | Writing a planning doc |
| Part 2 | The feedback loop | Diagnosing and correcting AI output |
| Part 3 | The full pattern live | Building, reviewing, shipping |

The pattern is simple. The skill is in the execution. And the execution gets better every time you do it.
