# Intent — Chore Tracker

*Why this project exists. The assistant optimizes for this when instructions run out.*

## Purpose

A simple daily chore tracker for two kids (ages 9 and 14) that runs on the home
network. Kids check off their own chores; parents manage the list and can fix mistakes.

## What done looks like

- [ ] Kids can visit the app on the home network, pick their name, and see their chores
- [ ] Daily chores show as complete or incomplete for the current day; reset each morning
- [ ] Weekly chores reset on a fixed day of the week
- [ ] Star points accumulate per completed chore — visible on the kid's screen
- [ ] Parent view: add, edit, remove chores; assign to one or both kids; fix mistakes
- [ ] Go HTTP server + plain HTML/JS frontend — no framework, no build step
- [ ] No logins — kid selection is a simple name pick on the home screen

## Values (when goals collide, prefer…)

- Simple over clever — home network app for kids, not a SaaS product
- Visible over hidden — kids should always see their progress and star totals clearly
- Parent control over kid autonomy — parent can always override any check
- Working over polished — get the core loop right before adding anything extra

## Non-goals

- No money or allowance tracking (may be phase 2)
- No streaks (phase 2 goal — acknowledged, deferred)
- No user accounts or passwords
- No external hosting or mobile app packaging
- No frontend framework or build toolchain
