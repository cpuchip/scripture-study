# Scripture Study Project — Gemini Context

The canonical project instructions are shared with Copilot and Claude. Read them first, then the Gemini-specific addendum below.

@.github/copilot-instructions.md

---

## Gemini / Antigravity IDE Addendum

This file is loaded automatically by the Gemini agent / Antigravity IDE on session start. Everything above (via the `@`-import of `.github/copilot-instructions.md`) applies equally to Gemini and other agents.

### Tool Naming in Antigravity IDE

Within Antigravity IDE, MCP tools are eagerly or lazily loaded. Eager tools are registered directly under the name `mcp_<server>_<tool>`. For lazily-loaded tools, call them via the `call_mcp_tool` tool.

| Need | Server Name | Tool Name |
|------|-------------|-----------|
| Search scriptures/talks | `gospel-engine-v2` | `gospel_search` |
| Get a scripture/talk | `gospel-engine-v2` | `gospel_get` |
| Browse content | `gospel-engine-v2` | `gospel_list` |
| Webster 1828 | `webster` | `define` / `webster_define` / `modern_define` |
| BYU citations | `byu-citations` | `byu_citations` |

### Core Commandment: Read Before Quoting

Every direct scripture or prophetic quote MUST be verified by calling `view_file` (or `gospel_get`) on the actual source file in `/gospel-library/` BEFORE writing it to the manuscript or provenance files. Memory is not a source. Search excerpts are pointers only.

Always follow the `source-verification` skill rules:
1. Locate the file in `/gospel-library/`
2. Call `view_file` to inspect the exact wording.
3. Keep quotes verbatim.
4. Record the verification in the `.scratch/provenance_*.md` files.
