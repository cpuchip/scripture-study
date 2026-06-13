#!/usr/bin/env python3
"""SessionEnd hook: mark this session's lane ended (claims become suspect)."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import (find_lane_by_session, now_str, read_input,
                          reground_counter, write_lane)


def main():
    data = read_input()
    sid = data.get("session_id", "")
    path, fm, body = find_lane_by_session(sid)
    if path is not None:
        fm["status"] = "ended"
        fm["last_active"] = now_str()
        write_lane(path, fm, body)
    # Prune this session's reground counter so .claude/cache/ doesn't accumulate
    # one stale file per ended session.
    if sid:
        try:
            reground_counter(sid).unlink(missing_ok=True)
        except OSError:
            pass


main()
