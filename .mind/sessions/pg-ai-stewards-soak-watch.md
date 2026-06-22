# pg-ai-stewards LOCAL-MODEL overnight soak — watch runbook + log

**Started 2026-06-19 ~05:52 UTC. REPORT DUE: 13:00 UTC (08:00 CDT) to Michael.**

Michael's ask: re-enable pg-ai-stewards on the **local FlexLLama models** (free fallback),
let it pick up the **books / ai-news / videos / vivint** autonomous work, **watch it work
through the phase wake-ups**, **report ~8am CDT**, and **if it goes off the track, pause the
whole thing**.

This file is the runbook for each wake-up (fresh context safe) AND the accumulating log.

---

## What was changed tonight (to set up the soak)

1. **Routed 5 pipeline families to local role-aliases** (LIVE jsonb edit on `stewards.pipelines.stages`;
   research-summary was already local). Casting = read/gather→`ingest` (gemma-12b),
   digest/synthesize/propose→`reason` (qwen3.6-27b), critique/review→`critic` (qwen3.6-27b).
   Families: book-digest, book-curate, playlist-digest, planning, research-write.
   - Role-aliases are local-FIRST with paid fallback: ingest=gemma→nemotron→kimi(opencode_go),
     reason/critic=qwen→kimi. So "fall back to local" = local is priority 0; opencode_go only if local fails.
   - ⚠ **This is a LIVE-ONLY change (drift from repo)** — an experiment ("see how they hold up").
     Morning decision: codify as a workspace overlay (if local holds up) OR revert. NOT yet in repo.
2. **proposals cap 50→70** (`reflect_guard_max_proposals_pending`) — headroom so a triage backlog
   (was 40/50) doesn't spuriously halt the experiment. Revert to 50 in the morning.
3. **Resumed autonomy** (`reflect_resume()`; was guard-paused on "6 consecutive failures" =
   opencode_go HTTP 429 GoUsageLimitError — the paid plan's usage cap; local fixes the root cause).
4. Proofs before sleeping: echo-test on `reason`→**flexllama done** (chat path); research-summary
   8151 = full tools-on pipeline proof (in flight at handoff).

## The rig (must stay up — local dispatch depends on it)
- `flexllama-stewards` container, OpenAI endpoint `host.docker.internal:8090` (from substrate).
  qwen3.6-27b q8@192k GPU0 (~47 tok/s) · gemma-12b f16@256k + nemotron-4b q8@512k GPU1.
- If `curl -s localhost:8090/v1/models` shows <3 models → rig down → ALL local dispatch fails →
  pause + note. (Restart: `pwsh projects/pg-ai-stewards-workspace/scripts/setup-flexllama.ps1`.)

## Schedules (UTC) — what should fire overnight
- book-digest-hourly `0 * * * *` (hourly) · book-curate-cron `0 */2` · vivint-reflect `0 */3`
  (planning) · playlist-digest-cron `0 */6` · ai-news-7am `0 13 * * 1-5` (= report time) ·
  science-news-weekly `0 13 * * 1` (Mon only — not tonight). Empty source → halt_on cancels (no waste).

---

## EACH WAKE-UP: run these checks
```bash
docker exec -i stewards-oss-pg psql -U stewards -d stewards <<'SQL'
-- 1. guard: is it about to / did it trip?
SELECT 'would_trip='||((stewards.reflect_guard_signals())->>'would_trip')||
       ' breach='||COALESCE((stewards.reflect_guard_signals())->>'breach','none')||
       ' paused='||stewards.config_get_text('autonomy_paused','?');
-- 2. recent autonomous runs: provider should be flexllama; watch failed/error
SELECT pipeline_family, status, count(*) FROM stewards.work_items
 WHERE actor IN ('scheduler','reflect-steward') AND updated_at > now()-interval '3 hours'
 GROUP BY 1,2 ORDER BY 1,2;
-- 3. which providers are actually serving chat (local vs paid fallback)
SELECT provider, status, count(*) FROM stewards.work_queue
 WHERE kind='chat' AND created_at > now()-interval '3 hours' GROUP BY 1,2 ORDER BY 1,2;
-- 4. autonomous spend (should be ~$0 on local; >$0 = paid fallback happening)
SELECT COALESCE(round(sum(ce.micro_dollars)/1e6,4),0) AS spend_3h FROM stewards.cost_events ce
 JOIN stewards.work_items w ON w.id=ce.work_item_id
 WHERE ce.at > now()-interval '3 hours' AND w.actor IN ('scheduler','reflect-steward');
-- 5. recent errors (diagnose)
SELECT provider, left(error,80), count(*) FROM stewards.work_queue
 WHERE error IS NOT NULL AND created_at > now()-interval '3 hours' GROUP BY 1,2 ORDER BY 3 DESC LIMIT 8;
SQL
```

## KNOWN-HARMLESS (do NOT pause for these)
The 3 background judges **engram-extractor** + **judge-brief** (deepseek-v4-flash) and
**watchman-consolidator** (kimi-k2.6) are hardcoded to opencode_go → they 429 (paid cap) and
error out. EXPECTED + harmless: $0 spend, non-blocking (main pipelines complete without them),
NOT scheduler/reflect-steward actors so they don't count toward the guard. Only effect =
context-engine auto-fold degraded (big local 192-256k windows absorb it). Left unedited on
purpose (untraced live fn edit, not at midnight). Morning report: recommend routing local.
**Do NOT pause for opencode_go 429s on these three.** Pause only if a MAIN pipeline
(book/playlist/curate/planning/research-summary) fails or spend climbs.

### ALSO known: the 15-min reaper false-kills slow local reads (NOT off-track)
The periodic reaper kills any chat in_progress >15min as "stale" (HARDCODED, cloud-speed-tuned).
A big-book `read` on gemma can legitimately exceed 15min on local → reaped → that book-digest
"fails" (error: "periodic reaper: stale in_progress >15min"). This is a SLOW-READ false-kill, NOT
a model failure — the book requeues, $0. **Do NOT pause for reaper-stale kills.** The ≥3-local-
failures pause rule means genuine model errors/garbage, NOT reaper kills. KEY REPORT FINDING:
make the reaper threshold a config + raise it for local (or chunk big reads). Not edited live (hardcoded fn).

### ALSO known: gemma single-slot contention under concurrent ingest (NOT off-track)
gemma runs `--parallel 1` and is the priority-0 `ingest` model. A long book `read` monopolizes
gemma's one slot (GPU1 92% util for 20+min); a CONCURRENT ingest call (e.g. planning context_gather)
then fails at the HTTP layer ("POST .../chat/completions: error sending request for url") = a client
timeout while gemma is busy. Worst at top-of-hour when schedules burst. nemotron (ingest pri-1, fast
223 tok/s, 512k) sits IDLE meanwhile. $0, requeues, rig stays up. **Do NOT pause for these** — it's a
throughput limit, not a breakage. REPORT FINDINGS (design, not midnight): (a) split ingest across
gemma+nemotron (or route book-read→gemma, gather→nemotron); (b) stagger schedule crons off the same
:00; (c) raise gemma `--parallel`; (d) raise the request/reaper timeout for local.

### ALSO known: qwen (reason) breaks grammar/structured-output stages (NOT off-track)
qwen3.6-27b is a REASONING model; on a grammar-constrained / structured-output stage it 500s with
"model produced output that does not match the expected peg-native format" (wq 8678, playlist-digest
`digest`). book-digest `digest` (also qwen) SUCCEEDS → it's that stage's response_format/grammar that
qwen's thinking tokens violate. Effect: the **videos (playlist-digest) leg likely persistently fails**
on qwen-reason. $0, requeues. **Do NOT pause** (1 leg, not a runaway). REPORT FINDING: route grammar/
structured stages to a non-thinking model (gemma/nemotron) or relax the grammar / strip thinking first.

## OFF-TRACK → pause the whole thing
`SELECT stewards.reflect_pause('off-track: <reason>');` then log it here + flag in the report.
Pause if ANY:
- guard would_trip / already paused (it self-pauses; confirm + note why).
- ≥3 NEW failures since last check on local providers (local models can't do the work).
- spend climbing past ~$1 (it's silently falling back to paid → not the experiment Michael wanted).
- in_flight runaway (≥6) or the same work_item stuck dispatching >30 min (loop/timeout).
- rig down (<3 models on :8090) and not recovering.
- outputs are garbage/empty (spot-check a digest's result content).

## REPORT (at/after 13:00 UTC) — deliver to Michael as text
Per intent (books / ai-news / videos / vivint): what ran, completed vs failed, which local model
did which phase, a quality spot-check (read 1 digest/proposal output), total spend (expect ~$0),
any pauses + why, and a recommendation: **codify the local routing as an overlay, or revert?**
Then this soak is done — delete the watch cron.

---

## CHECKPOINT LOG (append each wake-up: time, what ran, verdict)
- **05:52 UTC** — setup done (above); echo proof green on flexllama; guard reset (consec=0,
  would_trip=false); research-summary 8151 full-pipeline proof dispatched; burst due 06:00.
- **06:01 UTC** — ALL GREEN. ai-news-7am auto-fired on resume → gather+synthesize done, at REVIEW
  (local). full-pipeline proof at synthesize (gather did multi-round gemma tool-calls, all done).
  chat last 15m = **16 flexllama done + 2 in_progress, ZERO opencode_go**. spend 1h = **$0.0000**.
  guard would_trip=false, in_flight=3. Watch crons live: 16076a3f (hourly :22), b8a456ed (report 13:04).
  06:00 burst (book-digest/curate/playlist/vivint) not dispatched yet — next scheduler heartbeat; :22 tick will catch.
- **06:25 UTC (tick 1)** — HEALTHY, no pause. Full-pipeline proof **completed** end-to-end on local
  (gather→synthesize→review). 06:00 burst: **book-curate, planning, playlist-digest, 2× research-summary
  COMPLETED on local**; book-digest in_progress (~25m, plausible for a local book read — watch next tick).
  chat 30m = **71 flexllama done + 1 in_progress + 30 opencode_go:error**. Spend 3h = **$0.0000**.
  guard would_trip=false, in_flight=2. The 30 errors = the KNOWN-HARMLESS background judges (engram-extractor/
  judge-brief/watchman-consolidator, hardcoded to opencode_go) — diagnosed, not off-track, left unedited.
- **07:24 UTC (tick 2)** — HEALTHY, no pause. ALL autonomous runs last 3h **completed**: book-curate,
  **2× book-digest**, planning, playlist-digest, 2× research-summary (the prior in_progress book-digest
  finished + the 07:00 hourly ran). chat 70m = **46 flexllama done + 24 opencode_go:error (the 3 known judges)**.
  Spend **$0.0000**. guard would_trip=false. **Cancelled a pre-soak ZOMBIE** book-digest stuck at 'read'
  since 06-18 04:32 (~27h, created day before the soak; cascade killed 0 active rows) → in_flight 2→0.
  Local models are carrying the full slate cleanly.
- **08:24 UTC (tick 3)** — HEALTHY, no pause. last 90m: book-curate completed, book-digest completed(1),
  + the zombie cancelled(1), + **1 book-digest FAILED** = the **15-min reaper false-killed a slow gemma
  book `read`** ("stale in_progress >15min"; wq 8502 flexllama/gemma-12b at 08:00). NOT a model failure —
  slow-read false-kill; book requeues; $0. chat 70m = 7 flexllama done + 1 flexllama-error(the reap) +
  1 opencode_go(judge). spend **$0.0000**, guard clean, in_flight 0, nothing stuck. Activity lower (shelf
  draining). **KEY REPORT FINDING: 15-min reaper too short for big-book local reads → config-ize + raise.**
- **09:24 UTC (tick 4)** — HEALTHY w/ contention, no pause. last 75m: 1 book-digest completed-ish + 1
  book-digest in_progress (24m gemma read, GPU1 92% util — grinding, not hung) + **planning FAILED at
  context_gather** ("error sending request" = gemma busy w/ the read → concurrent ingest HTTP-timeout) +
  the tick-3 reaper item surfacing its auto-advance fail. chat 75m = 5 flexllama done + 1 flexllama-error
  + 1 in_progress + 1 opencode_go(judge). spend **$0.0000**, guard clean, rig UP (3 models), proposals 43/70.
  **2nd KEY FINDING: gemma --parallel 1 contends under concurrent ingest; nemotron idle.** vivint reruns 12:00.
- **10:24 UTC (tick 5)** — HEALTHY steady-state, no pause. last 75m: book-curate + book-digest completed;
  **19 flexllama done**. Same 2 known failures RECURRING (book-digest read→reaper, planning context_gather→
  gemma contention) — confirms both findings are systematic, not flukes. 8 opencode_go errors = judges.
  spend **$0.0000**, guard clean, in_flight 0, nothing stuck. Steady: local carries the load; hourly
  book-read + concurrent gathers are the two recurring trip points. No new failure modes.
- **11:25 UTC (tick 6)** — HEALTHY, cleaner hour, no pause. Only hourly book-digest ran (curate/vivint/
  playlist next at 12:00): 1 completed + 1 auto-advance hiccup (known reaper/advance class). **20 flexllama
  done, ZERO flexllama chat errors** this window. 7 opencode_go = judges. spend **$0.0000**, guard clean,
  in_flight 0, proposals 43 (no growth — vivint's 09:00 gather-fail added none; next vivint 12:00). No new modes.
- **12:24 UTC (tick 7, the 12:00 burst)** — HEALTHY, no pause. burst drove **49 flexllama done**; book-curate
  + **planning/vivint COMPLETED cleanly** (gather contention transient — fine this round); book-digest in_progress
  (critique). spend **$0.0000**, guard clean, proposals 45/70. **3rd KEY FINDING: playlist-digest `digest`
  FAILED — qwen3.6 'peg-native format' 500** (reasoning model breaks grammar/structured-output; wq 8678).
  → videos leg likely persistently fails on qwen-reason; route grammar stages to a non-thinking model. 14
  opencode_go = judges. Report fires 13:04.
- **13:24 UTC (REPORT, watch ENDED)** — soak ran ~7.5h on local, **$0.0000**. Totals: 244 flexllama chats
  done / 3 errored; 86 opencode_go errors = judges only. work_items: book-curate 4✓, book-digest 5✓/3✗,
  research-summary(ai-news) 3✓/0✗, planning(vivint) 2✓/1✗, playlist-digest(videos) 1✓/1✗. Quality HIGH
  (real cited AI digest + 5 philosophy book digests 5-20k chars). 3 findings = reaper-too-short / gemma
  contention / qwen-peg-grammar. Left RUNNING (not off-track), guard clean. Both watch crons deleted.
  Recommendation: codify local routing as overlay + the 3 fixes (currently live-only drift).
- **SESSION CLOSE** — parallelism CODIFIED (gemma q8 --parallel4) + doc-construction Phase 1 + playlist
  recast SHIPPED & e2e-proven on local; autonomy RESUMED (live on local, $0). Journal: OSS
  `.spec/journal/2026-06-19-local-soak-and-doc-construction.md`. Board updated. Watch ended.
- **13:xx UTC (post-report)** — Michael's reframe → 2 PLANS written (OSS `.spec/proposals/`):
  `agentic-doc-construction.md` (build artifact via tool-call diffs, chat=journal — addresses all 3
  findings) + `local-throughput-experiments.md` (parallelism). Now running the parallelism experiment;
  PAUSED autonomy for clean measurement (manual pause — reflect_resume to restart the digesters).
- **~13:5x UTC — PARALLELISM EXPERIMENT A done (results in local-throughput-experiments.md).** gemma
  --parallel 1 c2 = **WEDGED** (slot stuck 91% util, 600s timeout — concurrency HANGS the slot, worse than
  "contends"). gemma --parallel 2: c1=71 tok/s, **c2=155 agg (2×, no wedge)**, c4=153 (queues fine). Cost:
  single -25% (94→71), per-slot ctx 256k→131k (→ pairs w/ doc-construction page-in). **Verdict: adopt
  --parallel 2.** Rig is AT --parallel 2 now; autonomy still PAUSED; NOT codified (awaiting Michael's
  codify-vs-revert call + the gemma-window 131072 lockstep). throughput_test.py timeout 600→120.
- **~14:xx UTC — PARALLELISM CODIFIED + DOC-CONSTRUCTION PILOT: THESIS PROVEN.** (1) gemma q8
  --parallel 4 @524288 (4×131k, 248 tok/s) codified — live window 131072 + overlay + setup script,
  ws `b2eb206`. (2) Phase 1 doc-builder tools (`34-doc-builder.sql`: doc_create/append/patch/read/
  finalize over self-contained doc_drafts; chained in lib.rs) shipped + e2e-proven, OSS `6145d51`.
  (3) ★ THESIS PROVEN on the broken model: a tools-on build on **qwen** (which 500'd "peg-native
  format" one-shot) ran 8 flexllama chats / **0 errors / 0 peg failures** → built a coherent 2356-char
  digest via doc_* tool calls + a journal output. All 3 soak findings addressed by the one reframe.
  NEXT (wiring, thesis done): playlist_publish_draft bridge (handle, not body-arg) + recast playlist
  read→build + fresh-video live test. autonomy still PAUSED. Minor: inert doc-build-test pipeline left (FK-pinned).
