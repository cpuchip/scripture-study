# Becoming App — UX Improvements Spec

*Created: February 23, 2026*
*Agent: UX*

---

## Executive Summary

The Becoming app has strong bones — practices, memorization, study mode, reflections — but the daily mobile experience has friction that undermines the "super easy to use" goal. This spec identifies five categories of improvements: **critical flow fixes** (scroll reset, buried practices), **mobile-first redesign** (widget-based Today page), **memorization decay handling**, **navigation simplification**, and **consistency polish**.

---

## Current State Audit

### What You Have (19 practices)

| Category | Type | Count | Notes |
|----------|------|-------|-------|
| Scripture | memorize | 9 | D&C 93:29, Moroni 10:32-33, Alma 7:11-13, etc. |
| Study | memorize | 1 | "Comprehension is Ontological" |
| Morning | scheduled | 2 | Shave (every 2d), Clean pants (every 2d) |
| PT | tracker | 8 | Plank, chin tuck, goblin squats, bird dog, etc. |

### What's Working
- Practice types (habit, tracker, memorize, scheduled) cover real use cases
- Study Mode's adaptive difficulty ladder is genuinely sophisticated
- Undo on log actions (hover-to-undo pattern) avoids "are you sure?" friction
- Date navigation on Today page
- Category/pillar grouping and filtering

### What's Not Working

#### 1. Scroll Reset on Action (Critical — Mobile Killer)
**The problem:** Tapping a log button on the Today page calls `quickLog()` → `load()` → full re-render. The page scrolls to the top. On mobile, if your PT exercises are at the bottom, every single set-tap scrolls you back up and you have to find your place again.

**Impact:** This is the #1 reason the app feels frustrating on mobile. You're tapping 3 sets × 8 PT exercises = 24 taps, each one losing your place.

#### 2. Practices Buried in a Flat List
**The problem:** The Today page is one long list grouped by category. PT exercises (8 items × 3 sets each), memorize cards (9 items), scheduled tasks, and habits are all mixed into one scrollable page. On mobile, the PT section alone is longer than the viewport.

**Impact:** Can't see at a glance what's done vs. what needs attention. The categories help, but there's no way to collapse sections you've finished or focus on one type at a time.

#### 3. Memorize → Study Flow Friction
**The problem:** From Today, you tap a memorize card → Memorize page opens with that card selected → you do Review/Practice mode → you want to study → you click "Study" link → Study page loads, starts a *new session* with no memory of what you just reviewed. The flow is linear but the context is lost.

**More critically:** Going from Memorize to Study loses your filter state unless you construct the URL manually. The `studyLink` computed property builds the URL with query params, but it's buried — most users just click the nav link.

#### 4. Study Mode Doesn't Account for Time Away (Decay Problem)
**The problem:** SM-2 calculates `next_review` based on `time.Now()` when the review happens. If you miss 10 days of practice, the algorithm doesn't know. When you come back:
- Cards scheduled for day 3 are still showing "due" at day 13, but the algorithm treats your review as if you've been practicing continuously
- The ease factor and interval advance normally, even though the user has likely forgotten more than someone who's been reviewing daily
- There's no "rust" penalty for time away

**Example:** Moses 8:27 has `interval: 6, repetitions: 2, next_review: 2026-02-23`. If you skip 2 weeks and review on March 9, SM-2 will say "correct after hesitation? interval = 6 × 2.46 = 15 days." But after 2 weeks away, the user probably needs interval reset to something much shorter.

#### 5. Navigation Overload
**The problem:** The nav bar has 10 items on desktop, 10 items in the mobile hamburger menu: Today, Memorize, Practices, Reports, Pillars, Notes, Reflect, Library, Bookmarks, Tasks. For a daily-use mobile app, this is too many choices. Most sessions should be: open app → see Today → tap through a few things → done.

**The symptom:** The hamburger menu is a flat list with no visual hierarchy. The primary daily actions (Today, Memorize, Study) are at the same visual level as rarely-used pages (Pillars, Notes, Reports).

#### 6. Filtering Inconsistency
**The problem:** Three different filtering paradigms across pages:
- **Today:** Single-select filter chips (category OR pillar, not both). Toggle between group modes.
- **Memorize:** Multi-select Set-based filters for both categories AND pillars simultaneously. Tri-state pattern mentioned in UX phases doc but implemented differently here.
- **Study:** Gets filters from URL query params passed from Memorize. No visible filter UI of its own.

**The fix isn't necessarily making them identical** — each page has different needs — but the interaction pattern should feel consistent.

---

## Improvement Plan

### Phase 1: Critical Flow Fixes (Quick Wins — 1-2 days each)

#### 1.1 Fix Scroll Reset on Log Actions

**User Goal:** Tap a set button without losing my place.

**Root Cause:** `quickLog()` calls `await load()` which re-fetches ALL data and replaces `summary.value`, triggering a full re-render.

**Solution:** Optimistic UI update.

**Happy Path:**
1. User taps set button for "Plank — Set 2"
2. Immediately: button shows ✓, set count increments, completion stats update
3. In background: API call fires
4. If API fails: revert the optimistic update, show inline error toast

**States:**
- **Default:** Button shows set number, uncolored
- **Optimistic update:** Button shows ✓ green, counter incremented
- **Error rollback:** Button reverts, brief red flash, toast: "Couldn't save. Tap to retry."

**Implementation approach:**
- In `quickLog()`, mutate `summary.value` in-place (increment `log_count`, `total_sets`) before the API call
- Save scroll position with `window.scrollY` before mutation, restore after `nextTick`
- On API success: optionally re-fetch silently to sync any server-side state
- On API failure: revert the in-place mutation, show error

**Component changes:** DailyView.vue only — no new components.

#### 1.2 Collapsible Category Sections

**User Goal:** Hide completed sections so I can focus on what's left.

**Happy Path:**
1. User sees grouped sections (PT, Morning, Scripture, etc.)
2. Each group header is tappable — shows category name + completion count (e.g., "PT — 3/8")
3. Tap a completed group → it collapses to a single green bar: "PT ✓ 8/8"
4. Tap again → it expands

**States:**
- **Expanded (default):** Full practice list
- **Collapsed:** Single-line summary bar with completion count
- **Auto-collapse option:** Completed groups auto-collapse on next visit (stored in localStorage per date)

**Persistence:** Store collapsed state in localStorage keyed by date-string. On a new day, all sections start expanded.

**Component inventory:**
| Component | Type | Notes |
|-----------|------|-------|
| CollapsibleGroup | New | Wraps each category group, header + toggle + content slot |
| DailyView | Modified | Use CollapsibleGroup for each grouped section |

#### 1.3 Memorize/Study Context Handoff

**User Goal:** Flow from Today → Memorize → Study without losing context.

**Current flow:**
```
Today → Memorize (card selected via ?id=) → Study (loses context)
```

**Proposed flow:**
```
Today → Memorize (card selected via ?id=, filters preserved)
  └─ "Study All" button → Study (?category=...&pillar_ids=... preserved from filters)
  └─ "Study This Card" → Study (?card_id=15, starts with that card)
```

**Changes:**
- Memorize's "Study" link already computes filter params — make it more prominent (it's currently a small text link)
- Add a "Study this card" action per-card in Memorize view (bottom of card when flipped/after rating)
- Study mode: accept `?card_id=N` query param to start session focused on one card

---

### Phase 2: Mobile-First Today Redesign (Widget Approach — 1 week)

#### 2.1 Widget-Based Today Layout

**User Goal:** See my day at a glance. Act on things without scrolling. Know what's done and what's left.

**Design Philosophy:** The Today page should feel like a phone home screen — widgets at a glance, tap to drill in. Not a long scrollable list.

**Proposed Layout (mobile, top to bottom):**

```
┌─────────────────────────────────┐
│  Mon, February 23               │
│  8/19 completed                 │
│  [← prev]  [Jump to Today]  [next →] │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ 📖 Memorize          5/9 done  │
│ ┌───────────────────────────┐   │
│ │ Moses 8:27    ●●●○○  [→] │   │
│ │ D&C 93:29     ○○○○○  [→] │   │
│ └───────────────────────────┘   │
│ 6 cards due · Study All →       │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ 💪 PT Exercises       0/8 done │
│ ┌───────────────────────────┐   │
│ │ plank         [1][2]      │   │
│ │ chin tuck     [1][2][3]   │   │
│ │ goblin squats [1][2]      │   │
│ │ ...3 more                 │   │
│ └───────────────────────────┘   │
│ Expand all ↓                    │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ 🌅 Morning             1/2     │
│ ○ Shave · due in 2 days        │
│ ✓ Clean pants · next in 2 days │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ 💭 Daily Reflection    [▼]     │
│ "What's on your heart today?"  │
└─────────────────────────────────┘
```

**Key differences from current:**
1. **Each category is a card/widget** — self-contained, visually bounded
2. **Completion progress is per-widget** — "PT 0/8" tells you at a glance what needs work
3. **Long sections collapse** — PT shows first 4 items + "...3 more" + "Expand all"
4. **Memorize widget shows dot progress** (●○ filled/unfilled for daily reps target) instead of fraction text
5. **Tapping a set button does NOT re-render the whole page** (Phase 1.1 prerequisite)

**States:**
- **Default:** All widgets expanded, incomplete sections at top
- **Widget collapsed:** Tap header to collapse/expand, just shows title + progress
- **All done widget:** Category fully complete → compact green bar
- **Empty (new user):** Single widget: "No practices yet. Add your first →"

**Interaction Pattern:** Widgets, not modals or page transitions. Everything happens inline. The page stays stable.

**Sort Order:** Incomplete categories first. Within a category, incomplete items first, completed items dimmed at bottom.

**Accessibility:**
- Each widget is a `<section>` with `aria-label` (e.g., "PT Exercises, 0 of 8 complete")
- Collapse/expand buttons have `aria-expanded`
- Set buttons have `aria-label` (e.g., "Log set 2 of plank")
- Focus remains on the tapped button after optimistic update

**Component inventory:**
| Component | Type | Notes |
|-----------|------|-------|
| PracticeWidget | New | Card-style container for a category group |
| TrackerRow | New | Single tracker practice with inline set buttons |
| MemorizeWidget | New | Compact memorize section with dot progress |
| ScheduledRow | New | Single scheduled item with due state |
| DailyView | Modified | Refactored to use widget components |

#### 2.2 Quick-Action Bottom Bar (Mobile)

**User Goal:** Access the most important actions without opening the hamburger menu.

**Proposed mobile bottom nav (replaces hamburger for primary actions):**

```
┌──────┬──────┬──────┬──────┐
│ Today│ Study│ Memo │ More │
│  🏠  │  📖  │  🧠  │  ···  │
└──────┴──────┴──────┴──────┘
```

- **Today:** /today (home base)
- **Study:** /study (enter study session directly)
- **Memo:** /memorize (card review)
- **More:** Opens slide-up sheet with remaining pages (Practices, Reports, Pillars, Notes, Reflect, Library, Bookmarks, Tasks, Settings)

**Why:** On mobile, thumb-friendly bottom nav is the standard pattern (iOS tab bar, Android bottom nav). The hamburger menu requires reaching the top of the screen and then selecting from a flat list of 10 items. Three primary actions + overflow covers 95% of daily use.

**Desktop:** Keep the current top nav — it works fine with horizontal space.

**States:**
- **Active tab:** Highlighted icon + label (orange)
- **Badge on Study/Memo:** Show count of due cards (e.g., red dot with "6")
- **More sheet:** Slide-up panel with remaining nav items, grouped by frequency

**Accessibility:**
- `role="navigation"` with `aria-label="Primary navigation"`
- Active item: `aria-current="page"`
- Badge: `aria-label="6 cards due for review"`

---

### Phase 3: Memory Decay & Time-Away Handling (1 week)

#### 3.1 Decay Factor for SM-2

**The Problem:** SM-2 doesn't penalize for missed reviews. A card with `interval: 30, next_review: Jan 1` reviewed on Feb 1 (31 days late) gets treated the same as if reviewed on time.

**Proposed Solution: Overdue Ratio Decay**

When a review happens, calculate how overdue the card is and apply a decay penalty to the starting state.

```
overdue_days = review_date - next_review_date
overdue_ratio = overdue_days / interval

if overdue_ratio > 0.5:
    # Significantly overdue — penalize interval
    decay_factor = max(0.25, 1.0 - (overdue_ratio * 0.3))
    cfg.Interval = max(1, round(cfg.Interval * decay_factor))
    
    # Also reduce ease factor slightly (memory has degraded)
    cfg.EaseFactor = max(1.3, cfg.EaseFactor - 0.05 * min(overdue_ratio, 3.0))
```

**Example scenarios:**

| Card | Interval | Due | Reviewed | Overdue Ratio | Decay Factor | New Starting Interval |
|------|----------|-----|----------|---------------|--------------|----------------------|
| Moses 8:27 | 6d | Feb 23 | Feb 24 | 0.17 | No decay | 6d (normal SM-2) |
| Moses 8:27 | 6d | Feb 23 | Mar 2 | 1.17 | 0.65 | 4d → then SM-2 |
| D&C 93:29 | 1d | Feb 14 | Mar 1 | 15.0 | 0.25 (floor) | 1d (full reset) |
| John 17:3 | 218d | Sep 25 | Oct 30 | 0.16 | No decay | 218d (normal) |

**Key behavior:**
- Cards that are slightly late (< 50% of interval): no penalty, reviewed as normal
- Cards that are moderately late (50-200%): interval shrinks proportionally
- Cards that are severely late (>200%): near-full reset (interval → 1, ease → 1.3)

**UI indicator:**
- On the Memorize card, show a "rusty" badge if `overdue_ratio > 0.5`: "⚠ Haven't reviewed in 14 days — might be rusty"
- In Study mode, the card type label becomes "Rust removal" instead of "Confidence boost"

**Component changes:**
| Component | Type | Notes |
|-----------|------|-------|
| memorize.go | Modified | Add decay calculation in `SM2Review()` |
| MemorizeView.vue | Modified | Show "rusty" indicator on overdue cards |
| StudyView.vue | Modified | Label decayed cards differently |

#### 3.2 Welcome Back Screen

**User Goal:** When I haven't used the app in several days, I want to feel welcomed back, not overwhelmed by a wall of overdue items.

**Trigger:** If last activity is >3 days ago.

**Flow:**
```
[User opens app after 5 days away]
  → Welcome Back card at top of Today page
  │  "Welcome back! You've been away 5 days."
  │  "6 memorize cards may need a refresher."
  │  "Your PT exercises are still waiting."
  │  
  │  [Start Easy — Review 3 cards]  [Dive Right In]
  │
  └─ "Start Easy" → Study mode with overdue cards, 
     starting at lower difficulty levels
```

**States:**
- **Default:** Welcome back card appears at top, pushes normal content down
- **Dismissed:** User clicks "Dive Right In" → card disappears, full Today page shown
- **Easy mode:** User clicks "Start Easy" → routes to Study with `?mode=recovery`

**Why this matters:** After time away, the worst UX is seeing 9 overdue memorize cards, 8 PT exercises showing 0/3, and feeling like you've failed. The welcome-back card normalizes the gap and offers an easy re-entry.

---

### Phase 4: Navigation Simplification (2-3 days)

#### 4.1 Information Architecture Restructure

**Current pages (14 authenticated routes):** Today, Practices, Memorize, Study, Tasks, Notes, Reflections, Pillars, Reports, Sources, Bookmarks, Reader, History, Settings

**Proposed grouping by usage frequency:**

**Daily (bottom nav):**
- Today (home base, widget dashboard)
- Memorize / Study (combined entry point → fan out)

**Weekly (accessible from Today or "More"):**
- Practices (manage, create, edit)
- Reports (weekly/monthly view)
- Reflections (daily writing)

**Occasional ("More" menu):**
- Library + Bookmarks (combine into Library with bookmarks tab)
- Notes + Tasks (combine into Notes with a tasks tab, or keep separate behind "More")
- Pillars (setup, rarely changed after onboarding)
- Settings

**Key changes:**
1. **Merge Library + Bookmarks** — they're both about saved reading material. Library page gets a "Bookmarked" filter/tab.
2. **Tasks and Notes live behind "More"** — they're useful but not daily.
3. **Pillars is a setup page** — after onboarding, users rarely revisit. Move to Settings or behind "More."

#### 4.2 Desktop Nav Grouping

Instead of 10 flat links, group them:

```
Today   Study ▾   Track ▾   [settings icon]
         Memorize   Practices
         Study      Reports
                    Reflections
```

The dropdown groups reduce cognitive load from 10 items to 3 primary targets + dropdowns.

---

### Phase 5: Consistency & Polish (Ongoing)

#### 5.1 Unified Filter Pattern

Define one filter pattern used everywhere:

**Pattern: Filter Bar**
- Horizontal row of pills/chips
- Tap to include (filled)
- Tap again to exclude (filled + strikethrough) — if tri-state is needed
- Third tap to clear
- "Clear all" button when any filter is active
- Multi-select always (single-select is just multi-select with one selected)

**Apply consistently to:**
- Today page (category/pillar grouping)
- Memorize page (category/pillar filtering)
- Study page (show current filters, allow changing mid-session)
- Practices page (if filtered by status)

#### 5.2 Dark Mode Fix: Stop Using `!important` Overrides

**Current approach:** 200+ lines of `html.dark-mode .bg-white { ... !important }` overrides in App.vue.

**Better approach:** Tailwind v4 with CSS custom properties:
- Define light/dark tokens in `@theme` (e.g., `--color-surface`, `--color-surface-raised`, `--color-text-primary`)
- Components use `bg-(--color-surface)` instead of `bg-white`
- Dark mode swaps tokens, not individual class overrides
- Eliminates the `!important` cascade and makes dark mode automatic for new components

**This is a longer-term refactor** but should happen before building more pages.

#### 5.3 Loading States

**Current:** `<div v-if="loading">Loading...</div>` — plain text, no skeleton, layout shifts when content appears.

**Proposed:** Skeleton screens that match the widget layout:
- Today widgets: gray pulsing rectangles matching the card shapes
- Memorize card: card-shaped skeleton with placeholder text lines
- Practice list: row-shaped skeleton strips

**Quick win:** Even just adding `min-height` to the loading container to prevent layout shift would help.

#### 5.4 Toast Notification System

**Now:** No notification system. Success/error feedback is either missing or implicit (the UI re-renders with new data).

**Proposed:** A simple toast composable:

```typescript
// useToast.ts
const toast = useToast()
toast.success('Practice logged')
toast.error('Couldn't save. Tap to retry.', { action: () => retry() })
toast.undo('Deleted "Plank"', { onUndo: () => restore() })
```

**Rules (from our UX principles):**
- Max 2 toasts visible at once
- Auto-dismiss success toasts after 3 seconds
- Error toasts persist until dismissed
- Undo toasts persist for 5 seconds
- Position: bottom-center on mobile, bottom-right on desktop

---

## Priority Matrix

| Improvement | Impact | Effort | Priority |
|-------------|--------|--------|----------|
| 1.1 Fix scroll reset | 🔴 Critical | Small (1 day) | **P0 — Do first** |
| 1.2 Collapsible sections | 🟡 High | Small (1 day) | **P1** |
| 1.3 Memorize/Study handoff | 🟡 High | Small (half day) | **P1** |
| 2.1 Widget-based Today | 🟢 Transformative | Medium (3-4 days) | **P1** |
| 2.2 Mobile bottom nav | 🟡 High | Medium (2 days) | **P2** |
| 3.1 SM-2 decay factor | 🟡 High | Medium (2 days) | **P2** |
| 3.2 Welcome back screen | 🟢 Nice | Small (half day) | **P3** |
| 4.1 Nav restructure | 🟡 High | Medium (2 days) | **P2** |
| 5.1 Unified filters | 🟡 Medium | Medium (2 days) | **P3** |
| 5.2 Dark mode tokens | 🟡 Medium | Large (3-4 days) | **P3** |
| 5.3 Skeleton loading | 🟡 Medium | Small (1 day) | **P3** |
| 5.4 Toast system | 🟡 Medium | Small (1 day) | **P2** |

---

## Implementation Order

### Sprint 1: Stop the Bleeding (2-3 days)
1. **Fix scroll reset** — optimistic updates in `quickLog()`
2. **Collapsible category sections** — CollapsibleGroup component
3. **Memorize → Study context preservation** — query param handoff

### Sprint 2: Mobile-First Today (1 week)
4. **Widget-based Today layout** — rearchitect DailyView into widget components
5. **Mobile bottom nav** — replace hamburger with thumb-friendly tab bar
6. **Toast notification system** — useToast composable

### Sprint 3: Memory Intelligence (1 week)
7. **SM-2 decay factor** — overdue ratio penalty in memorize.go
8. **Welcome back screen** — detect time away, offer gentle re-entry
9. **Rust indicators** — visual cues for forgotten/decayed cards

### Sprint 4: Navigation & Polish (1 week)
10. **Nav restructure** — group by frequency, combine Library+Bookmarks
11. **Unified filter component** — extract shared filter bar
12. **Skeleton loading states** — layout-preserving loading

### Future
13. **Dark mode token refactor** — CSS custom properties instead of `!important` overrides
14. **PWA offline support** — service worker for practice logging while offline
15. **Haptic feedback** — navigator.vibrate() on mobile for set completion

---

## Open Questions

1. **Widget reordering?** Should users be able to drag-and-sort the Today widgets to put PT at the top? Or should the app auto-sort (incomplete first)?
2. **Memorize + Study merge?** These two pages do related things. Would one page with tabs (Review | Practice | Study) be clearer than two separate pages?
3. **PT timer?** Several exercises have timed holds (plank 30s, chin tuck 30s). Should the app include a countdown timer, or is that over-engineering?
4. **Streak tracking?** Showing a streak counter ("7 days in a row!") could motivate consistency. But it could also guilt after a gap. Worth testing?
5. **Library + Bookmarks merge** — does this make sense, or do you think of these as genuinely separate?

---

## Accessibility Audit Summary

### Good
- Semantic HTML (buttons are buttons, links are links)
- Keyboard-navigable practice rows
- Dark mode support

### Needs Work
- **No `aria-label` on icon-only buttons** — set completion buttons (✓/✕) have no accessible name
- **No focus management after log actions** — focus is lost when the page re-renders
- **No skip-to-content link** — standard a11y requirement
- **Color-only status indicators** — completed items are dimmed via opacity only, no icon/text change
- **Mobile hamburger menu** — no `aria-expanded`, no focus trap when open
- **Touch targets** — some filter chips and small buttons may be < 44×44px

### Quick Accessibility Wins
1. Add `aria-label` to all icon-only buttons in DailyView
2. Add `aria-live="polite"` region for completion stats (screen reader announces progress)
3. Add skip-to-content link in App.vue
4. Ensure all touch targets are >= 44×44px (especially set buttons and filter chips)
