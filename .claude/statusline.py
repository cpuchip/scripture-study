#!/usr/bin/env python3
"""Claude Code status line: model, context-window usage bar, rate limits.

Receives the statusline JSON on stdin (see code.claude.com/docs/en/statusline).
Prints one line: [Model] <bar> NN% ctx . 5h NN% . 7d NN%
Color thresholds: green <50%, yellow 50-79%, red >=80%.
"""
import json
import sys


def main():
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except AttributeError:
        pass

    try:
        d = json.load(sys.stdin)
    except Exception:
        print("[statusline: no input]")
        return

    parts = []

    model = (d.get("model") or {}).get("display_name")
    if model:
        parts.append(f"[{model}]")

    cw = d.get("context_window") or {}
    pct = cw.get("used_percentage")
    if pct is None:
        # null before the first API call and right after /compact
        parts.append("ctx —")
    else:
        p = max(0, min(100, int(pct)))
        bar = "▓" * (p // 10) + "░" * (10 - p // 10)
        color = "\033[32m" if p < 50 else "\033[33m" if p < 80 else "\033[31m"
        parts.append(f"{color}{bar} {p}% ctx\033[0m")

    rl = d.get("rate_limits") or {}
    for key, label in (("five_hour", "5h"), ("seven_day", "7d")):
        u = (rl.get(key) or {}).get("used_percentage")
        if u is not None:
            parts.append(f"{label} {int(u)}%")

    print(" · ".join(parts))


main()
