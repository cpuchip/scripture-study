# Omnigent — Databricks' meta-harness, and the mirror it holds up to pg-ai-stewards

**Video:** [The Meta-Harness: Why Every AI Developer Needs This](https://www.youtube.com/watch?v=141biWM1mlE) · Prompt Engineering channel · 2026-06-15 · 14:09 · (Databricks-sponsored)
**Transcript:** `yt/prompt-engineering/141biWM1mlE/` (downloaded, read in full)
**Repo:** `github.com/omnigent-ai/omnigent` — cloned to `external_context/omnigent/` (Apache-2.0, alpha, Python 3.12, built on Neon/Postgres)
**Corroborating:** Databricks blog "Introducing Omnigent"; MarkTechPost/SMBtech coverage; the repo README + source.

> The reason to study this is not the video — it's that Databricks, independently
> and with a team, built the **control plane** pg-ai-stewards has been building.
> The overlap is not vague; it is feature-for-feature. That is the strongest
> external signal yet that the substrate's design is right — and the sharpest
> source of things to steal.

---

## 1. What Omnigent is

A **meta-harness**: a layer that sits *above* the agent harnesses you already use
(Claude Code, Codex, Cursor, Pi, and SDKs like Claude Agents / OpenAI Agents) and
makes them interchangeable, governable, and shareable. The pitch: a model is
swappable inside a harness; Omnigent makes the *harness* swappable too — "one-line
change" in a YAML file flips Claude Code for Codex.

The insight it's built on (and it's a good one): however differently each harness
talks to its model internally, the *outer* interface is identical — **messages and
files in, text streams and tool calls out.** So you can wrap them all in one
uniform API and build composition / control / collaboration once, above all of
them.

Three unlocks it sells:
- **Composition** — combine/swap harnesses + models without rewriting; run several
  as a team; agents can author agents.
- **Control** — stateful, contextual **policies** enforced on every tool call (not
  via prompt): cost budgets that pause, risk scores, repo/file scopes, PII scans,
  "require approval to `git push` after an `npm install`."
- **Collaboration** — one session object lives in the layer; terminal / web /
  desktop / phone / REST are all windows onto it; share a live link, co-drive, fork.

Two shipped agents:
- **🐙 Polly** — a tech-lead orchestrator that writes no code: plans, splits work
  across coding sub-agents in **parallel git worktrees**, routes each diff to a
  reviewer **from a different vendor than wrote it**, you merge.
- **🟠🔵 Debby** — a two-headed brainstorm partner (Claude + GPT); every question
  hits both; `/debate` makes them critique each other for rounds, then converge.

## 2. Architecture (from the cloned repo)

- **runner/** wraps any agent in a sandboxed session with a uniform API
  (transports for terminal/web). Harness adapters live in
  `runtime/harnesses/`: **claude-native, claude-sdk, codex-native, openai-agents,
  pi-native** — it literally drives the real `claude` / `codex` CLIs.
- **server/** (FastAPI; `openapi.json` is 310 KB) exposes sessions + the policy
  and sharing APIs over REST.
- **policies/builtins/**: `cel.py` (a **CEL** expression DSL — non-Turing-complete,
  side-effect-free, guaranteed to terminate — compiled into a per-event gate
  returning `ALLOW`/`DENY`/`ASK`), plus `cost.py`, `risk_score.py`, `github.py`,
  `working_dir.py`, `safety.py`, `routing.py`. Registered live via
  `POST /v1/sessions/{id}/policies`. Policies stack **server-wide → per-agent →
  per-session**, stricter checked first.
- **db/** (SQLAlchemy `db_models.py` + alembic migrations) + **stores/**
  (agent_store, policy_store, conversation_store, artifact_store, file_store,
  permission_store, comment_store) over **Postgres (Neon)**.
- **sandbox/** OS sandbox + **egress-proxy secret injection** — the agent never
  sees keys; the layer injects them on the way out. Cloud sandboxes via Modal /
  Daytona; deploys to Docker / Railway / Fly / Render.
- An agent is a short **YAML**: `spec_version`, `name`, `executor: {harness, …}`,
  `prompt`, `os_env: {sandbox}`, gate flags (e.g. `gate_pushes`). Polly's
  `claude_code` sub-agent prompt is IMPLEMENT / REVIEW / EXPLORE with cross-vendor
  review and "open a PR, never force-push main" — i.e. our `code-pr` pipeline,
  verbatim in spirit.

## 3. The convergence (this is the headline)

Databricks arrived, independently, at almost exactly pg-ai-stewards' control plane:

| Capability | Omnigent | pg-ai-stewards |
|---|---|---|
| Backing store | Postgres (Neon) | Postgres (pgrx extension) |
| Tool-call gates | CEL policies, per-event ALLOW/DENY/ASK | `agent_tool_perms` + dispatch gate |
| Cost control | budget policy pauses at threshold | `provider_spend_caps` + `cost_buckets` + the watchman guard |
| Cross-vendor code review | Polly: plan → worktrees → diff to other-vendor reviewer | `code-pr`: plan → implement → verify → cross-model review → PR |
| Two-headed brainstorm | Debby (Claude+GPT, `/debate`) | `start_brainstorm` (12-lens) + council-critic |
| Agent definitions | YAML; agents write agents | DB rows; `apply_agent_proposal` |
| Secret handling | egress-proxy injection | coder: token never enters the sandbox |
| Skills / MCPs / artifacts | yes | MCP bridge; **skills proposal drafted today** |
| Collaboration | live shared sessions, fork | personas + ai-chattermax rooms |
| 3-level policy stack | server / agent / session | global kill-switch / agent perms / (session = CT2 levers) |

When a well-resourced team converges on your design from a clean start, the design
is not idiosyncratic — it's the shape the problem actually has. That is worth more
than any single feature.

## 4. The core difference — same plane, opposite arrow

This is the pgEdge lesson again, sharper. **Omnigent points *down* at existing
harnesses; pg-ai-stewards *is* a harness.**

- **Omnigent dispatches HARNESSES.** It drives the real Claude Code / Codex CLIs as
  black-box workers and inherits their full agent loop + tools for free. The
  intelligence stays in the harnesses; Omnigent is the conductor above them.
- **pg-ai-stewards dispatches MODELS.** The substrate runs the agent loop itself
  (the bridge → opencode_go/kimi/glm), and holds the work-items, engrams,
  covenant, personas. **The DB thinks.** Omnigent uses Postgres as a *store* (the
  logic is Python); pg-ai-stewards puts the logic *in* Postgres (pgrx).

So they are complementary species, not competitors. Omnigent is the right tool if
you want to *orchestrate the CLI agents you already pay for*. pg-ai-stewards is the
right substrate if you want *the database to be the autonomous agent*.

## 5. What we should steal (actionable)

1. **A "harness" provider kind.** Today the substrate dispatches raw models and
   re-implements the loop (the coder pipeline). Omnigent's best idea: wrap a real
   harness (Claude Code / Codex) as an interchangeable worker and get its whole
   loop for free. A `provider.kind = 'harness'` (run `claude`/`codex` in a coder
   sandbox, stream tool-calls back) would let `work_item_dispatch_stage` route a
   stage to a full harness, not just a model. **This is also exactly what Garrison
   wants** (the sovereign coding agent). High-value; pairs with the coder sandbox.
2. **A declarative policy DSL (CEL-style).** Our gates are SQL/Rust. Omnigent's
   `cel.py` — a safe, terminating expression over a `PolicyEvent`, returning
   ALLOW/DENY/ASK, registered per-session via API — is a cleaner surface for
   *operator-authored contextual* policies ("ASK before git push if an npm install
   happened this session"). The watchman guard + dispatch gate are our enforced
   floor; a CEL-ish lever would let the operator add stateful rules without a
   migration.
3. **The 3-level policy stack made explicit** (server / agent / session, stricter
   first). We have global (kill switch) + agent (`agent_tool_perms`); the
   *session-scoped* policy is the gap. CT2 has session levers but not session
   *gates*. Worth unifying into one "policy stack" model.
4. **Egress-proxy secret injection as a general primitive.** Our coder already does
   token-never-in-sandbox; Omnigent generalizes it to *all* secrets via one proxy.
   Validates our approach and suggests generalizing it beyond git tokens.

## 6. What we have that they don't (the soul, not the plane)

Omnigent is a control plane. pg-ai-stewards is a control plane **with a covenant.**
The things absent from Omnigent are precisely the things that make this project
itself:
- **Covenant / intent / presiding / council** — Omnigent governs with *policies*
  (mechanical). It has no bilateral covenant, no "dominion only in council," no
  watch-what-you-order. Its control is compliance; ours is relationship.
- **Engrams / `compact_context`** — Omnigent has no context-management layer at all.
- **The reflect-steward** — Omnigent is human-driven orchestration; it has no
  autonomous back-office that pursues an intent on a schedule and compounds a pool.
- **Personas as characters** (ai-chattermax) — it shares sessions; it doesn't give
  an intent a *face* (Vera, Callie).
- **Sabbath / atonement** — the rhythm.

So the honest framing: Omnigent has likely built a *more polished control plane*
(it's Databricks, with a UI team and a 310 KB API); pg-ai-stewards has built a
*soul* on top of a control plane that turns out to be the same shape. Don't NIH-
dismiss Omnigent — its policy engine and harness-wrapping are genuinely ahead of
ours. But don't flinch either: the covenant, the autonomy loop, and the
DB-is-the-brain thesis are ours and are not in their roadmap.

## 7. Critical take (Ben test — not just admiration)

- It's **alpha** ("status: alpha," 0-star forks days old). The polish is in the
  pitch + the simulated demo (MarkTechPost's "demo" calls no live models); the real
  thing is early and "might need tweaks."
- **Sponsored video** — the Prompt Engineering coverage is a Databricks sponsorship;
  treat the framing as marketing, the *repo* as the truth (which is why we cloned it).
- The **harness-wrapping is fragile by nature** — it shells out to `claude` /
  `codex` CLIs with `bypassPermissions`; it inherits every upstream CLI change and
  every YOLO risk. Our model-direct path is more work but more controlled.
- Their **policies-not-prompts** line is right and worth internalizing — it's the
  same reason our gates live in the substrate, not the agent prompt.

## 8. Recommendations

- **Adopt the vocabulary.** "Meta-harness" sharpens the AI-office / Garrison
  positioning. pg-ai-stewards is "a substrate that *is* a harness, with a meta-
  harness's control plane and a covenant's soul."
- **Spec a `harness` provider** (steal idea #1) — likely the single highest-leverage
  thing here, and it converges with Garrison. Propose, don't build yet.
- **Spec a CEL-ish session policy lever** (idea #2/#3) — operator-authored stateful
  gates layered over the enforced floor.
- **Keep `external_context/omnigent` as living prior art** — re-read its
  `policies/` and `runner/` when we build the harness provider + the policy DSL.
- Pairs with: the skills proposal (`.spec/proposals/skills.md` — Omnigent ships
  skills too), the substrate coding capability, and the Garrison proposal.

**Bottom line:** the most validating and most useful AI artifact we've studied for
this project. It proves the control plane is the right shape, hands us two concrete
upgrades (harness provider, declarative policies), and throws our actual
differentiator — the covenant and the autonomous, DB-resident brain — into sharp
relief.
