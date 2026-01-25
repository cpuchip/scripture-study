# Gospel Library Downloader

A terminal user interface (TUI) tool for downloading content from the Church of Jesus Christ of Latter-day Saints [Gospel Library](https://www.churchofjesuschrist.org/study) and converting it to Markdown for local AI-assisted scripture study.

## Features

- 📚 **Browse** Gospel Library content hierarchy interactively
- ✅ **Select** individual items or entire collections for download
- 📥 **Download** with automatic recursive crawling for nested content
- 📝 **Convert** to clean Markdown with:
  - Verse numbers formatted as `**1.**`
  - Footnotes with HTML anchors for linking
  - Relative cross-reference links between files
  - Audio links preserved
- 💾 **Cache** raw API responses to avoid re-downloading
- 🎯 **Quick Start** with `--standard` flag for scriptures + latest conference

## Installation

Requires Go 1.21 or later.

```bash
# From the scripture-study project root
cd scripts/gospel-library

# Build the binary
go build -o gospel-downloader.exe ./cmd/gospel-downloader

# Or run directly
go run ./cmd/gospel-downloader
```

## Usage

### Interactive TUI (Default)

```bash
# Launch the interactive browser
./gospel-downloader

# Or with go run
go run ./cmd/gospel-downloader
```

### TUI Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate list |
| `Enter` | Browse into collection/section |
| `Space` | Toggle selection (any item type) |
| `a` | Select all items in current view |
| `d` | Download selected items |
| `c` | Clear selection |
| `Backspace` | Go back |
| `/` | Filter/search |
| `q` | Quit |

### Status Indicators

| Icon | Meaning |
|------|---------|
| ` ` (blank) | Not selected, not downloaded |
| `●` | Selected (pending download) |
| `✓` | Already downloaded/cached |
| `◉` | Selected AND cached (will re-download) |

### Quick Download: Standard Works

Download all scriptures and the latest General Conference in one command:

```bash
./gospel-downloader --standard
```

This downloads:
- Book of Mormon
- Doctrine and Covenants
- Pearl of Great Price
- Old Testament
- New Testament
- October 2025 General Conference

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--lang` | `eng` | Language code (eng, spa, por, fra, deu, etc.) |
| `--cache` | `.gospel-cache` | Directory for cached API responses |
| `--output` | `gospel-library` | Directory for converted Markdown files |
| `--standard` | - | Download standard works + latest conference |
| `--cleanup` | - | Clear the cache directory |
| `--reset` | - | Clear both cache and output directories |

### Debug/Test Flags

| Flag | Description |
|------|-------------|
| `--test` | Test API with a sample request |
| `--test-cache` | Test API caching (run twice to see cache hit) |
| `--test-convert` | Test HTML to Markdown conversion |
| `--test-crawl` | Debug crawl to see what API returns |

## Output Structure

```
gospel-library/
└── eng/
    ├── scriptures/
    │   ├── bofm/
    │   │   ├── 1-ne/
    │   │   │   ├── 1.md      # 1 Nephi Chapter 1
    │   │   │   ├── 2.md
    │   │   │   └── ...
    │   │   ├── 2-ne/
    │   │   └── ...
    │   ├── dc-testament/
    │   │   └── dc/
    │   │       ├── 1.md      # D&C Section 1
    │   │       └── ...
    │   ├── pgp/
    │   ├── ot/
    │   └── nt/
    └── general-conference/
        └── 2025/
            └── 10/
                ├── 11nelson.md
                ├── 12uchtdorf.md
                └── ...
```

## Markdown Output Format

Each file includes YAML frontmatter and clean content:

```markdown
---
title: "1 Nephi 3"
audio: "https://assets.churchofjesuschrist.org/.../1-ne-3.mp3"
---

# 1 Nephi 3

**1.** And it came to pass that I, Nephi, returned from speaking 
with the Lord, to the tent of my father.

**2.** And it came to pass that he spake unto me, saying: Behold 
I have dreamed a dream...

---

## Footnotes

<a id="fn-2a"></a>**2a.** "dream" — See [Genesis 37:5](../../ot/gen/37.md)
```

## Cache Directory

Raw API responses are cached in `.gospel-cache/`:

```
.gospel-cache/
└── eng/
    ├── content/           # Full content JSON
    │   └── scriptures/
    │       └── bofm/
    │           └── 1-ne/
    │               └── 1.json
    └── collection/        # Navigation/catalog JSON
        └── scriptures/
            └── bofm.json
```

This allows:
- Re-running conversion without re-downloading
- Faster browsing of already-fetched content
- Offline access to cached content

## Examples

```bash
# Download Spanish content
./gospel-downloader --lang=spa

# Custom output directory
./gospel-downloader --output=./my-scriptures

# Clear everything and start fresh
./gospel-downloader --reset

# Clear cache but keep converted files
./gospel-downloader --cleanup
```

## Rate Limiting

The tool implements polite rate limiting (1-2 requests/second) to avoid overwhelming Church servers. Please be respectful of this shared resource.

## Copyright Notice

Downloaded content is copyrighted by The Church of Jesus Christ of Latter-day Saints and is for **personal study use only**. Do not redistribute downloaded content.

The cache and output directories are gitignored by default to prevent accidental commits of copyrighted material.

## Development

```bash
# Run tests
go test ./...

# Build
go build -o gospel-downloader.exe ./cmd/gospel-downloader

# Run with verbose output
go run ./cmd/gospel-downloader --test
```

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML conversion

## License

MIT License - See repository root for details.
