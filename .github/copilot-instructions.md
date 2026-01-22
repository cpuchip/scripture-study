# Scripture Study Project

This project is designed to facilitate AI-assisted scripture study, bringing new insights and tracking learning progress.

## Project Structure

```
scripture-study/
├── .github/
│   └── copilot-instructions.md  # This file
├── docs/
│   └── study_template.md        # Patterns and insights for effective AI-assisted study
├── scriptures/
│   └── {book}/
│       └── {subbook}.md         # Full scripture text by subbook
├── study/
│   └── {topic}.md               # Topic-based study notes with scripture references
├── journal/
│   └── {date}.md                # Personal findings, thoughts, and ideas
└── scripts/                     # Utility scripts for downloading/processing
```

## Folder Purposes

### `/docs/`
Contains meta-documentation about how to effectively study and collaborate:
- Patterns discovered during study sessions
- Templates for different types of study
- Notes on effective human-AI collaboration techniques
- Insights that improve future study sessions

### `/scriptures/`
Contains the actual scripture text converted to markdown format. Structure:
- `./scriptures/book/subbook.md` - Each subbook as a single file (preferred for AI context)
- This keeps entire books in context for better AI analysis

### `/study/`
Topic-based study documents where we:
- Pull in various scriptures by topic
- Understand scriptures in context
- Add cross-references and notes
- Build thematic understanding

### `/journal/`
Personal journal entries organized by date:
- `{date}.md` format (e.g., `2026-01-21.md`)
- Contains findings, thoughts, and ideas
- Searchable via VS Code text search

## Scripture Source

Scriptures are sourced from: https://github.com/beandog/lds-scriptures/
- Download latest release and convert to markdown
- Preferred formats for conversion: JSON or SQLite3

## AI Study Guidelines

When studying scriptures:
1. Provide historical and cultural context
2. Cross-reference related scriptures
3. Explain Hebrew/Greek word meanings when relevant
4. Consider multiple interpretations
5. Connect to practical application
6. Note any doctrinal significance

### Scripture Reference Links

When citing scriptures in study files, use markdown links to the source file with anchor to the chapter. This enables clicking directly to the scripture in preview mode.

**Format:** `[Book Chapter:Verse](relative/path/to/##_Book_Name.md#chapter-N)`

**Examples:**
- `[Moses 3:5](../scriptures/Pearl_of_Great_Price/01_Moses.md#chapter-3)`
- `[Genesis 1:1](../scriptures/Old_Testament/01_Genesis.md#chapter-1)`
- `[D&C 93:36](../scriptures/Doctrine_and_Covenants/Section_093.md)`
- `[1 Nephi 3:7](../scriptures/Book_of_Mormon/01_1_Nephi.md#chapter-3)`

Note: Underscores replace spaces in paths. Files are prefixed with numbers for canonical ordering (e.g., `01_Genesis.md`, `02_Exodus.md`). D&C sections use `Section_###.md` format (e.g., `Section_093.md`).

## Workflow

1. **Study Session**: Open or create a topic file in `/study/`
2. **Research**: Pull relevant scriptures and analyze with AI assistance
3. **Document**: Record insights and personal reflections in `/journal/`
4. **Review**: Use VS Code search to find past insights and connections
