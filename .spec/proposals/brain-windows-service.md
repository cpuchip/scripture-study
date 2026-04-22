---
workstream: WS2
status: proposed
brain_project: 6
created: 2026-04-02
last_updated: 2026-04-21
---

# Brain Windows Service: Systray App with Auto-Start

**Created:** 2026-04-02
**Status:** Draft — awaiting review

---

## Problem Statement

brain.exe runs as a foreground terminal process. Michael has to manually start it, and it dies when the terminal closes. He wants:

1. **Auto-start on login** — brain is ready when he sits down
2. **Easy to kill** — right-click → Exit, or kill the process
3. **Easy to disable** — toggle autostart without uninstalling
4. **Easy to update** — rebuild binary, restart, done

---

## Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Windows Service** (`golang.org/x/sys/windows/svc`) | Starts at boot, managed by `sc.exe`, auto-restart on crash | Requires admin to install, no user-level systray, runs before login, overkill for single-user | **No** |
| **Task Scheduler** (on login trigger) | Easy to set up, no code changes, can be disabled in UI | No systray presence, no way to stop/restart without Task Scheduler, invisible | **No** |
| **Systray wrapper** (separate binary) | Visible icon, start/stop/restart, Open Web UI, auto-start via registry | New binary to maintain, extra process | **Possible** |
| **Systray built into brain.exe** | Single binary, visible icon, all controls in right-click menu, auto-start via registry | Adds systray dependency, need to handle no-display environments (MCP mode) | **Yes** |

### Recommendation: Systray Built Into brain.exe

Add a `--systray` flag (or make it the default on Windows). When running in systray mode:
- brain.exe starts minimized to system tray
- Systray icon shows status (green = running, yellow = starting, red = error)
- Right-click menu: **Open Web UI** | **Restart** | **Disable Autostart** | **Exit**
- Registers itself in `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run` for auto-start on login
- Logs to file instead of stdout (when no terminal attached)

When running without `--systray` (e.g. from terminal, MCP mode, `brain exec`), behavior is unchanged.

---

## Design

### New Dependency

Use [`fyne.io/systray`](https://github.com/fyne-io/systray) — actively maintained fork (last commit: 2 weeks ago, 2026-03-19). The original `getlantern/systray` hasn't been updated in 3 years.

**CGo requirement:** Both libraries require `CGO_ENABLED=1`. On Windows, this means a C compiler (mingw-w64 via `choco install mingw` or MSYS2). brain.exe currently builds without CGo — this is an additive build requirement. Build tag isolation (`//go:build windows`) keeps the CGo dependency out of MCP and exec modes, but the binary will need CGo for the main build.

**Build flag:** Windows systray binaries need `-ldflags "-H=windowsgui"` to suppress the console window.

### Subcommand Structure

```
brain                   # default: start daemon (existing behavior)
brain --systray         # start daemon + systray icon
brain mcp               # MCP server mode (existing)
brain exec              # one-shot execution (existing)
brain install           # register autostart in Windows Registry
brain uninstall         # remove autostart registration
```

### Implementation

```go
// cmd/brain/systray.go (build-tagged for windows)

func runWithSystray() {
    systray.Run(onSystrayReady, onSystrayExit)
}

func onSystrayReady() {
    systray.SetIcon(iconData)
    systray.SetTitle("Brain")
    systray.SetTooltip("brain.exe — running")

    mOpen := systray.AddMenuItem("Open Web UI", "Open brain in browser")
    systray.AddSeparator()
    mStatus := systray.AddMenuItem("Status: Running", "")
    mStatus.Disable() // read-only status line
    mSessions := systray.AddMenuItem("Active Sessions: 0", "")
    mSessions.Disable()
    mEntries := systray.AddMenuItem("Pending Routes: 0", "")
    mEntries.Disable()
    systray.AddSeparator()
    mRestart := systray.AddMenuItem("Restart", "Restart brain daemon")
    mAutostart := systray.AddMenuItemCheckbox("Start on Login", "Auto-start when you sign in", isAutostartEnabled())
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Exit", "Stop brain and exit")

    // Start the daemon in a goroutine
    go func() {
        if err := runDaemon(); err != nil {
            systray.SetTooltip(fmt.Sprintf("brain.exe — error: %v", err))
        }
    }()

    // Handle menu clicks
    for {
        select {
        case <-mOpen.ClickedCh:
            openBrowser("http://localhost:" + port)
        case <-mRestart.ClickedCh:
            restartDaemon()
        case <-mAutostart.ClickedCh:
            toggleAutostart(mAutostart)
        case <-mQuit.ClickedCh:
            systray.Quit()
        }
    }
}
```

### Autostart Registration

```go
// cmd/brain/autostart_windows.go

const regKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
const regValue = "Brain"

func installAutostart() error {
    exe, _ := os.Executable()
    k, _, err := registry.CreateKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
    if err != nil {
        return err
    }
    defer k.Close()
    return k.SetStringValue(regValue, fmt.Sprintf(`"%s" --systray`, exe))
}

func removeAutostart() error {
    k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
    if err != nil {
        return err
    }
    defer k.Close()
    return k.DeleteValue(regValue)
}

func isAutostartEnabled() bool {
    k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
    if err != nil {
        return false
    }
    defer k.Close()
    _, _, err = k.GetStringValue(regValue)
    return err == nil
}
```

### Logging

When `--systray` is set and no terminal is attached, redirect log output to a file:

```go
logPath := filepath.Join(cfg.BrainDataDir, "brain.log")
f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
log.SetOutput(f)
```

Rotate or truncate on startup if > 10MB.

### Update Flow

1. Right-click → Exit (or `taskkill /IM brain.exe`)
2. Rebuild: `cd scripts/brain && go build -o brain.exe ./cmd/brain/`
3. brain.exe auto-starts on next login, or run `brain --systray` manually

For hot-reload during development: the web UI's shutdown endpoint (`/api/shutdown`) already exists. Could add a "Check for Updates" menu item that pulls latest binary from a known path and restarts.

---

## Phased Delivery

### Phase 1: Basic Systray (1 session)

1. Add `getlantern/systray` dependency
2. `--systray` flag on `brain` command
3. Systray icon with Open Web UI / Exit
4. File logging when in systray mode
5. Verify: `brain --systray` shows tray icon, `brain` without flag works as before

### Phase 2: Autostart + Controls (1 session)

1. `brain install` / `brain uninstall` subcommands
2. "Start on Login" checkbox in systray menu
3. "Restart" menu item
4. Status tooltip (running / error / starting)
5. Verify: install → reboot → brain starts automatically → right-click → Exit kills it

---

## Costs & Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| CGo dependency from systray | Medium | Both systray libs require CGo. Install mingw-w64 via `choco install mingw`. Build-tag isolation keeps CGo out of non-systray builds. |
| Build complexity (icon embedding, build tags) | Low | Single `//go:embed brain.ico` and `//go:build windows` tags |
| Conflicting instances | Low | Check if port is already in use at startup, show error in tooltip |
| MCP mode conflict | None | `--systray` is opt-in. MCP and exec subcommands never use it. |

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | brain should be always-on. Michael shouldn't have to remember to start it. |
| Covenant | Rules? | Easy to kill. Easy to disable. No admin required. No background surprises. |
| Line upon Line | Phasing? | Phase 1 (basic tray) → Phase 2 (autostart). Each stands alone. |
| Sabbath | When stop? | After Phase 1 — does it feel right to have brain always present? |
