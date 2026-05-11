# scripts/git-hooks

Git hooks for this repo. Install by symlink or copy into `.git/hooks/`.

## Install

Symlink (Linux/macOS):

```bash
ln -s ../../scripts/git-hooks/pre-commit .git/hooks/pre-commit
```

Copy (Windows, or if symlinks aren't supported):

```bash
cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Either works — the hooks just need to be in `.git/hooks/` and executable.

## Hooks

### `pre-commit`

Phase 5d / Phase C.3: re-seeds the substrate when `intent.yaml` or `.spec/covenant.yaml` changes.

- Detects staged changes to those files via `git diff --cached --name-only`
- For each changed file: docker cp to `pg-ai-stewards-dev`, then `SELECT stewards.seed_intents_from_yaml(...)` or `seed_covenant_from_yaml(...)`
- Idempotent — sha-comparison short-circuits no-op re-seeds
- Skips gracefully when the dev container isn't running (prints a manual-recovery hint, allows the commit)

The hook never blocks a commit. Worst case: the substrate misses a YAML change until manually re-seeded.
