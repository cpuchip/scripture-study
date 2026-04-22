# Brain UX QoL — Phase 8: Auto-Commit After Agent Sessions

**Status:** Deferred — needs design + safety review before building
**Parent:** [archive/brain-ux-quality-of-life.md](archive/brain-ux-quality-of-life.md) (Phases 1-7b shipped Apr 6)
**Created:** 2026-04-21 (split from parent during cleanup-2026-04-part2)

## Concept

When an agent session completes (pipeline stage transition), the brain evaluates what files were created/modified, generates a meaningful commit message, and commits the work.

## Key design questions

1. **Trigger:** Automatic on stage transition? Button in UI? Both?
2. **Scope:** Which files? All changes since last commit? Only files the agent touched? Need a way to track agent-modified files.
3. **Message:** AI-generated summary of the work? Template with entry title + stage? Human-editable before commit?
4. **Push:** Opt-in per commit? Global setting? Never auto-push?
5. **Safety:** Commits are permanent (in reflog). Push is more consequential. Need confirmation or at minimum clear audit trail.
6. **Architecture:** Pipeline post-stage hook (Go) vs. UI button + backend endpoint. Pipeline hook is cleaner for automation, button gives more control. Could start with button, add automation later.

## Prerequisite

Phase 7 (git status, shipped) gives the UI foundation. Auto-commit builds on knowing what changed.

## Revisit when

Phase 7 has been in daily use long enough to see whether auto-commit is actually needed or whether manual commits are fine.
