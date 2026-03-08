# Proactive Surfacing & Daily Digest

*Created: March 8, 2026*
*Priority: Near-term (features 1-3), Mid-term (feature 4 - semantic connections)*
*Depends on: brain.exe chromem-go vectors (already working), ibeco.me relay (already working)*
*This is the "why" for the brain project — helping you remember things over time.*

---

## Vision

The brain captures thoughts. Classification organizes them. But without **surfacing**, entries slowly become a graveyard of good intentions. Proactive surfacing makes brain a living system — it *remembers for you* and brings things back at the right time.

> "The value of a second brain isn't in storing — it's in surfacing the right thing at the right moment."

---

## Feature 1: Actions Due Today / Overdue

### What

Query brain entries with `due_date <= today`. Show them on the Today Screen (Plan 16) and in the daily digest notification.

### Implementation

**brain.exe** — already stores `due_date` in entries. Add a query:

```sql
SELECT * FROM entries 
WHERE user_id = ? AND due_date <= ? AND status != 'archived'
ORDER BY due_date ASC
```

**ibeco.me** — new endpoint or relay message:
- `GET /api/brain/entries/due?before={date}` — queries cached brain entries by due date
- Or relay message: `{"type": "entries_due_request", "before": "2026-03-08"}`

**brain-app Today Screen** — Already designed in Plan 16 as "Brain Actions" section.

### Effort: Small

---

## Feature 2: Stale People

### What

People you captured but haven't interacted with recently. "You mentioned Josh 3 weeks ago — want to follow up?"

### Implementation

**brain.exe** — query entries with category `person` (or `people`) where `updated_at` is older than N days:

```sql
SELECT * FROM entries 
WHERE user_id = ? AND category = 'person' AND status != 'archived'
  AND updated_at < datetime('now', '-14 days')
ORDER BY updated_at ASC
LIMIT 5
```

**Surfacing**: Show in daily digest as "People you haven't touched base with" section. Tap → opens the entry for review/update.

**Configuration**: Default threshold 14 days. Could make it configurable per entry (some people you check in with weekly, others monthly).

### Effort: Small

---

## Feature 3: Incomplete Subtasks Aged N Days

### What

Entries with unchecked subtasks that have been sitting for a while. "You started a project plan 10 days ago — 3 subtasks still undone."

### Implementation

**brain.exe**:

```sql
SELECT e.* FROM entries e
WHERE e.user_id = ? AND e.status != 'archived'
  AND EXISTS (
    SELECT 1 FROM sub_items si 
    WHERE si.entry_id = e.id AND si.done = 0
  )
  AND e.updated_at < datetime('now', '-7 days')
ORDER BY e.updated_at ASC
LIMIT 5
```

**Surfacing**: Show in daily digest as "Stalled projects" with count of remaining subtasks.

### Effort: Small

---

## Feature 4: Semantic Connections

### What

When you capture a new thought, brain searches its vector store for related past entries and surfaces them. "You wrote about covenant-keeping today — here's what you captured about covenants 6 weeks ago."

This is the **core differentiator**. This is what makes brain more than a note app. This is the *why*.

### How It Works

```
1. User captures new thought → brain.exe classifies it
2. After classification, brain.exe does vector search:
   - Embed the new entry's text with chromem-go
   - Search existing vectors for top-K similar entries (K=5, threshold=0.7)  
   - Filter out the entry itself and very recent entries (< 3 days old)
3. brain.exe stores connections as metadata on the entry:
   {"connections": [{"id": "abc", "title": "Covenants study", "similarity": 0.85}]}
4. brain.exe sends entry_updated with connections to relay
5. brain-app shows connections in the entry detail view
6. Daily digest surfaces "today's most interesting connections"
```

### UI: Connections in Entry Detail

In EditEntryScreen, below the subtask list:

```
┌─────────────────────────────────┐
│  🔗 Related Entries             │
│  ──────────────────────────     │
│  Covenants study (85% match)   │  ← Tap to open
│  Temple notes from Jan (79%)   │
│  D&C 84 reflection (72%)       │
└─────────────────────────────────┘
```

### UI: Daily Digest Connections

The morning digest highlights the most interesting connections from yesterday's entries:

```
🧠 Yesterday's Connections
───────────────────────
Your note about [patience with kids] connects to what you wrote about 
[Alma 32 — planting seeds] 3 weeks ago. Might be worth revisiting.
```

This is where the brain becomes a **companion** — not just storing, but actively connecting your thoughts into a web of meaning over time.

### Implementation Steps

1. **brain.exe: Post-classify vector search** — After `Classify()` returns, embed the entry text and search chromem-go for similar entries. Already have `VectorSearch()` in brain.exe.

2. **brain.exe: Store connections** — Add a `connections` field to entry metadata (JSON array of `{id, title, similarity}`). Or a separate `entry_connections` table if we want to query bidirectionally.

3. **Relay: Forward connections** — `entry_updated` payload already includes full entry JSON. Connections come along for free.

4. **brain-app: Show connections** — New "Related Entries" section in EditEntryScreen. Each item tappable → navigates to that entry.

5. **Daily digest: Surface best connections** — Morning digest query includes entries from yesterday that have connections with similarity > 0.8.

### Effort: Medium (the vector search already works; wiring it into the classify flow + UI is the work)

---

## Daily Digest Delivery

### Phase 1: In-App Digest (Near-term)

When user opens the Today Screen, the digest is computed and displayed at the top:

```
Good morning! Here's your brain today:
• 3 actions due (1 overdue)
• Josh — haven't connected in 18 days
• "Faith study" has 2 stalled subtasks
• Your note about patience connects to your Alma 32 study from 3 weeks ago
```

### Phase 2: Push Notification (Mid-term)

brain.exe compiles the digest and sends it through the relay to brain-app as a push notification. User sees a summary notification at their configured time (default: 7:00 AM).

Implementation requires:
- brain.exe scheduler (cron-like, runs digest queries at configured time)
- New relay message type: `daily_digest`
- brain-app notification handler (Android notification channel)

### Phase 3: Morning Email/Widget (Far-term)

- Android widget shows top 3 digest items
- Optional email digest for web-only users

---

## Scripture Memorization + Brain (Planning Task)

**Saved for future planning after Today Screen ships** (needs the ibeco.me ↔ brain bridge first).

Ideas to explore:
- When reviewing a memorize card, brain surfaces *other entries* related to that scripture's theme
- After a memorize session, brain captures "what I learned/connected" and links it back to the practice
- Spaced repetition intervals informed by how often the concept appears across brain entries (frequently referenced = better retention = longer intervals)
- Brain suggests *new* scriptures to memorize based on themes you're actively studying
- Vector search across scripture text + brain entries to find "you wrote about X, which connects to Y scripture you're memorizing"

This is where brain becomes a true **study companion** — not just storing your thoughts, but connecting them to your scripture learning in ways you wouldn't see on your own.

---

## Summary

| Feature | Effort | Priority | Depends On |
|---------|--------|----------|------------|
| 1. Due today/overdue | Small | Near-term | Today Screen (Plan 16) |
| 2. Stale people | Small | Near-term | Today Screen (Plan 16) |
| 3. Stalled subtasks | Small | Near-term | Today Screen (Plan 16) |
| 4. Semantic connections | Medium | Near-term | chromem-go (already working) |
| Daily digest (in-app) | Small | Near-term | Features 1-4 |
| Push notifications | Medium | Mid-term | Android notification setup |
| Scripture + brain synergy | TBD | Mid-term | Today Screen + memorize bridge |
