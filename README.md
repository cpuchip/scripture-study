# Scripture Study

AI-assisted scripture study for members of The Church of Jesus Christ of Latter-day Saints. This project combines local copies of gospel content with AI tools to enable deep, cross-referenced scripture study.

## 🎯 Purpose

This project facilitates:
- **Deep scripture study** with AI assistance for cross-referencing and context
- **Talk and lesson preparation** with templates and patterns from General Conference
- **Personal scripture journaling** with searchable, linked notes
- **Topic-based study** pulling together scriptures, conference talks, and manuals

## ⚠️ Important: Gospel Content Not Included

**This repository does not contain Church content.** The `gospel-library/` folder is in `.gitignore`.

Instead, this repo provides **tools to download your own copy** of Church materials from the official Gospel Library API. This approach:
- Respects intellectual property by not redistributing Church content
- Ensures you have the latest versions
- Gives you control over what content you download
- Works with the same public API the Church's own apps use

## 🚀 Quick Start

### Prerequisites

- **VS Code** with GitHub Copilot (Pro+ recommended for Claude access)
- **Go 1.21+** for running the download/indexing tools
- **Claude Opus 4.5** or better for AI-assisted study (via GitHub Copilot)

### Step 1: Download Gospel Content

All commands run from the repository root (uses `go.work` for module resolution):

```bash
# To run the downloader
go run .\scripts\gospel-library\cmd\gospel-downloader

# Download standard works + latest conference
go run .\scripts\gospel-library\cmd\gospel-downloader --standard

# Or use the interactive TUI to select specific content
go run .\scripts\gospel-library\cmd\gospel-downloader
```

See [scripts/gospel-library/README.md](scripts/gospel-library/README.md) for full documentation.

### Step 2: Set Up the MCP Server (Optional but Recommended)

The Gospel MCP server enables AI assistants to search and retrieve scripture content directly.

**Note:** The MCP server must be built and run from its own directory:

```bash
cd scripts/gospel-mcp

# Build with FTS5 full-text search support
go build -tags "fts5" -o gospel-mcp.exe ./cmd/gospel-mcp

# Index your downloaded content (--root points to repo root)
./gospel-mcp.exe index --root ../../

# Return to repo root
cd ../..
```

Add to your VS Code `settings.json`:
```json
{
  "mcp": {
    "servers": {
      "gospel": {
        "command": "/path/to/gospel-mcp.exe",
        "args": ["serve", "--db", "/path/to/gospel.db"]
      }
    }
  }
}
```

See [scripts/gospel-mcp/README.md](scripts/gospel-mcp/README.md) for full documentation.

### Step 3: Start Studying

Open the workspace in VS Code with GitHub Copilot enabled. The `.github/copilot-instructions.md` file automatically configures Claude with context about the project structure and study patterns.

## 📁 Project Structure

```
scripture-study/
├── docs/                    # Templates and study patterns
│   ├── study_template.md    # Personal scripture study sessions
│   ├── talk_template.md     # Sacrament meeting talk preparation
│   ├── lesson_template.md   # Sunday School/RS/EQ lesson prep
│   └── general-conference-examples.md  # Analysis of apostle talk patterns
│
├── study/                   # Personal study notes (private working copies)
│   ├── {topic}.md           # Topic-based studies
│   ├── talks/               # Conference talk analysis
│   └── cfm/                 # Come Follow Me notes
│
├── public/                  # Published versions for external linking
│   ├── study/               # Polished study documents
│   ├── lessons/             # Shareable lesson materials
│   └── callings/            # Calling-specific resources
│
├── callings/                # Calling-specific work
│   ├── sunday_school/       # Sunday School President materials
│   └── ward_council/        # Ward council resources
│
├── lessons/                 # Lesson preparation materials
│
├── books/                   # Additional study materials
│   └── lecture-on-faith/    # Lectures on Faith (public domain)
│
├── gospel-library/          # Downloaded Church content (NOT in git)
│   └── eng/
│       ├── scriptures/      # Standard works
│       ├── general-conference/  # Conference talks by year
│       ├── manual/          # Manuals and curriculum
│       └── liahona/         # Magazine content
│
├── scripts/                 # Tools
│   ├── gospel-library/      # Content downloader (TUI)
│   ├── gospel-mcp/          # MCP server for AI search
│   └── publish/             # Script to sync study/ → public/
│
└── .github/
    └── copilot-instructions.md  # AI context and collaboration guidelines
```

## 📝 Templates

### Study Template (`docs/study_template.md`)
For personal scripture study sessions. Includes:
- Spiritual/physical creation pattern for study planning
- Scripture gathering and cross-referencing
- Personal application and journaling

### Talk Template (`docs/talk_template.md`)
For sacrament meeting talk preparation. Based on analysis of 10+ General Conference talks:
- Opening patterns (story vs. scripture context)
- Structure options (thematic, metaphor, points)
- Climactic story placement
- Prayer/blessing closing format

### Lesson Template (`docs/lesson_template.md`)
For class instruction based on Teaching in the Savior's Way:
- Spiritual preparation checklist
- Discussion question development
- Invitation to act patterns

## 🤖 Using with GitHub Copilot

This project is designed for use with GitHub Copilot Pro+ and Claude Opus 4.5 (or better).

### How It Works

1. **Copilot Instructions**: The `.github/copilot-instructions.md` file provides Claude with:
   - Project structure and file locations
   - Scripture reference link formats
   - Study collaboration principles
   - Resource location quick reference

2. **MCP Server** (optional): Enables Claude to directly search and retrieve scripture content using:
   - `gospel_search` - Full-text search across all content
   - `gospel_get` - Retrieve specific verses by reference
   - `gospel_list` - Browse available content

3. **Local Content**: Downloaded markdown files can be read directly by Claude for context

### Example Prompts

```
"Let's study Moses 3:5 and its teaching about spiritual creation"

"Help me prepare a talk on covenants using the talk template"

"Search for conference talks about the ten virgins parable"

"Cross-reference D&C 93:36 with other scriptures about intelligence"
```

## 🔗 Public Folder

The `public/` directory contains published versions of study documents suitable for external sharing:
- Polished, reviewed content
- Stable links for sharing
- Synchronized from working directories via `scripts/publish/`
- go run .\scripts\publish\cmd\main.go

## 📖 Content Sources

| Source | Location | Description |
|--------|----------|-------------|
| Standard Works | `gospel-library/eng/scriptures/` | All scripture volumes |
| General Conference | `gospel-library/eng/general-conference/` | 1971–present |
| Manuals | `gospel-library/eng/manual/` | Come Follow Me, handbooks, etc. |
| Magazines | `gospel-library/eng/liahona/` | Current magazine content |
| Lectures on Faith | `books/lecture-on-faith/` | Historical study materials |

## ⚖️ Copyright Notice

**Gospel Library content** is © The Church of Jesus Christ of Latter-day Saints. This repository:
- Does **not** include or redistribute Church content
- Provides tools to download content from the Church's public API
- Enables the same access the official Gospel Library apps provide
- Stores downloaded content locally for personal study as organized markdownfiles with footnotes and references links

**Original content** in this repository (templates, study notes, scripts) is released under the MIT License.

## 🤝 Contributing

This is a personal study project, but the tools and templates may be useful to others. Feel free to:
- Fork and adapt for your own study
- Submit issues for tool improvements
- Share your own template variations

## 📚 Related Resources

- [Gospel Library](https://www.churchofjesuschrist.org/study) - Official Church study app
- [Come, Follow Me](https://www.churchofjesuschrist.org/study/come-follow-me) - Weekly curriculum
- [General Conference](https://www.churchofjesuschrist.org/study/general-conference) - Conference talks

---

*"Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."* — [D&C 130:18](gospel-library/eng/scriptures/dc-testament/dc/130.md)
