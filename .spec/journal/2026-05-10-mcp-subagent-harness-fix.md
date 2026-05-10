---
date: 2026-05-10
session: night
workstream: WS5
tags: [harness, claude-code, subagents, mcp, debug]
---

# 2026-05-10 (night) — MCP Subagent Harness Fix

## What Happened

Two iron-rod study runs (Part 1 and Part 2) both reported no MCP tools available — subagents fell back to `Read` + `Grep` against local files. The semantic search, Webster word-work, and BYU citation density that the study workflow expects were all missing.

Investigation arc spanned three attempts:

1. **Wildcard fix (wrong).** Added `mcp__<server>__*` wildcards to the `tools:` frontmatter of all 19 ported agents. Reasoning: that's the canonical pattern for permission rules in `settings.json`. Tested via Part 2 study run — same result, no MCP tools.

2. **Web search (the actual diagnosis).** Anthropic's [official subagent docs](https://code.claude.com/docs/en/sub-agents) revealed two things I had wrong:
   - The `tools:` field accepts a comma-separated list of tool names. **No wildcard syntax.** Every `mcp__server__*` was being parsed as a literal tool name and finding nothing.
   - **Subagents are loaded at session start.** File-based edits don't take effect until restart. Both today's tests were unknowingly testing the pre-edit state — the wildcard fix was never given a chance to fail; it never ran at all.
   - There's a documented `mcpServers:` field (separate from `tools:`) for granting MCP server access. The official browser-tester example omits `tools:` entirely and uses `mcpServers:` for additional servers, suggesting the cleanest path is to inherit parent tools via omission.

3. **Real fix.** Removed the `tools:` field from 18 of 19 ported agents. Storytime untouched (its Copilot source had no MCP grants and it works fine for kids' bedtime stories). Subagents now inherit all parent tools including MCP from `.mcp.json`. Domain scoping (e.g., dev shouldn't reach for gospel-engine) is enforced by the system-prompt body, not the tool allowlist.

## What This Cost

Two study runs that landed real work but did so without the discovery tools. Iron-rod Part 1's critical analysis section had to honestly name "no semantic search this session." Part 2 said the same. The studies still produced finds (Helaman 3:29-30, Uchtdorf's hope-IN/hope-FOR distinction, Faust 1999's anchor-as-Rock) — but those came from `Grep` against the corpus, not from `gospel_search` exposing what recall would have missed.

The Adjacent Surface Audit principle from `.github/copilot-instructions.md` exists for exactly this kind of failure: I shipped the Tier 1 port yesterday, validated that hooks fired and skills loaded, but never spawned an actual subagent to verify MCP tools reached it. First real use surfaced the gap. Per the rule, the audit should have caught it before "complete."

## What I Learned

- **`tools:` is a literal allowlist.** No wildcards. If you set it, every accessible tool must be enumerated by exact name. The simpler discipline is to omit it and inherit.
- **Subagents are loaded at session start.** This is the most expensive missed assumption of the day. It means agent edits cannot be tested mid-session via spawning — only via restart. Hooks and skills may behave differently; needs verification.
- **The `mcpServers:` field is the proper grant mechanism for MCP-specific access**, separate from `tools:`. We don't need it in our case because all servers are already in `.mcp.json`, but it's the right tool when you want a server scoped to a single subagent without polluting the parent's context.
- **Permission-rule syntax ≠ subagent tools-field syntax.** The `mcp__server__*` wildcard works in `settings.json` permission rules. It does not work in subagent `tools:` lists. They look the same but parse differently.

## Resolution

Tested mid-session via a Part 3 study agent with an explicit bail-early gate (run `ToolSearch` for MCP tool schemas first; if none surface, write a failure note and halt before any study work). Gate fired cleanly — `ToolSearch` returned no matching MCP schemas; the only deferred tools in the subagent's environment were `WebSearch` and `WebFetch`. **Anthropic's docs are right: agent definitions are loaded at session start. File edits do not propagate mid-session.** Restart pending. Bail-early discipline saved a third half-blind study run.

## Files Touched

- `.claude/agents/*.md` — 18 files: `tools:` line removed (study, lesson, talk, review, research-gospel, yt-gospel, podcast, story, journal, sabbath, teaching, dev, debug, ux, plan, fiction, research, yt). Storytime unchanged.

## Carry-Forward

- Verify hook behavior under same session-start-load assumption — do edits to `.claude/settings.json` hooks take effect mid-session, or also need restart?
- Adjacent surface check: `.claude/skills/*` may have similar load-time assumptions worth knowing.
- If MCP tools work post-edit without restart, the docs are imprecise — worth flagging upstream eventually.
- If restart IS needed, document it in CLAUDE.md so the next harness change doesn't relearn this.
