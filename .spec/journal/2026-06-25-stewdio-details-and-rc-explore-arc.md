# 2026-06-25 — Stewdio Details mode + the RC explore-repos arc (a long autonomous run)

One very long session in the **pg-ai-stewards** lane, Opus 4.8 at xhigh. It started
as a Stewdio polish ask and became a full security-reviewed feature arc, the back
half of it driven autonomously while Michael put kids to bed ("I know you are
capable of it"). Six shippable increments, all in the public OSS, all tested.

## The arc, in order

1. **Details mode** (`0c86d75`) — the "looks like nothing is happening, no thinking
   badge" fix + the pop-open details surface. A self-healing *working pulse*
   (driven off `chatSessionStatus`, names the current activity), a live
   *Activity / token-dispatch stream* (Models pane → "Activity"), and the
   `⚙ Dev` → `⚙ Details` relabel. The QA caught that my first pulse fix only
   *narrowed* the dead-air bug (forever → ≤15s) and that authoritative-`working`
   defeated `stop()`'s optimistic clear — both fixed, stop verified live.

2. **Zip-timeout diagnosis → import/clone arc ratified.** Michael's `pg-ai-stewards`
   repo zip "timed out." Traced (not guessed): `doc_import_corpus` ran ONE 180s
   synchronous extraction over the whole archive + embedded every file — wrong tool
   for a code repo. He ratified **RC-1** (explore public repos), **RC-2** (route a
   dropped code archive to explore, not embed), **RC-3** (lift the import cliff).

3. **RC-1** (`9924d9b`) — `/explore <url>`: clone a PUBLIC repo anonymously into the
   read-only sandbox, `research_codebase` answers, no DB embedding. The adversarial
   security QA caught **two real vulnerabilities**: a credential leak (`-c
   credential.helper=` doesn't clear a system/global helper → hermetic git env) and
   a token-exfiltration primitive (allow-list matched a substring anywhere in the
   URL → anchored host-rooted matching). Both fixed + oracle-covered.

4. **Activity feed shows non-LLM work** (`d562b5d`) — Michael's idea, and the lens
   for his own bug: the "Tools & sandboxes" pane shows `mcp_proxy` tool calls
   (doc-extract/ClamAV/coder) with status + duration. It immediately rendered
   `doc_extract → error → 120006ms` — the silent stall, made visible.

5. **RC-2** (detection `12c7512`, full `6901c3c`) — explore a *dropped* archive.
   The unpack reused the already-hardened `safeArchiveName`; I wrote a zip-slip
   oracle for the member names — and the adversarial QA found the **blocker** I'd
   missed: the *sandbox id* flowing into `rm -rf /worktrees/<id>`, where
   `sanitize("..")==".."` → `rm -rf /` on the bridge (pre-existing in `CloneRepo`
   too). Fixed in layers (sanitize hardening + `worktreeChildOK` strict-parent
   guard on every destructive op + tool-boundary id validation), plus HIGH
   reuse-skip + force-NetOff + MED structural scan/caps. The lesson is now a
   principle: *every untrusted input, not just the one you're thinking about.*

6. **RC-3** (`f61e780`) — the ~120s cliff was the bridge daemon's uniform
   `--call-timeout 120`. Added a `--slow-call-timeout 600` for an inherently-slow
   tool set + raised the converter/container to match. **Honest call:** I raised
   the cliff rather than rearchitect into an async job, because RC-2 already routes
   the common painful case (a code repo) to extract-free exploration; full async
   would be over-engineering across the Go/Rust boundary for a rare case.

**Deployed + verified live on the dev stack:** grant applied to pg, bridge
rebuilt + `refresh-tools`'d, schemas confirmed (`attachment_id` live, daemon logs
`slow-call-timeout=600s`), coder-runtime + doc-extract images present.

## What this session taught (kept)

- **The loop is the ceiling under context depth.** Michael couldn't compact mid-run,
  so the build → oracle → *adversarial QA* → verify-the-findings → fix → commit loop
  did the compensating. It caught three real security holes (two RC-1 vulns + the
  RC-2 `rm -rf /`) that green functional oracles all missed. Without it, an
  autonomous run at this depth would have shipped a catastrophe.
- **Verify the QA's findings, don't just trust them** (the running discipline): a
  grep caught that Digest is shelf-driven where the QA missed it; reading each
  finding caught which were real vs. plausible.
- **Stewardship caught a pre-existing bug:** the `rm -rf /` exposure lived in
  `CloneRepo` before RC-2 — fixed the sibling with the same guard, reported.

## Carry-forward (for "what's next")

- **His hands:** a model-driven e2e — drop the repo zip → it explores; or `/explore
  <url>`. Logic is oracle+QA-covered, plumbing verified live; I left the model run.
- **Deferred, named:** async corpus import (the true F1, if big DOC corpora become
  common — the slow-tool plumbing + activity feed are its foundation); RC-2 unpack
  *warnings* aren't surfaced (LOW nit); prune the dead opencode free-tier models
  from the overlay seeds (the engine self-heals via failover, just noise).
- **Stewdio focusing-release** S6–S9 (#253) were superseded by this arc and remain
  open if he wants them. The lane `.mind/sessions/pg-ai-stewards.md` has the full
  blow-by-blow; OSS journals `2026-06-2{4,5}-*` per increment.

Tasks #258/#259/#260/#261 ✅. It was good.
