#!/usr/bin/env python3
"""SessionStart hook: ground the session + claim its lane.

Replaces the original heredoc grounding hook — the canonical re-read
instruction lives here now, extended with the session-lane protocol
(2026-06-11, born from the duplicate-persona-host incident).
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import (SESSIONS, emit, find_lane_by_session, inbox_count,
                          now_str, parse_lane, read_input, slugify,
                          utf8_stdio, write_lane)

GROUNDING = (
    "Re-read intent.yaml, .spec/covenant.yaml, .mind/identity.md, "
    ".mind/active.md, and .mind/principles.md (if it exists) NOW before "
    "responding. Your grounding in these shapes everything that follows in "
    "this session. Use the Read tool, do not work from memory. "
    "NOTE: .mind/active.md is a lean in-flight board (closed arcs live in "
    ".spec/journal/ and .mind/archive/). Multi-session coordination lives in "
    ".mind/sessions/ — read its README once, keep your lane file current, "
    "and check the lanes before touching long-lived processes you didn't start."
)


def main():
    utf8_stdio()
    data = read_input()
    sid = data.get("session_id", "")
    title = (data.get("session_title") or "").strip()

    lane_line = ""
    if title:
        lane = slugify(title)
        SESSIONS.mkdir(parents=True, exist_ok=True)
        path = SESSIONS / f"{lane}.md"
        fm, body = parse_lane(path) if path.exists() else ({}, "")
        fm.update({"lane": lane, "session_id": sid, "status": "active",
                   "last_active": now_str()})
        fm.setdefault("started", now_str())
        if not body:
            body = "## Working on\n\n## Claims\n\n## Handoffs / notes\n"
        write_lane(path, fm, body)
        lane_line = f" Your lane: .mind/sessions/{lane}.md (claimed)."
        n = inbox_count(lane)
        if n:
            lane_line += (f" 📬 {n} signal(s) waiting — read "
                          f".mind/sessions/inbox/{lane}.md, act, then clear.")
    elif sid:
        existing, fm, _ = find_lane_by_session(sid)
        if existing:
            lane_line = f" Your lane: .mind/sessions/{existing.name}."
        else:
            lane_line = (f" This session has no title — /rename it, or create "
                         f"your lane per .mind/sessions/README.md with "
                         f"session_id {sid}.")

    emit("SessionStart", GROUNDING + lane_line)


main()
