# 2026-06-11 — Context statusline + post-compact grounding (small dev session)

**Session shape:** Michael asked how Claude Code plugins work and whether one could help manage the context window. Verified the current docs (code.claude.com) rather than answering from memory — worth it: the hooks surface grew capabilities I didn't know (`PostToolUse.updatedToolOutput` replaces tool results before the model sees them; `PreToolUse.updatedInput` rewrites tool args; statusline JSON carries `context_window.used_percentage` natively).

**Assessment he asked for:** the plugin packaging isn't worth it yet — the workspace already runs most of a context-steward as standalone hooks. Two standalone changes were worth it; he ratified both, shipped as root `3b2fab9`:

1. **`.claude/statusline.py`** — model + color-coded 10-char context bar + 5h/7d rate-limit %, wired via `statusLine` in `.claude/settings.json`. Tested all degraded inputs (null percentages early-session/post-compact, empty JSON, no stdin). UTF-8 stdout reconfigure for the Windows cp1252 gotcha (the Spin lesson).
2. **SessionStart matcher gap found while auditing settings:** `startup|resume|clear` did NOT include `compact` — so the re-read-durable-files grounding stayed silent at exactly the Ammon re-grounding moment (post-compaction, mid-marathon). One-word fix: matcher now `startup|resume|clear|compact`.

**Deliberately not built** (and why, so we don't re-litigate): PostToolUse output cap — the harness already persists oversized tool results to disk + `MAX_MCP_OUTPUT_TOKENS` covers MCP. PreCompact snapshot — a shell hook can only dump transcript tail; the Stop-hook + `.mind/active.md` discipline is better because the model writes state with judgment. PreCompact can't steer what the compaction summary keeps.

**Carry-forward:** Michael wants these patterns packaged as a **shareable plugin someday** — community adoption + his work environment. Migration is mechanical when the day comes (hooks → `hooks/hooks.json`, manifest, marketplace). Natural Working-with-AI teaching artifact. Memory: `project_claude_code_context_plugin`.

**Verify next session:** the statusline bar actually renders at the bottom of Claude Code (config was added mid-session; a restart guarantees pickup), and the grounding hook fires after the next auto-compact.
