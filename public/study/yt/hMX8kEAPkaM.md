# The wrong abstraction for AI


## Thesis

Jake Van Clief argues that "agents" are the wrong abstraction layer for building AI systems. The current industry pattern — routing data and tools to agent frameworks — inverts the right design priority. Instead, he proposes structuring data so that a single model can dynamically become "as many agents as you need" simply by navigating that data structure. In this view, context, prompts, and tools are not external attachments to an agent wrapper; they are intrinsic to the data structure itself, emerging naturally as the model traverses it.

## How it builds

The argument unfolds in three moves:

1. **Observation of scale**: Van Clief notes that a single research paper has drawn 36,000+ community members and 10M+ monthly views, suggesting strong market resonance for his alternative approach.

2. **Critique of the current pattern**: "Right now, everyone is taking data and tools and routing them to the agent." He identifies this as the dominant paradigm and calls it "the wrong abstraction."

3. **Proposed inversion**: "If you structure the data right, one model can become as many agents as you need as it navigates that data." Context, prompts, and tools exist inside the data structure rather than being bolted on. This makes the system model-agnostic and enables "a human in a much deeper place inside of the compute layer."

## Key passages

1. **"Agents are the wrong abstraction layer."** — The opening thesis, stated bluntly in the first second.

2. **"Right now, everyone is taking data and tools and routing them to the agent."** — The diagnosis of the prevailing paradigm: agents as the central hub that everything feeds into.

3. **"If you structure the data right, one model can become as many agents as you need as it navigates that data."** — The core alternative: data structure as the source of agent-like behavior, not a separate wrapper.

4. **"The context, the prompts, the tools, all exist inside the structure."** — The inversion: these are not external attachments but intrinsic to the data layout.

5. **"This also allows you to have a human in a much deeper place inside of the compute layer."** — The human-in-the-loop benefit: when the structure carries the intelligence, humans can intervene at the compute level rather than at the agent orchestration level.

## Themes

- **Abstraction layer selection matters**: Choosing where intelligence lives (in the agent wrapper vs. in the data structure) has cascading effects on flexibility, cost, and human leverage.
- **Data structure as program**: The idea that well-designed data can encode behavior, context, and tooling — making the model a navigator rather than a general-purpose worker.
- **Model agnosticism**: When intelligence is in the data structure, you're not locked into a specific model provider's agent framework.
- **Human-in-the-compute-layer**: A more efficient form of human oversight — intervening where the raw work happens rather than at the orchestration boundary.
- **Scale of ideas over scale of infrastructure**: The speaker's own 36K-member community built on a single paper suggests that conceptual clarity can outperform engineering complexity in market impact.

## Tensions & objections

The strongest objection: **data structure design is itself a form of agent engineering, just shifted upstream.** If you're designing a data structure that encodes context, prompts, and tool routing, you're still building an agent — you've just moved the complexity from the orchestration layer into the data layer. This doesn't eliminate the agent problem; it redistributes it.

Specific concerns:

- **Composability**: Can a single model truly navigate arbitrary data structures the way a purpose-built agent framework can compose tools, manage state, and handle error recovery? The claim that "one model can become as many agents as you need" is elegant but unproven at scale.

- **The hard problem remains**: Even with perfect data structure, you still need to solve alignment, reliability, and evaluation. A well-structured dataset doesn't prevent hallucination or goal drift.

- **Human-in-the-compute-layer is hard**: Van Clief claims this is "efficient" but doesn't explain how humans can effectively intervene at the compute layer without becoming the bottleneck the agent was supposed to solve.

- **The appeal may be premature**: The 36K community members and 10M views reflect excitement about a paradigm shift, not evidence that it works better in production. The agent abstraction, for all its flaws, has real engineering value in standardizing tool use, state management, and error handling.

## What's worth learning

1. **Audit your agent architecture for inversion**: When you design an agent system, ask: "Am I routing data to the agent, or am I structuring data so the agent can navigate it?" The former makes the agent a black box; the latter makes the structure the artifact you can reason about and iterate on.

2. **Design data structures that encode behavior**: Treat your data schema as a programming language. Can your data format express context windows, tool calls, and routing logic natively? If not, you're fighting the abstraction.

3. **Decouple from model providers early**: If your agent design is tightly coupled to a specific provider's tool-calling format or framework, you've built fragility into your core. Structure-first design keeps you portable.

4. **Identify where humans can actually add value in the compute layer**: Map the points in your pipeline where human judgment beats automation. Design your data structure to make those intervention points natural — not afterthoughts bolted onto an agent workflow.

5. **Test the single-model hypothesis**: Before investing in multi-agent orchestration, try building your system with one model navigating a well-designed data structure. Measure whether the complexity you're adding with agents actually improves outcomes, or just adds layers of indirection.