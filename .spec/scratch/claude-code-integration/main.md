# Claude Code Integration — Scratch File

**Binding Problem:** brain.exe agents currently depend solely on the GitHub Copilot SDK for LLM reasoning. This creates a single point of cost (1500 premium requests/month on Pro+) and a single point of capability. Adding Claude Code as an alternative backend would spread cost and provide access to different model capabilities (200K context, project caching, different agent loop).

**Created:** 2026-04-04

---

## Research Findings

### Two Integration Paths

**Path A: Claude Agent SDK (Python/TypeScript)**
- Programmatic library: `from claude_agent_sdk import query, ClaudeAgentOptions`
- Built-in tools: Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch
- Supports MCP servers, hooks, subagents, skills, CLAUDE.md
- **Billing: Anthropic Console API key** — NOT the $20/mo Pro subscription
- Available in Python and TypeScript only (no Go SDK)
- Async streaming interface
- Same agent loop and context management as Claude Code CLI

**Path B: Claude Code CLI subprocess**
- `claude -p "query" --output-format json` — print mode, non-interactive
- `claude -p --input-format stream-json --output-format stream-json` — streaming
- `--system-prompt` or `--system-prompt-file` for custom system prompts
- `--append-system-prompt` to add to default prompt (preserves Claude Code capabilities)
- `--mcp-config ./mcp.json` for MCP servers
- `--model claude-sonnet-4-6` or `--model opus` for model selection
- `--allowedTools` to whitelist tools without permission prompts
- `--dangerously-skip-permissions` for fully autonomous operation
- `--max-turns` to limit agentic turns
- `--max-budget-usd 5.00` for cost limiting
- `--bare` mode skips auto-discovery for faster scripted calls
- `--json-schema` for structured output
- **Billing: $20/mo Pro subscription** — usage-based, 5-hour session window
- Works from any language via subprocess
- Session persistence: `--session-id UUID` to resume conversations
- `--add-dir` for multi-directory access

### Decision: Path B (CLI subprocess)

**Rationale:**
1. Uses the $20/mo Pro subscription — the whole point is cost spreading from Copilot
2. Works from Go (brain.exe is Go) — no need for Python/TypeScript wrapper
3. Claude Code's built-in tools (file read/write/edit, bash, glob, grep) are exactly what research/plan agents need
4. MCP server support via `--mcp-config` means our existing MCP servers work
5. `--output-format json` gives structured responses parseable in Go
6. `--max-turns` and `--max-budget-usd` provide cost guardrails
7. Session persistence via `--session-id` enables multi-turn conversations

**Trade-offs:**
- Subprocess overhead (process spawn per call) vs in-process SDK
- Less fine-grained control over the agent loop
- Depends on Claude Code CLI being installed globally
- No Go SDK means we can't hook into event streams as deeply
- Usage-based billing is less predictable than per-request pricing

### Architecture Sketch

```
brain.exe
├── internal/ai/
│   ├── agent.go          (existing — Copilot SDK backend)
│   ├── claude_agent.go   (NEW — Claude Code CLI backend)
│   ├── pool.go           (modified — backend selection)
│   └── backend.go        (NEW — interface abstracting both backends)
```

**Backend interface:**
```go
type AgentBackend interface {
    Ask(ctx context.Context, prompt string, w io.Writer) (string, error)
    Close() error
}
```

**Copilot backend:** wraps existing Agent (Copilot SDK sessions)
**Claude backend:** wraps `claude -p` subprocess calls

**Selection logic in pipeline:**
- Research agent → Claude Code (Haiku equivalent? or Sonnet on Pro)
- Plan agent → Copilot SDK (Sonnet, 1.0 premium request)
- OR: configurable per-agent via config

### Claude Code Model Selection

On Pro ($20/mo):
- Default: Sonnet (most efficient for usage limits)
- Opus: Available but drains usage faster
- No Haiku equivalent — smallest is Sonnet
- `--model claude-sonnet-4-6` or `--model opus`

**Implication:** Can't replicate the Haiku (0.33 premium request) cost for research. Claude Code research would use Sonnet-equivalent usage. But since it's on a separate $20/mo budget, it's effectively free relative to Copilot budget.

### MCP Server Configuration

Brain's MCP servers can be passed via `--mcp-config`:
```json
{
  "mcpServers": {
    "becoming": {
      "command": "path/to/mcp.exe",
      "cwd": "path/to/becoming/"
    },
    "search-mcp": {
      "command": "path/to/search-mcp.exe",
      "args": ["serve"],
      "cwd": "path/to/search-mcp/"
    }
  }
}
```

### Working Directory + File Access

Claude Code operates in a working directory and can read/write files.
`--add-dir` adds additional directories.
For brain.exe pipeline agents, working directory = scripture-study repo root.

### Cost Tracking

Claude Code CLI doesn't expose per-call cost directly.
- `--output-format json` includes usage metadata in response
- `--max-budget-usd` provides a session-level cap
- No equivalent of Copilot's "premium request" multiplier
- Would need to track calls and estimate from usage patterns

### Open Questions

1. **Which agents should use Claude Code vs Copilot?**
   - Option A: Research on Claude Code (it's cheaper per Sonnet call when on Pro), Plan on Copilot
   - Option B: All pipeline agents on Claude Code, Copilot for interactive sessions only
   - Option C: Configurable per-agent with fallback

2. **How to handle Claude Code CLI not being installed?**
   - Graceful fallback to Copilot SDK
   - Check at startup: `claude --version`
   - Config flag: `AGENT_CLAUDE_CODE_ENABLED=true`

3. **Session management:**
   - Each research/plan pass = new session (stateless like current Copilot approach)
   - OR: maintain session across multi-turn pipeline operations?
   - Leaning toward stateless (simpler, matches current architecture)

4. **System prompt delivery:**
   - `--system-prompt-file` for governance docs (research-covenant.md, plan-covenant.md)
   - `--append-system-prompt` to keep Claude Code's built-in capabilities
   - Research prompt + governance = `--append-system-prompt-file`

5. **Context window:**
   - Claude Code has 200K context on Pro
   - Current research agent sometimes needs long scratch files + MCP results
   - 200K is generous — better than Copilot's effective limit

6. **Is `--bare` mode appropriate?**
   - Skips CLAUDE.md, skills, plugins, hooks
   - Faster startup for scripted calls
   - But loses project context benefits
   - Probably yes for pipeline agents (they have their own prompts)
