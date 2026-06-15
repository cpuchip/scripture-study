# Reflect-Steward Check-in

How to run a reflect-steward check-in for Michael — review what the autonomous
steward proposed, steer it, and use the kill switch. The review surface is
**Claude Code (me)**; these are the verbs I drive via `psql` against the OSS
stack. Load this when Michael says "check in on the steward," "reflect status,"
"what's the steward proposing," or asks to approve/decline/pause anything.

## Where it runs

The reflect-steward lives on the OSS stack (`stewards-oss-pg`). All verbs are SQL
functions in `stewards.*` (shipped in `22-reflect-steward.sql`). Invoke with:

```bash
docker exec -i -e PGUSER=stewards -e PGDATABASE=stewards stewards-oss-pg psql -At -c "SELECT ..."
```

## A check-in, step by step

1. **Status** — the dashboard:
   ```sql
   SELECT jsonb_pretty(stewards.reflect_status());
   ```
   Shows: `autonomy_paused`, `max_concurrent`, `in_flight`, `approved_waiting`,
   `proposals_pending`, `intents_paused`, `recent_reflect_runs`.
2. **The queue** — what it wants to do:
   ```sql
   SELECT slug, intent, pipeline, approved, binding_question FROM stewards.reflect_proposals();
   ```
   Present these to Michael in a readable list. **Recommend** which to approve,
   but the yes/no is his.
3. **Act on his word** (these are bin-3/4 — only on his explicit say-so):
   - Approve (queues it; the drain dispatches as capacity allows — does NOT fire it now):
     `SELECT stewards.reflect_approve('<slug>');`
   - Decline (cancels the proposal): `SELECT stewards.reflect_decline('<slug>', '<why>');`
   - Steer (note shapes the intent's next cycle): `SELECT stewards.reflect_steer('<intent>', '<note>');`

## The kill switch (use freely — it's reversible)

- **Stop everything now:** `SELECT stewards.reflect_pause('<why>');` → no new
  scheduled cycles, no drain dispatches. In-flight stages still finish (for those,
  use the emergency-stop brakes). Resume: `SELECT stewards.reflect_resume();`
- **Decommission one runaway intent** (the rest keep running):
  `SELECT stewards.reflect_pause_intent('<intent>', '<why>');` →
  `reflect_resume_intent('<intent>')` to lift.

## Capacity

Approved proposals dispatch only as in-flight work drops below
`reflect_max_concurrent` (default 2) — so a big approved batch never floods the
workers. Adjust: `SELECT stewards.config_set('reflect_max_concurrent','3'::jsonb, NULL);`

## The watchman wake (only after go-live)

Once Michael flips a schedule on (`UPDATE stewards.scheduled_pipelines SET
enabled=true WHERE slug='...'`), I self-schedule a wake every **2–5h**
(`ScheduleWakeup`) to:
1. `reflect_status()` — check `in_flight`, `approved_waiting`, spend, recent runs.
2. If anything looks runaway, drifting, or over-budget → `reflect_pause('watchman: <reason>')`
   and surface it to Michael at his next engagement.
3. Otherwise log a one-line "all nominal" and reschedule.

Keep the wake **cheap** — a status read + judgment, not deep work. The wake is
the human's proxy watch *between* Michael's own check-ins, per the presiding
covenant (Michael → Claude-watchman → reflect-steward → doers).

## What I never do without him

Flip a schedule from disabled→enabled (go-live), promote a proposal to real spend
beyond approving it for the capacity-gated drain, or let an intent run that he
hasn't blessed. Approve/decline/steer act only on his explicit instruction.
