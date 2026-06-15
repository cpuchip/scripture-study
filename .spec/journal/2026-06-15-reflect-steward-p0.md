# 2026-06-15 — Reflect-steward P0: dry-run + the fixes that close the loop

Two sessions, same arc. Michael ratified the reflect-steward design, said "go
build — dry-run report for the morning," then on waking: "proceed with fixing
and finishing the features." All on the OSS scratch stack; nothing autonomous
flipped on (his Hinge). Private journal — names the Vivint use-case, so it stays
out of the public repo; the public substrate-dev story lives in the OSS commit
messages (`b116ab9`, `dd34715`, `f8ea1bb`).

## The headline: the engine already existed; the work was making it run clean

The substrate's `planning` pipeline IS the reflect-steward loop (context_gather →
explore → synthesize → propose_work → review_plan), with
`enqueue_proposed_work_items` landing proposals as parked `agent_planning`
work-items. So P0 was configuration + bug-fixing, not a new pipeline. Created a
`vivint` intent (public-only scope in its values) and ran it.

## What the dry-run did + the bugs it surfaced (all now fixed + verified)

The Vivint cold-start cycle proposed a sensible *research plan* — 4
source-specific research items (BBB/Trustpilot/Reddit/app-store) + 1 synthesis —
grounded, valid, the review stage passing with a structured verdict. The value
was in the bugs it surfaced, all now fixed on fresh-container-verified commits:

1. **`research` agent never core-seeded** (`b116ab9`). 13's pipelines + echo-test
   ran on `agent_family='research'`, but a virgin core seeded no such agent; the
   digesters had drifted onto an unseeded `stewards-explore`. Unified on one
   core-seeded, web-capable `research` agent; virgin-smoke OK 3b guards it.
2. **`yt` MCP broken** — bridge rebuilt `WITH_YT` (7/7 servers OK).
3. **Finding B — the autonomy blocker** (`dd34715`). The review-verify gate only
   promotes a review stage to `verified` when its output starts with
   `REVIEW: passes|revised`; planning's `review_plan` emitted a dead (never-parsed)
   JSON verdict, so runs stalled at `maturity=planned` and proposals never
   auto-enqueued. Switched to the `REVIEW:` convention. **Proven:** a fresh cycle
   reached `verified` and auto-enqueued 3 proposals with no manual call (0 before).
4. **Context/memory + pool-read grants** (`dd34715`) — `research` got
   doc_search/get/similar + compact_context + remember/forget/context_* (minus the
   gated base-prompt editor) so the steward (and big-book/transcript digests)
   self-manage context.
5. **Finding A** (`f8ea1bb`) — `work_item_create` injects `input.today`, so no
   launch hard-fails on the resolver's missing-field behavior.

Full batch re-verified: rebuilt the image, virgin-smoke green on a fresh
container (all 6 OK blocks).

## The go-live bundle — proposed overnight, then BUILT with Michael (same day, `c285ff5`)

I deliberately did NOT pre-build the autonomous-control surface at 4am
(`dominion_in_council` + design choices that are Michael's). Presented the design;
Michael answered the two questions — kill switch = **global + per-intent**
("stop everything, then decommission the bad intent"); approve = **queue, not
auto-dispatch**, capacity-gated ("so we don't flood the work") — and said "lets
build." So `22-reflect-steward.sql` shipped: the kill switch, the approval queue +
capacity-gated drain (`reflect_drain_approved` hooked into the watchman tick,
respecting `reflect_max_concurrent`), the check-in verbs, and the `reflect-checkin`
skill. Functional proof on live (pause→0, approve queues, cap gates dispatch) +
virgin-smoke OK 7 on a fresh container. The Vivint schedule is seeded
`enabled=false`. **Only the flip remains (Michael's Hinge):** `UPDATE
scheduled_pipelines SET enabled=true` + the Claude-watchman wake.

The presiding chain made concrete: Michael authorized at the boundary (the design
+ the two decisions), I built within it, and going-live stayed his.

## Knowledge architecture (Michael's noodle, same day) + GO-LIVE

Michael then shaped the knowledge layer: a per-intent dedup ledger
(`intent_source_ledger` + `intent_sources_recent`/`record`, `1a24ac0`) so the
gatherer builds the pool UP instead of re-scrubbing, and **project-neighborhood
scoping** (`70a5fd0`) — `docs` tagged by project (via `work_item_create`, FK-safe
against `stewards.projects`), `pool_search` enforced-scoped to a project's
neighborhood; Vivint walled off, ai↔books pre-wired to cross-pollinate. doc_search
stays global for meta-studies. All virgin-smoke OK 7 + fresh-container green.

**GO-LIVE (Michael: "set it to auto + kick an immediate run"):** schedule enabled
(every 3h). The clean auto-proof run climbed raw→researched→planned→**verified**
and **auto-enqueued 2 proposals with no manual call** — the full loop works
unattended (dispatch→gather+record-sources→synthesize→propose→verify→auto-enqueue).
Watchman cron every 3h (session-only). `reflect_pause()` is the kill switch.

## The stumble + the lesson (durable: [[feedback_pg_ai_stewards_rebuild_discipline]])

The first launch run came back `raw`/0-proposals: re-applying `04` standalone for
Finding A had reverted `work_item_advance` to its pre-maturity-bump version
(consolidated chain = later-file-wins; a lone re-apply breaks the ordering). Repo
was always correct; fixed by re-applying `05→22` in order. **Never single-file
re-apply to a live consolidated chain — apply changed-file→end-of-chain, or
rebuild.**

## Carry-forward

- Report (private): `.spec/proposals/reflect-steward-p0-dryrun-report.md`.
- **Persistent watchman** = the real follow-up: the cron is session-only (durable
  not honored), so a substrate-internal watchdog (a scheduled pipeline that checks
  reflect_status + auto-pauses on runaway) is the right persistent mechanism.
- Minor: `on_maturity_verified` tries a `sabbath_dispatch` on an unseeded `plan`
  agent (caught/non-fatal) — same missing-agent shape as `research` was.
- Next intents (ai/books) when Michael's ready — projects + neighborhood already seeded.
- Live OSS stack carries earlier dev cruft (a `stewards-explore` row); repo is clean.

---

## PM update — the watchdog is built; M4 proven with Vera; cut-prep done

The "substrate-internal watchdog" named above as the persistent follow-up is
**built and live** (`23-reflect-watchman.sql`, OSS `082be5e`). It's deterministic
(no LLM), rides the bgworker heartbeat, and auto-pauses on in_flight / consecutive-
failure / 24h-spend / proposal-flood breaches, logging every trip. Michael loved
the framing — "a presiding agent of sorts." Live reading at ship: nominal, but
autonomous spend was **$8.51 / $10 (24h, 85%)** from today's heavy proof runs —
flagged to Michael as his dial (whether to bump the cap for the soak); I left the
principled $10 (raising a safety cap is his call, not mine — presiding covenant).

**M4 proven with a face:** dispatched a `persona-turn` on the OSS core as "Vera," a
Vivint CX-analyst persona, fed the gathered pool (BBB/ConsumerAffairs/Trustpilot
docs) + a PM question ("top 3 to fix this quarter?"). She answered in character,
grounded entirely in the pool with real cites (Trustpilot 3.9/59k; BBB Billing the
#2 bucket ~2,134/9k; JustUseApp #400 errors). The P0.5 "give the intent a face you
can talk to" architecture works e2e: gatherer writes the pool → host reads → tool-
free persona talks. Full ai-chattermax wiring (talk to Vera live in a room) is the
remaining richer piece — and it overlaps the CUT's gating prereq.

**The CUT's real blocker (discovered today):** ai-chattermax's persona-host still
dials the LIVE bridge. Stopping live without repointing takes chat.ibeco.me
personas dark. Parity was never the blocker — this is. Runbook:
`pg-ai-stewards-workspace/parity/cut-runbook.md` (gating prereq + data-carry table
+ sequence + rollback + D1-D4). Sabbath-tension resolved (promote-trigger wrapped).

---

## THE CUT — executed (2026-06-15 PM). OSS is the one substrate; live retired.

Michael: "lets cut over… full send." It turned out to be a real migration, not a
knob — and the diligence mattered.

**Two questions answered first:** Vera was a one-shot proof dispatch, never a
deployed chat persona (no key/room) — that's why she wasn't on his server. And the
dialing: persona-host is the connector and dials BOTH (wss→chat.ibeco.me +
Postgres→substrate); chat.ibeco.me never knows about the substrate. The swap = the
persona-host's STEWARDS_DSN.

**The real blocker (not parity):** the OSS bridge shipped 5 generic MCPs; the
tool-using personas need gospel/strongs/webster/byu/search (local binaries baked
into the LIVE bridge). The workspace compose.override was an empty P1 stub. So I
built `stewards-oss-bridge:michael` (bridge.Dockerfile FROM the live image →
inherits the domain binaries + webster data + yt-dlp/ffmpeg; OSS clean-room
substrate binaries swapped in; same Alpine/musl base = drop-in). gospel.db comes
from the /workspace mount.

**Sequence (all verified):** archive dump (5.6M) → 36/36 overlays onto OSS →
`fiction` agent + missing `persona_host.personas` (chattercode/pg-starlet) COPY'd
live→OSS (the self-seed only covers defaults; captured as overlay cut1) →
Michael-bridge up, refresh-tools 13/13 → persona-host swapped (stop live first →
no double-fire; verified no native exe either): callie/chip-assistant/dm-assistant/
npc-ally/chattercode all connected to chat.ibeco.me off OSS → enabled gospel+search
(overlay cut2; they were seeded dormant — the librarian's gospel error) → librarian
re-verified e2e: fetched + quoted Alma 32:21 correctly via gospel-mcp → .mcp.json
repointed (55434 + OSS exe) → live stopped (dev/bridge/ui, NOT removed).

**Lessons:** (1) persona world is owned by persona-host+ai-chattermax (self-seeds
the schema + default personas; runtime-added ones like chattercode must be carried).
(2) The cut's hard part was the bridge's domain-MCP coverage, never parity. (3)
Tables can have the same columns in DIFFERENT order across the live/clean-room
schemas — positional COPY fails; use explicit column lists.

**Carry-forward:** human chat-verify (drop a message → persona replies);
born-clean from-source rebuild of the domain MCPs (bridge is pragmatically FROM the
live image); remove the stopped live containers + archive the volume after the OSS
stack soaks clean. Recovery if needed: `docker start pg-ai-stewards-{dev,bridge,ui}`
+ revert .mcp.json + swap persona-host back.
