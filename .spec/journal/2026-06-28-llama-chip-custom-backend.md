# 2026-06-28 — llama-chip cuts the LM Studio runtime tether (custom backend: P0 → E1 → E2 → E4-design)

**Lane:** general-workspace. **Arc:** coding-model rig test → diagnose the tether → spec → P0 verify → E1 + E2 shipped → E4 designed. llama-chip was mid-session released from the pg-ai-stewards lane ("llama-chip is freed... you can work on b").

## What was done

**The setup (why).** Tested kdnuggets' "top 7 coding models 2026" on fermion (2×4090) through llama-chip. Three served fine — **Nemotron Cascade 2 30B-A3B** was the fast standout (~182 tok/s; the 3B-active MoE ~5× faster than dense 33B EXAONE). But **North Mini Code (`cohere2moe`) and DiffusionGemma (block-diffusion) crashed** — LM Studio's bundled `llama.cpp` (cuda12@2.22.0) doesn't have those architectures yet, and llama-chip can only use what LM Studio downloaded. That's the **tether**: LM Studio's compiled runtime lags upstream `ggml-org/llama.cpp` by weeks, so a brand-new arch crashes until LM Studio ships a newer build. llama-chip was born to escape a dying *wrapper* (FlexLLama); this finishes the job by escaping the *runtime cadence*.

**P0 — the inverse hypothesis, proven.** North Mini crashes on LM Studio 2.22.0 but **serves correct code @ ~137 tok/s on a fresh ggml-org `llama-server` (b9837)** — verified via an isolated throwaway llama-chip instance on a spare port/GPU, never touching the substrate's `:8090`. The custom-backend path was already ~90% built (`resolveBackend` accepted an explicit dir); P0 proved the premise.

**E1 — per-slot `backend` override.** `Slot.Backend` + a `backendFor()` helper so static config, `/api/ensure`, and profiles all honor it. Proven: one rig, two backends — stable models on LM Studio's vetted cuda12 + North Mini on a fresh ggml build, both healthy, both generating. This is the valuable one: test a bleeding-edge arch on one slot without restarting the whole rig.

**E2 — managed `pull-ggml` + `ggml@<tag>` resolution.** `llama-chip pull-ggml [bNNNN|latest]` self-fetches the binary + cudart zips from ggml-org releases into `~/.llama-chip/backends/ggml-<tag>/` (idempotent); `backend: "ggml@latest"` resolves it; `backends` lists managed builds. So nobody hand-downloads two zips.

**E4 — the `runner` field, designed (not built).** Explored `external_context/unsloth` (Michael downloaded it: "mimic what they're doing for managing things"). Unsloth Studio's backend is a Python peer to llama-chip, and it already solves block-diffusion serving cleanly. Folded its design into a concrete, build-ready E4 in the spec: route-by-GGUF-metadata (`general.architecture` prefix / `diffusion.canvas_length`) *before* arch resolution → a different binary behind the *same* `/v1`+`/health` (router untouched) → runner discovery via env-or-adjacent-to-`llama-server` → lazy `llama-server` resolution. Plus a "patterns to adopt" appendix (integrity-verified pulls, capability probe, discovery ladder, HF-as-registry).

## Surprises / lessons

**The cudart bug the happy path hid — `verify_real_path` earning its keep a third time.** `pull-ggml latest` "worked" only because b9837 was already present (idempotent skip), which bypassed the actual download. The first *genuine* download (an absent b9747) failed "no asset" — because the cudart zip name (`cudart-llama-bin-win-cuda-12.4-x64.zip`) *contains* the binary's `bin-win-cuda-12.4` fragment, so the matcher grabbed cudart as the binary and never found the real cudart. Build, vet, the serve test, and the idempotent path all passed it. Only forcing the real path — pulling a build that wasn't there — exposed it. Fixed by matching cudart first (the more specific name); re-verified with a real b9747 pull (both zips down, `llama-server.exe` + DLLs extracted). This is exactly the just-ratified covenant rule: verify on the REAL path, not a proxy the happy case skips.

**Unsloth Studio is AGPL, not "Apache core / AGPL UI."** The whole *backend* — the inference/serving/diffusion/installer code, i.e. everything worth mimicking — is AGPL-3.0; only the `unsloth/` training tree (the least useful part here) is Apache. So we reproduce the *designs* (not copyrightable) in our own Go and do not lift the files. Worth flagging loudly in the spec because the first assumption was wrong.

**Unsloth keeps NO model alias catalog.** A model is just an HF `repo_id` + a quant "variant" → blob-hash → `huggingface_hub` snapshot into the stock HF cache. That validates llama-chip's lean approach — if we ever add model-pulling, copy this shape, not a heavy catalog.

## Carry-forward

- **E4 runner field** — build when a diffusion `llama-server`-adjacent binary actually ships in a ggml-org (or our own) release, or wire Unsloth Studio's binary. Until then DiffusionGemma runs in Unsloth Studio.
- **Adopt-later patterns** (in the spec appendix): capability probe (`--help` → feature-gate flags per build), sha256-verified pulls, the binary-discovery ladder + owned-process reaping.
- **The rig is left stopped/free for coding-model experiments** (Michael's call, option b — "keep working on this"). pg-ai-stewards' `default_profile: dance-moe` boot-config is intact for whenever the substrate resumes.
- **MEMORY.md compacted** 20.5KB → 17.2KB (tightened every line; collapsed six superseded substrate-batch index lines into one pointer for future headroom).

## Commits (`cpuchip/llama-chip` main, all pushed)

`30291ee` spec · `a1b37be` default_profile (pg-ai-stewards) · `623f38a` **E1** per-slot backend · `9e6fd45` **E2** pull-ggml · `48eae29` E2 cudart-asset fix · `d511189` **E4** runner-field design grounded in Unsloth Studio.

Memory: `reference_local_coding_models_rig` (E1+E2 built), `project_llama_chip`. Spec: `projects/llama-chip/docs/custom-backend.md`.

## Evening: DiffusionGemma actually runs (E4 validated headless, Studio broken-then-repaired)

Michael: "lets do 2 it'd be VERY fun to try that style." Got **DiffusionGemma generating on the 4090** — and it validated the E4 runner-field design empirically: block-diffusion really is a *different binary* with a *different protocol*, exactly as the spec anticipated.

- **Model:** `unsloth/diffusiongemma-26B-A4B-it-GGUF` Q4_K_M (16GB, fits one 24GB card; Q8_0=25GB doesn't). 26B total / 4B active MoE — the "wacky fast" one.
- **The runner is NOT llama-server.** Both the plain and visual diffusion servers are **stdin/stdout workers** (the Python shim normally fronts them with /v1). But the visual-server's protocol is documented and simple, so I drove it raw — `echo <request-file-path> | NGL=99 llama-diffusion-gemma-visual-server.exe <gguf>` → it streams `F`(per-step canvas)/`C`(committed answer)/`STATS`/`DONE`. **No Python, no Studio, no shim** — the Go-shaped path Michael would prefer.
- **Result:** **~263 tok/s on the 4090** (vs 3.5 tok/s CPU — ~75×), correct iterative `nth_fib` + `print([nth_fib(i) for i in range(11)])`, with a visible `<|channel>thought` reasoning trace. The `F` records ARE the diffusion: the whole answer-canvas refines in parallel over ~10 steps/block, not left-to-right.

**Two bugs the verify-the-real-path habit caught (again):**
1. **Silent CPU fallback.** First GPU run showed 0 MiB GPU used — "on-device sampling unsupported; using host sampling." Unsloth's `ggml-cuda.dll` is built against **CUDA 13**, so it needs `cudart64_13/cublas64_13/cublasLt64_13.dll`; the CUDA-12 runtime I'd staged *loads* but ggml silently drops to CPU. Staged the 13 trio from Unsloth's `torch/lib` → CUDA backend loaded, 263 tok/s. The tell was nvidia-smi showing 0 MiB, not any error.
2. **I broke Studio, then fixed it.** My earlier `install.ps1 --local` *over a running Studio* (PID 55100 held a file lock) **split the venv** — moved the interpreter into a `.rollback` dir, left only the `unsloth.exe` trampoline in `unsloth_studio/` → "uv trampoline failed to canonicalize script path" (Michael's error). Accountability: a clean reinstall fixed it once the gutted dir was removed (`UV_VENV_CLEAR=1`); the full install pulls torch (cu130) regardless of `--no-torch`. Verified: `unsloth studio -p 8888` starts + serves healthy in ~8s, then stopped for a clean state. **Lesson: never reinstall a uv/Python app over its running process** — it mid-air-swaps the venv and corrupts it.

**Carry:** the `.unsloth/studio/unsloth_studio.rollback.20260628225822.*` dir is now orphaned cruft (safe to delete to reclaim a few GB). The headless `llama-diffusion-gemma-visual-server` recipe needs none of the Python stack. DiffusionGemma's GGUF lives at `~/.unsloth/models/diffusiongemma-26B-A4B-it-GGUF/`. Studio UI password is Michael's (not recorded here).
