# 2026-06-30 — loom's trust axis + session lifecycle, complete on the real path

**Lane:** general-workspace. **Arc:** continuing from the north-star council, loom marched through the wall (docker isolation), the reach (remote ssh), and the whole session lifecycle (resume, interrupt+steer, remote+isolate). Michael called it: feature-complete-enough to *use* and learn what it still needs.

## What was done

**The wall — `--isolate` (`ddeccf9`).** claude runs in a real directory with `Bash`/`Read`/`Write`, so a backend commanding it can touch the host. Isolation runs claude inside a docker container (`loom-claude`) that sees only `/work` (the repo) and the credentials file, mounted read-only. The one real fix: mounting all of `~/.claude` read-only blocked claude from writing its own session state; scoping the mount to just `.credentials.json:ro` fixed it, and the container was verified to see a stock Linux fs — no host `C:\Users`.

**The reach — `--remote` (`01ad6ee`, `dfbd988`, `1888e8d`).** ssh transport, the same wrapping pattern as isolate. The pipe worked on the first try — loom reached `cpuchip@172.17.1.230`, ran a command, and streamed the output back. But the far-side PATH was the catch: a non-interactive `ssh host "claude …"` uses a shell that skips the login profile, so claude's `~/.local/bin` install wasn't found. Michael's own independent ssh session was the diagnostic — `claude update` worked there interactively, which proved the binary was installed and the problem was PATH, not absence. Wrapping the remote command in a login shell (`bash -lc`) loads the profile and fixes it. Then it live-verified end to end: a Windows `loom.exe` drove a Claude Code agent on the remote Ubuntu box, its `→ Bash` tool-events streaming back, ~$0.12/turn.

**Durable sessions — `--resume` (`fdea626`).** claude persists a session to disk; `--resume <id>` reattaches. Therefore a session survives a process restart or a dropped ssh pipe — the piece that makes a *remote* session robust. Verified two ways I could run myself: a two-process oracle (process A remembers 73 and exits → a fresh process B `--resume`s the id and recalls 73) and the CLI end to end (`loom run` prints the session id → `loom run --resume` → 88, from a brand-new process).

**Interrupt + steer — `2fd1546`.** The protocol was undocumented, so I ran two things at once: a `claude-code-guide` agent and a live probe. The guide came back guessing `{"type":"interrupt"}`; the probe proved the real binary wants a `control_request` with `subtype:interrupt`, acked by a `control_response` success and terminated by a `result` of `subtype:error_during_execution`. The concurrency refactor split the session mutex — `turnMu` serializes turns, `ioMu` guards stdin writes — so the read loop holds no lock and `Interrupt()` can write while `SendStream` reads (race-checked clean). The oracle: interrupt a running turn (~0s to stop), then `Send "reply ALIVE"` on the still-live session → **ALIVE**, context intact. In the CLI, the first Ctrl-C during a turn now interrupts the agent, not loom.

**Reach + wall — `--remote --isolate` (`b880490`).** A sandboxed claude *on* the remote box. The subtlety was path resolution: docker runs on the far side of ssh, so its volume mounts must resolve there — `$HOME` expanded by the remote login shell, `--dir` a remote path. Built and unit-tested (the composed `ssh → docker` argv, and the no-`--dir`→`$HOME` fallback); live-verify waits on the `loom-claude` image being built on the remote.

## Lessons / surprises

- **The pipe worked; the far-side environment was the catch.** Both remote features had a clean transport and a mundane environment gotcha (PATH for `--remote`, the image-on-remote for `--remote --isolate`). The hard-looking part was easy; the boring part was where the real work hid.
- **A live probe beat the guide agent — [[feedback_verify_via_real_path]] again, on a protocol this time.** The guide's `{"type":"interrupt"}` was plausible and wrong for this binary. The probe settled it in one run against the actual claude. When docs are incomplete, the real tool is the authority, not a confident summary of the docs.
- **Clean division of labor held the covenant.** I verified everything I could reach from here — isolate, resume, interrupt, all local against the real claude — and handed live-verify for `--remote` (needs Michael's ssh-agent) and `--remote --isolate` (needs the remote image) to him with exact recipes. I never reported remote as working from a proxy; the `verify_real_path` clause stayed intact.
- **Two frames are now complete, and they're orthogonal.** The *trust axis* (direct / sandboxed / remote) and the *session lifecycle* (carry / resume / interrupt+steer) cross cleanly — a session at any trust level can be carried, resumed, and interrupted. That grid, not any single feature, is the deliverable.

## Carry-forward

- **The dogfood turn (Michael's framing).** loom is feature-complete-enough to *use* — the next learning comes from driving it on real work and finding the gaps, not from adding more capabilities up front.
- **The obvious first real user is pg-ai-stewards driving a remote claude session.** The open design question the dogfood should answer: what is loom's integration surface with the substrate — a subprocess it spawns, a long-lived service it calls, an MCP server? "loom is the substrate's hands" is settled; how the substrate *holds* loom is not.
- **`--remote --isolate` live-verify** — build `loom-claude` on workchip, then run it.
- **The next wall is honest zero-trust.** Isolation walls the filesystem but the container still holds the OAuth token and has network — a scoped/short-lived token plus egress limits is the real hardening. Named in the roadmap.
- Backlog unchanged: **panel role-routing** (doer→critic), `agy --isolate`, the `--agent`/`--agents` flag nit, `--events` through panel, a condenser for long sessions. Cascade-2 still held for a real coding test.

## Commits

`cpuchip/loom`: `ddeccf9` isolate · `01ad6ee`/`dfbd988`/`1888e8d` remote (+login-shell fix + live-verified) · `fdea626` resume · `2fd1546` interrupt+steer · `b880490` remote+isolate. Memory: `project_loom` (trust axis + session lifecycle + the probe-beats-guide note). Board: `.mind/active.md`. Prior arcs: `2026-06-30-loom-north-star-and-lore-research.md`, `2026-06-29-loom-and-the-agentic-harness-arc.md`.
