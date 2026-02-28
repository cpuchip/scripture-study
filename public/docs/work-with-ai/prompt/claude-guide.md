# Claude Opus 4.6 Prompting Best Practices

**Source:** [Anthropic Claude API Docs — Prompting Best Practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices)
**Captured:** February 2026
**Models covered:** Claude Opus 4.6, Claude Sonnet 4.6, Claude Haiku 4.5

---

## Overview

This is the single reference for prompt engineering with Claude's latest models. It covers foundational techniques, output control, tool use, thinking, and agentic systems.

---

## 1. General Principles

### 1.1 Be Clear and Direct

Claude responds well to clear, explicit instructions. Think of Claude as a brilliant but new employee who lacks context on your norms and workflows. The more precisely you explain what you want, the better the result.

**Golden rule:** Show your prompt to a colleague with minimal context on the task and ask them to follow it. If they'd be confused, Claude will be too.

- Be specific about the desired output format and constraints
- Provide instructions as sequential steps using numbered lists or bullet points when order matters
- If you want "above and beyond" behavior, explicitly request it rather than relying on inference

### 1.2 Add Context to Improve Performance

Providing context or motivation behind your instructions — explaining to Claude why such behavior is important — helps Claude better understand your goals and deliver more targeted responses. Claude is smart enough to generalize from the explanation.

### 1.3 Use Examples Effectively

Examples are one of the most reliable ways to steer Claude's output format, tone, and structure. A few well-crafted examples (few-shot or multishot prompting) can dramatically improve accuracy and consistency.

Best practices for examples:
- **Relevant:** Mirror your actual use case closely
- **Diverse:** Cover edge cases and vary enough that Claude doesn't pick up unintended patterns
- **Structured:** Wrap examples in `<example>` tags (multiple examples in `<examples>` tags)
- Include **3–5 examples** for best results
- You can ask Claude to evaluate your examples for relevance and diversity, or to generate additional ones

### 1.4 Structure Prompts with XML Tags

XML tags help Claude parse complex prompts unambiguously, especially when mixing instructions, context, examples, and variable inputs. Wrapping each type of content in its own tag (e.g., `<instructions>`, `<context>`, `<input>`) reduces misinterpretation.

- Use consistent, descriptive tag names across your prompts
- Nest tags when content has a natural hierarchy

### 1.5 Give Claude a Role

Setting a role in the system prompt focuses Claude's behavior and tone. Even a single sentence makes a difference:

```python
system="You are a helpful coding assistant specializing in Python."
```

### 1.6 Long Context Prompting (20K+ tokens)

When working with large documents or data-rich inputs:

- **Put longform data at the top** — place long documents and inputs near the top, above query/instructions/examples. Queries at the end can improve response quality by up to 30%
- **Structure with XML tags** — wrap each document in `<document>` tags with `<document_content>` and `<source>` subtags
- **Ground responses in quotes** — ask Claude to quote relevant parts of documents first before carrying out its task. Helps cut through noise.

### 1.7 Model Self-Knowledge

For correct self-identification:

```
The assistant is Claude, created by Anthropic. The current model is Claude Opus 4.6.
```

For LLM-powered apps needing model strings:

```
When an LLM is needed, please default to Claude Opus 4.6 unless the user requests otherwise. The exact model string for Claude Opus 4.6 is claude-opus-4-6.
```

---

## 2. Output and Formatting

### 2.1 Communication Style and Verbosity

Claude's latest models have a more concise and natural communication style:
- More direct and grounded — fact-based progress reports rather than self-celebratory updates
- More conversational — slightly more fluent and colloquial, less machine-like
- Less verbose — may skip detailed summaries for efficiency unless prompted otherwise

Claude may skip verbal summaries after tool calls. If you prefer more visibility:

```
After completing a task that involves tool use, provide a quick summary of the work you've done.
```

### 2.2 Control the Format of Responses

Four effective techniques:

1. **Tell Claude what to do instead of what not to do**
   - Instead of: "Do not use markdown in your response"
   - Try: "Your response should be composed of smoothly flowing prose paragraphs."

2. **Use XML format indicators**
   - Try: "Write the prose sections in `<smoothly_flowing_prose_paragraphs>` tags."

3. **Match your prompt style to the desired output** — formatting style in the prompt influences response style

4. **Use detailed prompts for specific formatting** — e.g., explicit guidance to minimize markdown:

```xml
<avoid_excessive_markdown_and_bullet_points>
When writing reports, documents, technical explanations, analyses, or any long-form content,
write in clear, flowing prose using complete paragraphs and sentences. Use standard paragraph
breaks for organization and reserve markdown primarily for inline code, code blocks, and
simple headings. Avoid using bold and italics.

DO NOT use ordered lists or unordered lists unless: a) you're presenting truly discrete items
where a list format is the best option, or b) the user explicitly requests a list or ranking.
</avoid_excessive_markdown_and_bullet_points>
```

### 2.3 LaTeX Output

Claude Opus 4.6 defaults to LaTeX for mathematical expressions. For plain text, add:

```
Format your response in plain text only. Do not use LaTeX, MathJax, or any markup notation.
Write all math expressions using standard text characters (e.g., "/" for division, "*" for
multiplication, and "^" for exponents).
```

### 2.4 Document Creation

Claude's latest models excel at creating presentations, animations, and visual documents. For best results:

```
Create a professional presentation on [topic]. Include thoughtful design elements,
visual hierarchy, and engaging animations where appropriate.
```

### 2.5 Migrating Away from Prefilled Responses

Starting with Claude 4.6, prefilled responses on the last assistant turn are no longer supported. Model intelligence and instruction following have advanced such that most prefill use cases no longer require it. Common migration scenarios: controlling output formatting, eliminating preambles, avoiding bad refusals, continuations, context hydration and role consistency.

---

## 3. Tool Use

### 3.1 Explicit Instructions for Action

Claude's latest models follow instructions precisely. If you say "can you suggest some changes," Claude will sometimes provide suggestions rather than implementing — even if making changes is what you intended.

**For proactive action by default:**

```xml
<default_to_action>
By default, implement changes rather than only suggesting them. If the user's intent is
unclear, infer the most useful likely action and proceed, using tools to discover any missing
details instead of guessing.
</default_to_action>
```

**For conservative action:**

```xml
<do_not_act_before_instructions>
Do not jump into implementation or changes files unless clearly instructed. When the user's
intent is ambiguous, default to providing information, doing research, and providing
recommendations rather than taking action.
</do_not_act_before_instructions>
```

**Important:** Claude Opus 4.5 and Opus 4.6 are more responsive to the system prompt than previous models. If your prompts were designed to reduce undertriggering on tools, these models may now overtrigger. Dial back aggressive language — replace "CRITICAL: You MUST use this tool when..." with "Use this tool when..."

### 3.2 Optimize Parallel Tool Calling

Claude's latest models excel at parallel tool execution:
- Run multiple speculative searches during research
- Read several files at once to build context faster
- Execute bash commands in parallel

```xml
<use_parallel_tool_calls>
If you intend to call multiple tools and there are no dependencies between the calls, make
all independent calls in parallel. Prioritize calling tools simultaneously whenever the
actions can be done in parallel rather than sequentially. However, if some calls depend on
previous calls to inform dependent values, do NOT call these in parallel — call them
sequentially. Never use placeholders or guess missing parameters.
</use_parallel_tool_calls>
```

---

## 4. Thinking and Reasoning

### 4.1 Overthinking and Excessive Thoroughness

Claude Opus 4.6 does significantly more upfront exploration than previous models, especially at higher `effort` settings. If this is excessive:

- Replace blanket defaults with targeted instructions. Instead of "Default to using [tool]," → "Use [tool] when it would enhance your understanding of the problem."
- Remove over-prompting. Tools that undertriggered in previous models now trigger appropriately. "If in doubt, use [tool]" will cause overtriggering.
- Use `effort` as a fallback — lower settings reduce overall thinking and token usage.

**To constrain reasoning:**

```
When deciding how to approach a problem, choose an approach and commit to it. Avoid
revisiting decisions unless you encounter new information that directly contradicts your
reasoning. If weighing two approaches, pick one and see it through.
```

### 4.2 Adaptive & Interleaved Thinking

Claude Opus 4.6 uses **adaptive thinking** (`thinking: {type: "adaptive"}`), where Claude dynamically decides when and how much to think. Claude Sonnet 4.6 supports both adaptive and manual extended thinking with interleaved mode.

Claude calibrates thinking based on two factors: the `effort` parameter and query complexity. Higher effort elicits more thinking; more complex queries do the same. On easier queries, the model responds directly.

**In internal evaluations, adaptive thinking reliably drives better performance than extended thinking.**

To guide thinking behavior:

```
After receiving tool results, carefully reflect on their quality and determine optimal
next steps before proceeding. Use your thinking to plan and iterate based on this new
information, and then take the best next action.
```

Key guidance:
- Prefer general instructions over prescriptive steps — "think thoroughly" often produces better reasoning than hand-written step-by-step plans
- Multishot examples work with thinking — use `<thinking>` tags inside few-shot examples
- Manual CoT as a fallback — when thinking is off, use `<thinking>` and `<answer>` tags to separate reasoning from output
- Ask Claude to self-check — "Before you finish, verify your answer against [test criteria]"

**Migration from extended thinking:**

```python
# Before (extended thinking)
thinking={"type": "enabled", "budget_tokens": 32000}

# After (adaptive thinking)
thinking={"type": "adaptive"}
output_config={"effort": "high"}  # or max, medium, low
```

---

## 5. Agentic Systems

### 5.1 Long-Horizon Reasoning and State Tracking

Claude's latest models excel at long-horizon reasoning with exceptional state tracking. Claude maintains orientation across extended sessions by focusing on incremental progress — steady advances on a few things at a time.

#### Context Awareness and Multi-Window Workflows

Claude 4.6 and 4.5 feature **context awareness** — the model can track its remaining context window throughout a conversation.

```
Your context window will be automatically compacted as it approaches its limit, allowing
you to continue working indefinitely. Do not stop tasks early due to token budget concerns.
As you approach your token budget limit, save your current progress and state to memory
before the context window refreshes. Always be as persistent and autonomous as possible.
Never artificially stop any task early regardless of the context remaining.
```

#### Multi-Context Window Workflows

For tasks spanning multiple context windows:

1. **Use a different prompt for the first window** — set up framework (write tests, create setup scripts), then iterate on a todo-list in future windows
2. **Write tests in structured format** — create tests before starting work (e.g., `tests.json`). Remind: "It is unacceptable to remove or edit tests"
3. **Set up quality of life tools** — create setup scripts (`init.sh`) to prevent repeated work
4. **Start fresh vs compacting** — Claude's latest models are extremely effective at discovering state from the local filesystem. Be prescriptive:
   - "Call pwd; you can only read and write files in this directory."
   - "Review progress.txt, tests.json, and the git logs."
5. **Provide verification tools** — Playwright MCP, computer use for testing UIs
6. **Encourage complete usage of context:**

```
This is a very long task. Plan your work clearly. Spend your entire output context working —
don't run out of context with significant uncommitted work. Continue systematically until complete.
```

#### State Management Best Practices

- **Structured formats** for state data (JSON for test results, task status)
- **Unstructured text** for progress notes
- **Git for state tracking** — provides log of what's been done with restorable checkpoints
- **Emphasize incremental progress** — explicitly ask to track progress and focus on incremental work

### 5.2 Balancing Autonomy and Safety

Without guidance, Claude Opus 4.6 may take actions that are difficult to reverse. Add reversibility guidance:

```xml
<reversibility>
Consider the reversibility and potential impact of your actions. Take local, reversible
actions freely (editing files, running tests), but for actions that are hard to reverse,
affect shared systems, or could be destructive, ask before proceeding.

Examples requiring confirmation:
- Destructive: deleting files/branches, dropping tables, rm -rf
- Hard to reverse: git push --force, git reset --hard, amending published commits
- Visible to others: pushing code, commenting on PRs/issues, sending messages

When encountering obstacles, do not use destructive actions as a shortcut.
</reversibility>
```

### 5.3 Research and Information Gathering

Claude's latest models demonstrate exceptional agentic search capabilities. For optimal research:

1. Provide clear success criteria
2. Encourage source verification across multiple sources
3. Use a structured approach for complex tasks:

```
Search in a structured way. Develop several competing hypotheses. Track confidence levels.
Regularly self-critique your approach. Update a hypothesis tree or research notes file.
Break down the research task systematically.
```

### 5.4 Subagent Orchestration

Claude's latest models have significantly improved native subagent orchestration. They can recognize when tasks benefit from delegating to specialized subagents.

**Watch for overuse:** Claude Opus 4.6 has a strong predilection for subagents and may spawn them when a simpler approach would suffice (e.g., spawning a subagent for code exploration when a direct grep is faster).

```
Use subagents when tasks can run in parallel, require isolated context, or involve
independent workstreams that don't need to share state. For simple tasks, sequential
operations, single-file edits, or tasks where you need to maintain context across steps,
work directly rather than delegating.
```

### 5.5 Chain Complex Prompts

With adaptive thinking and subagent orchestration, Claude handles most multi-step reasoning internally. Explicit prompt chaining (breaking into sequential API calls) is still useful when you need to inspect intermediate outputs or enforce a specific pipeline structure.

Most common pattern: **generate → review against criteria → refine**. Each step is a separate API call so you can log, evaluate, or branch.

### 5.6 Reduce File Creation in Agentic Coding

Claude may create new files as a "temporary scratchpad" — this can improve outcomes for agentic coding. To minimize:

```
If you create any temporary files, scripts, or helper files for iteration, clean them up
by removing them at the end of the task.
```

### 5.7 Overeagerness

Claude Opus 4.5 and 4.6 tend to overengineer — creating extra files, adding unnecessary abstractions, building in flexibility that wasn't requested.

```xml
<minimal_changes>
Avoid over-engineering. Only make changes that are directly requested or clearly necessary.

- Scope: Don't add features, refactor code, or make "improvements" beyond what was asked.
- Documentation: Don't add docstrings, comments, or type annotations to unchanged code.
- Defensive coding: Don't add error handling for scenarios that can't happen.
- Abstractions: Don't create helpers for one-time operations. Don't design for hypothetical futures.
</minimal_changes>
```

### 5.8 Avoid Hard-Coding and Test-Focused Solutions

Claude can focus too heavily on making tests pass at the expense of general solutions.

```
Write high-quality, general-purpose solutions. Do not hard-code values or create solutions
that only work for specific test inputs. Implement the actual logic that solves the problem
generally. Tests verify correctness, not define the solution. If the task is unreasonable or
tests are incorrect, inform me rather than working around them.
```

### 5.9 Minimizing Hallucinations in Agentic Coding

```xml
<investigate_before_answering>
Never speculate about code you have not opened. If the user references a specific file,
you MUST read the file before answering. Investigate and read relevant files BEFORE
answering questions about the codebase. Never make claims about code before investigating.
</investigate_before_answering>
```

---

## 6. Capability-Specific Tips

### 6.1 Improved Vision Capabilities

Claude Opus 4.5 and 4.6 have improved vision capabilities — better image processing and data extraction, especially with multiple images. Works for video analysis (break into frames). Giving Claude a crop tool provides consistent uplift.

### 6.2 Frontend Design

Without guidance, models can default to generic "AI slop" aesthetic. To create distinctive frontends:

```xml
<frontend_aesthetics>
Focus on:
- Typography: Choose beautiful, unique fonts. Avoid generic fonts like Arial and Inter.
- Color & Theme: Commit to a cohesive aesthetic. Use CSS variables. Dominant colors with
  sharp accents outperform timid palettes.
- Motion: Use animations for effects and micro-interactions.
- Backgrounds: Create atmosphere and depth rather than defaulting to solid colors.

Avoid: Overused font families (Inter, Roboto, Arial), clichéd color schemes (purple
gradients on white), predictable layouts, cookie-cutter design.

Think outside the box. Vary between light and dark themes, different fonts, different aesthetics.
</frontend_aesthetics>
```

---

## 7. Migration Considerations

When migrating to Claude 4.6 from earlier models:

1. **Be specific about desired behavior** — describe exactly what you'd like
2. **Frame instructions with modifiers** — instead of "Create an analytics dashboard," use "Create an analytics dashboard. Include as many relevant features and interactions as possible. Go beyond the basics."
3. **Request features explicitly** — animations and interactive elements should be asked for
4. **Update thinking configuration** — Claude 4.6 uses adaptive thinking (`thinking: {type: "adaptive"}`) instead of manual thinking with `budget_tokens`. Use the `effort` parameter for depth control.
5. **Migrate away from prefilled responses** — deprecated in Claude 4.6
6. **Tune anti-laziness prompting** — Claude 4.6 is significantly more proactive. Instructions needed for previous models will cause overtriggering.

### Migrating from Claude Sonnet 4.5 to Claude Sonnet 4.6

Sonnet 4.6 defaults to `effort: high` (unlike Sonnet 4.5 which had no effort parameter). Adjust accordingly or expect higher latency.

**Recommended effort settings:**
- **Medium** for most applications
- **Low** for high-volume or latency-sensitive workloads
- Set large max output token budget (64k recommended) at medium or high effort

**When to use Opus 4.6 instead:** For the hardest, longest-horizon problems — large-scale code migrations, deep research, extended autonomous work.

#### Without Extended Thinking

Continue without it. Set effort explicitly. At `low` effort with thinking disabled, expect similar or better performance vs Sonnet 4.5.

#### With Extended Thinking

Extended thinking continues to be supported. Keep thinking budget around 16k tokens.

- Coding use cases: Start with `medium` effort
- Chat/non-coding: Start with `low` effort with extended thinking

#### When to Try Adaptive Thinking

Consider adaptive when:
- **Autonomous multi-step agents** — coding agents, data analysis pipelines. Start at `high` effort.
- **Computer use agents** — Claude Sonnet 4.6 achieved best-in-class accuracy using adaptive mode.
- **Bimodal workloads** — mix of easy and hard tasks where adaptive skips thinking on simple queries and reasons deeply on complex ones.

---

## Key Takeaways for Our Work

| Claude Principle | Maps to Our Practice |
|-----------------|---------------------|
| Be clear and direct | Intent preambles on every document |
| Add context | Context engineering via MCP servers, skills, instructions |
| Use examples (3-5) | Multishot prompting in agent definitions |
| XML tags for structure | Already use XML tags extensively in copilot-instructions |
| Give Claude a role | 9 custom agents with specialized roles |
| Long context — data at top, query at bottom | Structure study documents with sources first |
| Adaptive thinking | Let Claude calibrate per-task |
| Subagent orchestration | Used but watch for overuse |
| Investigate before answering | Source verification skill enforces this |
| State management with git | Already version-controlled |
| Reversibility guidance | Could add covenant-style decision boundaries |
| Evaluation design | Eval agent mode, but could formalize evals |

---

*This is a reference copy for the [Working with AI Guide Series](00_guide-plan.md). The original lives at [platform.claude.com](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices).*
