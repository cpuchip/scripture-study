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

## Update 5 — P3.5 (the presiding chain): built + tested, structure proven live

Michael (asked "what's next?", chose "P3.5 — spawning + garrison watch"). Three
tested commits to `cpuchip/garrison`:

- **P3.5a spawning** (`spawn.go`): refactored `Loop` so the read→plan→edit→verify
  body (`runItem`) runs on a given work item, reused for single runs AND each
  spawned child. `RunWithSpawn` = decompose (planner emits `===TASK===` sub-tasks;
  atomic → no spawn) → a child loop per sub-task under the parent work item
  (sequential, shared tree, `--max-children` gate) → **preside** (every child
  outcome recorded to the parent; a blocked/failed child accounted for, D&C 121) →
  **integrate** (full oracle across the tree, then the presider's own edit-loop to
  finish). The ledger already carried the hierarchy + cost rollup.
- **P3.5b `garrison watch`** (`cmd/garrison/watch.go`): a read-only bubbletea TUI
  that tails the ledger and renders the live presider→child tree (status glyphs,
  polling 1s; WAL lets it read while a run writes). `renderTree` is pure (unit
  tested); `--once` prints once for non-tty. + `ledger.AllWorkItems`.
- Robustness: request timeout 300→600s, configurable (`GARRISON_REQUEST_TIMEOUT_SECONDS`).

**Structure proven LIVE:** a `--spawn` textkit run decomposed into 2 sub-agents
("WordFrequency", "TitleCase"), and `garrison watch --once` rendered the real
parent+2-children presiding tree from the ledger. The decompose → child-spawn →
preside → watch path all executed end-to-end on live data.

**But the full code-gen demo (P3.5c) is BLOCKED by a degraded rig.** Mid-run the
FlexLLama rig went ~20× slow — qwen at ~1.4 tok/s (115s for a 3-word reply), plus
EOFs — so the children's edit dispatches errored (no files written) and the
integrate call hit the timeout. Not a logic bug (unit tests prove the machinery;
the watch tree rendered correctly, even showing the children honestly marked
failed). **Don't hammer a degraded rig** (unattended-resilience lesson) — the full
code-generating spawn demo waits for the rig to recover (Michael may know why it's
loaded). Then NEXT = P4. Honest gotcha re-confirmed: a slow local rig is the
sovereign tool's real failure mode, not the code.

## Update 6 — council mode + Phase 4 BROWNFIELD complete (overnight run)

Michael: vision-expansion council (Garrison as a full Claude-Code/opencode-like
TUI — modes/flow, live stats [context pressure/time/cost], pause&chat, emergency
stop, brownfield+greenfield, tickets, multi-spawn, pg-ai-stewards context tools),
then "set you loose while I sleep." Roadmap reorg: **P4 Brownfield · P5 the TUI
shell (build-together flagship) · P6 context engine · P7 substrate-MCP · P8
tickets/parallel-spawn/multi-lang · P9 Tier-2 WASM/package.** Overnight goal = P4
(brownfield), chosen because it's the foundation the TUI drives AND every piece is
unit-testable with the mock dispatcher → a flaky rig can't waste the night.

Shipped (all rig-independent + green; ~8 tested commits to `cpuchip/garrison`):

- **Council mode** (`garrison council` / `chat` no-arg): interactive multi-turn
  REPL (history kept), `/run` synthesizes the discussion → a task → ratify →
  build, `/model`/`/reset`/`/help`/`/quit`. The "council before the flow" Michael
  asked for; the prompt→build CLI untouched.
- **R5 resilience**: `llm.IsTransient` (5xx/429/408/EOF/timeout/reset) +
  `Loop.dispatch` retries with exp backoff (MaxRetries=3). Built FIRST so the
  overnight run survives the rig's lockups. Every model call routes through it.
- **R1 read CONTENTS** into context (budget-capped, binary/size-skipped, re-read
  each edit iteration so the model sees its own prior edits).
- **R2 surgical edits** (`===EDIT===` SEARCH/REPLACE, exact-match, refuses
  not-found/escapes; protected-checked) alongside whole-file `===FILE===`.
- **R3 relevance-ranked read** (files ordered by task token-overlap so the
  relevant ones make the budget on a big repo; `ReadBudget` tunable). Semantic
  chromem half deferred (rig-dependent; keyword is the high-value primary for
  code).
- **R4 git** (`internal/vcs`, `--commit`): stages excluding `.garrison`,
  diff-stat, commits; off by default.
- **Tickets** (`--ticket FILE`): a ticket md/txt → task.

**LIVE brownfield proof (`docs/dogfood-03.md`):** an existing git repo with a
planted bug (`Add` returns a-b) → Garrison read it, made a 2-line surgical fix
(a+b, doc/package intact), strengthened the tests, passed the oracle, and
committed `4aebfc5` on top of existing history — `.garrison` correctly excluded,
all independently verified. And R5 carried it through a rig at ~1.4 tok/s (2m52s).
The greenfield→brownfield leap, on real models.

**Remaining toward the full coder tool:** P5 TUI (his flagship, build-TOGETHER —
needs his eyes + a healthy rig) · P6 context engine (the "context pressure" stat +
compaction) · P7 substrate-MCP power-up · multi-language oracle (Go-only today;
the riskiest change, held for a supervised pass to avoid regressing Go) · the
semantic half of R3. P3.5c full-spawn live demo still pending a fast rig (machinery
unit-tested + structure proven live). `.spec/`+`.mind/` records committed, NOT
pushed (root is Michael's).

## Update 7 — P5 (the TUI) + P6 (context engine) complete

Michael ("you are a most excellent steward. set a goal and finish p5 and p6, save
p7 for when im up") — overrode my earlier "build P5 together" caution; build it,
he tunes the UX when up. Six more tested commits, all rig-independent:

- **P6 — context engine** (`internal/contextx`): `EstimateTokens` /
  `EstimateMessages` (chars/token heuristic) + a `Pressure` gauge
  (used/window/pct + a text `Bar`) — the "context pressure" stat — and `Compact`,
  which summarizes the older middle of a conversation over a token budget via a
  `Summarizer`, keeping the system message + recent turns. Wired into council so a
  long council auto-compacts and stays responsive.
- **P5.1 — stats** (`agent.Stats` + `StatsFromLedger`): flow (status → council /
  do / verify), time on task, cumulative cost, child count, pressure — computed
  from the ledger, so a watcher reads it without touching the loop.
- **P5.2 — control** (`agent.Control` + `BasicControl`): `Wait` (pause), `Stopped`
  (emergency stop), `Injected` (steer). The loop checks it at each iteration
  boundary: pause holds, **stop cancels and ACCOUNTS in the ledger** (D&C 121),
  inject folds operator guidance into the next step.
- **P5.3 — `garrison drive`**: a bubbletea TUI that DRIVES a run (vs `watch`,
  which observes). A live stats bar (flow · pressure gauge · time · cost ·
  sub-agents), the presiding tree, a log tail, and keys `[space]` pause/resume /
  `[i]` inject / `[s]` emergency-stop / `[q]` quit. The loop runs in a goroutine
  wired to the control + logs to the TUI via a channel. `renderStatsBar` +
  `latestRoot` pure-tested; the interactive shell is built to spec — the one piece
  that wants Michael's eyes + a healthy rig to tune live.

**Garrison is now roughly Claude-Code-shaped:** greenfield + brownfield, sub-agent
spawning, an interactive council, and a driving TUI with modes/stats/pause/inject/
emergency-stop. The session arc (P3.5 → council → Phase 4 brownfield → P5+P6) is
~20 tested commits, every package green; README marks P1–P6.

**Next, with Michael:** P7 (substrate-MCP power-up — the `internal/mcp` client is
the waiting seam), then the multi-language oracle (held all along to avoid
regressing the proven Go path unsupervised). The driving TUI wants a live
shakedown (interactive). `.spec/`+`.mind/` records committed, not pushed.

---

## Update 8 — pg-ai-stewards local-model learnings applied (2026-06-22)

Acted on the inbox signal from the pg-ai-stewards doc-construction soak: *"Garrison
borrows the same `:8090` rig and already builds via edits — apply what transfers."*
The discipline that mattered here was **mapping against the actual loop code, not the
memory of it** — two assumptions were wrong until the code corrected them.

**What the code corrected:**
- The critic *already* reads the working tree from disk (`codeSnapshot` → `ReadTree`),
  so the substrate's "critic reads the artifact, not a passed blob" was already
  satisfied. No work.
- The obvious page-in optimization — "only re-echo the files that changed" — would
  **break** Garrison. Each loop iteration dispatches a single *stateless* user message
  (no growing conversation), so the model has no memory of a prior tree read; dropping
  unchanged files loses them entirely. The full-tree re-read is load-bearing, not
  waste. True page-in needs an on-demand `===READ===` tool loop → surfaced, not built.

**What transferred and shipped (cpuchip/garrison `0e020dd`, pushed):**
1. **Borrow the MoE the rig now serves.** The substrate flipped to the llama-chip
   `dance-moe` preset → the rig serves `qwen3.6-35b-a3b`, not the dense `qwen3.6-27b`
   Garrison still defaulted to. `garrison ping` *proved* it: planner/doer were
   `✗ not served` — a run would 404. Fixed config + `.env.example` + the live `.env`;
   ping now reads `✓`. (The MoE is also ~4× faster — the substrate's measured win.)
   **Inverse hypothesis honored: ping flipped `✗`→`✓` through the real path.**
2. **Kill the one-shot trap.** The system prompt called `===EDIT===` "preferred", but
   the per-iteration instruction and every feedback string commanded "re-output the
   complete files" — overriding the surgical-diff path that R2 already built and
   tested. That *is* the substrate's core doc-construction anomaly (no good agentic
   worker emits a large artifact one-shot). Redirected the prompts to diffs;
   `===FILE===` reserved for new files.
3. **Journal-as-output.** A `===JOURNAL===` convention + `ParseJournal`, captured as a
   distinct `journal`-role note — the presiding provenance trail, the way the digesters
   return a journal instead of the artifact. **Proven e2e on the live MoE:** a Reverse()
   run produced 3 journal notes, one per attempt ("created reverse.go + table test" →
   "added main.go to satisfy build" → "added doc comment to satisfy the detector") — the
   self-correction loop, narrated.
4. **Honest context gauge + rig docs.** `--parallel 2` splits `n_ctx` across slots →
   ~120k per request under `dance-moe`, not 192k. `DefaultWindow` 192k→120k so the
   gauge stops over-reporting headroom. README now documents the borrowed llama-chip
   rig, the per-slot tradeoff, the q8-KV / dedicated-VRAM rule (WDDM spill → ~3×
   slower), and the manual-host restart.

**Surfaced for a supervised pass** (`projects/garrison/docs/local-rig-learnings.md`):
true source-page-in as a gated `===READ===` tool loop + a manifest/outline initial
context, reusing the `internal/mcp` seam. Not urgent (32k read budget sits far inside
120k); the win is large brownfield repos + prompt-processing cost. Written up, not
built — same discipline that holds the multi-language oracle.

**Also:** gofmt'd Garrison's own source — it wasn't passing its own gofmt detector
(dogfooding integrity; pre-existing drift in `loop.go`'s Config block + `drive_test.go`).

A clean cross-pollination arc: one engine's overnight soak teaching its sibling. The
substrate proved these on digesters; Garrison is the same bet (Harness > Intelligence
on weak local models) wearing a coder's clothes, so the lessons fit almost one-to-one.
Root records (`.spec`/`.mind`) committed, not pushed — Michael pushes root.
