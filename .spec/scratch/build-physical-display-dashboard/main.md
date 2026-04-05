# Research: Build Physical Display Dashboard

**Entry ID:** build-physical-display-dashboard
**Category:** projects
**Captured:** 2026-04-05
**Related Project:** Space Center

---

## What This Is About

Build an LCARS-based (Star Trek: The Next Generation) dashboard on a Waveshare ESP32-S3 4.3" display. Display weather, integrate with ibeco.me for todo items, and test design patterns for control surfaces on the space ship bridge simulator.

## What Already Exists

[WORKSPACE] **Space Center Pipeline Test** (`.spec/scratch/space-center-pipeline-test/main.md`)
- Full documentation of the Space Center project as a test bed for the knowledge pipeline
- Confirms Space Center is Michael's vision: planetarium, science museum, starship bridge simulator
- Lists "Starship Bridge Simulator — Game Design" as one of the seed project ideas
- No existing dashboard or UI specifications yet

[WORKSPACE] **Private brain structure** exists at `/private-brain/ibeco.me` and appears to track a date (2026-03-21) — likely a todo/task system referenced in the project guidance

## External Context

[WEB] **LCARS Design** (Wikipedia)
- Library Computer Access/Retrieval System — Star Trek: The Next Generation computer OS interface
- Designed by Michael Okuda to minimize UI clutter (conveyed advanced technology)
- Used on PADDs (Personal Access Display Devices) — hand-held computers in TNG/DS9
- Often features button/panel layouts, rounded rectangles, color-coded buttons, minimalist aesthetic
- CSS/HTML LCARS frameworks exist on GitHub (LCARS 9, Modern LCARS with glass morphism)
- Note: CBS Television Studios holds copyright — fan recreations exist but unlicensed commercial use is restricted

[WEB] **LVGL (Light and Versatile Graphics Library)**
- Free, open-source embedded UI library for MCUs and MPUs
- Minimal requirements: ~100kB RAM, ~200-300kB flash for simple UIs
- 30+ built-in widgets (buttons, sliders, lists, etc.)
- No external dependencies — platform-independent UI code
- Official repo: github.com/lvgl/lvgl with examples and docs
- Active community forum at forum.lvgl.io

[WEB] **ESP32-S3 + LVGL Projects** (GitHub topics)
- Multiple Waveshare display + ESP32 projects exist
- Example: project_aura — Air-quality station with LVGL UI, MQTT, Home Assistant integration
- Example: PrintSphere — ESP32-S3 printer companion with circular display
- Several Waveshare ESP32 panel projects with LVGL
- Crypto price monitors and energy meters are common use cases
- Projects use C/C++ with ESP-IDF or Arduino frameworks

[WEB] **LVGL on ESP32-S3**
- LVGL is widely supported on ESP32/ESP32-S3
- Many open-source examples available on GitHub
- No major barriers to porting LVGL to Waveshare 4.3" display
- Display driver integration is key implementation detail

## Open Questions

**Hardware & Firmware**
1. What display driver does the Waveshare ESP32-S3 4.3" board use? (ILI9488, ST7789, or other?)
2. Is there existing Waveshare documentation or example code for this specific board?
3. Does LVGL work "out of the box" with this display, or will custom driver code be needed?
4. What development environment is preferred: PlatformIO, Arduino IDE, or ESP-IDF directly?

**LCARS Design & UI**
5. How faithful should the dashboard be to canonical LCARS design vs. simplified version for small screen?
6. What elements of LCARS are most recognizable/iconic? (Curved corner buttons? Color palette? Typography?)
7. How will page navigation work on a 4.3" screen — touch, swipe, or button controls?

**Integration & Data Fetching**
8. For weather: Should we use OpenWeatherMap (free tier available), National Weather Service (US, free), or another API?
9. For ibeco.me integration: Does ibeco.me expose an API? If not, what's the fallback (manual sync, direct database access)?
10. For scripture memorization: Should we pull from the gospel-library corpus, or is there a separate scripture API to query?

**Feature Scope**
11. Should the dashboard cycle between screens automatically, or require manual navigation?
12. How often should it refresh weather data (every 30 min? 1 hour? less frequently to conserve bandwidth)?
13. For todos/ibeco integration: Should the device allow checkbox interaction, or just display?

**Storage & Connectivity**
14. Should todo/scripture data be stored locally (flash) and synced periodically, or fetched fresh each time?
15. Is persistent WiFi connection expected, or on-demand connection for updates?

## Raw Sources

**Workspace**
- `.spec/scratch/space-center-pipeline-test/main.md` — Space Center project overview & test plan
- `/private-brain/ibeco.me` — I Become task system (needs clarification on API access)

**LCARS & Design**
- https://en.wikipedia.org/wiki/LCARS — Comprehensive LCARS background, copyright info, design history
- https://github.com/search?q=LCARS+UI+design — GitHub projects: LCARS 9, Modern LCARS with glass morphism
- https://lvgl.io — LVGL main homepage & product info
- https://github.com/lvgl/lvgl — Official LVGL repository (C library, MIT license)

**ESP32-S3 & Display Integration**
- https://github.com/topics/esp32-s3 — 843+ repositories tagged with ESP32-S3 topics
- https://github.com/topics/lvgl-esp32 — 63+ repositories with LVGL + ESP32 projects
- Notable projects found:
  - project_aura — Air quality station with LVGL UI, MQTT, Home Assistant integration
  - PrintSphere — Round ESP32-S3 printer companion with LVGL
  - Waveshare ESP32 panel projects (P4-86) with LVGL
  - Various energy monitors, crypto tickers, display examples

**Weather APIs**
- https://openweathermap.org/api — One Call API 3.0 (free tier with API key)
- https://weather.gov/documentation/services-web-api — National Weather Service API (US, free, no key)

**REST API for ESP32**
- https://github.com/search?q=ESP32+REST+API+client — Multiple lightweight REST client libraries

**Scripture Memorization & Spaced Repetition**
- https://github.com/search?q=scripture+memorization+app — Multiple scripture learning apps
- https://github.com/jacobcapper/openSRS — Open-source spaced repetition firmware (SM-2 algorithm) for ESP32
- Existing scripture memorization apps (web/mobile): Ponderize (Latter-day Saint community tool)

**I Become (ibeco.me)**
- https://ibeco.me — Website loads as React SPA; no documentation found on API or data export format
