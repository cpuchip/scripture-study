#!/usr/bin/env python3
"""PostToolUse hook: per-session re-grounding counter.

Counts tool uses PER SESSION (keyed by session_id) and, every THRESHOLD uses,
nudges the model to re-read its durable files. Per-session is the whole point:
Michael runs up to ~6 concurrent sessions against this one repo. A single shared
counter (the earlier cwd-relative / project-global designs) would fire ~Nx too
often and race on one file — session A getting a reground reminder because
sessions B–F did the tool calls. Each session_id gets its own counter under the
gitignored .claude/cache/ (see lanes_common.reground_counter).

Always safe: never raises, never blocks the tool pipeline. If session_id is
missing it falls back to a single 'default' bucket rather than crashing.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lanes_common import emit, read_input, reground_counter

THRESHOLD = 50
MSG = (
    "Re-grounding check: 50+ tool uses since last refresh. Re-read "
    ".mind/active.md (the in-flight board), .spec/covenant.yaml, and intent.yaml "
    "before continuing; also glance at your session lane + inbox "
    "(.mind/sessions/). Are your current actions still aligned with the intent? "
    "If you have drifted, course-correct now."
)


def main():
    data = read_input()
    path = reground_counter(data.get("session_id", ""))
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        try:
            n = int(path.read_text(encoding="utf-8").strip())
        except (OSError, ValueError):
            n = 0
        n += 1
        if n >= THRESHOLD:
            path.write_text("0\n", encoding="utf-8")
            emit("PostToolUse", MSG)
        else:
            path.write_text(f"{n}\n", encoding="utf-8")
    except Exception:
        pass  # a counter must never break a tool call


main()
