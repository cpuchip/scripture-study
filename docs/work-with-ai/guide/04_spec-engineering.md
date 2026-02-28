# Part 4: Specification Engineering — The Blueprint Layer

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [Planning Then Create (Gospel)](../01_planning-then-create-gospel.md), [Synthesis: What to Build](../intent/04_synthesis.md), [Nate's 5 Primitives](https://www.youtube.com/watch?v=BpibZSMGtdY), [eval.md § Discipline 4](../prompt/eval.md)
**Core thesis:** The practical skill going forward is not writing code or crafting prompts — it's the ability to describe an outcome with enough precision that an autonomous system can execute against it for days.

---

## Spiritual Creation Before Temporal

The book of Abraham describes a creation pattern that most readers skim past:

> "And the Gods said: Let us prepare the earth to bring forth grass... And the Gods organized the earth to bring forth grass from its own seed... And the Gods saw that they were obeyed."
> — Abraham 4:11-12, 18

Two things stand out:

1. **"Let us prepare" comes before "organized."** The Gods discussed, planned, and designed before executing. There was a spiritual creation (the plan) before the temporal creation (the thing).

2. **"They were obeyed."** The specification was so precise that the elements "obeyed" — they produced exactly what was designed. This isn't vague intention. It's precision engineering.

Moses 3:5 makes it even more explicit:

> "For I, the Lord God, created all things, of which I have spoken, spiritually, before they were naturally upon the face of the earth."

The blueprint came first. Every plant, every animal, every system was spiritually (conceptually, precisely) created before it was physically manifest. This is specification engineering — and it's the pattern the industry is rediscovering as the critical skill for autonomous AI.

---

## Why Specifications Matter Now

In a world of synchronous prompting, specifications were optional. You could iterate — prompt, evaluate, adjust, repeat. The human was in the loop for every step, catching errors and steering course corrections in real time.

In a world of autonomous agents, specifications are **the quality ceiling.**

An autonomous agent will take your specification and work for hours — reading files, writing code, running tests, making hundreds of micro-decisions — without checking in. When it's done, you evaluate the result. If the specification was imprecise, vague, or incomplete, every one of those micro-decisions was a coin flip.

Nate's framing is stark:

> "Real-time prompting rewards verbal fluency. Specification engineering rewards completeness of thinking."

These are different skills. A person who is great at live conversation with AI (prompt craft) may be terrible at writing a document so complete that an agent can execute against it unattended. And increasingly, the second skill is the one that matters.

---

## The Five Specification Primitives

Nate B Jones identifies [five primitives](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2090) — the building blocks of any specification that an agent can execute against:

### Primitive 1: Self-Contained Problem Statements

The specification must include everything the agent needs to understand the problem. No "you know what I mean." No references to conversations that aren't in the document. No context that exists only in someone's head.

**Weak:**
> Fix the authentication bug we discussed yesterday.

**Strong:**
> **Problem:** Users who log in with SSO are assigned a new session token, but when they navigate to the /dashboard route, the middleware rejects the token with a 401 because the `iss` (issuer) claim doesn't match the expected value. This affects all SSO providers (Google, GitHub) but not email/password login.
>
> **Expected behavior:** SSO tokens should be validated with the SSO issuer claim, not the default JWT issuer.
>
> **Relevant files:** `src/middleware/auth.ts`, `src/services/sso.ts`, `src/config/jwt.ts`
>
> **Constraints:** Do not modify the token structure for email/password login. All existing tests must continue to pass.

The second version can be executed by anyone — human or agent — who has access to the codebase. The first version requires context that exists only in yesterday's conversation.

### Primitive 2: Acceptance Criteria

What "done" looks like. Not "done" as in "I stopped working on it," but "done" as in "this objectively meets the requirements."

**Weak:**
> The search should be faster.

**Strong:**
> **Acceptance criteria:**
> - Search queries return results in under 200ms at P95 (measured by server-side timing, not client-side)
> - Pagination works correctly for result sets exceeding 100 items
> - Empty search queries return a helpful message instead of an error
> - Search results are ranked by relevance, with exact title matches appearing first
> - The existing search API contract is maintained — no breaking changes for current consumers

Acceptance criteria transform subjective evaluation ("is this good?") into objective verification ("does it pass?"). An agent can validate its own work against these criteria. Without them, "done" is a feeling, and the agent doesn't have feelings.

### Primitive 3: Constraint Architecture

Constraints define the solution space — what's required, what's forbidden, what's preferred, and what triggers escalation.

```markdown
## Constraints

### Must (non-negotiable)
- All database queries use parameterized statements (no SQL injection surface)
- API responses include proper CORS headers
- Error responses follow RFC 7807 (Problem Details)
- No new dependencies without explicit approval

### Must Not
- Must not modify the users table schema
- Must not expose internal error details in production responses
- Must not bypass the rate limiter for any endpoint

### Prefer (when possible)
- Prefer existing utility functions over new implementations
- Prefer immutable data patterns
- Prefer descriptive variable names over comments

### Escalation Triggers
- If the implementation requires a database migration, stop and get approval
- If a trade-off affects the values hierarchy (reliability vs. speed), present options
- If test coverage for a changed file drops below 80%, flag it
```

This is the most underused primitive. Most specs give requirements (musts) and acceptance criteria but skip the boundaries (must-nots), preferences, and escalation triggers. The result: agents make unconstrained decisions in areas you assumed were obvious.

The handbook's structure of priesthood keys is functionally a constraint architecture:

> "Priesthood keys ensure that God's work of salvation and exaltation is accomplished in an orderly manner. Those who hold priesthood keys direct the Lord's work within their areas of responsibility. This presiding authority is valid only for the specific responsibilities of the leader's calling."
> — [General Handbook 3.4.1.2](../../gospel-library/eng/manual/general-handbook/3-priesthood-principles.md)

Valid only for the specific responsibilities. Clear scope. Clear boundaries. Clear escalation when something exceeds jurisdiction.

### Primitive 4: Decomposition

Complex work must be broken into pieces small enough to be independently verifiable. Nate's heuristic: **each task should be completable and verifiable in under 2 hours.**

**Why 2 hours?** Because:
- You can verify the output while it's still fresh
- The agent makes fewer compounding errors in a shorter scope
- If something goes wrong, you lose hours, not days
- Each completed task provides a checkpoint

**Decomposition format:**

```markdown
## Tasks

### Task 1: Add SSO issuer to JWT config (30 min)
- Modify src/config/jwt.ts to accept an array of valid issuers
- Add SSO_ISSUER to environment variables
- Update the config validation to require SSO_ISSUER when SSO is enabled
- **Verify:** Config loads correctly with both issuers; app starts without errors

### Task 2: Update auth middleware (45 min)
- Modify src/middleware/auth.ts to check token issuer against the valid issuers array
- Maintain backward compatibility for non-SSO tokens
- **Verify:** Existing email/password auth tests pass; SSO tokens are accepted

### Task 3: Add SSO-specific test cases (45 min)
- Add integration tests for SSO token validation
- Include edge cases: expired SSO token, valid SSO token, SSO token with wrong issuer
- **Verify:** All new tests pass; code coverage does not decrease
```

Each task is independently verifiable. Each task can be completed by an agent in one focused session. Each task builds on the previous but can be evaluated on its own.

### Primitive 5: Evaluation Design

The specification includes test cases with known good outputs — so the agent (or the human reviewer) can objectively verify the implementation.

```markdown
## Evaluation Cases

### Case 1: Happy path — SSO login
Input: Valid Google SSO token with `iss: "accounts.google.com"`
Expected: 200 response, session created, dashboard accessible
Verify: GET /dashboard returns 200 with valid session cookie

### Case 2: Edge case — wrong issuer
Input: Token with `iss: "attacker.com"`
Expected: 401 response, no session created
Verify: No entry in sessions table

### Case 3: Regression — email/password login
Input: Standard JWT from email/password flow
Expected: Unchanged behavior — 200, session, dashboard
Verify: All existing auth integration tests pass

### Case 4: Configuration error
Input: App started without SSO_ISSUER env var when SSO is enabled
Expected: App refuses to start, logs clear error about missing config
Verify: Process exits with code 1 and error message includes "SSO_ISSUER"
```

Evaluation design closes the loop. The spec says what to build (problem statement), when it's done (acceptance criteria), how to build it (constraints + decomposition), and how to verify it (evaluation cases). An agent with all five primitives can work autonomously, verify its own output, and present trustworthy results.

---

## The `.spec/` Directory

Where do specifications live? The emerging best practice: **a `.spec/` directory in your project root.**

```
.spec/
  intent.md          ← Project-level purpose, values, constraints
  spec.md            ← Current source-of-truth specification
  tasks/
    001-sso-fix.md   ← Task files with status, inherited intent
    002-search-perf.md
  learnings/
    001-auth-testing.md  ← Reusable knowledge from past work
  deltas/
    2026-02-15-sso.md    ← Proposed changes (OpenSpec-style)
  archive/
    2026-02-10-schema.md ← Applied deltas, kept for reference
```

This structure is:
- **File-based** — just markdown, version-controlled, diffable, travels with branches
- **Human-readable** — any team member can review specs in a PR
- **Agent-readable** — any AI tool can parse the structure
- **Intent-linked** — every task references the intent doc
- **State-bearing** — task status, learnings, and change history all visible

The alternative — specs in Jira tickets, Notion pages, Confluence docs — means specifications live separately from code, require a different tool to access, and are invisible during code review. The `.spec/` directory collapses that gap.

---

## Your Entire Organization as Specification

Here's the insight that most people miss: **specifications aren't just for code projects.**

Nate at [26:42](https://www.youtube.com/watch?v=BpibZSMGtdY&t=1602):

> "Your entire organizational corpus — every process document, every handoff doc, every playbook — those are all specs now. They need to be written with the same precision."

If you have an agent handling customer support, the support playbook is a specification. If you have an agent generating reports, the reporting standards are a specification. If you have an agent managing inventory, your inventory procedures are a specification.

Every document that an agent might read and act on is, functionally, a spec. And most organizational documents were written for humans who fill in gaps with common sense, institutional knowledge, and the ability to ask a colleague. Agents don't have those backfills.

This means specification engineering isn't just a developer skill. It's a **communication skill that every organization needs.** The marketing team needs to write brand guidelines precise enough for an AI to follow. The legal team needs to write compliance rules explicit enough for an agent to respect. The operations team needs to write procedures complete enough for autonomous execution.

The General Handbook of The Church of Jesus Christ of Latter-day Saints is, in this light, a remarkable example of specification engineering at organizational scale. Consider how it defines the ward council:

> "The ward council seeks to help all ward members build spiritual strength, receive saving ordinances, keep covenants, and become consecrated followers of Jesus Christ. During ward council meetings, council members plan and coordinate this work."
> — [General Handbook 29.2.5](../../gospel-library/eng/manual/general-handbook/29-meetings-in-the-church.md)

That's a purpose statement for a recurring process. The handbook then specifies participants, frequency, agenda structure, decision-making authority, and escalation paths. It's a specification that tens of thousands of bishops worldwide execute against, with local adaptation, producing consistent outcomes across wildly different contexts.

That's what good specification engineering looks like at scale.

---

## Planner-Worker Architecture

As autonomous AI matures, a pattern is emerging: **the spec determines the quality ceiling.**

In "planner-worker" architectures:
1. A **planner agent** reads the specification and decomposes it into tasks
2. **Worker agents** execute individual tasks against the spec
3. A **reviewer agent** validates results against acceptance criteria
4. The cycle repeats until all tasks pass evaluation

The planner never writes code. The workers never see the full project scope. The reviewer never generates output. Each agent has a stewardship — a defined domain of responsibility.

In this architecture, if the spec is vague, the planner generates vague tasks. The workers execute those vague tasks in random directions. The reviewer can't evaluate something against criteria that don't exist. *Everything* depends on the specification.

This is why Nate says specification engineering is the highest-leverage discipline. It's not the most glamorous (writing a spec is less exciting than watching an agent build a feature). But it determines the ceiling for everything else.

---

## Real-Time Prompting vs. Specification Engineering

These are different cognitive skills, and recognizing the difference matters:

| Dimension | Real-Time Prompting | Specification Engineering |
|-----------|-------------------|--------------------------|
| **Rewards** | Verbal fluency, speed of iteration | Completeness of thinking, precision |
| **Time horizon** | Minutes (session) | Hours–days (project) |
| **Error cost** | Wasted morning | Wasted sprint (or worse) |
| **Feedback loop** | Immediate (you see the output) | Delayed (agent works, then you evaluate) |
| **Skill type** | Conversational intelligence | Writing intelligence |
| **Analogy** | Live debugging | Architecture document |

A developer who's amazing in a pairing session with AI (great prompt craft) may write terrible specifications — too vague, too assumption-heavy, too focused on the obvious and neglecting the edges.

The specification engineering discipline demands a different mode of thinking: **imagine everything that could go wrong, everything that's ambiguous, every decision the agent will need to make, and every way to verify the result.** It's the kind of thinking that makes good architects and good test engineers. It's not flashy, but it's where the value is moving.

---

## How to Start

### Level 1: Spec-Before-Code
Before your next feature, write a spec. Not a Jira ticket — a specification with all five primitives. Even if you "could just code it," write the spec first. Then have an agent implement it. Compare the result to what you would have built.

### Level 2: Spec-Based Review
After an agent completes a task, evaluate it against the spec, not against your instinct. Did it meet the acceptance criteria? Did it honor the constraints? Did the evaluation cases pass? If it feels wrong but passes the spec, the spec is incomplete — update it.

### Level 3: Living Specs
Keep your `.spec/` directory current. When the implementation drifts from the spec, update the spec (or fix the implementation). Stale specs are worse than no specs because they teach agents (and humans) to ignore specifications.

---

## What Specification Engineering Can't Do

Specifications define *what* to build. They don't define *why* it matters.

A perfectly-engineered specification for the wrong feature is still the wrong feature. Specifications without intent produce correct implementations of incorrect objectives. That's why intent engineering ([Part 3](03_intent-engineering.md)) comes before specification engineering in the hierarchy — the spec inherits its purpose from the intent.

And specifications alone don't handle the dynamics of trust, error recovery, reflection, or organizational alignment. Those patterns — which live beyond specification — are the subject of [Part 5](05_complete-cycle.md).

---

## Become

Writing a specification is an act of humility. It's admitting that your intent — no matter how clear to you — is not automatically clear to others. It's doing the hard work of translating "I know what I want" into "anyone can execute this."

That's a spiritual discipline as much as a professional one. Brigham Young said:

> "You educate a man; you educate a man. You educate a woman; you educate a generation."

Education is specification for humans — transferring knowledge precisely enough that others can act on it independently. The clearer the specification, the more autonomy the receiver has. The vaguer the specification, the more they depend on you.

Good specs set people free.

---

*Previous: [Part 3 — Intent Engineering](03_intent-engineering.md) | Next: [Part 5 — The Complete Cycle](05_complete-cycle.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
