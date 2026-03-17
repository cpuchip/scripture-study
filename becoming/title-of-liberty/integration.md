# Title of Liberty — Becoming App Integration

How the Title of Liberty rhythm maps to existing Becoming app features. Almost everything we need already exists.

---

## What Already Exists

The Becoming app (ibeco.me) has the building blocks. Title of Liberty doesn't need a badge engine, a rank tracker, or a troop management system. It needs habits, a place for quests, and a way to reflect.

| App Feature | Status | Title of Liberty Use |
|-------------|--------|---------------------|
| **Practices** (habit type) | ✅ Built | The 5 daily habits — Scripture, Prayer, Service, Hard Thing, Reflection |
| **Practice Logs** | ✅ Built | Daily habit completion tracking, streaks |
| **Tasks** | ✅ Built | Self-directed quests (title, description, status, reflection on completion) |
| **Memorization** (SM-2 spaced repetition) | ✅ Built | Scripture memorization — already works perfectly |
| **Pillars** (hierarchical, many-to-many) | ✅ Built | Four pillars: Faith & Covenant, Liberty & Service, Knowledge & Skill, Strength & Discipline |
| **Notes** | ✅ Built | Family council notes, quest documentation |
| **Reflections** | ✅ Built | Daily reflection habit, quarterly reviews, milestone recognition |
| **Reports** (heatmaps, practice reports) | ✅ Built | Habit consistency, streaks, growth patterns over time |

---

## Setup — What to Configure

No new code needed for Phase 1. Just configure existing features:

### 1. Pillars

Create (or rename) the 4 pillars. These parallel the default pillars with Book of Mormon framing:

| Title of Liberty Pillar | Default Pillar It Parallels |
|------------------------|-----------------------------|
| ⚔️ Faith & Covenant | Spiritual 🙏 |
| 🛡️ Liberty & Service | Social 🤝 |
| 🏗️ Knowledge & Skill | Intellectual 📚 |
| 💪 Strength & Discipline | Physical 💪 |

### 2. Daily Habits (Practices)

Create 5 practices, all as `habit` type with `daily` frequency:

| Practice | Pillar | Notes |
|----------|--------|-------|
| Scripture | ⚔️ Faith & Covenant | Personal reading — log with a note about what stood out |
| Prayer | ⚔️ Faith & Covenant | Personal prayer (not just family prayer) |
| Service | 🛡️ Liberty & Service | One act of kindness or service |
| Hard Thing | 💪 Strength & Discipline | Something difficult — physical, intellectual, or social |
| Reflection | All pillars | Brief note: what did I learn or how did I grow today |

### 3. Scripture Memorization

Add memorization cards as they come up naturally from reading. No pre-loaded 50-card deck. When a verse hits hard during family devotional, add it as a card.

Starting suggestions:
- [Alma 53:20–21](../../gospel-library/eng/scriptures/bofm/alma/53.md) — the standard
- [Alma 46:12](../../gospel-library/eng/scriptures/bofm/alma/46.md) — the title of liberty
- [1 Nephi 3:7](../../gospel-library/eng/scriptures/bofm/1-ne/3.md) — "I will go and do"
- [D&C 58:27](../../gospel-library/eng/scriptures/dc-testament/dc/58.md) — "anxiously engaged"

### 4. Quests (Tasks)

When someone picks a quest, create it as a task:
- **Title:** The quest name ("Memorize Alma 46:12–13", "Build a birdhouse", "Cook dinner weekly for a month")
- **Pillar:** Tag it to the relevant pillar(s)
- **Status:** Use the existing task lifecycle (not_started → in_progress → completed)
- **Notes:** Use the notes feature for check-ins, progress updates, reflections on completion

---

## What's New — Family Layer (webeco.me)

The only genuinely new features are for the family/community view. These are Phase 2 — not needed to start.

### Family Dashboard

A view on webeco.me where parents can see:
- Each family member's habit streaks (today, this week, longest streak)
- Active quests per person
- Recent reflections (opted-in — privacy matters, especially for teenagers)

This is a read-only aggregation view. Each person's data lives in their own ibeco.me account. The family dashboard just pulls it together.

### Family Journal

Shared notes from weekly family council. One entry per week:
- What we read in the Book of Mormon
- What stood out
- Growth we noticed in each other
- Quests underway

Could be a note in the app, or just a physical journal. The app version is a convenience.

### Quest Board

A shared view of everyone's active quests. Visible to the family so people can encourage each other, offer help, or be inspired by what others are working on.

---

## What We Don't Need

The v1 integration spec had:
- Badge tracking tables (badges, badge_requirements) — **removed**. Quests use existing tasks.
- Rank progress tables (rank_progress) — **removed**. Milestones are recognized in family council, recorded as reflections.
- Troop/family group management — **deferred**. Start with individual accounts. Family dashboard is the only group feature.
- Program template engine — **removed**. No need for a system that loads 59 badges into a database.
- 7 new MCP tools — **removed**. Existing tools (practices, tasks, reflections, notes) cover everything.

---

## Implementation Priority

| Priority | What | How |
|----------|------|-----|
| **Now** | Set up 5 habits as practices | Configure in app — no code |
| **Now** | Create/rename pillars | Configure in app — no code |
| **Soon** | Add first memorization cards | Configure in app — no code |
| **When quests start** | Create quests as tasks | Configure in app — no code |
| **Phase 2** | Family dashboard on webeco.me | New feature — aggregation view |
| **Phase 2** | Family journal on webeco.me | New feature — shared notes |
| **Phase 2** | Quest board on webeco.me | New feature — shared task view |

The pattern: **Phase 1 is configuration, not development.** Everything a single person or family needs to start already exists in the app. New development only becomes necessary when the family wants a shared view.

---

## Non-App Options

The app is optional. The system works without it:

| Component | App Version | Non-App Version |
|-----------|-------------|-----------------|
| Habit tracking | Practice logs with streaks | Physical board on the wall, or verbal check-in at dinner |
| Quests | Tasks with notes | Written in a personal journal |
| Memorization | Spaced repetition cards | Physical flashcards |
| Reflection | Daily reflection entries | Journal entry or dinner conversation |
| Family council | Shared note in app | Conversation around the dinner table |
| Family dashboard | webeco.me view | The parents just... know. They live in the same house. |

The system's power is in the rhythm and habits, not the technology. The app just makes tracking easier for people who like tracking.
