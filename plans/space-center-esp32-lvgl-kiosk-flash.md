# SquareLine LVGL Boot Reliability on ESP32-P4

**Binding question:** Does the ESP32-P4 panel reliably boot and display a frozen-version LVGL wayfinding screen using SquareLine Studio exports?

**Project:** —

**Date:** 2026-05-12

---

## The plan

Scaffold a minimal firmware project at `firmware/elecrow-p4-10/squareline-wayfinding/` based on Elecrow's ESP-IDF Lesson 9 base code. Pin LVGL to exactly 9.2.2 in `idf_component.yml` so IDF cannot float to a newer version. Create a SquareLine Studio 1.5.1 project at 1024×600 containing one static screen with placeholder labels and no images. Export the UI into `main/ui/` and wire `ui_init()` into `main.c`.

Configure the SDK exactly per Elecrow's documented SquareLine guide. Enable `CONFIG_IDF_EXPERIMENTAL_FEATURES`, set 200M PSRAM speed, use the vendor-provided larger partition table, and enable the 48-pt font. Build the project with ESP-IDF 5.4.2 on Windows and flash to COM3. This follows the vendor-validated integration path rather than Arduino or ESPHome.

Test cold-boot reliability by power-cycling the panel five consecutive times. Record time-to-visible-UI for each boot. The pass criteria is visible UI on every cold boot. Therefore the deliverable is a working repository folder plus a boot-reliability log documenting each cycle.

## Assumptions

- The 10.1" P4 board behaves identically to Elecrow's documented 7" P4 SquareLine guide for chip-level init (MIPI-DSI, GT911, PSRAM). If it does not, the display may remain white despite correct software configuration.
- "Frozen-version LVGL" means pinning LVGL to 9.2.2 via `idf_component.yml` and locking the SquareLine project to that version, preventing IDF from selecting a later release.
- The vendor-documented ESP-IDF path is required. Arduino and ESPHome are incompatible with SquareLine's export-to-IDF workflow.
- The panel receives supplemental USB power at 5 V/2 A. Power brown-outs are out of scope for this test.
- Boot reliability for this first test means "cold-boots to visible UI 5 consecutive times," not a statistical MTBF study.

## Risks

- **10.1" vs 7" board subtlety.** Elecrow's SquareLine guide targets the 7" variant. The 10.1" may use different MIPI-DSI panel timing or a backlight PWM profile that the BSP handles incorrectly. If the screen stays white, check panel initialization first.
- **LVGL 9.2.2 API mismatch in exported code.** Similar tools have shipped exports that call v8 APIs even when set to v9. If the build succeeds but the screen remains blank, inspect `screens.c` for API calls that do not match the pinned LVGL version.
- **PSRAM init fragility on cold boot.** The 200M PSRAM speed requires `CONFIG_IDF_EXPERIMENTAL_FEATURES`. Vendor tutorials enable it, but cold-boot reliability with this flag is unproven in our environment. If boot fails intermittently, drop PSRAM speed to 120M as the fallback.
- **Partition table surprise.** SquareLine exports with image assets need a larger partition table than default. Elecrow provides a custom table, but a missed or misnamed partition file in `sdkconfig` produces runtime asset-load crashes rather than a clean build error.

## Next steps

Create the SquareLine Studio placeholder project and export the UI sources. Scaffold the ESP-IDF project with pinned LVGL 9.2.2 and apply Elecrow's SDK configuration. Build, flash to COM3, and execute the five-cycle cold-boot test.