---
date: 2026-05-29
title: Public-OSS documentation pass (standalone-repo readiness, docs layer)
status: shipped — docs committed; build-extraction specced, not executed
---

# Public-OSS docs pass

## Ask + ratified scope

Michael: "update all of our documentation/readmes and guides for pg-ai-stewards
— get it ready to turn into a real standalone repo." Ratified via AskUserQuestion:
**public OSS**, **docs-only now** (build-decoupling specced separately),
**start now**.

## Inventory finding that reframed the task

"Standalone" is two layers, not one:
1. **Docs** (the ask) — and the existing docs were badly stale: `README.md` was
   the *pre-build research verdict* ("not yet a build project"); architecture.md
   dated 2026-05-06 (23 tables/67 fns vs. current 65/263).
2. **Build coupling** (the surprise) — it isn't independently buildable: the
   Dockerfiles build with `context: ../../..` and pull sibling modules via the
   shared `go.work`. Documentation can't fix that.

Key extraction finding: the Go modules have **no sibling source deps** (no
replace/require) — they just coexist in the parent go.work. So extraction is
plumbing (own go.work, self-contained Docker context, pluggable bundled MCP
servers, vendor the UI in), not code rewrites. Specced in
`.spec/proposals/standalone-extraction.md`.

## Shipped (docs layer)

- **README.md** rewritten for an outside reader (concept, capability table,
  architecture-at-a-glance, providers, quickstart pointer, gospel-study as ONE
  example, honest standalone-status note). Old research README preserved at
  `docs/history/2026-05-02-research-verdict.md`.
- **docs/architecture.md** state line refreshed to current counts.
- **LICENSE** (MIT), **.gitignore**, **CONTRIBUTING.md**, **QUICKSTART.md**.
- **docs/history/README.md** — provenance index.
- **.spec/proposals/standalone-extraction.md** — the build-decoupling spec.

## Not done (deliberately)

- Build extraction itself (specced, ratified as separate later work).
- Genericizing `.spec/` history content — left as an E.5 decision in the
  extraction proposal (leaning: keep, labeled internal provenance).
- A full table-by-table rewrite of architecture.md's body (only the headline
  state was stale; the map structure still holds — flagged growth inline).

## Carry-forward

- Execute `.spec/proposals/standalone-extraction.md` when ready to actually cut
  the repo (E.1 vendor UI → E.2 own go.work + contexts → E.3 pluggable MCP →
  E.4 optional corpus → E.5 subtree-split + CI).
- Add the CI check the proposal names: diff `extension_sql_file!` refs in
  lib.rs against the Dockerfile COPY list (the am1 gap that broke the J.11
  rebuild — would catch the next one).
