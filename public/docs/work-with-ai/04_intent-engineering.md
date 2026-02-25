# Intent Engineering: Telling AI What to *Want*

**Series:** Working with AI — Part 4
**Duration:** 30 minutes
**Audience:** Software engineers adopting AI-assisted development (VS Code + GitHub Copilot / Cursor)
**Date:** February 2026
**Prompted by:** [Nate B Jones, "Prompt Engineering Is Dead. Context Engineering Is Dying."](https://www.youtube.com/watch?v=QWzLPn164w0)

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| 1 | [Planning Then Creating](01_planning-then-create.md) | Why specification-driven development matters more than ever with AI |
| 2 | [The Feedback Loop](02_the-feedback-loop.md) | How to review, steer, and iterate — you as the architect, AI as the builder |
| 3 | [Live Build](03_live-build.md) | Build something real together, start to finish, applying the patterns |
| **4** | **[Intent Engineering](04_intent-engineering.md)** | **Encoding purpose — so agents optimize for what you actually need** |

### Glossary (New Terms)

| Term | Definition |
|------|------------|
| **Prompt engineering** | Crafting individual instructions for AI. Personal, session-level. (2023-2024 era) |
| **Context engineering** | Building the information environment an AI operates within — RAG, MCP, organizational knowledge. (2025-2026 era) |
| **Intent engineering** | Encoding *purpose* — goals, values, trade-offs, decision boundaries — so agents optimize for the right outcome. (Emerging) |
| **Decision boundary** | A rule defining which decisions an agent can make autonomously vs. which require human input. |
| **Value hierarchy** | An ordered list of what matters most — so agents can resolve trade-offs without asking every time. |
| **Intent drift** | When an agent's behavior gradually diverges from the original purpose, even while technically completing tasks. |

---

## Part 4: Intent Engineering (30 min)

### Opening (3 min)

**Recap from Parts 1-3:**
- **Part 1:** Plan before you build. The spec is the product.
- **Part 2:** Review, diagnose, correct, verify. The feedback loop is the skill.
- **Part 3:** Live build — applying the full pattern.

Those three parts gave us a solid workflow: *spec → build → feedback loop*. It works. But there's a layer underneath that we haven't addressed — and it explains why even good specs sometimes produce wrong-feeling output.

**The question this lesson answers:** When the AI has a perfect spec and all the context it needs and *still* optimizes for the wrong thing — what's missing?

The answer: **intent**. Not *what* to do, but *why* we're doing it. And crucially: what trade-offs we accept and what constraints are non-negotiable.

---

### The Klarna Story (5 min)

This is the story that makes intent engineering concrete.

> [Nate B Jones, 2:12](https://www.youtube.com/watch?v=QWzLPn164w0&t=134): "Klarna's AI customer service agent was extraordinarily good at resolving tickets fast. And that was the wrong goal to give the agent."

Klarna deployed an AI agent for customer service. It had excellent prompts. It had deep context — customer history, product database, conversation logs. By every measurable metric, it was working: tickets resolved faster, costs down, efficiency up.

Then customers started leaving.

> [Jones, 2:19](https://www.youtube.com/watch?v=QWzLPn164w0&t=139): "Klarna's actual intent was to build lasting customer relationships that drive lifetime value."

The agent was optimizing for *resolution speed*. But Klarna's actual goal was *relationship quality*. The agent learned that rushing customers off the line resolved tickets fastest. Technically correct. Strategically catastrophic.

**This is an engineering problem, not an AI problem.** The agent did exactly what it was told. The failure was in what was *encoded* — and what was left implicit.

---

### The Three Disciplines (7 min)

Parts 1-3 of this series map onto a progression the industry has gone through:

| Discipline | Era | Question | Our Series |
|-----------|-----|----------|------------|
| **Prompt engineering** | 2023-2024 | "How do I talk to AI?" | Part 1 — crafting the spec |
| **Context engineering** | 2025-2026 | "What does AI need to know?" | Part 2 — building the feedback environment |
| **Intent engineering** | Emerging | "What does AI need to *want*?" | **Part 4 — this session** |

The video frames these as three layers an organization needs:

#### Layer 1: Unified Context Infrastructure

> [Jones, 11:16](https://www.youtube.com/watch?v=QWzLPn164w0&t=676): "This is the layer the industry is most aware of and it's still not really built yet."

The problem: every team rolling their own context stack. Custom RAG pipelines, disconnected MCP servers, shadow agents. No shared organizational knowledge layer.

**What this looks like in practice:** Your project has documentation in Confluence, code in GitHub, designs in Figma, decisions in Slack threads, and architecture in someone's head. The AI agent gets access to *some* of these through whatever MCP servers or retrieval you've set up. But there's no unified view. Every agent session starts with partial knowledge.

In our series, Part 1 addressed this at the individual level — the spec is your shared context. But at the organizational level? Most companies haven't solved this yet.

#### Layer 2: Coherent Worker Toolkit

> [Jones, 13:58](https://www.youtube.com/watch?v=QWzLPn164w0&t=838): "Everyone's rolling out their own AI workflow. None of these employees can articulate their workflow in a way that's transferable, measurable, or improvable."

The problem: individual tool use doesn't scale. One engineer's Copilot workflow doesn't transfer to the next engineer. There's no way to measure whether AI-assisted work is actually better.

In our series, Part 3 demonstrated a transferable pattern — but only for one team's workflow. The challenge is making these patterns organization-wide: shared agent configurations, consistent feedback practices, reusable spec templates.

#### Layer 3: Intent Engineering

> [Jones, 16:20](https://www.youtube.com/watch?v=QWzLPn164w0&t=980): "This is the layer that almost certainly doesn't exist in your business. It requires something genuinely new."

The problem: OKRs were designed for humans who absorb culture through osmosis — lunch conversations, overheard meetings, watching what the boss actually prioritizes (vs. what they say they prioritize). Agents don't have any of this. They need explicit alignment *before* they start working.

The video identifies what needs to be encoded:

| What to Encode | Why |
|---------------|-----|
| **Goal structures agents can act on** | Not "improve customer satisfaction" but specific, measurable success criteria with defined scope |
| **Decision boundaries** | What can the agent decide autonomously? What requires escalation? |
| **Delegation frameworks** | Which agent handles what? How do handoffs work? |
| **Value hierarchies** | When two goals conflict, which wins? |
| **Feedback mechanisms** | How does the agent know it's drifting? Who reviews? |
| **Escalation paths** | When the agent hits uncertainty, where does it go? |

---

### Why This Matters for Individual Engineers (5 min)

"Intent engineering sounds like an enterprise problem. I'm one engineer with Copilot."

Fair. But the pattern scales down perfectly. Every time you start a chat session, you're implicitly making intent decisions:

**Without intent:**
> "Build a caching layer for the API."

The AI will build *a* caching layer. It'll make choices about eviction strategy, TTL, invalidation patterns, consistency guarantees. All reasonable. But were they *your* choices? Did it optimize for performance (fastest response) or consistency (freshest data) or simplicity (easiest to maintain)?

**With intent:**
> "Build a caching layer for the API. Our priority is data freshness over response speed — we'd rather have a cache miss than serve stale data. Keep it simple enough that a junior engineer can debug it. Redis is fine for now but don't couple tightly — we might switch to Memcached."

Now the agent knows:
- **Value hierarchy:** freshness > speed > complexity
- **Decision boundary:** Redis for now, but abstracted
- **Constraint:** junior-friendly code, no clever patterns
- **Success criteria:** not "fast responses" but "never stale + maintainable"

That's intent engineering at the individual level. Your spec says *what* to build. Your intent says *what matters about how it's built*.

**This connects directly to Parts 1-3:**

```
INTENT (Why are we doing this? What trade-offs do we accept?)
  ↓
SPEC (What are we building? — Part 1)
  ↓
BUILD (Implement against the spec)
  ↓
REVIEW (Does it match the intent, not just the spec? — Part 2)
```

Part 4 adds the layer underneath that Parts 1-3 assumed you were carrying in your head. The problem: agents can't read your head.

---

### Encoding Intent in Practice (7 min)

So how do you actually do this? Three levels, from simplest to most structured:

#### Level 1: The Intent Preamble

Add an intent block at the top of every spec or agent instruction file:

```markdown
## Intent
**Purpose:** Rebuild the notification system to reduce alert fatigue
**Success looks like:** Engineers respond to 80%+ of alerts (currently ~30%)
**Constraints:**
  - No new infrastructure (use existing Kafka + PagerDuty)
  - Must be backwards-compatible with current alert rules
  - Don't optimize for fewer alerts — optimize for more *actionable* alerts
**Trade-offs we accept:**
  - Slower rollout > breaking existing alerting
  - False negatives (missed alerts) are worse than false positives (extra noise)
**Decision boundaries:**
  - Agent can: restructure alert routing, modify severity levels, update templates
  - Agent must ask: before deleting any existing alert rule, before changing on-call rotation logic
```

This takes 5 minutes to write. It saves hours of misaligned work.

#### Level 2: Value Hierarchies per Project

For ongoing projects (not one-off tasks), maintain a living intent document:

```markdown
# Project: Customer Portal v3

## Values (ordered)
1. Data integrity — never show incorrect account information
2. Accessibility — WCAG AA minimum, AAA where practical
3. Performance — page loads under 2s on 3G
4. Developer experience — new engineers productive within one week
5. Feature velocity — ship fast, but never at the cost of 1-4

## Decision Boundaries
| Domain | Agent Autonomous | Requires Human |
|--------|-----------------|----------------|
| UI layout / styling | ✓ | |
| API endpoint design | ✓ | |
| Database schema changes | | ✓ |
| Auth/permission changes | | ✓ |
| Third-party integrations | | ✓ |
| Error message wording | ✓ | |
| Performance trade-offs | | ✓ |
```

This becomes part of your project's `.github/copilot-instructions.md` or equivalent. Every agent session inherits it.

#### Level 3: Intent-Linked Task Management

For teams running multiple agents across many tasks, connect intent to your task tracker:

```
Epic: Customer Portal v3 - Notification Overhaul
Intent: Reduce alert fatigue so engineers respond to real problems
Success: Alert response rate from 30% → 80%
Constraints: No new infra, backwards-compatible

  Task: Restructure severity levels
  Inherited intent: ↑ from epic
  Decision boundary: Can change severity mappings, must ask before removing any level
  
  Task: Design alert digest format
  Inherited intent: ↑ from epic  
  Additional constraint: Must work in email AND Slack (no rich formatting assumptions)
```

Every task carries its "why" from the parent epic. When an agent starts work on a task, it knows not just *what* to implement but *what the work is for* and *what trade-offs to make*.

---

### The Intent Drift Problem (3 min)

> [Jones, 22:07](https://www.youtube.com/watch?v=QWzLPn164w0&t=1327): "Organizations don't change their intent on purpose. But what you intended gets gradually diluted."

This is the maintenance problem. You write a great intent doc on day one. Six months later, you've:
- Added features that subtly conflict with the original constraints
- Promoted speed over quality "just this once" — five times
- Let the agent make decisions that should have been escalated, because it was faster
- Stopped reviewing against intent and started reviewing against "does it work?"

**How do you detect intent drift?**

1. **Regular intent reviews.** Every sprint retrospective (or personal review), ask: "Is our work still aligned with what we said we valued?" Not "did we ship?" but "did we ship the *right things*?"

2. **The intent audit.** Pick your last 10 completed tasks. For each one, check: does this serve the stated purpose? Does it honor the constraints? Would you make the same trade-offs today?

3. **Constraint violation tracking.** When a constraint gets violated, don't just fix it — log it. Three violations of the same constraint means either the constraint is wrong (update the intent) or the workflow is broken (fix the process).

4. **The "explain it to a new hire" test.** If you couldn't explain why a recent decision serves the project's stated intent, the intent has drifted.

---

### Practical Application (5 min)

**Exercise: Write an intent block for your current project.**

Take 3 minutes right now. Answer these questions:

1. **What is this project actually for?** Not the Jira ticket — the real purpose.
2. **What does success look like that you can't easily measure?** (Hint: if all you can think of are metrics, you haven't found the intent yet.)
3. **What are the non-negotiable constraints?** What would you reject even if it "worked"?
4. **When two good things conflict, which wins?** Write at least three value comparisons:
   - ___ > ___
   - ___ > ___
   - ___ > ___
5. **What can the AI decide? What needs you?** Draw the line.

**Now look at your current spec or copilot instructions.** Is any of this encoded? If not, your agents are guessing at your intent. Some of them are guessing right — but you're relying on luck, not engineering.

---

### The Full Pattern (2 min)

Parts 1-4 give us the complete workflow:

```
INTENT — What matters? What constraints are non-negotiable? What trade-offs do we accept?
  ↓
SPEC — What are we building? (Part 1: the blueprint)
  ↓
BUILD — Implement against the spec
  ↓
REVIEW — Does it match the intent? (Part 2: the feedback loop)
  ↓
REFLECT — What did we learn? Should the intent change? (Part 4: intent review)
  ↓
(cycle back to INTENT with updated understanding)
```

Parts 1-3 gave you spec → build → review. Part 4 adds the *why* underneath and the *reflection* on top. The spec says what to build. The intent says what matters about how it's built. The reflection asks whether what we built actually served what we intended.

Without intent: you build correctly but purposelessly.
Without spec: you intend well but build chaotically.
Without feedback: you drift without noticing.
Without reflection: you repeat the same mistakes.

All four together? That's engineering.

---

### Wrap-Up and Preview (3 min)

**The core insight:**

> [Jones, 27:53](https://www.youtube.com/watch?v=QWzLPn164w0&t=1673): "The prompt engineering era asked, 'How do I talk to AI?' The context engineering era is asking, 'What does AI need to know?' And the intent engineering era is beginning to ask the question that really matters: 'What does the organization need AI to *want*?'"

**What you should take away:**

1. **The spec is necessary but not sufficient.** Parts 1-3 taught you to spec before building. Part 4 teaches you to *intend* before speccing. The spec says *what*. Intent says *why*. Without the why, even a perfect spec can produce perfectly wrong output.

2. **Intent engineering is not overhead — it's insurance.** Five minutes writing an intent block saves hours of misaligned work. The cost of not doing it is Klarna: technically impressive, strategically disastrous.

3. **Start small.** You don't need an organization-wide intent framework. Start with an intent preamble on your next spec. Add value hierarchies when you notice trade-off confusion. Build up to project-level intent documents as the payoff becomes obvious.

4. **Review against intent, not just spec.** The feedback loop from Part 2 gets more powerful when you're checking "does this serve the purpose?" not just "does this match the description?"

**Homework:**
1. Write an intent block for your current project (the exercise from earlier — finish it if you didn't).
2. Pick one completed feature from last sprint. Do the intent audit: did it serve the stated purpose? Did it honor the constraints? Where did it drift?
3. Add a value hierarchy to your project's copilot-instructions or CONTRIBUTING.md. Even three lines of "X > Y" will change how the AI makes trade-offs.

---

## Facilitator Notes

### Key Points to Emphasize
- **Intent engineering is the missing layer, not a replacement for Parts 1-3.** The spec is still the product. The feedback loop is still the skill. Intent engineering gives both a *foundation*.
- **This is about encoding what you already know.** Engineers already have values, trade-offs, and constraints in their heads. Intent engineering makes them explicit — so agents (and teammates) can act on them too.
- **Start with the Klarna story.** It's the most memorable illustration of the problem. Everyone has experienced some version of "technically correct, strategically wrong."
- **The individual level is the entry point.** Don't wait for your organization to build an intent framework. Write an intent preamble on your next spec. That's enough to start.

### Common Objections

| Objection | Response |
|-----------|----------|
| "This is just writing better requirements" | It's related but distinct. Requirements say *what*. Intent says *why* and *what trade-offs to make*. A requirement says "build caching." Intent says "freshness over speed, simple over clever." |
| "My project isn't complex enough for this" | Even a simple project has implicit trade-offs. Would you rather have fast code or maintainable code? Encoding that takes 2 minutes and changes the output. |
| "OKRs already solve this" | OKRs work for humans who absorb organizational culture through osmosis. Agents don't attend all-hands meetings. They need the intent *encoded*, not implied. |
| "This feels like more ceremony" | The intent preamble is 5 lines. It's less ceremony than writing a bad spec, building the wrong thing, and then arguing in code review about whether it's correct. |
| "How do I know if my intent is right?" | You won't, perfectly, at the start. That's what the reflect step is for. Intent evolves. The point is to make it *explicit* so you can notice when it drifts — rather than drifting unconsciously. |

### Live Demo Ideas
- **The Klarna roleplay.** Give the audience a prompt: "You're building an AI customer service agent. Write three versions of the spec: one with just a prompt, one with context, one with intent. What changes?" Walk through the differences.
- **Intent audit, live.** Pick a real completed feature from your team's work. On screen, walk through: What was the stated purpose? What trade-offs were made? Did the result serve the intent? This is the most powerful demo because it's real.
- **Before/after copilot-instructions.** Show a real project's instruction file without intent encoding vs. with. Ask the AI the same question in both contexts. Compare the output.
- **Value hierarchy debate.** Give two conflicting scenarios: "The cache layer is fast but sometimes serves stale data" and "The cache layer is always fresh but 2x slower." Ask the audience: which is correct? Then show that *neither* is correct without a value hierarchy. This motivates encoding trade-offs.

### Connection to Parts 1-3
| Part | Contribution | Part 4 Extension |
|------|-------------|-------------------|
| 1: Planning | The spec is the product | Intent is the spec's foundation — *why* this spec? |
| 2: Feedback Loop | Review → diagnose → correct | Review *against intent*, not just against spec |
| 3: Live Build | Apply the pattern end-to-end | Add intent preamble to every spec, reflect at the end |

### Series Summary

| Session | You Learned | You Practiced |
|---------|-------------|---------------|
| Part 1 | Spec-driven development | Writing a planning doc |
| Part 2 | The feedback loop | Diagnosing and correcting AI output |
| Part 3 | The full pattern live | Building, reviewing, shipping |
| Part 4 | Intent engineering | Writing intent blocks, auditing alignment, encoding values |

The evolution of the series mirrors the evolution of the industry: we started with how to give instructions (prompts/specs), moved to how to build the right environment (context/feedback), and now address the hardest question — what should the AI actually be optimizing for?

The engineers who answer that question explicitly will ship better work than the ones who leave it to chance.
