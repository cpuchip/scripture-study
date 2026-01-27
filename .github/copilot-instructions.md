# Scripture Study Project

This project is designed to facilitate AI-assisted scripture study, bringing new insights and tracking learning progress.

## Project Structure

```
scripture-study/
├── .github/
│   └── copilot-instructions.md  # This file
├── books/
│   └── {collection}/            # Additional books (e.g., lecture-on-faith)
├── docs/
│   └── study_template.md        # Patterns and insights for effective AI-assisted study
├── gospel-library/
│   └── eng/
│       ├── broadcasts/          # Auxiliary training, recordings, misc events
│       ├── general-conference/  # Conference talks by year/session
│       ├── liahona/             # Magazine content
│       ├── manual/              # Lesson manuals and guides
│       ├── scriptures/          # Standard works
│       │   ├── ot/              # Old Testament
│       │   ├── nt/              # New Testament
│       │   ├── bofm/            # Book of Mormon
│       │   ├── dc-testament/    # Doctrine and Covenants
│       │   ├── pgp/             # Pearl of Great Price
│       │   └── {study-aids}/    # TG, BD, JST, maps, etc.
│       └── video/               # Video content
├── study/
│   ├── {topic}.md               # Topic-based study notes with scripture references
│   └── talks/
│       └── {YYYYMM}-{session}{speaker}.md  # Conference talk analysis notes
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

### `/gospel-library/`
Contains Church of Jesus Christ of Latter-day Saints content organized by language (currently `eng/`):

#### `/gospel-library/eng/scriptures/`
The standard works organized by volume:
- `ot/` - Old Testament (e.g., `gen/`, `ex/`, `isa/`)
- `nt/` - New Testament (e.g., `matt/`, `john/`, `acts/`)
- `bofm/` - Book of Mormon (e.g., `1-ne/`, `alma/`, `moro/`)
- `dc-testament/dc/` - Doctrine and Covenants sections
- `pgp/` - Pearl of Great Price (e.g., `moses/`, `abr/`, `js-h/`)
- Study aids: `tg/` (Topical Guide), `bd/` (Bible Dictionary), `gs/` (Guide to the Scriptures), `jst/` (Joseph Smith Translation)

Each book contains numbered chapter files (e.g., `1.md`, `2.md`, `3.md`).

#### `/gospel-library/eng/general-conference/`
Conference talks organized by year and session:
- `{year}/{month}/` - e.g., `2025/04/`, `2025/10/`
- Talk files named by session order and speaker: `57nelson.md`, `11oaks.md`

#### `/gospel-library/eng/manual/`
Lesson manuals and teaching guides including:
- Come, Follow Me curriculum
- Teaching in the Savior's Way
- Scripture study helps
- Stories and supplemental materials

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

## Content Source

Gospel Library content is downloaded and converted from the Church's official Gospel Library:
- Scriptures, manuals, conference talks, and other materials
- See `/scripts/gospel-library/` for download utilities

## AI Study Guidelines

When studying scriptures:
1. Provide historical and cultural context
2. Cross-reference related scriptures
3. Explain Hebrew/Greek word meanings when relevant
4. Consider multiple interpretations
5. Connect to practical application
6. Note any doctrinal significance

### Scripture Reference Links

When citing scriptures in study files, use markdown links to the source file. This enables clicking directly to the scripture in preview mode.

**Format:** `[Book Chapter:Verse](relative/path/to/gospel-library/eng/scriptures/{volume}/{book}/{chapter}.md)`

**Scripture Examples:**
- `[Moses 3:5](../gospel-library/eng/scriptures/pgp/moses/3.md)` - Pearl of Great Price
- `[Genesis 1:1](../gospel-library/eng/scriptures/ot/gen/1.md)` - Old Testament
- `[D&C 93:36](../gospel-library/eng/scriptures/dc-testament/dc/93.md)` - Doctrine and Covenants
- `[1 Nephi 3:7](../gospel-library/eng/scriptures/bofm/1-ne/3.md)` - Book of Mormon
- `[Matthew 5:14](../gospel-library/eng/scriptures/nt/matt/5.md)` - New Testament

**Path conventions:**
- Book abbreviations use lowercase with hyphens (e.g., `1-ne`, `2-cor`, `js-h`)
- Chapters are numbered files without leading zeros (e.g., `1.md`, `10.md`, `138.md`)
- D&C sections are in `dc-testament/dc/` folder

### Talk Reference Links

When citing conference talks:
- `[President Nelson, April 2025](../gospel-library/eng/general-conference/2025/04/57nelson.md)`

### Manual Reference Links

When citing manuals or lessons:
- `[Teaching in the Savior's Way](../gospel-library/eng/manual/teaching-in-the-saviors-way-2022/)`

## Workflows

### Study Session (Personal)
Personal scripture study for gaining insights and deepening understanding.

1. **Open**: Create or open a topic file in `/study/`
2. **Research**: Pull relevant scriptures and analyze with AI assistance
3. **Document**: Record insights and personal reflections in `/journal/`
4. **Review**: Use VS Code search to find past insights and connections

**Template:** Use [study_template.md](../docs/study_template.md) for structured study sessions.

### Talk Preparation (Sacrament Meeting)
Preparing talks for sacrament meeting presentations.

1. **Open**: Create a new file in `/journal/` for talk preparation (e.g., `2026-01-26-talk-charity.md`)
2. **Topic Research**: Search scriptures and conference talks for relevant content
3. **Outline**: Structure the talk with introduction, main points, and testimony
4. **Scripture Selection**: Choose key scriptures that support your message
5. **Personal Stories**: Include relevant personal experiences that illustrate principles
6. **Review**: Ensure talk fits within time constraints and flows naturally

**Template:** Use [talk_template.md](../docs/talk_template.md) for structured talk preparation.

### Lesson Planning (Class Instruction)
Preparing lessons for Sunday School, Relief Society, Elders Quorum, or other class settings.

1. **Open**: Browse `/gospel-library/eng/manual/` for the appropriate curriculum
2. **Study Manual**: Read the assigned lesson material thoroughly
3. **Prepare Questions**: Develop discussion questions that invite class participation and personal reflection
4. **Cross-Reference**: Find additional scriptures and talks that support lesson objectives
5. **Apply Principles**: Follow Teaching in the Savior's Way (`/gospel-library/eng/manual/teaching-in-the-saviors-way-2022/`):
   - Love those you teach
   - Teach by the Spirit
   - Teach the doctrine
   - Invite diligent learning
6. **Document**: Save lesson notes in `/journal/` with date and lesson topic

**Key Teaching Principles:**
- Ask questions that encourage pondering and discussion, not just yes/no answers
- Allow time for class members to share insights and experiences
- Focus on helping learners apply principles, not just covering content
- Invite the Spirit through testimony and relevant scripture

**Template:** Use [lesson_template.md](../docs/lesson_template.md) for structured lesson preparation.

### Talk Review (Conference Talk Analysis)
Analyzing a general conference talk to understand WHY it's effective as a teaching model.

1. **Select**: Choose a conference talk from `/gospel-library/eng/general-conference/` that moved or inspired you
2. **Read**: Read the talk carefully, noting what stood out emotionally and spiritually
3. **Analyze**: Evaluate the talk against the Teaching in the Savior's Way framework:
   - Focus on Jesus Christ (How does it point to the Savior?)
   - Love Those You Teach (How does the speaker show vulnerability and create safety?)
   - Teach by the Spirit (What invites the Spirit? Specificity, testimony, etc.)
   - Teach the Doctrine (How are scriptures and prophets used?)
   - Invite Diligent Learning (What invitations are given?)
4. **Identify Techniques**: Note rhetorical patterns, story placement, scripture density
5. **Personal Reflection**: Consider how to apply these techniques in your own teaching
6. **Document**: Save analysis in `/study/talks/` with naming convention `{YYYYMM}-{session}{speaker}.md`

**Example:** `202510-24brown.md` for Elder Brown's October 2025 talk

**Cross-Reference:** See [Teaching in the Savior's Way](../gospel-library/eng/manual/teaching-in-the-saviors-way-2022/) for the framework, and [general-conference-examples.md](../docs/general-conference-examples.md) for talk pattern analysis.
