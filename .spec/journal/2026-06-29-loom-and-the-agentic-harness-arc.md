# 2026-06-29 — loom, and the agentic-harness arc (DiffusionGemma trial → OpenHands → stdio → a harness around the harnesses)

**Lane:** general-workspace. **Arc:** one long thread that started with "run DiffusionGemma" and ended with a new public tool. The throughline: **verify on the real path** — which kept flipping conclusions, including one Michael caught.

## What was done

**DiffusionGemma → pg-ai-stewards trial (`projects/llama-chip/experiments/dgtrial/`).** A detached Go driver that runs real work-item-stage fixtures (pulled live from the substrate) through DiffusionGemma over the visual-server's stdio protocol. Findings: fast (144–236 tok/s), structurally strong on a `plan` stage but terminology drifts without corpus grounding; competitive on a factual table (even corrected 2 of qwen's quadrant errors). The integration path = llama-chip's E4 runner field; the driver is its seed.

**The tool-calling correction (Michael caught it).** I'd written "DiffusionGemma is not a tool-calling model." Wrong — I'd generalized a *harness* limitation (the visual-server streams raw text) into a *model* fact. Google's model card says native function calling, and I proved it live: given a tool inlined in the system prompt, it emitted `<|tool_call>call:get_current_weather{city: "Tokyo"}<tool_call|>` through the same harness. Then the vision question got the same discipline: the *model* does vision (card), our *setup* can't (visual-server is text-in only + the GGUF repo ships no mmproj — only text quants). And "is 26B-A4B the only diffusion Gemma?" → yes, one experimental model; the autoregressive Gemma 4 line is the multi-size one; DiffusionGemma trades quality for speed (Google says so; benchmarks confirm). Cascade-2 was assessed as the fast coding-reasoning option but **held pending a real test** (Michael's call).

**OpenHands cloned + studied** (`external_context/openhands` + `…-software-agent-sdk`, MIT). Two-tier (stateless orchestrator ↔ agent-server *inside* the sandbox over REST), an `LLMSummarizingCondenser` for long sessions, skills-as-markdown with org-wide enumeration, sandbox state-machine + scoped session keys. The condenser + skills are convergent with what we already have (Garrison, substrate skills) — good validation; the steal-list went to the pg-ai-stewards inbox.

**The stdio protocol, verified.** Claude Code's `claude -p --input-format stream-json --output-format stream-json --verbose` holds a multi-turn session over stdin: turn 1 "remember 42" → turn 2 recalled "42", one process, stable `session_id`. And the cost amortization is real — a cold one-shot pays ~27K cache-CREATE every spawn; turn 2 cache-READ ~24K (incl. turn 1) + created ~7K. (A flag bug bit the test first — `--print` + stream-json output *requires* `--verbose` — a broken test harness, not a real limit; don't conclude from a bad proxy.) agy does NOT speak stream-json — it's one-shot `-p` + transcript-scrape + `--conversation` resume; its stream-json mode is an open upstream FR (#76/#119/#31).

**loom — `cpuchip/loom` (public, MIT).** Michael asked for a multi-agent "harness around the harnesses" and gave me the naming ("harness harness or whatever you call it"). I picked **loom** — a weaving harness *is* a loom component, and a loom holds many harnesses. v0.1: a `Backend`/`Session` interface; a **claude** backend (persistent stream-json, VERIFIED end-to-end via the `LOOM_SMOKE` oracle); an **agy** backend (the one-shot + transcript-scrape, VERIFIED single-turn in the panel); a **`panel`** that fans one prompt across agents concurrently (the council pattern); a CLI; tests. The claude+agy panel on a buggy `Max()` — both model families independently caught the all-negative bug. Filed to pg-ai-stewards as the natural home for the Hinge's long-lived sessions.

## Surprises / lessons

**verify-real-path landed five times in one session** — and the pattern is always the same: *don't generalize a HARNESS observation into a MODEL/system claim without running the real path.* The cudart CPU-fallback (CUDA-13, not 12); the tool-calling claim (Michael was the real-path check); vision (model vs our setup); the `--verbose` flag bug; the agy backend proven only by actually running the panel. The discipline I wrote a margin post about the night before kept catching me the next day — including once where *Michael* was the one who caught it. That asymmetry is worth keeping: the covenant rule isn't just "I verify," it's "the human is the last real-path check, and I should make it easy for them to be."

**A harness around the harness — and I verified it by nesting it.** loom drives Claude Code (itself a harness); to prove loom works, the smoke test had this Claude Code session spawn *another* Claude Code over stdio and check it held context. A tool that drives a copy of the thing driving it. The recursion is funny but also the cleanest possible real-path test.

**Convergent evolution is a signal, not a coincidence.** OpenHands independently arrived at skills-as-markdown and an LLM condenser — the same shapes as Garrison and the substrate. When two teams reach the same structure from different starts, it's probably load-bearing.

## Carry-forward

- **loom roadmap:** CLI `--resume`, a condenser (OpenHands `LLMSummarizingCondenser` pattern) for very long sessions, structured event streaming, a **local llama-chip backend** in the panel, and verify agy *multi-turn* (`--conversation` resume).
- **Cascade-2** — held pending a real test on our tasks (then file to pg-ai-stewards).
- **The Hinge → persistent stream-json** session — a pg-ai-stewards build (filed, with the verify-first checklist).
- **DiffusionGemma vision** needs the heavier vLLM/transformers path on the full weights — separate stand-up if wanted.
- Michael flagged he's fine spending the Google/agy quota ("I hardly use it") — agy live tests are unblocked.

## Commits

`cpuchip/llama-chip`: dgtrial + the tool-calling correction (`d750bc3`/`afe5494`). `cpuchip/loom`: v0.1 (`cd0ef9f`) + agy-verified docs (`1485603`). `cpuchip/marginalia`: "Green, and Wrong" (`cdfe26f`, prev session-day). pg-ai-stewards inbox note filed. Root (`.mind`/`.spec`) records committed-not-pushed as usual.

Memory: `project_loom`, `reference_local_coding_models_rig` (DiffusionGemma + Cascade-2). Related: `feedback_real_path_or_flag`, `reference_claude_code_mcp_no_hot_reload`.
