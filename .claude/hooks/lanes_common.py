"""Shared helpers for the session-lane hooks (.mind/sessions/ protocol).

Imported by lane_*.py via sys.path trick (they all live in this folder).
Workspace root is resolved from this file's location: .claude/hooks/ -> root.
"""
import datetime
import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
SESSIONS = ROOT / ".mind" / "sessions"
INBOX = SESSIONS / "inbox"
TS_FMT = "%Y-%m-%dT%H:%M:%S"


def utf8_stdio():
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except AttributeError:
        pass


def read_input():
    try:
        return json.load(sys.stdin)
    except Exception:
        return {}


def now_str():
    return datetime.datetime.now().strftime(TS_FMT)


def slugify(title):
    s = re.sub(r"[^a-z0-9]+", "-", title.lower()).strip("-")
    return s or "untitled"


def parse_lane(path):
    """Return (frontmatter dict, body str) — body excludes the frontmatter block."""
    try:
        text = path.read_text(encoding="utf-8")
    except OSError:
        return {}, ""
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        return {}, text
    fm = {}
    for i, ln in enumerate(lines[1:], start=1):
        if ln.strip() == "---":
            return fm, "\n".join(lines[i + 1:]).lstrip("\n")
        if ":" in ln:
            k, v = ln.split(":", 1)
            fm[k.strip()] = v.strip()
    return fm, ""


def write_lane(path, fm, body):
    keys = ["lane", "session_id", "status", "started", "last_active"]
    front = "\n".join(f"{k}: {fm[k]}" for k in keys if k in fm)
    extra = "\n".join(f"{k}: {v}" for k, v in fm.items() if k not in keys)
    if extra:
        front += "\n" + extra
    path.write_text(f"---\n{front}\n---\n\n{body.rstrip()}\n", encoding="utf-8")


def find_lane_by_session(session_id):
    """Return (path, fm, body) of the lane claiming session_id, else (None, {}, '')."""
    if not session_id or not SESSIONS.is_dir():
        return None, {}, ""
    for p in sorted(SESSIONS.glob("*.md")):
        if p.name == "README.md":
            continue
        fm, body = parse_lane(p)
        if fm.get("session_id") == session_id:
            return p, fm, body
    return None, {}, ""


def inbox_count(lane):
    p = INBOX / f"{lane}.md"
    try:
        text = p.read_text(encoding="utf-8")
    except OSError:
        return 0
    return sum(1 for ln in text.splitlines() if ln.startswith("## "))


def emit(event, context):
    """Print hookSpecificOutput with additionalContext (no-op when empty)."""
    if not context:
        return
    print(json.dumps({"hookSpecificOutput": {
        "hookEventName": event, "additionalContext": context}}))
