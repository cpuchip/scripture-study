#!/usr/bin/env python3
"""UserPromptSubmit hook: lane heartbeat + pull-delivery of signals.

Per prompt: stamps last_active on this session's lane, then nudges the model
when (a) its inbox has mail, (b) the shared board changed since this lane's
previous heartbeat (a sibling session wrote it), or (c) no lane claims this
session yet (bootstrap for sessions already open when the protocol landed).
Silent when there is nothing to say.
"""
import datetime
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import (ROOT, TS_FMT, emit, find_lane_by_session,
                          inbox_count, now_str, read_input, utf8_stdio,
                          write_lane)


def main():
    utf8_stdio()
    data = read_input()
    sid = data.get("session_id", "")
    path, fm, body = find_lane_by_session(sid)

    if path is None:
        if sid:
            emit("UserPromptSubmit",
                 f"(Session-lane protocol: no lane in .mind/sessions/ claims "
                 f"session_id {sid}. If you are mid-task, create your lane — "
                 f".mind/sessions/<topic-slug>.md per the README — stamping "
                 f"this session_id, then continue with the user's request.)")
        return

    prev = fm.get("last_active", "")
    fm["last_active"] = now_str()
    fm["status"] = "active"
    write_lane(path, fm, body)

    notes = []
    n = inbox_count(fm.get("lane", path.stem))
    if n:
        notes.append(f"📬 {n} signal(s) in .mind/sessions/inbox/{path.stem}.md "
                     f"— read, act, then clear them.")
    try:
        board_mtime = (ROOT / ".mind" / "active.md").stat().st_mtime
        prev_ts = datetime.datetime.strptime(prev, TS_FMT).timestamp()
        if board_mtime > prev_ts:
            notes.append("The shared board (.mind/active.md) changed since "
                         "your last activity — a sibling session may have "
                         "written; re-read it before you next write to it.")
    except (OSError, ValueError):
        pass

    if notes:
        emit("UserPromptSubmit", "(" + " ".join(notes) + ")")


main()
