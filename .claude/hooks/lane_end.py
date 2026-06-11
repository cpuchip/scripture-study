#!/usr/bin/env python3
"""SessionEnd hook: mark this session's lane ended (claims become suspect)."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import find_lane_by_session, now_str, read_input, write_lane


def main():
    data = read_input()
    path, fm, body = find_lane_by_session(data.get("session_id", ""))
    if path is None:
        return
    fm["status"] = "ended"
    fm["last_active"] = now_str()
    write_lane(path, fm, body)


main()
