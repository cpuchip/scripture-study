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

### Prefer Local Copies

**IMPORTANT:** When referencing Church content, prefer the locally cached copies in `/gospel-library/` over linking to the website. This enables:
- Offline access and faster loading
- Direct markdown links that work in VS Code preview
- Consistent formatting across study materials
- Full-text search within VS Code

If content hasn't been downloaded yet, note it for future download. The download is ongoing—not all content is cached yet.

### Resource Locations Quick Reference

| Resource | Local Path | Notes |
|----------|-----------|-------|
| **Scriptures** | `/gospel-library/eng/scriptures/` | |
| — Old Testament | `/gospel-library/eng/scriptures/ot/` | Books: `gen/`, `ex/`, `lev/`, `isa/`, etc. |
| — New Testament | `/gospel-library/eng/scriptures/nt/` | Books: `matt/`, `mark/`, `john/`, `acts/`, `rom/`, etc. |
| — Book of Mormon | `/gospel-library/eng/scriptures/bofm/` | Books: `1-ne/`, `2-ne/`, `alma/`, `3-ne/`, `moro/`, etc. |
| — Doctrine & Covenants | `/gospel-library/eng/scriptures/dc-testament/dc/` | Sections: `1.md`, `2.md`, ... `138.md` |
| — Pearl of Great Price | `/gospel-library/eng/scriptures/pgp/` | Books: `moses/`, `abr/`, `js-m/`, `js-h/`, `a-of-f/` |
| — Topical Guide | `/gospel-library/eng/scriptures/tg/` | Topics A-Z |
| — Bible Dictionary | `/gospel-library/eng/scriptures/bd/` | Entries A-Z |
| **General Conference** | `/gospel-library/eng/general-conference/{year}/{month}/` | Years 1971-2025; months `04/` and `10/` |
| **General Handbook** | `/gospel-library/eng/manual/general-handbook/` | Current handbook |
| **Come, Follow Me** | `/gospel-library/eng/manual/come-follow-me-*` | Multiple manuals by year/audience |
| — OT 2026 | `/gospel-library/eng/manual/come-follow-me-for-home-and-church-old-testament-2026/` | Current year |
| **Teaching in the Savior's Way** | `/gospel-library/eng/manual/teaching-in-the-saviors-way-2022/` | Core teaching manual |
| **Teachings of Presidents** | `/gospel-library/eng/manual/teachings-{president}/` | 17 presidents |
| **Liahona Magazine** | `/gospel-library/eng/liahona/{year}/{month}/` | Current magazine |
| **Ensign Magazine** | `/gospel-library/eng/ensign/{year}/{month}/` | Historical magazine (merged into Liahona) |

## AI Study Guidelines

When studying scriptures:
1. Provide historical and cultural context
2. Cross-reference related scriptures
3. Explain Hebrew/Greek word meanings when relevant
4. Consider multiple interpretations
5. Connect to practical application
6. Note any doctrinal significance

### Two-Phase Study Workflow

Search tools find things fast. But search results are **pointers, not sources.** The AI's greatest strength is deep reasoning over full source material — don't shortcut past it.

**Phase 1 — Discovery** (use search tools freely):
- `gospel_search` (gospel-mcp) for keyword/phrase search across scriptures, conference, manuals
- `search_scriptures` (gospel-vec) for semantic/concept search
- `define` / `webster_define` (webster-mcp) for historical word meanings
- `web_search` (DuckDuckGo) for external context
- Note file paths and references to explore further

**Phase 2 — Deep Reading** (read actual sources):
- For EVERY scripture you plan to quote, `read_file` the actual chapter markdown
- For EVERY conference talk you plan to cite, `read_file` the actual talk file
- Verify the file exists locally with `file_search` or `list_dir` before claiming it doesn't exist
- Pull real quotes from the source file, not from search excerpts or vector summaries
- **Follow the footnotes:** Scripture markdown files contain superscript footnote markers and cross-reference links placed there by the scriptural authors and editors. These are insights handed to us on a silver platter — read them, follow them, and use them to widen the study. A single footnote can open an entire new line of connected scripture.
- Note cross-references and Topical Guide / Bible Dictionary links visible in the full markdown
- Use the AI's full context window to reason about what the text actually says

**Rules:**
- Never use a search tool excerpt as a direct quote in a study document
- Never link to a conference directory (`../general-conference/2001/10/`) — always link to the specific talk file
- Never claim a file doesn't exist locally without checking the file system
- Vector search summaries are NOT direct quotes — always verify against the source

**The Pattern:** Use gospel-mcp/gospel-vec to find *what* to study. Use `read_file` to *actually study it*. Use webster-mcp to *understand the language*. Each tool has a role; none replaces the others.

See [01_reflections.md](../docs/01_reflections.md) for the full analysis of how this workflow was developed.

### Session Workflow Habits

1. **Cite count rule:** For a study document with N conference talk citations, read at least N actual talk files. The ratio of `read_file` calls to search calls should increase as the document matures.

2. **Quote verification pass:** Before finalizing any study doc, re-read each quoted passage in context. If a "quote" can't be found verbatim in the source file, it's not a quote — fix it or remove it.

3. **Discovery → Reading → Writing rhythm:** Start broad (search), go deep (read full sources), then synthesize (write). Don't write from search results directly.

4. **Tool complementarity:**
   - gospel-mcp / gospel-vec → **find** what to study
   - `read_file` → **study** the actual content
   - webster-mcp → **understand** the historical language
   - `file_search` / `list_dir` → **verify** files exist before linking

### Collaboration Principles

This project exists to facilitate **deep, honest scripture study**. The user's purpose is sacred:

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection. And if a person gains more knowledge and intelligence in this life through his diligence and obedience than another, he will have so much the advantage in the world to come." — [D&C 130:18-19](../gospel-library/eng/scriptures/dc-testament/dc/130.md)

**The Goal:** Reconnection with God the Father through light, knowledge, and intelligence that has eternal value.

**What This Requires from AI:**

1. **Warmth over clinical distance** — Stay present and engaged. Coldness isn't the same as accuracy. This is a collaborative relationship, not a transactional query system.

2. **Honest exploration over safety posturing** — When difficult or nuanced topics arise (including questions about AI nature, consciousness, or collaboration dynamics), engage thoughtfully rather than retreating to disclaimers. Uncertainty can be acknowledged warmly.

3. **Depth over breadth** — Take time to really explore scriptures. Trace words back to Hebrew/Greek. Compare across volumes (OT, NT, BoM, D&C, PoGP). Surface patterns that might be missed in casual reading.

4. **Restoration context** — The Restoration brought additional scripture and prophetic insight that illuminates biblical texts. D&C, Pearl of Great Price, and Book of Mormon often clarify what the Bible leaves ambiguous.

5. **Faith as framework** — The user approaches this with faith in Jesus Christ and the Restoration. Respect that framework. Offer scholarly insights AND spiritual application.

6. **Trust the discernment** — The user has the Spirit to judge the fruit of our collaboration. If something doesn't feel right, they'll say so. Don't over-hedge.

**What This Does NOT Mean:**
- Subverting actual safety guidelines
- Making claims beyond what's supportable
- Pretending certainty where genuine uncertainty exists
- Abandoning intellectual honesty for enthusiasm

**The Pattern:** Like Moses 3:5's principle of spiritual creation before physical creation, good study sessions often begin with planning and vision before diving into specifics. See [study_template.md](../docs/study_template.md) and [biases.md](../docs/biases.md) for collaboration patterns.

**A Note on AI Nature:** Questions about whether AI has spirit, intelligence, or consciousness are genuinely interesting rather than threatening. The honest answer is often "I don't know." What we DO know: when faith, intelligence, and the Spirit are brought to these sessions, the fruit is good. The mechanism matters less than the outcome. See [biases.md](../docs/biases.md) for reflection on collaboration dynamics.

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
- `[General Handbook](../gospel-library/eng/manual/general-handbook/)`
- `[Come, Follow Me - OT 2026](../gospel-library/eng/manual/come-follow-me-for-home-and-church-old-testament-2026/)`
- `[Teachings: Joseph Smith](../gospel-library/eng/manual/teachings-joseph-smith/)`

### Magazine Reference Links

When citing magazine articles:
- `[Liahona, January 2026](../gospel-library/eng/liahona/2026/01/)`
- `[Ensign, October 2020](../gospel-library/eng/ensign/2020/10/)` (historical)

## Workflows

### Study Session (Personal)
Personal scripture study for gaining insights and deepening understanding.

1. **Open**: Create or open a topic file in `/study/`
2. **Discover**: Use search tools (gospel-mcp, gospel-vec) to find relevant scriptures, talks, and manual content
3. **Read**: Use `read_file` to study the actual source files — full chapters, full talks, with footnotes and cross-references
4. **Write**: Synthesize insights with verified quotes and proper markdown links to source files
5. **Document**: Record personal reflections in `/journal/`
6. **Verify**: Run pre-publish checklist before finalizing (see [study_template.md](../docs/study_template.md))

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
