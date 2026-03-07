# Brain — Sub-tasks / Checklists

*Created: July 2025*
*Status: Draft*
*Depends on: nothing (standalone data model extension)*

---

## Overview

Add checkable sub-items under any entry. Primary use case: a grocery list where each item can be checked off individually, or a project with discrete steps. Sub-tasks are lightweight — a text label, a done state, and an optional quantity/note.

**Design principle:** Sub-tasks belong to entries, not the other way around. An entry is the unit of thought. Sub-tasks are internal structure *within* that thought. Don't over-engineer — this is a checklist, not a project management tool.

---

## Data Model

### SQLite (brain.exe)

```sql
CREATE TABLE IF NOT EXISTS subtasks (
    id         TEXT PRIMARY KEY,
    entry_id   TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    done       INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subtasks_entry ON subtasks(entry_id);
```

**Why a separate table, not JSON in a column?**
- Enables SQL queries: "show me all entries with incomplete sub-tasks"
- Per-item update without rewriting the whole blob
- Proper cascade delete
- Future: potential for cross-entry sub-task references

### Go Types (store/types.go)

```go
type SubTask struct {
    ID        string    `json:"id"`
    EntryID   string    `json:"entry_id"`
    Text      string    `json:"text"`
    Done      bool      `json:"done"`
    SortOrder int       `json:"sort_order"`
    Created   time.Time `json:"created_at"`
    Updated   time.Time `json:"updated_at"`
}
```

Add to `Entry`:
```go
SubTasks []SubTask `json:"subtasks,omitempty" yaml:"subtasks,omitempty"`
```

### Dart Model (brain-app)

```dart
class SubTask {
  final String id;
  final String entryId;
  final String text;
  final bool done;
  final int sortOrder;

  SubTask({required this.id, required this.entryId, required this.text, this.done = false, this.sortOrder = 0});

  factory SubTask.fromJson(Map<String, dynamic> json) => SubTask(
    id: json['id'] ?? '',
    entryId: json['entry_id'] ?? '',
    text: json['text'] ?? '',
    done: json['done'] ?? false,
    sortOrder: json['sort_order'] ?? 0,
  );

  Map<String, dynamic> toJson() => {
    'id': id,
    'entry_id': entryId,
    'text': text,
    'done': done,
    'sort_order': sortOrder,
  };
}
```

Add to `HistoryEntry`:
```dart
final List<SubTask> subtasks;
```

---

## API Changes (brain.exe)

### Option A: Nested under entries (recommended)

Sub-tasks are always accessed through their parent entry.

```
GET    /api/entries/{id}                    → includes "subtasks" array in response
POST   /api/entries/{id}/subtasks           → create sub-task (body: {text, sort_order?})
PUT    /api/entries/{id}/subtasks/{sid}     → update sub-task (body: {text?, done?, sort_order?})
DELETE /api/entries/{id}/subtasks/{sid}     → delete sub-task
POST   /api/entries/{id}/subtasks/reorder   → bulk reorder (body: {ids: [...]})
```

`GetEntry` and `ListAll` include sub-tasks in the response. This avoids N+1 on the list view — a single LEFT JOIN or a follow-up query batched per page.

### Relay (ibeco.me)

Sub-task CRUD piggybacked on existing entry relay messages. Two approaches:

1. **Included in entry_updated** — when brain.exe processes a sub-task change, it sends the full entry (with subtasks) as an `entry_updated` message. ibeco.me cache already stores full entries.
2. **Dedicated sub-task messages** — overkill for now. Start with approach 1.

For relay-mode creates: `updateEntry()` already sends partial updates. Sub-task operations can use the same pattern, or we add `createSubTask()` / `updateSubTask()` to BrainApi that proxy through ibeco.me.

---

## Flutter UI

### Edit Entry Screen

Below the body field, add a "Sub-tasks" section:

```
┌──────────────────────────────────┐
│ Title: [Grocery list           ] │
│ Body:  [Weekly groceries       ] │
│                                  │
│ Sub-tasks                        │
│ ☑ Milk (2)                    ✕  │
│ ☑ Eggs                        ✕  │
│ ☐ Bread                       ✕  │
│ ☐ Butter                      ✕  │
│ [+ Add item                    ] │
│                                  │
│ Category: [actions ▼]            │
└──────────────────────────────────┘
```

- Checkbox toggles `done` inline (optimistic update, PUT to API)
- Tap text to edit inline
- ✕ button to delete (with undo toast, not confirmation dialog)
- Drag handle for reorder (long-press to enter reorder mode)
- "Add item" text field at the bottom — Enter key adds and clears
- Progress indicator: "3/5 done" shown as a subtitle or progress bar

### Entry List View

Show sub-task progress on the entry card when sub-tasks exist:

```
┌─────────────────────────────┐
│ 🛒 Grocery list             │
│ actions • 3/5 done ████░░   │
└─────────────────────────────┘
```

### AI Classification

Update the classifier prompt to handle sub-task-shaped input:

> If the input contains a list of items (bullets, numbers, or comma-separated), extract them as `subtasks` in the response.

Add to JSON schema:
```json
"subtasks": ["string array or omit"]
```

The classifier returns sub-task *text* only. brain.exe creates SubTask records from these strings after classification.

---

## Implementation Phases

### Phase 1: Data & API
1. Add `subtasks` table migration in `store/db.go`
2. Add `SubTask` type and `SubTasks` field on Entry
3. DB methods: `InsertSubTask`, `UpdateSubTask`, `DeleteSubTask`, `ReorderSubTasks`, `ListSubTasks`
4. Load subtasks in `GetEntry` and `ListAll` (batched query)
5. REST endpoints: POST/PUT/DELETE under `/api/entries/{id}/subtasks`
6. Update `handleBrainHistory` to include subtasks

### Phase 2: Flutter UI
1. Add `SubTask` model class and update `HistoryEntry`
2. SubTaskList widget (checkboxes, inline edit, delete, add)
3. Integrate into EditEntryScreen
4. Entry card progress indicator
5. API methods: `createSubTask()`, `updateSubTask()`, `deleteSubTask()`

### Phase 3: Smart Classification
1. Update classifier prompt to extract sub-tasks from list-shaped input
2. brain.exe: after classification, create SubTask records from `result.SubTasks`
3. Test with natural inputs: "grocery list: milk, eggs, bread, butter"

### Phase 4: Relay
1. ibeco.me: ensure entry_updated messages include subtasks
2. Relay mode sub-task CRUD (proxy through ibeco.me)
3. Offline queue support for sub-task operations

---

## Open Questions

- **Nesting?** Sub-sub-tasks? Almost certainly no. Keep it flat.
- **Quantity field?** "Milk (2)" — parse from text or add a dedicated `quantity` column? Start with parsing from text, add structured field if needed later.
- **Completion behavior?** When all sub-tasks are done, auto-mark the entry done? Probably yes for actions, no for other categories.
