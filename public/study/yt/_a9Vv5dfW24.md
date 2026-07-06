# VibeThinker 3B - Taking on Giant Models


## Thesis

Sam Witteveen examines VibeThinker 3B, a 3-billion-parameter reasoning model from Weibo AI Lab (Singapore) that claims to outperform much larger models—including Gemini 3 Pro, Claude Opus 4.5, GLM 5, and DeepSeek V 3.2—on hard math and coding tasks. The model is not trained from scratch; it starts from the older Qwen 2.5 Coder 3B and applies a novel post-training recipe.

The core thesis is that not all intelligence requires the same parameter count. Tasks involving verifiable reasoning (math, code) are fundamentally about search, constraint satisfaction, and error correction—not memorization. In these structured domains, a small model equipped with the right training can "figure things out" rather than "know things," competing with models hundreds of times its size.

## How it builds

Witteveen structures his analysis in three parts: (1) the model's claims and benchmark context, (2) the training methodology, and (3) a live demo.

**Benchmark framing:** He notes the model competes with models ~300× larger. On math benchmarks (AIME, AMIE 26), it is "on par if not beating" Claude Opus 4.5, Kimi 2.5, GLM 5, and Gemini 3 Pro. On coding benchmarks, it's "miles ahead" of other small models like Gemma 4 12B, Olmo, and Qwen 3/3.5. However, on knowledge benchmarks (GP-A diamond), it lags behind larger open reasoning models and proprietary models—confirming the paper's own distinction between "verifiable reasoning" and "broad knowledge."

**Training methodology:** The paper proposes a "spectrum to signal" principle: generate synthetic data with diverse solution strategies (the spectrum), then use reinforcement learning to amplify correct paths (the signal). The pipeline uses two-stage curriculum SFT (broad coverage → hard long problems only, discarding traces under 5,000 tokens), multi-domain RL (MGPO — Max Entropy Guided Policy Optimization), "long to short math RL" (accuracy-first, then reward shorter correct answers), and CLR (claim-level reliability, a test-time compute trick that generates many answers and selects the best).

**Demo:** Witteveen runs the model locally on a Dell MaxPro with RTX Pro 6000. It produces extremely long chains of thought on math, coding, and logic tasks. On general tasks (SVG drawing, long essays, long-context QA), it either over-thinks or produces poor results, revealing its narrow domain.

## Key passages

**On the core idea — verifiable reasoning vs. knowledge:**
"They're saying tasks that use verifiable reasoning, so this is things like math, things like code, are mostly a task around search and sort of constraint satisfaction, along with error correction."

**On the spectrum-to-signal principle:**
"They first built a bunch of sort of synthetic data with like diverse sets of solution strategies, that's what they call the spectrum. And then they use reinforcement learning to amplify the correct ones in there, which is the signal."

**On the two-stage SFT curriculum:**
"Stage one is like the broad coverage across math, code, STEM topics, chat. And stage two is retraining only on harder long problems. And they actually mention here that they throw out reasoning traces that are under 5,000 tokens."

**On long-to-short RL:**
"First they optimize for accuracy, and then later on they're basically giving rewards to shorter correct answers and penalizing the long ones."

**On test-time compute (CLR):**
"So often they're just under the proprietary models or the other big open weight models, by using this test time compute technique, that's what gets them over."

**On the demo — over-thinking on general tasks:**
"So the big thing then is to sort of see, okay, for things that it doesn't need that, does it still actually use really long chains of thought? And you can see even for the simple logic test, it's using a much larger amount of tokens than with the previous GLM 5.2 model."

## Themes

1. **Reasoning ≠ memorization.** The paper draws a sharp line between "verifiable reasoning" (search + constraint satisfaction) and "broad knowledge" (fact storage). Small models can compete on the former without the latter.

2. **Post-training > pre-training for narrow domains.** VibeThinker starts from an old base model (Qwen 2.5 Coder 3B) and achieves breakthrough performance through its training recipe alone—suggesting the post-training signal matters more than base model freshness for reasoning tasks.

3. **Diversity as a training signal.** The "spectrum to signal" approach treats diverse solution strategies as the raw material, with RL as the filter. This contrasts with the common approach of converging on one "best" reasoning path.

4. **Test-time compute as a force multiplier.** CLR (claim-level reliability) and long-to-short RL show that inference-time strategies—generating many answers, selecting the best, rewarding brevity—can close the gap with larger models without additional training.

5. **The narrowness trade-off.** The demo reveals the model's limits: it over-thinks everything, produces poor SVGs, and struggles with general tasks. Its strength is also its weakness—it's specialized, not general.

## Tensions & objections

**The strongest objection: test-time compute inflates the comparison.**

Witteveen himself flags this: "in some ways you could say that's not really fair because the other big models are perhaps not doing that." VibeThinker's benchmark wins rely partly on CLR—generating many answers at test time and selecting the best. If the comparison models (Gemini Pro, Claude Opus) were also evaluated with test-time compute, the gap would narrow significantly. The paper essentially compares VibeThinker's best-case inference (many samples + selection) against other models' single-pass outputs.

**Secondary objection: the base model is old.** Starting from Qwen 2.5 Coder 3B—a model released well before Qwen 3/3.5—means the model carries the limitations of an older architecture. Witteveen himself suggests the ideas might work "much better for a 9B model" or on "something like the Gemma 12B model." The 3B size is a proof of concept, not a production target.

**Tertiary objection: the demo reveals brittleness.** The model's tendency to over-think even simple tasks (producing thousands of thinking tokens for a basic logic question) and its failure on out-of-domain tasks (SVG drawing, long essays) suggests the training has created a model that is narrow and brittle rather than broadly capable. The benchmark wins are real but narrowly scoped.

## What's worth learning

1. **Separate reasoning from knowledge in your training data strategy.** If you're building a model for verifiable-reasoning tasks (math, code, logic), prioritize synthetic diversity + RL over massive pre-training. The spectrum-to-signal principle is a concrete recipe: generate many solution strategies, use RL to amplify correct ones.

2. **Discard short reasoning traces.** The two-stage SFT that throws out traces under 5,000 tokens is a simple but powerful signal: force the model to think deeply, not pattern-match shallowly. This is a data-filtering heuristic worth trying.

3. **Reward brevity after accuracy.** The "long to short math RL" phase—optimize for correctness first, then reward shorter correct answers—addresses the over-thinking problem. Apply this to any domain where your model generates verbose but correct outputs.

4. **Use test-time compute strategically.** CLR (generate many answers, select the best) is a cheap way to boost performance without retraining. For any narrow domain where you can verify answers, this is a free performance multiplier.

5. **Specialize before generalizing.** VibeThinker's demo shows that a narrow model can shine in its domain but fail outside it. The takeaway: build specialized reasoning models for specific domains (e.g., a design-focused agent model) rather than trying to make one small model do everything.