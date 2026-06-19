# Did Google steal my research?

- **The core thesis / claim**
The speaker claims that Google researchers independently arrived at the same architectural insight he published three months prior: agentic AI systems are more effectively built using simple folder hierarchies, Markdown files, and YAML front matter than via complex contemporary frameworks. He argues this is not a novel invention but a rediscovery of Unix-era design principles that he has already validated across hundreds of workspaces.

- **How it builds**
The talk opens with a provocative accusation of idea theft, then immediately softens it to parallel discovery. The speaker establishes chronological priority (a 21-page paper from three months ago), defines the minimalist technical approach (folders/YAML/Markdown), characterizes it as an "LLM wiki-style pattern," cites two Google researchers' recent paper as evidence of convergence, claims empirical superiority over competing agentic systems, reframes the overlap as inevitable because the approach is simply vintage Unix methodology, concedes that Google's specific tools may be worth integrating, and closes by inviting the researchers to collaborate.

- **Key passages**
> "Google stole my research. Or at the very least, their researchers are on the same idea that I had 3 months ago."
Gloss: The opening frames the talk as a priority dispute while allowing for independent discovery.

> "I believed that using folders, YAML, and Markdown was way more efficient to send AI to use tools and data than all these crazy things that people were building with agentics and agentic frameworks and stuff like that."
Gloss: The central technical argument that plain-text file systems outperform elaborate agentic frameworks.

> "And they specifically say, 'Just Markdowns, just files, just YAML front matter.'"
Gloss: Evidence that Google's recent work employs the exact same primitives he advocated.

> "I built hundreds of workspaces using this exact methodology, crushing any agentic systems that people have built."
Gloss: A claim of extensive practical validation and superior real-world performance.

> "I'm not angry because this is how it should be. I'm not teaching anything new. This is Unix methodology. This is methodology from the '60s and '70s..."
Gloss: The speaker deflates the grievance by locating the idea in decades-old computing tradition.

> "They are creating some interesting tools that I actually might use or integrate, we'll see."
Gloss: A concession that Google's implementation may have practical value beyond the shared core idea.

> "And if anyone knows the researchers at Google, just tell them to come my way. I'd love to work with them."
Gloss: The closing pivots from territoriality to enthusiastic collaboration.

- **Themes**
- **Convergent discovery** – Google's paper is treated as confirmation that the insight is inevitable rather than stolen.
- **Minimalism over complexity** – A recurring preference for "just files" instead of elaborate agentic frameworks.
- **Unix philosophy as AI architecture** – The argument that 1960s/70s file-system thinking is the correct foundation for modern agent routing.
- **Empirical superiority** – Repeated claims that this methodology "crushes" alternatives in production workspace deployments.
- **Open collaboration** – Despite the initial framing of theft, the tone shifts to welcoming partnership with the very researchers he accuses of copying him.

## Tensions & objections

**The null case — strongest objection to the thesis:**

The speaker's claim of empirical superiority is entirely anecdotal. "Hundreds of workspaces" and "crushing any agentic systems" sounds impressive, but the transcript offers zero benchmarks, no success metrics, no description of what tasks were tested, and no controlled comparison. This is the rhetorical structure of a testimonial, not evidence.

The approach almost certainly excels at document-retrieval and knowledge-management tasks — precisely the domains where folder hierarchies and Markdown files are a natural fit. But the speaker never acknowledges a single scenario where this architecture fails, struggles, or is suboptimal. Complex multi-step reasoning, dynamic tool discovery, error recovery, stateful multi-turn conversations, and tasks requiring real-time data may genuinely require more sophisticated architectures than static file trees.

**The "parallel discovery" argument cuts both ways:** If this is truly just rediscovering Unix methodology that "we should have been doing years ago," then the speaker's claim to priority is weakened — it's like claiming to have invented the wheel because you built a cart before someone else did. The fact that two Google researchers arrived at it independently could mean it's an obvious first attempt for anyone familiar with file systems, not a deep or hard-won insight.

**Selection bias is unaddressed:** The speaker's "hundreds of workspaces" may be predominantly similar tasks where file-based routing naturally works well. Without knowing the distribution of task types, the claim of universal superiority is unfalsifiable — we only hear about the wins.

**What the speaker cannot say without undermining his own framing:** He cannot simultaneously claim (a) this is ancient, obvious Unix methodology that nobody was doing, and (b) he has uniquely validated it across hundreds of production deployments. If it's so obvious, why did nobody else build hundreds of workspaces with it? If it required his unique insight to see the application, then it's not "just Unix methodology" — it's a specific architectural bet that deserves rigorous evaluation, not appeals to tradition.

## What's worth learning — and what we could do with it

1. **Prototype a file-system-first agent router.** Build a minimal agentic system where the "state" is a folder tree of Markdown files with YAML front matter, and the LLM routes itself by reading/writing files rather than calling a framework orchestrator. Test it on a real task (e.g., research synthesis, multi-step content generation) to see if the overhead is actually lower than LangChain/LlamaIndex equivalents.

2. **Run a structured bake-off.** Pick 3-5 representative tasks (document retrieval, multi-step reasoning, tool chaining, error recovery). Implement each with both the folder/YAML/Markdown approach and a mainstream agentic framework. Measure tokens consumed, latency, failure rate, and human debugging time. Turn the speaker's anecdote into reproducible data.

3. **Audit your own "agentic" stack for Unix violations.** List every abstraction layer between the LLM and the filesystem. Ask: does this layer enable a capability that raw files cannot, or is it ceremonial complexity? If the latter, strip it and measure the delta.

4. **Publish a negative-result log.** The speaker's claim is unfalsifiable because he only reports successes. Counter-model this by documenting specific tasks where the file-tree approach breaks down (e.g., real-time API polling, stateful multi-user sessions, dynamic tool discovery) and what architecture was required instead.

5. **Contact the Google researchers.** The speaker invited collaboration; the substrate could attempt to reach out to the authors of the cited paper to compare implementations, share benchmark protocols, or co-author a reconciliation note. This turns a territorial claim into an open research thread.

6. **Extract the "LLM wiki-style pattern" as a design spec.** Formalize the speaker's implicit schema: folder hierarchy = namespace, Markdown body = prompt/context, YAML front matter = metadata/routing tags. Write a one-page spec and test whether it generalizes across different LLM providers without framework lock-in.