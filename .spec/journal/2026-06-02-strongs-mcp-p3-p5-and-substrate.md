# Strong's Concordance MCP — P3–P5 + substrate deploy (v1 complete)

**Date:** 2026-06-02
**Workstream:** WS-tools (canon-walk toolchain)
**Mode:** dev
**Outcome:** v1 complete, live in both Claude Code and pg-ai-stewards.

## What was done

Resumed after a compaction with the explicit instruction "do p3/p4 so you're a
bit fresher." Read the resume runbook (`docs/next-steps.md`) and executed P3,
then surfaced P4 for ratification, then drove the full deploy + P5.

- **P3 `strongs_for_verse`** — `data/kjv-strongs.json.gz`, 31,102 verses / 66
  books. Build-data parser + `internal/concordance/verses.go` (ForVerse +
  ResolveRef with full book-name/abbrev/alias table) + the MCP tool (enriches
  each word's Strong's number with lemma + gloss from the lexicon).
- **P4** — registered in `.mcp.json` (Claude Code) and the substrate
  (`bridge.Dockerfile` build entry + `strongs1-mcp-seed.sql` migration + grants).
- **P5** — README usage + `data/ATTRIBUTION.md`.

Commits: strongs repo `6cb5c4d` (P3) → `02f44ba` (P5); main repo `cec7e36`
(substrate deploy artifacts).

## Surprises / discoveries

1. **The kaiserlik data is a mess, and the runbook's "parse via verse keys" was
   the right instinct for the wrong reason.** Not just an inconsistent top key —
   per-book *filenames don't match their contents*: `1Ch.json` is 6.8 MB and
   contains 1 Chronicles repeated ~16× (15,763 entries for a 942-verse book);
   7 of 67 files fail strict JSON parse because `bg`/`ch`/`sp` (Bulgarian /
   Chinese / Spanish) carry unescaped quotes; one file concatenates a second
   book. The fix: ignore filenames, regex-extract every `"ABBR|ch|vs":{"en":…}`
   pair across all files (the `en` field is clean + always first), and **dedup by
   verse key**. 102,140 raw pairs → 31,102 unique, **0 text conflicts** — every
   duplicate carried identical text. Per-book counts then matched KJV exactly.

2. **Markup hid behind clean sample verses.** Gen 1:1 / John 1:1 / John 3:16 are
   markup-free, so the first validation gate passed — but the moment I
   smoke-tested **Ps 23:1** (inverse-hypothesis / Agans 9, testing a verse I
   *didn't* hand-pick), `<em>is</em>` italics and `[[A Psalm of David.]]`
   superscription brackets leaked through (`David.]]` as a malformed word). KJV
   italics + Psalm superscriptions + `[fn]` footnotes pervade the OT/NT. Fixed
   the parser to strip all three, and — the real gap-closer — added a **global
   no-markup assertion across all 31,102 verses** so this class can never ship
   silently again. The lesson: a validation gate built only from clean,
   hand-picked examples validates nothing about the messy 99%.

3. **The runbook's P4 plan was stale, and the staleness mattered.** It assumed
   the bridge runs on the Windows host (drop a registry row → `.exe`). Ground
   truth: the bridge migrated into Docker on 2026-05-09 (`3e2-6`), and
   `mcp_servers.command` points at Linux binaries **baked into the bridge image**.
   So registration meant editing `bridge.Dockerfile`, rebuilding + restarting a
   live container mid-soak, and a ledger migration — a live deploy of the
   beyond-competence substrate, not an insert. That's why I stopped and surfaced
   (bin-3) with the corrected plan + verified Linux cross-compile before touching
   anything. Michael chose "go — I drive it now."

4. **The deploy landmine I'm most glad I caught.** The bridge entrypoint runs
   `set -e; stewards-cli migrate`, and migrate **exits 2 whenever it applies a
   migration while drift warnings exist**. Four pre-existing drift files
   (`4a-cost-tracking`, `h1-1-general-research-intent`, `j10`, `j11` — SQL edited
   in place after being recorded) mean *any* new migration applied via the
   entrypoint would exit non-zero → `set -e` → **bridge fails to start**. I
   pre-applied the migration via `stewards-cli migrate` before the restart, so
   the restart found nothing pending → "substrate is current" → exit 0 → clean
   start. Verified the steady-state exit code empirically before trusting it.
   This is real substrate fragility, not just a strongs quirk.

## Verification (the real path)

- Linux binary runs in the rebuilt image (`docker run --entrypoint … -stats` →
  19,570 lexicon + 31,102 verses).
- `refresh-tools` spawned strongs in-container and cataloged 3 tools.
- Auto-promotion: `mcp_tool_cache` (3) → `tool_defs` (3, active).
- `compose_tools('study')` array **contains** strongs_define / strongs_search /
  strongs_for_verse — the actual function that decides an agent's tools. (A first
  query falsely showed "0 rows" because I treated the scalar JSONB return as a
  table; digging in rather than trusting the negative result surfaced the truth —
  the tools are there.)

## Carry-forward

- **Confirm** the `.mcp.json` strongs tools appear on Michael's next Claude Code
  restart (first-run approval).
- **4 drift files** (`4a`/`h1-1`/`j10`/`j11`) want reconciliation — any future
  ledger migration applied via the bridge entrypoint hits the same exit-2 landmine.
- webster / byu-citations / becoming are **not** in the substrate `mcp_servers`
  registry currently (9 servers total) — strongs grants are correct regardless,
  but the "mirror the webster grants" framing doesn't literally apply there.
- H7462's STEPBible primary gloss reads oddly ("House of Shepherds" for *râʻâh*)
  — a first-row-wins quirk for a few entries; documented as "glosses are a
  starting point, not doctrine." Consider a v2 sense-preference pass.
- v2+: `strongs_occurrences` (every verse a lemma appears in).

## Relational note

Picked up cold after compaction and carried a multi-phase build to completion
without begging off (Ammon). The one genuine stop — the stale P4 substrate plan
— was the right place to surface (Michael had asked to be looped on integration
decisions), and he said go. The rest was stewardship: drive it, verify each
step via the real path, report, don't ask permission for the obviously-ratified.
