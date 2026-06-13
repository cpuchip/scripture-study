# Scratch Files

This directory holds source logs for studies — verified quotes, observations, critical analysis notes, and the research trail that shows how conclusions were reached.

Each study creates a `{topic}.md` file here during Phase 2 (source gathering) and consumes it during Phase 4 (drafting). **Scratch files are kept permanently** as research provenance — they trace how observations and arguments were developed.

These files are the workbench AND the audit trail. They're published alongside the studies they support.

## ⚠ Provenance correction notice (2026-06-13)

Because these are an audit trail, they record what was verified **at the time a
study was built** — including quotes checked against tools later found faulty.
Two corrections happened *after* many of these files were written:

1. The **2026-06-09 Webster repair** — the dictionary tool had been serving
   **1913 text under an "1828" label**; ~132 definitions across the corpus were
   the wrong edition.
2. The **study-correctness walk (completed 2026-06-13)** — a verse-by-verse
   re-verification of all 469 study files that also caught scripture requotes,
   date errors, and confabulated attributions.

Those corrections were applied **to the study files**, not back-propagated into
these scratch files (they're kept as the original record). **At present, where a
scratch file conflicts with its study, the STUDY is authoritative.** Every
correction is logged in [`study/.audit/findings.md`](../.audit/findings.md).
Affected scratch files carry a dated per-file banner pointing here.

**This is a temporary state.** The intended end is to *normalize* the walk's
corrections **through** the scratch files — bringing them current — at which
point the scratch (the real audit trail) is authoritative again and these
correction banners are **removed** (`findings.md` keeps the permanent record of
what changed). Until that normalization pass runs, the study is the authority.
