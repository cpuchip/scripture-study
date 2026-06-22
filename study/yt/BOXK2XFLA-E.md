# Don't build more AI agents until you watch this


## Thesis

The core claim: agents improve not by adding more tools and capabilities, but by pruning and maintaining the "harness" (the workbench of tools, permissions, prompts, and workflows) around them. Vercel's sales agent got better after deleting 80% of its tools, demonstrating that the harder question in 2026 is not whether you can build an agent, but whether you can keep the setup around it healthy as both the underlying model and the surrounding work environment change.

## How it builds

The video opens with Vercel's sales agent case study — built from observing a top performer's actual workflow, not a paper process. The agent filtered inbound messages, qualified leads, researched companies, drafted responses, and routed support questions, with human review. The key finding: the agent improved when tools were removed, not added.

From there the speaker lays out four principles:

1. **Models are moving** — the underlying model improves over time, so yesterday's harness can become wrong. A tool that helped a weaker model can confuse a stronger one; a guardrail that protected against an unreliable model can trap a better one.
2. **Agents inherit all the crud of the systems around them** — stale wikis, outdated processes, and wrong dashboard definitions are "very dangerous" because agents produce work from that mess convincingly.
3. **The biggest AI companies already know this** — OpenAI's Codex and Anthropic's Claude Code are strong not just because of their models, but because of carefully maintained harnesses. Better agents can help build better harnesses, creating a compounding flywheel.
4. **You need to ask: what is my harness?** — Every user of agentic tools has a harness (sources, prompts, permissions, verification loops). The mature question is what part of it will need to be deleted later.

The video closes with a five-point maintenance checklist and a recommendation of Stewart Brand's *The Maintenance of Everything* as the right frame for thinking about agents.

## Key passages

- "The beginner instinct is to add. The maintenance instinct is to ask what should be removed. That is the real agent story of 2026." — Frames the central shift from building to pruning.

- "Agents can also break when the model gets better and that is a different and new thing." — Challenges the conventional mental model that software only breaks when it gets worse.

- "A stale wiki that is annoying to you is incredibly dangerous to an agent because it doesn't know that and it just keeps on working." — Captures the unique risk agents pose when fed stale context.

- "Agents are a lot less like apps and more like sailboats. You don't just launch agents and walk away." — The sailboat analogy from Stewart Brand's *The Maintenance of Everything* as the right frame for agent lifecycle.

- "The companies that win are not the ones that build the perfect wrapper once. They're the ones that keep rebuilding the wrapper as the model and the work change." — The strategic implication for platform competition.

- "All the agent has to do is to keep working, and it will start to haunt your business." — The quiet danger of unmaintained agents producing convincing but wrong work.

## Themes

- **Harness as the real product** — The speaker repeatedly reframes the competitive landscape: what matters is not the model alone, but the workbench/harness around it. OpenAI and Anthropic are competing on harness quality, not just model quality.
- **Maintenance over building** — The dominant narrative is "can you build an agent?" The speaker argues the real question is "can you maintain it?" This includes pruning tools, updating sources, and adjusting permissions as models improve.
- **Two-directional breakage** — Agents break both when the world drifts (stale docs, changed processes) and when the model improves (old guardrails become constraints). This is a novel maintenance problem.
- **The flywheel** — Better models → better harnesses → more real work → more pressure to improve harnesses → better models. Platform companies that close this loop compound faster.
- **Simplicity as a maintenance virtue** — Referencing Stewart Brand: simplicity isn't just an aesthetic preference; it's a practical necessity for systems that must be maintained over time.
- **Personal harness awareness** — Even individual users have a "harness" (folders, memory, prompts, source docs, approval habits) that needs intentional maintenance.

## Tensions & objections

The strongest objection: **pruning is easy to advocate but hard to operationalize.** The video's core insight — that agents improve when you remove tools — is compelling in hindsight but offers little guidance on *what* to remove, *when*, or *how to test* that removal helped. In practice, teams face pressure to add capabilities (stakeholders want the agent to do more), and there's no clear signal that tells you a tool has become counterproductive until the agent starts producing bad work. The "maintenance instinct" requires ongoing measurement and experimentation that most organizations don't have infrastructure for.

Additionally, the Vercel case study is a single example from a well-resourced team that studied a top performer's workflow. The speaker acknowledges this is "not just a quirky sales automation story" but doesn't address how smaller teams without that luxury should approach harness design. The advice to "ask what is my harness" is valuable but abstract — many teams won't know how to answer it without the kind of observational rigor Vercel applied.

Finally, the flywheel argument (better models → better harnesses → more work → better harnesses) may be a moat for frontier labs but a trap for everyone else: if harness maintenance is the real competitive advantage, and only OpenAI and Anthropic are doing it well, then building custom agents may be a losing proposition for most organizations.

## What's worth learning

1. **Audit your agent's inputs regularly** — Check what sources the agent reads, whether they're current, and whether old sources have become misleading. Stale context is the agent's version of bad data.
2. **Review permissions after each model upgrade** — A permission that was harmless for a weaker model may be too broad for a stronger one, and a restriction that made sense before may now hold back a better model.
3. **Demand proof trails, not just conclusions** — Configure agents to link to sources, quote language, and report which sources they checked and which they couldn't access. An inspectable trail is the difference between trust and guesswork.
4. **Plan for deletion, not just addition** — When designing a harness, ask what parts will need to be removed as the model improves. Simplicity is a maintenance feature, not an aesthetic preference.
5. **Check whether the agent's job has silently changed** — Don't let an agent drift from summary to planning to recommendation without an intentional decision. Change the job on purpose if at all.
6. **Measure actual value, not just output** — Ask whether anyone reads the agent's output, whether it changes work, whether it saves time after review, or whether it creates another pile of work to manage.