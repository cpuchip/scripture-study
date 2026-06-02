# Proposal — Strong's Concordance MCP (Hebrew/Greek word-work for the Bible walks)

**Status:** ✅ RATIFIED + BUILDING (2026-06-02). **Source of truth is now the repo** — `projects/strongs-concordance-mcp/` (`github.com/cpuchip/strongs-concordance-mcp`), see its `README.md` + `docs/data-formats.md`. **P0–P2 shipped** (dual-lexicon pipeline + `strongs_define`/`strongs_search`, 19,570 entries, smoke-verified). **Remaining: P3** (`strongs_for_verse` via validated kaiserlik tagging) **+ P4** (register in `.mcp.json` + `stewards.mcp_servers`). Ratified options: dual lexicon (Strong's 1890 + STEPBible) · `for_verse` in v1 via validated KJV tagging · full build grant ([[feedback_strongs_mcp_stewardship]]). This proposal is kept as the origin record; do not duplicate the repo's spec here.
**Raised:** 2026-06-02, at the close of the Book of Mormon walk, while scoping the canon-walk series (PoGP → D&C → OT → NT). See `study/bom-walk/_workflow.md` → "the canon-walk series."
**Architectural twin:** `scripts/webster-mcp` (Go MCP server + bundled public-domain `data/`). Build pattern: the `mcp-server-go` skill. **Verified:** one stdio binary serves both Claude Code (`.mcp.json`) and pg-ai-stewards (`stewards.mcp_servers` connector, proxied by the bridge) — same as webster, no new substrate schema.

## The need (in one sentence)

For the OT/NT walks, make Hebrew/Greek word-work on the KJV as rich as `webster-mcp` makes 1828-English word-work on the Restoration text — so a chapter note can trace a load-bearing KJV word back to its original-language lemma and sense, the same way we trace "charity" or "virtue" through Webster 1828 today.

## Why it matters for the walk

The BoM walk leaned on `webster_define` exactly where the 1828 sense diverged from the modern one, and those were some of the richest moments (the "Therefore/But" word-work). The KJV has the same trap, deeper: the English is 1611, and underneath it is Hebrew and Greek that the English word often flattens (e.g. the four Greek loves behind one English "love"; *chesed* behind "lovingkindness/mercy"; *nephesh* behind "soul"). Without original-language access, the Bible walk's word-work would be **thinner** than the BoM walk's, not richer — backwards from the goal. Strong's is the standard bridge: it keys every KJV word to a numbered Hebrew/Greek lemma + gloss, usable without knowing the languages.

## Shape (mirrors webster-mcp)

A Go MCP server, `scripts/strongs-mcp`, registered in `.mcp.json`. Bundled public-domain data in `data/`. Candidate tools (final names TBD, parallel to `webster_define`):

- `strongs_define` — given a Strong's number (`H7225`, `G26`) → lemma, transliteration, part of speech, definition/gloss, and the KJV words it's translated as.
- `strongs_for_verse` — given a verse reference (e.g. "Genesis 1:1") → the word-by-word KJV→Strong's tagging, so a chapter note can see which words carry which lemmas.
- `strongs_search` — given a KJV English word → the Strong's number(s) behind it across occurrences (the reverse lookup).
- (stretch) `strongs_occurrences` — every verse where a lemma appears, for tracing a word across the canon.

## Data — sourcing is a research step (verify before asserting)

Strong's Concordance (James Strong, 1890) is **public domain**. Several open, machine-readable derivatives exist that pair Strong's numbers with a KJV word-tagging and Hebrew/Greek lexicons. **Do not assume a specific dataset is correct or complete — the build's first phase is to identify, license-check, and validate the source.** Known *candidates to evaluate* (confirm license + accuracy at build time, do not cite as settled):

- OpenScriptures Hebrew Bible (OSHB) + the OpenScriptures Greek tagging / morphology data.
- STEPBible open-licensed lexicon + tagging data (TAHOT / TAGNT).
- The various public-domain Strong's dictionary JSON/XML dumps (e.g. the openscriptures "strongs" repos).

Validation gate (same spirit as read-before-quoting): spot-check the dataset's gloss + lemma for a handful of known verses against a trusted reference before trusting it wholesale. A wrong concordance would silently corrupt every word-note built on it.

## Scope guardrails

- **Public-domain / open-license only.** No scraping a proprietary lexicon (BDB-full, Thayer's-with-modern-edits, etc. — confirm edition is PD).
- **Bundled + offline**, like webster-mcp — no live API dependency.
- **It's a reference index, not a translator.** It surfaces the original-language data; the *interpretation* stays mine-as-draft / Michael's-as-ratifier, same bin-1/2 frame as the walk. Strong's glosses are a starting point, not doctrine, and Strong's itself has known limitations (it glosses, it doesn't exegete) — note that in the README so future word-work doesn't over-trust a one-word gloss.

## Effort / sequencing

- Lift the webster-mcp skeleton (Go, stdio MCP, bundled `data/`, `internal/` loader). The novel work is the **data pipeline** (acquire → normalize → validate the KJV↔Strong's↔lexicon join), not the server plumbing.
- Sequence: **after the D&C walk, before the OT walk.** The PoGP and D&C walks don't need it (Restoration text; `webster-mcp` + `gospel_*` suffice).
- Build under the `dev` discipline + `mcp-server-go` skill; verify via the inverse-hypothesis (a wrong lookup should be reproducible and caught, not silently shipped).

## Open questions for Michael

1. Depth target for the OT/NT walks given the scale (~1,189 Bible chapters vs. 239) — does mini-study-per-chapter survive, or does the Bible walk shift to a lighter per-chapter note + heavier spin-offs? (Decide at that walk's planning; it shapes how heavily the concordance gets used.)
2. KJV-only, or also surface where the JST (already in `gospel-library/.../jst/`) revises the verse being word-studied?
