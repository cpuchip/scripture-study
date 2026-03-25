# Brain-App Polish — Gaps 1-4

*Created: March 8, 2026*
*Priority: Near-term (next 1-2 sessions)*
*Depends on: auto-refresh (entry_updated) already shipped*
*Status: **DONE** — All 4 gaps shipped. Offline caching layer added March 2026.*

---

## Gap 1: Entry Sync on App Launch

### Problem

When brain-app connects to ibeco.me, it fetches history via REST (`GET /api/brain/entries`). But if brain.exe classified entries while the app was closed, the WS connection is now live but the REST data is stale. The user sees old data until they pull-to-refresh.

### Solution

On WebSocket `auth_ok`, request an `entries_sync` from ibeco.me's cache. The protocol already supports this — ibeco.me has `handleEntriesSync()` in hub.go that caches synced entries. We just need the app to request the latest.

### Implementation

**brain_service.dart** — after `auth_ok`:
```dart
case MessageType.authOk:
  _setState(BrainConnectionState.connected);
  _reconnectAttempt = 0;
  _startPingTimer();
  // Request cached entries from relay
  _send({'type': 'entries_request', 'limit': 50});
  break;
```

**brain_service.dart** — new message type + handler:
```dart
static const entriesResponse = 'entries_response';

// In _onMessage switch:
case MessageType.entriesResponse:
  final entries = json['entries'] as List?;
  if (entries != null) {
    onEntriesSync?.call(entries.cast<Map<String, dynamic>>());
  }
  break;
```

**ibeco.me hub.go** — new `entries_request` handler:
- App sends `{"type": "entries_request", "limit": 50}`
- Hub queries `db.ListBrainEntries(userID, limit)` (already exists)
- Hub sends back `{"type": "entries_response", "entries": [...]}`

**home_screen.dart** — merge into HistoryScreen:
- `onEntriesSync` callback updates the cached entry list
- If HistoryScreen is open, it refreshes via the existing `entryUpdated` stream (or a new `entriesSync` stream)

### Effort: Small (1 session)

### Files touched:
- `scripts/becoming/internal/brain/hub.go` — new `entries_request` case in message switch
- `scripts/becoming/internal/brain/messages.go` — new message type constant
- `scripts/brain-app/lib/services/brain_service.dart` — send request on auth_ok, handle response
- `scripts/brain-app/lib/screens/home_screen.dart` — wire callback
- `scripts/brain-app/lib/screens/history_screen.dart` — merge synced entries

---

## Gap 2: Relay Subtask Error Recovery

### Problem

In relay mode, subtask operations (toggle, create, delete) return `{"status": "queued"}`. The app does an optimistic UI update. If brain.exe is offline and the message sits in the queue, or if the operation fails on the agent side, the app's UI diverges from reality. There's no feedback loop.

### Solution

Track optimistic operations by ID. When `entry_updated` arrives with the full entry payload (including subtasks), reconcile the optimistic state with the server state. If the operation failed (subtask state didn't change), revert the optimistic update and show a snackbar.

### Implementation

**brain_api.dart** — add an optimistic operation tracker:
```dart
// Track pending relay operations
final Map<String, _PendingOp> _pendingOps = {};

class _PendingOp {
  final String entryId;
  final String subtaskId;
  final String operation; // 'toggle', 'create', 'delete'
  final DateTime timestamp;
  final SubTask? previousState; // For rollback
}
```

**edit_entry_screen.dart** — in `_onEntryUpdated`:
```dart
// After applying server state, check for pending ops that are now resolved
// If a subtask we optimistically toggled is still in its old state in the
// server payload, the operation failed — show error snackbar
```

**Timeout**: If no `entry_updated` arrives within 30 seconds after a relay subtask op, show a warning: "Change may not have been saved — pull to refresh."

### Effort: Medium (needs careful state tracking)

### Files touched:
- `scripts/brain-app/lib/services/brain_api.dart` — pending ops tracker
- `scripts/brain-app/lib/screens/edit_entry_screen.dart` — reconciliation logic

---

## Gap 3: Classify Flow Polish

### Problem

In relay mode, pressing "AI Classify" shows: *"Classification requested — refresh to see results"*. But now that auto-refresh is live, the user doesn't need to manually refresh. Also, the `_classifying` spinner stops immediately even though the actual classification hasn't happened yet.

### Solution

1. Update the snackbar text to: *"Classification requested — update will appear automatically"*
2. Keep `_classifying = true` until either `entry_updated` arrives (success) or 30s timeout (fallback)
3. When `_onEntryUpdated` fires and category/title changed, show the existing success snackbar and clear `_classifying`

### Implementation

**edit_entry_screen.dart** — in the relay branch of `_classify()`:
```dart
// Relay mode: queued — result arrives async via entry_updated
// Keep _classifying true — _onEntryUpdated will clear it
setState(() {}); // don't clear _classifying
ScaffoldMessenger.of(context).showSnackBar(
  const SnackBar(
    content: Text('Classification requested — update will appear automatically'),
    behavior: SnackBarBehavior.floating,
  ),
);

// Safety timeout — clear spinner after 30s if no update arrives
Future.delayed(const Duration(seconds: 30), () {
  if (mounted && _classifying) {
    setState(() => _classifying = false);
  }
});
```

**edit_entry_screen.dart** — in `_onEntryUpdated()`:
```dart
// Clear classifying state when server update arrives
_classifying = false;
```

(This is already partially handled — `_classifying = false` is set in the non-dirty branch of `_onEntryUpdated`.)

### Effort: Small (15 minutes)

### Files touched:
- `scripts/brain-app/lib/screens/edit_entry_screen.dart` — snackbar text, spinner lifecycle, timeout

---

## Gap 4: Delete with Undo Toast

### Problem

Deleting an entry in HistoryScreen is permanent and immediate — no undo. Archive has an undo toast (great!), but delete doesn't.

### User Request

Keep the confirmation dialog (user wants awareness that this is a delete), BUT add an undo toast after confirmation as a double-check safety net.

### Solution

1. **Keep the existing delete confirmation dialog** — "Delete this entry?" with Cancel/Delete buttons
2. After user confirms, **soft-delete** — remove from local list immediately, show undo toast
3. If undo is tapped, re-insert the entry into the local list (it was never actually deleted)
4. If undo toast expires (5 seconds), fire the actual `deleteEntry()` API call
5. If the API call fails, re-insert the entry and show an error snackbar

### Implementation

**history_screen.dart** — replace `_deleteEntry`:
```dart
Future<void> _deleteEntry(HistoryEntry entry) async {
  // Show confirmation dialog (user's request: keep this)
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('Delete entry?'),
      content: Text('This will permanently delete "${entry.title ?? 'this entry'}"'),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
        TextButton(
          onPressed: () => Navigator.pop(ctx, true),
          child: const Text('Delete', style: TextStyle(color: Colors.red)),
        ),
      ],
    ),
  );
  if (confirmed != true) return;

  // Optimistic remove from local list
  final idx = _entries?.indexOf(entry) ?? -1;
  setState(() => _entries?.remove(entry));

  // Show undo toast — actual delete fires after toast expires
  bool undone = false;
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      content: Text('Deleted "${entry.title ?? 'entry'}"'),
      behavior: SnackBarBehavior.floating,
      duration: const Duration(seconds: 5),
      action: SnackBarAction(
        label: 'Undo',
        onPressed: () {
          undone = true;
          setState(() {
            if (idx >= 0 && idx <= (_entries?.length ?? 0)) {
              _entries?.insert(idx, entry);
            } else {
              _entries?.insert(0, entry);
            }
          });
        },
      ),
    ),
  ).closed.then((_) async {
    if (undone) return;
    // Toast expired without undo — actually delete
    try {
      await widget.api.deleteEntry(entry.id);
    } catch (e) {
      // Delete failed — re-insert
      if (mounted) {
        setState(() => _entries?.insert(0, entry));
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Delete failed: $e'), behavior: SnackBarBehavior.floating),
        );
      }
    }
  });
}
```

### Effort: Small (straightforward)

### Files touched:
- `scripts/brain-app/lib/screens/history_screen.dart` — rewrite `_deleteEntry` with dialog + undo toast pattern
