---
lane: space-center
session_id: 49f54f6f-9241-498a-b0f3-58d6a0b9e884
status: active
started: 2026-06-12T16:00:00
last_active: 2026-06-13T01:50:25
---

## Working on
- ✅ Built `apps/lcars-panel` — our own native LVGL 8.x LCARS bridge-aux UI for the ESP32-P4. (`50366e1`)
- ✅ FLASHED + BOOTING CLEAN on hardware (rev v1.3, 2026-06-13). Boot log: PSRAM 200MHz, display_init OK, UI built, panel live, ZERO underrun. Awaiting Michael's eyeball confirm of the actual screen.
- ✅ Fixes committed+pushed (`a44a177`): flash.ps1 port-clobber + PSRAM 20→200MHz + docs.
- ✅ Michael CONFIRMED it renders ("it works!"). Then 3 follow-ups DONE + flashed (`d1f2df3`, `.claude` untracked `f71f70a`): (1) colors → cpuchip.net palette (lavender chrome), (2) rounded top-left elbow + flat top/footer bar ends, (3) GT911 touch → 4 tappable rail buttons (select+highlight+footer readout). Awaiting Michael's tap-confirm.
- ★ GT911 on THIS board = backup I2C addr **0x14** (primary 0x5D NACKs); fallback handles it. IO40 reset safe (2-lane DSI leaves D3 free).
- Touch coord mapping (mirror/swap) unverified — if taps land wrong, adjust flags in lcars_touch.c.
- ✅ GAME SHIPPED (`186847b`): turned the panel into **SHIELD DEFENSE** — touch resource-management game (attacks fore/aft telegraphed in alert pill; SHLD-F/SHLD-A/REPAIR/RESET buttons; energy economy; 3 lives; score; difficulty ramp; **NVS-persisted high score**). Plus the cpuchip **concave LCARS elbow** (bg-circle overlay carves the inner corner) + button pressed-state feedback. Boots clean (SHIELD DEFENSE started hi=0). Awaiting Michael's playtest.
- Architecture: lcars_ui now presentational (render(state) + button-handler hook); lcars_game.c = engine (button taps → queue → game task, no shared lock). Files: lcars_game.{c,h} new.
- Tunables (attack rate/dmg/energy) at top of lcars_game.c — likely needs balance tuning after playtest.
- NEXT ideas: game balance tuning; sound (I2S amp on board); Antonio font; bridge-aux WiFi→Empty Epsilon mode (render seam ready).

## Gotchas found this session (2026-06-13, hardware bring-up)
- **esptool not in active python**: the workspace `.venv` shadows global python everywhere (even `py` defaults to it); had to `pip install esptool pyserial` into it. flash.ps1 errors were masked as generic "write_flash failed".
- **flash.ps1 `$port`/`$Port` collision**: PowerShell vars case-insensitive → finding the device clobbered "COM3" with the device object. Fixed → `$portDevice`.
- **PSRAM 200MHz needs `CONFIG_IDF_EXPERIMENTAL_FEATURES`** (IDF 5.4): without it `SPIRAM_SPEED_200M` silently → 20MHz → DSI underrun flood → garbled/black. Authoritative ref = `managed_components/espressif__esp_lcd_ek79007/test_apps/sdkconfig.defaults`.
- Non-interactive serial capture (miniterm blocks): `%TEMP%\serial_capture.py COM3 <secs>` (pyserial, pulses RTS reset then reads). Candidate to commit as a `scripts/` helper if reused.

## Claims
- (none — no long-lived processes; Docker build was one-shot)

## Handoffs / notes
- Picked up firmware/elecrow-p4-10. Host toolchain was validated 2026-04-23; built a NEW app this session.
- KEY architecture fact: panel runs native LVGL in C, NOT Vue. lcars/ Vue theme = design reference only (per bridge-build-guide §5). Built the LCARS look natively in lcars_ui.c using tokens.css palette.
- KEY build risk: enabled 32MB HEX PSRAM in sdkconfig.defaults (vendor Lesson07 ships `# CONFIG_SPIRAM is not set` — but the 1.2MB DSI framebuffer can't fit in 768KB internal RAM; PSRAM-off is the likely reason Lesson07's render was never confirmed). If screen is BLACK on first flash → suspect the PSRAM block first.
- Data seam ready for the WiFi/Empty-Epsilon leg: lcars_ui_set_status(hull/shields/energy/alert) — only mock_task in main.c gets replaced.
- Flash when board's connected: `.\scripts\build.ps1 -App apps/lcars-panel` then `.\scripts\flash.ps1 -App apps/lcars-panel -Monitor`.
