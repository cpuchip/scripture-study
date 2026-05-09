---
date: 2026-05-09
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3e, mcp, fetch-md, study-export, qol, autonomous]
priority: medium
carry_forward:
  - gospel-engine-v2 submodule has the same JSON-RPC notification fix in working tree but uncommitted — submodule pointer bump is its own session's work
  - fetch-md v2 with JS rendering (chromedp / playwright shell-out) when an agent first hits a SPA that returns sparse content
  - 3e.5 gospel_passthrough — wraps an outbound bridge call into an inbound stewards-mcp tool. Easy now that the bridge is real
  - 3f a.ibeco.me web UI design pass — Michael highlighted this as the next direction; producer side is now real
  - run another study to verify bridge stability under multi-tool agent flows (mysteries used gospel_get heavily; webster_define and other tools haven't been organically exercised yet)
  - cross-server tool name collisions still possible — none today, but a future MCP addition could collide on bare names
---

# 3e QoL sweep + first organic bridge use + autonomous queue

A long, productive run that closed the three carry-forwards from 3e.2.b/c
plus shipped two new pieces, then ran an autonomous queue while Michael
was out.

## The big arc

The most meaningful moment was the **mysteries-of-god study**. Michael
gave a real binding question — Nephi's claim to "a great knowledge of
the goodness and the mysteries of God" (1 Ne 1:1) — asking whether
mysteries are obtainable from the text alone or only through the
Spirit. We dispatched a `study-write` work_item on kimi-k2.6 and
watched the bridge log fill with `gospel_get` calls during the
review-stage self-check. The agent verified its quotations against the
corpus through the bridge, organically. That validates everything
3e.2.b/c built — the async fan-out, the completion pass, the bridge
daemon. Real work flowing through real architecture.

The output (16K chars, three-section study) auto-promoted into
`stewards.studies` via the 3c.3.5 trigger and dumped to
`study/mysteries-of-god-text-vs-spirit.md`. The agent emitted
`[slug](#)` placeholders for cross-references — it knew the sibling
studies existed in the corpus but didn't have file paths. I resolved
the 10 placeholders by hand for the first study, then built
`scripts/study-export/` to automate it for next time. That tool
indexes /study/ recursively, supports nested files via a hyphen-
flattened alias key, and resolves scripture refs ("Alma 40:3" →
gospel-library chapter path) using a 100-entry book map covering
OT/NT/BoM/D&C/PGP.

In parallel I built **fetch-md-mcp** as a new Go MCP server (Mozilla
Readability + JohannesKaufmann/html-to-markdown). Four tools:
fetch_url, fetch_urls, extract_links, fetch_url_raw. Plain HTTP only.
Granted to the research agent. The bridge cached its 4 tools, the
auto-promote trigger created tool_defs, and a synthetic test fetched
a Wikipedia page in 258ms. Eighth MCP server in the registry.

## What surprised

1. **The mysteries study's first reference was self-aware about the
   research limit.** The agent's section II opens with: "the corpus
   suggests something different. The mysteries are not esoteric facts
   about God. They are the knowledge of God himself." That's not
   confabulation — it's an actual synthesis from the corpus the agent
   read. When you watch a substrate agent do the work the
   source-verification skill demands, it changes how you think about
   what these tools are for.

2. **The "REVIEW: revised" stage marker bled through.** First time I
   read the study body in psql, line one was `REVIEW: revised\n\n# What
   Are the Mysteries...`. That's the agent's own provenance signal
   from the review stage. For the substrate row, useful. For the
   `/study/` export, noise. study-export now strips it via regex.

3. **psql `:variable` doesn't enter `DO $$ ... $$` blocks.** Already
   noted in 3e.2.b/c findings. Hit it again on verify-3e2-2.sql and
   immediately remembered. Memory wins.

4. **search-mcp was missing the `notifications/initialized` case
   AND sending error responses to notifications.** Two bugs in one
   server. The first (missing case) is what made bridge refresh fail
   on it specifically. The second (JSON-RPC violation) was lurking in
   five hand-rolled servers. Audit caught it as a class.

5. **Go's stdlib flag package stops at first positional.** First
   study-export run with `--out /tmp/file.md` after the slug positional
   silently dropped the flag. The tool defaulted to writing the
   workspace `study/` path, which happened to be the file I'd already
   committed, so the only diff was a trailing newline. Phew. Fix:
   permissive parsing — pull positional explicitly, then `flag.Parse`
   on the rest.

6. **gospel-engine-v2 is a git submodule.** Made the same JSON-RPC
   fix in its working tree but couldn't commit it from the parent —
   submodule history belongs in its own repo. Rebuilt the binary with
   `GOWORK=off go build` so the running gospel-mcp.exe carries the
   fix even though the submodule pointer is unchanged.

## Stewardship moments

- **Audit class found.** When search-mcp's JSON-RPC bug surfaced,
  boundary-tested whether to audit the same shape elsewhere. Five
  servers had the secondary-defense gap. Michael ratified the audit
  task; fixed all five, smoke-verified the two active ones, deferred
  the legacy three (build-skipped, source patched).

- **Submodule pointer bump deferred.** Instead of unilaterally
  bumping the gospel-engine-v2 submodule pointer in the parent (which
  would mean committing in the submodule's own repo first — a hard-
  to-reverse cross-repo action), left the working-tree edit in place
  and noted in commit message + this journal. Michael can sync the
  submodule on his terms.

- **Write-mutating tools left deny-by-default.** Even after broadening
  grants, kept `brain_create/update/delete`, `create_*`, `log_practice`,
  `review_card` off the substrate-agent surface. Substrate agents
  remain read-only on personal data. Boundary test: would Michael, if
  asked, want a kimi-k2.6 study run to be able to write to his brain
  without explicit oversight? No. So the writes stay gated.

## What this sets up

The substrate now has both directions of MCP, organic agent traffic
through the bridge, broad read-only grants across 7 agent families,
auto-promote keeping the tool surface in lockstep with the cache,
session crash recovery, and an export tool to publish substrate work
to the workspace cleanly. The producer side is meaningfully real.

The natural next moves Michael flagged: 3f web UI design pass (now
that producer is real), or 3d sandboxed git (he highlighted the line).
Both are design-heavy and warrant his judgment.

## Time

Roughly 4 hours across two halves: the morning 3e.2.b/c → 3e.2.d/e
build, then an afternoon QoL sweep + autonomous queue. Each commit
land-and-verify cycle stayed under 30 minutes thanks to the pgrx-rust
skill, the mcp-server-go skill, and the well-shaped bridge daemon.
The mcp_proxy round-trip latency (~258ms after session warmup) means
substrate agents using the bridge feel responsive enough that long
verify-and-revise loops are now practical.
