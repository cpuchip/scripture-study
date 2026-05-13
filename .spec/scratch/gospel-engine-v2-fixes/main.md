# Scratch: gospel-engine-v2 fixes — findings

*Started: 2026-05-13. Companion to `.spec/proposals/gospel-engine-v2-fixes.md`.*

## Binding problem

When the agent reaches for `gospel_*` tools during study sessions, it
hits a small set of repeating, papercut-shaped failures. Some are real
bugs (verse-range, indexer field corruption); some are documentation
drift (mode names, param names); some are missing surfaces that have
been "Open / High priority" in `docs/06_tool-use-observance.md` since
March 31 and never shipped. Each one individually is small. Together
they push the agent off the tool and onto `read_file` + `grep_search`
fallbacks, which is what the gospel-engine was supposed to replace.

## Source: prior journal entries

### `2026-04-20 give-away-all-my-sins-study.yaml`

> "Gospel MCPs and Webster define were all disabled… every gospel-engine
> and gospel-vec tool returned 'Tool currently disabled by the user.'
> Worked around with `grep_search` over `gospel-library/` with
> `includeIgnoredFiles: true`."

Rooted in: chatmode `tools:` allowlist using `gospel/*` prefix that
didn't match actual tool names. Carry-forward: "verify gospel MCPs
actually work in study mode."

**Status today (2026-05-13):** the agent allowlist patterns are
`'gospel-engine-v2/*'` (verified in `plan.agent.md`). The deferred
tool list in this session loaded `mcp_gospel-engine_*` tools and they
respond. **The April-20 allowlist symptom appears resolved.** What
remains is a different cluster of issues.

### `docs/06_tool-use-observance.md` (Mar 31 + Apr 20 entries)

Three items still listed as **High / Open**:

| # | Issue | Status |
|---|-------|--------|
| 1 | Verse-level retrieval in `gospel_get` (range support) | Open since Mar 31 |
| 2 | Cross-reference retrieval (85,590 edges, no MCP surface) | Open since Mar 31 |
| 3 | Agent instruction for `includeIgnoredFiles: true` on grep | Open |

Item #3 is now in `.github/copilot-instructions.md` ("Gospel Library
is gitignored… always pass `includeIgnoredFiles: true`"). Items #1
and #2 are still open.

## Live reproduction (2026-05-13 session)

Tested `mcp_gospel-engine_*` against the running engine.

### A. Mode enum drift — REPRODUCED

```text
mcp_gospel-engine_gospel_search { mode: "combined", ... }
→ ERROR: Your input to the tool was invalid (must be equal to one of the allowed values)
```

Server enum (`scripts/gospel-engine-v2/cmd/gospel-mcp/main.go` line ~280):
`["keyword", "semantic", "hybrid"]`.

Documented enum in `.github/copilot-instructions.md`:
> "`gospel_search` now does both via `mode: "keyword" | "semantic" | "combined"`."

And in `CLAUDE.md` (Claude Code addendum) likely same. The docs lie.

### B. `reference` vs `ref` name leak — REPRODUCED

```text
mcp_gospel-engine_gospel_get { ref: "D&C 93:24-30" }
→ ERROR: HTTP 400: provide either ref= or (type= and id=)
```

The MCP schema accepts `reference`, but when a wrong call comes in,
the server's HTTP-level error message mentions `ref=`. The agent
reasonably assumes the error names the actual param and retries with
`ref:`, which the MCP layer silently drops. Two passes lost to the
mismatch.

### C. Verse-range — REPRODUCED, still broken

```text
mcp_gospel-engine_gospel_get { reference: "D&C 93:24-30" }
→ Error: HTTP 404: not found
mcp_gospel-engine_gospel_get { reference: "D&C 93:24" }
→ OK
```

`getByReference` does `WHERE reference = $1` exact match against the
`scriptures` table. No range parsing. To get 7 verses requires 7
calls — the same 3-31 regression. Old gospel-mcp (Feb 15) had the fix:
`parseReference` split on `-`, then `WHERE verse >= ? AND verse <= ?`.

### D. Speaker field corrupted — REPRODUCED

```json
{
  "title": "\"I Will Give Away All My Sins to Know Thee\"",
  "speaker": "🎧 Listen to Audio",
  "year": 2026,
  "month": "04"
}
```

The indexer is grabbing the audio-link button text instead of the
speaker line. Affects credibility of search results — "Elder Wu" gets
labeled with a speaker icon string. Bug in
`scripts/gospel-engine-v2/internal/indexer/` (not yet read).

### E. No chapter-level fetch — REPRODUCED

```text
mcp_gospel-engine_gospel_get { book: "dc", chapter: 93, volume: "dc-testament" }
→ HTTP 400: provide either ref= or (type= and id=)
```

The schema only accepts `reference` (single verse) or `type+id` (DB
PK). To get a whole chapter the agent must `read_file` the markdown
or call `gospel_get` once per verse. May be intentional (per the
"`read_file` is for understanding, MCP is for discovery" pattern from
the Mar 31 entry), but the schema doesn't say so and the error
message gives no guidance.

### F. Cross-references not exposed — UNCHANGED

The DB has `cross_references` (85,590 rows) and `edges` tables. No
MCP tool returns them. To trace footnotes the agent reads the markdown
file directly.

## Other code observations

- `gospel_get` handler at `internal/api/server.go:153`. Single function,
  small surface — verse-range fix is ~15 lines.
- The MCP wrapper (`cmd/gospel-mcp/main.go:178`) and the HTTP server
  (`internal/api/server.go:54`) are both small. Param renaming + range
  parsing localized to those two files.
- `getByReference` returns a flat object. To return ranges, either
  return an array or wrap in `{verses: [...]}`. Need to check what
  callers expect.

## What's NOT a problem

- **Tool allowlist:** `'gospel-engine-v2/*'` pattern matches the
  loaded `mcp_gospel-engine_*` deferred tools in this session. The
  Apr-20 symptom is gone.
- **Auth:** ibeco.me-issued bearer token works against the hosted
  engine.
- **`gospel_search` mode `hybrid`:** works correctly, returns
  Wu's talk + Psalms 32:5 + Benson 1983 for the king's prayer query.
  Real semantic capacity is intact.

## Pointer

Findings folded into the existing proposal (which already covered
verse-range, cross-refs, and `includeIgnoredFiles`):
[`.spec/proposals/gospel-engine/phase1.5-ergonomics.md`](../../proposals/gospel-engine/phase1.5-ergonomics.md)
\u2014 see "Enhancements 4\u20136 (added 2026-05-13)".

This scratch is the live-reproduction evidence for those three new
enhancements (mode-enum drift, `ref`/`reference` leak, speaker
indexer bug).
