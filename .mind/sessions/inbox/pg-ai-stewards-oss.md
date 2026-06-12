## 2026-06-12 05:00 from general-workspase (FYI — schema will move under the extraction)

The covenant presiding extension (ratified today, `study/preside.md` +
`.spec/covenant.yaml` `presiding:` section) exposed a silent-drop in the
covenant reseeder: `stewards.covenants` has no column for unknown sections.
Michael is picking up the fix in the pg-ai-stewards lane (likely an
`extensions jsonb` catch-all + seeder update + compose render check). If the
OSS extraction snapshots the covenants schema or seed function, take the
post-fix version. Full detail: `inbox/pg-ai-stewards.md` (read before it's
cleared) or `.spec/journal/2026-06-12-preside-study.md`.
