# Codewright workspace roadmap — read-only researcher → working engineer → model-orchestrator

**Status:** A SHIPPED (2026-06-09); B + C DESIGN-ONLY (Michael's vision, captured this session).
**Binding question:** How does codewright grow from a read-only code-Q&A chat bot into an
engineer that works inside a repo+env, and eventually orchestrates other model-CLIs
(`agy -p`, `opencode`, `claude -p`) from its own container?
**Where:** pg-ai-stewards (substrate). Consumes the coder sandbox machinery (CC/CV2).

## The progression Michael named

```
A (done)        B (next)                C (the vision)
read-only   →   works in a repo+env  →  orchestrates other model-CLIs
research        clone/build/test/edit    agy -p · opencode · claude -p
allow-listed    persistent workspace     codewright as the conductor
```

## A — repo awareness + public-repo scope (SHIPPED 2026-06-09)

The live-chat gap: codewright couldn't tell anyone what it could see (the allow-list
was an invisible bridge env var) and was scoped to one repo.

- **`list_repos` tool** (stewards-mcp, mcp_proxy) — reads the SAME
  `CODER_REPO_ALLOWLIST`/`CODER_REPO_DENYLIST` the coder sandbox enforces, so the
  persona's answer == reality. codewright now answers "what can you look at?"
- **Scope = all Michael's PUBLIC repos.** `CODER_REPO_ALLOWLIST=github.com/cpuchip/`
  with a **deny-beats-allow** `CODER_REPO_DENYLIST=private-study`. The denylist is
  load-bearing: the bridge clones with a `GITHUB_TOKEN` that CAN reach private repos,
  so a broad allow substring would otherwise expose them. Future private repos →
  add to the denylist. (r14 + coder-mcp `repoAllowed` denylist check.)
- Still **read-only + ephemeral sandbox per call** — the security posture is unchanged.

## B — work inside a repo + env (the persistent workspace)

The jump from "reads a repo" to "an engineer working in one." Already 80% built: the
`coder` MCP (CC.1–6, CV2.1–2.3) has sandbox_start/read/write/edit/grep/glob/shell/
commit/push/open_pr against a repo-mounted worktree, and `code-pr` runs the full
clone→plan→implement→verify→pr cascade. **B is mostly wiring codewright to that, plus
making the workspace persistent.**

### ★ Measured latency (2026-06-09, live Engineering turn on ai-chattermax)

Where a research_codebase turn's ~47s actually goes (timestamps from
`chattercode-b03e5b16` / `subagent-20260609-231537`):

| step | time |
|---|---|
| 🐳 sandbox_start = **fire up Docker container + clone the repo** | **1.4s** |
| glob (list *.go) | 0.5s |
| grep (auth patterns) | 1.0s |
| read × 3 (full files) | 0.3–0.5s each |
| sandbox_stop | 0.8s |
| **all container + clone + search + read** | **~5–6s** |
| model (kimi) reasoning between calls + 11.9s final synthesis | **~40s** |
| **total** | **~47s** (a second live call was 67s — all model variance) |

**The headline: for a SMALL single repo, the container+clone is ~1.4s — ~3% of the
cycle. ~85–90% is the model thinking.** So the latency lever for small/single-repo Q&A
is **a faster model** (the v2 `model_override` we already designed), NOT a container
redesign. Conversational turns (no research) are 1–10s.

**BUT (Michael, 2026-06-09) the clone term is not constant — it scales, and that's
where the persistent container earns its keep:**
- **Repo size.** ai-chattermax is small (~1.4s shallow clone). A large repo / monorepo
  is 10s–60s+ to clone. As repo size grows, "clone every question" goes from negligible
  to dominant — and a warm clone (`git pull` deltas only) wins big.
- **Multi-repo working sets.** A question spanning a few repos = N clones per turn
  today. A persistent container holds N repos warm → one-time cost, then near-instant.
  Payoff ≈ (repo size × repo count × question frequency).
- **Statefulness.** Build artifacts, language-server indexes, accumulated edits — none
  survive an ephemeral sandbox; all persist in a warm one. This is B/C's real prize,
  separate from latency.

**Net:** persistent container ≠ "faster" for the small-single-repo case (model is the
lever there); it = "faster + stateful" for **large repos, multi-repo sessions, and any
work that accumulates** (build/test/index/edit). So the decision rule is a function of
the working set, not a flat yes/no.

Design points to settle at ratification:
- **Persistent vs ephemeral.** Today each call clones fresh (small repo ≈ 1.4s clone +
  ~45s model; large/multi-repo clone dominates). A persistent per-codewright sandbox
  keeps repos warm (clone once, `git pull` to update) — wins as repo size / count grow,
  and is the only way to keep accumulated state. Cost: lifecycle (when to pull/prune,
  disk), and a standing container is an attack surface even when idle.
- **★ The security inversion (the real gate).** Ephemeral + read-only + allow-listed
  is tight. A *persistent, writable* container driven by **kimi (a weaker model than
  the orchestrating Opus/Fable) reading untrusted repo content** is the prompt-
  injection-exfiltration profile the Google-MCP vet flagged. Mitigations to design in:
  keep write/push/PR gated (human Hinge on merge, as code-pr already does); no
  secrets in the workspace; the denylist still applies; consider a per-persona model
  floor (a stronger model when the persona can write).
- **Concurrency.** One persistent container + many rooms → serialize turns or pool
  workspaces. The current per-work-item worktree model already isolates; a persistent
  one needs a lock or a small pool.
- **Scope creep check.** B overlaps `code-pr`. The question is whether codewright
  *becomes* a code-pr front (chat → dispatch a code-pr work_item → report the draft
  PR) rather than getting its own parallel build machinery. Reuse, don't duplicate.

**Recommendation:** B v1 = codewright can *dispatch a code-pr work_item* from chat
("@codewright add a healthcheck to spin") → the existing cascade runs → it reports the
draft PR link. No new container architecture; the persistent workspace is a separate,
later optimization once the latency/warmth need is proven.

## C — codewright orchestrates other model-CLIs (the vision)

A persistent container with `agy -p` (Antigravity/Gemini 3.5 Flash), `opencode`, and
`claude -p` installed + authed, so codewright dispatches work to *other models' agents*
and composes their output — a conductor, not just a worker. This is the
[[project_council_review_beats_gift_matching]] idea made concrete: many doers, one
critic/composer, but with real CLI agents instead of single dispatches.

**★ The crux is auth — Michael's open question ("pre-setup the container??").** Each
CLI authenticates differently, and none should bake a secret into an image layer:

| CLI | Auth | Container approach |
|---|---|---|
| `claude -p` | `ANTHROPIC_API_KEY` (cleanest) OR Max-plan OAuth (interactive, hard to bake). Note: the 2026-06-15 change makes `claude -p` a SEPARATE credit pool from interactive. | Mount `ANTHROPIC_API_KEY` via env/secret at run; never in the image. |
| `opencode` | `opencode auth login` writes a token file (`~/.local/share/opencode/auth.json` or similar) OR an API key. | Pre-provision the auth file into a mounted secrets volume; or env key. |
| `agy -p` | Google account / Gemini API key (the agy-cli skill drives it headless). | Mount the credential; `agy` reads it. |

**The pattern:** a **long-lived container with a mounted secrets volume** (env file +
credential files), injected at `docker run`, never in the image. "Pre-setup the
container" = exactly this: provision credentials once into the mount, the container
reads them, the model never sees the raw secret (it just runs `claude -p "..."`).

Design points for C:
- **Cost governance.** Each CLI call is real spend across THREE billing pools (Anthropic
  credits, opencode sub, Gemini). The substrate's cost_buckets track opencode; claude/
  gemini need their own tracking or a hard per-container cap. Don't let a chat persona
  spawn unbounded `claude -p` calls.
- **Which model conducts?** codewright (kimi) deciding when to invoke claude vs opencode
  vs agy is itself a gift-matching/council question — and our finding says the *critic*
  is the lever, not the matching. So C's value may be "codewright drafts, claude -p
  reviews" more than "route each task to its best model."
- **Isolation.** Three model-CLIs with network + credentials in one container is a fat
  target. Egress controls, no host mounts beyond the secrets + workspace, and the
  denylist still governs what gets cloned.

**This is genuinely new architecture and deserves its own ratification + security pass.**
Not a bolt-on to A.

## Recommended sequence

1. **A — shipped.** Let it run; see how codewright behaves with real repo scope.
2. **B v1 — codewright dispatches code-pr from chat** (reuse the cascade; human Hinge on
   merge). Cheapest path to "works in a repo," no new container.
3. **Decide the persistent workspace** by feel (only if the clone-latency/warmth need is
   real after A+B).
4. **C — the model-CLI container** as its own ratified project, auth + cost + isolation
   designed up front. The auth pattern (mounted secrets volume, never in-image) is the
   anchor.

**One-liner:** A made codewright *honest about its reach*; B makes it *act in a repo*
(reuse code-pr); C makes it a *conductor of other models* (and the hard part there is
auth + cost across three billing pools, solved by a pre-provisioned secrets mount).
