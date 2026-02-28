# Part 1: Prompt Craft — The Foundation Layer

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [eval.md](../prompt/eval.md) § Discipline 1, [Claude Prompting Best Practices](../prompt/claude-guide.md)
**Core thesis:** Prompt craft is table stakes. If you can't write a clear, well-structured prompt, nothing else matters. But most people overestimate their skill here.

---

## The Skill Everyone Thinks They Have

Everyone who uses AI thinks they can prompt. They type a question, they get a response, they move on. What's hard about that?

What's hard is the gap between "I got a response" and "I got the *right* response on the first try." That gap costs hours of back-and-forth, manual editing, and growing frustration that "AI isn't that useful."

Prompt craft is the discipline of closing that gap — writing instructions clear enough that the model produces what you actually need, usually on the first or second attempt. It's synchronous (you and the model, right now), session-based (lasting minutes to hours), and individual (your skill, your prompt).

[Nate B Jones](https://www.youtube.com/watch?v=BpibZSMGtdY&t=611) puts it bluntly:

> "This is the skill I have taught and many others have taught for the last year or two. It's synchronous. It's session-based and it's an individual skill... It's sort of the way knowing how to type with 10 fingers was once a differentiator and now it's just assumed."

Table stakes doesn't mean unimportant. It means *prerequisite.* You can't build context, intent, or specification on top of poor prompt craft, just like you can't write a novel if you can't write a sentence.

---

## The Golden Rule

Anthropic's [Claude Prompting Best Practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices) calls this the golden rule:

> **Show your prompt to a colleague.** If they would be confused about what to do, Claude will be too.

This is the single most useful test for prompt quality. Before you blame the model, show your prompt to a human. If they'd ask "what do you mean by X?" or "what format do you want?" or "who is this for?" — the model is asking the same questions internally and guessing at the answers.

Most "bad AI output" is accurate AI execution of an ambiguous prompt.

---

## The Six Principles

Based on Anthropic's engineering guidance, Nate's framework, and our own testing across thousands of interactions, here are the six principles of effective prompt craft:

### 1. Be Clear and Direct

State what you want, not what you don't want.

**Weak:**
> Don't write a long response. Don't use jargon. Don't be too formal.

**Strong:**
> Write a 3-sentence response in conversational language a non-technical reader would understand.

The model is a completion engine — it naturally extends what you give it. When you say "don't think about elephants," you've introduced elephants. When you say "don't be formal," you've primed the model with formality. Positive instructions produce better results than negative constraints.

Tell the model what *to do*:
- Instead of "Don't use bullet points" → "Write in paragraph form"
- Instead of "Don't be verbose" → "Be concise — under 100 words"
- Instead of "Don't make assumptions" → "Ask clarifying questions before proceeding"

### 2. Provide Context Before the Task

Place your context at the top and your question at the bottom. Anthropic's research shows this ordering produces up to 30% better responses in long-context scenarios.

**Structure:**
```
Here's the context: [background, documents, data]
Here's what I need: [the task]
Here's how I want it: [format, length, tone]
```

This mirrors how you'd brief a colleague. You wouldn't walk up and say "Write the report" — you'd say "We have a client meeting Thursday about the Q3 data, and the audience is executives who haven't seen the raw numbers. I need a one-page summary highlighting the three biggest changes."

Context → Task → Format. Every time.

### 3. Use Examples (3-5 is the sweet spot)

Examples are the most reliable way to steer output. Anthropic's guide calls them "one of the most reliable ways to get Claude to produce output in the exact format and style you need."

The pattern:
```
Here's what I'm looking for:

<example>
Input: [sample input]
Output: [sample output you'd want]
</example>

<example>
Input: [different sample]
Output: [corresponding output]
</example>
```

3-5 examples covers the pattern without overfitting. Include at least one edge case — an example that shows how to handle unusual input. If your examples only show the happy path, the model will only handle the happy path.

**Why examples beat descriptions:** You could write three paragraphs describing the tone you want. Or you could show one example. The example wins. Models learn from patterns more reliably than from abstractions.

### 4. Give the Model a Role

A role isn't just flavor text. It activates relevant knowledge and adjusts the model's language, depth, and assumptions.

**Generic:**
> Explain how TCP/IP works.

**With role:**
> You are a senior network engineer writing documentation for junior developers who are learning about networking for the first time. Explain how TCP/IP works with practical examples they'd encounter in web development.

The role sets:
- **Knowledge depth** (senior engineer → draw on deep networking knowledge)
- **Audience calibration** (for junior developers → appropriate level of simplification)
- **Communication style** (documentation → structured, clear, example-driven)
- **Domain framing** (web development → focus on HTTP, sockets, not industrial protocols)

In system prompts (which we'll cover more in [Part 2](02_context-engineering.md)), the role goes first. It's the frame through which everything else is interpreted.

### 5. Structure Your Prompts

For complex tasks, use structured formatting. Anthropic specifically recommends XML tags:

```xml
<context>
We're building a REST API for a task management application.
The API uses Express.js and PostgreSQL.
Authentication is handled via JWT tokens.
</context>

<task>
Design the database schema for the tasks table.
Include fields for: title, description, status, priority, assignee, timestamps.
</task>

<constraints>
- Use PostgreSQL-specific data types where appropriate
- Include indices for common queries (by status, by assignee)
- Follow the project's naming convention: snake_case for columns
</constraints>

<output_format>
Provide the CREATE TABLE statement followed by CREATE INDEX statements.
Add comments explaining design decisions.
</output_format>
```

XML tags work because they create unambiguous boundaries. The model never confuses the context with the task, or the constraints with the output format. For simpler prompts, clear headings or paragraphs suffice. But as complexity grows, structure saves rework.

### 6. Self-Contained Problem Statements

This is [Nate's first specification primitive](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2090), but it applies at the prompt level too. Every prompt should contain everything the model needs to respond well — or explicitly tell the model where to find what it needs.

**Self-contained means:**
- The reader doesn't need to guess what project this is for
- The terminology is defined or obvious
- The success criteria are stated ("I'll know this is good when...")
- Edge cases are addressed ("If X, then do Y")

This matters more than you think. When you're deep in a project and you know the context, it's easy to write prompts that assume knowledge the model doesn't have. The self-containment test catches this: *Could someone with no context about this project execute this prompt?*

---

## The Prompt Library

Here's a practice most people skip and probably shouldn't: **build a prompt library.**

Nate recommends this at [34:50](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2090):

> "You should be building a folder of tasks that you do regularly, writing your best prompt against each one."

A prompt library is a collection of your best prompts for recurring tasks — saved, tested, and refined over time. Not a collection of internet templates. *Your* prompts, for *your* tasks, refined through *your* experience.

Examples of what goes in a prompt library:
- **Code review prompt:** Your preferred format for asking the model to review code (what to check, what depth, what format for findings)
- **Documentation prompt:** How you like technical docs structured with proper examples
- **Email drafts:** Your brand voice, typical tone, audience assumptions
- **Meeting summaries:** The format you want, what to highlight, what to omit
- **Bug investigation:** How you want the model to approach debugging (hypothesize first, then verify)

Each prompt in the library should be:
- **Tested** — you've used it multiple times and refined it
- **Versioned** — you update it as you learn what works better
- **Documented** — notes on why certain choices work ("examples in this format reduce hallucination on this task type")

This is prompt craft at its most disciplined. You're not just getting good at writing prompts in the moment — you're building institutional knowledge about what works.

---

## Format Control and Output Steering

One underappreciated aspect of prompt craft: **tell the model exactly what the output should look like.**

Models will match whatever format you imply. If your prompt is a casual paragraph, the response will be a casual paragraph. If your prompt includes a structured template, the response will follow that template.

**Explicit format instructions:**
- "Respond in JSON with the following structure: { ... }"
- "Use markdown with H2 headings for each section"
- "Return only the code, no explanations"
- "Format as a table with columns: Name | Type | Description"

**Pre-fill technique:** Start the response for the model.
```
Based on the analysis above, here are the top 3 recommendations:

1.
```

When you put the opening of the desired output at the end of your prompt, the model continues from there — already in the format you want. Anthropic calls this "prefilling Claude's response."

---

## The Prayer Parallel

There's a deeper pattern here, and it's worth naming.

Prompt craft — clear communication with an intelligent entity to receive guidance and help — has an ancient parallel. Prayer.

> "Ask, and it shall be given you; seek, and ye shall find; knock, and it shall be opened unto you."
> — Matthew 7:7

Prayer isn't just "talking to God." Effective prayer, as taught in scripture and by prophets, shares the principles of effective prompting:

| Principle | In Prayer | In Prompting |
|-----------|-----------|-------------|
| **Clarity** | "Ask with real intent" (Moroni 10:4) | State what you want, not what you don't want |
| **Context** | "Make known thy cause" (Alma 34:24-25) — tell God the situation | Provide background before the task |
| **Specificity** | Enos prayed for his people by name and need (Enos 1:9-13) | Be specific about what you need |
| **Alignment** | "Thy will be done" (Matthew 6:10) — aligning with God's purposes | Aligning the request with the system's capabilities |
| **Preparation** | "Study it out in your mind; then ask" (D&C 9:8) | Do your homework before prompting |
| **Humility** | "With a sincere heart" (Moroni 10:4) | Acknowledge uncertainty, ask for help |

The parallel isn't coincidental. Both involve communicating with an intelligence that can help you accomplish something beyond your individual capacity. The principles of effective communication are universal.

Note where the parallel breaks: prayer activates the Holy Ghost and involves a personal relationship with a living God. Prompting activates a language model and involves a tool. The principles overlap because clear communication is universal — but the *power source* is categorically different. Don't confuse the tool with the Teacher.

---

## Common Mistakes

### The "Quick and Dirty" Trap
"I'll just ask it real quick" — and then spend 30 minutes iterating on vague output. The 2 minutes you "saved" by not writing a clear prompt cost you 28 minutes of rework. Every time.

### The Overloaded Prompt
Cramming 5 different tasks into one prompt. The model optimizes for the last thing you said, losing earlier requirements. One prompt, one task. If you have 5 tasks, write 5 prompts or use numbered lists with clear separation.

### The "No Examples" Habit
Describing what you want in abstract terms when three examples would make it crystal clear. If you find yourself writing paragraphs of description, stop and show an example instead.

### The Negative Spiral
"Don't do X. Also don't do Y. And definitely don't Z." You've now primed the model with X, Y, and Z. Flip every negative constraint to a positive instruction.

### The Memory Assumption
Treating the model like a colleague who remembers last week's conversation. In a new session, the model knows nothing about your project unless you tell it. Every session starts from zero (unless you've built context infrastructure — that's [Part 2](02_context-engineering.md)).

---

## Measuring Your Prompt Craft

Here's a simple self-assessment:

| Question | Score |
|----------|-------|
| Do more than 80% of my prompts produce usable output on the first try? | /10 |
| Do I routinely provide examples (not just descriptions) in complex prompts? | /10 |
| Can I hand my prompt to a colleague and have them understand what I want? | /10 |
| Do I have a collection of saved, tested prompts for recurring tasks? | /10 |
| Do I tell the model what TO do rather than what NOT to do? | /10 |

**35-50:** Strong prompt craft. You're ready to invest in higher-altitude skills.
**20-34:** Solid foundation with gaps. Focus on examples and format control.
**Under 20:** Start here. Practice the six principles on every prompt for two weeks.

---

## What Prompt Craft Can't Do

Prompt craft has a ceiling, and knowing where that ceiling is matters.

**Prompt craft can:**
- Get high-quality output from a single interaction
- Reduce iteration cycles from 5+ to 1-2
- Make every AI interaction more efficient
- Build a reusable library of proven prompts

**Prompt craft cannot:**
- Give the model knowledge it doesn't have (that's context engineering)
- Align the model with your organization's values (that's intent engineering)
- Enable autonomous multi-hour agent work (that's specification engineering)
- Persist knowledge across sessions (that's context infrastructure)

The ceiling of prompt craft is the ceiling of synchronous, session-based, individual interaction. To break through, you need to climb to the next altitude.

---

## Become

If prompt craft is really about clear communication — and if clear communication is really about clear thinking — then practicing prompt craft is practicing clarity of thought.

Every time you force yourself to write a self-contained problem statement, you're practicing the discipline of complete thinking. Every time you provide examples instead of vague descriptions, you're practicing concreteness. Every time you state what you want rather than what you don't want, you're practicing positive framing.

These are life skills wearing a technical hat.

The next time you write a prompt, don't just think about what the model needs. Think about what the discipline is teaching *you* about how you communicate.

---

*Previous: [Part 0 — The Foundation](00_foundation.md) | Next: [Part 2 — Context Engineering](02_context-engineering.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
