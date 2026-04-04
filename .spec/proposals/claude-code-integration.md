# Claude Code Integration — Alternative Agent Backend

## Problem Statement

brain.exe's agent pipeline depends solely on the GitHub Copilot SDK for LLM reasoning. This creates two constraints:

1. **Cost concentration.** All agent work (research, plan, interactive) draws from a single pool of 1500 premium requests/month (Pro+). At 1.33 premium requests per pipeline entry, that's ~1125 entries/month maximum — but interactive Copilot sessions in VS Code compete for the same budget.
2. **Capability lock-in.** The Copilot SDK is excellent but has a specific context model. Claude Code offers different strengths: 200K context window, project-level caching, built-in file/bash tools, and a different agent loop.
3. **Experience gap.** Michael identified Claude Code fluency as a skill to develop (not just for cost, but for understanding the tool's model). Learning by integrating it into brain.exe is hands-on experience with a real production use case.

## Success Criteria

1. brain.exe can invoke Claude Code CLI for research and/or plan passes
2. Existing Copilot SDK path continues to work (no regression)
3. Backend selection is configurable per-agent (env var or config)
4. MCP servers are available to Claude Code agents
5. Usage is logged (calls, model, duration) for observability
6. Graceful fallback: if Claude Code isn't installed, uses Copilot SDK

## Constraints

- Claude Code must be installed globally (`claude` CLI available on PATH)
- Requires active Claude Pro subscription ($20/mo) or Anthropic API key
- Go only — no Python/TypeScript dependencies (use CLI subprocess, not Agent SDK)
- Must work on Windows (PowerShell)
- Cannot break existing pipeline tests

## Prior Art

- Current Copilot SDK integration: `internal/ai/agent.go`, `pool.go`
- Pipeline agents: `internal/pipeline/research.go` (research + plan)
- MCP server discovery: `internal/config/config.go`
- Billing model: PremiumRequestCost field on AgentConfig (Apr 4 overhaul)

## Proposed Approach

### Phase 1: Backend Abstraction + Claude Code Agent (one session)

1. **Define `AgentBackend` interface** in `internal/ai/backend.go`:
   ```go
   type AgentBackend interface {
       AskStreaming(ctx context.Context, prompt string, w io.Writer) (string, error)
       SessionInfo() SessionUsage
       Reset() error
   }
   ```

2. **Wrap existing Copilot agent** as `CopilotBackend` implementing the interface.

3. **Create `ClaudeCodeBackend`** in `internal/ai/claude_agent.go`:
   - Invokes `claude -p "prompt" --output-format json` as subprocess
   - Passes `--append-system-prompt-file` for governance docs
   - Passes `--mcp-config` with dynamically generated MCP config JSON
   - Passes `--model` from config
   - Passes `--max-turns` and optionally `--max-budget-usd`
   - Parses JSON output for response text and usage metadata
   - Logs BILLING line (model, duration, estimated cost)

4. **Config additions:**
   - `CLAUDE_CODE_ENABLED=true/false` (default: false)
   - `CLAUDE_CODE_MODEL=claude-sonnet-4-6` (default)
   - `CLAUDE_CODE_PATH=claude` (path to CLI, default: "claude")

5. **Pipeline integration:**
   - `research.go` checks config: if Claude Code enabled AND installed, use ClaudeCodeBackend for research agent
   - Plan agent stays on Copilot SDK initially (Sonnet 4.6 is well-tested)
   - Later: make per-agent backend configurable

### Phase 2: Testing + Validation (one session)

1. Unit tests for ClaudeCodeBackend with mock subprocess
2. Integration test: research pass on a test entry via Claude Code
3. Compare output quality: same entry researched by both backends
4. Compare cost: track usage on both sides for same work
5. Timing comparison: subprocess overhead vs SDK in-process

### Phase 3: Production Tuning (as needed)

1. Session reuse: `--session-id` for multi-turn if beneficial
2. `--bare` mode evaluation for pipeline agents
3. Model selection optimization (Sonnet vs Opus per task type)
4. Cost dashboard: unified view of Copilot + Claude Code usage

## Phased Delivery

| Phase | Scope | Delivers Value? |
|-------|-------|-------|
| 1 | Backend interface + Claude Code agent + config | Yes — can run research via Claude Code |
| 2 | Tests + quality/cost comparison | Yes — data for cost optimization decisions |
| 3 | Tuning + dashboard | Yes — production-ready dual-backend |

Phase 1 is small enough for one session. Phase 2 validates the approach.

## Verification Criteria

- `go test ./internal/ai/ ./internal/pipeline/` passes
- Research pass completes via Claude Code CLI on a test entry
- Copilot SDK research pass still works (no regression)
- BILLING log shows Claude Code calls with model/duration
- Graceful error when Claude Code not installed

## Costs and Risks

**Costs:**
- $20/mo Claude Pro subscription (Michael is already considering this)
- ~2 sessions of dev work (Phase 1 + Phase 2)
- Maintenance: two backends instead of one

**Risks:**
- Claude Code CLI interface could change (subprocess coupling)
- Usage-based billing is less predictable than per-request
- Pro usage limits might throttle pipeline agents during heavy batch runs
- Different agent loops might produce different quality outputs (need validation)

**Mitigations:**
- Graceful fallback means Copilot SDK is always available
- `--max-budget-usd` prevents runaway costs
- Phase 2 explicitly compares quality before committing to production use
