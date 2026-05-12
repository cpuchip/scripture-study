# Webcam Accessibility in Chrome Kiosk on Five Laptops

**Binding question:** Are the built-in webcams on all five repurposed laptops accessible to Chrome when running under a kiosk-mode user profile, independent of any ML stack validation?

**Project:** —

**Date:** 2026-05-12

---

## The plan

Run a hardware audit on all five laptops before touching kiosk configuration. Open Device Manager and Chrome’s `chrome://media-internals` under the current user to confirm the built-in webcam is enumerated. Document any machine that shows a driver warning or missing device as a hardware failure and remove it from further testing.

Create a local Windows user named `Kiosk`, enable auto-login through `netplwiz` or the Registry, and place a startup shortcut that launches Chrome with `--kiosk --app=<test-page-url>`. This approach works on Windows 11 Home and Pro alike—sidestepping the Assigned Access edition requirement—and gives the fastest path to a binary yes-or-no answer.

Build a single HTML test page that calls `navigator.mediaDevices.getUserMedia({video: true})`, enumerates devices with `navigator.mediaDevices.enumerateDevices()`, and renders the stream to a visible `<video>` element. Host the file locally on each machine or serve it from a static URL so the kiosk session has an immediate visual signal that the camera is alive.

Under the `Kiosk` profile, load the test page and grant camera access if Chrome prompts. Reboot the machine and verify the stream restarts without further interaction. If the permission does not persist, enable the system-wide camera privacy toggle under `Settings > Privacy & security > Camera` for the new user, then investigate Chrome policy registry keys such as `DefaultVideoCaptureSetting` or `VideoCaptureAllowedUrls` to force a silent allow.

Replicate the identical user account, auto-login setting, and Chrome shortcut on the remaining four laptops. Record the results in a pass/fail table that lists machine identifier, hardware detected, kiosk stream recovered after reboot, and notes. This table becomes the final answer to the binding question and unblocks downstream ML validation work.

## Assumptions

- **Windows edition is unspecified.** The mechanism must function on both Windows 11 Home and Pro; therefore Assigned Access is excluded.
- **Kiosk-mode means a dedicated local user with auto-login launching Chrome via `--kiosk`.** This excludes Windows Assigned Access, Chrome Enterprise policies, and third-party kiosk shells.
- **The `Kiosk` user profile is a standard, non-ephemeral local account.** Chrome site permissions and Windows privacy toggles, once set, persist across reboots.
- **Built-in webcams and inbox drivers are physically functional.** Any laptop whose camera does not appear in Device Manager is logged as a hardware failure and removed from further validation.
- **"Accessible" means Chrome enumerates the webcam and `getUserMedia` acquires a video stream.** A one-time permission grant during setup is acceptable; a prompt on every boot is not.

## Risks

- **Windows Camera privacy toggle defaults to off for new users.** The system-wide camera access switch may block Chrome entirely until it is manually enabled during setup.
- **Chrome on Windows does not auto-grant camera permissions like ChromeOS kiosk.** A grant that appears persistent in one session may evaporate after reboot. Watch for this during the reboot validation step; if it occurs, registry policy keys or flags are required.
- **Hardware or driver failure on one or more laptops.** The hardware audit catches this, but if even one machine fails, the "all five" criteria fails. A decision to source an external USB webcam or retire the machine must be surfaced immediately.
- **Windows 11 Home vs. Pro mismatch against future lock-down requirements.** Auto-login plus `--kiosk` answers the webcam question on any edition, but it is less tamper-resistant than Assigned Access. If the broader exhibit security context later demands Pro-only features, the fleet may need re-imaging or replacement.

## Next steps

Run the hardware audit on all five laptops first. Configure the `Kiosk` user, auto-login, and Chrome `--kiosk` shortcut on a single pilot machine and build the local test page. Validate that the camera stream survives a reboot without interaction, resolve any privacy-toggle or policy blocks, and replicate the exact configuration to the remaining four machines while recording the pass/fail table.