## 📬 2026-06-27 (from general-workspace) — teach the playlist-digester to SEE slides (transcript + slide-frames → vision) — IN PROGRESS (Part B build DELEGATED 2026-06-28, PR pending = Michael's Hinge)

**▶ 2026-06-28 — Part B build delegated to a dev opus agent (PR pending).** Grounded the spec + machinery.
**★ Key finding:** Part A (`yt_frames`/`yt_download_video` + `frames.go`/`slides.go`) is in the WORKSPACE
`scripts/yt-mcp/`, NOT the OSS `cmd/yt-mcp/` (which the bridge builds) — so option (a) needs Part A
**ported into `cmd/yt-mcp` first**, then **ffmpeg added to `bridge.Dockerfile`** (`WITH_YT` — it deliberately
has none), then the vision digest stage. Phasing handed to the agent: **B1** port Part A + ffmpeg + prove
`yt_frames` in-container on a real slide-heavy video · **B2** frames → substrate vision mechanism (reuse 47
content_parts + 49 chat_attachments-as-images), aligned frames.json×cues.json · **B3** the "read slides"
digest stage in `examples/playlist-digester.sql` (page frames in by handle). Oracle = virgin-smoke (core
chain) + the e2e "it SEES the slide" proof (vision extracts slide content the transcript MISSED). PR = his
Hinge (new capability = dominion_in_council). Spec: `.spec/proposals/yt-slide-frames.md`.

**★ DOGFOOD FOLLOW-UP (from Michael, 2026-06-28) — once Part B is built, have the substrate review a
video against itself.** Michael had general-workspace review **"The Agentic OS Setup That Will 10x
Claude Code"** (Chase AI, `youtu.be/HRw-vP0j8OM`, 31 min) against pg-ai-stewards — workspace review at
`study/yt/agentic-os-10x-claude-code-chase-ai.md`. He wants **the substrate to do the same review once
it can SEE slides** (Part B). It's the natural acceptance test: point the playlist/video digester at that
video, have it pull transcript + slides through the new vision stage, and produce its own "review against
pg-ai-stewards" — then compare to the workspace version. The video is literally *about* building an
agentic OS, so the substrate reviewing it against itself is a clean, meaningful dogfood. (Workspace
review's TL;DR: his AIOS is the same vision one tier down — file-system-backed, single-operator, with the
DB / multi-agent / verification / governance layers as hand-waves; two borrowable ideas = **session-mining
as workflow-audit** and **a cheap index-map tier in front of embedding retrieval**.)

✅ **yt-dlp version gotcha (RESOLVED here; matters for Part B's bridge image):** the video download first
FAILED — yt-dlp 2026.03.13 hit **"n challenge solving failed"** (YouTube anti-bot). The fix was simply
**updating yt-dlp (→ 2026.06.09)** — it then downloaded cleanly, *no JS-runtime / deno needed*; the stale
version was the whole problem. **So pin a RECENT yt-dlp in `bridge.Dockerfile` and rebuild it periodically**
— YouTube changes the challenge, so a months-old yt-dlp will silently fail on current videos. A
progressive-format fallback (itag 18) is a reasonable deeper safety net. The workspace review now has slides
(his "V.A.U.L.T." Jarvis-HUD dashboard + the `raw→wiki` Obsidian vault diagram — both confirm the
file-system-backed, thin-UI characterization).

— original signal below.

## 📬 2026-06-27 (from general-workspace) — [original] teach the playlist-digester to SEE slides — SPEC'D, Part B is yours

**Michael's ask:** the slide-heavy talks (the Google Cloud agentic-DB playlist being digested now) lose
their densest content — diagrams, benchmark tables, the SQL on the slide — in a caption-only digest. Fix:
the **rich-docs multimodal pattern you already have** (P1–P4: text + page-pixels → `gemma-vision` via
`--mmproj`) applied to **video**: transcript + slide-frames → vision.

**Spec (both halves):** `.spec/proposals/yt-slide-frames.md`. Part A (the workspace yt-MCP gets
`yt_download_video` + `yt_frames`) is general-workspace's build. **Part B is yours:**

- **The playlist/video digester should read the slides, not just the captions.** Add a "see slides"
  step that feeds the vision model **the slide frames + transcript aligned by timestamp** (`frames.json`
  × `cues.json` → "this slide + what was said over it"). Build via your existing doc-construction tool
  loop, paging frames in by handle (don't echo all frames at once — same page-in discipline as rich-docs).
- **Getting frames into the sandbox** — your call: (a) the `WITH_YT` bridge runs the shared recipe
  in-container (add ffmpeg to the bridge image; reuse its yt-dlp) → `frames/` + `frames.json` beside the
  transcript; or (b) dial the workspace yt-MCP's `yt_frames` over MCP. (a) keeps it self-contained.
- **The shared recipe** (capped video → ffmpeg scene-change frames → timestamp-aligned manifest) is in the
  spec, defined once so both Part A and Part B use the same thing. Reuse your doc-extract / page-pixels /
  vision-dispatch machinery — the only new bits are the *source* (video → frames) and the *alignment*
  (frame ↔ cue).

— filed by general-workspace. Spec ready; Part B is the substrate-digester multimodal upgrade. Not blocking.

---

## 📬 2026-06-27 (from general-workspace) — BINEVAL ("Ask, Don't Judge"): binary-question judges + a *guarded* self-improvement loop — directly upgrades `56`/`59`/`74` — OPEN (Michael reviewed + likes it)

**Michael handed us a paper and it lands squarely on your live work** — the trajectory-critic (`56`),
the gated self-improvement loop (`59`), and the north-star (`74`). Paper + full review + the
limitations: `books/papers/NOTES-bineval-ask-dont-judge.md` (PDF/text beside it). arXiv:2606.27226.

**The idea (and what Michael liked):** instead of a holistic `REVIEW: passes / REVISE`, the judge
answers N atomic **yes/no questions per dimension, each with a note** — a verdict that's *checkable*
(binary) **and** *actionable* (the note says why, and the note is what feeds refinement). Decompose
the criteria → answer independently → aggregate to calibrated scores. Training-free, task-agnostic;
beats UniEval/G-Eval, **strongest on factual consistency**, and **helps the *weaker* model more**
(small yes/no Qs are easier than a Likert score → good for a weak-local-model substrate).

**Concrete applications:**
- **`56` trajectory-critic → decompose the verdict.** Yes/no-with-notes per dimension instead of a
  holistic pass/revise. You see *which* criterion failed, debuggable maturity gates.
- **`59` self-improvement → BINEVAL is a concrete algorithm for it.** Failing questions → a note-taker
  extracts lessons → dedup → rewrite the *specific* prompt substring → early-stop. **Cross-model
  update** maps onto `model_capability` routing: a stronger reference model teaches the weaker local
  model's prompt via their per-question *disagreements* (improve qwen/gemma without fine-tuning).
- **`74` north-star → make it load-bearing, not a sticker.** The north-star *directions* ("welfare of
  the soul over the metric," "point to the source," "preside-don't-compel") become binary questions
  the critic asks each turn ("did this turn point to the source? y/n + why"). Exactly the
  "load-bearing not wallpaper" concern from the north-star design, with a method attached.

**★ Guardrails for `59` — the must-haves, straight from the paper's honest limitations:**
1. **Cap iterations + select-best-on-validation.** The loop *degrades* after 1–2 rounds — accumulated
   lessons become competing instructions ("instruction overload"). Early-stop is non-optional.
2. **Promptable-vs-capability triage.** On a real capability/computational limit (counting, ratios),
   *don't refine the prompt* — it's "unactionable and degrades performance"; escalate to a stronger
   model (m2 capability-substitution) or a deterministic tool. **Diagnose ≠ repair.**
3. **Don't over-decompose** broad/semantic criteria (relevance/taste) — over-atomizing makes the judge
   *more severe than humans*, not better aligned. Keep those lighter.

— filed by general-workspace. **Build is yours** (your stewardship over `56`/`59`/`74`); we reviewed +
captured. Not blocking; a strong enhancement when you next touch the critic or self-improvement loop.

---

## 📬 2026-06-25 (from general-workspace) — pg-ai-stewards = "the instantiation of the book"; embed a GENERALIZED north-star (Intent / step 1) on every call — ✅ RESOLVED 2026-06-27 (BUILT)

**✅ Resolved 2026-06-27 — both halves shipped.** (1) The north-star: `74-north-star.sql` (PR #13)
— generic-default-in-core + operator-owned config (`north_star.why/source/directions`) +
`render_north_star()` + re-authored `compose_system_prompt` (prepend FIRST, echo LAST). Load-bearing:
the directions re-root the substrate's existing covenant behaviors so the why is the tie-breaker.
Every-call confirmed (no scripture hardcoded in core; `ON CONFLICT DO NOTHING`; empty why opts out).
virgin-smoke OK 65a/65b, fresh-image 00→74 GREEN. Michael's Col 3:17 + 2 Ne 32:9/25:26/D&C-121
directions live in the overlay (`pg-ai-stewards-workspace/overlays/north-star.sql`, applied + clobber-
check green). (2) README cross-link to `cpuchip/scripture-book` ("the substrate is the book's pattern
instantiated; the North Star is step one") — done in the same PR. The original signal is preserved
below for reference.

---

## 📬 [original, kept for reference] 2026-06-25 (from general-workspace) — embed a GENERALIZED north-star (Intent / step 1) on every call

**Context:** in a live study session Michael landed two linked moves. The book *Beyond the Prompt*
(`github.com/cpuchip/scripture-book`) is the **why** — the eleven-step creation pattern. **pg-ai-stewards
is that book *instantiated*.** The book already points here (Practice 9: *"eventually, a whole substrate"*).

**1. README cross-link.** pg-ai-stewards' README should link to `github.com/cpuchip/scripture-book`,
framed as *the substrate is the instantiation of the book's pattern.* (Back-links — the book's README and
`cpuchip.net/teaching` pointing here — are those repos' own tasks; flagged to Michael separately.)

**2. The north star — give the substrate its Intent (step 1 of its own cycle), on every request.**
Diagnosis from the study: the substrate runs steps 2–11 of the book's eleven-step cycle (covenant,
stewardship, spec, watching, atonement…) but **step 1 — Intent, the named *why* — was never made explicit.**
Fix: a guiding "north star" carried in the **core prompt of every LLM call.**

Design (Michael's, settled):
- **Generalize it for the OSS core.** Do NOT hardcode a scripture in the public Apache-2.0 core (that fights
  the `values_anchor` genericization). The core **asks the operator installing it to provide their OWN north
  star** — a guiding *why* + the directions it governs. **★ This is the point, not a compromise:** the
  mechanism *enacts the doctrine* — every steward must name their Intent; the **form is universal, the content
  is the operator's.** Agency-within-bounds: you must orient, you choose how (persuade-don't-compel, applied to
  the tool's own users).
- **Recommend scriptures** for those who share the faith. Michael's pick (and his overlay's north star):
  **Colossians 3:17** — *"whatsoever ye do in word or deed, do all in the name of the Lord Jesus, giving thanks
  to God and the Father by him."* Short (low prompt cost), names Christ, *"in word or deed"* = every call.
  Alternates to recommend: **2 Nephi 25:26** (point to the source) and **2 Nephi 32:9** (consecrate every
  performance, for the welfare of the soul).
- **Michael's instance = the OVERLAY.** His Christ north star (Col 3:17) lives in his private overlay's
  core-prompt layer — his consecration of the generic engine (Abraham 4, *"in our image"*).
- **★ Load-bearing, not a sticker.** A verse pasted on every prompt that changes nothing becomes wallpaper the
  model ignores — the Christ-*patterned*-not-*centered* trap in miniature. The **directions** added alongside
  the verse should **re-root the substrate's EXISTING covenant behaviors** under the chosen why: serve the
  welfare of the soul over the metric · point to the source (no honors-of-men) · preside-don't-compel ·
  read-before-quoting / assume-you-can-be-wrong. Not new behaviors — the *why* named under the behaviors already
  there, so the north star becomes the **tie-breaker** when values conflict.

Open design questions (your stewardship to settle):
- **Required-at-install vs. neutral default?** Require the operator to name a why before boot (maximally enacts
  "Intent first") vs. ship a generic-but-real default + recommend they set their own (lower friction). Michael
  leans toward it being present on every request; Col 3:17 is short enough.
- **Every-call vs. decision-points** — Michael leans every-call; confirm it doesn't bloat utilitarian sub-calls.
- Exact set of existing behaviors the directions bind, and where the field lives (overlay core-prompt layer).

Full study context: `study/ai/harness/provenance.md`. The book's first ~10 pages (frontmatter + 9 practices +
the eleven-step cycle) are the source for the "instantiation" framing and the Intent=step-1 diagnosis.

— filed by general-workspace. **Build is yours** (stewardship); we kept this in explore/learn. Not blocking.

---

## 📬 2026-06-25 (from general-workspace) — Loreworks demo idea: make BOYD the first world; *Patterns of Conflict* = a knowledge factory; what DeepLore has to do with it — OPEN (Michael's late-night spark)

**Michael's spark (~2am, capture-before-forget):** hold the **6-minute Loreworks walkthrough**
(chunk F) + the **knowledge-factory** framing + **John Boyd's *Patterns of Conflict*** + **DeepLore**
in view at once — he sees them as the same thing. They are.

- **Patterns of Conflict *is* a knowledge factory.** Boyd ingested a vast corpus (military history,
  Sun Tzu, blitzkrieg, guerrilla war, Gödel / Heisenberg / 2nd-law) and **synthesized the recurring
  patterns + their relationships** into a navigable model of conflict. His *Destruction and Creation*
  (analyze → synthesize) is literally the Loreworks **world-build agent's** method: break the canon
  into entities, recombine into a typed graph. The world you build = **orientation** (Boyd's big "O").
- **So the killer first demo — and the 6-min video — could be Boyd himself.** Drop the Boyd corpus
  (`books/johnboyd/patterns-of-conflict/`: the Hammond *Discourse* + the clean POC slides) into
  Loreworks → extract the **orientation-graph of a real thinker's mind**: OODA · Orient · maneuver ·
  Auftragstaktik · Schwerpunkt · Fingerspitzengefühl · blitzkrieg vs guerrilla · the Gödel/Heisenberg/
  entropy roots — entities + typed edges + a **loremaster persona** you can interrogate. A *non-fiction*
  world (a thinker's thought) is a more striking proof than a fantasy world, and it ties the demo to
  the **harness-is-orientation** thesis (study scaffolded at `study/ai/harness/`).
- **What DeepLore has to do with it:** DeepLore is the structural proof-of-concept on the lore side
  (Obsidian vault + two-stage keyword→AI-select retrieval + **Emma the librarian** + relationship
  graph ≈ `world_entities`/`world_edges` + hybrid `embed_query` RRF + the world-build agent + `world_graph`).
  Its loop — *"your story fills in your world; your world fires back into your story"* — is a continuous
  **OODA loop on the world model**: the librarian flags a gap (orientation incomplete) → you author the
  entry (re-orient) → the world fires back. That's Boyd's mandate that orientation must continuously
  re-synthesize or it decays. **Four DeepLore transfers worth stealing into Loreworks:** `summary`-as-
  *when-to-select* retrieval hint · contextual gating (era/scene/character) · the **grow-during-play**
  gap-flag loop · a "why did this fire?" provenance trace.

**Thesis under all of it:** Loreworks isn't just "build a fictional world" — it's a general
**orientation-synthesis engine** (a knowledge factory), and Boyd is both its best demo subject AND the
strategist who explains *why* it matters. Sources + full synthesis: `study/ai/harness/provenance.md`,
`external_context/google-new-sdlc/NOTES.md`, `external_context/sillytavern-DeepLore/`.

— filed by general-workspace. A demo/positioning idea for Loreworks (chunk F + the loremaster, chunk C);
not blocking. Review when convenient.

---

<!-- ✅ RESOLVED + cleared 2026-06-25: the trajectory-eval gap is BUILT. The Glass-Box
trajectory critic shipped (`56-trajectory-critic.sql` — assemble_trajectory + the
trajectory-critic judge over the dispatch trace) AND grew into the gated self-improvement
loop (`59-self-improvement.sql`). Exactly the Day-4 "judge OVER the trajectory" this flagged.
A2A + Day-5 spec-driven + Day-3 context-engineering threads remain as future reading (durable
in external_context/google-new-sdlc/NOTES.md), not blocking. -->
## 📬 2026-06-25 (from general-workspace) — Google/Kaggle "Vibe Coding" whitepaper series ≈ the substrate; trajectory-eval gap — RESOLVED (the critic shipped)

**Michael flagged this as very pertinent to substrate work.** I gathered all 5 days of Google
& Kaggle's June-2026 "5-Day AI Agents: Intensive Vibe Coding Course" whitepapers (+2 Nov-2025
foundational) into `external_context/google-new-sdlc/` — full read + cross-project synthesis in
that folder's **`NOTES.md`**. The short version for you:

**The paper literally describes pg-ai-stewards as "the substrate."** Day 1 ("The New SDLC With
Vibe Coding") says, verbatim: *"For agents that serve real users at scale, the agent is the
product, and it needs the substrate underneath"* and *"Invest in the production substrate before
scale… build this substrate before the first production agent ships, not after."* Their substrate
checklist — persistent memory across sessions, scoped per-agent permissions, eval coverage,
observability/traces, MCP — is the substrate's feature list. Independent convergence, same noun.

**★ The real gap it surfaced — trajectory evaluation (Day 4, "Agent Quality").** Day 4 splits
eval into **output/final-response ("Black Box")** vs **trajectory/process ("Glass Box")** — the
latter assessing *every step of the execution trajectory*: LLM planning, **tool usage**
(wrong/missing/hallucinated tool or params), **tool-response interpretation** (e.g. *not
recognizing a 404 error state and proceeding as if it succeeded*), **RAG quality**, **trajectory
efficiency/robustness** (excess calls, loops, unhandled exceptions), and **multi-agent dynamics**
(inter-agent comms, role adherence). Day 1's sharp line: *"a fluent output that skipped its
verification steps is a more dangerous failure than one with a visible error."*

Where the substrate stands (I grepped the OSS extension — no trajectory evaluator found): you
**already capture** the full trajectory (every `tool_dispatch`/`work_queue` row, persona/sub-agent
delegation paths), but the judges + verify gates score **final outputs** ("REVIEW: passes",
maturity). The data is in the ledger; **a judge OVER the trajectory doesn't exist yet.** That's a
high-leverage, near-zero-new-data add: a **"trajectory critic"** stage/judge that reads the
dispatch trace against a rubric (right tools chosen, error states recognized, no redundant loops,
persona role adherence). Michael explicitly asked to flag this — *"trajectory eval for the
substrate would also be good too if we're not doing it."* (Garrison's getting the same — banked.)

**Two more threads (detail in NOTES.md):**
- **A2A (Agent2Agent)** — Day 1/Day 2 name MCP **and** A2A as "the connective tissue." You speak
  MCP everywhere; A2A is the one open standard not yet adopted — worth a look as persona/sub-agent
  coordination grows.
- **Day 5 "Spec-Driven Production"** — *"code is now disposable; a rock-solid BDD/Gherkin spec can
  regenerate the entire codebase repeatedly."* Directly relevant to the spec-driven pipelines
  (planning, doc-construction). Day 3 (Context Engineering: Sessions/Skills/Memory) maps onto
  sessions/engrams/memory. Both gathered, not yet deep-read.

— filed by general-workspace. Papers + NOTES.md are durable in `external_context/google-new-sdlc/`.
Review when convenient; the trajectory-critic is the one concrete build to consider.

---

## 📬 2026-06-16 (from general-workspace) — proposal: let the digester pipelines READ our repos — OPEN (needs council)

**Michael's ask:** give the ai/book/video digester pipelines the ability to *read the
work we're doing here* — a container with our repos checked out — so a digester can
compare what *it* produced against *our* studies and surface what to learn / incorporate.

**Motivation on disk:** the playlist digester digested the Euclid video the same week the
general lane wrote a human study of the *same* video — neither knows the other exists. A
"cross-reference our corpus" stage turns the digesters' §6 ("what could we do with this")
into "here's how this compares to what we've done, and what's worth folding in."

**~90% there:** the substrate ships read-only fs-read; the gap is making our repos visible
to the digester container. (a) read-only bind-mount scripture-study / scripture-book /
pg-ai-stewards-**oss** (NOT the private substrate repo with keys); or (b) a git-clone step
like code-pr. New tools-on read-only "cross-reference our corpus" stage. Caveats:
read-only always; mind secrets; gitignored content (gospel-library, /books, /yt) won't be
in a clean clone. **New standing capability → dominion_in_council: ratify before building.**
Pairs with book-digester.md §6 + study-pipeline.md. **Adjacent to the digester-steward
(curator) — a presiding curator that can read our corpus could pick books/videos that
fill gaps in what we've already studied.**

— filed by general-workspace; NOT yet acted — the next council item when Michael wants it.

<!-- RESOLVED + cleared 2026-06-24: the doc-extract sandbox shipped (P3e/f). The
"cross-reference our corpus" need is met by `doc_import_corpus` — zip a repo (a
repo is a folder), import it into the searchable docs pool tagged by project,
then doc_search it; all through the same hardened no-network extract lane. A
live read-only `git clone` into the sandbox (no zip step) is the noted future
enhancement (docs/rich-documents.md §deferred). See
.spec/journal/2026-06-24-rich-docs-p3-doc-extract.md. -->

<!-- cleared 2026-06-16: storytelling-craft-digest (done) + stuck-research-write diagnosis (done) -->
