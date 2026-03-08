# Brain — Rich Text / Markdown Body

*Created: July 2025*
*Status: **Done** (March 2026 — Edit/Preview toggle shipped in brain-app)*
*Depends on: nothing (body is already a text field)*

---

## Overview

Entry bodies are plain text today. Add markdown support so entries can have headers, bold, italics, lists, links, and code blocks. This is a **display and editing** feature — the storage is already compatible (body is `TEXT` in SQLite, `string` in Go/Dart).

**Design principle:** Markdown is progressive enhancement. Plain text entries stay valid. The editor should feel like writing text, not using a word processor. No toolbar unless the user wants it — keyboard shortcuts and inline syntax first.

---

## What Changes (and What Doesn't)

### Stays the same
- SQLite `body TEXT` column — no schema change
- Go `Entry.Body string` — no type change
- API request/response format — body is still a JSON string
- Classifier input — still receives raw text (stripped of markdown formatting)
- Search — FTS5 indexes the raw markdown text, which is fine (search matches content words regardless of formatting)

### Changes
- **Flutter**: render body as markdown in read/display contexts
- **Flutter**: markdown-aware editor for editing contexts
- **brain.exe web UI**: render markdown in the entry detail view (if/when the web UI is enhanced)
- **Archive export**: already writes `.md` files — markdown body is a natural fit

---

## Flutter Implementation

### Display: Markdown Rendering

Use `flutter_markdown` (or `markdown_widget`) to render body content:

```dart
// In entry detail / read mode
MarkdownBody(
  data: entry.text,
  selectable: true,
  styleSheet: MarkdownStyleSheet.fromTheme(Theme.of(context)).copyWith(
    p: Theme.of(context).textTheme.bodyMedium,
  ),
)
```

Where markdown appears:
- **Entry detail view** (read mode)
- **Entry list cards** — first line preview, rendered as plain text (strip markdown for preview)
- **Home screen cards** — plain text preview (no rendering in cards)

### Editing: Markdown Editor

Two options, increasing complexity:

#### Option A: Plain text + preview toggle (recommended for v1)

Keep the existing `TextField` for editing. Add a toggle button to switch between **Edit** (raw text) and **Preview** (rendered markdown).

```
┌──────────────────────────────────┐
│ [Edit] [Preview]                 │
│                                  │
│ ## Grocery thoughts              │
│                                  │
│ I've been thinking about **meal  │
│ planning** more seriously.       │
│                                  │
│ Key principles:                  │
│ - Buy ingredients, not meals     │
│ - Always have basics on hand     │
│ - Check what's expiring first    │
└──────────────────────────────────┘
```

This is the simplest path:
- No new packages for editing
- The user writes markdown directly (many people already think in markdown)
- Preview verifies formatting

#### Option B: Toolbar-assisted editor (v2)

Add a formatting toolbar above the keyboard:

```
┌──────────────────────────────────┐
│ B  I  ~S~  H  •  1.  ""  <>  🔗 │
└──────────────────────────────────┘
```

- **B** bold, **I** italic, **~S~** strikethrough
- **H** heading cycle (##, ###)
- **•** bullet list, **1.** numbered list
- **""** blockquote, **<>** code block
- **🔗** insert link

Each button wraps selected text or inserts markers at cursor. This mirrors how Obsidian mobile works.

Packages to evaluate:
- `fleather` — rich text editor that can export to markdown
- `super_editor` — more powerful but heavier
- Or just build a simple toolbar that manipulates the TextField controller

### Keyboard Shortcuts (physical keyboard)

For tablets / desktop:
- `Ctrl+B` → bold
- `Ctrl+I` → italic
- `Ctrl+Shift+H` → heading

---

## Classifier Interaction

The classifier receives body text for reclassification. Two approaches:

1. **Strip markdown before classify** — send plain text to the LLM. Simpler, avoids markdown syntax confusing the classifier.
2. **Send markdown as-is** — LLMs handle markdown fine. The classifier prompt already says "raw text."

**Recommendation:** Send as-is. Modern LLMs understand markdown natively. Stripping would lose structural information (e.g., a list is different from a paragraph).

---

## Search Interaction

FTS5 indexes the raw body text. Markdown syntax tokens (`**`, `##`, `-`) appear in the index but rarely match user queries. No changes needed — search naturally works with markdown content.

If noise becomes a problem later, add a pre-processing step that strips markdown before indexing. Not worth doing preemptively.

---

## Implementation Phases

### Phase 1: Read-only rendering
1. Add `flutter_markdown` package
2. Render body as markdown in entry detail view
3. Strip markdown for list card previews (first line plain text)
4. Verify existing entries render correctly (plain text is valid markdown)

### Phase 2: Edit + preview toggle
1. Add Edit/Preview toggle to EditEntryScreen
2. Edit mode: existing TextField (no change)
3. Preview mode: MarkdownBody rendering
4. Keyboard shortcut support for physical keyboards

### Phase 3: Formatting toolbar (optional, later)
1. Build toolbar widget with formatting buttons
2. TextField controller manipulation for wrap/insert
3. Position above keyboard using `MediaQuery.viewInsets`

---

## Open Questions

- **Voice capture → markdown?** When voice-captured text is classified, should the classifier format the body as markdown? Probably not by default — voice text is stream-of-consciousness. But the classifier could use markdown for structured outputs (lists, action items). Experiment.
- **Image links?** Once the Attachments feature ships, markdown body could reference images via `![alt](attachment://id)`. Cross-reference with Plan 13.
- **Web UI parity?** If brain.exe's web UI grows, it'll need markdown rendering too. Use a Go template with goldmark or similar. Low priority — the web UI is minimal today.
