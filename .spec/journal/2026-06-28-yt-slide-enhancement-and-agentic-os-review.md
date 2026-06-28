# 2026-06-27/28 — yt-MCP slide enhancement (Part A shipped) + the Agentic-OS review

**Lane:** general-workspace. **Arc:** spec → build → test → dogfood, all in one sitting.

## What was done

**Built the yt-MCP slide enhancement (Part A) — `scripts/yt-mcp/`, three commits live (not pushed):**
- `yt_download_video` — fetch the mp4 (720p cap, gitignored, never auto-called). Carries an
  `--ffmpeg-location` override so a *stale yt-dlp config* (Michael's pointed at a deleted ffmpeg dir)
  can't break the merge.
- `yt_frames` — ffmpeg slide extraction → `frames/*.png` + a timestamp-aligned `frames.json`. Three
  modes: `scene` (scene-change, one frame per slide), `interval`, `timestamps`. Returns the manifest,
  not the images; over-cap is evenly sampled (`frames.go`).
- `yt_slides` — the one-shot to *study* a slide talk: ensures transcript + video, **auto-picks the
  capture strategy (chapters → scene → interval)**, aligns each slide to the narration spoken over it
  (`frames.json` × `cues.json`, windowed to the next slide), and writes a readable `slides.md`
  (`slides.go`). 93b734ef + 4c3f6e7f.
- Spec covering both halves: `.spec/proposals/yt-slide-frames.md`. **Part B** (the substrate digester's
  multimodal "see the slides" upgrade) is pg-ai-stewards' — delegated, PR pending (their inbox).

**Tested live on the Cole Medin "New SDLC masterclass" (the video behind our harness study) — and it
genuinely enriched understanding.** The slides carried what the transcript flattens: the whitepaper's
**Figure 5 names "Trajectory Eval"** ("verify what it built AND how it got there") — independent
validation of the substrate trajectory-critic signal — plus the hard numbers (52.8%→66.5%, +13.7 pts
Terminal-Bench 2.0), Table 1's vibe→agentic spectrum, the conductor↔orchestrator framing, the CapEx/OpEx
crossover. Seeing the deck deepened my read of the source we've been building on.

**Reviewed "The Agentic OS Setup That Will 10x Claude Code" (Chase AI) against pg-ai-stewards** →
`study/yt/agentic-os-10x-claude-code-chase-ai.md`. **Verdict: it's our vision one tier down.** His
four-level AIOS maps almost 1:1 onto substrate components (skills+loop→personas+pipelines+`59`;
second-brain/`index.md`→engrams+RRF; dashboard+`claude -p`→Stewdio+the Hinge reviewer; distribution→
multi-tenancy) — but DB, retrieval, multi-agent dispatch, verification, and governance are *hand-waves*
where ours are the engineered core. His thesis ("the value is under the hood, not the dashboard")
validates our philosophy from outside; his V.A.U.L.T. Jarvis-HUD is the visual proof (gorgeous skin over
skills fired by headless `claude -p`). **Two borrowable ideas:** session-mining as a workflow-audit (mine
session history to propose pipelines/personas) and a cheap `index.md`-style map tier in front of embedding
retrieval. Left a dogfood note in the pg-ai-stewards inbox: after Part B, have the substrate review this
same video against itself.

**Researched + recorded: Claude Code has no MCP hot-reload.** `/mcp` reconnects but doesn't re-run
`tools/list`; editing `.mcp.json` mid-session isn't picked up; rebuilding a local MCP binary needs a full
**restart** (observed: a refresh picked up the new yt tools). Windows locks the running `.exe`, so I
rename-swapped the new binary into place (build to a new name → rename running exe out → move new in).
Memory: `reference_claude_code_mcp_no_hot_reload`.

## Surprises / lessons worth keeping

- **Scene-change under-fires on smooth-scroll screen-shares** (Excalidraw canvases, scrolling a page).
  It only caught 5 frames in 22 min of Cole Medin. Chapter markers are the best signal when present;
  interval is the fallback. That tuning finding *became* `yt_slides`'s auto-strategy.
- **A months-old yt-dlp silently fails video downloads** — "n challenge solving failed" (YouTube's
  anti-bot). Michael's instinct was exactly right: `pip install -U yt-dlp` (2026.03.13 → 2026.06.09)
  fixed it cleanly, **no deno / JS-runtime needed** — the stale version was the whole problem. Folded
  into the Part B note (the bridge must pin a *recent* yt-dlp and rebuild periodically) and a yt-MCP
  README prereq note. This will bite again; updating is the answer.
- **Read-before-quoting earned its keep again.** The transcript-digest subagent returned "verbatim"
  quotes that were actually *cleaned* — "90% of the value," "index MD," "99% of the way" don't appear
  literally in the auto-transcript. Verified against source, found the drift, and **paraphrased** the
  review instead of printing false-precision quotes. Auto-transcripts are themselves imperfect records,
  so quoting them verbatim is fraught regardless.
- **The whole thing was a clean dogfood:** built the slide tools, then used them to review a video *about
  building an agentic OS* against our agentic substrate. The tool, the review, and the subject rhymed.

## Relational

Easy, collaborative debugging. Michael caught the yt-dlp-update fix before I'd have escalated to deno —
his "i think it has an update command too" was the answer. The session moved spec → build → test →
review → fix without friction.

## Carry-forward

- **Part B (substrate slide-seeing digester)** — pg-ai-stewards' build (delegated, PR pending = his
  Hinge). The **dogfood** (substrate reviews the Chase AI video against itself) is queued behind it.
- **`yt_slides` goes live on the next Claude Code restart** (binary already swapped in; no MCP hot-reload).
- **Comments integration** — Michael wants it, but **security-gated**: prompt-injection (comments are
  untrusted input) is the load-bearing concern, plus toxicity/PII/spam. Deferred signal in the
  general-workspace inbox.
- **Transactional email/SMS notify service** — still parked (he'll want it for ibeco.me password-resets).
- Low-pri yt-MCP follow-up: bake a yt-dlp self-update / version-check or a progressive-format fallback
  (less urgent now that updating fixed it).

## Set down

- yt-MCP Part A is **done** — spec'd, built, live-verified (scene + interval + chapters all proven),
  documented, committed. Off the board.
- The "needs deno" theory for the n-challenge — **wrong, released.** It was just a stale yt-dlp.

---

## Addendum — Open Knowledge Format (OKF) research (same day)

Michael sent one more video (AI LABS, "Google's New Release Just Fixed AI Systems") and asked what
Google's **Open Knowledge Format** offers pg-ai-stewards. Researched against the primary spec, not the
video (the video was accurate; its "OKP" was a transcription slip): **OKF v0.1** (Google Cloud Data Cloud
team, 2026-06-12, repo `GoogleCloudPlatform/knowledge-catalog`) = knowledge as **a directory of markdown
files + YAML frontmatter** — one required field (`type`), a markdown-link concept graph, `index.md` for
progressive disclosure, `log.md` for history. *"If you can `cat` a file you can read OKF; if you can
`git clone` a repo you can ship it."* Positioned like MCP/skills — a standard Google expects every agent
to adopt.

**Finding (same shape as the Chase AI review): complementary, not competing.** OKF is knowledge AT REST
(portable interchange); pg-ai-stewards is knowledge IN MOTION (the live RRF semantic engine, which beats
OKF navigation on the query path). The use is to have the substrate **speak OKF at its edges, like it
already speaks MCP** — an `okf_export(intent)` / `okf_import(bundle)` boundary adapter (export turns a
Zion pool into a portable, git-versionable, any-agent-readable artifact; import ingests partner bundles).
Aligning the digesters' doc output to OKF frontmatter is good chunking hygiene and **reinforces the
"cheap index-map tier" idea from the Chase AI review — now flagged by two independent sources.**

**Honest caveat:** v0.1 is three weeks old, adoption unproven — "more an optimization than something you
need until it's a built-in standard." So it's a low-cost boundary adapter to keep on the shelf, built when
a sharing/ingest need is real — NOT a core change, the engine untouched. Writeup
`study/yt/open-knowledge-format-okf-for-pg-ai-stewards.md`; filed to the pg-ai-stewards inbox as a
council/stewardship candidate (a new standing capability = `dominion_in_council`).
