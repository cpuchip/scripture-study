# Brain — Attachments (Photos, Voice Memos, Files)

*Created: July 2025*
*Status: Draft — **Pushed to far-term** (March 2026). Requires S3/storage infrastructure decision. Focus is on brain-app polish, Today Screen, and proactive surfacing first.*
*Depends on: nothing (standalone data model extension)*

---

## Overview

Attach files to entries — photos, screenshots, voice recordings, PDFs, or any file. Primary use cases:

- **Snap a whiteboard** after a meeting and attach it to the "Project X planning" entry
- **Voice memo** for context — the raw audio that was transcribed, kept alongside the entry
- **Screenshot** of an error, a design, a receipt
- **Photo of a person** — attach to a "people" entry so you remember what they look like

This plan also covers **vision model integration** — using LM Studio's VLM (Vision Language Model) support to classify or describe images automatically.

---

## LM Studio Vision API — Key Research Finding

LM Studio's native REST API (`POST /api/v1/chat`) has **explicit image input support**:

```json
{
  "model": "qwen2-vl-2b-instruct",
  "input": [
    {"type": "message", "content": "Describe this image in detail."},
    {"type": "image", "data_url": "data:image/jpeg;base64,/9j/4AAQ..."}
  ],
  "temperature": 0.3
}
```

**Key details:**
- Input is an array of typed objects: `{"type": "message", "content": "..."}` and `{"type": "image", "data_url": "data:image/...;base64,..."}`
- Supports JPEG, PNG, WebP via base64 data URLs
- Requires a vision-capable model loaded in LM Studio (e.g., `qwen2-vl-2b-instruct`, `llava`, `moondream`)
- The OpenAI-compatible endpoint (`/v1/chat/completions`) may also support the standard `image_url` content block format, but the v1 REST API is the documented path

**Implication for brain.exe:** The current `ai.LMStudioClient` uses `/v1/chat/completions` (OpenAI-compat). For vision, we either:
1. Add a second method that calls `/api/v1/chat` (LM Studio native) for image inputs, or
2. Extend the OpenAI-compat call to include `image_url` content blocks (if LM Studio supports it)

Option 1 is more reliable since the native API has documented image support.

---

## Data Model

### File Storage Strategy

**Filesystem alongside SQLite, not blob storage in the DB.**

```
brain-data/
├── brain.db
└── attachments/
    ├── abc123.jpg
    ├── def456.webm
    └── ghi789.pdf
```

- Attachment files stored as `{attachment_id}.{ext}` in an `attachments/` directory next to the database
- SQLite stores metadata only (filename, mime type, size, entry association)
- This keeps the DB fast and backupable while allowing large files

### SQLite (brain.exe)

```sql
CREATE TABLE IF NOT EXISTS attachments (
    id          TEXT PRIMARY KEY,
    entry_id    TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    filename    TEXT NOT NULL,          -- original filename from upload
    mime_type   TEXT NOT NULL,          -- image/jpeg, audio/webm, etc.
    size_bytes  INTEGER NOT NULL,
    description TEXT,                   -- AI-generated or user-provided description
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attachments_entry ON attachments(entry_id);
```

### Go Types (store/types.go)

```go
type Attachment struct {
    ID          string    `json:"id"`
    EntryID     string    `json:"entry_id"`
    Filename    string    `json:"filename"`
    MimeType    string    `json:"mime_type"`
    SizeBytes   int64     `json:"size_bytes"`
    Description string    `json:"description,omitempty"`
    Created     time.Time `json:"created_at"`
}
```

Add to `Entry`:
```go
Attachments []Attachment `json:"attachments,omitempty" yaml:"attachments,omitempty"`
```

### Dart Model (brain-app)

```dart
class Attachment {
  final String id;
  final String entryId;
  final String filename;
  final String mimeType;
  final int sizeBytes;
  final String? description;

  Attachment({
    required this.id,
    required this.entryId,
    required this.filename,
    required this.mimeType,
    required this.sizeBytes,
    this.description,
  });

  bool get isImage => mimeType.startsWith('image/');
  bool get isAudio => mimeType.startsWith('audio/');

  factory Attachment.fromJson(Map<String, dynamic> json) => Attachment(
    id: json['id'] ?? '',
    entryId: json['entry_id'] ?? '',
    filename: json['filename'] ?? '',
    mimeType: json['mime_type'] ?? '',
    sizeBytes: json['size_bytes'] ?? 0,
    description: json['description'],
  );
}
```

Add to `HistoryEntry`:
```dart
final List<Attachment> attachments;
```

---

## API Changes (brain.exe)

### Attachment Endpoints

```
POST   /api/entries/{id}/attachments        → upload file (multipart/form-data)
GET    /api/entries/{id}/attachments         → list attachments for entry
GET    /api/attachments/{aid}/file           → download/serve the actual file
DELETE /api/entries/{id}/attachments/{aid}   → delete attachment + file
POST   /api/attachments/{aid}/describe       → trigger VLM description (images only)
```

### Upload Handler

```go
func (s *Server) handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form (max 20MB)
    // Validate mime type against allowlist
    // Generate UUID for attachment ID
    // Save file to attachments/{id}.{ext}
    // Insert metadata into attachments table
    // Return attachment JSON
}
```

**Security:**
- Allowlisted mime types: `image/jpeg`, `image/png`, `image/webp`, `image/gif`, `audio/webm`, `audio/mp4`, `audio/mpeg`, `application/pdf`
- Max file size: 20MB (configurable)
- Files served with `Content-Disposition: inline` for images, `attachment` for others
- No path traversal — files named by UUID only

### File Serving

```go
func (s *Server) handleServeAttachment(w http.ResponseWriter, r *http.Request) {
    // Look up attachment metadata
    // Set Content-Type from metadata
    // Serve file from disk
}
```

---

## Vision Model Integration

### Architecture

```
                                  ┌─────────────────┐
  image upload                    │   LM Studio      │
  ────────────►  brain.exe  ────► │   VLM model      │
                  │               │ (qwen2-vl-2b)    │
                  │               └────────┬──────────┘
                  │                        │
                  ◄────────────────────────┘
                  │  "A whiteboard showing
                  │   project timeline..."
                  ▼
              attachment.description = AI description
              + optionally update entry title/tags
```

### New AI Method

Add to `ai.LMStudioClient`:

```go
// CompleteWithImage sends a chat request with an image to LM Studio's native REST API.
// Uses POST /api/v1/chat with typed input array (message + image).
func (c *LMStudioClient) CompleteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (string, error) {
    base64Data := base64.StdEncoding.EncodeToString(imageData)
    dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

    payload := map[string]interface{}{
        "model": c.Model(),
        "input": []map[string]interface{}{
            {"type": "message", "content": prompt},
            {"type": "image", "data_url": dataURL},
        },
        "temperature": 0.3,
    }
    // POST to /api/v1/chat (native LM Studio endpoint)
    // Parse response.output[].content
}
```

**Note:** This requires a vision-capable model to be loaded. brain.exe should detect VLM availability (check if a VLM profile is registered) and gracefully skip image description when no VLM is available.

### Model Profile Extension

Add a `Vision` task type:

```go
const TaskVision Task = "vision"

// Register VLM models:
RegisterProfile(&ModelProfile{
    ID:    "qwen/qwen2-vl-2b-instruct",
    Name:  "Qwen2 VL 2B",
    Tasks: []Task{TaskVision},
    Temperature: 0.3,
})
```

### Describe Endpoint

`POST /api/attachments/{aid}/describe`:
1. Load attachment file from disk
2. Check if a VLM model profile exists
3. Call `CompleteWithImage` with prompt: "Describe this image concisely. What does it show?"
4. Save description to `attachments.description`
5. Return updated attachment metadata

**Auto-describe on upload** (optional): If a VLM is available and the upload is an image, automatically trigger description. Can be a config flag.

---

## Flutter UI

### Entry Detail — Attachment Display

```
┌──────────────────────────────────┐
│ Title: Project X Planning        │
│ Body: Meeting notes from...      │
│                                  │
│ Attachments                      │
│ ┌────────┐  ┌────────┐          │
│ │ 📷     │  │ 🎤     │          │
│ │ white- │  │ meeting │          │
│ │ board  │  │ -audio  │          │
│ └────────┘  └────────┘          │
│ [+ Attach file]                  │
└──────────────────────────────────┘
```

- **Images**: thumbnail grid with tap-to-fullscreen (zoom/pan)
- **Audio**: inline player with play/pause and duration
- **Other files**: icon + filename, tap to open in external app
- **Description**: shown below image thumbnail if AI description exists

### Attach Button

Opens a bottom sheet with options:
- 📷 Take photo (camera)
- 🖼 Choose from gallery
- 🎤 Record voice memo
- 📎 Pick file

Uses `image_picker` for camera/gallery, `file_picker` for files, and `record` (or similar) for voice memos.

### Upload Flow

1. User picks file → show upload progress indicator on the entry
2. Multipart POST to `/api/entries/{id}/attachments`
3. On success, add attachment to local entry state
4. If image + VLM available, show "AI is describing your image..." indicator
5. When description arrives, update attachment display

### Direct vs. Relay Mode

- **Direct mode**: upload directly to brain.exe
- **Relay mode**: upload to ibeco.me → relay to brain.exe (or store in ibeco.me if brain.exe is offline)

**Relay consideration:** Large file uploads (photos, voice memos) need multipart support in the relay. ibeco.me would need to either:
1. Forward the multipart upload to brain.exe via WebSocket (complex — binary frames)
2. Store attachments on ibeco.me's disk and sync to brain.exe later
3. **Direct-only for v1** — attachments only work in direct mode. Simplest path.

**Recommendation:** Start with direct-only uploads. Add relay support later as a separate enhancement.

---

## Implementation Phases

### Phase 1: Data & Storage
1. Add `attachments` table migration
2. `Attachment` Go type and DB methods (Insert, List, Get, Delete)
3. File storage helper (save to disk, read from disk, delete from disk)
4. REST endpoints: upload, list, serve, delete
5. Security: mime type allowlist, size limit, path safety

### Phase 2: Flutter UI (Images)
1. `Attachment` Dart model, update `HistoryEntry`
2. Image picker integration (camera + gallery)
3. Upload with progress indicator
4. Thumbnail grid in entry detail
5. Full-screen image viewer (zoom/pan)
6. API methods in BrainApi

### Phase 3: Vision Model
1. `CompleteWithImage` method on LMStudioClient (uses `/api/v1/chat`)
2. VLM model profile (TaskVision)
3. `/api/attachments/{aid}/describe` endpoint
4. Auto-describe on upload (configurable)
5. Display AI description in Flutter UI

### Phase 4: Audio & Other Files
1. Voice memo recording in Flutter
2. Inline audio player widget
3. File picker for arbitrary files
4. Generic file display (icon + name + size)

### Phase 5: Relay Support (later)
1. ibeco.me attachment storage
2. Sync attachments to brain.exe when online
3. Relay-mode upload flow

---

## Open Questions

- **Thumbnail generation?** Generate thumbnails server-side for faster list loading? Or let Flutter handle sizing? Start with Flutter-side resize, add server thumbnails if performance is a problem.
- **Voice memo transcription?** If a voice memo is attached, should brain.exe transcribe it using Whisper (or similar)? This would be powerful — attach audio, get text. But it's a separate feature. Note in body: "See voice memo" with auto-transcription as a future enhancement.
- **Attachment size for relay?** ibeco.me is on a VPS with limited disk. Need size quotas if relay-mode attachments land on ibeco.me.
- **Markdown cross-reference?** Rich text body (Plan 11) could reference attachments: `![whiteboard](attachment://abc123)`. Worth designing for but not blocking on.
