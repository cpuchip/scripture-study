---
workstream: WS5
status: deferred
created: 2026-05-08
last_updated: 2026-05-08
---

# Per-Model Prompt Tuning for Substrate Agents

> **Carry-forward (2026-05-11 consolidation):** Still deferred. Trigger
> for un-deferring: Batch H ("Use validation" — first real Phase A–F
> study runs) produces 2-3 cross-topic study runs. Then 3h.1 cross-topic
> validation has real data instead of being synthetic. Tracked at
> `projects/pg-ai-stewards/.spec/proposals/substrate-deferred-items.md`
> § C.2.

## Binding Problem

Different models have different default voices and tool-use tendencies. The same study agent prompt that produces well-voiced output on Opus 4.7 produces output on kimi-k2.6 with closing refrains, triadic flourishes, Latinate register, and confabulated revision notes. On qwen-3.6 it produces tool-name confusion, broken `(#)` link conventions, heavy mid-argument tables, and verbosity. The substrate's `(family, model_match)` agent variant table makes per-model prompt overrides structurally possible (Phase 3a shipped this for `watchman-consolidator`); the question is when and how to operationalize it across the full agent corpus.

## Status: deferred

A working prototype landed during the 2026-05-08 overnight session as Phase 3c.3.4 + 3c.3.4.1:

- `.stewards/kimi-k2.6/study.agent.md` — kimi-tuned study variant
- `.stewards/qwen-3.6/study.agent.md` — qwen-tuned study variant
- Five-way comparison memo at `study/.scratch/two-triplets-comparison-2026-05-08/comparison.md`

Both variants demonstrably outperformed the base prompt on the FtC/WtL binding question (run #4 cleared 5/6 measurable kimi signatures with active in-draft quote-correction; run #5 cleared all 12 targeted qwen signatures, 54% shorter / 16% fewer tokens / 61% faster than the qwen-base baseline).

**The prototype proves the mechanism works.** The full effort — generalizing the validation across study types, extending to other agent families (lesson, talk, journal, research), authoring variants for additional models (Sonnet, Gemini, GLM, etc.) — is **deferred to post-3e+ work** because:

1. The current prototype is **validated only on a single binding question** (FtC/WtL). The voice signatures we targeted may be partially specific to that question's structure (multi-source meta-study about parallel triplets). Cross-topic validation is the prerequisite to claiming generalized improvement.
2. Bigger surface gains exist upstream of voice. **3e (MCP server + client, with the former 3c.4 absorbed as 3e.2)** unlocks both the substrate as an external tool surface for IDE agents AND real scripture-quote verification at substrate-internal agent runtime. **3f (web UI)** unlocks the reading surface. Both move the substrate from "internally functional" to "externally useful" — voice tuning is downstream of them.
3. The prompt-tuning loop is **mechanical once instrumented**. With proper benchmarking (see References below), each new model can be onboarded against a fixed test set in a few hours.

## Where this sits in phase order

After 3e (MCP server + client, which absorbs the former 3c.4 gospel-engine HTTP work as sub-stage 3e.2) and 3f (web UI). Tentatively phased as:

- **3h.1** — Cross-topic validation suite. Run base + tuned variants of `.stewards/kimi-k2.6/study.agent.md` and `.stewards/qwen-3.6/study.agent.md` on three structurally distinct binding questions (focused exegesis, character study, modern-prophet talk analysis). Confirm tuned variants improve, regress, or are neutral on each. Adjust prompts based on findings.
- **3h.2** — Extend variant authoring to other study-adjacent agents (`lesson`, `talk`, `journal`, `research-gospel`). Each variant requires a baseline run + signature identification + tuned authoring + validation.
- **3h.3** — Onboard additional models. Each new model runs the cross-topic suite as both base and tuned, surfacing model-specific signatures. Cost-quality data informs default-provider selection per pipeline stage.
- **3h.4** — Migrate the test suite into a proper `study-bench` CLI (mirroring `classify-bench`'s shape — see References) so model evaluation is repeatable and produces structured comparison artifacts.

## Success criteria

1. A reusable evaluation rubric for study-agent output (extending the six kimi-shared and six qwen-specific signatures) that scores any given study against a model-neutral voice target.
2. Tuned variants for at least three models (kimi, qwen, one Anthropic) with cross-topic validation on three structurally distinct binding questions each.
3. A `study-bench` CLI tool that runs a binding question through N models × M prompt variants and produces a side-by-side report — the same shape as `classify-bench`.
4. Documented per-model cost-quality data sufficient to make pipeline-stage provider/model selection a checkable decision rather than a default.

## References — prior rubric work

The 2026-03 / 2026-04 brain-classification work pioneered the approach we should re-use here. **Read these first** before authoring the cross-topic study evaluation suite:

- **[classify-bench](classify-bench.md)** — proposes a CLI tool that classifies a fixed test dataset through 6 models (LM Studio + Copilot SDK), captures category, confidence, title, tags, latency. Human judgment against 5 criteria: category correctness, confidence calibration, title quality, actionability, latency. **The bench-CLI shape is what `study-bench` should mirror.**
- **[classifier-qwen-fix](classifier-qwen-fix.md)** — diagnoses qwen3.5-9b returning empty responses with `StructuredOutput: true` (thinking-model + grammar-sampling conflict at LM Studio). Fix: disable structured output, prompt-suppress thinking with `/no_think`. **Relevant practice: per-model adapter config, not just per-model prompt content.**
- **[archive/lm-studio-model-experiments/main.md](archive/lm-studio-model-experiments/main.md)** — comprehensive 5-phase test harness for 5 inference models + 2 embedding models. Explicit 0-5 scoring rubric on Accuracy, Depth, Citations, Hallucination (inverse), Usefulness. `results.tsv` structured logging. **The dimensional scoring rubric is directly portable to study-agent evaluation.**
- **[journal/2026-03-28--pass1-all-models-run.yaml](../journal/2026-03-28--pass1-all-models-run.yaml)** — Phase 1 outcome from the experiments above. Surfaced "thinking mode is a silent killer" — qwen and GLM allocate output tokens to invisible `<think>` blocks. **Same trap could appear in study work; the qwen-tuned variant should suppress thinking explicitly.**

Three patterns from this prior work that the prompt-tuning effort should preserve:

1. **Structured-data capture + human judgment.** The rubrics produced TSV/JSON-shaped data (token counts, latencies, dimensional scores), but Michael made the final quality call. The tools are evaluation aids, not decision-makers.
2. **Real test data, not synthetic.** classify-bench uses live brain entries; study-bench should use real binding questions Michael has actually asked, not contrived comparisons.
3. **Per-model adapter quirks matter as much as prompt content.** classifier-qwen-fix surfaces a structural-output bug that no prompt rewrite would catch. The substrate's `agents.response_format` + `agents.temperature` + `agents.top_p` columns exist for exactly this; per-model variants should set them deliberately, not inherit defaults.

## Out of scope

- Authoring variants for the watchman-consolidator agent — already done in Phase 3a, no observed need to re-tune.
- Authoring variants for non-study agents under WS6 specifically (lesson, talk, etc.) without a binding-question test set — premature.
- Bench tooling for embeddings — separate concern; current substrate uses `nomic-embed-text:v1.5` exclusively.

## Decision log

- **2026-05-08:** Prototype shipped (kimi + qwen study variants, both validated on FtC/WtL). Effort deferred to post-3e+ pending cross-topic validation. Recorded here so the validated prototypes don't atrophy.
