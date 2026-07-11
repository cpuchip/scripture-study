# cmux and the fleet-glass question — IndyDevDan digest + harness survey

**Video:** "SEE CMUX SOLVE Multi-Agent Orchestration (Claude Code and Pi Agent)" —
IndyDevDan, 2026-07-06, 30:29, https://www.youtube.com/watch?v=WAFUMBLOjHo
(transcript at `yt/indydevdan/WAFUMBLOjHo/`, digested by subagent 2026-07-09).
**Companion research:** cmux + dmux repos, read same session. Binding question:
what do these harnesses have that loom + the substrate don't, and is any of it
worth adopting or stealing?

## What the tools actually are

- **cmux** ([manaflow-ai/cmux](https://github.com/manaflow-ai/cmux), [cmux.com](https://cmux.com/),
  GPL-3.0, 24.1k★, very active): a native **macOS-only** terminal built on
  libghostty — not an orchestrator, a *programmable presentation layer*. Workspaces
  → surfaces (tabs) → panes, vertical tabs, per-workspace identity, integrated
  browser panes, iOS remote app. The load-bearing feature is the **scriptable CLI +
  Unix socket API**: create/split, send-keys, read-screen, screenshots, and
  lifecycle/notification events. Agents and their subagents render as *native
  panes* instead of hidden background processes.
- **dmux** ([standardagents/dmux](https://github.com/standardagents/dmux),
  [dmux.ai](https://dmux.ai/), MIT, 1.7k★, v5.10.0 July 2026): tmux + git
  worktrees + agent launch. Press `n`, type a prompt, pick agents → it creates
  worktree + branch + pane + launches the agent; merge back or open a PR from the
  pane menu. Supports ~12 agent CLIs. AI branch naming, file browser, lifecycle
  hooks. **Not in the video** — the video is cmux-only and never mentions worktrees.

## The video's spine

Dan's three orchestration problems: (1) no *programmatic* access to your agents,
(2) can't *see* agents, so you can't improve them, (3) booting an agent team by
hand kills the speed ("the thousandth agent run" test). His answer is cmux's
four-verb control loop — "send key … read the screen … open and close surfaces
and the loop repeats" — plus events, driven by a Claude Code orchestrator (Opus
4.8) over a three-tier team (orchestrator → leads → typed workers: plan / build /
build-frontend / test), booted with one `just fast-cc <feature>` command.
Communication is deliberately flat under the hierarchy ("any agent can prompt any
agent"). His filter: "if a tool does not have programmatic access, I just
completely ignore it." His verdict: adopts cmux alongside tmux; the risk is
immaturity — on camera, a completion event fired and his orchestrator never
registered it.

Sharpest line in the video: **"prompting in a black box in sub agents is a great
place to start. Terrible place to finish."** An agent you can't see is an agent
you can't improve.

## Where we already stand (honest audit against his three problems)

1. **Programmatic access — loom is *stronger* than cmux's model.** cmux scripts a
   TTY (send keys, scrape screen); loom returns structured JSON (text, session_id,
   cost, turns) across six backends with trust walls. Racing, resume, steering —
   all first-class. Nothing to adopt here; the video validates the design.
2. **See-to-improve — our real gap, named.** Loom sessions run headless; the
   foreman watches via transcripts and blind checks, but there is no single live
   glass over all running seats with jump-in-and-type. Stewdio's sessions view is
   the natural owner (loom serve already streams SSE), not a Mac terminal.
3. **Thousandth launch — cheap steal.** His justfile team recipes = named,
   reusable topology + one argument. Our full-foreman night booted four arcs by
   hand-writing briefs; a `team boot` recipe (role homes + worktrees + briefs from
   a template) is a morning's work and pays every fleet run after.

## The steals (recorded as candidate experiments)

- **FLEET-GLASS:** one live pane over all running loom/substrate seats — status,
  streaming tail, jump-in. Stewdio panel first (it already speaks to the
  substrate); a terminal fallback later if wanted. This is the see-to-improve gap.
- **TEAM-RECIPES:** one-tap boot of a named foreman topology (`just` or PowerShell
  script: worktrees + role homes + briefs-from-template + lane logging). Update
  the foreman skill with a recipes section when built.
- **RACE-SHAPE:** heterogeneous-fleet racing as a named foreman move — same
  prompt, N backends (loom's six make this unusually cheap for us), first answer
  that passes the oracle wins; tolerant of partial fleet failure (his Pi agents
  died mid-race; the fleet still won). Only lawful where a deterministic check
  names the winner — this composes with grindability, not around it.
- **Event-consumption lesson (free):** his on-camera bug — event emitted,
  orchestrator deaf — is the exact failure our pull-based inbox + statusline nudge
  was built to avoid. Keep pull; don't chase push.

## Verdicts

- **cmux: steal, don't adopt.** Mac-only kills it as a daily driver on Windows;
  the programmability it's famous for is the part loom already exceeds. What it
  actually teaches us is the *glass* — visibility as the precondition for the
  improvement loop.
- **dmux: skim, don't adopt.** It automates exactly the worktree-per-arc flow the
  full-foreman night ran by hand — but Claude Code's native worktree isolation +
  the foreman skill already cover it, and our fleets aren't tmux-shaped. Its pane
  menu (merge / PR per worktree) is a nice ergonomic reference for TEAM-RECIPES.
- **Tension worth keeping:** flat any-agent-prompts-any-agent sits against our
  "boss never implements" foreman discipline and the presiding covenant. His
  version has no guardrails and broke at the notification layer. If workers ever
  ping workers here, it goes through A2A with the watch intact — not around it.

## Sources

- Video + transcript: https://www.youtube.com/watch?v=WAFUMBLOjHo (local `yt/indydevdan/WAFUMBLOjHo/`)
- [manaflow-ai/cmux](https://github.com/manaflow-ai/cmux) · [cmux.com](https://cmux.com/)
- [standardagents/dmux](https://github.com/standardagents/dmux) · [dmux.ai](https://dmux.ai/)
- Surveyed but not load-bearing: [awesome-cli-coding-agents](https://github.com/bradAGI/awesome-cli-coding-agents),
  [awesome-agent-orchestrators](https://github.com/andyrewlee/awesome-agent-orchestrators),
  [claude-cmux-skill](https://github.com/ph3on1x/claude-cmux-skill), [cmux-agent-teams](https://github.com/hungrytech/cmux-agent-teams)
