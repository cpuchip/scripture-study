#!/usr/bin/env python3
"""Claude Code status line: lane, model, context-window bar, mail, rate limits.

Receives the statusline JSON on stdin (see code.claude.com/docs/en/statusline).
Prints one line: ⟨lane⟩ [Model] <bar> NN% ctx · 📬 N · 5h NN% · 7d NN%
Color thresholds: green <50%, yellow 50-79%, red >=80%.
Lane + mail come from the .mind/sessions/ protocol (one lane file per
session; the inbox badge is how PULL signal delivery stays visible).
"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SESSIONS = ROOT / ".mind" / "sessions"


def lane_for_session(session_id):
    if not session_id or not SESSIONS.is_dir():
        return None
    for p in sorted(SESSIONS.glob("*.md")):
        if p.name == "README.md":
            continue
        try:
            head = p.read_text(encoding="utf-8")[:400]
        except OSError:
            continue
        if f"session_id: {session_id}" in head:
            return p.stem
    return None


def inbox_count(lane):
    try:
        text = (SESSIONS / "inbox" / f"{lane}.md").read_text(encoding="utf-8")
    except OSError:
        return 0
    return sum(1 for ln in text.splitlines() if ln.startswith("## "))


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

    lane = lane_for_session(d.get("session_id", ""))
    if lane:
        parts.append(f"\033[36m⟨{lane}⟩\033[0m")

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

    if lane:
        n = inbox_count(lane)
        if n:
            parts.append(f"\033[35m📬 {n}\033[0m")

    rl = d.get("rate_limits") or {}
    for key, label in (("five_hour", "5h"), ("seven_day", "7d")):
        u = (rl.get(key) or {}).get("used_percentage")
        if u is not None:
            parts.append(f"{label} {int(u)}%")

    print(" · ".join(parts))


main()
