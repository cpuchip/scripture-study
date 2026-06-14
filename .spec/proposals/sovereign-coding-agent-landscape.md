# Garrison — Landscape & Design Inputs

**Date:** 2026-06-14 (overnight research, unattended)
**Status:** Research note — companion to [`sovereign-coding-agent.md`](./sovereign-coding-agent.md). Feeds the council's open questions; **decides nothing** (gather-and-evaluate only, per the unsupervised-scope rule).
**Sourcing:** Web-verified June 2026, not from memory. Stars/benchmarks are single-source summaries and move fast — treat as directional, verify any number before it reaches the book or a public claim. I did not run any of these tools; this is documentary, not empirical.

---

## Why this exists

Michael named three of these in passing: he doesn't love how **opencode** is put together, and hadn't tried **pi** or **hermes**. This note digs into all of them plus the rest of the field, aimed at Garrison's open questions — how each handles *local* models, how lean it is, how it survives weak-model tool-calling, and where the gap is that Garrison's governance niche would fill.

## The field (June 2026)

| Tool | License | Shape | Local models | Leanness | Note |
|------|---------|-------|--------------|----------|------|
| **pi** (M. Zechner) | MIT | terminal agent loop | yes (20+ providers incl. self-hosted) | **extreme** — 4 tools, ~300-word prompt | self-extends via TS/skills; no bundled MCP/plan-mode |
| **aider** | Apache-2.0 | git-native pair-programmer | yes (Ollama / OpenAI-compatible) | lean, focused | edit/diff/commit loop; the Aider benchmark is its own |
| **goose** (Block) | Apache-2.0 | framework + MCP extension bus | yes (Ollama, LM Studio, vLLM) | medium | model-routing by complexity; autonomous mode runs code **without approval** |
| **opencode** | open | multi-provider TS framework | yes (opencode.json custom provider) | heavy | huge adoption; config sprawl; ~78% slower than Claude Code on one timed task |
| **Codex CLI** (OpenAI) | open | official CLI | yes (`--oss`: Ollama/LM Studio/MLX) | medium | defaults to gpt-oss:20b (16GB) / 120b (80GB) |
| **Hermes Agent** (Nous) | open, self-hosted | persistent multi-platform agent | model-agnostic | heavy surface | 20+ chat platforms, 6 backends; persistent memory + cost-routing |

*Sources: [Morph open-source ranking](https://www.morphllm.com/ai-coding-assistant-open-source), [pi (dev.to)](https://dev.to/arshtechpro/pi-the-open-source-ai-coding-agent-you-probably-havent-tried-yet-2h0h) + [pi repo](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent), [goose review](https://www.openaitoolshub.org/en/blog/goose-ai-agent-block-review) + [goose providers](https://goose-docs.ai/docs/getting-started/providers/), [opencode review](https://ivern.ai/blog/opencode-review-open-source-ai-coding-agent-2026), [Codex+Ollama](https://ollama.com/blog/codex), [Hermes docs](https://hermes-agent.nousresearch.com/docs/).*

## The two Michael named, plus the one he disliked

**Pi — the lean exemplar, and the closest thing to Garrison's core already shipping.** Four built-in tools (read, write, edit, bash), a ~300-word system prompt, MIT, provider-agnostic, no SaaS backend. It deliberately omits plan-mode, to-do lists, and MCP, expecting you to add them as TypeScript extensions or *ask the agent to build them*. The lesson for Garrison: a four-tool, tiny-prompt agent loop is not a toy — it is a shipping product with users. That directly answers open question #2 ("how lean is lean"): **the irreducible core is roughly read/write/edit/bash + a verify step + the watch.** What Garrison adds on top is exactly what Pi leaves out on purpose: governance. Pi extends toward *features*; Garrison extends toward *the oracle, the critic, and the presiding ledger*.

**Hermes — the contrast, and a warning.** Nous Research's self-hosted agent has two ideas worth borrowing — persistent memory across sessions and cost-routing across models — wrapped in a surface Garrison should refuse: 20+ chat platforms (Telegram, Discord, WhatsApp, Signal…) and six execution backends. Its own documentation says it is "strongest when the workflow is narrow enough to inspect and repeat." That sentence is an unintentional argument *for* Garrison's gated, governed scope: the tool itself is best exactly where it is watched, and worst when it sprawls.

**opencode — why the instinct against it is right.** It is enormously popular and genuinely capable, but it is a multi-provider TypeScript framework whose cost is configuration sprawl and a heavy surface (one timed comparison had it ~78% slower than Claude Code on the same task; the broader 2026 ecosystem around it has been unstable, with sibling projects archived or changing governance mid-year). It is the "everything for everyone" maximalism Garrison's spec already rejects as architecture (a). The dislike is well-founded; it is the anti-Pi.

## The architectural cousin: goose

Block's **goose** is the closest in shape to Garrison's chosen (b)/(c): a framework that drives any model, with capabilities attached as **MCP extensions**, and the ability to route tasks between models by complexity. Two takeaways:

- **Borrow:** the MCP-as-extension-bus design and model-routing. Garrison's substrate already speaks MCP, so the engine (council / verify / compact / work-item) plugs into a goose-shaped loop cleanly. This is concrete support for open question #6 (plugin/MCP).
- **Correct:** goose's autonomous mode "executes code without explicit approval." That is precisely the ungated autonomy the presiding covenant and Garrison's gated-autonomy/oracle gates exist to prevent. Garrison is, in one line, **goose's MCP-framework shape plus the governance goose omits.**

## The model question (open Q #4) — the most actionable finding

The known failure mode of local models in agent loops is **tool-calling reliability**, and it is worst on general-purpose models — which is exactly what qwen3.6-27B (Michael's stated floor) is. The field's own selection criteria for agentic local models are "function calling, long context, structured output, reliable instruction following, and recovery when the first plan fails" ([Morph Ollama ranking](https://www.morphllm.com/best-ollama-models)).

The standout: **Devstral Small 2** (Mistral + All Hands AI, Apache-2.0, 24B, finetuned from Mistral-Small-3.1, 128k context, ~14GB VRAM at Q4 — fits an RTX 4090/3090 or a 32GB Mac). It is *purpose-built* for agentic software engineering — "excels at using tools to explore codebases, editing multiple files, powering SWE agents" — and scores ~58% on SWE-bench Verified (the full Devstral 2 reaches ~72.2%). ([Mistral](https://mistral.ai/news/devstral/), [Ollama](https://ollama.com/library/devstral-small-2), [HF](https://huggingface.co/mistralai/Devstral-Small-2505)).

**Design input:** Garrison should not assume one model for everything. The strong pattern is a **split** — a tool-tuned agentic model (Devstral-class) running the edit/verify *loop*, and a reasoner (qwen3.6-27B, GLM-5.1, or Kimi K2.6) reserved for the *planning* step. That maps directly onto the substrate's existing "one strong doer + a critic" lesson (D&C 88:122) and onto open question #4. gpt-oss:20b (16GB, via Codex's `--oss` path) is a viable secondary. Mirrors Spin's hard-won rule (non-thinking instruct models for the tight loop; reasoners where reasoning is wanted).

## What nobody else has — the Garrison gap, confirmed

The survey's clearest result: **every one of these is ungoverned or lightly governed.** They edit and run on model judgment plus, at most, a confirmation prompt. None has oracle-first verification as a *hard gate*, a council/critic pass at each stage, a presiding ledger that tracks a sub-agent chain, or covenant-bound dispatch. goose comes closest structurally and then explicitly *removes* the gate ("executes without approval"). The niche Garrison claimed in the spec is real and empty: the differentiator is governance, and the field has left it open.

## Inputs mapped to the open questions

- **#2 leanness** — Pi proves the floor: read/write/edit/bash + verify + the watch. Build out toward governance, not features.
- **#3 runtime** — every tool targets an **OpenAI-compatible `/v1` endpoint**, which LM Studio, Ollama, and vLLM all expose. Garrison should target that endpoint, not a specific runtime. Michael already runs LM Studio → primary; Ollama is then free.
- **#4 tool-calling** — Devstral-class tool-tuned model for the loop; reasoner for planning; structured output + retries + a forgiving parser. Do not trust a general model's raw tool-calls without the oracle gate.
- **#6 plugin / MCP** — goose validates MCP-as-extension-bus; the substrate already speaks MCP, so the engine attaches cleanly. The Claude Code plugin can be the luxury-mode client of the same MCP surface.
- **Borrow / reject, in one line:** borrow Pi's leanness, goose's MCP bus + model-routing, Hermes' persistent memory + cost-routing, aider's git-native verify-and-commit loop; reject opencode's config sprawl, goose's ungated autonomy, Hermes' 20-platform surface.

## Caveats (honesty pass)

- June-2026 web summaries; star counts and benchmark numbers are single-source and volatile. Verify before any public use.
- Documentary, not empirical — I ran none of these. The P1 dogfood should actually drive **pi + Devstral Small 2** as a baseline *before* Garrison writes a line, to feel the real floor.
- "pi" and "hermes" identified by search (Zechner's pi; Nous's Hermes Agent). If Michael meant something else by "pi agents," flag it and I'll re-run.

## Sources

- Open-source coding-agent rankings: [Morph](https://www.morphllm.com/ai-coding-assistant-open-source), [Nimbalyst local-first 14](https://nimbalyst.com/blog/best-local-first-ai-coding-tools-2026/)
- pi: [dev.to writeup](https://dev.to/arshtechpro/pi-the-open-source-ai-coding-agent-you-probably-havent-tried-yet-2h0h), [repo](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent), [Real Python ref](https://realpython.com/ref/ai-coding-tools/pi/)
- aider: [Morph](https://www.morphllm.com/ai-coding-assistant-open-source) (install + Ollama usage)
- goose: [review](https://www.openaitoolshub.org/en/blog/goose-ai-agent-block-review), [provider docs](https://goose-docs.ai/docs/getting-started/providers/)
- opencode: [ivern review](https://ivern.ai/blog/opencode-review-open-source-ai-coding-agent-2026), [dev.to 140k](https://dev.to/ji_ai/opencode-hit-140k-stars-why-terminal-agents-won-2026-aci)
- Codex CLI local: [Ollama blog](https://ollama.com/blog/codex), [OpenAI CLI reference](https://developers.openai.com/codex/cli/reference)
- Hermes Agent: [Nous docs](https://hermes-agent.nousresearch.com/docs/), [dev.to](https://dev.to/lynkr/hermes-lynkr-the-self-improving-agent-meets-the-universal-llm-proxy-3n11)
- Devstral: [Mistral](https://mistral.ai/news/devstral/), [Ollama](https://ollama.com/library/devstral-small-2), [Hugging Face](https://huggingface.co/mistralai/Devstral-Small-2505)
- Local model selection: [Morph Ollama ranking](https://www.morphllm.com/best-ollama-models), [HF open-source LLMs 2026](https://huggingface.co/blog/daya-shankar/open-source-llms)
