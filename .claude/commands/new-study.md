---
description: Start a new scripture study — scaffolds a study document from the template with discovery-reading-writing workflow
argument-hint: "[topic or scripture passage]"
---

Start a new scripture study on: **$ARGUMENTS**

If `$ARGUMENTS` is empty, ask the user for the topic and binding question before scaffolding.

If a more specialized study workflow exists, consider invoking it via `Agent(subagent_type=study, ...)` with the same arguments. Otherwise proceed with the steps below.

## Setup

1. Read the study template for structure: [docs/study_template.md](docs/study_template.md)
2. Check if a study on this topic already exists in `study/` — search related files with `Glob study/*.md` and `Grep` on the topic
3. Check existing studies for cross-connections: scan `study/` for related topics

## Scaffold

Create a new file at `study/{topic-slug}.md` with:

```markdown
# {Study Topic}

*Date: {YYYY-MM-DD}*

---

## Starting Questions

<!-- What are we trying to understand? What prompted this study? -->

## Discovery

<!-- Phase 1: Use search tools to find relevant scriptures, talks, and word studies -->
<!-- Note file paths — these are pointers to read in the Deep Reading phase -->

## Deep Reading

<!-- Phase 2: Read every source found above. Follow footnotes. Pull real quotes. -->

## Synthesis

<!-- What patterns emerged? What connections across volumes? -->

## Becoming

<!-- This is the most important section. What does this study ask of me? -->
<!-- Specific commitments, not abstractions. "Pray to see X" not "be more loving." -->
<!-- Connect to an existing becoming/ document if one fits, or start a new one. -->

## Sources Read

<!-- List every file you Read during this study -->
```

## Then Begin

After scaffolding, start **Phase 1 — Discovery**:
- Use `mcp__gospel-engine-v2__gospel_search` (modes: keyword, semantic, combined)
- Use `mcp__webster__webster_define` for key terms (especially Restoration-era vocabulary)
- Use `mcp__byu-citations__byu_citations` to discover who has cited key verses
- Note all file paths for Phase 2

Load the `quote-log` skill to externalize verified quotes immediately as you read.

Then move to **Phase 2 — Deep Reading** before writing anything in the Synthesis section. The `source-verification` skill is canonical: every quote verbatim, every citation count earned by a tool call this session.

End with the `becoming` skill to land the study somewhere personal.
