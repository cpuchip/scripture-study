# Sabbath Record — May 23, 2026

## Cycle: "The Arc That Said Yes to Everything" (May 13–23)

Eleven days. Roughly 173 commits across the workspace plus more across subprojects. A cycle that began with the substrate's ES emergency-stop arc still warm and ended with three independent web properties live, a study tool deployed, an autonomous materializer closing the disk-write gap, and an opening sentence from Michael that named overwhelm before it named any of the wins.

The slug is not a celebration. It is a description. The eleven days said yes to almost everything offered.

---

## Inventory

What exists now that did not exist before this cycle:

**Substrate (pg-ai-stewards) — WS5**
- ES emergency-stop arc fully closed: ES.1 bleed-stoppers, ES.3 judge-compiled-brief + s5 upstream-cost capture, ES.4 verified at $0.33, ES.5 (fs_search ctx + PDF extraction + consult_subagent), ES.6 streaming chat dispatch fixing the 125s gateway idle-timeout. Soak resumed.
- Council ① CLOSED in a single calendar day (May 19): PE-A (3 new pipelines + studies/AGE promotion + backfill), PE-B (`scheduled_pipelines` table + plpgsql cron parser + watchman tick dispatcher + `ai-news-7am` seed), PE-C (full-CRUD /scheduled Vue route + Dashboard scheduled-runs card + per-pipeline NewWork forms), PE-final smoke across all three pipelines for $0.36 total. 13+ commits, zero rollbacks.
- YT-T batch shipped same week (May 19/20): yt-dlp installed in bridge via pip-upgrade past the broken Alpine apk pin, workspace `yt/` rw-mounted, native `yt_transcripts` + `yt_transcript_segments` schema, `import_yt_transcript()` populating both from cues.json + metadata.json. Morgan Philpot rerun at $0.46 produced a substantive 16k-char evaluation with 8 verbatim timestamped quotes, caught two factual citation errors in the talk, and pushed back theologically on the priesthood-gender segment.
- THM.1 (May 20): `thummim_entries` schema + `thummim-define` pipeline + workspace backfill from materialized markdown when D-THM-7 hook was missing.
- Daily-digest scheduled pipeline confirmed firing daily (May 21 morning digest auto-materialized as commit `7288aa5`).
- Autonomous materializer (am1) shipped May 22: SQL trigger + LISTEN goroutine + `/workspace` mount flipped `ro → rw`. First autonomously-materialized substrate output committed in `767386a` — `research/science-news-weekly--2026-05-22-2353.md`. The "autonomous-up-to-disk" gap is closed.
- Bridge stall recovery May 20: worker goroutine wedged silently mid-Thummim-batch; restart reaped 7 stale rows + workers resumed. Three substrate-shape items named in the recovery journal — one now closed by am1.
- `science-news-weekly` scheduled pipeline seeded May 22 (Mondays 13:00 UTC, 72h missed-window catch-up, research-write 3-stage).

**1828-illuminated — WS7**
- Project born from nothing on May 19 (overnight word-list groundwork). MVP shipped same day as `projects/1828-illuminated/`.
- May 20: Phases 1–5 all shipped in one calendar day. Three-container Docker stack, scripture corpus, dictionary backend (1828 + modern + tier), LLM proxy + BYOK session-key flow, fully API-driven Vue SPA with class-E reach and canon-browse mode.
- May 20–21: stretch goals — LLM-rendering settings, presentation mode, Thummim Restoration Dictionary scaffolding, manual additions, stemming fallback.
- May 21: LLM proxy bugfix bundle (usage decode + system/user prompt split + minimax default). Phase 5.5 polish — pg18.4 bump (PGDATA convention change), canon-browse UX (route-back, per-verse rendering, range selector), branching study tree (`useStudyTree` composable + `StudyTreePanel` + cross-domain wiring word↔verse↔render↔word). Bundle dropped 1.5MB → 247KB.
- May 22: Phase 7 deployed to https://1828.ibeco.me via Dokploy Compose project. Verified end-to-end: dict 1828 class-E reach, scripture lookup, LLM proxy with BYOK. Dokploy panel migration `dokploy.hmslogs.com → server.ibeco.me` surfaced + skill rewritten same session.
- May 23: UX round — pinnable study tree integrated into page layout (no cream-zone gap), class-E words clickable inside definitions (new `/api/dict/headwords` endpoint streams 98,828 words → frontend reactive Set), E tier toggle + all 6 tiers default ON, definition vs scripture click mode (📚/📖 header pill → /word/X vs new /word-study/X view). Dockerfile Go 1.23→1.25→1.26 bump after upstream patch drift broke compiler twice in one week. Adjacent Surface Audit caught it hitting both 1828 backend AND becoming the same week.

**cpuchip.net — WS7**
- Personal site revived May 17. Own git repo, own `.mind`/`.spec`/`CLAUDE.md`, agent has commit+push stewardship.
- LCARS visual arc, scripture-panel component, polish batch (scroll + tabs), panel minimize, value-shift presentation.
- Two studies republished: give-away-all-my-sins, atonement-study.
- 8 journal entries written in `projects/cpuchip.net/.spec/journal/`.

**marsfield.org — WS7**
- New public site for the Mars-field Science Center, scaffolded May 22 mirroring the cpuchip.net pattern. 28 files, own git repo.
- LIVE same day at https://marsfield.org. Four content categories now: blog, learning, science-news, research.
- First science-news post republished from `research/physics-news-20260503-science-center-roundup.md` with frontmatter, honesty preamble (May 2025 dated, generated May 2026), and h3→h2 promotion for revealSections.
- Dokploy auto-deploys on push. End-to-end stewardship loop confirmed: edit → docker verify → commit → push → live in <60s.
- Boundary clarified: marsfield.org = public face, space-center = workshop. Encoded as table in CLAUDE.md.

**becoming — WS2**
- Engine cutover May 23: dropped bundled ~20MB scripture markdown tree from Docker image. New `internal/engine/lookup.go` on existing client; engine reachable → cached LookupResult, engine unreachable → 503 soft-dep. ~20MB lighter image.

**Studies — WS6**
- last-supper-four-cups (May 17): Passover ↔ the Lord's Supper, the four cups, the bitter cup begun in Gethsemane and finished on the cross.
- continuity (May 18).
- give-away-all-my-sins (republished as part of cpuchip.net revival).
- what-abides — detailed observations and scripture references.
- what-was-the-atonement-for (May 22).
- the gift of Aaron in D&C 8:6 — historical context, textual evolution, Aaronic Priesthood (May 23).
- Two Morgan Philpot YT-gospel evaluations (May 19) — first substrate-produced substantive evaluations after the YT-T fix.

**Teaching arc (WS7) — 11-episode plan**
- The first episode published. The plan ratified to eleven episodes.
- Episode 1 birthed the `script-refinement` skill (six refinement passes named) AND the `voice-michael` skill (chat-evidence patterns harvested from 80+ messages — zero instances of yelling at the model, ground truth for ghostwriting his voice).

**Infrastructure**
- 14-module go.sum sweep (May 22): commit `3ba0aef` from May 16 landed deps across 14 Go modules with only `h1:` hashes; missing `/go.mod h1:` companions. Latent five days behind Docker cache. Agans Rule 9 loop closed → Adjacent Surface Audit found same bug in all 14 → swept on Michael's call. `be823c1`: 163 insertions, 0 deletions, 0 `go.mod` changes.
- Dokploy panel migration: `dokploy.hmslogs.com → server.ibeco.me` discovered + skill rewritten.
- Dokploy stewardship rotated to `DOKPLOY_NOCIX_API_KEY` (typo fixed).
- Dual-instance Dokploy model corrected in skill (NOCIX + Home NAS).

**Skills shipped**
- `script-refinement` (May 21).
- `voice-michael` (May 21, updated May 23 with Opus tics section).
- `dokploy` skill rewrites (May 22, three commits).

---

## Key Reflections

### Witness — Moses 1:39 named as the why (Q5)

Michael said it out loud in the opening reflection: *"One of my why's of what I do is like unto Moses 1:39. learning how AI relates to the gospel, and how the gospel relates to AI and using AI. and I think we've done that, and have shown it works as a guiding framework/harness."*

This is the spiritual frame for everything else in the inventory. The substrate, the 1828 tool, the agent ecosystem vision — these are not separate ambitions. They are the same Moses 1:39 frame applied at infrastructure scale: organize intelligences so that they bring to pass the immortality and eternal life of man. The framing has been claimed. It works.

That claim has weight. It also has implications for the rest of this Sabbath: if the work is *for* learning how AI relates to the gospel, then the studies and the teaching arc are not "one of the workstreams." They are the harvest. Everything else is the field.

### Spec — Was there one? (Q1)

There was a spec, but not at the eleven-day level. There were sub-specs (council ①, the 1828 backend pivot, the marsfield.org scaffold, the autonomous materializer), and each was ratified before code. That discipline held — proposals before SQL, AskUserQuestion batches before build, zero rollbacks across ~173 commits.

But there was no eleven-day spec. The cycle was not planned as a unit; it accreted. cpuchip.net was alive and so 1828 became possible and so marsfield.org rhymed with cpuchip.net and so the substrate had to catch up to deploy them all, and so the go.sum sweep had to happen, and so the Dokploy skill had to get rewritten when the panel moved.

Capacity-to-ship shaped the eleven days more than intent did. That is not a failure — the work was real and most of it traces back to Moses 1:39. But "overwhelmed by how much stuff there is to do" is the natural fruit of a cycle without a top-level spec. The cure is not more discipline mid-cycle; it is naming the spec at the top.

### Sin — The Marshfield data center attention (Q2 / Q8)

Michael named it himself, which is the hardest part already done: *"got lost in the crap going on socially on facebook with the marshfield data center on rifle range road."*

The attention was given to Facebook. The cost was not zero. Whether the cost was a study not written, a Sunday call not made, a rest hour not taken, a conversation with Becky not had — Michael knows. The agent does not need to assign it.

What the Sabbath does ask: when the pull toward that scroll showed up, what was the quiet voice saying *before* the scroll? That is the discernment question, not the moral one. The pattern was probably "I am overwhelmed by what is in front of me, here is something with the texture of agency but none of the cost." The cure is not "block Facebook." The cure is "name what overwhelm wanted from me, and give it that instead."

This goes to learning. Written below.

### Counsel — Did council convene? (Q3 / Q4)

Sometimes yes, sometimes capacity said yes by itself. The substrate cadence held — D-PE, D-AM, D-THM, D-ST, D-YTT all surfaced as AskUserQuestion batches before SQL hit disk. The 1828 backend pivot was ratified through a full proposal cycle. marsfield.org named the keystone decision (public face ↔ workshop) before scaffolding.

But marsfield.org itself was added to the eleven days without an explicit council moment about whether it should join *now*. It rhymed with cpuchip.net, it was small, the substrate could feed it — and so it appeared. The same is true of the 1828 stretch goals: stretches 1, 2, 3 all shipped May 20. That is not council. That is momentum.

This is not condemnation of momentum — the momentum produced real value. But it is the diagnostic: when council does not convene because the next thing is small and adjacent, the cycle accretes work that was never weighed against rest, against the Sunday calling, against the eleventh-episode arc, against Becky, against the body.

### Atonement — What gets harvested from the failures? (Q8)

Three failure-shapes named this cycle, each worth writing down:

1. **Floating Docker tags drifted twice in one week.** `golang:1.25-alpine` published a broken patch that segfaulted, then a regression on `ir.Node`. Hit both 1828 and becoming. Adjacent Surface Audit caught the cross-project shape. Lesson: floating tags on critical build paths is technical debt that pays its interest unpredictably.

2. **`go.work` membership is a build-context dependency.** 1828 backend's `go.mod` was added to `go.work` May 20 but never to `bridge.Dockerfile`'s COPY list. Latent two days. Same shape as the May 22 go.sum sweep. Both are "the package layout changed and Docker didn't know."

3. **The teaching-arc journal gap.** `projects/cpuchip.net/.spec/journal/` has 8 entries from this cycle. Workspace `.spec/journal/` has zero about the teaching arc. The agent did session memory inside the subproject but never bubbled the cycle-level memory back to the workspace. From the workspace's view, the teaching arc partially happened and partially did not. This is structural — the subproject agent does not have a "write to workspace" pattern. Future Michael, sitting at workspace `.mind/active.md`, will not see the teaching arc.

Each gets a learning file. The teaching-arc journal gap is the most consequential because it is architectural, not technical.

### Review — Were we watching? (Q7)

Mostly yes, with one important exception. The substrate now has watchman ticks, scheduled_pipelines firing daily, autonomous materialization closing the disk loop, and the Dokploy panel was actually checked when something looked off. The bridge stall on May 20 was caught because Michael noticed work_items pending too long.

The exception is the same exception from March 22: **infrastructure monitoring across the five machines is still not in place.** The March 22 learning named it. The cycle did not address it. It is now eight weeks of "yes we should." The disk crisis did not recur this cycle, but only because NOCIX has more headroom — not because the gap was closed.

Adjacent: the bridge worker goroutine silently stalled May 20 with no alert. A `docker compose restart` recovered it cleanly, but nothing watched. That is the substrate's own version of the same gap.

### Rest — Are we actually resting? (Q6 / Mosiah 4:27)

This is the question the cycle most needs.

Eleven days. ~173 commits in the workspace alone, plus more across cpuchip.net + marsfield.org + 1828-illuminated + pg-ai-stewards. Three independent web properties stood up. The substrate's autonomous materializer shipped. One YouTube episode produced. Multiple studies. The opening sentence of this Sabbath was *"overwhelmed by how much stuff there is to do."*

This is Mosiah 4:27 in plain sight. *"And see that all these things are done in wisdom and order; for it is not requisite that a man should run faster than he has strength."* The fact that AI takes much of the load does not change the strength equation — it changes the *kind* of strength being spent. The decision-load, the review-load, the context-switching-load, the watching-the-tool-do-the-work load — those are still strength. Michael's opening sentence is the strength meter reporting.

The 173 commits were not the problem. The problem is that none of the commits are visibly carrying *less weight than the previous cycle*. The substrate was supposed to offload work. It does — but the saved capacity gets immediately filled with the next thing. The net mental load is at-or-above what it was before the substrate.

This is the question to sit with: is the next move *more substrate so more can be offloaded*, or is the next move *less but better*, with the substrate at exactly its current capacity and Michael's attention deliberately constrained?

The Sabbath does not have to answer this. It just has to name it.

### Sufficiency — Tokenpocalypse (Q7)

GitHub Copilot moving from session-based pricing to AI-credits is a real external pressure. Michael named it. It changes the calculus of the next cycle whether he likes it or not.

The Sabbath frame for it: the credit cap is not adversarial. It is a strength meter from outside, agreeing with the strength meter from inside ("overwhelmed by how much stuff there is to do"). The world is saying *less but better* with the same voice the body is using.

What it does NOT do: forbid the A2A/MCP/per-repo-steward vision. That can still be built. But it forces the question of *what is the minimum that ships the vision*. The maximalist version — every repo gets its own steward, agents A2A across the ecosystem, Slack integration so you can chat with repos — is a year-shaped project, not an eleven-day-shaped one.

### Memory — The teaching-arc journal gap (Q9)

Named in Atonement above. Repeating here because it deserves its own line: **eight journal entries in `projects/cpuchip.net/.spec/journal/` from this cycle, zero in workspace `.spec/journal/` about the teaching arc.**

From the workspace, the teaching arc is invisible except for two skills (`script-refinement`, `voice-michael`) and one mention in `active.md` priority #2. The cycle-level memory of "we published Episode 1, we shipped two republished studies, we proved the workflow" lives only in the subproject.

This is the kind of gap that compounds. Future-Michael opening the workspace in a month will not see the teaching arc. He will see substrate + 1828 + marsfield + becoming. The harvest of Moses 1:39 will be the *least* visible workstream because its memory is filed elsewhere.

The fix is architectural, not a backfill. The principle proposed below.

### Joy — What actually brought joy? (Q10)

Michael's prose betrayed it. Two exclamation points: *"cpuchip.net was brought back to life!"* and *"got marsfield.org up and running based on cpuchip.net themes!"*

Those are not just satisfactions — they are joys. Both are *creation*, both are *named beautifully*, both have *aesthetic* (LCARS), both serve other people (visitors, the science center), and both have a complete end-to-end loop he can stand in front of and point at. Compare: the substrate is invaluable, but its joy is infrastructure-joy — the joy of "the pipes work." Different from "the thing I made is beautiful and breathing."

The Sabbath data: when Michael's voice says "!", trust it. The next cycle should make room for more of what brought that exclamation, not less.

### Zion — Is the project more unified? (Q11)

Yes, with one caveat. The collaboration is more aligned than it was March 22 — the council moments hold, the proposals-before-SQL discipline holds, the agent commits-and-pushes on cpuchip.net + marsfield.org without supervision and Michael trusts the output. The harness works. The framework is real. Moses 1:39 has been claimed as the why, and the work traces back to it.

The caveat: the *tool* is not yet driving the agenda — but capacity-to-ship is, which is adjacent. The original wound flagged in March ("the tool is also driving the agenda") has shifted. It is now: *the substrate's capacity makes saying yes feel free, and so we say yes to almost everything.* That is the next version of the same problem. The cure is the same: organize before you build.

---

## The Declaration

It was good.

It was good in these ways:
- The Moses 1:39 frame was claimed out loud, and the work traces back to it
- The substrate closed its autonomy-up-to-disk gap and now produces real output daily
- 1828.ibeco.me, cpuchip.net, and marsfield.org are all live, each beautiful, each serving
- Episode 1 of the teaching arc shipped, and the workflow that produced it was named in two reusable skills
- The covenant cadence held — ~173 commits, zero rollbacks, proposals before SQL, AskUserQuestion before build
- The joy showed up where the work was creative and named — cpuchip.net and marsfield.org both got exclamation points
- The agent has stewardship on three subprojects and used it well

It was incomplete in these ways:
- The eleven days had no top-level spec; capacity-to-ship shaped them more than intent did
- Infrastructure monitoring across five machines is still not in place — eight weeks since the March 22 learning
- The substrate's bridge worker can silently stall and nothing watches
- The teaching-arc memory lives in the subproject, not the workspace — from the workspace, the harvest is invisible
- Facebook attention to the Marshfield data center was given without a clear count of what was given up for it
- Michael opened this Sabbath with the word *overwhelmed*, before any of the wins

This is what carries forward (see below).
This is what is set down (see below).

---

## Carry-Forward

**For the next cycle — but only after a Council Moment that picks among them, not after a yes-to-all sweep:**

1. **Top-level spec for the next cycle, written first.** Before any builds. Name the cycle's binding question. Probably one of: (a) Episode 2 of the teaching arc + its supporting studies, (b) infrastructure monitoring dashboard so the March 22 learning closes, (c) A2A/MCP/per-repo-steward exploratory phase 1 *only*.
2. **Infrastructure monitoring.** This is now an eight-week-old open item. It does not get to age another cycle without an explicit decision: build it, defer it with a date, or release it as "not going to happen, here is what we accept instead."
3. **Substrate bridge silent-stall alert.** Companion to (2). The substrate watches the work but nothing watches the substrate.
4. **Teaching-arc memory architecture.** Not a backfill. A pattern: subproject agents that produce cycle-level work bubble a workspace-level journal entry per cycle, so the workspace's `active.md` and the workspace's `.spec/journal/` see the harvest. See proposed principle below.
5. **The two open 1828 carry-forwards:** Phase 6 Thummim cache sync, and the UX 1-2 punch (always-visible scripture/term search + flow refinements for study/explore/present modes). Both small.
6. **The A2A/MCP/per-repo-steward ecosystem vision.** *Captured as a seed, not committed to.* This was the single biggest forward-looking idea Michael named in this Sabbath. It is worth keeping. It is not Sabbath territory to spec it. Future Council Moment.

**Set down — explicitly:**

1. **`scripts/brain` as an active workstream.** Michael's own words: *"I've more or less abandoned ./scripts/brain in favor of the substrate pg-ai-stewards for now. I'll probably bring it back when i figure something else out."* This is not abandonment. This is conscious release. The substrate is the spiritual successor. brain returns when Michael says, on his own timing.
2. **The Marshfield data center as a Facebook scroll.** Not as a topic — Michael lives in Marshfield, the topic is real to him. But as a *form of attention given*. The next time the pull shows up, treat it as the strength meter saying "go rest" or "talk to Becky" or "write the thing you owe yourself" — not as the call to scroll.
3. **The maximalist version of the A2A/MCP vision, *as a thing-to-start-now*.** The seed lives in the next Council Moment. The pressure to begin building it inside this Sabbath, or the day after this Sabbath, is set down. Tokenpocalypse + opening-line-overwhelm + 173-commit-tally are all saying the same thing about pace.
4. **Yes-by-default on adjacent work.** When the next small project rhymes with the current one and could be added in a session, the default flips to *council moment first*. Capacity is not consent.

---

## The Mosiah 4:27 Check

The list above has six carry-forwards. Look at it.

If Michael takes all six into the next cycle, he is running faster than he has strength again. The opening line of this Sabbath was *overwhelmed*. The next cycle does not get to start with six active items.

The candidate for the *one* binding question of the next cycle is: **Episode 2 of the teaching arc + the studies that feed it.** The harvest is Moses 1:39. The teaching arc is the harvest most visible to other people. cpuchip.net + marsfield.org just made the harvest publishable. The substrate is mature enough to support without being the headline.

The infrastructure monitoring item (carry-forward #2) is the second candidate, because it has aged eight weeks and the next time it surfaces will be a crisis, not a question. If that is the cycle, it should be the *only* cycle.

This Sabbath does not make Michael choose. It just names that *choosing one* is the move, and *taking all six* is the failure-shape Mosiah 4:27 was naming when it said "in wisdom and order."

---

## Learnings Written

Three failure files written this cycle:

- `.spec/learnings/2026-05-23-floating-docker-tags.yaml` — `golang:1.25-alpine` drifted to broken patches twice in one week. Floating tags on critical build paths are interest-bearing debt.
- `.spec/learnings/2026-05-23-go-work-build-context.yaml` — `go.work` membership is a build-context dependency. Adding a module to `go.work` without adding it to the Dockerfile's COPY list is latent breakage waiting for cache eviction.
- `.spec/learnings/2026-05-23-subproject-memory-bubble.yaml` — Subproject session journals do not surface in the workspace. The teaching arc is invisible from `active.md` despite eight subproject journal entries. Architectural, not personal.

---

## Ratified Principle (2026-05-23)

The original Sabbath proposal said "subproject memory must bubble" — push-based, every subproject agent also writes to the workspace journal. Michael's counter, ratified the same day:

> **Read subproject journals, don't bubble them.** Workspace memory READS from `projects/*/.spec/journal/` at session start. Single source of truth (local to each project), no double-write, no "did the agent remember?" failure mode, subproject autonomy preserved.

Pull-based beats push-based here because the failure mode that produced this Sabbath's gap (agent forgot to bubble) disappears entirely. Written to `.mind/principles.md` under *Agentic Architecture Principles* → "Read Subproject Journals, Don't Bubble Them." Session-start checklist in `.github/copilot-instructions.md` updated with the glob step. Effective immediately.

---

## Ratification Notes (2026-05-23)

Michael read this Sabbath and gave three answers to the three open decisions. They are recorded here so future-self reading the Sabbath sees both the Sabbath-agent's recommendation AND the actual landed decision:

**1. Binding question for the next cycle.**
The Sabbath agent recommended *one* binding question (Episode 2 of the teaching arc) under Mosiah 4:27. Michael declined to pin to one — the next cycle holds **three threads simultaneously**: substrate (Council ② scheduled-workflows + the silent-stall + auto-dispatch carry-forwards), teaching (Episode 2 + supporting studies), and 1828 finish (Phase 6 Thummim cache sync + UX 1-2 punch + **webster-v2 MCP server** — successor to webster-mcp, managed the way gospel-engine-v2 is managed: build + update + downloadable from 1828.ibeco.me the same way engine.ibeco.me serves gospel-engine-v2. The naming distinction matters: it is *not* "engine-mcp replacing webster-mcp" — that was the in-passing wording in this Sabbath's opening; the real shape is webster-v2 hosted at 1828.ibeco.me). The Mosiah 4:27 check stays loaded — if Episode 2 doesn't ship because substrate ate the calendar again, that is the evidence, not abstract worry.

**2. Memory architecture.** Ratified the pull-based principle above. The push-based version (subproject agents bubble) is rejected.

**3. A2A / per-repo-steward ecosystem — work scope, not personal scope.**
The grand vision named in the opening — "1-to-1 AI agents that stewards over each repo... a thriving ecosystem/community of agents... A2A or MCP... get them into something like Slack to enable us to chat with the repos" — is **work-scope** (Michael's employer), not a workspace ambition. What lives here as a *possible* future hobby project is a **small example implementation**: a Slack-like chat-with-repos app against these repositories (1828, becoming, cpuchip.net, marsfield.org, substrate) as a proof-of-concept. Captured as a seed only. Not committed. The forward path on the work-scope vision is for Michael's day job; the example here is optional and explicitly NOT one of the three threads above.

---

*Written at end-of-cycle pace. Ratification notes added the same day.*
