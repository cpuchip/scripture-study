---
name: agy-cli
description: Drive Google Antigravity's `agy` CLI (Gemini 3.5 Flash) headless from Claude Code for second-opinion, voice, or review passes — when you want a non-Claude model to read files and propose edits. Handles the two headless `-p` bugs (stdin-EOF hang, stdout drop). Use as a stopgap until the substrate redline pipeline lands. Claude-Code-only.
---

# Driving `agy` (Gemini) headless from Claude Code

`agy` is Google Antigravity's CLI (`C:\Users\cpuch\AppData\Local\agy\bin\agy.exe`). It runs **on the host with real filesystem access** and uses Michael's **Google subscription** — so it sidesteps the substrate's wall (the pg-ai-stewards lens sandbox can't read the manuscript; `agy` can). Use it to pull **Gemini 3.5 Flash (High)** in for a voice pass, a fresh-eyes review, or any "have another model look at these files and propose changes" task.

This is a **stopgap**. The durable path is the substrate `redline` pipeline (see `/.spec/proposals/substrate-multimodel-document-redline.md`). Until that ships, use this.

## The two bugs you MUST work around

Headless `agy -p` has two known issues (confirmed 2026-05-30; see the community MCP bridge built to work around them: github.com/SinanTufekci/Claude-Code-Antigravity-CLI-MCP-Server):

1. **Hangs in a non-TTY** — it waits on stdin for EOF. Fix: pipe `$null |` (PowerShell) / `</dev/null` (bash) so stdin is closed. Without this it hangs forever and you must `TaskStop` it.
2. **Drops the response from stdout** — exit code 0, empty pipe. The real answer is written to a transcript file on disk. Fix: read it from the newest `~/.gemini/antigravity-cli/brain/<conv-id>/.system_generated/logs/transcript.jsonl` — the last `"type":"PLANNER_RESPONSE"` entry's `content`.

## The recipe (PowerShell)

```powershell
# 1) DISPATCH — from the WORKSPACE ROOT so Antigravity loads the full workspace
#    instructions (the default cwd of the PowerShell tool is the book dir, not root).
Set-Location "C:\Users\cpuch\Documents\code\stuffleberry\scripture-study"
$null | agy -p "<PROMPT>" --dangerously-skip-permissions 2>&1
#    -> "PowerShell completed with no output" is EXPECTED (the stdout-drop bug). Good, not failure.

# 2) RECOVER the response from the newest transcript:
$t = Get-ChildItem "C:\Users\cpuch\.gemini\antigravity-cli\brain" -Recurse -Filter transcript.jsonl |
     Sort-Object LastWriteTime -Descending | Select-Object -First 1
Get-Content $t.FullName |
  Where-Object { $_ -match '"type":"PLANNER_RESPONSE"' } |
  ForEach-Object { ($_ | ConvertFrom-Json).content } |
  Where-Object { $_ -and $_.Trim().Length -gt 0 } |
  Select-Object -Last 1
```

Save the recovered text to `.draft/` (or a scratch file) so it persists — the `Read` tool may truncate a long single-line transcript field in display, but the file holds the whole thing.

## Prompt-writing rules (learned the hard way)

- **Give ABSOLUTE paths and say "do NOT search the filesystem; open only these exact paths."** Given a relative path or a bare filename, `agy` launches a slow system-wide async search that hangs the headless run. Absolute paths + no-search = it reads directly.
- **Run from the workspace root** (`Set-Location` first) so Antigravity picks up workspace context. For voice work, also point it explicitly at the rules: `read C:/.../.github/copilot-instructions.md (the 'Writing Voice' section)` — belt and suspenders.
- **Off-disk discipline:** `agy` *can* edit files (it has write tools). For review/voice passes, end the prompt with "Propose only. DO NOT edit any file." Treat output as a **menu to pick/adapt**, not drop-in — Gemini imports its own register (it drifts ornate against Michael's unadorned voice).
- **Verify-gate doctrine:** `agy`/Gemini has no `gospel_get`. Any proposed edit touching a scripture quote or doctrinal claim must pass your `gospel_get` verification before it lands. Forbid the model from altering quoted scripture.

## Useful flags

`-p`/`--print`/`--prompt` (non-interactive) · `--print-timeout` (default 5m) · `--dangerously-skip-permissions` (auto-approve tool calls — required headless) · `--add-dir <path>` (add a dir to the workspace, repeatable) · `-c`/`--continue`, `--conversation <id>` (resume) · `--sandbox`.

## Gotchas

- The Claude Code PowerShell tool **auto-backgrounds** long `agy` runs; with `$null` it exits cleanly and you get a completion notification (without `$null` it hangs — `TaskStop` it).
- Each PowerShell call is a **fresh session** — `$t`/vars don't persist between tool calls; re-fetch the transcript in the recovery call.
- The model (Gemini 3.5 Flash High) is set in Antigravity's **Model Selection** config, not a CLI flag — confirm via the transcript's `USER_SETTINGS_CHANGE` metadata if unsure.
- This spends Michael's **Google subscription** quota, not the substrate budget.
- A long final answer can span the transcript; if `Select-Object -Last 1` looks cut off, dump all `PLANNER_RESPONSE` contents and concatenate.
