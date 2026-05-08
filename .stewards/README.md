# `.stewards/` — model-specific prompt overrides for the substrate

This directory holds **model-specific agent prompts** that the
`pg-ai-stewards` substrate imports as `(family, model_match)` variants
in `stewards.agents`. They live here, in git, so we can revise them
deliberately and roll back when an iteration regresses.

## Why this exists

The base agent corpus at `.github/agents/*.agent.md` was authored for
Claude Opus 4.7 / Sonnet 4.6. Those prompts assume a model that is
literal but flexible, instruction-followable but capable of voice
imitation, and that defers to user judgement when uncertain.

Different models have different defaults. Kimi-k2.6 reaches for
symmetric pairs and triadic flourishes; Qwen-3.6 has a different
register; future models will have their own. The substrate already
supports per-variant agents via `(family, model_match)` as a PK —
e.g., a `study` family can have `model_match='*'` (the default
prompt) and `model_match='kimi-*'` (a kimi-tuned override).

This directory is where those overrides are authored and versioned.

## Structure

```
.stewards/
├── README.md                       (this file)
├── kimi-k2.6/                      (model-name-as-folder)
│   ├── README.md                   (model-specific amendment list)
│   ├── study.agent.md              (full prompt, kimi-tuned)
│   ├── watchman-consolidator.agent.md
│   └── …
├── qwen-3.6/                       (future)
└── opus-4.7/                       (future, if we ever override base)
```

**Folder name = model_match prefix.** `kimi-k2.6/` means files inside
target `model_match = 'kimi-*'`. `qwen-3.6/` would target `'qwen-*'`.
The convention is glob-style and matches the substrate's
`longest-glob-match-wins` perm resolution.

## File format

Each `*.agent.md` is a **complete prompt** (not a delta) — this keeps
the file directly importable without runtime layering, and lets git
diff show exactly what we changed across revisions.

YAML frontmatter is the same shape as `.github/agents/*.agent.md`,
plus one new field:

```yaml
---
description: 'Scripture study agent — kimi-k2.6 voice-tuned variant'
tools: [vscode, execute, read, ...]
model_match: 'kimi-*'        # NEW — controls which agents row this targets
base: '../../.github/agents/study.agent.md'   # NEW — pointer to source-of-truth
amendments:                                   # NEW — one-line summary of each delta
  - 'Forbid closing-refrain by function, not just form'
  - 'Anti-symmetry instruction in Phase 5 voice audit'
  - 'Anglo-Saxon over Latinate cut list'
---

(full prompt body follows; the body is what gets stored as agents.prompt)
```

Three rules for amendments:
1. **State the kimi-specific behavior in the prompt body.** The
   amendments list at the top is a changelog, not a substitute for the
   instruction itself.
2. **Reference the base file.** Anyone editing the kimi variant
   should also know what the default says, so they can decide whether
   the difference is still intentional.
3. **One-line amendment summaries.** Each entry should be readable in
   a git diff without re-reading the whole prompt.

## Import flow

> **Status (2026-05-08):** the existing `stewards-cli import --source agent:.github/agents/`
> hardcodes `model_match='*'` (see `cmd/stewards-cli/internal/importer/agents.go:126`).
> A follow-up extends the importer to honor `model_match` from frontmatter
> *or* derive it from the parent folder name. Once that lands:

```
stewards-cli import --source agent:.stewards/kimi-k2.6/
```

…will UPSERT each file's body into `stewards.agents (family, 'kimi-*')`.
The base `(family, '*')` row stays untouched; agent-resolution at chat
time uses longest-glob-match.

Until the importer extension lands, these files can be applied
manually by:

```sql
-- Example for one agent:
UPDATE stewards.agents
   SET prompt = pg_read_file('/work/.stewards/kimi-k2.6/study.agent.md')
 WHERE family = 'study' AND model_match = 'kimi-*';
```

…or by `psql` paste from the file body.

## Iteration discipline

These prompts are **dev artifacts** until the substrate's soak observation
window confirms the variant produces what we want. Workflow:

1. Author / revise the prompt file in this folder.
2. Re-import (or hand-apply) into `stewards.agents`.
3. Run a study via the `study-write` pipeline with a model matching the
   variant (`--provider opencode_go --model kimi-k2.6`).
4. Compare output against Michael's voice baseline (`study/give-away-all-my-sins.md`,
   `study/art-of-delegation.md`, `study/art-of-presidency.md`).
5. If improved: commit with a journal note describing what the change
   targeted and what it produced. If regressed: revert.

A variant graduates from "experimental" to "stable" when a study
produced under it requires no voice-cleanup before publishing. Until
then, every produced study should be reviewed by a person before
landing in `study/`.

## What this directory is **not**

- Not a place for one-off prompts or research notes. Use
  `.spec/scratch/` for those.
- Not a place for non-substrate prompts. Claude Code's `.claude/` and
  Copilot's `.github/agents/` remain canonical for IDE-side agents.
- Not a substitute for the base prompts in `.github/agents/`. Those
  are the source-of-truth; these are tuned variants.
