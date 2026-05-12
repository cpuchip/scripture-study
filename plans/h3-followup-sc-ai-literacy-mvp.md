# "Teach the Machine" — MVP AI-Literacy Exhibit

**Binding question:** What is the minimum-viable AI-literacy exhibit we could build for the Marsfield science center in 8 weeks with one staffer, ~$500 in NEW materials, and the existing hardware Michael already has: 5 laptops repurposed from the Bridge Simulator project + one 10" ESP32 panel?

**Project:** space-center

**Date:** 2026-05-12

---

## The plan

Build "Teach the Machine," a software-first exhibit using five repurposed Bridge Simulator laptops as interactive stations and the 10" ESP32 panel as a welcome kiosk. The laptops run browser-based machine learning in Chrome kiosk mode. The ESP32 panel displays a SquareLine-built LVGL interface explaining AI and routing visitors to stations. Physical construction is limited to 3D-printed prop trays, laptop cable-management brackets, and a kiosk stand using existing filament stock. New material spending targets $250 to $350 with a hard ceiling of $500.

Two laptops run Google's Teachable Machine with TensorFlow.js exports for fully offline image classification. Visitors train a model to distinguish 3D-printed space props, then test it against new objects in under a minute. One laptop runs Teachable Machine pose classification where a body gesture triggers a simple rocket-launch animation tuned for toddlers. The fourth laptop hosts "Fix It," a custom offline HTML page where visitors add biased training examples and watch the model fail. The fifth laptop hosts "Spot It," a scavenger-hunt page prompting visitors to identify AI embedded in everyday objects around the exhibit. The ESP32 panel never runs inference. It serves static content and wayfinding only.

All custom software is plain HTML, CSS, and JavaScript served locally. Teachable Machine exports are cached locally to eliminate dependency on venue WiFi. The Bambu Lab X1C prints prop holders, signage mounts, and the kiosk enclosure from existing filament. If a laptop's built-in webcam fails validation, the plan buys one inexpensive USB webcam for that unit. If more than two laptops fail, the exhibit scales to three active stations rather than exceed the $500 ceiling.

One staffer delivers the build in eight weeks at roughly 10 to 15 hours per week. Week 1 validates webcam operation, tests Teachable Machine offline export on all five machines, and confirms the ESP32 panel displays the LVGL UI. Weeks 2 through 4 build the "Fix It" and "Spot It" pages and finalize the kiosk interface. Weeks 5 and 6 integrate stations, print props and brackets, and run failure testing. Weeks 7 and 8 handle on-site installation, final debugging, and revision. No stage depends on external fabrication or long-lead shipping.

## Assumptions

- The five laptops each have a functional built-in webcam capable of running browser-based ML at interactive frame rates. This breaks if any camera is missing, physically damaged, or lacks drivers that Chrome can access under Windows 11.
- "AI literacy" for this audience means understanding that AI learns from examples, can be wrong, and is already embedded in daily life. It does not mean writing Python or studying neural-network math. If the venue expects coding or math, this exhibit misses the mark.
- The 10" ESP32 panel serves as a static welcome kiosk and wayfinding display only. If it must run inference or streaming video, the hardware and timeline fail.
- The exhibit operates on reliable venue WiFi or runs fully offline via TensorFlow.js exports from Teachable Machine. If the venue blocks offline file access or requires always-on internet for policy reasons, the software stack breaks.
- The Bambu Lab X1C can produce prop holders, signage mounts, and the kiosk enclosure from existing filament stock. If filament colors or structural requirements exceed on-hand inventory, the budget must absorb spool replacements.

## Risks

- Teachable Machine requires webcam browser permissions. Windows 11 privacy settings or driver issues could silently block cameras on event day with no IT staff to debug. Test every laptop under a fresh user profile identical to the venue's public setup.
- The ESP32-P4's ESP-IDF v5.4 toolchain is bleeding-edge. A single BSP or LVGL version mismatch could consume days of debug time for a component that represents only a fraction of visitor dwell time. Freeze LVGL and BSP versions in week 1 and do not upgrade.
- Toddlers ages four to six may not grasp the training metaphor. Without an immediate physical reward loop such as a sound or animation triggered by pose, they may disengage. The rocket-launch animation must run at full frame rate with zero perceptible latency.
- One staffer with a day job means only 10 to 15 hours per week. An unexpected laptop hardware failure or Windows update in week 7 leaves zero recovery buffer before the Month 4 pop-up target. Maintain a cold-spare laptop configuration through week 6.

## Next steps

First, validate webcam compatibility and Teachable Machine offline export across all five laptops. Second, build the "Fix It" bias-demo HTML page and the "Spot It" scavenger-hunt HTML page with local offline serving. Third, design and flash the LVGL kiosk UI to the ESP32-P4 panel using SquareLine Studio with frozen dependency versions. Fourth, 3D-print prop trays, signage mounts, laptop cable-management brackets, and the kiosk enclosure from existing filament. Fifth, integrate all stations, run failure testing under public-user conditions, and deploy on-site at Marsfield.