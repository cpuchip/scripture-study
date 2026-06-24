# Rich-docs P2 — attachments in chat (the MVP: upload + inline render + grounded answer)

**Date:** 2026-06-24
**Session:** pg-ai-stewards lane (continuing the rich-docs arc from P1)
**Michael:** "lets roll! pick up P1+" + (earlier) "make sure our UI supports these rich media types on chat display too"

## What shipped — the first real win

Attach an image to a Stewdio chat as injectable subject material; a vision model
reasons over it; the image renders inline in the conversation. The "work item
AND this image" shape from the spec.

**Substrate — `extension/48-chat-attachments.sql` (chain → 48):**
- `chat_attachments` table — durable, session-scoped, bytea + mime + kind
  (image/document) + extracted_text (P3) + byte_size + consumed_at. No FK to
  sessions (an upload can precede the first turn).
- `chat_attachment_parts(ids, session_id)` — assembles the 47 content_parts
  array from the stored bytes: an image → an `image_url` part with an inline data
  URL built **server-side** from the bytea (the base64 never round-trips through
  the app — same "read by handle, don't re-emit" discipline as page-in / the book
  corpus). Session-scoped (no cross-session injection), marks consumed_at, returns
  NULL when nothing resolves (text-only fallback).

**Go API — `cmd/stewards-ui/api/chat.go`:**
- `POST /api/chat/attach` (multipart) — stores the upload, returns {id, url, …}.
  25 MB cap, mime sniff, kind = image when image/*.
- `GET /api/chat/attachment/{id}` — serves the bytes (inline render source).
- `/api/chat/send` gains `attachment_ids` — when present, dispatch_chat_turn's
  6th arg becomes `chat_attachment_parts(ids, session)` → the vision alias fires
  (47). No attachments → NULL → text-only path unchanged.
- The SSE stream now emits `images[]` for a row (the image_url URLs from
  content_parts) so history re-renders inline on reconnect.

**UI — `ChatPanel.vue` + `api.ts`:**
- A 📎 attach button + hidden file input (accept image/*, multiple); staged
  thumbnails with remove; upload-on-send (uploads under the active session, then
  passes the ids to chatSend). A media-only turn (no text) is allowed.
- Inline rendering: images show in the user bubble (live via object URLs;
  on reload via the stream's `images[]` data URLs).

## Proven at every layer

- **OK 37** (virgin-smoke): table + chat_attachment_parts (session-scoped, the
  data-URL image_url part, consumed-marking, NULL-when-empty).
- **API e2e** (curl): upload the First Orbit screenshot → send with the id →
  gemma-4-26b-a4b (vision alias) quoted ALT 0m / Periapsis −600.0 km / TWR 0.00
  verbatim. `GET /api/chat/attachment/N` serves the bytes (200, image/png).
- **Full UI e2e** (playwright): selected the On Liberty doc → 📎 → uploaded the
  screenshot (staged thumbnail rendered) → sent → the image rendered inline in the
  user bubble and the assistant answered grounded in it (same telemetry verbatim).
  Screenshot `projects/pg-ai-stewards-oss/.playwright-cli/p2-image-chat.png`.

## The bug the e2e caught (verify-under-real-conditions)

The first API e2e errored: `chat HTTP 400: Failed to load image or audio file`.
**Root cause:** PostgreSQL's `encode(bytea, 'base64')` MIME-wraps at 76 chars with
newlines — a data URL with embedded `\n` fails to load. P1's proof used Python
base64 (unwrapped), so it never surfaced; only the real upload→bytea→encode path
did. Fix: `translate(encode(...), E'\n\r', '')`. Re-ran green. Lesson: the encode
path differs from how the bytes are produced — test the actual production path,
not a hand-built proxy.

## Scope calls / carry-forward

- **P2 = work-item chats** (attach an image while chatting about a work item). The
  spec's "empty AND work-item chats" — the empty-chat *entry point* rides P3's
  corpus/project lens picker (the natural place for a no-target chat). Surfaced,
  not silently dropped.
- Cost: history re-sends the image's content_parts every turn (the data URL is in
  the stored message); fine for MVP. Optimizations (cap old images; serve by id
  instead of base64-over-SSE for history) are follow-ups.
- `chat_attachments` has no retention/cleanup yet (spec §6 DB-bloat note) — P3+.
- **NEXT = P3** (documents + corpus-as-lens): PDF/office extraction (sandboxed —
  CVE surface) into `extracted_text`/chunks, and the empty-chat corpus/project
  picker. Then P4 (start_task carries the attachment + work_item into spawned work).

Autonomy stayed PAUSED. The digester-reads-our-repos council item remains OPEN
(deferred until sandboxed; dominion_in_council).
