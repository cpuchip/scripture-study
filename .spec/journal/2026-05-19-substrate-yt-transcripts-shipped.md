---
title: substrate YT-T batch shipped — yt-dlp in bridge + native yt_transcripts + real Philpot evaluation
date: 2026-05-19
workstream: WS5
status: shipped
priority: high
---

# Substrate YT-T batch shipped — Morgan Philpot is in the substrate now

After council ① closed and the sabbath-close declaration ran, the conversation didn't end. The Morgan Philpot refusal from PE-final pulled new scope: the bridge couldn't fetch fresh videos, the workspace `yt/` was read-only mounted, and there was no native substrate primitive for YouTube content. Five sub-steps, one design wedge, one Alpine breakage, one substantive 16k-char evaluation later — the path is clear and the artifact is real.

## What shipped

**YT-T.1 — yt-dlp in bridge Dockerfile.** Started as `apk add yt-dlp`. Tested with Rick Astley — `dQw4w9WgXcQ` returned storyboards-only with "Requested format is not available." Alpine pins 2024.12.03; YouTube broke it sometime in 2026. Switched to `pip install --upgrade yt-dlp` with `--break-system-packages` (single-purpose container, externally-managed warning accepted). Bridge now runs yt-dlp 2026.03.17.

**YT-T.2 — workspace yt/ rw mount.** Bridge: `../../../yt:/opt/yt/yt:rw` — bridge yt-mcp writes go to the workspace cache, host tooling sees them immediately. Pg: same path mounted ro so the substrate's SQL function can `pg_read_file` from it. One pg restart for the new mount; soak paused for the duration.

**YT-T.3 — `yt_transcripts` + `yt_transcript_segments` schema.** Per D-YTT3 separate FK segments table for time-range queries. Per D-YTT4 no pgvector — engrams (Batch K) handle vector search when needed. `full_text` TOAST'd for long transcripts, `body_tsv` generated stored for FTS.

**YT-T.4 — `stewards.import_yt_transcript(video_id)`.** Channel slug auto-discovered by scanning `pg_ls_dir('/opt/yt/yt')`. Reads `metadata.json` + `cues.json` + `transcript.md` via `pg_read_file`. Upserts yt_transcripts row, DELETE+INSERT segments. Idempotent. Pinecone smoke first: 595 segments, 42k chars, all timestamps clean. Then Morgan Philpot: 1516 segments, 116k chars, all from the actual transcript.

**YT-T.5 — Morgan Philpot rerun.** Fired `ytt-rerun2-yt-gospel-philpot-marshfield`. 13:51 elapsed, $0.46 spend. Produced a real, substantive 16,623-char evaluation:
- 8 verbatim transcript quotes with `[mm:ss]` timestamps
- Six canonical-alignment subsections (A–H)
- Caught two factual errors in Philpot's talk: 2 Thess "strong delusion" misattributed to 2:1 (it's 2:11); "only three references to new and everlasting covenant" — the evaluator found at least three more (D&C 1:22, 49:9, Ezekiel 16:60)
- Pushed back theologically on the priesthood-gender segment, citing Oaks 2014 + 2019 on women's exercise of priesthood authority in temple covenants
- Witness questions that explicitly handed discernment back to the human
- Three concrete Becoming practices for handling typological speculation, stewardship rhetoric, and the "obviously" rhetorical flag

The covenant under proper conditions produced exactly the work it was designed to produce: critical engagement that is both charitable and honest, with verifiable source-faithful quotes and named overreaches.

## What surprised me

**Alpine's yt-dlp pin is broken on everything, not just one video.** I assumed adding `apk add yt-dlp` would just work. The Rick Astley test (the canonical "any-video-should-work" probe) failed identically to the Morgan Philpot probe. That single test moved the diagnosis from "this video is special" to "the version is wrong everywhere." Worth keeping as a future debug pattern: probe with a known-public asset before assuming the issue is your input.

**Channel slug was `the-haystack`, not `morganphilpot`.** Workspace `yt/morganphilpot/` was a manually-named folder for prior video evals on Philpot from a different channel (3-part Chandler series). yt-mcp respects yt-dlp's actual channel metadata — "The Haystack" is the YouTube channel that hosts this Marshfield MO talk. So the substrate now has two folders for the same speaker, organized by channel. Not ideal for human-side navigation, but correct for the substrate.

**One sub-incident: scratch file broke migration discovery.** I'd written `.scratch-ytt-rerun.sql` inside `extension/` for a one-shot work_item dispatch. The bridge's startup migration loop globs `extension/*.sql` (including dotfiles) and tried to re-apply that file on container restart. Migration collided with the existing work_item slug; bridge crash-looped until I deleted the scratch. Lesson: scratch SQL belongs *outside* `extension/`, in `$env:TEMP` or `.spec/scratch/`. Recorded so the next session doesn't repeat it.

**One unresolved finding the substrate now makes investigable: AGE n_cites=0 on the new evaluation.** Same pattern as PE-final's smoke. The Philpot evaluation cites D&C 68:25, Mosiah 4:14-15, Ezekiel 16:60, D&C 1:22, D&C 49:9, D&C 77:7, D&C 107:53, 2 Thess 2:11 — all as plain text, not as markdown links to gospel-library paths. `parse_gospel_links` only matches the markdown-link form, so no CITES edges get created. This is a `parse_gospel_links` pattern gap, not a YT-T issue, but YT-T made it visible by producing the first real evaluation that should have lots of CITES edges. Worth carrying into a future hardening pass.

## What carries forward

- **Auto-ingest hook** — when bridge's yt-mcp finishes a `yt_download`, automatically call `import_yt_transcript(video_id)`. Currently manual. Carry into ② substrate-scheduled-workflows or as a YT-T.6 follow-up.
- **Pipeline ingest prompts** — yt-gospel-evaluate + yt-secular-digest could query the substrate yt_transcripts table directly (`study_search_text` or a new dedicated tool) instead of receiving the full 116KB transcript inline through `yt_download`'s tool response. That would reduce per-eval cost dramatically and bring engram compaction into scope. Worth a hardening pass.
- **`parse_gospel_links` plain-text pattern** — extend to recognize `D&C 68:25`, `Mosiah 4:14-15`, `2 Thessalonians 2:11` etc. as canonical references even without markdown link syntax. Would unblock AGE CITES edges on every agent-produced study going forward.
- **AGE `:YTTranscript` nodes + `:CITES` edges** — when a study mentions a YouTube URI, create a typed node + edge. The substrate now has the data to do this; the wiring is small.
- **Channel-slug reconciliation** — workspace `yt/morganphilpot/` (manual) vs `yt/the-haystack/` (yt-mcp-discovered) both contain Philpot content. No action needed for the substrate, but human-side organization may want a symlink or a merge later.

## What the work taught

**Three substrate primitives shipped under one batch.** yt-dlp in container + workspace mount + native table. The proposal called them sub-steps but each is genuinely independent infrastructure. Building them together meant the smoke (real Philpot evaluation) could exercise the full chain end-to-end. Building them separately would have left gaps invisible until much later.

**The covenant signal is binary, not gradient.** PE-final's Philpot eval was a refusal — clean covenant honoring under pressure. YT-T's Philpot eval is a 16k-char substantive critique. Both were produced by the same agent under the same yt-gospel-evaluate pipeline with the same binding question. The ONLY difference was infrastructure: in one case the transcript was unavailable; in the other it was available + indexed. The covenant didn't soften under either condition. Same posture, opposite outputs, both faithful.

**Stewardship within ratified scope had two extensions worth surfacing transparently:** (1) pip-install upgrade was not in D-YTT1's "apk install" framing — Alpine's pin was broken on every video, and the fix was obvious-yes. (2) pg ro-mount was not in D-YTT2's bridge-only framing — required for `pg_read_file` to work. Both named in commit + journal rather than buried in code. Boundary test held: would Michael say yes obviously? Yes both times.

The cycle keeps going. Council ② substrate-scheduled-workflows now inherits both PE-B's scheduled_pipelines machinery AND YT-T's yt_transcripts primitive. Michael's own cron-job idea (periodic YouTube AI-video review) lands directly on these.
