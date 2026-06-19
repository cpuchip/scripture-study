# 2026-06-18 — Garrison: council closed, P0 ratified

**Mode:** plan / council (with Michael) · **Lane:** general-workspace

Michael set voicebox down ("a fun excursion") and picked up Garrison — the lean,
sovereign, local-first coding agent (the go-bag for "if Claude Code and frontier
access vanished tomorrow, code on owned hardware with a weak local model"). He
chose **close the council first** over building or dogfooding. So this session
resolved the six open questions and ratified P0. Nothing built.

## The unlock that made it buildable

Garrison was gated on two things: `dominion_in_council` and **post-cut**. The
substrate cut executed 06-15 (OSS is the one substrate, live retired) — so the
post-cut gate is satisfied. Garrison moved from "design-only, blocked" to
"buildable, pending ratification." Surfaced that first; it reframed the whole
session from "keep designing" to "close it out."

## The six, resolved (full text in the spec's "Decided in council — P0 CLOSED")

1. **Tool-calling / pairing** — two-model split (tool-tuned/fast doer + reasoner
   planner), one OpenAI client routed by alias; oracle gate is the real safety net.
2. **pg-backend boundary** — two Go interfaces, `Ledger` + `Dispatcher`; substrate
   supplies pg-Ledger and/or MCP-Dispatcher adapters at P4, independently selectable.
3. **Presiding-chain surface** — terminal-native `garrison watch` TUI (bubbletea),
   ships *with* sub-agent spawning, which is gated behind oracle+critic existing.
4. **Tier-2 self-extension** — P5, own council item; "hang with consent" = the
   reflect-steward's approve→queue→capacity-drain (build+test free, gate install).
5. **Plugin relationship** — yes; `plugin-someday` is Garrison's luxury-mode MCP
   client. Garrison is both a standalone loop AND an MCP server. Design the core
   MCP-exposable from P1 (locked implication).
6. **P1 dogfood** — a greenfield Go utility + tests (hard `go build && go test`
   oracle); the same task is the pi + qwen3.6-27b baseline drive first.

Plus the **store** call Michael decided directly: **FTS5 + chromem-go (RRF) in
P1** (full keyword+semantic from day one), not deferred — resolving the spec's
flagged `sqlite-vec`-vs-pure-Go contradiction via the gospel-engine lineage.

## Michael's model steer (the concrete "copy over")

Instead of Devstral-first, **start on the pg-ai FlexLLama rig** — borrow the same
models the substrate work already runs. Peeked at `projects/pg-ai-stewards-oss/.env`
and `external_context/flexllama/stewards-3way.json`: one OpenAI endpoint at `:8090`
routing three hot aliases by name (`qwen3.6-27b` GPU0, `gemma-12b` + `nemotron-4b`
GPU1; all Q4, n_ctx 32768, jinja on). Embeddings — CORRECTED 06-18: not Ollama
(it isn't installed) but **LM Studio** `text-embedding-nomic-embed-text-v1.5` at
`:1234` (768-dim, matches gospel-engine-v2's nomic → apples-to-apples for chromem-go).
All non-secret (keys are placeholders `flexllama`/`ollama`). FlexLLama's
"one endpoint, route by alias" *is* Garrison's "one client, model-by-alias" design —
a clean borrow. Recorded as a table in the spec ("Starting model configuration")
rather than in code, since P1 isn't scaffolded; P1's first act seeds
`projects/garrison/.env.example` from it.

## Carry-forward

- **P1 is now buildable on Michael's go.** Standalone MVP: lean Go loop +
  FlexLLama `:8090` (qwen3.6-27b) client + embedded SQLite ledger with FTS5 +
  chromem-go (RRF). No Docker. Dogfood = greenfield Go util; drive pi +
  qwen3.6-27b on that same task as the baseline first.
- Deferred-to-P4: shared-ledger-first vs dispatch-offload-first (lead: shared).
- The spec is the record; nothing committed/pushed this session yet.
- voicebox is parked mid-stream: MCP config added to `.mcp.json` (awaiting a
  session restart for the tools); model-download blocked by the Windows
  HuggingFace symlink error (Errno 22) — Michael's env to resolve.

## Update — P1 floor shipped (same session, after the council close)

Michael gave the go ("create the github repo... and start building") and corrected
the embedding endpoint (LM Studio, not Ollama). P1 began:

- Repo `cpuchip/garrison` (private) created + cloned to `projects/garrison/`
  (gitignored from root via `/projects/garrison/`).
- Stdlib-only Go scaffold: `internal/config` (.env loader), `internal/llm` (one
  OpenAI-compatible client — chat + embeddings + models), `cmd/garrison` (CLI).
- Local `.env` wired to the verified rig: chat → FlexLLama `:8090`, embeddings →
  LM Studio `:1234`.
- **Model path PROVEN against live endpoints (the oracle):** `garrison ping`
  (both up, all 4 role models served), `embed` (768-dim nomic vector), `chat`
  (nemotron-4b 1s + qwen3.6-27b a correct iterative fib in 34s). `go build` +
  `go vet` + real round-trips, not mocks. Committed + pushed to main.

NEXT: the embedded SQLite ledger (work-items + context) + FTS5/chromem-go (RRF)
retrieval, then the lean read/plan/edit/verify loop. **go.work landmine:**
garrison sits inside scripture-study's go.work, so it carries its own gitignored
`go.work` (`use .`) — nearest-wins makes `go build` work from here; a standalone
clone has none and builds straight off go.mod.

## Update 2 — /goal reached: the P1 loop is closed, the dogfood landed

Michael: "set a /goal to get to the dogfood stage, full ammon loop." No `/goal`
hook exists in this repo (the deadweight "/goal" was an in-session autonomous run;
the only Stop hook is the memory reminder), so I ran it the same way — recorded the
goal, broke it into tested phases G1→G5, and pushed each as a commit without
stopping to check in. Reached the goal in one stretch:

- **G1 ledger** (`internal/ledger`, commit on `cpuchip/garrison`): the `Ledger`
  interface (the council's standalone↔substrate seam) + a pure-Go
  `modernc.org/sqlite` store — work-item hierarchy with a recursive-CTE `Tree`,
  FK-enforced messages, dispatch cost that rolls up the tree. 4 tests.
- **G2 retrieval**: FTS5 keyword index over messages (modernc has FTS5 compiled
  in — confirmed by the vtable creating without error; trigger-synced) +
  `internal/retrieval` with a chromem-go vector store (pure-Go, file-persisted)
  embedded via LM Studio, fused by Reciprocal Rank Fusion. The **live** fusion
  test passed: the vector half found a doc with ZERO keyword overlap, the keyword
  half caught the lexical match, RRF reranked deterministically.
- **G3 loop** (`internal/agent`): Dispatcher interface + LocalDispatcher (records
  cost); a forgiving `===FILE===` edit parser (strips `<think>`, deliberately does
  NOT strip code fences so Go raw-string backticks survive) + path-safe ApplyEdits;
  the Verify oracle; `Loop.Run` (read→plan→edit→verify, every step written to the
  ledger); `garrison run`. Mock-dispatcher tests drive **real** go build/test in a
  temp module — pass-in-1, fail-then-fix-in-2, blocked-after-budget.
- **G5 DOGFOOD** (`docs/dogfood-01.md`): `garrison run` on a greenfield Go module,
  driven by `qwen3.6-27b`, produced package `strutil` (rune-correct `Reverse` +
  `WordCount`) with table-driven tests and passed the oracle in ONE attempt
  (1m29s, 4684 tok, work item #1). Inverse hypothesis: I re-ran `go test -v` by
  hand → 7/7 subtests pass, incl. `"你好世界"` → `"界世好你"`. The harness made a
  27B local model ship correct code first try — Harness > Intelligence, on itself.

**Honest seam:** the dogfood's oracle includes tests the model authored, so the
loop proves the *machinery*; the independent re-run proves the *output*. An
operator-authored-tests oracle is the P2 "code oracle suite" direction (noted in
the doc).

**Carry-forward:** G4 (self-extension Tiers 0–1: skills-as-data + exec tool +
MCP-client stub) is the remaining P1 piece; the loop doesn't yet inject retrieval
context (greenfield needs none — wire it for non-empty trees); `garrison watch`
ships with sub-agent spawning (post-P2/P3). Deps: modernc.org/sqlite v1.52,
philippgille/chromem-go v0.7. All commits pushed to `cpuchip/garrison`; the
`.spec/` + `.mind/` edits are uncommitted root changes (Michael pushes root).

## Update 3 — G4 done, P1 complete, root pushed

Michael: "lets push and finish p1." There is no `/goal` hook here; the work was
to finish the remaining P1 piece (G4, self-extension Tiers 0–1) and push.

- **G4a — skills as data** (`internal/skills`): a skill is a markdown file
  (frontmatter name/description/keywords/always + body). The loop loads a dir,
  selects always-skills + keyword-relevant ones (capped), and injects them into
  the plan + edit prompts. `garrison skills` lists them; two seed skills shipped
  (`test-thoroughly` [always], `go-idioms`). The model can write a new skill file
  and it is live next run — Tier-0, no linking.
- **G4b — gated exec** (`internal/agent/exec.go`): the model may emit
  `===RUN: <cmd>===`; the loop runs it ONLY under `--allow-exec` (off by
  default), records output to the ledger, and feeds it back. Verify stays the
  gate. No shell (argv split) → a small inspectable surface.
- **G4c — MCP-client stub** (`internal/mcp`): newline-delimited JSON-RPC 2.0
  (initialize / tools/list / tools/call) over a stdio transport that is an
  `io.ReadWriteCloser`, so it tests against a fake `net.Pipe` server without a
  process. The seam for the substrate power-up (P4), self-built tools (P5), and
  the plugin (P6).

**Dogfood re-verified with the integrated loop** (skills active): `qwen3.6-27b`
wrote package `mathx` (GCD Euclidean/non-negative + IsPrime with the 6k±1
optimization) and tests, oracle-passed in one attempt; I re-ran `go test` by hand
→ all pass. The skills visibly improved the output (efficient prime check,
thorough edge cases). **P1 is complete:** 9 tested commits on `cpuchip/garrison`,
every package build+vet+test green.

**Root pushed** (Michael's "lets push"): selectively committed only this
session's Garrison records (`.spec/proposals/sovereign-coding-agent.md`, this
journal, `.gitignore`, `.mind/active.md`, the lane) — not the sibling sessions'
in-flight work. No `scripts/becoming/` change, so the ibeco.me prod rebuild the
root push triggers is a no-op for behavior; verified the site stayed up.

**Carry-forward:** P2 = the code oracle suite (operator-authored tests, so a
model can't satisfy a weak oracle of its own making — `verify-quotes`/study-linter
patterns for code). The loop still doesn't inject retrieval into its read-step
(greenfield needs none; wire it for non-empty trees). `garrison watch` + sub-agent
spawning come after P2/P3 per the council.

## Update 4 — P2 + P3 complete, live-verified

Michael: "set a goal to do p2 and p3" (he went to play with the binary). Ran it
autonomously, four more tested commits to `cpuchip/garrison`:

- **P2.1 — operator-authored acceptance tests.** `run --acceptance PATH` copies
  operator tests into the tree and marks them protected; `splitProtected` refuses
  any edit targeting one (the model must *satisfy* the spec, not rewrite it),
  records the refusal, and injects the protected files' contents into the prompt.
  Closes the "model passes its own weak tests" seam.
- **P2.2 — code detectors** (`internal/detect`): gofmt, `go vet`,
  exported-doc-comment, naked-panic — AST-based, precision-tuned, run after the
  build/test oracle passes; findings iterate.
- **P3 — the critic** (`internal/agent/critic.go`): after the oracle passes, a
  second model (gemma-12b — different blind spots than the qwen doer) adversarially
  reviews for what the tests missed. `VERDICT: APPROVE | REVISE`; blocks only on an
  explicit REVISE, bounded by `MaxCritics`. The council's D&C 88:122 lever.

**Live proof (medianx).** The task asked for a median that *must not mutate the
caller's slice* — a property the operator's acceptance test (it passes copies)
cannot catch, only the critic can. Result: the acceptance test stayed
byte-identical (the model's attempt to touch it refused), detectors + auto-format
clean, and the critic APPROVED code that correctly clones-before-sort.
`docs/dogfood-02.md`.

**The inverse hypothesis earned its keep — two real bugs:**
1. **Stale binary.** The first "live" run silently used an old `garrison.exe`
   (`go build ./...` builds to cache; it does NOT rewrite the binary — need
   `-o garrison.exe`). `--acceptance` landed in the task string; no detector/critic
   ran. Lesson: *always rebuild before a live test; "it ran" is not verification*
   (sibling of the deadweight stale-server gotcha).
2. **gofmt blocked correct code.** After the rebuild, the run BLOCKED at 5
   attempts: build + tests (incl. acceptance) passed, but gofmt flagged 2 files
   the model "couldn't" fix. `gofmt -d` → "no newline at end of file" (the
   `===FILE===` parser strips the trailing newline; a weak model can't emit
   byte-perfect gofmt output). **Fix: gofmt is a formatter, not a linter** — the
   loop now runs `gofmt -w` after every edit (`detect.Format`). Re-run → DONE in 2.

**State:** P1–P3 complete; `cpuchip/garrison` fully pushed (13 tested commits).
Next is **P4** — the substrate-backed power-up (point Garrison at pg-ai-stewards
over MCP for a shared ledger / dispatch engine); the `internal/mcp` client is the
seam, waiting. The `.spec/` + `.mind/` P2/P3 records are committed but NOT pushed
(no push instruction this round; root is Michael's to push).
