# Scripture Study Project — Claude Code

The canonical project instructions are shared with Copilot. Read them first, then the Claude-Code-specific addendum below.

@.github/copilot-instructions.md

---

## Claude Code Addendum

This file is loaded automatically by Claude Code on session start. Everything above (via the `@`-import) applies equally to Claude Code and Copilot. The notes in this section override or extend the shared instructions when the two harnesses diverge.

### Model & effort (Claude Code on Opus 4.8)

Claude Code runs **Claude Opus 4.8** (`claude-opus-4-8`) as of 2026-05-29. 4.8 builds on 4.7; the shared file's Foresight & Adjacent Surfaces tuning was written for 4.7 and applies to 4.8 unchanged.

**Effort is the main dial.** 4.8 is more effort-sensitive than any prior Opus, and the levels were recalibrated (`xhigh` = substantially more thinking than 4.7's `xhigh`; `medium` somewhat more, `high` somewhat less). Default is `high`. Michael's standing default is **`xhigh`** (set via `/model`), right for the dev / substrate / agentic work that dominates recent sessions. For a pure prose or study session, `high` is often better — it avoids overthinking a paragraph and keeps latency down. This is a per-session dial Michael owns: don't assume a level, and if a study session feels over-deliberated, that's the signal to suggest dropping to `high`.

### Tool naming differs from Copilot

The shared instructions list MCP tools using Copilot's deferred-tool naming (`mcp_gospel-engine_gospel_search`, `mcp_webster_webster_define`, etc.). **In Claude Code, use the standard MCP naming convention:** `mcp__<server>__<tool>`.

| Need | Claude Code tool name |
|------|----------------------|
| Search scriptures/talks | `mcp__gospel-engine-v2__gospel_search` |
| Get a scripture/talk | `mcp__gospel-engine-v2__gospel_get` |
| Browse content | `mcp__gospel-engine-v2__gospel_list` |
| Webster 1828 + modern | `mcp__webster__define` (or `webster_define`) |
| Web search (Exa) | `mcp__exa-search__web_search_exa` |
| Web search (DuckDuckGo) | `mcp__search__web_search` |
| YouTube | `mcp__yt__yt_download`, `mcp__yt__yt_get`, `mcp__yt__yt_list`, `mcp__yt__yt_search` |
| BYU citations | `mcp__byu-citations__byu_citations` (also `_bulk`, `_books`) |
| Brain entries | `mcp__becoming__brain_search` (also `_recent`, `_get`, `_create`, `_update`, `_delete`, `_stats`, `_tags`) |
| Practices/daily | `mcp__becoming__get_today` (also `_log_practice`, `_get_due_cards`, `_review_card`) |

The server names in `.mcp.json` are authoritative — Claude Code does **not** strip suffixes like `-v2` the way VS Code does. The full server name is part of the tool name. (Copilot's stripping of `-v2` is the source of the "trips us up repeatedly" gotcha in the shared instructions; that gotcha does not apply here.)

### Tool name mapping for shared instruction text

When the shared instructions or agent files reference Copilot tool names, mentally translate:

- `read_file` → `Read`
- `grep_search` → `Grep`
- `file_search` / `list_dir` → `Glob` / `Bash` (`ls`)
- `gospel_search` → `mcp__gospel-engine-v2__gospel_search`
- `webster_define` → `mcp__webster__define`
- `web_search_exa` → `mcp__exa-search__web_search_exa`
- `tool_search_tool_regex` → `ToolSearch` (Claude Code's deferred-tool loader)

### Gospel library is gitignored — same rule, different flag

Copilot's instruction is "pass `includeIgnoredFiles: true`." In Claude Code, the equivalent is: when using `Glob` or `Grep` on `gospel-library/` content, the tools respect `.gitignore` by default. Use `Bash` with `ls`/`rg` directly when you need to read inside ignored paths, or read the file by its known path with `Read`.

### Subagents in place of Copilot's "agent dropdown"

Copilot has a chat dropdown of custom agents (`.github/agents/*.agent.md`). Claude Code's equivalent is the **subagent system** — files in `.claude/agents/*.md` invoked via the `Agent` tool with `subagent_type: <name>`. The agent body is reusable; the frontmatter schema is different. Translated agents live in `.claude/agents/`.

When a translated agent does not yet exist for a workflow, fall back to following the shared principles in `.github/copilot-instructions.md` and the source agent file directly.

### Session lanes (multi-terminal coordination, 2026-06-11)

Michael runs several Claude Code sessions in parallel, topic-based. The
protocol lives at `.mind/sessions/README.md` — read it once. The short form:

- **Claim your lane** (`.mind/sessions/<topic>.md`). The SessionStart hook does
  this automatically from the session title; if a hook says "no lane claims
  session_id X", create the lane yourself with that id and your topic name.
- **Write only your own lane; read everyone's.** Check the lanes before
  killing/restarting long-lived processes you didn't start. Background shell
  launches are logged to your lane automatically.
- **Signal siblings** by appending to `.mind/sessions/inbox/<lane>.md`. Delivery
  is pull: they're nudged on next engagement; the statusline shows 📬. After
  acting on your own inbox, clear it.
- **`.mind/active.md` is a lean in-flight board** — closed arcs go to
  `.spec/journal/` (the record) and lines get deleted from the board. The old
  87K-token banner ledger is archived at `.mind/archive/`.

### Skills

`.claude/skills/*` and `.github/skills/*` are **independent copies — not symlinks.** They are kept as real files on purpose: Claude Code and Copilot have diverging needs, and a skill is allowed to drift between the two trees as those needs pull on it. A shared skill starts identical in both and may diverge over time; some skills live in only one tree (`council-moment`, `intent-check`, `pgrx-extension-bump`, `sabbath-close` are Claude-Code-only; `playwright-cli` differs deliberately between the two).

When creating or substantially editing a skill both harnesses use, write it to **both** `.github/skills/<name>/SKILL.md` and `.claude/skills/<name>/SKILL.md`. When the change is Claude-Code-specific, edit only the `.claude/` copy and let it drift.

### Slash commands (Copilot prompts)

Copilot's `.github/prompts/*.prompt.md` are not yet translated to Claude Code's `.claude/commands/`. When a command is needed, translate the body of the corresponding prompt and rewrite `${input:foo}` template variables to `$ARGUMENTS` / `$1`.
