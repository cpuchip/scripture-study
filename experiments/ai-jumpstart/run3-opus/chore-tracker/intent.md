# Intent — Chore Tracker

*Why this project exists. The assistant optimizes for this when instructions run out.
Written together in the first session (2026-06-12); revised whenever the vision sharpens.*

## Purpose

A simple home chore tracker for Michael's two kids (ages 9 and 14). Each kid opens it
on their own device over the home network, sees their chores for the day, and checks
off their own. It remembers from one day to the next and resets each morning. The point
is a kid-usable list that makes "what do I need to do today, and did I do it" obvious —
**nothing fancy.**

## What done looks like

- [ ] A single Go binary I run on my machine; kids reach it from their own devices on
      the home network (a LAN URL, no internet, no accounts to manage).
- [ ] Each kid picks who they are and sees only their own chores.
- [ ] Daily chores reset each morning; a few weekly chores reset weekly. A chore done
      today shows "done for the day" until the day rolls over.
- [ ] A kid taps a chore to mark it done themselves (no approval step, to start).
- [ ] Completing a chore earns simple **star points**; each kid has a running star
      total. No money, no allowance math.
- [ ] A parent can fix mistakes — uncheck something, add/rename/remove a chore.
- [ ] State survives a server restart (persists to disk).

## Values (when goals collide, prefer…)

- **Light over featureful.** "Nothing fancy" is the brief. Plain `net/http` + plain
  HTML/JS, no frameworks, no build step, no dependency I don't have to add.
- **Kid-usable over slick.** The 9-year-old has to be able to use it. Big tap targets,
  obvious state, readable on a tablet beats visual polish.
- **Boring and durable over clever.** Simplest thing that survives a restart and a
  forgetful Tuesday morning.

## Non-goals (for now — named so we don't drift into them)

- No money, allowance, or payout tracking.
- No parent-approval queue (a kid's check counts immediately). *Later-phase candidate.*
- No streaks yet. *Later-phase goal.*
- No accounts/passwords/auth — it's the home network, trusted devices.
- No cloud, no deploy target, no internet exposure.
- No SPA framework (React/Vue), no CSS framework, no npm.
