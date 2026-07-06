# Make ANY Model Think Like Fable in Minutes

## Critique

The digest is largely faithful — all five quotes verify against the transcript, and the structural summary of the pipeline is accurate. However, it flattens three things:

1. **The author's own hard ceiling.** The transcript contains an explicit concession — "a lot of these, again, we can't change cuz they come from the model weights itself" and "you still won't get Fable 5 performance" — that the digest buries in the final gloss rather than surfacing as a structural limitation of the entire method.

2. **The iterative, exploratory texture.** The author walks through the process as a series of small bets (check the blast radius → strip one file first to validate format → then scale). The digest presents it as a clean pipeline, which overstates how recipe-like it is.

3. **The comparison confound.** The author never addresses whether the Fable 5 and Opus 4.8 sessions in his logs were working on comparable problems. The "distance" metric he celebrates may reflect task-difficulty confounds, not pure model-behavior differences. The digest reproduces his framing without noting this gap.

No claims are unfaithful to what was said. The corrections below are additive, not reparative.

---

**The core thesis / claim**

Because the intelligence of Fable 5 resided in the model weights themselves, the author cannot restore Fable. However, he claims that by mining the JSONL session logs of past Fable 5 conversations (or open-source logs shared by others), one can extract measurable behavioral patterns—tool cadence, planning depth, action sequences—and distill them into a playbook. Injecting that playbook into new sessions via hooks or Claude MD can make existing models (Claude Opus, Codex, or open-source models) behave more like Fable without changing their underlying weights — though the author concedes this yields improved execution, not Fable-level performance.

**How it builds**

The argument moves from loss to workaround. It opens with the emotional hook of Fable withdrawal, then pivots to the raw material: JSONL log files that store every tool call, model response, and planning step. The author identifies the problem (bloat), proposes a stripping script, then builds a pipeline iteratively: first ask how many JSONL files exist to gauge the "blast radius," then write a stripping script and validate it on a single file before scaling, then isolate Fable transcripts → synthesize behavioral metrics as "real measured numbers" → run a side-by-side comparison against a baseline model like Opus 4.8 → quantify the "distance" in rhythm and tool use → distill the delta into a playbook. Finally, he shows how to operationalize the playbook through session-start hooks, the Claude Code Guide agent, or Claude MD skills, and offers his own pre-built playbook for those who lack data.

**Key passages**

> "the majority of your conversations, whether they're Codex or Claude code, live in what are called JSONL files on your computer... And within this metadata is a series of gold that you can mine."

*Gloss: The entire method rests on treating local session logs as a behavioral dataset rather than ephemeral chat history.*

> "Give me the behavioral patterns as real measured numbers, not just impressions... something that is tangible versus just an intangible objective look at the quality of the conversation."

*Gloss: The author insists on quantifiable metrics to avoid vague impressions when comparing model behaviors.*

> "how disciplined Fable seemed to be around using the right tools at the right time... You can learn a lot from its rhythms... the way it read and edited files, everything seemed to be a little bit more elegant and a little bit more refined and precise."

*Gloss: The target behavior is not just accuracy but a specific operational rhythm—tool selection, sequencing, and timing.*

> "Show me the distance between their rhythm, the tool call cadence, the action sequences, and the ratios like reads before edits and tests after edits."

*Gloss: The comparative analysis focuses on structural execution patterns rather than output quality alone.*

> "a lot of these, again, we can't change cuz they come from the model weights itself. But if you can implore or elicit Opus to think that much longer or plan a little bit longer to be a lot more thoughtful, you still won't get Fable 5 performance, but you can get a much stronger Opus execution."

*Gloss: The author's own ceiling — the playbook can elicit better execution from existing weights but cannot replicate the reasoning those weights enable.*

> "we can't clone the power of Fable 5, but you can do a few things to at least improve the models that you currently have in the meantime while we wait for all this play out."

*Gloss: The author explicitly bounds the claim: this is a performance boost, not a replication of Fable's raw intelligence.*

**Themes**

- **Behavioral mimicry without weight access:** The recurring tension between what the model *is* (weights) and what it can be *elicited* to do (behavior).
- **Log mining as prompt engineering:** Treating JSONL transcripts as training data for meta-prompts and playbooks.
- **Rhythm and discipline:** A preoccupation with the *order* and *cadence* of tool use—reads before edits, tests after edits, bash chaining—as the signature of Fable's superiority.
- **Context injection infrastructure:** The practical focus on hooks, Claude MD, and skills as delivery mechanisms for distilled behavior.
- **Community data surrogacy:** Using open-source Fable session logs when personal history is insufficient.

## Tensions & objections

**The strongest null case: rhythm is a symptom, not a cause.**

The entire method rests on the assumption that Fable's behavioral patterns (tool cadence, reads-before-edits, planning depth) are *transferable levers* — that if you make another model mimic those patterns, it will produce better outcomes. But the author himself admits the gap: "you still won't get Fable 5 performance" and "a lot of these, again, we can't change cuz they come from the model weights itself."

The objection is straightforward: Fable's disciplined rhythm is likely a *consequence* of its superior reasoning, not the *source* of it. A chess grandmaster thinks slowly and checks lines carefully — but telling a novice to "think longer and check more lines" does not produce grandmaster play, because the novice lacks the pattern recognition that makes those checks meaningful. Similarly, telling Opus to "think longer before acting" may produce more tokens of deliberation, but if the underlying reasoning capacity hasn't changed, those extra tokens are just more elaborate wandering.

**Specific vulnerabilities:**

1. **Correlation without controlled comparison.** The author compares Fable 5 and Opus 4.8 sessions from his own logs, but never addresses whether those sessions involved comparable tasks. If Fable was used on harder or more structured problems, the behavioral "distance" may reflect task differences, not model differences. The method treats observational data as if it were experimental.

2. **The playbook is just instructions.** At the end of the day, the distilled playbook is a set of natural-language instructions ("think longer," "read before editing," "test after edits"). But if a model *could* follow those instructions optimally, it would already do so — its failure to do so is a symptom of its weights, not a missing memo. Injecting the playbook is essentially re-prompting, and the author offers no evidence that the specific playbook he distills outperforms a generic "think step by step and verify your work" prompt.

3. **Survivorship in the logs.** The JSONL files represent sessions the author chose to have with Fable — likely the ones that went well or were interesting. The behavioral patterns extracted may reflect cherry-picked success cases rather than Fable's average behavior.

4. **The ceiling is the floor.** The author's own framing — "the second best thing until we get our super intelligence back" — implicitly concedes that this is a stopgap, not a solution. The strongest version of the objection is that the method is *correct but trivial*: of course you can nudge a model's behavior with better prompts; the interesting question is whether the nudge is large enough to matter, and the author provides no controlled measurement of outcome improvement, only behavioral description.

## What's worth learning — and what we could do with it

1. **Mine your own JSONL logs for behavioral baselines.** Write a script to strip and analyze your local Claude/Codex session logs to quantify your current model's tool cadence (reads per edit, tests per task, planning tokens per action). Establish a personal baseline before trying to change anything.

2. **Build a "session start" hook with bounded claims.** Draft a concise Claude MD skill or Codex agent prompt that injects 3-4 specific operational rules (e.g., "always read the full file before editing," "run tests immediately after edits," "write a 3-step plan before first tool call"). Treat it as a prompt-engineering experiment, not a model replacement.

3. **Run a controlled A/B on a repeated task.** Pick a standardized coding task you can give to your current model with and without the playbook. Measure outcome quality (passing tests, fewer retries) and behavioral metrics (tool call count, planning tokens). The author's method lacks this control; adding it turns anecdote into evidence.

4. **Curate a public "behavioral log" dataset.** If you have (or can find) high-quality Fable or other elite-model session logs, publish the stripped JSONL behavioral sequences as an open dataset. The community can then test whether the patterns generalize across different base models and tasks.

5. **Design the playbook as a checklist, not a personality.** Distill the target behavior into verifiable execution steps rather than vague style instructions. For example: "Before any edit tool call, emit a `read` or `view` call for the target file; after any edit, emit a `bash` test call within 3 turns." This makes the prompt auditable and the model's compliance measurable.

6. **Accept the ceiling and measure the floor.** Since the author concedes you cannot replicate Fable-level reasoning, set your success criterion as "better than baseline Opus/Codex on my specific tasks" rather than "Fable-like." Document the delta honestly; even a 10% reduction in retry loops is a real win.