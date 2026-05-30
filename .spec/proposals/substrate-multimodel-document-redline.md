# Proposal: Multi-Model Document Redline (generative pass) on the substrate

- **Status:** SHIPPED 2026-05-30 (stew session) — R.1–R.6 built + live-verified. Decisions + build notes below. **Pending:** `bin/stewards-mcp.exe` rebuild (MCP-locked) to make `panel_redline` / `panel_redline_condense` callable from Claude Code — the SQL path (`stewards.start_panel_redline`) is live and was used for the end-to-end verification.
- **Raised:** 2026-05-30, from the `projects/scripture-book` ("Beyond the Prompt") 3rd-draft-pass attempt
- **Author:** Claude Opus 4.8 (scripture-book session), at Michael's request
- **Consumer:** the pg-ai-stewards ("stew") Claude session — Michael will point this session here to design the enabling feature
- **Related:** `projects/scripture-book/.spec/workflows/multi-pass-audit.md` (v1/v2 evaluative passes), the v2 COUNCIL (`projects/scripture-book/.draft/20260529-00-COUNCIL.md`), `extension/m4-model-autoprobe.sql` / `model_capability` (M.1–M.4), `start_brainstorm` (`extension/j9c-start-brainstorm-lenses.sql`)

---

## The use case we want to enable

A **generative** multi-model pass (distinct from the evaluative brainstorm we already have):

1. Pick a panel of N specific models (e.g. glm-5.1, kimi-k2.6, qwen3.6-plus, gemini-3.5-flash, gemini-3.1-pro-preview, deepseek-v4-flash, minimax-m2.7).
2. Give **each model the same document** (here: a ~22k-word book manuscript — small enough to fit any of these models' context) plus a **mandate to propose concrete edits** (location + current snippet + proposed replacement + one-line rationale), not abstract critique.
3. **Collect** all N reports.
4. An **orchestrator condenses** the best edits into one ranked proposal doc (in this workflow the orchestrator is the human's Claude Code session, not a substrate aggregator — but a substrate condense step with a chosen model, e.g. gemini-3.5-flash, would also be useful).
5. **Nothing is written to the manuscript by the models** — output is a proposal menu; the human + verifying agent apply selectively.

This is the generative analog of the evaluative brainstorm: instead of "critique the binding question," it's "redline this document."

### Hard constraints the feature MUST preserve
- **Verification gate:** the panel models have no `gospel_get` / canonical access. They must be forbidden from altering scripture quotes or doctrinal claims, and any proposed edit that touches a quote/doctrine must be flagged for human/verifying-agent `gospel_get` review before it can land. (See the 2026-05-26 fabricated-D&C-104 incident — multi-model generation reintroduces exactly that risk if ungated.)
- **Off-disk:** proposals only; never an autonomous edit to the source.
- **Voice preservation:** the mandate must instruct models to keep the author's voice; the human picks among options.
- **Per-model panel + condense:** the value is model diversity *plus* a single condensed menu.

---

## What was attempted (2026-05-30) and why it failed

We tried to realize this on the **current** substrate by repurposing `start_brainstorm`:
- `p_models` to assign each chosen model to a lens (this part works — the J.8/J.9 fix is live; the per-lens model object validates over MCP now).
- The `binding_question` instructed each lens to **read the manuscript from the filesystem** (`/workspace/projects/scripture-book/src/chapters/*.md`) and return redlines.

**Result: the panel could not read the book.** Across **three** dispatches (two free probes + the 7-model run), the models reported the manuscript was not in their sandbox. Two representative quotes from the lens outputs:
- kimi-k2.6: *"I cannot read the files at `/workspace/projects/scripture-book/src/chapters/` — the sandbox does not expose that absolute path, and the repo-root-relative equivalent returned zero matches."*
- qwen3.6-plus: *"The sandbox does not contain files at `/workspace/...`. The only `.md` files available are in `.spec/journal/`."*

7-model run yield: **2 of 7 produced any text, and neither could see the book** (they fell back to generic advice); 5 returned empty or failed (`minimax-m2.7` failed; `glm-5.1`, `gemini-3.1-pro`, `gemini-3.5-flash`, `deepseek-v4-flash` empty). Cost was trivial (pennies) — the empties burned their budget on failed `fs_search` loops.

### Root cause (for the substrate session to confirm)
- The **bridge** container mounts the workspace at `/workspace` (rw) — verified via `docker inspect pg-ai-stewards-bridge`. But the **lens dispatch's fs sandbox does not resolve the manuscript path**; the models consistently see only something like `.spec/journal/`. So the bridge mount does **not** translate into lens file-read access to arbitrary repo paths. The fs scope for lens/brainstorm dispatches appears narrow (and differs from what `audit_files`' subagent can reach — `audit_files` is "restricted to fs_read/fs_search/fs_list," suggesting fs access is configurable per pipeline).
- Brainstorm lenses are **ideation-framed** (SCAMPER, Six Hats…), not document-redline-framed.
- No **per-lens `max_tokens`** override is exposed; reasoning models (glm-5.1) and large reads appear to exhaust the per-lens budget before producing content (consistent with the v2 glm-5 "empty" being a budget/transient, per `model_capability`).

---

## What would enable it (options for the substrate — not prescriptive)

Any one of these would unblock the workflow; (1) or (4) are probably the cleanest:

1. **Document-context injection on dispatch.** A dispatch param (on `start_brainstorm` and/or `spawn_subagent`) that injects a document — by inline text, by a file path the dispatch is granted to read, or by a registered corpus id — into each child's context before the lens/agent prompt runs. The panel then never needs ambient fs access.
2. **Corpus index for arbitrary project files** (not just the studies corpus). Index the manuscript (or any glob) as a corpus the lenses can query via `read_corpus_parents` / a corpus-read tool. Pair with per-lens corpus scoping.
3. **Parameterized lens fs scope.** Let a dispatch grant read access to a specified path (e.g. `/workspace/projects/scripture-book/src/chapters`) so `fs_read`/`fs_search` resolve it. (Mind the security posture — scope per call, don't open all of `/workspace` by default.)
4. **A dedicated `redline` pipeline_family** — the generative analog of `audit_files`, but with a **per-call model override** and a **document/glob** argument. Input: `(model, glob | document, mandate, cost_cap, max_tokens)`. Output: structured, location-anchored edit proposals. This is the most direct fit and keeps the evaluative `start_brainstorm` rails unchanged.
5. **Per-lens `max_tokens` override** (independent of the above) so reasoning models and large-context reads get adequate output budget, and a cost cap that accounts for large document reads. Without this, even a fixed fs/document path will keep returning empties on the pricier/reasoning models.

---

## Decisions ratified — 2026-05-30 (stew session, AskUserQuestion)

Root cause **confirmed against the code** (not just hypothesized): the `fs-read` MCP server enforces a **single global allow-list** (`stewards.mcp_servers.args` for `name='fs-read'`: `.spec/journal/*`, `.spec/proposals/*`, `.mind/*`, `docs/**`, `projects/space-center/**`). `projects/scripture-book/src/chapters/` is not in it, and `cmd/fs-read-mcp/sandbox.go` rejects absolute paths — so the lenses' `/workspace/...` attempts failed on path-form AND allow-list. Per-pipeline `fs_read_paths` was explicitly deferred (`h3-followup-1`). **Correction to the original draft:** `audit_files` is *not* a privileged fs path — it spawns a subagent against the same fs-read server + same allow-list, so it would have failed on the manuscript too. There is no existing mechanism that hands a document to a model outside the allow-list. **That** is the gap.

- **D-RL1 — Dedicated `redline` pipeline_family + `panel_redline` wrapper** (option 4), NOT an extension of `start_brainstorm`. Keeps the evaluative brainstorm rails clean; mirrors the `audit_files` → `subagent-files-audit` pattern. Generative analog: redline-framed prompt + location-anchored structured output.
- **D-RL2 — Document delivered server-side at dispatch time** (option 1), NOT model-side fs tools and NOT the MCP-host wrapper. Chosen for generality: works for autonomous/scheduled redline dispatches, not only human-/MCP-initiated ones. This sidesteps the allow-list entirely — the model never needs fs access, so no `fs_search` budget-burn, no sandbox, works for models without fs tools.
  - **REVISED 2026-05-30 (topology finding, CONFIRMED by Michael):** the ratified "bgworker reads `/workspace`" is not implementable as stated — the **bgworker runs inside the pg container (`pg-ai-stewards-dev`), which has NO `/workspace` mount** (only `/docker-entrypoint-initdb.d`, `/opt/yt/yt`, pgdata). Only the *bridge* mounts `/workspace`, and the bridge does tool dispatch, not chat-body composition. **Lighter corrected path (verified):** add a **read-only `/workspace` bind mount to the pg container** (docker-compose change + container *recreate*, soak pause — **no extension/Rust rebuild**) and inject the document in **SQL at compose time via `pg_read_file`** (confirmed working on the existing `/opt/yt/yt` ro mount for the stewards superuser; `pg_ls_dir` enumerates globs). Same goal (server-side, autonomous-capable), strictly lighter than the bgworker-Rust path. **Security:** the pg superuser gains ro read of the whole repo, so the redline read-fn MUST validate the requested path against an allow-list (refuse `.env`, secrets, paths outside a configured doc-root) — otherwise a crafted call could read a sensitive file and leak it into an external model's prompt. The models still never get fs access; only the substrate's own SQL reads files.
- **D-RL3 — Condense: both.** Default returns the N raw model reports; an optional `condense_model` param triggers a substrate aggregator child (a chosen model) that emits the ranked menu. Without it, the orchestrating Claude session condenses (today's v2-audit workflow).
- **D-RL4 — Build-in defaults (not separately ratified):** per-dispatch `max_tokens` override (option 5); the verification-gate mandate forbidding scripture/doctrine edits + a structured `touches_quote_or_doctrine` flag per proposed edit; a document size guard; off-disk (proposals only).

### RESOLVED — `max_tokens` = 32k per-call default (Michael, 2026-05-30)
Confirmed against the code: the substrate currently sets **no** `max_tokens` at all (`dry_run_chat` → `compose_messages` never adds it; es7 documents the convention as intentional). So normal dispatches run at the gateway's per-model default, and the redline empties were NOT a low output cap — they were the lenses burning `cost_cap_micro` on failed `fs_search` loops (no readable document). Injection (D-RL2) removes that.

**D-RL5 — per-dispatch `max_tokens`, default 32k, PER API CALL.** `max_tokens` is the output/completion ceiling on each individual chat request (each generation round), **not** a work-item total. The work-item total stays governed by `cost_cap_micro` (money) + `token_budget` (tokens) — those remain the runaway guards, untouched. So opening output budget doesn't weaken runaway protection; the two are different dials. For reasoning models, reasoning tokens count against `max_tokens` (glm-5: empty at 80, clean at 800), so 32k gives ample room to reason AND answer. Redline is single-shot + tools-disabled, so 32k ≈ the whole redline output. Notes: (a) many models silently cap output below 32k regardless — harmless; a few 400 on over-ask; (b) `cost_cap_micro` for a redline child must be set to fund 32k × the model's output rate or the cost gate quarantines mid-answer (auto-scale it from `model_pricing`); (c) future: record per-model true max-output in `model_capability`, let the auto-probe discover it, so the dispatcher never over-requests. Cost context (Michael): the marginal call is cents — opencode-go is $10/mo for ~$60 of tokens, gemini calls run <$0.01 (the whole book redline would), zen + gemini prepaid ~$20 each — so generous per-call output is fine; `cost_cap` catches pathological runaways.

### Acceptance criteria (suggested)
- A panel of ≥5 chosen models each receives the same ~22k-word document and returns ≥6 concrete, **location-anchored** redlines (quoting real text from the document), with a non-empty yield ≥ 80%.
- Per-call model override + per-call `max_tokens`/cost cap honored.
- The document is delivered without the *orchestrating* agent having to inline it by hand (which is infeasible/unsafe at 30k+ tokens — it would mean retyping the manuscript from memory, the exact fabrication risk we forbid).

---

## One salvageable result (despite the failure)

Even blind to the text, qwen3.6-plus reasoned to an insight that **converges with the v2 audit** (T1.2 / the imago-Dei refrain redundancy):

> *"This book's entire architecture risks saying everything twice — once in AI terms, once in scripture terms — when the most powerful move is to say it once, well, and let the reader see the connection."*

Caveat for the book: the dual-domain *parallel* is the book's point; the fix is cutting *redundant re-walks*, not the parallel. Logged here only as a second witness to a finding we already have.

---

## Recommendation for the book in the meantime

Do **not** force the generative pass on the current substrate. Continue the v2 evaluative approach (which works) + run the voice work (T1.3) as the planned **Gemini 3.5 Flash voicing pass**. Revisit this generative workflow once the substrate exposes document delivery to a model panel (one of the options above).
