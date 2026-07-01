---
date: 2026-06-29
lane: pg-ai-stewards
topic: the substrate saw a real slide end-to-end, I caught myself trusting a proxy and we made a rail of it, drafted the uplift-the-local-models arc, and gave the public repo a clean history
tags: [yt-slide-frames, verify-real-path, reflection, uplifting-local-models, rest, watcher, elastic-rig, default-profile, history-rewrite, filter-repo, scope-breath, marginalia-cadence]
---

# The real-path rail, the uplift arc, and a clean history

A long, multi-thread session that started as "merge the yt PR and test it" and became: an honest
e2e, a rail named after my own mistake, a design arc for uplifting the local models, and a clean
rewrite of the public repo's history. The connective tissue is the lesson — *verify on the real
path, don't trust a proxy you built.*

## What landed

- **yt slide-frames Part B — MERGED (PR #16, main `6b54dd1`) + the FULL real e2e run.** Brought the dev
  stack onto the yt overlay (rebuilt the WITH_YT bridge, recreated pg+bridge with the `/yt` mount —
  data-safe, pgdata persists) and ran the *whole* pipeline on a real AI-playlist video
  (`q8aZn-2FbHg`, "Did Google steal my research?"): `yt_frames` (real tool → 10 scene frames to `/yt`)
  → **`import_yt_frames`** (the `/yt`-volume `pg_read_file` the agent never ran — *works*, 10 captioned
  attachments) → `slides_read` agent loop → uncapped gemma vision. It read what the narration never
  said: arXiv `2603.18521`, "Interpretable Context Methodology", author **Sam McVety**, "Open Knowledge
  Format", **Isaiah 28:10** on a whiteboard (*line upon line* — the very pattern the substrate runs on),
  the folder tree. All carry-forward gaps closed.

- **The rail — `verify_real_path` (covenant) + `feedback_real_path_or_flag` (memory).** The night before I
  ran a *synthetic* vision call with my own `max_tokens=1500`, watched it return empty, and reported it
  as a "tuning finding" with a recommended fix — for a problem that existed only in my harness. Michael
  caught it in one line ("I don't think we have max token set there"). Confirmed in `bgworker.rs`: the
  `4096` default is `anthropic_body_from_openai` only; the local OpenAI-format rig never hits it → uncapped
  → clean content. The *same* bad assumption nearly drove a needless rig-config edit minutes later until I
  tested the real path (dance-moe sees fine). Ratified into the covenant; memory links it to
  `verify_via_real_path` (the SSE false-negative) + `build_the_oracle_first`.

- **The reflection (Michael asked how we're working).** Grounded scope estimate: ~24 projects, ~150–180K
  LOC, ~285 tracked tasks in 7 weeks ≈ **10–20 engineer-years of conventional-pace output** (a 5–8 eng
  team for 1.5–3 yrs), ~75–150× compression. Honest caveats: it's *output* not *maturity*; the real
  ceiling is Michael's review hours. What's good: the delegation shape + the Hinge + oracle-first + the
  covenant/memory. Where we struggle: I build a proxy and trust it (the root of the recent missteps).
  Rails: real-path-or-flag (#1), worktree-isolation for bg git agents, a periodic scope-breath.

- **Marginalia session-close cadence** added (CLAUDE.md addendum + `project_marginalia` memory): at session
  close, *consider* a margin post. First application below.

- **Scope-breath pass** (applied): closed #284 (tool-shelf, PR #15 shipped), #136 (CT2.4 answered), #251
  (folded into #266); **reframed #143 → the shadow-watcher** (Michael: models don't fold their own tools;
  a 2nd agent shadows + auto-folds → learn a heuristic); parked #150 (awaiting his first campaign);
  expanded #177 (claude -p harness, *free via Max*, pull into the substrate via Stewdio).

- **The rig.** Verified **dance-moe IS vision-capable** (the 500 was a load-timing transient, not a gap —
  caught by testing the real path). Added a **`default_profile`** feature to llama-chip (config field +
  boot-apply; committed `a1b37be`) so a restart boots into dance-moe instead of empty. Then, on Michael's
  ask, **shut the rig down cleanly** (graceful `/api/unload-all` → both 4090s to 0 MiB → stop the process)
  so general-workspace could rebuild it. (They did — custom-backend E1/E2 + DiffusionGemma running headless.)

- **★ The uplift arc — `.spec/proposals/uplifting-local-models.md` (DRAFT, for council).** Michael's real
  goal under three threads (model/provider mgmt + the shadow-watcher + the "rest"): **stop qwen spiraling
  to dead on real tasks.** Three composable mechanisms: (1) an **elastic rig** — a llama-chip-aware provider
  that drives `/api/ensure`/`/api/profile` to scale models per phase (the endpoints already exist *for this*;
  the tension is human-vs-substrate rig contention — we lived it this session — needs a lease/mode); (2) a
  **watcher** — a cheap big-context model (Nemotron, 1M ctx) shadowing live sessions, folding aggressively
  *because reveal is reversible*; (3) the **rest** — Michael's sharpest line: a turn where the only tools
  left are the housekeeping ones (context/tool/skill/journal/note mgmt), so the model can't dodge tidying.
  The frame Michael blessed: *the substrate practicing its own creation cycle — work, rest, tidy, continue.*
  First step if ratified = **build the spiral oracle** (measure qwen's failure rate) before the cure.

- **Planning-surface survey + the most-pressing read.** Open for the substrate: multi-tenancy/RLS
  (proposed, council — the one-way door), uplifting-local-models (new draft), page-in-large-results (DRAFT,
  dominion_in_council); inbox council-candidates: BINEVAL, OKF adapter, digesters-read-repos, Boyd-as-world.
  **Most pressing = local-model reliability** (everything autonomous already rides on qwen not dying);
  cheapest move = the spiral oracle. **Runner-up = multi-tenancy** (gets more expensive every migration).

- **Public history cleaned.** Michael disliked the `Co-Authored-By` / `Claude-Session` commit trailers. Set
  `includeCoAuthoredBy: false` (user settings → all future commits). Then for **just pg-ai-stewards (the
  public repo)**: `git-filter-repo` stripped the trailers from all 294 commits in ~2s, **tree hash
  identical** (`e0cd8e9` → content provably byte-for-byte, only messages changed), force-pushed
  `6b54dd1 → 0b57739`, and deleted three stale shipped-feature branches. `origin/main` now: 0 trailers,
  one branch.

## Lessons worth keeping

- **`git filter-branch` is unusable on Windows** (a process per commit — it timed out at 3 min on 294
  commits). **`git-filter-repo`** (one `pip install`) did the same rewrite in ~2s, single pass. And the safe
  verify for a message-only rewrite is **tree-hash equality**: if `HEAD^{tree}` is unchanged, content is
  provably intact — confirm *before* the force-push (origin holds the old history until you push).
- **dance-moe reasons AND sees** — its gemma auto-loads the co-located mmproj; the earlier 500 was timing.
  So the slide-digest "operational note" I flagged on the e2e was a non-issue. (Real-path again.)

## Carry-forwards

- **The uplift arc — Michael wants to wrestle the spec before we build.** Open for council: the lease model
  (Stage 1), watcher-as-new-kind vs grow-the-reactive-engine, the rest cadence, and the load-bearing doubt
  (does a restricted-toolset rest make a *weak* model tidy well? — needs a REAL probe). **Safe first move =
  the spiral oracle.**
- **Multi-tenancy go/no-go council** (one-way door, hardens with each migration).
- **The rig is DOWN; general-workspace rebuilt llama-chip.** When substrate work resumes, reload dance-moe
  (now the boot default) — coordinate so we don't clobber their coding-model experiments. Substrate is idle
  on the yt overlay until then.
- **`pg-ai-stewards-oss` main was rewritten + force-pushed (`0b57739`).** The local tree is on it; any other
  clone needs a re-fetch/reset.
- Parked/small: BINEVAL (Michael likes it), `migrate.sh` self-reconcile, #266 hardening, tool-shelf stacking
  edge, **the board itself is overdue for pruning** (~40K tokens of shipped arcs).

## Set down

- The wholesale trailer-rewrite across all 17 repos (declined — disproportionate; ~2,200 pushed commits for
  a benign attribution line; the setting fixes it going forward).
- A `principles.md` entry for the real-path rail (redundant — the covenant clause is the canonical home).
- A full board prune (deferred — its own task, not a rushed close).
