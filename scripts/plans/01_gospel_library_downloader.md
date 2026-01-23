# Gospel Library Downloader - Planning Document

> **Status:** Planning  
> **Created:** 2026-01-23  
> **Goal:** Download General Conference talks and other resources from the Church of Jesus Christ of Latter-day Saints Gospel Library for local AI-assisted study.

## Overview

The Gospel Library app syncs content from Church servers. We want to:
1. Discover the API the app uses
2. Download content locally (cached in `.cache/`)
3. Convert to Markdown for human/AI readability
4. Keep content `.gitignore`d (copyrighted material)
5. Provide sync capabilities to stay current

---

## 1. API Research

### ✅ VERIFIED API Endpoints (January 2026)

The Gospel Library website uses a well-structured API. Confirmed via browser network inspection:

#### Primary API Base URL
```
https://www.churchofjesuschrist.org/study/api/v3/language-pages/
```

#### Dynamic/Collection Pages (Browsing/Navigation)
```
GET https://www.churchofjesuschrist.org/study/api/v3/language-pages/type/dynamic?lang=eng&uri={path}
```
- Returns hierarchical collection structure with sections and entries
- Used for browsing conferences, topics, speakers, etc.
- Example paths: `/general-conference`, `/general-conference/2025/10`

**Response Structure:**
```json
{
  "collection": {
    "breadCrumbs": [...],
    "title": "General Conference Collection",
    "uri": "/general-conference",
    "sections": [
      {
        "title": "Conferences",
        "sectionKey": "general conference_0",
        "entries": [
          {
            "title": "October 2025",
            "uri": "/general-conference/2025/10",
            "type": "item",           // "item" = navigable content, "collection" = folder
            "src": "https://...",     // thumbnail image
            "archived": false,
            "category": "General Conference",
            "position": 128
          }
        ]
      },
      {
        "title": "Speakers",
        "entries": [...]
      },
      {
        "title": "Topics", 
        "entries": [...]
      }
    ]
  }
}
```

#### Content Pages (Actual Talk/Article Content)
```
GET https://www.churchofjesuschrist.org/study/api/v3/language-pages/type/content?lang=eng&uri={path}
```
- Returns full content for a specific item (talk, article, chapter)
- Example path: `/general-conference/2025/10/19oaks`

**Response Structure:**
```json
{
  "meta": {
    "title": "Introduction",
    "canonicalUrl": "/general-conference/2025/10/19oaks?lang=eng",
    "contentType": "text/html",
    "audio": [
      {
        "mediaUrl": "https://assets.churchofjesuschrist.org/...mp3",
        "variant": "audio"
      }
    ],
    "pageAttributes": {
      "data-content-type": "general-conference-talk",
      "data-uri": "/general-conference/2025/10/19oaks",
      "data-asset-id": "466660303a7741d3892c5f57382c7096"
    },
    "ogTagImageUrl": "https://...",
    "structuredData": "{\"@context\":\"https://schema.org\",\"@type\":\"WebPage\",\"datePublished\":\"2025-10-04T00:00:00.000Z\"}"
  },
  "content": "<header>...</header><p>Talk content as HTML...</p>",
  "pids": [...],
  "tableOfContentsUri": "...",
  "uri": "/general-conference/2025/10/19oaks",
  "verified": true,
  "restricted": false
}
```

#### Auxiliary Endpoints
```
# CSS customization rules (not needed for download)
GET https://www.churchofjesuschrist.org/study/api/v3/language-pages/misc/css-rules

# Notifications (not needed for download)
GET https://www.churchofjesuschrist.org/study/api/v3/language-pages/misc/notifications?lang=eng

# Site navigation data (header, footer, etc.)
GET https://www.churchofjesuschrist.org/services/platform/v4/resources/data?lang=eng
```

### Entry Types in Collections

| Type | Description |
|------|-------------|
| `item` | Navigable content page (talk, article, chapter) |
| `collection` | Folder/grouping (year range, speaker page) |
| `search` | Search widget (used for speaker/topic search) |

### Rate Limiting & Authentication

- ✅ Public content accessible without authentication
- Should implement polite rate limiting (1-2 requests/second)
- Respect robots.txt and terms of service
- User-Agent should identify our tool appropriately

---

## 1b. Content Types to Support

Target content from the Gospel Library:

| Content Type | Priority | Notes |
|--------------|----------|-------|
| General Conference | High | 1971-present, talks by session |
| Scriptures | High | With footnotes! Cross-references |
| Come, Follow Me | High | Current year curriculum |
| General Handbook | Medium | Leadership/policy reference |
| Books (Jesus the Christ, etc.) | Medium | Classic gospel study |
| Teaching in the Savior's Way | Medium | Teaching manual |
| Magazines (Ensign, Liahona) | Low | Historical articles |
| Other manuals | Low | As needed |

---

## 1c. Media Handling Strategy

| Media Type | Strategy |
|------------|----------|
| **Images** | Download locally to `.cache/gospel-library/images/`, embed with relative markdown links |
| **Audio (MP3)** | Link to church CDN URL (don't download, files are large) |
| **Video** | Link to church website URL |
| **PDF** | Link to church website URL |

### Image Storage
```
.cache/gospel-library/images/
└── general-conference/
    └── 2025/
        └── 10/
            └── 19oaks/
                └── image1.jpg
```

### Markdown Output Example
```markdown
![Speaker at pulpit](../../.cache/gospel-library/images/general-conference/2025/10/19oaks/image1.jpg)

**Media:**
- 🎧 [Listen to audio](https://assets.churchofjesuschrist.org/.../talk.mp3)
- 🎬 [Watch video](https://www.churchofjesuschrist.org/...)
- 📄 [Download PDF](https://www.churchofjesuschrist.org/.../talk.pdf)
```

---

## 2. Directory Structure

```
scripture-study/
├── .cache/                              # Raw downloaded files (gitignored)
│   └── gospel-library/
│       ├── catalog/                     # Cached catalog JSON from API
│       │   ├── eng_general-conference.json
│       │   ├── eng_scriptures.json
│       │   └── eng_come-follow-me.json
│       ├── content/                     # Raw API responses (JSON with HTML)
│       │   ├── general-conference/
│       │   │   └── 2025/
│       │   │       └── 10/
│       │   │           ├── 19oaks.json
│       │   │           └── 16uchtdorf.json
│       │   └── scriptures/
│       │       └── book-of-mormon/
│       ├── images/                      # Downloaded images
│       │   └── general-conference/
│       │       └── 2025/10/
│       │           └── 19oaks/
│       │               └── speaker.jpg
│       ├── metadata.json                # Sync state & timestamps
│       └── selections.json              # TUI selection state (what to download)
│
├── resources/                           # Converted markdown (gitignored)
│   └── gospel-library/
│       ├── general-conference/
│       │   └── 2025/
│       │       └── october/
│       │           ├── _index.md
│       │           ├── 19_oaks_introduction.md
│       │           └── 16_uchtdorf_do_your_part.md
│       ├── scriptures/
│       │   └── book-of-mormon/
│       │       └── 1-nephi/
│       │           ├── chapter-01.md
│       │           └── chapter-02.md
│       └── come-follow-me/
│           └── 2026/
│
├── scripts/
│   └── gospel-library/
│       ├── main.go                      # Entry point
│       ├── cmd/                         # CLI commands
│       │   └── root.go
│       ├── internal/
│       │   ├── api/                     # API client
│       │   │   ├── client.go
│       │   │   └── types.go
│       │   ├── cache/                   # Caching layer
│       │   │   └── cache.go
│       │   ├── convert/                 # HTML→Markdown conversion
│       │   │   ├── converter.go
│       │   │   └── footnotes.go
│       │   └── tui/                     # TUI screens
│       │       ├── app.go
│       │       ├── menu.go
│       │       ├── browser.go
│       │       └── progress.go
│       ├── config.yaml                  # User configuration
│       ├── go.mod
│       └── go.sum
│
└── scripts/plans/
    └── 01_gospel_library_downloader.md  # This document
```

---

## 3. Caching Strategy

### Cache Levels

1. **Catalog Cache** (`.cache/gospel-library/catalog/`)
   - Full content catalog
   - Refresh: Daily or on-demand
   - Used to determine what content exists

2. **Raw Content Cache** (`.cache/gospel-library/raw/`)
   - Original API responses
   - Refresh: Only when content updated (check via catalog metadata)
   - Preserves original data for re-conversion

3. **Sync Metadata** (`.cache/gospel-library/metadata.json`)
   ```json
   {
     "last_catalog_sync": "2026-01-23T10:00:00Z",
     "last_full_sync": "2026-01-20T08:00:00Z",
     "items": {
       "general-conference/2024/10/11nelson": {
         "last_downloaded": "2026-01-20T08:15:00Z",
         "version": "1.0.0",
         "converted": true
       }
     }
   }
   ```

### Cache Invalidation

- Compare catalog `version` or `lastModified` fields
- Only re-download changed content
- Keep old versions for comparison if needed

---

## 4. Conversion Pipeline

### HTML to Markdown Conversion

```
Raw JSON/HTML → Clean HTML → Markdown → Post-process
```

#### Steps:
1. **Extract** content from JSON response
2. **Clean** HTML (remove scripts, styles, tracking)
3. **Convert** to Markdown using html-to-markdown library
4. **Post-process**:
   - Add YAML frontmatter (title, author, date, source URL)
   - Fix scripture references to link to local scriptures
   - Format footnotes/endnotes
   - Extract and note media references

### Markdown Output Format

```markdown
---
title: "Talk Title"
author: "President Russell M. Nelson"
date: 2024-10-05
session: "Saturday Morning Session"
source: "https://www.churchofjesuschrist.org/study/general-conference/2024/10/11nelson"
audio: "https://assets.churchofjesuschrist.org/.../talk.mp3"
downloaded: 2026-01-23
---

# Talk Title

Talk content here with inline footnote markers[^1] that link to the bottom...

More content with another reference[^2] to scripture.

---

## Footnotes

[^1]: This is the footnote content explaining the marked word/phrase.

[^2]: See [1 Nephi 3:7](../scriptures/Book_of_Mormon/01_1_Nephi.md#chapter-3)
```

---

## 4b. Footnote Structure (VERIFIED ✅)

### API Response Structure

The API returns content with **separate `body` and `footnotes` sections**:

```json
{
  "content": {
    "head": "...",
    "body": "<p>...HTML content with footnote links...</p>",
    "footnotes": {
      "note1": { ... },
      "note7_a": { ... }
    }
  }
}
```

### Conference Talk Footnotes

**Body HTML Pattern:**
```html
<p>...the Lord is indeed hastening His work.<a class="note-ref" href="#note3"><sup class="marker" data-value="1"></sup></a></p>
```

**Footnote Object:**
```json
{
  "note3": {
    "id": "note3",
    "marker": "3.",
    "pid": "162706491",
    "context": "",                    // Empty for talks - no specific word marked
    "text": "<p>See <a class=\"scripture-ref\" href=\"/study/scriptures/dc-testament/dc/88?lang=eng&id=p73#p73\">Doctrine and Covenants 88:73</a>.</p>",
    "referenceUris": [
      {
        "type": "scripture-ref",
        "href": "/study/scriptures/dc-testament/dc/88?lang=eng&id=p73#p73",
        "text": "Doctrine and Covenants 88:73"
      }
    ]
  }
}
```

### Scripture Footnotes (More Complex)

**Body HTML Pattern:**
```html
<p class="verse" id="p7">
  <span class="verse-number">7 </span>And it came to pass that I, Nephi, said unto my father: I 
  <a class="study-note-ref" href="#note7_a"><sup class="marker" data-value="a"></sup>will</a> 
  go and do the things which the Lord hath commanded...
</p>
```

The **word following the `<a>` opening tag** is the annotated word (e.g., "will").

**Footnote Object:**
```json
{
  "note7_a": {
    "id": "note7_a",
    "marker": "7a",
    "pid": "128344597",
    "context": "will",              // 🎯 The annotated word is provided!
    "text": "<p><span data-note-category=\"tg\"><a class=\"scripture-ref\" href=\"/study/scriptures/tg/commitment?lang=eng\"><small>TG</small> Commitment</a>.</span></p>",
    "referenceUris": [
      {
        "type": "scripture-ref",
        "href": "/study/scriptures/tg/commitment?lang=eng",
        "text": "TG Commitment"
      }
    ]
  }
}
```

### Key Discovery: `context` Field

Scripture footnotes include a **`context` field** that contains the exact word/phrase being annotated! This eliminates the need for heuristics.

| Content Type | `context` Field | Annotated Word Location |
|--------------|-----------------|-------------------------|
| Conference Talks | Empty (`""`) | Footnote follows end of sentence |
| Scriptures | Populated (e.g., `"will"`) | Word is inside the `<a>` tag |

### Footnote Categories (Scriptures)

The `data-note-category` attribute indicates the type:
- `cross-ref` - Cross-reference to other scriptures
- `tg` - Topical Guide reference
- `bd` - Bible Dictionary
- `gs` - Guide to the Scriptures
- `ie` - Explanation ("i.e.")
- `or` - Alternative reading ("or")
- `heb` - Hebrew translation

### 🎯 Implementation Strategy

**For Conference Talks:**
```markdown
...the Lord is indeed hastening His work.[^3]

[^3]: See [D&C 88:73](../scriptures/Doctrine_and_Covenants/Section_088.md)
```

**For Scriptures (with context):**
```markdown
**7** And it came to pass that I, Nephi, said unto my father: I will[^7a] go and do...

[^7a]: **"will"** — TG [Commitment](../scriptures/tg/commitment.md)
```

### Conversion Logic

```go
type Footnote struct {
    ID            string        `json:"id"`
    Marker        string        `json:"marker"`
    PID           string        `json:"pid"`
    Context       string        `json:"context"`       // The annotated word (scriptures only)
    Text          string        `json:"text"`          // HTML content
    ReferenceURIs []RefURI      `json:"referenceUris"`
}

type RefURI struct {
    Type string `json:"type"`  // "scripture-ref", etc.
    Href string `json:"href"`
    Text string `json:"text"`
}
```

**Steps:**
1. Parse `body` HTML, find all `<a class="note-ref">` or `<a class="study-note-ref">` tags
2. Extract footnote ID from `href` (e.g., `#note7_a` → `note7_a`)
3. Look up footnote in `footnotes` object
4. Use `context` field if present; otherwise place footnote at sentence end
5. Convert footnote `text` HTML to markdown with local scripture links

---

## 5. TUI Application Design

The tool will be a **Terminal User Interface (TUI)** application for interactive content selection.

### Main Flow
```
┌─────────────────────────────────────────────────────────────┐
│  Gospel Library Downloader                          v1.0.0  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. [↻] Sync Catalog    (fetch latest content listing)     │
│  2. [□] Select Content  (browse and checkbox items)        │
│  3. [↓] Download        (download selected to cache)       │
│  4. [M] Convert         (convert cache to markdown)        │
│  5. [⚙] Settings        (language, rate limit, etc.)       │
│  6. [Q] Quit                                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Content Selection Screen
```
┌─────────────────────────────────────────────────────────────┐
│  Select Content to Download                    [Space]=Toggle│
├─────────────────────────────────────────────────────────────┤
│  ▼ General Conference                                       │
│    ▼ 2025                                                   │
│      [x] October 2025 (35 talks)                           │
│      [x] April 2025 (34 talks)                             │
│    ▶ 2024 (collapsed)                                       │
│    ▶ 2023                                                   │
│  ▶ Scriptures                                               │
│    [ ] Book of Mormon (with footnotes)                     │
│    [ ] Doctrine and Covenants                              │
│  ▶ Come, Follow Me                                          │
│    [x] 2026 - Book of Mormon                               │
│  ▶ Books & Manuals                                          │
│                                                             │
│  [Enter] Confirm    [A] Select All    [N] Select None      │
└─────────────────────────────────────────────────────────────┘
```

### Progress Screen
```
┌─────────────────────────────────────────────────────────────┐
│  Downloading...                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [████████████░░░░░░░░] 60%  (42/70 items)                 │
│                                                             │
│  Current: general-conference/2025/10/16uchtdorf            │
│  Speed: 2.1 items/sec                                       │
│  ETA: ~15 seconds                                           │
│                                                             │
│  ✓ Downloaded: 42                                           │
│  ⏭ Skipped (cached): 12                                     │
│  ✗ Errors: 0                                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### CLI Flags (for automation/scripting)
```bash
# Launch TUI (default)
go run ./scripts/gospel-library/...

# Non-interactive modes
go run ./scripts/gospel-library/... --sync              # Just sync catalog
go run ./scripts/gospel-library/... --download-all      # Download everything selected in config
go run ./scripts/gospel-library/... --convert           # Convert all cached content
go run ./scripts/gospel-library/... --lang spa          # Use Spanish
```

### Go TUI Libraries
- [`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) - TUI framework
- [`github.com/charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles) - Pre-built components (lists, spinners, progress bars)
- [`github.com/charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) - Styling

---

## 6. Configuration

### `scripts/gospel-library/config.yaml`
```yaml
# API Settings
api:
  base_url: "https://www.churchofjesuschrist.org/study/api/v3"
  language: "eng"
  rate_limit: 1.0  # requests per second
  user_agent: "ScriptureStudy-Downloader/1.0 (personal study tool)"

# Cache Settings
cache:
  directory: ".cache/gospel-library"
  catalog_ttl: "24h"

# Output Settings
output:
  directory: "resources/gospel-library"
  format: "markdown"

# Content to Download
content:
  include:
    - "general-conference"
    - "come-follow-me"
    # - "manual/gospel-principles"
  exclude:
    - "*/media/*"
  years:
    min: 2020  # Only download from 2020 onwards
    max: null  # Up to current

# Scripture Reference Linking
references:
  link_to_local: true
  scriptures_path: "../scriptures"
```

---

## 7. .gitignore Updates

Add to `.gitignore`:
```gitignore
# Gospel Library Downloads (copyrighted content)
.cache/
resources/gospel-library/

# Keep the scripts and plans
!scripts/gospel-library/
!scripts/plans/
```

---

## 8. Implementation Phases

### Phase 1: Research & Prototype ✅ (Mostly Done)
- [x] Test API endpoints manually (curl/browser)
- [x] Document actual API response structures
- [ ] Verify rate limiting requirements (test with real downloads)
- [ ] Create basic Go HTTP client

### Phase 2: Core Infrastructure
- [ ] Project scaffolding (`./scripts/gospel-library/`)
- [ ] HTTP client with rate limiting
- [ ] Caching layer (raw JSON storage)
- [ ] Sync metadata tracking
- [ ] Image downloader

### Phase 3: Catalog & Navigation
- [ ] Parse collection/dynamic API responses
- [ ] Build content tree structure
- [ ] Identify all content types (talks, scriptures, manuals, etc.)
- [ ] Selection state persistence

### Phase 4: TUI Application
- [ ] Basic bubbletea app structure
- [ ] Main menu screen
- [ ] Content browser with checkboxes
- [ ] Progress/download screen
- [ ] Settings screen

### Phase 5: Content Download
- [ ] Download selected content to cache
- [ ] Download associated images
- [ ] Skip already-cached content
- [ ] Error handling & retry logic

### Phase 6: Markdown Conversion
- [ ] HTML to Markdown converter
- [ ] YAML frontmatter generation
- [ ] Footnote extraction & formatting
- [ ] Scripture reference linking (to local files)
- [ ] Image path rewriting
- [ ] Media link generation (audio/video/PDF)

### Phase 7: Polish & Documentation
- [ ] CLI flags for automation
- [ ] Configuration file support
- [ ] README with usage instructions
- [ ] Error messages and logging

---

## 9. Technical Considerations

### Go Libraries to Use
- `net/http` - HTTP client
- `encoding/json` - JSON parsing
- `github.com/JohannesKaufmann/html-to-markdown` - HTML conversion
- `gopkg.in/yaml.v3` - YAML config/frontmatter
- **TUI Framework:**
  - `github.com/charmbracelet/bubbletea` - TUI framework (Elm architecture)
  - `github.com/charmbracelet/bubbles` - Components (list, progress, spinner)
  - `github.com/charmbracelet/lipgloss` - Styling
- `golang.org/x/time/rate` - Rate limiting

### Error Handling
- Network failures: Retry with exponential backoff
- Rate limiting: Respect 429 responses, back off
- Partial downloads: Track progress, resume capability
- Invalid content: Log and skip, don't fail entire sync

### Performance
- Concurrent downloads (configurable, default 2-3)
- Skip already-downloaded content
- Incremental syncs based on catalog changes

---

## 10. Ethical Considerations

1. **Respect Terms of Service** - Use for personal study only
2. **Rate Limiting** - Don't overwhelm servers
3. **No Redistribution** - Downloaded content stays local (.gitignore)
4. **Attribution** - Always include source URLs in converted content
5. **Copyright Notice** - Add copyright notice to converted files

---

## 11. Open Questions

- [x] ~~Does the API require authentication for any content?~~ **No, public content is accessible**
- [x] ~~What is the actual catalog structure?~~ **Documented above - `collection.sections[].entries[]`**
- [ ] Are there versioning/changelog endpoints for detecting updates? **Unknown - may need to track by comparing content hashes**
- [x] ~~Should we support multiple languages?~~ **Default to English, add `--lang` flag for others**
- [x] ~~How to handle media?~~ **See Media Handling section below**
- [x] ~~Should we integrate with existing `scripts/convert/` structure?~~ **No, separate tool at `./scripts/gospel-library/`**
- [x] ~~How far back does General Conference data go?~~ **1971 - sufficient**
- [x] ~~What other content types are available?~~ **See Content Types section below**
- [x] ~~**Footnote linking strategy**~~ **VERIFIED - API provides `context` field for scriptures with annotated word; talks use sentence-end footnotes. See Section 4b.**

---

## 12. Next Steps

1. ~~**API Exploration** - Use browser dev tools or proxy to capture actual API calls from churchofjesuschrist.org/study~~ ✅ **DONE**
2. **Create Prototype** - Simple Go script to download one talk
3. **Document Findings** - Update this plan with actual API details
4. **Build Incrementally** - Start with General Conference, expand later

---

## Appendix: Sample API Calls

### Get General Conference Collection Structure
```bash
curl -s "https://www.churchofjesuschrist.org/study/api/v3/language-pages/type/dynamic?lang=eng&uri=/general-conference" | jq '.collection.sections[0].entries[:3]'
```

### Get October 2025 Conference Content List
```bash
curl -s "https://www.churchofjesuschrist.org/study/api/v3/language-pages/type/dynamic?lang=eng&uri=/general-conference/2025/10" | jq '.collection'
```

### Get Specific Talk Content
```bash
curl -s "https://www.churchofjesuschrist.org/study/api/v3/language-pages/type/content?lang=eng&uri=/general-conference/2025/10/19oaks" | jq '.meta'
```

---

## References

- [Church of Jesus Christ Study Website](https://www.churchofjesuschrist.org/study)
- [Gospel Library App](https://www.churchofjesuschrist.org/pages/mobileapps/gospellibrary)
- Community reverse engineering efforts (various GitHub projects)
