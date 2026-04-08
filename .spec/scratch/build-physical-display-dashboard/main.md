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

---

## Plan

**Scope:** 4-5 sessions
**Complexity:** medium

### What to Build

Build an LCARS-style dashboard for the Waveshare ESP32-S3 4.3" touch display with:
- HTML/CSS prototype first (for design validation)
- C/LVGL firmware for ESP32-S3 hardware
- Weather integration (National Weather Service API)  
- ibeco.me integration (todos, scripture memorization)
- Docker development environment
- Auto-cycling screens with pause functionality

**Files/Packages:**
- `projects/space-center/dashboard/` - Main project directory
- `projects/space-center/dashboard/prototype/` - HTML/CSS demo
- `projects/space-center/dashboard/firmware/` - ESP32 firmware
- `projects/space-center/dashboard/docs/` - Hardware specs, API docs

### Phases

1. **Phase 1: HTML/CSS Prototype** (1 session)
   - Deliverable: Working HTML/CSS demo with LCARS styling
   - Files: `prototype/index.html`, `prototype/style.css`, `prototype/script.js`
   - Test touch interactions, screen cycling, LCARS aesthetic

2. **Phase 2: Development Environment** (1 session)
   - Deliverable: Docker container with ESP-IDF + LVGL toolchain
   - Files: `Dockerfile`, `docker-compose.yml`, development docs
   - Verify compilation and flashing workflow

3. **Phase 3: Basic LVGL Display** (1 session)  
   - Deliverable: LVGL "Hello World" running on hardware
   - Files: `firmware/main.c`, LVGL configuration, display drivers
   - Establish touch input and basic UI framework

4. **Phase 4: API Integration** (1-2 sessions)
   - Deliverable: WiFi connectivity + data fetching from APIs
   - Files: WiFi setup, HTTP client code, JSON parsing
   - Weather from NWS API, todos/scripture from ibeco.me API

### Scenarios

**HTML Prototype Validation:**
- When user loads prototype in browser, then LCARS-styled interface displays
- When user touches screen areas (simulated), then visual feedback appears
- When auto-cycle timer expires, then screen transitions to next view

**Hardware Display:**
- When device boots, then LVGL UI displays on 4.3" screen  
- When user touches display, then touch coordinates are detected and processed
- When WiFi connects, then weather data fetches from NWS API
- When device cycles screens, then transitions between weather/todos/scripture views

**API Integration:**
- When device requests weather, then current conditions display for configured location
- When device fetches todos, then active tasks from selected ibeco.me groups appear
- When device shows scripture, then memorization verses display with touch interactions
- When user touches todo checkbox, then API call marks task complete

**Dashboard Operation:**
- When device starts auto-cycle, then screens change every 30 seconds
- When user touches pause button, then auto-cycle stops until un-paused
- When device refreshes data, then weather updates hourly, todos/scripture on each cycle

### Decisions Needed

1. **Hardware Configuration:** Touch vs non-touch version?
   - Touch version: Interactive checkboxes, pause controls, manual navigation
   - Non-touch: Display-only, simpler code, lower cost
   - Trade-off: Interactivity vs complexity

2. **Development Framework:** Arduino IDE vs ESP-IDF?
   - Arduino: Easier learning curve, more examples available online
   - ESP-IDF: More professional, better performance control, official tooling
   - Trade-off: Development speed vs capabilities

3. **LCARS Fidelity:** How faithful to original design?
   - High fidelity: Curves, specific colors, authentic fonts (complex CSS/LVGL)
   - Simplified: Rounded rectangles, color scheme, general aesthetic (easier)
   - Trade-off: Visual appeal vs development time

4. **Data Storage:** Local caching vs live fetching?
   - Live: Always current, simpler code, requires constant WiFi
   - Cached: Works offline, more complex sync, handles network failures  
   - Trade-off: Reliability vs simplicity

### Risks

**Hardware Complexity:** ESP32-S3 uses CH422G I/O expander for display control
- Mitigation: Follow Waveshare examples exactly, use their provided libraries

**LVGL Learning Curve:** LVGL has different paradigms than web development
- Mitigation: Start with simple examples, build complexity gradually

**API Rate Limits:** NWS and ibeco.me may have usage restrictions
- Mitigation: Cache responses, implement reasonable request intervals

**Touch Calibration:** Capacitive touch may need device-specific calibration
- Mitigation: Use proven touch libraries, implement calibration routine

**Power Management:** Always-on display may drain battery quickly
- Mitigation: Implement display dimming, consider sleep cycles for battery operation

### Dependencies

**Hardware:**
- Waveshare ESP32-S3-Touch-LCD-4.3 development board (confirmed available)
- USB-C cable for programming and power
- WiFi network access for API connectivity

**Software:**
- Docker for development environment consistency
- ESP-IDF v5.3+ (per Waveshare documentation)
- LVGL v8.4.0 library (required for ESP32-S3-Touch-LCD-4.3)
- ESP32_Display_Panel library (Waveshare specific)

**External Services:**
- National Weather Service API (free, no key required)
- ibeco.me API (requires API key from becoming-mcp configuration)
- Local WiFi network with internet access

**Workspace Integration:**
- ibeco.me API documentation in `./scripts/becoming`
- Space Center project structure in `projects/space-center`
- Existing knowledge pipeline infrastructure

### Who Benefits? (Consecration Check)

**Primary Users:**
- Michael (Space Center developer) - Visual progress tracker, design validation
- Space Center visitors - Future bridge console UI patterns and aesthetics  
- Family/household - Weather and task visibility in common areas

**Broader Stewardship:**
- Space Center project - Proof of concept for physical control interfaces
- Dashboard concept - Reusable patterns for other IoT displays
- LCARS design exploration - Authentic sci-fi aesthetic for educational/entertainment use

This serves the larger vision of the Space Center as an educational destination that makes space science engaging and accessible.

### How Does This Integrate? (Zion Check)

**Extends Existing Work:**
- Builds on Space Center project architecture and documentation
- Integrates with existing ibeco.me API and data structures
- Uses established development patterns (Docker containers, Git workflows)

**Creates New Capabilities:**
- Physical display interfaces (new domain for scripture-study workspace)
- Hardware-software integration patterns
- LVGL/embedded GUI development expertise

**Complements Without Competing:**
- Different interface for existing ibeco.me data (web → physical display)
- Proof of concept for Space Center simulator consoles (research vs production)
- Expands project scope without duplicating existing functionality

**Foundation for Future Work:**
- Hardware patterns for other embedded displays
- LCARS design system for Space Center simulator
- IoT integration approaches for physical spaces

The physical dashboard connects digital task tracking to ambient awareness — making spiritual commitments visible in daily life.
