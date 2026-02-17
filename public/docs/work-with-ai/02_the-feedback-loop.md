# The Feedback Loop: Reviewing, Steering, and Iterating with AI

**Series:** Working with AI — Part 2 of 3
**Duration:** 30 minutes
**Audience:** Software engineers adopting AI-assisted development (VS Code + GitHub Copilot / Cursor)
**Date:** February 2026

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| 1 | [Planning Then Creating](01_planning-then-create.md) | Why specification-driven development matters more than ever with AI |
| **2** | **[The Feedback Loop](02_the-feedback-loop.md)** | **How to review, steer, and iterate — you as the architect, AI as the builder** |
| 3 | [Live Build](03_live-build.md) | Build something real together, start to finish, applying the patterns |

### Glossary

| Term | Definition |
|------|------------|
| **Session** | One prompt-and-response cycle. You say something, the AI processes and responds with text, tool calls, file edits, etc. |
| **Chat session** | The full conversation containing multiple sessions. Your ongoing back-and-forth in one chat window. |
| **Spec / Blueprint** | The planning document created collaboratively before implementation begins. |
| **Feedback loop** | Review → diagnose → correct → verify → repeat. |

---

## Part 2: The Feedback Loop (30 min)

### Opening (3 min)

**Recap from Part 1:** You learned the pattern — envision, specify, build, watch and steer. You wrote a planning doc before writing code. The spec is the product.

But the spec doesn't guarantee a perfect first draft. The AI will misunderstand. It will drift. It will solve the wrong problem elegantly. And that's *fine* — as long as you know how to steer.

**The question this lesson answers:** When the AI generates something that's close but not right, how do you course-correct efficiently without starting over?

The answer is a feedback loop: **review → diagnose → correct → verify → repeat**. This is the core skill of AI-assisted development. Anyone can prompt. The engineers who produce great work are the ones who *review and steer*.

**An important caveat before we start:** The *goal* of the feedback loop is to make itself unnecessary. As your specs improve, as your patterns stabilize, as the AI's track record earns trust — you iterate less and less. I've had sessions where 1,112 lines across 13 files compiled and type-checked on the first try. Zero corrections needed. That wasn't luck — it was the payoff of a thorough planning document (Part 1) and a codebase with established patterns the AI could follow. Part 2 isn't about creating work. It's about developing the capacity to notice when something *isn't* right, so that when everything *is* right, you can move with confidence.

---

### Why AI Output Needs Review (5 min)

**The AI is a brilliant first-drafter, not an infallible implementer.**

Even with a perfect spec, the AI will:

| Failure Mode | What It Looks Like | Why It Happens |
|---|---|---|
| **Drift** | Implementation gradually diverges from spec | The AI optimizes locally — each function looks reasonable, but the system drifts from the original architecture |
| **Hallucination** | Invents API calls, libraries, or patterns that don't exist | Trained on patterns, not truth — it produces plausible-looking code that references nonexistent things |
| **Wrong abstraction** | Solves the problem but with the wrong structure | It doesn't know your codebase conventions or your team's preferences unless you tell it |
| **Missing edge cases** | Works for the happy path, breaks on boundaries | It generates what you asked for, not what you *meant* |
| **Stale patterns** | Uses deprecated APIs or outdated idioms | Training data has a cutoff — it may not know your framework's latest version |

**Key insight:** These failures aren't bugs in the AI. They're the *expected behavior* of any system that generates code from natural language. Natural language is ambiguous. Code is precise. The gap between the two is where your review skills matter.

**Analogy:** Think of the AI as a brilliant junior developer who just joined the team. They're smart, fast, and eager. But they don't know:
- Your project's history and the decisions behind it
- Which patterns you've already tried and abandoned
- The implicit conventions your team follows
- The edge cases that burned you last quarter

Everything a junior does gets reviewed. AI output should be treated the same way. 

---

### The Review Process (10 min)

Here's the review process that works consistently:

#### Step 1: Read Before You Run

Don't execute AI-generated code blindly. Read it first. You're looking for:

- **Does it match the spec?** Compare the implementation to your planning doc. Every function should correspond to something in the plan.
- **Does it fit the existing codebase?** Look at naming conventions, error handling patterns, file organization. Does it follow what's already there?
- **Does it make reasonable assumptions?** The AI will fill gaps in your spec with assumptions. Are those assumptions correct?
- **Was the methodology sound?** A comprehensive-looking result can mask an incomplete process. In one research session, the AI produced a thorough, well-sourced document — but had only used keyword search, never semantic search. The output *looked* complete, but an entire class of relevant results was missing because the wrong search tool was used. Reviewing *how* something was produced is as important as reviewing *what* was produced.

> Treat the AI's output as a *pull request*, not a *deployment*.

#### Step 2: Diagnose the Gap

When something is wrong, figure out *what kind* of wrong it is before correcting:

| Type of Problem | Your Response |
|---|---|
| **Spec was unclear** | Clarify the spec, then ask for a redo. The AI followed your words — the words were ambiguous. |
| **AI missed context** | Point it to the relevant existing code or document. "Look at how we handle this in `auth.go` and follow the same pattern." |
| **Wrong approach entirely** | Start that section fresh. Don't try to patch a fundamentally wrong architecture — describe what you actually want. |
| **Small inaccuracy** | Targeted correction. "Change the error handling in `parseConfig` to return the error instead of logging and continuing." |
| **Missing capability** | You have a tool or approach you didn't use. Zero-result searches should prompt a tool review, not a retreat to fallbacks. "Try semantic search instead of keyword search for conceptual queries." |

**The most common mistake:** Trying to fix a wrong-approach problem with targeted corrections. If the AI built a REST API and you wanted GraphQL, don't try to tweak the REST code into GraphQL. Describe what you want and let it rebuild.

#### Step 3: Give Corrections That Stick

This is where most people struggle. Your corrections need to be:

**Specific, not vague.**
- Bad: "This doesn't look right."
- Good: "The `processOrder` function is mutating the input struct directly. It should create a copy, modify the copy, and return it. We treat all inputs as immutable in this codebase."

**Contextual, not abstract.**
- Bad: "Use better error handling."
- Good: "Follow the error handling pattern in `handlers/user.go` — wrap errors with `fmt.Errorf` and the function name, and return them to the caller instead of logging here."

**Directional, not just critical.**
- Bad: "This API design is wrong."
- Good: "The API should be organized by resource, not by action. Instead of `/createUser` and `/deleteUser`, use `POST /users` and `DELETE /users/{id}`. See the existing routes in `routes.go` for the pattern."

**The principle:** You're not just telling the AI what's wrong — you're teaching it how things work *in your project*. Every correction adds to the session's context. Good corrections compound.

#### Step 4: Verify the Fix

After the AI corrects, check:
- Did it actually fix the thing you pointed out?
- Did it break something else in the process?
- Did it apply the pattern consistently, or just in the one place you mentioned?

This is where working in small increments pays off. If you asked for one function and it's wrong, the blast radius is small. If you asked for 500 lines and something's off, finding the problem is much harder.

#### Real Example: The Range Parsing Bug

Here's a real example from my own work. I built an MCP server — a tool that lets an AI fetch structured content from a local database. I had a function called `content_get` that retrieves specific entries by ID.

**The spec said:** "Accept entry ranges like `section-93:24-30` and return all entries in the range."

**What the AI built:** A `parseReference` function that extracted the section and entry number. It worked perfectly for single entries. But when I passed `section-93:24-30`, it only returned entry 24. It parsed `24-30` and took just the first number.

**How I found it:** I used the tool. I asked for `section-93:24-30` and got one entry back. The spec said I should get seven. That tells me exactly what to look for.

**The correction:** "The `parseReference` function needs to split the entry portion on `-` to extract a start and end entry. Add an `EndEntry` field. Then `getEntryRange` should query `WHERE entry_id >= ? AND entry_id <= ?` and return all matching entries."

**After the fix:** `section-93:24-30` returns all seven entries. But I also tested `section-93:24` (single entry, no range) to make sure the fix didn't break the simple case.

**The pattern:** Use → discover the gap → diagnose (spec vs. implementation) → give a specific correction → verify the fix → test adjacent cases.

---

### When the AI Gets Stuck (7 min)

Sometimes the feedback loop breaks down. You correct, the AI "fixes" it, but introduces a new problem. You correct that, and the original problem comes back. You're going in circles.

**Recognizing the spiral:**
- You've given the same correction more than twice
- Each "fix" introduces a new, different bug
- The AI starts apologizing excessively (a sign it's lost)
- The output is getting *longer* but not *better*

**When you're spiraling, stop correcting and do one of these:**

#### Strategy 1: Start fresh on that section
Don't try to patch something that's structurally wrong. Describe what you want from scratch, with the lessons you've learned from the failed attempts included in your description.

"Let's start over on the authentication module. Here's what I need:
- JWT-based auth with refresh tokens
- Middleware pattern (not in each handler)
- Store refresh tokens in the database, not in memory
- Follow the error handling pattern from `handlers/user.go`"

You're carrying forward the corrections as *specifications*. Every failed attempt taught you something — encode that in the next prompt.

#### Strategy 2: Break it smaller
If the AI can't get a whole module right, ask for one function at a time. Verify each function before moving to the next.

"Just write the `validateToken` function for now. It takes a JWT string, validates the signature using the key in `config.SecretKey`, checks the expiration, and returns the claims or an error."

Once that works, move to the next function. The AI now has a verified piece of code as context, which helps it get the next piece right.

#### Strategy 3: Show, don't tell
If the AI keeps misunderstanding a pattern, write *one example yourself* and say "follow this pattern."

```go
// This is how we handle errors in this codebase:
func getUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("getUser: %w", err)
    }
    return user, nil
}
// Now write deleteUser following the same pattern.
```

One concrete example communicates more than ten paragraphs of description.

#### Strategy 4: Review Your Tool Selection
Sometimes the problem isn't the AI's logic — it's the *tools* being used. If a search returns zero results, the instinct should be "try a different search tool" before "there's nothing to find."

**Real example — tool familiarity bias:** In a research session, the AI ran two conceptual queries through a keyword search engine. Both returned zero results. Instead of switching to the semantic vector search tool — which was available and designed for exactly this kind of conceptual query — the AI fell back to direct file reads. When the user asked "did you try the vector search?", semantic searches immediately found relevant documents the keyword search had missed entirely. None were obscure. The tool was wrong, not the query.

The principle: AI tends toward familiar tools. Zero results should trigger a tool-selection review — not capitulation.

#### Strategy 5: Clear the context
In long sessions, the AI's context window fills up with earlier conversation — including the wrong code and the corrections. All that noise can actively confuse it. Starting a new chat with a clean, refined prompt sometimes works better than continuing to correct.

Bring your spec, bring the lessons learned, leave the failed attempts behind.

#### Real Example: Cross-Reference Scoping

Another real example from the same tool. My content database stores cross-references — links connecting one entry to related entries. When I asked for cross-references for a specific entry, I was getting cross-references for the *entire section*.

**Diagnosis:** The function was running its regex on the entire references block, ignoring which entry each reference belonged to. Each reference had an anchor like `<a id="fn-9a">` where `9` is the entry number, but the code didn't filter by that.

**First correction:** "Only extract cross-references for the requested entry by checking the anchor ID against the entry number."

**The fix worked — but revealed a second problem.** Now the cross-references were correct per entry, but the *index* (a pre-built database) still had the old, section-wide cross-references. The fix only worked for real-time parsing, not for indexed data.

**Second correction:** "This change requires re-indexing. The indexer needs the same entry-scoping logic applied at index time."

**The lesson:** Fixing a bug can reveal a deeper architectural issue. The first fix was correct but incomplete. Watching the behavior after the fix — *verifying* — revealed what else needed to change. You can't design this away. You discover it by using the tool and paying attention.

---

### The Trust Gradient (3 min)

Not all AI output deserves the same level of scrutiny. Over time, you develop a feel for when to read every line and when to scan-and-move.

| Situation | Trust Level | Your Response |
|---|---|---|
| **Boilerplate / scaffolding** | High | Scan structure, verify naming, move on |
| **Business logic** | Medium | Read carefully, verify against spec, test edge cases |
| **Security / auth / crypto** | Low | Read every line. Test adversarially. Consider getting a human review too. |
| **Infrastructure / deployment** | Low | Verify against your actual environment. AI doesn't know your infra. |
| **New pattern / unfamiliar territory** | Low | Read, research, verify. The AI may be confidently wrong. |
| **Pattern it's done correctly 5 times already** | High | Spot check, move on |

**Real example — preemptive pattern following:** In one session, I needed a new middleware called `auth.Optional` — same as the existing `auth.Required` but allowing unauthenticated requests through. Rather than designing it from scratch, the AI read the existing `auth.Required` code and produced `auth.Optional` by following the same structure while removing the 401 rejection. Same variable names, same session lookup, same dev-mode handling. A reviewer would think the same person wrote both. No correction needed. That's what happens when the AI has good context: it follows existing patterns preemptively, and the review step becomes "saw it was correct, moved on."

**The principle:** Trust is earned through verification, not assumed. As you verify more output in a session, you calibrate. You learn where this particular AI, in this particular codebase, gets things right and where it struggles.

This is the same process you use with a new team member. At first, you review everything. Over time, you learn their strengths and weaknesses. You review more carefully in their weak areas. You trust their strong areas.

---

### Wrap-Up and Preview (2 min)

**The feedback loop:**
1. **Review** — Read before you run. Treat AI output as a PR.
2. **Diagnose** — What kind of problem is it? Spec gap, missing context, wrong approach, or small bug?
3. **Correct** — Be specific, contextual, and directional. Every correction teaches.
4. **Verify** — Check the fix. Check that nothing else broke. Test adjacent cases.
5. **Repeat** — Until the output matches the spec.

**When you're stuck:**
- Start fresh on that section (carry lessons forward as specs)
- Break it smaller (one function at a time)
- Show, don't tell (write one example, let the AI follow it)
- Clear the context (new chat with clean prompt)

**Next session: Live Build**
- Pick a project from someone's homework doc
- Build it together in real-time, start to finish
- Apply the full pattern: plan → spec → build → review → steer → ship
- Everyone watches the feedback loop happen live

**Homework:**
Take the planning doc you wrote after Part 1. Start building it with AI. When the AI produces something wrong, **don't just fix it yourself**. Practice the feedback loop:
1. Diagnose what kind of problem it is
2. Write a specific, contextual correction
3. Verify the fix
4. Note what worked and what didn't

Bring your notes to the next session.

---

## Facilitator Notes

### Key Points to Emphasize
- **Review is the skill.** Prompting is easy. Reviewing well is hard. This is where engineering judgment matters most.
- **Corrections are specifications.** Every correction you give adds to the session's working context. Good corrections compound into better output throughout the session.
- **The feedback loop is not a failure.** Needing to correct the AI isn't a sign that AI tools are broken. It's the normal workflow. The question is whether you steer efficiently or waste time fighting the tool.
- **The spiral is the enemy.** Learn to recognize when you're going in circles and switch strategies before wasting 30 minutes on a doomed approach.

### Common Objections

| Objection | Response |
|-----------|----------|
| "If I have to review everything, what's the point?" | You review *output*, not *syntax*. Reading code for correctness is much faster than writing it from scratch. You're 5-10x faster, not infinitely faster. |
| "I spent an hour fighting the AI on something I could have written in 20 minutes" | That happens. The skill is recognizing it early and switching strategies. As you calibrate, you'll know when to steer and when to just write it yourself. |
| "The AI keeps making the same mistake" | That's a signal to either (1) show it an example, (2) break the task smaller, or (3) start a fresh context. Repeating the same correction is the least effective strategy. |
| "I don't know if the AI's code is correct — I don't know the language well enough" | Then you need to learn enough of the language to review. AI shifts the skill from *writing* to *reading and evaluating*. You still need domain knowledge. |

### Live Demo Ideas
- **Live correction demo:** Have the AI write a function, intentionally give a vague spec, then walk through diagnosing and correcting the output
- **Spiral demo:** Show a real example of going in circles, then demonstrate the "start fresh with lessons learned" strategy
- **Trust gradient exercise:** Show three different pieces of AI output (boilerplate, business logic, auth) and have the audience discuss what level of review each needs
- **Before/after corrections:** Show a vague correction ("fix the error handling") vs. a specific one, and compare the AI's responses

### Series Roadmap
- **Part 1: Planning Then Creating** — The spec-driven pattern. Done.
- **Part 2: The Feedback Loop** — Today. Reviewing, diagnosing, correcting, verifying.
- **Part 3: Live Build** — Apply the full pattern to a real project, live. The audience sees envision → specify → build → review → steer in action.
