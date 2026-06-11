#!/usr/bin/env python3
"""PostToolUse hook (Bash|PowerShell): auto-log background-process claims.

When a session launches a background command, the claim lands in ITS lane
file — so "who owns persona-host.exe" is a one-line read, never forensics
on process parents again (the 2026-06-11 duplicate-host lesson).
Always silent.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import find_lane_by_session, now_str, read_input, write_lane


def main():
    data = read_input()
    tool_input = data.get("tool_input") or {}
    if not tool_input.get("run_in_background"):
        return
    path, fm, body = find_lane_by_session(data.get("session_id", ""))
    if path is None:
        return
    cmd = (tool_input.get("command") or "").replace("\n", " ")[:160]
    claim = f"- {now_str()} background ({data.get('tool_name', '?')}): {cmd}"
    if "## Claims" in body:
        body = body.replace("## Claims", f"## Claims\n{claim}", 1)
    else:
        body = body.rstrip() + f"\n\n## Claims\n{claim}\n"
    write_lane(path, fm, body)


main()
