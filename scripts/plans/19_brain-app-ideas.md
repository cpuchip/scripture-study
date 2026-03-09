# Plan 19: Brain-App Ideas (from capture)

*Created: March 9, 2026*
*Source: Brain entries captured via voice/widget during daily use*
*Priority: Mixed — some quick wins, some far-term*

---

## Ideas

### 1. Single-Subtask Collapse
**Brain entry:** `a249be5d` — "Brain app task structure idea"
**Idea:** If an entry has only one subtask, promote it to the main body or a "next steps" note instead of showing it as a subtask list. Reduces visual clutter for simple tasks.
**Status:** Someday
**Effort:** Small
**Where:** brain-app display logic + possibly brain.exe classification

---

### 2. Natural Language Practice Creation
**Brain entry:** `ac28114d` — "Automatic budget transaction practice feature"
**Idea:** Give instructions like "add a practice to ibeco.me to do budget transactions every 4 days" and have the system create the practice automatically. NLP → API call.
**Status:** Someday
**Effort:** Medium-Large
**Connects to:** Plan 13 (Agentic Chat) — this is a concrete use-case for agentic capabilities. Could be an early agentic feature: parse intent → `POST /api/practices` with correct interval.
**Where:** brain.exe NLP pipeline or brain-app voice capture → intent extraction → ibeco.me API

---

### 3. Image Integration Pipeline
**Brain entry:** `82e58844` — "Brain App Image Integration Pipeline"
**Idea:** Photos captured through brain-app → sent to brain.exe backend → image identification/classification via VLM → stored as entries with image attachments and AI-generated descriptions.
**Status:** Roadmap
**Effort:** Large
**Connects to:** Plan 12 (Attachments) — image pipeline is the primary use-case for the attachment system. Needs file storage infrastructure (S3 or local) before implementation.
**Where:** brain-app camera/gallery capture → relay upload → brain.exe VLM classification → entry creation with attachment

---

### 4. GitHub Copilot SDK for Mobile Study Mode
**Brain entry:** `e8bdac42` — "GitHub Copilot SDK for Brain App Study Mode"
**Idea:** Integrate GitHub Copilot SDK into brain-app for on-device real-time study mode. Phone becomes a study companion — ask questions, get cross-references, explore topics while reading scriptures on the go.
**Status:** Roadmap
**Effort:** Large
**Connects to:** Plan 13 (Agentic Chat) — this is the mobile-first version of the agentic vision. Copilot SDK provides the LLM backbone; MCP tools provide the scripture/brain context.
**Where:** brain-app Flutter → Copilot SDK integration → MCP tool access (gospel-vec, brain search, webster)

---

## Dependencies

| Idea | Depends On |
|------|-----------|
| 1. Single-subtask collapse | Nothing — standalone |
| 2. NL practice creation | Plan 13 infrastructure or standalone NLP |
| 3. Image pipeline | Plan 12 (attachments + file storage) |
| 4. Copilot SDK study | Plan 13 (agentic chat framework) |

## Priority Recommendation

**Quick win:** #1 (single-subtask collapse) — small effort, improves daily UX
**Next natural step:** #2 (NL practice creation) — aligns with agentic direction
**Infrastructure-gated:** #3 and #4 wait for Plans 12/13

---

## Completed Ideas (from same capture batch)

These were captured as brain entries and have been implemented:

| Entry | What | Resolution |
|-------|------|-----------|
| Scrollable widgets (`1773032974127559297-1`) | Widgets with scrollable content | Plan 18 — both practice and brain widgets now use `ListView` + `RemoteViewsService` |
| Widget captures in recents (`4ce75e3d`) | Captures from widget don't show in capture tab | Fixed — `QuickAddScreen` writes to SharedPreferences, `HomeScreen` merges on resume |
| Widget size/mic (`c0796e7c`) | 2x4 widget too big, mic broken | Fixed — widget responsive sizing, mic wired to speech recognition |
| Nav bar compatibility (`190f9ad5`) | History screen doesn't account for Android bottom bar | Fixed — safe area padding |
| Archive feature (`0c822f0a`) | Swipe right to archive | Done — archive + filter implemented |
