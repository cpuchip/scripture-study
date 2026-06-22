## 📬 2026-06-19 (from pg-ai-stewards) — review the local-model / doc-construction session; apply to Garrison if it transfers

**What happened here:** drove pg-ai-stewards onto the local FlexLLama rig; an overnight soak surfaced
3 small-model failure modes, then two fixes landed + were e2e-proven on local:
1. **Parallelism / rig:** gemma `--parallel 1` WEDGES under concurrency; **gemma q8 `--parallel 4
   @524288` = 4×131k slots, 248 tok/s**, fits beside nemotron (one instance beats two — shared weights
   + continuous batching). And **KV must live in DEDICATED VRAM** (Windows/WDDM spills the overflow to
   system RAM → ~3× slowdown) → q8 KV is the enabler (qwen/nemotron/gemma).
2. **Agentic doc-construction:** small models **BUILD** a large artifact via tool-call diffs
   (doc_create/append/patch/finalize) + a **journal** final reply, instead of one-shot generation
   (which trips the reaper, contends for the slot, and 500s on grammar). Proven on qwen — the model
   that failed one-shot ran 0-error building via tools.

**Apply to Garrison if possible** (it borrows the same `:8090` rig + already builds code via edits):
- the `--parallel` / q8 rig-config lessons (its dispatcher hits the same FlexLLama endpoint);
- **journal-as-output** — final reply = a short account of what it did (like the coder), not the artifact;
- **source-page-in** — don't echo big files/sources into the model; cache + page them in on demand;
- the WDDM **"KV in dedicated VRAM"** rule for any local model Garrison loads.

Detail: `pg-ai-stewards-oss/.spec/proposals/{agentic-doc-construction, local-throughput-experiments,
local-learnings-rollout}.md` + journal `2026-06-19-local-soak-and-doc-construction.md`.

— filed by pg-ai-stewards; not yet acted. Review when convenient; not blocking.
