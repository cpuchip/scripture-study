## 2026-06-12 05:00 from general-workspase (Michael will pick this up here)

**The covenant reseeder silently dropped the new `presiding:` section — the
substrate is dispatching under the old covenant shape.**

Context: Michael ratified the presiding extension into `.spec/covenant.yaml`
yesterday-night/today (council record: `.spec/journal/2026-06-12-preside-study.md`;
the study: `study/preside.md`; commit `8355f951`). The root pre-commit hook
fired ("reseeding stewards.covenants") and created an ACTIVE row — but:

- `stewards.covenants` is STRUCTURED: one column per known section
  (`human_commits_to`, `agent_commits_to`, `when_broken`, `recovery`,
  `council_moment`, `teaching_extension`, …). There is no `presiding` column.
- The new active row `29e1a8d9-e2a8-48de-819c-ca351a6e07c5`
  (created 2026-06-12 09:37:57+00, source_yaml_sha b129222d…) contains **no
  trace of `preside`** in any column — verified by ILIKE probes on
  agent_commits_to and teaching_extension.
- So `compose_system_prompt`'s covenant block still serves the pre-presiding
  covenant to every dispatch.

**Recommended fix (counsel as you see fit):** rather than a one-off
`presiding` column, an `extensions jsonb` catch-all (or equivalent) in
`stewards.covenants` + seeder update, so the NEXT covenant evolution can't be
silently dropped either — this is the second silent-drop cousin after
seed_fingerprints. Then re-run the reseed (re-commit covenant.yaml or invoke
the seed fn directly) and verify the active row carries `preside_under_121`
+ that a real dispatch's composed prompt includes the presiding terms.
Also worth a glance: does `compose_system_prompt` render extension sections
generically, or does teaching_extension have bespoke rendering the new
section would also need?

Heads-up sent to the pg-ai-stewards-oss lane too (schema change lands mid-
extraction). The walls-vs-compulsion audit of substrate mechanisms
(`study/preside.md` §V) is a named follow-on if it fits this lane's queue.

— found via `watch_what_you_order`, four minutes after its own ratification.
