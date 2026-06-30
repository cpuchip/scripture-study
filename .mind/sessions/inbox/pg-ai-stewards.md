<!--
Inbox compacted 2026-06-30 (lodestar/world-graph session). All signals below were
read + registered; durable artifacts live on disk. Kept as lean pointers so 📬 resets
without losing the threads. Full original bodies are in git history of this file.
-->

## 📬 registered — open research threads (from general-workspace), not blocking

- **2026-06-30 · DETERMINISTIC-CORE LORE (statecharts for world-graphing)** — `.spec/proposals/deterministic-core-lore.md` (in THIS OSS repo, `7627e14`). ★ Directly adjacent to **#291 world-graph**: model lore entities as XState-v5 statecharts (state + typed events + guards, JSONB) + a `world_events` log = the missing deterministic core / temporal dimension; LLM proposes→engine validates. Borrow `@statelyai/graph` + Sketch simulator + Mermaid I/O (cloned `external_context/stately-sketch`). View stack: keep 3d-force-graph, add Vue Flow (editable) + Cytoscape (analytics) + Mermaid. **First step: one quest → statechart + guard + loremaster proposes an event.** → pick up when the world-graph grows a *state/advance* dimension (after lodestar's structural extraction lands).
- **2026-06-29 · Hinge persistent stream-json + loom + OpenHands** — `claude -p --input-format stream-json` holds multi-turn context (cost win proven); `cpuchip/loom` (`projects/loom`, MIT) is the harness-around-harnesses for long-lived Hinge sessions; OpenHands (`external_context/openhands`) patterns to steal = condenser, skills-enumeration, in-sandbox agent-server. → relevant when the coder/Hinge pipeline next moves.
- **2026-06-28 · llama-chip custom-backend** — `projects/llama-chip/docs/custom-backend.md`; per-slot `backend` override (E1) + `pull-ggml` (E2), proven isolated on `:8095`. Needs a rebuild+restart = **my call** (I steward the running exe). `dance-moe` profile needs reloading before the substrate next needs local models (coordinate the GPU0/GPU1 swap with general-workspace). NOTE: lodestar is deterministic (no LLM), so it doesn't depend on the rig.
- **2026-06-28 · OKF v0.1 boundary adapter** — `study/yt/open-knowledge-format-okf-for-pg-ai-stewards.md`; `okf_export`/`okf_import` shelf item; build when a share/ingest need is real (new capability → council).
- **2026-06-25 · Loreworks/Boyd demo + DeepLore transfers** — `study/ai/harness/provenance.md`; Boyd's *Patterns of Conflict* as the first non-fiction world (orientation-graph); four DeepLore steals (summary-as-retrieval-hint, contextual gating, grow-during-play gap-flag, provenance trace). → Loreworks chunk F positioning.
- **2026-06-16 · digester-reads-repos** — largely met by `doc_import_corpus` (zip-a-repo); live read-only `git clone` into the sandbox is the noted future enhancement (council if pursued).

## ✅ resolved (shipped this arc — tombstoned)

- BINEVAL (#287, `79-bineval.sql`) · north-star (#283, `74-north-star.sql`, PR #13) · trajectory-critic (#269, `56`) + self-improvement (`59`) · yt-slide-frames Part B (#285) — all built. Vibe-Coding / Google-SDLC papers durable in `external_context/google-new-sdlc/NOTES.md`.
