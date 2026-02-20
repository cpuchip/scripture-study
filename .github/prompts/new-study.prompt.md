---
name: new-study
description: "Start a new scripture study — scaffolds a study document from the template with discovery-reading-writing workflow"
agent: study
argument-hint: "[topic or scripture passage]"
tools: [read, edit, search, "gospel/*", "gospel-vec/*", "webster/*", "search/*"]
---

Start a new scripture study on the given topic.

## Setup

1. Read the study template for structure: [docs/study_template.md](../../docs/study_template.md)
2. Check if a study on this topic already exists in `study/` — search for related files
3. Check existing studies for cross-connections: scan `study/*.md` filenames for related topics

## Scaffold

Create a new file at `study/${input:slug:topic-slug}.md` with:

```markdown
# ${input:topic:Study Topic}

*Date: ${input:date:YYYY-MM-DD}*

---

## Starting Questions

<!-- What are we trying to understand? What prompted this study? -->

## Discovery

<!-- Phase 1: Use search tools to find relevant scriptures, talks, and word studies -->
<!-- Note file paths — these are pointers to read in the Deep Reading phase -->

## Deep Reading

<!-- Phase 2: read_file every source found above. Follow footnotes. Pull real quotes. -->

## Synthesis

<!-- What patterns emerged? What connections across volumes? -->

## Becoming

<!-- This is the most important section. What does this study ask of me? -->
<!-- Specific commitments, not abstractions. "Pray to see X" not "be more loving." -->
<!-- Connect to an existing becoming/ document if one fits, or start a new one. -->

## Sources Read

<!-- List every file you read_file'd during this study -->
```

## Then Begin

After scaffolding, start **Phase 1 — Discovery**:
- Use `gospel_search` for keyword searches
- Use `search_scriptures` for semantic searches
- Use `webster_define` for key terms (especially Restoration-era vocabulary)
- Note all file paths for Phase 2

Then move to **Phase 2 — Deep Reading** before writing anything in the Synthesis section.
